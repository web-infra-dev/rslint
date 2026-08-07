package utils

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ModuleIndex answers, for any file in one Program, the questions every
// import rule asks about a file in isolation: which modules it references,
// what they resolve to, and what the file says about its own exports. Each
// answer is a function of that file's own syntax, so it is computed once for
// the whole lint run however many files ask for it, and however many times.
//
// It holds only file-local facts. What a rule then does with them — which
// references it treats as graph edges, how it walks between files — stays in
// the rule, because those choices belong to the rule's options rather than to
// the Program.
type ModuleIndex struct {
	ctx      rule.RuleContext
	settings *ModuleSettings

	mu      sync.Mutex
	refs    map[refKey][]ModuleReference
	exports map[*ast.SourceFile]*LocalExports
}

// refKey pairs a file with the module syntax the caller wants recognized.
// Rules disagree about whether require() and define() count, and the answer
// differs accordingly.
type refKey struct {
	file    *ast.SourceFile
	options ModuleReferenceOptions
}

// ModuleSettings is the `import/` settings block compiled once. The raw
// settings are re-read for every reference otherwise, which means recompiling
// the import/ignore patterns and re-deriving the external module folders each
// time.
type ModuleSettings struct {
	ignore          []*regexp.Regexp
	externalFolders []string
	key             string
}

// indexKey identifies one ModuleIndex. Only the settings appear: they are the
// one input that changes every answer the index gives.
type indexKey struct {
	settings string
}

// IndexFor returns the Program's module index, building it on the first rule
// of the run that asks for it.
func IndexFor(ctx rule.RuleContext) *ModuleIndex {
	settings := CompileModuleSettings(ctx.Settings)
	index, _ := ctx.ProgramCache.Load(indexKey{settings: settings.key}, func() any {
		return &ModuleIndex{
			ctx:      ctx,
			settings: settings,
			refs:     make(map[refKey][]ModuleReference),
			exports:  make(map[*ast.SourceFile]*LocalExports),
		}
	}).(*ModuleIndex)
	return index
}

// Settings returns the compiled `import/` settings this index was built with.
func (index *ModuleIndex) Settings() *ModuleSettings {
	if index == nil {
		return nil
	}
	return index.settings
}

// Files returns every file of the Program, in the Program's own order. A
// file's position in this slice is stable for the lifetime of the index, so
// callers that need a dense numbering of the Program can adopt it.
func (index *ModuleIndex) Files() []*ast.SourceFile {
	if index == nil || index.ctx.Program == nil {
		return nil
	}
	return index.ctx.Program.SourceFiles()
}

// Refs returns file's module references in source order, with each
// reference's target already resolved. The result is shared with every other
// caller and must not be modified.
func (index *ModuleIndex) Refs(file *ast.SourceFile, options ModuleReferenceOptions) []ModuleReference {
	if index == nil || file == nil {
		return nil
	}

	key := refKey{file: file, options: options}
	index.mu.Lock()
	refs, ok := index.refs[key]
	index.mu.Unlock()
	if ok {
		return refs
	}

	// Computed outside the lock: collection is pure, so two files racing on
	// their first request cost one redundant collection at worst, which is
	// cheaper than serializing every file in the run behind one mutex.
	refs = collectModuleReferences(index.ctx, file, options, index.settings)

	index.mu.Lock()
	if existing, ok := index.refs[key]; ok {
		refs = existing
	} else {
		index.refs[key] = refs
	}
	index.mu.Unlock()
	return refs
}

// Exports returns what file says about its own exports, with every module
// specifier resolved and none of them followed. The result is shared with
// every other caller and must not be modified.
func (index *ModuleIndex) Exports(file *ast.SourceFile) *LocalExports {
	if index == nil || file == nil {
		return nil
	}

	index.mu.Lock()
	exports, ok := index.exports[file]
	index.mu.Unlock()
	if ok {
		return exports
	}

	exports = collectLocalExports(index.ctx, file, index.settings)

	index.mu.Lock()
	if existing, ok := index.exports[file]; ok {
		exports = existing
	} else {
		index.exports[file] = exports
	}
	index.mu.Unlock()
	return exports
}

// CompileModuleSettings compiles the `import/` settings that decide which
// references survive collection and which resolved paths count as external.
func CompileModuleSettings(settings map[string]interface{}) *ModuleSettings {
	compiled := &ModuleSettings{
		externalFolders: ExternalModuleFolders(settings),
		key:             moduleSettingsKey(settings),
	}
	for _, pattern := range settingsStringList(settings, "import/ignore") {
		if expression, err := regexp.Compile(pattern); err == nil {
			compiled.ignore = append(compiled.ignore, expression)
		}
	}
	return compiled
}

func moduleSettingsKey(settings map[string]interface{}) string {
	if settings == nil {
		return ""
	}
	return fmt.Sprint(settings["import/ignore"], "\x00", settings["import/external-module-folders"])
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
		if expression.MatchString(fileName) {
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
