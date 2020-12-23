declare var Vditor: any;

/** 用来获取 vditor 实例的工具元素 */
const d = document.createElement("div");
const id = "test__" + Date.now();
d.setAttribute("id", id);
d.style.display = "none";
document.body.appendChild(d);

const vditor = new Vditor(id).vditor;

export function vditorRender(previewElement: Element) {
  Vditor.setContentTheme(vditor.options.preview.theme.current, vditor.options.preview.theme.path);
  Vditor.codeRender(previewElement);
  Vditor.highlightRender(JSON.stringify(vditor.options.preview.hljs), previewElement, vditor.options.cdn);
  Vditor.mathRender(previewElement, {
    cdn: vditor.options.cdn,
    math: JSON.stringify(vditor.options.preview.math),
  });
  Vditor.mermaidRender(previewElement, vditor.options.cdn);
  Vditor.flowchartRender(previewElement, vditor.options.cdn);
  Vditor.graphvizRender(previewElement, vditor.options.cdn);
  Vditor.chartRender(previewElement, vditor.options.cdn);
  Vditor.mindmapRender(previewElement, vditor.options.cdn);
  Vditor.abcRender(previewElement, vditor.options.cdn);
  Vditor.mediaRender(previewElement);
}
