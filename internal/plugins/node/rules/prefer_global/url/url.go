package url

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed url.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/url.js
var PreferGlobalURLRule = rule.Rule{
	Name:   "node/prefer-global/url",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		preferModule := len(options) > 0 && options[0] == "never"
		message := rule.RuleMessage{
			Id:          "preferGlobal",
			Description: "Unexpected use of 'require(\"url\").URL'. Use the global variable 'URL' instead.",
		}
		if preferModule {
			message = rule.RuleMessage{
				Id:          "preferModule",
				Description: "Unexpected use of the global variable 'URL'. Use 'require(\"url\").URL' instead.",
			}
		}
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			var references []*ast.Node
			url := &nodeutil.ReferenceTrace{Read: func(node *ast.Node) {
				// JSDoc types are comments in ESLint, not global reads.
				if !utils.IsJSDocSyntaxNode(node) {
					references = append(references, node)
				}
			}}
			tracker := nodeutil.NewReferenceTracker(ctx)
			if preferModule {
				tracker.TrackGlobals(map[string]*nodeutil.ReferenceTrace{"URL": url})
			} else {
				tracker.TrackModules(map[string]*nodeutil.ReferenceTrace{
					"url": {Properties: map[string]*nodeutil.ReferenceTrace{"URL": url}},
				})
			}
			slices.SortStableFunc(references, func(a, b *ast.Node) int { return cmp.Compare(a.Pos(), b.Pos()) })
			for _, node := range references {
				ctx.ReportNode(node, message)
			}
		}}
	},
}
