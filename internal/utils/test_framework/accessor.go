package test_framework

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// This file is the framework-neutral counterpart of eslint-plugin-jest's
// src/rules/utils/accessors.ts and its replaceAccessorFixer helper.
//
// It exists because ast.Node.Pos() is not ESTree's node.range[0]: Pos() points
// at the end of the previous token, so it includes the node's leading trivia
// (whitespace, newlines, comments). A fix range built straight from Pos() will
// silently overwrite that trivia. Report ranges are unaffected — ctx.ReportNode
// already routes through utils.TrimNodeTextRange — so only hand-built ranges
// need these helpers.

// IsAccessorNode reports whether node is one of the node kinds GetMemberEntries
// stores in MemberEntry.Node.
func IsAccessorNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindPrivateIdentifier,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return true
	default:
		return false
	}
}

// IsComputedIdentifierAccessor reports whether node is an identifier written as
// a computed key — the `b` in `a[b]` rather than the `b` in `a.b`.
//
// GetMemberEntries reports such a key as a member named after the identifier's
// text, which matches eslint-plugin-jest's getNodeChain. That keeps chains like
// `expect(a).not[x]()` parseable, which upstream relies on, but the name is the
// identifier's runtime *value*, not its text. Rewriting the node would rename a
// variable reference rather than a member, so fixes must not touch it.
func IsComputedIdentifierAccessor(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}

	child := node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindParenthesizedExpression {
			child = parent
			continue
		}
		return parent.Kind == ast.KindElementAccessExpression &&
			parent.AsElementAccessExpression().ArgumentExpression == child
	}
	return false
}

// AccessorReceiverAndParent returns the receiver and accessor expression that
// own entry. Parentheses around a computed key are transparent.
func AccessorReceiverAndParent(entry *MemberEntry) (*ast.Node, *ast.Node) {
	if entry == nil || entry.Node == nil {
		return nil, nil
	}

	parent := entry.Node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return nil, nil
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		property := parent.AsPropertyAccessExpression()
		if property.Name() != entry.Node {
			return nil, nil
		}
		return property.Expression, parent
	case ast.KindElementAccessExpression:
		element := parent.AsElementAccessExpression()
		if ast.SkipParentheses(element.ArgumentExpression) != entry.Node {
			return nil, nil
		}
		return element.Expression, parent
	default:
		return nil, nil
	}
}

// AccessorQuestionDotToken returns the optional-chain token owned by accessor.
func AccessorQuestionDotToken(accessor *ast.Node) *ast.Node {
	if accessor == nil {
		return nil
	}

	switch accessor.Kind {
	case ast.KindPropertyAccessExpression:
		return accessor.AsPropertyAccessExpression().QuestionDotToken
	case ast.KindElementAccessExpression:
		return accessor.AsElementAccessExpression().QuestionDotToken
	default:
		return nil
	}
}

func removeAccessorSyntaxRanges(
	sourceFile *ast.SourceFile,
	comments []*ast.CommentRange,
	accessor *ast.Node,
	start int,
) []core.TextRange {
	if !utils.HasCommentInSpan(comments, start, accessor.End()) {
		return []core.TextRange{core.NewTextRange(start, accessor.End())}
	}

	ranges := []core.TextRange{}
	for _, token := range utils.TokensOfNode(sourceFile, accessor) {
		if token.Start >= start && token.End <= accessor.End() {
			ranges = append(ranges, token.Range())
		}
	}
	return ranges
}

