import { ETAPAS } from './tipos.ts'

export const ETAPA: Record<string, string> = {
  na_fila: 'Na fila',
  contatado: 'Contatado',
  abriu: 'Abriu o link',
  respondeu: 'Respondeu',
  interessado: 'Interessado',
  proposta: 'Proposta',
  fechado: 'Fechado',
  perdido: 'Perdido',
  saiu: 'Saiu',
}

export const DOR: Record<string, string> = {
  sem_site: 'Sem site',
  site_agregador: 'Site é rede social ou agregador',
  site_ruim: 'Site não abre',
  reclamacao: 'Reclamação nas avaliações',
}

export const VARIANTE: Record<string, string> = { texto: 'Texto', link: 'Link' }

export type Degrau = { rotulo: string; n: number; conversao: number }

// perdido e saiu não têm degrau próprio: quem saiu conta até o degrau mais fundo que alcançou (saidasPorDegrau, que a
// query acha pelas datas). Somar tudo em na_fila subcontava contatado, porque todo perdido vem de contatado.
export function funil(
  porEtapa: Record<string, number>,
  saidasPorDegrau: Record<string, number> = {},
): { degraus: Degrau[]; saidas: number } {
  const escada = ETAPAS.filter((e) => e !== 'perdido' && e !== 'saiu')
  let acumulado = 0
  const n: number[] = Array(escada.length).fill(0)
  for (let i = escada.length - 1; i >= 0; i--) {
    acumulado += (porEtapa[escada[i]] ?? 0) + (saidasPorDegrau[escada[i]] ?? 0)
    n[i] = acumulado
  }
  const saidas = Object.values(saidasPorDegrau).reduce((a, b) => a + b, 0)
  const degraus = escada.map((e, i) => {
    const base = n[Math.max(0, i - 1)]
    return { rotulo: ETAPA[e], n: n[i], conversao: base ? Math.round((n[i] / base) * 100) : 0 }
  })
  return { degraus, saidas }
}
