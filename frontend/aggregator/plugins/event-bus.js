import Vue from 'vue'

// L'aggregator deve usare LO STESSO bus degli step.
//
// La chiave su window è il contratto: se qui si scrivesse un nome diverso,
// l'aggregator avrebbe un bus tutto suo e non sentirebbe nulla di quello
// che gli step si dicono. In un progetto vero questa costante sta in una
// libreria condivisa, importata da tutti.
const CHIAVE = '__FEU_EVENT_BUS__'

if (!window[CHIAVE]) {
  window[CHIAVE] = new Vue()
}

export const bus = window[CHIAVE]

export const EVENTI = {
  DATI_CONDUCENTE: 'dati-conducente-cambiati',
  GARANZIE_SCELTE: 'garanzie-scelte',
  PREVENTIVO_CALCOLATO: 'preventivo-calcolato',
  STEP_COMPLETATO: 'step-completato',
}

// inject rende il bus disponibile in ogni componente come this.$bus
export default (context, inject) => {
  inject('bus', bus)
  inject('eventi', EVENTI)
}
