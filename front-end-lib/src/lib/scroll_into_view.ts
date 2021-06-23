import { md2website } from "./md2website.global";

export function scrollIntoView(url: string) {
  const hash = url.split("#").pop();
  const target = md2website.fragment.querySelector(`[data-n-id="${hash}"]`);
  if (target) {
    location.hash = hash;

    var highlightClassName = ["hash_selected", "hash_selected-highlight"];
    target.classList.add(...highlightClassName);
    target.scrollIntoView();
    setTimeout(() => {
      target.classList.remove(...highlightClassName);
    }, 2000);
  } else {
    console.warn("无法定位到", hash, "所指向的元素");
  }
}
