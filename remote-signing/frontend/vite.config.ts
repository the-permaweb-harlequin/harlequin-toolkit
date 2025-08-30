import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import {nodePolyfills} from 'vite-plugin-node-polyfills'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    nodePolyfills({
      include: ['buffer', 'crypto', 'stream', 'util'],
      exclude: ['http'],
      globals: {
        Buffer: true,
        global: true,
        process: true,
      },
      overrides: {
        fs: 'memfs',
      },
      protocolImports: true,
    }),
  ],
  define: {
    global: 'globalThis',
  },
  resolve: {
    alias: {
      // Path alias for components
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    minify: false,
    sourcemap: true,
  },
})
