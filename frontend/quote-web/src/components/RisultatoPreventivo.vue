<script setup>
// Questo componente usa la COMPOSITION API con <script setup>.
//
// È lo stile moderno di Vue 3, e vale la pena vederlo affiancato alla
// Options API degli altri due componenti. Le differenze pratiche:
//
//   Options API                    Composition API
//   -------------------------      --------------------------------
//   props: { ... }                 const props = defineProps({ ... })
//   data() { return { x: 1 } }     const x = ref(1)   -> si legge x.value
//   computed: { y() {...} }        const y = computed(() => ...)
//   methods: { f() {...} }         function f() {...}
//   this.x                         x.value  (nel template basta x)
//
// Tutto ciò che è dichiarato qui dentro è automaticamente disponibile nel
// template: non serve esportare niente.
//
// Quando conviene: nei componenti grandi la Options API sparpaglia la
// logica di una stessa funzionalità tra data, computed e methods, mentre
// con la Composition API la tieni tutta insieme.

import { computed } from 'vue'

const props = defineProps({
  preventivo: { type: Object, required: true },
})

// computed() prende una funzione e restituisce un valore reattivo: si
// ricalcola solo quando cambia props.preventivo.
const rate = computed(() => {
  const t = props.preventivo.totale
  return {
    annuale: t,
    semestrale: t / 2,
    mensile: t / 12,
  }
})

const etichette = {
  eta: 'Età del conducente',
  classe: 'Classe di merito',
  provincia: 'Provincia di residenza',
  patente: 'Anzianità di patente',
}

function euro(v) {
  return v.toFixed(2).replace('.', ',') + ' €'
}

// Un coefficiente sopra 1 aumenta il premio, sotto 1 lo riduce.
function segno(v) {
  if (v > 1.001) return 'aumenta'
  if (v < 0.999) return 'riduce'
  return 'neutro'
}
</script>

<template>
  <section class="risultato">
    <h2>Preventivo</h2>

    <table class="voci">
      <thead>
        <tr>
          <th>Garanzia</th>
          <th class="num">Base</th>
          <th class="num">Prezzo</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="v in preventivo.voci" :key="v.codice">
          <td>{{ v.nome }}</td>
          <td class="num base">{{ euro(v.prezzoBase) }}</td>
          <td class="num">{{ euro(v.prezzo) }}</td>
        </tr>
      </tbody>
      <tfoot>
        <tr>
          <td colspan="2">Imponibile</td>
          <td class="num">{{ euro(preventivo.imponibile) }}</td>
        </tr>
        <tr>
          <td colspan="2">Imposte</td>
          <td class="num">{{ euro(preventivo.imposte) }}</td>
        </tr>
        <tr class="totale">
          <td colspan="2">Totale annuo</td>
          <td class="num">{{ euro(preventivo.totale) }}</td>
        </tr>
      </tfoot>
    </table>

    <div class="rate">
      <span>Semestrale: <strong>{{ euro(rate.semestrale) }}</strong></span>
      <span>Mensile: <strong>{{ euro(rate.mensile) }}</strong></span>
    </div>

    <details class="dettaglio">
      <summary>Come è stato calcolato</summary>
      <p class="nota">
        I coefficienti si applicano solo alla RCA. Le garanzie accessorie
        hanno un prezzo fisso, indipendente dal profilo di rischio.
      </p>
      <ul>
        <li v-for="(valore, chiave) in preventivo.coefficienti" :key="chiave">
          {{ etichette[chiave] || chiave }}:
          <strong>&times;{{ valore.toFixed(2) }}</strong>
          <span :class="['effetto', segno(valore)]">{{ segno(valore) }}</span>
        </li>
      </ul>
    </details>
  </section>
</template>

<style scoped>
.risultato {
  margin-top: 1.5rem;
  padding: 1.25rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: #fafafa;
}

h2 {
  margin-top: 0;
  font-size: 1.15rem;
}

.voci {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.voci th,
.voci td {
  padding: 0.45rem 0.3rem;
  border-bottom: 1px solid #eee;
  text-align: left;
}

.num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.base {
  color: #999;
}

.totale td {
  border-top: 2px solid #ccc;
  border-bottom: 0;
  font-weight: 700;
  font-size: 1rem;
}

.rate {
  display: flex;
  gap: 1.5rem;
  margin-top: 0.8rem;
  color: #555;
  font-size: 0.85rem;
}

.dettaglio {
  margin-top: 1rem;
  font-size: 0.85rem;
}

.dettaglio summary {
  cursor: pointer;
  color: #1f6feb;
}

.nota {
  color: #666;
}

.dettaglio ul {
  margin: 0.4rem 0 0;
  padding-left: 1.1rem;
}

.dettaglio li {
  margin-bottom: 0.2rem;
}

.effetto {
  margin-left: 0.4rem;
  font-size: 0.75rem;
  padding: 0.05rem 0.4rem;
  border-radius: 3px;
}

.effetto.aumenta {
  background: #fdeaea;
  color: #a33;
}

.effetto.riduce {
  background: #e9f7ee;
  color: #276749;
}

.effetto.neutro {
  background: #eee;
  color: #666;
}
</style>
