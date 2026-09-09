package no_extraneous_require

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_extraneous_require.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-extraneous-require.js
var NoExtraneousRequireRule = rule.Rule{
	Name:   "node/no-extraneous-require",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// tsgo records normalized identifiers, including escaped spellings.
		// Include global objects because their require property can be computed.
		mayRequire := false
		for _, name := range []string{"require", "global", "globalThis", "self", "window"} {
			if ctx.SourceFile.HasIdentifier(name) {
				mayRequire = true
				break
			}
		}
		if !mayRequire {
			return nil
		}
		p := ctx.Program()
		fileName := ctx.SourceFile.FileName()
		if p == nil || fileName == "<input>" {
			return nil
		}
		pkg := nodeutil.FindPackage(p, fileName)
		if pkg == nil {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				calls := rule.TrackGlobalCalls(ctx, map[string]*rule.GlobalCallTrace{
					"require": {Call: true, Members: map[string]*rule.GlobalCallTrace{
						"resolve": {Call: true},
					}},
				}, rule.GlobalCallOptions{})
				// ESLint presents diagnostics in source order, including when
				// alias traversal reaches a later call before a direct call.
				slices.SortStableFunc(calls, func(a, b *ast.Node) int { return cmp.Compare(a.Pos(), b.Pos()) })
				allowed := nodeutil.StringListSetting("allowModules", opts, ctx.Settings)
				var evaluator *utils.StaticStringEvaluator
				var resolution *nodeutil.ResolutionOptions
				targets := map[string]string{}
				for _, node := range calls {
					args := node.AsCallExpression().Arguments
					if args == nil || len(args.Nodes) == 0 {
						continue
					}
					source := utils.ESTreeRuntimeExpression(args.Nodes[0])
					specifier, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(source))
					if !ok {
						if evaluator == nil {
							evaluator = utils.NewStaticStringEvaluatorWithoutScope()
						}
						specifier, ok = evaluator.EvalToString(source)
					}
					// The shared evaluator cannot fold compound BigInt values;
					// this uncommon module-name form is documented in the rule.
					if !ok {
						continue
					}
					name, found := targets[specifier]
					if !found {
						var resource string
						name, resource = nodeutil.ImportModuleName(specifier)
						if name != "" && (pkg.AllowsDependency(p, name) || slices.Contains(allowed, name) || nodeutil.HasTypeScriptAlias(p, fileName, resource)) {
							name = ""
						}
						if name != "" {
							if resolution == nil {
								value := nodeutil.RequireResolutionOptions(p, fileName, opts, ctx.Settings)
								resolution = &value
							}
							// As with no-extraneous-import, only resolverConfig.modules
							// is supported; other overrides have documented differences.
							if nodeutil.ResolveModule(p, resource, fileName, *resolution) == "" {
								name = ""
							}
						}
						targets[specifier] = name
					}
					if name != "" {
						ctx.ReportNode(source, rule.RuleMessage{
							Id: "extraneous", Description: `"` + name + `" is extraneous.`,
							Data: map[string]string{"moduleName": name},
						})
					}
				}
			},
		}
	},
}
