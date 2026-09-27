import assert from 'node:assert/strict'
import { test } from 'node:test'
import { deveCobrar, proximaCobranca, rotuloDias } from '../lib/cobrar.ts'

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

test('dias desde o contato contam só dias inteiros', () => {
  assert.equal(rotuloDias('2026-09-27T02:00:00Z', agora), 'hoje')
  assert.equal(rotuloDias('2026-09-26T12:00:00Z', agora), 'há 1 dia')
  assert.equal(rotuloDias(new Date('2026-09-23T13:00:00Z'), agora), 'há 3 dias')
})

test('próxima cobrança é a do contato mais antigo que ainda não fez 3 dias', () => {
  const candidatos = [
    { ...base, contatado_em: '2026-09-26T08:00:00Z' },
    { ...base, contatado_em: '2026-09-25T10:30:00Z' },
    { ...base, contatado_em: '2026-09-27T09:00:00Z' },
  ]
  assert.deepEqual(proximaCobranca(candidatos, agora), new Date('2026-09-28T10:30:00Z'))
})

test('quem já pode ser cobrado, respondeu ou recebeu follow-up não marca a próxima cobrança', () => {
  const candidatos = [
    base,
    { ...base, contatado_em: '2026-09-26T00:00:00Z', respondeu_em: '2026-09-26T05:00:00Z' },
    { ...base, contatado_em: '2026-09-26T00:00:00Z', follow_up_em: '2026-09-26T05:00:00Z' },
    { ...base, contatado_em: null },
  ]
  assert.equal(proximaCobranca(candidatos, agora), null)
  assert.equal(proximaCobranca([], agora), null)
})
