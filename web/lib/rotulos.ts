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

// perdido e saiu são saída do funil, sem etapa seguinte. ponytail: os dois somam só no primeiro
// degrau — nem todo saiu passou por abriu/respondeu antes, então não dá pra repor mais fundo na escada.
export function funil(porEtapa: Record<string, number>): { degraus: Degrau[]; saidas: number } {
  const escada = ETAPAS.filter((e) => e !== 'perdido' && e !== 'saiu')
  let acumulado = 0
  const n: number[] = Array(escada.length).fill(0)
  for (let i = escada.length - 1; i >= 0; i--) {
    acumulado += porEtapa[escada[i]] ?? 0
    n[i] = acumulado
  }
  const saidas = (porEtapa.perdido ?? 0) + (porEtapa.saiu ?? 0)
  n[0] += saidas
  const degraus = escada.map((e, i) => {
    const base = n[Math.max(0, i - 1)]
    return { rotulo: ETAPA[e], n: n[i], conversao: base ? Math.round((n[i] / base) * 100) : 0 }
  })
  return { degraus, saidas }
}
