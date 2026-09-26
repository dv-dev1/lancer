package lead

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestClassificarSite(t *testing.T) {
	casos := []struct {
		uri  string
		want Dor
	}{
		{"", SemSite},
		{"https://instagram.com/x", SiteAgregador},
		{"http://www.instagram.com/x", SiteAgregador},
		{"instagram.com/x", SiteAgregador},
		{"https://www.ifood.com.br/delivery/loja", SiteAgregador},
		{"https://wa.me/5583999999999", SiteAgregador},
		{"https://linktr.ee/x", SiteAgregador},
		{"https://docesabor.com.br", ""},
		{"https://notinstagram.com", ""},
	}
	for _, c := range casos {
		if got := ClassificarSite(c.uri); got != c.want {
			t.Errorf("ClassificarSite(%q) = %q, want %q", c.uri, got, c.want)
		}
	}
}

func TestCelular(t *testing.T) {
	casos := []struct {
		tel      string
		want     string
		wantBool bool
	}{
		{"+55 83 99876-5432", "5583998765432", true},
		{"+55 83 3222-1234", "558332221234", false},
		{"", "", false},
		{"(83) 99876-5432", "5583998765432", true},
	}
	for _, c := range casos {
		got, ok := Celular(c.tel)
		if got != c.want || ok != c.wantBool {
			t.Errorf("Celular(%q) = (%q, %v), want (%q, %v)", c.tel, got, ok, c.want, c.wantBool)
		}
	}
}

func TestPortePassa(t *testing.T) {
	casos := []struct {
		nota float64
		n    int
		want bool
	}{
		{4.0, 30, true},
		{3.9, 500, false},
		{4.8, 29, false},
		{4.5, -1, true},  // R12: avaliações desconhecida (-1) cai pra só a nota
		{3.9, -1, false}, // idem, mas nota abaixo do mínimo
	}
	for _, c := range casos {
		if got := PortePadrao.Passa(c.nota, c.n); got != c.want {
			t.Errorf("PortePadrao.Passa(%v, %v) = %v, want %v", c.nota, c.n, got, c.want)
		}
	}
}

func TestPontuar(t *testing.T) {
	semSite300 := Lead{Nota: 4.0, Avaliacoes: 300, Dores: []Dor{SemSite}}
	siteRuim300 := Lead{Nota: 4.0, Avaliacoes: 300, Dores: []Dor{SiteRuim}}
	semDor300 := Lead{Nota: 4.0, Avaliacoes: 300}
	if !(Pontuar(semSite300) > Pontuar(siteRuim300)) {
		t.Errorf("sem site (%d) devia pontuar mais que site ruim (%d)", Pontuar(semSite300), Pontuar(siteRuim300))
	}
	if !(Pontuar(siteRuim300) > Pontuar(semDor300)) {
		t.Errorf("site ruim (%d) devia pontuar mais que sem dor (%d)", Pontuar(siteRuim300), Pontuar(semDor300))
	}

	semSite100 := Lead{Nota: 4.0, Avaliacoes: 100, Dores: []Dor{SemSite}}
	if !(Pontuar(semSite300) > Pontuar(semSite100)) {
		t.Errorf("com a mesma dor, 300 avaliações (%d) devia pontuar mais que 100 (%d)", Pontuar(semSite300), Pontuar(semSite100))
	}

	// R12: avaliações desconhecida (-1) não pode pontuar negativo nem diferente de 0 avaliações conhecidas.
	desconhecida := Lead{Nota: 4.0, Avaliacoes: -1, Dores: []Dor{SemSite}}
	zero := Lead{Nota: 4.0, Avaliacoes: 0, Dores: []Dor{SemSite}}
	if Pontuar(desconhecida) < 0 {
		t.Errorf("Pontuar com Avaliacoes desconhecida (-1) não pode ser negativo: %d", Pontuar(desconhecida))
	}
	if Pontuar(desconhecida) != Pontuar(zero) {
		t.Errorf("Avaliacoes desconhecida (-1) devia pontuar igual a 0 conhecidas: %d vs %d", Pontuar(desconhecida), Pontuar(zero))
	}
}

func TestSortearVarianteDistribuicao(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	texto := 0
	for range 1000 {
		if SortearVariante(r) == Texto {
			texto++
		}
	}
	if texto < 450 || texto > 550 {
		t.Errorf("esperava entre 450 e 550 Texto em 1000 sorteios, veio %d", texto)
	}
}

func TestSlug(t *testing.T) {
	got := Slug("Pão de Açúcar & Cia", "ABCDEF1234")
	if !strings.HasPrefix(got, "pao-de-acucar-cia-") {
		t.Errorf("Slug(...) = %q, want prefix %q", got, "pao-de-acucar-cia-")
	}
	if !strings.HasSuffix(got, "-1234") {
		t.Errorf("Slug(...) = %q, want suffix com os 4 últimos caracteres do ID", got)
	}
}

func TestMontarMensagem(t *testing.T) {
	const remetente = "Daniel"
	const gancho = "vi que ainda não tem site"
	const link = "https://lancer.exemplo/p/padaria-abcd"
	const fraseSaida = "se não fizer sentido, é só me avisar que não mando mais nada"

	texto := MontarMensagem(remetente, gancho, Texto, link)
	if strings.Contains(texto, link) {
		t.Errorf("variante Texto não devia conter o link, mensagem: %q", texto)
	}
	if !strings.Contains(texto, remetente) || !strings.Contains(texto, fraseSaida) {
		t.Errorf("variante Texto devia conter remetente e a frase de saída, mensagem: %q", texto)
	}

	comLink := MontarMensagem(remetente, gancho, Link, link)
	if !strings.Contains(comLink, link) {
		t.Errorf("variante Link devia conter o link, mensagem: %q", comLink)
	}
	if !strings.Contains(comLink, remetente) || !strings.Contains(comLink, fraseSaida) {
		t.Errorf("variante Link devia conter remetente e a frase de saída, mensagem: %q", comLink)
	}
}

func TestGanchoValido(t *testing.T) {
	casos := []struct {
		gancho string
		want   bool
	}{
		{"vi que ainda não tem site", true},
		{"R$ 50", false},
		{"https://", false},
		{strings.Repeat("a", 241), false},
	}
	for _, c := range casos {
		if got := GanchoValido(c.gancho); got != c.want {
			t.Errorf("GanchoValido(%q) = %v, want %v", c.gancho, got, c.want)
		}
	}
}

func TestGanchoPadrao(t *testing.T) {
	for _, d := range []Dor{SemSite, SiteAgregador, SiteRuim, Reclamacao} {
		g := GanchoPadrao(d, "detalhe qualquer")
		if g == "" || !GanchoValido(g) {
			t.Errorf("GanchoPadrao(%q, ...) = %q, want gancho válido e não vazio", d, g)
		}
	}
}
