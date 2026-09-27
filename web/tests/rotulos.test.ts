import assert from 'node:assert/strict'
import { test } from 'node:test'
import { funil } from '../lib/rotulos.ts'

test('funil conta quem chegou pelo menos até a etapa, a conversão sobre a anterior, e soma as saídas', () => {
  const f = funil({
    na_fila: 2,
    contatado: 3,
    abriu: 1,
    respondeu: 2,
    interessado: 1,
    proposta: 1,
    fechado: 2,
    perdido: 3,
    saiu: 1,
  })
  assert.deepEqual(
    f.degraus.map((d) => [d.rotulo, d.n, d.conversao]),
    [
      ['Na fila', 16, 100],
      ['Contatado', 10, 63],
      ['Abriu o link', 7, 70],
      ['Respondeu', 6, 86],
      ['Interessado', 4, 67],
      ['Proposta', 3, 75],
      ['Fechado', 2, 67],
    ],
  )
  assert.equal(f.saidas, 4)
})

test('funil vazio não divide por zero', () => {
  assert.deepEqual(
    funil({}).degraus.map((d) => d.conversao),
    [0, 0, 0, 0, 0, 0, 0],
  )
})
