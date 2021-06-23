import App from "./App.svelte";

const allEL = document.createElement("div");
allEL.classList.add(".markdown-body")
const app = new App({
  target: allEL,
});

if (document.body === null) {
  window.addEventListener("load", mount);
} else {
  mount();
}

function mount() {
  document.body.appendChild(allEL);
}

export default app;
