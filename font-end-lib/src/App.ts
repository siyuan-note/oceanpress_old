import { defineComponent, ref, watchEffect } from "vue";
import { vditorRender } from "./lib/vditor";
export const contentEL = ref(document.createElement("div"));
export default defineComponent({
  name: "App",
  setup(props, ctx) {
    const articleEl = ref(null as null | HTMLElement);
    watchEffect(() => {
      if (contentEL.value && articleEl.value) {
        Array.from(contentEL.value.children).forEach((el) => articleEl.value!.appendChild(el));
        vditorRender(articleEl.value);
      }
    });

    return {
      contentEL,
      articleEl,
    };
  },
});
