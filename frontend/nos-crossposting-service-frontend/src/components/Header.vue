<template>
  <header>
    <a class="logo" href="/">
      <img src="../assets/logo.svg"/>
    </a>
    <nav>
      <ul>
        <li>
          <a href="https://nos.social">Download Nos</a>
        </li>
        <li>
          <a href="https://github.com/planetary-social/nos-crossposting-service">Source code</a>
        </li>
      </ul>
    </nav>
    <LogoutButton v-if="userIsLoggedIn"></LogoutButton>
  </header>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {useStore} from "vuex";
import LogoutButton from "@/components/LogoutButton.vue";

@Options({
  components: {LogoutButton}
})
export default class Header extends Vue {
  private readonly store = useStore();

  get userIsLoggedIn(): boolean {
    return !!this.store.state.user;
  }
}
</script>

<style scoped lang="scss">
header {
  display: flex;
  flex-flow: row nowrap;
  align-items: center;

  .logo {
    display: block;
  }

  nav {
    font-size: 24px;
    font-weight: 600;
    flex: 1;
    text-align: right;

    ul {
      margin: 0;
      padding: 0;
      list-style-type: none;

      li {
        display: inline-block;
        margin: 0;
        padding: 0 1em;

        &:last-child {
          padding-right: 0;
        }

        &:first-child {
          padding-left: 0;
        }

        a {
          color: #9379BF;
          text-decoration: none;

          &:hover {
            color: #fff;
          }
        }
      }
    }
  }

  .logout-button {
    margin-left: 1em;
  }
}

@media screen and (max-width: 1200px) {
  header {
    flex-flow: column nowrap;
    text-align: center;

    nav {
      margin: 1em 0;
      text-align: center;

      ul {
        li {
          display: block;
          padding: .5em 0;
        }
      }
    }

    .logout-button {
      margin: 1em 0;
    }
  }
}

</style>
