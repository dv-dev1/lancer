package maps

import "testing"

func TestIdDoLink(t *testing.T) {
	casos := []struct {
		href string
		want string
	}{
		{"/maps/place/Padaria/data=!4m6!3m5!1s0x7a3:0x9b2!8m2!3d-7.1!4d-34.8!16s%2Fg%2F11abc!19sChIJabc123?hl=pt-BR", "ChIJabc123"},
		{"/maps/place/Sem19s/data=!4m6!3m5!1s0x7a3:0x9b2!8m2!3d-7.1!4d-34.8", "0x7a3:0x9b2"},
		{"/maps/place/SemNada/data=!4m6!3m5", ""},
	}
	for _, c := range casos {
		if got := idDoLink(c.href); got != c.want {
			t.Errorf("idDoLink(%q) = %q, want %q", c.href, got, c.want)
		}
	}
}

func TestNota(t *testing.T) {
	casos := []struct {
		s    string
		want float64
	}{
		{"4,5 estrelas", 4.5},
		{"5 estrelas", 5},
		{"1 estrela", 1},
		{"", 0},
	}
	for _, c := range casos {
		if got := nota(c.s); got != c.want {
			t.Errorf("nota(%q) = %v, want %v", c.s, got, c.want)
		}
	}
}

func TestContagem(t *testing.T) {
	casos := []struct {
		s    string
		want int
	}{
		{"279 avaliações", 279},
		{"1.234 avaliações", 1234},
		{"", 0},
	}
	for _, c := range casos {
		if got := contagem(c.s); got != c.want {
			t.Errorf("contagem(%q) = %v, want %v", c.s, got, c.want)
		}
	}
}

func TestTelefoneDoItem(t *testing.T) {
	casos := []struct {
		s    string
		want string
	}{
		{"phone:tel:083991355466", "83991355466"},
		{"phone:tel:08333221234", "8333221234"},
	}
	for _, c := range casos {
		if got := telefoneDoItem(c.s); got != c.want {
			t.Errorf("telefoneDoItem(%q) = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestEhCaptcha(t *testing.T) {
	casos := []struct {
		url, texto string
		want       bool
	}{
		{"https://www.google.com/sorry/index?continue=x", "", true},
		{"https://www.google.com/maps/place/x", "detectamos tráfego incomum na sua rede", true},
		{"https://www.google.com/maps/place/x", "we have detected unusual traffic from your network", true},
		{"https://www.google.com/maps/place/x", "Padaria Boa Vista", false},
	}
	for _, c := range casos {
		if got := ehCaptcha(c.url, c.texto); got != c.want {
			t.Errorf("ehCaptcha(%q, %q) = %v, want %v", c.url, c.texto, got, c.want)
		}
	}
}
