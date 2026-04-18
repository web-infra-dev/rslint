package no_lone_blocks

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-lone-blocks
var NoLoneBlocksRule = rule.Rule{
	Name: "no-lone-blocks",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var loneBlocks []*ast.Node

		report := func(node *ast.Node) {
			parent := node.Parent
			id := "redundantBlock"
			description := "Block is redundant."
			if parent != nil && (parent.Kind == ast.KindBlock || parent.Kind == ast.KindClassStaticBlockDeclaration) {
				id = "redundantNestedBlock"
				description = "Nested block is redundant."
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          id,
				Description: description,
			})
		}

		isLoneBlock := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil {
				return false
			}
			switch parent.Kind {
			case ast.KindBlock, ast.KindSourceFile:
				return true
			case ast.KindCaseClause, ast.KindDefaultClause:
				clause := parent.AsCaseOrDefaultClause()
				if clause == nil || clause.Statements == nil {
					return false
				}
				statements := clause.Statements.Nodes
				return len(statements) != 1 || statements[0] != node
			}
			return false
		}

		markLoneBlock := func(node *ast.Node) {
			if len(loneBlocks) == 0 {
				return
			}
			if loneBlocks[len(loneBlocks)-1] == node.Parent {
				loneBlocks = loneBlocks[:len(loneBlocks)-1]
			}
		}

		return rule.RuleListeners{
			ast.KindBlock: func(node *ast.Node) {
				if isLoneBlock(node) {
					loneBlocks = append(loneBlocks, node)
				}
			},
			rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
				if len(loneBlocks) > 0 && loneBlocks[len(loneBlocks)-1] == node {
					loneBlocks = loneBlocks[:len(loneBlocks)-1]
					report(node)
					return
				}
				parent := node.Parent
				if parent == nil {
					return
				}
				if parent.Kind == ast.KindBlock {
					parentBlock := parent.AsBlock()
					if parentBlock != nil && parentBlock.Statements != nil && len(parentBlock.Statements.Nodes) == 1 {
						report(node)
					}
				}
			},
			ast.KindVariableStatement: func(node *ast.Node) {
				varStmt := node.AsVariableStatement()
				if varStmt == nil || varStmt.DeclarationList == nil {
					return
				}
				// let, const, using, and await using are block-scoped.
				if varStmt.DeclarationList.Flags&ast.NodeFlagsBlockScoped != 0 {
					markLoneBlock(node)
				}
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				if utils.IsInStrictMode(node, ctx.SourceFile) {
					markLoneBlock(node)
				}
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				markLoneBlock(node)
			},
		}
	},
}
