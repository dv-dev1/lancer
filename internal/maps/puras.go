// Package maps coleta leads no Google Maps com navegador headless: sem API paga, fora dos termos de uso.
package maps

import (
	"regexp"
	"strconv"
	"strings"
)

var reNumero = regexp.MustCompile(`[\d,]+`)
var reContagem = regexp.MustCompile(`[\d.]+`)

// idDoLink prefere o !19s(ChIJ…); nem todo resultado tem, então cai no par !1s(0x…:0x…).
func idDoLink(href string) string {
	if id := extrairEntre(href, "!19s"); id != "" {
		return id
	}
	return extrairEntre(href, "!1s")
}

func extrairEntre(s, marcador string) string {
	i := strings.Index(s, marcador)
	if i < 0 {
		return ""
	}
	resto := s[i+len(marcador):]
	fim := strings.IndexAny(resto, "!?")
	if fim < 0 {
		fim = len(resto)
	}
	return resto[:fim]
}

func nota(s string) float64 {
	m := reNumero.FindString(s)
	if m == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(strings.Replace(m, ",", ".", 1), 64)
	return v
}

// contagem: o ponto em "1.234" é separador de milhar, não decimal.
func contagem(s string) int {
	m := reContagem.FindString(s)
	if m == "" {
		return 0
	}
	n, _ := strconv.Atoi(strings.ReplaceAll(m, ".", ""))
	return n
}

// telefoneDoItem tira o 0 de tronco que o Maps antepõe ao DDD.
func telefoneDoItem(s string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, "phone:tel:"), "0")
}

func ehCaptcha(url, texto string) bool {
	if strings.Contains(url, "/sorry/") {
		return true
	}
	t := strings.ToLower(texto)
	return strings.Contains(t, "tráfego incomum") || strings.Contains(t, "unusual traffic")
}
