package main

import (
	"flag"
	"io"
	"os"
	"testing"
)

// TestRunValidaFlags cobre o R-fecho-11: --cidade vazia e --limite <= 0 saem com código 2, antes de
// qualquer variável de ambiente ou Chrome — por isso só cobre os casos inválidos: um caso válido
// chegaria em maps.Novo (sobe Chrome de verdade), que este pacote de teste não pode fazer rodar.
func TestRunValidaFlags(t *testing.T) {
	casos := []struct {
		nome string
		args []string
		want int
	}{
		{"cidade vazia", []string{"coleta", "--nicho=confeitaria", "--cidade="}, 2},
		{"limite zero", []string{"coleta", "--nicho=confeitaria", "--bairro=Manaíra", "--limite=0"}, 2},
		{"limite negativo", []string{"coleta", "--nicho=confeitaria", "--bairro=Manaíra", "--limite=-1"}, 2},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			argsOrig, flagsOrig := os.Args, flag.CommandLine
			t.Cleanup(func() { os.Args, flag.CommandLine = argsOrig, flagsOrig })
			os.Args = c.args
			flag.CommandLine = flag.NewFlagSet(c.args[0], flag.ContinueOnError)
			flag.CommandLine.SetOutput(io.Discard)

			if got := run(); got != c.want {
				t.Errorf("run() = %d, want %d", got, c.want)
			}
		})
	}
}
