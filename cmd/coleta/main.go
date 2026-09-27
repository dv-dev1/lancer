// Comando coleta busca os leads por nicho × cidade/bairro da Paraíba e imprime a fila pronta;
// com --servir fica de plantão rodando os pedidos feitos pelo painel.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dv-dev1/lancer/internal/banco"
	"github.com/dv-dev1/lancer/internal/custo"
	"github.com/dv-dev1/lancer/internal/lead"
	"github.com/dv-dev1/lancer/internal/llm"
	"github.com/dv-dev1/lancer/internal/maps"
	"github.com/dv-dev1/lancer/internal/site"
)

// main só chama run() e repassa o código de saída: os.Exit não roda defer, então o cancelamento do
// browser (defer cancelNavegador em run) precisa terminar antes do processo morrer (R10).
func main() {
	os.Exit(run())
}

func run() int {
	nicho := flag.String("nicho", "", "restaurante|confeitaria|loja|servico")
	cidade := flag.String("cidade", cidadePadrao, "cidade da Paraíba")
	bairro := flag.String("bairro", "", "bairro (vazio = a cidade inteira)")
	limite := flag.Int("limite", 10, "quantidade de leads qualificados")
	// ponytail: --dry-run nem lê o banco; reabre lugares já vistos, aceitável num teste manual.
	dryRun := flag.Bool("dry-run", false, "não usa banco: nem lê já vistos, nem grava (fase 1)")
	servirFlag := flag.Bool("servir", false, "fica de plantão rodando os pedidos de coleta feitos pelo painel")
	flag.Parse()

	if !*servirFlag && strings.TrimSpace(*cidade) == "" {
		fmt.Fprintln(os.Stderr, "--cidade não pode ser vazia")
		return 2
	}
	if !*servirFlag && *limite <= 0 {
		fmt.Fprintln(os.Stderr, "--limite precisa ser maior que zero")
		return 2
	}

	cfg := config{
		chaveOpenAI:    os.Getenv("OPENAI_API_KEY"),
		remetente:      os.Getenv("LANCER_REMETENTE"),
		chavePageSpeed: os.Getenv("PAGESPEED_API_KEY"), // opcional: PageSpeed funciona sem chave em volume baixo
		previewURL:     os.Getenv("LANCER_PREVIEW_URL"),
	}
	var faltando []string
	if cfg.chaveOpenAI == "" {
		faltando = append(faltando, "OPENAI_API_KEY")
	}
	if cfg.remetente == "" {
		faltando = append(faltando, "LANCER_REMETENTE")
	}
	if *servirFlag && os.Getenv("DATABASE_URL") == "" {
		faltando = append(faltando, "DATABASE_URL")
	}
	if len(faltando) > 0 {
		fmt.Fprintf(os.Stderr, "faltam variáveis de ambiente: %s\n", strings.Join(faltando, ", "))
		return 1
	}

	// Ctrl-C/SIGTERM cancela ctx em vez de matar o processo na marra: sem isso o Chrome (subido logo
	// abaixo) fica órfão, já que main não roda defer nenhum depois de um os.Exit no meio do caminho.
	ctx, pararSinal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer pararSinal()

	var b *banco.Banco
	if !*dryRun && os.Getenv("DATABASE_URL") != "" {
		var err error
		b, err = banco.Abrir(ctx, os.Getenv("DATABASE_URL"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro:", err)
			return 1
		}
	}

	if *servirFlag {
		return servir(ctx, cfg, b)
	}
	if _, err := rodada(ctx, cfg, b, entrada{Nicho: *nicho, Cidade: *cidade, Bairro: *bairro, Limite: *limite}, 0); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	return 0
}

type config struct {
	chaveOpenAI, remetente, chavePageSpeed, previewURL string
}

// rodada é uma coleta completa, do Chrome até a gravação; b nil = sem banco (fase 1 ou --dry-run).
func rodada(ctx context.Context, cfg config, b *banco.Banco, e entrada, pedido int64) ([]lead.Lead, error) {
	// 90s: teto de segurança maior que o timeout interno do PageSpeed (60s) no internal/site,
	// que já gerencia seu próprio deadline via contexto — o client não pode cortar antes dele.
	cliente := &http.Client{Timeout: 90 * time.Second}
	conta := &custo.Conta{}

	ctx, cancelNavegador, err := maps.Novo(ctx)
	if err != nil {
		return nil, err
	}
	defer cancelNavegador()
	mc := &maps.Coletor{} // Pausa zero já cai no padrão de produção (maps.PausaPadrao + sorteio)
	sc := &site.Checador{HTTP: cliente, PageSpeedBase: "https://www.googleapis.com", Chave: cfg.chavePageSpeed, NotaMinima: 0.5}
	lc := &llm.Cliente{Chave: cfg.chaveOpenAI, Base: "https://api.openai.com", HTTP: cliente, Conta: conta}
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

	e.Remetente, e.PreviewURL = cfg.remetente, cfg.previewURL
	if b != nil {
		e.JaVistos = b.JaVistos
	}

	leads, descartes, cm, err := coletar(ctx, mc, sc, lc, r, e)
	if err != nil {
		return nil, err
	}

	// imprime antes de gravar: a coleta do dia não se perde na tela se a gravação falhar depois.
	imprimir(os.Stdout, leads, descartes, *conta, cm)
	if b == nil {
		return leads, nil
	}

	if err := b.Gravar(ctx, leads, paraDescartesBanco(descartes), pedido); err != nil {
		return nil, err
	}
	unidades := int64(conta.TokensEntrada + conta.TokensSaida)
	if err := b.SomarCusto(ctx, time.Now(), "openai", unidades, conta.Teto()); err != nil {
		return nil, err
	}
	if _, err := b.MarcarPerdidos(ctx); err != nil {
		return nil, err
	}
	return leads, nil
}

// esperaPedido é o intervalo entre consultas à tabela de pedidos quando não há nenhum pendente.
const esperaPedido = 5 * time.Second

// servir roda os pedidos do painel um por vez até Ctrl-C. Captcha para tudo: marca os pendentes
// como erro e sai, porque a trava do dia vale para o plantão também.
func servir(ctx context.Context, cfg config, b *banco.Banco) int {
	// o que ficou "rodando" é de um plantão anterior que morreu no meio.
	if err := b.FalharPedidos(ctx, "coletor interrompido no meio da coleta", "rodando"); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	fmt.Println("coletor de plantão: esperando pedidos do painel (Ctrl-C para sair)")
	for {
		p, err := b.PegarPedido(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return 0
			}
			fmt.Fprintln(os.Stderr, "erro:", err)
			return 1
		}
		if p == nil {
			select {
			case <-ctx.Done():
				return 0
			case <-time.After(esperaPedido):
			}
			continue
		}

		fmt.Printf("\npedido %d: %s em %s, limite %d\n", p.ID, p.Nicho, lugarDaBusca(p.Cidade, p.Bairro), p.Limite)
		leads, err := rodada(ctx, cfg, b, entrada{Nicho: p.Nicho, Cidade: p.Cidade, Bairro: p.Bairro, Limite: p.Limite}, p.ID)
		// contexto novo: com Ctrl-C o ctx já morreu, e o pedido ainda precisa sair de "rodando".
		fim := context.Background()
		switch {
		case errors.Is(err, errCaptcha):
			fmt.Fprintln(os.Stderr, "erro:", err)
			if err := b.FalharPedidos(fim, mensagemCaptcha, "rodando", "pendente"); err != nil {
				fmt.Fprintln(os.Stderr, "erro:", err)
			}
			return 1
		case err != nil:
			fmt.Fprintln(os.Stderr, "erro:", err)
			if err := b.FalharPedidos(fim, err.Error(), "rodando"); err != nil {
				fmt.Fprintln(os.Stderr, "erro:", err)
				return 1
			}
			if ctx.Err() != nil {
				return 0
			}
		default:
			if err := b.ConcluirPedido(fim, p.ID, len(leads)); err != nil {
				fmt.Fprintln(os.Stderr, "erro:", err)
				return 1
			}
		}
	}
}

// paraDescartesBanco converte pro tipo de internal/banco: banco não pode importar package main (R18).
func paraDescartesBanco(descartes []descarte) []banco.Descarte {
	bds := make([]banco.Descarte, len(descartes))
	for i, d := range descartes {
		bds[i] = banco.Descarte{PlaceID: d.PlaceID, Motivo: d.Motivo}
	}
	return bds
}
