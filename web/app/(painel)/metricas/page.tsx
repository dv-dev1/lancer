import type { Metadata } from 'next'
import { sql } from '@/lib/db.ts'
import { custoPorFechado, type LeadParaMetrica, porDor, porNicho, porVariante, totais } from '@/lib/metricas.ts'
import { DOR, funil, VARIANTE } from '@/lib/rotulos.ts'
import { FaixaNumeros, Funil, TabelaCorte } from './relatorio.tsx'

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
  const t = totais(leads)

  return (
    <div className="max-w-[48rem] space-y-4 p-4 lg:p-10">
      <FaixaNumeros
        itens={[
          { rotulo: 'Contatados', valor: String(t.contatados) },
          { rotulo: 'Respondidos', valor: String(t.respondidos) },
          { rotulo: 'Taxa de resposta', valor: t.taxa, n: t.contatados },
          { rotulo: 'Custo por fechado', valor: custoPorFechado(custo.total, fechados) },
        ]}
      />
      <Funil degraus={f.degraus} saidas={f.saidas} />
      <div className="grid gap-4 lg:grid-cols-3">
        <TabelaCorte titulo="Por variante" cortes={porVariante(leads)} rotulos={VARIANTE} />
        <TabelaCorte titulo="Por nicho" cortes={porNicho(leads)} />
        <TabelaCorte titulo="Por dor" cortes={porDor(leads)} rotulos={DOR} />
      </div>
    </div>
  )
}
