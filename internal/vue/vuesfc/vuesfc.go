// Package vuesfc splits a Vue Single File Component into its top-level blocks
// and projects its script blocks onto a text the TypeScript parser can read.
//
// It ports the block-splitting half of `@vue/compiler-sfc`'s `parse()` (Vue
// 3.5), the part that turns an SFC into a `<template>`, some `<script>`s, some
// `<style>`s, and custom blocks, together with the `pad: 'space'` projection
// that same function offers its callers. Upstream:
// https://github.com/vuejs/core/blob/main/packages/compiler-sfc/src/parse.ts
//
// # Why the projection preserves offsets
//
// [Extract] returns a text of exactly the input's byte length: every byte
// outside a script block becomes a space, and line terminators are kept where
// they were. So a position in the returned text is the same position in the
// `.vue` file on disk, and a range computed against the parsed AST needs no
// translation before it is reported to a user or applied as an autofix. This
// is upstream's own `pad: 'space'` mode ("if `pad === 'space'`, whitespace
// replaces all characters"), used here for every block rather than only for
// the text preceding one.
//
// A multi-byte character outside a script block becomes as many spaces as it
// had bytes, which keeps every byte offset exact. On a line that holds both
// template and script text this shifts the UTF-16 columns upstream would
// report for that one line; a block on its own line, which is every SFC in
// practice, is unaffected.
//
// # What the port covers, and what it leaves out
//
// Covered: top-level blocks, `lang`, `setup` and `src` attributes, quoted and
// unquoted attribute values, self-closing blocks, HTML comments, ASCII
// case-insensitive tag names, and the raw-text content model HTML gives
// `<script>` and `<style>`.
//
// Left out: everything upstream does after splitting, such as compiler errors,
// `<script setup>` binding analysis, source maps, style preprocessing, and
// loading a block's external `src`. A block carrying `src` contributes no
// inline content here, exactly as upstream generates no AST for one.
//
// Tag and attribute names are compared by folding ASCII only, because that is
// the fold HTML specifies for them. This is deliberately not
// ecmascript.StringToLowerCase: that answers what JavaScript's
// String#toLowerCase would say about arbitrary text, which is a different
// question from what HTML says two names are. Attribute values are not folded
// at all, matching both HTML and upstream's exact `lang` comparison.
package vuesfc

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/vue/htmlscan"
)

// BlockKind classifies a top-level SFC block by its tag name.
type BlockKind uint8

const (
	// BlockOther is a custom block, such as `<i18n>` or `<docs>`.
	BlockOther BlockKind = iota
	BlockTemplate
	BlockScript
	BlockStyle
)

// Block is one top-level element of an SFC.
//
// Content is the range between the start tag and the end tag, matching
// upstream's `block.content = source.slice(loc.start.offset, loc.end.offset)`.
// It is empty for a self-closing or unterminated block. Outer spans the whole
// element including both tags.
type Block struct {
	Kind    BlockKind
	Tag     string
	Content core.TextRange
	Outer   core.TextRange
	// Lang is the `lang` attribute value, empty when the attribute is absent.
	Lang string
	// Setup reports the presence of a bare `setup` attribute, which is what
	// makes a script block a `<script setup>` upstream.
	Setup bool
	// External reports a `src` attribute. Such a block's content lives in
	// another file, so Content is empty even when the tags are not.
	External bool
}

// Result is one SFC's split blocks and the script projection built from them.
type Result struct {
	// Text is the input with every byte outside a script block blanked. Its
	// length always equals the input's.
	Text string
	// ScriptKind is the script kind the projection must be parsed with. It is
	// [core.ScriptKindJS] for an SFC with no script block, because an empty
	// projection parses the same either way.
	ScriptKind core.ScriptKind
	// Blocks holds every top-level block in source order.
	Blocks []Block
	// Comments holds every top-level HTML comment in source order. A comment
	// inside a block's content is not reported here.
	Comments []core.TextRange
}

