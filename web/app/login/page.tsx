import type { Metadata } from 'next'
import { destinoSeguro } from '@/lib/sessao.ts'
import { Formulario } from './formulario.tsx'

export const metadata: Metadata = { title: 'Entrar' }

export default async function Login({ searchParams }: { searchParams: Promise<{ de?: string }> }) {
  const { de } = await searchParams
  return (
    <main className="flex min-h-dvh flex-col justify-center px-6 pt-[max(env(safe-area-inset-top),2.5rem)] pb-[max(env(safe-area-inset-bottom),2.5rem)] lg:px-24">
      <div className="w-full max-w-[21rem] space-y-14">
        <h1 className="font-semibold text-[3.25rem] leading-none tracking-[-0.04em] lg:text-[4.5rem]">lancer</h1>
        <Formulario de={destinoSeguro(de)} />
      </div>
    </main>
  )
}
