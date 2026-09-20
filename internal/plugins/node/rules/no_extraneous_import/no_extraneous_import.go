package no_extraneous_import

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
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
		pkg := packagejson.FindNearestValid(p, fileName)
		if pkg == nil {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		allowed := nodeutil.StringListSetting("allowModules", opts, ctx.Settings)
		var resolutionOptions [2]*moduleresolver.Options
		extraneousName := func(specifier string, typeOnly bool) string {
			name, resource := nodeutil.ImportModuleName(specifier)
			if name == "" || nodeutil.AllowsDependency(p, pkg, name) || slices.Contains(allowed, name) ||
				typeOnly && nodeutil.AllowsDependency(p, pkg, module.GetTypesPackageName(name)) ||
				nodeutil.HasTypeScriptAlias(p, fileName, resource) {
				return ""
			}
			index := 0
			if typeOnly {
				index = 1
			}
			if resolutionOptions[index] == nil {
				resolution := nodeutil.ImportResolutionOptions(ctx, typeOnly, opts)
				resolutionOptions[index] = &resolution
			}
			if moduleresolver.ResolveModule(p, resource, fileName, *resolutionOptions[index]) != "" {
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
		return nodeutil.VisitImports(ctx, nodeutil.ImportVisitorOptions{}, func(source *ast.Node, specifier string, typeOnly bool) {
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
		})
	},
}
