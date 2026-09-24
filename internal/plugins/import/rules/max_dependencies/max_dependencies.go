package max_dependencies

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed max_dependencies.schema.json
var schemaJSON []byte

// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/max-dependencies.js
var MaxDependenciesRule = rule.Rule{
	Name:   "import/max-dependencies",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		limit := float64(10)
		ignoreTypeImports := false
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				switch value := option["max"].(type) {
				case float64:
					limit = value
				case int:
					limit = float64(value)
				}
				ignoreTypeImports, _ = option["ignoreTypeImports"].(bool)
			}
		}

		dependencies := make(map[string]struct{})
		var lastSource *ast.Node
		listeners := import_utils.VisitModules(func(source, node *ast.Node) {
			// Even ignored type imports and duplicate paths become the report site.
			lastSource = source
			if ignoreTypeImports && ast.IsImportDeclaration(node) &&
				node.ImportClause() != nil && node.ImportClause().IsTypeOnly() {
				return
			}
			// Once over the limit, only the last source can affect the report.
			if float64(len(dependencies)) <= limit {
				dependencies[source.Text()] = struct{}{}
			}
		}, import_utils.VisitModulesOptions{Commonjs: true, ESModule: true})
		listeners[rule.ListenerOnExit(ast.KindEndOfFile)] = func(_ *ast.Node) {
			// A negative max with no module source crashes upstream. There is no
			// dependency to report in that case, so leave such files unreported.
			if lastSource != nil && float64(len(dependencies)) > limit {
				ctx.ReportNode(lastSource, rule.RuleMessage{
					Description: "Maximum number of dependencies (" + ecmascript.NumberToString(limit) + ") exceeded.",
				})
			}
		}
		return listeners
	},
}
