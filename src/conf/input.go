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

// 对于 [/s/S].rss.xml 的文档不再生成对应的 html 文件
var RssNoOutputHtml bool = true

// 是否处于开发模式
var IsDev bool = false

func init() {
	SourceDir_ := flag.String("SourceDir", "", "笔记本所在目录")
	OutDir_ := flag.String("OutDir", "", "将结果输出到此目录")
	TemplateDir_ := flag.String("TemplateDir", "./views/", "模板文件所在目录")
	SqlitePath_ := flag.String("SqlitePath", "", "思源 sqlite db 路径")
	AssetsDir_ := flag.String("AssetsDir", "", "思源存放资源文件的路径，已弃用，现在采取和思源一样的措施")
	assetsDir_ := flag.String("assetsDir", "", "思源存放资源文件的路径，已弃用，请使用 AssetsDir")
	IsDev_ := flag.Bool("IsDev", false, "设置为开发模式，请不要开启此选项")
	RssNoOutputHtml_ := flag.Bool("RssNoOutputHtml", true, "对于 [/s/S].rss.xml 的文档不再生成对应的 html 文件")
	flag.Parse()

	SourceDir, _ = filepath.Abs(*SourceDir_)
	OutDir, _ = filepath.Abs(*OutDir_)
	TemplateDir, _ = filepath.Abs(*TemplateDir_)
	SqlitePath, _ = filepath.Abs(*SqlitePath_)
	AssetsDir, _ = filepath.Abs(*AssetsDir_)
	IsDev = *IsDev_
	RssNoOutputHtml = *RssNoOutputHtml_

	// TODO： 再过上几个版本删掉此提示
	if *assetsDir_ != "" {
		parameterChangeHint("assetsDir 参数已更名为 AssetsDir , 请修正")
	}
	if *AssetsDir_ != "" {
		parameterChangeHint("AssetsDir 已弃用，现在采取和思源一样的措施，先寻找笔记本下的 assets 目录，再寻找笔记本上一层（工作空间）下的 assets 目录 , 请修正")
	}

}

func parameterChangeHint(msg string) {
	panic("\n=== 参数变更提示\n " + msg + " \n===")
}
