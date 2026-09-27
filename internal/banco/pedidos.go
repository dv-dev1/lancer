package banco

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Pedido struct {
	ID                    int64
	Cidade, Bairro, Nicho string
	Limite                int
}

// PegarPedido marca o pendente mais antigo como rodando e o devolve; nil quando não há nenhum.
// skip locked: dois coletores ao mesmo tempo nunca pegam o mesmo pedido.
func (b *Banco) PegarPedido(ctx context.Context) (*Pedido, error) {
	var p Pedido
	err := b.pool.QueryRow(ctx, `
		update pedidos set estado = 'rodando'
		where id = (select id from pedidos where estado = 'pendente' order by id limit 1 for update skip locked)
		returning id, cidade, bairro, nicho, limite
	`).Scan(&p.ID, &p.Cidade, &p.Bairro, &p.Nicho, &p.Limite)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("banco: PegarPedido: %w", err)
	}
	return &p, nil
}

func (b *Banco) ConcluirPedido(ctx context.Context, id int64, novos int) error {
	if _, err := b.pool.Exec(ctx, `update pedidos set estado = 'pronto', novos = $2, terminado_em = now() where id = $1`, id, novos); err != nil {
		return fmt.Errorf("banco: ConcluirPedido: %w", err)
	}
	return nil
}

// ponytail: filtra por estado, não por id: vale com um coletor só (fora de escopo ter vários).
// FalharPedidos marca como erro os pedidos nos estados dados: o que rodava quando a coleta caiu, ou
// todos os pendentes quando o captcha para o dia.
func (b *Banco) FalharPedidos(ctx context.Context, motivo string, estados ...string) error {
	if _, err := b.pool.Exec(ctx, `update pedidos set estado = 'erro', erro = $1, terminado_em = now() where estado = any($2)`, motivo, estados); err != nil {
		return fmt.Errorf("banco: FalharPedidos: %w", err)
	}
	return nil
}
