package maps

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

var ErrCaptcha = errors.New("maps: captcha")

const BaseGoogleMaps = "https://www.google.com"

const PausaPadrao = 3 * time.Second

// timeoutElemento teta WaitVisible/Click pra nunca travar pra sempre; subir o valor não reduz a
// falha na aba Avaliações (round 2: falha o teto inteiro, não uns segundos a mais — seletor
// ausente, não lentidão). var, não const: o teste da fixture "sem aba" encolhe isso.
var timeoutElemento = 15 * time.Second

// timeoutChamada teta a função inteira (Buscar/Abrir/AvaliacoesDoAberto), não só waits/clicks: o
// ctx de main.go não tem deadline própria, então sem isso um Navigate/Evaluate que nunca responde
// trava a rodada pra sempre. var: o teste de prazo encolhido usa isso.
var timeoutChamada = 60 * time.Second

// esperarClicar roda uma ação de espera/clique com teto próprio, sem depender do contexto do chamador ter deadline.
func esperarClicar(ctx context.Context, acao chromedp.Action) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutElemento)
	defer cancel()
	return chromedp.Run(ctx, acao)
}

// esperaAposClique e esperaAposRolagem são vars (não const) pra os testes encolherem e o
// fixture não levar minutos pra rodar.
var (
	esperaAposClique  = 2 * time.Second
	esperaAposRolagem = 1500 * time.Millisecond
)

type Resultado struct {
	ID, Nome, URL string
}

type Lugar struct {
	ID, Nome, Endereco, Telefone, Site string
	Fechado                            bool
	Nota                               float64
	// -1 é "desconhecida": a aba Avaliações não apareceu (visão signed-out às vezes reduz a só
	// "Visão geral"/"Sobre", sem nenhum sinal de contagem em lugar nenhum — achado no round 3).
	Avaliacoes int
}

type Coletor struct {
	Base string
	// Pausa zero usa o padrão (PausaPadrao + sorteio); negativo desliga a pausa (só em teste).
	Pausa time.Duration
}

func (c *Coletor) base() string {
	if c.Base == "" {
		return BaseGoogleMaps
	}
	return c.Base
}

// pausar soma até 1s de sorteio à pausa configurada. Zero (Coletor{} não configurado) cai no
// padrão de produção — esquecer de setar Pausa não pode desligar a única defesa contra captcha.
// Negativo desliga a pausa de propósito: é o que os testes usam pra rodar rápido.
func pausar(base time.Duration) time.Duration {
	if base < 0 {
		return 0
	}
	if base == 0 {
		base = PausaPadrao
	}
	return base + time.Duration(rand.Int64N(int64(time.Second)))
}

func Novo(ctx context.Context) (context.Context, context.CancelFunc, error) {
	// 60s em vez dos 20s padrão: Chrome frio no runner do CI, com dois pacotes subindo juntos, passa dos 20s.
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("lang", "pt-BR"), chromedp.WSURLReadTimeout(60*time.Second))
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	cancelar := func() {
		cancelBrowser()
		cancelAlloc()
	}
	// aloca o browser aqui, sem timeout: se a 1ª chamada de chromedp.Run fosse a de Buscar/Abrir, o
	// fim do timeoutChamada dela derrubaria o browser inteiro (doc do Run).
	if err := chromedp.Run(browserCtx); err != nil {
		cancelar()
		return nil, nil, fmt.Errorf("maps: não consegui abrir o Chrome: %w", err)
	}
	return browserCtx, cancelar, nil
}

func (c *Coletor) checarCaptcha(ctx context.Context) (bool, error) {
	var urlAtual, texto string
	if err := chromedp.Run(ctx,
		chromedp.Location(&urlAtual),
		chromedp.Evaluate(jsTextoDoBody, &texto),
	); err != nil {
		return false, err
	}
	return ehCaptcha(urlAtual, texto), nil
}

