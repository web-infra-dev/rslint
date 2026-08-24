package no_lonely_if

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// NoLonelyIfRule disallows `if` statements as the only statement in `else`
// blocks.
// https://eslint.org/docs/latest/rules/no-lonely-if
var NoLonelyIfRule = rule.Rule{
	Name:   "no-lonely-if",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sf := ctx.SourceFile

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindBlock {
					return
				}
				if len(parent.AsBlock().Statements.Nodes) != 1 {
					return
				}

				grandparent := parent.Parent
				if grandparent == nil || grandparent.Kind != ast.KindIfStatement {
					return
				}
				if grandparent.AsIfStatement().ElseStatement != parent {
					return
				}

				if utils.AreBracesNecessary(sf, parent) {
					return
				}

				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
					Id:          "unexpectedLonelyIf",
					Description: "Unexpected if as the only statement in an else block.",
				}, func() []rule.RuleFix {
					fix, ok := buildLonelyIfFix(sf, parent, node)
					if !ok {
						return nil
					}
					return []rule.RuleFix{fix}
				})
			},
		}
	},
}

// buildLonelyIfFix mirrors ESLint's fixer: it replaces the `else { if (...) {...} }`
// block with `else if (...) {...}`. It refuses to fix (returns ok=false) when
// comments sit between the else block's braces and the inner if statement, or
// when removing the braces would change the code's meaning due to ASI.
func buildLonelyIfFix(sf *ast.SourceFile, elseBlock, ifNode *ast.Node) (rule.RuleFix, bool) {
	text := sf.Text()

	blockRange := utils.TrimNodeTextRange(sf, elseBlock)
	innerRange := utils.BracedNodeInnerRange(sf, elseBlock)
	nodeRange := utils.TrimNodeTextRange(sf, ifNode)

	// An unterminated `else {` recovers as a Block without a closing brace, so
	// the inner if reaches past the brace pair's interior and there is no pair
	// to unwrap.
	if nodeRange.Pos() < innerRange.Pos() || nodeRange.End() > innerRange.End() {
		return rule.RuleFix{}, false
	}

	// Don't fix if there are any non-whitespace characters interfering (e.g. comments).
	if !ecmascript.IsBlank(text[innerRange.Pos():nodeRange.Pos()]) ||
		!ecmascript.IsBlank(text[nodeRange.End():innerRange.End()]) {
		return rule.RuleFix{}, false
	}

	thenStatement := ifNode.AsIfStatement().ThenStatement
	if thenStatement.Kind != ast.KindBlock {
		lastIfToken, hasLastIfToken := utils.PreviousTokenBefore(sf, thenStatement, thenStatement.End())
		tokenAfterElseBlock, hasTokenAfter := utils.TokenAtOrAfter(sf, blockRange.End())

		if hasLastIfToken && lastIfToken.Kind != ast.KindSemicolonToken && hasTokenAfter {
			r, _ := utf8.DecodeRuneInString(tokenAfterElseBlock.Text)

			// `<` is absent from ESLint's set because no JavaScript statement can
			// start with it. A TypeScript one can (`<T>x;`), and it continues the
			// preceding expression across a line break, so an unbraced if body
			// would swallow the statement that follows the else block.
			if utils.IsSameLine(sf, thenStatement.End(), tokenAfterElseBlock.Start) ||
				strings.ContainsRune("([/+`-<", r) ||
				lastIfToken.Kind == ast.KindPlusPlusToken ||
				lastIfToken.Kind == ast.KindMinusMinusToken {
				// Fixing would change the semantics of the code due to ASI.
				return rule.RuleFix{}, false
			}
		}
	}

	elseKeyword, hasElseKeyword := utils.TokenBeforePosition(sf, blockRange.Pos())
	prefix := ""
	if hasElseKeyword && elseKeyword.End == blockRange.Pos() {
		prefix = " "
	}

	return rule.RuleFixReplaceRange(
		core.NewTextRange(blockRange.Pos(), blockRange.End()),
		prefix+text[nodeRange.Pos():nodeRange.End()],
	), true
}
