import { scrollIntoView } from "./lib/scroll_into_view";

import { vditorRender } from "./lib/vditor";

export function scrollIntoHash() {
  scrollIntoView(location.href);
}

async function render() {
  let old = null as any;
  while (1) {
    const mdContent = document.getElementById("static_app_llej");
    if (mdContent!=null) {
      // 在用户选中一些元素后 隐藏 a 标签后面的图片，便于用户复制
      if(window.getSelection().type == "Range"){
        mdContent.classList.add("eventSelection")
      }else{
        mdContent.classList.remove("eventSelection")
      }
    }
    if (mdContent === null || old === mdContent) {
      await new Promise((s) => setTimeout(s, 80));
    } else {
      console.log("[render] ", mdContent);
      old = mdContent;

      /** ═════════🏳‍🌈 渲染 md 🏳‍🌈═════════  */
      vditorRender(mdContent);

      /** ═════════🏳‍🌈 块引用在当前页的跳转 🏳‍🌈═════════  */
      scrollIntoView(location.href);
      mdContent.addEventListener("click", (e) => {
        const el = e.target as any;
        let path = el.href || el.src;
        if (path) {
          scrollIntoView(path);
        }
      });
    }
  }
}

render();
