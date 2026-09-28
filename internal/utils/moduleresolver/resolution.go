package moduleresolver

import (
	"encoding/json"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/tailscale/hujson"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// Options selects runtime files independently of the compiler's
// declaration-file preference. Nil search lists use Node defaults; empty lists
// disable that search. Conditions selects active export conditions. Package
// traversal and export-path validation stay with tsgo.
type Options struct {
	Extensions       []string            `json:"extensions"`
	Modules          []string            `json:"modules"`
	Paths            []string            `json:"paths"`
	Conditions       []string            `json:"conditions"`
	ExtensionAliases map[string][]string `json:"extensionAliases"`
	Aliases          []Alias             `json:"aliases"`
	Fallbacks        []Alias             `json:"fallbacks,omitempty"`
	MainFields       []MainField         `json:"mainFields"`
	MainFiles        []string            `json:"mainFiles"`
	AliasFields      [][]string          `json:"aliasFields"`
	FullySpecified   bool                `json:"fullySpecified,omitempty"`
	// Legacy Node resolvers use literal filenames and main/index entries,
	// without package exports/imports or resource query handling.
	IgnoreExports    bool `json:"ignoreExports,omitempty"`
	LiteralPaths     bool `json:"literalPaths,omitempty"`
	PreserveSymlinks bool `json:"preserveSymlinks,omitempty"`
	// Local imports disable directory lookup unless entry options enable it.
	// Require callers retain the ordinary Node directory lookup.
	NoDirectory bool `json:"noDirectory"`
	// NodeExports applies Node's own `exports`/`imports` target rules instead
	// of enhanced-resolve's: a target must name an existing file exactly, with
	// no extension or directory lookup, and is read as a URL after any pattern
	// is substituted. A fallback array skips every entry that is not a valid
	// target, including null and other non-string values.
	NodeExports bool `json:"nodeExports,omitempty"`
}

// DefaultExtensions returns the runtime lookup defaults in search order.
func DefaultExtensions() []string { return []string{".js", ".json", ".node", ".mjs", ".cjs"} }

type resolutionKey struct{ name, file, options string }

// Result separates the filesystem path from resource suffixes and failures.
// Empty Path and Error can represent a builtin or an explicitly ignored alias.
type Result struct {
	Path, ResourceSuffix, Error string
	recursive                   bool
	terminal                    bool
}

// ResolveModule resolves a runtime package through this generation's FS.
// It does not load the result into the Program or fall back to @types packages.
func ResolveModule(p *program.Program, name, containingFile string, options Options) string {
	resolved, _ := ResolveModuleWithError(p, name, containingFile, options)
	return resolved
}

// ResolveModuleWithError also preserves the resolution failure for missing
// module diagnostics. Package traversal and exports selection still use tsgo.
func ResolveModuleWithError(p *program.Program, name, containingFile string, options Options) (string, string) {
	result := Resolve(p, name, containingFile, options)
	return result.Path, result.Error
}

// Resolve caches runtime lookup within the immutable Program generation.
// It neither loads source files nor consults plugin settings.
func Resolve(p *program.Program, name, containingFile string, options Options) Result {
	if p.FS() == nil {
		return Result{}
	}
	encoded, err := json.Marshal(options)
	if err != nil {
		return Result{}
	}
	result := program.Cached(p, resolutionKey{name, containingFile, string(encoded)}, func() Result {
		if options.Extensions == nil {
			options.Extensions = DefaultExtensions()
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
func (resolver *nodeResolver) resolveRequest(name string) Result {
	p, containingFile, options := resolver.program, resolver.fileName, resolver.options
	originalName := name
	// enhanced-resolve separates resource queries/fragments from the path.
	if index := strings.IndexAny(name, "?#"); index > 0 && !resolver.mainTarget && !options.LiteralPaths {
		name = name[:index]
	}
	if options.NodeExports && !tspath.IsExternalModuleNameRelative(name) {
		// Node lets a pattern match an empty segment and reads it as a single
		// separator, as a plain path lookup does.
		for strings.Contains(name, "//") {
			name = strings.ReplaceAll(name, "//", "/")
		}
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
	var terminal bool
bases:
	for _, base := range append(slices.Clone(options.Paths), tspath.GetDirectoryPath(containingFile)) {
		base = tspath.ResolvePath(p.CurrentDirectory(), base)
		resolveError = "Can't resolve '" + originalName + "' in '" + base + "'"
		recursive = false
		terminal = false
		if !resolver.mainTarget {
			if result, matched := resolver.aliasField(name, base, false); matched {
				if result.Error == "" && result.Path != "" {
					return result
				}
				resolveError = result.Error
				recursive = result.recursive
				terminal = result.terminal
				continue
			}
		}
		// A POSIX backslash is a filename character. Do not let tsgo's
		// separator normalization select a different file or package. Node
		// reads a backslash after the package name as `/` only once an exports
		// or imports target turns the request into a URL.
		literalBackslash := strings.Contains(name, `\`) && tspath.GetRootLength(base) == 1 && !tspath.IsRootedDiskPath(name)
		if literalBackslash && (!options.NodeExports || !nodeURLBackslash(name)) {
			continue
		}
		// Explicit fallbacks replace the default builtin exemption and must
		// run after ordinary package lookup, like any other fallback.
		if modules.IsNodeBuiltin(name) && len(options.Fallbacks) == 0 {
			resolveError = ""
			continue
		}
		for _, folder := range folders {
			view := &nodeResolutionFS{
				FS: p.FS(), folder: folder, base: base, options: options, resolver: resolver,
				explicitExtension: path.Ext(name), resolved: map[string]string{},
				request:        name,
				noModuleSearch: len(options.Modules) == 0 && !tspath.IsExternalModuleNameRelative(name),
				fullySpecified: options.FullySpecified,
			}
			// Bare package entry points retain main/index lookup. Nested paths and
			// relative requests must name a file when fullySpecified is enabled.
			if view.fullySpecified && !tspath.IsExternalModuleNameRelative(name) && !strings.HasPrefix(name, "#") {
				if packageName, rest := module.ParsePackageName(name); packageName != "" && rest == "" {
					// Self-references go directly through exports, without the
					// external package-entry lookup that relaxes fullySpecified.
					owner := packagejson.FindNearest(p, tspath.ResolvePath(base, "__import__.js"))
					view.fullySpecified = false
					if owner != nil && owner.Field("name") == packageName {
						switch owner.Field("exports") {
						case nil, false, "", float64(0):
						default:
							view.fullySpecified = true
						}
					}
				}
			}
			if options.NoDirectory {
				view.blockedDirectory = view.physical(tspath.ResolvePath(base, name))
			}
			resolver := view.newResolver(p.CurrentDirectory())
			result, _ := resolver.ResolveModuleName(name, tspath.ResolvePath(base, "__import__.js"), core.ResolutionModeCommonJS, nil)
			if view.failure.Error != "" {
				resolveError, recursive, terminal = view.failure.Error, view.failure.recursive, view.failure.terminal
				break
			}
			if !view.unresolved && result != nil && result.IsResolved() {
				if literalBackslash && !view.nodeTarget {
					continue bases
				}
				if view.builtin {
					resolveError = ""
					continue bases
				}
				return Result{Path: view.Realpath(result.ResolvedFileName), ResourceSuffix: view.resourceSuffix}
			}
			if view.exportsFile != "" {
				terminal = true
				request := name
				if view.importsTarget != "" {
					request = view.importsTarget
				}
				packageName, _ := module.ParsePackageName(request)
				subpath := "." + strings.TrimPrefix(request, packageName)
				conditions, err := json.Marshal(options.Conditions)
				if err != nil {
					return Result{Error: err.Error()}
				}
				resolveError = `"` + subpath + `" is not exported under the conditions ` + string(conditions) + " from package " + tspath.GetDirectoryPath(view.exportsFile) + " (see exports field in " + view.exportsFile + ")"
				if view.unresolved {
					resolveError = "Package path " + subpath + " is exported from package " + tspath.GetDirectoryPath(view.exportsFile) + ", but no valid target file was found (see exports field in " + view.exportsFile + ")"
				}
				// A found package's exports failure cannot fall through to a
				// different copy in a later module directory.
				break
			} else if view.importsFile != "" && !view.importsMatched {
				terminal = true
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
	return Result{Error: resolveError, recursive: recursive, terminal: terminal}
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
	options           Options
	resolver          *nodeResolver
	resourceSuffix    string
	failure           Result
	explicitExtension string
	resolved          map[string]string
	activeDirectories map[string]bool
	unresolved        bool
	request           string
	noModuleSearch    bool
	blockedDirectory  string
	fullySpecified    bool
	exportsFile       string
	importsFile       string
	importsMatched    bool
	importsTarget     string
	builtin           bool
	// nodeTarget records that an exports or imports target was selected, and
	// exportsKeys and importsKeys the keys that match a request without a
	// pattern. NodeExports only.
	nodeTarget  bool
	exportsKeys map[string]bool
	importsKeys map[string]bool
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
	if f.fullySpecified {
		return ""
	}
	for _, extension := range f.options.Extensions {
		if found := f.probeFile(name + extension); found != "" {
			return found
		}
	}
	return ""
}

func (f *nodeResolutionFS) aliasFile(name string) (string, bool) {
	result, matched := f.resolver.alias(name, f.options.Aliases)
	if !matched {
		result, matched = f.resolver.aliasField(name, tspath.GetDirectoryPath(name), true)
	}
	if !matched {
		return "", false
	}
	return f.redirectedFile(name, result), true
}

func (f *nodeResolutionFS) redirectedFile(name string, result Result) string {
	if result.recursive || result.terminal {
		f.failure = result
	}
	if result.Error != "" {
		return ""
	}
	if result.Path == "" {
		f.builtin = true
		return name
	}
	f.resourceSuffix = result.ResourceSuffix
	return result.Path
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
		if f.fullySpecified {
			return false
		}
		result := f.resolver.mainEntry(tspath.GetDirectoryPath(physical))
		if result.recursive || result.terminal {
			f.failure = result
			return true
		}
		resolved = f.redirectedFile(physical, result)
	case strings.HasSuffix(physical, nodeTargetSuffix), strings.HasSuffix(physical, nodeExportSuffix):
		if f.fullySpecified && strings.HasSuffix(physical, nodeTargetSuffix) {
			return false
		}
		target := strings.TrimSuffix(strings.TrimSuffix(physical, nodeTargetSuffix), nodeExportSuffix)
		// Relative imports maps honor extension aliases; package main and
		// exports targets retain their explicitly selected extensions.
		applyAlias := strings.HasPrefix(f.request, "#") && f.importsTarget == ""
		// Node accepts an exports or imports target only as an existing file.
		exactTarget := f.options.NodeExports && strings.HasSuffix(physical, nodeExportSuffix)
		if exactTarget {
			if path, ok := f.nodeTargetPath(target); ok {
				resolved = f.probeFile(path)
			}
			f.nodeTarget = f.nodeTarget || resolved != ""
		} else {
			resolved = f.probe(target, applyAlias)
		}
		if resolved == "" && !exactTarget && !f.options.NoDirectory && !f.fullySpecified && !strings.HasSuffix(target, "/") && f.FS.DirectoryExists(target) && !f.activeDirectories[target] {
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
	case (f.options.MainFiles != nil || f.fullySpecified) && f.isDirectoryIndex(name, physical):
		if f.fullySpecified {
			return false
		}
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

// encodedSeparator is Node's own check for an escaped path separator. It only
// folds ASCII letters, which RE2 and JavaScript fold alike.
var encodedSeparator = regexp.MustCompile(`(?i)%2F|%5C`)

// invalidSegment is Node's check for a `.`, `..` or `node_modules` segment,
// spelled with or without percent-escapes, between either separator. Node
// rejects an exports or imports target, or a request's pattern match, that
// contains one.
var invalidSegment = esregexp.MustCompile(`(^|\\|\/)((\.|%2e)(\.|%2e)?|(n|%6e|%4e)(o|%6f|%4f)(d|%64|%44)(e|%65|%45)(_|%5f)(m|%6d|%4d)(o|%6f|%4f)(d|%64|%44)(u|%75|%55)(l|%6c|%4c)(e|%65|%45)(s|%73|%53))(\\|\/|$)`, "i")

// nodeValidTarget reports whether Node accepts a relative target's segments.
// Other target shapes are validated by tsgo.
func nodeValidTarget(target string) bool {
	return !strings.HasPrefix(target, "./") || !invalidSegment.Test(target[2:])
}

// nodeURLBackslash reports a request whose backslashes Node may read as
// separators: one that goes through an imports map, or one whose package name
// has none, so that only the package's exports can see them.
func nodeURLBackslash(name string) bool {
	if strings.HasPrefix(name, "#") {
		return true
	}
	packageName, _ := module.ParsePackageName(strings.ReplaceAll(name, `\`, "/"))
	return !strings.Contains(name[:min(len(packageName), len(name))], `\`)
}

// nodeExactKeys collects the keys of an exports or imports map that Node
// matches without a pattern. A string or array exports only a package's root.
func nodeExactKeys(value hujson.Value) map[string]bool {
	keys := map[string]bool{".": true}
	if object, ok := value.Value.(*hujson.Object); ok {
		for _, member := range object.Members {
			if key := nodePackageMemberName(member); !strings.Contains(key, "*") {
				keys[key] = true
			}
		}
	}
	return keys
}

// invalidPatternMatch reports a request whose pattern match Node rejects.
// Node checks only the text a pattern matched; the key's own text comes from
// the package and names valid segments, so the whole subpath is checked.
func (f *nodeResolutionFS) invalidPatternMatch() bool {
	if strings.HasPrefix(f.request, "#") && !f.importsKeys[f.request] && invalidSegment.Test(f.request) {
		return true
	}
	request := f.request
	if f.importsTarget != "" {
		request = f.importsTarget
	}
	if strings.HasPrefix(request, "#") || tspath.IsExternalModuleNameRelative(request) {
		return false
	}
	packageName, _ := module.ParsePackageName(request)
	subpath := request[min(len(packageName), len(request)):]
	return subpath != "" && !f.exportsKeys["."+subpath] && invalidSegment.Test(subpath)
}

// nodeTargetPath converts an exports or imports target, after any pattern is
// substituted, to the file Node reads. The part inside the package is a URL:
// an escaped `/` or `\` anywhere in it is invalid, a query or fragment is
// dropped, a backslash separates segments, and percent-escapes are decoded to
// UTF-8 before the path is normalized.
func (f *nodeResolutionFS) nodeTargetPath(target string) (string, bool) {
	root := ""
	for _, owner := range []string{f.exportsFile, f.importsFile} {
		if owner == "" {
			continue
		}
		directory := tspath.GetDirectoryPath(owner) + "/"
		if strings.HasPrefix(target, directory) && len(directory) > len(root) {
			root = directory
		}
	}
	if root == "" {
		return target, true
	}
	if f.invalidPatternMatch() || f.unexportedDirectory() {
		return "", false
	}
	inner := target[len(root):]
	if encodedSeparator.MatchString(inner) {
		return "", false
	}
	if index := strings.IndexAny(inner, "?#"); index >= 0 {
		inner = inner[:index]
	}
	decoded, err := url.PathUnescape(strings.ReplaceAll(inner, `\`, "/"))
	// A file path cannot end in a separator.
	if err != nil || !utf8.ValidString(decoded) || strings.HasSuffix(decoded, "/") {
		return "", false
	}
	return tspath.NormalizePath(root + decoded), true
}

// unexportedDirectory reports a request ending in `/`, such as `pkg/` or
// `pkg/sub/`, that no exact key exports. In Node a pattern never matches an
// empty string, and one that matches such a request names a directory rather
// than a file, so neither resolves.
func (f *nodeResolutionFS) unexportedDirectory() bool {
	// A `/` in a query or fragment is not part of the path.
	endsInSeparator := func(request string) bool {
		if index := strings.IndexAny(request, "?#"); index >= 0 {
			request = request[:index]
		}
		return strings.HasSuffix(request, "/")
	}
	if strings.HasPrefix(f.request, "#") && endsInSeparator(f.request[1:]) && !f.importsKeys[f.request] {
		return true
	}
	request := f.request
	if f.importsTarget != "" {
		request = f.importsTarget
	}
	if strings.HasPrefix(request, "#") || tspath.IsExternalModuleNameRelative(request) {
		return false
	}
	packageName, _ := module.ParsePackageName(request)
	subpath := request[min(len(packageName), len(request)):]
	return endsInSeparator(subpath) && !f.exportsKeys["."+subpath]
}

func (f *nodeResolutionFS) DirectoryExists(name string) bool {
	if f.noModuleSearch && tspath.GetBaseFileName(name) == "node_modules" {
		return false
	}
	physical := strings.TrimSuffix(strings.TrimSuffix(f.physical(name), nodeTargetSuffix), nodeExportSuffix)
	if f.fullySpecified {
		if tspath.IsExternalModuleNameRelative(f.request) {
			if physical == tspath.ResolvePath(f.base, f.request) {
				return false
			}
		} else if strings.HasSuffix(name, "/node_modules/"+strings.TrimRight(f.request, "/")) {
			return false
		}
	}
	return (f.blockedDirectory == "" || strings.TrimRight(physical, "/") != strings.TrimRight(f.blockedDirectory, "/")) && f.FS.DirectoryExists(physical)
}

func (f *nodeResolutionFS) Realpath(name string) string {
	resolved := f.resolved[name]
	if resolved == "" {
		resolved = f.physical(name)
	}
	if f.options.PreserveSymlinks {
		return resolved
	}
	return f.FS.Realpath(resolved)
}

func (f *nodeResolutionFS) ReadFile(name string) (string, bool) {
	text, ok := f.FS.ReadFile(f.physical(name))
	if !ok || !strings.HasSuffix(name, "/package.json") {
		return text, ok
	}
	if !json.Valid([]byte(text)) {
		f.unresolved = true
		f.failure = Result{Error: "Invalid package.json at " + f.physical(name), terminal: true}
		return "", false
	}
	value, err := hujson.Parse([]byte(text))
	if err != nil {
		return text, true
	}
	if object, ok := value.Value.(*hujson.Object); ok {
		if f.options.IgnoreExports {
			object.Members = slices.DeleteFunc(object.Members, func(member hujson.ObjectMember) bool {
				key := nodePackageMemberName(member)
				return key == "exports" || key == "imports"
			})
		}
		if imports := value.Find("/imports"); imports != nil {
			filterNodeConditions(imports, f.options.Conditions)
			markNodeImportTargets(imports, f.options)
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
						f.exportsKeys = nodeExactKeys(member.Value)
					}
					filterNodeConditions(&member.Value, f.options.Conditions)
					if !prepareNodeExports(&member.Value, f.options) {
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
	f.importsKeys = nodeExactKeys(hujson.Value{Value: entries})
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
					f.fullySpecified = false
				}
			}
		} else {
			f.importsMatched = target.Value != nil
		}
	}
}

// Let tsgo select imports-map keys and conditions, including redirects to Node
// builtins that have no file for a compiler resolver to find.
func markNodeImportTargets(value *hujson.Value, options Options) {
	switch v := value.Value.(type) {
	case hujson.Literal:
		if v.Kind() == '"' {
			target := v.String()
			if options.NodeExports && !nodeValidTarget(target) {
				value.Value = hujson.Literal("null")
				return
			}
			if index := strings.IndexAny(target, "?#"); index > 0 && isURLTarget(target, options) {
				target = target[:index]
				value.Value = hujson.String(target)
			}
			if modules.IsNodeBuiltin(target) {
				value.Value = hujson.String("./" + nodeBuiltinTarget)
			} else if strings.HasPrefix(target, "./") {
				markNodeTargets(value, nodeExportSuffix)
			}
		}
	case *hujson.Object:
		for i := range v.Members {
			markNodeImportTargets(&v.Members[i].Value, options)
		}
	case *hujson.Array:
		if !prepareNodeExports(value, options) {
			value.Value = &hujson.Array{}
			return
		}
		for i := range v.Elements {
			markNodeImportTargets(&v.Elements[i], options)
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
// NodeExports skips primitive entries as well and flattens a nested array in
// place, as Node tries its entries in order before moving on.
func prepareNodeExports(value *hujson.Value, options Options) bool {
	conditions := options.Conditions
	switch object := value.Value.(type) {
	case *hujson.Object:
		for i := range object.Members {
			if !prepareNodeExports(&object.Members[i].Value, options) {
				object.Members[i].Value.Value = hujson.Literal("null")
			}
		}
	case *hujson.Array:
		targets := make([]hujson.Value, 0, len(object.Elements))
		for _, target := range object.Elements {
			switch candidate := target.Value.(type) {
			case hujson.Literal:
				if candidate.Kind() != '"' {
					if options.NodeExports {
						continue
					}
					return false
				}
				if options.NodeExports && !nodeValidTarget(candidate.String()) {
					continue
				}
			case *hujson.Array:
				if !options.NodeExports {
					continue
				}
			case *hujson.Object:
				selected, ok := nodeExportArrayCondition(candidate, conditions)
				if !ok {
					continue
				}
				target = selected
				if literal, ok := target.Value.(hujson.Literal); ok && (literal.Kind() != '"' || literal.String() == "") {
					continue
				}
				if literal, ok := target.Value.(hujson.Literal); ok && options.NodeExports && !nodeValidTarget(literal.String()) {
					continue
				}
			}
			if !prepareNodeExports(&target, options) {
				return false
			}
			if nested, ok := target.Value.(*hujson.Array); ok && options.NodeExports {
				targets = append(targets, nested.Elements...)
				continue
			}
			targets = append(targets, target)
		}
		object.Elements = targets
	case hujson.Literal:
		if object.Kind() == '"' {
			target := object.String()
			if options.NodeExports && !nodeValidTarget(target) {
				return false
			}
			if index := strings.IndexAny(target, "?#"); index > 0 && isURLTarget(target, options) {
				value.Value = hujson.String(target[:index])
			}
		}
	}
	return true
}

// isURLTarget reports a target whose `?` and `#` are removed before tsgo
// sees it. NodeExports keeps them: a bare imports target is a package request,
// taken literally, and a relative target is read as a URL only once a pattern
// has been substituted into it.
func isURLTarget(_ string, options Options) bool {
	return !options.NodeExports
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
