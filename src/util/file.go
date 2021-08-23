package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/siyuan-note/oceanpress/src/conf"
)

// WriteFile 确保文件路径存在，且不超出 conf.outDir
func WriteFile(targetPath string, data []byte, perm os.FileMode) error {
	rel, err := filepath.Rel(conf.OutDir, targetPath)
	if err != nil {
		return err
	}
	if strings.Contains(rel, "..") {
		// 不是子路径
		return errors.New("要写入的文件路径不是 conf.OutDir 的子路径")
	}
	dir := path.Dir(targetPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0700) // Create your file
	}
	return ioutil.WriteFile(targetPath, data, perm)

}
