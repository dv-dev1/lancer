package maps

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// temChrome cobre os dois jeitos comuns de achar o binário: no PATH (Linux) ou o .app do macOS.
func temChrome() bool {
	for _, nome := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
		if _, err := exec.LookPath(nome); err == nil {
			return true
		}
	}
	_, err := os.Stat("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	return err == nil
}

const paginaBusca = `<html><body><div role="feed">
  <a href="/maps/place/Loja+Um/data=!19sChIJum111" aria-label="Loja Um">Loja Um</a>
  <a href="/maps/place/Loja+Dois/data=!19sChIJdois222" aria-label="Loja Dois">Loja Dois</a>
  <a href="/maps/place/Loja+Tres/data=!19sChIJtres333" aria-label="Loja Três">Loja Três</a>
</div></body></html>`

const paginaCaptcha = `<html><body><p>Nossos sistemas detectaram tráfego incomum vindo da sua rede.</p></body></html>`

// paginaLugar monta o fixture do lugar: completo tem telefone e site, sem site não tem o data-item-id authority.
func paginaLugar(comSite bool) string {
	site := `<a data-item-id="authority" href="https://exemplo.com.br">exemplo.com.br</a>`
	if !comSite {
		site = ""
	}
	return `<html><body>
  <h1>Confeitaria Teste</h1>
  <span role="img" aria-label="4,5 estrelas">4,5</span>
  <span aria-label="279 avaliações">279 avaliações</span>
  <button data-item-id="phone:tel:083991355466">(83) 99135-5466</button>
  ` + site + `
  <button data-item-id="address" aria-label="Endereço: Rua das Flores, 100 - Manaíra">Rua das Flores, 100</button>
  <button role="tab" aria-label="Avaliações">Avaliações</button>
  <div>
    <div>
      <div>
        <span>As avaliações destacam bom atendimento e preços justos para a região, com elogios recorrentes ao brigadeiro.+18</span>
        <span>Resumo feito com o Gemini</span>
      </div>
    </div>
  </div>
  <div style="overflow-y:auto;height:100px">
    <button aria-label="Classificar avaliações">Classificar</button>
    <div data-review-id="r1"><span aria-label="5 estrelas">5 estrelas</span><span lang="pt">Muito bom, recomendo!</span></div>
    <div data-review-id="r2"><span aria-label="2 estrelas">2 estrelas</span><span lang="pt">Demorou muito para o pedido chegar…
Mais</span></div>
    <div data-review-id="r3"><span aria-label="1 estrela">1 estrela</span><span lang="pt">Atendimento péssimo</span></div>
    <div data-review-id="r2"><span aria-label="2 estrelas">2 estrelas</span><span lang="pt">Demorou muito para o pedido chegar…
Mais</span></div>
  </div>
</body></html>`
}

func novoServidorFixture() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/maps/search/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(paginaBusca))
	})
	mux.HandleFunc("/maps/place/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(paginaLugar(!strings.Contains(r.URL.RawQuery, "semsite"))))
	})
	mux.HandleFunc("/sorry/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(paginaCaptcha))
	})
	return httptest.NewServer(mux)
}

