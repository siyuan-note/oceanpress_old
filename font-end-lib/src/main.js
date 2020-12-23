import { createApp } from 'vue';
import App, { contentEL } from './App.vue';

function hashAndUpdate(e) {
    scrollIntoView(e.newURL);
}

function scrollIntoView(url) {
    const hash = url.split("#").pop();
    const target = document.querySelector(`[data-block-id="${hash}"]`);
    if (target) {
        var highlightClassName = [ 'hash_selected', 'hash_selected-highlight' ];
        target.classList.add(...highlightClassName);
        target.scrollIntoView();
        setTimeout(() => {
            target.classList.remove(...highlightClassName);

        }, 2000);
    }
}
window.addEventListener('hashchange', hashAndUpdate);
window.addEventListener("load", (e) => {

    const appEL = document.getElementById("app");
    if (appEL) {
        [ ...appEL.children ].forEach(el => contentEL.value.appendChild(el));
        createApp({ ...App }).mount(appEL);
    } else {
        throw "没有找到可供挂载 App 的元素";
    }


    // 当前链接可能指向了某一个块，尝试跳过去
    scrollIntoView(location.href);
    document.addEventListener("click", (e) => {
        if (e.target instanceof HTMLAnchorElement) {
            scrollIntoView(e.target.href);
        }
    });
});
