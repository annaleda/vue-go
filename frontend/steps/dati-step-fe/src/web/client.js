// Bundle di HYDRATION per la vista web.
//
// Questo file gira NEL BROWSER. Hypernova ha già mandato l'HTML renderizzato
// lato server; questo script lo "riattacca" a Vue, così i campi diventano
// reattivi e gli eventi funzionano.
//
// Il nome della view deve corrispondere ESATTAMENTE a quello usato in
// src/index.js lato server: è la chiave con cui hypernova-vue ritrova il
// markup e i dati serializzati nella pagina.
import { renderVuex } from 'hypernova-vue'
import App from './App.vue'
import { createStore } from '../store'

renderVuex(process.env.ID_VIEW_WEB || 'dati-step_web', App, createStore)
