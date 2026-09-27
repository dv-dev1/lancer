import assert from 'node:assert/strict'
import { test } from 'node:test'
import { seguinte, vizinhos } from '../lib/mesa.ts'

const ids = ['7', '3', '12', '5']

test('lead do meio mostra a posição e os dois vizinhos', () => {
  assert.deepEqual(vizinhos(ids, '3'), { atual: '3', posicao: 2, total: 4, anterior: '7', proximo: '12' })
})

test('a fila não dá a volta: o primeiro não tem anterior e o último não tem próximo', () => {
  assert.deepEqual(vizinhos(ids, '7'), { atual: '7', posicao: 1, total: 4, anterior: null, proximo: '3' })
  assert.deepEqual(vizinhos(ids, '5'), { atual: '5', posicao: 4, total: 4, anterior: '12', proximo: null })
})

test('lead ausente, inválido, repetido na URL ou fora da lista abre o primeiro', () => {
  const primeiro = { atual: '7', posicao: 1, total: 4, anterior: null, proximo: '3' }
  assert.deepEqual(vizinhos(ids, undefined), primeiro)
  assert.deepEqual(vizinhos(ids, 'abc'), primeiro)
  assert.deepEqual(vizinhos(ids, ['3', '12']), primeiro)
  assert.deepEqual(vizinhos(ids, '999'), primeiro)
})

test('fila vazia não tem lead atual', () => {
  assert.equal(vizinhos([], '3'), null)
})

test('depois de enviar abre o próximo; no fim da fila, o anterior; sozinho, nenhum', () => {
  assert.equal(seguinte(ids, '3'), '12')
  assert.equal(seguinte(ids, '5'), '12')
  assert.equal(seguinte(['9'], '9'), null)
  assert.equal(seguinte([], '9'), null)
})
