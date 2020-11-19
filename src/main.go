package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/2234839/md2website/src/util"
	"github.com/88250/lute"
	copy "github.com/otiai10/copy"
)

func main() {
	// 流程 1  用户输入 {源目录 输出目录}
	util.Log("----- 流程 1 用户输入 -----")
	sourceDir := SourceDir
	outDir := OutDir
	util.Log("sourceDir:" + sourceDir + "\n" + "outDir:" + outDir)

	// 流程 2  copy 源目录中资源文件至输出目录
	util.Log("----- 流程 2 copy 资源 -----")
	copy.Copy(sourceDir, outDir, copy.Options{
		// 跳过一些不必要的目录以及 md 文件
		Skip: func(src string) (bool, error) {
			return (isSkipPath(src) || strings.HasSuffix(src, ".md")), nil
		},
	})
	util.Log("copy 完成")

	// 流程 3  遍历源目录 生成 html 到输出目录
	util.Log("----- 流程 3 生成 html -----")
	luteEngine := lute.New()

	var entityList []fileEntity
	filepath.Walk(sourceDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if isSkipPath(path) || (!info.IsDir() && !strings.HasSuffix(path, ".md")) {
				return nil
			} else {
				relativePath := strings.ReplaceAll(path[len(sourceDir):], string(os.PathSeparator), "/")
				var virtualPath string
				if info.IsDir() {
					virtualPath = relativePath
				} else {
					virtualPath = relativePath[0:len(relativePath)-29] + ".html"
				}

				entityList = append(entityList, fileEntity{
					path:         path,
					info:         info,
					relativePath: relativePath,
					virtualPath:  virtualPath,
				})
				return nil
			}
		})

	fmt.Println("开始生成html,共", len(entityList), "项")

	for _, entity := range entityList {
		info := entity.info
		path := entity.path
		relativePath := entity.relativePath
		virtualPath := entity.virtualPath

		if info.IsDir() {
			// 这里要生成一个类似于当前目录菜单的东西
			targetPath := filepath.Join(outDir, relativePath, "index.html")
			// 当前目录的 子路径 不包含更深层级的
			sonList := fileEntityListFilter(entityList, func(f fileEntity) bool {
				return strings.HasPrefix(f.virtualPath, virtualPath) &&
					// 这个条件去除了间隔一层以上的其他路径
					strings.LastIndex(f.virtualPath[len(virtualPath):], "/") == 0
			})
			type sonEntityI struct {
				WebPath string
			}
			var sonEntityList []sonEntityI
			for _, sonEntity := range sonList {
				webPath := sonEntity.virtualPath[len(virtualPath):]
				if sonEntity.info.IsDir() {
					webPath += "/index.html"
				}
				sonEntityList = append(sonEntityList, sonEntityI{
					WebPath: webPath,
				})
			}

			type menuInfo struct {
				SonEntityList []sonEntityI
				PageTitle     string
			}
			html := ExecTemplate("menu", menuInfo{
				SonEntityList: sonEntityList,
				PageTitle:     "菜单页",
			})
			ioutil.WriteFile(targetPath, []byte(html), 0777)
			fmt.Println(relativePath, len(sonEntityList))
		} else {
			// 这里的 targetPath 是有问题的，他隐含了一个条件就是md文档一定包含id
			targetPath := filepath.Join(outDir, relativePath[0:len(relativePath)-29]) + ".html"
			mdByte, err := ioutil.ReadFile(path)
			if err != nil {
				util.Log("读取文件失败", err)
			}
			mdStr := string(mdByte)

			html := luteEngine.MarkdownStr("", mdStr)
			ioutil.WriteFile(targetPath, []byte(html), 0777)
			// fmt.Println(relativePath, info.Size())
		}
	}
	// End
	util.Log("----- End -----")

}

func isSkipPath(path string) bool {
	return strings.Contains(path, ".git")
}

type fileEntity struct {
	path         string
	relativePath string
	// 最终要可以访问的路径
	virtualPath string
	info        os.FileInfo
}

// go 怎么写类似于其他语言泛型的过滤方式 ？// https://medium.com/@habibridho/here-is-why-no-one-write-generic-slice-filter-in-go-8b3d1063674e
func fileEntityListFilter(list []fileEntity, test func(fileEntity) bool) (ret []fileEntity) {
	for _, s := range list {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
