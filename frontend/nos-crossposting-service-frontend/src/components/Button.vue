<template>
  <button @click="onClick" :class="{ disabled: disabled }">{{ text }}</button>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';

@Options({
  props: {
    text: String,
    disabled: Boolean,
  },
})
export default class Button extends Vue {
  text!: string;
  disabled!: boolean;

  onClick(event: any): void {
    this.$emit('click', event)
  }
}
</script>

<style scoped lang="scss">
button {
  background-image: linear-gradient(180deg, #F08508 0%, #F43F75 100%);
  padding: 33px; // input has 30px padding + 3px border
  border: 0;
  border-radius: 10px;
  color: #FFF;
  font-size: 32px;
  font-weight: 700;
  text-shadow: 0px 2px 2px rgba(0, 0, 0, 0.25);

  &:hover {
    cursor: pointer;
  }

  &.disabled {
    cursor: not-allowed;
  }
}
</style>
