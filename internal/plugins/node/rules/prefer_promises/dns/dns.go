// cspell:ignore Naptr
package dns

import (
	"cmp"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-promises/dns.js
var PreferPromisesDNSRule = rule.Rule{
	Name:   "node/prefer-promises/dns",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			type diagnostic struct {
				node *ast.Node
				name string
			}
			var diagnostics []diagnostic
			properties := make(map[string]*nodeutil.ReferenceTrace, 18)
			for _, name := range []string{
				"lookup", "lookupService", "getServers", "resolve", "resolve4", "resolve6",
				"resolveAny", "resolveCname", "resolveMx", "resolveNaptr", "resolveNs",
				"resolvePtr", "resolveSoa", "resolveSrv", "resolveTxt", "reverse", "setServers",
			} {
				properties[name] = &nodeutil.ReferenceTrace{Call: func(node *ast.Node) {
					diagnostics = append(diagnostics, diagnostic{node, name})
				}}
			}
			properties["Resolver"] = &nodeutil.ReferenceTrace{Construct: func(node *ast.Node) {
				diagnostics = append(diagnostics, diagnostic{node, "Resolver"})
			}}
			nodeutil.NewReferenceTracker(ctx).TrackModules(map[string]*nodeutil.ReferenceTrace{
				"dns": {Properties: properties},
			})
			// Module loaders and ESM imports are traced in separate passes.
			slices.SortStableFunc(diagnostics, func(a, b diagnostic) int {
				return cmp.Compare(a.node.Pos(), b.node.Pos())
			})
			for _, diagnostic := range diagnostics {
				message := rule.RuleMessage{
					Id:          "preferPromisesNew",
					Description: "Use 'new dns.promises.Resolver()' instead.",
				}
				if diagnostic.name != "Resolver" {
					message = rule.RuleMessage{
						Id:          "preferPromises",
						Description: "Use 'dns.promises." + diagnostic.name + "()' instead.",
					}
				}
				ctx.ReportNode(diagnostic.node, message)
			}
		}}
	},
}
