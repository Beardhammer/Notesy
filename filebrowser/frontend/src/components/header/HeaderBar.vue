<template>
  <header>
    <img v-if="showLogo !== undefined" :src="logoURL" />
    <action
      v-if="showMenu !== undefined"
      class="menu-button"
      icon="menu"
      :label="$t('buttons.toggleSidebar')"
      @action="openSidebar()"
    />

    <slot />

    <div class="header-center" v-if="showBoardSwitcher !== undefined">
      <board-switcher />
    </div>

    <div id="dropdown" :class="{ active: this.currentPromptName === 'more' }">
      <slot name="actions" />
    </div>

    <action
      v-if="this.$slots.actions"
      id="more"
      icon="more_vert"
      :label="$t('buttons.more')"
      @action="$store.commit('showHover', 'more')"
    />

    <div
      class="overlay"
      v-show="this.currentPromptName == 'more'"
      @click="$store.commit('closeHovers')"
    />

    <my-tasks></my-tasks>

    <div icon="badge" class="username">{{ username }}</div>

  </header>
</template>

<script>
import { logoURL } from "@/utils/constants";

import Action from "@/components/header/Action.vue";
import BoardSwitcher from "@/components/header/BoardSwitcher.vue";
import MyTasks from "@/components/MyTasks.vue";
import { mapGetters } from "vuex";

export default {
  name: "header-bar",
  props: ["showLogo", "showMenu", "showBoardSwitcher"],
  components: {
    Action,
    BoardSwitcher,
    MyTasks,
  },
  data: function () {
    return {
      logoURL,
      username: this.$store.state.user.username
    };
  },
  methods: {
    openSidebar() {
      this.$store.commit("showHover", "sidebar");
    },
  },
  computed: {
    ...mapGetters(["currentPromptName"])
  },
};
</script>

<style scoped>
.header-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1;
}
</style>
