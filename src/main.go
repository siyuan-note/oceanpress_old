package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	copy "github.com/otiai10/copy"
	conf "github.com/siyuan-note/oceanpress/src/conf"
	"github.com/siyuan-note/oceanpress/src/historical_data_processing"
	oceanpress "github.com/siyuan-note/oceanpress/src/render"
	"github.com/siyuan-note/oceanpress/src/sqlite"
	store "github.com/siyuan-note/oceanpress/src/store"
	structAll "github.com/siyuan-note/oceanpress/src/struct"
	"github.com/siyuan-note/oceanpress/src/template"
	"github.com/siyuan-note/oceanpress/src/util"
)

func main() {
	util.RunningLog("0", "=== 🛬 开始转换 🛫 ===")
	// 流程 1  用户输入 {源目录 输出目录}
	util.RunningLog("1", "用户输入")
	sourceDir := conf.SourceDir
	outDir := conf.OutDir
	workspaceDir := path.Join(filepath.ToSlash(conf.SourceDir), "../")
	util.RunningLog("1.1", "sourceDir:"+sourceDir)
	util.RunningLog("1.2", "outDir:"+outDir)
	util.RunningLog("1.3", "viewsDir:"+conf.TemplateDir)
	util.RunningLog("1.4", "SqlitePath:"+conf.SqlitePath)
	tempDbPath := path.Join(filepath.ToSlash(sourceDir), "../oceanPressTemp.db")
	err := copy.Copy(conf.SqlitePath, tempDbPath)
	if err != nil {
		util.DevLog("copy 数据库失败", err)
	}
	conf.SqlitePath = tempDbPath

	// 流程 2  copy 源目录中资源文件至输出目录
	util.RunningLog("2", "copy 资源到 outDir")
	// copy views 中的资源文件
	copy.Copy(path.Join(conf.TemplateDir, "./assets"), path.Join(outDir, "./assets"))
	copy.Copy(path.Join(workspaceDir, "./widgets"), path.Join(outDir, "./assets/widgets"))
	util.RunningLog("2.1", "copy 完成")
	util.RunningLog("2.2", "copy widgets")

	// 流程 3  遍历源目录 生成 html 到输出目录
	util.RunningLog("3", "生成 html")

	// 转换数据结构 filepath => entityList
	util.RunningLog("3.1", "收集转换生成所需数据")
	structAll.NoteStore = store.DirToStruct(
		sourceDir,
		conf.SqlitePath,
		template.TemplateRender,
		func(db sqlite.DbResult, FindFileEntityFromID structAll.FindFileEntityFromID, structToHTML func(interface{}) string) func(entity structAll.FileEntity) (html string, xml string) {
			// luteEngine lute 实例
			var luteEngine = lute.New()

			/** 对引用块进行渲染 */
			luteEngine.SetBlockRef(true)
			// /** 渲染 id （渲染为空） */
			luteEngine.SetKramdownIAL(true)
			// /** 标题的链接 a 标签渲染 */
			luteEngine.SetHeadingAnchor(true)
			luteEngine.SetKramdownIALIDRenderName("data-n-id")
			luteEngine.SetKramdownSpanIAL(true)

			// FileEntityToHTML entity 转 html
			FileEntityToHTML := func(entity structAll.FileEntity) (html string, xml string) {
				context := oceanpress.Context{}
				context.Db = db
				context.BaseEntity = entity
				context.FindFileEntityFromID = FindFileEntityFromID
				context.LuteEngine = luteEngine
				context.StructToHTML = structToHTML
				// 当(*oceanpress.Options)之后的代码报错的时候，请按照查阅 [0dev](./render/0dev.md) 同步相关代码
				renderer := oceanpress.NewOceanPressRenderer(entity.Tree, (*oceanpress.Options)(luteEngine.RenderOptions), &context)

				return renderer.Render()
			}
			return FileEntityToHTML
		})

	util.RunningLog("3.2", "复制资源文件")

	for _, entity := range structAll.NoteStore.StructList {
		if entity.Tree == nil {
			// 目录
		} else {
			if conf.IsDev {
				// 开发模式下跳过资源的 copy
				HandlingAssets(entity.Tree.Root, outDir, entity)
			} else {
				HandlingAssets(entity.Tree.Root, outDir, entity)
			}
		}
	}

	util.RunningLog("3.3", "从文件到数据结构转换完毕，开始生成html,共", len(structAll.NoteStore.StructList), "项")
	historical_data_processing.GenerateRedirectFile()
	for _, entity := range structAll.NoteStore.StructList {
		info := entity.Info
		virtualPath := entity.VirtualPath()

		LevelRoot := entity.RootPath()

		if info.IsDir() {
			if conf.IsDev {
				continue
			}
			// 这里要生成一个类似于当前目录菜单的东西
			targetPath := filepath.Join(outDir, virtualPath, "index.html")
			// 当前目录的 子路径 不包含更深层级的
			sonList := fileEntityListFilter(structAll.NoteStore.StructList, func(f structAll.FileEntity) bool {
				return strings.HasPrefix(f.VirtualPath(), virtualPath) &&
					// 这个条件去除了间隔一层以上的其他路径
					strings.LastIndex(f.VirtualPath()[len(virtualPath):], "/") == 0
			})

			var sonEntityList []template.SonEntityI
			for _, sonEntity := range sonList {
				if conf.RssNoOutputHtml && strings.HasSuffix(sonEntity.Name, ".rss.xml") {
					continue
				}
				webPath := sonEntity.VirtualPath()[len(virtualPath):]
				var name string
				if sonEntity.Info.IsDir() {
					name = webPath + "/"
					webPath += "/index.html"
				} else {
					name = sonEntity.Name
				}

				sonEntityList = append(sonEntityList, template.SonEntityI{
					WebPath: webPath,
					Name:    name,
					IsDir:   sonEntity.Info.IsDir(),
				})
			}
			var menuInfo = (template.MenuInfo{
				SonEntityList: sonEntityList,
				PageTitle:     "菜单页",
				LevelRoot:     LevelRoot,
			})
			html := menuInfo.Render()
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				os.MkdirAll(filepath.Join(outDir, virtualPath), 0700) // Create your file
			}
			ioutil.WriteFile(targetPath, []byte(html), 0777)
		} else {
			startT := time.Now()
			targetPath := filepath.Join(outDir, virtualPath)

			rawHTML, xml := entity.Output()
			if len(rawHTML) != 0 {
				html := template.ArticleRender(template.ArticleInfo{
					Content:   template.HTML(rawHTML),
					PageTitle: entity.Name,
					LevelRoot: LevelRoot,
				})
				if strings.HasSuffix(entity.Name, ".rss.xml") == false || conf.RssNoOutputHtml {
					var err = ioutil.WriteFile(targetPath, []byte(html), 0777)
					if err != nil {
						util.Log(err)
					}
				}
			}
			if len(xml) != 0 {
				targetPath := filepath.Join(outDir, virtualPath[0:len(virtualPath)-5])
				var err = ioutil.WriteFile(targetPath, []byte(xml), 0777)
				if err != nil {
					util.Log(err)
				}
			}
			tc := time.Since(startT)
			// 大于 x00 ms 的
			if tc > 1000_000_000 {
				util.DevLog("渲染耗时高", tc, targetPath)
			}

		}
	}
	// End
	util.Log("----- End -----")

}

