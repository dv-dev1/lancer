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

// paginaLugar monta um fixture de página de lugar compatível com os seletores de internal/maps.
// comAba=false modela a visão signed-out reduzida (sem aba de Avaliações, achada ao vivo no round 3
// do internal/maps): é o caminho do R12, então nesse caso nem histograma, nem resumo, nem
// avaliações existem no DOM — só a "Visão geral"/"Sobre".
func paginaLugar(nome, notaTexto, telefoneDigitos, siteHref string, fechado, comAba bool, estrelas [5]int, resumo string, reviews []reviewFix) string {
	var b strings.Builder
	b.WriteString("<html><body>\n  <h1>" + nome + "</h1>\n")
	if fechado {
		b.WriteString("  <p>Fechado permanentemente</p>\n")
	}
	b.WriteString(`  <span role="img" aria-label="` + notaTexto + ` estrelas">` + notaTexto + "</span>\n")
	if telefoneDigitos != "" {
		b.WriteString(`  <button data-item-id="phone:tel:0` + telefoneDigitos + `">telefone</button>` + "\n")
	}
	if siteHref != "" {
		b.WriteString(`  <a data-item-id="authority" href="` + siteHref + `">site</a>` + "\n")
	}
	b.WriteString(`  <button data-item-id="address" aria-label="Endereço: Rua Teste, 1 - Bairro">Rua Teste, 1</button>` + "\n")
	if !comAba {
		b.WriteString(`  <button role="tab" aria-label="Sobre">Sobre</button>` + "\n</body></html>")
		return b.String()
	}
	b.WriteString(`  <button role="tab" aria-label="Avaliações">Avaliações</button>` + "\n  <div>\n")
	for i, estrelaN := range []int{5, 4, 3, 2, 1} {
		b.WriteString(fmt.Sprintf(`    <span aria-label="%d estrelas, %d avaliações"></span>`+"\n", estrelaN, estrelas[i]))
	}
	b.WriteString("  </div>\n")
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

// feedItemAbsoluto aponta pra fora do servidor da busca: simula o Abrir de um lugar isolado
// falhando por conexão recusada, sem precisar encolher o timeout de internal/maps (não dá pra
// tocar nesse pacote a partir daqui, e o timeout de produção é de 15s).
func feedItemAbsoluto(base, id, nome string) string {
	return fmt.Sprintf(`<a href="%s/maps/place/x/data=!19s%s" aria-label="%s">%s</a>`, base, id, nome, nome)
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

	// porta morta: sobe e fecha na hora, só pra ter um endereço que recusa conexão (id-F).
	portaMorta := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	urlMorta := portaMorta.URL
	portaMorta.Close()

	aberturas := map[string]int{}
	lugares := map[string]string{
		"id-C": paginaLugar("Confeitaria Fechada", "4,5", "83999990001", "", true, true, [5]int{60, 20, 10, 7, 3}, "", nil),
		"id-D": paginaLugar("Confeitaria Sem Celular", "4,5", "", "", false, true, [5]int{60, 20, 10, 7, 3}, "", nil),
		"id-A": paginaLugar("Doceria A", "4,5", "83998765432", siteProprio.URL, false, true, [5]int{70, 15, 8, 4, 3},
			"Clientes elogiam o atendimento e reclamam da demora no whatsapp.",
			[]reviewFix{{"r1", "5", "Muito bom, recomendo!"}, {"r2", "2", "demoraram pra responder no whatsapp"}}),
		// sem aba de Avaliações: visão signed-out reduzida (R12) — Avaliacoes vira -1, porte usa só a nota.
		"id-G": paginaLugar("Doceria G", "4,2", "83977776666", "", false, false, [5]int{}, "", nil),
		"id-B": paginaLugar("Doceria B", "4,0", "83988887777", "", false, true, [5]int{20, 5, 2, 2, 1}, "", nil),
		"id-E": paginaLugar("Doceria E", "4,5", "83900001111", "", false, true, [5]int{60, 20, 10, 7, 3}, "", nil),
	}
	// id-A duplicado nos dois termos testa o dedup; id-F (porta morta) testa erro isolado que não
	// derruba a coleta; id-E só existe pra provar que --limite 3 para antes de abri-lo.
	feeds := map[string]string{
		"confeitaria em Manaíra, João Pessoa - PB": feedHTML(
			feedItem("id-C", "Confeitaria Fechada"),
			feedItemAbsoluto(urlMorta, "id-F", "Doceria F"),
			feedItem("id-A", "Doceria A"),
			feedItem("id-D", "Confeitaria Sem Celular"),
		),
		"doceria em Manaíra, João Pessoa - PB": feedHTML(
			feedItem("id-A", "Doceria A"),
			feedItem("id-G", "Doceria G"),
			feedItem("id-B", "Doceria B"),
			feedItem("id-E", "Doceria E"),
		),
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
	// mais folgado que antes: id-G (sem aba) come os 15s de timeout de produção do internal/maps.
	ctx, cancelTimeout := context.WithTimeout(ctx, 150*time.Second)
	defer cancelTimeout()

	mc := &maps.Coletor{Base: mapsSrv.URL, Pausa: -1} // negativo desliga a pausa; zero cairia no padrão de produção (3-4s)
	sc := &site.Checador{HTTP: siteProprio.Client(), PageSpeedBase: pageSpeed.URL, Chave: "", NotaMinima: 0.5}
	lc := &llm.Cliente{Chave: "chave-teste", Base: openaiSrv.URL, HTTP: openaiSrv.Client(), Conta: conta}
	r := rand.New(rand.NewPCG(1, 2))

	leads, descartes, cm, err := coletar(ctx, mc, sc, lc, r, entrada{
		Nicho:      "confeitaria",
		Bairro:     "Manaíra",
		Limite:     3,
		Remetente:  "Daniel",
		PreviewURL: "",
	})
	if err != nil {
		t.Fatalf("coletar: %v", err)
	}

	// mesmo lugar em dois termos sai 1 vez, e --limite 3 para de abrir depois de 3 qualificados (A, G, B).
	if aberturas["id-A"] != 1 {
		t.Errorf("aberturas[id-A] = %d, want 1 (dedup entre termos)", aberturas["id-A"])
	}
	if aberturas["id-E"] != 0 {
		t.Errorf("id-E foi aberto; limite=3 já tinha sido atingido em id-A, id-G e id-B")
	}
	if cm.Buscas != 2 {
		t.Errorf("cm.Buscas = %d, want 2", cm.Buscas)
	}
	if cm.LugaresAbertos != 6 {
		t.Errorf("cm.LugaresAbertos = %d, want 6 (C, F, A, D, G, B — E não conta)", cm.LugaresAbertos)
	}

	// só quem passou no porte (A, G, B) chama a LLM — fechado/sem celular/erro isolado não gastam avaliação.
	if len(corposLLM) != 3 {
		t.Fatalf("chamadas à LLM = %d, want 3 (id-A, id-G e id-B passaram no porte)", len(corposLLM))
	}

	// descartes: id-C (fechado), id-F (erro isolado ao abrir — não é captcha, não derruba a coleta) e id-D (sem celular).
	if len(descartes) != 3 {
		t.Fatalf("descartes = %+v, want 3", descartes)
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
	if !strings.HasPrefix(motivos["Doceria F"], "erro:") {
		t.Errorf("motivo de Doceria F = %q, want prefixo \"erro:\" (Abrir falhou, não é captcha, a coleta segue)", motivos["Doceria F"])
	}

	if len(leads) != 3 {
		t.Fatalf("leads = %+v, want 3", leads)
	}
	porID := map[string]lead.Lead{}
	for _, l := range leads {
		porID[l.PlaceID] = l
	}
	for i := 1; i < len(leads); i++ {
		if leads[i-1].Pontuacao < leads[i].Pontuacao {
			t.Errorf("leads não ordenados por pontuação desc: %+v", leads)
		}
	}

	// PageSpeed fora do ar não descarta o lead: id-A segue vivo (com a dor de reclamação vinda da LLM).
	leadA, ok := porID["id-A"]
	if !ok {
		t.Fatal("id-A não qualificou")
	}
	if len(leadA.Dores) != 1 || leadA.Dores[0] != lead.Reclamacao {
		t.Errorf("Dores de id-A = %v, want [reclamacao]", leadA.Dores)
	}
	if leadA.Telefone != "5583998765432" {
		t.Errorf("Telefone de id-A = %q, want 5583998765432", leadA.Telefone)
	}
	if leadA.Avaliacoes != 100 {
		t.Errorf("Avaliacoes de id-A = %d, want 100 (soma do histograma da aba)", leadA.Avaliacoes)
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

	// R12: id-G não tem aba de Avaliações → Avaliacoes -1, mas qualifica pelo porte só na nota (4,2 >= 4,0).
	leadG, ok := porID["id-G"]
	if !ok {
		t.Fatal("id-G não qualificou (R12: devia passar no porte só pela nota, sem contagem de avaliações)")
	}
	if leadG.Avaliacoes != -1 {
		t.Errorf("Avaliacoes de id-G = %d, want -1 (desconhecida, sem aba)", leadG.Avaliacoes)
	}
	if len(leadG.Dores) != 1 || leadG.Dores[0] != lead.SemSite {
		t.Errorf("Dores de id-G = %v, want [sem_site]", leadG.Dores)
	}
	// a LLM precisa saber que "sem_site" já foi detectado antes dela (doc de llm.Entrada.Dores).
	if !strings.Contains(string(corposLLM[1]), "sem_site") {
		t.Errorf("corpo da 2ª chamada à LLM (id-G) não menciona sem_site: %s", corposLLM[1])
	}

	// gancho vazio da LLM (id-B) cai no GanchoPadrao da dor de maior peso (R2): SemSite.
	leadB, ok := porID["id-B"]
	if !ok {
		t.Fatal("id-B não qualificou")
	}
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
	for _, precisa := range []string{"Doceria A", "Doceria B", "Doceria G", "fechado", "sem celular", "(?)", "Maps: 2 buscas, 6 lugares abertos"} {
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

	mc := &maps.Coletor{Base: srv.URL, Pausa: -1}
	r := rand.New(rand.NewPCG(1, 2))

	_, _, _, err := coletar(ctx, mc, nil, nil, r, entrada{Nicho: "confeitaria", Bairro: "Manaíra", Limite: 2, Remetente: "Daniel"})
	if err == nil || err.Error() != mensagemCaptcha {
		t.Fatalf("coletar com captcha: err = %v, want %q", err, mensagemCaptcha)
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
