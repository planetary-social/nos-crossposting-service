<template>
  <ul class="notifications">
    <li class="notification"
        :class="[notification.style]"
        v-for="(notification, index) in notifications" :key="index">
      <div class="text">
        {{ notification.text }}
      </div>
      <div class="button" @click="dismiss(index)">
        X
      </div>
    </li>
  </ul>
</template>

<script lang="ts">
import {Options, Vue} from 'vue-class-component';
import {useStore} from "vuex";
import {Mutation} from "@/store";

@Options({
  components: {}
})
export default class Notifications extends Vue {

  private readonly store = useStore();

  get notifications(): Notifications[] {
    return this.store.state.notifications;
  }

  dismiss(index: number): void {
    this.store.commit(Mutation.DismissNotification, index);
  }
}

</script>

<style scoped lang="scss">
$red: #ef6155;
$green: #2ecc71;

.notifications {
  margin: 0;
  padding: 0;

  .notification {
    list-style-type: none;
    padding: 1em;
    margin: 1em;
    border-radius: 10px;
    color: #fff;
    overflow: hidden;
    display: flex;
    flex-flow: row nowrap;
    align-items: center;

    .text {
      flex: 1;
    }

    .button {
      cursor: pointer;
    }

    &.error {
      background-color: $red;
    }

    &.success {
      background-color: $green;
    }
  }
}
</style>
