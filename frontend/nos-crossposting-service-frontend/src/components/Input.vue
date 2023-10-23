<template>
  <input type="text" :placeholder="placeholder">
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';

@Options({
  props: {
    placeholder: String
  },
})
export default class Input extends Vue {
  placeholder!: string
}
</script>

<style scoped lang="scss">
input {
  border-radius: 10px;
  border: 3px solid rgba(255, 255,255, 0.15);
  padding: .5em;
  margin: 1em 0;
  background-color: #342255;
  min-width: 500px;
  color: rgba(255, 255, 255, 0.25);
  font-weight: 700;
  font-size: 32px;
}
</style>
