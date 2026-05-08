package reactutil

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IndentLineStart returns the byte offset of the first character on the
// line containing pos, derived from the source file's ECMA line map.
func IndentLineStart(lineMap []core.TextPos, pos int) int {
	line := scanner.ComputeLineOfPosition(lineMap, pos)
	return int(lineMap[line])
}

// IndentLeading counts consecutive `indentChar` bytes at the start of the
// line containing pos. Mirrors upstream ESLint's
// `getText(node, node.loc.start.column).match(/^[ ]+|^[\t]+/)` slice that
// every eslint-plugin-react indent-style rule uses to read a line's leading
// indent.
func IndentLeading(text string, lineMap []core.TextPos, pos int, indentChar byte) int {
	start := IndentLineStart(lineMap, pos)
	count := 0
	for i := start; i < len(text); i++ {
		if text[i] != indentChar {
			break
		}
		count++
	}
	return count
}

// NodeStartIndent returns the leading indent (count of `indentChar`) on
// the line containing the trimmed start of `node`.
func NodeStartIndent(sf *ast.SourceFile, node *ast.Node, indentChar byte) int {
	trimmed := utils.TrimNodeTextRange(sf, node)
	return IndentLeading(sf.Text(), sf.ECMALineMap(), trimmed.Pos(), indentChar)
}

// NodeEndIndent returns the leading indent (count of `indentChar`) on
// the line containing the trimmed end of `node`. Steps one byte back from
// `End()` to land inside the node so that a node ending exactly at a
// newline is still attributed to its own last line.
func NodeEndIndent(sf *ast.SourceFile, node *ast.Node, indentChar byte) int {
	trimmed := utils.TrimNodeTextRange(sf, node)
	pos := trimmed.End()
	if pos > 0 {
		pos--
	}
	return IndentLeading(sf.Text(), sf.ECMALineMap(), pos, indentChar)
}

// NodeStartUTF16Column returns the 0-based UTF-16 character column of
// the trimmed start of `node`. Used by indent-style rules whose option
// semantics are character-based (e.g. eslint-plugin-react's
// `jsx-indent-props` `'first'` mode, which aligns to the visual column of
// the first prop). Counting bytes here would diverge from upstream ESLint
// when the source contains multi-byte characters before the node.
func NodeStartUTF16Column(sf *ast.SourceFile, node *ast.Node) int {
	trimmed := utils.TrimNodeTextRange(sf, node)
	_, char := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, trimmed.Pos())
	return int(char)
}

// IsNodeFirstInLine reports whether node is the first non-whitespace
// (and non-comma) source character on its line. Walks back from the
// trimmed start over spaces, tabs, commas and CR; a leading newline (or
// start-of-file) means yes, anything else means no. The comma skip
// mirrors upstream's `getFirstNodeInLine` behaviour for array-literal
// siblings.
func IsNodeFirstInLine(sf *ast.SourceFile, node *ast.Node) bool {
	text := sf.Text()
	trimmed := utils.TrimNodeTextRange(sf, node)
	i := trimmed.Pos() - 1
	for i >= 0 {
		c := text[i]
		if c == '\n' {
			return true
		}
		if c == ' ' || c == '\t' || c == ',' || c == '\r' {
			i--
			continue
		}
		return false
	}
	return true
}

// WrongIndentMessage builds the standard `wrongIndent` rule message used
// by indent-style rules ported from eslint-plugin-react. The message text
// matches the upstream
// `Expected indentation of {needed} {type} {characters} but found {gotten}.`
// template; `characters` switches between "character" and "characters"
// based on whether `needed` is 1.
func WrongIndentMessage(needed, gotten int, indentType string) rule.RuleMessage {
	characters := "characters"
	if needed == 1 {
		characters = "character"
	}
	return rule.RuleMessage{
		Id: "wrongIndent",
		Description: fmt.Sprintf(
			"Expected indentation of %d %s %s but found %d.",
			needed, indentType, characters, gotten,
		),
		Data: map[string]string{
			"needed":     strconv.Itoa(needed),
			"type":       indentType,
			"characters": characters,
			"gotten":     strconv.Itoa(gotten),
		},
	}
}

// ReportIndentReplaceLeading emits a `wrongIndent` diagnostic anchored at
// `node` and provides an autofix that replaces the whitespace from the
// start of the node's line up to the node's trimmed start with `needed`
// repetitions of `indentChar`. This is the default fix shape used by
// upstream's
// `replaceTextRange([node.range[0] - node.loc.start.column, node.range[0]], …)`
// pattern.
//
// When `needed` is negative (only reachable via a negative `indentSize`
// option such as `[-2]`), the diagnostic is emitted WITHOUT a fix. The
// message still reports the negative `needed` verbatim — that matches
// upstream ESLint, which calls `' '.repeat(needed)` inside its fix
// lambda; the lambda throws `RangeError`, ESLint discards the fix but
// keeps the diagnostic intact. Skipping the fix here avoids the
// `strings: negative Repeat count` panic in Go (which would crash the
// whole linter, not just this one rule).
func ReportIndentReplaceLeading(ctx rule.RuleContext, node *ast.Node, needed, gotten int, indentChar byte, indentType string) {
	msg := WrongIndentMessage(needed, gotten, indentType)
	if needed < 0 {
		ctx.ReportNode(node, msg)
		return
	}
	indent := strings.Repeat(string(indentChar), needed)
	trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
	startPos := trimmed.Pos()
	lineStart := IndentLineStart(ctx.SourceFile.ECMALineMap(), startPos)
	ctx.ReportNodeWithFixes(node, msg, rule.RuleFix{
		Text:  indent,
		Range: core.NewTextRange(lineStart, startPos),
	})
}
