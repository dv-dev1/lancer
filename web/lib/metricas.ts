export type LeadParaMetrica = {
  variante: string
  nicho: string
  dores: string[]
  contatado_em: string | Date | null
  respondeu_em: string | Date | null
}

export type Corte = { chave: string; n: number; respostas: number; taxa: string }

export function taxa(respostas: number, n: number): string {
  return n === 0 ? '—' : `${Math.round((respostas / n) * 100)}%`
}

// Cada corte é uma pergunta separada sobre o mesmo lead: as dores não são exclusivas entre si.
function porCorte(leads: LeadParaMetrica[], chavesDe: (l: LeadParaMetrica) => string[]): Corte[] {
  const contagem = new Map<string, { n: number; respostas: number }>()
  for (const l of leads) {
    if (!l.contatado_em) continue
    for (const chave of chavesDe(l)) {
      const c = contagem.get(chave) ?? { n: 0, respostas: 0 }
      c.n++
      if (l.respondeu_em) c.respostas++
      contagem.set(chave, c)
    }
  }
  return [...contagem.entries()]
    .map(([chave, { n, respostas }]) => ({ chave, n, respostas, taxa: taxa(respostas, n) }))
    .sort((a, b) => b.n - a.n)
}

export const porVariante = (leads: LeadParaMetrica[]) => porCorte(leads, (l) => [l.variante])
export const porNicho = (leads: LeadParaMetrica[]) => porCorte(leads, (l) => [l.nicho])
export const porDor = (leads: LeadParaMetrica[]) => porCorte(leads, (l) => l.dores)

export function custoPorFechado(totalUsd: number, fechados: number): string {
  if (fechados === 0) return '—'
  return (totalUsd / fechados).toLocaleString('pt-BR', { style: 'currency', currency: 'USD', minimumFractionDigits: 4 })
}
