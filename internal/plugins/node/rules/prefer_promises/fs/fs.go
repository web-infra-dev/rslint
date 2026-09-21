package fs

import (
	"cmp"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-promises/fs.js
var PreferPromisesFSRule = rule.Rule{
	Name:   "node/prefer-promises/fs",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			type diagnostic struct {
				node *ast.Node
				name string
			}
			var diagnostics []diagnostic
			properties := make(map[string]*nodeutil.ReferenceTrace, 30)
			for _, name := range []string{
				"access", "copyFile", "open", "rename", "truncate", "rmdir", "mkdir",
				"readdir", "readlink", "symlink", "lstat", "stat", "link", "unlink",
				"chmod", "lchmod", "lchown", "chown", "utimes", "realpath", "mkdtemp",
				"writeFile", "appendFile", "readFile", "cp", "glob", "lutimes", "opendir",
				"rm", "statfs",
			} {
				properties[name] = &nodeutil.ReferenceTrace{Call: func(node *ast.Node) {
					diagnostics = append(diagnostics, diagnostic{node, name})
				}}
			}
			nodeutil.NewReferenceTracker(ctx).TrackModules(map[string]*nodeutil.ReferenceTrace{
				"fs": {Properties: properties},
			})
			// Module loaders and ESM imports are traced in separate passes.
			slices.SortStableFunc(diagnostics, func(a, b diagnostic) int {
				return cmp.Compare(a.node.Pos(), b.node.Pos())
			})
			for _, diagnostic := range diagnostics {
				ctx.ReportNode(diagnostic.node, rule.RuleMessage{
					Id:          "preferPromises",
					Description: "Use 'fs.promises." + diagnostic.name + "()' instead.",
				})
			}
		}}
	},
}
