import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { nextTick } from 'vue'

// Vuetify
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { aliases, mdi } from 'vuetify/iconsets/mdi'
import '@mdi/font/css/materialdesignicons.css'

// 自定义样式
import './style.css'
import App from './App.vue'

// 现代化的亮色和暗色主题
const lightTheme = {
  dark: false,
  colors: {
    background: '#f5f5f5',
    surface: '#ffffff',
    primary: '#0078d4',
    'primary-darken-1': '#106ebe',
    secondary: '#5c2d91',
    'secondary-darken-1': '#4b2474',
    error: '#d13438',
    info: '#0078d4',
    success: '#107c10',
    warning: '#ffb900',
    'on-surface': '#1f1f1f',
    'surface-variant': '#f3f3f3',
  }
}



// 创建 Vuetify 实例
const vuetify = createVuetify({
  components,
  directives,
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: {
      mdi,
    },
  },
  theme: {
    defaultTheme: 'lightTheme',
    themes: {
      lightTheme,
    },
  },
  defaults: {
    VCard: {
      elevation: 0,
      rounded: 'lg',
    },
    VBtn: {
      rounded: 'lg',
      style: 'text-transform: none;',
    },
    VChip: {
      rounded: 'lg',
    },
    VTextField: {
      variant: 'outlined',
      density: 'comfortable',
      rounded: 'lg',
    },
    VSelect: {
      variant: 'outlined',
      density: 'comfortable',
      rounded: 'lg',
    },
    VTextarea: {
      variant: 'outlined',
      rounded: 'lg',
    },
    VDialog: {
      scrim: false,
      persistent: false,
    },
    VSnackbar: {
      scrim: false,
    },
  },
})

// 创建 Pinia 状态管理
const pinia = createPinia()

// 创建应用实例
const app = createApp(App)

// 注册插件
app.use(pinia)
app.use(vuetify)

// 挂载应用
app.mount('#app')

// 在应用挂载后立即初始化配置
nextTick(() => {
  // 初始化配置store
  if (window.runtime && window.runtime.EventsOn) {
    import('./stores/config').then(({ useConfigStore }) => {
      const configStore = useConfigStore()
      configStore.initialize()
    })
  }
})
