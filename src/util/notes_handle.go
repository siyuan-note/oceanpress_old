package util

import (
	"path/filepath"
	"strings"
)

func IsSkipPath(path string) bool {
	p := filepath.ToSlash(path)
	return strings.Contains(p, "/.git/") || strings.Contains(p, "/.siyuan/backup/")
}

var NotesSuffix = ".sy"

// IsNotes 判断该路径是不是笔记
func IsNotes(path string) bool {
	return strings.HasSuffix(path, ".sy") || strings.HasSuffix(path, ".md")
}