type resultadoBruto struct {
	Href string `json:"href"`
	Nome string `json:"nome"`
}

// Buscar rola o feed até ter max resultados ou até ele parar de crescer duas vezes seguidas.
func (c *Coletor) Buscar(ctx context.Context, consulta string, max int) ([]Resultado, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutChamada)
	defer cancel()

	if p := pausar(c.Pausa); p > 0 {
		if err := chromedp.Run(ctx, chromedp.Sleep(p)); err != nil {
			return nil, err
		}
	}

	urlBusca := c.base() + "/maps/search/" + url.PathEscape(consulta) + "?hl=pt-BR"
	if err := chromedp.Run(ctx, chromedp.Navigate(urlBusca)); err != nil {
		return nil, err
	}
	if capturado, err := c.checarCaptcha(ctx); err != nil {
		return nil, err
	} else if capturado {
		return nil, ErrCaptcha
	}
	if err := esperarClicar(ctx, chromedp.WaitVisible(selFeed, chromedp.ByQuery)); err != nil {
		// sem feed: às vezes o Maps pula direto pro lugar (resultado único), às vezes é "nenhum
		// resultado" — nenhum dos dois é falha nossa, então não propaga o erro do WaitVisible.
		return c.resultadoUnicoOuVazio(ctx)
	}

	anterior, paradas := 0, 0
	for {
		var atual int
		js := fmt.Sprintf(jsRolarFeedTpl, esperaAposRolagem.Milliseconds())
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &atual, comAwaitPromise)); err != nil {
			return nil, err
		}
		if atual >= max {
			break
		}
		if atual <= anterior {
			paradas++
			if paradas >= 2 {
				break
			}
		} else {
			paradas = 0
		}
		anterior = atual
	}

	var brutos []resultadoBruto
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsExtrairResultados, &brutos)); err != nil {
		return nil, err
	}

	var paginaURL string
	if err := chromedp.Run(ctx, chromedp.Location(&paginaURL)); err != nil {
		return nil, err
	}
	base, _ := url.Parse(paginaURL)

	vistos := map[string]bool{}
	var resultados []Resultado
	for _, b := range brutos {
		id := idDoLink(b.Href)
		if id == "" || vistos[id] {
			continue
		}
		vistos[id] = true
		resultados = append(resultados, Resultado{ID: id, Nome: b.Nome, URL: resolverURL(base, b.Href)})
		if len(resultados) >= max {
			break
		}
	}
	return resultados, nil
}

// resultadoUnicoOuVazio: página de lugar vira resultado único; qualquer outra coisa é "nenhum resultado", não erro.
func (c *Coletor) resultadoUnicoOuVazio(ctx context.Context) ([]Resultado, error) {
	var urlAtual string
	if err := chromedp.Run(ctx, chromedp.Location(&urlAtual)); err != nil {
		return nil, err
	}
	id := idDoLink(urlAtual)
	if id == "" || !strings.Contains(urlAtual, "/maps/place/") {
		return nil, nil
	}
	var nome string
	if err := esperarClicar(ctx, chromedp.Text(selH1, &nome, chromedp.ByQuery)); err != nil {
		return nil, nil // h1 não carregou a tempo: trata como sem resultado, não como erro do lugar
	}
	return []Resultado{{ID: id, Nome: nome, URL: urlAtual}}, nil
}

func resolverURL(base *url.URL, href string) string {
	ref, err := url.Parse(href)
	if err != nil || base == nil {
		return href
	}
	return base.ResolveReference(ref).String()
}

type lugarBruto struct {
	Nome         string `json:"nome"`
	NotaTexto    string `json:"notaTexto"`
	TelefoneItem string `json:"telefoneItem"`
	Site         string `json:"site"`
	Endereco     string `json:"endereco"`
	Fechado      bool   `json:"fechado"`
}

