package default_case_last

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/default-case-last
var DefaultCaseLastRule = rule.Rule{
	Name: "default-case-last",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				switchStmt := node.AsSwitchStatement()
				if switchStmt == nil {
					return
				}

				caseBlock := switchStmt.CaseBlock
				if caseBlock == nil {
					return
				}

				clauses := caseBlock.AsCaseBlock()
				if clauses == nil || clauses.Clauses == nil {
					return
				}

				clauseNodes := clauses.Clauses.Nodes
				if len(clauseNodes) == 0 {
					return
				}

				// Find the default clause
				for i, clause := range clauseNodes {
					if clause.Kind == ast.KindDefaultClause {
						// If it's not the last clause, report
						if i != len(clauseNodes)-1 {
							ctx.ReportNode(clause, rule.RuleMessage{
								Id:          "notLast",
								Description: "Default clause should be the last clause.",
							})
						}
						return
					}
				}
			},
		}
	},
}
