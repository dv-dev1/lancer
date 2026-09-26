// Package places fala com a Google Places API (New): busca IDs, detalha e lê avaliações.
package places

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dv-dev1/lancer/internal/custo"
)

type Cliente struct {
	Chave string
	Base  string
	HTTP  *http.Client
	Conta *custo.Conta
}

type Lugar struct {
	ID         string
	Nome       string
	Endereco   string
	Telefone   string
	Site       string
	Status     string
	MapsURI    string
	Nota       float64
	Avaliacoes int
}

// maxPaginas limita o custo de uma busca ruim; 3 páginas de 20 já cobrem o bairro na prática.
const maxPaginas = 3

type buscaCorpo struct {
	TextQuery    string `json:"textQuery"`
	LanguageCode string `json:"languageCode"`
	RegionCode   string `json:"regionCode"`
	PageSize     int    `json:"pageSize"`
	PageToken    string `json:"pageToken,omitempty"`
}

type buscaResposta struct {
	Places []struct {
		ID string `json:"id"`
	} `json:"places"`
	NextPageToken string `json:"nextPageToken"`
}

func (c *Cliente) BuscarIDs(ctx context.Context, consulta string) ([]string, error) {
	var ids []string
	var token string
	for pagina := 0; pagina < maxPaginas; pagina++ {
		corpo := buscaCorpo{
			TextQuery:    consulta,
			LanguageCode: "pt-BR",
			RegionCode:   "BR",
			PageSize:     20,
			PageToken:    token,
		}
		var resp buscaResposta
		if err := c.chamar(ctx, http.MethodPost, "/v1/places:searchText", "places.id,nextPageToken", corpo, &resp); err != nil {
			return nil, err
		}
		c.Conta.IDs += len(resp.Places)
		for _, p := range resp.Places {
			ids = append(ids, p.ID)
		}
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return ids, nil
}

type detalheResposta struct {
	ID          string `json:"id"`
	DisplayName struct {
		Text string `json:"text"`
	} `json:"displayName"`
	FormattedAddress         string  `json:"formattedAddress"`
	InternationalPhoneNumber string  `json:"internationalPhoneNumber"`
	WebsiteURI               string  `json:"websiteUri"`
	Rating                   float64 `json:"rating"`
	UserRatingCount          int     `json:"userRatingCount"`
	BusinessStatus           string  `json:"businessStatus"`
	GoogleMapsURI            string  `json:"googleMapsUri"`
}

const mascaraDetalhes = "id,displayName,formattedAddress,internationalPhoneNumber,websiteUri,rating,userRatingCount,businessStatus,googleMapsUri"

func (c *Cliente) Detalhar(ctx context.Context, id string) (Lugar, error) {
	var resp detalheResposta
	if err := c.chamar(ctx, http.MethodGet, caminhoLugar(id), mascaraDetalhes, nil, &resp); err != nil {
		return Lugar{}, err
	}
	c.Conta.Detalhes++
	return Lugar{
		ID:         resp.ID,
		Nome:       resp.DisplayName.Text,
		Endereco:   resp.FormattedAddress,
		Telefone:   resp.InternationalPhoneNumber,
		Site:       resp.WebsiteURI,
		Status:     resp.BusinessStatus,
		MapsURI:    resp.GoogleMapsURI,
		Nota:       resp.Rating,
		Avaliacoes: resp.UserRatingCount,
	}, nil
}

type avaliacoesResposta struct {
	Reviews []struct {
		Text struct {
			Text string `json:"text"`
		} `json:"text"`
	} `json:"reviews"`
}

func (c *Cliente) Avaliacoes(ctx context.Context, id string) ([]string, error) {
	var resp avaliacoesResposta
	if err := c.chamar(ctx, http.MethodGet, caminhoLugar(id), "reviews", nil, &resp); err != nil {
		return nil, err
	}
	c.Conta.Atmosfera++
	textos := make([]string, 0, len(resp.Reviews))
	for _, r := range resp.Reviews {
		if r.Text.Text != "" {
			textos = append(textos, r.Text.Text)
		}
	}
	return textos, nil
}

func caminhoLugar(id string) string {
	return "/v1/places/" + url.PathEscape(id) + "?languageCode=pt-BR"
}

func (c *Cliente) chamar(ctx context.Context, metodo, caminho, mascara string, corpo, resp any) error {
	var leitor io.Reader
	if corpo != nil {
		b, err := json.Marshal(corpo)
		if err != nil {
			return err
		}
		leitor = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, metodo, c.Base+caminho, leitor)
	if err != nil {
		return err
	}
	req.Header.Set("X-Goog-Api-Key", c.Chave)
	req.Header.Set("X-Goog-FieldMask", mascara)
	if corpo != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return erroHTTP(res)
	}
	return json.NewDecoder(res.Body).Decode(resp)
}

// erroHTTP trunca em 200 bytes: a resposta de erro do Google pode vir com corpo grande.
func erroHTTP(res *http.Response) error {
	corpo, _ := io.ReadAll(io.LimitReader(res.Body, 200))
	if res.StatusCode == http.StatusForbidden {
		return fmt.Errorf("places: status %d: %s (ative a Places API (New) no projeto e confira a chave)", res.StatusCode, corpo)
	}
	return fmt.Errorf("places: status %d: %s", res.StatusCode, corpo)
}
