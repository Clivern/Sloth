/** @format */

import Vue from "vue";
import Vuex from "vuex";
import health from "./health.module";
import auth from "./auth.module";

Vue.use(Vuex);

export default new Vuex.Store({
	modules: {
		health,
		auth,
	},
});
