package main

import (
	"os"
	"path/filepath"

	"github.com/2234839/md2website/src/util"
)

// 用于获取用户输入的参数

var curPath = os.Args[0]

// SourceDir 源目录
var SourceDir, _ = filepath.Abs(os.Args[1])

// OutDir 输出到的目录
var OutDir, _ = filepath.Abs(os.Args[2])

// TemplateDir 模板文件所在的目录
var TemplateDir, _ = filepath.Abs(os.Args[3])

func init() {
	util.Log(curPath)
}
