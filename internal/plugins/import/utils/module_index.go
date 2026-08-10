package utils

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

// ModuleIndex answers what each file of one Program says about its own
// exports. That answer depends only on the file's syntax, so it is derived
// once for the whole lint run however many files ask for it.
//
// What a file *imports* is not here: the core module graph on RuleContext
// answers that for every rule, not just these. This holds the part that is
// specific to eslint-plugin-import — export maps, and the `import/` settings
// that decide which references count.
//
// The index lives in the Program cache, so it must not hold the Program it was
// built for: an entry that did would keep its Program reachable forever, the
// weak key's cleanup would never run, and every editor keystroke — each of
// which produces a new Program — would strand one. The Program is therefore
// passed in by whoever is asking; every file of one run shares it, so any
// caller's is the same one.
type ModuleIndex struct {
	settings *ModuleSettings

	exports rslint_utils.LazyMap[*ast.SourceFile, *localExports]

	mu sync.Mutex
	// exportMaps holds the fully merged export maps that turned out not to
	// depend on where the query that built them started. The builder decides
	// which those are; see exportMapOf.
	exportMaps map[*ast.SourceFile]*ExportMap
}

// indexKey identifies one index. Only the settings appear: they are the one
// input that changes every answer it gives.
type indexKey struct {
	settings string
}

// IndexFor returns the Program's import index for these settings, building it
// on the first rule of the run that asks for it.
func IndexFor(ctx rule.RuleContext) *ModuleIndex {
	settings := SettingsFor(ctx)
	return rule.CachedByProgram(ctx.Program, indexKey{settings: settings.Key()}, func() *ModuleIndex {
		return newModuleIndex(settings)
	})
}

func newModuleIndex(settings *ModuleSettings) *ModuleIndex {
	return &ModuleIndex{
		settings:   settings,
		exportMaps: make(map[*ast.SourceFile]*ExportMap),
	}
}

// localExportsOf returns what file says about its own exports, with every
// module specifier resolved and none of them followed. The result is shared
// with every other caller and must not be modified.
func (index *ModuleIndex) localExportsOf(program *compiler.Program, file *ast.SourceFile) *localExports {
	if index == nil || file == nil {
		return nil
	}
	return index.exports.Get(file, func() *localExports {
		return collectLocalExports(program, file, index.settings)
	})
}

// cachedExportMap returns the merged export map of a file whose closure was
// found to be free of re-export cycles, or nil when there is none to reuse.
// The result is read-only: every caller shares it.
func (index *ModuleIndex) cachedExportMap(file *ast.SourceFile) *ExportMap {
	if index == nil {
		return nil
	}
	index.mu.Lock()
	defer index.mu.Unlock()
	return index.exportMaps[file]
}

func (index *ModuleIndex) storeExportMap(file *ast.SourceFile, exports *ExportMap) {
	if index == nil {
		return
	}
	index.mu.Lock()
	defer index.mu.Unlock()
	if _, ok := index.exportMaps[file]; !ok {
		index.exportMaps[file] = exports
	}
}
