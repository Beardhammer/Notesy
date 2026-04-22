<template>
  <div id="editor-container" :class="{ 'header-collapsed': headerCollapsed }">
    <header-bar>
      <action icon="close" :label="$t('buttons.close')" @action="close()" />
      <title>{{ req.name }}</title>
      <action
        :icon="headerCollapsed ? 'expand_more' : 'expand_less'"
        :label="headerCollapsed ? 'Expand' : 'Collapse'"
        @action="headerCollapsed = !headerCollapsed"
      />
    </header-bar>

    <div id="editor" class="iframe-container">
      <iframe :src="codi_file_url" frameborder="0"></iframe>
    </div>
  </div>
</template>

<style scoped>
  #editor-container {
    height: 100vh;
    width: 100vw;
  }

  #editor-container >>> header {
    transition: height 0.15s ease, padding 0.15s ease;
  }

  #editor-container.header-collapsed >>> header {
    height: 2em;
    padding: 0.15em 0.5em;
  }

  #editor-container.header-collapsed >>> header title,
  #editor-container.header-collapsed >>> header .username {
    display: none;
  }

  #editor-container.header-collapsed #editor {
    margin-top: -2em;
    transition: margin-top 0.15s ease;
  }

  iframe {
    height: 100% !important;
    overflow: auto;
    width: 100vw;
  }
</style>

<script>
import { mapState } from "vuex";
import { codiMd } from "@/utils/constants";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import url from "@/utils/url";

export default {
  name: "codimdeditor",
  components: {
    HeaderBar,
    Action,
  },
  data: function () {
    return {
      headerCollapsed: false,
    };
  },
  computed: {
    ...mapState(["req", "user", "jwt"]),
    codi_file_url() {
      const filePath     = this.$route.path.split('/').slice(2, -1).join('/');
      const fileNameNoMD = this.$route.path.split('/').slice(-1)[0].split('.').slice(0,-1).join('.');
      const parts        = [filePath, fileNameNoMD].filter(Boolean);
      const noteId       = encodeURIComponent(parts.join('/'));
      const baseUrl      = codiMd.url.replace(/\/+$/, '');
      const fullUrl      = baseUrl + "/" + noteId + "?edit";
      return fullUrl;
    }
  },
  methods: {
    close() {
      this.$store.commit("updateRequest", {});
      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
  }
};
</script>
