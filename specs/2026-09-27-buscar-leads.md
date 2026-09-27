# buscar leads pelo painel

## Intenção

1. Uma seção nova no painel, **Buscar**, com um formulário: cidade da Paraíba (João Pessoa por padrão, com sugestões das maiores), bairro opcional, nicho e limite (1 a 20). Enviar grava um pedido no banco; o painel não roda a coleta.
2. Um coletor de plantão, `go run ./cmd/coleta --servir`, pega os pedidos pendentes um por vez, roda a mesma coleta do terminal e marca o pedido como pronto (com quantos leads novos) ou com erro. No Mac agora, na VM na fase 3.
3. A lista de pedidos mostra o estado de cada um (esperando, buscando, pronto, erro) e se atualiza sozinha enquanto houver pedido em andamento. Pronto leva aos leads daquele pedido (`/leads?pedido=<id>`); os novos também entram na Fila.
4. A coleta aceita cidade: `--cidade` (padrão "João Pessoa"), e `--bairro` passa a ser opcional. A busca no Maps vira "confeitaria em Manaíra, Campina Grande - PB", ou "confeitaria em Campina Grande - PB" sem bairro. O lead guarda a cidade; sem bairro, a cidade vai também no campo bairro, para o painel continuar mostrando o lugar.
5. Captcha continua parando a coleta do dia: o pedido vira erro com a mensagem de captcha, os outros pendentes também, e o coletor sai.

## Critério de aceite

```bash
cd ~/lancer && go vet ./... && go test ./...
# ok em todos (os testes de banco rodam contra o Neon com DATABASE_URL)

cd ~/lancer/web && npm test && npm run lint && npx tsc --noEmit && npm run build
# verde
```

Ponta a ponta, com o painel no `npm run dev` e o coletor no `go run ./cmd/coleta --servir`:
- pedir "confeitaria, Campina Grande, sem bairro, limite 2" em `/buscar`: o pedido passa de esperando para buscando e para pronto sem recarregar a página;
- o link do pedido pronto abre `/leads?pedido=<id>` só com os leads dele, com bairro "Campina Grande";
- pedido com nicho fora da lista ou limite 0 é recusado no servidor.

## Fora de escopo

- Cidade fora da Paraíba.
- Cancelar pedido em andamento, ou vários coletores em paralelo.
- Timer diário da VM (fase 3) e o rodízio nicho × bairro.
- Mudar a mensagem ("desenvolvedor aqui de João Pessoa") conforme a cidade.
- Mostrar se o coletor está vivo: pedido esperando mostra o comando para subir o coletor.
