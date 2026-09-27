import type { Metadata } from 'next'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { LeadFoco } from '@/lib/consultas.ts'
import { sql } from '@/lib/db.ts'
import { Despacho } from '../../despacho.tsx'
import { PainelFoco } from '../../foco.tsx'
import { CaminhoEtapas, Conversa, type Mensagem } from './caminho.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Lead' }

type LeadFicha = LeadFoco & { etapa: string; saiu_de: string | null }

export default async function FichaLead({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) notFound()
  const [[l], mensagens] = await Promise.all([
    sql()`select id, nome, nicho, bairro, telefone, nota, avaliacoes, dores, detalhes, pontuacao, variante,
        mensagem, etapa, saiu_de
      from leads where id = ${id}`,
    sql()`select id, autor, texto, criado_em from mensagens where lead_id = ${id} order by criado_em`,
  ])
  if (!l) notFound()
  const lead = { ...l, id: String(l.id) } as LeadFicha
  const conversa = mensagens as Mensagem[]

  return (
    <>
      <div className="px-4 pt-6 lg:px-10 lg:pt-9">
        <Link href="/leads" className="inline-block text-sm text-suave transition-colors duration-150 hover:text-texto">
          ← Leads
        </Link>
      </div>
      <PainelFoco lead={lead}>
        <CaminhoEtapas id={lead.id} etapa={lead.etapa} saiuDe={lead.saiu_de} />
        <Despacho
          key={lead.id}
          modo="ficha"
          id={lead.id}
          telefone={lead.telefone}
          mensagem={lead.mensagem}
          variante={lead.variante}
          depois={lead.etapa === 'na_fila' ? `/leads/${lead.id}` : null}
        />
        <Conversa mensagens={conversa} />
      </PainelFoco>
    </>
  )
}
