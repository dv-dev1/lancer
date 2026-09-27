package banco

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dv-dev1/lancer/internal/lead"
)

// abrirBancoTeste isola cada teste num schema Postgres próprio (dropado no fim): assim os testes
// rodam contra o Neon real sem um pisar no dado do outro. Não usa Abrir: o Neon aceita search_path
// no RuntimeParams do startup packet (application_name, por exemplo, funciona) mas o ignora — só o
// SET explícito por conexão nova (AfterConnect) isola o schema de fato.
func abrirBancoTeste(t *testing.T) *Banco {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL não configurada, pulando teste de banco")
	}
	ctx := context.Background()

	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("conectar: %v", err)
	}

	nomeSchema := fmt.Sprintf("teste_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "create schema "+nomeSchema); err != nil {
		admin.Close()
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		defer admin.Close()
		if _, err := admin.Exec(context.Background(), "drop schema "+nomeSchema+" cascade"); err != nil {
			t.Errorf("drop schema %s: %v", nomeSchema, err)
		}
	})

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "set search_path to "+nomeSchema)
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("NewWithConfig: %v", err)
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		t.Fatalf("aplicar schema: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return &Banco{pool: pool}
}

func leadTeste(placeID string) lead.Lead {
	return lead.Lead{
		PlaceID:   placeID,
		Slug:      placeID + "-slug",
		Nicho:     "confeitaria",
		Bairro:    "Manaíra",
		Nome:      "Doceria " + placeID,
		Telefone:  "5583999999999",
		Nota:      4.5,
		Dores:     []lead.Dor{lead.SemSite},
		Detalhes:  map[lead.Dor]string{lead.SemSite: ""},
		Pontuacao: 10,
		Variante:  lead.Texto,
		Mensagem:  "oi",
	}
}

func TestGravarDuasVezesNaoDuplica(t *testing.T) {
	b := abrirBancoTeste(t)
	ctx := context.Background()
	l := leadTeste("place-dup")

	if err := b.Gravar(ctx, []lead.Lead{l}, nil); err != nil {
		t.Fatalf("1ª gravação: %v", err)
	}
	if err := b.Gravar(ctx, []lead.Lead{l}, nil); err != nil {
		t.Fatalf("2ª gravação: %v", err)
	}

	var n int
	if err := b.pool.QueryRow(ctx, "select count(*) from leads where place_id = $1", l.PlaceID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("count(leads) = %d, want 1 (Gravar repetido não duplica)", n)
	}
}

func TestJaVistosEnxergaLeadsEVistos(t *testing.T) {
	b := abrirBancoTeste(t)
	ctx := context.Background()

	if err := b.Gravar(ctx, []lead.Lead{leadTeste("place-lead")}, []Descarte{{PlaceID: "place-descarte", Motivo: "fechado"}}); err != nil {
		t.Fatal(err)
	}

	vistos, err := b.JaVistos(ctx, []string{"place-lead", "place-descarte", "place-novo"})
	if err != nil {
		t.Fatal(err)
	}
	if !vistos["place-lead"] {
		t.Error(`vistos["place-lead"] = false, want true (veio de leads)`)
	}
	if !vistos["place-descarte"] {
		t.Error(`vistos["place-descarte"] = false, want true (veio de vistos)`)
	}
	if vistos["place-novo"] {
		t.Error(`vistos["place-novo"] = true, want false (nunca visto)`)
	}
}

func TestMarcarPerdidosSoAposTresDiasSemResposta(t *testing.T) {
	b := abrirBancoTeste(t)
	ctx := context.Background()
	agora := time.Now()

	casos := []struct {
		placeID     string
		etapa       string
		followUp    time.Time
		comResposta bool
	}{
		{"perdido-1", "contatado", agora.Add(-4 * 24 * time.Hour), false},  // 4 dias sem resposta: perde
		{"abriu-1", "abriu", agora.Add(-4 * 24 * time.Hour), false},        // abriu o preview e não respondeu: perde
		{"recente-1", "contatado", agora.Add(-1 * 24 * time.Hour), false},  // só 1 dia: continua contatado
		{"respondeu-1", "contatado", agora.Add(-4 * 24 * time.Hour), true}, // 4 dias mas respondeu: continua contatado
	}
	for _, c := range casos {
		if err := b.Gravar(ctx, []lead.Lead{leadTeste(c.placeID)}, nil); err != nil {
			t.Fatal(err)
		}
		var respondeuEm any
		if c.comResposta {
			respondeuEm = agora
		}
		if _, err := b.pool.Exec(ctx,
			"update leads set etapa = $4, follow_up_em = $2, respondeu_em = $3 where place_id = $1",
			c.placeID, c.followUp, respondeuEm, c.etapa); err != nil {
			t.Fatal(err)
		}
	}

	n, err := b.MarcarPerdidos(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("MarcarPerdidos = %d, want 2", n)
	}

	linhas, err := b.pool.Query(ctx, "select place_id, etapa || coalesce('<-' || saiu_de, '') from leads where place_id = any($1)",
		[]string{"perdido-1", "abriu-1", "recente-1", "respondeu-1"})
	if err != nil {
		t.Fatal(err)
	}
	defer linhas.Close()
	etapas := map[string]string{}
	for linhas.Next() {
		var id, etapa string
		if err := linhas.Scan(&id, &etapa); err != nil {
			t.Fatal(err)
		}
		etapas[id] = etapa
	}
	want := map[string]string{"perdido-1": "perdido<-contatado", "abriu-1": "perdido<-abriu", "recente-1": "contatado", "respondeu-1": "contatado"}
	for id, w := range want {
		if etapas[id] != w {
			t.Errorf("etapa[%s] = %q, want %q", id, etapas[id], w)
		}
	}
}

func TestSomarCustoSoma(t *testing.T) {
	b := abrirBancoTeste(t)
	ctx := context.Background()
	dia := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)

	if err := b.SomarCusto(ctx, dia, "openai", 100, 0.01); err != nil {
		t.Fatal(err)
	}
	if err := b.SomarCusto(ctx, dia, "openai", 50, 0.02); err != nil {
		t.Fatal(err)
	}

	var unidades int64
	var usd float64
	if err := b.pool.QueryRow(ctx, "select unidades, usd from custos where dia = $1 and api = $2", dia, "openai").Scan(&unidades, &usd); err != nil {
		t.Fatal(err)
	}
	if unidades != 150 {
		t.Errorf("unidades = %d, want 150", unidades)
	}
	if usd < 0.0299 || usd > 0.0301 {
		t.Errorf("usd = %f, want ~0.03", usd)
	}
}
