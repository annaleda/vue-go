import { createApp } from 'vue'
import App from './App.vue'
import './style.css'

// Il punto di ingresso dell'applicazione.
//
// createApp(App) crea l'istanza, .mount('#app') la attacca al div in
// index.html. Da qui in poi Vue controlla quel pezzo di DOM.
//
// In Vue 2 la stessa cosa si scriveva:
//    new Vue({ render: h => h(App) }).$mount('#app')
// Il codice dei micro-frontend Poste è ancora in questa forma.
createApp(App).mount('#app')
