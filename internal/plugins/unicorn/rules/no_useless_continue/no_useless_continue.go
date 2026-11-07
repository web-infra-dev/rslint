package no_useless_continue

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
					return []rule.RuleFix{rule.RuleFixRemoveRange(unicornutil.StatementRemovalRange(ctx.SourceFile, textRange))}
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
