import { defineConfig } from 'vite'
import path from 'path'
import vue from '@vitejs/plugin-vue'


// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '/static': path.resolve(__dirname, './public/static'),
      // Only including mediaplayer to reduce bundle size.
      'dashjs': require.resolve('dashjs/dist/dash.mediaplayer.min.js')
    },
  },
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 80,
    hmr: {
        clientPort: 80,
    },
    fs: {
      allow: ['..'], // Allow serving static content from imported pacakges.
    },
  },
  build: {
    chunkSizeWarningLimit: 1500,
  }
})
