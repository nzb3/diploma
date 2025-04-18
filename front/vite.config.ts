import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@components': '/src/components',
      '@': '/src',
      '@views': '/src/views',
      '@layouts': '/src/layouts',
      '@pages': '/src/pages',
      '@services': '/src/services',
      '@types': '/src/types',
      '@utils': '/src/utils'
    }
  },
  base: './'
})
