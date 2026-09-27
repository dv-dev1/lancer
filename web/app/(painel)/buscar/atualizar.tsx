'use client'

import { useRouter } from 'next/navigation'
import { useEffect } from 'react'

// Enquanto houver pedido esperando ou buscando, relê a página do servidor a cada poucos segundos.
export function Atualizar({ ativo }: { ativo: boolean }) {
  const router = useRouter()
  useEffect(() => {
    if (!ativo) return
    const t = setInterval(() => router.refresh(), 4000)
    return () => clearInterval(t)
  }, [ativo, router])
  return null
}
