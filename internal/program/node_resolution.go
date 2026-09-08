package program

import (
	"encoding/json"
	"path"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/tailscale/hujson"
)

// NodeResolutionOptions selects runtime files independently of the compiler's
// declaration-file preference. Nil extension/module lists use Node defaults;
// empty lists disable that search. Conditions adds active export conditions to
// CommonJS's require condition. Package traversal and exports stay with tsgo.
type NodeResolutionOptions struct {
	Extensions       []string
	Modules          []string
	Paths            []string
	Conditions       []string
	ExtensionAliases map[string][]string
}

type nodeResolutionKey struct{ name, file, options string }

// ResolveNodeModule resolves a runtime package through this generation's FS.
// It does not load the result into the Program or fall back to @types packages.
func (p *Program) ResolveNodeModule(name, containingFile string, options NodeResolutionOptions) string {
	if p.FS() == nil {
		return ""
	}
	encoded, err := json.Marshal([]any{options.Extensions, options.Modules, options.Paths, options.Conditions, options.ExtensionAliases})
	if err != nil {
		return ""
	}
	return Cached(p, nodeResolutionKey{name, containingFile, string(encoded)}, func() string {
		if options.Extensions == nil {
			options.Extensions = []string{".js", ".json", ".node", ".mjs", ".cjs"}
		}
		if options.Modules == nil {
			options.Modules = []string{"node_modules"}
		}
		if options.Conditions == nil {
			options.Conditions = []string{"node", "require", "import"}
		}
		// enhanced-resolve separates resource queries/fragments from the path.
		if index := strings.IndexAny(name, "?#"); index > 0 {
			name = name[:index]
		}
		for _, base := range append(slices.Clone(options.Paths), tspath.GetDirectoryPath(containingFile)) {
			for _, folder := range options.Modules {
				view := &nodeResolutionFS{
					FS: p.FS(), folder: folder, options: options,
					explicitExtension: path.Ext(name), resolved: map[string]string{},
				}
				resolver := view.newResolver(p.CurrentDirectory())
				base = tspath.ResolvePath(p.CurrentDirectory(), base)
				result, _ := resolver.ResolveModuleName(name, tspath.ResolvePath(base, "__import__.js"), core.ResolutionModeCommonJS, nil)
				if !view.invalidPackage && result != nil && result.IsResolved() {
					return view.Realpath(result.ResolvedFileName)
				}
			}
		}
		return ""
	})
}

type nodeResolutionHost struct {
	fs  vfs.FS
	cwd string
}

func (h *nodeResolutionHost) FS() vfs.FS                  { return h.fs }
func (h *nodeResolutionHost) GetCurrentDirectory() string { return h.cwd }

// The private view projects runtime candidates onto tsgo's JavaScript probes.
// A marked package target retains its exact extension (including .node or
// .d.ts); unmarked index/file probes use the caller's extension list. This
// reuses tsgo's package/exports traversal without exposing a second source host.
const nodeTargetSuffix = ".__rslint_node_target__.js"

type nodeResolutionFS struct {
	vfs.FS
	folder            string
	options           NodeResolutionOptions
	explicitExtension string
	resolved          map[string]string
	activeDirectories map[string]bool
	invalidPackage    bool
}

func (f *nodeResolutionFS) newResolver(cwd string) *module.Resolver {
	return module.NewResolver(&nodeResolutionHost{f, cwd}, &core.CompilerOptions{
		ModuleResolution: core.ModuleResolutionKindBundler,
		NoDtsResolution:  core.TSTrue, ResolveJsonModule: core.TSTrue,
		CustomConditions: f.options.Conditions,
	}, "", "", append(slices.Clone(f.options.Extensions), f.explicitExtension))
}

func (f *nodeResolutionFS) physical(name string) string {
	name = strings.ReplaceAll(name, nodeTargetSuffix+"/", "/")
	if index := strings.LastIndex(name, "/node_modules"); index >= 0 && (len(name) == index+13 || name[index+13] == '/') {
		rest := strings.TrimPrefix(name[index+13:], "/")
		if tspath.IsRootedDiskPath(f.folder) {
			return tspath.ResolvePath(f.folder, rest)
		}
		return tspath.ResolvePath(name[:index+1], f.folder, rest)
	}
	return name
}

