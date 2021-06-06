package render

import (
	"bytes"
	"errors"
	"html/template"
	"path"
	"strings"

	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/html"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/render"
	luteUtil "github.com/88250/lute/util"
	structAll "github.com/siyuan-note/oceanpress/src/struct"
	"github.com/siyuan-note/oceanpress/src/util"
)

func (r *OceanPressRender) Render() (output []byte) {
	output = r.BaseRenderer.Render()
	output = append(output, r.RenderFootnotes()...)

	var refHTML string
	curID := r.context.BaseEntity.Tree.ID

	sql := `SELECT
	"refs".block_id AS "id"
FROM
	"refs"
	LEFT JOIN blocks ON "refs".block_id = blocks.id
WHERE
	def_block_id = /** 被引用块的 id */
	'` + curID + `'
	AND /** 当前文档内对当前文档的引用不显示在反链中 */
	"blocks".root_id != '` + curID + `';`
	// 底部反链
	content := r.SqlRender(sql, false, true)
	if len(content) > 0 {
		// TODO: 这里也应该使用模板，容后再做
		refHTML = `<h2>链接到此文档的相关文档</h2>` + content
	}
	output = append(output, []byte(refHTML)...)
	return
}

// renderNodeToHTML 将指定节点渲染为 html
func (r *OceanPressRender) renderNodeToHTML(node *ast.Node, headerIncludes bool) string {
	// 收集块
	var nodes []*ast.Node
	ast.Walk(node, func(n *ast.Node, entering bool) ast.WalkStatus {
		if entering {
			nodes = append(nodes, n)
			if ast.NodeHeading == node.Type && headerIncludes {
				// 支持“标题块”引用
				children := headingChildren(n)
				nodes = append(nodes, children...)
			}
		}
		return ast.WalkSkipChildren
	})

	// 渲染块
	root := &ast.Node{Type: ast.NodeDocument}
	luteEngine := lute.New()
	tree := &parse.Tree{Root: root, Context: &parse.Context{ParseOption: luteEngine.ParseOptions}}
	tree.Context.ParseOption.KramdownBlockIAL = false // 关闭 IAL

	if node.ID != "" {
		err := r.context.push(node.ID)
		if err != nil {
			return "oceanpress 渲染错误：「循环引用」"
		} else {
			defer r.context.pop(node.ID)
		}
	}

	renderer := NewOceanPressRenderer(tree, (*Options)(luteEngine.RenderOptions), r.context)
	// renderer2 := render.NewFormatRenderer(tree, luteEngine.RenderOptions)
	renderer.Writer = &bytes.Buffer{}
	// renderer.NodeWriterStack = append(renderer.NodeWriterStack, renderer.Writer) // 因为有可能不是从 root 开始渲染，所以需要初始化
	for _, node := range nodes {
		ast.Walk(node, func(n *ast.Node, entering bool) ast.WalkStatus {
			rendererFunc := renderer.RendererFuncs[n.Type]
			return rendererFunc(n, entering)
		})
	}
	html := strings.TrimSpace(renderer.Writer.String())
	return html
}

