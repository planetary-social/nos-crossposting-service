<template>
  <h1>
    Nos crossposting service
    {{ user }}
  </h1>
  <router-view/>
</template>

<style lang="scss">
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

nav {
  padding: 30px;

  a {
    font-weight: bold;
    color: #2c3e50;

    &.router-link-exact-active {
      color: #42b983;
    }
  }
}
</style>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {useStore} from "vuex";
import {APIService} from "@/services/APIService";

@Options({})
export default class App extends Vue {

  private readonly apiService = new APIService(useStore());

  created() {
    this.loadCurrentUser();
  }

  private loadCurrentUser(): void {
    this.apiService.refreshCurrentUser()
  }
}
</script>
