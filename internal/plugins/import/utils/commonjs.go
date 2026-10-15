package utils

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/collections"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/tailscale/hujson"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// IsCommonJSModule identifies CommonJS implementations without inferring their
// format from a generic .d.ts file, which may describe an ESM entry or live in a
// separate @types package. Unknown implementations retain normal export checks.
func IsCommonJSModule(ctx rule.RuleContext, source modules.Source) bool {
	p := ctx.Program()
	if !p.IsValid() || !p.OwnsSourceFile(ctx.SourceFile) {
		return false
	}
	link := resolveExportLinkForLookup(p, ctx.SourceFile, IndexFor(ctx).settings, source)
	if !link.Resolved {
		return false
	}
	if link.Target != nil && link.Target.IsDeclarationFile {
		request := source.Specifier().Text()
		_, subpath := module.ParsePackageName(request)
		if tspath.IsDeclarationFileName(request) && (tspath.IsExternalModuleNameRelative(request) || strings.HasPrefix(request, "#") || subpath != "") {
			// An explicitly requested declaration is not its adjacent .js file,
			// including when package exports substitute it through a wildcard.
			// A bare package name ending in .d.ts does not identify a file.
			return tspath.TryGetExtensionFromPath(link.Path) == tspath.ExtensionDcts
		}
		// Reuse the compiler's paths, custom conditions and request mode, but
		// omit declaration lookup and the implicit "types" export condition.
		// The cached resolver owns only options and a filesystem host, never p.
		resolver := program.Cached(p, implementationResolverKey{}, func() *module.Resolver {
			options := p.Options().Clone()
			options.NoDtsResolution = core.TSTrue
			options.CustomConditions = slices.DeleteFunc(slices.Clone(options.CustomConditions), func(condition string) bool {
				return condition == "types" || strings.HasPrefix(condition, "types@")
			})
			paths := &collections.OrderedMap[string, []string]{}
			for name, targets := range options.Paths.Entries() {
				implementations := slices.DeleteFunc(slices.Clone(targets), tspath.IsDeclarationFileName)
				if len(implementations) != 0 {
					paths.Set(name, implementations)
				}
			}
			options.Paths = paths
			host := rslint_utils.CreateCompilerHost(p.CurrentDirectory(), implementationFS{FS: p.FS(), suffixes: options.ModuleSuffixes})
			return module.NewResolver(host, options, "", "", nil)
		})
		mode := p.GetModeForUsageLocation(ctx.SourceFile, source.Specifier())
		resolved, _ := resolver.ResolveModuleName(source.Specifier().Text(), ctx.SourceFile.FileName(), mode, nil)
		if resolved.IsResolved() && !tspath.IsDeclarationFileName(resolved.ResolvedFileName) {
			if _, declarationMain := declarationMainSource(resolved.ResolvedFileName, p.Options().ModuleSuffixes); !declarationMain {
				return isCommonJSImplementation(p, resolved.ResolvedFileName)
			}
		}
		// Explicit declaration extensions remain useful when only types are
		// installed. A generic declaration's package type is not evidence of
		// the runtime format, even if it declares an explicit default export.
		return tspath.TryGetExtensionFromPath(link.Path) == tspath.ExtensionDcts
	}
	// Explicit file extensions take precedence over their package's type.
	switch tspath.TryGetExtensionFromPath(link.Path) {
	case tspath.ExtensionMjs, tspath.ExtensionMts, tspath.ExtensionDmts:
		return false
	case tspath.ExtensionCjs, tspath.ExtensionCts, tspath.ExtensionDcts:
		return true
	}
	return link.Target != nil && p.SourceFileMetadata(link.Target).PackageJsonType == "commonjs"
}

type implementationResolverKey struct{}
type commonJSImplementationKey string

// NoDtsResolution omits implicit declarations, but explicit paths and
// typesVersions can still redirect it to types or unrelated implementations.
// Keep a runtime-only view without changing the Program's filesystem.
type implementationFS struct {
	vfs.FS
	suffixes []string
}

// Preserve an explicit declaration main as an opaque resolution result. Simply
// hiding it lets tsgo substitute a sibling .js or fall back to index.js, neither
// of which is the selected entry. This virtual path is never read or parsed.
const declarationMainMarker = ".__rslint_declaration_main__"

