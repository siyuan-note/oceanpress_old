/** ═════════🏳‍🌈 复制编译后的资源到 views 中的资源目录 🏳‍🌈═════════  */
//@ts-check
import fs from "fs";
import fse from "fs-extra";
import { join } from "path";

// const indexHtml = fs.readFileSync("./dist/index.html").toString();

// const assetsCode = `<!-- front-end-lib 生成的资源 star -->\n${indexHtml.match(/\/title>([\s\S]*)<\/head>/)[ 1 ]}\n<!-- front-end-lib 生成的资源 end -->`;

// const headHtmlPath = "../src/views/head.html";
// const headHtml = fs.readFileSync(headHtmlPath).toString();


// const newHeadHtml = headHtml
//     .replace(/<!-- front-end-lib 生成的资源 star -->([\s\S]*)<!-- front-end-lib 生成的资源 end -->/, assetsCode)
//     .replace(`src="./assets`, `src="{{.LevelRoot}}assets/front-end-lib`)

// fs.writeFileSync(headHtmlPath, newHeadHtml);


const font_end_lib = '../src/views/assets/front-end-lib/';
const assets = './public/build/';
fse.emptyDirSync(font_end_lib);
fse.copy(assets, font_end_lib);