package main

import (
	"bytes"
	"strings"

	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/render"
)

// LuteEngine lute 实例
var LuteEngine = lute.New()

/** 用于从 md 文档中解析获得一些结构性信息 */
var mdStructuredLuteEngine = lute.New()

/** 当前被处理的 entity */
var baseEntity FileEntity

func init() {
	/** 对引用块进行渲染 */
	LuteEngine.SetBlockRef(true)
	// /** 渲染 id （渲染为空） */
	LuteEngine.SetKramdownIAL(true)
	// /** 标题的链接 a 标签渲染 */
	LuteEngine.SetHeadingAnchor(true)
	LuteEngine.SetKramdownIALIDRenderName("data-block-id")

	mdStructuredLuteEngine.SetKramdownIAL(true)
	mdStructuredLuteEngine.SetKramdownIALIDRenderName("data-block-id")
	// 当前正在被渲染的块的 id
	var id string
	/** 获取块的id */
	var getBlockID = func(n *ast.Node, _ bool) (string, ast.WalkStatus) {
		id = n.TokensStr()
		return "", ast.WalkContinue
	}
	LuteEngine.Md2HTMLRendererFuncs[ast.NodeBlockRefID] = getBlockID
	LuteEngine.Md2HTMLRendererFuncs[ast.NodeBlockEmbedID] = getBlockID

	/** 块引用渲染 */
	LuteEngine.Md2HTMLRendererFuncs[ast.NodeBlockRefText] = func(n *ast.Node, entering bool) (string, ast.WalkStatus) {
		var html string
		if entering {
			fileEntity, _ := FindFileEntityFromID(id)
			var src string
			if fileEntity.path != "" {
				src = FileEntityRelativePath(baseEntity, fileEntity, id)
			}

			html = `<a href="` + src + `" class="c-block-ref" data-block-type="` + n.Type.String() + `">` + n.Text() + `</a>`
		}
		return html, ast.WalkSkipChildren
	}
	/** 嵌入块渲染 */
	LuteEngine.Md2HTMLRendererFuncs[ast.NodeBlockEmbedText] = func(n *ast.Node, entering bool) (string, ast.WalkStatus) {

		var html string
		if entering {
			fileEntity, _ := FindFileEntityFromID(id)
			var src string
			test := baseEntity
			if fileEntity.path != "" {
				src = FileEntityRelativePath(baseEntity, fileEntity, id)
			}

			html = `<div title="尚未开发完成，完成后应该直接渲染对应部分的数据">
			<a href="` + src + `">块引用-> ` + n.ID + "||" + id + "||" + n.Text() + test.id + `</a>
		  </div>`
		}
		return html, ast.WalkSkipChildren
	}
}

// MdStructInfo md 结构信息
type MdStructInfo struct {
	blockID   string
	blockType string
	mdContent string
}

// GetMdStructInfo 从 md 获取结构信息
func GetMdStructInfo(name string, md string) []MdStructInfo {

	luteEngine := mdStructuredLuteEngine
	tree := parse.Parse(name, []byte(md), luteEngine.Options)

	var infoList []MdStructInfo
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if entering {
			return ast.WalkContinue
		}

		if nil == n.FirstChild {
			return ast.WalkSkipChildren
		}
		content := renderBlockMarkdown(n)
		infoList = append(infoList, MdStructInfo{blockID: n.IALAttr("id"), blockType: n.Type.String(), mdContent: content})

		return ast.WalkContinue
	})
	return infoList
}

func renderBlockMarkdown(node *ast.Node) string {
	root := &ast.Node{Type: ast.NodeDocument}
	luteEngine := mdStructuredLuteEngine

	tree := &parse.Tree{Root: root, Context: &parse.Context{Option: luteEngine.Options}}
	renderer := render.NewFormatRenderer(tree)
	renderer.Writer = &bytes.Buffer{}
	renderer.NodeWriterStack = append(renderer.NodeWriterStack, renderer.Writer)
	ast.Walk(node, func(n *ast.Node, entering bool) ast.WalkStatus {
		rendererFunc := renderer.RendererFuncs[n.Type]
		return rendererFunc(n, entering)
	})
	return strings.TrimSpace(renderer.Writer.String())
}

// FileEntityToHTML 转 html
func FileEntityToHTML(entity FileEntity) string {
	baseEntity = entity
	return LuteEngine.MarkdownStr("", entity.mdStr)
}

// FileEntityRelativePath 计算他们变成 html 文件之后的相对路径
func FileEntityRelativePath(base FileEntity, cur FileEntity, id string) string {
	// 减一是因为 路径开头必有 / 而这里只需要跳到这一层
	count := strings.Count(base.relativePath, "/") - 1

	l2 := strings.Split(cur.relativePath, "/")

	url := strings.Repeat("../", count)
	url += strings.Join(l2[1:], "/")
	url = FilePathToWebPath(url)
	url += "#" + id
	return url
}

// FilePathToWebPath 将相对文件路径转为 web路径，主要是去除文件中的id 以及添加 .html
func FilePathToWebPath(filePath string) string {
	// 这里的判定其实未必准，但先不考虑用户自定义如此后缀的情况
	if strings.HasSuffix(filePath, ".sy.md") {
		return filePath[0:len(filePath)-29] + ".html"
	} else if strings.HasSuffix(filePath, ".md") {
		return filePath[0:len(filePath)-3] + ".html"
	} else {
		// 大概率是空
		return filePath
	}
}
