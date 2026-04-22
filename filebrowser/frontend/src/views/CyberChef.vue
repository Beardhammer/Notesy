<template>
  <div class="cyberchef-page">
    <header class="cyberchef-header">
      <button class="team-share-btn" @click="goHome">
        <i class="material-icons">home</i>
        <span>Team Share</span>
      </button>
      <span class="cyberchef-title">CyberChef</span>
    </header>
    <iframe
      class="cyberchef-frame"
      src="/cyberchef/"
      title="CyberChef"
      referrerpolicy="no-referrer"
    />
  </div>
</template>

<script>
export default {
  name: "cyberchef",
  methods: {
    goHome() {
      this.$router.push({ path: "/files/" });
    },
  },
};
</script>

<style scoped>
.cyberchef-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  position: fixed;
  top: 0;
  left: 0;
  z-index: 999;
  background: #fff;
}
.cyberchef-header {
  position: static;
  height: 3em;
  display: flex;
  align-items: center;
  padding: 0 1em;
  background-color: #1c1c1e;
  color: #fff;
  border-bottom: 1px solid #343434;
  flex-shrink: 0;
}
.team-share-btn {
  display: flex;
  align-items: center;
  gap: 0.4em;
  background: none;
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: 4px;
  color: #fff;
  padding: 0.3em 0.8em;
  cursor: pointer;
  font-size: 0.9em;
}
.team-share-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}
.team-share-btn .material-icons {
  font-size: 1.2em;
}
.cyberchef-title {
  margin-left: 1em;
  font-size: 1.1em;
  font-weight: 500;
}
.cyberchef-frame {
  flex: 1;
  width: 100%;
  border: 0;
  display: block;
}
</style>
