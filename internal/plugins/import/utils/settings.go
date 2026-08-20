package utils

import (
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// ModuleSettings is the `import/` settings block compiled once per Program
// generation and configuration. The raw settings are re-read for every
// reference otherwise, which means recompiling the import/ignore patterns and
// re-deriving the external module folders each time.
type ModuleSettings struct {
	ignore          []*esregexp.RegExp
	internalRegex   *esregexp.RegExp
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
// element is quoted so list boundaries survive the encoding. Internal-regex
// classification is intentionally excluded because it does not change those
// indexes; compiledModuleSettingsCacheKey adds it for SettingsFor's cache.
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
	key := moduleSettingsKey(settings)
	if pattern, ok := settingString(settings, "import/internal-regex"); ok {
		key += "\x00" + strconv.Quote(pattern)
	}
	return key
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

// IsExternalPath reports whether a resolved path or unresolved bare specifier
// should be treated as external.
func (compiled *ModuleSettings) IsExternalPath(specifier string, resolvedPath string) bool {
	if compiled == nil {
		return false
	}
	for _, folder := range compiled.externalFolders {
		if pathContainsSegment(resolvedPath, folder) {
			return true
		}
	}
	return specifier != "" && !tspath.IsExternalModuleNameRelative(specifier) && resolvedPath == ""
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
	// Keep empty strings: Node's path.resolve(packagePath, "") yields the
	// package root, so upstream intentionally treats every resolved path under
	// that package as external for this configuration.
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

func pathContainsSegment(fileName string, segment string) bool {
	if fileName == "" || segment == "" {
		return false
	}
	normalizedFileName := "/" + strings.Trim(tspath.NormalizePath(fileName), "/") + "/"
	normalizedSegment := strings.Trim(tspath.NormalizeSlashes(segment), "/")
	return strings.Contains(normalizedFileName, "/"+normalizedSegment+"/")
}
