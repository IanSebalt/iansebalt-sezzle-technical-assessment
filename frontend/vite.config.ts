import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// Dev requests are proxied so the browser always talks to a same-origin /api, exactly as it does
// in the container where nginx fronts the backend. That keeps CORS out of the picture entirely.
const backendUrl = process.env.BACKEND_URL ?? 'http://localhost:9080'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': { target: backendUrl, changeOrigin: true },
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/tests/setup.ts'],
    include: ['src/tests/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'text-summary'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/main.tsx', 'src/tests/**', 'src/**/*.d.ts'],
    },
  },
})
