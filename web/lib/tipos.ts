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
