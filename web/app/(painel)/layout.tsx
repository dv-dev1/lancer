import Link from 'next/link'
import { sair } from '../login/acoes.ts'
import { Menu } from './menu.tsx'

export default function LayoutPainel({ children }: { children: React.ReactNode }) {
  return (
    <>
      <header className="border-b border-borda">
        <nav className="mx-auto flex max-w-6xl items-center gap-5 px-4 py-4 sm:gap-8 sm:px-6">
          <Link href="/" className="text-xl font-medium tracking-tight">
            lancer
          </Link>
          <Menu />
          <form action={sair} className="ml-auto">
            <button type="submit" className="text-sm text-suave transition-colors duration-150 hover:text-white">
              Sair
            </button>
          </form>
        </nav>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">{children}</main>
    </>
  )
}
