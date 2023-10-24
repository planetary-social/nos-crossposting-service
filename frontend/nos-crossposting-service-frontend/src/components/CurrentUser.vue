<template>
  <div class="current-user">
    <div class="user-info">
      <img class="image" :src="user?.twitterProfileImageURL">
      <div class="name">{{ user.twitterName }}</div>
      <div class="username">@{{ user.twitterUsername }}</div>
      <a class="logout-button" @click="logout">
        <img src="../assets/logout_on_dark.svg">
      </a>
    </div>
    <Checkmark></Checkmark>
  </div>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {User} from "@/dto/User";
import Checkmark from "@/components/Checkmark.vue";
import {APIService} from "@/services/APIService";
import {useStore} from "vuex";
import {Mutation} from "@/store";

@Options({
  components: {Checkmark},
  props: {
    user: User
  },
})
export default class CurrentUser extends Vue {
  user!: User

  private readonly apiService = new APIService(useStore())
  private readonly store = useStore();

  logout(): void {
    this.apiService.logoutCurrentUser()
        .catch(() => {
          this.store.commit(Mutation.PushNotificationError, "Error logging out the user.");
        });
  }
}
</script>

<style scoped lang="scss">
.current-user {
  display: grid;
  grid-template-columns: auto 50px;
  grid-template-rows: auto;
  grid-template-areas:
    "user-info checkmark";
  align-items: center;
  margin: 1.5em 0;

  .user-info {
    grid-area: user-info;
    border-radius: 10px;
    display: grid;
    grid-template-columns: auto 1fr auto;
    grid-template-rows: auto;
    grid-template-areas:
    "image name logout-button"
    "image username logout-button";
    align-items: center;

    border: 3px solid #9379BF;
    gap: 5px 20px;
    padding: 20px;
    background-color: #2A1B45;

    .image {
      grid-area: image;
      border-radius: 75px;
      height: 75px;
    }

    .username, .name {
      font-size: 28px;
      font-style: normal;
    }

    .username {
      grid-area: username;
      font-weight: 500;
      color: #9379BF;
    }

    .name {
      grid-area: name;
      font-weight: 700;
      color: #fff;
    }

    .logout-button {
      grid-area: logout-button;
      cursor: pointer;
    }
  }

  .checkmark {
    grid-area: checkmark;
    padding: 0 15px;
  }
}
</style>
