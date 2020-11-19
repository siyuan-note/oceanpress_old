package main

import (
	"bytes"
	"html/template"
)

// HTMLtemplate 包含一些模板
var HTMLtemplate = template.Must(template.ParseGlob(TemplateDir + "*.html"))

func init() {
	template.Must(HTMLtemplate.ParseGlob(TemplateDir + "*/*.html"))
}

// ExecTemplate 执行模板
/**
 * 现有模板 menu
 */
func ExecTemplate(name string, data interface{}) string {
	buf := new(bytes.Buffer)
	HTMLtemplate.ExecuteTemplate(buf, name, data)
	return buf.String()
}
