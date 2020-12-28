/** 暴露到全局的一些配置 */
export const md2website = {
  gotoClick(e:any) {
    console.log("gotoClick");
  },
};

globalThis["md2website"] = md2website;
