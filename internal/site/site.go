// Package site decide se o site do lead está ruim: fora do ar, sem HTTPS ou lento no mobile.
package site

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Checador struct {
	HTTP          *http.Client
	PageSpeedBase string
	Chave         string
	NotaMinima    float64
}

const (
	timeoutSite      = 10 * time.Second
	timeoutPageSpeed = 60 * time.Second
)

// Checar nunca devolve erro: a falha do próprio site é o achado, não um problema nosso.
func (c *Checador) Checar(ctx context.Context, alvo string) (ruim bool, motivo string) {
	if !strings.Contains(alvo, "://") {
		alvo = "https://" + alvo
	}

	ctxSite, cancel := context.WithTimeout(ctx, timeoutSite)
	defer cancel()
	req, err := http.NewRequestWithContext(ctxSite, http.MethodGet, alvo, nil)
	if err != nil {
		return true, "fora do ar"
	}
	res, err := c.HTTP.Do(req)
	if err != nil || res.StatusCode >= 400 {
		return true, "fora do ar"
	}
	defer res.Body.Close()

	// checa o esquema após seguir redirecionamentos: o real é o que o navegador do lead veria.
	if res.Request.URL.Scheme != "https" {
		return true, "sem HTTPS"
	}

	nota, ok := c.notaPageSpeed(ctx, res.Request.URL.String())
	if !ok {
		return false, ""
	}
	if nota < c.NotaMinima {
		return true, fmt.Sprintf("lento no celular (nota %d/100)", int(nota*100))
	}
	return false, ""
}

func (c *Checador) notaPageSpeed(ctx context.Context, alvo string) (nota float64, ok bool) {
	ctxPS, cancel := context.WithTimeout(ctx, timeoutPageSpeed)
	defer cancel()

	consulta := url.Values{
		"url":      {alvo},
		"strategy": {"mobile"},
		"category": {"performance"},
		"key":      {c.Chave},
	}
	destino := c.PageSpeedBase + "/pagespeedonline/v5/runPagespeed?" + consulta.Encode()
	req, err := http.NewRequestWithContext(ctxPS, http.MethodGet, destino, nil)
	if err != nil {
		return 0, false
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return 0, false
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return 0, false
	}

	var resp struct {
		LighthouseResult struct {
			Categories struct {
				Performance struct {
					Score float64 `json:"score"`
				} `json:"performance"`
			} `json:"categories"`
		} `json:"lighthouseResult"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return 0, false
	}
	return resp.LighthouseResult.Categories.Performance.Score, true
}