func declarationMainSource(name string, suffixes []string) (string, bool) {
	index := strings.LastIndex(name, declarationMainMarker)
	if index < 0 {
		return "", false
	}
	original, ending := name[:index], name[index+len(declarationMainMarker):]
	if ending == ".js" {
		return original, true
	}
	for _, suffix := range suffixes {
		if ending == suffix+".js" {
			extension := tspath.TryGetExtensionFromPath(original)
			return strings.TrimSuffix(original, extension) + suffix + extension, true
		}
	}
	return "", false
}

func (fs implementationFS) FileExists(name string) bool {
	if original, ok := declarationMainSource(name, fs.suffixes); ok {
		return fs.FS.FileExists(original)
	}
	return !tspath.IsDeclarationFileName(name) && fs.FS.FileExists(name)
}

func (fs implementationFS) Realpath(name string) string {
	if original, ok := declarationMainSource(name, fs.suffixes); ok {
		return fs.FS.Realpath(original) + declarationMainMarker + ".js"
	}
	return fs.FS.Realpath(name)
}

func (fs implementationFS) ReadFile(name string) (string, bool) {
	text, ok := fs.FS.ReadFile(name)
	if !ok || tspath.GetBaseFileName(name) != "package.json" {
		return text, ok
	}
	value, err := hujson.Parse([]byte(text))
	if err != nil {
		return text, true
	}
	object, ok := value.Value.(*hujson.Object)
	if !ok {
		return text, true
	}
	object.Members = slices.DeleteFunc(object.Members, func(member hujson.ObjectMember) bool {
		name, ok := member.Name.Value.(hujson.Literal)
		return ok && name.String() == "typesVersions"
	})
	for i := range object.Members {
		member := &object.Members[i]
		name, ok := member.Name.Value.(hujson.Literal)
		if !ok {
			continue
		}
		key := name.String()
		if target, ok := member.Value.Value.(hujson.Literal); key == "main" && ok && target.Kind() == '"' && tspath.IsDeclarationFileName(target.String()) {
			member.Value.Value = hujson.String(target.String() + declarationMainMarker + ".js")
		}
		if key == "exports" || key == "imports" {
			original, wasLiteral := member.Value.Value.(hujson.Literal)
			blockDeclarationTargets(&member.Value)
			if key == "exports" && wasLiteral && original.Kind() == '"' && tspath.IsDeclarationFileName(original.String()) {
				// tsgo treats a root null as absent; preserve the blocked entry.
				member.Value.Value = &hujson.Object{Members: []hujson.ObjectMember{{
					Name: hujson.Value{Value: hujson.String(".")}, Value: hujson.Value{Value: hujson.Literal("null")},
				}}}
			}
		}
	}
	return string(value.Pack()), true
}

func blockDeclarationTargets(value *hujson.Value) {
	switch target := value.Value.(type) {
	case hujson.Literal:
		if target.Kind() == '"' && tspath.IsDeclarationFileName(target.String()) {
			value.Value = hujson.Literal("null")
		}
	case *hujson.Object:
		for i := range target.Members {
			blockDeclarationTargets(&target.Members[i].Value)
		}
	case *hujson.Array:
		for i := range target.Elements {
			blockDeclarationTargets(&target.Elements[i])
		}
	}
}

func isCommonJSImplementation(p *program.Program, fileName string) bool {
	return program.Cached(p, commonJSImplementationKey(fileName), func() bool {
		switch tspath.TryGetExtensionFromPath(fileName) {
		case tspath.ExtensionMjs, tspath.ExtensionMts:
			return false
		case tspath.ExtensionCjs, tspath.ExtensionCts:
			return true
		case tspath.ExtensionJson:
			// JSON also exposes a default object rather than named ESM exports.
			return true
		case tspath.ExtensionJs, tspath.ExtensionJsx, tspath.ExtensionTs, tspath.ExtensionTsx:
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
		// Untyped .js packages may contain ESM syntax (including Node's syntax
		// detection). Do not equate a missing package type with CommonJS.
		file := p.GetSourceFile(fileName)
		if file == nil {
			text, ok := p.FS().ReadFile(fileName)
			if !ok {
				return false
			}
			file = parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: fileName}, text, core.GetScriptKindFromFileName(fileName))
		}
		// Forced module detection and automatic JSX detection install synthetic
		// markers. Only authored imports, exports and import.meta imply ESM.
		indicator := file.ExternalModuleIndicator
		return indicator == nil || !ast.IsExternalModuleIndicator(indicator) && !ast.IsImportMeta(indicator)
	})
}
