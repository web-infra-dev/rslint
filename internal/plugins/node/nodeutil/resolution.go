package nodeutil

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
	"github.com/web-infra-dev/rslint/internal/program"
)

// ResolutionOptions selects runtime files independently of the compiler's
// declaration-file preference. Nil extension/module lists use Node defaults;
// empty lists disable that search. Conditions adds active export conditions to
// CommonJS's require condition. Package traversal and export-path validation
// stay with tsgo.
type ResolutionOptions struct {
	Extensions       []string
	Modules          []string
	Paths            []string
	Conditions       []string
	ExtensionAliases map[string][]string
}

type nodeResolutionKey struct{ name, file, options string }

// ResolveModule resolves a runtime package through this generation's FS.
// It does not load the result into the Program or fall back to @types packages.
func ResolveModule(p *program.Program, name, containingFile string, options ResolutionOptions) string {
	if p.FS() == nil {
		return ""
	}
	encoded, err := json.Marshal([]any{options.Extensions, options.Modules, options.Paths, options.Conditions, options.ExtensionAliases})
	if err != nil {
		return ""
	}
	return program.Cached(p, nodeResolutionKey{name, containingFile, string(encoded)}, func() string {
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
				if !view.unresolved && result != nil && result.IsResolved() {
					return view.Realpath(result.ResolvedFileName)
				}
			}
		}
		return ""
	})
}

// The private view projects runtime candidates onto tsgo's file probes.
// A marked package target retains its exact extension (including .node or
// .d.ts); unmarked index/file probes use the caller's extension list. This
// reuses tsgo's package/exports traversal without exposing a second source host.
const nodeTargetSuffix = ".__rslint_node_target__.js"

// An extra component preserves every original export-path component for tsgo's
// validation. A TS probe checks the file directly, without requiring the
// original target to exist as a directory first.
const nodeExportSuffix = "/.__rslint_node_export__.ts"
const nodeDirectoryExportSuffix = "/.__rslint_node_directory_export__.ts"

type nodeResolutionFS struct {
	vfs.FS
	folder            string
	options           ResolutionOptions
	explicitExtension string
	resolved          map[string]string
	activeDirectories map[string]bool
	unresolved        bool
}

func (f *nodeResolutionFS) newResolver(cwd string) *module.Resolver {
	host := compiler.NewCompilerHost(cwd, f, "", nil, nil, nil)
	return module.NewResolver(host, &core.CompilerOptions{
		ModuleResolution: core.ModuleResolutionKindBundler,
		NoDtsResolution:  core.TSTrue, ResolveJsonModule: core.TSTrue,
		CustomConditions: f.options.Conditions,
		// Keep probes in the projected namespace. The caller maps the selected
		// file via Realpath; tsgo must not remap an already physical path.
		PreserveSymlinks: core.TSTrue,
	}, "", "", append(slices.Clone(f.options.Extensions), f.explicitExtension))
}

func (f *nodeResolutionFS) physical(name string) string {
	name = strings.ReplaceAll(name, nodeTargetSuffix+"/", "/")
	name = strings.ReplaceAll(name, nodeExportSuffix+"/", "/")
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
	case strings.HasSuffix(physical, nodeDirectoryExportSuffix):
		// Upstream rejects directory targets with a trailing separator.
		f.unresolved = true
		return true
	case strings.HasSuffix(physical, nodeTargetSuffix), strings.HasSuffix(physical, nodeExportSuffix):
		target := strings.TrimSuffix(strings.TrimSuffix(physical, nodeTargetSuffix), nodeExportSuffix)
		resolved = f.probe(target, false)
		if resolved == "" && !strings.HasSuffix(target, "/") && f.FS.DirectoryExists(target) && !f.activeDirectories[target] {
			// enhanced-resolve permits directory exports. Ask tsgo to resolve
			// their main/index using its regular relative-directory traversal.
			// Guard package main cycles before creating a nested resolver.
			if f.activeDirectories == nil {
				f.activeDirectories = map[string]bool{}
			}
			f.activeDirectories[target] = true
			logicalTarget := strings.TrimSuffix(strings.TrimSuffix(name, nodeTargetSuffix), nodeExportSuffix)
			directory := tspath.GetDirectoryPath(logicalTarget)
			result, _ := f.newResolver(directory).ResolveModuleName("./"+tspath.GetBaseFileName(logicalTarget), tspath.ResolvePath(directory, "__import__.js"), core.ResolutionModeCommonJS, nil)
			delete(f.activeDirectories, target)
			if result != nil && result.IsResolved() {
				resolved = f.Realpath(result.ResolvedFileName)
			}
		}
		if resolved == "" && strings.HasSuffix(physical, nodeExportSuffix) {
			// Stop tsgo after its first valid exports target, even when that
			// target is missing. The caller discards this synthetic result.
			// This leaves target validation and condition selection in tsgo.
			f.unresolved = true
			return true
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
	return f.FS.DirectoryExists(strings.TrimSuffix(strings.TrimSuffix(f.physical(name), nodeTargetSuffix), nodeExportSuffix))
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
		f.unresolved = true
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
					if !prepareNodeExports(&member.Value, f.options.Conditions) {
						// A blocked selection must retain an exports map: a bare
						// null field permits tsgo's legacy main/index fallback.
						member.Value.Value = &hujson.Object{Members: []hujson.ObjectMember{{
							Name: hujson.Value{Value: hujson.String(".")}, Value: hujson.Value{Value: hujson.Literal("null")},
						}}}
					}
				}
				suffix := nodeTargetSuffix
				if key == "exports" {
					suffix = nodeExportSuffix
				}
				markNodeTargets(&member.Value, suffix)
			}
		}
	}
	return string(value.Pack()), true
}

