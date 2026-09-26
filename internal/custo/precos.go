package custo

// Fonte: developers.google.com/maps/billing-and-pricing/pricing, revalidado em 2026-09-26.
// Text Search Essentials (IDs Only) é grátis, sem limite — não entra no teto.
const (
	precoDetalhesPorMil  = 20.00 // Place Details Enterprise; 1.000 chamadas grátis por mês
	precoAtmosferaPorMil = 25.00 // Place Details Enterprise + Atmosphere; 1.000 chamadas grátis por mês
)

// Fonte: mesma tabela de geniai-agente/cerebro/agente/grafo.py:43, revalidado em 2026-09-26.
const (
	precoTokenEntradaPorMilhao = 0.40 // gpt-4.1-mini, tokens de entrada
	precoTokenSaidaPorMilhao   = 1.60 // gpt-4.1-mini, tokens de saída
)

// Teto soma o custo em US$ como se a cota grátis mensal já estivesse zerada: é o pior caso do dia, não a média do mês.
func (c Conta) Teto() float64 {
	return float64(c.Detalhes)/1000*precoDetalhesPorMil +
		float64(c.Atmosfera)/1000*precoAtmosferaPorMil +
		float64(c.TokensEntrada)/1_000_000*precoTokenEntradaPorMilhao +
		float64(c.TokensSaida)/1_000_000*precoTokenSaidaPorMilhao
}
