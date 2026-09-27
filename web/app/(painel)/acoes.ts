'use server'

import { revalidatePath } from 'next/cache'
import { RedirectType, redirect } from 'next/navigation'
import { sql } from '@/lib/db.ts'
import { exigirSessao } from '@/lib/guarda.ts'
import { validarPedido } from '@/lib/pedidos.ts'
import { destinoSeguro } from '@/lib/sessao.ts'
import { alcancou, ETAPAS } from '@/lib/tipos.ts'

async function desfazerContato(id: string) {
  await sql()`update leads set etapa = 'na_fila', contatado_em = null, follow_up_em = null, respondeu_em = null,
    saiu_de = null where id = ${id}`
}

async function gravarEtapa(id: string, etapa: string) {
  if (!(ETAPAS as readonly string[]).includes(etapa)) throw new Error(`etapa inválida: ${etapa}`)
  if (etapa === 'na_fila') return desfazerContato(id)
  // As datas só se preenchem uma vez (pílula clicada de novo não reseta), e pular degrau preenche os de trás.
  // Na direita do set, `etapa` ainda é a antiga: saiu_de guarda de onde o lead saiu, para o funil.
  await sql()`update leads set etapa = ${etapa},
    contatado_em = case when ${alcancou(etapa, 'contatado')}::boolean then coalesce(contatado_em, now()) else contatado_em end,
    respondeu_em = case when ${alcancou(etapa, 'respondeu')}::boolean then coalesce(respondeu_em, now()) else respondeu_em end,
    saiu_de = case when ${etapa} in ('perdido', 'saiu') and etapa not in ('perdido', 'saiu') then etapa else saiu_de end
    where id = ${id}`
}

export async function mudarEtapa(form: FormData) {
  await exigirSessao()
  await gravarEtapa(String(form.get('id')), String(form.get('etapa')))
  revalidatePath('/', 'layout')
}

export async function marcarEnviado(form: FormData) {
  await exigirSessao()
  await gravarEtapa(String(form.get('id')), 'contatado')
  revalidatePath('/', 'layout')
  // replace: voltar no navegador não deve reabrir o lead que já saiu da fila.
  redirect(destinoSeguro(form.get('depois')), RedirectType.replace)
}

// R22: só marca o envio; quem tira o lead da régua (perdido, 3 dias depois) é o MarcarPerdidos do coletor.
export async function marcarFollowUp(form: FormData) {
  await exigirSessao()
  const id = String(form.get('id'))
  await sql()`update leads set follow_up_em = now() where id = ${id} and respondeu_em is null`
  revalidatePath('/', 'layout')
  redirect(destinoSeguro(form.get('depois') ?? '/cobrar'), RedirectType.replace)
}

export async function pedirColeta(form: FormData) {
  await exigirSessao()
  const r = validarPedido(form)
  if ('erro' in r) redirect(`/buscar?erro=${encodeURIComponent(r.erro)}`, RedirectType.replace)
  const { cidade, bairro, nicho, limite } = r.pedido
  await sql()`insert into pedidos (cidade, bairro, nicho, limite) values (${cidade}, ${bairro}, ${nicho}, ${limite})`
  revalidatePath('/buscar')
  redirect('/buscar', RedirectType.replace)
}
