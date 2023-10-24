<template>
  <a href="/login" class="log-in-with-twitter-button">
    <div class="img-wrapper">
      <img src="../assets/login_with_x.svg">
    </div>
  </a>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';

@Options({})
export default class LogInWithTwitterButton extends Vue {
}
</script>

<style scoped lang="scss">
a {
  display: inline-block;
  margin: 1.5em 0;
  padding: 5px;
  border-radius: 10px;
  background: linear-gradient(#F08508, #F43F75);

  .img-wrapper {
    padding: 1em 2em;
    background-color: #170D2B;

    img {
      display: block;
      height: 52px;
    }
  }

  &:hover {
    .img-wrapper {
      background-color: transparent;
    }
  }
}
</style>
