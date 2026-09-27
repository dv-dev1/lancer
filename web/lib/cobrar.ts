// R21: texto fixo do follow-up único (fora de escopo: mais de um follow-up).
export const TEXTO_FOLLOW_UP =
  'Oi, tudo bem? Passando só pra saber se você chegou a ver minha mensagem. Se fizer sentido conversar, é só responder por aqui.'

export type LeadParaCobranca = {
  contatado_em: string | Date | null
  respondeu_em: string | Date | null
  follow_up_em: string | Date | null
}

const DIA_MS = 24 * 60 * 60 * 1000
const TRES_DIAS_MS = 3 * DIA_MS

export function deveCobrar(lead: LeadParaCobranca, agora: Date): boolean {
  if (!lead.contatado_em || lead.respondeu_em || lead.follow_up_em) return false
  return new Date(lead.contatado_em).getTime() <= agora.getTime() - TRES_DIAS_MS
}

export function rotuloDias(desde: string | Date, agora: Date): string {
  const dias = Math.floor((agora.getTime() - new Date(desde).getTime()) / DIA_MS)
  if (dias < 1) return 'hoje'
  return dias === 1 ? 'há 1 dia' : `há ${dias} dias`
}

export function proximaCobranca(candidatos: LeadParaCobranca[], agora: Date): Date | null {
  const quando = candidatos
    .filter((l) => l.contatado_em && !l.respondeu_em && !l.follow_up_em)
    .map((l) => new Date(l.contatado_em as string | Date).getTime() + TRES_DIAS_MS)
    .filter((t) => t > agora.getTime())
  return quando.length ? new Date(Math.min(...quando)) : null
}
