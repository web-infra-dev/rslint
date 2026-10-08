package no_import_module_exports

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

//go:embed no_import_module_exports.schema.json
var schemaJSON []byte

const message = "Cannot use import declarations in modules that export using CommonJS (module.exports = 'foo' or exports.bar = 'hi')"

// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-import-module-exports.js
var NoImportModuleExportsRule = rule.Rule{
	Name:   "import/no-import-module-exports",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		fileName := tspath.NormalizePath(ctx.SourceFile.FileName())
		if program := ctx.Program(); program != nil {
			fileName = tspath.ResolvePath(program.CurrentDirectory(), fileName)
		}
		entryPoint := packageEntryPoint(ctx)
		var exceptions []*minimatch3.Matcher
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				for _, pattern := range utils.ToStringSlice(option["exceptions"]) {
					exceptions = append(exceptions, minimatch3.New(pattern, minimatch3.Options{}))
				}
			}
		}

		isException := false
		for _, exception := range exceptions {
			if exception.Match(fileName) {
				isException = true
				break
			}
		}

		hasCommonJSExport := false
		checkMember := func(node *ast.Node) bool {
			if utils.IsInJsxTagName(node) {
				return false
			}
			object, _ := utils.MemberExpressionParts(node)
			object = utils.ESTreeRuntimeExpression(object)
			if object == nil || object.Kind != ast.KindIdentifier {
				return false
			}
			name := object.Text()
			return (name == "module" || name == "exports") && isCommonJSExportReference(ctx, name)
		}

		if !isException && entryPoint != fileName {
			utils.VisitDescendants(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
				if !hasCommonJSExport && (node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression) {
					hasCommonJSExport = checkMember(node)
				}
				return true
			})
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				if hasCommonJSExport {
					ctx.ReportNode(node, rule.RuleMessage{Description: message})
				}
			},
		}
	},
}

func isCommonJSExportReference(ctx rule.RuleContext, name string) bool {
	manager := scopeanalysis.Declarations(ctx)
	var objectScope *scope.Scope
	for index := len(manager.Scopes) - 1; index >= 0; index-- {
		candidate := manager.Scopes[index]
		if len(candidate.Declarations(name)) > 0 {
			objectScope = candidate
			break
		}
	}
	if objectScope == nil {
		return true
	}
	declarations := objectScope.Declarations(name)
	if len(declarations) > 0 && declarations[0].Kind == scope.DefImport {
		return false
	}
	// rslint collapses eslint-scope's top-level module scope into Global.
	return objectScope == manager.Global && ast.IsExternalModule(ctx.SourceFile)
}

func packageEntryPoint(ctx rule.RuleContext) string {
	program := ctx.Program()
	if program == nil || ctx.SourceFile == nil {
		return ""
	}
	pkg := packagejson.FindNearest(program, ctx.SourceFile.FileName())
	if pkg == nil {
		return ""
	}
	result := moduleresolver.Resolve(program, pkg.Directory(), ctx.SourceFile.FileName(), moduleresolver.Options{
		Extensions:    []string{".js", ".json", ".node"},
		MainFields:    []moduleresolver.MainField{{Name: []string{"main"}, ForceRelative: true}},
		MainFiles:     []string{"index"},
		IgnoreExports: true,
		LiteralPaths:  true,
	})
	if result.Error != "" || result.Path == "" {
		return ""
	}
	return tspath.NormalizePath(result.Path)
}
