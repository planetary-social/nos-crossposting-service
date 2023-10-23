<template>
  <div class="home">
    <Explanation/>

    <div v-if="!loading && !user">
      <div class="step">
        1. Link your X account:
      </div>

      <LogInWithTwitterButton/>
    </div>

    <div v-if="!loading && user">
      <div class="step">
        1. Logged in as
      </div>

      <CurrentUser :user="user"/>
    </div>

    <div class="step">
      2. Your nostr identities:
    </div>

    <Input placeholder="Paste your npub address"></Input>

    <div v-if="loading">
      Loading...
    </div>

    <div v-if="!loading && user">
      <div v-if="!publicKeys">
        Loading public keys...
      </div>

      <div v-if="publicKeys">
        <ul v-if="publicKeys.publicKeys?.length > 0">
          <li v-for="publicKey in publicKeys.publicKeys" :key="publicKey.npub">
            {{ publicKey.npub }}
          </li>
        </ul>

        <p v-if="publicKeys.publicKeys?.length == 0">
          You haven't added any public keys yet.
        </p>
      </div>

      <input placeholder="npub..." v-model="npub">
      <button @click="addPublicKey">Link public key</button>
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
import {APIService} from "@/services/APIService";
import {PublicKeys} from "@/dto/PublicKeys";
import {AddPublicKeyRequest} from "@/dto/AddPublicKeyRequest";
import Input from "@/components/Input.vue";


@Options({
  components: {
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

  get loading(): boolean {
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
</style>
