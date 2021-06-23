/** 暴露到全局的一些配置 */
export const md2website = {
  gotoClick(e:Event) {
    console.log("gotoClick");
  },
  fragment:<DocumentFragment> document
};

globalThis["md2website"] = md2website;
