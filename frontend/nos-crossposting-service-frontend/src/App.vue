<template>
  <div class="wrapper">
    <Header></Header>
    <div class="content">
      <Notifications></Notifications>
      <router-view/>
    </div>
  </div>
</template>

<style lang="scss">
@font-face {
  font-family: 'Clarity City';
  font-style: normal;
  font-weight: 400;
  font-display: swap;
  src: url("./assets/fonts/claritycity/ClarityCity-Regular.otf") format('opentype');
}

@font-face {
  font-family: 'Clarity City';
  font-style: normal;
  font-weight: 500;
  font-display: swap;
  src: url("./assets/fonts/claritycity/ClarityCity-Medium.otf") format('opentype');
}

@font-face {
  font-family: 'Clarity City';
  font-style: normal;
  font-weight: 600;
  font-display: swap;
  src: url("./assets/fonts/claritycity/ClarityCity-SemiBold.otf") format('opentype');
}

@font-face {
  font-family: 'Clarity City';
  font-style: normal;
  font-weight: 700;
  font-display: swap;
  src: url("./assets/fonts/claritycity/ClarityCity-Bold.otf") format('opentype');
}

html, body {
  padding: 0;
  margin: 0;
  color: #fff;
  min-height: 100vh;

  font-family: 'Clarity City', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
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
  .wrapper {
    padding: 4em;

    .content {
      padding: 0 4em;
    }
  }
}

@media screen and (max-width: 1200px) {
  #app {
    .wrapper {
      .content {
        padding: 0;
      }
    }
  }
}
</style>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {useStore} from "vuex";
import {APIService} from "@/services/APIService";
import Header from "@/components/Header.vue";
import Notifications from "@/components/Notifications.vue";
import {Mutation} from "@/store";

@Options({
  components: {
    Notifications,
    Header,
  }
})
export default class App extends Vue {
  private readonly apiService = new APIService(useStore());
  private readonly store = useStore();

  created(): void {
    this.apiService.refreshCurrentUser()
        .catch(() => {
          this.store.commit(Mutation.PushNotificationError, "Error loading the user.");
        });
  }
}
</script>
