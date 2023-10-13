<template>
  <div class="current-user">
    You are logged in as {{ user.accountID }}.
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
</style>
