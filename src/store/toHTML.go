package store

import (
	"bytes"
	"errors"
	"html/template"
	"path"
	"sort"
	"strings"

	sqlite "github.com/2234839/md2website/src/sqlite"
	"github.com/2234839/md2website/src/util"
	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/html"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/render"
	luteUtil "github.com/88250/lute/util"
)

// EmbeddedBlockInfo 嵌入块所需信息
type EmbeddedBlockInfo struct {
	AEmbeddedBlockInfo int
	Title              interface{}
	Src                string
	Content            interface{}
}

// BlockRefInfo 块引用所需信息
type BlockRefInfo struct {
	ABlockRefInfo int
	Title         interface{}
	Src           string
}

// Generate TODO: 这里的逻辑有点混乱，需要整理
func Generate(db sqlite.DbResult, FindFileEntityFromID FindFileEntityFromID, structToHTML func(interface{}) string) func(entity FileEntity) string {
	// luteEngine lute 实例
	var luteEngine = lute.New()

	/** 当前被处理的 entity */
	var baseEntity FileEntity
	/** 对引用块进行渲染 */
	luteEngine.SetBlockRef(true)
	// /** 渲染 id （渲染为空） */
	luteEngine.SetKramdownIAL(true)
	// /** 标题的链接 a 标签渲染 */
	luteEngine.SetHeadingAnchor(true)
	luteEngine.SetKramdownIALIDRenderName("data-block-id")

	// FileEntityToHTML entity 转 html
	FileEntityToHTML := func(entity FileEntity) string {
		renderer, r := NewOceanpressRenderer(entity.Tree, luteEngine.RenderOptions, db, FindFileEntityFromID, structToHTML, baseEntity, luteEngine)

		r.context.baseEntity = entity

		// 在每个文档的底部显示反链
		curID := entity.Tree.ID
		var refHTML string
		content := r.SqlRender(`SELECT "refs".block_id as "ref_id", blocks.* FROM "refs"

	LEFT JOIN blocks
	ON "refs".block_id = blocks.id

	WHERE
	def_block_id = /** 被引用块的 id */ '`+curID+`';`, false)
		if len(content) > 0 {
			// TODO: 这里也应该使用模板，容后再做
			refHTML = `<h2>链接到此文档的相关文档</h2>` + content
		}
		output := renderer.Render()
		html := string(output)
		return html + refHTML
	}
	return FileEntityToHTML
}

type OceanpressRenderer struct {
	*render.BaseRenderer
	context Context
}

