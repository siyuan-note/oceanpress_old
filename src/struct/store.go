package structAll

import (
	"os"
	"strings"
	"time"

	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	"github.com/siyuan-note/oceanpress/src/conf"
	"github.com/siyuan-note/oceanpress/src/util"
)

// StructInfo md 结构信息
type StructInfo struct {
	BlockID    string
	BlockType  string
	Node       *ast.Node
	FileEntity *FileEntity
}

func (r *StructInfo) GetCreated() time.Time {
	return util.IDTimeStrToTime(r.BlockID)
}
func (r *StructInfo) GetUpdate() time.Time {
	updated, _, _ := util.FindAttr(r.Node.KramdownIAL, "updated")
	return util.IDTimeStrToTime(updated)
}

// FileEntity md 文件被解析后的结构
type FileEntity struct {
	// 文档名 title
	Name string
	// 文件绝对路径
	Path string
	// 相对源目录的路径
	RelativePath   string
	NotesCode      string
	Info           os.FileInfo
	StructInfoList []StructInfo
	Tree           *parse.Tree
	Output         func() (html string, xml string)
}

// FilePathToWebPath 将相对文件路径转为 web路径，主要是去除文件中的id 以及添加 .html
func FilePathToWebPath(filePath string) string {
	if util.IsNotes(filePath) {
		return filePath[0:len(filePath)-len(util.NotesSuffix)] + ".html"
	}
	// 大概率是空
	return filePath
}

func (r *FileEntity) FileEntityRelativePath(target FileEntity, id string) string {
	base := r
	// 减一是因为 路径开头必有 / 而这里只需要跳到这一层
	count := strings.Count(base.VirtualPath(), "/")
	if strings.HasPrefix(base.VirtualPath(), "/") {
		count--
	}
	l2 := strings.Split(target.VirtualPath(), "/")
	url := strings.Repeat("../", count)
	url += strings.Join(l2[1:], "/")
	url = FilePathToWebPath(url)
	url += "#" + id
	return url
}

// VirtualPath 是最终要在浏览器中可以访问的路径
func (r *FileEntity) VirtualPath() (path string) {

	// 使用文档名作为路径名
	if conf.OutMode == "title" {
		entries := strings.Split(r.RelativePath, "/")
		var virtualPath = []string{}
		for _, v := range entries {
			id := v
			if strings.HasSuffix(v, util.NotesSuffix) {
				id = v[:len(v)-len(util.NotesSuffix)]
			}
			if util.IsID(id) {
				FileEntity, _, err := NoteStore.FindFileEntityFromID(id)
				if err == nil {
					virtualPath = append(virtualPath, FileEntity.Name)
					continue
				}
			}
			virtualPath = append(virtualPath, id)
		}
		path = strings.Join(virtualPath, "/")
		if util.IsNotes(r.RelativePath) {
			path += ".html"
		}
		return path
	} else {
		// 直接使用 ID 作为路径名
		if conf.OutMode != "id" {
			util.Warn("OutMode 参数的值在预设之外，默认采用 id 模式")
		}
		if util.IsNotes(r.RelativePath) {
			return r.RelativePath[0:len(r.RelativePath)-len(util.NotesSuffix)] + ".html"
		}
		return r.RelativePath
	}
}

// RootPath 获取当前对象相对于 root 目录的路径
func (r *FileEntity) RootPath() string {
	Level := strings.Count(r.VirtualPath(), "/") - 1
	if r.Info.IsDir() {
		Level++
	}
	// relativePath 通过 LevelRoot 可以跳转到生成目录，即根目录
	var LevelRoot = "./"
	if Level > 0 {
		LevelRoot += strings.Repeat("../", Level)
	}
	return LevelRoot
}

// FindFileEntityFromID 通过 id 返回对应实体
type FindFileEntityFromID func(id string) (FileEntity, StructInfo, error)

// DirToStructRes DirToStruct 的返回值定义
type DirToStructRes struct {
	StructList           []FileEntity
	FindFileEntityFromID FindFileEntityFromID
}

// EmbeddedBlockInfo 嵌入块所需信息
type EmbeddedBlockInfo struct {
	AEmbeddedBlockInfo int
	Title              interface{}
	Src                string
	Content            interface{}
}

// BlockRefInfo 块引用所需信息
type BlockRefInfo struct {
	ABlockRefInfo int
	Title         interface{}
	Src           string
}

type RssItem struct {
	Guid          string
	Title         string
	Link          string
	Published     string
	Updated       string
	Created       string
	LastBuildDate string
	Description   string
	ContentBase   string
	ContentHTML   string
}

// RssInfo 块引用所需信息
type RssInfo struct {
	ARssInfo      int
	LastBuildDate string
	List          []RssItem
}
