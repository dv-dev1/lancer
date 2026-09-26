// Package custo acumula contadores de uso da coleta; o preço por contador entra na T5.
package custo

type Conta struct {
	IDs           int
	Detalhes      int
	Atmosfera     int
	TokensEntrada int
	TokensSaida   int
}
