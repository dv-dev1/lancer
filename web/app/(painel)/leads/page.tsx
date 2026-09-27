import type { Metadata } from 'next'
import Link from 'next/link'
import { sql } from '@/lib/db.ts'
import { abaAtual } from '@/lib/quadro.ts'
import { DOR, ETAPA, VARIANTE } from '@/lib/rotulos.ts'
import { ESCADA } from '@/lib/tipos.ts'
import { botao } from '../ui.tsx'
import { type CartaoLead, type Coluna, Quadro, type Saida } from './quadro.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Leads' }

type Filtro = string | string[] | undefined
type Filtros = { nicho?: Filtro; variante?: Filtro; dor?: Filtro; aba?: Filtro }

const primeiro = (v: Filtro) => (Array.isArray(v) ? v[0] : v) || null

function hrefComFiltro(f: Filtros, patch: Partial<Record<keyof Filtros, string>>): string {
  const q = new URLSearchParams(Object.entries({ ...f, ...patch }).filter(([, v]) => v) as [string, string][])
  return q.size ? `/leads?${q}` : '/leads'
}

export default async function Leads({ searchParams }: { searchParams: Promise<Filtros> }) {
  const f = await searchParams
  const nicho = primeiro(f.nicho)
  const variante = primeiro(f.variante)
  const dor = primeiro(f.dor)
  const [porEtapaBruto, leads, nichos] = await Promise.all([
    sql()`select etapa, count(*)::int as n from leads group by etapa`,
    // O limit corta quantos cartões aparecem, não a contagem real de cada coluna (essa vem do group by acima).
    sql()`select id, nome, nicho, bairro, etapa, dores, pontuacao from leads
      where etapa not in ('perdido', 'saiu')
        and (${nicho}::text is null or nicho = ${nicho})
        and (${variante}::text is null or variante = ${variante})
        and (${dor}::text is null or ${dor} = any(dores))
      order by pontuacao desc limit 200`,
    sql()`select distinct nicho from leads order by nicho`,
  ])
  const porEtapa = Object.fromEntries(porEtapaBruto.map((e) => [e.etapa as string, e.n as number]))
  const opcoesNicho = Object.fromEntries(nichos.map((n) => [n.nicho as string, n.nicho as string]))
  const aba = abaAtual(primeiro(f.aba) ?? undefined, porEtapa)
  const colunas: Coluna[] = ESCADA.map((etapa) => ({
    etapa,
    rotulo: ETAPA[etapa],
    n: porEtapa[etapa] ?? 0,
    href: hrefComFiltro(f, { aba: etapa }),
    leads: (leads as CartaoLead[]).filter((l) => l.etapa === etapa),
  }))
  const saidas: Saida[] = ['perdido', 'saiu']
    .map((etapa) => ({ etapa, rotulo: ETAPA[etapa], n: porEtapa[etapa] ?? 0 }))
    .filter((s) => s.n > 0)

  return (
    <div className="space-y-4 p-4 lg:p-8">
      <nav className="flex flex-wrap items-center gap-2.5">
        <Chips grupo="Nicho" chave="nicho" opcoes={opcoesNicho} f={f} />
        <Chips grupo="Variante" chave="variante" opcoes={VARIANTE} f={f} />
        <Chips grupo="Dor" chave="dor" opcoes={DOR} f={f} />
      </nav>
      <Quadro colunas={colunas} aba={aba} saidas={saidas} />
    </div>
  )
}

function Chips({
  grupo,
  chave,
  opcoes,
  f,
}: {
  grupo: string
  chave: keyof Filtros
  opcoes: Record<string, string>
  f: Filtros
}) {
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <span className="w-20 text-apagado text-xs uppercase tracking-widest">{grupo}</span>
      {[['', 'Todas'], ...Object.entries(opcoes)].map(([valor, rotulo]) => {
        const ativo = (f[chave] ?? '') === valor
        return (
          <Link
            key={valor}
            href={hrefComFiltro(f, { [chave]: valor })}
            className={botao(ativo ? 'primario' : 'secundario', 'h-8 px-3 text-xs')}
          >
            {rotulo}
          </Link>
        )
      })}
    </div>
  )
}
