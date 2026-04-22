<template>
  <div class="kanban-container">
    <header-bar showMenu showBoardSwitcher>
      <p>Kanban</p>
      <action icon="file_download" label="Export Backup" @action="exportBackup" />
    </header-bar>

    <div class="kanban-board">
      <div
        class="kanban-column"
        :class="'kanban-column--' + col.id"
        v-for="col in columns"
        :key="col.id"
        @dragover.prevent
        @drop="onDrop(col.id, $event)"
      >
        <h3>
          {{ col.label }}
          <span class="count">{{ columnTasks(col.id).length }}</span>
        </h3>
        <div class="kanban-cards">
          <div
            class="kanban-card"
            v-for="task in columnTasks(col.id)"
            :key="task.id"
            draggable="true"
            @dragstart="onDragStart(task, $event)"
            @click="editTask(task)"
          >
            <div class="card-header">
              <span class="card-title">{{ task.title }}</span>
              <div class="card-avatars" v-if="getAssigneeList(task).length">
                <span
                  class="avatar-chip"
                  v-for="name in getAssigneeList(task)"
                  :key="name"
                  :title="name"
                  :style="{ backgroundColor: avatarColor(name) }"
                >{{ initials(name) }}</span>
              </div>
            </div>
            <span class="card-dates" v-if="task.startDate || task.endDate">
              <i class="material-icons" style="font-size: 14px"
                >calendar_today</i
              >
              {{ task.startDate }}{{ task.endDate ? " - " + task.endDate : "" }}
            </span>
            <span class="card-board" v-if="isAllBoards && task.boardName">
              <i class="material-icons" style="font-size: 12px">dashboard</i>
              {{ task.boardName }}
            </span>
          </div>
        </div>
        <button class="add-task-btn" @click="showNewTask(col.id)">
          <i class="material-icons">add</i>
        </button>
      </div>
    </div>

    <!-- Task Modal -->
    <div class="modal-overlay" v-if="modal" @click.self="closeModal">
      <div class="modal-card">
        <h3>{{ editing ? "Edit Task" : "New Task" }}</h3>
        <label>Title</label>
        <input v-model="form.title" type="text" placeholder="Task title" />
        <label>Description</label>
        <textarea
          v-model="form.description"
          placeholder="Description"
          rows="3"
        ></textarea>
        <label>Column</label>
        <select v-model="form.column">
          <option value="todo">To Do</option>
          <option value="inprogress">In Progress</option>
          <option value="done">Done</option>
          <option value="blocked">Blocked</option>
        </select>
        <label>Assign To</label>
        <div class="assignee-list">
          <label
            class="assignee-check"
            v-for="u in assignableUsers"
            :key="u.username"
          >
            <input
              type="checkbox"
              :value="u.username"
              v-model="form.assignedTo"
            />
            {{ u.username }}
          </label>
        </div>
        <label>Start Date</label>
        <input v-model="form.startDate" type="date" />
        <label>End Date/Due Date</label>
        <input v-model="form.endDate" type="date" />
        <div class="modal-actions">
          <button class="button button--flat" @click="closeModal">
            Cancel
          </button>
          <button
            v-if="editing"
            class="button button--flat button--red"
            @click="removeTask"
          >
            Delete
          </button>
          <button class="button button--flat button--blue" @click="saveTask">
            Save
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import * as kanbanApi from "@/api/kanban";
import { users as usersApi } from "@/api";

