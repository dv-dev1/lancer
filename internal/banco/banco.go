// Package banco persiste leads e descartes duráveis no Postgres, pra uma segunda coleta nunca reabrir um lugar já visto.
package banco

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dv-dev1/lancer/internal/lead"
)

//go:embed schema.sql
var schema string

type Banco struct {
	pool *pgxpool.Pool
}

// Descarte é a versão de banco.Gravar: cmd/coleta converte o próprio tipo pra este (R18 — banco não
// pode importar package main).
type Descarte struct {
	PlaceID, Motivo string
}

// Abrir cria o pool e aplica o schema embutido. pool.Exec sem argumento usa o protocolo simples do
// pgx, que aceita os vários "create table" do schema.sql numa chamada só.
func Abrir(ctx context.Context, url string) (*Banco, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("banco: abrir: %w", err)
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("banco: aplicar schema: %w", err)
	}
	return &Banco{pool: pool}, nil
}

func (b *Banco) JaVistos(ctx context.Context, ids []string) (map[string]bool, error) {
	linhas, err := b.pool.Query(ctx, `
		select place_id from leads where place_id = any($1)
		union
		select place_id from vistos where place_id = any($1)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("banco: JaVistos: %w", err)
	}
	defer linhas.Close()

	vistos := map[string]bool{}
	for linhas.Next() {
		var id string
		if err := linhas.Scan(&id); err != nil {
			return nil, fmt.Errorf("banco: JaVistos: %w", err)
		}
		vistos[id] = true
	}
	return vistos, linhas.Err()
}

// pedido 0 = coleta do terminal, sem pedido do painel.
func (b *Banco) Gravar(ctx context.Context, leads []lead.Lead, descartes []Descarte, pedido int64) error {
	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("banco: Gravar: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, l := range leads {
		var avaliacoes any // R12: -1 do coletor (aba de avaliações ausente) vira NULL
		if l.Avaliacoes >= 0 {
			avaliacoes = l.Avaliacoes
		}
		dores := make([]string, len(l.Dores))
		for i, d := range l.Dores {
			dores[i] = string(d)
		}
		detalhes, err := json.Marshal(l.Detalhes)
		if err != nil {
			return fmt.Errorf("banco: Gravar: detalhes de %s: %w", l.PlaceID, err)
		}
		_, err = tx.Exec(ctx, `
			insert into leads (place_id, slug, nicho, cidade, bairro, nome, telefone, site, endereco, nota, avaliacoes, dores, detalhes, pontuacao, variante, mensagem, pedido_id)
			values ($1, $2, $3, coalesce(nullif($4, ''), 'João Pessoa'), $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, nullif($17, 0))
			on conflict (place_id) do nothing
		`, l.PlaceID, l.Slug, l.Nicho, l.Cidade, l.Bairro, l.Nome, l.Telefone, l.Site, l.Endereco, l.Nota, avaliacoes, dores, detalhes, l.Pontuacao, string(l.Variante), l.Mensagem, pedido)
		if err != nil {
			return fmt.Errorf("banco: Gravar: lead %s: %w", l.PlaceID, err)
		}
	}

	for _, d := range descartes {
		if d.PlaceID == "" { // erro isolado, não durável — não entra em vistos, volta a ser tentado
			continue
		}
		if _, err := tx.Exec(ctx, `insert into vistos (place_id, motivo) values ($1, $2) on conflict do nothing`, d.PlaceID, d.Motivo); err != nil {
			return fmt.Errorf("banco: Gravar: descarte %s: %w", d.PlaceID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("banco: Gravar: %w", err)
	}
	return nil
}

func (b *Banco) SomarCusto(ctx context.Context, dia time.Time, api string, unidades int64, usd float64) error {
	_, err := b.pool.Exec(ctx, `
		insert into custos (dia, api, unidades, usd) values ($1, $2, $3, $4)
		on conflict (dia, api) do update set unidades = custos.unidades + excluded.unidades, usd = custos.usd + excluded.usd
	`, dia, api, unidades, usd)
	if err != nil {
		return fmt.Errorf("banco: SomarCusto: %w", err)
	}
	return nil
}

func (b *Banco) MarcarPerdidos(ctx context.Context) (int64, error) {
	tag, err := b.pool.Exec(ctx, `
		update leads set etapa = 'perdido', saiu_de = etapa
		where etapa in ('contatado', 'abriu') and follow_up_em < now() - interval '3 days' and respondeu_em is null
	`)
	if err != nil {
		return 0, fmt.Errorf("banco: MarcarPerdidos: %w", err)
	}
	return tag.RowsAffected(), nil
}
