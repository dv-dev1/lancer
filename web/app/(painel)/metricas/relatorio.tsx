import { amostraPequena, type Corte } from '@/lib/metricas.ts'
import type { Degrau } from '@/lib/rotulos.ts'
import { Etiqueta } from '../ui.tsx'

const BLOCO = 'rounded-md border border-linha bg-superficie p-5'

function Amostra({ n }: { n: number }) {
  const pequena = amostraPequena(n)
  return (
    <span className={`text-xs ${pequena ? 'text-aviso' : 'text-apagado'}`}>
      n={n}
      {pequena ? ' · amostra pequena' : ''}
    </span>
  )
}

export function FaixaNumeros({ itens }: { itens: { rotulo: string; valor: string; n?: number }[] }) {
  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
      {itens.map((it) => (
        <div key={it.rotulo} className={`${BLOCO} space-y-1`}>
          <p className="text-apagado text-xs uppercase tracking-widest">{it.rotulo}</p>
          <p className="font-mono text-2xl text-texto tabular-nums">{it.valor}</p>
          {it.n !== undefined && <Amostra n={it.n} />}
        </div>
      ))}
    </div>
  )
}

export function Funil({ degraus, saidas }: { degraus: Degrau[]; saidas: number }) {
  const topo = Math.max(1, degraus[0]?.n ?? 1)
  return (
    <section className={`${BLOCO} space-y-3`}>
      <Etiqueta>Funil</Etiqueta>
      <p className="text-apagado text-xs">Leads que chegaram até cada etapa · % sobre a etapa anterior</p>
      {degraus.map((d) => (
        <div key={d.rotulo} className="grid grid-cols-[7rem_1fr_4.5rem] items-center gap-x-3 text-sm">
          <span className="text-suave">{d.rotulo}</span>
          <div className="h-2 rounded-sm bg-linha">
            <div className="h-full rounded-sm bg-acento" style={{ width: `${Math.max(2, (d.n / topo) * 100)}%` }} />
          </div>
          <span className="text-right font-mono tabular-nums">
            <span className="text-texto">{d.n}</span> <span className="text-apagado text-xs">{d.conversao}%</span>
          </span>
        </div>
      ))}
      {saidas > 0 && <p className="text-apagado text-xs">Perdidos ou saíram: {saidas}</p>}
    </section>
  )
}

export function TabelaCorte({
  titulo,
  cortes,
  rotulos,
}: {
  titulo: string
  cortes: Corte[]
  rotulos?: Record<string, string>
}) {
  return (
    <section className={`${BLOCO} space-y-2`}>
      <Etiqueta>{titulo}</Etiqueta>
      {cortes.length === 0 && <p className="text-sm text-suave">Sem leads contatados ainda.</p>}
      {cortes.map((c) => (
        <div key={c.chave} className="flex items-center justify-between text-sm">
          <span className="text-suave">{rotulos?.[c.chave] ?? c.chave}</span>
          <span className="font-mono text-texto tabular-nums">
            {c.taxa} <Amostra n={c.n} />
          </span>
        </div>
      ))}
    </section>
  )
}
