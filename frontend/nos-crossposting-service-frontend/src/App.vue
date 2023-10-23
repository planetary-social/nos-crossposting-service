<template>
  <div class="wrapper">
    <Header></Header>

    {{ user }}
    <router-view/>
  </div>
</template>

<style lang="scss">
html, body {
  padding: 0;
  margin: 0;
  color: #fff;
  height: 100%;
}

html {
  background: linear-gradient(#160f24, #281945);
}

body {
  background-image: url("./assets/background.svg");
  background-position: right center;
  background-repeat: no-repeat;
  background-size: auto 100%;
}

#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;

  .wrapper {
    padding: 4em;
  }
}
</style>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {useStore} from "vuex";
import {APIService} from "@/services/APIService";
import Header from "@/components/Header.vue";

@Options({
  components: {
    Header,
  }
})
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
