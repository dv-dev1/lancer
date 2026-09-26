package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/url"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/dv-dev1/lancer/internal/custo"
	"github.com/dv-dev1/lancer/internal/lead"
	"github.com/dv-dev1/lancer/internal/llm"
	"github.com/dv-dev1/lancer/internal/maps"
	"github.com/dv-dev1/lancer/internal/site"
)

// termosPorNicho traduz a flag --nicho nos termos de busca no Maps.
var termosPorNicho = map[string][]string{
	"restaurante": {"restaurante", "lanchonete"},
	"confeitaria": {"confeitaria", "doceria"},
	"loja":        {"loja de roupas", "loja de presentes"},
	"servico":     {"barbearia", "salão de beleza", "clínica de estética"},
}

// prioridadeDor decide o GanchoPadrao quando a LLM não devolve gancho (ruling R2):
// a dor de maior peso primeiro, na ordem de Pontuar.
var prioridadeDor = []lead.Dor{lead.SemSite, lead.SiteAgregador, lead.SiteRuim, lead.Reclamacao}

// mensagemCaptcha é o texto exato que a coleta mostra quando o Google barra com captcha.
const mensagemCaptcha = "Google pediu captcha — coleta parada; tente amanhã"

// maxPorBusca é o teto de resultados por termo de busca no feed do Maps.
// ponytail: fixo em 20; sobe a constante se um nicho pedir mais que isso por termo.
const maxPorBusca = 20

type entrada struct {
	Nicho      string
	Bairro     string
	Limite     int
	Remetente  string
	PreviewURL string
}

type descarte struct {
	Nome   string
	Motivo string
}

// contagemMaps registra o uso do Maps: é grátis (navegador, sem API paga), mas ainda vale mostrar na saída.
type contagemMaps struct {
	Buscas         int
	LugaresAbertos int
}

// erroSistemico devolve não-nil só pra falhas que abortam a coleta inteira (hoje, só captcha).
// Qualquer outro erro de um lugar isolado vira descarte — não pode derrubar o resto da rodada.
func erroSistemico(err error) error {
	if errors.Is(err, maps.ErrCaptcha) {
		return errors.New(mensagemCaptcha)
	}
	return nil
}

// coletar busca, filtra e monta a mensagem dos leads do dia. Os cortes baratos (fechado, celular, porte)
// acontecem antes de avaliações e LLM, que são o SKU caro.
func coletar(ctx context.Context, mc *maps.Coletor, sc *site.Checador, lc *llm.Cliente, r *rand.Rand, e entrada) ([]lead.Lead, []descarte, contagemMaps, error) {
	termos, ok := termosPorNicho[e.Nicho]
	if !ok {
		return nil, nil, contagemMaps{}, fmt.Errorf("nicho %q inválido; use restaurante, confeitaria, loja ou servico", e.Nicho)
	}

	resultados, buscas, err := buscarDeduplicados(ctx, mc, termos, e.Bairro)
	if err != nil {
		return nil, nil, contagemMaps{}, err
	}

	var leads []lead.Lead
	var descartes []descarte
	qualificados := 0
	abertos := 0
	for _, res := range resultados {
		if qualificados >= e.Limite {
			break
		}

		lugar, err := mc.Abrir(ctx, res.URL)
		abertos++
		if err != nil {
			if sistemico := erroSistemico(err); sistemico != nil {
				return nil, nil, contagemMaps{}, sistemico
			}
			// falha num lugar isolado (página fora do ar, blip de rede) não pode derrubar a coleta inteira.
			descartes = append(descartes, descarte{res.Nome, fmt.Sprintf("erro: %v", err)})
			continue
		}
		if lugar.Fechado {
			descartes = append(descartes, descarte{lugar.Nome, "fechado"})
			continue
		}
		celular, ok := lead.Celular(lugar.Telefone)
		if !ok {
			descartes = append(descartes, descarte{lugar.Nome, "sem celular"})
			continue
		}
		if !lead.PortePadrao.Passa(lugar.Nota, lugar.Avaliacoes) {
			descartes = append(descartes, descarte{lugar.Nome, "fora do porte"})
			continue
		}
		qualificados++

		dores, detalhes, err := classificarESitear(ctx, sc, lugar)
		if err != nil {
			return nil, nil, contagemMaps{}, err
		}

		resumo, negativas, err := mc.AvaliacoesDoAberto(ctx)
		if err != nil {
			if sistemico := erroSistemico(err); sistemico != nil {
				return nil, nil, contagemMaps{}, sistemico
			}
			descartes = append(descartes, descarte{lugar.Nome, fmt.Sprintf("erro: %v", err)})
			continue
		}
		saida, err := lc.Analisar(ctx, llm.Entrada{Nicho: e.Nicho, Dores: detalhes, Avaliacoes: montarAvaliacoes(resumo, negativas)})
		if err != nil {
			descartes = append(descartes, descarte{lugar.Nome, fmt.Sprintf("erro: %v", err)})
			continue
		}
		if saida.Reclamacao != "" {
			dores = append(dores, lead.Reclamacao)
			detalhes[lead.Reclamacao] = saida.Reclamacao
		}

		if len(dores) == 0 {
			descartes = append(descartes, descarte{lugar.Nome, "sem dor"})
			continue
		}

		gancho := saida.Gancho
		if gancho == "" {
			principal := dorPrincipal(dores)
			gancho = lead.GanchoPadrao(principal, detalhes[principal])
		}

		l := lead.Lead{
			PlaceID:    lugar.ID,
			Nome:       lugar.Nome,
			Nicho:      e.Nicho,
			Bairro:     e.Bairro,
			Telefone:   celular,
			Site:       lugar.Site,
			Nota:       lugar.Nota,
			Avaliacoes: lugar.Avaliacoes,
			Dores:      dores,
			Detalhes:   detalhes,
		}
		l.Pontuacao = lead.Pontuar(l)
		l.Variante = lead.SortearVariante(r)
		l.Slug = lead.Slug(l.Nome, l.PlaceID)
		l.Mensagem = lead.MontarMensagem(e.Remetente, gancho, l.Variante, linkPreview(e.PreviewURL, l.Slug))
		leads = append(leads, l)
	}

	sort.Slice(leads, func(i, j int) bool { return leads[i].Pontuacao > leads[j].Pontuacao })
	return leads, descartes, contagemMaps{Buscas: buscas, LugaresAbertos: abertos}, nil
}

