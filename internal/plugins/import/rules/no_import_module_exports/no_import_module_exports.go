package no_import_module_exports

import (
	_ "embed"
	"runtime"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
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
		var exceptionPatterns []string
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				exceptionPatterns = utils.ToStringSlice(option["exceptions"])
			}
		}
		exceptions := compileExceptionMatchers(exceptionPatterns, runtime.GOOS == "windows")

		isException := false
		for _, exception := range exceptions {
			if exception.Match(fileName) {
				isException = true
				break
			}
		}

		references := scopeanalysis.References(ctx, map[string]struct{}{"exports": {}, "module": {}})
		resolvedReferences := make(map[*ast.Node]bool, len(references.References))
		for _, reference := range references.References {
			resolvedReferences[reference.Identifier] = reference.Resolved() != nil
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
			return (name == "module" || name == "exports") && !resolvedReferences[object]
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

func compileExceptionMatchers(patterns []string, windows bool) []*minimatch3.Matcher {
	matchers := make([]*minimatch3.Matcher, 0, len(patterns))
	for _, pattern := range patterns {
		if windows {
			pattern = strings.ReplaceAll(pattern, `\`, "/")
		}
		matchers = append(matchers, minimatch3.New(pattern, minimatch3.Options{}))
	}
	return matchers
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
	return resolveNodePackageEntry(program, pkg)
}

func resolveNodePackageEntry(program *program.Program, pkg *packagejson.Package) string {
	if main, ok := pkg.Field("main").(string); ok && main != "" {
		if runtime.GOOS == "windows" {
			main = strings.ReplaceAll(main, `\`, "/")
		}
		mainPath := tspath.ResolvePath(pkg.Directory(), main)
		if entry := probeNodeFile(program, mainPath); entry != "" {
			return entry
		}
		if entry := probeNodeFile(program, tspath.ResolvePath(mainPath, "index")); entry != "" {
			return entry
		}
	}
	return probeNodeFile(program, tspath.ResolvePath(pkg.Directory(), "index"))
}

func probeNodeFile(program *program.Program, base string) string {
	for _, extension := range []string{"", ".js", ".json", ".node"} {
		candidate := base + extension
		if !program.FileExists(candidate) {
			continue
		}
		if realPath := program.FS().Realpath(candidate); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
		return tspath.NormalizePath(candidate)
	}
	return ""
}
