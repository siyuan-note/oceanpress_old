package store

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	sqlite "github.com/2234839/md2website/src/sqlite"
	"github.com/2234839/md2website/src/util"
	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	protyle "github.com/88250/protyle"
)

// StructInfo md 结构信息
type StructInfo struct {
	blockID   string
	blockType string
	node      *ast.Node
}

// FileEntity md 文件被解析后的结构
type FileEntity struct {
	Name string
	// 文件绝对路径
	Path string
	// 相对源目录的路径
	RelativePath string
	// 最终要在浏览器中可以访问的路径
	VirtualPath    string
	NotesCode      string
	Info           os.FileInfo
	StructInfoList []StructInfo
	Tree           *parse.Tree
	ToHTML         func() string
}

// RootPath 获取当前对象相对于 root 目录的路径
func (r *FileEntity) RootPath() string {
	Level := strings.Count(r.VirtualPath, "/") - 1
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

func addKramdownIAL(node *ast.Node) {
	ctx := addKramdownIALContext{}
	addAll(node, &ctx)
	// 文档最后更新时间
	node.KramdownIAL = append(node.KramdownIAL, []string{"updated", strconv.Itoa(ctx.docUpdated)})
	//TODO: 所有块 id 都应该改成 data-block-id 这里应该要看下 lute 内是如何实现的，不应该这里还要写死
	for _, v := range node.KramdownIAL {
		if v[0] == "id" {
			node.KramdownIAL = append(node.KramdownIAL, []string{"data-block-id", v[1]})
		}
	}
}

type addKramdownIALContext struct {
	docUpdated int
}

// addAll 遍历整颗树，附加一些数据到 KramdownIAL
func addAll(node *ast.Node, ctx *addKramdownIALContext) {
	for _, v := range node.KramdownIAL {
		// 获取文档最后更新时间
		if v[0] == "updated" {
			updated, _ := strconv.Atoi(v[1])
			if updated > ctx.docUpdated {
				ctx.docUpdated = updated
			}
		}
	}
	node.KramdownIAL = append(node.KramdownIAL, []string{"data-type", node.Type.String()})

	if node.Next != nil {
		addAll(node.Next, ctx)
	}
	if node.FirstChild != nil {
		addAll(node.FirstChild, ctx)
	}
	for _, n := range node.Children {
		addAll(n, ctx)
	}
}

// DirToStruct 从 目录 转为更可用的结构
func DirToStruct(dir string, dbPath string, structToHTML func(interface{}) string) DirToStructRes {
	db := sqlite.InitDb(dbPath)
	/** 用于从 md 文档中解析获得一些结构性信息 */
	var mdStructuredLuteEngine = lute.New()
	mdStructuredLuteEngine.SetKramdownIAL(true)
	mdStructuredLuteEngine.SetKramdownIALIDRenderName("data-block-id")

	// GetStructInfoByNotesCode 从 NotesCode 获取结构信息 suffix 是文件后缀目前支持 .sy 和 .md
	GetStructInfoByNotesCode := func(notesCode string, suffix string) ([]StructInfo, *parse.Tree) {
		var tree *parse.Tree
		var err error
		if suffix == ".sy" {
			var needFix bool
			tree, needFix, err = protyle.ParseJSON(lute.New(), []byte(notesCode))
			if needFix {
				util.Debugger("needFix", needFix)
			}
		} else if suffix == ".md" {
			tree = parse.Parse("", []byte(notesCode), mdStructuredLuteEngine.ParseOptions)
		}
		addKramdownIAL(tree.Root)
		if err != nil {
			panic(err)
		}
		var infoList []StructInfo
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if entering {
				return ast.WalkContinue
			}

			if nil == n.FirstChild {
				return ast.WalkSkipChildren
			}
			infoList = append(infoList, StructInfo{
				blockID:   n.IALAttr("id"),
				blockType: n.Type.String(),
				node:      n,
			})
			return ast.WalkContinue
		})
		return infoList, tree
	}

	// StructList 解析后的所有对象
	var StructList []FileEntity
	// FindFileEntityFromID 通过id找到对应的数据 这里之后要改一下，用 map 会比 双重for 好一些
	FindFileEntityFromID := func(id string) (FileEntity, StructInfo, error) {
		var fileEntity FileEntity
		var mdInfo StructInfo
		for _, entity := range StructList {
			for _, info := range entity.StructInfoList {
				if info.blockID == id {
					fileEntity = entity
					mdInfo = info
					break
				} else {
					continue
				}
			}
			// 这个代表已经找到了
			if fileEntity.Path != "" {
				break
			}
		}
		if fileEntity.Path == "" {
			var msg = "未找到id " + id + " 对应的fileEntity"
			util.Warn(msg)
			return fileEntity, mdInfo, errors.New(msg)
		}
		return fileEntity, mdInfo, nil
	}

	FileEntityToHTML := Generate(db, FindFileEntityFromID, structToHTML)

	// FileToFileEntity 通过文件路径以及文件信息获取他的结构信息
	FileToFileEntity := func(sourceDir string, path string, info os.FileInfo) FileEntity {
		relativePath := strings.ReplaceAll(path[len(sourceDir):], string(os.PathSeparator), "/")
		var virtualPath string
		var notesCode string
		var name string
		var StructInfo []StructInfo
		var tree *parse.Tree
		if strings.Contains(path, "css 变量") {
			// util.Debugger(path)
		}
		if info.IsDir() {
			virtualPath = relativePath
		} else {
			virtualPath = FilePathToWebPath(relativePath)
			mdByte, err := ioutil.ReadFile(path)
			if err != nil {
				util.Warn("读取文件失败", err)
			}
			notesCode = string(mdByte)
			StructInfo, tree = GetStructInfoByNotesCode(notesCode, filepath.Ext(path))
			if util.IsNotes(relativePath) {
				baseName := filepath.Base(relativePath)
				name = baseName[:len(baseName)-3]
			}
		}
		entity := FileEntity{
			Path:           path,
			Info:           info,
			RelativePath:   relativePath,
			VirtualPath:    virtualPath,
			NotesCode:      notesCode,
			StructInfoList: StructInfo,
			Name:           name,
			Tree:           tree,
		}
		entity.ToHTML = func() string {
			return FileEntityToHTML(entity)
		}

		return entity
	}

	filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			p := filepath.ToSlash(path)
			if err != nil {
				return err
			} else if util.IsSkipPath(p) || (!info.IsDir() && !util.IsNotes(p)) {
				return nil
			} else {
				StructList = append(StructList, FileToFileEntity(dir, p, info))
				return nil
			}
		})
	return DirToStructRes{
		StructList:           StructList,
		FindFileEntityFromID: FindFileEntityFromID,
	}
}

// FilePathToWebPath 将相对文件路径转为 web路径，主要是去除文件中的id 以及添加 .html
func FilePathToWebPath(filePath string) string {
	if util.IsNotes(filePath) {
		return filePath[0:len(filePath)-len(util.NotesSuffix)] + ".html"
	}
	// 大概率是空
	return filePath
}
