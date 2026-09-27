import type { Metadata } from 'next'
import Link from 'next/link'
import { deveCobrar, proximaCobranca, rotuloDias, TEXTO_FOLLOW_UP } from '@/lib/cobrar.ts'
import { candidatosCobranca, tamanhoDaFila } from '@/lib/consultas.ts'
import { seguinte, vizinhos } from '@/lib/mesa.ts'
import { Despacho } from '../despacho.tsx'
import { PainelFoco } from '../foco.tsx'
import { ColunaFila, hrefLead, Mesa, TopoDeck, Vazio } from '../mesa.tsx'
import { botao } from '../ui.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Cobrar' }

// O servidor pode rodar em UTC; a hora mostrada é a de João Pessoa.
const DIA = new Intl.DateTimeFormat('pt-BR', {
  weekday: 'long',
  day: '2-digit',
  month: '2-digit',
  timeZone: 'America/Fortaleza',
})
const HORA = new Intl.DateTimeFormat('pt-BR', { hour: '2-digit', minute: '2-digit', timeZone: 'America/Fortaleza' })

export default async function Cobrar({ searchParams }: { searchParams: Promise<{ lead?: string | string[] }> }) {
  const [{ lead }, candidatos] = await Promise.all([searchParams, candidatosCobranca()])
  const agora = new Date()
  const leads = candidatos.filter((l) => deveCobrar(l, agora))
  const ids = leads.map((l) => l.id)
  const v = vizinhos(ids, lead)
  if (!v) return <CobrarVazio proxima={proximaCobranca(candidatos, agora)} esperando={candidatos.length} />
  const atual = leads[v.posicao - 1]
  const itens = leads.map((l) => ({
    id: l.id,
    nome: l.nome,
    pontuacao: l.pontuacao,
    meta: (
      <>
        <span className="font-mono">{rotuloDias(l.contatado_em, agora)}</span> · {l.bairro}
      </>
    ),
  }))

  return (
    <Mesa coluna={<ColunaFila rotulo="Cobrar" base="/cobrar" itens={itens} atual={v.atual} />}>
      <TopoDeck base="/cobrar" v={v} />
      <PainelFoco
        lead={atual}
        hrefFicha={`/leads/${atual.id}`}
        extra={
          <span>
            contatado <span className="text-texto font-mono">{rotuloDias(atual.contatado_em, agora)}</span>
          </span>
        }
      >
        <Despacho
          key={atual.id}
          modo="cobrar"
          id={atual.id}
          telefone={atual.telefone}
          mensagem={TEXTO_FOLLOW_UP}
          depois={hrefLead('/cobrar', seguinte(ids, v.atual)) ?? '/cobrar'}
          anterior={hrefLead('/cobrar', v.anterior)}
          proximo={hrefLead('/cobrar', v.proximo)}
        />
      </PainelFoco>
    </Mesa>
  )
}

async function CobrarVazio({ proxima, esperando }: { proxima: Date | null; esperando: number }) {
  const fila = await tamanhoDaFila()
  return (
    <Vazio frase="Nada para cobrar agora.">
      {proxima ? (
        <p>
          A próxima cobrança abre{' '}
          <span className="text-texto">
            {DIA.format(proxima)}, às <span className="font-mono">{HORA.format(proxima)}</span>
          </span>
          , quando o contato mais antigo completa 3 dias sem resposta.{' '}
          {esperando === 1 ? '1 lead contatado está' : `${esperando} leads contatados estão`} nessa espera.
        </p>
      ) : (
        <p>Nenhum lead contatado esperando resposta. A cobrança começa 3 dias depois do primeiro envio.</p>
      )}
      {fila > 0 && (
        <Link href="/" className={botao('secundario')}>
          Ir para a Fila <span className="font-mono">{fila}</span>
        </Link>
      )}
    </Vazio>
  )
}