func (r *OceanpressRenderer) NodeDocument(node *ast.Node, entering bool) ast.WalkStatus {
	return r.GeneterateRenderFunction(func(n *ast.Node, entering bool, src string, fileEntity FileEntity, mdInfo StructInfo, html string) string {
		r.context.refID = n.TokensStr()
		if n.ID == "20210325155155-2wk7rxv" {
			util.Log("debugger")
		}
		var dataString string
		for _, item := range n.KramdownIAL {
			name := item[0]
			value := item[1]
			dataString += "data-block-" + name + "=\"" + value + "\" "
		}
		if entering {
			return "<main " + dataString + ">"

		}
		return "</main>"
	})(node, entering)
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

/** 块引用渲染,类似于超链接 */
func (r *OceanpressRenderer) NodeBlockRef(node *ast.Node, entering bool) ast.WalkStatus {
	if entering == false {
		return ast.WalkContinue
	}
	var refID string
	root := getRootByNode(node)
	currentEntity, _, _ := r.context.FindFileEntityFromID(root.ID)

	var targetNodeStructInfo StructInfo
	var targetEntity FileEntity

	var src string
	var title string
	hasEmbedText := false
	children := getAllNextByNode(node.FirstChild)
	for _, n := range children {
		if n.Type == ast.NodeBlockRefID {
			// 这里应该每个 NodeBlockRef 都包含了，意味着一般一定执行
			refID = n.TokensStr()

			targetEntity, targetNodeStructInfo, _ = r.context.FindFileEntityFromID(refID)
			if targetEntity.Path != "" {
				src = FileEntityRelativePath(currentEntity, targetEntity, refID)
			}
		}
		if n.Type == ast.NodeBlockRefText {
			// NodeBlockRef 内不一定有 NodeBlockRefText
			hasEmbedText = true
			title = n.TokensStr()
		}
	}

	if hasEmbedText == false && targetNodeStructInfo.node != nil {
		if targetNodeStructInfo.node.Type == ast.NodeDocument {
			title = targetEntity.Name
		} else {
			title = r.context.luteEngine.HTML2Text(r.renderNodeToHTML(targetNodeStructInfo.node, false))
		}
	}
	r.WriteHTML(r.context.structToHTML(BlockRefInfo{
		Src:   src,
		Title: title,
	}))
	return ast.WalkSkipChildren
}

/** 嵌入块渲染 */
func (r *OceanpressRenderer) NodeBlockQueryEmbed(node *ast.Node, entering bool) ast.WalkStatus {
	if entering == false {
		return ast.WalkContinue
	}
	var sql string
	var html string
	for _, n := range getAllNextByNode(node.FirstChild) {
		if n.Type == ast.NodeBlockQueryEmbedScript {
			// 这里应该每个 NodeBlockQueryEmbed 都包含了，意味着期望一定执行
			sql = n.TokensStr()
			html = r.SqlRender(sql, true)
		}
	}
	r.WriteHTML(html)
	return ast.WalkSkipChildren
}

/** 超级块渲染 */
func (r *OceanpressRenderer) NodeSuperBlock(node *ast.Node, entering bool) ast.WalkStatus {
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
	sort.Slice(children, func(i, j int) bool {
		n := children[i]
		return n.Type != ast.NodeSuperBlockCloseMarker && n.Type != ast.NodeSuperBlockOpenMarker
	})
	for _, n := range children {
		html += r.renderNodeToHTML(n, true)
	}
	// TODO： 采用模板
	html = "<div data-type=\"NodeSuperBlock\" data-sb-layout=\"" + layout + "\" >" + html + "</div>"
	r.WriteHTML(html)
	return ast.WalkSkipChildren
}

// 代码块渲染
func (r *OceanpressRenderer) NodeCodeBlock(node *ast.Node, entering bool) ast.WalkStatus {
	r.context.rawRenderer.Newline()
	noHighlight := false
	var language string
	if nil != node.FirstChild.Next && 0 < len(node.FirstChild.Next.CodeBlockInfo) {
		language = luteUtil.BytesToStr(node.FirstChild.Next.CodeBlockInfo)
		noHighlight = r.context.rawRenderer.NoHighlight(language)
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

			r.context.rawRenderer.Tag("div", attrs, false)
			r.context.rawRenderer.Tag("div", [][]string{{"spin", "1"}}, false)
			r.context.rawRenderer.Tag("/div", nil, false)
			r.context.rawRenderer.Tag("/div", nil, false)
			return ast.WalkSkipChildren
		}

		attrs := [][]string{{"class", "code-block"}, {"data-language", language}}
		r.context.rawRenderer.Tag("pre", attrs, false)
		r.context.rawRenderer.WriteString("<code>")
	} else {
		if noHighlight {
			return ast.WalkSkipChildren
		}

		r.context.rawRenderer.Tag("/code", nil, false)
		r.context.rawRenderer.Tag("/pre", nil, false)
	}
	return ast.WalkContinue
}

// SqlRender 通过 sql 渲染出 html
func (r *OceanpressRenderer) SqlRender(sql string, headerIncludes bool) string {
	sql = util.HTMLEntityDecoder(sql)
	ids := r.context.db.SQLToID(sql)
	var html string
	for _, id := range ids {
		fileEntity, mdInfo, err := r.context.FindFileEntityFromID(id)
		if err != nil {
			return ""
		}
		err = r.context.push(mdInfo.blockID)
		if err != nil {
			// TODO: 这里应该要处理成点击展开，或者换一个更好的显示
			html = "error: 循环引用 "
		} else {
			var src string
			if fileEntity.Path != "" {
				src = FileEntityRelativePath(r.context.baseEntity, fileEntity, id)
			}
			// 修改 base 路径以使用 ../ 这样的形式指向根目录再深入到待解析的md文档所在的路径 ,就在下面一点点会再重置回去
			r.context.luteEngine.RenderOptions.LinkBase = strings.Repeat("../", strings.Count(r.context.baseEntity.RelativePath, "/")-1) + "." + path.Dir(fileEntity.RelativePath)
			html += r.context.structToHTML(EmbeddedBlockInfo{
				Src:     src,
				Title:   src,
				Content: template.HTML(r.renderNodeToHTML(mdInfo.node, headerIncludes)),
			})
			r.context.luteEngine.RenderOptions.LinkBase = ""
			r.context.pop(mdInfo.blockID)
		}

	}
	return html
}

