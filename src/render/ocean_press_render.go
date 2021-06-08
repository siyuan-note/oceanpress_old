// Lute - 一款结构化的 Markdown 引擎，支持 Go 和 JavaScript
// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
// 原始地址 https://github.com/88250/lute/blob/master/render/protyle_preview_renderer.go
package render

import (
	"bytes"
	"errors"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/siyuan-note/oceanpress/src/sqlite"
	structAll "github.com/siyuan-note/oceanpress/src/struct"
	oceanUtil "github.com/siyuan-note/oceanpress/src/util"

	"github.com/88250/lute"
	"github.com/88250/lute/html"

	"github.com/88250/lute/ast"
	"github.com/88250/lute/lex"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/util"
)

type Context struct {
	Db                   sqlite.DbResult
	FindFileEntityFromID func(id string) (structAll.FileEntity, structAll.StructInfo, error)
	StructToHTML         func(interface{}) string

	RefID   string
	idStack *[]string
	push    func(id string) error
	pop     func(id string)

	// 第一层级的引用
	TopRefId *[]string

	BaseEntity structAll.FileEntity
	LuteEngine *lute.Lute
}
type OceanPressRender struct {
	*BaseRenderer
	context *Context
}

func NewOceanPressRenderer(tree *parse.Tree, options *Options,
	context *Context,
) *OceanPressRender {
	// 嵌入块的 id
	refID := context.BaseEntity.Tree.ID
	if context.idStack == nil {
		context.idStack = &[]string{}
	}
	if context.TopRefId == nil {
		context.TopRefId = &[]string{}
	}
	// 或许应该在一个 render 之后pop掉当前push的
	push := func(id string) error {
		if id == "" {
			return nil
		}
		count := 0
		for _, item := range *context.idStack {
			if item == id {
				count++
			}
		}
		if count >= 2 {
			oceanUtil.Warn("<循环引用>", id)
			return errors.New("循环引用")
		}
		t := append(*context.idStack, id)
		context.idStack = &t
		return nil
	}
	pop := func(id string) {
		stack := *context.idStack
		t := stack[:len(*context.idStack)-1]
		context.idStack = &t
	}
	context.pop = pop
	context.push = push

	context.RefID = refID

	ret := &OceanPressRender{NewBaseRenderer(tree, options), context}
	ret.RendererFuncs[ast.NodeDocument] = ret.renderDocument
	ret.RendererFuncs[ast.NodeParagraph] = ret.renderParagraph
	ret.RendererFuncs[ast.NodeText] = ret.renderText
	ret.RendererFuncs[ast.NodeCodeSpan] = ret.renderCodeSpan
	ret.RendererFuncs[ast.NodeCodeSpanOpenMarker] = ret.renderCodeSpanOpenMarker
	ret.RendererFuncs[ast.NodeCodeSpanContent] = ret.renderCodeSpanContent
	ret.RendererFuncs[ast.NodeCodeSpanCloseMarker] = ret.renderCodeSpanCloseMarker
	ret.RendererFuncs[ast.NodeCodeBlock] = ret.renderCodeBlock
	ret.RendererFuncs[ast.NodeCodeBlockFenceOpenMarker] = ret.renderCodeBlockOpenMarker
	ret.RendererFuncs[ast.NodeCodeBlockFenceInfoMarker] = ret.renderCodeBlockInfoMarker
	ret.RendererFuncs[ast.NodeCodeBlockCode] = ret.renderCodeBlockCode
	ret.RendererFuncs[ast.NodeCodeBlockFenceCloseMarker] = ret.renderCodeBlockCloseMarker
	ret.RendererFuncs[ast.NodeMathBlock] = ret.renderMathBlock
	ret.RendererFuncs[ast.NodeMathBlockOpenMarker] = ret.renderMathBlockOpenMarker
	ret.RendererFuncs[ast.NodeMathBlockContent] = ret.renderMathBlockContent
	ret.RendererFuncs[ast.NodeMathBlockCloseMarker] = ret.renderMathBlockCloseMarker
	ret.RendererFuncs[ast.NodeInlineMath] = ret.renderInlineMath
	ret.RendererFuncs[ast.NodeInlineMathOpenMarker] = ret.renderInlineMathOpenMarker
	ret.RendererFuncs[ast.NodeInlineMathContent] = ret.renderInlineMathContent
	ret.RendererFuncs[ast.NodeInlineMathCloseMarker] = ret.renderInlineMathCloseMarker
	ret.RendererFuncs[ast.NodeEmphasis] = ret.renderEmphasis
	ret.RendererFuncs[ast.NodeEmA6kOpenMarker] = ret.renderEmAsteriskOpenMarker
	ret.RendererFuncs[ast.NodeEmA6kCloseMarker] = ret.renderEmAsteriskCloseMarker
	ret.RendererFuncs[ast.NodeEmU8eOpenMarker] = ret.renderEmUnderscoreOpenMarker
	ret.RendererFuncs[ast.NodeEmU8eCloseMarker] = ret.renderEmUnderscoreCloseMarker
	ret.RendererFuncs[ast.NodeStrong] = ret.renderStrong
	ret.RendererFuncs[ast.NodeStrongA6kOpenMarker] = ret.renderStrongA6kOpenMarker
	ret.RendererFuncs[ast.NodeStrongA6kCloseMarker] = ret.renderStrongA6kCloseMarker
	ret.RendererFuncs[ast.NodeStrongU8eOpenMarker] = ret.renderStrongU8eOpenMarker
	ret.RendererFuncs[ast.NodeStrongU8eCloseMarker] = ret.renderStrongU8eCloseMarker
	ret.RendererFuncs[ast.NodeBlockquote] = ret.renderBlockquote
	ret.RendererFuncs[ast.NodeBlockquoteMarker] = ret.renderBlockquoteMarker
	ret.RendererFuncs[ast.NodeHeading] = ret.renderHeading
	ret.RendererFuncs[ast.NodeHeadingC8hMarker] = ret.renderHeadingC8hMarker
	ret.RendererFuncs[ast.NodeHeadingID] = ret.renderHeadingID
	ret.RendererFuncs[ast.NodeList] = ret.renderList
	ret.RendererFuncs[ast.NodeListItem] = ret.renderListItem
	ret.RendererFuncs[ast.NodeThematicBreak] = ret.renderThematicBreak
	ret.RendererFuncs[ast.NodeHardBreak] = ret.renderHardBreak
	ret.RendererFuncs[ast.NodeSoftBreak] = ret.renderSoftBreak
	ret.RendererFuncs[ast.NodeHTMLBlock] = ret.renderHTML
	ret.RendererFuncs[ast.NodeInlineHTML] = ret.renderInlineHTML
	ret.RendererFuncs[ast.NodeLink] = ret.renderLink
	ret.RendererFuncs[ast.NodeImage] = ret.renderImage
	ret.RendererFuncs[ast.NodeBang] = ret.renderBang
	ret.RendererFuncs[ast.NodeOpenBracket] = ret.renderOpenBracket
	ret.RendererFuncs[ast.NodeCloseBracket] = ret.renderCloseBracket
	ret.RendererFuncs[ast.NodeOpenParen] = ret.renderOpenParen
	ret.RendererFuncs[ast.NodeCloseParen] = ret.renderCloseParen
	ret.RendererFuncs[ast.NodeOpenBrace] = ret.renderOpenBrace
	ret.RendererFuncs[ast.NodeCloseBrace] = ret.renderCloseBrace
	ret.RendererFuncs[ast.NodeLinkText] = ret.renderLinkText
	ret.RendererFuncs[ast.NodeLinkSpace] = ret.renderLinkSpace
	ret.RendererFuncs[ast.NodeLinkDest] = ret.renderLinkDest
	ret.RendererFuncs[ast.NodeLinkTitle] = ret.renderLinkTitle
	ret.RendererFuncs[ast.NodeStrikethrough] = ret.renderStrikethrough
	ret.RendererFuncs[ast.NodeStrikethrough1OpenMarker] = ret.renderStrikethrough1OpenMarker
	ret.RendererFuncs[ast.NodeStrikethrough1CloseMarker] = ret.renderStrikethrough1CloseMarker
	ret.RendererFuncs[ast.NodeStrikethrough2OpenMarker] = ret.renderStrikethrough2OpenMarker
	ret.RendererFuncs[ast.NodeStrikethrough2CloseMarker] = ret.renderStrikethrough2CloseMarker
	ret.RendererFuncs[ast.NodeTaskListItemMarker] = ret.renderTaskListItemMarker
	ret.RendererFuncs[ast.NodeTable] = ret.renderTable
	ret.RendererFuncs[ast.NodeTableHead] = ret.renderTableHead
	ret.RendererFuncs[ast.NodeTableRow] = ret.renderTableRow
	ret.RendererFuncs[ast.NodeTableCell] = ret.renderTableCell
	ret.RendererFuncs[ast.NodeEmoji] = ret.renderEmoji
	ret.RendererFuncs[ast.NodeEmojiUnicode] = ret.renderEmojiUnicode
	ret.RendererFuncs[ast.NodeEmojiImg] = ret.renderEmojiImg
	ret.RendererFuncs[ast.NodeEmojiAlias] = ret.renderEmojiAlias
	ret.RendererFuncs[ast.NodeFootnotesDefBlock] = ret.renderFootnotesDefBlock
	ret.RendererFuncs[ast.NodeFootnotesDef] = ret.renderFootnotesDef
	ret.RendererFuncs[ast.NodeFootnotesRef] = ret.renderFootnotesRef
	ret.RendererFuncs[ast.NodeToC] = ret.renderToC
	ret.RendererFuncs[ast.NodeBackslash] = ret.renderBackslash
	ret.RendererFuncs[ast.NodeBackslashContent] = ret.renderBackslashContent
	ret.RendererFuncs[ast.NodeHTMLEntity] = ret.renderHtmlEntity
	ret.RendererFuncs[ast.NodeYamlFrontMatter] = ret.renderYamlFrontMatter
	ret.RendererFuncs[ast.NodeYamlFrontMatterOpenMarker] = ret.renderYamlFrontMatterOpenMarker
	ret.RendererFuncs[ast.NodeYamlFrontMatterContent] = ret.renderYamlFrontMatterContent
	ret.RendererFuncs[ast.NodeYamlFrontMatterCloseMarker] = ret.renderYamlFrontMatterCloseMarker
	ret.RendererFuncs[ast.NodeBlockRef] = ret.renderBlockRef
	ret.RendererFuncs[ast.NodeBlockRefID] = ret.renderBlockRefID
	ret.RendererFuncs[ast.NodeBlockRefSpace] = ret.renderBlockRefSpace
	ret.RendererFuncs[ast.NodeBlockRefText] = ret.renderBlockRefText
	ret.RendererFuncs[ast.NodeMark] = ret.renderMark
	ret.RendererFuncs[ast.NodeMark1OpenMarker] = ret.renderMark1OpenMarker
	ret.RendererFuncs[ast.NodeMark1CloseMarker] = ret.renderMark1CloseMarker
	ret.RendererFuncs[ast.NodeMark2OpenMarker] = ret.renderMark2OpenMarker
	ret.RendererFuncs[ast.NodeMark2CloseMarker] = ret.renderMark2CloseMarker
	ret.RendererFuncs[ast.NodeSup] = ret.renderSup
	ret.RendererFuncs[ast.NodeSupOpenMarker] = ret.renderSupOpenMarker
	ret.RendererFuncs[ast.NodeSupCloseMarker] = ret.renderSupCloseMarker
	ret.RendererFuncs[ast.NodeSub] = ret.renderSub
	ret.RendererFuncs[ast.NodeSubOpenMarker] = ret.renderSubOpenMarker
	ret.RendererFuncs[ast.NodeSubCloseMarker] = ret.renderSubCloseMarker
	ret.RendererFuncs[ast.NodeKramdownBlockIAL] = ret.renderKramdownBlockIAL
	ret.RendererFuncs[ast.NodeKramdownSpanIAL] = ret.renderKramdownSpanIAL
	ret.RendererFuncs[ast.NodeBlockQueryEmbed] = ret.renderBlockQueryEmbed
	ret.RendererFuncs[ast.NodeBlockQueryEmbedScript] = ret.renderBlockQueryEmbedScript
	ret.RendererFuncs[ast.NodeBlockEmbed] = ret.renderBlockEmbed
	ret.RendererFuncs[ast.NodeBlockEmbedID] = ret.renderBlockEmbedID
	ret.RendererFuncs[ast.NodeBlockEmbedSpace] = ret.renderBlockEmbedSpace
	ret.RendererFuncs[ast.NodeBlockEmbedText] = ret.renderBlockEmbedText
	ret.RendererFuncs[ast.NodeTag] = ret.renderTag
	ret.RendererFuncs[ast.NodeTagOpenMarker] = ret.renderTagOpenMarker
	ret.RendererFuncs[ast.NodeTagCloseMarker] = ret.renderTagCloseMarker
	ret.RendererFuncs[ast.NodeLinkRefDefBlock] = ret.renderLinkRefDefBlock
	ret.RendererFuncs[ast.NodeLinkRefDef] = ret.renderLinkRefDef
	ret.RendererFuncs[ast.NodeSuperBlock] = ret.renderSuperBlock
	ret.RendererFuncs[ast.NodeSuperBlockOpenMarker] = ret.renderSuperBlockOpenMarker
	ret.RendererFuncs[ast.NodeSuperBlockLayoutMarker] = ret.renderSuperBlockLayoutMarker
	ret.RendererFuncs[ast.NodeSuperBlockCloseMarker] = ret.renderSuperBlockCloseMarker
	ret.RendererFuncs[ast.NodeGitConflict] = ret.renderGitConflict
	ret.RendererFuncs[ast.NodeGitConflictOpenMarker] = ret.renderGitConflictOpenMarker
	ret.RendererFuncs[ast.NodeGitConflictContent] = ret.renderGitConflictContent
	ret.RendererFuncs[ast.NodeGitConflictCloseMarker] = ret.renderGitConflictCloseMarker
	ret.RendererFuncs[ast.NodeIFrame] = ret.renderIFrame
	ret.RendererFuncs[ast.NodeVideo] = ret.renderVideo
	ret.RendererFuncs[ast.NodeAudio] = ret.renderAudio
	ret.RendererFuncs[ast.NodeKbd] = ret.renderKbd
	ret.RendererFuncs[ast.NodeKbdOpenMarker] = ret.renderKbdOpenMarker
	ret.RendererFuncs[ast.NodeKbdCloseMarker] = ret.renderKbdCloseMarker
	ret.RendererFuncs[ast.NodeUnderline] = ret.renderUnderline
	ret.RendererFuncs[ast.NodeUnderlineOpenMarker] = ret.renderUnderlineOpenMarker
	ret.RendererFuncs[ast.NodeUnderlineCloseMarker] = ret.renderUnderlineCloseMarker
	ret.RendererFuncs[ast.NodeBr] = ret.renderBr
	return ret
}

