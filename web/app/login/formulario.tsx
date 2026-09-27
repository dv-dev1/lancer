'use client'

import { useActionState } from 'react'
import { entrar } from './acoes.ts'

export function Formulario({ de }: { de: string }) {
  const [tentativa, acao, enviando] = useActionState(entrar, null)
  return (
    <form action={acao} className="space-y-4">
      <input type="hidden" name="de" value={de} />
      {/* O React limpa o form depois da action; quem errou a senha não precisa redigitar o usuário. */}
      <Campo nome="usuario" rotulo="Usuário" tipo="text" autocomplete="username" inicial={tentativa?.usuario} />
      <Campo nome="senha" rotulo="Senha" tipo="password" autocomplete="current-password" />
      {tentativa && (
        <p role="alert" className="surgir text-sm text-quente">
          {tentativa.erro}
        </p>
      )}
      <button
        type="submit"
        disabled={enviando}
        className="botao-primario w-full py-2.5 text-sm font-medium transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
      >
        {enviando ? 'Entrando…' : 'Entrar'}
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
    <label className="block space-y-1.5">
      <span className="text-xs uppercase tracking-widest text-suave">{rotulo}</span>
      <input
        name={nome}
        type={tipo}
        autoComplete={autocomplete}
        defaultValue={inicial}
        required
        className="w-full rounded-xl border border-borda bg-fundo/60 px-4 py-2.5 text-sm text-white outline-none transition-colors duration-150 focus:border-ciano/60"
      />
    </label>
  )
}
