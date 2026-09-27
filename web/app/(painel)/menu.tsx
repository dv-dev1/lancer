'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

const ITENS = [
  { href: '/', rotulo: 'Fila' },
  { href: '/cobrar', rotulo: 'Cobrar' },
  { href: '/leads', rotulo: 'Leads' },
  { href: '/metricas', rotulo: 'Métricas' },
]

export function Menu() {
  const rota = usePathname()
  return ITENS.map(({ href, rotulo }) => {
    const ativo = href === '/' ? rota === '/' : rota.startsWith(href)
    return (
      <Link
        key={href}
        href={href}
        aria-current={ativo ? 'page' : undefined}
        className={`relative text-sm transition-colors duration-150 hover:text-white ${ativo ? 'text-white' : 'text-suave'}`}
      >
        {rotulo}
        {ativo && <span className="gradiente-marca absolute -bottom-[1.1rem] left-0 h-0.5 w-full rounded-full" />}
      </Link>
    )
  })
}