func (r *OceanPressRender) renderBr(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteString("<br />")
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderUnderline(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderUnderlineOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteString("<u>")
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderUnderlineCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteString("</u>")
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderKbd(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderKbdOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteString("<kbd>")
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderKbdCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteString("</kbd>")
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderVideo(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("div", [][]string{{"class", "iframe"}}, false)
		tokens := node.Tokens
		if r.Options.Sanitize {
			tokens = sanitize(tokens)
		}
		tokens = r.tagSrcPath(tokens)
		r.Write(tokens)
		r.Tag("/div", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderAudio(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("div", [][]string{{"class", "iframe"}}, false)
		tokens := node.Tokens
		if r.Options.Sanitize {
			tokens = sanitize(tokens)
		}
		tokens = r.tagSrcPath(tokens)
		r.Write(tokens)
		r.Tag("/div", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderIFrame(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("div", [][]string{{"class", "iframe"}}, false)
		tokens := node.Tokens
		if r.Options.Sanitize {
			tokens = sanitize(tokens)
		}
		tokens = r.tagSrcPath(tokens)
		r.Write(tokens)
		r.Tag("/div", nil, false)
	}
	return ast.WalkContinue
}

// func (r *OceanPressRender) Render() (output []byte) {
// 	output = r.BaseRenderer.Render()
// 	output = append(output, r.RenderFootnotes()...)
// 	return
// }

func (r *OceanPressRender) renderGitConflictCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(node.Tokens)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderGitConflictContent(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(html.EscapeHTML(node.Tokens))
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderGitConflictOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(node.Tokens)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderGitConflict(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	if entering {
		attrs := [][]string{{"class", "language-git-conflict"}}
		r.handleKramdownBlockIAL(node)
		attrs = append(attrs, node.KramdownIAL...)
		r.Tag("div", attrs, false)
	} else {
		r.Tag("/div", nil, false)
	}
	return ast.WalkContinue
}

// func (r *OceanPressRender) renderSuperBlock(node *ast.Node, entering bool) ast.WalkStatus {
// 	return ast.WalkContinue
// }

func (r *OceanPressRender) renderSuperBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkSkipChildren
}

func (r *OceanPressRender) renderSuperBlockLayoutMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkSkipChildren
}

func (r *OceanPressRender) renderSuperBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkSkipChildren
}

func (r *OceanPressRender) renderLinkRefDefBlock(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkSkipChildren
}

func (r *OceanPressRender) renderLinkRefDef(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkSkipChildren
}

func (r *OceanPressRender) renderTag(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.TextAutoSpacePrevious(node)
	} else {
		r.TextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderTagOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("em", node.Parent.KramdownIAL, false)
		r.WriteByte(lex.ItemCrosshatch)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderTagCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteByte(lex.ItemCrosshatch)
		r.Tag("/em", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderKramdownBlockIAL(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderKramdownSpanIAL(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMark(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.TextAutoSpacePrevious(node)
	} else {
		r.TextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMark1OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("mark", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMark1CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/mark", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMark2OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("mark", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMark2CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/mark", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderSup(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderSupOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("sup", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderSupCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/sup", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderSub(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderSubOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("sub", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderSubCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/sub", nil, false)
	}
	return ast.WalkContinue
}

// func (r *OceanPressRender) renderBlockQueryEmbed(node *ast.Node, entering bool) ast.WalkStatus {
// 	if entering {
// 		r.Newline()
// 		r.Tag("div", nil, false)
// 	} else {
// 		r.Tag("/div", nil, false)
// 		r.Newline()
// 	}
// 	return ast.WalkContinue
// }

func (r *OceanPressRender) renderBlockQueryEmbedScript(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteByte(lex.ItemDoublequote)
		r.Write(node.Tokens)
		r.WriteByte(lex.ItemDoublequote)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBlockEmbed(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.handleKramdownBlockIAL(node)
		r.Tag("div", node.KramdownIAL, false)
	} else {
		r.Tag("/div", nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBlockEmbedID(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBlockEmbedSpace(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBlockEmbedText(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteByte(lex.ItemDoublequote)
		r.Write(html.EscapeHTML(node.Tokens))
		r.WriteByte(lex.ItemDoublequote)
	}
	return ast.WalkContinue
}

// func (r *OceanPressRender) renderBlockRef(node *ast.Node, entering bool) ast.WalkStatus {
// 	return ast.WalkContinue
// }

func (r *OceanPressRender) renderBlockRefID(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBlockRefSpace(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBlockRefText(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteByte(lex.ItemDoublequote)
		r.Write(node.Tokens)
	} else {
		r.WriteByte(lex.ItemDoublequote)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderYamlFrontMatterCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteString("</code></pre>")
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderYamlFrontMatterContent(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(html.EscapeHTML(node.Tokens))
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderYamlFrontMatterOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		attrs := [][]string{{"class", "vditor-yml-front-matter"}}
		attrs = append(attrs, node.Parent.KramdownIAL...)
		r.Tag("pre", attrs, false)
		r.WriteString("<code class=\"language-yaml\">")
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderYamlFrontMatter(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	return ast.WalkContinue
}

func (r *OceanPressRender) renderHtmlEntity(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(html.EscapeHTML(node.Tokens))
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBackslashContent(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(html.EscapeHTML(node.Tokens))
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBackslash(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderToC(node *ast.Node, entering bool) ast.WalkStatus {
	return r.BaseRenderer.renderToC(node, entering)
}

func (r *OceanPressRender) renderFootnotesRef(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		idx, _ := r.Tree.FindFootnotesDef(node.Tokens)
		idxStr := strconv.Itoa(idx)
		r.Tag("sup", [][]string{{"class", "footnotes-ref"}, {"id", "footnotes-ref-" + node.FootnotesRefId}}, false)
		r.Tag("a", [][]string{{"href", "#footnotes-def-" + idxStr}}, false)
		r.WriteString(idxStr)
		r.Tag("/a", nil, false)
		r.Tag("/sup", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderFootnotesDefBlock(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) RenderFootnotes() []byte {
	if 1 > len(r.FootnotesDefs) {
		return nil
	}

	buf := bytes.Buffer{}
	buf.WriteString("<div class=\"footnotes-defs-div\">")
	buf.WriteString("<hr class=\"footnotes-defs-hr\" />\n")
	buf.WriteString("<ol class=\"footnotes-defs-ol\">")
	for i, def := range r.FootnotesDefs {
		buf.WriteString("<li id=\"footnotes-def-" + strconv.Itoa(i+1) + "\">")
		footnotesTree := &parse.Tree{Name: "", Context: r.Tree.Context}
		footnotesTree.Context.Tree = footnotesTree
		footnotesTree.Root = &ast.Node{Type: ast.NodeDocument}
		footnotesTree.Root.AppendChild(def)
		defRenderer := NewOceanPressRenderer(footnotesTree, r.Options, r.context)
		lc := footnotesTree.Root.LastDeepestChild()
		for i = len(def.FootnotesRefs) - 1; 0 <= i; i-- {
			ref := def.FootnotesRefs[i]
			gotoRef := " <a href=\"#footnotes-ref-" + ref.FootnotesRefId + "\" class=\"vditor-footnotes__goto-ref\">↩</a>"
			link := &ast.Node{Type: ast.NodeInlineHTML, Tokens: util.StrToBytes(gotoRef)}
			lc.InsertAfter(link)
		}
		defRenderer.RenderingFootnotes = true
		defContent, _ := defRenderer.Render()
		buf.Write([]byte(defContent))
		buf.WriteString("</li>\n")
	}
	buf.WriteString("</ol></div>")
	return buf.Bytes()
}

func (r *OceanPressRender) renderFootnotesDef(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if !r.RenderingFootnotes {
			var found bool
			for _, n := range r.FootnotesDefs {
				if bytes.EqualFold(node.Tokens, n.Tokens) {
					found = true
					break
				}
			}
			if !found {
				r.FootnotesDefs = append(r.FootnotesDefs, node)
			}
			return ast.WalkSkipChildren
		}
	}
	return ast.WalkContinue
}

// func (r *OceanPressRender) renderCodeBlock(node *ast.Node, entering bool) ast.WalkStatus {
// 	r.Newline()

// 	noHighlight := false
// 	var language string
// 	if nil != node.FirstChild.Next && 0 < len(node.FirstChild.Next.CodeBlockInfo) {
// 		language = util.BytesToStr(node.FirstChild.Next.CodeBlockInfo)
// 		noHighlight = r.NoHighlight(language)
// 	}

// 	if entering {
// 		if noHighlight {
// 			var attrs [][]string
// 			tokens := html.EscapeHTML(node.FirstChild.Next.Next.Tokens)
// 			tokens = bytes.ReplaceAll(tokens, util.CaretTokens, nil)
// 			tokens = bytes.TrimSpace(tokens)
// 			attrs = append(attrs, []string{"data-content", util.BytesToStr(tokens)})
// 			attrs = append(attrs, []string{"data-subtype", language})
// 			r.Tag("div", attrs, false)
// 			r.Tag("div", [][]string{{"spin", "1"}}, false)
// 			r.Tag("/div", nil, false)
// 			r.Tag("/div", nil, false)
// 			return ast.WalkSkipChildren
// 		}

// 		attrs := [][]string{{"class", "code-block"}, {"data-language", language}}
// 		r.Tag("pre", attrs, false)
// 		r.WriteString("<code>")
// 	} else {
// 		if noHighlight {
// 			return ast.WalkSkipChildren
// 		}

// 		r.Tag("/code", nil, false)
// 		r.Tag("/pre", nil, false)
// 	}
// 	return ast.WalkContinue
// }

func (r *OceanPressRender) renderCodeBlockCode(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(html.EscapeHTML(node.Tokens))
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCodeBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCodeBlockInfoMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCodeBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmojiAlias(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmojiImg(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmojiUnicode(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(node.Tokens)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmoji(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderInlineMathCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/span", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderInlineMathContent(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderInlineMathOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		tokens := html.EscapeHTML(node.Next.Tokens)
		r.Tag("span", [][]string{{"data-subtype", "math"}, {"data-content", util.BytesToStr(tokens)}}, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderInlineMath(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMathBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMathBlockContent(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMathBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderMathBlock(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	if entering {
		var attrs [][]string
		tokens := html.EscapeHTML(node.FirstChild.Next.Tokens)
		tokens = bytes.ReplaceAll(tokens, util.CaretTokens, nil)
		tokens = bytes.TrimSpace(tokens)
		attrs = append(attrs, []string{"data-content", util.BytesToStr(tokens)})
		attrs = append(attrs, []string{"data-subtype", "math"})
		r.Tag("div", attrs, false)
		r.Tag("div", [][]string{{"spin", "1"}}, false)
		r.Tag("/div", nil, false)
		r.Tag("/div", nil, false)
		r.Newline()
	}
	return ast.WalkSkipChildren
}

func (r *OceanPressRender) renderTableCell(node *ast.Node, entering bool) ast.WalkStatus {
	tag := "td"
	if ast.NodeTableHead == node.Parent.Parent.Type {
		tag = "th"
	}
	if entering {
		var attrs [][]string
		switch node.TableCellAlign {
		case 1:
			attrs = append(attrs, []string{"align", "left"})
		case 2:
			attrs = append(attrs, []string{"align", "center"})
		case 3:
			attrs = append(attrs, []string{"align", "right"})
		}
		r.Tag(tag, attrs, false)
	} else {
		r.Tag("/"+tag, nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderTableRow(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("tr", nil, false)
		r.Newline()
	} else {
		r.Tag("/tr", nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderTableHead(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("thead", nil, false)
		r.Newline()
	} else {
		r.Tag("/thead", nil, false)
		r.Newline()
		if nil != node.Next {
			r.Tag("tbody", nil, false)
		}
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderTable(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.handleKramdownBlockIAL(node)
		r.Tag("table", node.KramdownIAL, false)
		r.Newline()
	} else {
		if nil != node.FirstChild.Next {
			r.Tag("/tbody", nil, false)
		}
		r.Newline()
		r.Tag("/table", nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrikethrough(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.TextAutoSpacePrevious(node)
	} else {
		r.TextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrikethrough1OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("del", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrikethrough1CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/del", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrikethrough2OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("del", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrikethrough2CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/del", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderLinkTitle(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderLinkDest(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderLinkSpace(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderLinkText(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		var tokens []byte
		if r.Options.AutoSpace {
			tokens = r.Space(node.Tokens)
		} else {
			tokens = node.Tokens
		}
		r.Write(html.EscapeHTML(tokens))
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCloseBrace(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderOpenBrace(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCloseParen(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderOpenParen(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCloseBracket(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderOpenBracket(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBang(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderImage(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if 0 == r.DisableTags {
			attrs := [][]string{{"class", "img"}}
			if style := node.IALAttr("parent-style"); "" != style {
				attrs = append(attrs, []string{"style", style})
			}
			r.Tag("span", attrs, false)
			r.WriteString("<img src=\"")
			destTokens := node.ChildByType(ast.NodeLinkDest).Tokens
			destTokens = r.LinkPath(destTokens)
			if "" != r.Options.ImageLazyLoading {
				r.Write(html.EscapeHTML(util.StrToBytes(r.Options.ImageLazyLoading)))
				r.WriteString("\" data-src=\"")
			}
			r.Write(html.EscapeHTML(destTokens))
			r.WriteString("\" alt=\"")
		}
		r.DisableTags++
		return ast.WalkContinue
	}

	r.DisableTags--
	if 0 == r.DisableTags {
		r.WriteByte(lex.ItemDoublequote)
		title := node.ChildByType(ast.NodeLinkTitle)
		var titleTokens []byte
		if nil != title && nil != title.Tokens {
			titleTokens = html.EscapeHTML(title.Tokens)
			r.WriteString(" title=\"")
			r.Write(titleTokens)
			r.WriteByte(lex.ItemDoublequote)
		}
		ial := r.NodeAttrsStr(node)
		if "" != ial {
			r.WriteString(" " + ial)
		}
		r.WriteString(" />")
		if 0 < len(titleTokens) {
			r.Tag("span", [][]string{{"class", "protyle-action__title"}}, false)
			r.Write(titleTokens)
			r.Tag("/span", nil, false)
		}
		r.Tag("/span", nil, false)

		if r.Options.Sanitize {
			buf := r.Writer.Bytes()
			idx := bytes.LastIndex(buf, []byte("<img src="))
			imgBuf := buf[idx:]
			if r.Options.Sanitize {
				imgBuf = sanitize(imgBuf)
			}
			r.Writer.Truncate(idx)
			r.Writer.Write(imgBuf)
		}
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderLink(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.LinkTextAutoSpacePrevious(node)

		dest := node.ChildByType(ast.NodeLinkDest)
		destTokens := dest.Tokens
		destTokens = r.LinkPath(destTokens)
		attrs := [][]string{{"href", util.BytesToStr(html.EscapeHTML(destTokens))}}
		if title := node.ChildByType(ast.NodeLinkTitle); nil != title && nil != title.Tokens {
			attrs = append(attrs, []string{"title", util.BytesToStr(html.EscapeHTML(title.Tokens))})
		}
		r.Tag("a", attrs, false)
	} else {
		r.Tag("/a", nil, false)

		r.LinkTextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderHTML(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		tokens := node.Tokens
		if r.Options.Sanitize {
			tokens = sanitize(tokens)
		}
		tokens = r.tagSrcPath(tokens)
		r.Write(tokens)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderInlineHTML(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		tokens := node.Tokens
		if r.Options.Sanitize {
			tokens = sanitize(tokens)
		}
		r.Write(tokens)
	}
	return ast.WalkContinue
}

// func (r *OceanPressRender) renderDocument(node *ast.Node, entering bool) ast.WalkStatus {
// 	return ast.WalkContinue
// }

func (r *OceanPressRender) renderParagraph(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.handleKramdownBlockIAL(node)
		var attrs [][]string
		attrs = append(attrs, node.KramdownIAL...)
		if r.Options.ChineseParagraphBeginningSpace && ast.NodeDocument == node.Parent.Type {
			attrs = append(attrs, []string{"class", "indent--2"})
		}
		r.Tag("p", attrs, false)
	} else {
		r.Tag("/p", nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderText(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		var tokens []byte
		if r.Options.AutoSpace {
			tokens = r.Space(node.Tokens)
		} else {
			tokens = node.Tokens
		}

		if r.Options.FixTermTypo {
			tokens = r.FixTermTypo(tokens)
		}
		r.Write(html.EscapeHTML(tokens))
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCodeSpan(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if r.Options.AutoSpace {
			if text := node.PreviousNodeText(); "" != text {
				lastc, _ := utf8.DecodeLastRuneInString(text)
				if unicode.IsLetter(lastc) || unicode.IsDigit(lastc) {
					r.WriteByte(lex.ItemSpace)
				}
			}
		}
	} else {
		if r.Options.AutoSpace {
			if text := node.NextNodeText(); "" != text {
				firstc, _ := utf8.DecodeRuneInString(text)
				if unicode.IsLetter(firstc) || unicode.IsDigit(firstc) {
					r.WriteByte(lex.ItemSpace)
				}
			}
		}
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCodeSpanOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("code", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCodeSpanContent(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Write(html.EscapeHTML(node.Tokens))
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderCodeSpanCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/code", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmphasis(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.TextAutoSpacePrevious(node)
	} else {
		r.TextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmAsteriskOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("em", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmAsteriskCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/em", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmUnderscoreOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("em", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderEmUnderscoreCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/em", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrong(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.TextAutoSpacePrevious(node)
	} else {
		r.TextAutoSpaceNext(node)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrongA6kOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("strong", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrongA6kCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/strong", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrongU8eOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("strong", node.Parent.KramdownIAL, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderStrongU8eCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("/strong", nil, false)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBlockquote(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.handleKramdownBlockIAL(node)
		r.Tag("blockquote", node.KramdownIAL, false)
		r.Newline()
	} else {
		r.Newline()
		r.WriteString("</blockquote>")
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderBlockquoteMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

// func (r *OceanPressRender) renderHeading(node *ast.Node, entering bool) ast.WalkStatus {
// 	if entering {
// 		r.Newline()
// 		level := headingLevel[node.HeadingLevel : node.HeadingLevel+1]
// 		r.WriteString("<h" + level)
// 		id := HeadingID(node)
// 		if r.Options.ToC || r.Options.HeadingID || r.Options.KramdownBlockIAL {
// 			r.WriteString(" id=\"" + id + "\"")
// 			if r.Options.KramdownBlockIAL {
// 				if "id" != r.Options.KramdownIALIDRenderName && 0 < len(node.KramdownIAL) {
// 					r.WriteString(" " + r.Options.KramdownIALIDRenderName + "=\"" + node.KramdownIAL[0][1] + "\"")
// 				}
// 				if 1 < len(node.KramdownIAL) {
// 					exceptID := node.KramdownIAL[1:]
// 					for _, attr := range exceptID {
// 						r.WriteString(" " + attr[0] + "=\"" + attr[1] + "\"")
// 					}
// 				}
// 			}
// 		}
// 		r.WriteString(">")
// 	} else {
// 		if r.Options.HeadingAnchor {
// 			id := HeadingID(node)
// 			r.Tag("a", [][]string{{"id", "vditorAnchor-" + id}, {"class", "vditor-anchor"}, {"href", "#" + id}}, false)
// 			r.WriteString(`<svg viewBox="0 0 16 16" version="1.1" width="16" height="16"><path fill-rule="evenodd" d="M4 9h1v1H4c-1.5 0-3-1.69-3-3.5S2.55 3 4 3h4c1.45 0 3 1.69 3 3.5 0 1.41-.91 2.72-2 3.25V8.59c.58-.45 1-1.27 1-2.09C10 5.22 8.98 4 8 4H4c-.98 0-2 1.22-2 2.5S3 9 4 9zm9-3h-1v1h1c1 0 2 1.22 2 2.5S13.98 12 13 12H9c-.98 0-2-1.22-2-2.5 0-.83.42-1.64 1-2.09V6.25c-1.09.53-2 1.84-2 3.25C6 11.31 7.55 13 9 13h4c1.45 0 3-1.69 3-3.5S14.5 6 13 6z"></path></svg>`)
// 			r.Tag("/a", nil, false)
// 		}
// 		r.WriteString("</h" + headingLevel[node.HeadingLevel:node.HeadingLevel+1] + ">")
// 		r.Newline()
// 	}
// 	return ast.WalkContinue
// }

func (r *OceanPressRender) renderHeadingC8hMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderHeadingID(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *OceanPressRender) renderList(node *ast.Node, entering bool) ast.WalkStatus {
	tag := "ul"
	if 1 == node.ListData.Typ || (3 == node.ListData.Typ && 0 == node.ListData.BulletChar) {
		tag = "ol"
	}
	if entering {
		r.Newline()
		var attrs [][]string
		r.renderListStyle(node, &attrs)
		if 0 == node.ListData.BulletChar && 1 != node.ListData.Start {
			attrs = append(attrs, []string{"start", strconv.Itoa(node.ListData.Start)})
		}
		r.handleKramdownBlockIAL(node)
		attrs = append(attrs, node.KramdownIAL...)
		r.Tag(tag, attrs, false)
		r.Newline()
	} else {
		r.Newline()
		r.Tag("/"+tag, nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderListItem(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		var attrs [][]string
		r.handleKramdownBlockIAL(node)
		attrs = append(attrs, node.KramdownIAL...)
		if 3 == node.ListData.Typ && nil != node.FirstChild && ((ast.NodeTaskListItemMarker == node.FirstChild.Type) ||
			(nil != node.FirstChild.FirstChild && ast.NodeTaskListItemMarker == node.FirstChild.FirstChild.Type)) {
			taskListItemMarker := node.FirstChild.FirstChild
			if nil == taskListItemMarker {
				taskListItemMarker = node.FirstChild
			}
			taskClass := "protyle-task"
			if taskListItemMarker.TaskListItemChecked {
				taskClass += " protyle-task--done"
			}
			attrs = append(attrs, []string{"class", taskClass})
		}
		r.Tag("li", attrs, false)
	} else {
		r.Tag("/li", nil, false)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderTaskListItemMarker(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		var attrs [][]string
		if node.TaskListItemChecked {
			attrs = append(attrs, []string{"checked", ""})
		}
		attrs = append(attrs, []string{"disabled", ""}, []string{"type", "checkbox"})
		r.Tag("input", attrs, true)
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderThematicBreak(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Newline()
		r.Tag("hr", nil, true)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderHardBreak(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Tag("br", nil, true)
		r.Newline()
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) renderSoftBreak(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if r.Options.SoftBreak2HardBreak {
			r.Tag("br", nil, true)
			r.Newline()
		} else {
			r.Newline()
		}
	}
	return ast.WalkContinue
}

func (r *OceanPressRender) handleKramdownBlockIAL(node *ast.Node) {
	// 已经统一在 store 中处理了
	// if r.Options.KramdownBlockIAL && "id" != r.Options.KramdownIALIDRenderName && 0 < len(node.KramdownIAL) {
	// 	// 第一项必须是 ID
	// 	node.KramdownIAL[0][0] = r.Options.KramdownIALIDRenderName
	// }
}
