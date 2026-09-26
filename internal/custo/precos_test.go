package custo

import "testing"

func TestTeto(t *testing.T) {
	c := Conta{Detalhes: 1000, Atmosfera: 1000, TokensEntrada: 1_000_000, TokensSaida: 1_000_000}
	got := c.Teto()
	want := 20.00 + 25.00 + 0.40 + 1.60
	if got != want {
		t.Errorf("Teto() = %v, want %v", got, want)
	}
}

func TestTetoZerado(t *testing.T) {
	if got := (Conta{}).Teto(); got != 0 {
		t.Errorf("Teto() de Conta zerada = %v, want 0", got)
	}
}

func TestTetoProporcional(t *testing.T) {
	c := Conta{Detalhes: 500, Atmosfera: 200}
	got := c.Teto()
	want := 500.0/1000*20.00 + 200.0/1000*25.00
	if got != want {
		t.Errorf("Teto() = %v, want %v", got, want)
	}
}
