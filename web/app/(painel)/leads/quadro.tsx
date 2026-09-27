import Link from 'next/link'
import { DOR } from '@/lib/rotulos.ts'
import type { Etapa } from '@/lib/tipos.ts'
import { Etiqueta } from '../ui.tsx'

export type CartaoLead = {
  id: string
  nome: string
  nicho: string
  bairro: string
  etapa: Etapa
  dores: string[]
  pontuacao: number
}
export type Coluna = { etapa: Etapa; rotulo: string; n: number; href: string; leads: CartaoLead[] }
export type Saida = { etapa: string; rotulo: string; n: number }

export function Quadro({ colunas, aba, saidas }: { colunas: Coluna[]; aba: Etapa; saidas: Saida[] }) {
  return (
    <div className="space-y-4">
      <div role="tablist" aria-label="Etapa" className="flex gap-1 overflow-x-auto lg:hidden">
        {colunas.map((c) => (
          <Aba key={c.etapa} coluna={c} ativa={c.etapa === aba} />
        ))}
      </div>
      <div className="lg:grid lg:grid-cols-[repeat(7,minmax(11rem,1fr))] lg:gap-0 lg:overflow-x-auto lg:rounded-md lg:border lg:border-linha">
        {colunas.map((c) => (
          <ColunaBoard key={c.etapa} coluna={c} ativa={c.etapa === aba} />
        ))}
      </div>
      {saidas.length > 0 && <Saidas saidas={saidas} />}
    </div>
  )
}

function Aba({ coluna, ativa }: { coluna: Coluna; ativa: boolean }) {
  return (
    <Link
      href={coluna.href}
      role="tab"
      aria-selected={ativa}
      className={`shrink-0 rounded-md px-3 py-1.5 text-sm transition-colors duration-150 ${
        ativa ? 'bg-elevada text-acento' : 'text-suave hover:text-texto'
      }`}
    >
      {coluna.rotulo} <span className="font-mono text-xs">{coluna.n}</span>
    </Link>
  )
}

function ColunaBoard({ coluna, ativa }: { coluna: Coluna; ativa: boolean }) {
  return (
    <div
      className={`${ativa ? 'flex' : 'hidden lg:flex'} min-w-0 flex-col gap-3 border-linha border-r px-3 py-3 last:border-r-0 lg:h-full lg:overflow-y-auto`}
    >
      <div className="flex items-center justify-between">
        <Etiqueta>{coluna.rotulo}</Etiqueta>
        <span className="font-mono text-apagado text-xs">{coluna.n}</span>
      </div>
      <div className="space-y-2">
        {coluna.leads.map((l) => (
          <Cartao key={l.id} lead={l} />
        ))}
        {coluna.leads.length === 0 && <p className="text-apagado text-xs">Nenhum lead.</p>}
      </div>
    </div>
  )
}

function Cartao({ lead }: { lead: CartaoLead }) {
  return (
    <Link
      href={`/leads/${lead.id}`}
      className="block space-y-1 rounded-md border border-linha bg-superficie px-3 py-2.5 text-sm transition-colors duration-150 hover:border-linha-forte hover:bg-elevada"
    >
      <div className="flex items-center justify-between gap-2">
        <span className="truncate font-medium text-texto">{lead.nome}</span>
        <span className="font-mono text-acento text-xs">{lead.pontuacao}</span>
      </div>
      <p className="truncate text-apagado text-xs">
        {lead.nicho} · {lead.bairro}
      </p>
      {lead.dores.length > 0 && (
        <p className="truncate text-suave text-xs">{lead.dores.map((d) => DOR[d] ?? d).join(', ')}</p>
      )}
    </Link>
  )
}

function Saidas({ saidas }: { saidas: Saida[] }) {
  return (
    <details className="text-xs">
      <summary className="cursor-pointer select-none text-apagado">Saídas</summary>
      <ul className="mt-2 space-y-1">
        {saidas.map((s) => (
          <li key={s.etapa} className="text-suave">
            {s.rotulo}: <span className="font-mono text-texto">{s.n}</span>
          </li>
        ))}
      </ul>
    </details>
  )
}
