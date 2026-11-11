package utils

import (
	"encoding/json"
	"maps"
	"path"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

type importResolver struct {
	name          string
	options       *moduleresolver.PreparedOptions
	defaultConfig bool
}

type importResolverConfig struct {
	resolvers []importResolver
	core      []string
	err       string
}

// ImportResolver adapts import/resolver settings to existing resolution
// services. Configuration policy stays here; filesystem lookup stays in the
// shared resolver and the Program. Immutable configuration is shared within a
// generation; the instance keeps the current file and request context.
type ImportResolver struct {
	*importResolverConfig
	ctx      rule.RuleContext
	patterns *module.ParsedPatterns
}

type importResolverConfigKey struct{ settings, cwd string }
type importResolverEnvironmentKey struct{ cwd string }
type importResolverPresetKey struct{ preset, cwd string }

func NewImportResolver(ctx rule.RuleContext) *ImportResolver {
	cwd := ctx.ProcessCurrentDirectory()
	if cwd == "" {
		cwd = ctx.Program().CurrentDirectory()
	}
	settings := ctx.Settings
	config := rule.CachedByEnvironment(ctx, importResolverEnvironmentKey{cwd}, func() *importResolverConfig {
		build := func() *importResolverConfig { return compileImportResolverConfig(settings, cwd) }
		// Only the first file for an immutable environment prepares this key.
		// Unbound Go contexts retain content-based caching and observe settings
		// mutations when constructing another resolver.
		if preset := resolverConfigPreset(settings); preset != "" {
			return rule.CachedByProgram(ctx, importResolverPresetKey{preset, cwd}, build)
		}
		identity := resolverConfigIdentity(configuredResolvers(settings))
		core := settings["import/core-modules"]
		encoded, err := json.Marshal([2]any{identity, core})
		if err == nil && resolverConfigCacheable(identity) && resolverSettingCacheable(core) {
			return rule.CachedByProgram(ctx, importResolverConfigKey{string(encoded), cwd}, build)
		}
		// Preserve non-JSON and custom Go representations without caching an
		// incomplete identity.
		return build()
	})
	return &ImportResolver{ctx: ctx, importResolverConfig: config}
}

// Common option-free configurations have a fixed identity. Direct Go callers
// need not allocate an equivalent JSON key on every rule invocation either.
func resolverConfigPreset(settings map[string]any) string {
	if settings["import/core-modules"] != nil {
		return ""
	}
	raw := settings["import/resolver"]
	if (raw == nil || raw == false || raw == "") && settings["import/resolve"] == nil {
		return "default"
	}
	switch value := raw.(type) {
	case string:
		if value == "typescript" || value == "eslint-import-resolver-typescript" {
			return value
		}
	case map[string]any:
		if len(value) == 1 {
			for _, preset := range [...]struct{ name, configured string }{
				{"typescript", "typescript configured"},
				{"eslint-import-resolver-typescript", "eslint-import-resolver-typescript configured"},
			} {
				if rawConfig, ok := value[preset.name]; ok {
					// Keep defaultConfig in the identity without traversing options.
					if config, _ := rawConfig.(map[string]any); len(config) != 0 {
						return preset.configured
					}
					return preset.name
				}
			}
		}
	}
	return ""
}

// resolverConfigIdentity retains resolver order, whether each resolver has any
// options, and the Node options actually read by the native adapter. The values
// of TypeScript project lists and unknown resolver options do not affect native
// resolution and need not be traversed.
func resolverConfigIdentity(raw any) any {
	switch value := raw.(type) {
	case []any:
		if value == nil {
			return value
		}
		entries := make([]any, len(value))
		for i, entry := range value {
			entries[i] = resolverConfigIdentity(entry)
		}
		return entries
	case map[string]any:
		if value == nil {
			return value
		}
		entries := make(map[string]any, len(value))
		for name, rawConfig := range value {
			config, _ := rawConfig.(map[string]any)
			// Whether any option is present controls suggestion eligibility
			// (importResolver.defaultConfig), even when the options themselves
			// do not affect lookup.
			defaultConfig := len(config) == 0
			if name != "node" && name != "eslint-import-resolver-node" {
				entries[name] = defaultConfig
				continue
			}
			relevant := make(map[string]any)
			for _, option := range [...]string{"extensions", "preserveSymlinks", "moduleDirectory", "paths"} {
				if setting, ok := config[option]; ok {
					relevant[option] = setting
				}
			}
			entries[name] = []any{defaultConfig, relevant}
		}
		return entries
	default:
		return raw
	}
}

// Resolver arrays accept []any, while option lists also accept []string.
// Custom Go types may encode as the same JSON but have a different meaning to
// the settings parser. Only cache the representations whose interpretation is
// preserved by JSON. Marshal has already rejected cyclic settings.
func resolverConfigCacheable(raw any) bool {
	switch value := raw.(type) {
	case nil, bool, float64:
		return true
	case string:
		return utf8.ValidString(value)
	case []any:
		if value == nil {
			return false
		}
		for _, entry := range value {
			if !resolverConfigCacheable(entry) {
				return false
			}
		}
		return true
	case map[string]any:
		return resolverSettingCacheable(value)
	default:
		return false
	}
}

func resolverSettingCacheable(raw any) bool {
	switch value := raw.(type) {
	case nil, bool, float64:
		return true
	case string:
		return utf8.ValidString(value)
	case []string:
		if value == nil {
			return false
		}
		for _, entry := range value {
			if !utf8.ValidString(entry) {
				return false
			}
		}
		return true
	case []any:
		if value == nil {
			return false
		}
		for _, entry := range value {
			if !resolverSettingCacheable(entry) {
				return false
			}
		}
		return true
	case map[string]any:
		if value == nil {
			return false
		}
		for key, entry := range value {
			if !utf8.ValidString(key) || !resolverSettingCacheable(entry) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func configuredResolvers(settings map[string]any) any {
	raw := settings["import/resolver"]
	if raw == nil || raw == false || raw == "" {
		if legacy := settings["import/resolve"]; legacy != nil {
			raw = map[string]any{"node": legacy}
		} else {
			// Unlike upstream's node-only default, let the TypeScript project
			// resolve what node cannot, such as tsconfig paths and `.js` -> `.ts`
			// imports. Node stays first so the paths it finds keep upstream's view.
			raw = []any{"node", "typescript"}
		}
	}
	return raw
}

func compileImportResolverConfig(settings map[string]any, cwd string) *importResolverConfig {
	r := &importResolverConfig{core: slices.Clone(settingsStringList(settings, "import/core-modules"))}
	raw := configuredResolvers(settings)
	set := func(name string, config map[string]any) {
		entry := importResolver{name: name, defaultConfig: len(config) == 0}
		if name == "node" || name == "eslint-import-resolver-node" {
			entry.options = moduleresolver.PrepareOptions(nodeResolveOptions(cwd, config))
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
			if !resolver.options.PreserveSymlinks() {
				// Node resolves from the real importer directory as well as
				// returning the real target when symlink preservation is off.
				directory := r.ctx.Program().FS().Realpath(tspath.GetDirectoryPath(fileName))
				fileName = tspath.ResolvePath(directory, tspath.GetBaseFileName(fileName))
			}
			result := resolver.options.Resolve(r.ctx.Program(), name, fileName)
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
var typescriptFallbackOptions = moduleresolver.PrepareOptions(moduleresolver.Options{
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
})

func (r *ImportResolver) resolveTypeScriptFallback(name string) (string, bool) {
	// Upstream drops the query and fragment before matching paths, so
	// `@icons/x?raw` substitutes `x` rather than `x?raw`. The leading `#` of
	// package imports is part of the name, so the scan starts after it.
	start := 0
	if strings.HasPrefix(name, "#") {
		start = 1
	}
	if i := strings.IndexAny(name[start:], "?#"); i >= 0 {
		name = name[:start+i]
	}
	p := r.ctx.Program()
	fileName := r.ctx.SourceFile.FileName()
	directory := p.FS().Realpath(tspath.GetDirectoryPath(fileName))
	fileName = tspath.ResolvePath(directory, tspath.GetBaseFileName(fileName))
	resolve := func(request string) (string, bool) {
		result := typescriptFallbackOptions.Resolve(p, request, fileName)
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

func nodeResolveOptions(cwd string, config map[string]any) moduleresolver.Options {
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
		opts.Modules = slices.Clone(settingsStringList(config, "moduleDirectory"))
	}
	for _, path := range settingsStringList(config, "paths") {
		opts.Modules = append(opts.Modules, tspath.ResolvePath(cwd, path))
	}
	return opts
}

// EquivalentNodeCandidate verifies a plain JavaScript request under default
// Node resolution, independently of the legacy import resolver's exports-blind
// lookup. Unknown runtime policy is a reason to omit a suggestion, not to guess.
func (r *ImportResolver) EquivalentNodeCandidate(source modules.Source, candidate, packageRoot string) bool {
	configured := r.ctx.Settings["import/resolver"]
	// The implicit TypeScript fallback does not alter this independent Node
	// runtime check. Explicit resolver chains retain their unknown policy.
	defaults := (configured == nil || configured == false || configured == "") && r.ctx.Settings["import/resolve"] == nil
	if r.err != "" || len(r.resolvers) == 0 || !r.resolvers[0].defaultConfig || len(r.core) != 0 {
		return false
	}
	if len(r.resolvers) != 1 && !defaults {
		return false
	}
	switch r.resolvers[0].name {
	case "node", "eslint-import-resolver-node":
	default:
		return false
	}
	// A configured legacy resolver is also outside the default Node contract.
	if legacy := r.ctx.Settings["import/resolve"]; legacy != nil {
		config, ok := legacy.(map[string]any)
		if !ok || len(config) != 0 {
			return false
		}
	}
	switch path.Ext(r.ctx.SourceFile.FileName()) {
	case ".js", ".cjs", ".mjs":
	default:
		return false
	}
	if source.TypeOnly() || source.Kind() == modules.ModuleReferenceAMD || source.Attributes().State != modules.AttributesNone {
		return false
	}
	imported := source.Specifier().Text()
	if strings.ContainsAny(imported+candidate, `\%?#`) {
		return false
	}
	// Do not select a conditional or fallback export without the project's
	// runtime conditions. Flat string targets (including patterns) are verified
	// by the existing resolver; this does not invert exports into guessed aliases.
	pkg := packagejson.Read(r.ctx.Program(), packageRoot)
	if pkg == nil || !plainExports(pkg.Field("exports")) {
		return false
	}
	require := source.Kind() == modules.ModuleReferenceRequire
	if !require && path.Ext(r.ctx.SourceFile.FileName()) == ".cjs" {
		return false
	}
	if require {
		// tsgo can mark .cjs as an external module with a synthetic SourceFile
		// indicator. Only authored ESM syntax excludes Node's ordinary require.
		if path.Ext(r.ctx.SourceFile.FileName()) == ".mjs" || r.ctx.SourceFile.ExternalModuleIndicator != nil && r.ctx.SourceFile.ExternalModuleIndicator.Kind != ast.KindSourceFile {
			return false
		}
		if path.Ext(r.ctx.SourceFile.FileName()) == ".js" {
			owner := packagejson.FindNearest(r.ctx.Program(), r.ctx.SourceFile.FileName())
			if owner != nil && owner.Field("type") == "module" {
				return false
			}
		}
		callee := utils.ESTreeCallCallee(source.Declaration().AsCallExpression().Expression)
		if r.ctx.Refs.ResolveInFile(callee) != nil {
			return false
		}
	}
	condition := "import"
	if require {
		condition = "require"
	}
	options := moduleresolver.Options{
		Extensions: []string{".js", ".json", ".node"},
		Conditions: []string{"node", condition},
		MainFields: []moduleresolver.MainField{{Name: []string{"main"}, ForceRelative: true}},
		MainFiles:  []string{"index"}, NodeMain: true, NodeExports: true,
		FullySpecified: !require, NoDirectory: !require, LiteralPaths: require,
	}
	// Default Node resolves relative to the real importer directory and
	// canonicalizes symlinks. Do not compare two merely equal lexical spellings.
	fileName := r.ctx.SourceFile.FileName()
	directory := r.ctx.Program().FS().Realpath(tspath.GetDirectoryPath(fileName))
	fileName = tspath.ResolvePath(directory, tspath.GetBaseFileName(fileName))
	original := moduleresolver.Resolve(r.ctx.Program(), imported, fileName, options)
	if original.Error != "" || original.Path == "" || original.ResourceSuffix != "" {
		return false
	}
	switch path.Ext(original.Path) {
	case ".js", ".cjs", ".mjs":
	default:
		return false
	}
	proposed := moduleresolver.Resolve(r.ctx.Program(), candidate, fileName, options)
	return proposed.Error == "" && proposed.Path == original.Path && proposed.ResourceSuffix == ""
}

func plainExports(value any) bool {
	switch exports := value.(type) {
	case nil, string:
		return true
	case map[string]any:
		for subpath, target := range exports {
			if subpath != "." && !strings.HasPrefix(subpath, "./") {
				return false
			}
			if _, ok := target.(string); !ok {
				return false
			}
		}
		return true
	default:
		return false
	}
}
