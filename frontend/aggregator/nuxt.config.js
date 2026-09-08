// Aggregator Nuxt 2.
//
// Il suo mestiere è comporre: non contiene nessuna logica di business, non
// sa cosa fanno gli step. Chiede a ciascuno il proprio HTML già renderizzato
// e lo mette in pagina.
//
// Nel progetto Poste fra aggregator e step c'è un ulteriore servizio in Go,
// il views-hub, che tiene la mappa step -> endpoint e decide quale vista
// (web o mobile) chiedere. Qui la mappa sta nella configurazione: un livello
// in meno, stesso principio.

export default {
  // Nuxt 2 in modalità universale: render lato server + hydration.
  ssr: true,
  target: 'server',

  head: {
    title: 'Preventivo auto — micro-frontend',
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    ],
  },

  css: ['~/assets/main.css'],

  // Solo client: l'EventBus non serve durante il render server-side, e
  // `window` lì non esiste.
  plugins: [{ src: '~/plugins/event-bus.js', mode: 'client' }],

  components: true,

  // Le probe di Kubernetes hanno bisogno di un endpoint che non renderizzi
  // una pagina intera: serverMiddleware risponde prima di arrivare a Nuxt.
  serverMiddleware: [
    { path: '/health/liveness', handler: '~/server-middleware/health.js' },
    { path: '/health/readiness', handler: '~/server-middleware/health.js' },
  ],

  // publicRuntimeConfig è leggibile anche dal browser; privateRuntimeConfig
  // solo dal server. Gli endpoint degli step sono nomi di servizio interni
  // (dati-step-fe, garanzie-step-fe) che dal browser non si risolvono:
  // devono restare privati, ed è il server a chiamarli.
  privateRuntimeConfig: {
    steps: [
      {
        id: 'dati',
        endpoint: process.env.DATI_STEP_URL || 'http://localhost:3031',
        viewWeb: 'dati-step_web',
        viewMobile: 'dati-step_mobile',
      },
      {
        id: 'garanzie',
        endpoint: process.env.GARANZIE_STEP_URL || 'http://localhost:3032',
        viewWeb: 'garanzie-step_web',
        viewMobile: 'garanzie-step_mobile',
      },
    ],
  },

  publicRuntimeConfig: {
    canale: process.env.CANALE || 'web',
  },

  // Il proxy gira le chiamate del browser verso i servizi Go e verso i
  // bundle di hydration degli step.
  modules: ['@nuxtjs/proxy'],
  proxy: {
    '/api/garanzie': {
      target: process.env.CATALOG_URL || 'http://localhost:8081',
      pathRewrite: { '^/api/garanzie': '/api/v1/garanzie' },
    },
    '/api/preventivi': {
      target: process.env.QUOTE_URL || 'http://localhost:8080',
      pathRewrite: { '^/api/preventivi': '/api/v1/preventivi' },
    },
    // I bundle di hydration sono serviti dagli step stessi, sotto /<id>/.
    // È il percorso che ogni step dichiara in returnMeta.src.
    '/dati': { target: process.env.DATI_STEP_URL || 'http://localhost:3031' },
    '/garanzie': { target: process.env.GARANZIE_STEP_URL || 'http://localhost:3032' },
  },

  server: {
    port: process.env.PORT || 8080,
    host: '0.0.0.0',
  },

  build: {},
}
