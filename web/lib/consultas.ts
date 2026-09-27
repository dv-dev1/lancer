import { deveCobrar } from './cobrar.ts'
import { sql } from './db.ts'

export type LeadFoco = {
  // bigint vem do driver como string; o id só é usado como texto (href, input hidden), nunca em conta.
  id: string
  nome: string
  nicho: string
  bairro: string
  telefone: string
  nota: string | number
  avaliacoes: number | null
  pontuacao: number
  dores: string[]
  detalhes: Record<string, string>
  variante: string
  mensagem: string
}

export type LeadCobranca = LeadFoco & {
  contatado_em: string | Date
  respondeu_em: string | Date | null
  follow_up_em: string | Date | null
}

export async function leadsDaFila(): Promise<LeadFoco[]> {
  const linhas = await sql()`select id, nome, nicho, bairro, telefone, nota, avaliacoes, pontuacao, dores, detalhes,
      variante, mensagem
    from leads where etapa = 'na_fila' order by pontuacao desc, id`
  return linhas.map((l) => ({ ...l, id: String(l.id) }) as LeadFoco)
}

// A régua dos 3 dias é o deveCobrar, não o SQL: a query só traz quem ainda pode entrar nela.
export async function candidatosCobranca(): Promise<LeadCobranca[]> {
  const linhas = await sql()`select id, nome, nicho, bairro, telefone, nota, avaliacoes, pontuacao, dores, detalhes,
      variante, mensagem, contatado_em, respondeu_em, follow_up_em
    from leads
    where etapa in ('contatado', 'abriu') and contatado_em is not null and respondeu_em is null
      and follow_up_em is null
    order by contatado_em, id`
  return linhas.map((l) => ({ ...l, id: String(l.id) }) as LeadCobranca)
}

export async function tamanhoDaFila(): Promise<number> {
  const [{ n }] = await sql()`select count(*)::int as n from leads where etapa = 'na_fila'`
  return n as number
}

export async function contagens(agora = new Date()): Promise<{ fila: number; cobrar: number }> {
  const [fila, candidatos] = await Promise.all([tamanhoDaFila(), candidatosCobranca()])
  return { fila, cobrar: candidatos.filter((l) => deveCobrar(l, agora)).length }
}
