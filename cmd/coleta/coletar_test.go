package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dv-dev1/lancer/internal/custo"
	"github.com/dv-dev1/lancer/internal/lead"
	"github.com/dv-dev1/lancer/internal/llm"
	"github.com/dv-dev1/lancer/internal/places"
	"github.com/dv-dev1/lancer/internal/site"
)

// respostaOpenAI monta o corpo de chat/completions no formato que o cliente llm espera.
func respostaOpenAI(content string, tokensEntrada, tokensSaida int) string {
	b, _ := json.Marshal(content)
	return fmt.Sprintf(`{"choices":[{"message":{"content":%s}}],"usage":{"prompt_tokens":%d,"completion_tokens":%d}}`, b, tokensEntrada, tokensSaida)
}

func detalheJSON(id, siteProprioURL string) string {
	switch id {
	case "id-C": // fechado: descartado antes de custar avaliação
		return `{"id":"id-C","displayName":{"text":"Confeitaria Fechada"},"businessStatus":"CLOSED_PERMANENTLY","internationalPhoneNumber":"+55 83 99999-0001","rating":4.5,"userRatingCount":100}`
	case "id-A": // site próprio (PageSpeed dele cai) + reclamação vinda da LLM
		return fmt.Sprintf(`{"id":"id-A","displayName":{"text":"Doceria A"},"businessStatus":"OPERATIONAL","internationalPhoneNumber":"+55 83 99876-5432","websiteUri":%q,"rating":4.5,"userRatingCount":100}`, siteProprioURL)
	case "id-D": // sem celular: descartado antes de custar avaliação
		return `{"id":"id-D","displayName":{"text":"Confeitaria Sem Celular"},"businessStatus":"OPERATIONAL","internationalPhoneNumber":"","rating":4.5,"userRatingCount":100}`
	case "id-B": // sem site: gancho vem vazio da LLM, usa GanchoPadrao
		return `{"id":"id-B","displayName":{"text":"Doceria B"},"businessStatus":"OPERATIONAL","internationalPhoneNumber":"+55 83 98888-7777","websiteUri":"","rating":4.0,"userRatingCount":30}`
	}
	return `{}`
}

