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
// Implicit bindings are facts, not synthetic TypeScript symbols.
type RefStoreInit struct {
	implicitBindings       []string
	nonGlobalTopLevelScope bool
}

func (i RefStoreInit) defines(name string) bool {
	for _, candidate := range i.implicitBindings {
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

var commonJSRefStoreInit = RefStoreInit{
	implicitBindings:       []string{"arguments"},
	nonGlobalTopLevelScope: true,
}

// ResolveLanguageDefaults resolves the concrete Globals and RefStore
// initialization supplied by ESLint's default language selection for fileName.
// It deliberately does not read authored sourceType, package.json, or
// TypeScript-flavoured extensions.
//
// JavaScript module files contribute only their non-global top-level scope.
// An exact, case-sensitive .cjs extension contributes CommonJS's four globals,
// non-global wrapper scope, and implicit wrapper arguments binding. Other
// extensions return zero values.
func ResolveLanguageDefaults(fileName string) (GlobalsInit, RefStoreInit) {
	switch tspath.GetAnyExtensionFromPath(fileName, nil, false) {
	case tspath.ExtensionJs, tspath.ExtensionMjs:
		return GlobalsInit{}, moduleRefStoreInit
	case tspath.ExtensionCjs:
		return commonJSGlobalsInit, commonJSRefStoreInit
	}
	return GlobalsInit{}, RefStoreInit{}
}
