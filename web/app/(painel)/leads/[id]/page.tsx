import type { Metadata } from 'next'
import { revalidatePath } from 'next/cache'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { Fragment } from 'react'
import { formatarTelefone, linkWhatsApp } from '@/lib/abordagem.ts'
import { sql } from '@/lib/db.ts'
import { exigirSessao } from '@/lib/guarda.ts'
import { DOR, ETAPA, VARIANTE } from '@/lib/rotulos.ts'
import { alcancou, ETAPAS } from '@/lib/tipos.ts'
import { Selo, SeloEtapa } from '../../ui.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Lead' }

async function mudarEtapa(form: FormData) {
  'use server'
  await exigirSessao()
  const id = String(form.get('id'))
  const etapa = String(form.get('etapa'))
  if (!(ETAPAS as readonly string[]).includes(etapa)) throw new Error(`etapa inválida: ${etapa}`)
  // As datas só se preenchem uma vez (pílula clicada de novo não reseta), e pular degrau preenche os de trás.
  // Na direita do set, `etapa` ainda é a antiga: saiu_de guarda de onde o lead saiu, para o funil.
  await sql()`update leads set etapa = ${etapa},
    contatado_em = case when ${alcancou(etapa, 'contatado')}::boolean then coalesce(contatado_em, now()) else contatado_em end,
    respondeu_em = case when ${alcancou(etapa, 'respondeu')}::boolean then coalesce(respondeu_em, now()) else respondeu_em end,
    saiu_de = case when ${etapa} in ('perdido', 'saiu') and etapa not in ('perdido', 'saiu') then etapa else saiu_de end
    where id = ${id}`
  revalidatePath(`/leads/${encodeURIComponent(id)}`)
}

export default async function FichaLead({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) notFound()
  const [[l], mensagens] = await Promise.all([
    sql()`select id, nome, nicho, bairro, telefone, site, endereco, nota, avaliacoes, dores, detalhes,
        pontuacao, variante, mensagem, etapa, contatado_em, follow_up_em, respondeu_em, criado_em
      from leads where id = ${id}`,
    sql()`select id, autor, texto, criado_em from mensagens where lead_id = ${id} order by criado_em`,
  ])
  if (!l) notFound()
  const detalhes = l.detalhes as Record<string, string>

  return (
    <div className="grid gap-6 md:grid-cols-[1fr_1.4fr]">
      <div className="space-y-4">
        <Link href="/leads" className="inline-block text-sm text-suave transition-colors duration-150 hover:text-white">
          ← Leads
        </Link>
        <section className="cartao surgir space-y-4 p-5">
          <div className="space-y-3">
            <h1 className="titulo-gradiente text-3xl font-medium tracking-[-0.05em]">{l.nome}</h1>
            <div className="flex flex-wrap items-center gap-2">
              <SeloEtapa etapa={l.etapa} rotulo={ETAPA[l.etapa] ?? l.etapa} />
              <Selo>{VARIANTE[l.variante] ?? l.variante}</Selo>
            </div>
            <div className="flex flex-wrap items-center gap-4 text-sm">
              <span className="text-suave">
                {l.nicho} · {l.bairro}
              </span>
              <a
                className="text-ciano hover:underline"
                href={linkWhatsApp(l.telefone, '')}
                target="_blank"
                rel="noopener"
              >
                {formatarTelefone(l.telefone)}
              </a>
            </div>
            <p className="text-sm text-suave">
              Pontuação {l.pontuacao} · nota {l.nota} ({l.avaliacoes ?? '?'})
            </p>
          </div>
          <a
            className="botao-primario inline-block px-5 py-2 text-sm text-white transition-[filter] duration-150 hover:brightness-110"
            href={linkWhatsApp(l.telefone, l.mensagem)}
            target="_blank"
            rel="noopener"
          >
            Abrir no WhatsApp
          </a>
        </section>

        <section className="cartao surgir space-y-2 p-5">
          <h2 className="mb-2 text-xs uppercase tracking-widest text-suave">Dores detectadas</h2>
          {(l.dores as string[]).map((d) => (
            <p key={d} className="text-sm">
              <Selo>{DOR[d] ?? d}</Selo>
              {detalhes?.[d] && <span className="ml-2 text-suave">{detalhes[d]}</span>}
            </p>
          ))}
        </section>

        <form action={mudarEtapa} className="cartao surgir p-5">
          <input type="hidden" name="id" value={l.id} />
          <h2 className="mb-3 text-xs uppercase tracking-widest text-suave">Etapa</h2>
          <div className="flex flex-wrap gap-1.5">
            {ETAPAS.map((e) => (
              <button
                key={e}
                type="submit"
                name="etapa"
                value={e}
                className={`rounded-full border px-3 py-1 text-xs transition-colors duration-150 ${
                  l.etapa === e
                    ? 'botao-primario border-transparent'
                    : 'border-borda text-suave hover:border-ciano/40 hover:text-white'
                }`}
              >
                {ETAPA[e]}
              </button>
            ))}
          </div>
        </form>
      </div>

      <section className="space-y-2">
        <h2 className="font-medium">Mensagens</h2>
        {mensagens.map((m) => (
          <Fragment key={m.id}>
            <div
              className={`max-w-[85%] whitespace-pre-wrap rounded-2xl p-3 text-sm ${
                m.autor === 'lead' ? 'cartao' : 'ml-auto border border-ciano/30 bg-ciano/10'
              }`}
            >
              <div className="mb-1 text-[10px] uppercase tracking-wide text-suave">
                {m.autor === 'lead' ? 'Lead' : 'Eu'}
              </div>
              {m.texto}
            </div>
          </Fragment>
        ))}
        {mensagens.length === 0 && <p className="text-sm text-suave">Nenhuma mensagem registrada ainda.</p>}
      </section>
    </div>
  )
}
