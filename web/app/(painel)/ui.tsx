const VARIANTES = {
  primario: 'bg-acento text-tinta-acento hover:bg-acento/85',
  secundario: 'border border-linha-forte text-texto hover:border-apagado hover:bg-elevada',
  discreto: 'text-suave hover:bg-elevada hover:text-texto',
} as const

export type VarianteBotao = keyof typeof VARIANTES

// A altura vem no `extra` para o chamador trocar sem brigar com uma classe h-* fixa aqui.
export function botao(variante: VarianteBotao, extra = 'h-10'): string {
  return `inline-flex items-center justify-center gap-2.5 whitespace-nowrap rounded-md px-4 text-sm font-medium transition-colors duration-150 disabled:cursor-not-allowed disabled:opacity-50 ${VARIANTES[variante]} ${extra}`
}

// No celular não há teclado físico: a tecla só aparece do lg para cima, salvo quando `sempre`.
export function Tecla({ children, sempre = false }: { children: React.ReactNode; sempre?: boolean }) {
  return (
    <kbd
      className={`${sempre ? 'inline-flex' : 'hidden lg:inline-flex'} h-5 min-w-5 items-center justify-center rounded-[4px] border border-current/30 px-1 font-mono font-normal text-xs leading-none`}
    >
      {children}
    </kbd>
  )
}

export function Etiqueta({ children, id }: { children: React.ReactNode; id?: string }) {
  return (
    <h2 id={id} className="font-medium text-[11px] text-apagado uppercase tracking-[0.08em]">
      {children}
    </h2>
  )
}

const COR_ETAPA: Record<string, string> = {
  fechado: 'border-positivo/40 text-positivo',
  perdido: 'border-linha text-apagado',
  saiu: 'border-linha text-apagado',
}

export function Cabecalho({ titulo, descricao }: { titulo: string; descricao?: string }) {
  return (
    <header className="space-y-1">
      <h1 className="font-semibold text-2xl tracking-[-0.02em]">{titulo}</h1>
      {descricao && <p className="text-sm text-suave">{descricao}</p>}
    </header>
  )
}

export function SeloEtapa({ etapa, rotulo }: { etapa: string; rotulo: string }) {
  return <Selo cor={COR_ETAPA[etapa]}>{rotulo}</Selo>
}

export function Selo({ children, cor }: { children: React.ReactNode; cor?: string }) {
  return (
    <span className={`inline-block rounded-[4px] border px-2 py-0.5 text-xs ${cor ?? 'border-linha-forte text-suave'}`}>
      {children}
    </span>
  )
}