func (f *nodeResolutionFS) probe(name string, applyAlias bool) string {
	if aliases, ok := f.options.ExtensionAliases[path.Ext(name)]; ok && applyAlias {
		for _, extension := range aliases {
			candidate := strings.TrimSuffix(name, path.Ext(name)) + extension
			if f.FS.FileExists(candidate) {
				return candidate
			}
		}
		return ""
	}
	if f.FS.FileExists(name) {
		return name
	}
	for _, extension := range f.options.Extensions {
		if candidate := name + extension; f.FS.FileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func (f *nodeResolutionFS) FileExists(name string) bool {
	physical := f.physical(name)
	var resolved string
	switch {
	case strings.HasSuffix(physical, "/package.json"):
		return f.FS.FileExists(physical)
	case strings.HasSuffix(physical, nodeTargetSuffix):
		target := strings.TrimSuffix(physical, nodeTargetSuffix)
		resolved = f.probe(target, false)
		if resolved == "" && !strings.HasSuffix(target, "/") && f.FS.DirectoryExists(target) && !f.activeDirectories[target] {
			// enhanced-resolve permits directory exports. Ask tsgo to resolve
			// their main/index using its regular relative-directory traversal.
			// Guard package main cycles before creating a nested resolver.
			if f.activeDirectories == nil {
				f.activeDirectories = map[string]bool{}
			}
			f.activeDirectories[target] = true
			directory := tspath.GetDirectoryPath(target)
			result, _ := f.newResolver(directory).ResolveModuleName("./"+tspath.GetBaseFileName(target), tspath.ResolvePath(directory, "__import__.js"), core.ResolutionModeCommonJS, nil)
			delete(f.activeDirectories, target)
			if result != nil && result.IsResolved() {
				resolved = f.Realpath(result.ResolvedFileName)
			}
		}
	case f.explicitExtension != "" && strings.HasSuffix(physical, f.explicitExtension):
		resolved = f.probe(physical, true)
	case strings.HasSuffix(physical, ".js"):
		resolved = f.probe(strings.TrimSuffix(physical, ".js"), false)
	}
	if resolved != "" {
		f.resolved[name] = resolved
	}
	return resolved != ""
}

func (f *nodeResolutionFS) DirectoryExists(name string) bool {
	return f.FS.DirectoryExists(strings.TrimSuffix(f.physical(name), nodeTargetSuffix))
}

func (f *nodeResolutionFS) Realpath(name string) string {
	if resolved := f.resolved[name]; resolved != "" {
		return f.FS.Realpath(resolved)
	}
	return f.FS.Realpath(f.physical(name))
}

func (f *nodeResolutionFS) ReadFile(name string) (string, bool) {
	text, ok := f.FS.ReadFile(f.physical(name))
	if !ok || !strings.HasSuffix(name, "/package.json") {
		return text, ok
	}
	if !json.Valid([]byte(text)) {
		f.invalidPackage = true
		return "", false
	}
	value, err := hujson.Parse([]byte(text))
	if err != nil {
		return text, true
	}
	if object, ok := value.Value.(*hujson.Object); ok {
		// Runtime lookup must not select TypeScript's versioned declarations.
		object.Members = slices.DeleteFunc(object.Members, func(member hujson.ObjectMember) bool {
			return nodePackageMemberName(member) == "typesVersions"
		})
		for i := range object.Members {
			member := &object.Members[i]
			key := nodePackageMemberName(*member)
			if key == "main" || key == "exports" {
				if key == "exports" {
					if literal, ok := member.Value.Value.(hujson.Literal); ok && (string(literal) == "null" || string(literal) == "false" || string(literal) == `""`) {
						continue
					}
					selected := selectNodeExports(&member.Value, f.options.Conditions)
					literal, _ := member.Value.Value.(hujson.Literal)
					if !selected || string(literal) == "null" {
						// A blocked selection must retain an exports map: a bare
						// null field permits tsgo's legacy main/index fallback.
						member.Value.Value = &hujson.Object{Members: []hujson.ObjectMember{{
							Name: hujson.Value{Value: hujson.String(".")}, Value: hujson.Value{Value: hujson.Literal("null")},
						}}}
					}
				}
				markNodeTargets(&member.Value)
			}
		}
	}
	return string(value.Pack()), true
}

// Exports selects a target before checking whether the file exists. tsgo also
// tries subsequent conditions/array entries when declarations are missing;
// runtime lookup must not turn such a broken first target into a valid import.
func selectNodeExports(value *hujson.Value, conditions []string) bool {
	switch object := value.Value.(type) {
	case *hujson.Object:
		if len(object.Members) > 0 && strings.HasPrefix(nodePackageMemberName(object.Members[0]), ".") {
			for i := range object.Members {
				if !selectNodeExports(&object.Members[i].Value, conditions) {
					object.Members[i].Value.Value = hujson.Literal("null")
				}
			}
			return true
		}
		for i := range object.Members {
			key := nodePackageMemberName(object.Members[i])
			if key == "default" || key == "require" || slices.Contains(conditions, key) {
				candidate := object.Members[i].Value
				if selectNodeExports(&candidate, conditions) {
					value.Value = candidate.Value
					return true
				}
			}
		}
		return false
	case *hujson.Array:
		var selected *hujson.Value
		for _, candidate := range object.Elements {
			if literal, ok := candidate.Value.(hujson.Literal); ok && literal.Kind() != '"' {
				// Upstream rejects primitive array entries, even after a valid
				// target. A null condition value outside an array simply blocks it.
				value.Value = hujson.Literal("null")
				return true
			}
			if _, nested := candidate.Value.(*hujson.Array); nested {
				continue
			}
			if selectNodeExports(&candidate, conditions) {
				if literal, ok := candidate.Value.(hujson.Literal); ok && literal.Kind() == '"' && selected == nil && validNodeExportTarget(literal.String()) {
					selected = &candidate
				}
			}
		}
		if selected != nil {
			value.Value = selected.Value
			return true
		}
		return false
	case hujson.Literal:
		if object.Kind() == '"' {
			target := object.String()
			if index := strings.IndexAny(target, "?#"); index > 0 {
				value.Value = hujson.String(target[:index])
			}
		}
	}
	return true
}

func validNodeExportTarget(target string) bool {
	if !strings.HasPrefix(target, "./") {
		return false
	}
	// Use the same path component rules as tsgo before selecting an array
	// target. Invalid paths may fall through; missing files may not.
	for _, part := range tspath.GetPathComponents(target, "")[2:] {
		if part == "." || part == ".." || part == "node_modules" {
			return false
		}
	}
	return true
}

func nodePackageMemberName(member hujson.ObjectMember) string {
	name, ok := member.Name.Value.(hujson.Literal)
	if !ok {
		return ""
	}
	return name.String()
}

func markNodeTargets(value *hujson.Value) {
	switch v := value.Value.(type) {
	case hujson.Literal:
		if v.Kind() == '"' {
			value.Value = hujson.String(v.String() + nodeTargetSuffix)
		}
	case *hujson.Object:
		for i := range v.Members {
			markNodeTargets(&v.Members[i].Value)
		}
	case *hujson.Array:
		for i := range v.Elements {
			markNodeTargets(&v.Elements[i])
		}
	}
}

type nearestConfigKey string
type compilerOptionsKey string

// ReadCompilerOptions parses an explicit config through this generation's FS.
// Returned options are cached and must be treated as immutable.
func (p *Program) ReadCompilerOptions(fileName string) *core.CompilerOptions {
	if p.FS() == nil {
		return nil
	}
	fileName = tspath.ResolvePath(p.CurrentDirectory(), fileName)
	return Cached(p, compilerOptionsKey(fileName), func() *core.CompilerOptions {
		host := compiler.NewCompilerHost(p.CurrentDirectory(), p.FS(), p.DefaultLibraryPath(), nil, nil, nil)
		parsed, _ := tsoptions.GetParsedCommandLineOfConfigFile(fileName, &core.CompilerOptions{}, nil, host, nil)
		if parsed != nil {
			return parsed.CompilerOptions()
		}
		return nil
	})
}

// NearestCompilerOptions reads the nearest tsconfig using the same immutable
// FS and tsgo config parser, including extends. It does not create a Program.
func (p *Program) NearestCompilerOptions(fileName string) *core.CompilerOptions {
	if p.FS() == nil {
		return nil
	}
	return Cached(p, nearestConfigKey(tspath.GetDirectoryPath(fileName)), func() *core.CompilerOptions {
		for directory := tspath.GetDirectoryPath(fileName); directory != ""; {
			config := tspath.ResolvePath(directory, "tsconfig.json")
			if p.FS().FileExists(config) {
				return p.ReadCompilerOptions(config)
			}
			parent := tspath.GetDirectoryPath(directory)
			if parent == directory {
				break
			}
			directory = parent
		}
		return nil
	})
}
