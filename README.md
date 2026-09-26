# lancer

Ferramenta interna de prospecção de leads para João Pessoa - PB. Todo dia ela varre o Google Places por nicho × bairro, detecta as dores de cada estabelecimento (sem site, site que é rede social, site próprio ruim, reclamação nas avaliações) e devolve a fila do dia com mensagem pronta.

## De onde vem o nome

"Lancer" é o "free lance" original: o mercenário medieval que vendia a própria lança pra quem pagasse. A ferramenta faz o equivalente em prospecção — sai atrás de quem ainda não foi abordado.

## Exemplo de saída (fictício)

O comando abaixo roda a coleta de verdade. O bloco a seguir é **exemplo, com nomes e telefones trocados** — a primeira rodada real substitui isto aqui.

```
$ go run ./cmd/coleta --nicho confeitaria --bairro "Manaíra" --limite 2

pontos  nome                    dores       nota (n)   variante  telefone
65      Doceria Maria Doce      reclamacao  4.5 (100)  texto     5583999998888
46      Doce Sabor Presente     sem_site    4.0 (30)   link      5583988887777

[Doceria Maria Doce] doceria-maria-doce-a1b2
oi, aqui é o Daniel. vi que vocês demoram bastante pra responder no whatsapp. faz sentido pra vocês? se não fizer sentido, é só me avisar que não mando mais nada.

[Doce Sabor Presente] doce-sabor-presente-c3d4
oi, aqui é o Daniel. vi que ainda não tem site. faz sentido pra vocês? se não fizer sentido, é só me avisar que não mando mais nada. montei um exemplo rápido pra ver: <preview>/p/doce-sabor-presente-c3d4

descartes:
- Confeitaria Fechada: fechado
- Confeitaria Sem Celular: sem celular

conta: {IDs:6 Detalhes:4 Atmosfera:2 TokensEntrada:150 TokensSaida:50} — teto US$ 0.13 (cota grátis já considerada estourada)
```

## Variáveis de ambiente

| Variável | Obrigatória | Uso |
|---|---|---|
| `GOOGLE_MAPS_API_KEY` | sim | Places API (New) e PageSpeed Insights |
| `OPENAI_API_KEY` | sim | análise das avaliações (gpt-4.1-mini) |
| `LANCER_REMETENTE` | sim | nome usado na abertura da mensagem |
| `LANCER_PREVIEW_URL` | não | base do link de preview; vazio vira o placeholder `<preview>/p/<slug>` |

Faltando alguma obrigatória, o comando sai com erro nomeando qual falta.

## Rodar

```bash
go vet ./... && go test ./...
set -a && source .env && set +a && go run ./cmd/coleta --nicho confeitaria --bairro "Manaíra" --limite 5
```

`--nicho` aceita `restaurante`, `confeitaria`, `loja` ou `servico`. `--limite` é opcional, padrão 10.
