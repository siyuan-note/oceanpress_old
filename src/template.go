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

// ArticleInfo 结构
type ArticleInfo struct {
	PageTitle string
	Content   interface{}
	LevelRoot string
}

var globalF = template.FuncMap{"unescaped": unescaped}

var articleTemplate = HTMLtemplate.New("article").Funcs(globalF)

// ArticleRender 渲染文章html
func ArticleRender(info ArticleInfo) string {
	return ExecTemplate(articleTemplate, info)
}

var menuTemplate = HTMLtemplate.New("menu").Funcs(globalF)

type sonEntityI struct {
	WebPath string
}

// MenuInfo 菜单结构
type MenuInfo struct {
	SonEntityList []sonEntityI
	PageTitle     string
}

// MenuRender 渲染菜单
func MenuRender(info MenuInfo) string {
	return ExecTemplate(menuTemplate, info)
}
