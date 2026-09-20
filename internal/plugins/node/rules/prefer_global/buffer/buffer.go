package buffer

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/referencetracker"
)

//go:embed buffer.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/buffer.js
var PreferGlobalBufferRule = rule.Rule{
	Name:   "node/prefer-global/buffer",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		preferModule := len(options) > 0 && options[0] == "never"
		message := rule.RuleMessage{
			Id:          "preferGlobal",
			Description: "Unexpected use of 'require(\"buffer\").Buffer'. Use the global variable 'Buffer' instead.",
		}
		if preferModule {
			message = rule.RuleMessage{
				Id:          "preferModule",
				Description: "Unexpected use of the global variable 'Buffer'. Use 'require(\"buffer\").Buffer' instead.",
			}
		}
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			var references []*ast.Node
			trace := map[string]*nodeutil.ReferenceTrace{
				"Buffer": {Read: func(node *ast.Node) {
					references = append(references, node)
				}},
			}
			if preferModule {
				referencetracker.New(ctx).TrackGlobals(trace)
			} else {
				nodeutil.NewReferenceTracker(ctx).TrackModules(map[string]*nodeutil.ReferenceTrace{
					"buffer": {Properties: trace},
				})
			}
			// Tracking visits entry points separately; diagnostics follow source order.
			slices.SortStableFunc(references, func(a, b *ast.Node) int {
				return cmp.Compare(a.Pos(), b.Pos())
			})
			for _, node := range references {
				ctx.ReportNode(node, message)
			}
		}}
	},
}
