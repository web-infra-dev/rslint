package utils

import (
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// ModuleSettings is the `import/` settings block compiled once per Program and
// configuration. The raw settings are re-read for every reference otherwise,
// which means recompiling the import/ignore patterns and re-deriving the
// external module folders each time.
type ModuleSettings struct {
	ignore          []*esregexp.RegExp
	externalFolders []string
	key             string
}

// settingsKey identifies one compiled ModuleSettings in the Program cache.
type settingsKey struct {
	settings string
}

// SettingsFor returns the compiled `import/` settings of this rule
// configuration, compiling them on the first rule of the run that asks for
// them. Only the cache key is derived per call; compiling the settings —
// notably their regexps — happens inside the build, once per Program and key.
func SettingsFor(ctx rule.RuleContext) *ModuleSettings {
	key := moduleSettingsKey(ctx.Settings)
	settings := ctx.Settings
	return rule.CachedByProgram(ctx.Program, settingsKey{settings: key}, func() *ModuleSettings {
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
	return compiled
}

// moduleSettingsKey encodes the settings compileModuleSettings reads into one
// string: two settings maps share a key exactly when they compile alike. Every
// element is quoted so that list boundaries survive the encoding, and the
// lists are read through the same coercions the compilation applies.
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
		if expression.Test(fileName) {
			return true
		}
	}
	return false
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
// module folders, defaulting to node_modules when the setting names none.
func externalModuleFolders(settings map[string]interface{}) []string {
	var folders []string
	for _, folder := range settingsStringList(settings, "import/external-module-folders") {
		if folder != "" {
			folders = append(folders, folder)
		}
	}
	if len(folders) == 0 {
		return []string{"node_modules"}
	}
	return folders
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

func pathContainsSegment(fileName string, segment string) bool {
	if fileName == "" || segment == "" {
		return false
	}
	normalizedFileName := "/" + strings.Trim(tspath.NormalizePath(fileName), "/") + "/"
	normalizedSegment := strings.Trim(tspath.NormalizeSlashes(segment), "/")
	return strings.Contains(normalizedFileName, "/"+normalizedSegment+"/")
}
