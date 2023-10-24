<template>
  <a class="logout-button" @click="logout">
    <div class="label">
      Logout
    </div>
    <img class="icon" src="../assets/logout_on_light.svg"/>
  </a>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {APIService} from "@/services/APIService";
import {useStore} from "vuex";

@Options({})
export default class LogOutButton extends Vue {
  private readonly apiService = new APIService(useStore());

  logout(): void {
    this.apiService.logoutCurrentUser()
  }

}
</script>

<style scoped lang="scss">
a {
  padding: 15px;
  border-radius: 10px;
  background-color: #fff;
  color: #19072C;
  font-size: 24px;
  display: flex;
  align-items: center;
  font-weight: 700;
  cursor: pointer;

  .icon {
    margin-left: 10px;
  }
}
</style>
