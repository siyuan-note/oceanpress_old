<script lang="ts">
    import { scrollIntoView } from "../lib/scroll_into_view";

    import { vditorRender } from "../lib/vditor";
    async function render() {
        const mdContent = document.getElementById("static_app_llej");
        if (mdContent === null) {
            setTimeout(render, 80);
        } else {
            console.log(mdContent, mdContent.textContent.length);
            /** ═════════🏳‍🌈 渲染 md 🏳‍🌈═════════  */
            vditorRender(mdContent);

            /** ═════════🏳‍🌈 快引用在当前页的跳转 🏳‍🌈═════════  */
            scrollIntoView(location.href);
            mdContent.addEventListener("click", (e) => {
                if (e.target instanceof HTMLAnchorElement) {
                    scrollIntoView(e.target.href);
                }
            });
        }
    }

    render();
</script>

<svelte:options tag="md2website-md" />
