<template>
  <div>
    <div v-if="progress" class="progress">
      <div v-bind:style="{ width: this.progress + '%' }"></div>
    </div>
    <sidebar></sidebar>
    <main>
      <router-view></router-view>
      <shell v-if="isExecEnabled && isLogged && user.perm.execute" />
    </main>
    <prompts></prompts>
    <upload-files></upload-files>
    <shoutbox v-if="isLogged" />
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import Sidebar from "@/components/Sidebar.vue";
import Prompts from "@/components/prompts/Prompts.vue";
import Shell from "@/components/Shell.vue";
import UploadFiles from "../components/prompts/UploadFiles.vue";
import Shoutbox from "@/components/shoutbox/ShoutBox.vue";
import { enableExec } from "@/utils/constants";

export default {
  name: "layout",
  components: {
    Sidebar,
    Prompts,
    Shell,
    UploadFiles,
    Shoutbox,
  },
  computed: {
    ...mapGetters(["isLogged", "progress", "currentPrompt"]),
    ...mapState(["user"]),
    isExecEnabled: () => enableExec,
  },
  watch: {
    $route: function () {
      this.$store.commit("resetSelected");
      this.$store.commit("multiple", false);
      if (this.currentPrompt?.prompt !== "success")
        this.$store.commit("closeHovers");
    },
    isLogged(value) {
      if (value) this.$store.dispatch("shoutbox/connect");
      else this.$store.dispatch("shoutbox/disconnect");
    },
  },
  mounted() {
    if (this.isLogged) this.$store.dispatch("shoutbox/connect");
    window.addEventListener("message", this.onMessage);
  },
  beforeDestroy() {
    this.$store.dispatch("shoutbox/disconnect");
    window.removeEventListener("message", this.onMessage);
  },
  methods: {
    onMessage(ev) {
      const data = ev.data;
      if (!data || typeof data !== "object") return;
      if (data.type === "shoutbox.attachLine" && data.path && data.line) {
        this.$store.dispatch("shoutbox/attachLine", {
          path: String(data.path),
          line: parseInt(data.line, 10),
          snippet: data.snippet ? String(data.snippet) : "",
        });
      }
    },
  },
};
</script>
