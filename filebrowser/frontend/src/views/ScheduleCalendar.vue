<template>
  <div class="schedule-container">
    <header-bar showMenu showBoardSwitcher>
      <action icon="chevron_left" label="Previous Month" @action="prevMonth" />
      <p>{{ monthName }} {{ year }}</p>
      <action icon="chevron_right" label="Next Month" @action="nextMonth" />
      <action icon="add" label="New Event" @action="showNewEvent" />
      <action icon="file_download" label="Export Backup" @action="exportBackup" />
    </header-bar>

    <div class="calendar-grid">
      <div class="day-header" v-for="d in dayNames" :key="d">{{ d }}</div>
      <div
        class="calendar-day"
        v-for="(day, idx) in calendarDays"
        :key="idx"
        :class="{
          'other-month': !day.currentMonth,
          today: day.isToday,
        }"
        @click="onDayClick(day)"
      >
        <span class="day-number">{{ day.date }}</span>
        <div
          class="day-event"
          v-for="item in day.items"
          :key="item.type + '-' + item.id"
          :style="{ backgroundColor: item.color }"
          @click.stop="editItem(item)"
        >
          {{ item.title }}
        </div>
      </div>
    </div>

    <!-- Event Modal -->
    <div class="modal-overlay" v-if="modal" @click.self="closeModal">
      <div class="modal-card">
        <h3>{{ editing ? "Edit Event" : "New Event" }}</h3>
        <label>Title</label>
        <input v-model="form.title" type="text" placeholder="Event title" />
        <label>Date</label>
        <input v-model="form.date" type="date" />
        <label>End Date/Due Date (optional)</label>
        <input v-model="form.endDate" type="date" />
        <label>Color</label>
        <div class="color-picker">
          <span
            v-for="c in colorOptions"
            :key="c"
            class="color-swatch"
            :style="{ backgroundColor: c }"
            :class="{ selected: form.color === c }"
            @click="form.color = c"
          ></span>
        </div>
        <div class="modal-actions">
          <button class="button button--flat" @click="closeModal">
            Cancel
          </button>
          <button
            v-if="editing && editingType === 'event'"
            class="button button--flat button--red"
            @click="removeEvent"
          >
            Delete
          </button>
          <button
            v-if="!editing || editingType === 'event'"
            class="button button--flat button--blue"
            @click="saveEvent"
          >
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

const columnColors = {
  todo: "#42A5F5",
  inprogress: "#FFA726",
  done: "#66BB6A",
  blocked: "#EF5350",
};

