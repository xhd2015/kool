import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    // to avoid issue like:
    //   Cannot read properties of null (reading 'useContext')
    dedupe: [
      "react",
      "react-dom",
      "react-router",
      "react-router-dom",
    ],
    alias: {
      '@external_src': fileURLToPath(new URL('./external_src', import.meta.url)),
    },
  },
  server: {
    strictPort: true,
  },
})
