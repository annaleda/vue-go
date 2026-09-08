import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

Vue.use(Vuex)

// In sviluppo e in produzione le chiamate passano sempre da /api: è
// l'aggregator (o nginx, o il gateway) a girarle al servizio giusto.
const API = process.env.API_BASE || '/api'

export function createStore() {
  return new Vuex.Store({
    state: {
      garanzie: [],
      selezionate: ['RCA'],
      // I dati del conducente NON stanno qui dentro: arrivano dall'altro
      // micro-frontend attraverso l'EventBus. Questo step non ha idea di
      // dove siano stati raccolti, e non deve averla.
      datiConducente: null,
      datiValidi: false,
      preventivo: null,
      errore: null,
      caricamento: false,
    },

    mutations: {
      setGaranzie(state, g) { state.garanzie = g },
      setSelezionate(state, s) { state.selezionate = s },
      setDatiConducente(state, { dati, valido }) {
        state.datiConducente = dati
        state.datiValidi = valido
      },
      setPreventivo(state, p) { state.preventivo = p },
      setErrore(state, e) { state.errore = e },
      setCaricamento(state, v) { state.caricamento = v },
    },

    actions: {
      async caricaGaranzie({ commit }) {
        try {
          const { data } = await axios.get(`${API}/garanzie`)
          commit('setGaranzie', data)
        } catch (e) {
          commit('setErrore', 'Impossibile caricare le garanzie')
        }
      },

      async calcola({ state, commit }) {
        if (!state.datiValidi || !state.datiConducente) return null
        if (state.selezionate.length === 0) return null

        commit('setCaricamento', true)
        commit('setErrore', null)
        try {
          const { data } = await axios.post(`${API}/preventivi`, {
            ...state.datiConducente,
            garanzie: state.selezionate,
          })
          commit('setPreventivo', data)
          return data
        } catch (e) {
          const msg =
            (e.response && e.response.data && e.response.data.errore) ||
            'Errore nel calcolo del preventivo'
          commit('setErrore', msg)
          commit('setPreventivo', null)
          return null
        } finally {
          commit('setCaricamento', false)
        }
      },
    },
  })
}
