// Package empty_brace_spaces ports eslint-plugin-unicorn's
// `empty-brace-spaces` rule.
package empty_brace_spaces

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var messageEmptyBraceSpaces = rule.RuleMessage{
	Id:          "empty-brace-spaces",
	Description: "Do not add spaces between braces.",
}

// isStaticBlockBody reports whether a Block is the body tsgo nests inside a
// `static {}` block. ESTree has no such node — upstream sees a single
// StaticBlock — so the outer KindClassStaticBlockDeclaration listener already
// covers this brace pair and observing the nested Block too would report the
// identical span twice.
func isStaticBlockBody(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindClassStaticBlockDeclaration
}

// hasChildren reports whether the node already holds content, in which case
// the braces are not an empty pair. `SemicolonClassElement` entries are stray
// `;` between class members and have no ESTree counterpart, so they keep a
// class body "empty" the way upstream's `node.body` does.
func hasChildren(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		return len(node.AsObjectLiteralExpression().Properties.Nodes) > 0
	case ast.KindBlock, ast.KindClassStaticBlockDeclaration:
		// tsgo keeps the Block of a `static {}` block on the declaration's
		// own `Body` field, while `Block.Body()` is the node itself.
		body := node
		if node.Kind == ast.KindClassStaticBlockDeclaration {
			body = node.AsClassStaticBlockDeclaration().Body
		}
		if body == nil {
			return false
		}
		return len(body.AsBlock().Statements.Nodes) > 0
	default:
		members := node.Members()
		for _, member := range members {
			if member != nil && !ast.IsSemicolonClassElement(member) {
				return true
			}
		}
		return false
	}
}

// innerBraceRange returns the span strictly between the node's opening and
// closing braces. The opening brace is located with a scan bounded by
// node.End(), because a `static {}` block starts at the `static` keyword and
// the enclosing class body's brace comes first. The closing brace is always
// the node's last character for every node kind this rule inspects.
func innerBraceRange(node *ast.Node, sourceFile *ast.SourceFile) (core.TextRange, bool) {
	start := scanner.SkipTrivia(sourceFile.Text(), utils.TrimNodeTextRange(sourceFile, node).Pos())
	end := node.End()

	s := scanner.GetScannerForSourceFile(sourceFile, start)
	for s.TokenStart() < end {
		if s.Token() == ast.KindOpenBraceToken {
			return core.NewTextRange(s.TokenEnd(), end-1), true
		}
		s.Scan()
	}
	return core.TextRange{}, false
}

// isWhitespaceOnly reports whether text is made up solely of ECMAScript
// whitespace, mirroring upstream's `/^\s+$/` test. `ecmascript.IsBlank`
// matches JavaScript's `\s` exactly, unlike `strings.TrimSpace`.
func isWhitespaceOnly(text string) bool {
	return text != "" && ecmascript.IsBlank(text)
}

// EmptyBraceSpacesRule disallows whitespace between the braces of an empty
// block, class body, static block, or object literal.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/empty-brace-spaces.js
var EmptyBraceSpacesRule = rule.Rule{
	Name:   "unicorn/empty-brace-spaces",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		check := func(node *ast.Node) {
			if isStaticBlockBody(node) || hasChildren(node) {
				return
			}

			inner, ok := innerBraceRange(node, ctx.SourceFile)
			if !ok {
				return
			}
			if !isWhitespaceOnly(ctx.SourceFile.Text()[inner.Pos():inner.End()]) {
				return
			}

			ctx.ReportRangeWithFixes(inner, messageEmptyBraceSpaces, rule.RuleFixRemoveRange(inner))
		}

		return rule.RuleListeners{
			ast.KindBlock:                       check,
			ast.KindClassDeclaration:            check,
			ast.KindClassExpression:             check,
			ast.KindClassStaticBlockDeclaration: check,
			ast.KindObjectLiteralExpression:     check,
		}
	},
}