// RemoveAccessorEntryRanges removes entry without requiring the caller to
// retain the full parsed member slice. It derives the next accessor or call
// directly from the AST, which is required for semantic entries originating in
// an alias initializer rather than the final call site's Members.
func RemoveAccessorEntryRanges(
	sourceFile *ast.SourceFile,
	comments []*ast.CommentRange,
	entry *MemberEntry,
) ([]core.TextRange, bool) {
	if sourceFile == nil || entry == nil {
		return nil, false
	}
	receiver, accessor := AccessorReceiverAndParent(entry)
	if receiver == nil || accessor == nil {
		return nil, false
	}

	questionDot := AccessorQuestionDotToken(accessor)
	start := receiver.End()
	if questionDot != nil {
		start = questionDot.End()
	}
	ranges := removeAccessorSyntaxRanges(sourceFile, comments, accessor, start)
	if len(ranges) == 0 || questionDot == nil {
		return ranges, len(ranges) > 0
	}

	next := accessor.Parent
	for next != nil && next.Kind == ast.KindParenthesizedExpression {
		next = next.Parent
	}
	if next == nil {
		return nil, false
	}

	switch next.Kind {
	case ast.KindPropertyAccessExpression:
		property := next.AsPropertyAccessExpression()
		if property.Expression != accessor {
			return nil, false
		}
		connector, ok := utils.TokenAtOrAfter(sourceFile, accessor.End())
		if !ok || (connector.Text != "." && connector.Text != "?.") {
			return nil, false
		}
		nextNameStart := utils.TrimNodeTextRange(sourceFile, property.Name()).Pos()
		if utils.HasCommentInSpan(comments, accessor.End(), nextNameStart) {
			ranges = append(ranges, connector.Range())
		} else {
			ranges = append(ranges, core.NewTextRange(accessor.End(), nextNameStart))
		}
	case ast.KindElementAccessExpression:
		element := next.AsElementAccessExpression()
		if element.Expression != accessor {
			return nil, false
		}
		if nextQuestionDot := AccessorQuestionDotToken(next); nextQuestionDot != nil {
			nextQuestionDotRange := utils.TrimNodeTextRange(sourceFile, nextQuestionDot)
			if utils.HasCommentInSpan(comments, accessor.End(), nextQuestionDotRange.End()) {
				ranges = append(ranges, nextQuestionDotRange)
			} else {
				ranges = append(ranges, core.NewTextRange(accessor.End(), nextQuestionDotRange.End()))
			}
		}
	case ast.KindCallExpression:
		call := next.AsCallExpression()
		if call.Expression != accessor {
			return nil, false
		}
		if call.QuestionDotToken != nil {
			callQuestionDotRange := utils.TrimNodeTextRange(sourceFile, call.QuestionDotToken)
			if utils.HasCommentInSpan(comments, accessor.End(), callQuestionDotRange.End()) {
				ranges = append(ranges, callQuestionDotRange)
			} else {
				ranges = append(ranges, core.NewTextRange(accessor.End(), callQuestionDotRange.End()))
			}
		}
	case ast.KindVariableDeclaration:
		declaration := next.AsVariableDeclaration()
		if ast.SkipParentheses(declaration.Initializer) != accessor {
			return nil, false
		}
		ranges = append(ranges, utils.TrimNodeTextRange(sourceFile, questionDot))
	default:
		return nil, false
	}
	return ranges, true
}

// AccessorRange returns the accessor's own source range, excluding leading
// trivia. For string and template literals the range includes the delimiters.
func AccessorRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, bool) {
	if sourceFile == nil || !IsAccessorNode(node) {
		return core.TextRange{}, false
	}
	return utils.TrimNodeTextRange(sourceFile, node), true
}

// AccessorValueRange returns the range of the accessor's textual value: the
// token itself for an identifier, and the span between the delimiters for a
// string or template literal.
func AccessorValueRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, bool) {
	tokenRange, ok := AccessorRange(sourceFile, node)
	if !ok {
		return core.TextRange{}, false
	}

	switch node.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return tokenRange, true
	default:
		// Both `'x'` and `` `x` `` use a single-character delimiter per side.
		// Guard against unterminated literals in recovered parses.
		if tokenRange.End()-tokenRange.Pos() < 2 {
			return core.TextRange{}, false
		}
		return core.NewTextRange(tokenRange.Pos()+1, tokenRange.End()-1), true
	}
}

// AccessorReplacement returns the range and text that rewrite an accessor to
// name. The accessor's written form is preserved: a bare identifier stays bare,
// and a string or template literal keeps its original delimiters because only
// the text between them is replaced.
//
// Preserving the delimiters is a deliberate divergence from upstream's
// replaceAccessorFixer, which always emits `'name'`. Normalizing quotes makes an
// autofix rewrite `["toEqual"]` into `['toStrictEqual']`, which introduces a
// fresh `quotes` violation in a double-quoted codebase. jest/no-alias-methods
// already preserved delimiters before this helper existed; this makes the whole
// plugin consistent with that.
//
// Returns false for accessors that cannot meaningfully be renamed: a computed
// identifier key, whose text is a variable reference rather than the member name
// (see IsComputedIdentifierAccessor), and a private identifier, which is never a
// matcher or a framework API member. Callers should still report those — only
// the fix has to be withheld.
func AccessorReplacement(sourceFile *ast.SourceFile, node *ast.Node, name string) (core.TextRange, string, bool) {
	if node == nil || IsComputedIdentifierAccessor(node) {
		return core.TextRange{}, "", false
	}

	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		valueRange, ok := AccessorValueRange(sourceFile, node)
		if !ok {
			return core.TextRange{}, "", false
		}
		return valueRange, name, true
	default:
		return core.TextRange{}, "", false
	}
}
