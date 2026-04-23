<template>
  <div class="shoutbox-root">
    <button
      v-if="!open"
      class="shoutbox-fab"
      @click="setOpen(true)"
      :aria-label="'Open shoutbox'"
    >
      <i class="material-icons">chat</i>
      <span v-if="unread > 0" class="shoutbox-fab-badge">{{ unread > 99 ? "99+" : unread }}</span>
      <span v-if="!connected" class="shoutbox-fab-dot" title="disconnected"></span>
    </button>

    <div
      v-if="open"
      class="shoutbox-panel"
      :style="panelStyle"
    >
      <button
        class="shoutbox-ctrl-btn shoutbox-drag-handle"
        @mousedown="startDrag"
        :title="'Drag to move'"
        :aria-label="'Drag to move'"
      >
        <i class="material-icons">drag_indicator</i>
      </button>
      <div class="shoutbox-window-controls">
        <button class="shoutbox-ctrl-btn" @click="toggleMute" :aria-label="muted ? 'Unmute' : 'Mute'" :title="muted ? 'Unmute' : 'Mute'">
          <i class="material-icons">{{ muted ? "notifications_off" : "notifications" }}</i>
        </button>
        <button class="shoutbox-ctrl-btn" @click="minimize" :aria-label="'Minimize'" :title="'Minimize'">
          <i class="material-icons">remove</i>
        </button>
      </div>
      <message-list />
      <compose-box />
    </div>

    <toast-host />
  </div>
</template>

<script>
import { mapState, mapMutations } from "vuex";
import MessageList from "./MessageList.vue";
import ComposeBox from "./ComposeBox.vue";
import ToastHost from "./ToastHost.vue";

const PANEL_W = 360;
const PANEL_H = 500;
const DRAG_THRESHOLD_PX = 3;

export default {
  name: "ShoutBox",
  components: { MessageList, ComposeBox, ToastHost },
  data: () => ({
    dragging: false,
    dragStartX: 0,
    dragStartY: 0,
    dragStartLeft: 0,
    dragStartTop: 0,
    dragMoved: false,
  }),
  computed: {
    ...mapState("shoutbox", ["open", "unread", "connected", "muted", "position"]),
    panelStyle() {
      if (this.position && typeof this.position.x === "number") {
        return {
          left: this.position.x + "px",
          top: this.position.y + "px",
          right: "auto",
          bottom: "auto",
        };
      }
      return {};
    },
  },
  methods: {
    ...mapMutations("shoutbox", ["setOpen", "toggleMute", "setPosition"]),
    minimize() {
      this.setOpen(false);
    },
    startDrag(e) {
      const panel = this.$el.querySelector(".shoutbox-panel");
      if (!panel) return;
      const rect = panel.getBoundingClientRect();
      e.preventDefault();
      this.dragStartX = e.clientX;
      this.dragStartY = e.clientY;
      this.dragStartLeft = rect.left;
      this.dragStartTop = rect.top;
      this.dragging = true;
      this.dragMoved = false;
      document.addEventListener("mousemove", this.onDrag);
      document.addEventListener("mouseup", this.endDrag);
    },
    onDrag(e) {
      if (!this.dragging) return;
      const dx = e.clientX - this.dragStartX;
      const dy = e.clientY - this.dragStartY;
      if (!this.dragMoved) {
        if (Math.abs(dx) < DRAG_THRESHOLD_PX && Math.abs(dy) < DRAG_THRESHOLD_PX) return;
        this.dragMoved = true;
      }
      const maxX = window.innerWidth - PANEL_W;
      const maxY = window.innerHeight - PANEL_H;
      const x = Math.max(0, Math.min(maxX, this.dragStartLeft + dx));
      const y = Math.max(0, Math.min(maxY, this.dragStartTop + dy));
      this.setPosition({ x, y });
    },
    endDrag() {
      this.dragging = false;
      this.dragMoved = false;
      document.removeEventListener("mousemove", this.onDrag);
      document.removeEventListener("mouseup", this.endDrag);
    },
  },
};
</script>

<style scoped>
.shoutbox-root {
  position: fixed;
  right: 1.5em;
  bottom: 1.5em;
  z-index: 10001; /* Above #editor-container which is z-index: 9999 */
  font-family: inherit;
}
.shoutbox-fab {
  width: 56px; height: 56px; border-radius: 50%;
  background: #2196f3; color: white; border: none; cursor: pointer;
  box-shadow: 0 2px 8px rgba(0,0,0,0.35);
  display: flex; align-items: center; justify-content: center;
  position: relative;
}
.shoutbox-fab:hover { background: #1976d2; }
.shoutbox-fab-badge {
  position: absolute; top: -4px; right: -4px;
  background: #f44336; color: white; border-radius: 10px;
  padding: 2px 6px; font-size: 11px; font-weight: bold;
}
.shoutbox-fab-dot {
  position: absolute; bottom: 4px; right: 4px;
  width: 8px; height: 8px; border-radius: 50%; background: #ff9800;
}
.shoutbox-panel {
  position: fixed;
  right: 1.5em;
  bottom: 5.5em;
  width: 360px; height: 500px;
  background: #2f3136;
  color: #dcddde;
  border-radius: 8px;
  box-shadow: 0 6px 24px rgba(0,0,0,0.5);
  display: flex; flex-direction: column;
  overflow: hidden;
}
@media (max-width: 600px) {
  .shoutbox-panel {
    position: fixed !important;
    inset: 0 !important;
    width: 100vw !important; height: 100vh !important;
    border-radius: 0;
  }
}

.shoutbox-drag-handle {
  position: absolute;
  top: 6px;
  left: 6px;
  z-index: 2;
  cursor: move;
}
.shoutbox-drag-handle:hover {
  background: #202225;
  color: #fff;
}

.shoutbox-window-controls {
  position: absolute;
  top: 6px;
  right: 6px;
  display: flex;
  gap: 2px;
  z-index: 2;
}
.shoutbox-ctrl-btn {
  background: rgba(32,34,37,0.7);
  color: #dcddde;
  border: none;
  cursor: pointer;
  border-radius: 50%;
  width: 24px; height: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transition: background 0.15s, color 0.15s;
}
.shoutbox-ctrl-btn:hover {
  background: #202225;
  color: #fff;
}
.shoutbox-ctrl-btn .material-icons { font-size: 16px; }
</style>
