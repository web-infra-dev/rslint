package no_case_declarations

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var unexpectedMessage = rule.RuleMessage{
	Id:          "unexpected",
	Description: "Unexpected lexical declaration in case block.",
}

var addBracketsMessage = rule.RuleMessage{
	Id:          "addBrackets",
	Description: "Add {} brackets around the case block.",
}

// https://eslint.org/docs/latest/rules/no-case-declarations
var NoCaseDeclarationsRule = rule.Rule{
	Name:   "no-case-declarations",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCaseBlock: func(node *ast.Node) {
				caseBlock := node.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}

				// Scan a switch in one callback instead of dispatching one listener
				// for every case/default clause.
				for _, clauseNode := range caseBlock.Clauses.Nodes {
					clause := clauseNode.AsCaseOrDefaultClause()
					if clause == nil || clause.Statements == nil || len(clause.Statements.Nodes) == 0 {
						continue
					}

					statements := clause.Statements.Nodes
					firstStatement := statements[0]
					lastStatement := statements[len(statements)-1]
					buildSuggestion := func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message: addBracketsMessage,
							FixesArr: []rule.RuleFix{
								rule.RuleFixInsertBefore(ctx.SourceFile, firstStatement, "{ "),
								rule.RuleFixInsertAfter(lastStatement, " }"),
							},
						}}
					}

					for _, statement := range statements {
						if isLexicalDeclaration(statement) {
							ctx.ReportNodeWithDeferredSuggestions(statement, unexpectedMessage, buildSuggestion)
						}
					}
				}
			},
		}
	},
}

// isLexicalDeclaration checks if a statement is a lexical declaration
// (let, const, class, function declaration)
func isLexicalDeclaration(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		return true
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		if varStmt != nil && varStmt.DeclarationList != nil {
			// BlockScoped covers let, const, using, and await using, matching
			// upstream's check for every variable declaration kind except var.
			return varStmt.DeclarationList.Flags&ast.NodeFlagsBlockScoped != 0
		}
	}
	return false
}
