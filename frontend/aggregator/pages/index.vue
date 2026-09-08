<script>
import axios from 'axios'

export default {
  name: 'PaginaFunnel',

  // asyncData gira SUL SERVER al primo caricamento.
  //
  // È qui che avviene la composizione: prima che il browser riceva qualcosa,
  // l'aggregator interroga tutti gli step e ne raccoglie l'HTML.
  //
  // Le chiamate partono in parallelo con Promise.all: farle in sequenza
  // significherebbe sommare le latenze, e con dieci step si vedrebbe.
  async asyncData({ $config, req }) {
    const steps = ($config && $config.steps) || []
    const canale = 'web'

    const richieste = steps.map(async (step) => {
      const nomeVista = canale === 'web' ? step.viewWeb : step.viewMobile
      try {
        // Il protocollo Hypernova: POST /batch con una mappa di viste.
        // Si possono chiedere più viste in una sola chiamata — da qui il
        // nome "batch".
        const { data } = await axios.post(
          `${step.endpoint}/batch`,
          { [step.id]: { name: nomeVista, data: {} } },
          { timeout: 5000 }
        )

        const risultato = data && data.results && data.results[step.id]
        if (!risultato || risultato.error) {
          return { id: step.id, nome: nomeVista, html: null, script: null }
        }
        return {
          id: step.id,
          nome: nomeVista,
          html: risultato.html,
          // meta.src è l'URL del bundle di hydration, dichiarato dallo step
          script: (risultato.meta && risultato.meta.src) || null,
        }
      } catch (e) {
        // Uno step che non risponde non deve far cadere l'intera pagina:
        // si mostra il resto e si segnala il buco. È il vantaggio pratico
        // della composizione — il guasto resta circoscritto.
        return { id: step.id, nome: nomeVista, html: null, script: null, errore: true }
      }
    })

    return { viste: await Promise.all(richieste) }
  },

  data() {
    return {
      totale: null,
      completati: [],
    }
  },

  // mounted gira solo nel browser: qui l'aggregator si mette in ascolto
  // di quello che gli step si dicono fra loro.
  mounted() {
    this.$bus.$on(this.$eventi.PREVENTIVO_CALCOLATO, this.onPreventivo)
    this.$bus.$on(this.$eventi.STEP_COMPLETATO, this.onStepCompletato)
  },

  beforeDestroy() {
    this.$bus.$off(this.$eventi.PREVENTIVO_CALCOLATO, this.onPreventivo)
    this.$bus.$off(this.$eventi.STEP_COMPLETATO, this.onStepCompletato)
  },

  methods: {
    onPreventivo(prev) {
      this.totale = prev && prev.totale
    },
    onStepCompletato({ step }) {
      if (!this.completati.includes(step)) this.completati.push(step)
    },
    euro(v) {
      return Number(v).toFixed(2).replace('.', ',') + ' €'
    },
  },
}
</script>

<template>
  <div class="funnel">
    <header class="barra">
      <div class="titolo">
        <strong>Preventivo auto</strong>
        <span class="nota">composto da {{ viste.length }} micro-frontend</span>
      </div>

      <!--
        Questo totale è la dimostrazione che la comunicazione funziona:
        l'aggregator non ha calcolato nulla e non ha chiamato nessuna API.
        Ha solo ascoltato un evento emesso da uno step che non conosce.
      -->
      <div v-if="totale !== null" class="totale">
        Totale: <strong>{{ euro(totale) }}</strong>
      </div>
    </header>

    <main class="contenuto">
      <template v-for="vista in viste">
        <NovaView
          v-if="vista.html"
          :key="vista.id"
          :nome="vista.nome"
          :html="vista.html"
          :script="vista.script"
        />
        <div v-else :key="vista.id + '-err'" class="guasto">
          Lo step <code>{{ vista.id }}</code> non è disponibile.
        </div>
      </template>
    </main>

    <footer class="pie">
      <span v-if="completati.length">Step completati: {{ completati.join(', ') }}</span>
      <span v-else>Nessuno step completato</span>
    </footer>
  </div>
</template>

<style scoped>
.funnel {
  max-width: 780px;
  margin: 0 auto;
  padding: 1.5rem 1rem 3rem;
}

.barra {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  padding: 0.9rem 1.1rem;
  margin-bottom: 1.2rem;
  border-radius: 6px;
  background: #1f2937;
  color: #fff;
}

.nota {
  display: block;
  font-size: 0.78rem;
  color: #9ca3af;
}

.totale {
  font-size: 1.05rem;
}

.guasto {
  padding: 0.8rem 1rem;
  margin-bottom: 1rem;
  border: 1px dashed #c0392b;
  border-radius: 6px;
  background: #fdf0ef;
  color: #922;
  font-size: 0.88rem;
}

.pie {
  margin-top: 1.2rem;
  color: #6b7280;
  font-size: 0.8rem;
}
</style>