// tsgo owns exports paths, patterns and ordinary condition selection. Only
// normalize enhanced-resolve's array behavior: primitive entries invalidate an
// array; nested arrays and unmatched/null conditional entries are skipped.
func prepareNodeExports(value *hujson.Value, conditions []string) bool {
	switch object := value.Value.(type) {
	case *hujson.Object:
		for i := range object.Members {
			if !prepareNodeExports(&object.Members[i].Value, conditions) {
				object.Members[i].Value.Value = hujson.Literal("null")
			}
		}
	case *hujson.Array:
		targets := make([]hujson.Value, 0, len(object.Elements))
		for _, target := range object.Elements {
			switch candidate := target.Value.(type) {
			case hujson.Literal:
				if candidate.Kind() != '"' {
					return false
				}
			case *hujson.Array:
				continue
			case *hujson.Object:
				selected, ok := nodeExportArrayCondition(candidate, conditions)
				if !ok {
					continue
				}
				target = selected
				if literal, ok := target.Value.(hujson.Literal); ok && (literal.Kind() != '"' || literal.String() == "") {
					continue
				}
			}
			if !prepareNodeExports(&target, conditions) {
				return false
			}
			targets = append(targets, target)
		}
		object.Elements = targets
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

// A conditional null inside an exports array is skipped by enhanced-resolve,
// while tsgo treats it as a blocked resolution. Select only these array entries
// here; ordinary conditional exports remain untouched for tsgo to select.
func nodeExportArrayCondition(object *hujson.Object, conditions []string) (hujson.Value, bool) {
	for _, member := range object.Members {
		key := nodePackageMemberName(member)
		if key == "default" || key == "require" || slices.Contains(conditions, key) {
			if nested, ok := member.Value.Value.(*hujson.Object); ok {
				if selected, found := nodeExportArrayCondition(nested, conditions); found {
					return selected, true
				}
			} else {
				return member.Value, true
			}
		}
	}
	return hujson.Value{}, false
}

func nodePackageMemberName(member hujson.ObjectMember) string {
	name, ok := member.Name.Value.(hujson.Literal)
	if !ok {
		return ""
	}
	return name.String()
}

func markNodeTargets(value *hujson.Value, suffix string) {
	switch v := value.Value.(type) {
	case hujson.Literal:
		if v.Kind() == '"' {
			if suffix == nodeExportSuffix && strings.HasSuffix(v.String(), "/") {
				suffix = nodeDirectoryExportSuffix
			}
			value.Value = hujson.String(v.String() + suffix)
		}
	case *hujson.Object:
		for i := range v.Members {
			markNodeTargets(&v.Members[i].Value, suffix)
		}
	case *hujson.Array:
		for i := range v.Elements {
			markNodeTargets(&v.Elements[i], suffix)
		}
	}
}

type nearestConfigKey string
type compilerOptionsKey string

// readCompilerOptions parses an explicit config through this generation's FS.
// Returned options are cached and must be treated as immutable.
func readCompilerOptions(p *program.Program, fileName string) *core.CompilerOptions {
	if p.FS() == nil {
		return nil
	}
	fileName = tspath.ResolvePath(p.CurrentDirectory(), fileName)
	return program.Cached(p, compilerOptionsKey(fileName), func() *core.CompilerOptions {
		host := compiler.NewCompilerHost(p.CurrentDirectory(), p.FS(), p.DefaultLibraryPath(), nil, nil, nil)
		parsed, _ := tsoptions.GetParsedCommandLineOfConfigFile(fileName, &core.CompilerOptions{}, nil, host, nil)
		if parsed != nil {
			return parsed.CompilerOptions()
		}
		return nil
	})
}

// nearestCompilerOptions reads the nearest tsconfig using the same immutable
// FS and tsgo config parser, including extends. It does not create a Program.
func nearestCompilerOptions(p *program.Program, fileName string) *core.CompilerOptions {
	directory := tspath.GetDirectoryPath(fileName)
	if p.FS() == nil || directory == "" {
		return nil
	}
	return program.Cached(p, nearestConfigKey(directory), func() *core.CompilerOptions {
		config, found := tspath.ForEachAncestorDirectory(directory, func(directory string) (string, bool) {
			config := tspath.ResolvePath(directory, "tsconfig.json")
			return config, p.FS().FileExists(config)
		})
		if found {
			return readCompilerOptions(p, config)
		}
		return nil
	})
}
