package maps

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"time"

	"github.com/chromedp/chromedp"
)

var ErrCaptcha = errors.New("maps: captcha")

// BaseGoogleMaps é o Base de produção; os testes apontam Coletor.Base para um httptest.
const BaseGoogleMaps = "https://www.google.com"

// PausaPadrao é o valor recomendado pra Coletor.Pausa fora de teste.
const PausaPadrao = 3 * time.Second

// esperaAposClique e esperaAposRolagem são vars (não const) pra os testes encolherem e o
// fixture não levar minutos pra rodar.
var (
	esperaAposClique  = 2 * time.Second
	esperaAposRolagem = 1500 * time.Millisecond
)

type Resultado struct {
	ID, Nome, URL string
}

type Lugar struct {
	ID, Nome, Endereco, Telefone, Site string
	Fechado                            bool
	Nota                               float64
	Avaliacoes                         int
}

type Coletor struct {
	Base  string
	Pausa time.Duration
}

func (c *Coletor) base() string {
	if c.Base == "" {
		return BaseGoogleMaps
	}
	return c.Base
}

// pausar soma até 1s de sorteio à pausa configurada; Pausa zero (usado nos testes) não pausa nada.
func pausar(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	return base + time.Duration(rand.Int64N(int64(time.Second)))
}

// Novo aloca um Chrome headless com uma única aba, em pt-BR.
func Novo(ctx context.Context) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("lang", "pt-BR"))
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	return browserCtx, func() {
		cancelBrowser()
		cancelAlloc()
	}
}

func (c *Coletor) checarCaptcha(ctx context.Context) (bool, error) {
	var urlAtual, texto string
	if err := chromedp.Run(ctx,
		chromedp.Location(&urlAtual),
		chromedp.Evaluate(`document.body ? document.body.innerText : ""`, &texto),
	); err != nil {
		return false, err
	}
	return ehCaptcha(urlAtual, texto), nil
}

type resultadoBruto struct {
	Href string `json:"href"`
	Nome string `json:"nome"`
}

// Buscar rola o feed até ter max resultados ou até ele parar de crescer duas vezes seguidas.
func (c *Coletor) Buscar(ctx context.Context, consulta string, max int) ([]Resultado, error) {
	if p := pausar(c.Pausa); p > 0 {
		if err := chromedp.Run(ctx, chromedp.Sleep(p)); err != nil {
			return nil, err
		}
	}

	urlBusca := c.base() + "/maps/search/" + url.PathEscape(consulta) + "?hl=pt-BR"
	if err := chromedp.Run(ctx, chromedp.Navigate(urlBusca)); err != nil {
		return nil, err
	}
	if capturado, err := c.checarCaptcha(ctx); err != nil {
		return nil, err
	} else if capturado {
		return nil, ErrCaptcha
	}
	if err := chromedp.Run(ctx, chromedp.WaitVisible(`div[role="feed"]`, chromedp.ByQuery)); err != nil {
		return nil, err
	}

	anterior, paradas := 0, 0
	for {
		var atual int
		js := fmt.Sprintf(jsRolarFeedTpl, esperaAposRolagem.Milliseconds())
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &atual, comAwaitPromise)); err != nil {
			return nil, err
		}
		if atual >= max {
			break
		}
		if atual <= anterior {
			paradas++
			if paradas >= 2 {
				break
			}
		} else {
			paradas = 0
		}
		anterior = atual
	}

	var brutos []resultadoBruto
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsExtrairResultados, &brutos)); err != nil {
		return nil, err
	}

	var paginaURL string
	if err := chromedp.Run(ctx, chromedp.Location(&paginaURL)); err != nil {
		return nil, err
	}
	base, _ := url.Parse(paginaURL)

	vistos := map[string]bool{}
	var resultados []Resultado
	for _, b := range brutos {
		id := idDoLink(b.Href)
		if id == "" || vistos[id] {
			continue
		}
		vistos[id] = true
		resultados = append(resultados, Resultado{ID: id, Nome: b.Nome, URL: resolverURL(base, b.Href)})
		if len(resultados) >= max {
			break
		}
	}
	return resultados, nil
}

