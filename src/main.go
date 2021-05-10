package main

import (
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	store "github.com/2234839/md2website/src/store"
	"github.com/2234839/md2website/src/util"
	copy "github.com/otiai10/copy"
)

func main() {

	util.RunningLog("0", "=== 🛬 开始转换 🛫 ===")
	// 流程 1  用户输入 {源目录 输出目录}
	util.RunningLog("1", "用户输入")
	sourceDir := SourceDir
	outDir := OutDir
	util.RunningLog("1.1", "sourceDir:"+sourceDir)
	util.RunningLog("1.2", "outDir:"+outDir)
	util.RunningLog("1.3", "viewsDir:"+TemplateDir)
	util.RunningLog("1.4", "dbPath:"+SqlitePath)

	// 流程 2  copy 源目录中资源文件至输出目录
	util.RunningLog("2", "copy 资源到 outDir")

	copy.Copy(sourceDir, outDir, copy.Options{
		// 跳过一些不必要的目录以及 md 文件
		Skip: func(src string) (bool, error) {
			return (util.IsSkipPath(src) || util.IsNotes(src)), nil
		},
	})
	// copy views 中的资源文件
	copy.Copy(path.Join(TemplateDir, "./assets"), path.Join(outDir, "./assets"))
	util.RunningLog("2.1", "copy 完成")

	// 流程 3  遍历源目录 生成 html 到输出目录
	util.RunningLog("3", "生成 html")

	// 转换数据结构 filepath => entityList

	noteStore := store.DirToStruct(sourceDir, SqlitePath, TemplateRender)

	util.RunningLog("3.1", "从文件到数据结构转换完毕，开始生成html,共", len(noteStore.StructList), "项")

	for _, entity := range noteStore.StructList {
		info := entity.Info
		relativePath := entity.RelativePath
		virtualPath := entity.VirtualPath

		Level := strings.Count(relativePath, "/") - 1
		if info.IsDir() {
			Level++
		}
		// relativePath 通过 LevelRoot 可以跳转到生成目录，即根目录
		var LevelRoot = "./"
		if Level > 0 {
			LevelRoot += strings.Repeat("../", Level)
		}

		if info.IsDir() {
			// 这里要生成一个类似于当前目录菜单的东西
			targetPath := filepath.Join(outDir, relativePath, "index.html")
			// 当前目录的 子路径 不包含更深层级的
			sonList := fileEntityListFilter(noteStore.StructList, func(f store.FileEntity) bool {
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

			type menuInfo struct {
				SonEntityList []sonEntityI
				PageTitle     string
			}
			html := MenuRender(MenuInfo{
				SonEntityList: sonEntityList,
				PageTitle:     "菜单页",
				LevelRoot:     LevelRoot,
			})
			ioutil.WriteFile(targetPath, []byte(html), 0777)
		} else {
			targetPath := filepath.Join(outDir, relativePath[0:len(relativePath)-3]) + ".html"

			// rawHTML := mdtransform.FileEntityToHTML(entity)
			rawHTML := entity.ToHTML()

			html := ArticleRender(ArticleInfo{
				Content:   template.HTML(rawHTML),
				PageTitle: entity.Name,
				LevelRoot: LevelRoot,
			})
			var err = ioutil.WriteFile(targetPath, []byte(html), 0777)
			if err != nil {
				util.Log(err)
			}

		}
	}
	// End
	util.Log("----- End -----")

}

// go 怎么写类似于其他语言泛型的过滤方式 ？// https://medium.com/@habibridho/here-is-why-no-one-write-generic-slice-filter-in-go-8b3d1063674e
func fileEntityListFilter(list []store.FileEntity, test func(store.FileEntity) bool) (ret []store.FileEntity) {
	for _, s := range list {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
