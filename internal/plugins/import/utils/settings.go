package utils

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// ModuleSettings is the `import/` settings block compiled once per Program
// generation and configuration. The raw settings are re-read for every
// reference otherwise, which means recompiling the import/ignore patterns and
// re-deriving the external module folders each time.
type ModuleSettings struct {
	ignore          []*esregexp.RegExp
	internalRegex   *esregexp.RegExp
	coreModules     map[string]struct{}
	externalFolders []string
	key             string
}

// settingsKey identifies one compiled ModuleSettings in the source cache.
type settingsKey struct {
	settings string
}

// SettingsFor returns the compiled `import/` settings of this rule
// configuration, compiling them on the first rule of the run that asks for
// them. Only the cache key is derived per call; compiling the settings —
// notably their regexps — happens inside the build, once per Program and key.
func SettingsFor(ctx rule.RuleContext) *ModuleSettings {
	key := compiledModuleSettingsCacheKey(ctx.Settings)
	settings := ctx.Settings
	return rule.CachedByProgram(ctx, settingsKey{settings: key}, func() *ModuleSettings {
		return compileModuleSettings(settings)
	})
}

// compileModuleSettings compiles the `import/` settings that decide which
// references count and which resolved paths are external.
func compileModuleSettings(settings map[string]interface{}) *ModuleSettings {
	compiled := &ModuleSettings{
		externalFolders: externalModuleFolders(settings),
		key:             moduleSettingsKey(settings),
	}
	if names := settingsStringList(settings, "import/core-modules"); len(names) != 0 {
		compiled.coreModules = make(map[string]struct{}, len(names))
		for _, name := range names {
			compiled.coreModules[name] = struct{}{}
		}
	}
	for _, pattern := range settingsStringList(settings, "import/ignore") {
		if expression, err := esregexp.Compile(pattern, ""); err == nil {
			compiled.ignore = append(compiled.ignore, expression)
		}
	}
	if pattern, ok := settingString(settings, "import/internal-regex"); ok && pattern != "" {
		if expression, err := esregexp.Compile(pattern, ""); err == nil {
			compiled.internalRegex = expression
		}
	}
	return compiled
}

// moduleSettingsKey encodes the settings used by shared module indexes. Every
// element is quoted so list boundaries survive the encoding. Classification-
// only settings are intentionally excluded because they do not change those
// indexes; compiledModuleSettingsCacheKey adds them for SettingsFor's cache.
func moduleSettingsKey(settings map[string]interface{}) string {
	var key strings.Builder
	for _, pattern := range settingsStringList(settings, "import/ignore") {
		key.WriteString(strconv.Quote(pattern))
	}
	key.WriteByte('\x00')
	for _, folder := range externalModuleFolders(settings) {
		key.WriteString(strconv.Quote(folder))
	}
	return key.String()
}

func compiledModuleSettingsCacheKey(settings map[string]interface{}) string {
	var key strings.Builder
	key.WriteString(moduleSettingsKey(settings))
	key.WriteByte('\x00')
	for _, name := range settingsStringList(settings, "import/core-modules") {
		key.WriteString(strconv.Quote(name))
	}
	key.WriteByte('\x00')
	if pattern, ok := settingString(settings, "import/internal-regex"); ok {
		key.WriteString(strconv.Quote(pattern))
	}
	return key.String()
}

// Key identifies the settings these were compiled from, for callers that key
// their own derived state by them.
func (compiled *ModuleSettings) Key() string {
	if compiled == nil {
		return ""
	}
	return compiled.key
}

// IsIgnoredPath reports whether `import/ignore` covers a resolved import
// target path.
func (compiled *ModuleSettings) IsIgnoredPath(fileName string) bool {
	if compiled == nil {
		return false
	}
	for _, expression := range compiled.ignore {
		if expression.TestOrTimeout(fileName) {
			return true
		}
	}
	return false
}

// IsInternalSpecifier reports whether `import/internal-regex` classifies the
// written module specifier as internal.
func (compiled *ModuleSettings) IsInternalSpecifier(specifier string) bool {
	return compiled != nil && compiled.internalRegex != nil && compiled.internalRegex.TestOrTimeout(specifier)
}

// IsNodeBuiltinSpecifier reports whether the complete written specifier is a
// Node.js builtin. TypeScript-aware ESLint resolvers perform this exact check
// before filesystem resolution, so an installed package or paths mapping with
// the same name cannot shadow the builtin.
func IsNodeBuiltinSpecifier(specifier string) bool {
	return specifier != "" && core.NodeCoreModules()[specifier]
}

// IsCoreModuleSpecifier reports whether the specifier's package root is a
// Node.js builtin or is listed by `import/core-modules`. Resolution remains a
// caller concern for non-exact builtin subpath specifiers and configured core
// modules, because a resolved project module can override their lexical
// classification.
func (compiled *ModuleSettings) IsCoreModuleSpecifier(specifier string) bool {
	if specifier == "" {
		return false
	}
	base := baseModuleName(specifier)
	if core.NodeCoreModules()[base] {
		return true
	}
	if compiled == nil {
		return false
	}
	_, configured := compiled.coreModules[base]
	return configured
}

