import Link from 'next/link'
import type { ReactNode } from 'react'
import type { Vizinhos } from '@/lib/mesa.ts'
import { Etiqueta } from './ui.tsx'

export type ItemFila = { id: string; nome: string; meta: string | ReactNode; pontuacao: number }

export const hrefLead = (base: string, id: string | null) => (id ? `${base}?lead=${id}` : null)

export function Mesa({ coluna, children }: { coluna: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="lg:grid lg:h-full lg:grid-cols-[18rem_minmax(0,1fr)] lg:grid-rows-[minmax(0,1fr)]">
      {coluna}
      <div className="min-w-0 lg:overflow-y-auto">{children}</div>
    </div>
  )
}

export function ColunaFila({
  rotulo,
  base,
  itens,
  atual,
}: {
  rotulo: string
  base: string
  itens: ItemFila[]
  atual: string
}) {
  return (
    <aside aria-label={rotulo} className="hidden min-h-0 flex-col border-linha border-r bg-superficie lg:flex">
      <div className="flex h-12 shrink-0 items-center justify-between border-linha border-b px-4">
        <Etiqueta>{rotulo}</Etiqueta>
        <span className="font-mono text-apagado text-xs">{itens.length}</span>
      </div>
      <ol className="min-h-0 flex-1 overflow-y-auto">
        {itens.map((item) => {
          const ehAtual = item.id === atual
          return (
            <li key={item.id}>
              <Link
                href={`${base}?lead=${item.id}`}
                replace
                scroll={false}
                prefetch={false}
                aria-current={ehAtual ? 'true' : undefined}
                data-atual={ehAtual || undefined}
                className={`group grid grid-cols-[minmax(0,1fr)_auto] gap-x-3 gap-y-0.5 border-b border-b-linha border-l px-4 py-3 transition-colors duration-150 ${
                  ehAtual ? 'border-l-acento bg-elevada' : 'border-l-transparent hover:bg-elevada'
                }`}
              >
                <span className="truncate font-medium text-sm text-texto">{item.nome}</span>
                <span className={`font-mono text-sm ${ehAtual ? 'text-acento' : 'text-suave'}`}>{item.pontuacao}</span>
                {/* apagado não passa AA sobre elevada: no atual e no hover o metadado sobe para suave. */}
                <span
                  className={`col-span-2 truncate text-xs ${ehAtual ? 'text-suave' : 'text-apagado group-hover:text-suave'}`}
                >
                  {item.meta}
                </span>
              </Link>
            </li>
          )
        })}
      </ol>
    </aside>
  )
}

export function TopoDeck({ base, v }: { base: string; v: Vizinhos }) {
  return (
    <div className="sticky top-0 z-10 flex h-12 items-center gap-2 border-linha border-b bg-fundo px-1 lg:px-10">
      <Seta href={hrefLead(base, v.anterior)} rotulo="Lead anterior" movel="‹" tecla="←" className="lg:ml-auto" />
      <p className="flex-1 text-center font-mono text-sm text-suave lg:order-first lg:flex-none lg:text-left">
        <span className="text-texto">{v.posicao}</span> de {v.total}
      </p>
      <Seta href={hrefLead(base, v.proximo)} rotulo="Próximo lead" movel="›" tecla="→" />
    </div>
  )
}

function Seta({
  href,
  rotulo,
  movel,
  tecla,
  className = '',
}: {
  href: string | null
  rotulo: string
  movel: string
  tecla: string
  className?: string
}) {
  const forma = `inline-flex size-11 items-center justify-center rounded-md text-xl lg:size-8 lg:border lg:border-linha-forte lg:font-mono lg:text-sm ${className}`
  const glifo = (
    <>
      <span className="lg:hidden">{movel}</span>
      <span className="hidden lg:inline">{tecla}</span>
    </>
  )
  if (!href)
    return (
      <span aria-hidden className={`${forma} text-linha-forte lg:border-linha`}>
        {glifo}
      </span>
    )
  return (
    <Link
      href={href}
      replace
      scroll={false}
      prefetch={false}
      aria-label={rotulo}
      className={`${forma} text-suave transition-colors duration-150 hover:bg-elevada hover:text-texto`}
    >
      {glifo}
    </Link>
  )
}

export function Vazio({ frase, children }: { frase: string; children: React.ReactNode }) {
  return (
    <div className="max-w-[36rem] space-y-4 px-4 py-10 lg:px-10 lg:py-14">
      <p className="font-semibold text-2xl text-texto tracking-[-0.02em]">{frase}</p>
      <div className="space-y-6 text-[15px] text-suave leading-relaxed">{children}</div>
    </div>
  )
}
