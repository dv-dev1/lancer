import { type NextRequest, NextResponse } from 'next/server'
import { COOKIE_SESSAO, validarToken } from './lib/sessao.ts'

export async function proxy(req: NextRequest) {
  const logado = await validarToken(req.cookies.get(COOKIE_SESSAO)?.value)
  const naTelaDeLogin = req.nextUrl.pathname === '/login'
  if (logado && naTelaDeLogin) return NextResponse.redirect(new URL('/', req.url))
  // Action com sessão vencida: um 307 viraria HTML no lugar de RSC e erro na tela. A própria action checa a sessão
  // (exigirSessao) e manda ao login por um redirect que o cliente do Next entende.
  if (logado || naTelaDeLogin || req.headers.has('next-action')) return NextResponse.next()
  const login = new URL('/login', req.url)
  login.searchParams.set('de', req.nextUrl.pathname + req.nextUrl.search)
  return NextResponse.redirect(login)
}

export const config = { matcher: ['/((?!_next/static|_next/image).*)'] }
