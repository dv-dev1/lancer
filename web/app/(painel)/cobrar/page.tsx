import type { Metadata } from 'next'
import { revalidatePath } from 'next/cache'
import Link from 'next/link'
import { linkWhatsApp } from '@/lib/abordagem.ts'
import { deveCobrar, TEXTO_FOLLOW_UP } from '@/lib/cobrar.ts'
import { sql } from '@/lib/db.ts'
import { exigirSessao } from '@/lib/guarda.ts'
import { Cabecalho } from '../ui.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Cobrar' }

// R22: só marca o envio; quem tira o lead da régua (perdido, 3 dias depois) é o MarcarPerdidos do coletor.
async function marcarFollowUp(form: FormData) {
  'use server'
  await exigirSessao()
  const id = String(form.get('id'))
  await sql()`update leads set follow_up_em = now() where id = ${id}`
  revalidatePath('/cobrar')
}

type Candidato = {
  // bigint vem do driver como string; o id só é usado como texto (href, input hidden), nunca em conta.
  id: string | number
  nome: string
  telefone: string
  contatado_em: string | null
  respondeu_em: string | null
  follow_up_em: string | null
}

export default async function Cobrar() {
  // A régua dos 3 dias é o deveCobrar, não o SQL: a query só traz quem ainda pode entrar nela.
  const candidatos = (await sql()`select id, nome, telefone, contatado_em, respondeu_em, follow_up_em from leads
    where contatado_em is not null and respondeu_em is null and follow_up_em is null`) as Candidato[]
  const agora = new Date()
  const leads = candidatos.filter((l) => deveCobrar(l, agora))

  return (
    <div className="space-y-6">
      <Cabecalho
        titulo="Cobrar"
        descricao={`${leads.length} ${leads.length === 1 ? 'lead' : 'leads'} sem resposta há 3 dias`}
      />
      <div className="cartao surgir divide-y divide-borda">
        {leads.map((l) => (
          <div key={l.id} className="flex flex-wrap items-center justify-between gap-4 p-4">
            <Link className="font-medium text-white hover:text-ciano" href={`/leads/${encodeURIComponent(l.id)}`}>
              {l.nome}
            </Link>
            <div className="flex items-center gap-2">
              <a
                className="botao-primario px-4 py-1.5 text-sm transition-[filter] duration-150 hover:brightness-110"
                href={linkWhatsApp(l.telefone, TEXTO_FOLLOW_UP)}
                target="_blank"
                rel="noopener"
              >
                Abrir no WhatsApp
              </a>
              <form action={marcarFollowUp}>
                <input type="hidden" name="id" value={l.id} />
                <button
                  type="submit"
                  className="rounded-full border border-borda px-4 py-1.5 text-sm text-suave transition-colors duration-150 hover:text-white"
                >
                  Follow-up enviado
                </button>
              </form>
            </div>
          </div>
        ))}
        {leads.length === 0 && <p className="py-10 text-center text-suave">Nenhum lead para cobrar.</p>}
      </div>
    </div>
  )
}
