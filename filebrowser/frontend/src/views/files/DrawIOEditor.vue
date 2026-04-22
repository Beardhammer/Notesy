<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="$t('buttons.close')" @action="close()" />
      <title>{{ req.name }}</title>
    </header-bar>

    <breadcrumbs base="/files" noLink />
    <div id='warning-banner'><i icon='warning'/>Collaboration is NOT supported.  Tread carefully, your changes matter &lt;3</div>
    <div id="editor">
      <iframe :src="drawio_url" ref="drawioIframe" frameborder="0"></iframe>
    </div>
  </div>
</template>

<style scoped>
  #editor-container {
    height: 100vh;
    width: 100vw;
  }

  iframe {
    height: 100% !important;
    overflow: auto;
    width: 100vw;
  }

  #warning-banner {
    width: 100%;
    background-color: #ffcc00; /* Yellow background for warning */
    color: black; /* Text color */
    text-align: center; /* Center the text */
    padding: 10px 0; /* Some padding for top and bottom */
    font-size: 1.2em; /* Slightly larger text */
    border-top: 2px solid red; /* Optional: red bottom border for emphasis */
    border-bottom: 2px solid red; /* Optional: red bottom border for emphasis */
  }
</style>

<script>
import { mapState } from "vuex";
import url from "@/utils/url";
import { baseURL, drawIO } from "@/utils/constants";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";

export default {
  name: "drawioeditor",
  components: {
    HeaderBar,
    Action,
    Breadcrumbs,
  },


  data: function () {
    return {};
  },


  computed: {
    ...mapState(["req", "user", "jwt"]),
    breadcrumbs() {
      let parts = this.$route.path.split("/");

      if (parts[0] === "") {
        parts.shift();
      }

      if (parts[parts.length - 1] === "") {
        parts.pop();
      }

      let breadcrumbs = [];

      for (let i = 0; i < parts.length; i++) {
        breadcrumbs.push({ name: decodeURIComponent(parts[i]) });
      }

      breadcrumbs.shift();

      if (breadcrumbs.length > 3) {
        while (breadcrumbs.length !== 4) {
          breadcrumbs.shift();
        }

        breadcrumbs[0].name = "...";
      }

      return breadcrumbs;
    },
    drawio_url() {
      // Serialize the parameters into &param=value
      const serializeParams = (params) => {
        return Object.keys(params).map(key => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`).join('&');
      }

      // define the embed config
      let urlParams = {
        embed: 1,
        dark: 1,
        proto: 'json',
        local: 1, // local mode only
        spin: 1, // spinner while loading XML
        autosave: 0,  // disable autosave
        saveAndExit: 0, // remove buttons
        noSaveBtn: 0, /// show the save button
        noExitBtn: 1, // remove buttons
        modified: 'Modified',
        gapi: 0, // disable google integration
        db: 0, // disable dropbox integration
        od: 0, // disable onedrive integration
        tr: 0, // disable trello integration
        gl: 0, // disable github integration
        drive: 0 // disable gitlab integration

      };

      // Append the URL params to drawio embed
      let url = `${drawIO.url}?${serializeParams(urlParams)}`
      return url
    },
    fileUrl() {
      return `${window.location.protocol}//${window.location.host}${baseURL}/api/raw${url.encodePath(this.req.path)}`
    },
  },


  beforeDestroy() {
    window.removeEventListener('message', this.receiveMessage, false);
  },


  mounted: async function () {
    this.initDrawio();
  },


  methods: {
    close() {
      if (this.detectModification()) { return }
      this.$store.commit("updateRequest", {});
      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },

    // Draw IO init
    initDrawio() {
      window.addEventListener('message', this.receiveMessage, false);
    },

    // Load diagram into drawio
    async fetchAndLoadDiagram() {
      console.log('Fetching diagram: ', this.fileUrl)
      const xmlUrl = this.fileUrl;
      const xml = await this.fetchDiagramXML(xmlUrl);
      this.postMessageToDrawio('load', { xml });
    },

    // Request with exception handling
    async fetchDiagramXML(url) {
      try {
        const response = await fetch(url, {cache: "no-cache"});
        return await response.text();
      } catch (error) {
        console.error('Error fetching diagram XML:', error);
      }
    },

    // Send messages to drawio
    postMessageToDrawio(action, data = {}) {
      const iframe = this.$refs.drawioIframe;
      if (iframe && iframe.contentWindow) {
        iframe.contentWindow.postMessage(JSON.stringify({ action, ...data }), '*');
      }
    },

    // message handler for event
    receiveMessage(event) {
      // ignore events not from drawio
      const drawio_host = new URL(this.drawio_url);
      if (event.origin !== drawio_host.origin) {
        return;
      }

      const message = JSON.parse(event.data);
      switch(message.event) {
        case 'init':
          console.log('Draw.io initialized');
          this.fetchAndLoadDiagram();
          break;
        case 'save':
          this.saveFile(message.xml);
          console.log('Draw.io diagram saved');
          this.$showSuccess(this.$t("success.drawIoSaved"));
          break;
      }
    },

    // save the XML using the callback URL from FB
    async saveFile(data) {
      const url = `${window.location.protocol}//${window.location.host}${baseURL}/api/drawio/callback?auth=${this.jwt}&save=${encodeURIComponent(this.req.path)}`

      const response = await fetch(url, {
        method: "POST",
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({'diagram': data})
      })

        // Check if the request was successful
      if (!response.ok) {
        console.log("Error saving file: ", response.status);
        throw new Error(`HTTP error! status: ${response.status}`);
      } else {
        console.log("File saved!")
      }
    },

    detectModification() {
      const iframe = this.$refs.drawioIframe;
      const changes = iframe.contentWindow.document.getElementsByClassName('geStatus')[0].textContent.length>0
      if (changes) {
        return !confirm(this.$t("success.drawIoChangesDetected"))
      } else {
        return false
      }
    }
  },
};
</script>
