import { scrollIntoView } from "./lib/scroll_into_view";

import { vditorRender } from "./lib/vditor";

export function scrollIntoHash() {
  scrollIntoView(location.href);
}

async function render() {
  let old = null as any;
  while (1) {
    const mdContent = document.getElementById("static_app_llej");
    if (mdContent === null || old === mdContent) {
      await new Promise((s) => setTimeout(s, 80));
    } else {
      console.log("[render] ", mdContent);
      old = mdContent;

      /** ═════════🏳‍🌈 渲染 md 🏳‍🌈═════════  */
      vditorRender(mdContent);

      /** ═════════🏳‍🌈 快引用在当前页的跳转 🏳‍🌈═════════  */
      scrollIntoView(location.href);
      mdContent.addEventListener("click", (e) => {
        const el = e.target as any;
        scrollIntoView(el.href || el.src);
      });
    }
  }
}

render();
