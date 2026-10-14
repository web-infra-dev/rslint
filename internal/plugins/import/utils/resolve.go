package utils

import (
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

type importResolver struct {
	name          string
	options       moduleresolver.Options
	defaultConfig bool
}

// ImportResolver adapts import/resolver settings to existing resolution
// services. Configuration policy stays here; filesystem lookup stays in the
// shared resolver and the Program. It is constructed once per rule invocation.
type ImportResolver struct {
	ctx       rule.RuleContext
	resolvers []importResolver
	core      []string
	err       string
}

func NewImportResolver(ctx rule.RuleContext) *ImportResolver {
	r := &ImportResolver{ctx: ctx, core: settingsStringList(ctx.Settings, "import/core-modules")}
	raw := ctx.Settings["import/resolver"]
	if raw == nil || raw == false || raw == "" {
		raw = map[string]any{"node": ctx.Settings["import/resolve"]}
	}
	set := func(name string, config map[string]any) {
		entry := importResolver{name: name, defaultConfig: len(config) == 0}
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
		default:
			return "", false, "unable to load resolver \"" + resolver.name + "\"."
		}
	}
	return "", false, ""
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

// EquivalentNodeCandidate verifies a plain JavaScript request under default
// Node resolution, independently of the legacy import resolver's exports-blind
// lookup. Unknown runtime policy is a reason to omit a suggestion, not to guess.
func (r *ImportResolver) EquivalentNodeCandidate(source modules.Source, candidate, packageRoot string) bool {
	if len(r.resolvers) != 1 || !r.resolvers[0].defaultConfig || len(r.core) != 0 {
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
		MainFiles:  []string{"index"}, NodeExports: true,
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
