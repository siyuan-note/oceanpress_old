# md2website

将 MarkDown 文件转换为 html 的静态站点

[点击这里查看生成后的效果](https://2234839.github.io/md2website/) 静态文件位于 `./docs/`

## 运行方式 run

`go run .\src\main.go (sourceDir) (outDir)`

```go
go run .\src\main.go "C:\\Users\\llej\\AppData\\Local\\Programs\\SiYuan\\resources\\guide\\思源笔记用户指南" "D:\\code\\md2website\\docs"
```

## 待完成的功能点（按优先级降序排序）

* [ ] 块引用链接 `70%`
* [ ] 嵌入块渲染
* [ ] 菜单页面美化
* [ ] 代码块渲染
* [ ] 块引用当前页面预览
* [ ] 块链接可 copy
* [ ] 书签页
* [ ] 标签页
* [ ] 反链
* [ ] 嵌入块查询渲染