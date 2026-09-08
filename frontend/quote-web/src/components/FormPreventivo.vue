<script>
// Options API, come App.vue.
//
// Questo componente mostra le tre cose che si usano di più in Vue:
//   - v-model     per legare un campo del form a una variabile
//   - v-for       per ripetere un elemento su una lista
//   - computed    per un valore derivato che si ricalcola da solo

export default {
  name: 'FormPreventivo',

  // Le props sono i dati che arrivano dal padre. Dichiararle con il tipo
  // e il valore di default non è obbligatorio, ma Vue avvisa in console
  // quando qualcuno passa la cosa sbagliata: vale i trenta secondi che costa.
  props: {
    garanzie: { type: Array, default: () => [] },
    disabilitato: { type: Boolean, default: false },
  },

  // Dichiarare gli eventi emessi è documentazione eseguibile: chi legge il
  // componente sa cosa aspettarsi senza cercare gli $emit nel codice.
  emits: ['calcola'],

  data() {
    return {
      etaConducente: 40,
      provincia: 'BO',
      classeDiMerito: 14,
      anniPatente: 15,
      selezionate: ['RCA'],
    }
  },

  computed: {
    // Una computed è una funzione che si comporta come un dato: Vue la
    // ricalcola solo quando cambia qualcosa da cui dipende, e mette in
    // cache il risultato. Da preferire sempre a un metodo chiamato nel
    // template, che verrebbe rieseguito a ogni ridisegno.
    valido() {
      return (
        this.etaConducente >= 18 &&
        this.etaConducente <= 100 &&
        this.classeDiMerito >= 1 &&
        this.classeDiMerito <= 18 &&
        this.provincia.trim().length === 2 &&
        this.selezionate.length > 0
      )
    },

    stimaIndicativa() {
      return this.garanzie
        .filter((g) => this.selezionate.includes(g.codice))
        .reduce((somma, g) => somma + g.prezzoBase, 0)
    },
  },

  methods: {
    invia() {
      if (!this.valido) return

      // $emit manda un evento al padre. Il padre lo intercetta con
      // @calcola="onCalcola" e riceve questo oggetto come argomento.
      this.$emit('calcola', {
        etaConducente: Number(this.etaConducente),
        provincia: this.provincia.trim().toUpperCase(),
        classeDiMerito: Number(this.classeDiMerito),
        anniPatente: Number(this.anniPatente),
        garanzie: this.selezionate,
      })
    },
  },
}
</script>

<template>
  <!-- .prevent evita che il browser ricarichi la pagina inviando il form -->
  <form class="form" @submit.prevent="invia">
    <fieldset :disabled="disabilitato">
      <legend>Dati del conducente</legend>

      <div class="griglia">
        <label>
          Età
          <!--
            v-model lega il campo alla variabile nei due sensi: scrivi nel
            campo e la variabile cambia, cambi la variabile e il campo si
            aggiorna. È zucchero sintattico per :value + @input.

            .number converte la stringa in numero: senza, "40" resterebbe
            testo e i confronti numerici darebbero risultati sorprendenti.
          -->
          <input v-model.number="etaConducente" type="number" min="18" max="100" />
        </label>

        <label>
          Provincia
          <input v-model="provincia" type="text" maxlength="2" placeholder="BO" />
        </label>

        <label>
          Classe di merito
          <input v-model.number="classeDiMerito" type="number" min="1" max="18" />
          <small>1 è la migliore, 18 la peggiore</small>
        </label>

        <label>
          Anni di patente
          <input v-model.number="anniPatente" type="number" min="0" max="80" />
        </label>
      </div>
    </fieldset>

    <fieldset :disabled="disabilitato">
      <legend>Garanzie</legend>

      <p v-if="garanzie.length === 0" class="vuoto">
        Catalogo non ancora caricato…
      </p>

      <!--
        v-for ripete l'elemento per ogni voce della lista.

        :key serve a Vue per riconoscere gli elementi tra un ridisegno e
        l'altro. Senza, in liste che cambiano ordine Vue riusa i nodi
        sbagliati e ti ritrovi le checkbox spuntate sulla riga sbagliata.
        Va messa sempre, e deve essere stabile e univoca.
      -->
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
        <span class="prezzo">{{ g.prezzoBase.toFixed(2) }} &euro;</span>
      </label>
    </fieldset>

    <div class="azioni">
      <span class="stima" v-if="stimaIndicativa > 0">
        Prezzo base selezionato: {{ stimaIndicativa.toFixed(2) }} &euro;
      </span>
      <button type="submit" :disabled="!valido || disabilitato">
        Calcola preventivo
      </button>
    </div>
  </form>
</template>

<style scoped>
.form {
  margin: 1.5rem 0;
}

fieldset {
  border: 1px solid #ddd;
  border-radius: 6px;
  margin-bottom: 1rem;
  padding: 1rem;
}

legend {
  padding: 0 0.4rem;
  font-weight: 600;
  font-size: 0.9rem;
}

.griglia {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 0.9rem;
}

label {
  display: block;
  font-size: 0.85rem;
}

input[type='number'],
input[type='text'] {
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
  font-size: 0.75rem;
}

.garanzia {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.35rem 0;
  border-bottom: 1px solid #f0f0f0;
}

.garanzia .nome {
  flex: 1;
}

.garanzia em {
  color: #888;
  font-size: 0.78rem;
}

.garanzia .prezzo {
  color: #555;
  font-variant-numeric: tabular-nums;
}

.vuoto {
  color: #888;
  font-size: 0.85rem;
}

.azioni {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.stima {
  color: #666;
  font-size: 0.85rem;
}

button {
  padding: 0.6rem 1.2rem;
  border: 0;
  border-radius: 4px;
  background: #1f6feb;
  color: #fff;
  font-size: 0.95rem;
  cursor: pointer;
}

button:disabled {
  background: #bbb;
  cursor: not-allowed;
}
</style>