// Scripts returns the blocks that contribute content to Text, in source order.
func (r Result) Scripts() []Block {
	var scripts []Block
	for _, block := range r.Blocks {
		if contributesContent(block) {
			scripts = append(scripts, block)
		}
	}
	return scripts
}

// contributesContent reports whether a block's text is copied into the
// projection. Scripts and the projection must agree on this, or a caller would
// read a range the projection blanked.
func contributesContent(block Block) bool {
	return block.Kind == BlockScript &&
		!block.External &&
		block.Content.Len() > 0 &&
		supportedLang(block.Lang)
}

// Extract splits text into top-level blocks and builds the offset-preserving
// script projection described in the package comment.
//
// It never fails: text that holds no recognizable block yields an all-blank
// projection, which is the right answer for an SFC with only a template and a
// style. A rule that asks about the file rather than its code (a filename
// rule, say) still runs against that file.
func Extract(text string) Result {
	result := Result{ScriptKind: core.ScriptKindJS}
	scan(text, &result)
	result.Text = project(text, result.Blocks)
	result.ScriptKind = scriptKind(result.Blocks)
	return result
}

// scan walks the top level of an SFC, collecting blocks and comments. Text
// between blocks is not markup Vue keeps, so it is skipped.
func scan(text string, result *Result) {
	for index := 0; index < len(text); {
		next := strings.IndexByte(text[index:], '<')
		if next < 0 {
			return
		}
		index += next

		switch {
		case strings.HasPrefix(text[index:], "<!--"):
			end := htmlscan.CommentEnd(text, index)
			result.Comments = append(result.Comments, core.NewTextRange(index, end))
			index = end
		case strings.HasPrefix(text[index:], "<!"), strings.HasPrefix(text[index:], "<?"):
			index = htmlscan.DeclarationEnd(text, index)
		case strings.HasPrefix(text[index:], "</"):
			// A stray end tag at the top level closes nothing.
			index = htmlscan.DeclarationEnd(text, index)
		default:
			block, end, ok := readBlock(text, index)
			if !ok {
				// Not a start tag after all: a bare `<` in top-level text.
				index++
				continue
			}
			result.Blocks = append(result.Blocks, block)
			index = end
		}
	}
}

// readBlock reads one top-level element beginning at the `<` at start. It
// reports false when start does not begin a start tag.
func readBlock(text string, start int) (Block, int, bool) {
	nameEnd := htmlscan.TagNameEnd(text, start+1)
	if nameEnd == start+1 {
		return Block{}, start, false
	}
	tag := htmlscan.Fold(text[start+1 : nameEnd])

	startTag := htmlscan.ReadStartTag(text, nameEnd)

	block := Block{Kind: blockKind(tag), Tag: tag}
	for _, attribute := range startTag.Attributes {
		switch htmlscan.Fold(attribute.Name) {
		case "lang":
			// The value is kept verbatim. HTML folds an attribute name, not
			// its value, and upstream compares `lang` against "ts"/"tsx"
			// exactly, so `lang="TS"` names no language Vue compiles, and
			// must name none here either.
			block.Lang = attribute.Value
		case "setup":
			// `setup` is a boolean attribute: upstream keys off its presence,
			// so `setup` and `setup=""` both mark a script setup block.
			block.Setup = true
		case "src":
			block.External = attribute.HasValue && attribute.Value != ""
		}
	}

	if startTag.SelfClosing || htmlscan.IsVoid(tag) {
		block.Outer = core.NewTextRange(start, startTag.End)
		block.Content = core.NewTextRange(startTag.End, startTag.End)
		return block, startTag.End, true
	}

	contentEnd, outerEnd := findEndTag(text, tag, startTag.End)
	block.Content = core.NewTextRange(startTag.End, contentEnd)
	block.Outer = core.NewTextRange(start, outerEnd)
	return block, outerEnd, true
}

