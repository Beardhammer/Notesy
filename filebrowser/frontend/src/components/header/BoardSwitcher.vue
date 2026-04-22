<template>
  <div class="board-switcher" v-if="boards.length > 0" ref="switcher">
    <button class="board-current" @click="toggleDropdown">
      <i class="material-icons">dashboard</i>
      <span class="board-name">{{ currentBoardName }}</span>
      <i class="material-icons arrow">arrow_drop_down</i>
    </button>

    <button class="board-add" @click="showCreate" title="New Board">
      <i class="material-icons">add</i>
    </button>

    <div class="board-dropdown" v-if="open">
      <div
        class="board-option"
        :class="{ active: currentBoardId === 'all' }"
        @click="selectBoard('all')"
      >
        <i class="material-icons">select_all</i>
        <span>All Boards</span>
      </div>
      <div class="board-divider"></div>
      <div
        class="board-option"
        v-for="b in boards"
        :key="b.id"
        :class="{ active: currentBoardId === b.id }"
      >
        <div class="board-option-main" @click="selectBoard(b.id)">
          <i class="material-icons">view_kanban</i>
          <span>{{ b.name }}</span>
        </div>
        <div class="board-option-actions">
          <button @click.stop="showRename(b)" title="Rename">
            <i class="material-icons">edit</i>
          </button>
          <button @click.stop="confirmDelete(b)" title="Delete" v-if="boards.length > 1">
            <i class="material-icons">delete</i>
          </button>
        </div>
      </div>
    </div>

    <!-- Create/Rename Modal -->
    <div class="modal-overlay" v-if="modal" @click.self="closeModal">
      <div class="modal-card">
        <h3>{{ modalMode === 'create' ? 'New Board' : 'Rename Board' }}</h3>
        <label>Board Name</label>
        <input
          v-model="modalName"
          type="text"
          placeholder="Board name"
          @keyup.enter="saveModal"
          ref="modalInput"
        />
        <div class="modal-actions">
          <button class="button button--flat" @click="closeModal">Cancel</button>
          <button class="button button--flat button--blue" @click="saveModal">Save</button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div class="modal-overlay" v-if="deleteModal" @click.self="deleteModal = false">
      <div class="modal-card">
        <h3>Delete Board</h3>
        <p>Delete "{{ deleteTarget && deleteTarget.name }}" and all its tasks and events? This cannot be undone.</p>
        <div class="modal-actions">
          <button class="button button--flat" @click="deleteModal = false">Cancel</button>
          <button class="button button--flat button--red" @click="doDelete">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import * as kanbanApi from "@/api/kanban";

export default {
  name: "board-switcher",
  data() {
    return {
      open: false,
      modal: false,
      modalMode: "create",
      modalName: "",
      renameTarget: null,
      deleteModal: false,
      deleteTarget: null,
    };
  },
  computed: {
    boards() {
      return this.$store.state.boards;
    },
    currentBoardId() {
      return this.$store.state.currentBoardId;
    },
    currentBoardName() {
      if (this.currentBoardId === "all") return "All Boards";
      const b = this.boards.find((b) => b.id === this.currentBoardId);
      return b ? b.name : "Select Board";
    },
  },
  async created() {
    await this.loadBoards();
    document.addEventListener("click", this.onClickOutside);
  },
  beforeDestroy() {
    document.removeEventListener("click", this.onClickOutside);
  },
  methods: {
    onClickOutside(e) {
      if (this.$refs.switcher && !this.$refs.switcher.contains(e.target)) {
        this.open = false;
      }
    },
    async loadBoards() {
      try {
        const boards = await kanbanApi.getBoards();
        this.$store.commit("setBoards", boards || []);
        // If no current board selected, pick the first one
        if (
          !this.currentBoardId ||
          (!this.boards.find((b) => b.id === this.currentBoardId) &&
            this.currentBoardId !== "all")
        ) {
          if (this.boards.length > 0) {
            this.$store.commit("setCurrentBoardId", this.boards[0].id);
          }
        }
      } catch {
        this.$store.commit("setBoards", []);
      }
    },
    toggleDropdown() {
      this.open = !this.open;
    },
    selectBoard(id) {
      this.$store.commit("setCurrentBoardId", id);
      this.open = false;
    },
    showCreate() {
      this.modalMode = "create";
      this.modalName = "";
      this.modal = true;
      this.open = false;
      this.$nextTick(() => {
        if (this.$refs.modalInput) this.$refs.modalInput.focus();
      });
    },
    showRename(b) {
      this.modalMode = "rename";
      this.modalName = b.name;
      this.renameTarget = b;
      this.modal = true;
      this.open = false;
      this.$nextTick(() => {
        if (this.$refs.modalInput) this.$refs.modalInput.focus();
      });
    },
    closeModal() {
      this.modal = false;
      this.modalName = "";
      this.renameTarget = null;
    },
    async saveModal() {
      if (!this.modalName.trim()) return;
      try {
        if (this.modalMode === "create") {
          const b = await kanbanApi.createBoard(this.modalName.trim());
          await this.loadBoards();
          this.$store.commit("setCurrentBoardId", b.id);
        } else {
          await kanbanApi.updateBoard(
            this.renameTarget.id,
            this.modalName.trim()
          );
          await this.loadBoards();
        }
        this.closeModal();
      } catch (e) {
        this.$showError(e);
      }
    },
    confirmDelete(b) {
      this.deleteTarget = b;
      this.deleteModal = true;
      this.open = false;
    },
    async doDelete() {
      try {
        await kanbanApi.deleteBoard(this.deleteTarget.id);
        const wasSelected = this.currentBoardId === this.deleteTarget.id;
        this.deleteModal = false;
        this.deleteTarget = null;
        await this.loadBoards();
        if (wasSelected && this.boards.length > 0) {
          this.$store.commit("setCurrentBoardId", this.boards[0].id);
        }
      } catch (e) {
        this.$showError(e);
      }
    },
  },
};
</script>

