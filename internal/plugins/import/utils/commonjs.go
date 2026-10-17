package utils

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// IsCommonJSModule identifies the implementation's format independently of a
// generic declaration file, which may describe ESM or live in @types.
func IsCommonJSModule(ctx rule.RuleContext, source modules.Source) bool {
	return isCommonJSImplementation(ctx.Program(), implementationPath(ctx, source))
}

// HasCommonJSNamespaceDefault exempts native Node defaults and declarations
// describing the CommonJS namespace. An authored ES default retains the normal
// named-member check: transpilers may unwrap it through __esModule, and deciding
// that from arbitrary CommonJS assignments would require executing the module.
func HasCommonJSNamespaceDefault(ctx rule.RuleContext, source modules.Source) bool {
	p := ctx.Program()
	path := implementationPath(ctx, source)
	if !isCommonJSImplementation(p, path) {
		return false
	}
	options := p.OptionsForFile(ctx.SourceFile)
	kind := options.GetEmitModuleKind()
	nodeModule := kind >= core.ModuleKindNode16 && kind <= core.ModuleKindNodeNext
	mode := p.GetModeForUsageLocation(ctx.SourceFile, source.Specifier())
	if nodeModule && mode == core.ResolutionModeESM {
		// Node always exposes module.exports itself, including __esModule.
		return true
	}
	if (kind == core.ModuleKindCommonJS || nodeModule && mode == core.ResolutionModeCommonJS) &&
		!compilerOptionsESModuleInteropForOptions(options) {
		// Without __importDefault, TypeScript emits require(...).default.
		return false
	}
	index := IndexFor(ctx)
	link := resolveExportLinkForLookup(p, ctx.SourceFile, index.settings, source)
	if link.Target == nil || !link.Target.IsDeclarationFile {
		return false
	}
	// Get returns positively known declarations, including re-exports. Has
	// deliberately treats unknown star targets as exporting every name, which
	// is useful for validation but is not evidence of an interop marker.
	exports, available := GetExportMap(ctx, source)
	return available && exports.Get(defaultExportName) == nil && exports.Get("__esModule") == nil
}

func implementationPath(ctx rule.RuleContext, source modules.Source) string {
	p := ctx.Program()
	if !p.IsValid() || !p.OwnsSourceFile(ctx.SourceFile) {
		return ""
	}
	link := resolveExportLinkForLookup(p, ctx.SourceFile, IndexFor(ctx).settings, source)
	if !link.Resolved {
		return ""
	}
	if link.Target != nil && link.Target.IsDeclarationFile {
		request := source.Specifier().Text()
		_, subpath := module.ParsePackageName(request)
		if tspath.IsDeclarationFileName(request) && (tspath.IsExternalModuleNameRelative(request) || strings.HasPrefix(request, "#") || subpath != "") {
			return link.Path
		}
		if path := moduleresolver.ResolveImplementation(p, ctx.SourceFile, source); path != "" && !tspath.IsDeclarationFileName(path) {
			return path
		}
	}
	return link.Path
}

type commonJSImplementationKey string
type implementationSourceKey string

// Runtime inspection reads the generation's filesystem even when the compiler
// already loaded this path. An editor can retain the old dependency AST while
// a symlink alias supplies newer, unsaved text in the generation's overlay.
func implementationSource(p *program.Program, path string) *ast.SourceFile {
	return program.Cached(p, implementationSourceKey(path), func() *ast.SourceFile {
		text, ok := p.FS().ReadFile(path)
		if !ok {
			return nil
		}
		if file := p.GetSourceFile(path); file != nil && file.Text() == text {
			return file
		}
		kind := core.GetScriptKindFromFileName(path)
		if kind == core.ScriptKindUnknown {
			// Explicit Node package targets may be extensionless JavaScript.
			kind = core.ScriptKindJS
		}
		return parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path}, text, kind)
	})
}

func authoredModuleIndicator(file *ast.SourceFile) bool {
	indicator := file.ExternalModuleIndicator
	return indicator != nil && (ast.IsExternalModuleIndicator(indicator) || ast.IsImportMeta(indicator))
}

func isCommonJSImplementation(p *program.Program, fileName string) bool {
	if fileName == "" || !p.IsValid() {
		return false
	}
	return program.Cached(p, commonJSImplementationKey(fileName), func() bool {
		switch tspath.TryGetExtensionFromPath(fileName) {
		case tspath.ExtensionMjs, tspath.ExtensionMts, tspath.ExtensionDmts:
			return false
		case tspath.ExtensionCjs, tspath.ExtensionCts, tspath.ExtensionDcts, tspath.ExtensionJson:
			return true
		case "", tspath.ExtensionJs, tspath.ExtensionJsx, tspath.ExtensionTs, tspath.ExtensionTsx:
		default:
			return false
		}
		if pkg := packagejson.FindNearest(p, fileName); pkg != nil {
			switch pkg.Field("type") {
			case "module":
				return false
			case "commonjs":
				return true
			}
		}
		file := implementationSource(p, fileName)
		// Forced module detection and JSX markers are not authored ESM syntax.
		return file != nil && !authoredModuleIndicator(file)
	})
}
