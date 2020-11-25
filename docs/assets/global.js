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
window.addEventListener("load", () => {
    // 当前链接可能指向了某一个块，尝试跳过去
    scrollIntoView(location.href);

    /** 用来获取 vditor 实例的工具元素 */
    const d = document.createElement("div");
    const id = "test__" + Date.now();
    d.setAttribute("id", id);
    d.style.display = "none";
    document.body.appendChild(d);

    const vditor = new Vditor(id).vditor;
    const previewElement = document.querySelector('.vditor-reset');
    Vditor.setContentTheme(vditor.options.preview.theme.current, vditor.options.preview.theme.path);
    Vditor.codeRender(previewElement);
    Vditor.highlightRender(JSON.stringify(vditor.options.preview.hljs), previewElement, vditor.options.cdn);
    Vditor.mathRender(previewElement, {
        cdn: vditor.options.cdn,
        math: JSON.stringify(vditor.options.preview.math)
    });
    Vditor.mermaidRender(previewElement, vditor.options.cdn);
    Vditor.flowchartRender(previewElement, vditor.options.cdn);
    Vditor.graphvizRender(previewElement, vditor.options.cdn);
    Vditor.chartRender(previewElement, vditor.options.cdn);
    Vditor.mindmapRender(previewElement, vditor.options.cdn);
    Vditor.abcRender(previewElement, vditor.options.cdn);
    Vditor.mediaRender(previewElement);
});