export default {
  name: "schedule-calendar",
  components: { HeaderBar, Action },
  data() {
    const now = new Date();
    return {
      year: now.getFullYear(),
      month: now.getMonth(),
      tasks: [],
      events: [],
      modal: false,
      editing: false,
      editingId: null,
      editingType: null,
      form: {
        title: "",
        date: "",
        endDate: "",
        color: "#42A5F5",
      },
      dayNames: ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"],
      colorOptions: [
        "#42A5F5",
        "#66BB6A",
        "#FFA726",
        "#EF5350",
        "#AB47BC",
        "#26C6DA",
        "#8D6E63",
        "#78909C",
      ],
    };
  },
  computed: {
    boardId() {
      return this.$store.state.currentBoardId;
    },
    monthName() {
      return new Date(this.year, this.month).toLocaleString("default", {
        month: "long",
      });
    },
    calendarDays() {
      const days = [];
      const firstDay = new Date(this.year, this.month, 1);
      const lastDay = new Date(this.year, this.month + 1, 0);
      const startWeekday = firstDay.getDay();
      const today = new Date();
      const todayStr = this.formatDate(today);

      // Previous month padding
      const prevLastDay = new Date(this.year, this.month, 0);
      for (let i = startWeekday - 1; i >= 0; i--) {
        const d = prevLastDay.getDate() - i;
        const dateStr = this.formatDate(
          new Date(this.year, this.month - 1, d)
        );
        days.push({
          date: d,
          fullDate: dateStr,
          currentMonth: false,
          isToday: dateStr === todayStr,
          items: this.getItemsForDate(dateStr),
        });
      }

      // Current month
      for (let d = 1; d <= lastDay.getDate(); d++) {
        const dateStr = this.formatDate(new Date(this.year, this.month, d));
        days.push({
          date: d,
          fullDate: dateStr,
          currentMonth: true,
          isToday: dateStr === todayStr,
          items: this.getItemsForDate(dateStr),
        });
      }

      // Next month padding
      const remaining = 7 - (days.length % 7);
      if (remaining < 7) {
        for (let d = 1; d <= remaining; d++) {
          const dateStr = this.formatDate(
            new Date(this.year, this.month + 1, d)
          );
          days.push({
            date: d,
            fullDate: dateStr,
            currentMonth: false,
            isToday: dateStr === todayStr,
            items: this.getItemsForDate(dateStr),
          });
        }
      }

      return days;
    },
  },
  watch: {
    boardId() {
      this.loadData();
    },
  },
  async created() {
    await this.loadData();
  },
  methods: {
    formatDate(d) {
      const y = d.getFullYear();
      const m = String(d.getMonth() + 1).padStart(2, "0");
      const day = String(d.getDate()).padStart(2, "0");
      return `${y}-${m}-${day}`;
    },
    async loadData() {
      try {
        const tasks = await kanbanApi.getTasks(this.boardId);
        this.tasks = tasks || [];
      } catch {
        this.tasks = [];
      }
      try {
        const events = await kanbanApi.getAllEvents(this.boardId);
        this.events = events || [];
      } catch {
        this.events = [];
      }
    },
    getItemsForDate(dateStr) {
      const items = [];

      // Kanban tasks with dates
      for (const t of this.tasks) {
        if (t.startDate === dateStr || (t.startDate && t.endDate && dateStr >= t.startDate && dateStr <= t.endDate) || (!t.startDate && t.endDate === dateStr)) {
          items.push({
            id: t.id,
            title: t.title,
            color: columnColors[t.column] || "#42A5F5",
            type: "task",
          });
        }
      }

      // Events
      for (const e of this.events) {
        if (e.date === dateStr || (e.date && e.endDate && dateStr >= e.date && dateStr <= e.endDate)) {
          items.push({
            id: e.id,
            title: e.title,
            color: e.color || "#42A5F5",
            type: "event",
          });
        }
      }

      return items;
    },
    prevMonth() {
      if (this.month === 0) {
        this.month = 11;
        this.year--;
      } else {
        this.month--;
      }
    },
    nextMonth() {
      if (this.month === 11) {
        this.month = 0;
        this.year++;
      } else {
        this.month++;
      }
    },
    onDayClick(day) {
      this.editing = false;
      this.editingId = null;
      this.editingType = null;
      this.form = {
        title: "",
        date: day.fullDate,
        endDate: "",
        color: "#42A5F5",
      };
      this.modal = true;
    },
    editItem(item) {
      if (item.type === "event") {
        const ev = this.events.find((e) => e.id === item.id);
        if (!ev) return;
        this.editing = true;
        this.editingId = ev.id;
        this.editingType = "event";
        this.form = {
          title: ev.title,
          date: ev.date,
          endDate: ev.endDate || "",
          color: ev.color || "#42A5F5",
        };
        this.modal = true;
      }
      // Kanban tasks are view-only on calendar; edit them on the Kanban board
    },
    closeModal() {
      this.modal = false;
    },
    async saveEvent() {
      if (!this.form.title.trim() || !this.form.date) return;
      try {
        if (this.editing && this.editingType === "event") {
          const ev = this.events.find((e) => e.id === this.editingId);
          const bid = ev && ev.boardId ? ev.boardId : this.boardId;
          await kanbanApi.updateEvent(bid, this.editingId, this.form);
        } else {
          const bid = this.boardId === "all" ? this.$store.state.boards[0]?.id : this.boardId;
          await kanbanApi.createEvent(bid, this.form);
        }
        await this.loadData();
        this.closeModal();
      } catch (e) {
        this.$showError(e);
      }
    },
    async removeEvent() {
      try {
        const ev = this.events.find((e) => e.id === this.editingId);
        const bid = ev && ev.boardId ? ev.boardId : this.boardId;
        await kanbanApi.deleteEvent(bid, this.editingId);
        await this.loadData();
        this.closeModal();
      } catch (e) {
        this.$showError(e);
      }
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
    showNewEvent() {
      this.editing = false;
      this.editingId = null;
      this.editingType = null;
      const today = new Date();
      this.form = {
        title: "",
        date: this.formatDate(today),
        endDate: "",
        color: "#42A5F5",
      };
      this.modal = true;
    },
  },
};
</script>

<style scoped>
.schedule-container {
  width: 100%;
  height: calc(100vh - 4em);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  grid-template-rows: auto;
  grid-auto-rows: 1fr;
  gap: 1px;
  background: var(--divider, #e0e0e0);
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.day-header {
  background: var(--surfaceSecondary, #f5f5f5);
  padding: 0.5em;
  text-align: center;
  font-weight: 600;
  font-size: 0.85em;
}

.calendar-day {
  background: var(--surfacePrimary, #fff);
  padding: 0.5em;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.calendar-day:hover {
  background: var(--surfaceSecondary, #fafafa);
}

.calendar-day.other-month {
  opacity: 0.4;
}

.calendar-day.today .day-number {
  background: #1e88e5;
  color: #fff;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.day-number {
  font-size: 0.82em;
  font-weight: 500;
  margin-bottom: 0.2em;
}

.day-event {
  font-size: 0.72em;
  color: #fff;
  padding: 1px 4px;
  border-radius: 3px;
  margin-bottom: 1px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  cursor: pointer;
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
  width: 380px;
  max-width: 90vw;
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

.modal-card input {
  width: 100%;
  padding: 0.5em;
  border: 1px solid var(--divider, #ccc);
  border-radius: 4px;
  font-size: 0.9em;
  box-sizing: border-box;
}

.color-picker {
  display: flex;
  gap: 0.5em;
  margin-top: 0.25em;
}

.color-swatch {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color 0.15s;
}

.color-swatch.selected {
  border-color: #333;
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
