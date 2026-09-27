import assert from 'node:assert/strict'
import { test } from 'node:test'
import {
  AMOSTRA_MINIMA,
  amostraPequena,
  custoPorFechado,
  porDor,
  porNicho,
  porVariante,
  taxa,
  totais,
} from '../lib/metricas.ts'

test('taxa arredonda a porcentagem e não divide por zero', () => {
  assert.equal(taxa(1, 4), '25%')
  assert.equal(taxa(0, 0), '—')
})

test('cortes só contam quem foi contatado; sem contato não entra nem no n', () => {
  const leads = [
    {
      variante: 'texto',
      nicho: 'confeitaria',
      dores: ['sem_site'],
      contatado_em: '2026-01-01',
      respondeu_em: '2026-01-02',
    },
    { variante: 'texto', nicho: 'confeitaria', dores: ['sem_site'], contatado_em: '2026-01-01', respondeu_em: null },
    { variante: 'link', nicho: 'salao', dores: ['site_ruim'], contatado_em: null, respondeu_em: null },
  ]
  assert.deepEqual(porVariante(leads), [{ chave: 'texto', n: 2, respostas: 1, taxa: '50%' }])
})

test('lead com duas dores conta nos dois cortes de dor', () => {
  const leads = [
    {
      variante: 'texto',
      nicho: 'confeitaria',
      dores: ['sem_site', 'reclamacao'],
      contatado_em: '2026-01-01',
      respondeu_em: '2026-01-02',
    },
  ]
  assert.deepEqual(porDor(leads), [
    { chave: 'sem_site', n: 1, respostas: 1, taxa: '100%' },
    { chave: 'reclamacao', n: 1, respostas: 1, taxa: '100%' },
  ])
})

test('porNicho agrupa pelo nicho do lead', () => {
  const leads = [{ variante: 'texto', nicho: 'confeitaria', dores: [], contatado_em: '2026-01-01', respondeu_em: null }]
  assert.deepEqual(porNicho(leads), [{ chave: 'confeitaria', n: 1, respostas: 0, taxa: '0%' }])
})

test('custo por fechado dá — sem nenhum lead fechado, e divide certo quando há', () => {
  assert.equal(custoPorFechado(10, 0), '—')
  assert.equal(
    custoPorFechado(10, 2),
    (5).toLocaleString('pt-BR', { style: 'currency', currency: 'USD', minimumFractionDigits: 4 }),
  )
})

test('totais contam contatados e respondidos, com a taxa sobre os contatados', () => {
  const leads = [
    { variante: 'texto', nicho: 'confeitaria', dores: [], contatado_em: '2026-01-01', respondeu_em: '2026-01-02' },
    { variante: 'link', nicho: 'confeitaria', dores: [], contatado_em: '2026-01-01', respondeu_em: null },
    { variante: 'link', nicho: 'salao', dores: [], contatado_em: '2026-01-03', respondeu_em: null },
    { variante: 'texto', nicho: 'salao', dores: [], contatado_em: null, respondeu_em: null },
  ]
  assert.deepEqual(totais(leads), { contatados: 3, respondidos: 1, taxa: '33%' })
})

test('resposta sem contato registrado não entra nos totais, como nos cortes', () => {
  const leads = [{ variante: 'texto', nicho: 'salao', dores: [], contatado_em: null, respondeu_em: '2026-01-02' }]
  assert.deepEqual(totais(leads), { contatados: 0, respondidos: 0, taxa: '—' })
})

test('amostra é pequena abaixo de 10 contatados e deixa de ser a partir de 10', () => {
  assert.equal(AMOSTRA_MINIMA, 10)
  assert.equal(amostraPequena(0), true)
  assert.equal(amostraPequena(9), true)
  assert.equal(amostraPequena(10), false)
})
