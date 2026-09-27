import assert from 'node:assert/strict'
import { test } from 'node:test'
import { acaoDaTecla, type EventoTecla } from '../lib/teclas.ts'

const tecla = (key: string, resto: Partial<EventoTecla> = {}): EventoTecla => ({
  key,
  tag: 'BODY',
  editavel: false,
  modificador: false,
  repetida: false,
  ...resto,
})

test('↵ abre o WhatsApp e as setas andam pela fila', () => {
  assert.equal(acaoDaTecla(tecla('Enter'), 'fila'), 'abrir')
  assert.equal(acaoDaTecla(tecla('ArrowRight'), 'fila'), 'proximo')
  assert.equal(acaoDaTecla(tecla('ArrowLeft'), 'cobrar'), 'anterior')
})

test('E confirma na fila e na ficha; F confirma no cobrar; maiúscula vale igual', () => {
  assert.equal(acaoDaTecla(tecla('e'), 'fila'), 'confirmar')
  assert.equal(acaoDaTecla(tecla('E'), 'fila'), 'confirmar')
  assert.equal(acaoDaTecla(tecla('e'), 'ficha'), 'confirmar')
  assert.equal(acaoDaTecla(tecla('f'), 'cobrar'), 'confirmar')
  assert.equal(acaoDaTecla(tecla('F'), 'cobrar'), 'confirmar')
})

test('a tecla de um modo não confirma no outro', () => {
  assert.equal(acaoDaTecla(tecla('f'), 'fila'), null)
  assert.equal(acaoDaTecla(tecla('e'), 'cobrar'), null)
})

test('a ficha não é lista: as setas não fazem nada', () => {
  assert.equal(acaoDaTecla(tecla('ArrowRight'), 'ficha'), null)
  assert.equal(acaoDaTecla(tecla('ArrowLeft'), 'ficha'), null)
})

test('digitar dentro de campo não dispara nada, nem o "e" da mensagem', () => {
  for (const tag of ['TEXTAREA', 'INPUT', 'SELECT']) {
    assert.equal(acaoDaTecla(tecla('e', { tag }), 'fila'), null)
    assert.equal(acaoDaTecla(tecla('Enter', { tag }), 'fila'), null)
    assert.equal(acaoDaTecla(tecla('ArrowRight', { tag }), 'fila'), null)
  }
  assert.equal(acaoDaTecla(tecla('e', { tag: 'DIV', editavel: true }), 'fila'), null)
})

test('atalho do sistema e tecla segurada não disparam', () => {
  assert.equal(acaoDaTecla(tecla('e', { modificador: true }), 'fila'), null)
  assert.equal(acaoDaTecla(tecla('ArrowRight', { modificador: true }), 'fila'), null)
  assert.equal(acaoDaTecla(tecla('ArrowRight', { repetida: true }), 'fila'), null)
})

test('↵ com foco em link ou botão fica com o navegador; as letras seguem valendo', () => {
  assert.equal(acaoDaTecla(tecla('Enter', { tag: 'A' }), 'fila'), null)
  assert.equal(acaoDaTecla(tecla('Enter', { tag: 'BUTTON' }), 'fila'), null)
  assert.equal(acaoDaTecla(tecla('e', { tag: 'BUTTON' }), 'fila'), 'confirmar')
})

test('tecla sem ação devolve null', () => {
  assert.equal(acaoDaTecla(tecla('x'), 'fila'), null)
  assert.equal(acaoDaTecla(tecla(' '), 'fila'), null)
})
