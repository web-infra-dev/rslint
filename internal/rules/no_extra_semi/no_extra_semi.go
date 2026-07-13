package no_extra_semi

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-extra-semi
var NoExtraSemiRule = rule.Rule{
	Name: "no-extra-semi",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindEmptyStatement: func(node *ast.Node) {
				parent := node.Parent
				if parent != nil {
					allowedParentTypes := []ast.Kind{
						ast.KindForStatement,
						ast.KindForInStatement,
						ast.KindForOfStatement,
						ast.KindWhileStatement,
						ast.KindDoStatement,
						ast.KindIfStatement,
						ast.KindLabeledStatement,
						ast.KindWithStatement,
					}
					for _, allowed := range allowedParentTypes {
						if parent.Kind == allowed {
							return
						}
					}
				}

				report(ctx, node)
			},
			ast.KindSemicolonClassElement: func(node *ast.Node) {
				report(ctx, node)
			},
		}
	},
}

func report(ctx rule.RuleContext, node *ast.Node) {
	msg := rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unnecessary semicolon.",
	}

	if isFixable(ctx, node) {
		ctx.ReportNodeWithFixes(node, msg, rule.RuleFixRemove(ctx.SourceFile, node))
	} else {
		ctx.ReportNode(node, msg)
	}
}

func isFixable(ctx rule.RuleContext, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}

	var stmts []*ast.Node
	switch parent.Kind {
	case ast.KindSourceFile, ast.KindModuleBlock:
		stmts = parent.Statements()
	case ast.KindBlock:
		// Only check directives in function bodies (block statements cannot have directives)
		if parent.Parent != nil && ast.IsFunctionLike(parent.Parent) {
			stmts = parent.Statements()
		} else {
			return true
		}
	default:
		return true
	}

	idx := -1
	for i, s := range stmts {
		if s == node {
			idx = i
			break
		}
	}
	if idx != -1 && idx < len(stmts)-1 && ast.IsPrologueDirective(stmts[idx+1]) {
		return false
	}
	return true
}
