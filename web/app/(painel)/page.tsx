import type { Metadata } from 'next'
import Link from 'next/link'
import { deveCobrar } from '@/lib/cobrar.ts'
import { candidatosCobranca, leadsDaFila } from '@/lib/consultas.ts'
import { seguinte, vizinhos } from '@/lib/mesa.ts'
import { Despacho } from './despacho.tsx'
import { PainelFoco } from './foco.tsx'
import { ColunaFila, hrefLead, Mesa, TopoDeck, Vazio } from './mesa.tsx'
import { botao } from './ui.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Fila' }

export default async function Fila({ searchParams }: { searchParams: Promise<{ lead?: string | string[] }> }) {
  const [{ lead }, leads] = await Promise.all([searchParams, leadsDaFila()])
  const ids = leads.map((l) => l.id)
  const v = vizinhos(ids, lead)
  if (!v) return <FilaVazia />
  const atual = leads[v.posicao - 1]
  const itens = leads.map((l) => ({ id: l.id, nome: l.nome, pontuacao: l.pontuacao, meta: `${l.nicho} · ${l.bairro}` }))

  return (
    <Mesa coluna={<ColunaFila rotulo="Fila" base="/" itens={itens} atual={v.atual} />}>
      <TopoDeck base="/" v={v} />
      <PainelFoco lead={atual} hrefFicha={`/leads/${atual.id}`}>
        <Despacho
          key={atual.id}
          modo="fila"
          id={atual.id}
          telefone={atual.telefone}
          mensagem={atual.mensagem}
          variante={atual.variante}
          depois={hrefLead('/', seguinte(ids, v.atual)) ?? '/'}
          anterior={hrefLead('/', v.anterior)}
          proximo={hrefLead('/', v.proximo)}
        />
      </PainelFoco>
    </Mesa>
  )
}

async function FilaVazia() {
  const agora = new Date()
  const pendentes = (await candidatosCobranca()).filter((l) => deveCobrar(l, agora)).length
  if (pendentes > 0)
    return (
      <Vazio frase="Fila vazia.">
        <p>
          Ninguém esperando o primeiro contato, mas{' '}
          {pendentes === 1 ? '1 lead já pode ser cobrado' : `${pendentes} leads já podem ser cobrados`}.
        </p>
        <Link href="/cobrar" className={botao('primario')}>
          Ir para o Cobrar <span className="font-mono">{pendentes}</span>
        </Link>
      </Vazio>
    )
  return (
    <Vazio frase="Fila vazia.">
      <p>Ninguém esperando o primeiro contato. Uma coleta nova enche a fila:</p>
      <code className="block whitespace-pre-wrap break-words rounded-md border border-linha bg-superficie px-4 py-3 font-mono text-[13px] text-texto">
        go run ./cmd/coleta --nicho confeitaria --bairro "Manaíra"
      </code>
    </Vazio>
  )
}
