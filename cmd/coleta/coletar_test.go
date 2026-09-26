package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dv-dev1/lancer/internal/custo"
	"github.com/dv-dev1/lancer/internal/lead"
	"github.com/dv-dev1/lancer/internal/llm"
	"github.com/dv-dev1/lancer/internal/maps"
	"github.com/dv-dev1/lancer/internal/site"
)

// respostaOpenAI monta o corpo de chat/completions no formato que o cliente llm espera.
func respostaOpenAI(content string, tokensEntrada, tokensSaida int) string {
	b, _ := json.Marshal(content)
	return fmt.Sprintf(`{"choices":[{"message":{"content":%s}}],"usage":{"prompt_tokens":%d,"completion_tokens":%d}}`, b, tokensEntrada, tokensSaida)
}

// temChrome cobre os dois jeitos comuns de achar o binário: no PATH (Linux) ou o .app do macOS.
// Duplicado de internal/maps/maps_test.go: teste de um pacote não importa o _test.go de outro.
func temChrome() bool {
	for _, nome := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
		if _, err := exec.LookPath(nome); err == nil {
			return true
		}
	}
	_, err := os.Stat("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	return err == nil
}

type reviewFix struct {
	ID, Estrelas, Texto string
}

// paginaLugar monta um fixture de página de lugar compatível com os seletores de internal/maps
// (mesma estrutura de internal/maps/maps_test.go, parametrizada pro cenário de cada teste).
func paginaLugar(nome, notaTexto, avaliacoesTexto, telefoneDigitos, siteHref string, fechado bool, resumo string, reviews []reviewFix) string {
	var b strings.Builder
	b.WriteString("<html><body>\n  <h1>" + nome + "</h1>\n")
	if fechado {
		b.WriteString("  <p>Fechado permanentemente</p>\n")
	}
	b.WriteString(`  <span role="img" aria-label="` + notaTexto + ` estrelas">` + notaTexto + "</span>\n")
	b.WriteString(`  <span aria-label="` + avaliacoesTexto + ` avaliações">` + avaliacoesTexto + " avaliações</span>\n")
	if telefoneDigitos != "" {
		b.WriteString(`  <button data-item-id="phone:tel:0` + telefoneDigitos + `">telefone</button>` + "\n")
	}
	if siteHref != "" {
		b.WriteString(`  <a data-item-id="authority" href="` + siteHref + `">site</a>` + "\n")
	}
	b.WriteString(`  <button data-item-id="address" aria-label="Endereço: Rua Teste, 1 - Bairro">Rua Teste, 1</button>` + "\n")
	b.WriteString(`  <button role="tab" aria-label="Avaliações">Avaliações</button>` + "\n")
	if resumo != "" {
		b.WriteString("  <div>\n    <div>\n      <div>\n        <span>" + resumo + "+5</span>\n        <span>Resumo feito com o Gemini</span>\n      </div>\n    </div>\n  </div>\n")
	}
	b.WriteString(`  <div style="overflow-y:auto;height:100px">` + "\n    " + `<button aria-label="Classificar avaliações">Classificar</button>` + "\n")
	for _, rv := range reviews {
		b.WriteString(fmt.Sprintf("    <div data-review-id=%q><span aria-label=\"%s estrelas\">%s estrelas</span><span lang=\"pt\">%s</span></div>\n", rv.ID, rv.Estrelas, rv.Estrelas, rv.Texto))
	}
	b.WriteString("  </div>\n</body></html>")
	return b.String()
}

func feedItem(id, nome string) string {
	return fmt.Sprintf(`<a href="/maps/place/x/data=!19s%s" aria-label="%s">%s</a>`, id, nome, nome)
}

func feedHTML(itens ...string) string {
	return "<html><body><div role=\"feed\">\n" + strings.Join(itens, "\n") + "\n</div></body></html>"
}

