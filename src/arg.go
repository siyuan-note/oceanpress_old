package main

import "os"

// 用于获取用户输入的参数

// SourceDir 源目录
var SourceDir = os.Args[1]

// OutDir 输出到的目录
var OutDir = os.Args[2]

// TemplateDir 模板文件所在的目录
var TemplateDir = "./views/"
