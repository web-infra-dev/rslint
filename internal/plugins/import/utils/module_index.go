package utils

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ModuleIndex answers what each file of one Program says about its own
// exports. That answer depends only on the file's syntax, so it is derived
// once for the whole lint run however many files ask for it.
//
// What a file *imports* is not here: the core module graph on RuleContext
// answers that for every rule, not just these. This holds the part that is
// specific to eslint-plugin-import — export maps, and the `import/` settings
// that decide which references count.
type ModuleIndex struct {
	settings *ModuleSettings

	mu      sync.Mutex
	exports map[*ast.SourceFile]*localExports
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
	return CachedByProgram(ctx.Program, indexKey{settings: settings.Key()}, func() *ModuleIndex {
		return newModuleIndex(settings)
	})
}

func newModuleIndex(settings *ModuleSettings) *ModuleIndex {
	return &ModuleIndex{
		settings:   settings,
		exports:    make(map[*ast.SourceFile]*localExports),
		exportMaps: make(map[*ast.SourceFile]*ExportMap),
	}
}

// localExportsOf returns what file says about its own exports, with every
// module specifier resolved and none of them followed. The result is shared
// with every other caller and must not be modified.
func (index *ModuleIndex) localExportsOf(ctx rule.RuleContext, file *ast.SourceFile) *localExports {
	if index == nil || file == nil {
		return nil
	}

	index.mu.Lock()
	exports, ok := index.exports[file]
	index.mu.Unlock()
	if ok {
		return exports
	}

	exports = collectLocalExports(ctx, file, index.settings)

	index.mu.Lock()
	if existing, ok := index.exports[file]; ok {
		exports = existing
	} else {
		index.exports[file] = exports
	}
	index.mu.Unlock()
	return exports
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
