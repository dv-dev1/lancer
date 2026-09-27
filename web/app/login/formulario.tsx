'use client'

import { useActionState } from 'react'
import { Tecla } from '../(painel)/ui.tsx'
import { entrar } from './acoes.ts'

export function Formulario({ de }: { de: string }) {
  const [tentativa, acao, enviando] = useActionState(entrar, null)
  return (
    <form action={acao} className="space-y-8">
      <input type="hidden" name="de" value={de} />
      {/* O React limpa o form depois da action; quem errou a senha não precisa redigitar o usuário. */}
      <Campo nome="usuario" rotulo="Usuário" tipo="text" autocomplete="username" inicial={tentativa?.usuario} />
      <Campo nome="senha" rotulo="Senha" tipo="password" autocomplete="current-password" />
      {tentativa && (
        <p role="alert" className="text-aviso text-sm">
          {tentativa.erro}
        </p>
      )}
      <button
        type="submit"
        disabled={enviando}
        className="flex h-11 w-full items-center justify-between rounded-md bg-acento px-4 font-medium text-sm text-tinta-acento transition-colors duration-150 hover:bg-acento/85 disabled:cursor-wait disabled:opacity-60"
      >
        {enviando ? 'Entrando…' : 'Entrar'}
        <Tecla sempre>↵</Tecla>
      </button>
    </form>
  )
}

function Campo({
  nome,
  rotulo,
  tipo,
  autocomplete,
  inicial,
}: {
  nome: string
  rotulo: string
  tipo: string
  autocomplete: string
  inicial?: string
}) {
  return (
    <label className="block space-y-1">
      <span className="block font-medium text-[11px] text-apagado uppercase tracking-[0.08em]">{rotulo}</span>
      {/* A sombra interna engrossa a linha no foco sem mexer na altura do campo. */}
      <input
        name={nome}
        type={tipo}
        autoComplete={autocomplete}
        defaultValue={inicial}
        required
        className="block w-full rounded-none border-linha-forte border-b bg-transparent px-0 py-2 text-base text-texto transition-colors duration-150 focus:border-acento focus:shadow-[inset_0_-1px_0_var(--color-acento)] focus:outline-hidden"
      />
    </label>
  )
}
