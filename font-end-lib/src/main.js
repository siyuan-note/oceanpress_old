//@ts-check
import { createApp } from 'vue';
import App, { contentEL } from './App.vue';
import { scrollIntoView } from "./lib/scroll_into_view";
function hashAndUpdate(e) {
    scrollIntoView(e.newURL);
}

window.addEventListener('hashchange', hashAndUpdate);
console.log("开始执行");
window.addEventListener("load", (e) => {
    console.log("内页脚本加载成功");
    const appEL = document.getElementById("static_app_llej");
    if (appEL) {
        [ ...appEL.children ].forEach(el => contentEL.value.appendChild(el));
        createApp({ ...App }).mount(appEL);
    } else {
        throw "没有找到可供挂载 App 的元素";
    }

});
