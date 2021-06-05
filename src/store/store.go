package store

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	protyle "github.com/88250/protyle"
	sqlite "github.com/siyuan-note/oceanpress/src/sqlite"
	structAll "github.com/siyuan-note/oceanpress/src/struct"
	"github.com/siyuan-note/oceanpress/src/util"
)

func addKramdownIAL(node *ast.Node) {
	ctx := addKramdownIALContext{}
	addAll(node, &ctx)
	// 文档最后更新时间
	node.KramdownIAL = append(node.KramdownIAL, []string{"updated", strconv.Itoa(ctx.docUpdated)})
	//TODO: 所有块 id 都应该改成 data-block-id 这里应该要看下 lute 内是如何实现的，不应该这里还要写死
	for _, v := range node.KramdownIAL {
		if v[0] == "id" {
			node.KramdownIAL = append(node.KramdownIAL, []string{"data-n-id", v[1]})
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
	if node.IALAttr("id") != "" {
		node.SetIALAttr("data-n-id", node.IALAttr("id"))
	}
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
func DirToStruct(dir string,
	dbPath string,
	structToHTML func(interface{}) string,
	Generate func(db sqlite.DbResult, FindFileEntityFromID structAll.FindFileEntityFromID, structToHTML func(interface{}) string) func(entity structAll.FileEntity) string) structAll.DirToStructRes {
	db := sqlite.InitDb(dbPath)
	/** 用于从 md 文档中解析获得一些结构性信息 */
	var mdStructuredLuteEngine = lute.New()
	mdStructuredLuteEngine.SetKramdownIAL(true)
	mdStructuredLuteEngine.SetKramdownIALIDRenderName("data-n-id")

	// GetStructInfoByNotesCode 从 NotesCode 获取结构信息 suffix 是文件后缀目前支持 .sy 和 .md
	GetStructInfoByNotesCode := func(notesCode string, suffix string) ([]structAll.StructInfo, *parse.Tree) {
		var tree *parse.Tree
		var err error
		if suffix == ".sy" {
			tree, _, err = protyle.ParseJSON(lute.New(), []byte(notesCode))
		} else if suffix == ".md" {
			tree = parse.Parse("", []byte(notesCode), mdStructuredLuteEngine.ParseOptions)
		}
		addKramdownIAL(tree.Root)
		if err != nil {
			panic(err)
		}
		var infoList []structAll.StructInfo
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if entering {
				return ast.WalkContinue
			}

			if nil == n.FirstChild {
				return ast.WalkSkipChildren
			}
			infoList = append(infoList, structAll.StructInfo{
				BlockID:   n.IALAttr("id"),
				BlockType: n.Type.String(),
				Node:      n,
			})
			return ast.WalkContinue
		})
		return infoList, tree
	}

	// StructList 解析后的所有对象
	var StructList []structAll.FileEntity
	StructInfoMap := make(map[string]*structAll.StructInfo)
	// FindFileEntityFromID 通过id找到对应的数据
	FindFileEntityFromID := func(id string) (structAll.FileEntity, structAll.StructInfo, error) {
		var fileEntity structAll.FileEntity
		var mdInfo structAll.StructInfo
		info, ok := StructInfoMap[id]
		if ok {
			return *info.FileEntity, *info, nil
		} else {
			var msg = "未找到id " + id + " 对应的fileEntity"
			return fileEntity, mdInfo, errors.New(msg)
		}
	}

	FileEntityToHTML := Generate(db, FindFileEntityFromID, structToHTML)

	// FileToFileEntity 通过文件路径以及文件信息获取他的结构信息
	FileToFileEntity := func(sourceDir string, path string, info os.FileInfo) structAll.FileEntity {
		relativePath := strings.ReplaceAll(path[len(sourceDir):], string(os.PathSeparator), "/")
		var virtualPath string
		var notesCode string
		var name string
		var StructInfo []structAll.StructInfo
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
		entity := structAll.FileEntity{
			Path:           path,
			Info:           info,
			RelativePath:   relativePath,
			VirtualPath:    virtualPath,
			NotesCode:      notesCode,
			StructInfoList: StructInfo,
			Name:           name,
			Tree:           tree,
		}
		for i := 0; i < len(entity.StructInfoList); i++ {
			entity.StructInfoList[i].FileEntity = &entity
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
				fileEntity := FileToFileEntity(dir, p, info)
				StructList = append(StructList, fileEntity)
				return nil
			}
		})
	// 构建 StructInfoMap
	for i := 0; i < len(StructList); i++ {
		fileEntity := StructList[i]
		for j := 0; j < len(fileEntity.StructInfoList); j++ {
			info := fileEntity.StructInfoList[j]
			if info.BlockID != "" {
				StructInfoMap[info.BlockID] = &info
			}
		}
	}
	return structAll.DirToStructRes{
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
