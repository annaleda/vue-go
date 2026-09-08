// Server Hypernova per lo step "dati conducente".
//
// Questo file è insieme:
//   - il server HTTP che espone il protocollo Hypernova (POST /batch)
//   - il registro delle "view" che questo micro-frontend sa renderizzare
//
// Non esiste un server.js sorgente: questo file viene impacchettato da
// webpack (target node) e il risultato è dist/server.js.

import hypernova from 'hypernova/server'
import { renderVuex } from 'hypernova-vue/server'
import express from 'express'
import path from 'path'

import WebApp from './web/App.vue'
import MobileApp from './mobile/App.vue'
import { createStore } from './store'

const STEP_ID = process.env.STEP_ID || 'dati'
const ID_VIEW_WEB = process.env.ID_VIEW_WEB || 'dati-step_web'
const ID_VIEW_MOBILE = process.env.ID_VIEW_MOBILE || 'dati-step_mobile'
const PORT = Number(process.env.PORT || 3031)

hypernova({
  devMode: process.env.NODE_ENV !== 'production',

  // getComponent è il cuore del protocollo.
  //
  // Il chiamante fa POST /batch con {"nome": {"name": "...", "data": {...}}}
  // e Hypernova, per ogni voce, chiama questa funzione. Quello che si
  // restituisce viene renderizzato lato server e torna indietro come
  // { html, meta }.
  getComponent(name, data, context) {
    // returnMeta è il canale per dire al chiamante QUALE SCRIPT caricare
    // per l'hydration. Senza, il chiamante riceverebbe HTML morto.
    const returnMeta = (data && data.returnMeta) || {}

    if (name === ID_VIEW_MOBILE) {
      returnMeta.src = `/${STEP_ID}/mobile_client.js`
      return renderVuex(ID_VIEW_MOBILE, MobileApp, createStore)
    }
    if (name === ID_VIEW_WEB) {
      returnMeta.src = `/${STEP_ID}/web_client.js`
      return renderVuex(ID_VIEW_WEB, WebApp, createStore)
    }
    // Nome sconosciuto: si restituisce null e Hypernova risponde con un
    // errore per quella singola voce, senza far fallire l'intero batch.
    return null
  },

  port: PORT,

  // createApplication permette di aggiungere le proprie route a quelle che
  // Hypernova installa da solo.
  createApplication() {
    const app = express()

    // I bundle di hydration vanno serviti da qualche parte. Il percorso
    // /<STEP_ID>/ corrisponde a quello messo in returnMeta.src qui sopra.
    app.use(`/${STEP_ID}`, express.static(path.join(process.cwd(), 'dist')))

    app.get('/health/liveness', (req, res) => res.json({ status: 'UP' }))
    app.get('/health/readiness', (req, res) => res.json({ status: 'UP' }))

    return app
  },
})

// eslint-disable-next-line no-console
console.log(`dati-step-fe: Hypernova in ascolto sulla porta ${PORT}`)