func TestColetarFluxoCompleto(t *testing.T) {
	// site próprio do id-A: no ar, HTTPS OK; o PageSpeed dele cai (testa "PageSpeed fora do ar não descarta o lead").
	siteProprio := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer siteProprio.Close()

	pageSpeed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer pageSpeed.Close()

	buscaPorTermo := map[string][]string{
		"confeitaria em Manaíra, João Pessoa - PB": {"id-C", "id-A", "id-D"},
		"doceria em Manaíra, João Pessoa - PB":     {"id-A", "id-B", "id-E"},
	}
	detalheChamadas := map[string]int{}
	avaliacaoChamadas := map[string]int{}

	placesSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var corpo struct {
				TextQuery string `json:"textQuery"`
			}
			json.NewDecoder(r.Body).Decode(&corpo)
			ids := buscaPorTermo[corpo.TextQuery]
			ps := make([]map[string]string, len(ids))
			for i, id := range ids {
				ps[i] = map[string]string{"id": id}
			}
			json.NewEncoder(w).Encode(map[string]any{"places": ps})
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/v1/places/")
		if r.Header.Get("X-Goog-FieldMask") == "reviews" {
			avaliacaoChamadas[id]++
			switch id {
			case "id-A":
				fmt.Fprint(w, `{"reviews":[{"text":{"text":"demoraram pra responder no whatsapp"}}]}`)
			default:
				fmt.Fprint(w, `{}`)
			}
			return
		}
		detalheChamadas[id]++
		fmt.Fprint(w, detalheJSON(id, siteProprio.URL))
	}))
	defer placesSrv.Close()

	var chamadasLLM int
	var corpoSegundaChamadaLLM []byte
	openaiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadasLLM++
		if chamadasLLM == 1 {
			fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"demora pra responder no whatsapp","gancho":"vi que voces demoram bastante pra responder no whatsapp"}`, 100, 30))
			return
		}
		corpoSegundaChamadaLLM, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"","gancho":""}`, 50, 20))
	}))
	defer openaiSrv.Close()

	conta := &custo.Conta{}
	pl := &places.Cliente{Chave: "chave-teste", Base: placesSrv.URL, HTTP: placesSrv.Client(), Conta: conta}
	sc := &site.Checador{HTTP: siteProprio.Client(), PageSpeedBase: pageSpeed.URL, Chave: "chave-teste", NotaMinima: 0.5}
	lc := &llm.Cliente{Chave: "chave-teste", Base: openaiSrv.URL, HTTP: openaiSrv.Client(), Conta: conta}
	r := rand.New(rand.NewPCG(1, 2))

	leads, descartes, err := coletar(context.Background(), pl, sc, lc, r, entrada{
		Nicho:      "confeitaria",
		Bairro:     "Manaíra",
		Limite:     2,
		Remetente:  "Daniel",
		PreviewURL: "",
	})
	if err != nil {
		t.Fatalf("coletar: %v", err)
	}

	// mesmo lugar em dois termos sai 1 vez, e --limite 2 para de detalhar depois de 2 qualificados.
	if detalheChamadas["id-A"] != 1 {
		t.Errorf("detalheChamadas[id-A] = %d, want 1 (dedup entre termos)", detalheChamadas["id-A"])
	}
	if _, tocou := detalheChamadas["id-E"]; tocou {
		t.Errorf("id-E foi detalhado; limite=2 já tinha sido atingido em id-A e id-B")
	}
	if len(detalheChamadas) != 4 {
		t.Errorf("detalheChamadas = %v, want 4 ids (C, A, D, B)", detalheChamadas)
	}

	// lugar fechado (id-C) ou sem celular (id-D) não gastam avaliação: só quem passou no porte (A, B) gasta.
	if len(avaliacaoChamadas) != 2 || avaliacaoChamadas["id-A"] != 1 || avaliacaoChamadas["id-B"] != 1 {
		t.Errorf("avaliacaoChamadas = %v, want só id-A e id-B, 1 vez cada", avaliacaoChamadas)
	}
	if conta.Atmosfera != 2 {
		t.Errorf("conta.Atmosfera = %d, want 2 (nº de quem passou no porte)", conta.Atmosfera)
	}

	// descartes: id-C (fechado) e id-D (sem celular), nenhuma avaliação gasta com eles.
	if len(descartes) != 2 {
		t.Fatalf("descartes = %+v, want 2", descartes)
	}
	motivos := map[string]string{descartes[0].Nome: descartes[0].Motivo, descartes[1].Nome: descartes[1].Motivo}
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
	if leadA.Detalhes[lead.Reclamacao] != "demora pra responder no whatsapp" {
		t.Errorf("Detalhes[reclamacao] de id-A = %q", leadA.Detalhes[lead.Reclamacao])
	}
	if !strings.Contains(leadA.Mensagem, "vi que voces demoram bastante pra responder no whatsapp") {
		t.Errorf("Mensagem de id-A não usou o gancho da LLM: %q", leadA.Mensagem)
	}

	// a LLM precisa saber que "sem_site" já foi detectado antes dela (doc de llm.Entrada.Dores).
	if !strings.Contains(string(corpoSegundaChamadaLLM), "sem_site") {
		t.Errorf("corpo da 2ª chamada à LLM não menciona sem_site: %s", corpoSegundaChamadaLLM)
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
	imprimir(&buf, leads, descartes, *conta)
	saida := buf.String()
	for _, precisa := range []string{"Doceria A", "Doceria B", "fechado", "sem celular", "teto"} {
		if !strings.Contains(saida, precisa) {
			t.Errorf("saída impressa não contém %q:\n%s", precisa, saida)
		}
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
	_, _, err := coletar(context.Background(), nil, nil, nil, nil, entrada{Nicho: "salgadinho"})
	if err == nil {
		t.Fatal("esperava erro para nicho inválido")
	}
}
