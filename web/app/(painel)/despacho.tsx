'use client'

import { useRouter } from 'next/navigation'
import { type RefObject, useEffect, useId, useRef, useState } from 'react'
import { useFormStatus } from 'react-dom'
import { linkWhatsApp, semPreview } from '@/lib/abordagem.ts'
import { VARIANTE } from '@/lib/rotulos.ts'
import { acaoDaTecla, type Modo, TECLA_CONFIRMAR } from '@/lib/teclas.ts'
import { marcarEnviado, marcarFollowUp } from './acoes.ts'
import { AvisoPreview } from './foco.tsx'
import { botao, Etiqueta, Tecla } from './ui.tsx'

const CONFIRMAR: Record<Modo, { acao: (form: FormData) => Promise<void>; rotulo: string; titulo: string }> = {
  fila: { acao: marcarEnviado, rotulo: 'Enviei', titulo: 'Mensagem' },
  ficha: { acao: marcarEnviado, rotulo: 'Enviei', titulo: 'Mensagem' },
  cobrar: { acao: marcarFollowUp, rotulo: 'Follow-up enviado', titulo: 'Follow-up' },
}

// Montar com key={lead.id}: sem isso o texto editado de um lead vaza para o próximo.
export function Despacho({
  modo,
  id,
  telefone,
  mensagem,
  variante,
  depois,
  anterior = null,
  proximo = null,
}: {
  modo: Modo
  id: string
  telefone: string
  mensagem: string
  variante?: string
  // null esconde a confirmação: na ficha, o E só vale para quem ainda está na fila.
  depois: string | null
  anterior?: string | null
  proximo?: string | null
}) {
  const [texto, setTexto] = useState(mensagem)
  const router = useRouter()
  const link = useRef<HTMLAnchorElement>(null)
  const form = useRef<HTMLFormElement>(null)
  const pendente = useRef(false)
  const idTitulo = useId()
  const { acao, rotulo, titulo } = CONFIRMAR[modo]

  // O lead atual pode estar abaixo da dobra da coluna; a navegação usa scroll={false} e não o traria à vista.
  useEffect(() => {
    document.querySelector('[data-atual]')?.scrollIntoView({ block: 'nearest' })
  }, [])

  useEffect(() => {
    function aoTeclar(e: KeyboardEvent) {
      if (e.defaultPrevented || e.isComposing) return
      const alvo = e.target instanceof HTMLElement ? e.target : null
      const tecla = acaoDaTecla(
        {
          key: e.key,
          tag: alvo?.tagName ?? '',
          editavel: alvo?.isContentEditable ?? false,
          modificador: e.ctrlKey || e.metaKey || e.altKey,
          repetida: e.repeat,
        },
        modo,
      )
      const destino = tecla === 'proximo' ? proximo : tecla === 'anterior' ? anterior : null
      if (tecla === 'abrir') link.current?.click()
      else if (tecla === 'confirmar' && form.current && !pendente.current) form.current.requestSubmit()
      else if (destino) router.replace(destino, { scroll: false })
      else return
      e.preventDefault()
    }
    window.addEventListener('keydown', aoTeclar)
    return () => window.removeEventListener('keydown', aoTeclar)
  }, [modo, anterior, proximo, router])

  return (
    <>
      <section aria-labelledby={idTitulo} className="space-y-3">
        <div className="flex items-baseline justify-between gap-4">
          <Etiqueta id={idTitulo}>{titulo}</Etiqueta>
          {variante && (
            <span className="text-apagado text-xs">variante {(VARIANTE[variante] ?? variante).toLowerCase()}</span>
          )}
        </div>
        <textarea
          aria-labelledby={idTitulo}
          value={texto}
          onChange={(e) => setTexto(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Escape') e.currentTarget.blur()
          }}
          rows={6}
          className="field-sizing-content block min-h-36 w-full resize-none rounded-md border border-linha-forte bg-elevada px-4 py-3 text-[15px] text-texto leading-relaxed transition-colors duration-150 focus:border-acento focus:outline-hidden"
        />
        <p className="text-apagado text-xs">
          A edição vale só para este envio.
          <span className="hidden lg:inline">
            {' '}
            <Tecla>Esc</Tecla> sai do texto e devolve as teclas.
          </span>
        </p>
        {semPreview(texto) && <AvisoPreview />}
      </section>
      <div className="fixed inset-x-0 bottom-[calc(3.5rem+env(safe-area-inset-bottom))] z-10 flex gap-2 border-linha border-t bg-fundo px-4 py-3 lg:static lg:border-0 lg:bg-transparent lg:p-0">
        <a
          ref={link}
          href={linkWhatsApp(telefone, texto)}
          target="_blank"
          rel="noopener"
          className={botao('primario', 'h-12 flex-1 lg:h-10 lg:flex-none')}
        >
          <Tecla>↵</Tecla>
          Abrir WhatsApp
        </a>
        {depois !== null && (
          <form ref={form} action={acao} className="flex flex-1 lg:flex-none">
            <input type="hidden" name="id" value={id} />
            <input type="hidden" name="depois" value={depois} />
            <Confirmar tecla={TECLA_CONFIRMAR[modo]} rotulo={rotulo} pendente={pendente} />
          </form>
        )}
      </div>
    </>
  )
}

function Confirmar({ tecla, rotulo, pendente }: { tecla: string; rotulo: string; pendente: RefObject<boolean> }) {
  const { pending } = useFormStatus()
  // O atalho chama requestSubmit fora do botão: sem espelhar o estado, dois E seguidos gravariam duas vezes.
  useEffect(() => {
    pendente.current = pending
  }, [pending, pendente])
  return (
    <button type="submit" disabled={pending} className={botao('secundario', 'h-12 w-full lg:h-10 lg:w-auto')}>
      <Tecla>{tecla}</Tecla>
      {pending ? 'Gravando…' : rotulo}
    </button>
  )
}
