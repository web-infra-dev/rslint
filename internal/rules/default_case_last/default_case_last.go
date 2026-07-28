package default_case_last

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/default-case-last
var DefaultCaseLastRule = rule.Rule{
	Name: "default-case-last",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindDefaultClause: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindCaseBlock {
					return
				}

				caseBlock := parent.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}

				clauses := caseBlock.Clauses.Nodes
				if len(clauses) == 0 || clauses[len(clauses)-1] == node {
					return
				}

				for _, clause := range clauses {
					if clause.Kind != ast.KindDefaultClause {
						continue
					}
					if clause != node {
						return
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "notLast",
						Description: "Default clause should be the last clause.",
					})
					return
				}
			},
		}
	},
}
