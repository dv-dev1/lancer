import { type NeonQueryFunction, neon } from '@neondatabase/serverless'

let conexao: NeonQueryFunction<false, false> | undefined

// Preguiçoso: o next build importa este módulo sem DATABASE_URL.
export function sql(): NeonQueryFunction<false, false> {
  conexao ??= neon(process.env.DATABASE_URL ?? '')
  return conexao
}
