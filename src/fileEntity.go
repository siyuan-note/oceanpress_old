package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/2234839/md2website/src/util"
)

// FileEntityList 解析后的所有对象
var FileEntityList []FileEntity

// FileEntity md 文件被解析后的结构
type FileEntity struct {
	// 这个 id 是有很大可能为空的
	id   string
	name string
	// 文件绝对路径
	path string
	// 相对源目录的路径
	relativePath string
	// 最终要在浏览器中可以访问的路径
	virtualPath      string
	mdStr            string
	info             os.FileInfo
	MdStructInfoList []MdStructInfo
}

// FileToFileEntity 通过文件路径以及文件信息获取他的结构信息
func FileToFileEntity(path string, info os.FileInfo) FileEntity {
	sourceDir := SourceDir
	relativePath := strings.ReplaceAll(path[len(sourceDir):], string(os.PathSeparator), "/")
	var virtualPath string
	var mdStr string
	var name string
	var id string
	var mdStructInfo []MdStructInfo
	if info.IsDir() {
		virtualPath = relativePath
	} else {
		virtualPath = FilePathToWebPath(relativePath)
		mdByte, err := ioutil.ReadFile(path)
		if err != nil {
			util.Log("读取文件失败", err)
		}
		mdStr = string(mdByte)
		mdStructInfo = GetMdStructInfo("", mdStr)
		if strings.HasSuffix(relativePath, ".sy.md") {
			baseName := filepath.Base(relativePath)
			name = baseName[:len(baseName)-29]
			id = baseName[len(baseName)-28 : len(baseName)-6]
		}
	}

	return FileEntity{
		path:             path,
		info:             info,
		relativePath:     relativePath,
		virtualPath:      virtualPath,
		mdStr:            mdStr,
		MdStructInfoList: mdStructInfo,
		id:               id,
		name:             name,
	}
}

// FindFileEntityFromID 通过id找到对应的数据
func FindFileEntityFromID(id string) (FileEntity, MdStructInfo) {
	var fileEntity FileEntity
	var mdInfo MdStructInfo
	for _, entity := range FileEntityList {
		if entity.id == id {
			fileEntity = entity
			break
		}
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
		if fileEntity.path != "" {
			break
		}
	}
	if fileEntity.path == "" {
		test := FileEntityList
		util.Log("未找到对应fileEntity", id, len(test))
	}
	return fileEntity, mdInfo
}