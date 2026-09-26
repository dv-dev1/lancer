package maps

import "github.com/chromedp/cdproto/runtime"

// comAwaitPromise faz o Evaluate esperar a Promise resolver: usado nos scripts que rolam e
// esperam a lista carregar antes de contar os itens de novo.
func comAwaitPromise(p *runtime.EvaluateParams) *runtime.EvaluateParams {
	return p.WithAwaitPromise(true)
}

// seletores usados direto pelo chromedp (WaitVisible/Click), fora dos scripts JS: mesmo motivo de morar aqui.
const selFeed = `div[role="feed"]`
const selH1 = `h1`
const selAbaAvaliacoes = `button[role="tab"][aria-label^="Avaliações"]`

const jsTextoDoBody = `document.body ? document.body.innerText : ""`

const jsExtrairResultados = `
Array.from(document.querySelectorAll('div[role="feed"] a[href*="/maps/place/"]')).map(function(a){
  return {href: a.getAttribute("href") || "", nome: a.getAttribute("aria-label") || ""};
})
`

// rola o feed até o fim e espera %dms — o Google injeta mais itens de forma assíncrona.
const jsRolarFeedTpl = `
(function(){
  return new Promise(function(resolve){
    var feed = document.querySelector('div[role="feed"]');
    if (!feed) { resolve(0); return; }
    feed.scrollTop = feed.scrollHeight;
    setTimeout(function(){
      resolve(document.querySelectorAll('div[role="feed"] a[href*="/maps/place/"]').length);
    }, %d);
  });
})()
`

// sem contagem de avaliações aqui: signed-out o Google não mostra na visão geral, só no histograma da aba (jsContagemAvaliacoes).
const jsExtrairLugar = `
(function(){
  function attr(sel, nome){ var el = document.querySelector(sel); return el ? (el.getAttribute(nome) || "") : ""; }
  var h1 = document.querySelector("h1");
  return {
    nome: h1 ? h1.textContent : "",
    notaTexto: attr('[role="img"][aria-label*="estrela"]', "aria-label"),
    telefoneItem: attr('button[data-item-id^="phone:tel:"]', "data-item-id"),
    site: attr('a[data-item-id="authority"]', "href"),
    endereco: attr('button[data-item-id="address"]', "aria-label").replace(/^Endereço:\s*/, ""),
    fechado: document.body.innerText.indexOf("Fechado permanentemente") !== -1
  };
})()
`

// as 5 barras do histograma ("N estrelas, M avaliações"); somar as 5 (contagem() no Go) dá o total.
const jsContagemAvaliacoes = `
Array.from(document.querySelectorAll("[aria-label]")).map(function(e){ return e.getAttribute("aria-label") || ""; })
  .filter(function(l){ return /^\d+\s*estrelas?,/i.test(l); })
  .map(function(l){ return l.split(",")[1] || ""; })
`

// acha o ancestral com scroll do botão "Classificar avaliações": estável, ao contrário das classes.
const jsAchaPainelRolavel = `
function achaPainelRolavel(){
  var botao = document.querySelector('button[aria-label="Classificar avaliações"]');
  if (!botao) return null;
  var el = botao.parentElement;
  while (el) {
    var estilo = getComputedStyle(el);
    if (estilo.overflowY === "auto" || estilo.overflowY === "scroll") return el;
    el = el.parentElement;
  }
  return null;
}
`

const jsRolarAvaliacoesTpl = jsAchaPainelRolavel + `
(function(){
  return new Promise(function(resolve){
    var painel = achaPainelRolavel();
    if (!painel) { resolve(0); return; }
    painel.scrollTop = painel.scrollHeight;
    setTimeout(function(){ resolve(document.querySelectorAll("[data-review-id]").length); }, %d);
  });
})()
`

const jsExtrairAvaliacoes = `
Array.from(document.querySelectorAll("[data-review-id]")).map(function(el){
  var estrelasEl = el.querySelector('[aria-label*="estrela"]');
  var textoEl = el.querySelector("[lang]");
  var texto = textoEl ? (textoEl.textContent || "") : "";
  return {
    id: el.getAttribute("data-review-id") || "",
    estrelas: estrelasEl ? (estrelasEl.getAttribute("aria-label") || "") : "",
    texto: texto.replace(/…\n?Mais\s*$/, "").trim()
  };
})
`

// sobe até 4 níveis do rótulo "Resumo feito com o Gemini" até achar o parágrafo real (>80 chars): o rótulo é irmão do texto, não contêiner.
const jsResumoGemini = `
(function(){
  var alvo = null;
  var todos = document.querySelectorAll("*");
  for (var i = 0; i < todos.length; i++) {
    var el = todos[i];
    if (el.children.length === 0 && el.textContent.trim() === "Resumo feito com o Gemini") { alvo = el; break; }
  }
  if (!alvo) return "";
  var atual = alvo;
  for (var p = 0; p < 4 && atual.parentElement; p++) {
    atual = atual.parentElement;
    if (atual.innerText && atual.innerText.length > 80) break;
  }
  var texto = atual.innerText || "";
  var idx = texto.indexOf("Resumo feito com o Gemini");
  var resumo = idx >= 0 ? texto.slice(0, idx) : texto;
  return resumo.replace(/\+\d+\s*$/, "").trim();
})()
`
