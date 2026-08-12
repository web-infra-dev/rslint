package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

const defaultExportName = "default"

type ExportMeta struct {
	Namespace *ExportMap
}

// ExportMap records the statically visible exports of an ES module. It does
// not synthesize default exports from compiler interop settings; HasExport
// keeps that looser existence check for rules that need it.
//
// A map handed to a rule is read-only and shared: one map answers for its file
// however many files import it, and files are linted concurrently. That is why
// nothing here that writes is exported — only the builder in this package, which
// owns a map until it publishes it, may fill one in.
type ExportMap struct {
	exports    map[string]*ExportMeta
	hasUnknown bool
}

func newExportMap() *ExportMap {
	return &ExportMap{
		exports: make(map[string]*ExportMeta),
	}
}

func (m *ExportMap) Size() int {
	if m == nil {
		return 0
	}
	size := len(m.exports)
	if m.hasUnknown {
		size++
	}
	return size
}

func (m *ExportMap) Has(name string) bool {
	if m == nil {
		return false
	}
	if _, ok := m.exports[name]; ok {
		return true
	}
	return m.hasUnknown
}

func (m *ExportMap) Get(name string) *ExportMeta {
	if m == nil {
		return nil
	}
	return m.exports[name]
}

func (m *ExportMap) set(name string, meta *ExportMeta) {
	if m == nil || name == "" {
		return
	}
	if meta == nil {
		meta = &ExportMeta{}
	}
	m.exports[name] = meta
}

func (m *ExportMap) addUnknown() {
	if m != nil {
		m.hasUnknown = true
	}
}

func (m *ExportMap) mergeFrom(other *ExportMap, includeDefault bool) {
	if m == nil || other == nil {
		return
	}
	for name, meta := range other.exports {
		if !includeDefault && name == defaultExportName {
			continue
		}
		m.set(name, meta)
	}
	if other.hasUnknown {
		m.addUnknown()
	}
}

// GetExportMap resolves moduleSpecifier from ctx.SourceFile and returns a
// recursive export map. The second result is false when no export map is
// available, matching eslint-plugin-import's "imports == null" branch.
//
// The map is read-only and may be shared with every other file of the run that
// imports the same module; so may any ExportMeta.Namespace reached through it.
func GetExportMap(ctx rule.RuleContext, moduleSpecifier *ast.Node) (*ExportMap, bool) {
	if !ctx.HasProgram() || ctx.SourceFile == nil {
		return nil, false
	}
	return getExportMap(ctx.SourceFile, moduleSpecifier, newExportBuilder(IndexFor(ctx), ctx.Program()))
}

// exportBuilder carries one query's traversal state over the per-file export
// structures the index holds. building maps a file to the map being filled
// for it, so a module that re-exports its way back to itself sees the partial
// map rather than recursing forever; seen does the same for the name lookup.
//
// It also decides which of the maps it builds outlive the query. A map is
// only a property of its file when nothing in the file's re-export closure
// forms a cycle: inside a cycle, what a module ends up exporting depends on
// which module the query entered from, so those maps stay private to the
// query that built them.
type exportBuilder struct {
	index *ModuleIndex
	// runtime is the source environment every file this builder reaches belongs
	// to. It is
	// held here, on a value that lasts one query, rather than on the index,
	// which can last as long as the Program cache entry and must not keep a
	// Program alive.
	runtime  rslint_utils.ModuleResolutionRuntime
	building map[*ast.SourceFile]*ExportMap
	seen     map[exportKey]bool
	// onStack holds the files whose maps are still being filled, so that
	// reaching one again is recognized as a cycle rather than as reuse.
	onStack map[*ast.SourceFile]bool
	// unstable holds the files this query finished but found to be in or
	// below a cycle, so that reusing one keeps its dependents unstable too.
	unstable map[*ast.SourceFile]bool
	// sawCycle reports whether the file currently being built has reached a
	// cycle so far. It is saved and restored around each nested build.
	sawCycle bool
}

func newExportBuilder(index *ModuleIndex, runtime rslint_utils.ModuleResolutionRuntime) *exportBuilder {
	return &exportBuilder{
		index:    index,
		runtime:  runtime,
		building: make(map[*ast.SourceFile]*ExportMap),
		seen:     make(map[exportKey]bool),
		onStack:  make(map[*ast.SourceFile]bool),
		unstable: make(map[*ast.SourceFile]bool),
	}
}

