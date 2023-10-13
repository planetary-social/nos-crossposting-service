<template>
  <div class="home">
    <Explanation/>
    <LogInWithTwitterButton/>
  </div>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import Explanation from '@/components/Explanation.vue';
import LogInWithTwitterButton from "@/components/LogInWithTwitterButton.vue";

@Options({
  components: {
    LogInWithTwitterButton,
    Explanation,
  },
})
export default class HomeView extends Vue {
}
</script>
