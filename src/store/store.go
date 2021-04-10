package store

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	sqlite "github.com/2234839/md2website/src/sqlite"
	"github.com/2234839/md2website/src/util"
	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/render"
)

// MdStructInfo md 结构信息
type MdStructInfo struct {
	blockID   string
	blockType string
	mdContent string
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
	VirtualPath      string
	MdStr            string
	Info             os.FileInfo
	MdStructInfoList []MdStructInfo

	ToHTML func() string
}

// FindFileEntityFromID 通过 id 返回对应实体
type FindFileEntityFromID func(id string) (FileEntity, MdStructInfo, error)

// DirToStructRes DirToStruct 的返回值定义
type DirToStructRes struct {
	StructList           []FileEntity
	FindFileEntityFromID FindFileEntityFromID
}

// DirToStruct 从 目录 转为更可用的结构
func DirToStruct(dir string, dbPath string, structToHTML func(interface{}) string) DirToStructRes {
	db := sqlite.InitDb(dbPath)
	/** 用于从 md 文档中解析获得一些结构性信息 */
	var mdStructuredLuteEngine = lute.New()
	mdStructuredLuteEngine.SetKramdownIAL(true)
	mdStructuredLuteEngine.SetKramdownIALIDRenderName("data-block-id")

	renderBlockMarkdown := func(node *ast.Node) string {
		root := &ast.Node{Type: ast.NodeDocument}
		luteEngine := mdStructuredLuteEngine

		tree := &parse.Tree{Root: root, Context: &parse.Context{ParseOption: luteEngine.ParseOptions}}
		renderer := render.NewFormatRenderer(tree, luteEngine.RenderOptions)
		renderer.Writer = &bytes.Buffer{}
		renderer.NodeWriterStack = append(renderer.NodeWriterStack, renderer.Writer)
		ast.Walk(node, func(n *ast.Node, entering bool) ast.WalkStatus {
			rendererFunc := renderer.RendererFuncs[n.Type]
			return rendererFunc(n, entering)
		})
		return strings.TrimSpace(renderer.Writer.String())
	}

	// GetMdStructInfo 从 md 获取结构信息
	GetMdStructInfo := func(name string, md string) []MdStructInfo {
		luteEngine := mdStructuredLuteEngine
		tree := parse.Parse(name, []byte(md), luteEngine.ParseOptions)

		var infoList []MdStructInfo
		ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
			if entering {
				return ast.WalkContinue
			}

			if nil == n.FirstChild {
				return ast.WalkSkipChildren
			}
			content := renderBlockMarkdown(n)
			if strings.Contains(n.Text(), "岁，一事无成，未来还有希望吗？") {
				// 这里有一个 bug 待 lute 修复
			}
			infoList = append(infoList, MdStructInfo{
				blockID:   n.IALAttr("id"),
				blockType: n.Type.String(),
				mdContent: content,
				node:      n,
			})
			return ast.WalkContinue
		})
		return infoList
	}

	// StructList 解析后的所有对象
	var StructList []FileEntity
	// FindFileEntityFromID 通过id找到对应的数据 这里之后要改一下，用 map 会比 双重for 好一些
	FindFileEntityFromID := func(id string) (FileEntity, MdStructInfo, error) {
		var fileEntity FileEntity
		var mdInfo MdStructInfo
		for _, entity := range StructList {
			for _, info := range entity.MdStructInfoList {
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
		var mdStr string
		var name string
		var mdStructInfo []MdStructInfo
		if info.IsDir() {
			virtualPath = relativePath
		} else {
			virtualPath = FilePathToWebPath(relativePath)
			mdByte, err := ioutil.ReadFile(path)
			if err != nil {
				util.Warn("读取文件失败", err)
			}
			mdStr = string(mdByte)
			mdStructInfo = GetMdStructInfo("", mdStr)
			if strings.HasSuffix(relativePath, ".md") {
				baseName := filepath.Base(relativePath)
				name = baseName[:len(baseName)-3]
			}
		}
		entity := FileEntity{
			Path:             path,
			Info:             info,
			RelativePath:     relativePath,
			VirtualPath:      virtualPath,
			MdStr:            mdStr,
			MdStructInfoList: mdStructInfo,
			Name:             name,
		}
		entity.ToHTML = func() string {
			return FileEntityToHTML(entity)
		}
		// if strings.Contains(virtualPath, "文章分享到的") {
		// 	util.Log("debugger")
		// }
		return entity
	}

	filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if isSkipPath(path) || (!info.IsDir() && !strings.HasSuffix(path, ".md")) {
				return nil
			} else {
				StructList = append(StructList, FileToFileEntity(dir, path, info))
				return nil
			}
		})
	return DirToStructRes{
		StructList:           StructList,
		FindFileEntityFromID: FindFileEntityFromID,
	}
}
func isSkipPath(path string) bool {
	return strings.Contains(path, ".git")
}

// FilePathToWebPath 将相对文件路径转为 web路径，主要是去除文件中的id 以及添加 .html
func FilePathToWebPath(filePath string) string {
	if strings.HasSuffix(filePath, ".md") {
		return filePath[0:len(filePath)-3] + ".html"
	}
	// 大概率是空
	return filePath
}