// montarAvaliacoes é o que a LLM lê: o resumo do Gemini (quando existe) na frente das negativas.
func montarAvaliacoes(resumo string, negativas []string) []string {
	var avaliacoes []string
	if resumo != "" {
		avaliacoes = append(avaliacoes, resumo)
	}
	return append(avaliacoes, negativas...)
}

func buscarDeduplicados(ctx context.Context, mc *maps.Coletor, termos []string, bairro string) ([]maps.Resultado, int, error) {
	vistos := map[string]bool{}
	var resultados []maps.Resultado
	buscas := 0
	for _, termo := range termos {
		consulta := fmt.Sprintf("%s em %s, João Pessoa - PB", termo, bairro)
		encontrados, err := mc.Buscar(ctx, consulta, maxPorBusca)
		buscas++
		if err != nil {
			if sistemico := erroSistemico(err); sistemico != nil {
				return nil, buscas, sistemico
			}
			return nil, buscas, err
		}
		for _, res := range encontrados {
			if !vistos[res.ID] {
				vistos[res.ID] = true
				resultados = append(resultados, res)
			}
		}
	}
	return resultados, buscas, nil
}

// classificarESitear aplica a dor de site (sem site, agregador, ou Checar no site próprio).
// Só site próprio (ClassificarSite == "") gasta uma checagem de PageSpeed.
func classificarESitear(ctx context.Context, sc *site.Checador, lugar maps.Lugar) ([]lead.Dor, map[lead.Dor]string, error) {
	var dores []lead.Dor
	detalhes := map[lead.Dor]string{}

	switch dor := lead.ClassificarSite(lugar.Site); dor {
	case lead.SemSite:
		dores = append(dores, lead.SemSite)
		detalhes[lead.SemSite] = "" // precisa existir como chave: é o que informa a LLM que a dor já foi detectada
	case lead.SiteAgregador:
		dores = append(dores, lead.SiteAgregador)
		detalhes[lead.SiteAgregador] = dominioDe(lugar.Site)
	default:
		if ruim, motivo := sc.Checar(ctx, lugar.Site); ruim {
			dores = append(dores, lead.SiteRuim)
			detalhes[lead.SiteRuim] = motivo
		}
	}
	return dores, detalhes, nil
}

// dominioDe extrai só o domínio do agregador (ex.: "instagram.com"), nunca a URL completa do perfil.
func dominioDe(uri string) string {
	if !strings.Contains(uri, "://") {
		uri = "https://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	return strings.TrimPrefix(u.Hostname(), "www.")
}

func dorPrincipal(dores []lead.Dor) lead.Dor {
	presentes := map[lead.Dor]bool{}
	for _, d := range dores {
		presentes[d] = true
	}
	for _, d := range prioridadeDor {
		if presentes[d] {
			return d
		}
	}
	return dores[0]
}

// linkPreview cai no placeholder "<preview>" quando LANCER_PREVIEW_URL não foi configurada.
func linkPreview(previewURL, slug string) string {
	base := previewURL
	if base == "" {
		base = "<preview>"
	}
	return base + "/p/" + slug
}

func imprimir(w io.Writer, leads []lead.Lead, descartes []descarte, conta custo.Conta, cm contagemMaps) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "pontos\tnome\tdores\tnota (n)\tvariante\ttelefone")
	for _, l := range leads {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%.1f (%d)\t%s\t%s\n", l.Pontuacao, l.Nome, doresParaTexto(l.Dores), l.Nota, l.Avaliacoes, l.Variante, l.Telefone)
	}
	tw.Flush()

	for _, l := range leads {
		fmt.Fprintf(w, "\n[%s] %s\n%s\n", l.Nome, l.Slug, l.Mensagem)
	}

	if len(descartes) > 0 {
		fmt.Fprintln(w, "\ndescartes:")
		for _, d := range descartes {
			fmt.Fprintf(w, "- %s: %s\n", d.Nome, d.Motivo)
		}
	}

	fmt.Fprintf(w, "\nMaps: %d buscas, %d lugares abertos (grátis)\n", cm.Buscas, cm.LugaresAbertos)
	fmt.Fprintf(w, "conta: %+v — teto US$ %.2f\n", conta, conta.Teto())
}

func doresParaTexto(dores []lead.Dor) string {
	nomes := make([]string, len(dores))
	for i, d := range dores {
		nomes[i] = string(d)
	}
	return strings.Join(nomes, ",")
}
