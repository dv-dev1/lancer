'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

type Contagens = { fila: number; cobrar: number }

const SECOES: { href: string; rotulo: string; contagem?: keyof Contagens }[] = [
  { href: '/', rotulo: 'Fila', contagem: 'fila' },
  { href: '/cobrar', rotulo: 'Cobrar', contagem: 'cobrar' },
  { href: '/leads', rotulo: 'Leads' },
  { href: '/metricas', rotulo: 'Métricas' },
  { href: '/buscar', rotulo: 'Buscar' },
]

// Um DOM só: no celular cada link é uma célula da barra inferior; do lg para cima, uma linha do trilho.
export function Menu({ contagens }: { contagens: Contagens }) {
  const rota = usePathname()
  return SECOES.map(({ href, rotulo, contagem }) => {
    const ativo = href === '/' ? rota === '/' : rota.startsWith(href)
    return (
      <Link
        key={href}
        href={href}
        aria-current={ativo ? 'page' : undefined}
        className={`relative flex h-14 items-center justify-center gap-1.5 text-[13px] transition-colors duration-150 lg:h-9 lg:justify-between lg:rounded-md lg:px-3 lg:text-sm ${
          ativo
            ? 'text-acento before:absolute before:top-0 before:left-1/2 before:h-0.5 before:w-8 before:-translate-x-1/2 before:bg-acento lg:bg-elevada lg:before:hidden'
            : 'text-suave hover:text-texto lg:hover:bg-superficie'
        }`}
      >
        {rotulo}
        {contagem && (
          <span className={`font-mono text-xs ${ativo ? 'text-acento' : 'text-apagado'}`}>{contagens[contagem]}</span>
        )}
      </Link>
    )
  })
}
