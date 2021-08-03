package store

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/88250/lute"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/lex"
	"github.com/88250/lute/parse"
	luteUtil "github.com/88250/lute/util"
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
		entity.Output = func() (html string, xml string) {
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

var openCurlyBrace = []byte("{")
var closeCurlyBrace = []byte("}")

// 解析 KramdownSpan 以下四个函数代码来自 github.com\88250\lute\parse\inline_attribute_list.go
func parseKramdownSpanIAL(tokens []byte) (pos int, ret [][]string) {
	pos = bytes.Index(tokens, closeCurlyBrace)
	if curlyBracesStart := bytes.Index(tokens, []byte("{:")); 0 == curlyBracesStart && curlyBracesStart+2 < pos {
		tokens = tokens[curlyBracesStart+2:]
		curlyBracesEnd := bytes.Index(tokens, closeCurlyBrace)
		if 3 > curlyBracesEnd {
			return
		}

		tokens = tokens[:curlyBracesEnd]
		for {
			valid, remains, attr, name, val := TagAttr(tokens)
			if !valid {
				break
			}

			tokens = remains
			if 1 > len(attr) {
				break
			}

			nameStr := strings.ReplaceAll(string(name), luteUtil.Caret, "")
			valStr := strings.ReplaceAll(string(val), luteUtil.Caret, "")
			ret = append(ret, []string{nameStr, valStr})
		}
	}
	return
}

func TagAttr(tokens []byte) (valid bool, remains, attr, name, val []byte) {
	valid = true
	remains = tokens
	var whitespaces []byte
	var i int
	var token byte
	for i, token = range tokens {
		if !lex.IsWhitespace(token) {
			break
		}
		whitespaces = append(whitespaces, token)
	}
	if 1 > len(whitespaces) {
		return
	}
	tokens = tokens[i:]

	var attrName []byte
	tokens, attrName = parseAttrName(tokens)
	if 1 > len(attrName) {
		return
	}

	var valSpec []byte
	valid, tokens, valSpec = parseAttrValSpec(tokens)
	if !valid {
		return
	}

	remains = tokens
	attr = append(attr, whitespaces...)
	attr = append(attr, attrName...)
	attr = append(attr, valSpec...)
	if nil != valSpec {
		name = attrName
		val = valSpec[2 : len(valSpec)-1]
	}
	return
}
func parseAttrName(tokens []byte) (remains, attrName []byte) {
	remains = tokens
	if !lex.IsASCIILetter(tokens[0]) && lex.ItemUnderscore != tokens[0] && lex.ItemColon != tokens[0] {
		return
	}
	attrName = append(attrName, tokens[0])
	tokens = tokens[1:]
	var i int
	var token byte
	for i, token = range tokens {
		if !lex.IsASCIILetterNumHyphen(token) && lex.ItemUnderscore != token && lex.ItemDot != token && lex.ItemColon != token {
			break
		}
		attrName = append(attrName, token)
	}
	if 1 > len(attrName) {
		return
	}

	remains = tokens[i:]
	return
}
func parseAttrValSpec(tokens []byte) (valid bool, remains, valSpec []byte) {
	valid = true
	remains = tokens
	var i int
	var token byte
	for i, token = range tokens {
		if !lex.IsWhitespace(token) {
			break
		}
		valSpec = append(valSpec, token)
	}
	if lex.ItemEqual != token {
		valSpec = nil
		return
	}
	valSpec = append(valSpec, token)
	tokens = tokens[i+1:]
	if 1 > len(tokens) {
		valid = false
		return
	}

	for i, token = range tokens {
		if !lex.IsWhitespace(token) {
			break
		}
		valSpec = append(valSpec, token)
	}
	token = tokens[i]
	valSpec = append(valSpec, token)
	tokens = tokens[i+1:]
	closed := false
	if lex.ItemDoublequote == token { // A double-quoted attribute value consists of ", zero or more characters not including ", and a final ".
		for i, token = range tokens {
			valSpec = append(valSpec, token)
			if lex.ItemDoublequote == token {
				closed = true
				break
			}
		}
	} else if lex.ItemSinglequote == token { // A single-quoted attribute value consists of ', zero or more characters not including ', and a final '.
		for i, token = range tokens {
			valSpec = append(valSpec, token)
			if lex.ItemSinglequote == token {
				closed = true
				break
			}
		}
	} else { // An unquoted attribute value is a nonempty string of characters not including whitespace, ", ', =, <, >, or `.
		for i, token = range tokens {
			if lex.ItemGreater == token {
				i-- // 大于字符 > 不计入 valSpec
				break
			}
			valSpec = append(valSpec, token)
			if lex.IsWhitespace(token) {
				// 属性使用空白分隔
				break
			}
			if lex.ItemDoublequote == token || lex.ItemSinglequote == token || lex.ItemEqual == token || lex.ItemLess == token || lex.ItemGreater == token || lex.ItemBacktick == token {
				closed = false
				break
			}
			closed = true
		}
	}

	if !closed {
		valid = false
		valSpec = nil
		return
	}

	remains = tokens[i+1:]
	return
}
