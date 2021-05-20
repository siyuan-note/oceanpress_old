import Vditor from "vditor";

/** 用来获取 vditor 实例的工具元素 */
const d = document.createElement("div");
const id = "test__" + Date.now();
d.setAttribute("id", id);
d.style.display = "none";
const vditor = new Vditor(d, {
  cache: { id: id },
  cdn: "https://cdn.jsdelivr.net/npm/vditor@3.8.4",
}).vditor;

function generalAdaptation(
  adapterTarget: {
    getMathElements: (element: HTMLElement) => NodeListOf<Element>;
    getCode: (mathElement: Element) => string;
  },
  type: string,
) {
  Object.assign(adapterTarget, {
    getMathElements: (element) => {
      return element.querySelectorAll(`[data-subtype=${type}]`);
    },
    getCode: (element) => (element as HTMLElement).dataset.content,
  });
}

generalAdaptation(Vditor.adapter.mathRenderAdapter, "math");
{
  //流程图
  generalAdaptation(Vditor.adapter.mermaidRenderAdapter, "mermaid");
  Vditor.adapter.mermaidRenderAdapter.getCode = (element) => {
    element.innerHTML = (element as HTMLElement).dataset.content;
    return element.textContent;
  };
}
{
  //脑图
  generalAdaptation(Vditor.adapter.mindmapRenderAdapter, "mindmap");
  Vditor.adapter.mindmapRenderAdapter.getCode = (el) =>
    (el as HTMLElement).dataset.parseContent;
}
generalAdaptation(Vditor.adapter.chartRenderAdapter, "echarts");
generalAdaptation(Vditor.adapter.abcRenderAdapter, "abc");
generalAdaptation(Vditor.adapter.graphvizRenderAdapter, "graphviz");
Vditor.adapter.graphvizRenderAdapter.getMathElements = (e) => {
  debugger;
  return e.querySelectorAll(`[data-subtype=graphviz]`);
};
generalAdaptation(Vditor.adapter.flowchartRenderAdapter, "flowchart");
generalAdaptation(Vditor.adapter.plantumlRenderAdapter, "plantuml");

export function vditorRender(previewElement: HTMLElement) {
  const cdn = vditor.options.cdn.replace(/-adapter\d+/, "");
  Vditor.setContentTheme(
    vditor.options.preview.theme.current,
    vditor.options.preview.theme.path,
  );
  Vditor.codeRender(previewElement);
  Vditor.highlightRender(
    //@ts-expect-error
    JSON.stringify(vditor.options.preview.hljs),
    previewElement,
    cdn,
  );
  Vditor.mathRender(previewElement, {
    cdn: cdn,
    //@ts-expect-error
    math: JSON.stringify(vditor.options.preview.math),
  });
  //@ts-expect-error
  Vditor.mermaidRender(previewElement, cdn);
  Vditor.flowchartRender(previewElement, cdn);
  Vditor.graphvizRender(previewElement, cdn);
  //@ts-expect-error
  Vditor.chartRender(previewElement, cdn);
  //@ts-expect-error
  Vditor.mindmapRender(previewElement, cdn);
  Vditor.abcRender(previewElement, cdn);
  Vditor.mediaRender(previewElement);
  Vditor.plantumlRender(previewElement, cdn)
}
