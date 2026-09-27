import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import { COOKIE_SESSAO, validarToken } from './sessao.ts'

// Server action é chamada pelo id, de qualquer rota: toda action que grava começa por aqui.
export async function exigirSessao(): Promise<string> {
  const usuario = await validarToken((await cookies()).get(COOKIE_SESSAO)?.value)
  if (!usuario) redirect('/login')
  return usuario
}
