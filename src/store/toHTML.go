package store

import (
	"bytes"
	"errors"
	"html/template"
	"path"
	"strings"

	sqlite "github.com/2234839/md2website/src/sqlite"
	"github.com/2234839/md2website/src/util"
	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/render"
)

// EmbeddedBlockInfo 嵌入块所需信息
type EmbeddedBlockInfo struct {
	AEmbeddedBlockInfo int
	Title              string
	Src                string
	Content            interface{}
}

// BlockRefInfo 块引用所需信息
type BlockRefInfo struct {
	ABlockRefInfo int
	Title         interface{}
	Src           string
}

// Generate
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

	// 嵌入块的 id
	var refID string
	/** 获取块的id */
	var getBlockID = func(n *ast.Node, _ bool) (string, ast.WalkStatus) {
		refID = n.TokensStr()
		return "", ast.WalkContinue
	}

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

	luteEngine.Md2HTMLRendererFuncs[ast.NodeBlockRefID] = getBlockID
	luteEngine.Md2HTMLRendererFuncs[ast.NodeBlockEmbedID] = getBlockID

	/** 块引用渲染,类似于超链接 */
	luteEngine.Md2HTMLRendererFuncs[ast.NodeBlockRefText] = func(n *ast.Node, entering bool) (string, ast.WalkStatus) {
		var html string
		if entering {
			fileEntity, mdInfo, err := FindFileEntityFromID(refID)
			if err != nil {
				return "", ast.WalkContinue
			}
			var src string
			if fileEntity.Path != "" {
				src = FileEntityRelativePath(baseEntity, fileEntity, refID)
			}
			var title = template.HTML(n.Text())
			// TODO 循环引用的处理应该抽离出去
			err = push(mdInfo.blockID)
			if err != nil {
				// TODO: 这里应该要处理成点击展开，或者换一个更好的显示
				html = "error:块引用 循环引用 "
			} else {
				t := strings.TrimSpace(string(title))
				// 锚文本模板变量处理 使用定义块内容文本填充。
				if t == "{{.text}}" {
					// 如定义块是文档块，则使用文档名填充。
					if mdInfo.blockType == "NodeDocument" {
						title = template.HTML(fileEntity.Name)
					} else {

						title = template.HTML(luteEngine.MarkdownStr("", renderNodeMarkdown(mdInfo.node, false)))

					}
				}

				// template.HTML(luteEngine.MarkdownStr("", renderNodeMarkdown(mdInfo.node)))
				html = structToHTML(BlockRefInfo{
					Src:   src,
					Title: title,
				})
				pop(mdInfo.blockID)
			}
		}
		return html, ast.WalkSkipChildren
	}
	/** 嵌入块渲染 */
	luteEngine.Md2HTMLRendererFuncs[ast.NodeBlockEmbedText] = func(n *ast.Node, entering bool) (string, ast.WalkStatus) {

		var html string
		if entering {
			fileEntity, mdInfo, err := FindFileEntityFromID(refID)
			if err != nil {
				return "", ast.WalkContinue
			}
			// TODO 循环引用的处理应该抽离出去
			err = push(mdInfo.blockID)
			if err != nil {
				// TODO: 这里应该要处理成点击展开，或者换一个更好的显示
				html = "error:嵌入块 循环引用 "
			} else {
				var src string
				if fileEntity.Path != "" {
					src = FileEntityRelativePath(baseEntity, fileEntity, refID)
				}
				// 修改 base 路径以使用 ../ 这样的形式指向根目录再深入到待解析的md文档所在的路径 ,就在下面一点点会再重置回去
				luteEngine.RenderOptions.LinkBase = strings.Repeat("../", strings.Count(baseEntity.RelativePath, "/")-1) + "." + path.Dir(fileEntity.RelativePath)
				html = structToHTML(EmbeddedBlockInfo{
					Src:   src,
					Title: n.Text(),
					// 这里涉及到一个套娃问题，还有 baseEntity 该怎么处理。以及他们的路径怎么办
					Content: template.HTML(luteEngine.MarkdownStr("", renderNodeMarkdown(mdInfo.node, true))),
				})
				luteEngine.RenderOptions.LinkBase = ""
				pop(mdInfo.blockID)
			}

		}
		return html, ast.WalkSkipChildren
	}

	/** 嵌入块查询渲染, 这个东西也应该包含在一个前端组件中 */
	luteEngine.Md2HTMLRendererFuncs[ast.NodeBlockQueryEmbedScript] = func(n *ast.Node, entering bool) (string, ast.WalkStatus) {
		var html string
		if entering {
			sql := n.TokensStr()
			ids := db.SQLToID(sql)

			curID := getNodeRelativeBlockID(n)

			for _, id := range ids {
				if id == curID {
					// 排除当前块，查询出自身并不方便阅读
					continue
				}

				fileEntity, mdInfo, err := FindFileEntityFromID(id)
				if err != nil {
					return "", ast.WalkContinue
				}
				err = push(mdInfo.blockID)
				if err != nil {
					// TODO: 这里应该要处理成点击展开，或者换一个更好的显示
					html = "error: 循环引用 "
				} else {
					var src string
					if fileEntity.Path != "" {
						src = FileEntityRelativePath(baseEntity, fileEntity, id)
					}
					// 修改 base 路径以使用 ../ 这样的形式指向根目录再深入到待解析的md文档所在的路径 ,就在下面一点点会再重置回去
					luteEngine.RenderOptions.LinkBase = strings.Repeat("../", strings.Count(baseEntity.RelativePath, "/")-1) + "." + path.Dir(fileEntity.RelativePath)
					html += structToHTML(EmbeddedBlockInfo{
						Src: src,
						// Title: sql, // 显示 sql 似乎不太好
						Title: src,
						// 这里涉及到一个套娃问题，还有 baseEntity 该怎么处理。以及他们的路径怎么办
						Content: template.HTML(luteEngine.MarkdownStr("", renderNodeMarkdown(mdInfo.node, true))),
					})
					luteEngine.RenderOptions.LinkBase = ""
					pop(mdInfo.blockID)
				}

			}

		}
		return html, ast.WalkSkipChildren
	}
	// FileEntityToHTML 转 html
	FileEntityToHTML := func(entity FileEntity) string {
		baseEntity = entity
		return luteEngine.MarkdownStr("", entity.MdStr)
	}
	return FileEntityToHTML
}

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
	tree.Context.ParseOption.KramdownIAL = false // 关闭 IAL
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
