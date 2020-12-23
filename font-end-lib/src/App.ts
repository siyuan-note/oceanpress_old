import { defineComponent, ref, watchEffect } from "vue";
export const contentEL = ref(document.createElement("div"));
export default defineComponent({
  name: "App",
  setup(props, ctx) {
    const articleEl = ref(null as null | HTMLElement);
    watchEffect(() => {
      if (contentEL.value && articleEl.value) {
        Array.from(contentEL.value.children).forEach((el) => articleEl.value!.appendChild(el));
        console.log("[contentEL.value]", contentEL.value.childElementCount, contentEL.value);
        console.log("[articleEl]", articleEl.value);
      }
    });

    return {
      contentEL,
      articleEl,
    };
  },
  data() {
    return { a: 33 };
  },
});
