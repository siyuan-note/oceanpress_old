import { defineComponent, onMounted, ref, watchEffect } from "vue";
import { scrollIntoView } from "./lib/scroll_into_view";
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

        // 当前链接可能指向了某一个块，尝试跳过去
        scrollIntoView(location.href);
      }
    });

    const onAnchorClick = (e: Event) => {
      if (e.target instanceof HTMLAnchorElement) {
        scrollIntoView(e.target.href);
      }
    };
    return {
      contentEL,
      articleEl,
      onAnchorClick,
    };
  },
});
