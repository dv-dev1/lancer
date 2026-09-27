import type { Metadata } from 'next'
import Link from 'next/link'
import { linkWhatsApp } from '@/lib/abordagem.ts'
import { sql } from '@/lib/db.ts'
import { VARIANTE } from '@/lib/rotulos.ts'
import { Cabecalho, Selo } from './ui.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Fila' }

export default async function Fila() {
  const leads = await sql()`select id, nome, nicho, bairro, telefone, mensagem, pontuacao, variante
    from leads where etapa = 'na_fila' order by pontuacao desc`

  return (
    <div className="space-y-6">
      <Cabecalho
        titulo="Fila"
        descricao={`${leads.length} ${leads.length === 1 ? 'lead' : 'leads'} esperando o primeiro contato`}
      />
      <div className="cartao surgir overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-borda text-xs uppercase tracking-widest text-suave">
            <tr>
              <th className="px-4 py-3 font-normal">Lead</th>
              <th className="font-normal">Pontuação</th>
              <th className="font-normal">Variante</th>
              <th className="pr-4 font-normal" />
            </tr>
          </thead>
          <tbody>
            {leads.map((l) => (
              <tr key={l.id} className="border-b border-borda last:border-0">
                <td className="px-4 py-3">
                  <Link className="font-medium text-white hover:text-ciano" href={`/leads/${encodeURIComponent(l.id)}`}>
                    {l.nome}
                  </Link>
                  <div className="text-suave">
                    {l.nicho} · {l.bairro}
                  </div>
                </td>
                <td className="tabular-nums">{l.pontuacao}</td>
                <td>
                  <Selo>{VARIANTE[l.variante] ?? l.variante}</Selo>
                </td>
                <td className="pr-4">
                  <a
                    className="botao-primario inline-block px-4 py-1.5 text-sm transition-[filter] duration-150 hover:brightness-110"
                    href={linkWhatsApp(l.telefone, l.mensagem)}
                    target="_blank"
                    rel="noopener"
                  >
                    Abrir no WhatsApp
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {leads.length === 0 && <p className="py-10 text-center text-suave">Nenhum lead na fila.</p>}
      </div>
    </div>
  )
}
