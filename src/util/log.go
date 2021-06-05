package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	conf "github.com/siyuan-note/oceanpress/src/conf"
)

// Log 打印
func Log(a ...interface{}) {
	fmt.Println(a...)
}

// DevLog 仅在开发模式下打印
func DevLog(a ...interface{}) {
	if conf.IsDev {
		fmt.Println(a...)
	}
}

// Debugger 在这个函数内加上断点，方便调试
func Debugger(a ...interface{}) {
	if conf.IsDev {
		var log []interface{}
		log = append(log, "[debug]")
		fmt.Println(append(log, a...)...) // 在开发的时候可以在这里加一个断点，方便调试
	}
}

// Warn 警告
func Warn(a ...interface{}) {
	var log []interface{}
	log = append(log, "[warn]")
	fmt.Println(append(log, a...)...)
}

// RunningLog 打印稍微带一些格式的运行时流程信息
func RunningLog(serialNumber string, a ...interface{}) {
	l := strings.Split(serialNumber, ".")
	var log []interface{}
	if len(l) == 1 {
		log = append(log, serialNumber+".")
		log = append(log, a...)
	} else if len(l) > 1 {
		log = append(log, strings.Repeat("  ", len(l)-1), serialNumber)
		log = append(log, a...)
	}
	Log(log...)
}

// RenderError 用于渲染失败的地方
func RenderError() string {
	// TODO: 非调试模式应该返回空字符串
	return "<此处渲染失败>"
}

// HTMLEntityDecoder
func HTMLEntityDecoder(s string) string {
	re3, _ := regexp.Compile("&#\\d+?;")
	rep := re3.ReplaceAllStringFunc(s, func(s string) string {
		asciiCode, _ := strconv.ParseInt(s[2:len(s)-1], 10, 64)
		v := string(asciiCode)
		return v
	})
	return rep
}