// Abrir também clica na aba Avaliações (só lá tem a contagem, pra quem não está logado). Se a aba
// não aparecer no teto, não é falha do lugar: devolve Avaliacoes -1 (R12), sem erro — por isso
// AvaliacoesDoAberto, chamada logo depois, não clica de novo.
func (c *Coletor) Abrir(ctx context.Context, destino string) (Lugar, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutChamada)
	defer cancel()

	if p := pausar(c.Pausa); p > 0 {
		if err := chromedp.Run(ctx, chromedp.Sleep(p)); err != nil {
			return Lugar{}, err
		}
	}

	if err := chromedp.Run(ctx, chromedp.Navigate(destino)); err != nil {
		return Lugar{}, err
	}
	if capturado, err := c.checarCaptcha(ctx); err != nil {
		return Lugar{}, err
	} else if capturado {
		return Lugar{}, ErrCaptcha
	}
	if err := esperarClicar(ctx, chromedp.WaitVisible(selH1, chromedp.ByQuery)); err != nil {
		return Lugar{}, err
	}

	var bruto lugarBruto
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsExtrairLugar, &bruto)); err != nil {
		return Lugar{}, err
	}

	var urlAtual string
	if err := chromedp.Run(ctx, chromedp.Location(&urlAtual)); err != nil {
		return Lugar{}, err
	}

	lugar := Lugar{
		ID:         idDoLink(urlAtual),
		Nome:       bruto.Nome,
		Endereco:   bruto.Endereco,
		Telefone:   telefoneDoItem(bruto.TelefoneItem),
		Site:       bruto.Site,
		Fechado:    bruto.Fechado,
		Nota:       nota(bruto.NotaTexto),
		Avaliacoes: -1,
	}

	if err := esperarClicar(ctx, chromedp.Click(selAbaAvaliacoes, chromedp.ByQuery)); err != nil {
		return lugar, nil
	}
	if err := chromedp.Run(ctx, chromedp.Sleep(esperaAposClique)); err != nil {
		return Lugar{}, err
	}
	if capturado, err := c.checarCaptcha(ctx); err != nil {
		return Lugar{}, err
	} else if capturado {
		return Lugar{}, ErrCaptcha
	}

	var partesHistograma []string
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsContagemAvaliacoes, &partesHistograma)); err != nil {
		return Lugar{}, err
	}
	avaliacoes := 0
	for _, p := range partesHistograma {
		avaliacoes += contagem(p)
	}
	lugar.Avaliacoes = avaliacoes

	return lugar, nil
}

type avaliacaoBruta struct {
	ID       string `json:"id"`
	Estrelas string `json:"estrelas"`
	Texto    string `json:"texto"`
}

// AvaliacoesDoAberto exige que Abrir tenha sido chamada logo antes: é o Abrir que clica na aba.
// ponytail: rola 3x fixo em vez de medir crescimento; teto — lugar com muitas avaliações positivas
// pode não juntar as 10 negativas. Melhorar rolando até parar de crescer, como o loop de Buscar.
func (c *Coletor) AvaliacoesDoAberto(ctx context.Context) (string, []string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutChamada)
	defer cancel()

	var resumo string
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsResumoGemini, &resumo)); err != nil {
		return "", nil, err
	}

	for i := 0; i < 3; i++ {
		var atual int
		js := fmt.Sprintf(jsRolarAvaliacoesTpl, esperaAposRolagem.Milliseconds())
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &atual, comAwaitPromise)); err != nil {
			return "", nil, err
		}
	}

	var brutas []avaliacaoBruta
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsExtrairAvaliacoes, &brutas)); err != nil {
		return "", nil, err
	}

	vistos := map[string]bool{}
	var negativas []string
	for _, b := range brutas {
		if b.ID == "" || vistos[b.ID] {
			continue
		}
		vistos[b.ID] = true
		if n := nota(b.Estrelas); n >= 1 && n <= 3 && len(negativas) < 10 {
			negativas = append(negativas, b.Texto)
		}
	}
	return resumo, negativas, nil
}