// go 怎么写类似于其他语言泛型的过滤方式 ？// https://medium.com/@habibridho/here-is-why-no-one-write-generic-slice-filter-in-go-8b3d1063674e
func fileEntityListFilter(list []structAll.FileEntity, test func(structAll.FileEntity) bool) (ret []structAll.FileEntity) {
	for _, s := range list {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

// TODO: 逻辑不够严谨，还是存在错误 例如 file:///D:/code/doc/docHTML/工具/文本处理/简单文本处理.html 中的资源地址无法访问
func HandlingAssets(node *ast.Node, outDir string, fileEntity structAll.FileEntity) {
	if node.Next != nil {
		HandlingAssets(node.Next, outDir, fileEntity)
	}
	if node.FirstChild != nil {
		HandlingAssets(node.FirstChild, outDir, fileEntity)
	}
	for _, n := range node.Children {
		HandlingAssets(n, outDir, fileEntity)
	}

	if node != nil && node.Type == ast.NodeLinkDest {
		dest := node.TokensStr()
		decodeUrl, err := url.QueryUnescape(dest)
		if err == nil {
			dest = decodeUrl
		}

		if strings.HasPrefix(filepath.ToSlash(dest), "assets/") {
			// 笔记本中的资源目录
			sourceDir := filepath.ToSlash(conf.SourceDir)
			workspaceDir := path.Join(sourceDir, "../")

			level := 0

			for {
				assetsPath := path.Join(filepath.ToSlash(filepath.Dir(fileEntity.Path)), strings.Repeat("../", level), dest)
				matched, _ := filepath.Rel(workspaceDir+"*", assetsPath)
				// 判断资源文件是否处于工作空间内，超出工作空间的不处理
				if len(matched) > 0 {
					_, err := os.Stat(assetsPath)
					if err == nil {
						matched, _ := filepath.Match(sourceDir+"*", assetsPath)
						if matched {
							// 资源文件在笔记本内，这里重写链接地址即可
							var p string
							if level == 0 {
								p = path.Join(fileEntity.RootPath(), dest)
							} else {
								p = path.Join(strings.Repeat("../", level), dest)
							}
							node.Tokens = []byte(p)
						} else {
							// 在工作空间内
							node.Tokens = []byte(path.Join(fileEntity.RootPath(), dest))
						}
						// TODO 因为会有多个链接指向同一个资源，所以下面的写法会导致多余的 copy 需要优化
						err := copy.Copy(assetsPath, path.Join(outDir, dest))
						if err != nil {
							util.Warn("复制资源文件失败", err)
						}
						return
					} else {
						// 当前路径不存在
						level += 1
						continue
					}
				} else {
					// 超出了工作空间范围
					util.Warn("没有在工作空间内找到资源文件", err)
					break
				}
			}
		}
	}
}
