# md2website

将 MarkDown 文件转换为 html 的静态站点

[点击这里查看生成后的效果](https://2234839.github.io/md2website/) 静态文件位于 `./docs/`

## 运行方式 run

[下载地址](https://github.com/2234839/md2website/releases)

| 平台    | 命令                                                      | 备注                         |
| ------- | --------------------------------------------------------- | ---------------------------- |
| Windows | `.\windows_md2website.exe (sourceDir) (outDir) (viewDir)` |                              |
| Linux   | `.\linux_md2website (sourceDir) (outDir) (viewDir)`       |                              |
| Mac     | `.\mac_md2website (sourceDir) (outDir) (viewDir)`         | 我没有尝试过，因为我没有 mac |
| 源码    | `go run .\src\ (sourceDir) (outDir) (viewDir)`            |                              |

`(sourceDir)` 是文档所在目录, `(outDir)` 是你要输出的目录， `(viewDir)` 是视图文件的目录可以直接使用 `./src/views/` (src 内携带了一些 .go 文件，这个是可以不要的，重点在于 views 目录下的文件，可以自行修改其中的文件来定制一些效果)

示例（使用源码，重点是后面的三个参数）：

```go
go run .\src\ "C:/Users/llej/AppData/Local/Programs/SiYuan/resources/guide/思源笔记用户指南" "D:/code/md2website/docs" "./src/views/"
```

## 待完成的功能点（按优先级降序排序）

| 可用 | 功能名 | 大致进度 |
| --- | --- | --- |
| ❎🔨 | [#4 菜单页面美化](https://github.com/2234839/md2website/issues/4) | `15%` |
| ✅🔨 | [#2 嵌入块渲染](https://github.com/2234839/md2website/issues/2) 目前不支持循环引用 | `60%` |
| ⭕ | 目录树 |  |
| ⭕ | 页面 header 与 footer |  |
| ⭕ | 块引用当前页面预览 |  |
| ⭕ | 块链接可 copy |  |
| ⭕ | 书签页 |  |
| ⭕ | 标签页 |  |
| ⭕ | 反链 |  |
| ✅🔨 | [#1 块引用链接](https://github.com/2234839/md2website/issues/1) | `92%` |
| ⭕ | 嵌入块查询渲染 |  |
| ✅ | 支持 {.text} 这样的锚文本 | `100%` |
| ✅ | [#3 代码高亮 以及 数学公式和脑图等渲染](https://github.com/2234839/md2website/issues/3) [点击这里查看生成后的效果](https://2234839.github.io/md2website/Markdown%20%e4%bd%bf%e7%94%a8%e6%8c%87%e5%8d%97/Markdown%20%e5%ae%8c%e6%95%b4%e7%a4%ba%e4%be%8b.html#%E6%95%B0%E5%AD%A6%E5%85%AC%E5%BC%8F) 、还需要修改 vditor 等资源的引用为本地文件（不是很重要之后再说） | `100%` |

1. ✅ 表示基本可以使用了
2. 🔨 表示还在修改中
3. ❎ 表示有一小部分功能可以使用了但还存在比较大的问题
4. ❌ 表示不可用
5. ⭕ 表示尚未开始

## 相关

### 使用本产品的网站

> 欢迎您向我[提交您的网站地址](mailto:admin@shenzilong.cn)，我将在审阅后添加到此处

- [崮生 • 一些随笔 🎨](https://shenzilong.cn)

## 开发

项目组织结构

### /font-end-lib

> 这个目录下是一个使用 svelte 的前端项目，目的是为了提供前端技术上的支持

这个项目主要的思路是提供 web components 组件的 js 文件，然后再 html 的模板中使用

### build

`go build -o md2website.exe .\src\`
