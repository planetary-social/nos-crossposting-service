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


@Options({
  components: {
    CurrentUser,
    LogInWithTwitterButton,
    Explanation,
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
