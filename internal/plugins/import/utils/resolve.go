package utils

import (
	"maps"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
}

func NewImportResolver(ctx rule.RuleContext) *ImportResolver {
	r := &ImportResolver{ctx: ctx, core: settingsStringList(ctx.Settings, "import/core-modules")}
	raw := ctx.Settings["import/resolver"]
	if raw == nil || raw == false || raw == "" {
		raw = map[string]any{"node": ctx.Settings["import/resolve"]}
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
func (r *ImportResolver) Resolve(source *ast.Node) (string, bool, string) {
	return r.ResolveName(source.Text(), source)
}

// ResolveName resolves a candidate spelling in the original reference's context.
// Rules comparing paths need this without constructing or mutating AST nodes.
func (r *ImportResolver) ResolveName(name string, source *ast.Node) (string, bool, string) {
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
