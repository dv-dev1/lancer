import type { Metadata } from 'next'
import { sql } from '@/lib/db.ts'
import { custoPorFechado, type LeadParaMetrica, porDor, porNicho, porVariante } from '@/lib/metricas.ts'
import { DOR, funil, VARIANTE } from '@/lib/rotulos.ts'
import { Cabecalho } from '../ui.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Métricas' }

export default async function Metricas() {
  const [etapas, saidas, leadsBrutos, [custo]] = await Promise.all([
    sql()`select etapa, count(*)::int as n from leads group by etapa`,
    sql()`select coalesce(saiu_de, 'na_fila') as degrau, count(*)::int as n
          from leads where etapa in ('perdido', 'saiu') group by 1`,
    sql()`select variante, nicho, dores, contatado_em, respondeu_em from leads`,
    sql()`select coalesce(sum(usd), 0)::float as total from custos`,
  ])
  const leads = leadsBrutos as LeadParaMetrica[]
  const porEtapa = Object.fromEntries(etapas.map((e) => [e.etapa as string, e.n as number]))
  const f = funil(porEtapa, Object.fromEntries(saidas.map((s) => [s.degrau as string, s.n as number])))
  const fechados = porEtapa.fechado ?? 0
  const topoFunil = Math.max(1, f.degraus[0]?.n ?? 1)

  return (
    <div className="space-y-8">
      <Cabecalho titulo="Métricas" />

      <section className="cartao surgir p-5">
        <h2 className="font-medium">Funil</h2>
        <p className="mb-5 text-xs text-suave">Leads que chegaram até cada etapa · % sobre a etapa anterior</p>
        {f.degraus.map((d) => (
          <div key={d.rotulo} className="mb-2 grid grid-cols-[8rem_1fr_4.5rem] items-center gap-x-3 text-sm">
            <span className="text-suave">{d.rotulo}</span>
            <div className="flex h-6 justify-center">
              <div
                className="gradiente-marca h-full rounded-md"
                style={{ width: `${Math.max(1.5, (d.n / topoFunil) * 100)}%` }}
              />
            </div>
            <span className="text-right tabular-nums">
              <span className="text-white">{d.n}</span>
              <span className="ml-1.5 text-xs text-suave">{d.conversao}%</span>
            </span>
          </div>
        ))}
        {f.saidas > 0 && <p className="mt-4 text-xs text-suave">Perdidos ou saíram: {f.saidas}</p>}
      </section>

      <div className="grid gap-4 lg:grid-cols-3">
        <TabelaCorte titulo="Por variante" cortes={porVariante(leads)} rotulos={VARIANTE} />
        <TabelaCorte titulo="Por nicho" cortes={porNicho(leads)} />
        <TabelaCorte titulo="Por dor" cortes={porDor(leads)} rotulos={DOR} />
      </div>

      <section className="cartao surgir p-5">
        <h2 className="mb-4 font-medium">Custo</h2>
        <div className="flex gap-8 text-sm">
          <p>
            Total{' '}
            <span className="ml-2 tabular-nums text-white">
              {custo.total.toLocaleString('pt-BR', { style: 'currency', currency: 'USD', minimumFractionDigits: 4 })}
            </span>
          </p>
          <p>
            Por fechado <span className="ml-2 tabular-nums text-white">{custoPorFechado(custo.total, fechados)}</span>
          </p>
        </div>
      </section>
    </div>
  )
}

function TabelaCorte({
  titulo,
  cortes,
  rotulos,
}: {
  titulo: string
  cortes: { chave: string; n: number; respostas: number; taxa: string }[]
  rotulos?: Record<string, string>
}) {
  return (
    <section className="cartao surgir p-5">
      <h2 className="mb-4 font-medium">{titulo}</h2>
      {cortes.length === 0 && <p className="text-sm text-suave">Sem leads contatados ainda.</p>}
      {cortes.map((c) => (
        <div key={c.chave} className="mb-2 flex items-center justify-between text-sm">
          <span className="text-suave">{rotulos?.[c.chave] ?? c.chave}</span>
          <span className="tabular-nums text-white">
            {c.taxa} <span className="text-xs text-suave">(n={c.n})</span>
          </span>
        </div>
      ))}
    </section>
  )
}
