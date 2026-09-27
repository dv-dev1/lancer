# lancer

Ferramenta interna de prospecção de leads para João Pessoa - PB. Todo dia ela varre o Google Maps (pelo navegador, sem API paga) por nicho × bairro, detecta as dores de cada estabelecimento (sem site, site que é rede social, site próprio ruim, reclamação nas avaliações) e devolve a fila do dia com mensagem pronta.

Abrir o Google Maps num navegador automatizado foge dos termos de uso do Google — é a troca feita conscientemente pra não depender de API paga. Rodar com moderação (a `Pausa` entre requisições já tenta reduzir o risco) e parar se aparecer captcha (a própria coleta já para sozinha nesse caso).

## De onde vem o nome

"Lancer" é o "free lance" original: o mercenário medieval que vendia a própria lança pra quem pagasse. A ferramenta faz o equivalente em prospecção — sai atrás de quem ainda não foi abordado.

## Exemplo de saída (saída real, com nomes e telefones trocados)

O bloco abaixo é a saída de verdade da 3ª rodada de teste (bairro Tambaú), só com os nomes dos
estabelecimentos e os telefones trocados por fictícios — dores, pontos, variante, mensagens,
descartes e rodapé são exatamente como saíram.

```
$ go run ./cmd/coleta --nicho confeitaria --bairro "Tambaú" --limite 3

pontos  nome                       dores           nota (n)  variante  telefone
69      Doce Recanto - Tambaú      sem_site        4.6 (?)   texto     5583990001111
65      Padaria Bela Vista         site_agregador  4.5 (?)   link      5583990002222
65      Restaurante Sabor Caseiro  sem_site        4.5 (?)   texto     5583990003333

[Doce Recanto - Tambaú] doce-recanto-tambau-xVvM
oi, aqui é o Daniel, desenvolvedor aqui de João Pessoa. vi que vocês não têm site, o que dificulta encontrar informações completas sobre os produtos. faz sentido pra vocês? se não fizer sentido, é só me avisar que não mando mais nada.

[Padaria Bela Vista] padaria-bela-vista-PPvE
oi, aqui é o Daniel, desenvolvedor aqui de João Pessoa. vi que o cardápio está só no Instagram, e o cliente precisa abrir o app para ver. faz sentido pra vocês? se não fizer sentido, é só me avisar que não mando mais nada. montei um exemplo rápido pra ver: <preview>/p/padaria-bela-vista-PPvE

[Restaurante Sabor Caseiro] restaurante-sabor-caseiro-ejTg
oi, aqui é o Daniel, desenvolvedor aqui de João Pessoa. vi que vocês não têm site, e isso pode dificultar os clientes que buscam informações rápidas. faz sentido pra vocês? se não fizer sentido, é só me avisar que não mando mais nada.

descartes:
- Doce Trufa - Tambaú: sem celular
- Padaria Boa Vista Tambaú: sem celular
- Doce Mel Confeitaria - João Pessoa | Bolos e Doces |: sem dor
- Bela Vista Delicatessen & Panificadora: sem celular
- Padaria Praiana - Padaria e Mini-Mercado: sem celular
- Doce Amore: sem celular

Maps: 2 buscas, 9 lugares abertos (grátis)
conta: {TokensEntrada:1902 TokensSaida:99} — teto US$ 0.00
```

Se o Google barrar a coleta com captcha, o comando para a rodada inteira com "Google pediu captcha — coleta parada; tente amanhã" — não adianta insistir no mesmo dia.

## Variáveis de ambiente

| Variável | Obrigatória | Uso |
|---|---|---|
| `OPENAI_API_KEY` | sim | análise das avaliações (gpt-4.1-mini) |
| `LANCER_REMETENTE` | sim | nome usado na abertura da mensagem |
| `LANCER_PREVIEW_URL` | não | base do link de preview; vazio vira o placeholder `<preview>/p/<slug>` |
| `PAGESPEED_API_KEY` | não | PageSpeed Insights; ver abaixo |
| `DATABASE_URL` | não | Postgres (Neon); vazio, ou com `--dry-run`, a coleta não lê nem grava banco |

Faltando alguma obrigatória, o comando sai com erro nomeando qual falta.

`PAGESPEED_API_KEY` é grátis e não pede cartão nem billing (basta ativar a API no Google Cloud Console).
Sem ela, o PageSpeed costuma devolver 429 (limite de uso sem chave) e a checagem de lentidão simplesmente
não roda pro lead — não é tratada como site ruim nem derruba a coleta, só fica sem esse dado.

## Rodar

```bash
go vet ./... && go test ./...
set -a && source .env && set +a && go run ./cmd/coleta --nicho confeitaria --bairro "Manaíra" --limite 5
```

`--nicho` aceita `restaurante`, `confeitaria`, `loja` ou `servico`. `--limite` é opcional, padrão 10.

Precisa de Chrome ou Chromium instalado (o `internal/maps` abre um headless local); sem isso, os testes de ponta a ponta pulam com aviso.

## Painel (fase 2)

A coleta grava os leads no Neon; o painel (`web/`, Next) lê as mesmas tabelas e mostra a fila do dia,
a tela de cobrança, a lista de leads com filtros e as métricas de resposta por variante, nicho e dor.

Com banco, quem já foi aberto (virou lead ou foi descartado por motivo que não muda) é pulado antes de abrir o
Maps. Rodapé real de duas rodadas seguidas, mesmo nicho e bairro: os 8 lugares da 1ª (5 leads e 3 descartes)
não são abertos de novo na 2ª.

```
$ go run ./cmd/coleta --nicho confeitaria --bairro "Manaíra" --limite 5   # 1ª rodada
...
Maps: 2 buscas, 8 lugares abertos (grátis)
já vistos: 0

$ go run ./cmd/coleta --nicho confeitaria --bairro "Manaíra" --limite 5   # 2ª rodada
...
Maps: 2 buscas, 6 lugares abertos (grátis)
já vistos: 8
```

```bash
cd web && npm run dev
```

Variáveis (em `web/.env.local`, symlink para o `.env` da raiz): `DATABASE_URL`, `AUTH_SECRET`,
`DASHBOARD_USER`, `DASHBOARD_PASSWORD`.
