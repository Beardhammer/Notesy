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

    <div id="editor"></div>
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

#editor {
  height: calc(100vh - 4em);
  transition: height 0.15s ease;
}

#editor-container.header-collapsed #editor {
  height: calc(100vh - 2em);
}
</style>

<script>
import { mapState } from "vuex";
import url from "@/utils/url";
import { baseURL, onlyOffice } from "@/utils/constants";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";

export default {
  name: "onlyofficeeditor",
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
  },
  created() {
    window.addEventListener("keydown", this.keyEvent);
  },
  beforeDestroy() {
    window.removeEventListener("keydown", this.keyEvent);
    this.editor.destroyEditor();
  },
  mounted: function () {
    let onlyofficeScript = document.createElement("script");
    onlyofficeScript.setAttribute(
      "src",
      `${onlyOffice.url}/web-apps/apps/api/documents/api.js`
    );
    document.head.appendChild(onlyofficeScript);

    /*eslint-disable */
    onlyofficeScript.onload = () => {
      let fileUrl = `http://filebrowser${baseURL}/api/onlyoffice/dl/${this.jwt}${url.encodePath(
        this.req.path
      )}`;

      // create a key from the last modified timestamp and the reversed file path (most specific part first)
      // replace all special characters (only these symbols are supported: 0-9, a-z, A-Z, -._=)
      // and truncate it (max length is 20 characters)
      const key = (
        Date.parse(this.req.modified).valueOf()
        + url
          .encodePath(this.req.path.split('/').reverse().join(''))
          .replaceAll(/[!~[\]*'()/,;:\-%+. ]/g, "")
      ).substring(0, 20);

      const config = {
        document: {
          fileType: this.req.extension.substring(1),
          key: key,
          title: this.req.name,
          url: fileUrl,
          permissions: {
            edit: this.user.perm.modify,
            download: this.user.perm.download,
            print: this.user.perm.download
          }
        },
        editorConfig: {
          callbackUrl: `http://filebrowser${baseURL}/api/onlyoffice/save/${this.jwt}${url.encodePath(this.req.path)}`,
          user: {
            id: `${this.user.username}`,
            name: `${this.user.username}`
          },
          customization: {
            autosave: true,
            forcesave: true
          },
          lang: this.user.locale,
          mode: this.user.perm.modify ? "edit" : "view"
        }
      };

      if(onlyOffice.jwtSecret != "") {
        fetch(`${baseURL}/api/onlyoffice/token`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Auth": this.jwt,
          },
          body: JSON.stringify(config),
        })
          .then((res) => res.text())
          .then((token) => {
            config.token = token;
            this.editor = new DocsAPI.DocEditor("editor", config);
          });
      } else {
        this.editor = new DocsAPI.DocEditor("editor", config);
      }
    };
    /*eslint-enable */
  },
  methods: {
    back() {
      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
    keyEvent(event) {
      if (!event.ctrlKey && !event.metaKey) {
        return;
      }

      if (String.fromCharCode(event.which).toLowerCase() !== "s") {
        return;
      }

      event.preventDefault();
      this.save();
    },
    close() {
      this.$store.commit("updateRequest", {});

      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
  },
};
</script>
