/** ═════════🏳‍🌈 复制编译后的资源到 views 中的资源目录 🏳‍🌈═════════  */
//@ts-check
import fs from "fs";
import fse from "fs-extra";
import { join } from "path";

const indexHtml = fs.readFileSync("./dist/index.html").toString();

const assetsCode = `<!-- font-end-lib 生成的资源 star -->\n${indexHtml.match(/\/title>([\s\S]*)<\/head>/)[ 1 ]}\n<!-- font-end-lib 生成的资源 end -->`;

const headHtmlPath = "../src/views/head.html";
const headHtml = fs.readFileSync(headHtmlPath).toString();


const newHeadHtml = headHtml
    .replace(/<!-- font-end-lib 生成的资源 star -->([\s\S]*)<!-- font-end-lib 生成的资源 end -->/, assetsCode)
    .replace(`src="./assets`, `src="{{.LevelRoot}}assets/font-end-lib`)
    /** 在使用 file 模式的时候是没有办法访问 type="module" 资源的，~~但幸好也是不需要的~~
     * 在使用多个脚本的时候这样会导致可能存在的变量名冲突，这块有点难受， file 模式不应该放弃，但在用户引入其他脚本导致冲突的时候怎么办？。再包裹一层？
     *
      */
    .replace(`type="module"`, `defer`);
fs.writeFileSync(headHtmlPath, newHeadHtml);


const font_end_lib = '../src/views/assets/font-end-lib/';
const assets = './dist/assets/';
fse.emptyDirSync(font_end_lib);

// fse.copy(assets, font_end_lib);

const dir = fse.readdirSync(assets);

dir.forEach(path => {
    const filePath = join(assets, path);
    if (path.endsWith(".js")) {
        const rawJs = fse.readFileSync(filePath).toString();
        const js = `(function(){${rawJs}})()`;

        fse.writeFileSync(join(font_end_lib, path), js);
        fse.writeFileSync("./test",js);
    } else {
        console.log("copy");
        fse.copy(filePath, font_end_lib);
    }
});