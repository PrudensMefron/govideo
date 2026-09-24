import { createApp } from 'vue'
import App from './App.vue'
import './style.css'

let theme = 'system'
try { theme = localStorage.getItem('govideo-theme') || 'system' } catch { /* storage may be unavailable */ }
document.documentElement.dataset.theme = theme === 'system'
  ? (matchMedia('(prefers-color-scheme: dark)').matches ? 'govideo-dark' : 'govideo-light')
  : `govideo-${theme}`

createApp(App).mount('#app')
