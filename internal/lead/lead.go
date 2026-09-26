// Package lead contém as regras puras de classificação, pontuação e mensagem de um lead.
package lead

import (
	"math/rand/v2"
	"net/url"
	"strings"
	"unicode/utf8"
)

type Dor string

const (
	SemSite       Dor = "sem_site"
	SiteAgregador Dor = "site_agregador"
	SiteRuim      Dor = "site_ruim"
	Reclamacao    Dor = "reclamacao"
)

type Variante string

const (
	Texto Variante = "texto"
	Link  Variante = "link"
)

type Lead struct {
	PlaceID    string
	Nome       string
	Nicho      string
	Bairro     string
	Telefone   string
	Site       string
	Nota       float64
	Avaliacoes int
	Dores      []Dor
	Detalhes   map[Dor]string
	Pontuacao  int
	Variante   Variante
	Slug       string
	Mensagem   string
}

var agregadores = []string{
	"instagram.com", "facebook.com", "fb.com", "linktr.ee", "wa.me", "whatsapp.com",
	"ifood.com.br", "goomer.app", "anota.ai", "cardapioweb.com", "linkin.bio",
	"beacons.ai", "taplink.cc", "bio.link",
}

func ClassificarSite(uri string) Dor {
	if uri == "" {
		return SemSite
	}
	host := hostDe(uri)
	for _, d := range agregadores {
		if host == d || strings.HasSuffix(host, "."+d) {
			return SiteAgregador
		}
	}
	return ""
}

// hostDe extrai o host mesmo quando a URI vem sem esquema (place.WebsiteUri às vezes vem assim).
func hostDe(uri string) string {
	if !strings.Contains(uri, "://") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

func soDigitos(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ponytail: DDD 55 (Santa Maria/RS) é ambíguo com o DDI 55 já presente; sem caso real em João
// Pessoa, não vale desambiguar agora.
func Celular(tel string) (string, bool) {
	digitos := soDigitos(tel)
	if digitos == "" {
		return "", false
	}
	if !strings.HasPrefix(digitos, "55") {
		digitos = "55" + digitos
	}
	resto := digitos[2:]
	if len(resto) < 2 {
		return digitos, false
	}
	aposDDD := resto[2:]
	return digitos, len(aposDDD) == 9 && aposDDD[0] == '9'
}

type Porte struct {
	NotaMin       float64
	AvaliacoesMin int
}

var PortePadrao = Porte{NotaMin: 4.0, AvaliacoesMin: 30}

func (p Porte) Passa(nota float64, n int) bool {
	return nota >= p.NotaMin && n >= p.AvaliacoesMin
}

// ponytail: heurística ingênua (pesos a olho); calibrar pela taxa de resposta real por dor.
func Pontuar(l Lead) int {
	n := l.Avaliacoes
	if n > 500 {
		n = 500
	}
	pontos := n/5 + int((l.Nota-4)*50)
	for _, d := range l.Dores {
		switch d {
		case SemSite, SiteAgregador:
			pontos += 40
		case SiteRuim:
			pontos += 25
		case Reclamacao:
			pontos += 20
		}
	}
	return pontos
}

func SortearVariante(r *rand.Rand) Variante {
	if r.IntN(2) == 0 {
		return Texto
	}
	return Link
}

var acentos = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
	"Á", "A", "À", "A", "Â", "A", "Ã", "A", "Ä", "A",
	"É", "E", "È", "E", "Ê", "E", "Ë", "E",
	"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
	"Ó", "O", "Ò", "O", "Ô", "O", "Õ", "O", "Ö", "O",
	"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
	"Ç", "C", "Ñ", "N",
)

// os 4 últimos caracteres do PlaceID evitam colisão entre estabelecimentos de nome igual.
func Slug(nome, placeID string) string {
	s := strings.ToLower(acentos.Replace(nome))
	var b strings.Builder
	hifen := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			hifen = false
		default:
			if !hifen && b.Len() > 0 {
				b.WriteByte('-')
				hifen = true
			}
		}
	}
	base := strings.TrimRight(b.String(), "-")
	sufixo := placeID
	if len(sufixo) > 4 {
		sufixo = sufixo[len(sufixo)-4:]
	}
	return base + "-" + sufixo
}

// preço e link ficam fora do primeiro contato por decisão de produto.
func GanchoValido(g string) bool {
	if utf8.RuneCountInString(g) > 240 {
		return false
	}
	return !strings.Contains(g, "R$") && !strings.Contains(g, "http")
}

// fallback para quando a LLM falha ou não roda.
func GanchoPadrao(d Dor, detalhe string) string {
	switch d {
	case SemSite:
		return "vi que ainda não tem site"
	case SiteAgregador:
		return "vi que o contato de vocês hoje é só o " + detalhe
	case SiteRuim:
		return "tentei abrir o site de vocês e não carregou direito"
	case Reclamacao:
		return "vi um comentário sobre " + detalhe + " nas avaliações"
	default:
		return detalhe
	}
}

// a frase de opt-out é fixa: nunca reenviar sem essa saída.
func MontarMensagem(remetente, gancho string, v Variante, link string) string {
	msg := "oi, aqui é o " + remetente + ". " + gancho + ". faz sentido pra vocês?"
	msg += " se não fizer sentido, é só me avisar que não mando mais nada."
	if v == Link {
		msg += " montei um exemplo rápido pra ver: " + link
	}
	return msg
}
