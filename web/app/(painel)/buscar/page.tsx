import type { Metadata } from 'next'
import Link from 'next/link'
import { sql } from '@/lib/db.ts'
import { CIDADES_PB, ESTADO_PEDIDO, lugarDoPedido, NICHO } from '@/lib/pedidos.ts'
import { pedirColeta } from '../acoes.ts'
import { botao, Etiqueta } from '../ui.tsx'
import { Atualizar } from './atualizar.tsx'

export const dynamic = 'force-dynamic'
export const metadata: Metadata = { title: 'Buscar' }

type Pedido = {
  id: string
  cidade: string
  bairro: string
  nicho: string
  limite: number
  estado: string
  novos: number | null
  erro: string | null
}

const CAMPO =
  'block h-10 w-full rounded-md border border-linha-forte bg-elevada px-3 text-sm text-texto transition-colors duration-150 focus:border-acento focus:outline-hidden'

export default async function Buscar({ searchParams }: { searchParams: Promise<{ erro?: string }> }) {
  const { erro } = await searchParams
  const pedidos = (await sql()`select id, cidade, bairro, nicho, limite, estado, novos, erro from pedidos
    order by id desc limit 20`) as Pedido[]
  const ativo = pedidos.some((p) => p.estado === 'pendente' || p.estado === 'rodando')

  return (
    <div className="max-w-[40rem] space-y-10 p-4 lg:p-10">
      <Atualizar ativo={ativo} />
      <form action={pedirColeta} className="space-y-5">
        <Etiqueta>Nova busca no Google Maps</Etiqueta>
        <div className="grid gap-4 sm:grid-cols-2">
          <Campo rotulo="Cidade" id="cidade">
            <input id="cidade" name="cidade" list="cidades-pb" required defaultValue="João Pessoa" className={CAMPO} />
            <datalist id="cidades-pb">
              {CIDADES_PB.map((c) => (
                <option key={c} value={c} />
              ))}
            </datalist>
          </Campo>
          <Campo rotulo="Bairro (opcional)" id="bairro">
            <input id="bairro" name="bairro" placeholder="a cidade inteira" className={CAMPO} />
          </Campo>
          <Campo rotulo="Nicho" id="nicho">
            <select id="nicho" name="nicho" className={CAMPO}>
              {Object.entries(NICHO).map(([valor, rotulo]) => (
                <option key={valor} value={valor}>
                  {rotulo}
                </option>
              ))}
            </select>
          </Campo>
          <Campo rotulo="Quantos leads" id="limite">
            <input
              id="limite"
              name="limite"
              type="number"
              min={1}
              max={20}
              defaultValue={5}
              required
              className={`${CAMPO} font-mono`}
            />
          </Campo>
        </div>
        {erro && (
          <p role="alert" className="text-aviso text-sm">
            {erro}
          </p>
        )}
        <button type="submit" className={botao('primario')}>
          Buscar leads
        </button>
      </form>

      <section className="space-y-3">
        <Etiqueta>Pedidos</Etiqueta>
        {pedidos.length === 0 && <p className="text-sm text-suave">Nenhuma busca ainda.</p>}
        <ul className="divide-y divide-linha rounded-md border border-linha bg-superficie">
          {pedidos.map((p) => (
            <li key={p.id} className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 px-4 py-3">
              <div className="min-w-0">
                <p className="text-sm text-texto">
                  {NICHO[p.nicho] ?? p.nicho} · {lugarDoPedido(p.cidade, p.bairro)}
                </p>
                <p className="text-suave text-xs">
                  até <span className="font-mono">{p.limite}</span> leads
                </p>
              </div>
              <Situacao p={p} />
            </li>
          ))}
        </ul>
        {pedidos.some((p) => p.estado === 'pendente') && (
          <p className="text-suave text-xs">
            Pedido esperando só anda com o coletor ligado:{' '}
            <code className="font-mono text-texto">go run ./cmd/coleta --servir</code>
          </p>
        )}
      </section>
    </div>
  )
}

function Campo({ rotulo, id, children }: { rotulo: string; id: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="block text-suave text-xs">
        {rotulo}
      </label>
      {children}
    </div>
  )
}

function Situacao({ p }: { p: Pedido }) {
  if (p.estado === 'pronto') {
    return p.novos ? (
      <Link href={`/leads?pedido=${p.id}`} className="text-positivo text-sm hover:underline">
        <span className="font-mono">{p.novos}</span> {p.novos === 1 ? 'lead novo' : 'leads novos'} →
      </Link>
    ) : (
      <span className="text-sm text-suave">Nenhum lead novo</span>
    )
  }
  if (p.estado === 'erro') {
    return <span className="max-w-full text-aviso text-sm break-words">{p.erro}</span>
  }
  return (
    <span className={`text-sm ${p.estado === 'rodando' ? 'text-acento' : 'text-suave'}`}>
      {ESTADO_PEDIDO[p.estado]}…
    </span>
  )
}