<style scoped>
.board-switcher {
  display: flex;
  align-items: center;
  position: relative;
  gap: 0.25em;
}

.board-current {
  display: flex;
  align-items: center;
  gap: 0.35em;
  background: transparent;
  border: 1px solid var(--divider, #ccc);
  border-radius: 6px;
  padding: 0.3em 0.5em;
  cursor: pointer;
  font-size: 0.9em;
  color: var(--textPrimary, #333);
  transition: background 0.15s;
}

.board-current:hover {
  background: var(--surfaceSecondary, #f5f5f5);
}

.board-current .material-icons {
  font-size: 18px;
}

.board-current .arrow {
  font-size: 20px;
  margin-left: -0.2em;
}

.board-name {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.board-add {
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px dashed var(--divider, #ccc);
  border-radius: 6px;
  width: 30px;
  height: 30px;
  cursor: pointer;
  color: var(--textSecondary, #757575);
  transition: background 0.15s, color 0.15s;
}

.board-add:hover {
  background: var(--surfaceSecondary, #f5f5f5);
  color: var(--textPrimary, #333);
}

.board-add .material-icons {
  font-size: 18px;
}

.board-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  background: var(--surfacePrimary, #fff);
  border: 1px solid var(--divider, #e0e0e0);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  min-width: 220px;
  z-index: 1000;
  padding: 0.35em 0;
}

.board-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5em 0.75em;
  cursor: pointer;
  transition: background 0.1s;
  font-size: 0.88em;
}

.board-option:hover {
  background: var(--surfaceSecondary, #f5f5f5);
}

.board-option.active {
  background: rgba(30, 136, 229, 0.08);
  color: #1e88e5;
}

.board-option-main {
  display: flex;
  align-items: center;
  gap: 0.5em;
  flex: 1;
  min-width: 0;
}

.board-option-main span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.board-option .material-icons {
  font-size: 18px;
}

.board-option-actions {
  display: flex;
  gap: 0.15em;
  opacity: 0;
  transition: opacity 0.15s;
}

.board-option:hover .board-option-actions {
  opacity: 1;
}

.board-option-actions button {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
  color: var(--textSecondary, #757575);
  display: flex;
  align-items: center;
}

.board-option-actions button:hover {
  background: var(--divider, #e0e0e0);
  color: var(--textPrimary, #333);
}

.board-option-actions button .material-icons {
  font-size: 16px;
}

.board-divider {
  height: 1px;
  background: var(--divider, #e0e0e0);
  margin: 0.25em 0;
}

/* Modals */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-card {
  background: var(--surfacePrimary, #fff);
  border-radius: 8px;
  padding: 1.5em;
  width: 360px;
  max-width: 90vw;
}

.modal-card h3 {
  margin: 0 0 1em 0;
}

.modal-card p {
  margin: 0 0 1em 0;
  font-size: 0.9em;
  color: var(--textSecondary, #555);
}

.modal-card label {
  display: block;
  font-size: 0.85em;
  font-weight: 500;
  margin: 0.75em 0 0.25em 0;
}

.modal-card input {
  width: 100%;
  padding: 0.5em;
  border: 1px solid var(--divider, #ccc);
  border-radius: 4px;
  font-size: 0.9em;
  box-sizing: border-box;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5em;
  margin-top: 1.5em;
}

.button--red {
  color: #e53935 !important;
}

.button--blue {
  color: #1e88e5 !important;
}
</style>
