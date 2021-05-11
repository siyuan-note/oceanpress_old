package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/2234839/md2website/src/util"
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
var assetsDir string

func init() {
	SourceDir_ := flag.String("SourceDir", "", "笔记本所在目录")
	OutDir_ := flag.String("OutDir", "", "将结果输出到此目录")
	TemplateDir_ := flag.String("TemplateDir", "./views/", "模板文件所在目录")
	SqlitePath_ := flag.String("SqlitePath", "", "思源 sqlite db 路径")
	assetsDir_ := flag.String("assetsDir", "", "思源存放资源文件的路径")
	flag.Parse()

	SourceDir, _ = filepath.Abs(*SourceDir_)
	OutDir, _ = filepath.Abs(*OutDir_)
	TemplateDir, _ = filepath.Abs(*TemplateDir_)
	SqlitePath, _ = filepath.Abs(*SqlitePath_)
	assetsDir, _ = filepath.Abs(*assetsDir_)
	util.Log(flag.NArg(), SourceDir, OutDir, TemplateDir, SqlitePath, assetsDir)
}