func (builder *exportBuilder) moduleRuntime() rslint_utils.ModuleResolutionRuntime {
	if builder == nil {
		return nil
	}
	return builder.runtime
}

func getExportMap(origin *ast.SourceFile, moduleSpecifier *ast.Node, builder *exportBuilder) (*ExportMap, bool) {
	runtime := builder.moduleRuntime()
	if runtime == nil || origin == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return nil, false
	}

	link := resolveExportLink(runtime, origin, builder.index.settings, moduleSpecifier)
	if !link.Resolved {
		return nil, false
	}
	return builder.exportMapOf(link.Target), true
}

// exportMapOf applies a file's export steps in source order. The map is
// registered before the steps run, so a dependency that re-exports its way
// back here merges whatever is complete at that point — which is what
// re-walking the statements would also produce.
//
// A file whose closure turned out to be free of cycles is handed to the index
// afterwards, because its map cannot depend on where the query started: a
// dependency can only be half-built when it is also waiting on this file,
// which is exactly the case a cycle-free closure rules out.
func (builder *exportBuilder) exportMapOf(sourceFile *ast.SourceFile) *ExportMap {
	if sourceFile == nil {
		return newExportMap()
	}
	if existing := builder.building[sourceFile]; existing != nil {
		if builder.onStack[sourceFile] || builder.unstable[sourceFile] {
			builder.sawCycle = true
		}
		return existing
	}
	if cached := builder.index.cachedExportMap(sourceFile); cached != nil {
		return cached
	}

	exports := newExportMap()
	builder.building[sourceFile] = exports
	builder.onStack[sourceFile] = true
	enclosing := builder.sawCycle
	builder.sawCycle = false

	local := builder.index.localExportsOf(builder.moduleRuntime(), sourceFile)
	for _, step := range local.Steps {
		builder.applyStep(exports, local, step)
	}

	delete(builder.onStack, sourceFile)
	if builder.sawCycle {
		builder.unstable[sourceFile] = true
	} else {
		builder.index.storeExportMap(sourceFile, exports)
	}
	builder.sawCycle = enclosing || builder.sawCycle
	return exports
}

func (builder *exportBuilder) applyStep(exports *ExportMap, local *localExports, step exportStep) {
	switch step.Kind {
	case exportStepNames:
		for _, name := range step.Names {
			exports.set(name, nil)
		}

	case exportStepLocalDefault:
		if meta, ok := builder.namespaceImportMeta(local, step.Local); ok {
			exports.set(defaultExportName, meta)
		} else {
			exports.set(defaultExportName, nil)
		}

	case exportStepStar:
		if !step.Link.Resolved {
			exports.addUnknown()
			return
		}
		exports.mergeFrom(builder.exportMapOf(step.Link.Target), false)

	case exportStepNamed:
		var dependency *ExportMap
		if step.FromModule && step.Link.Resolved {
			dependency = builder.exportMapOf(step.Link.Target)
		}
		for _, spec := range step.Specs {
			if !step.FromModule {
				if meta, ok := builder.namespaceImportMeta(local, spec.LocalIdent); ok {
					exports.set(spec.Exported, meta)
				} else {
					exports.set(spec.Exported, nil)
				}
				continue
			}
			if !spec.LocalOK {
				continue
			}
			if dependency == nil {
				exports.set(spec.Exported, nil)
				continue
			}
			if !dependency.Has(spec.Local) {
				continue
			}
			exports.set(spec.Exported, dependency.Get(spec.Local))
		}
	}
}

// namespaceImportMeta finds the `import * as localName` the file re-exports
// under some other name. Every import declaration before the matching one has
// its export map built along the way, so by the time a match is returned the
// maps of all earlier imports exist and stay visible to the rest of this
// query.
func (builder *exportBuilder) namespaceImportMeta(local *localExports, localName string) (*ExportMeta, bool) {
	if localName == "" {
		return nil, false
	}
	for _, binding := range local.Imports {
		if !binding.Link.Resolved {
			continue
		}
		imports := builder.exportMapOf(binding.Link.Target)
		if binding.NamespaceName == localName {
			return &ExportMeta{Namespace: imports}, true
		}
	}
	return nil, false
}

func moduleExportName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	return rslint_utils.GetStaticPropertyName(node)
}
