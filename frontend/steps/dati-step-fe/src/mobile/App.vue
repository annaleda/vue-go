<script>
import FormDati from '../components/FormDati.vue'

export default {
  name: 'MobileApp',
  components: { FormDati },
}
</script>

<template>
  <div class="step-mobile">
    <FormDati :web="false" />
  </div>
</template>
