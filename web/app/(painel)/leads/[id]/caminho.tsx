import { ETAPA } from '@/lib/rotulos.ts'
import { alcancou, ESCADA, type Etapa } from '@/lib/tipos.ts'
import { mudarEtapa } from '../../acoes.ts'
import { Etiqueta } from '../../ui.tsx'

export type Mensagem = { id: string | number; autor: string; texto: string }

export function CaminhoEtapas({ id, etapa, saiuDe }: { id: string; etapa: string; saiuDe: string | null }) {
  // Quem saiu não tem degrau "atual": a referência para marcar o que já foi feito vira o degrau de onde saiu.
  const referencia = etapa === 'perdido' || etapa === 'saiu' ? (saiuDe ?? 'na_fila') : etapa

  return (
    <form action={mudarEtapa} className="space-y-3">
      <input type="hidden" name="id" value={id} />
      <Etiqueta>Etapa</Etiqueta>
      <ol className="flex flex-col gap-1.5 lg:flex-row lg:gap-1.5">
        {ESCADA.map((degrau) => (
          <li key={degrau} className="min-w-0 lg:flex-1">
            <Degrau etapa={degrau} atual={etapa === degrau} feito={alcancou(referencia, degrau)} />
          </li>
        ))}
      </ol>
      <div className="flex gap-1.5">
        <SaidaBotao etapa="perdido" atual={etapa === 'perdido'} />
        <SaidaBotao etapa="saiu" atual={etapa === 'saiu'} />
      </div>
    </form>
  )
}

function Degrau({ etapa, atual, feito }: { etapa: Etapa; atual: boolean; feito: boolean }) {
  return (
    <button
      type="submit"
      name="etapa"
      value={etapa}
      className={`block w-full min-w-0 break-words rounded-md border px-3 py-2 text-left text-sm transition-colors duration-150 lg:text-center ${
        atual
          ? 'border-acento text-acento'
          : feito
            ? 'border-linha-forte text-texto hover:bg-elevada'
            : 'border-linha text-suave hover:bg-elevada hover:text-texto'
      }`}
    >
      {ETAPA[etapa]}
    </button>
  )
}

function SaidaBotao({ etapa, atual }: { etapa: 'perdido' | 'saiu'; atual: boolean }) {
  return (
    <button
      type="submit"
      name="etapa"
      value={etapa}
      className={`rounded-md border px-3 py-1.5 text-xs transition-colors duration-150 ${
        atual ? 'border-linha-forte text-texto' : 'border-linha text-apagado hover:text-suave'
      }`}
    >
      {ETAPA[etapa]}
    </button>
  )
}

export function Conversa({ mensagens }: { mensagens: Mensagem[] }) {
  return (
    <section className="space-y-3">
      <Etiqueta>Conversa</Etiqueta>
      {mensagens.length === 0 ? (
        <p className="text-apagado text-sm">Nenhuma mensagem registrada ainda.</p>
      ) : (
        <ol className="space-y-4 border-linha border-l pl-4">
          {mensagens.map((m) => (
            <li key={m.id} className="space-y-0.5">
              <p className="text-apagado text-xs uppercase tracking-wide">{m.autor === 'lead' ? 'Lead' : 'Eu'}</p>
              <p className="whitespace-pre-wrap break-words text-[15px] text-texto">{m.texto}</p>
            </li>
          ))}
        </ol>
      )}
    </section>
  )
}
