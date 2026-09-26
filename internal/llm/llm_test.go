package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dv-dev1/lancer/internal/custo"
	"github.com/dv-dev1/lancer/internal/lead"
)

type corpoRequisicao struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	ResponseFormat struct {
		Type       string `json:"type"`
		JSONSchema struct {
			Name   string `json:"name"`
			Strict bool   `json:"strict"`
			Schema struct {
				Type       string `json:"type"`
				Properties map[string]struct {
					Type string `json:"type"`
				} `json:"properties"`
				Required             []string `json:"required"`
				AdditionalProperties bool     `json:"additionalProperties"`
			} `json:"schema"`
		} `json:"json_schema"`
	} `json:"response_format"`
}

func respostaOpenAI(content string, promptTokens, completionTokens int) string {
	b, _ := json.Marshal(content)
	return fmt.Sprintf(`{"choices":[{"message":{"content":%s}}],"usage":{"prompt_tokens":%d,"completion_tokens":%d}}`, b, promptTokens, completionTokens)
}

func TestAnalisarRequisicaoESchema(t *testing.T) {
	var recebido corpoRequisicao
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer chave-teste" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer chave-teste")
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&recebido); err != nil {
			t.Fatalf("decode corpo: %v", err)
		}
		fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"demora no whatsapp","gancho":"vi que vocês demoram bastante pra responder no whatsapp"}`, 120, 40))
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
	saida, err := c.Analisar(context.Background(), Entrada{
		Nicho:      "restaurante",
		Dores:      map[lead.Dor]string{lead.SemSite: ""},
		Avaliacoes: []string{"demoraram muito pra responder no whatsapp"},
	})
	if err != nil {
		t.Fatalf("Analisar: %v", err)
	}

	if recebido.Model != "gpt-4.1-mini" {
		t.Errorf("Model = %q, want %q (padrão quando Modelo vazio)", recebido.Model, "gpt-4.1-mini")
	}
	js := recebido.ResponseFormat.JSONSchema
	if recebido.ResponseFormat.Type != "json_schema" || js.Name != "analise" || !js.Strict {
		t.Errorf("response_format inesperado: %+v", recebido.ResponseFormat)
	}
	if js.Schema.Type != "object" {
		t.Errorf("schema.type = %q, want object", js.Schema.Type)
	}
	if js.Schema.Properties["reclamacao"].Type != "string" || js.Schema.Properties["gancho"].Type != "string" {
		t.Errorf("schema.properties inesperado: %+v", js.Schema.Properties)
	}
	if len(js.Schema.Required) != 2 || js.Schema.Required[0] != "reclamacao" || js.Schema.Required[1] != "gancho" {
		t.Errorf("schema.required = %v, want [reclamacao gancho]", js.Schema.Required)
	}
	if js.Schema.AdditionalProperties {
		t.Errorf("schema.additionalProperties = true, want false")
	}

	if saida.Reclamacao != "demora no whatsapp" {
		t.Errorf("Reclamacao = %q, want %q", saida.Reclamacao, "demora no whatsapp")
	}
	if saida.Gancho != "vi que vocês demoram bastante pra responder no whatsapp" {
		t.Errorf("Gancho = %q", saida.Gancho)
	}
	if conta.TokensEntrada != 120 || conta.TokensSaida != 40 {
		t.Errorf("Conta = %+v, want TokensEntrada=120 TokensSaida=40", conta)
	}
}

func TestPromptSistemaRegras(t *testing.T) {
	var recebido corpoRequisicao
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&recebido)
		fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"","gancho":"vi que ainda não tem site"}`, 1, 1))
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
	if _, err := c.Analisar(context.Background(), Entrada{Nicho: "loja"}); err != nil {
		t.Fatalf("Analisar: %v", err)
	}

	if len(recebido.Messages) == 0 || recebido.Messages[0].Role != "system" {
		t.Fatalf("Messages = %+v, want [0] com role system", recebido.Messages)
	}
	prompt := recebido.Messages[0].Content

	proibidas := []string{"agregador", "presença digital", "captação", "visibilidade"}
	for _, p := range proibidas {
		if !strings.Contains(prompt, p) {
			t.Errorf("prompt não cita a palavra proibida %q pra instruir a LLM a nunca usá-la", p)
		}
	}
	if !strings.Contains(prompt, "nunca") {
		t.Error("prompt devia deixar claro que essas palavras nunca podem aparecer")
	}
	if !strings.Contains(prompt, "primeira pessoa do singular") {
		t.Error("prompt devia pedir primeira pessoa do singular (nunca \"percebemos\")")
	}
	if !strings.Contains(prompt, "nome comum da plataforma") {
		t.Error("prompt devia pedir o nome comum da plataforma, nunca o host")
	}
	if !strings.Contains(prompt, "Instagram") || !strings.Contains(prompt, "iFood") {
		t.Error("prompt devia dar exemplos de nome comum de plataforma (Instagram, iFood, ...)")
	}
	if !strings.Contains(prompt, "consequência concreta") {
		t.Error("prompt devia pedir a consequência concreta pra quem compra, não marketing genérico")
	}
	if !strings.Contains(prompt, "1 frase curta") {
		t.Error("prompt devia limitar o gancho a 1 frase curta")
	}
}

