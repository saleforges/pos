// clients/pos/next.config.ts
import type { NextConfig } from 'next'
import path from 'path'

const nextConfig: NextConfig = {
  output: 'standalone',
  turbopack: {
    root: path.join(__dirname, '..'),
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.API_PROXY_TARGET ?? 'http://localhost:8080'}/:path*`,
      },
    ]
  },
}

export default nextConfig
