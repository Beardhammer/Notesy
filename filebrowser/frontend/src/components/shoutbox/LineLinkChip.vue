<template>
  <a class="shoutbox-chip" :href="href" @click.prevent="go" :title="snippet || display">
    <i class="material-icons">description</i>
    <span class="shoutbox-chip-label">{{ display }}</span>
    <span v-if="snippet" class="shoutbox-chip-snippet">"{{ snippet }}"</span>
  </a>
</template>

<script>
import url from "@/utils/url";

export default {
  name: "LineLinkChip",
  props: {
    path: { type: String, required: true },
    line: { type: Number, required: true },
    rangeEnd: { type: Number, default: null },
    snippet: { type: String, default: "" },
  },
  computed: {
    display() {
      const base = this.path.split("/").pop();
      return this.rangeEnd ? `${base}:${this.line}-${this.rangeEnd}` : `${base}:${this.line}`;
    },
    href() {
      const p = this.path.replace(/^\/+/, "");
      const parts = [`line=${this.line}`];
      if (this.snippet) parts.push(`text=${encodeURIComponent(this.snippet)}`);
      return `/files/${url.encodePath(p)}?${parts.join("&")}`;
    },
  },
  methods: {
    go() {
      this.$router.push(this.href);
    },
  },
};
</script>

<style scoped>
.shoutbox-chip {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 1px 8px; border-radius: 12px;
  background: #1e3a5f; color: #8ec8ff; text-decoration: none;
  font-size: 0.9em; cursor: pointer;
  max-width: 100%;
}
.shoutbox-chip i { font-size: 14px; }
.shoutbox-chip:hover { background: #2a4f7d; color: #b3d8ff; }
.shoutbox-chip-label { font-weight: 500; }
.shoutbox-chip-snippet {
  color: #b9bbbe;
  font-style: italic;
  font-size: 0.85em;
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
