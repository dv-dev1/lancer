// Package custo acumula contadores de uso da coleta; o preço por contador entra na T5.
package custo

type Conta struct {
	TokensEntrada int
	TokensSaida   int
}
