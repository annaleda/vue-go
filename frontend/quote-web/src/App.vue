<script>
// Questo componente usa la OPTIONS API: lo stato sta in data(), la logica in
// methods, i cicli di vita in mounted() e così via.
//
// È scritto così di proposito: è la forma di Vue 2, quella che trovi nei
// micro-frontend Poste. L'alternativa moderna (Composition API) è mostrata
// in RisultatoPreventivo.vue, così vedi le due sintassi affiancate.

import FormPreventivo from './components/FormPreventivo.vue'
import RisultatoPreventivo from './components/RisultatoPreventivo.vue'
import { caricaGaranzie, calcolaPreventivo } from './api.js'

export default {
  name: 'App',

  // I componenti figli vanno dichiarati per poterli usare nel template.
  components: { FormPreventivo, RisultatoPreventivo },

  // data() DEVE essere una funzione che restituisce un oggetto, non un
  // oggetto. Se fosse un oggetto, tutte le istanze del componente
  // condividerebbero lo stesso stato.
  //
  // Tutto ciò che sta qui dentro è REATTIVO: se cambia, la parte di
  // interfaccia che lo usa si ridisegna da sola. Non serve dire a Vue
  // "aggiorna": è la differenza principale rispetto a manipolare il DOM.
  data() {
    return {
      garanzie: [],
      preventivo: null,
      errore: null,
      caricamento: false,
    }
  },

  // mounted() scatta quando il componente è stato inserito nella pagina.
  // È il punto giusto per le chiamate HTTP iniziali.
  async mounted() {
    try {
      this.garanzie = await caricaGaranzie()
    } catch (e) {
      this.errore = `Impossibile caricare le garanzie: ${e.message}`
    }
  },

  methods: {
    // Riceve la richiesta dal form figlio (vedi l'evento nel template).
    async onCalcola(richiesta) {
      this.caricamento = true
      this.errore = null
      this.preventivo = null
      try {
        this.preventivo = await calcolaPreventivo(richiesta)
      } catch (e) {
        this.errore = e.message
      } finally {
        // finally viene eseguito sempre: senza, un errore lascerebbe
        // l'interfaccia bloccata sullo stato "sto caricando".
        this.caricamento = false
      }
    },
  },
}
</script>

<template>
  <main class="pagina">
    <header>
      <h1>Preventivo assicurazione auto</h1>
      <p class="sottotitolo">
        Vue 3 nel browser, due servizi Go dietro.
      </p>
    </header>

    <!--
      Comunicazione tra componenti, in due direzioni:

      :garanzie="garanzie"       PROP: dal padre al figlio (dati che scendono)
      @calcola="onCalcola"       EVENTO: dal figlio al padre (notifiche che salgono)

      È la regola fondamentale: un figlio non modifica mai direttamente i dati
      del padre, gli manda un evento e il padre decide. Quando i componenti da
      coordinare diventano tanti si passa a uno store condiviso (Vuex o Pinia),
      che è quello che fa il progetto Poste con Vuex.
    -->
    <FormPreventivo
      :garanzie="garanzie"
      :disabilitato="caricamento"
      @calcola="onCalcola"
    />

    <!-- v-if aggiunge e rimuove l'elemento dal DOM in base alla condizione -->
    <p v-if="caricamento" class="stato">Calcolo in corso…</p>
    <p v-if="errore" class="errore">{{ errore }}</p>

    <RisultatoPreventivo v-if="preventivo" :preventivo="preventivo" />
  </main>
</template>

<style scoped>
/*
  scoped significa che questo CSS vale SOLO per questo componente: Vue
  aggiunge un attributo univoco agli elementi e lo usa nei selettori.
  Senza scoped, .pagina qui dentro colpirebbe qualunque .pagina dell'app.
*/
.pagina {
  max-width: 720px;
  margin: 0 auto;
  padding: 2rem 1rem 4rem;
}

h1 {
  margin-bottom: 0.25rem;
  font-size: 1.6rem;
}

.sottotitolo {
  margin-top: 0;
  color: #666;
  font-size: 0.9rem;
}

.stato {
  color: #555;
}

.errore {
  padding: 0.75rem 1rem;
  border-left: 3px solid #c0392b;
  background: #fdf0ef;
  color: #922;
}
</style>
