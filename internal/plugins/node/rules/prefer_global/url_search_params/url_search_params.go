package url_search_params

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed url_search_params.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/url-search-params.js
var URLSearchParamsRule = rule.Rule{
	Name:   "node/prefer-global/url-search-params",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		preferModule := len(options) > 0 && options[0] == "never"
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			message := rule.RuleMessage{
				Id:          "preferGlobal",
				Description: "Unexpected use of 'require(\"url\").URLSearchParams'. Use the global variable 'URLSearchParams' instead.",
			}
			if preferModule {
				message = rule.RuleMessage{
					Id:          "preferModule",
					Description: "Unexpected use of the global variable 'URLSearchParams'. Use 'require(\"url\").URLSearchParams' instead.",
				}
			}
			var nodes []*ast.Node
			trace := &nodeutil.ReferenceTrace{Read: func(node *ast.Node) {
				nodes = append(nodes, node)
			}}
			tracker := nodeutil.NewReferenceTracker(ctx)
			if preferModule {
				tracker.TrackGlobals(map[string]*nodeutil.ReferenceTrace{"URLSearchParams": trace})
			} else {
				tracker.TrackModules(map[string]*nodeutil.ReferenceTrace{
					"url": {Properties: map[string]*nodeutil.ReferenceTrace{"URLSearchParams": trace}},
				})
			}
			slices.SortStableFunc(nodes, func(a, b *ast.Node) int { return cmp.Compare(a.Pos(), b.Pos()) })
			for _, node := range nodes {
				ctx.ReportNode(node, message)
			}
		}}
	},
}
