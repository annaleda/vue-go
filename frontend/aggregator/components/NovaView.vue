<script>
// Il "ponte di hydration".
//
// Riceve due cose da un micro-frontend renderizzato lato server:
//   html    il markup già pronto, prodotto da Hypernova
//   script  l'URL del bundle che riporta in vita quel markup
//
// Le mette entrambe in pagina. Da quel momento il componente Vue di quel
// micro-frontend è vivo nel browser, con il suo store e i suoi eventi —
// ma questo componente non sa nulla di cosa contenga.
//
// È l'equivalente di NovaHike.vue nell'aggregator di Poste.

export default {
  name: 'NovaView',

  props: {
    html: { type: String, required: true },
    script: { type: String, default: '' },
    nome: { type: String, default: '' },
  },
}
</script>

<template>
  <div class="nova-view">
    <!--
      v-html inserisce HTML grezzo senza escaparlo.

      È pericoloso in generale (è la porta d'ingresso classica per XSS) ed è
      accettabile qui solo perché il markup arriva da un nostro servizio
      interno, chiamato dal server, non da un input dell'utente. Se questo
      HTML potesse essere influenzato da chi naviga, andrebbe sanificato.
    -->
    <div v-html="html"></div>

    <!--
      Il bundle di hydration va caricato DOPO che il markup è in pagina,
      perché lo script cerca il proprio nodo per chiave e id.

      `:is="'script'"` è il modo di far generare a Vue un tag <script>:
      scritto direttamente nel template verrebbe interpretato dal compilatore
      invece che emesso. `async` evita di bloccare il rendering del resto.
    -->
    <component v-if="script" :is="'script'" :src="script" async></component>
  </div>
</template>

<style scoped>
.nova-view {
  margin-bottom: 1rem;
}
</style>