// IsScopedModuleSpecifier mirrors eslint-plugin-import's lexical scoped-name
// check. It counts UTF-16 code units, as JavaScript RegExp does without the
// Unicode flag; this matters for an astral character immediately after `@`.
func IsScopedModuleSpecifier(specifier string) bool {
	if len(specifier) < 2 || specifier[0] != '@' {
		return false
	}

	afterAt := specifier[1:]
	slash := strings.IndexByte(afterAt, '/')
	segment := afterAt
	if slash >= 0 {
		segment = afterAt[:slash]
	}
	segmentUnits := ecmascript.StringCodeUnitCount(segment)
	return segmentUnits >= 2 || segmentUnits == 1 && slash >= 0 && slash+1 < len(afterAt) && afterAt[slash+1] != '/'
}

// IsExternalPath reports whether a resolved path or unresolved bare specifier
// should be treated as external.
func (compiled *ModuleSettings) IsExternalPath(specifier string, resolvedPath string) bool {
	if compiled == nil {
		return false
	}
	for _, folder := range compiled.externalFolders {
		// Resolving an empty configured folder from the importing package
		// yields that package root. Together with the outside-package check in
		// eslint-plugin-import, it classifies every resolved target as external.
		if folder == "" && resolvedPath != "" {
			return true
		}
		if pathContainsSegment(resolvedPath, folder) {
			return true
		}
	}
	return specifier != "" && !tspath.IsExternalModuleNameRelative(specifier) && resolvedPath == ""
}

// IsExternalPathFromPackage classifies a resolved target relative to the
// importing package. A target outside that package is external; a target
// inside it is external only when it is below a configured external-module
// folder. Relative folders resolve from packagePath.
func (compiled *ModuleSettings) IsExternalPathFromPackage(packagePath, resolvedPath string, caseSensitive bool) bool {
	if compiled == nil || resolvedPath == "" {
		return false
	}
	compareOptions := tspath.ComparePathsOptions{UseCaseSensitiveFileNames: caseSensitive}
	if packagePath != "" && !tspath.ContainsPath(packagePath, resolvedPath, compareOptions) {
		return true
	}
	for _, folder := range compiled.externalFolders {
		folderPath := folder
		if !tspath.IsRootedDiskPath(folderPath) {
			if packagePath == "" {
				if pathContainsSegment(resolvedPath, folderPath) {
					return true
				}
				continue
			}
			folderPath = tspath.ResolvePath(packagePath, folderPath)
		}
		if tspath.ContainsPath(folderPath, resolvedPath, compareOptions) {
			return true
		}
	}
	return false
}

// externalModuleFolders returns eslint-plugin-import's configured external
// module folders. The default applies only when the setting is absent or not
// an array; an explicit empty array disables it, matching JavaScript truthiness.
func externalModuleFolders(settings map[string]interface{}) []string {
	if settings == nil {
		return []string{"node_modules"}
	}
	raw, configured := settings["import/external-module-folders"]
	if !configured {
		return []string{"node_modules"}
	}
	var values []string
	switch typed := raw.(type) {
	case []string:
		values = typed
	case []interface{}:
		values = make([]string, 0, len(typed))
		for _, item := range typed {
			if value, ok := item.(string); ok {
				values = append(values, value)
			}
		}
	default:
		return []string{"node_modules"}
	}
	// Keep empty strings: resolving one from packagePath yields the package
	// root, so this configuration classifies every resolved path below it as
	// external.
	return values
}

func settingsStringList(settings map[string]interface{}, name string) []string {
	if settings == nil {
		return nil
	}
	switch typed := settings[name].(type) {
	case []string:
		return typed
	case []interface{}:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if value, ok := item.(string); ok {
				values = append(values, value)
			}
		}
		return values
	}
	return nil
}

func settingString(settings map[string]interface{}, name string) (string, bool) {
	if settings == nil {
		return "", false
	}
	value, ok := settings[name].(string)
	return value, ok
}

// baseModuleName mirrors eslint-plugin-import's package-root extraction,
// including its treatment of malformed scoped names such as `@scope`.
func baseModuleName(name string) string {
	if IsScopedModuleSpecifier(name) {
		firstSlash := strings.IndexByte(name, '/')
		if firstSlash < 0 {
			return name + "/undefined"
		}
		if nextSlash := strings.IndexByte(name[firstSlash+1:], '/'); nextSlash >= 0 {
			return name[:firstSlash+1+nextSlash]
		}
		return name
	}
	if slash := strings.IndexByte(name, '/'); slash >= 0 {
		return name[:slash]
	}
	return name
}

func pathContainsSegment(fileName string, segment string) bool {
	if fileName == "" || segment == "" {
		return false
	}
	normalizedFileName := "/" + strings.Trim(tspath.NormalizePath(fileName), "/") + "/"
	normalizedSegment := strings.Trim(tspath.NormalizeSlashes(segment), "/")
	return strings.Contains(normalizedFileName, "/"+normalizedSegment+"/")
}
