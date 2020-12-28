package main

import (
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/2234839/md2website/src/util"
	copy "github.com/otiai10/copy"
)

func main() {
	util.RunningLog("0", "=== 🛬 开始转换 🛫 ===")
	// 流程 1  用户输入 {源目录 输出目录}
	util.RunningLog("1", "用户输入")
	sourceDir := SourceDir
	outDir := OutDir
	util.RunningLog("1.1", "sourceDir:" + sourceDir)
	util.RunningLog("1.2", "outDir:" + outDir)
	util.RunningLog("1.3", "viewsDir:" + TemplateDir)

	// 流程 2  copy 源目录中资源文件至输出目录
	util.RunningLog("2", "copy 资源到 outDir")

	copy.Copy(sourceDir, outDir, copy.Options{
		// 跳过一些不必要的目录以及 md 文件
		Skip: func(src string) (bool, error) {
			return (isSkipPath(src) || strings.HasSuffix(src, ".md")), nil
		},
	})
	// copy views 中的资源文件
	copy.Copy(path.Join(TemplateDir, "./assets"), path.Join(outDir, "./assets"))
	util.RunningLog("2.1", "copy 完成")

	// 流程 3  遍历源目录 生成 html 到输出目录
	util.RunningLog("3", "生成 html")

	// 转换数据结构 filepath => entityList
	filepath.Walk(sourceDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if isSkipPath(path) || (!info.IsDir() && !strings.HasSuffix(path, ".md")) {
				return nil
			} else {
				FileEntityList = append(FileEntityList, FileToFileEntity(path, info))
				return nil
			}
		})

	util.RunningLog("3.1", "从文件到数据结构转换完毕，开始生成html,共", len(FileEntityList), "项")

	for _, entity := range FileEntityList {
		info := entity.info
		relativePath := entity.relativePath
		virtualPath := entity.virtualPath

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
			sonList := fileEntityListFilter(FileEntityList, func(f FileEntity) bool {
				return strings.HasPrefix(f.virtualPath, virtualPath) &&
					// 这个条件去除了间隔一层以上的其他路径
					strings.LastIndex(f.virtualPath[len(virtualPath):], "/") == 0
			})

			var sonEntityList []sonEntityI
			for _, sonEntity := range sonList {
				webPath := sonEntity.virtualPath[len(virtualPath):]
				var name string
				if sonEntity.info.IsDir() {
					name = webPath + "/"
					webPath += "/index.html"
				} else {
					name = sonEntity.name
				}

				sonEntityList = append(sonEntityList, sonEntityI{
					WebPath: webPath,
					Name:    name,
					IsDir:   sonEntity.info.IsDir(),
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

			rawHTML := FileEntityToHTML(entity)

			html := ArticleRender(ArticleInfo{
				Content:   template.HTML(rawHTML),
				PageTitle: entity.name,
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

func isSkipPath(path string) bool {
	return strings.Contains(path, ".git")
}

// go 怎么写类似于其他语言泛型的过滤方式 ？// https://medium.com/@habibridho/here-is-why-no-one-write-generic-slice-filter-in-go-8b3d1063674e
func fileEntityListFilter(list []FileEntity, test func(FileEntity) bool) (ret []FileEntity) {
	for _, s := range list {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
