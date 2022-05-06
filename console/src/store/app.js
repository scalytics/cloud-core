import { inject } from "vue";
import { defineStore } from "pinia";

const baseURL = import.meta.env.DEV ? "http://api.infinimesh.local" : window.location.origin.replace("console.", "api.")

export const useAppStore = defineStore("app", {
  state: () => ({
    token: "",
    me: {
      title: "",
    },
    theme: "dark",
    theme_pick: "system",
  }),
  getters: {
    base_url: () => baseURL,
    logged_in: (state) => state.token !== "",
    http: (state) => {
      return inject("axios").create({
        baseURL,
        headers: {
          Authorization: `Bearer ${state.token}`,
        },
      });
    },
  },
  actions: {
    logout() {
      this.$reset();
    },
  },

  persist: {
    enabled: true,
    strategies: [{ storage: localStorage, key: "infinimesh" }],
  },
});
