package main

import (
	"strconv"

	"github.com/88250/lute"
	"github.com/88250/lute/ast"
)

// LuteEngine lute 实例
var LuteEngine = lute.New()

func init() {
	/** 对引用块进行渲染 */
	LuteEngine.SetBlockRef(true)
	// /** 渲染 id （渲染为空） */
	LuteEngine.SetKramdownIAL(true)
	// /** 标题的链接 a 标签渲染 */
	LuteEngine.SetHeadingAnchor(true)

	LuteEngine.Md2HTMLRendererFuncs[ast.NodeBlockRefText] = func(n *ast.Node, entering bool) (string, ast.WalkStatus) {
		var html string
		if entering {

			html = `<a href="" class="c-block-ref" data-block-type="` + strconv.Itoa(int(n.Type)) + `">` + n.Text() + `</a>`
		}
		return html, ast.WalkSkipChildren
	}
}
