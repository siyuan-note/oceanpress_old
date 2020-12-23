package util

import (
	"fmt"
	"strings"
)

// Log 打印
func Log(a ...interface{}) {
	fmt.Println(a...)
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
