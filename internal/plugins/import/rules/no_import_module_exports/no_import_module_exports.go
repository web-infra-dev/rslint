package no_import_module_exports

import (
	_ "embed"
	"path"
	"runtime"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
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

		disabled := isException || entryPoint == fileName
		var importDeclarations []*ast.Node
		hasCommonJSExport := false
		checkMember := func(node *ast.Node) {
			if disabled || hasCommonJSExport {
				return
			}
			if utils.IsInJsxTagName(node) {
				return
			}
			object, _ := utils.MemberExpressionParts(node)
			object = utils.ESTreeRuntimeExpression(object)
			if object == nil || object.Kind != ast.KindIdentifier {
				return
			}
			name := object.Text()
			if name != "module" && name != "exports" {
				return
			}
			symbol := ctx.Refs.ResolveInFile(object)
			hasCommonJSExport = symbol == nil || symbol.Flags&ast.SymbolFlagsModuleExports != 0
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				if !disabled {
					importDeclarations = append(importDeclarations, node)
				}
			},
			ast.KindPropertyAccessExpression: checkMember,
			ast.KindElementAccessExpression:  checkMember,
			rule.ListenerOnExit(ast.KindEndOfFile): func(_ *ast.Node) {
				if !hasCommonJSExport {
					return
				}
				for _, declaration := range importDeclarations {
					ctx.ReportNode(declaration, rule.RuleMessage{Description: message})
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
	return resolveNodePackageEntryForPlatform(program, pkg, runtime.GOOS == "windows")
}

func resolveNodePackageEntryForPlatform(program *program.Program, pkg *packagejson.Package, windows bool) string {
	if main, ok := pkg.Field("main").(string); ok && main != "" {
		// The Program filesystem uses TypeScript path identity, which treats a
		// backslash as a separator on every platform. On POSIX, Node instead
		// treats it as a literal filename character. Such a target cannot be
		// probed without aliasing it to a different slash-separated file, so
		// conservatively take Node's package-index fallback.
		if !windows && strings.Contains(main, `\`) {
			return probeNodeIndex(program, pkg.Directory())
		}
		mainPath := resolveNodeMainPath(pkg.Directory(), main, windows)
		if entry := probeNodeMainFile(program, mainPath); entry != "" {
			return entry
		}
		if entry := probeNodeIndex(program, mainPath); entry != "" {
			return entry
		}
	}
	return probeNodeIndex(program, pkg.Directory())
}

// resolveNodeMainPath matches the host platform's path.resolve semantics.
// TypeScript paths treat backslashes as separators on every platform, while
// Node treats them as ordinary filename characters on POSIX.
func resolveNodeMainPath(directory string, main string, windows bool) string {
	if windows {
		return tspath.ResolvePath(directory, main)
	}
	if path.IsAbs(main) {
		return path.Clean(main)
	}
	return path.Join(directory, main)
}

func probeNodeMainFile(program *program.Program, base string) string {
	if entry := nodeFile(program, base); entry != "" {
		return entry
	}
	return probeNodeExtensions(program, base)
}

func probeNodeIndex(program *program.Program, directory string) string {
	return probeNodeExtensions(program, tspath.ResolvePath(directory, "index"))
}

func probeNodeExtensions(program *program.Program, base string) string {
	for _, extension := range []string{".js", ".json", ".node"} {
		candidate := base + extension
		if entry := nodeFile(program, candidate); entry != "" {
			return entry
		}
	}
	return ""
}

func nodeFile(program *program.Program, candidate string) string {
	if !program.FileExists(candidate) {
		return ""
	}
	if realPath := program.FS().Realpath(candidate); realPath != "" {
		return tspath.NormalizePath(realPath)
	}
	return tspath.NormalizePath(candidate)
}