// renderBlockRef 块引用渲染,类似于超链接
func (r *OceanPressRender) renderBlockRef(node *ast.Node, entering bool) ast.WalkStatus {
	if entering == false {
		return ast.WalkContinue
	}
	var refID string
	root := getRootByNode(node)
	currentEntity, _, _ := r.FindFileEntityFromID(root.ID)

	var targetNodeStructInfo structAll.StructInfo
	var targetEntity structAll.FileEntity

	var src string
	var title string
	var findErr error = nil
	hasEmbedText := false
	children := getAllNextByNode(node.FirstChild)

	for _, n := range children {
		if n.Type == ast.NodeBlockRefID {
			// 这里应该每个 NodeBlockRef 都包含了，意味着一般一定执行
			refID = n.TokensStr()
			targetEntity, targetNodeStructInfo, findErr = r.FindFileEntityFromID(refID)
			if targetEntity.Path != "" {
				src = currentEntity.FileEntityRelativePath(targetEntity, refID)
			}
		}
		if n.Type == ast.NodeBlockRefText {
			// NodeBlockRef 内不一定有 NodeBlockRefText
			hasEmbedText = true
			title = n.TokensStr()
		}
	}

	if hasEmbedText == false {
		name, _, err := FindAttr(targetNodeStructInfo.Node.KramdownIAL, "name")
		// 对于命名块优先渲染他的名字 而非内容
		if err == nil {
			title = name
		} else if targetNodeStructInfo.Node != nil {
			// 渲染引用块的内容文本
			if targetNodeStructInfo.Node.Type == ast.NodeDocument {
				title = targetEntity.Name
			} else {
				html := r.renderNodeToHTML(targetNodeStructInfo.Node, false)
				title = r.HTML2Text(html)
			}
		}
	}
	title = strings.ReplaceAll(title, "\n", "")
	// findErr 本身已经会发出警告了
	if findErr == nil && strings.TrimSpace(title) == "" {
		util.Warn("<块引用渲染为空>", r.context.BaseEntity.RelativePath+" 中的块引用 "+refID)
	}
	r.WriteString(r.context.StructToHTML(structAll.BlockRefInfo{
		Src:   src,
		Title: title,
	}))

	return ast.WalkSkipChildren
}

// renderBlockQueryEmbed 嵌入块渲染
func (r *OceanPressRender) renderBlockQueryEmbed(node *ast.Node, entering bool) ast.WalkStatus {
	if entering == false {
		return ast.WalkContinue
	}
	var sql string
	var html string
	for _, n := range getAllNextByNode(node.FirstChild) {
		if n.Type == ast.NodeBlockQueryEmbedScript {
			// 这里应该每个 NodeBlockQueryEmbed 都包含了，意味着期望一定执行
			sql = n.TokensStr()
			html = r.SqlRender(sql, true, false)
		}
	}
	r.WriteString(html)
	return ast.WalkSkipChildren
}

// NodeSuperBlock 超级块渲染
func (r *OceanPressRender) renderSuperBlock(node *ast.Node, entering bool) ast.WalkStatus {
	if entering == false {
		return ast.WalkContinue
	}
	var layout string
	var html string
	children := getAllNextByNode(node.FirstChild)
	for _, n := range children {
		if n.Type == ast.NodeSuperBlockLayoutMarker {
			layout = n.TokensStr()
		}
	}
	for _, n := range children {
		html += r.renderNodeToHTML(n, false)
	}
	r.Newline()
	node.KramdownIAL = append(node.KramdownIAL, []string{"data-sb-layout", layout})
	r.Tag("div", node.KramdownIAL, false)
	r.WriteString(html)
	r.Tag("/div", nil, false)
	return ast.WalkSkipChildren
}

