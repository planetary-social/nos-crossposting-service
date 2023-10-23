<template>
  <input type="text" :placeholder="placeholder" :value="modelValue" @input="onInput">
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';

@Options({
  props: {
    placeholder: String,
    modelValue: String,
  },
  emits: [
    'update:modelValue',
  ],
})
export default class Input extends Vue {
  placeholder!: string
  modelValue!: string;

  onInput(event: InputEvent): void {
    this.$emit('update:modelValue', (event.target as HTMLInputElement).value);
  }
}
</script>

<style scoped lang="scss">
input {
  border-radius: 10px;
  border: 3px solid rgba(255, 255, 255, 0.15);
  padding: 16px;
  background-color: #342255;
  min-width: 500px;
  font-weight: 700;
  font-size: 32px;
  line-height: 32px;
  color: #fff;

  &::placeholder {
    color: rgba(255, 255, 255, 0.25);
  }
}
</style>
