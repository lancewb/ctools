import { createApp } from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import router from './router' // 稍后创建
import './style.css' // 可选，保留 wails 默认样式或清空

createApp(App).use(vuetify).use(router).mount('#app')