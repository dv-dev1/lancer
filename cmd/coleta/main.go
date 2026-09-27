// Comando coleta busca os leads do dia por nicho × bairro em João Pessoa e imprime a fila pronta.
package main

import (
	"context"
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
	bairro := flag.String("bairro", "", "bairro de João Pessoa")
	limite := flag.Int("limite", 10, "quantidade de leads qualificados")
	// ponytail: --dry-run nem lê o banco; reabre lugares já vistos, aceitável num teste manual.
	dryRun := flag.Bool("dry-run", false, "não usa banco: nem lê já vistos, nem grava (fase 1)")
	flag.Parse()

	if *bairro == "" {
		fmt.Fprintln(os.Stderr, "--bairro é obrigatório")
		return 2
	}
	if *limite <= 0 {
		fmt.Fprintln(os.Stderr, "--limite precisa ser maior que zero")
		return 2
	}

	chaveOpenAI := os.Getenv("OPENAI_API_KEY")
	remetente := os.Getenv("LANCER_REMETENTE")
	chavePageSpeed := os.Getenv("PAGESPEED_API_KEY") // opcional: PageSpeed funciona sem chave em volume baixo

	var faltando []string
	if chaveOpenAI == "" {
		faltando = append(faltando, "OPENAI_API_KEY")
	}
	if remetente == "" {
		faltando = append(faltando, "LANCER_REMETENTE")
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

	// 90s: teto de segurança maior que o timeout interno do PageSpeed (60s) no internal/site,
	// que já gerencia seu próprio deadline via contexto — o client não pode cortar antes dele.
	cliente := &http.Client{Timeout: 90 * time.Second}
	conta := &custo.Conta{}

	ctx, cancelNavegador, err := maps.Novo(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	defer cancelNavegador()
	mc := &maps.Coletor{} // Pausa zero já cai no padrão de produção (maps.PausaPadrao + sorteio)
	sc := &site.Checador{HTTP: cliente, PageSpeedBase: "https://www.googleapis.com", Chave: chavePageSpeed, NotaMinima: 0.5}
	lc := &llm.Cliente{Chave: chaveOpenAI, Base: "https://api.openai.com", HTTP: cliente, Conta: conta}
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

	e := entrada{
		Nicho:      *nicho,
		Bairro:     *bairro,
		Limite:     *limite,
		Remetente:  remetente,
		PreviewURL: os.Getenv("LANCER_PREVIEW_URL"),
	}
	if b != nil {
		e.JaVistos = b.JaVistos
	}

	leads, descartes, cm, err := coletar(ctx, mc, sc, lc, r, e)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}

	// imprime antes de gravar: a coleta do dia não se perde na tela se a gravação falhar depois.
	imprimir(os.Stdout, leads, descartes, *conta, cm)
	if b == nil {
		return 0
	}

	if err := b.Gravar(ctx, leads, paraDescartesBanco(descartes)); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	unidades := int64(conta.TokensEntrada + conta.TokensSaida)
	if err := b.SomarCusto(ctx, time.Now(), "openai", unidades, conta.Teto()); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	if _, err := b.MarcarPerdidos(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	return 0
}

// paraDescartesBanco converte pro tipo de internal/banco: banco não pode importar package main (R18).
func paraDescartesBanco(descartes []descarte) []banco.Descarte {
	bds := make([]banco.Descarte, len(descartes))
	for i, d := range descartes {
		bds[i] = banco.Descarte{PlaceID: d.PlaceID, Motivo: d.Motivo}
	}
	return bds
}
