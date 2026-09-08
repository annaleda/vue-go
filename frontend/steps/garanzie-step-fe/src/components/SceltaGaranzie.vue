<script>
// Questo è il micro-frontend interessante, perché DIPENDE DA UN ALTRO
// senza conoscerlo.
//
// I dati del conducente li raccoglie `dati-step-fe`, che è un servizio
// diverso, un repository diverso, un bundle diverso. Qui arrivano solo
// attraverso l'EventBus: nessun import, nessuna chiamata, nessuna
// conoscenza reciproca. È tutto il senso dei micro-frontend.

import { bus, EVENTI } from '../event-bus'

export default {
  name: 'SceltaGaranzie',

  props: {
    web: { type: Boolean, default: true },
  },

  computed: {
    garanzie() { return this.$store.state.garanzie },
    preventivo() { return this.$store.state.preventivo },
    errore() { return this.$store.state.errore },
    caricamento() { return this.$store.state.caricamento },
    datiValidi() { return this.$store.state.datiValidi },

    selezionate: {
      get() { return this.$store.state.selezionate },
      set(v) { this.$store.commit('setSelezionate', v) },
    },

    puoCalcolare() {
      return this.datiValidi && this.selezionate.length > 0 && !this.caricamento
    },
  },

  // mounted gira SOLO nel browser, mai durante il rendering server-side.
  // È quindi il posto giusto sia per iscriversi al bus sia per le chiamate
  // HTTP: lato server non ci sarebbe nessuno con cui parlare.
  mounted() {
    this.$store.dispatch('caricaGaranzie')

    // Ci si mette in ascolto dell'altro step.
    bus.$on(EVENTI.DATI_CONDUCENTE, this.onDatiConducente)
  },

  // Disiscriversi è obbligatorio.
  //
  // Il bus vive su `window` e sopravvive al componente: se non si rimuove
  // l'ascoltatore, resta agganciato a un componente distrutto. È una perdita
  // di memoria, e nelle applicazioni a lunga vita si accumula.
  beforeDestroy() {
    bus.$off(EVENTI.DATI_CONDUCENTE, this.onDatiConducente)
  },

  methods: {
    onDatiConducente(payload) {
      this.$store.commit('setDatiConducente', payload)
      // Se i dati cambiano, il preventivo calcolato prima non vale più.
      this.$store.commit('setPreventivo', null)
    },

    async calcola() {
      const prev = await this.$store.dispatch('calcola')
      if (prev) {
        // Si avvisa il resto della pagina: l'aggregator mostra il totale
        // nella barra in alto senza sapere come è stato calcolato.
        bus.$emit(EVENTI.PREVENTIVO_CALCOLATO, prev)
        bus.$emit(EVENTI.STEP_COMPLETATO, { step: 'garanzie' })
      }
    },

    euro(v) {
      return Number(v).toFixed(2).replace('.', ',') + ' €'
    },
  },
}
</script>

<template>
  <div class="scelta" :class="{ mobile: !web }">
    <h2>Garanzie</h2>

    <p v-if="!datiValidi" class="avviso">
      Completa prima i dati del conducente.
    </p>

    <p v-if="garanzie.length === 0" class="vuoto">Caricamento catalogo…</p>

    <label v-for="g in garanzie" :key="g.codice" class="garanzia">
      <input
        type="checkbox"
        :value="g.codice"
        v-model="selezionate"
        :disabled="g.obbligatoria"
      />
      <span class="nome">
        {{ g.nome }}
        <em v-if="g.obbligatoria">(obbligatoria)</em>
      </span>
      <span class="prezzo">{{ euro(g.prezzoBase) }}</span>
    </label>

    <div class="azioni">
      <button type="button" :disabled="!puoCalcolare" @click="calcola">
        {{ caricamento ? 'Calcolo…' : 'Calcola preventivo' }}
      </button>
    </div>

    <p v-if="errore" class="errore">{{ errore }}</p>

    <div v-if="preventivo" class="risultato">
      <div class="riga" v-for="v in preventivo.voci" :key="v.codice">
        <span>{{ v.nome }}</span>
        <span class="num">{{ euro(v.prezzo) }}</span>
      </div>
      <div class="riga totale">
        <span>Totale annuo (imposte incluse)</span>
        <span class="num">{{ euro(preventivo.totale) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.scelta {
  padding: 1rem 1.25rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: #fff;
}

h2 {
  margin: 0 0 0.8rem;
  font-size: 1.05rem;
}

.avviso {
  padding: 0.5rem 0.75rem;
  border-left: 3px solid #d69e2e;
  background: #fffaf0;
  color: #975a16;
  font-size: 0.85rem;
}

.vuoto {
  color: #888;
  font-size: 0.85rem;
}

.garanzia {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.35rem 0;
  border-bottom: 1px solid #f0f0f0;
  font-size: 0.9rem;
}

.garanzia .nome { flex: 1; }
.garanzia em { color: #888; font-size: 0.78rem; }
.garanzia .prezzo { color: #555; font-variant-numeric: tabular-nums; }

.azioni {
  margin-top: 0.9rem;
}

button {
  padding: 0.55rem 1.1rem;
  border: 0;
  border-radius: 4px;
  background: #1f6feb;
  color: #fff;
  font-size: 0.9rem;
  cursor: pointer;
}

button:disabled {
  background: #bbb;
  cursor: not-allowed;
}

.errore {
  margin-top: 0.7rem;
  padding: 0.5rem 0.75rem;
  border-left: 3px solid #c0392b;
  background: #fdf0ef;
  color: #922;
  font-size: 0.85rem;
}

.risultato {
  margin-top: 1rem;
  padding-top: 0.6rem;
  border-top: 1px solid #eee;
  font-size: 0.9rem;
}

.riga {
  display: flex;
  justify-content: space-between;
  padding: 0.25rem 0;
}

.riga.totale {
  margin-top: 0.4rem;
  padding-top: 0.5rem;
  border-top: 2px solid #ccc;
  font-weight: 700;
}

.num { font-variant-numeric: tabular-nums; }
</style>
