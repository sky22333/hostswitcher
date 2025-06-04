import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  build: {
    rollupOptions: {
      // 确保外部依赖正确处理
      external: [],
    },
  },
  // 确保开发服务器也能正确解析别名
  server: {
    fs: {
      strict: false,
    },
  },
})
