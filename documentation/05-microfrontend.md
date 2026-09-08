# 05 — Micro-frontend con Hypernova

[← Kubernetes e Helm](04-kubernetes-e-helm.md) · [Indice](../README.md#documentazione)

Il progetto contiene **la stessa applicazione fatta in due modi**, e girano
insieme così puoi tenerle aperte affiancate:

| | Dove | Cos'è |
|---|---|---|
| **SPA** | `frontend/quote-web` — porta 3000 | Una sola app Vue 3 che fa tutto |
| **Micro-frontend** | `frontend/steps/*` + `frontend/aggregator` — porta 3001 | La stessa cosa spezzata in due step indipendenti, ricomposti a runtime |

La seconda è l'architettura del canale digitale di Poste. Questo documento
spiega come funziona.

---

## 1. Il problema che risolve

Una SPA è un'unica unità di rilascio. Se venti persone in quattro team lavorano
sullo stesso funnel di vendita, ogni modifica richiede di ricostruire e
rilasciare tutto, e un errore in uno step blocca il rilascio degli altri.

L'idea dei micro-frontend è la stessa dei microservizi, applicata alla UI:
**ogni pezzo di interfaccia è un servizio autonomo**, con il suo repository,
la sua pipeline, il suo container, il suo rilascio.

Il problema diventa allora: come si rimettono insieme?

---

## 2. Le strade possibili

| Approccio | Come | Limiti |
|---|---|---|
| **iframe** | Ogni pezzo in un iframe | Isolamento totale ma comunicazione scomoda, stili e altezze un incubo |
| **Module Federation** (webpack 5) | I bundle si importano a vicenda a runtime | Tutto lato client, e serve la stessa versione del framework |
| **Web Components** | Ogni pezzo è un custom element | Indipendente dal framework, ma niente SSR |
| **SSR componibile** (Hypernova) | Ogni pezzo espone un servizio che renderizza HTML | Un servizio Node per pezzo, quindi costo di infrastruttura |

Poste ha scelto l'ultima, e questo progetto la riproduce.

---

## 3. Come funziona Hypernova

**Hypernova** è un progetto open source di Airbnb. In sostanza è un
**protocollo HTTP per chiedere a un servizio di renderizzare un componente**.

```
POST /batch
{ "vista1": { "name": "dati-step_web", "data": {} } }

→ { "success": true,
    "results": {
      "vista1": {
        "html": "<div data-hypernova-key=\"datistep_web\" data-hypernova-id=\"2d4f...\">…</div>
                 <script type=\"application/json\" data-hypernova-key=…>{props}</script>",
        "meta": { "src": "/dati/web_client.js" }
      } } }
```

Tre cose in una risposta:

1. **`html`** — il markup già renderizzato, con dentro gli attributi
   `data-hypernova-key` e `data-hypernova-id` che serviranno a ritrovarlo;
2. **il blocco `<script type="application/json">`** — le props serializzate,
   perché il browser deve ricostruire lo stesso stato che aveva il server;
3. **`meta.src`** — l'URL del bundle che riporta in vita quel markup.

Si chiama `/batch` perché in una sola chiamata si possono chiedere più viste.

### Il lato server

`frontend/steps/dati-step-fe/src/index.js` è insieme il server e il registro
delle viste:

```js
hypernova({
  getComponent(name, data, context) {
    const returnMeta = (data && data.returnMeta) || {}

    if (name === ID_VIEW_MOBILE) {
      returnMeta.src = `/${STEP_ID}/mobile_client.js`
      return renderVuex(ID_VIEW_MOBILE, MobileApp, createStore)
    }
    if (name === ID_VIEW_WEB) {
      returnMeta.src = `/${STEP_ID}/web_client.js`
      return renderVuex(ID_VIEW_WEB, WebApp, createStore)
    }
    return null
  },
  port: PORT,
  createApplication() { /* express, static, health */ },
})
```

**Due viste per ogni step**, `web` e `mobile`, registrate con nomi diversi e
bundle diversi. Lo stesso step deve rendersi nel portale desktop e dentro
l'app, con layout differenti.

### Il lato client

`src/web/client.js` è tre righe, e gira nel browser:

```js
import { renderVuex } from 'hypernova-vue'
import App from './App.vue'
import { createStore } from '../store'

renderVuex('dati-step_web', App, createStore)
```

Cerca nella pagina il nodo con quella chiave, legge le props dal blocco JSON,
e ci monta sopra Vue. Da lì in poi il componente è vivo: i campi diventano
reattivi, gli eventi funzionano.

**Il nome deve coincidere esattamente** fra server e client. Se sono diversi
non succede nulla di visibile: la pagina appare corretta ma è HTML morto.

---

## 4. La build: tre bundle da un solo comando

`webpack.common.js` esporta **due configurazioni**, e `webpack.prod.js` le dà
a webpack come array:

```js
module.exports = [
  merge(server, { mode: 'production' }),   // target: 'node'
  merge(client, { mode: 'production' }),   // target: 'web', due entry
]
```

Risultato in `dist/`:

```
server.js          5,7 KB    il server Hypernova (gira in Node)
web_client.js      109 KB    hydration della vista desktop
mobile_client.js   109 KB    hydration della vista mobile
```

`server.js` è minuscolo perché `webpack-node-externals` esclude i
`node_modules`: sul server sono già su disco, impacchettarli sarebbe inutile
e romperebbe i moduli nativi.

Nella config client invece serve dire a webpack che alcuni moduli Node non
esistono nel browser:

```js
node: { fs: 'empty', net: 'empty', module: 'empty' },
```

E il CSS va trattato diversamente nei due target: nel bundle server si scarta
(`null-loader`), nel client si inietta (`vue-style-loader`).

---

## 5. L'aggregator

`frontend/aggregator` è un'applicazione Nuxt 2 che non contiene logica di
business. Il suo unico mestiere è comporre.

In `pages/index.vue`, `asyncData` gira **sul server** prima che il browser
riceva qualcosa:

```js
const richieste = steps.map(async (step) => {
  const { data } = await axios.post(
    `${step.endpoint}/batch`,
    { [step.id]: { name: nomeVista, data: {} } },
    { timeout: 5000 }
  )
  const r = data.results[step.id]
  return { id: step.id, html: r.html, script: r.meta && r.meta.src }
})

return { viste: await Promise.all(richieste) }
```

**In parallelo, non in sequenza.** Con due step la differenza non si vede, con
dieci sì: le latenze si sommerebbero.

E se uno step non risponde, si cattura l'errore e si restituisce `html: null`.
La pagina mostra gli altri e segnala il buco — **il guasto resta circoscritto**,
che è il vantaggio pratico più concreto della composizione.

Poi `components/NovaView.vue` mette in pagina i due pezzi:

```html
<div v-html="html"></div>
<component v-if="script" :is="'script'" :src="script" async></component>
```

`:is="'script'"` è il modo di far generare a Vue un tag `<script>`: scritto
direttamente nel template verrebbe interpretato dal compilatore invece che
emesso.

> Nel progetto Poste fra aggregator e step c'è un ulteriore servizio in Go,
> il **views-hub**, che tiene la mappa step → endpoint e decide quale vista
> chiedere in base al canale. Qui quella mappa sta in `nuxt.config.js`: un
> livello in meno, stesso principio.

---

## 6. L'EventBus, e il tranello

I due step devono comunicare: `garanzie-step-fe` ha bisogno dei dati raccolti
da `dati-step-fe`. Ma sono servizi diversi, bundle diversi, e non si conoscono.

La soluzione è un bus di eventi condiviso. E qui c'è **il tranello più
insidioso di tutta l'architettura**.

Se ogni bundle facesse semplicemente:

```js
export const bus = new Vue()      // SBAGLIATO
```

...ognuno avrebbe la **propria istanza**. Girerebbero nella stessa pagina,
sembrerebbe tutto a posto, ma gli eventi emessi da uno non arriverebbero mai
all'altro. Nessun errore in console, nessun indizio: funziona tutto tranne la
comunicazione.

L'unico posto davvero condiviso fra bundle diversi è `window`:

```js
const CHIAVE = '__FEU_EVENT_BUS__'

function getBus() {
  if (typeof window === 'undefined') return new Vue()   // SSR: bus usa e getta
  if (!window[CHIAVE]) window[CHIAVE] = new Vue()
  return window[CHIAVE]
}

export const bus = getBus()
```

Il controllo su `typeof window` non è opzionale: durante il rendering
server-side `window` non esiste e il modulo esploderebbe. Va bene restituire
un bus locale, perché lato server nessuno deve comunicare con nessuno — gli
eventi servono solo dopo l'hydration.

### I nomi degli eventi sono il contratto

```js
export const EVENTI = {
  DATI_CONDUCENTE: 'dati-conducente-cambiati',
  PREVENTIVO_CALCOLATO: 'preventivo-calcolato',
  STEP_COMPLETATO: 'step-completato',
}
```

In un'architettura dove i pezzi si parlano **solo** per evento, questi nomi
sono l'unica interfaccia che hanno. Un errore di battitura non dà nessun
errore: semplicemente l'evento non arriva. Tenerli in un posto solo è la
difesa minima; in un progetto vero starebbero in una libreria condivisa.

### Disiscriversi è obbligatorio

```js
mounted() {
  bus.$on(EVENTI.DATI_CONDUCENTE, this.onDatiConducente)
},
beforeDestroy() {
  bus.$off(EVENTI.DATI_CONDUCENTE, this.onDatiConducente)
},
```

Il bus vive su `window` e sopravvive ai componenti. Senza `$off`, l'ascoltatore
resta agganciato a un componente distrutto: è una perdita di memoria che nelle
applicazioni a lunga vita si accumula.

### Il limite di fondo: il bus non ha memoria

Se uno step si idrata **dopo** che un altro ha già emesso il suo evento, quel
messaggio è perso: nessuno era in ascolto. Nel progetto si rimedia
ri-emettendo in `mounted()`, ma è una toppa.

> **Ed è esattamente il motivo per cui in Poste lo stato vero non sta
> nell'EventBus ma nel Funnel Orchestrator.** Il bus serve a notificare
> ("è cambiato qualcosa"), non a conservare. Chi arriva tardi interroga
> l'orchestratore, non il bus.

---

## 7. Il costo, misurato

Vale la pena guardare i numeri prima di innamorarsi dell'architettura.

| | SPA (`quote-web`) | Micro-frontend |
|---|---|---|
| Servizi da gestire | 1 | 3 (2 step + aggregator) |
| Immagine | 48 MB (nginx + statici) | ~250 MB ciascuna (Node + node_modules) |
| Memoria richiesta | 32 Mi | 256 Mi per servizio |
| Runtime | nessuno, file statici | processo Node che renderizza a ogni richiesta |
| Avvio | immediato | **~20 secondi** (Hypernova fa il fork di 4 worker) |
| Comunicazione interna | props ed eventi Vue | bus globale su `window` |

Il salto è netto: da un container nginx da 48 MB a tre processi Node da 256 Mi
ciascuno. In `values.yaml` del chart queste differenze sono scritte esplicite.

**Quando conviene**: quando i team sono tanti e devono rilasciare in modo
indipendente. È un costo tecnico pagato per un beneficio organizzativo.

**Quando no**: se l'applicazione la mantiene un team solo, la SPA è più
semplice, più leggera e più veloce.

---

## 8. Due cose scoperte costruendo questo progetto

**`@babel/runtime` va dichiarato a mano.** Il primo build è fallito così:

```
Module not found: Error: Can't resolve '@babel/runtime/helpers/defineProperty'
  in '/app/node_modules/nova-helpers/lib'
```

`nova-helpers`, dipendenza di `hypernova-vue`, lo richiede senza dichiararlo.
Va aggiunto alle proprie dipendenze. È il motivo per cui `@babel/runtime`
compare nel `package.json` dei micro-frontend Poste: non è una scelta, è una
toppa obbligata.

**Hypernova gira in cluster mode.** All'avvio fa il fork di più worker e
impiega una ventina di secondi a mettersi effettivamente in ascolto:

```
Worker #1 is now online
...
Worker #1 is now connected to 0.0.0.0:3031
```

Su Kubernetes serve quindi una `startupProbe` generosa, altrimenti la liveness
uccide il pod prima che sia pronto. Nel chart è `startupFailureThreshold: 30`
contro i 15 dei servizi Go.

---

## 9. Provarlo

```bash
docker compose up -d --build
```

| | |
|---|---|
| **Micro-frontend composti** | **http://localhost:3001** |
| SPA monolitica (confronto) | http://localhost:3000 |
| Step "dati" da solo | http://localhost:3031 |
| Step "garanzie" da solo | http://localhost:3032 |

Sull'aggregator: **cambia i dati nel primo riquadro e premi *Calcola* nel
secondo**. Il totale compare nella barra nera in alto, che appartiene
all'aggregator — tre bundle diversi che si parlano solo per evento.

Interrogare direttamente il protocollo:

```bash
curl -s -X POST http://localhost:3031/batch \
  -H "Content-Type: application/json" \
  -d '{"v":{"name":"dati-step_web","data":{}}}' | head -c 600
```

Vedrai l'HTML renderizzato lato server con `data-hypernova-key` e il blocco
JSON delle props.

Per vedere il **guasto circoscritto** in azione:

```bash
docker compose stop garanzie-step-fe
```

Ricarica l'aggregator: il primo step funziona, al posto del secondo compare
l'avviso. Con una SPA sarebbe caduta l'intera pagina.

---

## 10. La corrispondenza con Poste

| Questo progetto | Poste |
|---|---|
| `dati-step-fe`, `garanzie-step-fe` | i 14 `*-step-fe` di welfare-card, area-riservata, cross-selling |
| `aggregator` (Nuxt) | `feu/hike/aggregator` e `portal-aggregator` |
| mappa step in `nuxt.config.js` | **views-hub**, un servizio in Go |
| `event-bus.js` su `window` | `EventBus` esportato da `feu-ui-components-lib` |
| store Vuex per step | idem |
| viste `web` e `mobile` | idem, `ID_VIEW_WEB` / `ID_VIEW_MOBILE` |
| — | **Funnel Orchestrator**: lo stato persistente, che qui non c'è |

L'ultima riga è la differenza più importante. Qui lo stato vive nella pagina e
sparisce ricaricando. In Poste vive in un servizio esterno, ed è ciò che
permette a un cliente di riprendere il percorso il giorno dopo, da un altro
canale.

---

[← Kubernetes e Helm](04-kubernetes-e-helm.md) · [Indice](../README.md#documentazione)
