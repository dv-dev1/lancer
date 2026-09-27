export type Modo = 'fila' | 'cobrar' | 'ficha'
export type AcaoTecla = 'abrir' | 'confirmar' | 'proximo' | 'anterior'
export type EventoTecla = { key: string; tag: string; editavel: boolean; modificador: boolean; repetida: boolean }

const CAMPOS = new Set(['INPUT', 'TEXTAREA', 'SELECT'])
// ↵ num link ou botão já é o clique dele; abrir o WhatsApp junto dispararia duas ações.
const CLICAVEIS = new Set(['A', 'BUTTON'])
export const TECLA_CONFIRMAR: Record<Modo, string> = { fila: 'E', cobrar: 'F', ficha: 'E' }

export function acaoDaTecla(t: EventoTecla, modo: Modo): AcaoTecla | null {
  if (t.repetida || t.modificador || t.editavel || CAMPOS.has(t.tag)) return null
  if (t.key === 'Enter') return CLICAVEIS.has(t.tag) ? null : 'abrir'
  if (t.key.toUpperCase() === TECLA_CONFIRMAR[modo]) return 'confirmar'
  if (modo === 'ficha') return null
  if (t.key === 'ArrowRight') return 'proximo'
  if (t.key === 'ArrowLeft') return 'anterior'
  return null
}
