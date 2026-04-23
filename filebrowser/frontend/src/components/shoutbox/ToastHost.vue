<template>
  <div class="toast-host">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="toast"
      @click="openShoutbox"
    >
      <strong>{{ t.author }}</strong>
      <span class="toast-body">{{ preview(t.body) }}</span>
    </div>
  </div>
</template>

<script>
import { mapState, mapMutations } from "vuex";

export default {
  name: "ToastHost",
  computed: {
    ...mapState("shoutbox", { toasts: "pendingToasts" }),
  },
  watch: {
    toasts: {
      handler(list) {
        list.forEach((t) => {
          if (!t._timer) {
            t._timer = setTimeout(() => this.dequeueToast(t.id), 4000);
          }
        });
      },
      immediate: true,
      deep: true,
    },
  },
  methods: {
    ...mapMutations("shoutbox", ["dequeueToast", "setOpen"]),
    preview(body) {
      const text = body.replace(/\[\[([^\]]+?):\d+(?:-\d+)?\]\]/g, "[link]");
      return text.length > 80 ? text.slice(0, 77) + "…" : text;
    },
    openShoutbox() {
      this.setOpen(true);
    },
  },
};
</script>

<style scoped>
.toast-host {
  position: fixed;
  left: 1em;
  bottom: 1em;
  z-index: 1050;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.toast {
  background: #2f3136;
  color: #dcddde;
  padding: 10px 14px;
  border-radius: 6px;
  box-shadow: 0 4px 14px rgba(0,0,0,0.5);
  max-width: 320px;
  cursor: pointer;
  font-size: 0.9em;
  border-left: 3px solid #2196f3;
}
.toast:hover { background: #34373c; }
.toast strong { color: #fff; margin-right: 6px; }
.toast-body { color: #dcddde; }
</style>
