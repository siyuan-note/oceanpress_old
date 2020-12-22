package main

import (
	"bytes"
	"html/template"
)

// HTMLtemplate 包含一些模板
var HTMLtemplate = template.Must(template.ParseGlob(TemplateDir + "*.html"))

func init() {
	template.Must(HTMLtemplate.ParseGlob(TemplateDir + "*/*.html"))
}
func unescaped(x string) interface{} { return template.HTML(x) }

// ExecTemplate 执行模板
/**
 * 现有模板 menu
 */
func ExecTemplate(t *template.Template, data interface{}) string {
	buf := new(bytes.Buffer)
	t.Execute(buf, data)
	return buf.String()
}

var globalF = template.FuncMap{"unescaped": unescaped}

var articleTemplate = HTMLtemplate.New("article").Funcs(globalF)
var menuTemplate = HTMLtemplate.New("menu").Funcs(globalF)
var embeddedBlockTemplate = HTMLtemplate.New("embeddedBlock").Funcs(globalF)
var blockRefTemplate = HTMLtemplate.New("blockRef").Funcs(globalF)

type sonEntityI struct {
	WebPath string
	IsDir   bool
	Name    string
}

// ArticleInfo 结构
type ArticleInfo struct {
	PageTitle string
	Content   interface{}
	LevelRoot string
}

// MenuInfo 菜单结构
type MenuInfo struct {
	SonEntityList []sonEntityI
	PageTitle     string
	LevelRoot string
}

// EmbeddedBlockInfo 嵌入块所需信息
type EmbeddedBlockInfo struct {
	Title   string
	Src     string
	Content interface{}
}

// BlockRefInfo 块引用所需信息
type BlockRefInfo struct {
	Title string
	Src   string
}

// ArticleRender 渲染文章html
func ArticleRender(info ArticleInfo) string {
	return ExecTemplate(articleTemplate, info)
}

// MenuRender 渲染菜单
func MenuRender(info MenuInfo) string {
	return ExecTemplate(menuTemplate, info)
}

// EmbeddedBlockRender 渲染嵌入块
func EmbeddedBlockRender(info EmbeddedBlockInfo) string {
	return ExecTemplate(embeddedBlockTemplate, info)
}

// BlockRefRender 渲染块引用
func BlockRefRender(info BlockRefInfo) string {
	return ExecTemplate(blockRefTemplate, info)
}
