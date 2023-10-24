<template>
  <input type="text" :placeholder="placeholder" :value="modelValue"
         @input="onInput" :class="{ disabled: disabled }" :disabled="disabled">
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';

@Options({
  props: {
    placeholder: String,
    modelValue: String,
    disabled: Boolean,
  },
  emits: [
    'update:modelValue',
  ],
})
export default class Input extends Vue {
  placeholder!: string
  modelValue!: string;
  disabled!: boolean;

  onInput(event: InputEvent): void {
    this.$emit('update:modelValue', (event.target as HTMLInputElement).value);
  }
}
</script>

<style scoped lang="scss">
input {
  border-radius: 10px;
  border: 3px solid #fff;
  padding: 30px;
  background-color: #1B122D;
  min-width: 500px;
  font-weight: 700;
  font-size: 32px;
  line-height: 32px;
  color: #fff;

  &::placeholder {
    color: #fff;
  }

  &.disabled {
    cursor: not-allowed;
    background-color: #342255;
    border-color: rgba(255, 255, 255, 0.15);

    &::placeholder {
      color: rgba(255, 255, 255, 0.25);
    }
  }
}
</style>
