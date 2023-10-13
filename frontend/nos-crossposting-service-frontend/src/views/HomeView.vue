<template>
  <div class="home">
    <div v-if="loading">
      Loading...
    </div>

    <div v-if="!loading && !user">
      <Explanation/>
      <LogInWithTwitterButton/>
    </div>

    <div v-if="!loading && user">
      <CurrentUser :user="user"/>
    </div>

  </div>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {useStore} from "vuex";
import Explanation from '@/components/Explanation.vue';
import LogInWithTwitterButton from "@/components/LogInWithTwitterButton.vue";
import {User} from "@/dto/User";
import CurrentUser from "@/components/CurrentUser.vue";


@Options({
  components: {
    CurrentUser,
    LogInWithTwitterButton,
    Explanation,
  },
})
export default class HomeView extends Vue {

  private readonly store = useStore();

  get loading(): boolean {
    return this.store.state.user === undefined;
  }

  get user(): User {
    return this.store.state.user;
  }
}
</script>
