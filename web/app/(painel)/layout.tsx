import { contagens } from '@/lib/consultas.ts'
import { Trilho } from './trilho.tsx'

// Consulta o banco a cada pedido; sem isto o build tentaria pré-renderizar e precisaria do DATABASE_URL.
export const dynamic = 'force-dynamic'

export default async function LayoutPainel({ children }: { children: React.ReactNode }) {
  const { fila, cobrar } = await contagens()
  return (
    <div className="lg:grid lg:h-dvh lg:grid-cols-[12rem_minmax(0,1fr)] lg:grid-rows-[minmax(0,1fr)]">
      <Trilho fila={fila} cobrar={cobrar} />
      <main className="min-w-0 pb-[calc(3.5rem+env(safe-area-inset-bottom))] lg:overflow-y-auto lg:pb-0">
        {children}
      </main>
    </div>
  )
}
