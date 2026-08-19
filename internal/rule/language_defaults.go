package rule

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// languageGlobal is one global supplied by the resolved language defaults,
// independently of authored languageOptions.globals.
type languageGlobal struct {
	name   string
	access utils.GlobalAccess
}

// GlobalsInit is immutable initialization data supplied to Globals by the
// resolved language defaults. Its zero value supplies no additional names.
//
// The backing entries are package-owned static data. Callers must treat values
// returned by ResolveLanguageDefaults as read-only.
type GlobalsInit struct {
	entries []languageGlobal
}

func (i GlobalsInit) access(name string) utils.GlobalAccess {
	for _, entry := range i.entries {
		if entry.name == name {
			return entry.access
		}
	}
	return utils.GlobalAccessUnset
}

// RefStoreInit is immutable initialization data supplied to RefStore by the
// resolved language defaults. Its zero value adds no reference or scope facts.
// Implicit wrapper bindings are facts, not synthetic TypeScript symbols.
type RefStoreInit struct {
	implicitWrapperBindings []string
	nonGlobalTopLevelScope  bool
	globalTopLevelScope     bool
}

func (i RefStoreInit) hasImplicitWrapperBinding(name string) bool {
	for _, candidate := range i.implicitWrapperBindings {
		if candidate == name {
			return true
		}
	}
	return false
}

// languageGlobalCatalog is the set ApplyTo must gate even when the resolved
// defaults supply no entry. This prevents a caller-provided seed (or a reused
// destination map) from leaking one source's defaults into another.
var languageGlobalCatalog = []languageGlobal{
	{name: "exports", access: utils.GlobalAccessWritable},
	{name: "global", access: utils.GlobalAccessReadonly},
	{name: "module", access: utils.GlobalAccessReadonly},
	{name: "require", access: utils.GlobalAccessReadonly},
}

var commonJSGlobalsInit = GlobalsInit{entries: languageGlobalCatalog}

var moduleRefStoreInit = RefStoreInit{nonGlobalTopLevelScope: true}

var scriptRefStoreInit = RefStoreInit{globalTopLevelScope: true}

var commonJSRefStoreInit = RefStoreInit{
	implicitWrapperBindings: []string{"arguments"},
	nonGlobalTopLevelScope:  true,
}

var javascriptSourceExtensions = []string{
	tspath.ExtensionJs,
	tspath.ExtensionJsx,
	tspath.ExtensionMjs,
	tspath.ExtensionCjs,
}

func isJavaScriptSourceExtension(fileName string) bool {
	ext := tspath.GetAnyExtensionFromPath(fileName, nil, false)
	return tspath.ExtensionIsOneOf(ext, javascriptSourceExtensions)
}

// ResolveLanguageDefaults resolves the concrete Globals and RefStore
// initialization supplied by ESLint's default language selection for fileName,
// together with effective language options.
//
// An omitted source type is filled from the filename for JavaScript files:
// .js/.mjs select module, .cjs selects commonjs. Other extensions, including
// .jsx/.ts/.tsx/.cts, keep the empty value. Authored sourceType then selects
// the inits on every extension. module contributes a non-global top-level
// scope. commonjs contributes writable exports and read-only
// global/module/require everywhere; on espree-parsed extensions
// (.js/.jsx/.mjs/.cjs) it adds the non-global wrapper scope and the
// wrapper-local arguments binding, while TypeScript-flavoured extensions stop
// at the four globals, matching typescript-eslint's scope manager. script
// forces a global program scope even when module syntax is present; the
// still-empty TypeScript/JSX value contributes no defaults.
func ResolveLanguageDefaults(fileName string, languageOptions LanguageOptions) (GlobalsInit, RefStoreInit, LanguageOptions) {
	if languageOptions.SourceType == "" {
		switch tspath.GetAnyExtensionFromPath(fileName, nil, false) {
		case tspath.ExtensionJs, tspath.ExtensionMjs:
			languageOptions.SourceType = "module"
		case tspath.ExtensionCjs:
			languageOptions.SourceType = "commonjs"
		}
	}
	switch languageOptions.SourceType {
	case "commonjs":
		if isJavaScriptSourceExtension(fileName) {
			return commonJSGlobalsInit, commonJSRefStoreInit, languageOptions
		}
		return commonJSGlobalsInit, RefStoreInit{}, languageOptions
	case "module":
		return GlobalsInit{}, moduleRefStoreInit, languageOptions
	case "script":
		return GlobalsInit{}, scriptRefStoreInit, languageOptions
	}
	return GlobalsInit{}, RefStoreInit{}, languageOptions
}
