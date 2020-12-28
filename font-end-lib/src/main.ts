import App from "./App.svelte";

const allEL = document.createElement("div");
const app = new App({
  target: allEL,
});

document.body.appendChild(allEL);

export default app;
