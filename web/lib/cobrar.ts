// R21: texto fixo do follow-up único (fora de escopo: mais de um follow-up).
export const TEXTO_FOLLOW_UP =
  'Oi, tudo bem? Passando só pra saber se você chegou a ver minha mensagem. Se fizer sentido conversar, é só responder por aqui.'

export type LeadParaCobranca = {
  contatado_em: string | Date | null
  respondeu_em: string | Date | null
  follow_up_em: string | Date | null
}

const TRES_DIAS_MS = 3 * 24 * 60 * 60 * 1000

export function deveCobrar(lead: LeadParaCobranca, agora: Date): boolean {
  if (!lead.contatado_em || lead.respondeu_em || lead.follow_up_em) return false
  return new Date(lead.contatado_em).getTime() <= agora.getTime() - TRES_DIAS_MS
}
