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
      dedupe: ['react', 'react-dom'], // Force Vite to use the root installation for these packages
    },
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET,
          changeOrigin: true,
          secure: false,
          rewrite: (path) => path.replace(/^\/api/, ''),
          configure: (proxy) => {
            proxy.on('proxyReq', (proxyReq) => {
              // Strip Origin header so backend sees a same-origin request,
              // avoiding CORS 403 when localhost:5173 isn't in its allow-list.
              proxyReq.removeHeader('origin');
            });
            proxy.on('proxyRes', (proxyRes) => {
              // Strip Secure flag from Set-Cookie in dev (browser is on HTTP).
              // In production, Caddy/ingress handles TLS so Secure works correctly.
              const cookies = proxyRes.headers['set-cookie'];
              if (cookies) {
                proxyRes.headers['set-cookie'] = cookies.map((c) =>
                  c.replace(/;\s*Secure/gi, ''),
                );
              }
            });
          },
        },
      },
    },
  }
})
