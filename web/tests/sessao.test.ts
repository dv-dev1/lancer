import assert from 'node:assert/strict'
import { test } from 'node:test'
import { SignJWT } from 'jose'
import { credenciaisValidas, criarToken, destinoSeguro, validarToken } from '../lib/sessao.ts'

process.env.AUTH_SECRET = 'segredo-de-teste-com-tamanho-suficiente'
process.env.DASHBOARD_USER = 'lancer'
process.env.DASHBOARD_PASSWORD = 'senha-certa'

const assinar = (segredo: string, exp: string | number) =>
  new SignJWT()
    .setProtectedHeader({ alg: 'HS256' })
    .setSubject('lancer')
    .setExpirationTime(exp)
    .sign(new TextEncoder().encode(segredo))

test('token criado no login abre o painel', async () => {
  assert.equal(await validarToken(await criarToken('lancer')), 'lancer')
})

test('token adulterado, de outro segredo, vencido ou sem assinatura não abre', async () => {
  const bom = await criarToken('lancer')
  const [cabecalho, corpo, assinatura] = bom.split('.')
  const outroCorpo = Buffer.from(JSON.stringify({ sub: 'admin', exp: 9999999999 })).toString('base64url')
  const semAssinatura = `${Buffer.from('{"alg":"none"}').toString('base64url')}.${corpo}.`
  assert.equal(await validarToken(`${cabecalho}.${outroCorpo}.${assinatura}`), null)
  assert.equal(await validarToken(await assinar('outro-segredo-qualquer-bem-comprido', '7d')), null)
  assert.equal(
    await validarToken(await assinar(process.env.AUTH_SECRET as string, Math.floor(Date.now() / 1000) - 60)),
    null,
  )
  assert.equal(await validarToken(semAssinatura), null)
  assert.equal(await validarToken(undefined), null)
})

test('sem AUTH_SECRET configurado, nada abre', async () => {
  const token = await criarToken('lancer')
  const segredo = process.env.AUTH_SECRET
  delete process.env.AUTH_SECRET
  try {
    assert.equal(await validarToken(token), null)
  } finally {
    process.env.AUTH_SECRET = segredo
  }
})

test('credenciais certas abrem; usuário ou senha errados, ou sem variável de ambiente, não', () => {
  assert.equal(credenciaisValidas('lancer', 'senha-certa'), 'lancer')
  assert.equal(credenciaisValidas('lancer', 'senha-errada'), null)
  assert.equal(credenciaisValidas('outro', 'senha-certa'), null)
  assert.equal(credenciaisValidas('lancer', ''), null)
})

test('depois do login só volta para caminho do próprio painel', () => {
  assert.equal(destinoSeguro('/leads?etapa=na_fila'), '/leads?etapa=na_fila')
  assert.equal(destinoSeguro('//evil.com'), '/')
  assert.equal(destinoSeguro('/\\evil.com'), '/')
  assert.equal(destinoSeguro('/\t/evil.com'), '/')
  assert.equal(destinoSeguro('/\n/evil.com'), '/')
  assert.equal(destinoSeguro('https://evil.com'), '/')
  assert.equal(destinoSeguro(null), '/')
})
