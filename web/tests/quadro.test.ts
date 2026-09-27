import assert from 'node:assert/strict'
import { test } from 'node:test'
import { abaAtual } from '../lib/quadro.ts'

test('aba pedida e válida vence mesmo vazia', () => {
  assert.equal(abaAtual('proposta', { proposta: 0, na_fila: 5 }), 'proposta')
})

test('aba ausente ou inválida cai na primeira etapa do funil com leads', () => {
  assert.equal(abaAtual(undefined, { na_fila: 0, contatado: 3 }), 'contatado')
  assert.equal(abaAtual('etapa-que-nao-existe', { na_fila: 0, contatado: 3 }), 'contatado')
})

test('sem etapa nenhuma com leads, cai em na_fila', () => {
  assert.equal(abaAtual(undefined, {}), 'na_fila')
})
