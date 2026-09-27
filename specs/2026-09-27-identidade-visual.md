# identidade visual do painel

## Intenção

1. O painel deixa de ser o esqueleto do GeniAI com outra cor. Some o menu no topo, a página centralizada, o título grande com subtítulo, o cartão com tabela dentro e o botão pílula. O formato passa a ser uma **Mesa de foco**: um lead por vez, como triagem de caixa de entrada.
2. A identidade é dark, moderna e minimalista. Grafite quente com um acento âmbar, e a cor sempre significa algo: âmbar é ação e avanço, verde é resposta, coral é aviso. Nada de gradiente, brilho, vidro ou símbolo; a marca é só a palavra `lancer`.
3. O trabalho do dia fica no teclado:
   - na Fila e no Cobrar, o lead atual mostra por que abordar e a mensagem, que dá para editar antes de sair;
   - `↵` abre o WhatsApp, `E` marca enviado e passa ao próximo, `←`/`→` andam pela fila;
   - cada botão mostra a própria tecla.
4. No celular é um lead por tela, como um baralho, com as ações fixas no polegar e a navegação numa barra inferior; nada rola para o lado.
5. Leads vira um quadro do funil (uma coluna por etapa); a ficha usa a mesma mesa, com a etapa como caminho de 7 degraus e as saídas à parte; Métricas vira um relatório que avisa quando o `n` é pequeno demais.

## Direção

- **Tese:** o painel é a mesa do prospector: um lance por vez, com o porquê e a mensagem na frente e a ação na mão (teclado ou polegar). Recusa o dashboard de cartões num container centralizado.
- **Mundo:** app de tela cheia, sem container centralizado. Superfícies separadas por tom e linha de 1 px, sem sombra, vidro nem gradiente. Um acento âmbar. Botões retangulares (canto de 4 a 6 px, nunca pílula) com a tecla ao lado; rótulos de seção em caixa alta pequena; números em mono.
- **Paleta** (contraste medido sobre o fundo):

  | papel | cor | uso |
  |---|---|---|
  | fundo | `#0f0f0e` | página |
  | superfície | `#171716` | blocos, linhas da fila |
  | elevada | `#1f1f1d` | hover, campo, menu aberto |
  | linha | `#2c2b28` | divisória |
  | linha forte | `#3d3c38` | borda de campo e de botão secundário |
  | texto | `#ecebe7` | 16,1:1 |
  | suave | `#a3a19a` | texto secundário, 7,4:1 |
  | apagado | `#85837c` | metadado, saídas do funil |
  | acento | `#f2b45c` | ação primária, item ativo, degrau alcançado, 10,5:1 |
  | positivo | `#8fd1a1` | respondeu, fechado, taxa |
  | aviso | `#f08a6c` | variante link sem preview, amostra pequena, erro |

- **Tipo:** Geist no texto e Geist Mono em número, telefone, pontuação, contagem e tecla, as duas grátis pelo `next/font/google`. O nome do lead é o maior texto da tela; não há título de página.
- **Movimento:** só transição de cor (150 ms). Respeita `prefers-reduced-motion`.

## UX por tela

- **Casca:**
  - no computador (≥ 1024 px), tela cheia em grade. Um trilho vertical à esquerda leva a marca `lancer`, as seções Fila, Cobrar, Leads e Métricas (com as contagens de Fila e Cobrar) e "Sair" no pé;
  - no celular, o trilho vira barra inferior, e marca e Sair vão para uma barra fina no topo.
- **Fila (`/?lead=<id>`):**
  - no computador, coluna estreita com a fila (o atual marcado no âmbar) e o painel de foco. O painel traz o nome, a linha de dados (nicho · bairro · nota (n) · pontuação), POR QUE (dores com o detalhe) e MENSAGEM (textarea editável; a edição vale só para este envio);
  - ações `↵ Abrir WhatsApp`, `E Enviei` e `←`/`→`, com a posição ("2 de 10");
  - "Enviei" reusa a gravação de etapa (→ `contatado`) e abre o próximo lead;
  - a variante `link` mostra aviso em coral enquanto o preview não existe (fase 3);
  - no celular: cabeçalho "2 de 10" com ‹ ›, e as duas ações fixas embaixo.
- **Cobrar (`/cobrar?lead=<id>`):**
  - a mesma mesa, com "há N dias" desde o contato, o texto do follow-up e `F Follow-up enviado`;
  - Cobrar vazio diz quando é a próxima cobrança; Fila vazia leva ao Cobrar quando há pendentes.
- **Leads (`/leads`):**
  - quadro com uma coluna por etapa (na fila → fechado; a contagem real no topo de cada coluna) e as saídas (perdido, saiu) recolhidas;
  - cartões compactos com pontuação e dores; filtros de nicho, variante e dor numa barra;
  - no celular, uma aba por etapa.
- **Ficha (`/leads/<id>`):**
  - o mesmo painel de foco, mais a etapa como caminho de 7 degraus (o atual no âmbar) com Perdido e Saiu à parte, e a conversa em linha do tempo;
  - voltar para "Na fila" funciona como desfazer: limpa as datas de contato, follow-up e resposta e o `saiu_de`.
- **Métricas:** relatório numa coluna:
  - faixa de números grandes (contatados, respondidos, taxa, custo por fechado) com o `n`;
  - funil de uma cor;
  - taxas por variante, nicho e dor, com "amostra pequena" em coral quando `n < 10`.
- **Login:** tela cheia, a palavra `lancer`, campos só com linha inferior e botão âmbar retangular com `↵`.

## Critério de aceite

```bash
cd ~/lancer/web && ! grep -rnE '#010f17|#0099ff|#00c0fa|#0daadf|#08fbd0|gradiente-marca|titulo-gradiente|botao-primario|rounded-full|font-inter|Inter\(|max-w-6xl|Cabecalho' app lib proxy.ts
# sem saída: nenhum resto do esqueleto nem do visual do GeniAI

cd ~/lancer/web && node --test tests/tema.test.ts tests/mesa.test.ts tests/teclas.test.ts
# contraste ≥ 4.5:1 lido do globals.css, sem gradient/blur; vizinhos da fila; mapa de teclas

cd ~/lancer/web && npm test && npm run lint && npx tsc --noEmit && npm run build
# verde, com o build sem DATABASE_URL
```

No `npm run dev`, em 390 px e 1440 px, com sessão:
- `/`, `/cobrar`, `/leads`, `/leads/<id>`, `/metricas` e `/login` sem rolagem lateral (`scrollWidth` ≤ largura);
- em `/`: `↵` abre `https://wa.me/...` com o texto editado; digitar "e" dentro da mensagem não envia; `→` muda o lead e o "2 de N";
- `E` tira o lead da fila e a ficha mostra Contatado; voltar para "Na fila" pela ficha o devolve sem datas;
- capturas de cada tela nos dois tamanhos, mostradas ao Daniel.

## Fora de escopo

- O visual do preview `/p/<slug>` (fase 3): ele leva a cara do negócio do lead, não a do lancer.
- Tema claro e alternância de tema.
- Logo desenhado, favicon elaborado e ícone de app.
- Mudar regra de negócio ou query, fora o "Enviei" (reusa a gravação de etapa) e o desfazer ao voltar para "Na fila".
- Guardar a mensagem editada no banco: a edição vale só para o envio.
- Arrastar cartão no quadro; animação de entrada; gráfico interativo.
- O mockup `2026-09-27-identidade-visual.mockup.html` ficou superado por esta versão e não entra no commit.
