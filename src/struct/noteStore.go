package structAll

import (
	"strings"

	"github.com/siyuan-note/oceanpress/src/conf"
	"github.com/siyuan-note/oceanpress/src/util"
)

// 来自 main.go 的生成
var NoteStore = DirToStructRes{}

// PathResolve 处理不同模式下路径的变换
func PathResolve(path string) string {
	// 使用文档名作为路径名
	if conf.OutMode == "title" {
		entries := strings.Split(path, "/")
		var virtualPath = []string{}
		for _, v := range entries {
			if v == "" {
				virtualPath = append(virtualPath, v)
				continue
			}
			suffix := ""
			fragment := strings.Split(v, ".")
			id := fragment[0]
			if len(fragment) > 1 {
				suffix = "." + fragment[1]
			}
			if util.IsID(id) {
				FileEntity, _, err := NoteStore.FindFileEntityFromID(id)
				if err == nil {
					virtualPath = append(virtualPath, FileEntity.Name+suffix)
					continue
				}
			}
			virtualPath = append(virtualPath, v)
		}
		path = strings.Join(virtualPath, "/")
		return path
	} else {
		// 直接使用 ID 作为路径名
		if conf.OutMode != "id" {
			util.Warn("OutMode 参数的值在预设之外，默认采用 id 模式")
		}
		return path
	}
}
