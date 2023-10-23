<template>
  <div class="current-user">
    <img class="image" :src="user?.twitterProfileImageURL">
    <div class="name">{{ user.twitterName }}</div>
    <div class="username">@{{ user.twitterUsername }}</div>
    <a class="logout-button">
      <img src="../assets/logout_on_dark.svg">
    </a>
  </div>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {User} from "@/dto/User";

@Options({
  props: {
    user: User
  },
})
export default class CurrentUser extends Vue {
  user!: User
}
</script>

<style scoped lang="scss">
.current-user {
  margin: 1.5em 0;
  border-radius: 10px;
  width: 500px;
  display: grid;
  grid-template-columns: auto 1fr auto;
  grid-template-rows: auto;
  grid-template-areas:
    "image name logout-button"
    "image username logout-button";
  align-items: center;

  border: 3px solid #9379BF;
  gap: 20px;
  padding: 20px;

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
  }
}
</style>