func TestColetarFluxoCompleto(t *testing.T) {
	if !temChrome() {
		t.Skip("Chrome/Chromium não encontrado, pulando teste com navegador real")
	}

	// site próprio do id-A: no ar, HTTPS OK; o PageSpeed dele cai (testa "PageSpeed fora do ar não descarta o lead").
	siteProprio := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer siteProprio.Close()

	pageSpeed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer pageSpeed.Close()

	aberturas := map[string]int{}
	lugares := map[string]string{
		"id-C": paginaLugar("Confeitaria Fechada", "4,5", "100", "83999990001", "", true, "", nil),
		"id-D": paginaLugar("Confeitaria Sem Celular", "4,5", "100", "", "", false, "", nil),
		"id-A": paginaLugar("Doceria A", "4,5", "100", "83998765432", siteProprio.URL, false,
			"Clientes elogiam o atendimento e reclamam da demora no whatsapp.",
			[]reviewFix{{"r1", "5", "Muito bom, recomendo!"}, {"r2", "2", "demoraram pra responder no whatsapp"}}),
		"id-B": paginaLugar("Doceria B", "4,0", "30", "83988887777", "", false, "", nil),
		"id-E": paginaLugar("Doceria E", "4,5", "100", "83977776666", "", false, "", nil),
	}
	// mesmo id-A nos dois termos testa o dedup; id-E só existe pra provar que --limite 2 para antes de abri-lo.
	feeds := map[string]string{
		"confeitaria em Manaíra, João Pessoa - PB": feedHTML(feedItem("id-C", "Confeitaria Fechada"), feedItem("id-A", "Doceria A"), feedItem("id-D", "Confeitaria Sem Celular")),
		"doceria em Manaíra, João Pessoa - PB":     feedHTML(feedItem("id-A", "Doceria A"), feedItem("id-B", "Doceria B"), feedItem("id-E", "Doceria E")),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/maps/search/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, feeds[strings.TrimPrefix(r.URL.Path, "/maps/search/")])
	})
	mux.HandleFunc("/maps/place/", func(w http.ResponseWriter, r *http.Request) {
		for id, html := range lugares {
			if strings.Contains(r.URL.Path, id) {
				aberturas[id]++
				fmt.Fprint(w, html)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})
	mapsSrv := httptest.NewServer(mux)
	defer mapsSrv.Close()

	var corposLLM [][]byte
	openaiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corpo, _ := io.ReadAll(r.Body)
		corposLLM = append(corposLLM, corpo)
		if len(corposLLM) == 1 {
			fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"demora pra responder no whatsapp","gancho":"vi que voces demoram bastante pra responder no whatsapp"}`, 100, 30))
			return
		}
		fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"","gancho":""}`, 50, 20))
	}))
	defer openaiSrv.Close()

	conta := &custo.Conta{}
	ctx, cancelNovo := maps.Novo(context.Background())
	defer cancelNovo()
	ctx, cancelTimeout := context.WithTimeout(ctx, 90*time.Second)
	defer cancelTimeout()

	mc := &maps.Coletor{Base: mapsSrv.URL, Pausa: 0}
	sc := &site.Checador{HTTP: siteProprio.Client(), PageSpeedBase: pageSpeed.URL, Chave: "", NotaMinima: 0.5}
	lc := &llm.Cliente{Chave: "chave-teste", Base: openaiSrv.URL, HTTP: openaiSrv.Client(), Conta: conta}
	r := rand.New(rand.NewPCG(1, 2))

	leads, descartes, cm, err := coletar(ctx, mc, sc, lc, r, entrada{
		Nicho:      "confeitaria",
		Bairro:     "Manaíra",
		Limite:     2,
		Remetente:  "Daniel",
		PreviewURL: "",
	})
	if err != nil {
		t.Fatalf("coletar: %v", err)
	}

	// mesmo lugar em dois termos sai 1 vez, e --limite 2 para de abrir depois de 2 qualificados.
	if aberturas["id-A"] != 1 {
		t.Errorf("aberturas[id-A] = %d, want 1 (dedup entre termos)", aberturas["id-A"])
	}
	if aberturas["id-E"] != 0 {
		t.Errorf("id-E foi aberto; limite=2 já tinha sido atingido em id-A e id-B")
	}
	if cm.Buscas != 2 {
		t.Errorf("cm.Buscas = %d, want 2", cm.Buscas)
	}
	if cm.LugaresAbertos != 4 {
		t.Errorf("cm.LugaresAbertos = %d, want 4 (C, A, D, B)", cm.LugaresAbertos)
	}

	// lugar fechado (id-C) ou sem celular (id-D) não gastam avaliação: só quem passou no porte (A, B) chama a LLM.
	if len(corposLLM) != 2 {
		t.Fatalf("chamadas à LLM = %d, want 2 (só id-A e id-B passaram no porte)", len(corposLLM))
	}

	// descartes: id-C (fechado) e id-D (sem celular), antes de custar qualquer avaliação.
	if len(descartes) != 2 {
		t.Fatalf("descartes = %+v, want 2", descartes)
	}
	motivos := map[string]string{}
	for _, d := range descartes {
		motivos[d.Nome] = d.Motivo
	}
	if motivos["Confeitaria Fechada"] != "fechado" {
		t.Errorf("motivo de Confeitaria Fechada = %q, want fechado", motivos["Confeitaria Fechada"])
	}
	if motivos["Confeitaria Sem Celular"] != "sem celular" {
		t.Errorf("motivo de Confeitaria Sem Celular = %q, want \"sem celular\"", motivos["Confeitaria Sem Celular"])
	}

	// leads qualificados: A e B, ordenados por pontuação (A tem mais avaliações e nota maior).
	if len(leads) != 2 {
		t.Fatalf("leads = %+v, want 2", leads)
	}
	if leads[0].PlaceID != "id-A" || leads[1].PlaceID != "id-B" {
		t.Errorf("ordem dos leads = [%s %s], want [id-A id-B] (pontuação desc)", leads[0].PlaceID, leads[1].PlaceID)
	}
	if leads[0].Pontuacao < leads[1].Pontuacao {
		t.Errorf("leads não ordenados por pontuação desc: %+v", leads)
	}

	// PageSpeed fora do ar não descarta o lead: id-A segue vivo (com a dor de reclamação vinda da LLM).
	leadA := leads[0]
	if len(leadA.Dores) != 1 || leadA.Dores[0] != lead.Reclamacao {
		t.Errorf("Dores de id-A = %v, want [reclamacao]", leadA.Dores)
	}
	if leadA.Telefone != "5583998765432" {
		t.Errorf("Telefone de id-A = %q, want 5583998765432", leadA.Telefone)
	}
	if leadA.Detalhes[lead.Reclamacao] != "demora pra responder no whatsapp" {
		t.Errorf("Detalhes[reclamacao] de id-A = %q", leadA.Detalhes[lead.Reclamacao])
	}
	if !strings.Contains(leadA.Mensagem, "vi que voces demoram bastante pra responder no whatsapp") {
		t.Errorf("Mensagem de id-A não usou o gancho da LLM: %q", leadA.Mensagem)
	}

	// resumo do Gemini + avaliação negativa do Maps chegam na LLM; a positiva fica de fora (llm.Entrada.Avaliacoes).
	corpoA := string(corposLLM[0])
	if !strings.Contains(corpoA, "Clientes elogiam o atendimento e reclamam da demora no whatsapp") {
		t.Errorf("corpo da 1ª chamada à LLM não contém o resumo do Gemini: %s", corpoA)
	}
	if !strings.Contains(corpoA, "demoraram pra responder no whatsapp") {
		t.Errorf("corpo da 1ª chamada à LLM não contém a avaliação negativa: %s", corpoA)
	}
	if strings.Contains(corpoA, "Muito bom, recomendo") {
		t.Errorf("corpo da 1ª chamada à LLM contém a avaliação positiva (só as negativas deviam ir): %s", corpoA)
	}

	// a LLM precisa saber que "sem_site" já foi detectado antes dela (doc de llm.Entrada.Dores).
	if !strings.Contains(string(corposLLM[1]), "sem_site") {
		t.Errorf("corpo da 2ª chamada à LLM não menciona sem_site: %s", corposLLM[1])
	}

	// gancho vazio da LLM (id-B) cai no GanchoPadrao da dor de maior peso (R2): SemSite.
	leadB := leads[1]
	if len(leadB.Dores) != 1 || leadB.Dores[0] != lead.SemSite {
		t.Errorf("Dores de id-B = %v, want [sem_site]", leadB.Dores)
	}
	if !strings.Contains(leadB.Mensagem, lead.GanchoPadrao(lead.SemSite, "")) {
		t.Errorf("Mensagem de id-B não usou o GanchoPadrao: %q", leadB.Mensagem)
	}

	// LANCER_PREVIEW_URL vazio: link cai no placeholder "<preview>" (R3).
	if leadB.Variante == lead.Link && !strings.Contains(leadB.Mensagem, "<preview>/p/"+leadB.Slug) {
		t.Errorf("Mensagem de id-B (variante link) sem o placeholder <preview>: %q", leadB.Mensagem)
	}

	var buf strings.Builder
	imprimir(&buf, leads, descartes, *conta, cm)
	saida := buf.String()
	for _, precisa := range []string{"Doceria A", "Doceria B", "fechado", "sem celular", "Maps: 2 buscas, 4 lugares abertos"} {
		if !strings.Contains(saida, precisa) {
			t.Errorf("saída impressa não contém %q:\n%s", precisa, saida)
		}
	}
}

