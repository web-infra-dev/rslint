package utils

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// IsCommonJSModule identifies the implementation's format independently of a
// generic declaration file, which may describe ESM or live in @types.
func IsCommonJSModule(ctx rule.RuleContext, source modules.Source) bool {
	p, path := commonJSImplementation(ctx, source)
	return isCommonJSImplementation(p, path)
}

// HasCommonJSNamespaceDefault exempts native Node defaults and declarations
// describing the CommonJS namespace. An authored ES default retains the normal
// named-member check: transpilers may unwrap it through __esModule, and deciding
// that from arbitrary CommonJS assignments would require executing the module.
func HasCommonJSNamespaceDefault(ctx rule.RuleContext, source modules.Source) bool {
	p := ctx.Program()
	runtime, path := commonJSImplementation(ctx, source)
	if !isCommonJSImplementation(runtime, path) {
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

// Only a declaration-backed import needs external runtime evidence. Ordinary
// source imports retain the compiler's AST, module graph and export semantics.
func commonJSImplementation(ctx rule.RuleContext, source modules.Source) (*program.Program, string) {
	p := ctx.Program()
	if !p.IsValid() || !p.OwnsSourceFile(ctx.SourceFile) {
		return p, ""
	}
	link := resolveExportLinkForLookup(p, ctx.SourceFile, IndexFor(ctx).settings, source)
	if !link.Resolved {
		return p, ""
	}
	if link.Target != nil && link.Target.IsDeclarationFile {
		request := source.Specifier().Text()
		_, subpath := module.ParsePackageName(request)
		if tspath.IsDeclarationFileName(request) && (tspath.IsExternalModuleNameRelative(request) || strings.HasPrefix(request, "#") || subpath != "") {
			return p, link.Path
		}
		runtime := commonJSRuntimeView(ctx)
		mode := p.GetModeForUsageLocation(ctx.SourceFile, source.Specifier())
		if path := resolveRuntimeImplementation(runtime, request, ctx.SourceFile.FileName(), mode); path != "" && !tspath.IsDeclarationFileName(path) {
			return runtime, path
		}
	}
	return p, link.Path
}

type commonJSRuntimeViewKey struct{}

// The existing rootless Program provides package lookup, resolution and caches
// without copying compiler ASTs or constructing another type-checking graph.
// Its lifetime is one file's lint pass, so a reused editor Program cannot retain
// old implementation text, negative lookups or package metadata.
func commonJSRuntimeView(ctx rule.RuleContext) *program.Program {
	return rule.CachedByFile(ctx, commonJSRuntimeViewKey{}, func() *program.Program {
		p := ctx.Program()
		host := compiler.NewCompilerHost(p.CurrentDirectory(), p.ExternalFileSystem(), p.DefaultLibraryPath(), nil, nil, nil)
		runtime, _ := program.NewFromRoots(program.RootOptions{
			Host: host, CompilerOptions: p.OptionsForFile(ctx.SourceFile), SingleThreaded: true,
		})
		return runtime
	})
}

type commonJSImplementationKey string
type implementationSourceKey string

func implementationSource(p *program.Program, path string) *ast.SourceFile {
	if file := p.GetSourceFile(path); file != nil {
		return file
	}
	return program.Cached(p, implementationSourceKey(path), func() *ast.SourceFile {
		text, ok := p.FS().ReadFile(path)
		if !ok {
			return nil
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
