import type { NextConfig } from 'next'

// O AGENTS.md gerado é sobre o Next em si, não sobre o Lancer, e duplicaria o AGENTS.md da raiz do repo.
const nextConfig: NextConfig = { agentRules: false }

export default nextConfig
