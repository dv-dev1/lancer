import assert from 'node:assert/strict'
import { test } from 'node:test'
import { formatarTelefone, linkWhatsApp } from '../lib/abordagem.ts'

test('link do wa.me leva só os dígitos do telefone e o texto codificado', () => {
  assert.equal(linkWhatsApp('+55 (83) 8746-8188', 'Olá, Ana!'), 'https://wa.me/558387468188?text=Ol%C3%A1%2C%20Ana!')
})

test('número sem código do país ganha o 55; número que já vem com o 55 (o coletor) não dobra', () => {
  assert.equal(linkWhatsApp('(83) 99999-0000', ''), 'https://wa.me/5583999990000?text=')
  assert.equal(linkWhatsApp('5583999990000', ''), 'https://wa.me/5583999990000?text=')
})

test('telefone aparece no formato brasileiro', () => {
  assert.equal(formatarTelefone('5583991000137'), '+55 83 99100-0137')
  assert.equal(formatarTelefone('558387468188'), '+55 83 8746-8188')
  assert.equal(formatarTelefone('123'), '123')
})
