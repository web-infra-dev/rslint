package no_useless_catch

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-useless-catch
var NoUselessCatchRule = rule.Rule{
	Name: "no-useless-catch",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCatchClause: func(node *ast.Node) {
				catchClause := node.AsCatchClause()
				if catchClause == nil {
					return
				}

				// The catch must have a parameter
				if catchClause.VariableDeclaration == nil {
					return
				}

				// The parameter must be a simple identifier (not destructured)
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name() == nil {
					return
				}
				paramName := varDecl.Name()
				if paramName.Kind != ast.KindIdentifier {
					return
				}
				catchParamText := paramName.AsIdentifier().Text

				// The body must have exactly one statement
				if catchClause.Block == nil {
					return
				}
				block := catchClause.Block.AsBlock()
				if block == nil || block.Statements == nil || len(block.Statements.Nodes) != 1 {
					return
				}

				// That statement must be a ThrowStatement
				stmt := block.Statements.Nodes[0]
				if stmt == nil || stmt.Kind != ast.KindThrowStatement {
					return
				}

				// The throw argument must be an Identifier
				throwStmt := stmt.AsThrowStatement()
				if throwStmt == nil || throwStmt.Expression == nil {
					return
				}
				if throwStmt.Expression.Kind != ast.KindIdentifier {
					return
				}

				// The thrown identifier must match the catch parameter name
				thrownText := throwStmt.Expression.AsIdentifier().Text
				if thrownText != catchParamText {
					return
				}

				// Determine whether the parent TryStatement has a finally block
				tryStmt := node.Parent
				if tryStmt == nil || tryStmt.Kind != ast.KindTryStatement {
					return
				}
				tryData := tryStmt.AsTryStatement()
				if tryData == nil {
					return
				}

				if tryData.FinallyBlock != nil {
					// Has finally: report on the catch clause
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unnecessaryCatchClause",
						Description: "Unnecessary catch clause.",
					})
				} else {
					// No finally: report on the try statement
					ctx.ReportNode(tryStmt, rule.RuleMessage{
						Id:          "unnecessaryCatch",
						Description: "Unnecessary try/catch wrapper.",
					})
				}
			},
		}
	},
}
