import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // 自动按需引入 Vuetify 组件和指令，解决 hoisted vnodes 报错
    vuetify({ autoImport: true }),
  ],
  define: {
    // 解决 Feature flag 警告
    __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false',
  },
  // Wails 开发模式下的端口配置保持不变（如果有的话）
  server: {
    port: 34115,
    fs: {
      // 允许为项目根目录之外的文件提供服务
      strict: false,
    },
  }
})