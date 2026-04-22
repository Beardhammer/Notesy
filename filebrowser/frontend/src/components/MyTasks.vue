<template>
  <div class="my-tasks" v-if="isLogged && myTasks.length">
    <span class="my-tasks-label">Your tasks:</span>
    <span
      class="my-tasks-chip"
      v-for="t in myTasks"
      :key="t.id"
      :class="'chip--' + t.column"
      @click="goToTask(t)"
    >{{ t.title }}</span>
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import * as kanbanApi from "@/api/kanban";

export default {
  name: "my-tasks",
  data() {
    return {
      tasks: [],
    };
  },
  computed: {
    ...mapState(["user"]),
    ...mapGetters(["isLogged"]),
    myTasks() {
      const me = this.user && this.user.username;
      if (!me) return [];
      return this.tasks.filter((t) => {
        const assignees = t.assignedTo || "";
        if (Array.isArray(assignees)) return assignees.includes(me);
        return assignees
          .split(",")
          .map((s) => s.trim())
          .includes(me);
      });
    },
  },
  watch: {
    isLogged: {
      immediate: true,
      handler(val) {
        if (val) this.loadTasks();
      },
    },
    $route() {
      if (this.isLogged) this.loadTasks();
    },
  },
  methods: {
    async loadTasks() {
      try {
        this.tasks = (await kanbanApi.getTasks("all")) || [];
      } catch {
        this.tasks = [];
      }
    },
    goToTask(task) {
      if (this.$route.path !== "/kanban") {
        this.$router.push({ path: "/kanban" }, () => {});
      }
    },
  },
};
</script>

<style scoped>
.my-tasks {
  display: flex;
  align-items: center;
  gap: 0.35em;
  margin-left: auto;
  margin-right: 0.5em;
  overflow-x: auto;
  max-width: 50vw;
}

.my-tasks-label {
  font-size: 0.75em;
  font-weight: 600;
  white-space: nowrap;
  color: var(--textSecondary, #999);
}

.my-tasks-chip {
  font-size: 0.7em;
  padding: 0.15em 0.5em;
  border-radius: 10px;
  white-space: nowrap;
  cursor: pointer;
  color: #fff;
  font-weight: 500;
  transition: opacity 0.15s;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.my-tasks-chip:hover {
  opacity: 0.8;
}

.chip--todo {
  background: #42A5F5;
}

.chip--inprogress {
  background: #FFA726;
}

.chip--done {
  background: #66BB6A;
}

.chip--blocked {
  background: #EF5350;
}
</style>
