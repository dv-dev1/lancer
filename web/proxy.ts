import { type NextRequest, NextResponse } from 'next/server'
import { COOKIE_SESSAO, validarToken } from './lib/sessao.ts'

export async function proxy(req: NextRequest) {
  const logado = await validarToken(req.cookies.get(COOKIE_SESSAO)?.value)
  const naTelaDeLogin = req.nextUrl.pathname === '/login'
  if (logado && naTelaDeLogin) return NextResponse.redirect(new URL('/', req.url))
  // Action passa sem sessão porque um 307 viraria HTML no lugar de RSC; ela mesma chama exigirSessao. Só POST: num
  // GET o Next ignora o header e renderiza a página, e liberar qualquer método abria o painel sem login.
  const ehAction = req.method === 'POST' && req.headers.has('next-action')
  if (logado || naTelaDeLogin || ehAction) return NextResponse.next()
  const login = new URL('/login', req.url)
  login.searchParams.set('de', req.nextUrl.pathname + req.nextUrl.search)
  return NextResponse.redirect(login)
}

export const config = { matcher: ['/((?!_next/static|_next/image).*)'] }
