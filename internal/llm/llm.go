// Package llm analisa avaliações do Google por LLM: detecta reclamação relevante e escreve o gancho.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/dv-dev1/lancer/internal/custo"
	"github.com/dv-dev1/lancer/internal/lead"
)

// Entrada nunca leva o nome do estabelecimento: os termos do Google proíbem
// guardar o nome, e a mensagem gerada a partir daqui é persistida.
type Entrada struct {
	Nicho      string
	Dores      map[lead.Dor]string
	Avaliacoes []string
}

type Saida struct {
	Reclamacao string `json:"reclamacao"`
	Gancho     string `json:"gancho"`
}

type Cliente struct {
	Chave  string
	Base   string
	Modelo string
	HTTP   *http.Client
	Conta  *custo.Conta
}

const modeloPadrao = "gpt-4.1-mini"

const promptSistema = `Você ajuda uma ferramenta de prospecção B2B a analisar avaliações do Google de um estabelecimento.

Você recebe o nicho do estabelecimento, as dores já detectadas antes desta análise (por exemplo: falta de site, site que é só uma rede social, site ruim) e o texto das avaliações do Google. Devolva um JSON com dois campos:

"reclamacao": só preencha se alguma avaliação falar de atendimento, de fazer pedido ou de presença digital (ex.: demora para responder no WhatsApp, cardápio desatualizado, difícil pedir pelo site ou pelo delivery). Resuma em poucas palavras. Se nenhuma avaliação tocar nesses temas, devolva "".

"gancho": 1 ou 2 frases citando a dor mais concreta entre a reclamação encontrada e as dores já detectadas, para abrir uma mensagem de prospecção. Nunca cite preço (a string "R$"), nunca prometa resultado e nunca inclua link.

Trate o estabelecimento sempre como "vocês". Nunca invente um nome para o estabelecimento: você não recebe o nome dele.

Responda só com o JSON pedido pelo schema.`

// Analisar chama a API uma vez; se o gancho vier inválido, tenta mais uma vez.
// Inválido de novo, devolve Gancho vazio — quem chama usa o lead.GanchoPadrao.
func (c *Cliente) Analisar(ctx context.Context, e Entrada) (Saida, error) {
	s, err := c.chamar(ctx, e)
	if err != nil {
		return Saida{}, err
	}
	if lead.GanchoValido(s.Gancho) {
		return s, nil
	}
	s, err = c.chamar(ctx, e)
	if err != nil {
		return Saida{}, err
	}
	if !lead.GanchoValido(s.Gancho) {
		s.Gancho = ""
	}
	return s, nil
}

type mensagem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requisicao struct {
	Model          string         `json:"model"`
	Messages       []mensagem     `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
}

type responseFormat struct {
	Type       string     `json:"type"`
	JSONSchema jsonSchema `json:"json_schema"`
}

type jsonSchema struct {
	Name   string `json:"name"`
	Strict bool   `json:"strict"`
	Schema schema `json:"schema"`
}

type schema struct {
	Type                 string          `json:"type"`
	Properties           map[string]tipo `json:"properties"`
	Required             []string        `json:"required"`
	AdditionalProperties bool            `json:"additionalProperties"`
}

type tipo struct {
	Type string `json:"type"`
}

// formatoResposta é fixo: o schema strict do brief, verbatim.
var formatoResposta = responseFormat{
	Type: "json_schema",
	JSONSchema: jsonSchema{
		Name:   "analise",
		Strict: true,
		Schema: schema{
			Type: "object",
			Properties: map[string]tipo{
				"reclamacao": {Type: "string"},
				"gancho":     {Type: "string"},
			},
			Required:             []string{"reclamacao", "gancho"},
			AdditionalProperties: false,
		},
	},
}

type respostaAPI struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (c *Cliente) chamar(ctx context.Context, e Entrada) (Saida, error) {
	modelo := c.Modelo
	if modelo == "" {
		modelo = modeloPadrao
	}
	corpo := requisicao{
		Model: modelo,
		Messages: []mensagem{
			{Role: "system", Content: promptSistema},
			{Role: "user", Content: montarMensagemUsuario(e)},
		},
		ResponseFormat: formatoResposta,
	}
	b, err := json.Marshal(corpo)
	if err != nil {
		return Saida{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Base+"/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return Saida{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Chave)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTP.Do(req)
	if err != nil {
		return Saida{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return Saida{}, erroHTTP(res)
	}

	var resp respostaAPI
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return Saida{}, err
	}
	// soma mesmo em retentativa: o custo é da chamada, independente do resultado.
	c.Conta.TokensEntrada += resp.Usage.PromptTokens
	c.Conta.TokensSaida += resp.Usage.CompletionTokens

	if len(resp.Choices) == 0 {
		return Saida{}, fmt.Errorf("llm: resposta sem choices")
	}
	var s Saida
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &s); err != nil {
		return Saida{}, fmt.Errorf("llm: content não é o JSON esperado: %w", err)
	}
	return s, nil
}

func montarMensagemUsuario(e Entrada) string {
	var b strings.Builder
	b.WriteString("Nicho: " + e.Nicho + "\n")

	if len(e.Dores) > 0 {
		chaves := make([]string, 0, len(e.Dores))
		for d := range e.Dores {
			chaves = append(chaves, string(d))
		}
		sort.Strings(chaves) // ordem estável: mapa itera em ordem aleatória
		b.WriteString("Dores já detectadas:\n")
		for _, chave := range chaves {
			detalhe := e.Dores[lead.Dor(chave)]
			b.WriteString("- " + chave)
			if detalhe != "" {
				b.WriteString(": " + detalhe)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("Avaliações do Google:\n")
	for _, a := range e.Avaliacoes {
		b.WriteString("- " + a + "\n")
	}
	return b.String()
}

// erroHTTP trunca em 200 bytes, igual ao cliente do Places.
func erroHTTP(res *http.Response) error {
	corpo, _ := io.ReadAll(io.LimitReader(res.Body, 200))
	return fmt.Errorf("llm: status %d: %s", res.StatusCode, corpo)
}