// findEndTag locates the end tag closing a block whose content starts at
// contentStart. It returns the content's end and the element's end.
//
// A raw-text element ends at its first end tag, as HTML's content model says.
// Any other element is matched by depth, so a nested `<template>` does not
// close its parent, and a comment is skipped so that a commented-out end tag
// does not either. An unterminated block runs to the end of the file, which is
// what upstream's error recovery also yields.
func findEndTag(text string, tag string, contentStart int) (int, int) {
	depth := 1
	raw := htmlscan.IsRawText(tag)

	for index := contentStart; index < len(text); {
		next := strings.IndexByte(text[index:], '<')
		if next < 0 {
			break
		}
		index += next

		if !raw && strings.HasPrefix(text[index:], "<!--") {
			index = htmlscan.CommentEnd(text, index)
			continue
		}

		if strings.HasPrefix(text[index:], "</") {
			nameEnd := htmlscan.TagNameEnd(text, index+2)
			if htmlscan.Fold(text[index+2:nameEnd]) == tag {
				depth--
				if depth == 0 {
					return index, htmlscan.DeclarationEnd(text, nameEnd)
				}
			}
			index = nameEnd
			continue
		}

		if !raw {
			nameEnd := htmlscan.TagNameEnd(text, index+1)
			if nameEnd > index+1 && htmlscan.Fold(text[index+1:nameEnd]) == tag {
				if startTag := htmlscan.ReadStartTag(text, nameEnd); !startTag.SelfClosing {
					depth++
					index = startTag.End
					continue
				}
			}
			index = max(nameEnd, index+1)
			continue
		}

		index++
	}
	return len(text), len(text)
}

// project builds the offset-preserving text described in the package comment:
// the input's length, script content verbatim at its own offsets, and spaces
// everywhere else with line terminators kept in place.
func project(text string, blocks []Block) string {
	projected := make([]byte, len(text))
	for index := range len(text) {
		if character := text[index]; character == '\n' || character == '\r' {
			projected[index] = character
			continue
		}
		projected[index] = ' '
	}
	for _, block := range blocks {
		if !contributesContent(block) {
			continue
		}
		copy(projected[block.Content.Pos():block.Content.End()],
			text[block.Content.Pos():block.Content.End()])
	}
	return string(projected)
}

// scriptKind picks the one script kind the whole projection is parsed with.
//
// An SFC may hold both a `<script>` and a `<script setup>`, and Vue requires
// them to agree on `lang`. When they disagree anyway, the more permissive kind
// wins so that neither block's syntax is rejected: JSX and TSX admit what JS
// and TS do, and TypeScript admits what JavaScript does.
func scriptKind(blocks []Block) core.ScriptKind {
	kind := core.ScriptKindJS
	typescript := false
	jsx := false

	for _, block := range blocks {
		if !contributesContent(block) {
			continue
		}
		switch block.Lang {
		case "ts":
			typescript = true
		case "tsx":
			typescript = true
			jsx = true
		case "jsx":
			jsx = true
		}
	}

	switch {
	case typescript && jsx:
		kind = core.ScriptKindTSX
	case typescript:
		kind = core.ScriptKindTS
	case jsx:
		kind = core.ScriptKindJSX
	}
	return kind
}

// supportedLang reports whether a script block's `lang` names a language the
// TypeScript parser reads. A block in anything else (CoffeeScript, say)
// contributes no content, so its text is never parsed as if it were JavaScript.
func supportedLang(lang string) bool {
	switch lang {
	case "", "js", "javascript", "ts", "typescript", "jsx", "tsx":
		return true
	}
	return false
}

func blockKind(tag string) BlockKind {
	switch tag {
	case "template":
		return BlockTemplate
	case "script":
		return BlockScript
	case "style":
		return BlockStyle
	}
	return BlockOther
}

// Extension is the file extension of a Vue Single File Component.
const Extension = ".vue"

// IsFile reports whether path names a Vue Single File Component.
//
// The extension is matched by folding ASCII, because the filesystems rslint
// runs on disagree about whether a name's case is part of its identity, and a
// component saved as `App.VUE` is the same kind of file either way.
func IsFile(path string) bool {
	if len(path) <= len(Extension) {
		return false
	}
	return htmlscan.Fold(path[len(path)-len(Extension):]) == Extension
}
