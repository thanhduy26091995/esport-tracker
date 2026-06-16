import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './style.css'
import App from './App.vue'
import router from './router'
import { i18n } from './plugins/i18n'
import { useLocaleStore } from './stores/localeStore'

const app = createApp(App)
const pinia = createPinia()
const localeStore = useLocaleStore(pinia)

i18n.global.locale.value = localeStore.getLocale()

app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(i18n)

// Forward unhandled Vue errors to Umami as custom events
app.config.errorHandler = (err, _instance, info) => {
  console.error('[Vue error]', err, info)
  try {
    window.umami?.track('js-error', {
      message: err instanceof Error ? err.message : String(err),
      info,
      url: window.location.pathname,
    })
  } catch {}
}

app.mount('#app')
