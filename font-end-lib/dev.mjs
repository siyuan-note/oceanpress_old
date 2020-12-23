/** ═════════🏳‍🌈 删除资源文件，为了调试 🏳‍🌈═════════  */
//@ts-check
import fs from "fs";
import fse from "fs-extra";

fse.emptyDirSync('../docs/assets/font-end-lib/');

const headHtmlPath = "../src/views/head.html";
const headHtml = fs.readFileSync(headHtmlPath).toString();
const assetsCode = `<!-- font-end-lib 生成的资源 star --><script type="module" src="http://localhost:3000/vite/client"></script>
<script type="module" src="http://localhost:3000/src/main.js"></script><!-- font-end-lib 生成的资源 end -->`;

const newHeadHtml = headHtml
    .replace(/<!-- font-end-lib 生成的资源 star -->([\s\S]*)<!-- font-end-lib 生成的资源 end -->/, assetsCode);
fs.writeFileSync(headHtmlPath, newHeadHtml);