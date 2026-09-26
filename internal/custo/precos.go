package custo

// Fonte: mesma tabela de geniai-agente/cerebro/agente/grafo.py:43, revalidado em 2026-09-26.
const (
	precoTokenEntradaPorMilhao = 0.40 // gpt-4.1-mini, tokens de entrada
	precoTokenSaidaPorMilhao   = 1.60 // gpt-4.1-mini, tokens de saída
)

// Teto soma o custo em US$ da OpenAI. A coleta no Maps é grátis (navegador, sem API paga) e não entra aqui.
func (c Conta) Teto() float64 {
	return float64(c.TokensEntrada)/1_000_000*precoTokenEntradaPorMilhao +
		float64(c.TokensSaida)/1_000_000*precoTokenSaidaPorMilhao
}
