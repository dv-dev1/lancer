// O coletor já salva o telefone com o 55 (internal/lead.Celular); um número digitado à mão pode vir
// sem ele. Só completa quando o tamanho é de DDD + número (10 ou 11), pra nunca dobrar o prefixo.
function internacional(telefone: string): string {
  const d = telefone.replace(/\D/g, '')
  return d.length === 10 || d.length === 11 ? `55${d}` : d
}

export function linkWhatsApp(telefone: string, texto: string): string {
  return `https://wa.me/${internacional(telefone)}?text=${encodeURIComponent(texto)}`
}

export function formatarTelefone(telefone: string): string {
  const m = internacional(telefone).match(/^55(\d{2})(\d{4,5})(\d{4})$/)
  return m ? `+55 ${m[1]} ${m[2]}-${m[3]}` : telefone
}
