export type Vizinhos = {
  atual: string
  posicao: number
  total: number
  anterior: string | null
  proximo: string | null
}

// `atual` vem cru da URL: ausente, inválido ou repetido (?lead=1&lead=2) cai no primeiro da lista.
export function vizinhos(ids: readonly string[], atual: unknown): Vizinhos | null {
  if (ids.length === 0) return null
  const i = Math.max(0, typeof atual === 'string' ? ids.indexOf(atual) : -1)
  return {
    atual: ids[i],
    posicao: i + 1,
    total: ids.length,
    anterior: ids[i - 1] ?? null,
    proximo: ids[i + 1] ?? null,
  }
}

export function seguinte(ids: readonly string[], atual: unknown): string | null {
  const v = vizinhos(ids, atual)
  return v ? (v.proximo ?? v.anterior) : null
}
