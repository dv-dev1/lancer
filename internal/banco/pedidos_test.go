package banco

import (
	"context"
	"testing"

	"github.com/dv-dev1/lancer/internal/lead"
)

func TestCicloDoPedido(t *testing.T) {
	b := abrirBancoTeste(t)
	ctx := context.Background()
	if _, err := b.pool.Exec(ctx, `insert into pedidos (cidade, bairro, nicho, limite) values ('Campina Grande', '', 'confeitaria', 2), ('João Pessoa', 'Manaíra', 'loja', 3)`); err != nil {
		t.Fatal(err)
	}

	p, err := b.PegarPedido(ctx)
	if err != nil || p == nil || p.Cidade != "Campina Grande" || p.Limite != 2 {
		t.Fatalf("1º PegarPedido = %+v, %v; quero o de Campina Grande", p, err)
	}
	l := leadTeste("place-pedido")
	l.Cidade = "Campina Grande"
	if err := b.Gravar(ctx, []lead.Lead{l}, nil, p.ID); err != nil {
		t.Fatal(err)
	}
	if err := b.ConcluirPedido(ctx, p.ID, 1); err != nil {
		t.Fatal(err)
	}
	var cidade string
	var pedidoID int64
	if err := b.pool.QueryRow(ctx, `select cidade, pedido_id from leads where place_id = 'place-pedido'`).Scan(&cidade, &pedidoID); err != nil {
		t.Fatal(err)
	}
	if cidade != "Campina Grande" || pedidoID != p.ID {
		t.Errorf("lead gravado com cidade %q e pedido %d; quero Campina Grande e %d", cidade, pedidoID, p.ID)
	}

	if err := b.FalharPedidos(ctx, "captcha", "pendente"); err != nil {
		t.Fatal(err)
	}
	if p, err := b.PegarPedido(ctx); err != nil || p != nil {
		t.Errorf("PegarPedido depois do captcha = %+v, %v; quero nil", p, err)
	}

	var estados []string
	if err := b.pool.QueryRow(ctx, `select array_agg(estado || ':' || coalesce(novos::text, erro) order by id) from pedidos`).Scan(&estados); err != nil {
		t.Fatal(err)
	}
	if len(estados) != 2 || estados[0] != "pronto:1" || estados[1] != "erro:captcha" {
		t.Errorf("estados = %v; quero [pronto:1 erro:captcha]", estados)
	}

	var semPedido *int64
	if err := b.Gravar(ctx, []lead.Lead{leadTeste("place-terminal")}, nil, 0); err != nil {
		t.Fatal(err)
	}
	if err := b.pool.QueryRow(ctx, `select pedido_id from leads where place_id = 'place-terminal'`).Scan(&semPedido); err != nil || semPedido != nil {
		t.Errorf("coleta do terminal gravou pedido_id %v, %v; quero NULL", semPedido, err)
	}
}
