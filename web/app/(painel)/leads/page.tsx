import type { Metadata } from 'next'
import Link from 'next/link'
import { sql } from '@/lib/db.ts'
import { DOR, ETAPA, VARIANTE } from '@/lib/rotulos.ts'
import { Cabecalho, Selo, SeloEtapa } from '../ui.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Leads' }

type Filtros = { nicho?: string; etapa?: string; variante?: string; dor?: string }

export default async function Leads({ searchParams }: { searchParams: Promise<Filtros> }) {
  const f = await searchParams
  const nicho = f.nicho || null
  const etapa = f.etapa || null
  const variante = f.variante || null
  const dor = f.dor || null
  const [leads, nichos] = await Promise.all([
    sql()`select id, nome, nicho, bairro, etapa, variante, dores, pontuacao from leads
      where (${nicho}::text is null or nicho = ${nicho})
        and (${etapa}::text is null or etapa = ${etapa})
        and (${variante}::text is null or variante = ${variante})
        and (${dor}::text is null or ${dor} = any(dores))
      order by pontuacao desc limit 200`,
    sql()`select distinct nicho from leads order by nicho`,
  ])
  const opcoesNicho = Object.fromEntries(nichos.map((n) => [n.nicho as string, n.nicho as string]))

  return (
    <div className="space-y-6">
      <Cabecalho
        titulo="Leads"
        descricao={`${leads.length} ${leads.length === 1 ? 'lead' : 'leads'} · ordenados pela pontuação`}
      />
      <nav className="cartao surgir space-y-2.5 p-4">
        <Chips grupo="Nicho" chave="nicho" opcoes={opcoesNicho} f={f} />
        <Chips grupo="Etapa" chave="etapa" opcoes={ETAPA} f={f} />
        <Chips grupo="Variante" chave="variante" opcoes={VARIANTE} f={f} />
        <Chips grupo="Dor" chave="dor" opcoes={DOR} f={f} />
      </nav>
      <div className="cartao surgir overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-borda text-xs uppercase tracking-widest text-suave">
            <tr>
              <th className="px-4 py-3 font-normal">Lead</th>
              <th className="font-normal">Pontuação</th>
              <th className="font-normal">Variante</th>
              <th className="font-normal">Etapa</th>
              <th className="hidden pr-4 font-normal md:table-cell">Dores</th>
            </tr>
          </thead>
          <tbody>
            {leads.map((l) => (
              <tr
                key={l.id}
                className="relative border-b border-borda transition-colors duration-150 last:border-0 hover:bg-superficie"
              >
                <td className="px-4 py-3">
                  {/* O after cobre a linha inteira: clicar em qualquer ponto abre a ficha. */}
                  <Link
                    className="font-medium text-white after:absolute after:inset-0 hover:text-ciano"
                    href={`/leads/${encodeURIComponent(l.id)}`}
                  >
                    {l.nome}
                  </Link>
                  <div className="text-suave">
                    {l.nicho} · {l.bairro}
                  </div>
                </td>
                <td className="tabular-nums">{l.pontuacao}</td>
                <td>
                  <Selo>{VARIANTE[l.variante] ?? l.variante}</Selo>
                </td>
                <td>
                  <SeloEtapa etapa={l.etapa} rotulo={ETAPA[l.etapa] ?? l.etapa} />
                </td>
                <td className="hidden pr-4 md:table-cell">
                  {(l.dores as string[]).map((d) => DOR[d] ?? d).join(', ') || '—'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {leads.length === 0 && <p className="py-10 text-center text-suave">Nenhum lead com esses filtros.</p>}
      </div>
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
  const href = (valor: string) => {
    const q = new URLSearchParams(Object.entries({ ...f, [chave]: valor }).filter(([, v]) => v) as [string, string][])
    return q.size ? `/leads?${q}` : '/leads'
  }
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <span className="w-20 text-xs uppercase tracking-widest text-suave">{grupo}</span>
      {[['', 'Todas'], ...Object.entries(opcoes)].map(([valor, rotulo]) => (
        <Link
          key={valor}
          href={href(valor)}
          className={`rounded-full border px-3 py-1 text-xs transition-colors duration-150 ${
            (f[chave] ?? '') === valor
              ? 'botao-primario border-transparent'
              : 'border-borda text-suave hover:border-ciano/40 hover:text-white'
          }`}
        >
          {rotulo}
        </Link>
      ))}
    </div>
  )
}
