import Link from 'next/link'
import { sair } from '../login/acoes.ts'
import { Menu } from './menu.tsx'

function Marca() {
  return (
    <Link href="/" className="font-semibold text-[15px] text-texto tracking-[-0.02em]">
      lancer
    </Link>
  )
}

function Sair() {
  return (
    <form action={sair}>
      <button type="submit" className="text-apagado text-sm transition-colors duration-150 hover:text-texto">
        Sair
      </button>
    </form>
  )
}

export function Trilho({ fila, cobrar }: { fila: number; cobrar: number }) {
  return (
    <>
      <header className="border-linha border-b pt-[env(safe-area-inset-top)] lg:hidden">
        <div className="flex h-11 items-center justify-between px-4">
          <Marca />
          <Sair />
        </div>
      </header>
      <nav
        aria-label="Seções"
        className="fixed inset-x-0 bottom-0 z-20 grid grid-cols-5 border-linha border-t bg-fundo pb-[env(safe-area-inset-bottom)] lg:static lg:flex lg:flex-col lg:gap-0.5 lg:border-t-0 lg:border-r lg:px-3 lg:py-5"
      >
        <div className="hidden lg:mb-7 lg:block lg:px-3">
          <Marca />
        </div>
        <Menu contagens={{ fila, cobrar }} />
        <div className="hidden lg:mt-auto lg:block lg:px-3">
          <Sair />
        </div>
      </nav>
    </>
  )
}
