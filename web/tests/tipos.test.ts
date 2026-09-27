import assert from 'node:assert/strict'
import { test } from 'node:test'
import { alcancou } from '../lib/tipos.ts'

test('clicar no próprio degrau marca ele', () => {
  assert.equal(alcancou('contatado', 'contatado'), true)
  assert.equal(alcancou('respondeu', 'respondeu'), true)
})

test('pular direto para um degrau adiante conta os degraus anteriores', () => {
  assert.equal(alcancou('interessado', 'respondeu'), true)
  assert.equal(alcancou('interessado', 'contatado'), true)
  assert.equal(alcancou('respondeu', 'contatado'), true)
  assert.equal(alcancou('fechado', 'respondeu'), true)
})

test('degrau anterior, na fila e saída não marcam contato nem resposta', () => {
  assert.equal(alcancou('contatado', 'respondeu'), false)
  assert.equal(alcancou('na_fila', 'contatado'), false)
  assert.equal(alcancou('perdido', 'contatado'), false)
  assert.equal(alcancou('saiu', 'respondeu'), false)
})
