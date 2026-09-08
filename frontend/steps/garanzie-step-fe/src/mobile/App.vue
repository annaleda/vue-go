<script>
import SceltaGaranzie from '../components/SceltaGaranzie.vue'

export default {
  name: 'MobileApp',
  components: { SceltaGaranzie },
}
</script>

<template>
  <div class="step-mobile">
    <SceltaGaranzie :web="false" />
  </div>
</template>
