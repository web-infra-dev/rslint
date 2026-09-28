package no_useless_continue

import (
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var NoUselessContinueRule = rule.Rule{
	Name:   "unicorn/no-useless-continue",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindContinueStatement: func(node *ast.Node) {
				if node.AsContinueStatement().Label != nil || !isUselessContinue(node) {
					return
				}

				textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				ctx.ReportRangeWithDeferredFixes(textRange, rule.RuleMessage{
					Id:          "no-useless-continue",
					Description: "Unnecessary `continue` statement.",
				}, func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixRemoveRange(statementRemovalRange(ctx.SourceFile.Text(), textRange))}
				})
			},
		}
	},
}

func isUselessContinue(node *ast.Node) bool {
	// A bare loop body needs its continue statement to remain valid syntax.
	if node.Parent == nil || node.Parent.Kind != ast.KindBlock {
		return false
	}

	for current := node; current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindBlock:
			block := parent.AsBlock()
			if block.Statements == nil || len(block.Statements.Nodes) == 0 || block.Statements.Nodes[len(block.Statements.Nodes)-1] != current {
				return false
			}
			current = parent
		case ast.KindIfStatement:
			ifStatement := parent.AsIfStatement()
			if ifStatement.ThenStatement != current && ifStatement.ElseStatement != current {
				return false
			}
			current = parent
		case ast.KindForStatement:
			return parent.AsForStatement().Statement == current
		case ast.KindForInStatement, ast.KindForOfStatement:
			return parent.AsForInOrOfStatement().Statement == current
		case ast.KindWhileStatement:
			return parent.AsWhileStatement().Statement == current
		case ast.KindDoStatement:
			return parent.AsDoStatement().Statement == current
		default:
			return false
		}
	}
	return false
}

// statementRemovalRange mirrors Unicorn's removeStatement: remove a whole
// otherwise empty line, but leave leading and trailing comments intact.
func statementRemovalRange(text string, statement core.TextRange) core.TextRange {
	start, end := statement.Pos(), statement.End()
	// Only whitespace can extend the removal range. Stop at the first token
	// or comment instead of scanning the whole line for each continue.
	lineStart := start
	for lineStart > 0 && text[lineStart-1] != '\r' && text[lineStart-1] != '\n' {
		r, size := utf8.DecodeLastRuneInString(text[:lineStart])
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			return statement
		}
		lineStart -= size
	}
	lineEnd := end
	for lineEnd < len(text) && text[lineEnd] != '\r' && text[lineEnd] != '\n' {
		r, size := utf8.DecodeRuneInString(text[lineEnd:])
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			return statement
		}
		lineEnd += size
	}

	end = lineEnd
	if start == 0 {
		if end < len(text) {
			if text[end] == '\r' && end+1 < len(text) && text[end+1] == '\n' {
				end += 2
			} else {
				end++
			}
		}
	} else {
		start = lineStart - 1
		if start > 0 && text[start-1] == '\r' && text[start] == '\n' {
			start--
		}
	}
	return core.NewTextRange(start, end)
}
