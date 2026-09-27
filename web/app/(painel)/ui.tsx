const COR_ETAPA: Record<string, string> = {
  fechado: 'border-[#08fbd0]/40 text-[#08fbd0]',
  perdido: 'border-borda text-suave',
  saiu: 'border-borda text-suave',
}

export function Cabecalho({ titulo, descricao }: { titulo: string; descricao?: string }) {
  return (
    <header className="surgir space-y-2">
      <h1 className="titulo-gradiente text-4xl font-medium tracking-[-0.05em] sm:text-5xl">{titulo}</h1>
      {descricao && <p className="text-sm text-suave">{descricao}</p>}
    </header>
  )
}

export function SeloEtapa({ etapa, rotulo }: { etapa: string; rotulo: string }) {
  return <Selo cor={COR_ETAPA[etapa]}>{rotulo}</Selo>
}

export function Selo({ children, cor }: { children: React.ReactNode; cor?: string }) {
  return (
    <span className={`inline-block rounded-full border px-2.5 py-0.5 text-xs ${cor ?? 'border-borda text-texto'}`}>
      {children}
    </span>
  )
}
