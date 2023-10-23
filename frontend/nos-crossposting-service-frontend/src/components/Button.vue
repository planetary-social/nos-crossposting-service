<template>
  <button @click="$emit('click', $event)">{{ text }}</button>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';

@Options({
  props: {
    text: String
  },
})
export default class Button extends Vue {
  text!: string
}
</script>

<style scoped lang="scss">
button {
  background-image: linear-gradient(180deg, #F08508 0%, #F43F75 100%);
  padding: 19px; // input has 16px padding + 3px border
  border: 0;
  border-radius: 10px;
  color: #FFF;
  font-size: 32px;
  font-weight: 700;

  &:hover {
    cursor: pointer;
  }
}
</style>
