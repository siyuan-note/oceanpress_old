/** ═════════🏳‍🌈 复制编译后的资源到 views 中的资源目录 🏳‍🌈═════════  */
//@ts-check
import fs from "fs";
import fse from "fs-extra";

const indexHtml = fs.readFileSync("./dist/index.html").toString();

const assetsCode = `<!-- font-end-lib 生成的资源 star -->\n${indexHtml.match(/\/title>([\s\S]*)<\/head>/)[ 1 ]}\n<!-- font-end-lib 生成的资源 end -->`;

const headHtmlPath = "../src/views/head.html";
const headHtml = fs.readFileSync(headHtmlPath).toString();


const newHeadHtml = headHtml
    .replace(/<!-- font-end-lib 生成的资源 star -->([\s\S]*)<!-- font-end-lib 生成的资源 end -->/, assetsCode)
    .replace(`src="./assets`, `src="{{.LevelRoot}}assets/font-end-lib`)
    /** 在使用 file 模式的时候是没有办法访问 type="module" 资源的，但幸好也是不需要的  */
    .replace(`type="module"`, `defer`);
fs.writeFileSync(headHtmlPath, newHeadHtml);

fse.emptyDirSync('../src/views/assets/font-end-lib/');
fse.copy('./dist/assets/', '../src/views/assets/font-end-lib/');