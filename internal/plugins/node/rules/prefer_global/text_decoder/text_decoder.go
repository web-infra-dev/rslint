package text_decoder

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed text_decoder.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/text-decoder.js
var TextDecoderRule = rule.Rule{
	Name:   "node/prefer-global/text-decoder",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		preferModule := len(options) > 0 && options[0] == "never"
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			message := rule.RuleMessage{
				Id:          "preferGlobal",
				Description: "Unexpected use of 'require(\"util\").TextDecoder'. Use the global variable 'TextDecoder' instead.",
			}
			if preferModule {
				message = rule.RuleMessage{
					Id:          "preferModule",
					Description: "Unexpected use of the global variable 'TextDecoder'. Use 'require(\"util\").TextDecoder' instead.",
				}
			}
			var nodes []*ast.Node
			trace := &nodeutil.ReferenceTrace{Read: func(node *ast.Node) {
				nodes = append(nodes, node)
			}}
			tracker := nodeutil.NewReferenceTracker(ctx)
			if preferModule {
				tracker.TrackGlobals(map[string]*nodeutil.ReferenceTrace{"TextDecoder": trace})
			} else {
				tracker.TrackModules(map[string]*nodeutil.ReferenceTrace{
					"util": {Properties: map[string]*nodeutil.ReferenceTrace{"TextDecoder": trace}},
				})
			}
			slices.SortStableFunc(nodes, func(a, b *ast.Node) int { return cmp.Compare(a.Pos(), b.Pos()) })
			for _, node := range nodes {
				ctx.ReportNode(node, message)
			}
		}}
	},
}
