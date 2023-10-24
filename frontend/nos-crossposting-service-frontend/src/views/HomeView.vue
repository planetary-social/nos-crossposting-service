<template>
  <div class="home">
    <Explanation/>

    <div v-if="!loadingUser && !user">
      <div class="step">
        <div class="text">
          1. Link your X account:
        </div>
      </div>

      <LogInWithTwitterButton/>
    </div>

    <div v-if="!loadingUser && user">
      <div class="step">
        <div class="text">
          1. Logged in as
        </div>
      </div>

      <CurrentUser :user="user"/>
    </div>

    <div class="step">
      <div class="text">
        2. Your nostr identities:
      </div>
      <ul class="actions">
        <li v-if="publicKeys && publicKeys.publicKeys?.length > 0 && !editingPublicKeys">
          <a @click="startEditingPublicKeys">
            Edit
          </a>
        </li>
        <li v-if="editingPublicKeys">
          <a @click="endEditingPublicKeys">
            Done
          </a>
        </li>
      </ul>
    </div>

    <div class="public-keys-wrapper" v-if="!loadingUser && user">
      <div v-if="!publicKeys">
        Loading public keys...
      </div>

      <ul class="public-keys"
          v-if="publicKeys && publicKeys.publicKeys?.length > 0">
        <li v-for="publicKey in publicKeys.publicKeys" :key="publicKey.npub">
          <a @click="scheduleDelete(publicKey)" v-if="editingPublicKeys"
             class="delete-public-key-button">
            <img src="../assets/delete.svg"/>
          </a>
          <div class="npub">{{ publicKey.npub }}</div>
          <Checkmark v-if="!editingPublicKeys"></Checkmark>
        </li>
      </ul>
    </div>

    <div class="link-npub-form">
      <Input placeholder="Paste your npub address" v-model="npub"
             :disabled="formDisabled"></Input>
      <Button text="Add" @buttonClick="addPublicKey"
              :disabled="formDisabled"></Button>
    </div>
  </div>
</template>

<script lang="ts">
import {Watch} from 'vue-property-decorator'
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
import {Mutation} from "@/store";
import {PublicKey} from "@/dto/PublicKey";


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
  editingPublicKeys = false;
  publicKeysToRemove: PublicKey[] = [];

  get loadingUser(): boolean {
    return this.store.state.user === undefined;
  }

  get user(): User {
    return this.store.state.user;
  }

  get formDisabled(): boolean {
    return this.loadingUser || !this.user;
  }

  @Watch('user')
  watchUser(newUser: CurrentUser): void {
    if (newUser) {
      this.reloadPublicKeys();
    } else {
      this.publicKeys = {publicKeys: []};
    }
  }

  startEditingPublicKeys(): void {
    this.publicKeysToRemove = [];
    this.editingPublicKeys = true;
  }

  endEditingPublicKeys(): void {
    for (const publicKey of this.publicKeysToRemove) {
      this.apiService.deletePublicKey(publicKey)
          .catch(() =>
              this.store.commit(Mutation.PushNotificationError, "Error removing a public key.")
          );
    }

    this.publicKeysToRemove = [];
    this.editingPublicKeys = false;
    this.reloadPublicKeys();
  }

  scheduleDelete(publicKey: PublicKey): void {
    const index = this.publicKeys?.publicKeys?.indexOf(publicKey);
    if (index !== undefined && index >= 0) {
      this.publicKeys?.publicKeys?.splice(index, 1);
      // force vue update
      this.publicKeys = {
        publicKeys: [...this.publicKeys?.publicKeys || []],
      }
    }
    this.publicKeysToRemove.push(publicKey);
  }

  addPublicKey(): void {
    this.apiService.addPublicKey(new AddPublicKeyRequest(this.npub))
        .then(() => {
          this.npub = "";
          this.cancelEditingPublicKeysWithoutReloading();
          this.reloadPublicKeys();
        })
        .catch(() => {
          this.store.commit(Mutation.PushNotificationError, "Error adding the public key.");
        });
  }

  private cancelEditingPublicKeysWithoutReloading(): void {
    this.publicKeysToRemove = [];
    this.editingPublicKeys = false;
    this.reloadPublicKeys();
  }

  private reloadPublicKeys(): void {
    this.publicKeys = null;
    this.apiService.publicKeys()
        .then(response => {
          this.publicKeys = response.data;
        })
        .catch(error => {
          if (error.response && error.response.status === 401) {
            return;
          }
          this.store.commit(Mutation.PushNotificationError, "Error loading the public keys.");
        });
  }
}
</script>

<style scoped lang="scss">
.step {
  font-size: 28px;
  margin-top: 2em;
  display: flex;

  .text {
    font-weight: 700;
  }

  .actions {
    padding: 0;
    margin: 0 0 0 1em;
    display: flex;
    color: #F08508;
    list-style-type: none;

    li {
      padding: 0;
      margin: 0;

      a {
        text-decoration: none;

        &:hover {
          cursor: pointer;
          color: #fff;
        }
      }
    }
  }
}

.link-npub-form, .public-keys-wrapper {
  margin: 28px 0;
}

button {
  margin-left: 1em;
}

.current-user {
  width: 550px;
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

    display: flex;
    align-items: center;
    flex-flow: row nowrap;

    &:first-child {
      margin-top: 0;
    }

    &:last-child {
      margin-bottom: 0;
    }

    .npub {
      width: 300px;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  .delete-public-key-button {
    display: block;
    cursor: pointer;
    margin-right: .5em;

    img {
      display: block;
    }
  }
}

.link-npub-form {
  display: flex;

  .input {
    width: 500px;
  }
}

@media screen and (max-width: 1200px) {
  .current-user {
    width: 100%;
  }

  .log-in-with-twitter-button {
    display: block;
  }

  .link-npub-form {
    display: flex;
    flex-flow: column nowrap;

    .input {
      width: auto;
    }

    .button {
      width: auto;
      margin: 1em 0;
    }
  }
}
</style>
