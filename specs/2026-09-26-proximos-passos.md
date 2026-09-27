# Próximos passos

Atualizado em 2026-09-26. A fase 1 (coleta no terminal) está na `main` (`942c0f3`), com CI verde. Este arquivo junta o que falta fazer e o que precisa estar pronto antes de cada passo. O detalhe de cada tarefa está no plano, `specs/2026-09-26-lancer.plano.md`, e os critérios de aceite na spec, `specs/2026-09-26-lancer.md`.

## Fase 2 — banco e painel

### O que vamos fazer

1. **Rever a decisão D1 antes do schema.** Ela manda guardar só o `place_id` e buscar nome, site e nota ao vivo. A regra veio dos termos da Places API, que saiu do projeto (decisão D2). Com o Maps no navegador, buscar ao vivo significa abrir o Maps de novo a cada consulta do painel, o que aumenta o risco de captcha. A recomendação é guardar nome e telefone junto com o que derivamos.
2. **T6 — banco e coleta persistente** (`db/schema.sql`, Go com `pgx/v5`):
   - tabelas `leads`, `vistos`, `mensagens`, `visitas` e `custos`, com a etapa do funil em `leads` (`na_fila`, `contatado`, `abriu`, `respondeu`, `interessado`, `proposta`, `fechado`, `perdido`, `saiu`);
   - a coleta pula o `place_id` que já está em `leads` ou em `vistos` **antes** de abrir a página no Maps, para não gastar navegação com quem já foi visto;
   - grava o lead e o descarte, e `--dry-run` só imprime;
   - no fim de cada rodada, quem passou 3 dias do follow-up sem responder vira `perdido`;
   - teste com Postgres real, que só roda quando existe `DATABASE_URL`.
3. **T7 — painel web** (`web/`, Next 16, partindo da base do painel da GeniAI: login, sessão, layout e `linkWhatsApp`):
   - `/`: fila do dia por pontuação, com o botão "Abrir no WhatsApp";
   - `/cobrar`: contatado há 3 dias ou mais, sem resposta e sem follow-up, com o texto pronto;
   - `/leads` e `/leads/[id]`: filtros por nicho, etapa, variante e dor; ficha com dores, mensagens e as etapas manuais;
   - `/metricas`: funil e taxa de resposta por variante, nicho e dor, sempre com o `n` ao lado.

### O que precisa vir do Daniel

| Item | Para quê | Como |
|---|---|---|
| `DATABASE_URL` | T6 e T7 | criar um projeto novo no Neon, plano Free (console.neon.tech → New project), copiar a connection string e colar em `~/lancer/.env` como `DATABASE_URL=` |
| Decisão D1 | schema da T6 | guardar nome e telefone (recomendado) ou só o `place_id` |
| Usuário e senha do painel | login do painel | escolher e colocar em `~/lancer/.env` como `DASHBOARD_USER=` e `DASHBOARD_PASSWORD=` |

O `AUTH_SECRET` do painel é gerado localmente, sem ação do Daniel.

### O que é revalidado antes de começar

- Cota do plano Free do Neon, e se o cadastro pede cartão (a regra é custo zero).
- Versão atual do `pgx/v5` e do Next 16 que a base da GeniAI usa.
- Instalar as dependências (`go get github.com/jackc/pgx/v5` e `npm install` em `web/`) só com ok, porque instalar pacote é ação de faixa vermelha.

### Aceite da fase 2

```bash
cd ~/lancer/web && npm test && npm run lint && npx tsc --noEmit && npm run build
# verde
cd ~/lancer && go run ./cmd/coleta --nicho confeitaria --bairro "Manaíra" --limite 5 && go run ./cmd/coleta --nicho confeitaria --bairro "Manaíra" --limite 5
# a segunda rodada não abre de novo nenhum place_id da primeira
```

Depois, o painel local (`npm run dev`) mostra a fila, e o botão wa.me abre a mensagem certa.

## Fase 3 — escuta, preview e deploy

### O que vamos fazer

1. **T8 — escuta** (`cmd/escuta`): whatsmeow pareado por QR, com a sessão em SQLite local. Só escuta, nunca envia. Mensagem de telefone que não é lead volta na hora, sem gravar nem logar. A primeira mensagem do Daniel marca `contatado`, a segunda grava o follow-up, e a resposta do lead marca `respondeu`. Contato `@lid` passa pelo mapeamento LID → telefone do whatsmeow.
2. **T9 — preview** (`web/app/p/[slug]`): quatro templates (cardápio, vitrine, catálogo, serviços) com os dados reais do lead e itens de exemplo marcados com selo, mais o `opengraph-image`. A visita é registrada, e o robô do WhatsApp que monta o card do link é ignorado.
3. **T10 — deploy**:
   - VM Oracle Always Free (Ubuntu ARM): binários `linux/arm64`, `lancer-escuta.service` com `Restart=always` e `lancer-coleta.timer` diário às 07:00 (`America/Fortaleza`), girando nicho × bairro dentro de um orçamento diário;
   - painel no Cloudflare `*.workers.dev` com `@opennextjs/cloudflare`, com `DATABASE_URL` e `AUTH_SECRET` em secrets.

### O que precisa vir do Daniel

| Item | Para quê | Observação |
|---|---|---|
| Celular com o WhatsApp pessoal | parear a escuta | escanear o QR que aparece no terminal |
| Conta Oracle Cloud com a VM ARM Always Free | rodar a coleta e a escuta | a Oracle costuma pedir cartão no cadastro só para verificar identidade; como a regra é custo zero, isso precisa de decisão antes da T10 |
| Conta Cloudflare | publicar o painel | plano grátis |
| Um segundo número de WhatsApp | teste ponta a ponta | faz o papel de lead |

### Riscos e revalidações antes

- **Captcha na VM.** IP de datacenter tende a receber captcha do Google antes do IP de casa. Antes de ligar o timer, rodar uma coleta manual na VM. Se o captcha vier logo, a coleta continua no Mac e só a escuta vai para a VM.
- OpenNext com Next 16.
- Driver SQLite do whatsmeow que compile para `linux/arm64` sem cgo.

### Aceite da fase 3

```bash
cd ~/lancer && go test ./...
curl -s -o /dev/null -w '%{http_code}\n' "https://<app>.workers.dev/p/<slug>"
# 200
```

Ponta a ponta, com o segundo número no papel de lead: enviar pelo wa.me do painel leva a `contatado`; responder leva a `respondeu`; abrir o preview registra a visita; o card do WhatsApp não conta como visita.

## Pendências herdadas da fase 1

- **Contagem de avaliações.** No layout do Maps sem login a contagem vem como "(?)", e o porte usa só a nota (R12). Rever quando houver dado de resposta suficiente.
- **Ajustes pequenos adiados:**
  - slug com "--" quando o `place_id` tem hífen nos quatro últimos caracteres (o link funciona);
  - falta teste do `Pontuar` com várias dores e teste de URL sem esquema;
  - `hostDe` engole o erro do `url.Parse`;
  - "lento no celular" trunca a nota em vez de arredondar;
  - `classificarESitear` sempre devolve erro `nil`;
  - a coleta de avaliações não para depois de juntar 10 negativas;
  - os testes mudam variáveis globais do pacote.
- **CI:**
  - as actions `checkout@v4` e `setup-go@v5` rodam em Node 20, que está obsoleto, e precisam subir de versão;
  - `ubuntu-latest` passa a ser Ubuntu 26 em 19/10/2026: conferir se o passo do AppArmor para o sandbox do Chrome continua valendo.
