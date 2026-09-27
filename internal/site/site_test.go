package site

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestChecarForaDoAr(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &Checador{HTTP: srv.Client(), NotaMinima: 0.5}
	ruim, motivo := c.Checar(context.Background(), srv.URL)
	if !ruim || motivo != "fora do ar" {
		t.Errorf("Checar = (%v, %q), want (true, \"fora do ar\")", ruim, motivo)
	}
}

func TestChecarLentoNoCelular(t *testing.T) {
	site := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer site.Close()
	ps := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"lighthouseResult":{"categories":{"performance":{"score":0.3}}}}`)
	}))
	defer ps.Close()

	c := &Checador{HTTP: site.Client(), PageSpeedBase: ps.URL, Chave: "chave-teste", NotaMinima: 0.5}
	ruim, motivo := c.Checar(context.Background(), site.URL)
	if !ruim || !strings.Contains(motivo, "lento") {
		t.Errorf("Checar = (%v, %q), want (true, contém \"lento\")", ruim, motivo)
	}
}

func TestChecarSiteOK(t *testing.T) {
	site := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer site.Close()
	ps := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"lighthouseResult":{"categories":{"performance":{"score":0.9}}}}`)
	}))
	defer ps.Close()

	c := &Checador{HTTP: site.Client(), PageSpeedBase: ps.URL, Chave: "chave-teste", NotaMinima: 0.5}
	ruim, motivo := c.Checar(context.Background(), site.URL)
	if ruim || motivo != "" {
		t.Errorf("Checar = (%v, %q), want (false, \"\")", ruim, motivo)
	}
}

func TestChecarSemHTTPS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Checador{HTTP: srv.Client(), NotaMinima: 0.5}
	ruim, motivo := c.Checar(context.Background(), srv.URL)
	if !ruim || motivo != "sem HTTPS" {
		t.Errorf("Checar = (%v, %q), want (true, \"sem HTTPS\")", ruim, motivo)
	}
}

func TestPageSpeedSemChaveNaoMandaKeyVazio(t *testing.T) {
	var query url.Values
	ps := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		fmt.Fprint(w, `{"lighthouseResult":{"categories":{"performance":{"score":0.9}}}}`)
	}))
	defer ps.Close()

	c := &Checador{HTTP: ps.Client(), PageSpeedBase: ps.URL, Chave: "", NotaMinima: 0.5}
	if _, ok := c.notaPageSpeed(context.Background(), "https://exemplo.com.br"); !ok {
		t.Fatal("notaPageSpeed devia ter sucesso")
	}
	if _, existe := query["key"]; existe {
		t.Errorf("query = %v, não devia ter \"key\" quando Chave é vazia", query)
	}
}

func TestChecarNotaAusenteNaoEhLento(t *testing.T) {
	site := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer site.Close()
	ps := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"lighthouseResult":{"categories":{"performance":{"score":null}}}}`)
	}))
	defer ps.Close()

	c := &Checador{HTTP: site.Client(), PageSpeedBase: ps.URL, Chave: "chave-teste", NotaMinima: 0.5}
	ruim, motivo := c.Checar(context.Background(), site.URL)
	if ruim || motivo != "" {
		t.Errorf("Checar = (%v, %q), want (false, \"\") — nota ausente é indeterminado, não pune o lead", ruim, motivo)
	}
}

func TestChecarBloqueioNaoEhForaDoAr(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests} {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))
		c := &Checador{HTTP: srv.Client(), NotaMinima: 0.5}
		ruim, motivo := c.Checar(context.Background(), srv.URL)
		if ruim || motivo != "" {
			t.Errorf("status %d: Checar = (%v, %q), want (false, \"\") — bloqueio é indeterminado, não fora do ar", status, ruim, motivo)
		}
		srv.Close()
	}
}

func TestChecarPageSpeedFalha(t *testing.T) {
	site := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer site.Close()
	ps := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ps.Close()

	c := &Checador{HTTP: site.Client(), PageSpeedBase: ps.URL, Chave: "chave-teste", NotaMinima: 0.5}
	ruim, motivo := c.Checar(context.Background(), site.URL)
	if ruim || motivo != "" {
		t.Errorf("Checar = (%v, %q), want (false, \"\") — falha nossa não pune o lead", ruim, motivo)
	}
}
