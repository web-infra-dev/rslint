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
// import rule asks about a file in isolation: which modules it references and
// which files those resolve to. Each answer is a function of that file's own
// syntax, so it is computed once for the whole lint run however many files ask
// for it, and however many times.
//
// It holds only file-local facts. What a rule then does with them — which
// references it treats as graph edges, how it walks between files — stays in
// the rule, because those choices belong to the rule's options rather than to
// the Program.
type ModuleIndex struct {
	ctx      rule.RuleContext
	options  ModuleReferenceOptions
	settings *ModuleSettings

	mu   sync.Mutex
	refs map[*ast.SourceFile][]ModuleReference
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

// IndexKey identifies one ModuleIndex. Only what changes the references
// themselves belongs here: the syntax the collector recognizes, and the
// settings that drop references as they are collected.
type IndexKey struct {
	Options  ModuleReferenceOptions
	Settings string
}

// IndexFor returns the Program's module index for these collection options,
// building it on the first rule of the run that asks for it. Rules that
// disagree about which module syntax counts get their own index.
func IndexFor(ctx rule.RuleContext, options ModuleReferenceOptions) *ModuleIndex {
	settings := CompileModuleSettings(ctx.Settings)
	key := IndexKey{Options: options, Settings: settings.key}
	index, _ := ctx.ProgramCache.Load(key, func() any {
		return &ModuleIndex{
			ctx:      ctx,
			options:  options,
			settings: settings,
			refs:     make(map[*ast.SourceFile][]ModuleReference),
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

// Files returns every file of the Program, in the Program's own order. The
// index of a file in this slice is stable for the lifetime of the index, so
// callers that need a dense numbering of the Program can adopt it.
func (index *ModuleIndex) Files() []*ast.SourceFile {
	if index == nil || index.ctx.Program == nil {
		return nil
	}
	return index.ctx.Program.SourceFiles()
}

// Refs returns file's module references in source order, with each reference's
// target already resolved. The result is shared with every other caller and
// must not be modified.
func (index *ModuleIndex) Refs(file *ast.SourceFile) []ModuleReference {
	if index == nil || file == nil {
		return nil
	}

	index.mu.Lock()
	refs, ok := index.refs[file]
	index.mu.Unlock()
	if ok {
		return refs
	}

	// Collected outside the lock: collection is pure, so two files racing on
	// their first request cost one redundant collection at worst, which is
	// cheaper than serializing every file in the run behind one mutex.
	refs = collectModuleReferences(index.ctx, file, index.options, index.settings)

	index.mu.Lock()
	if existing, ok := index.refs[file]; ok {
		refs = existing
	} else {
		index.refs[file] = refs
	}
	index.mu.Unlock()
	return refs
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
