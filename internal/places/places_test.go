package places

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dv-dev1/lancer/internal/custo"
)

func TestBuscarIDsDuasPaginas(t *testing.T) {
	var chamadas int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		if got := r.Header.Get("X-Goog-Api-Key"); got != "chave-teste" {
			t.Errorf("X-Goog-Api-Key = %q, want %q", got, "chave-teste")
		}
		if got := r.Header.Get("X-Goog-FieldMask"); got != "places.id,nextPageToken" {
			t.Errorf("X-Goog-FieldMask = %q, want %q", got, "places.id,nextPageToken")
		}
		var corpo struct {
			TextQuery    string `json:"textQuery"`
			LanguageCode string `json:"languageCode"`
			RegionCode   string `json:"regionCode"`
			PageSize     int    `json:"pageSize"`
			PageToken    string `json:"pageToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&corpo); err != nil {
			t.Fatalf("decode corpo: %v", err)
		}
		if corpo.TextQuery != "padaria em Manaíra" || corpo.LanguageCode != "pt-BR" || corpo.RegionCode != "BR" || corpo.PageSize != 20 {
			t.Errorf("corpo inesperado: %+v", corpo)
		}
		if chamadas == 1 && corpo.PageToken != "" {
			t.Errorf("pageToken na 1ª chamada = %q, want vazio", corpo.PageToken)
		}
		if chamadas == 2 && corpo.PageToken != "pagina-2" {
			t.Errorf("pageToken na 2ª chamada = %q, want %q", corpo.PageToken, "pagina-2")
		}

		ids := make([]map[string]string, 20)
		for i := range ids {
			ids[i] = map[string]string{"id": fmt.Sprintf("id-%d-%d", chamadas, i)}
		}
		resp := map[string]any{"places": ids}
		if chamadas == 1 {
			resp["nextPageToken"] = "pagina-2"
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
	ids, err := c.BuscarIDs(context.Background(), "padaria em Manaíra")
	if err != nil {
		t.Fatalf("BuscarIDs: %v", err)
	}
	if len(ids) != 40 {
		t.Errorf("len(ids) = %d, want 40", len(ids))
	}
	if chamadas != 2 {
		t.Errorf("chamadas = %d, want 2", chamadas)
	}
	if conta.IDs != 40 {
		t.Errorf("conta.IDs = %d, want 40", conta.IDs)
	}
}

func TestDetalhar(t *testing.T) {
	const mascaraQuerida = "id,displayName,formattedAddress,internationalPhoneNumber,websiteUri,rating,userRatingCount,businessStatus,googleMapsUri"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("languageCode"); got != "pt-BR" {
			t.Errorf("languageCode = %q, want pt-BR", got)
		}
		if got := r.Header.Get("X-Goog-FieldMask"); got != mascaraQuerida {
			t.Errorf("X-Goog-FieldMask = %q, want %q", got, mascaraQuerida)
		}
		fmt.Fprint(w, `{"id":"abc123","displayName":{"text":"Padaria Boa Vista"},"formattedAddress":"Rua X, 100"}`)
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
	lugar, err := c.Detalhar(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Detalhar: %v", err)
	}
	if lugar.Nome != "Padaria Boa Vista" {
		t.Errorf("Nome = %q, want %q", lugar.Nome, "Padaria Boa Vista")
	}
	if lugar.Endereco != "Rua X, 100" {
		t.Errorf("Endereco = %q, want %q", lugar.Endereco, "Rua X, 100")
	}
	if lugar.Telefone != "" || lugar.Site != "" || lugar.Nota != 0 || lugar.Avaliacoes != 0 || lugar.Status != "" || lugar.MapsURI != "" {
		t.Errorf("campos ausentes deviam ser zero: %+v", lugar)
	}
	if conta.Detalhes != 1 {
		t.Errorf("conta.Detalhes = %d, want 1", conta.Detalhes)
	}
}

func TestAvaliacoes(t *testing.T) {
	casos := []struct {
		nome   string
		corpo  string
		querem []string
	}{
		{"vazias", `{}`, []string{}},
		{"filtra texto vazio", `{"reviews":[{"text":{"text":"ótimo atendimento"}},{"text":{"text":""}}]}`, []string{"ótimo atendimento"}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("X-Goog-FieldMask"); got != "reviews" {
					t.Errorf("X-Goog-FieldMask = %q, want reviews", got)
				}
				fmt.Fprint(w, c.corpo)
			}))
			defer srv.Close()

			conta := &custo.Conta{}
			cli := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
			textos, err := cli.Avaliacoes(context.Background(), "abc123")
			if err != nil {
				t.Fatalf("Avaliacoes: %v", err)
			}
			if len(textos) != len(c.querem) {
				t.Fatalf("textos = %v, want %v", textos, c.querem)
			}
			for i, txt := range textos {
				if txt != c.querem[i] {
					t.Errorf("textos[%d] = %q, want %q", i, txt, c.querem[i])
				}
			}
			if conta.Atmosfera != 1 {
				t.Errorf("conta.Atmosfera = %d, want 1", conta.Atmosfera)
			}
		})
	}
}

func TestErro403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error":{"status":"PERMISSION_DENIED","message":"chave sem permissão"}}`)
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
	_, err := c.BuscarIDs(context.Background(), "padaria em Manaíra")
	if err == nil {
		t.Fatal("esperava erro no 403")
	}
	if !strings.Contains(err.Error(), "PERMISSION_DENIED") {
		t.Errorf("erro sem PERMISSION_DENIED: %v", err)
	}
	if !strings.Contains(err.Error(), "ative a Places API (New) no projeto e confira a chave") {
		t.Errorf("erro sem a dica do 403: %v", err)
	}
}
