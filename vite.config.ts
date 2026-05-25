import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { configDefaults } from 'vitest/config'

const cfAsyncModuleScriptPlugin = () => ({
  name: 'cfasync-module-script',
  transformIndexHtml(html: string) {
    return html.replace(
      /<script\s+type="module"(?![^>]*data-cfasync)/g,
      '<script data-cfasync="false" type="module"',
    )
  },
})

// https://vite.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag.startsWith('cap-'),
        },
      },
    }),
    cfAsyncModuleScriptPlugin(),
  ],
  esbuild: mode === 'production' ? { drop: ['console', 'debugger'] } : {},
  test: {
    environment: 'jsdom',
    globals: true,
    css: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
    exclude: [...configDefaults.exclude, '.worktrees/**', 'scripts/test/**', '.github/scripts/test/**'],
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-qrcode': ['qrcode'],
          'vendor-vue-i18n': ['vue-i18n'],
        },
      },
    },
  },
  server: {
    host: '0.0.0.0', // 监听所有网络接口
    port: 5173,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/sitemap.xml': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/robots.txt': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    }
  },
}))
