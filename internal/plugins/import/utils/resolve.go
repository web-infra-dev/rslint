package utils

import (
	"maps"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

type importResolver struct {
	name    string
	options moduleresolver.Options
}

// ImportResolver adapts import/resolver settings to existing resolution
// services. Configuration policy stays here; filesystem lookup stays in the
// shared resolver and the Program. It is constructed once per rule invocation.
type ImportResolver struct {
	ctx       rule.RuleContext
	resolvers []importResolver
	core      []string
	err       string
	patterns  *module.ParsedPatterns
}

func NewImportResolver(ctx rule.RuleContext) *ImportResolver {
	r := &ImportResolver{ctx: ctx, core: settingsStringList(ctx.Settings, "import/core-modules")}
	raw := ctx.Settings["import/resolver"]
	if raw == nil || raw == false || raw == "" {
		if legacy := ctx.Settings["import/resolve"]; legacy != nil {
			raw = map[string]any{"node": legacy}
		} else {
			// Unlike upstream's node-only default, let the TypeScript project
			// resolve what node cannot, such as tsconfig paths and `.js` -> `.ts`
			// imports. Node stays first so the paths it finds keep upstream's view.
			raw = []any{"node", "typescript"}
		}
	}
	set := func(name string, config map[string]any) {
		entry := importResolver{name: name}
		if name == "node" || name == "eslint-import-resolver-node" {
			entry.options = nodeResolveOptions(ctx, config)
		}
		// Upstream reduces settings into a Map: the last configuration wins
		// without changing the resolver's original position.
		for i := range r.resolvers {
			if r.resolvers[i].name == name {
				r.resolvers[i] = entry
				return
			}
		}
		r.resolvers = append(r.resolvers, entry)
	}
	var add func(any)
	add = func(raw any) {
		switch value := raw.(type) {
		case string:
			set(value, nil)
		case nil:
			// A null entry contributes no keys to the upstream resolver Map.
		case []any:
			for _, entry := range value {
				add(entry)
			}
		case map[string]any:
			// JSON object settings have no retained key order. Arrays preserve
			// resolver precedence when more than one resolver is configured.
			for _, name := range slices.Sorted(maps.Keys(value)) {
				config, _ := value[name].(map[string]any)
				set(name, config)
			}
		default:
			r.err = "invalid resolver config"
		}
	}
	add(raw)
	return r
}

// Resolve returns a path, success (including builtins), and a configuration
// error. A missing module is not a configuration error. JavaScript resolver
// plugins cannot execute in the native rule runtime.
func (r *ImportResolver) Resolve(source modules.Source) (string, bool, string) {
	return r.ResolveName(source.Specifier().Text(), source)
}

// ResolveName resolves a candidate spelling in the original reference's context.
// Rules comparing paths need this without constructing or mutating AST nodes.
func (r *ImportResolver) ResolveName(name string, source modules.Source) (string, bool, string) {
	if slices.Contains(r.core, name) {
		return "", true, ""
	}
	if r.err != "" {
		return "", false, r.err
	}
	for _, resolver := range r.resolvers {
		switch resolver.name {
		case "node", "eslint-import-resolver-node":
			if modules.IsNodeBuiltin(name) {
				return "", true, ""
			}
			fileName := r.ctx.SourceFile.FileName()
			if !resolver.options.PreserveSymlinks {
				// Node resolves from the real importer directory as well as
				// returning the real target when symlink preservation is off.
				directory := r.ctx.Program().FS().Realpath(tspath.GetDirectoryPath(fileName))
				fileName = tspath.ResolvePath(directory, tspath.GetBaseFileName(fileName))
			}
			result := moduleresolver.Resolve(r.ctx.Program(), name, fileName, resolver.options)
			if result.Error == "" {
				return result.Path, true, ""
			}
		case "typescript", "eslint-import-resolver-typescript":
			if modules.IsNodeBuiltin(name) {
				return "", true, ""
			}
			if path, _, ok := r.ctx.Program().ResolveModuleNameAt(r.ctx.SourceFile, name, source); ok {
				return path, true, ""
			}
			if path, ok := r.resolveTypeScriptFallback(name); ok {
				return path, true, ""
			}
		default:
			return "", false, "unable to load resolver \"" + resolver.name + "\"."
		}
	}
	return "", false, ""
}

// cspell:ignore unrs fesm
// eslint-import-resolver-typescript v4.4.4 resolves with unrs-resolver: tsconfig
// paths first, then enhanced-resolve lookup with the options below. TypeScript
// only loads modules it can type, so it rejects targets such as `./style.css`
// or `@/style.css` that exist on disk; retry those as upstream does.
var typescriptFallbackOptions = moduleresolver.Options{
	Extensions: []string{".ts", ".tsx", ".d.ts", ".js", ".jsx", ".json", ".node"},
	ExtensionAliases: map[string][]string{
		".js":  {".ts", ".tsx", ".d.ts", ".js"},
		".ts":  {".ts", ".d.ts", ".js"},
		".jsx": {".tsx", ".d.ts", ".jsx"},
		".tsx": {".tsx", ".d.ts", ".jsx", ".js"},
		".cjs": {".cts", ".d.cts", ".cjs"},
		".cts": {".cts", ".d.cts", ".cjs"},
		".mjs": {".mts", ".d.mts", ".mjs"},
		".mts": {".mts", ".d.mts", ".mjs"},
	},
	MainFields: []moduleresolver.MainField{
		{Name: []string{"types"}, ForceRelative: true},
		{Name: []string{"typings"}, ForceRelative: true},
		{Name: []string{"fesm2020"}, ForceRelative: true},
		{Name: []string{"fesm2015"}, ForceRelative: true},
		{Name: []string{"esm2020"}, ForceRelative: true},
		{Name: []string{"es2020"}, ForceRelative: true},
		{Name: []string{"module"}, ForceRelative: true},
		{Name: []string{"jsnext:main"}, ForceRelative: true},
		{Name: []string{"main"}, ForceRelative: true},
	},
	Conditions: []string{"types", "import", "esm2020", "es2020", "es2015", "require", "node", "node-addons", "browser", "default"},
	Modules:    []string{"node_modules"},
}

func (r *ImportResolver) resolveTypeScriptFallback(name string) (string, bool) {
	p := r.ctx.Program()
	fileName := r.ctx.SourceFile.FileName()
	directory := p.FS().Realpath(tspath.GetDirectoryPath(fileName))
	fileName = tspath.ResolvePath(directory, tspath.GetBaseFileName(fileName))
	resolve := func(request string) (string, bool) {
		result := moduleresolver.Resolve(p, request, fileName, typescriptFallbackOptions)
		return result.Path, result.Error == "" && result.Path != ""
	}
	options := p.Options()
	if options != nil && options.Paths.Size() > 0 && !tspath.IsExternalModuleNameRelative(name) {
		if r.patterns == nil {
			r.patterns = module.TryParsePatterns(options.Paths)
		}
		// A matched pattern's targets are tried in order before package lookup.
		if pattern := module.MatchPatternOrExact(r.patterns, name); pattern.IsValid() {
			base := options.GetPathsBasePath(p.CurrentDirectory())
			for _, target := range options.Paths.GetOrZero(pattern.Text) {
				if path, ok := resolve(tspath.ResolvePath(base, strings.Replace(target, "*", pattern.MatchedText(name), 1))); ok {
					return path, true
				}
			}
		}
	}
	return resolve(name)
}

func nodeResolveOptions(ctx rule.RuleContext, config map[string]any) moduleresolver.Options {
	opts := moduleresolver.Options{
		Extensions:    []string{".mjs", ".js", ".json", ".node"},
		IgnoreExports: true, LiteralPaths: true, PreserveSymlinks: true,
		MainFields: []moduleresolver.MainField{
			{Name: []string{"module"}, ForceRelative: true},
			// cspell:ignore jsnext
			{Name: []string{"jsnext:main"}, ForceRelative: true},
			{Name: []string{"main"}, ForceRelative: true},
		},
	}
	if _, ok := config["extensions"]; ok {
		opts.Extensions = settingsStringList(config, "extensions")
	}
	if value, ok := config["preserveSymlinks"].(bool); ok {
		opts.PreserveSymlinks = value
	}
	opts.Modules = []string{"node_modules"}
	if value, ok := config["moduleDirectory"].(string); ok {
		opts.Modules = []string{value}
	} else if _, ok := config["moduleDirectory"]; ok {
		opts.Modules = settingsStringList(config, "moduleDirectory")
	}
	cwd := ctx.ProcessCurrentDirectory()
	if cwd == "" {
		cwd = ctx.Program().CurrentDirectory()
	}
	for _, path := range settingsStringList(config, "paths") {
		opts.Modules = append(opts.Modules, tspath.ResolvePath(cwd, path))
	}
	return opts
}
