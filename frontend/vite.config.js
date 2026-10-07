import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    // WebView2 – свежий Chromium: можно не транспилировать в старый JS
    target: 'es2022',
    rollupOptions: {
      output: {
        // React и набор иконок меняются редко – отдельные чанки, основной код меньше
        manualChunks: { react: ['react', 'react-dom'], icons: ['./src/icons.js'] },
      },
    },
  },
  test: { environment: 'node' },
})
