<script>
// Vue 2, Options API. È la sintassi che trovi nei micro-frontend Poste.
//
// Differenze rispetto a Vue 3 che si notano subito:
//   - il template deve avere UNA SOLA radice
//   - lo store è Vuex, non Pinia
//   - non esiste <script setup>

import { bus, EVENTI } from '../event-bus'

export default {
  name: 'FormDati',

  props: {
    web: { type: Boolean, default: true },
  },

  computed: {
    // mapState/mapGetters di Vuex farebbero lo stesso in meno righe, ma
    // scritto per esteso si vede cosa succede: si legge dallo store e si
    // scrive con una mutation.
    eta: {
      get() { return this.$store.state.etaConducente },
      set(v) { this.$store.dispatch('aggiorna', { campo: 'etaConducente', valore: v }) },
    },
    provincia: {
      get() { return this.$store.state.provincia },
      set(v) { this.$store.dispatch('aggiorna', { campo: 'provincia', valore: v }) },
    },
    classe: {
      get() { return this.$store.state.classeDiMerito },
      set(v) { this.$store.dispatch('aggiorna', { campo: 'classeDiMerito', valore: v }) },
    },
    anniPatente: {
      get() { return this.$store.state.anniPatente },
      set(v) { this.$store.dispatch('aggiorna', { campo: 'anniPatente', valore: v }) },
    },
    dati() { return this.$store.getters.dati },
    valido() { return this.$store.getters.valido },
  },

  watch: {
    // Ogni volta che i dati cambiano, si avvisa il resto della pagina.
    //
    // deep e immediate: `deep` perché si osserva un oggetto e non un valore
    // singolo, `immediate` per emettere subito lo stato iniziale — così un
    // altro step che si monta dopo non resta a mani vuote.
    dati: {
      handler(nuovi) {
        bus.$emit(EVENTI.DATI_CONDUCENTE, { dati: nuovi, valido: this.valido })
      },
      deep: true,
      immediate: true,
    },
  },

  mounted() {
    // Ri-emissione al montaggio: se questo step si idrata DOPO l'altro,
    // l'evento iniziale emesso in `immediate` è andato perso perché nessuno
    // era ancora in ascolto.
    //
    // È il limite di un bus di eventi puro: non ha memoria. Chi arriva tardi
    // non sa cosa si è detto prima. Le soluzioni sono ri-emettere (come qui),
    // oppure usare uno stato condiviso invece del bus — ed è esattamente il
    // motivo per cui in Poste lo stato vero sta nel Funnel Orchestrator e
    // non nell'EventBus.
    bus.$emit(EVENTI.DATI_CONDUCENTE, { dati: this.dati, valido: this.valido })
  },
}
</script>

<template>
  <!-- Vue 2: una sola radice, obbligatoria -->
  <div class="form-dati" :class="{ mobile: !web }">
    <h2>Dati del conducente</h2>

    <div class="griglia">
      <label>
        Età
        <input v-model.number="eta" type="number" min="18" max="100" />
      </label>

      <label>
        Provincia
        <input v-model="provincia" type="text" maxlength="2" placeholder="BO" />
      </label>

      <label>
        Classe di merito
        <input v-model.number="classe" type="number" min="1" max="18" />
        <small>1 è la migliore, 18 la peggiore</small>
      </label>

      <label>
        Anni di patente
        <input v-model.number="anniPatente" type="number" min="0" max="80" />
      </label>
    </div>

    <p class="stato" :class="valido ? 'ok' : 'ko'">
      {{ valido ? 'Dati completi' : 'Completa i dati per proseguire' }}
    </p>
  </div>
</template>

<style scoped>
.form-dati {
  padding: 1rem 1.25rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: #fff;
}

.form-dati.mobile .griglia {
  grid-template-columns: 1fr;
}

h2 {
  margin: 0 0 0.8rem;
  font-size: 1.05rem;
}

.griglia {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 0.8rem;
}

label {
  display: block;
  font-size: 0.85rem;
}

input {
  display: block;
  width: 100%;
  margin-top: 0.25rem;
  padding: 0.45rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 0.95rem;
}

small {
  color: #888;
  font-size: 0.72rem;
}

.stato {
  margin: 0.8rem 0 0;
  font-size: 0.8rem;
}

.stato.ok {
  color: #276749;
}

.stato.ko {
  color: #975a16;
}
</style>
