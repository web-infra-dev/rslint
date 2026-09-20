package console

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed console.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-global/console.js
var PreferGlobalConsoleRule = rule.Rule{
	Name:   "node/prefer-global/console",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		preferModule := len(options) > 0 && options[0] == "never"
		message := rule.RuleMessage{
			Id:          "preferGlobal",
			Description: "Unexpected use of 'require(\"console\")'. Use the global variable 'console' instead.",
		}
		if preferModule {
			message = rule.RuleMessage{
				Id:          "preferModule",
				Description: "Unexpected use of the global variable 'console'. Use 'require(\"console\")' instead.",
			}
		}
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			var references []*ast.Node
			trace := map[string]*nodeutil.ReferenceTrace{
				"console": {Read: func(node *ast.Node) {
					// JSDoc types are comments in ESLint, not global reads.
					if !utils.IsJSDocSyntaxNode(node) && !ast.IsExclusivelyTypeOnlyImportOrExport(node) {
						references = append(references, node)
					}
				}},
			}
			tracker := nodeutil.NewReferenceTracker(ctx)
			if preferModule {
				tracker.TrackGlobals(trace)
			} else {
				tracker.TrackModules(trace)
			}
			slices.SortStableFunc(references, func(a, b *ast.Node) int { return cmp.Compare(a.Pos(), b.Pos()) })
			for _, node := range references {
				ctx.ReportNode(node, message)
			}
		}}
	},
}
