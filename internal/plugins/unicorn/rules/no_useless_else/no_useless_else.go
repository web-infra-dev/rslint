package no_useless_else

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

const messageID = "no-useless-else"

var unexpectedElse = rule.RuleMessage{
	Id:          messageID,
	Description: "Unexpected `else` after a statement that exits.",
}

var NoUselessElseRule = rule.Rule{
	Name:   "unicorn/" + messageID,
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var roots map[*ast.Node]map[*ast.Node]bool
		var analysis *exitAnalysis
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindIfStatement): func(node *ast.Node) {
				statement := node.AsIfStatement()
				if !isCandidate(node) {
					return
				}
				root := cfg.RootOf(node)
				if root == nil {
					return
				}
				if analysis == nil {
					analysis = &exitAnalysis{ctx: &ctx}
					roots = make(map[*ast.Node]map[*ast.Node]bool)
				}
				exits, analyzed := roots[root]
				if !analyzed {
					exits = analysis.analyze(root)
					roots[root] = exits
				}
				if !exits[node] {
					return
				}
				token, ok := utils.TokenAtOrAfter(ctx.SourceFile, statement.ThenStatement.End())
				if !ok || token.Kind != ast.KindElseKeyword {
					return
				}
				ctx.ReportRangeWithDeferredFixes(token.Range(), unexpectedElse, func() []rule.RuleFix {
					return buildFix(&ctx, node)
				})
			},
		}
	},
}

func isCandidate(node *ast.Node) bool {
	stmt := node.AsIfStatement()
	return stmt.ThenStatement != nil && stmt.ElseStatement != nil &&
		node.Parent != nil && node.Parent.Kind != ast.KindModuleBlock && node.Parent.CanHaveStatements()
}
