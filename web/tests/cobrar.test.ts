import assert from 'node:assert/strict'
import { test } from 'node:test'
import { deveCobrar } from '../lib/cobrar.ts'

const agora = new Date('2026-09-27T12:00:00Z')
const base = { contatado_em: '2026-09-24T12:00:00Z', respondeu_em: null, follow_up_em: null }

test('3 dias exatos sem resposta e sem follow-up já pode cobrar', () => {
  assert.equal(deveCobrar(base, agora), true)
})

test('menos de 3 dias ainda não cobra', () => {
  assert.equal(deveCobrar({ ...base, contatado_em: '2026-09-25T12:00:00Z' }, agora), false)
})

test('quem já respondeu não entra na régua de cobrança', () => {
  assert.equal(deveCobrar({ ...base, respondeu_em: '2026-09-25T00:00:00Z' }, agora), false)
})

test('follow-up já enviado sai da régua, pra não cobrar duas vezes', () => {
  assert.equal(deveCobrar({ ...base, follow_up_em: '2026-09-25T00:00:00Z' }, agora), false)
})

test('sem contatado_em (ainda na fila) não entra na régua', () => {
  assert.equal(deveCobrar({ ...base, contatado_em: null }, agora), false)
})