func (r *OceanpressRenderer) 模板复制粘贴用(node *ast.Node, entering bool) ast.WalkStatus {
	return r.GeneterateRenderFunction(func(n *ast.Node, entering bool, src string, fileEntity FileEntity, mdInfo StructInfo, html string) string {
		return ""
	})(node, entering)
}

func (r *OceanpressRenderer) titleRenderer(n *ast.Node, entering bool, src string, fileEntity FileEntity, mdInfo StructInfo, html string) template.HTML {
	var title = template.HTML(n.Text())
	t := string(title)

	// 锚文本模板变量处理 使用定义块内容文本填充。
	if strings.Contains(t, "{{.text}}") {
		var title2 template.HTML
		// 如定义块是文档块，则使用文档名填充。
		if mdInfo.blockType == "NodeDocument" {
			title2 = template.HTML(fileEntity.Name)
		} else {
			title2 = template.HTML(
				r.context.luteEngine.HTML2Text(
					r.context.luteEngine.MarkdownStr("", renderNodeMarkdown(mdInfo.node, false)),
				),
			)
		}
		title = template.HTML(strings.ReplaceAll(t, "{{.text}}", string(title2)))
	}
	title = template.HTML(strings.ReplaceAll(string(title), "\n", ""))
	return title
}

// HOC, 内部处理了循环引用的问题， 生成一个渲染函数，
func (r *OceanpressRenderer) GeneterateRenderFunction(render func(n *ast.Node, entering bool, src string, fileEntity FileEntity, mdInfo StructInfo, html string) string) func(n *ast.Node, entering bool) ast.WalkStatus {
	return func(n *ast.Node, entering bool) ast.WalkStatus {
		var html string
		if entering {
			fileEntity, mdInfo, err := r.context.FindFileEntityFromID(r.context.refID)
			if err != nil {
				return ast.WalkContinue
			}
			err = r.context.push(mdInfo.blockID)
			if err != nil {
				html = "error: 循环引用 "
			} else {
				var src string
				if fileEntity.Path != "" {
					src = FileEntityRelativePath(r.context.baseEntity, fileEntity, r.context.refID)
				}
				// 修改 base 路径以使用 ../ 这样的形式指向根目录再深入到待解析的md文档所在的路径 ,就在下面一点点会再重置回去
				r.context.luteEngine.RenderOptions.LinkBase = strings.Repeat("../", strings.Count(r.context.baseEntity.RelativePath, "/")-1) + "." + path.Dir(fileEntity.RelativePath)
				html = render(n, entering, src, fileEntity, mdInfo, "")
				r.context.luteEngine.RenderOptions.LinkBase = ""
				r.context.pop(mdInfo.blockID)
			}

		}
		r.WriteHTML(html)
		return ast.WalkSkipChildren
	}
}

// ===== 下面是一些工具函数

// FileEntityRelativePath 计算他们变成 html 文件之后的相对路径
func FileEntityRelativePath(base FileEntity, cur FileEntity, id string) string {
	// 减一是因为 路径开头必有 / 而这里只需要跳到这一层
	count := strings.Count(base.RelativePath, "/")
	if strings.HasPrefix(base.RelativePath, "/") {
		count--
	}
	l2 := strings.Split(cur.RelativePath, "/")
	url := strings.Repeat("../", count)
	url += strings.Join(l2[1:], "/")
	url = FilePathToWebPath(url)
	url += "#" + id
	return url
}

