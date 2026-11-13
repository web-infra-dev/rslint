package no_useless_else

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func buildFix(ctx *rule.RuleContext, node *ast.Node) []rule.RuleFix {
	sf := ctx.SourceFile
	text := sf.Text()
	stmt := node.AsIfStatement()
	consequent, alternate := stmt.ThenStatement, stmt.ElseStatement
	thenRange := utils.TrimNodeTextRange(sf, consequent)
	elseRange := utils.TrimNodeTextRange(sf, alternate)
	if thenRange.End() >= elseRange.Pos() || elseRange.Pos() >= elseRange.End() ||
		hasDirectScopedDeclaration(alternate) {
		return nil
	}

	lineStarts := sf.ECMALineMap()
	afterLine := scanner.ComputeLineOfPosition(lineStarts, elseRange.End())
	afterLineEnd := len(text)
	if afterLine+1 < len(lineStarts) {
		afterLineEnd = utils.LineContentEnd(text, int(lineStarts[afterLine+1]))
	}
	comments := ctx.Comments.All()
	if utils.HasCommentInSpan(comments, thenRange.End(), elseRange.Pos()) ||
		!ecmascript.IsBlank(text[elseRange.End():afterLineEnd]) {
		return nil
	}

	bodyRange := elseRange
	if alternate.Kind == ast.KindBlock {
		if text[elseRange.Pos()] != '{' || text[elseRange.End()-1] != '}' ||
			hasUnsafeMultilineToken(sf, alternate) {
			return nil
		}
		bodyRange = utils.BracedNodeInnerRange(sf, alternate)
		// Directive comments can acquire a new target after the closing brace
		// disappears. Keep their original placement.
		for _, comment := range utils.CommentsInSpan(comments, bodyRange.Pos(), bodyRange.End()) {
			if utils.IsDirectiveComment(comment.Kind, ecmascript.StringTrim(utils.CommentValue(text, comment))) {
				return nil
			}
		}
	}
	body := text[bodyRange.Pos():bodyRange.End()]
	if !ecmascript.IsBlank(body) {
		first, hasFirst := utils.TokenAtOrAfter(sf, bodyRange.Pos())
		if consequent.Kind != ast.KindBlock && hasFirst && first.Start < bodyRange.End() &&
			unsafeBoundary(sf, thenRange.End(), first.Text) {
			return nil
		}
		if next, found := utils.TokenAtOrAfter(sf, elseRange.End()); found &&
			unsafeBoundary(sf, bodyRange.End(), next.Text) {
			return nil
		}

		indent := lineIndent(text, lineStarts, utils.TrimNodeTextRange(sf, node).Pos())
		if alternate.Kind == ast.KindBlock {
			firstContent := bodyRange.Pos() + len(body) - len(strings.TrimLeftFunc(body, ecmascript.IsWhiteSpaceOrLineTerminator))
			bodyIndent := lineIndent(text, lineStarts, firstContent)
			body = adjustBodyIndent(text, lineStarts, bodyRange, bodyIndent, indent)
		}
		linebreak := "\n"
		if len(lineStarts) > 1 {
			nextLine := int(lineStarts[1])
			linebreak = text[utils.LineContentEnd(text, nextLine):nextLine]
		}
		body = linebreak + indent + body
	} else {
		body = ""
	}
	// Reserve the consequent so overlapping fixes cannot invalidate the exit.
	replacement := text[thenRange.Pos():thenRange.End()] + body
	return []rule.RuleFix{rule.RuleFixReplaceRange(thenRange.WithEnd(elseRange.End()), replacement)}
}

func hasDirectScopedDeclaration(node *ast.Node) bool {
	if node.Kind == ast.KindBlock {
		for _, stmt := range node.Statements() {
			if isScopedDeclaration(stmt) {
				return true
			}
		}
		return false
	}
	return isScopedDeclaration(node)
}

func isScopedDeclaration(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindVariableStatement:
		return node.AsVariableStatement().DeclarationList.Flags&ast.NodeFlagsBlockScoped != 0
	case ast.KindClassDeclaration, ast.KindFunctionDeclaration, ast.KindEnumDeclaration,
		ast.KindImportEqualsDeclaration, ast.KindInterfaceDeclaration, ast.KindModuleDeclaration, ast.KindTypeAliasDeclaration:
		return true
	}
	return false
}

func hasUnsafeMultilineToken(sf *ast.SourceFile, node *ast.Node) bool {
	for _, token := range utils.TokensOfNode(sf, node) {
		if token.Kind != ast.KindJsxText && token.Kind != ast.KindJsxTextAllWhiteSpaces &&
			ecmascript.ContainsLineTerminator(sf.Text(), token.Start, token.End) {
			return true
		}
	}
	return false
}

func unsafeBoundary(sf *ast.SourceFile, end int, following string) bool {
	previous, ok := utils.TokenBeforePosition(sf, end)
	if !ok || previous.Kind == ast.KindSemicolonToken {
		return false
	}
	// Match Unicorn's conservative guard even when the preceding statement,
	// such as a bare return, would allow ASI before the following token.
	return following != "" && strings.ContainsRune("[(/<`+-", rune(following[0])) ||
		unicornutil.NeedsSemicolonAfter(sf, previous, following, false)
}

func lineIndent(text string, lineStarts []core.TextPos, pos int) string {
	start := int(lineStarts[scanner.ComputeLineOfPosition(lineStarts, pos)])
	end := start
	for end < pos && (text[end] == ' ' || text[end] == '\t') {
		end++
	}
	return text[start:end]
}

func adjustBodyIndent(text string, lineStarts []core.TextPos, body core.TextRange, from, to string) string {
	var result strings.Builder
	pos := body.Pos()
	firstLine := scanner.ComputeLineOfPosition(lineStarts, pos)
	for line := firstLine + 1; line < len(lineStarts) && int(lineStarts[line]) <= body.End(); line++ {
		next := int(lineStarts[line])
		contentEnd := utils.LineContentEnd(text, next)
		writeLine := text[pos:contentEnd]
		if pos != body.Pos() {
			writeLine = adjustLineIndent(writeLine, from, to)
		}
		result.WriteString(writeLine)
		result.WriteString(text[contentEnd:next])
		pos = next
	}
	last := text[pos:body.End()]
	if pos != body.Pos() {
		last = adjustLineIndent(last, from, to)
	}
	result.WriteString(last)
	return ecmascript.StringTrim(result.String())
}

func adjustLineIndent(line, from, to string) string {
	if ecmascript.IsBlank(line) {
		return ""
	}
	if rest, ok := strings.CutPrefix(line, from); ok {
		return to + rest
	}
	return line
}
