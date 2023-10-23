<template>
  <div class="explanation">
    <div class="large">
      Start posting your Nostr content on X
    </div>
    <div class="small">
      Log in to your X account and add your Nostr identities so that everything you publish on Nostr will be mirrored on X
    </div>
  </div>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';

@Options({})
export default class Explanation extends Vue {
}
</script>

<style scoped lang="scss">
.explanation {
  text-align: left;
  margin: 10em 0;
  max-width: 700px;

  .large {
    font-size: 60px;
    line-height: 84px;
    font-weight: 700;
  }

  .small {
    color: #9379BF;
    font-size: 26px;
    font-weight: 500;
    line-height: 36px;
  }
}
</style>
