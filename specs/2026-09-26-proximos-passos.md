# Próximos passos

Atualizado em 2026-09-27. A fase 1 (coleta no terminal) está na `main`, com CI verde, e a fase 2 (banco e painel) está pronta na branch `fase-2`. Este arquivo junta o que falta fazer e o que precisa estar pronto antes de cada passo. O detalhe de cada tarefa está no plano, `specs/2026-09-26-lancer.plano.md`, e os critérios de aceite na spec, `specs/2026-09-26-lancer.md`.

## Fase 2 — banco e painel (pronta na branch `fase-2`)

- **Banco:** database `lancer` no projeto Neon `neon-bronze-leaf`. A org Neon do Daniel é gerenciada pela Vercel e recusa projeto criado pelo CLI, por isso não houve projeto novo. O schema mora em `internal/banco/schema.sql` e a própria coleta o aplica.
- **Dedupe entre dias:** a coleta pula quem já foi visto antes de abrir o Maps. A 2ª rodada em Manaíra mostrou `já vistos: 8` e não reabriu nenhum lugar da 1ª.
- **Painel** (`web/`): fila com o botão wa.me, `/cobrar`, `/leads`, ficha com as pílulas de etapa e `/metricas`. Por enquanto roda só local (`cd web && npm run dev`); o deploy é da fase 3.
- **Falta:** push da `fase-2` e merge na `main`, com ok do Daniel.

### Antes de mandar a primeira mensagem

- A variante `link` ainda leva o texto literal `<preview>/p/<slug>`, porque o preview é da fase 3 (T9). Até lá, mande só lead da variante `texto`, ou apague a última frase antes de enviar.
- Até a escuta existir (fase 3), o funil anda pelas pílulas da ficha e pelo botão "follow-up enviado" do `/cobrar`.

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
  - slug com "--" quando o `place_id` tem hífen nos oito últimos caracteres (o link funciona);
  - falta teste do `Pontuar` com várias dores e teste de URL sem esquema;
  - `hostDe` engole o erro do `url.Parse`;
  - "lento no celular" trunca a nota em vez de arredondar;
  - `classificarESitear` sempre devolve erro `nil`;
  - a coleta de avaliações não para depois de juntar 10 negativas;
  - os testes mudam variáveis globais do pacote.
- **Adiados na fase 2:**
  - o custo do dia usa o fuso local (a VM vai rodar em `America/Fortaleza`);
  - se `SomarCusto` falhar depois do `Gravar`, o custo daquela rodada se perde (fração de centavo);
  - o funil só sabe de onde o lead saiu (`saiu_de`), não quando passou por cada degrau.
- **CI:**
  - `ubuntu-latest` passa a ser Ubuntu 26 em 19/10/2026: conferir se o passo do AppArmor para o sandbox do Chrome continua valendo.
