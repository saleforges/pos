import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        '@': '/src',
      },
    },
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET,
          changeOrigin: true,
          secure: false,
          rewrite: (path) => path.replace(/^\/api/, ''),
          configure: (proxy) => {
            proxy.on('proxyRes', (proxyRes) => {
              // Strip Secure flag from Set-Cookie in dev (browser is on HTTP).
              // In production, Caddy/ingress handles TLS so Secure works correctly.
              const cookies = proxyRes.headers['set-cookie'];
              if (cookies) {
                proxyRes.headers['set-cookie'] = Array.isArray(cookies)
                  ? cookies.map((c) => c.replace(/;\s*Secure/gi, ''))
                  : cookies.replace(/;\s*Secure/gi, '');
              }
            });
          },
        },
      },
    },
  }
})
