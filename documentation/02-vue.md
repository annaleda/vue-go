# 02 — Vue

[← Go](01-go.md) · [Indice](../README.md#documentazione) · [Docker →](03-docker.md)

Vue da zero, sul codice di `frontend/quote-web`.

---

## 1. L'idea di fondo: la reattività

In JavaScript classico si manipola il DOM a mano: prendi l'elemento, cambi
il testo, aggiorni la tabella. Il rischio è che lo stato dell'applicazione e
quello che si vede a schermo divergano.

Vue rovescia il rapporto: **dichiari come lo stato si traduce in interfaccia**,
e quando lo stato cambia l'interfaccia si aggiorna da sola.

```js
data() {
  return { preventivo: null }
}
```

```html
<RisultatoPreventivo v-if="preventivo" :preventivo="preventivo" />
```

Nessuno scrive "ora mostra il risultato". Si assegna `this.preventivo` e la
tabella compare. Vue sa quali parti del DOM dipendono da quel dato e ridisegna
solo quelle.

---

## 2. Il Single File Component

Un file `.vue` contiene tre sezioni, e tutte e tre riguardano lo stesso
componente:

```vue
<script>  // la logica
export default { ... }
</script>

<template>  <!-- la struttura -->
  <div>...</div>
</template>

<style scoped>  /* l'aspetto */
</style>
```

`scoped` nello style significa che quel CSS vale **solo per questo
componente**: Vue aggiunge un attributo univoco agli elementi e lo inserisce
nei selettori. Senza, una classe `.form` colpirebbe qualunque `.form` della
pagina. È il modo in cui Vue evita il problema storico dei CSS globali.

---

## 3. Le due sintassi

Vue 3 ne ha due, e in questo progetto ci sono entrambe di proposito.

### Options API — `App.vue`, `FormPreventivo.vue`

Lo stato in `data()`, la logica in `methods`, i valori derivati in `computed`.

```js
export default {
  props: { garanzie: { type: Array, default: () => [] } },
  data() {
    return { etaConducente: 40, selezionate: ['RCA'] }
  },
  computed: {
    valido() { return this.etaConducente >= 18 && this.selezionate.length > 0 }
  },
  methods: {
    invia() { this.$emit('calcola', { ... }) }
  },
}
```

È la forma di **Vue 2**, e resta pienamente supportata in Vue 3. La trovi in
tutto il codice esistente scritto prima del 2021.

### Composition API — `RisultatoPreventivo.vue`

```vue
<script setup>
import { computed } from 'vue'

const props = defineProps({ preventivo: { type: Object, required: true } })

const rate = computed(() => ({
  annuale: props.preventivo.totale,
  mensile: props.preventivo.totale / 12,
}))

function euro(v) { return v.toFixed(2) + ' €' }
</script>
```

Tutto ciò che è dichiarato nello `<script setup>` è automaticamente
disponibile nel template.

### La corrispondenza

| Options API | Composition API |
|---|---|
| `props: { ... }` | `const props = defineProps({ ... })` |
| `data() { return { x: 1 } }` | `const x = ref(1)` → si legge `x.value` |
| `computed: { y() {...} }` | `const y = computed(() => ...)` |
| `methods: { f() {...} }` | `function f() {...}` |
| `mounted() {...}` | `onMounted(() => {...})` |
| `this.x` | `x.value` (nel template basta `x`) |

**Quale usare?** Nei componenti piccoli è indifferente. In quelli grandi la
Options API sparpaglia la logica di una stessa funzionalità fra `data`,
`computed` e `methods`, mentre la Composition API la tiene insieme — ed è
il motivo per cui è nata.

> Un dettaglio che confonde all'inizio: con `ref()` si scrive `x.value` nel
> codice ma solo `x` nel template. Vue "srotola" automaticamente i ref nei
> template.

---

## 4. Le direttive che servono davvero

### `v-model` — legare un campo a una variabile

```html
<input v-model.number="etaConducente" type="number" />
```

Legame nei due sensi: scrivi nel campo e la variabile cambia, cambi la
variabile e il campo si aggiorna.

Il modificatore **`.number` non è opzionale** quando serve un numero: senza,
`"40"` resta una stringa, e `"9" > "10"` è vero perché il confronto è
alfabetico. È un errore facile e insidioso.

### `v-for` — ripetere su una lista

```html
<label v-for="g in garanzie" :key="g.codice">
  {{ g.nome }}
</label>
```

**`:key` va messa sempre.** Serve a Vue per riconoscere gli elementi fra un
ridisegno e l'altro. Senza, in una lista che cambia ordine Vue riusa i nodi
sbagliati e ti ritrovi la checkbox spuntata sulla riga sbagliata. La chiave
deve essere stabile e univoca: **mai l'indice** se la lista può riordinarsi.

### `v-if` — mostrare o no

```html
<p v-if="caricamento">Calcolo in corso…</p>
```

`v-if` aggiunge e rimuove l'elemento dal DOM. Esiste anche `v-show`, che lo
lascia lì e cambia solo `display`: meglio quando la condizione si alterna
spesso, peggio quando l'elemento è pesante e raramente visibile.

### I due punti e la chiocciola

```html
<FormPreventivo :garanzie="garanzie" @calcola="onCalcola" />
```

- `:` è la scorciatoia di `v-bind:` — passa un **valore JavaScript**.
  Senza i due punti, `garanzie="garanzie"` passerebbe la stringa `"garanzie"`.
- `@` è la scorciatoia di `v-on:` — ascolta un **evento**.

Modificatore utile: `@submit.prevent` chiama `preventDefault()` e impedisce
al browser di ricaricare la pagina inviando il form.

---

## 5. Props giù, eventi su

È la regola fondamentale della comunicazione fra componenti.

```
        App.vue  (ha lo stato)
           │  :garanzie          ▲  @calcola
           ▼                     │
      FormPreventivo             │
           └─────────────────────┘
```

Il **padre passa dati** al figlio con le props. Il **figlio notifica** il
padre con un evento, e il padre decide cosa fare.

```js
// nel figlio
this.$emit('calcola', { etaConducente: 40, ... })
```

```html
<!-- nel padre -->
<FormPreventivo @calcola="onCalcola" />
```

**Un figlio non modifica mai direttamente le props.** Sono in sola lettura,
e Vue avvisa in console se ci provi. Il motivo è che altrimenti diventerebbe
impossibile capire chi ha cambiato cosa.

Quando i componenti da coordinare sono tanti e sparsi, passare props di
livello in livello diventa scomodo (*prop drilling*). Lì si introduce uno
**store** condiviso: Pinia in Vue 3, Vuex in Vue 2.

---

## 6. `computed` invece di metodi

```js
computed: {
  stimaIndicativa() {
    return this.garanzie
      .filter(g => this.selezionate.includes(g.codice))
      .reduce((somma, g) => somma + g.prezzoBase, 0)
  }
}
```

Una `computed` si usa nel template come se fosse un dato (`{{ stimaIndicativa }}`,
senza parentesi) ma è calcolata. La differenza rispetto a un metodo è la
**cache**: si ricalcola solo quando cambia qualcosa da cui dipende, mentre
un metodo verrebbe rieseguito a ogni ridisegno.

Regola pratica: se è un valore derivato dallo stato, è una computed.

---

## 7. Il ciclo di vita

```js
async mounted() {
  this.garanzie = await caricaGaranzie()
}
```

`mounted()` scatta quando il componente è stato inserito nella pagina. È il
punto giusto per le chiamate HTTP iniziali. Gli altri che si usano davvero
sono `created()` (prima del DOM) e `unmounted()` (per disiscriversi da
timer ed event listener, altrimenti restano appesi).

---

## 8. Le chiamate HTTP

Stanno tutte in `src/api.js`, separate dai componenti. È una buona abitudine:
i componenti restano concentrati sulla presentazione, e il giorno che cambia
un endpoint si tocca un file solo.

```js
const res = await fetch(`${BASE}/garanzie`)
if (!res.ok) { throw new Error(...) }
return res.json()
```

**Attenzione a `fetch`**: non lancia eccezioni sui 4xx e 5xx. Una risposta
`500` arriva come risposta normale, e `res.ok` è l'unico modo di accorgersene.
È la differenza più insidiosa rispetto ad axios, che invece rifiuta la promise.

Lato componente, il pattern è sempre questo:

```js
async onCalcola(richiesta) {
  this.caricamento = true
  this.errore = null
  try {
    this.preventivo = await calcolaPreventivo(richiesta)
  } catch (e) {
    this.errore = e.message
  } finally {
    this.caricamento = false     // sempre, anche in caso di errore
  }
}
```

Senza `finally`, un errore lascerebbe l'interfaccia bloccata sullo stato
"sto caricando".

---

## 9. Vite e il proxy

Vite fa due mestieri diversi:

- **in sviluppo** serve i moduli al browser senza impacchettarli, quindi
  parte in un istante e ricarica solo il file che hai toccato;
- **in build** produce i file statici ottimizzati in `dist/`.

Il proxy in `vite.config.js` risolve un problema concreto:

```js
proxy: {
  '/api/preventivi': {
    target: 'http://localhost:8080',
    rewrite: (path) => path.replace(/^\/api\/preventivi/, '/api/v1/preventivi'),
  },
}
```

In sviluppo il frontend gira sulla 5173 e i servizi sull'8080: per il browser
sono **origini diverse**, e senza intervento le chiamate verrebbero bloccate
dal CORS. Il proxy fa sì che il browser veda una sola origine.

In produzione lo stesso lavoro lo fa nginx (`nginx.conf`), e in Kubernetes il
gateway. Il percorso `/api/...` resta identico nei tre casi: è per questo che
il codice del frontend non cambia mai fra sviluppo e produzione.

---

## 10. Vue 2 e Vue 3

Vue 2 è **fuori supporto da dicembre 2023**, ma resta moltissimo codice in
giro. Le differenze che incontri leggendolo:

| | Vue 2 | Vue 3 |
|---|---|---|
| Avvio | `new Vue({ render: h => h(App) }).$mount('#app')` | `createApp(App).mount('#app')` |
| Store | Vuex | Pinia (Vuex funziona ancora) |
| Reattività | `Object.defineProperty` | `Proxy` |
| Aggiungere proprietà | serviva `Vue.set(obj, 'k', v)` | funziona e basta |
| Radice del template | **una sola** | più elementi permessi |
| Composition API | no (c'era un plugin) | sì |
| Build | webpack / vue-cli | Vite |

Il limite dell'unica radice in Vue 2 dipendeva da come era implementata la
reattività, e obbligava a un `<div>` contenitore inutile in ogni componente.

---

## 11. Dove guardare nel codice

| Concetto | File |
|---|---|
| montaggio dell'app | `src/main.js` |
| Options API, stato, ciclo di vita, props giù/eventi su | `src/App.vue` |
| `v-model`, `v-for`, `computed`, `$emit` | `src/components/FormPreventivo.vue` |
| Composition API, `script setup`, `defineProps` | `src/components/RisultatoPreventivo.vue` |
| chiamate HTTP separate dai componenti | `src/api.js` |
| proxy di sviluppo | `vite.config.js` |

## 12. Comandi

```bash
npm install       # installa le dipendenze
npm run dev       # sviluppo su http://localhost:5173, ricarica a caldo
npm run build     # produce dist/
npm run preview   # serve dist/ per provare la build
```

> Se `npm install` non trova i pacchetti, npm probabilmente punta a un
> registry aziendale: `npm install --registry https://registry.npmjs.org`

---

[← Go](01-go.md) · [Indice](../README.md#documentazione) · [Docker →](03-docker.md)
