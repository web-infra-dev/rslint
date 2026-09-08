package no_extraneous_import

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_extraneous_import.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-extraneous-import.js
var NoExtraneousImportRule = rule.Rule{
	Name:   "node/no-extraneous-import",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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
		allowed := nodeutil.StringListSetting("allowModules", opts, ctx.Settings)
		var resolutionOptions [2]*program.NodeResolutionOptions
		extraneousName := func(specifier string, typeOnly bool) string {
			name, resource := nodeutil.ImportModuleName(specifier)
			if name == "" || pkg.AllowsDependency(p, name) || slices.Contains(allowed, name) ||
				typeOnly && pkg.AllowsDependency(p, module.GetTypesPackageName(name)) ||
				nodeutil.HasTypeScriptAlias(p, fileName, resource) {
				return ""
			}
			index := 0
			if typeOnly {
				index = 1
			}
			if resolutionOptions[index] == nil {
				resolution := nodeutil.ImportResolutionOptions(p, fileName, typeOnly, opts, ctx.Settings)
				resolutionOptions[index] = &resolution
			}
			if p.ResolveNodeModule(resource, fileName, *resolutionOptions[index]) != "" {
				return name
			}
			return ""
		}
		type targetKey struct {
			specifier string
			typeOnly  bool
		}
		// Resolution is immutable within a file, but every occurrence still
		// needs its own diagnostic. Type-only and runtime imports may differ.
		targets := map[targetKey]string{}
		check := func(source *ast.Node, typeOnly bool) {
			if source == nil {
				return
			}
			var specifier string
			switch source.Kind {
			case ast.KindStringLiteral:
				specifier = source.Text()
			case ast.KindBigIntLiteral:
				specifier = utils.NormalizeBigIntLiteral(source.Text())
			case ast.KindNumericLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
				var ok bool
				specifier, ok = utils.NewStaticStringEvaluatorWithoutScope().EvalToString(source)
				if !ok {
					return
				}
			default:
				return
			}
			key := targetKey{specifier, typeOnly}
			name, found := targets[key]
			if !found {
				name = extraneousName(specifier, typeOnly)
				targets[key] = name
			}
			if name != "" {
				ctx.ReportNode(source, rule.RuleMessage{
					Id: "extraneous", Description: `"` + name + `" is extraneous.`,
					Data: map[string]string{"moduleName": name},
				})
			}
		}
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				check(declaration.ModuleSpecifier, declaration.ImportClause != nil && declaration.ImportClause.AsImportClause().IsTypeOnly())
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				declaration := node.AsExportDeclaration()
				check(declaration.ModuleSpecifier, declaration.IsTypeOnly)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Expression.Kind == ast.KindImportKeyword && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					// ESTree strips parentheses, but preserves template literals and
					// authored TS wrappers. Type import expressions are a different AST.
					check(utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0]), false)
				}
			},
		}
	},
}
