package process

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed process.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/process.js
var ProcessRule = rule.Rule{
	Name:   "node/prefer-global/process",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		preferModule := len(options) > 0 && options[0] == "never"
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			var nodes []*ast.Node
			traces := map[string]*nodeutil.ReferenceTrace{
				"process": {Read: func(node *ast.Node) { nodes = append(nodes, node) }},
			}
			tracker := nodeutil.NewReferenceTracker(ctx)
			var message rule.RuleMessage
			if preferModule {
				message = rule.RuleMessage{
					Id:          "preferModule",
					Description: "Unexpected use of the global variable 'process'. Use 'require(\"process\")' instead.",
				}
				tracker.TrackGlobals(traces)
			} else {
				message = rule.RuleMessage{
					Id:          "preferGlobal",
					Description: "Unexpected use of 'require(\"process\")'. Use the global variable 'process' instead.",
				}
				tracker.TrackModules(traces)
			}
			slices.SortStableFunc(nodes, func(a, b *ast.Node) int { return cmp.Compare(a.Pos(), b.Pos()) })
			for _, node := range nodes {
				ctx.ReportNode(node, message)
			}
		}}
	},
}
