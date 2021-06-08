package main

import (
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	copy "github.com/otiai10/copy"
	conf "github.com/siyuan-note/oceanpress/src/conf"
	oceanpress "github.com/siyuan-note/oceanpress/src/render"
	"github.com/siyuan-note/oceanpress/src/sqlite"
	store "github.com/siyuan-note/oceanpress/src/store"
	structAll "github.com/siyuan-note/oceanpress/src/struct"
	"github.com/siyuan-note/oceanpress/src/util"
)

func main() {
	util.RunningLog("0", "=== 🛬 开始转换 🛫 ===")
	// 流程 1  用户输入 {源目录 输出目录}
	util.RunningLog("1", "用户输入")
	sourceDir := conf.SourceDir
	outDir := conf.OutDir
	util.RunningLog("1.1", "sourceDir:"+sourceDir)
	util.RunningLog("1.2", "outDir:"+outDir)
	util.RunningLog("1.3", "viewsDir:"+conf.TemplateDir)
	util.RunningLog("1.4", "SqlitePath:"+conf.SqlitePath)
	util.RunningLog("1.5", "AssetsDir:"+conf.AssetsDir)

	// 流程 2  copy 源目录中资源文件至输出目录
	util.RunningLog("2", "copy 资源到 outDir")

	copy.Copy(sourceDir, outDir, copy.Options{
		// 跳过一些不必要的目录以及 md 文件
		Skip: func(src string) (bool, error) {
			return (util.IsSkipPath(src) || util.IsNotes(src)), nil
		},
	})
	// copy views 中的资源文件
	copy.Copy(path.Join(conf.TemplateDir, "./assets"), path.Join(outDir, "./assets"))
	util.RunningLog("2.1", "copy 完成")

	// 流程 3  遍历源目录 生成 html 到输出目录
	util.RunningLog("3", "生成 html")

	// 转换数据结构 filepath => entityList
	util.RunningLog("3.1", "收集转换生成所需数据")

	noteStore := store.DirToStruct(
		sourceDir,
		conf.SqlitePath,
		TemplateRender,
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

			// FileEntityToHTML entity 转 html
			FileEntityToHTML := func(entity structAll.FileEntity) (html string, xml string) {
				context := oceanpress.Context{}
				context.Db = db
				context.BaseEntity = entity
				context.FindFileEntityFromID = FindFileEntityFromID
				context.LuteEngine = luteEngine
				context.StructToHTML = structToHTML
				renderer := oceanpress.NewOceanPressRenderer(entity.Tree, (*oceanpress.Options)(luteEngine.RenderOptions), &context)

				return renderer.Render()
			}
			return FileEntityToHTML
		})
	util.RunningLog("3.2", "复制资源文件")
	for _, entity := range noteStore.StructList {
		if entity.Tree == nil {
			// 目录
		} else {
			HandlingAssets(entity.Tree.Root, outDir, entity.RootPath())
		}
	}

	util.RunningLog("3.3", "从文件到数据结构转换完毕，开始生成html,共", len(noteStore.StructList), "项")

	for _, entity := range noteStore.StructList {
		info := entity.Info
		relativePath := entity.RelativePath
		virtualPath := entity.VirtualPath

		LevelRoot := entity.RootPath()

		if info.IsDir() {
			if conf.IsDev {
				continue
			}
			// 这里要生成一个类似于当前目录菜单的东西
			targetPath := filepath.Join(outDir, relativePath, "index.html")
			// 当前目录的 子路径 不包含更深层级的
			sonList := fileEntityListFilter(noteStore.StructList, func(f structAll.FileEntity) bool {
				return strings.HasPrefix(f.VirtualPath, virtualPath) &&
					// 这个条件去除了间隔一层以上的其他路径
					strings.LastIndex(f.VirtualPath[len(virtualPath):], "/") == 0
			})

			var sonEntityList []sonEntityI
			for _, sonEntity := range sonList {
				webPath := sonEntity.VirtualPath[len(virtualPath):]
				var name string
				if sonEntity.Info.IsDir() {
					name = webPath + "/"
					webPath += "/index.html"
				} else {
					name = sonEntity.Name
				}

				sonEntityList = append(sonEntityList, sonEntityI{
					WebPath: webPath,
					Name:    name,
					IsDir:   sonEntity.Info.IsDir(),
				})
			}
			var menuInfo = (MenuInfo{
				SonEntityList: sonEntityList,
				PageTitle:     "菜单页",
				LevelRoot:     LevelRoot,
			})
			html := menuInfo.Render()
			ioutil.WriteFile(targetPath, []byte(html), 0777)
		} else {
			startT := time.Now()
			targetPath := filepath.Join(outDir, relativePath[0:len(relativePath)-3]) + ".html"

			rawHTML, xml := entity.Output()
			if len(rawHTML) != 0 {
				html := ArticleRender(ArticleInfo{
					Content:   template.HTML(rawHTML),
					PageTitle: entity.Name,
					LevelRoot: LevelRoot,
				})
				if conf.RssNoOutputHtml == false {
					var err = ioutil.WriteFile(targetPath, []byte(html), 0777)
					if err != nil {
						util.Log(err)
					}
				}
			}
			if len(xml) != 0 {
				targetPath := filepath.Join(outDir, relativePath[0:len(relativePath)-3])
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
func HandlingAssets(node *ast.Node, outDir string, rootPath string) {
	if node.Next != nil {
		HandlingAssets(node.Next, outDir, rootPath)
	}
	if node.FirstChild != nil {
		HandlingAssets(node.FirstChild, outDir, rootPath)
	}
	for _, n := range node.Children {
		HandlingAssets(n, outDir, rootPath)
	}

	if node != nil && node.Type == ast.NodeLinkDest {
		dest := node.TokensStr()

		if strings.HasPrefix(filepath.ToSlash(dest), "assets/") {
			err := copy.Copy(path.Join(path.Join(conf.AssetsDir, dest[len("assets/"):])), path.Join(outDir, dest))
			if err != nil {
				util.Warn("复制资源文件失败", err)
			}

			node.Tokens = []byte(path.Join(rootPath, dest))

		}
	}
}
