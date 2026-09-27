import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { test } from 'node:test'

const app = new URL('../app/', import.meta.url)
const css = readFileSync(new URL('globals.css', app), 'utf8')
const cores: Record<string, string> = Object.fromEntries(
  [...css.matchAll(/--color-([a-z-]+):\s*(#[0-9a-f]{6})/g)].map(([, nome, hex]) => [nome, hex]),
)

function luminancia(hex: string): number {
  const [r, g, b] = [1, 3, 5].map((i) => {
    const c = Number.parseInt(hex.slice(i, i + 2), 16) / 255
    return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

function contraste(a: string, b: string): number {
  const [claro, escuro] = [luminancia(a), luminancia(b)].sort((x, y) => y - x)
  return (claro + 0.05) / (escuro + 0.05)
}

// apagado sobre elevada dá ~4,35:1: metadado em superfície elevada usa suave.
const PARES: [string, string[]][] = [
  ...['texto', 'suave', 'acento', 'positivo', 'aviso'].map((t): [string, string[]] => [
    t,
    ['fundo', 'superficie', 'elevada'],
  ]),
  ['tinta-acento', ['acento']],
  ['apagado', ['fundo', 'superficie']],
]

test('todo par de texto e fundo da paleta passa AA (4,5:1)', () => {
  for (const [texto, fundos] of PARES) {
    for (const fundo of fundos) {
      assert.ok(cores[texto] && cores[fundo], `--color-${texto} ou --color-${fundo} ausente do globals.css`)
      const c = contraste(cores[texto], cores[fundo])
      assert.ok(c >= 4.5, `${texto} sobre ${fundo}: ${c.toFixed(2)}:1`)
    }
  }
})

test('o CSS não tem gradiente, blur nem vidro', () => {
  assert.doesNotMatch(css, /gradient\(|blur\(|backdrop/)
})

test('nenhuma tela usa utilitário de gradiente, blur ou vidro', () => {
  const telas = readdirSync(app, { recursive: true, encoding: 'utf8' }).filter((f) => f.endsWith('.tsx'))
  for (const tela of telas) {
    assert.doesNotMatch(
      readFileSync(new URL(tela, app), 'utf8'),
      /\bbg-(?:linear|radial|conic|gradient)-|\bblur-|backdrop-/,
      tela,
    )
  }
})
