import { ESCADA, type Etapa } from './tipos.ts'

// Aba pedida e válida vale mesmo vazia (o usuário escolheu ver aquele degrau). Sem pedido ou
// pedido inválido, o quadro abre onde tem gente esperando; sem ninguém em lugar nenhum, na fila.
export function abaAtual(pedida: string | undefined, contagens: Record<string, number>): Etapa {
  if (pedida && (ESCADA as readonly string[]).includes(pedida)) return pedida as Etapa
  return ESCADA.find((e) => contagens[e] > 0) ?? 'na_fila'
}
