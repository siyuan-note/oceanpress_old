package template

import (
	"bytes"
	"html/template"
	"path"
	textTemplate "text/template"

	conf "github.com/siyuan-note/oceanpress/src/conf"
	structAll "github.com/siyuan-note/oceanpress/src/struct"
	"github.com/siyuan-note/oceanpress/src/util"
)

// HTMLtemplate 包含一些模板
var globalF = template.FuncMap{"unescaped": unescaped}

var menuTemplate *template.Template
var HTMLtemplate *template.Template
var articleTemplate *template.Template
var embeddedBlockTemplate *template.Template
var blockRefTemplate *template.Template
var rssTemplate *textTemplate.Template
var redirectTemplate *template.Template

func init() {
	HTMLtemplate = template.Must(template.ParseGlob(path.Join(conf.TemplateDir, "./*.html")))
	articleTemplate = HTMLtemplate.New("article").Funcs(globalF)
	embeddedBlockTemplate = HTMLtemplate.New("embeddedBlock").Funcs(globalF)
	blockRefTemplate = HTMLtemplate.New("blockRef").Funcs(globalF)
	menuTemplate = HTMLtemplate.New("menu").Funcs(globalF)
	redirectTemplate = HTMLtemplate.New("redirect").Funcs(globalF)
	template.Must(HTMLtemplate.ParseGlob(path.Join(conf.TemplateDir, "./*/*.html")))

	// ParseGlob(path.Join(conf.TemplateDir, "./*.xml"))
	rssTemplate = textTemplate.Must(textTemplate.ParseGlob(path.Join(conf.TemplateDir, "./*.xml")))
	// c := textTemplate.Must(textTemplate)
}
func unescaped(x string) interface{} { return template.HTML(x) }

// ExecTemplate 执行模板
/**
 * 现有模板 menu
 */
func ExecTemplate(t *template.Template, data interface{}) string {
	buf := new(bytes.Buffer)
	err := t.Execute(buf, data)
	if err != nil {
		util.Warn("<html模板执行失败>", err)
	}
	return buf.String()
}

type SonEntityI struct {
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
	SonEntityList []SonEntityI
	PageTitle     string
	LevelRoot     string
}

func (r *MenuInfo) Render() string {
	return MenuRender(*r)
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
func EmbeddedBlockRender(info structAll.EmbeddedBlockInfo) string {
	return ExecTemplate(embeddedBlockTemplate, info)
}

// BlockRefRender 渲染块引用
func BlockRefRender(info structAll.BlockRefInfo) string {
	return ExecTemplate(blockRefTemplate, info)
}

// TemplateRender 将数据通过模板进行渲染 目前支持 EmbeddedBlockInfo BlockRefInfo 的处理
func TemplateRender(info interface{}) string {
	EmbeddedBlock, ok := info.(structAll.EmbeddedBlockInfo)
	if ok {
		return EmbeddedBlockRender(EmbeddedBlock)
	}

	BlockRef, ok := info.(structAll.BlockRefInfo)
	if ok {
		return BlockRefRender(BlockRef)
	}

	Redirect, ok := info.(structAll.RedirectInfo)
	if ok {
		return ExecTemplate(redirectTemplate, Redirect)
	}
	RssInfo, ok := info.(structAll.RssInfo)
	if ok {
		buf := new(bytes.Buffer)
		err := rssTemplate.ExecuteTemplate(buf, "RSS", RssInfo)
		if err != nil {
			util.Warn("<rss模板执行失败>", err)
		}
		r := buf.String()
		return r
	}
	util.Warn("没有找到对应的 template render", info)
	return "[渲染错误]没有找到对应的 template render"
}

func HTML(s string) template.HTML {
	return template.HTML(s)
}
