package custo

import "testing"

func TestTeto(t *testing.T) {
	c := Conta{TokensEntrada: 1_000_000, TokensSaida: 1_000_000}
	got := c.Teto()
	want := 0.40 + 1.60
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
	c := Conta{TokensEntrada: 500_000, TokensSaida: 200_000}
	got := c.Teto()
	want := 500_000.0/1_000_000*0.40 + 200_000.0/1_000_000*1.60
	if got != want {
		t.Errorf("Teto() = %v, want %v", got, want)
	}
}
