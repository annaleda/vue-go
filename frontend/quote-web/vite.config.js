import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Vite fa due mestieri diversi:
//   - in sviluppo serve i file al browser senza impacchettarli, quindi parte
//     in un istante e ricarica solo il modulo che hai toccato
//   - in build produce i file statici ottimizzati in dist/
//
// È il successore di webpack (che è quello usato nei micro-frontend Poste).
export default defineConfig({
  plugins: [vue()],

  server: {
    port: 5173,
    // Il proxy evita i problemi di CORS in sviluppo: il browser vede una
    // sola origine (localhost:5173) e Vite gira le chiamate /api al backend.
    //
    // È lo stesso ruolo che ha proxy.conf.json nei progetti Angular, e in
    // produzione lo stesso lavoro lo fa nginx o il gateway.
    proxy: {
      '/api/preventivi': {
        target: process.env.QUOTE_URL || 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/preventivi/, '/api/v1/preventivi'),
      },
      '/api/garanzie': {
        target: process.env.CATALOG_URL || 'http://localhost:8081',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/garanzie/, '/api/v1/garanzie'),
      },
    },
  },
})
