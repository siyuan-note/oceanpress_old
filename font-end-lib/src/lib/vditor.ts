import Vditor from "vditor";

/** 用来获取 vditor 实例的工具元素 */
const d = document.createElement("div");
const id = "test__" + Date.now();
d.setAttribute("id", id);
d.style.display = "none";
const vditor = new Vditor(d, { cache: { id: id } }).vditor;

export function vditorRender(previewElement: HTMLElement) {
  Vditor.setContentTheme(vditor.options.preview.theme.current, vditor.options.preview.theme.path);
  Vditor.codeRender(previewElement);
  //@ts-expect-error
  Vditor.highlightRender(JSON.stringify(vditor.options.preview.hljs), previewElement, vditor.options.cdn);
  Vditor.mathRender(previewElement, {
    cdn: vditor.options.cdn,
  //@ts-expect-error
    math: JSON.stringify(vditor.options.preview.math),
  });
  //@ts-expect-error
  Vditor.mermaidRender(previewElement, vditor.options.cdn);
  Vditor.flowchartRender(previewElement, vditor.options.cdn);
  Vditor.graphvizRender(previewElement, vditor.options.cdn);
  //@ts-expect-error
  Vditor.chartRender(previewElement, vditor.options.cdn);
  //@ts-expect-error
  Vditor.mindmapRender(previewElement, vditor.options.cdn);
  Vditor.abcRender(previewElement, vditor.options.cdn);
  Vditor.mediaRender(previewElement);
}
