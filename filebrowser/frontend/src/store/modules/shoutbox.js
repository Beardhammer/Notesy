const MAX_MESSAGES = 200;

function loadPosition() {
  try {
    const raw = localStorage.getItem("shoutboxPosition");
    if (!raw) return null;
    const p = JSON.parse(raw);
    if (typeof p.x === "number" && typeof p.y === "number") return p;
  } catch (e) {
    // ignore
  }
  return null;
}

const state = () => ({
  messages: [],
  open: false,
  unread: 0,
  muted: localStorage.getItem("shoutboxMuted") === "1",
  lastId: 0,
  connected: false,
  eventSource: null,
  pendingToasts: [], // consumed by ToastHost.vue
  position: loadPosition(), // {x, y} or null = default corner
  pendingInsert: "", // text to splice into compose draft (from codimd attach button)
});

const mutations = {
  setOpen(state, value) {
    state.open = value;
    if (value) state.unread = 0;
  },
  setConnected(state, value) {
    state.connected = value;
  },
  setEventSource(state, es) {
    state.eventSource = es;
  },
  setMessages(state, msgs) {
    state.messages = msgs.slice(-MAX_MESSAGES);
    state.lastId = state.messages.length
      ? state.messages[state.messages.length - 1].id
      : 0;
  },
  appendMessage(state, msg) {
    if (state.messages.some((m) => m.id === msg.id)) return; // dedupe
    state.messages.push(msg);
    if (state.messages.length > MAX_MESSAGES) state.messages.shift();
    state.lastId = msg.id;
  },
  bumpUnread(state) {
    state.unread += 1;
  },
  toggleMute(state) {
    state.muted = !state.muted;
    localStorage.setItem("shoutboxMuted", state.muted ? "1" : "0");
  },
  enqueueToast(state, msg) {
    state.pendingToasts.push(msg);
    if (state.pendingToasts.length > 3) state.pendingToasts.shift();
  },
  dequeueToast(state, id) {
    state.pendingToasts = state.pendingToasts.filter((m) => m.id !== id);
  },
  setPosition(state, pos) {
    state.position = pos;
    if (pos) localStorage.setItem("shoutboxPosition", JSON.stringify(pos));
  },
  setPendingInsert(state, text) {
    state.pendingInsert = text || "";
  },
  clearPendingInsert(state) {
    state.pendingInsert = "";
  },
};

const actions = {
  async connect({ commit, dispatch, state, rootState }) {
    if (state.connected) return;
    try {
      const res = await fetch("/api/shouts", {
        headers: { "X-Auth": rootState.jwt },
      });
      const msgs = await res.json();
      commit("setMessages", msgs || []);
    } catch (e) {
      // ignored — SSE will retry
    }
    dispatch("openStream");
  },
  openStream({ commit, state, rootState }) {
    if (state.eventSource) return;
    // EventSource cannot set custom headers — JWT goes via ?auth=, which http/auth.go accepts.
    const url = `/api/shouts/stream?auth=${encodeURIComponent(rootState.jwt || "")}`;
    const es = new EventSource(url);
    es.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data);
        commit("appendMessage", msg);
        const user = rootState.user;
        if (
          msg.author !== (user && user.username) &&
          !state.open &&
          !state.muted
        ) {
          commit("bumpUnread");
          commit("enqueueToast", msg);
        }
      } catch (e) {
        // skip malformed frame
      }
    };
    es.onopen = () => commit("setConnected", true);
    es.onerror = () => commit("setConnected", false);
    commit("setEventSource", es);
  },
  disconnect({ commit, state }) {
    if (state.eventSource) state.eventSource.close();
    commit("setEventSource", null);
    commit("setConnected", false);
  },
  async send({ rootState }, body) {
    const res = await fetch("/api/shouts", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Auth": rootState.jwt,
      },
      body: JSON.stringify({ body }),
    });
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },
  attachLine({ commit }, { path, line, snippet }) {
    if (!path || !line) return;
    const chip = `[[${path}:${line}]]`;
    // If the user selected text, quote it after the chip so the message reads
    // `[[file.md:42]] "the highlighted text"`.
    const cleanSnippet = (snippet || "")
      .replace(/"/g, "'")
      .replace(/\s+/g, " ")
      .trim();
    const text = cleanSnippet ? `${chip} "${cleanSnippet}"` : chip;
    commit("setPendingInsert", text);
    commit("setOpen", true);
  },
};

export default { namespaced: true, state, mutations, actions };
