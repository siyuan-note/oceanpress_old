import { md2website } from "./lib/md2website.global";
import { scrollIntoView } from "./lib/scroll_into_view";

import { vditorRender } from "./lib/vditor";

export function scrollIntoHash() {
  scrollIntoView(location.href);
}
async function render() {
  let old = null as any;
  while (1) {
    const mdContent = md2website.fragment.getElementById("static_app_llej");
    if (mdContent != null) {
      // 在用户选中一些元素后 隐藏 a 标签后面的图片，便于用户复制
      if (window.getSelection().type == "Range") {
        mdContent.classList.add("eventSelection");
      } else {
        mdContent.classList.remove("eventSelection");
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
      md2website.gotoClick = (e) => {
        const el = e.target as any;
        let path = el.href || el.src;
        if (path) {
          scrollIntoView(path);
        }
      };
      /** ═════════🏳‍🌈 渲染 命名、别名等 🏳‍🌈═════════  */
      Array.from(md2website.fragment.querySelectorAll("[data-n-id]")).map(
        (el) => {
          const attrFragment = document.createDocumentFragment();
          function addItem(name: string): boolean {
            const value = el.getAttribute(name);
            if (value) {
              const el = document.createElement("div");
              el.textContent = value;
              el.classList.add("protyle-attr--" + name);
              attrFragment.appendChild(el);
              return true;
            } else {
              return false;
            }
          }
          addItem("name");
          addItem("alias");
          if (addItem("memo")) {
            attrFragment.lastElementChild.setAttribute(
              "title",
              attrFragment.lastElementChild.textContent
            );
            attrFragment.lastElementChild.textContent = "";
          }
          addItem("bookmark");

          if (attrFragment.childNodes.length > 0) {
            const attrDiv = document.createElement("div");
            attrDiv.classList.add("protyle-attr");
            attrDiv.appendChild(attrFragment);
            el.appendChild(attrDiv);
          }
        }
      );
    }
  }
}

render();