func TestColetorComFixture(t *testing.T) {
	if !temChrome() {
		t.Skip("Chrome/Chromium não encontrado, pulando teste com navegador real")
	}

	// waits curtos só no teste: a produção precisa do tempo real pro Google carregar assíncrono.
	esperaAposClique = 200 * time.Millisecond
	esperaAposRolagem = 200 * time.Millisecond

	srv := novoServidorFixture()
	defer srv.Close()

	ctx, cancelNovo := Novo(context.Background())
	defer cancelNovo()
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()

	c := &Coletor{Base: srv.URL, Pausa: 0}

	t.Run("busca ate atingir o max", func(t *testing.T) {
		resultados, err := c.Buscar(ctx, "confeitaria em Manaíra", 2)
		if err != nil {
			t.Fatalf("Buscar: %v", err)
		}
		if len(resultados) != 2 {
			t.Fatalf("len(resultados) = %d, want 2", len(resultados))
		}
	})

	t.Run("busca ate parar de crescer", func(t *testing.T) {
		resultados, err := c.Buscar(ctx, "confeitaria em Manaíra", 10)
		if err != nil {
			t.Fatalf("Buscar: %v", err)
		}
		if len(resultados) != 3 {
			t.Fatalf("len(resultados) = %d, want 3 (só existem 3 no fixture)", len(resultados))
		}
		vistos := map[string]bool{}
		for _, r := range resultados {
			if r.ID == "" {
				t.Errorf("resultado sem ID: %+v", r)
			}
			if vistos[r.ID] {
				t.Errorf("ID duplicado: %s", r.ID)
			}
			vistos[r.ID] = true
			if r.Nome == "" || r.URL == "" {
				t.Errorf("resultado incompleto: %+v", r)
			}
		}
	})

	t.Run("lugar completo", func(t *testing.T) {
		lugar, err := c.Abrir(ctx, srv.URL+"/maps/place/x/data=!19sChIJcompleto123")
		if err != nil {
			t.Fatalf("Abrir: %v", err)
		}
		if lugar.ID != "ChIJcompleto123" {
			t.Errorf("ID = %q, want %q", lugar.ID, "ChIJcompleto123")
		}
		if lugar.Nome != "Confeitaria Teste" {
			t.Errorf("Nome = %q, want %q", lugar.Nome, "Confeitaria Teste")
		}
		if lugar.Nota != 4.5 {
			t.Errorf("Nota = %v, want 4.5", lugar.Nota)
		}
		if lugar.Avaliacoes != 279 {
			t.Errorf("Avaliacoes = %v, want 279", lugar.Avaliacoes)
		}
		if lugar.Telefone != "83991355466" {
			t.Errorf("Telefone = %q, want %q", lugar.Telefone, "83991355466")
		}
		if lugar.Site != "https://exemplo.com.br" {
			t.Errorf("Site = %q, want %q", lugar.Site, "https://exemplo.com.br")
		}
		if lugar.Endereco != "Rua das Flores, 100 - Manaíra" {
			t.Errorf("Endereco = %q, want %q", lugar.Endereco, "Rua das Flores, 100 - Manaíra")
		}
		if lugar.Fechado {
			t.Errorf("Fechado = true, want false")
		}

		resumo, negativas, err := c.AvaliacoesDoAberto(ctx)
		if err != nil {
			t.Fatalf("AvaliacoesDoAberto: %v", err)
		}
		wantResumo := "As avaliações destacam bom atendimento e preços justos para a região, com elogios recorrentes ao brigadeiro."
		if resumo != wantResumo {
			t.Errorf("resumo = %q, want %q", resumo, wantResumo)
		}
		if len(negativas) != 2 {
			t.Fatalf("len(negativas) = %d, want 2 (r2 dedup + r3): %v", len(negativas), negativas)
		}
		for _, n := range negativas {
			if strings.Contains(n, "Mais") {
				t.Errorf("texto não cortou o sufixo \"…\\nMais\": %q", n)
			}
		}
	})

	t.Run("lugar sem site", func(t *testing.T) {
		lugar, err := c.Abrir(ctx, srv.URL+"/maps/place/x?semsite=1")
		if err != nil {
			t.Fatalf("Abrir: %v", err)
		}
		if lugar.Site != "" {
			t.Errorf("Site = %q, want vazio", lugar.Site)
		}
		if lugar.Nome != "Confeitaria Teste" {
			t.Errorf("Nome = %q, want %q", lugar.Nome, "Confeitaria Teste")
		}
	})

	t.Run("captcha", func(t *testing.T) {
		_, err := c.Abrir(ctx, srv.URL+"/sorry/index")
		if err != ErrCaptcha {
			t.Fatalf("Abrir em página de captcha: err = %v, want ErrCaptcha", err)
		}
	})
}

// TestMapsAoVivo é a checagem rápida pra quando o Google mudar o layout; só roda sob pedido
// porque bate no Google Maps real (fora dos termos de uso, sujeito a captcha).
func TestMapsAoVivo(t *testing.T) {
	if os.Getenv("LANCER_MAPS_AO_VIVO") != "1" {
		t.Skip("defina LANCER_MAPS_AO_VIVO=1 pra rodar contra o Google Maps real")
	}

	ctx, cancelNovo := Novo(context.Background())
	defer cancelNovo()
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()

	c := &Coletor{Pausa: PausaPadrao}
	resultados, err := c.Buscar(ctx, "confeitaria em Manaíra, João Pessoa - PB", 5)
	if err != nil {
		t.Fatalf("Buscar: %v", err)
	}
	if len(resultados) == 0 {
		t.Fatal("nenhum resultado")
	}

	lugar, err := c.Abrir(ctx, resultados[0].URL)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	if lugar.Nome == "" {
		t.Error("Nome vazio")
	}
	if lugar.Nota == 0 {
		t.Error("Nota vazia")
	}
	if lugar.Avaliacoes == 0 {
		t.Error("Avaliacoes vazia")
	}
}
