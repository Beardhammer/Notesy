import Vue from "vue";
import Vuex from "vuex";
import mutations from "./mutations";
import getters from "./getters";
import upload from "./modules/upload";
import shoutbox from "./modules/shoutbox";

Vue.use(Vuex);

const state = {
  user: null,
  req: {},
  oldReq: {},
  clipboard: {
    key: "",
    items: [],
  },
  jwt: "",
  progress: 0,
  loading: false,
  reload: false,
  selected: [],
  multiple: false,
  prompts: [],
  showShell: false,
  boards: [],
  currentBoardId: "",
};

export default new Vuex.Store({
  strict: true,
  state,
  getters,
  mutations,
  modules: { upload, shoutbox },
});