// renderNodeToHTML 将指定节点渲染为 html
func (r *OceanpressRenderer) renderNodeToHTML(node *ast.Node, headerIncludes bool) string {
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

	renderer, _ := r.forkNewOceanpressRenderer(tree)
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

// 将 Node 渲染为 md 对于 header 节点特殊处理，会将他的 child 包含进来
func renderNodeMarkdown(node *ast.Node, headerIncludes bool) string {
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
	// tree.Context.ParseOption.KramdownIAL = false // 关闭 IAL
	tree.Context.ParseOption.KramdownBlockIAL = false // 关闭 IAL
	renderer := render.NewFormatRenderer(tree, luteEngine.RenderOptions)
	renderer.Writer = &bytes.Buffer{}
	renderer.NodeWriterStack = append(renderer.NodeWriterStack, renderer.Writer) // 因为有可能不是从 root 开始渲染，所以需要初始化
	for _, node := range nodes {
		ast.Walk(node, func(n *ast.Node, entering bool) ast.WalkStatus {
			rendererFunc := renderer.RendererFuncs[n.Type]
			return rendererFunc(n, entering)
		})
	}
	return strings.TrimSpace(renderer.Writer.String())
}

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

// 获取和 node 最相关的 block ID,从node 一直往他的父级找
func getNodeRelativeBlockID(node *ast.Node) string {
	curID := node.IALAttr("id")
	cursor := node
	for true {
		if curID != "" {
			break
		} else {
			if cursor.Parent != nil {
				cursor = cursor.Parent
				curID = cursor.IALAttr("id")
			} else {
				break
			}
		}
	}
	return curID
}

type Context struct {
	db                   sqlite.DbResult
	FindFileEntityFromID FindFileEntityFromID
	structToHTML         func(interface{}) string

	refID string
	push  func(id string) error
	pop   func(id string)

	baseEntity  FileEntity
	luteEngine  *lute.Lute
	rawRenderer *render.ProtylePreviewRenderer
}

func (r *OceanpressRenderer) WriteHTML(html string) {
	// r.WriteString(html) // 这里是没有用的
	// 因为实际上调用 rawRenderer 的渲染，r 只是提供一些渲染函数，所以实际要输出结果要调用 rawRenderer 的 WriteString
	r.context.rawRenderer.WriteString(html)
}

// forkNewOceanpressRenderer 从当前的 OceanpressRenderer 派生出一个新的，基于之前的配置
func (r *OceanpressRenderer) forkNewOceanpressRenderer(tree *parse.Tree) (*render.ProtylePreviewRenderer, *OceanpressRenderer) {
	r1, r2 := NewOceanpressRenderer(tree, r.context.luteEngine.RenderOptions, r.context.db, r.context.FindFileEntityFromID, r.context.structToHTML, r.context.baseEntity, r.context.luteEngine)
	return r1, r2
}
func NewOceanpressRenderer(tree *parse.Tree, options *render.Options,
	db sqlite.DbResult,
	FindFileEntityFromID FindFileEntityFromID,
	structToHTML func(interface{}) string,

	baseEntity FileEntity,
	luteEngine *lute.Lute,

) (*render.ProtylePreviewRenderer, *OceanpressRenderer) {

	// rawRenderer := render.NewHtmlRenderer(tree, options)
	rawRenderer := render.NewProtylePreviewRenderer(tree, options)
	// 嵌入块的 id
	var refID string
	var idStack []string

	push := func(id string) error {
		for _, item := range idStack {
			if item == id {
				util.Warn("循环引用", id)
				return errors.New("循环引用")
			}
		}
		idStack = append(idStack, id)
		return nil
	}
	pop := func(id string) {
		idStack = idStack[:len(idStack)-1]
	}
	var context = Context{
		db,
		FindFileEntityFromID,
		structToHTML,
		refID,
		push,
		pop,
		baseEntity,
		luteEngine,
		rawRenderer,
	}

	ret2 := &OceanpressRenderer{
		render.NewBaseRenderer(tree, options),
		context,
	}

	rawRenderer.RendererFuncs[ast.NodeBlockRef] = ret2.NodeBlockRef
	rawRenderer.RendererFuncs[ast.NodeBlockQueryEmbed] = ret2.NodeBlockQueryEmbed
	rawRenderer.RendererFuncs[ast.NodeSuperBlock] = ret2.NodeSuperBlock
	rawRenderer.RendererFuncs[ast.NodeCodeBlock] = ret2.NodeCodeBlock
	return rawRenderer, ret2
}
