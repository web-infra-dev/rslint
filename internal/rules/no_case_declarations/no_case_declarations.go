package no_case_declarations

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var unexpectedMessage = rule.RuleMessage{
	Id:          "unexpected",
	Description: "Unexpected lexical declaration in case clause.",
}

// https://eslint.org/docs/latest/rules/no-case-declarations
var NoCaseDeclarationsRule = rule.Rule{
	Name: "no-case-declarations",
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
					if clause == nil || clause.Statements == nil {
						continue
					}
					for _, statement := range clause.Statements.Nodes {
						if isLexicalDeclaration(statement) {
							ctx.ReportNode(statement, unexpectedMessage)
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
			// NOTE: We also check `using` (TC39 Stage 3) since it has the same block-scoping
			// semantics as let/const. ESLint does not check `using` yet.
			return varStmt.DeclarationList.Flags&ast.NodeFlagsBlockScoped != 0
		}
	}
	return false
}
