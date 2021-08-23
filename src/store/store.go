package store

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

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
	id, _, _ := util.FindAttr(node.KramdownIAL, "id")
	node.KramdownIAL = append(node.KramdownIAL, []string{"data-n-id", id})
}

type addKramdownIALContext struct {
	docUpdated int
}

// addAll 遍历整颗树，附加一些数据到 KramdownIAL. 目前有 id 最后更新时间
func addAll(node *ast.Node, ctx *addKramdownIALContext) {
	if node.Type == ast.NodeDocument {
		if ctx.docUpdated == 0 {
			time := util.TimeFromID(node.ID)
			updated, _ := strconv.Atoi(time)
			ctx.docUpdated = updated
		}
	}
	updatedAttr, _, err := util.FindAttr(node.KramdownIAL, "updated")
	if err == nil {
		// 获取文档最后更新时间
		updated, _ := strconv.Atoi(updatedAttr)
		if updated > ctx.docUpdated {
			ctx.docUpdated = updated
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
	structToHTML func(interface{}) (res string),
	Generate func(
		db sqlite.DbResult,
		FindFileEntityFromID structAll.FindFileEntityFromID,
		structToHTML func(interface{}) (res string),
	) func(entity structAll.FileEntity) (html string, xml string)) structAll.DirToStructRes {
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

		if err != nil {
			util.Warn("<无法打开文件>", err)
		}
		var infoList []structAll.StructInfo
		if tree != nil && tree.Root != nil {
			addKramdownIAL(tree.Root)
			ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
				if entering {
					return ast.WalkContinue
				}
				infoList = append(infoList, structAll.StructInfo{
					BlockID:   n.ID,
					BlockType: n.Type.String(),
					Node:      n,
				})

				if nil == n.FirstChild {
					return ast.WalkSkipChildren
				}
				return ast.WalkContinue
			})
		}
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
	FileToFileEntity := func(sourceDir string, path string, info os.FileInfo) (entity structAll.FileEntity, err error) {
		relativePath := filepath.ToSlash(path[len(sourceDir):])
		var notesCode string
		var name string
		var StructInfo []structAll.StructInfo
		var tree *parse.Tree
		if info.IsDir() {

		} else {
			mdByte, err := ioutil.ReadFile(path)
			if err != nil {
				util.Warn("<读取文件失败>", err)
				return entity, err
			} else {
				notesCode = string(mdByte)
				StructInfo, tree = GetStructInfoByNotesCode(notesCode, filepath.Ext(path))
				if util.IsNotes(relativePath) {
					name, _, err = util.FindAttr(tree.Root.KramdownIAL, "title")
					if err != nil {
						util.Warn(path + " 没有标题")
					}
				}
			}
		}
		entity = structAll.FileEntity{
			Path:           path,
			Info:           info,
			RelativePath:   relativePath,
			NotesCode:      notesCode,
			StructInfoList: StructInfo,
			Name:           name,
			Tree:           tree,
		}
		for i := 0; i < len(entity.StructInfoList); i++ {
			entity.StructInfoList[i].FileEntity = &entity
		}
		entity.Output = func() (html string, xml string) {
			return FileEntityToHTML(entity)
		}
		return entity, nil
	}

	filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			p := filepath.ToSlash(path)
			if err != nil {
				return err
			} else if util.IsSkipPath(p) || (!info.IsDir() && !util.IsNotes(p)) {
				return nil
			} else {
				fileEntity, err := FileToFileEntity(dir, p, info)
				if err == nil {
					StructList = append(StructList, fileEntity)
					return nil
				} else {
					// 由于思源锁定文件或其他程序锁定文件导致解析失败的 Entity 就不参与后面的流程了
					return nil
				}
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
