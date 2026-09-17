package nodeutil

import (
	"encoding/json"
	"path"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/tailscale/hujson"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// ResolutionOptions selects runtime files independently of the compiler's
// declaration-file preference. Nil search lists use Node defaults; empty lists
// disable that search. Conditions selects active export conditions. Package
// traversal and export-path validation stay with tsgo.
type ResolutionOptions struct {
	Extensions        []string            `json:"extensions"`
	Modules           []string            `json:"modules"`
	Paths             []string            `json:"paths"`
	Conditions        []string            `json:"conditions"`
	ExtensionAliases  map[string][]string `json:"extensionAliases"`
	Aliases           []moduleAlias       `json:"aliases"`
	AliasesConfigured bool                `json:"aliasesConfigured"`
	MainFields        []nodeMainField     `json:"mainFields"`
	MainFiles         []string            `json:"mainFiles"`
	AliasFields       [][]string          `json:"aliasFields"`
	// Local imports disable directory lookup unless entry options enable it.
	// Require callers retain the ordinary Node directory lookup.
	NoDirectory bool `json:"noDirectory"`
}

type nodeResolutionKey struct{ name, file, options string }
type nodeResolution struct {
	path, resourceSuffix, resolveError string
	recursive                          bool
}

// ResolveModule resolves a runtime package through this generation's FS.
// It does not load the result into the Program or fall back to @types packages.
func ResolveModule(p *program.Program, name, containingFile string, options ResolutionOptions) string {
	resolved, _ := ResolveModuleWithError(p, name, containingFile, options)
	return resolved
}

// ResolveModuleWithError also preserves the resolution failure for missing
// module diagnostics. Package traversal and exports selection still use tsgo.
func ResolveModuleWithError(p *program.Program, name, containingFile string, options ResolutionOptions) (string, string) {
	result := resolveModuleCached(p, name, containingFile, options)
	return result.path, result.resolveError
}

func resolveModuleCached(p *program.Program, name, containingFile string, options ResolutionOptions) nodeResolution {
	if p.FS() == nil {
		return nodeResolution{}
	}
	encoded, err := json.Marshal(options)
	if err != nil {
		return nodeResolution{}
	}
	result := program.Cached(p, nodeResolutionKey{name, containingFile, string(encoded)}, func() nodeResolution {
		if options.Extensions == nil {
			options.Extensions = []string{".js", ".json", ".node", ".mjs", ".cjs"}
		}
		if options.Modules == nil {
			options.Modules = []string{"node_modules"}
		}
		if options.Conditions == nil {
			options.Conditions = []string{"node", "require", "import"}
		}
		resolver := nodeResolver{program: p, fileName: containingFile, options: options}
		return resolver.resolve(name)
	})
	return result
}

// resolveRequest keeps package traversal and file probes in tsgo.
func (resolver *nodeResolver) resolveRequest(name string) nodeResolution {
	p, containingFile, options := resolver.program, resolver.fileName, resolver.options
	originalName := name
	// enhanced-resolve separates resource queries/fragments from the path.
	if index := strings.IndexAny(name, "?#"); index > 0 && !resolver.mainTarget {
		name = name[:index]
	}
	folders := options.Modules
	if len(folders) == 0 {
		// Package imports and self-references do not use module directories.
		// Run tsgo once with external directory probes disabled.
		folders = []string{"node_modules"}
	}
	if tspath.IsExternalModuleNameRelative(name) {
		// Module search directories govern bare packages, not local files.
		folders = []string{"node_modules"}
	}
	var resolveError string
	var recursive bool
bases:
	for _, base := range append(slices.Clone(options.Paths), tspath.GetDirectoryPath(containingFile)) {
		base = tspath.ResolvePath(p.CurrentDirectory(), base)
		resolveError = "Can't resolve '" + originalName + "' in '" + base + "'"
		recursive = false
		if !resolver.mainTarget {
			if result, matched := resolver.aliasField(name, base, false); matched {
				if result.resolveError == "" && result.path != "" {
					return result
				}
				resolveError = result.resolveError
				recursive = result.recursive
				continue
			}
		}
		// A POSIX backslash is a filename character. Do not let tsgo's
		// separator normalization select a different file or package.
		if strings.Contains(name, `\`) && tspath.GetRootLength(base) == 1 && !tspath.IsRootedDiskPath(name) {
			continue
		}
		if isNodeBuiltin(name) {
			resolveError = ""
			continue
		}
		for _, folder := range folders {
			view := &nodeResolutionFS{
				FS: p.FS(), folder: folder, base: base, options: options, resolver: resolver,
				explicitExtension: path.Ext(name), resolved: map[string]string{},
				request:        name,
				noModuleSearch: len(options.Modules) == 0 && !tspath.IsExternalModuleNameRelative(name),
			}
			if options.NoDirectory {
				view.blockedDirectory = view.physical(tspath.ResolvePath(base, name))
			}
			resolver := view.newResolver(p.CurrentDirectory())
			result, _ := resolver.ResolveModuleName(name, tspath.ResolvePath(base, "__import__.js"), core.ResolutionModeCommonJS, nil)
			if view.recursiveError != "" {
				resolveError, recursive = view.recursiveError, true
				break
			}
			if !view.unresolved && result != nil && result.IsResolved() {
				if view.builtin {
					resolveError = ""
					continue bases
				}
				return nodeResolution{path: view.Realpath(result.ResolvedFileName), resourceSuffix: view.resourceSuffix}
			}
			if view.exportsFile != "" {
				request := name
				if view.importsTarget != "" {
					request = view.importsTarget
				}
				packageName, _ := module.ParsePackageName(request)
				subpath := "." + strings.TrimPrefix(request, packageName)
				conditions, err := json.Marshal(options.Conditions)
				if err != nil {
					return nodeResolution{resolveError: err.Error()}
				}
				resolveError = `"` + subpath + `" is not exported under the conditions ` + string(conditions) + " from package " + tspath.GetDirectoryPath(view.exportsFile) + " (see exports field in " + view.exportsFile + ")"
				if view.unresolved {
					resolveError = "Package path " + subpath + " is exported from package " + tspath.GetDirectoryPath(view.exportsFile) + ", but no valid target file was found (see exports field in " + view.exportsFile + ")"
				}
				// A found package's exports failure cannot fall through to a
				// different copy in a later module directory.
				break
			} else if view.importsFile != "" && !view.importsMatched {
				resolveError = "Package import " + name + " is not imported from package " + tspath.GetDirectoryPath(view.importsFile) + " (see imports field in " + view.importsFile + ")"
			}
			if view.importsFile != "" {
				if name == "#" {
					resolveError = "Request should have at least 2 characters"
				} else if strings.HasSuffix(name, "/") {
					resolveError = "Resolving to directories is not possible with the imports field (request was " + name + ")"
				}
			}
		}
	}
	return nodeResolution{resolveError: resolveError, recursive: recursive}
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
const nodeBuiltinTarget = "__rslint_node_builtin__.ts"

type nodeResolutionFS struct {
	vfs.FS
	folder            string
	base              string
	options           ResolutionOptions
	resolver          *nodeResolver
	resourceSuffix    string
	recursiveError    string
	explicitExtension string
	resolved          map[string]string
	activeDirectories map[string]bool
	unresolved        bool
	request           string
	noModuleSearch    bool
	blockedDirectory  string
	exportsFile       string
	importsFile       string
	importsMatched    bool
	importsTarget     string
	builtin           bool
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
			name = tspath.ResolvePath(f.folder, rest)
		} else {
			name = tspath.ResolvePath(name[:index+1], f.folder, rest)
		}
	}
	// Node encodes filesystem paths as UTF-8, replacing each unpaired UTF-16
	// surrogate. Keep the original JavaScript value for matching and messages.
	if !utf8.ValidString(name) {
		name = string(utf16.Decode(ecmascript.StringCodeUnits(name)))
	}
	return name
}

func (f *nodeResolutionFS) probe(name string, applyAlias bool) string {
	if result, matched := f.aliasFile(name); matched {
		return result
	}

	if aliases, ok := f.options.ExtensionAliases[path.Ext(name)]; ok && applyAlias && !f.resolver.mainTarget {
		for _, extension := range aliases {
			candidate := strings.TrimSuffix(name, path.Ext(name)) + extension
			if found := f.probeFile(candidate); found != "" {
				return found
			}
		}
		return ""
	}
	if f.FS.FileExists(name) {
		return name
	}
	for _, extension := range f.options.Extensions {
		if found := f.probeFile(name + extension); found != "" {
			return found
		}
	}
	return ""
}

func (f *nodeResolutionFS) aliasFile(name string) (string, bool) {
	result, matched := f.resolver.alias(name)
	if !matched {
		result, matched = f.resolver.aliasField(name, tspath.GetDirectoryPath(name), true)
	}
	if !matched {
		return "", false
	}
	return f.redirectedFile(name, result), true
}

func (f *nodeResolutionFS) redirectedFile(name string, result nodeResolution) string {
	if result.recursive {
		f.recursiveError = result.resolveError
	}
	if result.resolveError != "" {
		return ""
	}
	if result.path == "" {
		f.builtin = true
		return name
	}
	f.resourceSuffix = result.resourceSuffix
	return result.path
}

func (f *nodeResolutionFS) probeFile(name string) string {
	if result, matched := f.aliasFile(name); matched {
		return result
	}
	if f.FS.FileExists(name) {
		return name
	}
	return ""
}

func (f *nodeResolutionFS) FileExists(name string) bool {
	physical := f.physical(name)
	var resolved string
	switch {
	case strings.HasPrefix(f.request, "#") && tspath.GetBaseFileName(physical) == nodeBuiltinTarget:
		f.builtin = true
		return true
	case strings.HasSuffix(physical, "/package.json"):
		return f.FS.FileExists(physical)
	case strings.HasSuffix(physical, nodeDirectoryExportSuffix):
		// Upstream rejects directory targets with a trailing separator.
		f.unresolved = true
		return true
	case strings.HasSuffix(physical, "/"+nodeMainTarget+nodeTargetSuffix):
		result := f.resolver.mainEntry(tspath.GetDirectoryPath(physical))
		if result.recursive {
			f.recursiveError = result.resolveError
			return true
		}
		resolved = f.redirectedFile(physical, result)
	case strings.HasSuffix(physical, nodeTargetSuffix), strings.HasSuffix(physical, nodeExportSuffix):
		target := strings.TrimSuffix(strings.TrimSuffix(physical, nodeTargetSuffix), nodeExportSuffix)
		// Relative imports maps honor extension aliases; package main and
		// exports targets retain their explicitly selected extensions.
		applyAlias := strings.HasPrefix(f.request, "#") && f.importsTarget == ""
		resolved = f.probe(target, applyAlias)
		if resolved == "" && !f.options.NoDirectory && !strings.HasSuffix(target, "/") && f.FS.DirectoryExists(target) && !f.activeDirectories[target] {
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
	case f.options.MainFiles != nil && f.isDirectoryIndex(name, physical):
		for _, entry := range f.options.MainFiles {
			if strings.Contains(entry, `\`) && tspath.GetRootLength(physical) == 1 && !tspath.IsRootedDiskPath(entry) {
				continue
			}
			if resolved = f.probe(tspath.ResolvePath(tspath.GetDirectoryPath(physical), entry), false); resolved != "" {
				break
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
	if f.noModuleSearch && tspath.GetBaseFileName(name) == "node_modules" {
		return false
	}
	physical := strings.TrimSuffix(strings.TrimSuffix(f.physical(name), nodeTargetSuffix), nodeExportSuffix)
	return (f.blockedDirectory == "" || strings.TrimRight(physical, "/") != strings.TrimRight(f.blockedDirectory, "/")) && f.FS.DirectoryExists(physical)
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
		if imports := value.Find("/imports"); imports != nil {
			filterNodeConditions(imports, f.options.Conditions)
			markNodeImportTargets(imports, f.options.Conditions)
			if entries, ok := imports.Value.(*hujson.Object); ok && f.importsFile == "" && strings.HasPrefix(f.request, "#") {
				f.recordImports(entries, f.physical(name))
			}
		}
		request := f.request
		if f.importsTarget != "" {
			request = f.importsTarget
		}
		packageName, _ := module.ParsePackageName(request)
		selfReference := false
		if name := value.Find("/name"); name != nil {
			if literal, ok := name.Value.(hujson.Literal); ok && literal.Kind() == '"' {
				selfReference = literal.String() == packageName
			}
		}
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
					if selfReference || strings.HasSuffix(name, "/"+packageName+"/package.json") {
						f.exportsFile = f.physical(name)
					}
					filterNodeConditions(&member.Value, f.options.Conditions)
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
		if f.options.MainFields != nil {
			object.Members = slices.DeleteFunc(object.Members, func(member hujson.ObjectMember) bool {
				return nodePackageMemberName(member) == "main"
			})
			if len(f.options.MainFields) != 0 {
				object.Members = append(object.Members, hujson.ObjectMember{
					Name:  hujson.Value{Value: hujson.String("main")},
					Value: hujson.Value{Value: hujson.String("./" + nodeMainTarget + nodeTargetSuffix)},
				})
			}
		}
	}
	return string(value.Pack()), true
}

// Remember the selected request for diagnostics; tsgo still resolves its target.
func (f *nodeResolutionFS) recordImports(entries *hujson.Object, fileName string) {
	f.importsFile = fileName
	var best string
	for _, entry := range entries.Members {
		key := nodePackageMemberName(entry)
		pattern := core.TryParsePattern(key)
		if !pattern.IsValid() || !pattern.Matches(f.request) {
			continue
		}
		if best != "" && key != f.request && (best == f.request || module.ComparePatternKeys(key, best) >= 0) {
			continue
		}
		best = key
		target := entry.Value
		if conditions, ok := target.Value.(*hujson.Object); ok {
			target, _ = nodeExportArrayCondition(conditions, f.options.Conditions)
		}
		f.importsMatched, f.importsTarget = false, ""
		if literal, ok := target.Value.(hujson.Literal); ok {
			if literal.Kind() == '"' && literal.String() != "" {
				f.importsMatched = true
				name := literal.String()
				if pattern.StarIndex >= 0 {
					name = strings.ReplaceAll(name, "*", pattern.MatchedText(f.request))
				}
				if !tspath.IsExternalModuleNameRelative(name) {
					f.importsTarget = name
				}
			}
		} else {
			f.importsMatched = target.Value != nil
		}
	}
}

// Let tsgo select imports-map keys and conditions, including redirects to Node
// builtins that have no file for a compiler resolver to find.
func markNodeImportTargets(value *hujson.Value, conditions []string) {
	switch v := value.Value.(type) {
	case hujson.Literal:
		if v.Kind() == '"' {
			target := v.String()
			if index := strings.IndexAny(target, "?#"); index > 0 {
				target = target[:index]
				value.Value = hujson.String(target)
			}
			if isNodeBuiltin(target) {
				value.Value = hujson.String("./" + nodeBuiltinTarget)
			} else if strings.HasPrefix(target, "./") {
				markNodeTargets(value, nodeExportSuffix)
			}
		}
	case *hujson.Object:
		for i := range v.Members {
			markNodeImportTargets(&v.Members[i].Value, conditions)
		}
	case *hujson.Array:
		if !prepareNodeExports(value, conditions) {
			value.Value = &hujson.Array{}
			return
		}
		for i := range v.Elements {
			markNodeImportTargets(&v.Elements[i], conditions)
			if target, ok := v.Elements[i].Value.(hujson.Literal); ok && target.Kind() == '"' && !tspath.IsExternalModuleNameRelative(target.String()) {
				// Package targets stop on resolution failure just like the
				// marked file targets; tsgo must not try a later array entry.
				v.Elements = v.Elements[:i+1]
				if i == 0 {
					value.Value = target
				}
				break
			}
		}
	}
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
		if key == "default" || slices.Contains(conditions, key) {
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
			// Preserve invalid targets for tsgo's validation; appending a
			// slash would turn "." into an apparently valid "./..." target.
			if suffix == nodeExportSuffix && !strings.HasPrefix(v.String(), "./") {
				return
			}
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

// tsgo activates the require condition for CommonJS internally. Remove inactive
// conditions before handing it exports/imports, so an explicit override can
// replace that default as well as add custom conditions. Keep subpath keys.
func filterNodeConditions(value *hujson.Value, conditions []string) {
	switch object := value.Value.(type) {
	case *hujson.Object:
		hasPathKeys := slices.ContainsFunc(object.Members, func(member hujson.ObjectMember) bool {
			key := nodePackageMemberName(member)
			return strings.HasPrefix(key, ".") || strings.HasPrefix(key, "#")
		})
		if !hasPathKeys {
			object.Members = slices.DeleteFunc(object.Members, func(member hujson.ObjectMember) bool {
				key := nodePackageMemberName(member)
				return !strings.HasPrefix(key, ".") && !strings.HasPrefix(key, "#") && key != "default" && !slices.Contains(conditions, key)
			})
		}
		for i := range object.Members {
			filterNodeConditions(&object.Members[i].Value, conditions)
		}
	case *hujson.Array:
		for i := range object.Elements {
			filterNodeConditions(&object.Elements[i], conditions)
		}
	}
}
