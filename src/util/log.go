package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Log 打印
func Log(a ...interface{}) {
	fmt.Println(a...)
}

// Debugger 在这个函数内加上断点，方便调试
func Debugger(a ...interface{}) {
	var log []interface{}
	log = append(log, "[debug]")
	fmt.Println(append(log, a...)...)
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
