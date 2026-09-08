import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

// createStore è una FUNZIONE, non un'istanza. E non è un dettaglio di stile.
//
// In SSR il server gestisce più richieste nello stesso processo Node. Se lo
// store fosse un singleton creato all'import, tutte le richieste
// condividerebbero lo stesso stato: i dati di un utente finirebbero nella
// pagina di un altro.
//
// Ogni render deve avere il proprio store fresco, quindi si passa una
// factory e Hypernova la chiama a ogni richiesta.
export function createStore() {
  return new Vuex.Store({
    state: {
      etaConducente: 40,
      provincia: 'BO',
      classeDiMerito: 14,
      anniPatente: 15,
    },

    // Le mutation sono l'unico modo di cambiare lo stato, e sono
    // SINCRONE. È questa regola che permette agli strumenti di sviluppo di
    // ricostruire la storia dei cambiamenti.
    mutations: {
      setCampo(state, { campo, valore }) {
        state[campo] = valore
      },
    },

    // Le action possono essere asincrone e non toccano lo stato
    // direttamente: fanno commit di una mutation.
    actions: {
      aggiorna({ commit }, payload) {
        commit('setCampo', payload)
      },
    },

    // I getter sono i valori derivati: l'equivalente di computed, ma
    // sullo store.
    getters: {
      dati(state) {
        return {
          etaConducente: Number(state.etaConducente),
          provincia: String(state.provincia || '').trim().toUpperCase(),
          classeDiMerito: Number(state.classeDiMerito),
          anniPatente: Number(state.anniPatente),
        }
      },
      valido(state, getters) {
        const d = getters.dati
        return (
          d.etaConducente >= 18 &&
          d.etaConducente <= 100 &&
          d.classeDiMerito >= 1 &&
          d.classeDiMerito <= 18 &&
          d.provincia.length === 2
        )
      },
    },
  })
}