// renderHeading h-number 标题块渲染
func (r *OceanPressRender) renderHeading(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		level := headingLevel[node.HeadingLevel : node.HeadingLevel+1]
		idStr := HeadingID(node)
		var attrs [][]string
		if r.Options.ToC || r.Options.HeadingID || r.Options.KramdownBlockIAL {
			attrs = append(attrs, []string{"id", idStr})
			if r.Options.KramdownBlockIAL {
				if 1 < len(node.KramdownIAL) {
					attrs = append(attrs, node.KramdownIAL[1:]...)
				}
			}
		}
		r.Newline()
		r.Tag("h"+level, attrs, false)
	} else {
		if r.Options.HeadingAnchor {
			id := HeadingID(node)
			r.Tag("a", [][]string{{"id", "vditorAnchor-" + id}, {"class", "vditor-anchor"}, {"href", "#" + id}}, false)
			// r.WriteString(`<svg viewBox="0 0 16 16" version="1.1" width="16" height="16"><path fill-rule="evenodd" d="M4 9h1v1H4c-1.5 0-3-1.69-3-3.5S2.55 3 4 3h4c1.45 0 3 1.69 3 3.5 0 1.41-.91 2.72-2 3.25V8.59c.58-.45 1-1.27 1-2.09C10 5.22 8.98 4 8 4H4c-.98 0-2 1.22-2 2.5S3 9 4 9zm9-3h-1v1h1c1 0 2 1.22 2 2.5S13.98 12 13 12H9c-.98 0-2-1.22-2-2.5 0-.83.42-1.64 1-2.09V6.25c-1.09.53-2 1.84-2 3.25C6 11.31 7.55 13 9 13h4c1.45 0 3-1.69 3-3.5S14.5 6 13 6z"></path></svg>`)
			r.Tag("img", [][]string{{"src", r.context.BaseEntity.RootPath() + "assets/icon/alink.png"}}, true)
			r.Tag("/a", nil, false)
		}
		r.Tag("/h"+headingLevel[node.HeadingLevel:node.HeadingLevel+1], nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

// 代码块渲染 这里定制的目的是为了 附加 mindmap 解析后的数据
func (r *OceanPressRender) renderCodeBlock(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	noHighlight := false
	var language string
	if nil != node.FirstChild.Next && 0 < len(node.FirstChild.Next.CodeBlockInfo) {
		language = luteUtil.BytesToStr(node.FirstChild.Next.CodeBlockInfo)
		noHighlight = r.NoHighlight(language)
	}

	if entering {
		if noHighlight {
			var attrs [][]string
			tokens := html.EscapeHTML(node.FirstChild.Next.Next.Tokens)
			tokens = bytes.ReplaceAll(tokens, luteUtil.CaretTokens, nil)
			tokens = bytes.TrimSpace(tokens)

			content := luteUtil.BytesToStr(tokens)
			attrs = append(attrs, []string{"data-content", content})
			if language == "mindmap" { // 图标数据 parser ，protyle 是引入了 lute 来做这个，我打算在编译的时候 parse
				eChartsData := html.EscapeString(render.EChartsMindmapStr(content))
				attrs = append(attrs, []string{"data-parse-content", eChartsData})
			}
			attrs = append(attrs, []string{"data-subtype", language})

			r.Tag("div", attrs, false)
			r.Tag("div", [][]string{{"spin", "1"}}, false)
			r.Tag("/div", nil, false)
			r.Tag("/div", nil, false)
			return ast.WalkSkipChildren
		}

		attrs := [][]string{{"class", "code-block"}, {"data-language", language}}
		r.Tag("pre", attrs, false)
		r.WriteString("<code>")
	} else {
		if noHighlight {
			return ast.WalkSkipChildren
		}

		r.Tag("/code", nil, false)
		r.Tag("/pre", nil, false)
	}
	return ast.WalkContinue
}

// 文档根节点
func (r *OceanPressRender) renderDocument(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("main", node.KramdownIAL, false)
	} else {
		r.Tag("/main", nil, false)
	}
	return ast.WalkContinue
}

// ========= 附加在 OceanPressRender 上的工具方法
func (r *OceanPressRender) Tag(name string, attrs [][]string, selfclosing bool) {
	if r.DisableTags > 0 {
		return
	}
	id, idIndex, _ := FindAttr(attrs, "id")
	nId, nIdIndex, _ := FindAttr(attrs, "data-n-id")

	var attrsTemp [][]string
	if idIndex != nIdIndex && id == nId && id != "" {
		for i, v := range attrs {
			if i == idIndex {
				// 过滤掉这一项
			} else {
				attrsTemp = append(attrsTemp, v)
			}
		}
		attrs = attrsTemp
	}

	r.WriteString("<")
	r.WriteString(name)
	if 0 < len(attrs) {
		for _, attr := range attrs {
			attrName := attr[0]
			attrValue := attr[1]
			r.WriteString(" " + attrName + "=\"" + attrValue + "\"")
		}
	}
	if selfclosing {
		r.WriteString(" /")
	}
	r.WriteString(">")
}

// SqlRender 通过 sql 渲染出 html
func (r *OceanPressRender) SqlRender(sql string, headerIncludes bool, removeDuplicate bool) string {
	sql = util.HTMLEntityDecoder(sql)
	ids := r.context.Db.SQLToID(sql)
	if removeDuplicate {
		var ret []string
		retIncludes := func(id string) bool {
			count := 0
			for _, id2 := range ret {
				if id == id2 {
					count += 1
				}
			}
			return count > 0
		}
		for _, id := range ids {
			if retIncludes(id) == false {
				ret = append(ret, id)
			}
		}
		ids = ret
	}

	var html string
	for _, id := range ids {
		fileEntity, mdInfo, err := r.FindFileEntityFromID(id)
		if err != nil {
			continue
		}
		// err = r.context.push(mdInfo.BlockID)
		if err != nil {
			// TODO: 这里应该要处理成点击展开，或者换一个更好的显示
			html = "error: 循环引用 "
		} else {
			var src string
			if fileEntity.Path != "" {
				src = r.context.BaseEntity.FileEntityRelativePath(fileEntity, id)
			}
			// 修改 base 路径以使用 ../ 这样的形式指向根目录再深入到待解析的md文档所在的路径 ,就在下面一点点会再重置回去
			r.context.LuteEngine.RenderOptions.LinkBase = strings.Repeat("../", strings.Count(r.context.BaseEntity.RelativePath, "/")-1) + "." + path.Dir(fileEntity.RelativePath)
			html += r.context.StructToHTML(structAll.EmbeddedBlockInfo{
				Src:     src,
				Title:   src,
				Content: template.HTML(r.renderNodeToHTML(mdInfo.Node, headerIncludes)),
			})
			r.context.LuteEngine.RenderOptions.LinkBase = ""
			// r.context.pop(mdInfo.BlockID)
		}

	}
	return html
}

// FindFileEntityFromID 附加了人性化警告
func (r *OceanPressRender) FindFileEntityFromID(id string) (structAll.FileEntity, structAll.StructInfo, error) {
	a, b, err := r.context.FindFileEntityFromID(id)
	if err != nil {
		util.Warn("<没有找到对应块>", r.context.BaseEntity.Name+"("+r.context.BaseEntity.RelativePath+") 引用了 "+id+" 但没有找到该块")
	}
	return a, b, err
}
func (r *OceanPressRender) HTML2Text(dom string) string {
	lute := r.context.LuteEngine
	tree := lute.HTML2Tree(dom)
	if nil == tree {
		return ""
	}
	buf := &bytes.Buffer{}

	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		switch n.Type {
		case ast.NodeText, ast.NodeLinkText, ast.NodeBlockRefText, ast.NodeBlockEmbedText, ast.NodeFootnotesRef, ast.NodeCodeSpanContent:
			buf.Write(n.Tokens)
		}
		return ast.WalkContinue
	})
	return buf.String()
}

// =========

func headingChildren(heading *ast.Node) (ret []*ast.Node) {
	currentLevel := heading.HeadingLevel
	var blocks []*ast.Node
	for n := heading.Next; nil != n; n = n.Next {
		if ast.NodeHeading == n.Type {
			if currentLevel >= n.HeadingLevel {
				break
			}
		}
		blocks = append(blocks, n)
	}
	ret = append(ret, blocks...)
	return
}
func getRootByNode(node *ast.Node) *ast.Node {
	if node.Parent == nil {
		return node
	} else {
		return getRootByNode(node.Parent)
	}
}

// getAllNextByNode 获取一个节点的 所有后续节点
func getAllNextByNode(node *ast.Node) []*ast.Node {
	var list []*ast.Node
	// if node == nil {
	// 	return list
	// } else {
	// 	return append(list, getAllNextByNode(node.Next)...)
	// }
	c := node
	for {
		if c == nil {
			break
		} else {
			list = append(list, c)
			c = c.Next
		}
	}
	return list
}
func FindAttr(attrs [][]string, name string) (string, int, error) {
	for i, kv := range attrs {
		if name == kv[0] {
			return kv[1], i, nil
		}
	}
	return "", 0, errors.New("没有找到对应的 attr")
}
