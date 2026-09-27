// Mesmas chaves do termosPorNicho do cmd/coleta: nicho fora daqui o coletor recusa.
export const NICHO: Record<string, string> = {
  confeitaria: 'Confeitaria',
  restaurante: 'Restaurante',
  loja: 'Loja',
  servico: 'Serviço',
}

// Só sugestão no <datalist>: qualquer cidade da Paraíba passa na validação.
export const CIDADES_PB = [
  'João Pessoa',
  'Campina Grande',
  'Santa Rita',
  'Patos',
  'Bayeux',
  'Sousa',
  'Cabedelo',
  'Cajazeiras',
  'Guarabira',
  'Sapé',
  'Mamanguape',
  'Queimadas',
  'Conde',
  'Monteiro',
  'Pombal',
  'Esperança',
]

export const ESTADO_PEDIDO: Record<string, string> = {
  pendente: 'Esperando o coletor',
  rodando: 'Buscando',
  pronto: 'Pronto',
  erro: 'Erro',
}

export type NovoPedido = { cidade: string; bairro: string; nicho: string; limite: number }

// O texto vai para a busca do Maps: só letra, número, espaço e pontuação de nome de lugar.
const NOME_DE_LUGAR = /^[\p{L}\d .,'-]{2,60}$/u

export function validarPedido(form: FormData): { pedido: NovoPedido } | { erro: string } {
  const cidade = String(form.get('cidade') ?? '').trim()
  const bairro = String(form.get('bairro') ?? '').trim()
  const nicho = String(form.get('nicho') ?? '')
  const limite = Number(form.get('limite'))
  if (!NOME_DE_LUGAR.test(cidade)) return { erro: 'Cidade inválida.' }
  if (bairro && !NOME_DE_LUGAR.test(bairro)) return { erro: 'Bairro inválido.' }
  if (!Object.hasOwn(NICHO, nicho)) return { erro: 'Nicho inválido.' }
  if (!Number.isInteger(limite) || limite < 1 || limite > 20) return { erro: 'O limite vai de 1 a 20.' }
  return { pedido: { cidade, bairro, nicho, limite } }
}

export const lugarDoPedido = (cidade: string, bairro: string) => (bairro ? `${bairro}, ${cidade}` : cidade)
