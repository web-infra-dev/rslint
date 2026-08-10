package rule

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// fileLanguageGlobal is one global supplied by a file's default language
// environment, independently of authored languageOptions.globals.
type fileLanguageGlobal struct {
	name   string
	access utils.GlobalAccess
}

// FileLanguageGlobals is the immutable set of globals supplied by a file's
// default language environment. Its zero value supplies no additional names.
//
// The backing entries are package-owned static data. Callers must treat values
// returned by ResolveFileLanguageDefaults as read-only.
type FileLanguageGlobals struct {
	entries []fileLanguageGlobal
}

func (g FileLanguageGlobals) access(name string) utils.GlobalAccess {
	for _, entry := range g.entries {
		if entry.name == name {
			return entry.access
		}
	}
	return utils.GlobalAccessUnset
}

// FileScopeDefaults is the immutable scope initialization supplied by a
// file's default language environment. Its zero value adds no scope facts.
// Implicit bindings are facts, not synthetic TypeScript symbols.
type FileScopeDefaults struct {
	implicitBindings       []string
	nonGlobalTopLevelScope bool
}

func (d FileScopeDefaults) defines(name string) bool {
	for _, candidate := range d.implicitBindings {
		if candidate == name {
			return true
		}
	}
	return false
}

// fileLanguageGlobalCatalog is the set ApplyTo must gate even when a
// particular file receives no entry. This prevents a caller-provided seed (or
// a reused destination map) from leaking one file's defaults into another.
var fileLanguageGlobalCatalog = []fileLanguageGlobal{
	{name: "exports", access: utils.GlobalAccessWritable},
	{name: "global", access: utils.GlobalAccessReadonly},
	{name: "module", access: utils.GlobalAccessReadonly},
	{name: "require", access: utils.GlobalAccessReadonly},
}

var commonJSFileLanguageGlobals = FileLanguageGlobals{entries: fileLanguageGlobalCatalog}

var moduleFileScopeDefaults = FileScopeDefaults{nonGlobalTopLevelScope: true}

var commonJSFileScopeDefaults = FileScopeDefaults{
	implicitBindings:       []string{"arguments"},
	nonGlobalTopLevelScope: true,
}

// ResolveFileLanguageDefaults resolves the concrete data supplied by ESLint's
// default language environment for fileName. It deliberately does not read
// authored sourceType, package.json, or TypeScript-flavoured extensions.
//
// JavaScript module files contribute only their non-global top-level scope.
// An exact, case-sensitive .cjs extension contributes CommonJS's four globals,
// non-global wrapper scope, and implicit wrapper arguments binding. Other
// extensions return zero values.
func ResolveFileLanguageDefaults(fileName string) (FileLanguageGlobals, FileScopeDefaults) {
	switch tspath.GetAnyExtensionFromPath(fileName, nil, false) {
	case tspath.ExtensionJs, tspath.ExtensionMjs:
		return FileLanguageGlobals{}, moduleFileScopeDefaults
	case tspath.ExtensionCjs:
		return commonJSFileLanguageGlobals, commonJSFileScopeDefaults
	}
	return FileLanguageGlobals{}, FileScopeDefaults{}
}
