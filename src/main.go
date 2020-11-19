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
	util.Log("----- 流程 1 -----")
	sourceDir := "C:\\Users\\llej\\AppData\\Local\\Programs\\SiYuan\\resources\\guide\\思源笔记用户指南"
	outDir := "D:\\res\\go2website_test_out"
	util.Log("sourceDir:" + sourceDir + "\n" + "outDir:" + outDir)

	// 流程 2  copy 源目录中资源文件至输出目录
	util.Log("----- 流程 2 -----")
	copy.Copy(sourceDir, outDir, copy.Options{
		// 跳过一些不必要的目录以及 md 文件
		Skip: func(src string) (bool, error) {
			return (isSkipPath(src) || strings.HasSuffix(src, ".md")), nil
		},
	})
	util.Log("copy 完成")

	// 流程 3  遍历源目录 生成 html 到输出目录
	util.Log("----- 流程 3 -----")
	luteEngine := lute.New()
	filepath.Walk(sourceDir,
		func(path string, info os.FileInfo, err error) error {
			relativePath := path[len(sourceDir):]
			if err != nil {
				return err
			} else if isSkipPath(path) || !strings.HasSuffix(relativePath, ".md") {
				return nil
			} else {
				if info.IsDir() {
					// 这里应该要生成一个类似于当前目录菜单的东西
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
					fmt.Println(relativePath, info.Size())
				}
				return nil
			}
		})
	// End
}

func isSkipPath(path string) bool {
	return strings.Contains(path, ".git")
}
