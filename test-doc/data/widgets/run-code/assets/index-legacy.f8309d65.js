System.register(["./vendor-legacy.3cc3c1bb.js"],(function(){"use strict";var e,t,n,r,i,s,o,a,u,l,c,d,p,v,g,m,f,b;return{setters:[function(y){e=y.c,t=y.d,n=y.r,r=y.o,i=y.w,s=y.a,o=y.b,a=y.e,u=y.u,l=y.f,c=y.g,d=y.h,p=y.i,v=y.j,g=y.k,m=y.l,f=y.m,b=y.n}],execute:function(){const y=l();var S=t({props:{inside:{type:String,default:"inside",required:!0},source:{type:String,default:""},title:{type:String,default:""},preamble:{type:String,default:""},version:{type:String,default:"16.6.0"}},emits:["update:inside","update:source","update:title","update:preamble","update:version"],setup(t,{emit:l}){const c=t,{inside:d,source:p,title:v,preamble:g,version:m}=function(t,n){const r={};for(const i in t)r[i]=e({get:()=>t[i],set(e){n(`update:${i}`,e)}});return r}(c,l),f=n(null),b=e({get:()=>m.value.trim().split(".").map(((e,t)=>0===t?e:"x")).join("."),set(e){m.value=e}});return r((async()=>{const e=window.RunKit.createNotebook({element:f.value,preamble:g.value,title:v.value,gutterStyle:d.value,source:p.value,nodeVersion:b.value});e.onLoad=async t=>{e.evaluate()},e.onSave=async()=>{e.getSource().then((e=>p.value=e)),e.getNodeVersion().then((e=>b.value=e))},i((()=>e.setSource(p.value))),i((()=>e.setTitle(v.value))),i((()=>e.setPreamble(g.value))),i((()=>e.setNodeVersion(b.value)))})),(e,t)=>(s(),o("pre",{class:"embed","data-gutter":u(d)},[a("div",{ref:f},null,512),y],8,["data-gutter"]))}}),h=t({setup(e){f.apiCache=!0;const t=c({source:"",inside:"inside",title:"",preamble:"",version:"16.6.0"});return d((async()=>{const e=p();e?v.getBlockAttr(e,"custom-config").then((e=>{if(e){const n=g(e.value);n?Object.assign(t,JSON.parse(n)):console.error("decode 失败, code:",e.value)}})).catch((e=>{console.error("查询config失败",e)})):console.error("获取当前挂件快id失败"),m()===m.env.siYuan&&i((()=>{v.setBlockAttrs({id:p(),attrs:{"custom-config":JSON.stringify(t)}}).then((e=>{}))}))})),(e,n)=>(s(),o(S,{source:u(t).source,"onUpdate:source":n[1]||(n[1]=e=>u(t).source=e),inside:u(t).inside,"onUpdate:inside":n[2]||(n[2]=e=>u(t).inside=e),title:u(t).title,"onUpdate:title":n[3]||(n[3]=e=>u(t).title=e),preamble:u(t).preamble,"onUpdate:preamble":n[4]||(n[4]=e=>u(t).preamble=e),version:u(t).version,"onUpdate:version":n[5]||(n[5]=e=>u(t).version=e)},null,8,["source","inside","title","preamble","version"]))}});const N=b(h);N.mount("#app"),N.config}}}));
//# sourceMappingURL=index-legacy.f8309d65.js.map
