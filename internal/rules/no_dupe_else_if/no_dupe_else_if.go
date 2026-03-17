package no_dupe_else_if

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-dupe-else-if
var NoDupeElseIfRule = rule.Rule{
	Name: "no-dupe-else-if",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				// Only check if this IfStatement is in an "else if" position:
				// its parent must be an IfStatement whose ElseStatement is this node.
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindIfStatement {
					return
				}
				parentIf := parent.AsIfStatement()
				if parentIf == nil || parentIf.ElseStatement != node {
					return
				}

				// Get the token signature of the current condition.
				ifStmt := node.AsIfStatement()
				if ifStmt == nil || ifStmt.Expression == nil {
					return
				}
				currentCondition := getExpressionTokenSignature(ctx.SourceFile, ifStmt.Expression)

				// Walk up the if-else-if chain collecting all prior conditions.
				ancestor := parent
				for ancestor != nil && ancestor.Kind == ast.KindIfStatement {
					ancestorIf := ancestor.AsIfStatement()
					if ancestorIf == nil || ancestorIf.Expression == nil {
						break
					}

					priorCondition := getExpressionTokenSignature(ctx.SourceFile, ancestorIf.Expression)
					if currentCondition == priorCondition {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "unexpected",
							Description: "This branch can never execute. Its condition is a duplicate or covered by previous conditions in the if-else-if chain.",
						})
						return
					}

					// Move to the next ancestor in the chain.
					// The ancestor must itself be an "else if" for the chain to continue.
					grandparent := ancestor.Parent
					if grandparent == nil || grandparent.Kind != ast.KindIfStatement {
						break
					}
					grandparentIf := grandparent.AsIfStatement()
					if grandparentIf == nil || grandparentIf.ElseStatement != ancestor {
						break
					}
					ancestor = grandparent
				}
			},
		}
	},
}

// getExpressionTokenSignature produces a canonical string from an expression's
// tokens (skipping comments and whitespace), matching ESLint's equalTokens approach.
func getExpressionTokenSignature(sourceFile *ast.SourceFile, expr *ast.Node) string {
	var result strings.Builder
	sourceText := sourceFile.Text()
	first := true
	utils.ForEachToken(expr, func(token *ast.Node) {
		trimmedRange := utils.TrimNodeTextRange(sourceFile, token)
		start := trimmedRange.Pos()
		end := trimmedRange.End()
		if start < end {
			if !first {
				result.WriteByte(' ')
			}
			result.WriteString(sourceText[start:end])
			first = false
		}
	}, sourceFile)
	return result.String()
}