export default {
  name: "kanban-board",
  components: { HeaderBar, Action },
  data() {
    return {
      tasks: [],
      userList: [],
      modal: false,
      editing: false,
      editingId: null,
      form: {
        title: "",
        description: "",
        column: "todo",
        assignedTo: [],
        startDate: "",
        endDate: "",
      },
      dragTask: null,
      columns: [
        { id: "todo", label: "To Do" },
        { id: "inprogress", label: "In Progress" },
        { id: "done", label: "Done" },
        { id: "blocked", label: "Blocked" },
      ],
    };
  },
  computed: {
    assignableUsers() {
      return this.userList.filter(
        (u) => u.username.toLowerCase() !== "admin"
      );
    },
    boardId() {
      return this.$store.state.currentBoardId;
    },
    isAllBoards() {
      return this.boardId === "all";
    },
  },
  watch: {
    boardId() {
      this.loadTasks();
    },
  },
  async created() {
    await this.loadTasks();
    try {
      this.userList = await usersApi.getAll();
    } catch {
      this.userList = [];
    }
  },
  methods: {
    async loadTasks() {
      try {
        this.tasks = await kanbanApi.getTasks(this.boardId);
        if (!this.tasks) this.tasks = [];
      } catch {
        this.tasks = [];
      }
    },
    columnTasks(colId) {
      return this.tasks
        .filter((t) => t.column === colId)
        .sort((a, b) => a.position - b.position);
    },
    showNewTask(colId) {
      this.editing = false;
      this.editingId = null;
      this.form = {
        title: "",
        description: "",
        column: colId || "todo",
        assignedTo: [],
        startDate: "",
        endDate: "",
      };
      this.modal = true;
    },
    editTask(task) {
      this.editing = true;
      this.editingId = task.id;
      let assignees = task.assignedTo || [];
      if (typeof assignees === "string") {
        assignees = assignees ? assignees.split(",").map((s) => s.trim()) : [];
      }
      this.form = {
        title: task.title,
        description: task.description,
        column: task.column,
        assignedTo: assignees,
        startDate: task.startDate || "",
        endDate: task.endDate || "",
      };
      this.modal = true;
    },
    closeModal() {
      this.modal = false;
    },
    async saveTask() {
      if (!this.form.title.trim()) return;
      const payload = {
        ...this.form,
        assignedTo: Array.isArray(this.form.assignedTo)
          ? this.form.assignedTo.join(", ")
          : this.form.assignedTo,
      };
      try {
        if (this.editing) {
          await kanbanApi.updateTask(this.boardId, this.editingId, payload);
        } else {
          const bid = this.isAllBoards ? this.$store.state.boards[0]?.id : this.boardId;
          await kanbanApi.createTask(bid, payload);
        }
        await this.loadTasks();
        this.closeModal();
      } catch (e) {
        this.$showError(e);
      }
    },
    async removeTask() {
      try {
        const task = this.tasks.find((t) => t.id === this.editingId);
        const bid = task && task.boardId ? task.boardId : this.boardId;
        await kanbanApi.deleteTask(bid, this.editingId);
        await this.loadTasks();
        this.closeModal();
      } catch (e) {
        this.$showError(e);
      }
    },
    getAssigneeList(task) {
      if (!task.assignedTo) return [];
      if (Array.isArray(task.assignedTo)) return task.assignedTo;
      return task.assignedTo.split(",").map((s) => s.trim()).filter(Boolean);
    },
    initials(name) {
      return name.charAt(0).toUpperCase();
    },
    avatarColor(name) {
      const colors = [
        "#42A5F5", "#66BB6A", "#FFA726", "#AB47BC",
        "#26C6DA", "#EF5350", "#8D6E63", "#5C6BC0",
      ];
      let hash = 0;
      for (let i = 0; i < name.length; i++) {
        hash = name.charCodeAt(i) + ((hash << 5) - hash);
      }
      return colors[Math.abs(hash) % colors.length];
    },
    async exportBackup() {
      try {
        const tasks = await kanbanApi.getTasks(this.boardId) || [];
        const events = await kanbanApi.getAllEvents(this.boardId) || [];
        const backup = {
          exportedAt: new Date().toISOString(),
          tasks,
          events,
        };
        const blob = new Blob([JSON.stringify(backup, null, 2)], {
          type: "application/json",
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `filebrowser-backup-${new Date().toISOString().slice(0, 10)}.json`;
        a.click();
        URL.revokeObjectURL(url);
      } catch (e) {
        this.$showError(e);
      }
    },
    onDragStart(task, event) {
      this.dragTask = task;
      event.dataTransfer.effectAllowed = "move";
      event.dataTransfer.setData("text/plain", task.id.toString());
    },
    async onDrop(colId, event) {
      event.preventDefault();
      if (!this.dragTask) return;

      const task = this.dragTask;
      this.dragTask = null;

      if (task.column === colId) return;

      try {
        const bid = task.boardId || this.boardId;
        await kanbanApi.updateTask(bid, task.id, {
          title: task.title,
          description: task.description,
          column: colId,
          position: task.position,
          assignedTo: task.assignedTo,
          startDate: task.startDate,
          endDate: task.endDate,
          createdBy: task.createdBy,
          createdAt: task.createdAt,
          boardId: task.boardId,
        });
        await this.loadTasks();
      } catch (e) {
        this.$showError(e);
      }
    },
  },
};
</script>

<style scoped>
.kanban-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.kanban-board {
  display: flex;
  gap: 1em;
  padding: 1em;
  flex: 1;
  overflow-x: auto;
  min-height: 0;
}

.kanban-column {
  flex: 1;
  min-width: 220px;
  background: var(--surfaceSecondary, #f5f5f5);
  border-radius: 8px;
  padding: 0.75em;
  display: flex;
  flex-direction: column;
}

.kanban-column--todo {
  box-shadow: 0 0 12px 2px rgba(66, 165, 245, 0.35);
  border: 1px solid rgba(66, 165, 245, 0.4);
}

.kanban-column--inprogress {
  box-shadow: 0 0 12px 2px rgba(255, 167, 38, 0.35);
  border: 1px solid rgba(255, 167, 38, 0.4);
}

.kanban-column--done {
  box-shadow: 0 0 12px 2px rgba(102, 187, 106, 0.35);
  border: 1px solid rgba(102, 187, 106, 0.4);
}

.kanban-column--blocked {
  box-shadow: 0 0 12px 2px rgba(239, 83, 80, 0.35);
  border: 1px solid rgba(239, 83, 80, 0.4);
}

.kanban-column h3 {
  margin: 0 0 0.5em 0;
  font-size: 0.95em;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.5em;
}

.kanban-column h3 .count {
  background: var(--surfacePrimary, #e0e0e0);
  border-radius: 50%;
  width: 22px;
  height: 22px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8em;
}

.kanban-cards {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.5em;
}

.kanban-card {
  background: var(--surfacePrimary, #fff);
  border: 1px solid var(--divider, #e0e0e0);
  border-radius: 6px;
  padding: 0.75em;
  cursor: grab;
  display: flex;
  flex-direction: column;
  gap: 0.3em;
  transition: box-shadow 0.15s;
}

.kanban-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.kanban-card:active {
  cursor: grabbing;
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5em;
}

.card-title {
  font-weight: 500;
  font-size: 0.9em;
  flex: 1;
  min-width: 0;
}

.card-avatars {
  display: flex;
  flex-shrink: 0;
  margin-left: auto;
}

.card-avatars .avatar-chip {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  color: #fff;
  font-size: 0.7em;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-left: -6px;
  border: 2px solid var(--surfacePrimary, #fff);
  cursor: default;
}

.card-avatars .avatar-chip:first-child {
  margin-left: 0;
}

.card-dates {
  font-size: 0.78em;
  color: var(--textSecondary, #757575);
  display: flex;
  align-items: center;
  gap: 0.25em;
}

.card-board {
  font-size: 0.72em;
  color: var(--textSecondary, #999);
  display: flex;
  align-items: center;
  gap: 0.2em;
  margin-top: 0.15em;
}

/* Modal */
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
  width: 420px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-card h3 {
  margin: 0 0 1em 0;
}

.modal-card label {
  display: block;
  font-size: 0.85em;
  font-weight: 500;
  margin: 0.75em 0 0.25em 0;
}

.modal-card input,
.modal-card textarea,
.modal-card select {
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

.add-task-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 0.4em;
  margin-top: 0.5em;
  border: 2px dashed var(--divider, #ccc);
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  color: var(--textSecondary, #757575);
  transition: background 0.15s, color 0.15s;
}

.add-task-btn:hover {
  background: var(--surfacePrimary, #e8e8e8);
  color: var(--textPrimary, #333);
}

.assignee-list {
  max-height: 140px;
  overflow-y: auto;
  border: 1px solid var(--divider, #ccc);
  border-radius: 4px;
  padding: 0.35em 0.5em;
}

.assignee-check {
  display: flex !important;
  align-items: center;
  gap: 0.4em;
  font-weight: 400 !important;
  margin: 0.2em 0 !important;
  cursor: pointer;
}

.assignee-check input[type="checkbox"] {
  width: auto;
}

.button--red {
  color: #e53935 !important;
}

.button--blue {
  color: #1e88e5 !important;
}
</style>