func resolverURL(base *url.URL, href string) string {
	ref, err := url.Parse(href)
	if err != nil || base == nil {
		return href
	}
	return base.ResolveReference(ref).String()
}

type lugarBruto struct {
	Nome            string `json:"nome"`
	NotaTexto       string `json:"notaTexto"`
	AvaliacoesTexto string `json:"avaliacoesTexto"`
	TelefoneItem    string `json:"telefoneItem"`
	Site            string `json:"site"`
	Endereco        string `json:"endereco"`
	Fechado         bool   `json:"fechado"`
}

func (c *Coletor) Abrir(ctx context.Context, destino string) (Lugar, error) {
	if p := pausar(c.Pausa); p > 0 {
		if err := chromedp.Run(ctx, chromedp.Sleep(p)); err != nil {
			return Lugar{}, err
		}
	}

	if err := chromedp.Run(ctx, chromedp.Navigate(destino)); err != nil {
		return Lugar{}, err
	}
	if capturado, err := c.checarCaptcha(ctx); err != nil {
		return Lugar{}, err
	} else if capturado {
		return Lugar{}, ErrCaptcha
	}
	if err := chromedp.Run(ctx, chromedp.WaitVisible(`h1`, chromedp.ByQuery)); err != nil {
		return Lugar{}, err
	}

	var bruto lugarBruto
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsExtrairLugar, &bruto)); err != nil {
		return Lugar{}, err
	}

	var urlAtual string
	if err := chromedp.Run(ctx, chromedp.Location(&urlAtual)); err != nil {
		return Lugar{}, err
	}

	return Lugar{
		ID:         idDoLink(urlAtual),
		Nome:       bruto.Nome,
		Endereco:   bruto.Endereco,
		Telefone:   telefoneDoItem(bruto.TelefoneItem),
		Site:       bruto.Site,
		Fechado:    bruto.Fechado,
		Nota:       nota(bruto.NotaTexto),
		Avaliacoes: contagem(bruto.AvaliacoesTexto),
	}, nil
}

type avaliacaoBruta struct {
	ID       string `json:"id"`
	Estrelas string `json:"estrelas"`
	Texto    string `json:"texto"`
}

// AvaliacoesDoAberto vale para o lugar aberto por último (Abrir). Rola o painel de avaliações
// 3 vezes: o número de itens carregados por rolagem varia e não vale a pena medir crescimento aqui.
func (c *Coletor) AvaliacoesDoAberto(ctx context.Context) (string, []string, error) {
	if err := chromedp.Run(ctx, chromedp.Click(`button[role="tab"][aria-label^="Avaliações"]`, chromedp.ByQuery)); err != nil {
		return "", nil, err
	}
	if err := chromedp.Run(ctx, chromedp.Sleep(esperaAposClique)); err != nil {
		return "", nil, err
	}
	if capturado, err := c.checarCaptcha(ctx); err != nil {
		return "", nil, err
	} else if capturado {
		return "", nil, ErrCaptcha
	}

	var resumo string
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsResumoGemini, &resumo)); err != nil {
		return "", nil, err
	}

	for i := 0; i < 3; i++ {
		var atual int
		js := fmt.Sprintf(jsRolarAvaliacoesTpl, esperaAposRolagem.Milliseconds())
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &atual, comAwaitPromise)); err != nil {
			return "", nil, err
		}
	}

	var brutas []avaliacaoBruta
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsExtrairAvaliacoes, &brutas)); err != nil {
		return "", nil, err
	}

	vistos := map[string]bool{}
	var negativas []string
	for _, b := range brutas {
		if b.ID == "" || vistos[b.ID] {
			continue
		}
		vistos[b.ID] = true
		if n := nota(b.Estrelas); n >= 1 && n <= 3 && len(negativas) < 10 {
			negativas = append(negativas, b.Texto)
		}
	}
	return resumo, negativas, nil
}
