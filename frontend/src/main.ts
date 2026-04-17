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
app.mount('#app')
