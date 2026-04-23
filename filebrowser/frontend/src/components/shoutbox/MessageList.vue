<template>
  <div ref="list" class="shoutbox-list" @scroll="onScroll">
    <div v-for="m in messages" :key="m.id" class="shoutbox-msg">
      <div class="shoutbox-msg-meta">
        <span class="shoutbox-msg-author">{{ m.author }}</span>
        <span class="shoutbox-msg-time">{{ formatTime(m.createdAt) }}</span>
      </div>
      <div class="shoutbox-msg-body">
        <template v-for="(tok, i) in parse(m.body)">
          <line-link-chip
            v-if="tok.type === 'link'"
            :key="i"
            :path="tok.path"
            :line="tok.line"
            :range-end="tok.rangeEnd"
            :snippet="tok.snippet"
          />
          <span v-else :key="i">{{ tok.value }}</span>
        </template>
      </div>
    </div>
    <div v-if="messages.length === 0" class="shoutbox-empty">
      No messages yet. Say something!
    </div>
  </div>
</template>

<script>
import { mapState } from "vuex";
import { parseBody } from "./parseBody";
import LineLinkChip from "./LineLinkChip.vue";

export default {
  name: "MessageList",
  components: { LineLinkChip },
  data: () => ({ stickToBottom: true }),
  computed: { ...mapState("shoutbox", ["messages"]) },
  watch: {
    messages() {
      this.$nextTick(() => {
        if (this.stickToBottom) this.scrollToBottom();
      });
    },
  },
  mounted() {
    this.scrollToBottom();
  },
  methods: {
    parse: parseBody,
    formatTime(ts) {
      const d = new Date(ts * 1000);
      return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    },
    scrollToBottom() {
      const el = this.$refs.list;
      if (el) el.scrollTop = el.scrollHeight;
    },
    onScroll() {
      const el = this.$refs.list;
      if (!el) return;
      this.stickToBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 30;
    },
  },
};
</script>

<style scoped>
.shoutbox-list {
  flex: 1;
  overflow-y: auto;
  padding: 2.25em 0.75em 0.5em 0.75em;
  background: #36393f;
  color: #dcddde;
}
.shoutbox-list::-webkit-scrollbar { width: 8px; }
.shoutbox-list::-webkit-scrollbar-track { background: #2e3035; }
.shoutbox-list::-webkit-scrollbar-thumb { background: #202225; border-radius: 4px; }
.shoutbox-msg { margin-bottom: 0.6em; }
.shoutbox-msg-meta {
  font-size: 0.75em;
  color: #8e9297;
  display: flex;
  gap: 0.5em;
  margin-bottom: 2px;
}
.shoutbox-msg-author { font-weight: 600; color: #fff; }
.shoutbox-msg-body {
  font-size: 0.92em;
  line-height: 1.4;
  word-wrap: break-word;
  color: #dcddde;
}
.shoutbox-empty {
  color: #8e9297;
  text-align: center;
  margin-top: 2em;
  font-size: 0.85em;
  font-style: italic;
}
</style>
