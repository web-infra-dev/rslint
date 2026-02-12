package no_case_declarations

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-case-declarations
var NoCaseDeclarationsRule = rule.Rule{
	Name: "no-case-declarations",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkClause := func(node *ast.Node) {
			clause := node.AsCaseOrDefaultClause()
			if clause == nil || clause.Statements == nil {
				return
			}

			for _, stmt := range clause.Statements.Nodes {
				if isLexicalDeclaration(stmt) {
					ctx.ReportNode(stmt, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Unexpected lexical declaration in case clause.",
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCaseClause:    checkClause,
			ast.KindDefaultClause: checkClause,
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
			flags := varStmt.DeclarationList.Flags
			if flags&ast.NodeFlagsLet != 0 || flags&ast.NodeFlagsConst != 0 || flags&ast.NodeFlagsUsing != 0 {
				return true
			}
		}
	}
	return false
}
