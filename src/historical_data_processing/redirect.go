package historical_data_processing

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/siyuan-note/oceanpress/src/conf"
	structAll "github.com/siyuan-note/oceanpress/src/struct"
	"github.com/siyuan-note/oceanpress/src/template"
	"github.com/siyuan-note/oceanpress/src/util"
)

type item struct {
	ID   string
	Path string
}

func generateIDPathJSON() (IDPathJSON []item) {
	for _, entity := range structAll.NoteStore.StructList {
		if !entity.Info.IsDir() {
			path := entity.VirtualPath()
			IDPathJSON = append(IDPathJSON, item{ID: entity.Tree.ID, Path: path})
		}
	}
	return IDPathJSON
}
func writeRedirectFile(oldItem item) {
	entity, _, err := structAll.NoteStore.FindFileEntityFromID(oldItem.ID)
	if err == nil {
		tierCount := strings.Count(oldItem.Path, "/") // 获取文件夹层级
		rootPath := ""
		// i=1 的原因是默认路径前置了一个斜杠，要跳过这个
		for i := 1; i < tierCount; i++ {
			rootPath += "../"
		}
		redirectPath := path.Join(rootPath, entity.VirtualPath()+"#"+oldItem.ID)
		html := template.TemplateRender(structAll.RedirectInfo{
			RedirectPath: redirectPath,
			Title:        entity.Name,
		})
		err := util.WriteFile(path.Join(conf.OutDir, oldItem.Path), []byte(html), 0777)
		if err != nil {
			util.Warn("<写重定向html失败>", err)
		}
	}

}

func GenerateRedirectFile() {
	targetPath := filepath.Join(conf.OutDir, "assets/historical_data", "redirect.json")

	IDPathJSON := generateIDPathJSON()
	oldJSONByte, err := ioutil.ReadFile(targetPath)
	if err == nil {
		var oldJSON []item
		err := json.Unmarshal(oldJSONByte, &oldJSON)
		if err == nil {
			for _, oldItem := range oldJSON {
				exist := false
				for _, item := range IDPathJSON {
					if item.ID == oldItem.ID {
						exist = true
						if item.Path != oldItem.Path {
							// 文档是被移动的，将这条记录添加到 redirect.json 中
							IDPathJSON = append(IDPathJSON, oldItem)
							// 文档被移动，进行重定向
							writeRedirectFile(oldItem)
							break
						} else {
							// 和历史相同，不用管
						}
					}
				}
				if !exist {
					// 输出带id重定向 html
					writeRedirectFile(oldItem)
				}
			}
		}
	}

	jsonByte, err := json.Marshal(IDPathJSON)
	if err != nil {
		util.Warn("<生成用于重定向的历史数据失败>", err)
	} else {
		err := util.WriteFile(targetPath, jsonByte, 0777)
		if err != nil {
			util.Warn("<写入用于重定向的历史数据失败>", err)
		}
	}
}
