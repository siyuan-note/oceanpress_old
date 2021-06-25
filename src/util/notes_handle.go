package util

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func IsSkipPath(path string) bool {
	p := filepath.ToSlash(path)
	return strings.Contains(p, "/.git/") || strings.Contains(p, "/.siyuan")
}

var NotesSuffix = ".sy"

// IsNotes 判断该路径是不是笔记
func IsNotes(path string) bool {
	return strings.HasSuffix(path, ".sy") || strings.HasSuffix(path, ".md")
}

func FindAttr(attrs [][]string, name string) (string, int, error) {
	for i, kv := range attrs {
		if name == kv[0] {
			return kv[1], i, nil
		}
	}
	return "", 0, errors.New("没有找到对应的 attr")
}

// IDTimeStrToTime IDTime=>20210607114722
func IDTimeStrToTime(IDTime string) time.Time {
	l := len(IDTime)
	sec, _ := strconv.Atoi(IDTime[l-2:])
	min, _ := strconv.Atoi(IDTime[l-4 : l-2])
	hour, _ := strconv.Atoi(IDTime[l-6 : l-4])
	day, _ := strconv.Atoi(IDTime[l-8 : l-6])
	month, _ := strconv.Atoi(IDTime[l-10 : l-8])
	year, _ := strconv.Atoi(IDTime[:l-10])

	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
}