func TestAnalisarModeloCustomizado(t *testing.T) {
	var recebido corpoRequisicao
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&recebido)
		fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"","gancho":"vi que ainda não tem site"}`, 10, 5))
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, Modelo: "gpt-4.1", HTTP: srv.Client(), Conta: conta}
	if _, err := c.Analisar(context.Background(), Entrada{Nicho: "loja"}); err != nil {
		t.Fatalf("Analisar: %v", err)
	}
	if recebido.Model != "gpt-4.1" {
		t.Errorf("Model = %q, want %q", recebido.Model, "gpt-4.1")
	}
}

func TestAnalisarRetentativaComRS(t *testing.T) {
	var chamadas int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		if chamadas == 1 {
			fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"","gancho":"um combo por R$ 20 pra vocês"}`, 50, 20))
			return
		}
		fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"","gancho":"vi que ainda não tem site"}`, 60, 25))
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
	saida, err := c.Analisar(context.Background(), Entrada{Nicho: "loja"})
	if err != nil {
		t.Fatalf("Analisar: %v", err)
	}
	if chamadas != 2 {
		t.Fatalf("chamadas = %d, want 2 (1ª com R$ força retentativa)", chamadas)
	}
	if saida.Gancho != "vi que ainda não tem site" {
		t.Errorf("Gancho = %q, want o da 2ª chamada", saida.Gancho)
	}
	if conta.TokensEntrada != 110 || conta.TokensSaida != 45 {
		t.Errorf("Conta = %+v, want soma das 2 chamadas (110, 45)", conta)
	}
}

func TestAnalisarDuasInvalidasDevolveVazio(t *testing.T) {
	var chamadas int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		fmt.Fprint(w, respostaOpenAI(`{"reclamacao":"","gancho":"promoção de R$ 10"}`, 10, 10))
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
	saida, err := c.Analisar(context.Background(), Entrada{Nicho: "loja"})
	if err != nil {
		t.Fatalf("Analisar: %v", err)
	}
	if chamadas != 2 {
		t.Errorf("chamadas = %d, want 2 (não tenta uma 3ª vez)", chamadas)
	}
	if saida.Gancho != "" {
		t.Errorf("Gancho = %q, want vazio após 2 tentativas inválidas", saida.Gancho)
	}
}

func TestAnalisarErroHTTP(t *testing.T) {
	var chamadas int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, strings.Repeat("x", 300))
	}))
	defer srv.Close()

	conta := &custo.Conta{}
	c := &Cliente{Chave: "chave-teste", Base: srv.URL, HTTP: srv.Client(), Conta: conta}
	_, err := c.Analisar(context.Background(), Entrada{Nicho: "loja"})
	if err == nil {
		t.Fatal("esperava erro no 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("erro sem o status: %v", err)
	}
	if strings.Count(err.Error(), "x") > 200 {
		t.Errorf("erro devia truncar o corpo em 200 bytes: %v", err)
	}
	if chamadas != 1 {
		t.Errorf("chamadas = %d, want 1 (erro HTTP não tenta de novo)", chamadas)
	}
}
