package no_duplicate_case

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
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

				sourceText := ctx.SourceFile.Text()
				var previousTests []string

				for _, clause := range caseBlock.Clauses.Nodes {
					if clause.Kind != ast.KindCaseClause {
						continue
					}

					caseClause := clause.AsCaseOrDefaultClause()
					if caseClause == nil || caseClause.Expression == nil {
						continue
					}

					testText := normalizeExpressionText(sourceText[caseClause.Expression.Pos():caseClause.Expression.End()])

					for _, prev := range previousTests {
						if prev == testText {
							ctx.ReportNode(clause, rule.RuleMessage{
								Id:          "unexpected",
								Description: "Duplicate case label.",
							})
							break
						}
					}

					previousTests = append(previousTests, testText)
				}
			},
		}
	},
}

// normalizeExpressionText normalizes source text for comparison by collapsing
// whitespace and removing comments outside of string literals.
func normalizeExpressionText(text string) string {
	var result strings.Builder
	result.Grow(len(text))
	i := 0
	lastWasSpace := false
	var inString byte // 0 = not in string, otherwise the quote character

	for i < len(text) {
		ch := text[i]

		// Inside a string literal: copy verbatim until closing quote
		if inString != 0 {
			result.WriteByte(ch)
			lastWasSpace = false
			if ch == '\\' && i+1 < len(text) {
				i++
				result.WriteByte(text[i])
			} else if ch == inString {
				inString = 0
			}
			i++
			continue
		}

		// String literal start
		if ch == '"' || ch == '\'' || ch == '`' {
			inString = ch
			result.WriteByte(ch)
			lastWasSpace = false
			i++
			continue
		}

		// Skip single-line comments
		if i+1 < len(text) && ch == '/' && text[i+1] == '/' {
			i += 2
			for i < len(text) && text[i] != '\n' {
				i++
			}
			continue
		}
		// Skip multi-line comments
		if i+1 < len(text) && ch == '/' && text[i+1] == '*' {
			i += 2
			for i+1 < len(text) && (text[i] != '*' || text[i+1] != '/') {
				i++
			}
			if i+1 < len(text) {
				i += 2
			}
			continue
		}

		// Collapse whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if !lastWasSpace && result.Len() > 0 {
				result.WriteByte(' ')
				lastWasSpace = true
			}
			i++
			continue
		}

		result.WriteByte(ch)
		lastWasSpace = false
		i++
	}
	return strings.TrimSpace(result.String())
}
