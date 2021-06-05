package conf

import (
	"flag"
	"os"
	"path/filepath"
)

// 用于获取用户输入的参数
var curPath = os.Args[0]

// SourceDir 源目录
var SourceDir string

// OutDir 输出到的目录
var OutDir string

// TemplateDir 模板文件所在的目录
var TemplateDir string

// sqlite db 文件的路径
var SqlitePath string
var AssetsDir string

// 是否处于开发模式
var IsDev bool = false

func init() {
	SourceDir_ := flag.String("SourceDir", "", "笔记本所在目录")
	OutDir_ := flag.String("OutDir", "", "将结果输出到此目录")
	TemplateDir_ := flag.String("TemplateDir", "./views/", "模板文件所在目录")
	SqlitePath_ := flag.String("SqlitePath", "", "思源 sqlite db 路径")
	AssetsDir_ := flag.String("AssetsDir", "", "思源存放资源文件的路径")
	assetsDir_ := flag.String("assetsDir", "", "思源存放资源文件的路径")
	IsDev_ := flag.Bool("IsDev", false, "思源存放资源文件的路径")
	flag.Parse()

	SourceDir, _ = filepath.Abs(*SourceDir_)
	OutDir, _ = filepath.Abs(*OutDir_)
	TemplateDir, _ = filepath.Abs(*TemplateDir_)
	SqlitePath, _ = filepath.Abs(*SqlitePath_)
	AssetsDir, _ = filepath.Abs(*AssetsDir_)

	// TODO： 再过上几个版本删掉此提示
	if *assetsDir_ != "" {
		panic("\n===\n assetsDir 参数已更名为 AssetsDir , 请修正 \n===")
	}
	IsDev = *IsDev_
}
