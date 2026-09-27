import type { Metadata, Viewport } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'
import './globals.css'

const geist = Geist({ subsets: ['latin'], variable: '--font-geist' })
const geistMono = Geist_Mono({ subsets: ['latin'], variable: '--font-geist-mono' })

export const metadata: Metadata = { title: { template: '%s · lancer', default: 'lancer' } }
export const viewport: Viewport = { themeColor: '#0f0f0e', colorScheme: 'dark', viewportFit: 'cover' }

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="pt-BR" className={`${geist.variable} ${geistMono.variable}`}>
      <body className="min-h-dvh bg-fundo font-sans text-texto antialiased">{children}</body>
    </html>
  )
}
