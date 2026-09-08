import { renderVuex } from 'hypernova-vue'
import App from './App.vue'
import { createStore } from '../store'

renderVuex(process.env.ID_VIEW_MOBILE || 'garanzie-step_mobile', App, createStore)
