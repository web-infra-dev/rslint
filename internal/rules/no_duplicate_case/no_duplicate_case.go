package no_duplicate_case

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-duplicate-case
var NoDuplicateCaseRule = rule.Rule{
	Name: "no-duplicate-case",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				switchStmt := node.AsSwitchStatement()
				if switchStmt == nil || switchStmt.CaseBlock == nil {
					return
				}

				caseBlock := switchStmt.CaseBlock.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}

				seen := make(map[string]bool)

				for _, clause := range caseBlock.Clauses.Nodes {
					if clause.Kind != ast.KindCaseClause {
						continue
					}

					caseClause := clause.AsCaseOrDefaultClause()
					if caseClause == nil || caseClause.Expression == nil {
						continue
					}

					testText := getExpressionTokenSignature(ctx.SourceFile, caseClause.Expression)

					if seen[testText] {
						ctx.ReportNode(clause, rule.RuleMessage{
							Id:          "unexpected",
							Description: "Duplicate case label.",
						})
					} else {
						seen[testText] = true
					}
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
