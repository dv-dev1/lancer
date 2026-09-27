import Link from 'next/link'
import type { LeadFoco } from '@/lib/consultas.ts'
import { DOR } from '@/lib/rotulos.ts'
import { Etiqueta } from './ui.tsx'

function nota(valor: string | number, avaliacoes: number | null): string {
  const n = Number(valor).toLocaleString('pt-BR', { minimumFractionDigits: 1, maximumFractionDigits: 1 })
  return `${n} (${avaliacoes ?? '?'})`
}

function Ponto() {
  return (
    <span aria-hidden className="text-apagado">
      ·
    </span>
  )
}

export function PainelFoco({
  lead,
  hrefFicha,
  extra,
  children,
}: {
  lead: LeadFoco
  hrefFicha?: string
  extra?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <article className="max-w-[42rem] space-y-9 px-4 pt-6 pb-28 lg:px-10 lg:pt-9 lg:pb-12">
      <header className="space-y-2.5">
        <h1 className="break-words font-semibold text-[1.875rem] leading-[1.1] tracking-[-0.025em] lg:text-[2.5rem]">
          {hrefFicha ? (
            <Link
              href={hrefFicha}
              className="decoration-linha-forte underline-offset-[6px] transition-colors duration-150 hover:underline"
            >
              {lead.nome}
            </Link>
          ) : (
            lead.nome
          )}
        </h1>
        <p className="flex flex-wrap items-baseline gap-x-2 gap-y-1 text-sm text-suave">
          <span>{lead.nicho}</span>
          <Ponto />
          <span>{lead.bairro}</span>
          <Ponto />
          <span>
            nota <span className="font-mono text-texto">{nota(lead.nota, lead.avaliacoes)}</span>
          </span>
          <Ponto />
          <span>
            pontuação <span className="font-mono text-texto">{lead.pontuacao}</span>
          </span>
          {extra && (
            <>
              <Ponto />
              {extra}
            </>
          )}
        </p>
      </header>
      <section aria-labelledby="por-que" className="space-y-3">
        <Etiqueta id="por-que">Por que</Etiqueta>
        <ul className="space-y-3">
          {lead.dores.map((d) => (
            <li key={d} className="space-y-0.5 break-words">
              <p className="text-[15px] text-texto">{DOR[d] ?? d}</p>
              {lead.detalhes?.[d] && <p className="text-sm text-suave">{lead.detalhes[d]}</p>}
            </li>
          ))}
        </ul>
      </section>
      {children}
    </article>
  )
}

export function AvisoPreview() {
  return (
    <p className="text-aviso text-sm">
      O link de preview desta mensagem ainda não existe (fase 3). Apague a frase com{' '}
      <code className="font-mono">&lt;preview&gt;</code> antes de enviar.
    </p>
  )
}
