<script>
// Radice della vista WEB (desktop).
//
// Perché esistono due App quasi identiche, una web e una mobile?
// Perché lo stesso step deve rendersi in due contesti diversi: nel portale
// desktop e dentro l'app mobile, con layout diversi. Sono registrate come
// DUE VIEW HYPERNOVA DISTINTE, con bundle di hydration separati.
//
// La differenza qui è minima (una prop), ma nel progetto Poste i due layout
// divergono parecchio.

import FormDati from '../components/FormDati.vue'

export default {
  name: 'WebApp',
  components: { FormDati },
}
</script>

<template>
  <div class="step-web">
    <FormDati :web="true" />
  </div>
</template>
