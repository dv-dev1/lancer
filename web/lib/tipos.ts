export const ETAPAS = [
  'na_fila',
  'contatado',
  'abriu',
  'respondeu',
  'interessado',
  'proposta',
  'fechado',
  'perdido',
  'saiu',
] as const

export type Etapa = (typeof ETAPAS)[number]

export const ESCADA: readonly Etapa[] = ETAPAS.filter((e) => e !== 'perdido' && e !== 'saiu')

export function alcancou(etapa: string, degrau: Etapa): boolean {
  const i = ESCADA.indexOf(etapa as Etapa)
  return i >= 0 && i >= ESCADA.indexOf(degrau)
}
