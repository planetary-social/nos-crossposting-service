<template>
  <div class="home">
    <Explanation/>

    <div v-if="!loadingUser && !user">
      <div class="step">
        1. Link your X account:
      </div>

      <LogInWithTwitterButton/>
    </div>

    <div v-if="!loadingUser && user">
      <div class="step">
        1. Logged in as
      </div>

      <CurrentUser :user="user"/>
    </div>

    <div class="step">
      2. Your nostr identities:
    </div>

    <div class="public-keys-wrapper" v-if="!loadingUser && user">
      <div v-if="!publicKeys">
        Loading public keys...
      </div>

      <ul class="public-keys"
          v-if="publicKeys && publicKeys.publicKeys?.length > 0">
        <li v-for="publicKey in publicKeys.publicKeys" :key="publicKey.npub">
          <div class="npub">{{ publicKey.npub }}</div>
          <Checkmark></Checkmark>
        </li>
      </ul>
    </div>

    <form>
      <Input placeholder="Paste your npub address" v-model="npub"></Input>
      <Button text="Add" @click="addPublicKey"></Button>
    </form>
  </div>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {useStore} from "vuex";
import Explanation from '@/components/Explanation.vue';
import LogInWithTwitterButton from "@/components/LogInWithTwitterButton.vue";
import {User} from "@/dto/User";
import CurrentUser from "@/components/CurrentUser.vue";
import {APIService} from "@/services/APIService";
import {PublicKeys} from "@/dto/PublicKeys";
import {AddPublicKeyRequest} from "@/dto/AddPublicKeyRequest";
import Input from "@/components/Input.vue";
import Button from "@/components/Button.vue";
import Checkmark from "@/components/Checkmark.vue";


@Options({
  components: {
    Checkmark,
    Button,
    CurrentUser,
    LogInWithTwitterButton,
    Explanation,
    Input,
  },
})
export default class HomeView extends Vue {

  private readonly apiService = new APIService(useStore());
  private readonly store = useStore();

  publicKeys: PublicKeys | null = null;
  npub = "";

  get loadingUser(): boolean {
    return this.store.state.user === undefined;
  }

  get user(): User {
    return this.store.state.user;
  }

  created() {
    this.apiService.publicKeys()
        .then(response => {
          this.publicKeys = response.data;
        })
  }

  addPublicKey(): void {
    this.apiService.addPublicKey(new AddPublicKeyRequest(this.npub))
        .then(response => {
          console.log("added", response);
        })
        .catch(error => {
          console.log("error", error);
        })
  }
}
</script>

<style scoped lang="scss">
.step {
  font-size: 28px;
  margin-top: 2em;
}

form, .public-keys-wrapper {
  margin: 28px 0;
}

button {
  margin-left: 1em;
}

.public-keys {
  list-style-type: none;
  font-size: 28px;
  margin: 0;
  padding: 0;

  li {
    margin: 1em 0;
    padding: 0;
    color: #9379BF;
    font-style: normal;
    font-weight: 700;

    &:first-child {
      margin-top: 0;
    }

    &:last-child {
      margin-bottom: 0;
    }

    .npub, .checkmark {
      display: inline-block;
    }

    .npub {
      width: 300px;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}
</style>
