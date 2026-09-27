import assert from 'node:assert/strict'
import { test } from 'node:test'
import { lugarDoPedido, validarPedido } from '../lib/pedidos.ts'

const form = (campos: Record<string, string>) => {
  const f = new FormData()
  for (const [k, v] of Object.entries(campos)) f.set(k, v)
  return f
}

test('pedido válido sai aparado, com bairro opcional', () => {
  assert.deepEqual(
    validarPedido(form({ cidade: '  Campina Grande ', bairro: ' ', nicho: 'confeitaria', limite: '2' })),
    {
      pedido: { cidade: 'Campina Grande', bairro: '', nicho: 'confeitaria', limite: 2 },
    },
  )
})

test('recusa nicho fora da lista, limite fora de 1 a 20 e cidade vazia ou estranha', () => {
  const base = { cidade: 'João Pessoa', bairro: 'Manaíra', nicho: 'loja', limite: '5' }
  for (const ruim of [
    { nicho: 'padaria' },
    { nicho: 'toString' },
    { limite: '0' },
    { limite: '21' },
    { limite: '2.5' },
    { cidade: '' },
    { cidade: 'x'.repeat(61) },
    { bairro: '<script>' },
  ]) {
    assert.ok('erro' in validarPedido(form({ ...base, ...ruim })), JSON.stringify(ruim))
  }
})

test('lugar do pedido igual ao da busca no Maps', () => {
  assert.equal(lugarDoPedido('Campina Grande', ''), 'Campina Grande')
  assert.equal(lugarDoPedido('João Pessoa', 'Manaíra'), 'Manaíra, João Pessoa')
})
