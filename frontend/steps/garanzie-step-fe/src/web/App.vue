<script>
import SceltaGaranzie from '../components/SceltaGaranzie.vue'

export default {
  name: 'WebApp',
  components: { SceltaGaranzie },
}
</script>

<template>
  <div class="step-web">
    <SceltaGaranzie :web="true" />
  </div>
</template>