// TestColetarCaptchaAborta cobre o Ruling novo: ErrCaptcha de qualquer chamada do Maps aborta a rodada inteira.
func TestColetarCaptchaAborta(t *testing.T) {
	if !temChrome() {
		t.Skip("Chrome/Chromium não encontrado, pulando teste com navegador real")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body><p>Nossos sistemas detectaram tráfego incomum vindo da sua rede.</p></body></html>`)
	}))
	defer srv.Close()

	ctx, cancelNovo := maps.Novo(context.Background())
	defer cancelNovo()
	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	mc := &maps.Coletor{Base: srv.URL, Pausa: 0}
	r := rand.New(rand.NewPCG(1, 2))

	_, _, _, err := coletar(ctx, mc, nil, nil, r, entrada{Nicho: "confeitaria", Bairro: "Manaíra", Limite: 2, Remetente: "Daniel"})
	if err == nil || err.Error() != mensagemCaptcha {
		t.Fatalf("coletar com captcha: err = %v, want %q", err, mensagemCaptcha)
	}
}

func TestDominioDe(t *testing.T) {
	casos := []struct{ uri, want string }{
		{"https://www.instagram.com/padoca", "instagram.com"},
		{"instagram.com/padoca", "instagram.com"},
		{"https://ifood.com.br/delivery/loja", "ifood.com.br"},
	}
	for _, c := range casos {
		if got := dominioDe(c.uri); got != c.want {
			t.Errorf("dominioDe(%q) = %q, want %q", c.uri, got, c.want)
		}
	}
}

func TestDorPrincipal(t *testing.T) {
	casos := []struct {
		dores []lead.Dor
		want  lead.Dor
	}{
		{[]lead.Dor{lead.Reclamacao, lead.SemSite}, lead.SemSite},
		{[]lead.Dor{lead.SiteRuim, lead.Reclamacao}, lead.SiteRuim},
		{[]lead.Dor{lead.Reclamacao}, lead.Reclamacao},
	}
	for _, c := range casos {
		if got := dorPrincipal(c.dores); got != c.want {
			t.Errorf("dorPrincipal(%v) = %q, want %q", c.dores, got, c.want)
		}
	}
}

func TestColetarNichoInvalido(t *testing.T) {
	_, _, _, err := coletar(context.Background(), nil, nil, nil, nil, entrada{Nicho: "salgadinho"})
	if err == nil {
		t.Fatal("esperava erro para nicho inválido")
	}
}
