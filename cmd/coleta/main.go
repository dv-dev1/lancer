// Comando coleta busca os leads do dia por nicho × bairro em João Pessoa e imprime a fila pronta.
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dv-dev1/lancer/internal/custo"
	"github.com/dv-dev1/lancer/internal/llm"
	"github.com/dv-dev1/lancer/internal/maps"
	"github.com/dv-dev1/lancer/internal/site"
)

func main() {
	nicho := flag.String("nicho", "", "restaurante|confeitaria|loja|servico")
	bairro := flag.String("bairro", "", "bairro de João Pessoa")
	limite := flag.Int("limite", 10, "quantidade de leads qualificados")
	flag.Parse()

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
		os.Exit(1)
	}

	// 90s: teto de segurança maior que o timeout interno do PageSpeed (60s) no internal/site,
	// que já gerencia seu próprio deadline via contexto — o client não pode cortar antes dele.
	cliente := &http.Client{Timeout: 90 * time.Second}
	conta := &custo.Conta{}

	ctx, cancelNavegador := maps.Novo(context.Background())
	defer cancelNavegador()
	mc := &maps.Coletor{} // Pausa zero já cai no padrão de produção (maps.PausaPadrao + sorteio)
	sc := &site.Checador{HTTP: cliente, PageSpeedBase: "https://www.googleapis.com", Chave: chavePageSpeed, NotaMinima: 0.5}
	lc := &llm.Cliente{Chave: chaveOpenAI, Base: "https://api.openai.com", HTTP: cliente, Conta: conta}
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

	leads, descartes, cm, err := coletar(ctx, mc, sc, lc, r, entrada{
		Nicho:      *nicho,
		Bairro:     *bairro,
		Limite:     *limite,
		Remetente:  remetente,
		PreviewURL: os.Getenv("LANCER_PREVIEW_URL"),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}

	imprimir(os.Stdout, leads, descartes, *conta, cm)
}
