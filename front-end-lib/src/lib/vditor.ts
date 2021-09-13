import Vditor from "vditor";

/** 用来获取 vditor 实例的工具元素 */
const d = document.createElement("div");
const id = "test__" + Date.now();
d.setAttribute("id", id);
d.style.display = "none";
const vditor = new Vditor(d, {
  cache: { id: id },
  // cdn: "https://cdn.jsdelivr.net/npm/vditor@3.8.4",
})
function generalAdaptation(
  adapterTarget: {
    getElements: (element: HTMLElement) => NodeListOf<Element>;
    getCode: (mathElement: Element) => string;
  },
  type: string,
) {
  Object.assign(adapterTarget, {
    getElements: (element) => {
      return element.querySelectorAll(`[data-subtype=${type}]`);
    },
    getCode: (element) => (element as HTMLElement).dataset.content,
  });
}

generalAdaptation(Vditor.adapterRender.mathRenderAdapter, "math");
{
  //流程图
  generalAdaptation(Vditor.adapterRender.mermaidRenderAdapter, "mermaid");
  Vditor.adapterRender.mermaidRenderAdapter.getCode = (element) => {
    element.innerHTML = (element as HTMLElement).dataset.content;
    return element.textContent;
  };
}
{
  //脑图
  generalAdaptation(Vditor.adapterRender.mindmapRenderAdapter, "mindmap");
  Vditor.adapterRender.mindmapRenderAdapter.getCode = (el) =>
    (el as HTMLElement).dataset.parseContent;
}
generalAdaptation(Vditor.adapterRender.chartRenderAdapter, "echarts");
generalAdaptation(Vditor.adapterRender.abcRenderAdapter, "abc");
generalAdaptation(Vditor.adapterRender.graphvizRenderAdapter, "graphviz");
Vditor.adapterRender.graphvizRenderAdapter.getElements = (e) => {
  return e.querySelectorAll(`[data-subtype=graphviz]`);
};
generalAdaptation(Vditor.adapterRender.flowchartRenderAdapter, "flowchart");
generalAdaptation(Vditor.adapterRender.plantumlRenderAdapter, "plantuml");

export async function vditorRender(previewElement: HTMLElement) {
  // const cdn = v.options.cdn.replace(/-adapter\d+/, "");
  console.log(vditor);
  while(!vditor.vditor){
    await sleep(3)
  }
  const v=vditor.vditor
  Vditor.setContentTheme(
    v.options.preview.theme.current,
    v.options.preview.theme.path,
  );
  Vditor.codeRender(previewElement);
  Vditor.highlightRender(
    //@ts-expect-error
    JSON.stringify(v.options.preview.hljs),
    previewElement,
  );
  Vditor.mathRender(previewElement, {
    //@ts-expect-error
    math: JSON.stringify(v.options.preview.math),
  });
  //@ts-expect-error
  Vditor.mermaidRender(previewElement, );
  Vditor.flowchartRender(previewElement, );
  Vditor.graphvizRender(previewElement, );
  //@ts-expect-error
  Vditor.chartRender(previewElement, );
  //@ts-expect-error
  Vditor.mindmapRender(previewElement, );
  Vditor.abcRender(previewElement, );
  Vditor.mediaRender(previewElement);
  Vditor.plantumlRender(previewElement, )
}
function sleep(ms:number){
  return new Promise((s)=>{
    setTimeout(s,ms)
  })
}