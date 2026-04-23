<template>
  <div class="shoutbox-compose">
    <p v-if="error" class="shoutbox-error">{{ error }}</p>
    <div class="shoutbox-compose-row">
      <textarea
        ref="ta"
        v-model="draft"
        rows="2"
        placeholder="Type a message…"
        @keydown.enter.exact.prevent="send"
      />
      <button
        type="button"
        class="shoutbox-send-btn"
        :disabled="!canSend"
        @click="send"
        :aria-label="'Send'"
      >
        <i class="material-icons">send</i>
      </button>
    </div>
  </div>
</template>

<script>
import { mapState, mapMutations } from "vuex";

export default {
  name: "ComposeBox",
  data: () => ({
    draft: "",
    error: "",
  }),
  computed: {
    ...mapState("shoutbox", ["pendingInsert"]),
    canSend() {
      return this.draft.trim().length > 0 && this.draft.length <= 2000;
    },
  },
  watch: {
    pendingInsert(text) {
      if (!text) return;
      const ta = this.$refs.ta;
      const sep = this.draft && !this.draft.endsWith(" ") ? " " : "";
      const start = ta ? ta.selectionStart : this.draft.length;
      const end = ta ? ta.selectionEnd : this.draft.length;
      this.draft = this.draft.slice(0, start) + sep + text + this.draft.slice(end);
      this.clearPendingInsert();
      this.$nextTick(() => {
        if (ta) {
          ta.focus();
          const pos = start + sep.length + text.length;
          ta.setSelectionRange(pos, pos);
        }
      });
    },
  },
  methods: {
    ...mapMutations("shoutbox", ["clearPendingInsert"]),
    async send() {
      if (!this.canSend) return;
      this.error = "";
      const body = this.draft;
      try {
        await this.$store.dispatch("shoutbox/send", body);
        this.draft = "";
      } catch (e) {
        this.error = "couldn't send — try again";
      }
    },
  },
};
</script>

<style scoped>
.shoutbox-compose {
  border-top: 1px solid #18191c;
  padding: 0.5em;
  background: #2f3136;
}
.shoutbox-compose-row { display: flex; gap: 6px; align-items: flex-end; }
.shoutbox-compose textarea {
  flex: 1;
  resize: none;
  border: 1px solid #18191c;
  border-radius: 4px;
  padding: 6px 8px;
  font: inherit;
  background: #40444b;
  color: #dcddde;
  outline: none;
}
.shoutbox-compose textarea:focus { border-color: #2196f3; }
.shoutbox-compose textarea::placeholder { color: #72767d; }
.shoutbox-send-btn {
  background: #2196f3;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 10px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
}
.shoutbox-send-btn:hover:not([disabled]) { background: #1976d2; }
.shoutbox-send-btn[disabled] { opacity: 0.4; cursor: not-allowed; }
.shoutbox-send-btn .material-icons { font-size: 18px; }
.shoutbox-error { color: #f48687; font-size: 0.85em; margin: 0 0 0.25em 0; }
</style>
