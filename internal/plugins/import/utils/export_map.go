package utils

import (
	"iter"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

const defaultExportName = "default"

type exportBindingKind uint8

const (
	exportBindingLocal exportBindingKind = iota
	exportBindingExternal
)

// exportBinding identifies the source binding of an export so `export *`
// merges can distinguish re-exporting the same binding through multiple paths
// from colliding distinct bindings with the same exported name.
type exportBinding struct {
	file *ast.SourceFile
	name string
	kind exportBindingKind
}

type ExportMeta struct {
	Namespace  *ExportMap
	doc        *exportDocSource
	originFile *ast.SourceFile
	originName string
	originKind exportBindingKind
	// originLocal reports whether (originFile, originName, originKind) already
	// identifies the terminal binding (rather than an indirect exported-name
	// reference from local.ValueImports resolved via resolveBindingOrigin).
	originLocal bool
	// unresolved distinguishes a declared re-export from one whose value is
	// available. Names must retain missing re-exports for import/export.
	unresolved bool
	// synthetic marks a default placeholder fabricated from implicitDefault or
	// NodeDefault when no explicit default export existed yet. Explicit exports
	// outrank synthetic defaults during cyclic SCC convergence.
	synthetic bool
	// conflict marks an ambiguous `export *` collision between two distinct
	// bindings of the same name. It retains Has() == true for export listing
	// while clearing Namespace so consumers do not falsely validate a member
	// against either conflicting side.
	conflict bool
}

func (m *ExportMeta) originBinding() exportBinding {
	return exportBinding{
		file: m.originFile,
		name: m.originName,
		kind: m.originKind,
	}
}

const (
	exportRankMissing    = 0
	exportRankUnresolved = 1
	exportRankSynthetic  = 2
	exportRankExplicit   = 3
	exportRankConflict   = 4
)

func exportMetaRank(meta *ExportMeta) int {
	switch {
	case meta == nil:
		return exportRankMissing
	case meta.unresolved:
		return exportRankUnresolved
	case meta.synthetic:
		return exportRankSynthetic
	case meta.conflict:
		return exportRankConflict
	default:
		return exportRankExplicit
	}
}

// ExportMap records the statically visible exports of an ES module. It does
// not enumerate synthetic defaults from CommonJS interop; HasDefault
// includes them when checking whether a default import has a value.
//
// A map handed to a rule is read-only and can share its underlying data with
// other importers, which may be linted concurrently. That is why
// nothing here that writes is exported — only the builder in this package, which
// owns a map until it publishes it, may fill one in.
type ExportMap struct {
	exports         map[string]*ExportMeta
	names           []string
	hasUnknown      bool
	reexports       map[string]string
	implicitDefault bool
	docFile         *ast.SourceFile
	defaultDoc      *exportDocSource
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

// HasDefault checks that the default export resolves to a value, including
// TypeScript interop defaults. Merely declaring a missing re-export does not
// count, unlike Has, which keeps those names for export validation.
func (m *ExportMap) HasDefault() bool {
	if m == nil {
		return false
	}
	meta := m.Get(defaultExportName)
	return m.implicitDefault || (meta != nil && !meta.unresolved)
}

// ReexportSource returns the resolved file of an explicit `export { … } from`.
// Star exports and local exports have no direct re-export source.
func (m *ExportMap) ReexportSource(name string) string {
	if m == nil {
		return ""
	}
	return m.reexports[name]
}

// Names enumerates statically known export names, without exposing mutable
// metadata or inventing names for unresolved star exports. Names retain their
// discovery order so diagnostics from re-exports are deterministic.
func (m *ExportMap) Names() iter.Seq[string] {
	if m == nil {
		return slices.Values([]string(nil))
	}
	return slices.Values(m.names)
}

func (m *ExportMap) set(name string, meta *ExportMeta) {
	if m == nil {
		return
	}
	if meta == nil {
		meta = &ExportMeta{}
	}
	if _, exists := m.exports[name]; !exists {
		m.names = append(m.names, name)
	}
	m.exports[name] = meta
}

func (m *ExportMap) addUnknown() {
	if m != nil {
		m.hasUnknown = true
	}
}

type originResolution uint8

const (
	originNotFound originResolution = iota
	originFound
	originAmbiguous
)

type directOriginTarget struct {
	binding  exportBinding
	terminal bool
}

type localOriginIndex struct {
	direct      map[string]directOriginTarget
	starTargets []*ast.SourceFile
}

func (builder *exportBuilder) localOriginIndexOf(file *ast.SourceFile, local *localExports) *localOriginIndex {
	if builder.originIndex != nil {
		if cached, ok := builder.originIndex[local]; ok {
			return cached
		}
	}
	directCap := 0
	for _, step := range local.Steps {
		switch step.Kind {
		case exportStepNames:
			directCap += len(step.Names)
		case exportStepLocalDefault:
			directCap++
		case exportStepNamed:
			directCap += len(step.Specs)
		}
	}
	idx := &localOriginIndex{}
	if directCap > 0 {
		idx.direct = make(map[string]directOriginTarget, directCap)
	}
	for _, step := range local.Steps {
		builder.originStepScans++
		switch step.Kind {
		case exportStepNames:
			kind := exportBindingLocal
			if step.NamespaceExport {
				kind = exportBindingExternal
			}
			for _, name := range step.Names {
				originName := name
				if name == defaultExportName && step.Local != "" {
					originName = step.Local
				}
				idx.direct[name] = directOriginTarget{
					binding:  exportBinding{file: file, name: originName, kind: kind},
					terminal: true,
				}
			}
		case exportStepLocalDefault:
			idx.direct[defaultExportName] = directOriginTarget{
				binding:  exportBinding{file: file, name: defaultExportName},
				terminal: true,
			}
		case exportStepNamed:
			for _, spec := range step.Specs {
				if step.FromModule {
					if !spec.LocalOK {
						continue
					}
					if !step.Link.Resolved || step.Link.View == moduleViewDefaultOnly || step.Link.Target == nil {
						idx.direct[spec.Exported] = directOriginTarget{
							binding:  exportBinding{file: file, name: spec.Exported, kind: exportBindingExternal},
							terminal: true,
						}
					} else {
						idx.direct[spec.Exported] = directOriginTarget{
							binding: exportBinding{file: step.Link.Target, name: spec.Local},
						}
					}
					continue
				}
				if spec.LocalIdent != "" {
					if imp, ok := local.ValueImports[spec.LocalIdent]; ok && imp.Link.Resolved && imp.Link.View != moduleViewDefaultOnly && imp.Link.Target != nil {
						idx.direct[spec.Exported] = directOriginTarget{
							binding: exportBinding{file: imp.Link.Target, name: imp.ImportedName},
						}
					} else {
						idx.direct[spec.Exported] = directOriginTarget{
							binding:  exportBinding{file: file, name: spec.LocalIdent},
							terminal: true,
						}
					}
				} else {
					idx.direct[spec.Exported] = directOriginTarget{
						binding:  exportBinding{file: file, name: spec.Exported},
						terminal: true,
					}
				}
			}
		case exportStepStar:
			if step.Link.Resolved && step.Link.View != moduleViewDefaultOnly && step.Link.Target != nil {
				idx.starTargets = append(idx.starTargets, step.Link.Target)
			}
		}
	}
	if builder.originIndex == nil {
		builder.originIndex = make(map[*localExports]*localOriginIndex)
	}
	builder.originIndex[local] = idx
	return idx
}

func (builder *exportBuilder) resolveBindingOriginTop(binding exportBinding) (exportBinding, originResolution) {
	if binding.file == nil {
		return exportBinding{}, originNotFound
	}
	local := builder.index.localExportsOf(builder.program(), binding.file)
	if local == nil {
		return exportBinding{}, originNotFound
	}
	idx := builder.localOriginIndexOf(binding.file, local)
	if target, ok := idx.direct[binding.name]; ok && target.terminal {
		return target.binding, originFound
	}
	if builder.originVisited == nil {
		builder.originVisited = make(map[exportBinding]bool)
	} else {
		clear(builder.originVisited)
	}
	return builder.resolveBindingOrigin(binding, builder.originVisited)
}

func (builder *exportBuilder) resolveBindingOrigin(
	binding exportBinding,
	visited map[exportBinding]bool,
) (exportBinding, originResolution) {
	if binding.file == nil {
		return exportBinding{}, originNotFound
	}
	if visited[binding] {
		return exportBinding{}, originNotFound
	}
	visited[binding] = true

	local := builder.index.localExportsOf(builder.program(), binding.file)
	if local == nil {
		return exportBinding{}, originNotFound
	}
	idx := builder.localOriginIndexOf(binding.file, local)
	if target, ok := idx.direct[binding.name]; ok {
		if target.terminal {
			return target.binding, originFound
		}
		return builder.resolveBindingOrigin(target.binding, visited)
	}
	if binding.name == defaultExportName {
		if local.ImplicitDefault {
			return exportBinding{file: binding.file, name: defaultExportName}, originFound
		}
		return exportBinding{}, originNotFound
	}
	var resolved exportBinding
	found := false
	for _, targetFile := range idx.starTargets {
		candidate, state := builder.resolveBindingOrigin(exportBinding{file: targetFile, name: binding.name}, visited)
		switch state {
		case originAmbiguous:
			return exportBinding{}, originAmbiguous
		case originFound:
			if !found {
				resolved = candidate
				found = true
			} else if resolved != candidate {
				return exportBinding{}, originAmbiguous
			}
		}
	}
	if found {
		return resolved, originFound
	}
	return exportBinding{}, originNotFound
}

func (builder *exportBuilder) sameExportBinding(a, b *ExportMeta) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil || a.originFile == nil || b.originFile == nil {
		return false
	}
	oa := a.originBinding()
	ob := b.originBinding()
	if a.originLocal && b.originLocal {
		return oa == ob
	}
	if !a.originLocal {
		var state originResolution
		oa, state = builder.resolveBindingOriginTop(oa)
		if state != originFound {
			return false
		}
	}
	if !b.originLocal {
		var state originResolution
		ob, state = builder.resolveBindingOriginTop(ob)
		if state != originFound {
			return false
		}
	}
	return oa == ob
}

var conflictExportMeta = &ExportMeta{conflict: true}

// mergeStarMeta computes the monotone join of cur and incoming along an
// `export *` edge. Same-binding paths preserve Namespace and documentation
// metadata regardless of star order; distinct populated bindings (whether
// explicit or synthetic) collide to conflictExportMeta (Namespace: nil),
// matching ECMAScript's ambiguous star export omission without dropping the
// name from Has()/Names().
func (builder *exportBuilder) mergeStarMeta(cur, incoming *ExportMeta) (*ExportMeta, bool) {
	if incoming == nil || cur == incoming {
		return cur, false
	}
	curRank := exportMetaRank(cur)
	inRank := exportMetaRank(incoming)
	if curRank == exportRankConflict {
		return cur, false
	}
	if inRank == exportRankConflict {
		return incoming, true
	}
	if curRank < exportRankSynthetic || inRank < exportRankSynthetic {
		if inRank > curRank {
			return incoming, true
		}
		return cur, false
	}
	if !builder.sameExportBinding(cur, incoming) {
		return conflictExportMeta, true
	}
	if inRank > curRank {
		return incoming, true
	}
	if inRank == curRank && curRank == exportRankExplicit {
		needNamespace := cur.Namespace == nil && incoming.Namespace != nil
		needDoc := cur.doc == nil && incoming.doc != nil
		switch {
		case needNamespace && needDoc:
			return incoming, true
		case needNamespace:
			if cur.doc == nil || incoming.doc != nil {
				return incoming, true
			}
			merged := *incoming
			merged.doc = cur.doc
			return &merged, true
		case needDoc:
			if cur.Namespace == nil {
				return incoming, true
			}
			merged := *cur
			merged.doc = incoming.doc
			return &merged, true
		}
	}
	return cur, false
}

func shouldUpgradeNamed(cur, next *ExportMeta) bool {
	if next == nil || cur == next {
		return false
	}
	curRank := exportMetaRank(cur)
	nextRank := exportMetaRank(next)
	if nextRank != curRank {
		return nextRank > curRank
	}
	return curRank == exportRankExplicit && cur.Namespace == nil && next.Namespace != nil
}

func (builder *exportBuilder) mergeStarFrom(m *ExportMap, other *ExportMap, skip map[string]bool) {
	if m == nil || other == nil {
		return
	}
	for name := range other.Names() {
		if name == defaultExportName {
			continue
		}
		if skip[name] {
			continue
		}
		if next, changed := builder.mergeStarMeta(m.exports[name], other.exports[name]); changed {
			m.set(name, next)
		}
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
// GetExportMap returns the effective export view of one authored
// module request, including loader semantics selected by import attributes.
func GetExportMap(ctx rule.RuleContext, source modules.Source) (*ExportMap, bool) {
	if !ctx.Program().IsValid() || ctx.SourceFile == nil {
		return nil, false
	}
	return getExportMap(ctx.SourceFile, source, newExportBuilder(IndexFor(ctx), ctx.Program()))
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
	// sourceProgram is the source environment every file this builder reaches
	// belongs to. It is
	// held here, on a value that lasts one query, rather than on the index,
	// which can last as long as the Program cache entry and must not keep a
	// Program alive.
	sourceProgram *program.Program
	building      map[*ast.SourceFile]*ExportMap
	seen          map[exportKey]bool
	// visitIndex maps every file currently on componentStack to its 1-based
	// Tarjan discovery index. A file remains on componentStack while it may
	// depend on an unfinished ancestor; once the component root finishes,
	// convergeSCC propagates missing cyclic re-exports in-place across all
	// members' existing ExportMaps without evicting or rebuilding them.
	visitIndex     map[*ast.SourceFile]int
	componentStack []*ast.SourceFile
	nextIndex      int
	// lowLink tracks the smallest visitIndex reachable from the file currently
	// being built. It is saved and restored around each nested build.
	lowLink int
	// unstable holds the files this query finished but found to be in or
	// below a cycle, so that reusing one keeps its dependents unstable too.
	unstable map[*ast.SourceFile]bool
	// buildCount records how many files actually ran exportMapOf in this query.
	buildCount int
	// namedSpecifierChecks counts named re-export specifiers evaluated across
	// applyStep and convergeSCC during this query.
	namedSpecifierChecks int
	// sccQueueEvents counts (member, name) propagation events pushed to the
	// convergeSCC work queue during this query.
	sccQueueEvents int
	// originIndex lazily caches per-localExports direct/star lookup tables for
	// resolveBindingOrigin so wide barrels resolve binding origins in O(1).
	originIndex map[*localExports]*localOriginIndex
	// originVisited is reused across multi-hop resolveBindingOrigin calls.
	originVisited map[exportBinding]bool
	// originStepScans counts local.Steps entries indexed for origin lookups.
	originStepScans int
	// defaultImport selects eslint-plugin-import/default's text filtering and
	// explicit interop behavior without changing the authored export index.
	defaultImport bool
	// sawCycle reports whether the file currently being built has reached a
	// cycle so far. It is saved and restored around each nested build.
	sawCycle bool
}

func newExportBuilder(index *ModuleIndex, sourceProgram *program.Program) *exportBuilder {
	return &exportBuilder{
		index:         index,
		sourceProgram: sourceProgram,
	}
}

func (builder *exportBuilder) program() *program.Program {
	if builder == nil {
		return nil
	}
	return builder.sourceProgram
}

func getExportMap(origin *ast.SourceFile, source modules.Source, builder *exportBuilder) (*ExportMap, bool) {
	sourceProgram := builder.program()
	if !sourceProgram.IsValid() || origin == nil || source.Specifier() == nil || !ast.IsStringLiteralLike(source.Specifier()) {
		return nil, false
	}

	link := resolveExportLink(sourceProgram, origin, builder.index.settings, source)
	if !link.Resolved {
		return nil, false
	}
	if link.View == moduleViewDefaultOnly {
		return defaultOnlyExportMap(), true
	}
	exports := builder.exportMapOf(link.Target)
	if link.NodeDefault && !exports.implicitDefault {
		// The cached map describes the module. This edge's default must not
		// leak into a CommonJS consumer querying the same module later.
		view := *exports
		view.implicitDefault = true
		return &view, true
	}
	return exports, true
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
		if idx := builder.visitIndex[sourceFile]; idx > 0 {
			builder.sawCycle = true
			if idx < builder.lowLink {
				builder.lowLink = idx
			}
		} else if builder.unstable[sourceFile] {
			builder.sawCycle = true
		}
		return existing
	}
	if cached := builder.index.cachedExportMap(sourceFile); cached != nil {
		return cached
	}
	// Cached queries need no traversal state.
	if builder.building == nil {
		builder.building = make(map[*ast.SourceFile]*ExportMap)
		builder.visitIndex = make(map[*ast.SourceFile]int)
		builder.unstable = make(map[*ast.SourceFile]bool)
	}

	builder.nextIndex++
	builder.buildCount++
	myIndex := builder.nextIndex
	prevLowLink := builder.lowLink
	prevSawCycle := builder.sawCycle
	stackPos := len(builder.componentStack)

	exports := newExportMap()
	builder.building[sourceFile] = exports
	builder.visitIndex[sourceFile] = myIndex
	builder.componentStack = append(builder.componentStack, sourceFile)
	builder.lowLink = myIndex
	builder.sawCycle = false

	local := builder.index.localExportsOf(builder.program(), sourceFile)
	exports.implicitDefault = local.ImplicitDefault
	exports.docFile = sourceFile
	exports.defaultDoc = local.DefaultDoc
	for _, step := range local.Steps {
		builder.applyStep(exports, local, step)
	}

	childLowLink := builder.lowLink
	childSawCycle := builder.sawCycle
	if childLowLink < myIndex {
		// This module depended on an unfinished ancestor still on
		// componentStack. Keep it in building so all references share one
		// ExportMap pointer; the component root will converge the SCC in-place.
		if prevLowLink > 0 {
			builder.lowLink = min(prevLowLink, childLowLink)
		} else {
			builder.lowLink = 0
		}
	} else {
		// Component root: finalize this strongly connected component. All
		// members stay in building (never evicted or rebuilt), and any cycle
		// converges their ExportMaps in-place so both sibling lookups and live
		// namespace references observe the complete exports.
		scc := builder.componentStack[stackPos:]
		for _, member := range scc {
			delete(builder.visitIndex, member)
			if childSawCycle {
				builder.unstable[member] = true
			}
		}
		if childSawCycle && (len(scc) > 1 || hasSelfExportStep(local, sourceFile)) {
			builder.convergeSCC(scc)
		}
		builder.componentStack = builder.componentStack[:stackPos]
		builder.lowLink = prevLowLink
	}

	if childSawCycle {
		builder.unstable[sourceFile] = true
	} else {
		builder.index.storeExportMap(sourceFile, exports)
	}
	builder.sawCycle = prevSawCycle || childSawCycle
	return exports
}

func hasSelfExportStep(local *localExports, self *ast.SourceFile) bool {
	if local == nil {
		return false
	}
	for _, step := range local.Steps {
		if step.Kind == exportStepNamed && step.FromModule &&
			step.Link.Resolved && step.Link.View != moduleViewDefaultOnly && step.Link.Target == self {
			return true
		}
	}
	return false
}

type explicitOwner struct {
	stepIdx int
	specIdx int
}

type sccNameKey struct {
	member int
	name   string
}

type sccNamedEdge struct {
	target      int
	exported    string
	nodeDefault bool
}

// convergeSCC propagates cyclic `export *` and explicit `export { ... } from`
// edges in-place across the completed strongly connected component until a
// fixed point is reached. Updates flow along a name-keyed work queue (named
// re-export edges indexed by (depIdx, spec.Local) and star edges indexed by
// depIdx) plus a separate hasUnknown queue. Because every transition strictly
// increases a finite monotone lattice (missing < unresolved < synthetic <
// explicit(plain) < explicit(namespace, same binding) < conflict, or
// !hasUnknown -> hasUnknown) and reserves all explicitly exported names
// against star re-exports regardless of statement order, convergence is
// guaranteed in O(edges) work and existing ExportMeta.Namespace pointers stay
// valid.
func (builder *exportBuilder) convergeSCC(scc []*ast.SourceFile) {
	n := len(scc)
	if n == 0 {
		return
	}
	prog := builder.program()
	maps := make([]*ExportMap, n)
	locals := make([]*localExports, n)
	sccIndex := make(map[*ast.SourceFile]int, n)
	for i, sf := range scc {
		maps[i] = builder.building[sf]
		locals[i] = builder.index.localExportsOf(prog, sf)
		sccIndex[sf] = i
	}

	// Track the last explicit (stepIdx, specIdx) that declared each export
	// name per member so:
	//  1. `export *` never writes to any name the module explicitly exports
	//     (local declarations or explicit named re-exports, including
	//     unresolved ones), regardless of whether the explicit export appears
	//     before or after `export *`.
	//  2. An earlier explicit re-export never overwrites a later explicit
	//     export for the same name (even within the same export statement).
	explicitOwners := make([]map[string]explicitOwner, n)
	for i, local := range locals {
		var owners map[string]explicitOwner
		setOwner := func(name string, stepIdx, specIdx int) {
			if owners == nil {
				owners = make(map[string]explicitOwner)
			}
			owners[name] = explicitOwner{stepIdx: stepIdx, specIdx: specIdx}
		}
		for stepIdx, step := range local.Steps {
			switch step.Kind {
			case exportStepNames:
				for _, name := range step.Names {
					setOwner(name, stepIdx, -1)
				}
			case exportStepLocalDefault:
				setOwner(defaultExportName, stepIdx, -1)
			case exportStepNamed:
				for specIdx, spec := range step.Specs {
					if step.FromModule && !spec.LocalOK {
						continue
					}
					setOwner(spec.Exported, stepIdx, specIdx)
				}
			}
		}
		explicitOwners[i] = owners
	}

	// Index intra-SCC propagation edges:
	// - starEdges[depIdx] lists distinct SCC members `i != depIdx` that do
	//   `export * from dep`.
	// - namedEdges[(depIdx, spec.Local)] lists winning-owner explicit
	//   re-exports `export { Local as Exported } from dep`.
	starEdges := make([][]int, n)
	var namedEdges map[sccNameKey][]sccNamedEdge
	for i, local := range locals {
		owners := explicitOwners[i]
		for stepIdx, step := range local.Steps {
			switch step.Kind {
			case exportStepStar:
				if !step.Link.Resolved || step.Link.View == moduleViewDefaultOnly {
					continue
				}
				depIdx, inSCC := sccIndex[step.Link.Target]
				if !inSCC || depIdx == i {
					continue
				}
				// Outer loop iterates i in strictly increasing order, so O(1)
				// tail comparison suffices to deduplicate.
				if targets := starEdges[depIdx]; len(targets) == 0 || targets[len(targets)-1] != i {
					starEdges[depIdx] = append(targets, i)
				}
			case exportStepNamed:
				if !step.FromModule || !step.Link.Resolved || step.Link.View == moduleViewDefaultOnly {
					continue
				}
				depIdx, inSCC := sccIndex[step.Link.Target]
				if !inSCC {
					continue
				}
				for specIdx, spec := range step.Specs {
					if !spec.LocalOK {
						continue
					}
					owner := owners[spec.Exported]
					if owner.stepIdx != stepIdx || owner.specIdx != specIdx {
						continue
					}
					if namedEdges == nil {
						namedEdges = make(map[sccNameKey][]sccNamedEdge)
					}
					key := sccNameKey{member: depIdx, name: spec.Local}
					namedEdges[key] = append(namedEdges[key], sccNamedEdge{
						target:      i,
						exported:    spec.Exported,
						nodeDefault: step.Link.NodeDefault,
					})
				}
			}
		}
	}

	// Propagate hasUnknown along star edges (each member enters the queue at
	// most once).
	var unknownQueue []int
	for i := range n {
		if maps[i] != nil && maps[i].hasUnknown && len(starEdges[i]) > 0 {
			unknownQueue = append(unknownQueue, i)
		}
	}
	for head := 0; head < len(unknownQueue); {
		depIdx := unknownQueue[head]
		head++
		if head == len(unknownQueue) {
			unknownQueue = unknownQueue[:0]
			head = 0
		}
		for _, targetIdx := range starEdges[depIdx] {
			target := maps[targetIdx]
			if target == nil || target.hasUnknown {
				continue
			}
			target.addUnknown()
			if len(starEdges[targetIdx]) > 0 {
				unknownQueue = append(unknownQueue, targetIdx)
			}
		}
	}

	// Propagate changed (memberIdx, name) events in FIFO order.
	var queue []sccNameKey
	var inQueue map[sccNameKey]bool
	enqueue := func(member int, name string, meta *ExportMeta) {
		key := sccNameKey{member: member, name: name}
		hasStar := name != defaultExportName && len(starEdges[member]) > 0
		hasNamed := len(namedEdges[key]) > 0 && meta != nil && (!meta.unresolved || name == defaultExportName)
		if !hasStar && !hasNamed {
			return
		}
		if inQueue == nil {
			inQueue = make(map[sccNameKey]bool)
		}
		if inQueue[key] {
			return
		}
		inQueue[key] = true
		queue = append(queue, key)
		builder.sccQueueEvents++
	}

	namedEdgeWouldUpdate := func(dep *ExportMap, depMeta *ExportMeta, name string, edge sccNamedEdge) bool {
		target := maps[edge.target]
		if target == nil || dep == nil || depMeta == nil {
			return false
		}
		cur := target.exports[edge.exported]
		if name == defaultExportName && (dep.implicitDefault || edge.nodeDefault) && depMeta.unresolved {
			return exportMetaRank(cur) < exportRankSynthetic
		}
		if depMeta.unresolved {
			return false
		}
		return shouldUpgradeNamed(cur, depMeta)
	}

	namedEdgeNeedsUpdate := func(dep *ExportMap, depMeta *ExportMeta, name string, edge sccNamedEdge) (*ExportMeta, bool) {
		if !namedEdgeWouldUpdate(dep, depMeta, name, edge) {
			return nil, false
		}
		if name == defaultExportName && (dep.implicitDefault || edge.nodeDefault) && depMeta.unresolved {
			return &ExportMeta{
				doc:        dep.defaultDoc,
				originFile: dep.docFile,
				originName: defaultExportName,
				synthetic:  true,
			}, true
		}
		return depMeta, true
	}

	for i := range n {
		m := maps[i]
		if m == nil {
			continue
		}
		hasStarEdges := len(starEdges[i]) > 0
		for name := range m.Names() {
			meta := m.exports[name]
			needsEnqueue := false
			if meta != nil && (!meta.unresolved || name == defaultExportName) {
				if edges := namedEdges[sccNameKey{member: i, name: name}]; len(edges) > 0 {
					for _, edge := range edges {
						if namedEdgeWouldUpdate(m, meta, name, edge) {
							needsEnqueue = true
							break
						}
					}
				}
			}
			if !needsEnqueue && hasStarEdges && name != defaultExportName {
				for _, targetIdx := range starEdges[i] {
					if _, isExplicit := explicitOwners[targetIdx][name]; isExplicit {
						continue
					}
					target := maps[targetIdx]
					if target != nil {
						if _, changed := builder.mergeStarMeta(target.exports[name], meta); changed {
							needsEnqueue = true
							break
						}
					}
				}
			}
			if needsEnqueue {
				enqueue(i, name, meta)
			}
		}
	}

	for head := 0; head < len(queue); {
		ev := queue[head]
		queue[head] = sccNameKey{}
		head++
		if head == len(queue) {
			queue = queue[:0]
			head = 0
		} else if head >= 64 && head*2 >= len(queue) {
			nRemain := copy(queue, queue[head:])
			clear(queue[nRemain:])
			queue = queue[:nRemain]
			head = 0
		}
		delete(inQueue, ev)
		dep := maps[ev.member]
		if dep == nil {
			continue
		}
		depMeta := dep.exports[ev.name]

		if ev.name != defaultExportName {
			for _, targetIdx := range starEdges[ev.member] {
				if _, isExplicit := explicitOwners[targetIdx][ev.name]; isExplicit {
					continue
				}
				target := maps[targetIdx]
				if target == nil {
					continue
				}
				if next, changed := builder.mergeStarMeta(target.exports[ev.name], depMeta); changed {
					target.set(ev.name, next)
					enqueue(targetIdx, ev.name, next)
				}
			}
		}

		if depMeta != nil && (!depMeta.unresolved || ev.name == defaultExportName) {
			for _, edge := range namedEdges[ev] {
				builder.namedSpecifierChecks++
				if next, changed := namedEdgeNeedsUpdate(dep, depMeta, ev.name, edge); changed {
					maps[edge.target].set(edge.exported, next)
					enqueue(edge.target, edge.exported, next)
				}
			}
		}
	}
}

func (builder *exportBuilder) applyStep(exports *ExportMap, local *localExports, step exportStep) {
	switch step.Kind {
	case exportStepNames:
		kind := exportBindingLocal
		if step.NamespaceExport {
			kind = exportBindingExternal
		}
		for _, name := range step.Names {
			originName := name
			if name == defaultExportName && step.Local != "" {
				originName = step.Local
			}
			exports.set(name, &ExportMeta{
				doc:         step.Docs[name],
				originFile:  exports.docFile,
				originName:  originName,
				originKind:  kind,
				originLocal: true,
			})
		}

	case exportStepLocalDefault:
		if meta, ok := builder.namespaceImportMeta(local, step.Local); ok {
			meta.doc = step.Docs[defaultExportName]
			meta.originFile = exports.docFile
			meta.originName = defaultExportName
			meta.originLocal = true
			exports.set(defaultExportName, meta)
		} else {
			exports.set(defaultExportName, &ExportMeta{
				doc:         step.Docs[defaultExportName],
				originFile:  exports.docFile,
				originName:  defaultExportName,
				originLocal: true,
			})
		}

	case exportStepStar:
		if !step.Link.Resolved {
			exports.addUnknown()
			return
		}
		builder.mergeStarFrom(exports, builder.exportMapForLink(step.Link), local.ExplicitNames)

	case exportStepNamed:
		var dependency *ExportMap
		if step.FromModule && step.Link.Resolved {
			dependency = builder.exportMapForLink(step.Link)
		}
		for _, spec := range step.Specs {
			if !step.FromModule {
				if meta, ok := builder.namespaceImportMeta(local, spec.LocalIdent); ok {
					meta.originFile = exports.docFile
					meta.originName = spec.LocalIdent
					meta.originLocal = true
					exports.set(spec.Exported, meta)
				} else if imp, ok := local.ValueImports[spec.LocalIdent]; ok && imp.Link.Resolved && imp.Link.View != moduleViewDefaultOnly && imp.Link.Target != nil {
					exports.set(spec.Exported, &ExportMeta{
						originFile: imp.Link.Target,
						originName: imp.ImportedName,
					})
				} else {
					originName := spec.LocalIdent
					if originName == "" {
						originName = spec.Exported
					}
					exports.set(spec.Exported, &ExportMeta{
						originFile:  exports.docFile,
						originName:  originName,
						originLocal: true,
					})
				}
				continue
			}
			if !spec.LocalOK {
				continue
			}
			builder.namedSpecifierChecks++
			if dependency == nil {
				exports.set(spec.Exported, &ExportMeta{unresolved: true})
				continue
			}
			if exports.reexports == nil {
				exports.reexports = make(map[string]string)
			}
			exports.reexports[spec.Exported] = step.Link.Path
			// An explicit re-export declares its public name even when its
			// target is missing. Resolving the target only supplies metadata.
			meta := dependency.Get(spec.Local)
			if spec.Local == defaultExportName && (dependency.implicitDefault || step.Link.NodeDefault) && (meta == nil || meta.unresolved) {
				meta = &ExportMeta{
					doc:        dependency.defaultDoc,
					originFile: dependency.docFile,
					originName: defaultExportName,
					synthetic:  true,
				}
			}
			if meta == nil {
				meta = &ExportMeta{unresolved: true}
				if spec.Local == defaultExportName {
					meta.doc = dependency.defaultDoc
				}
			}
			exports.set(spec.Exported, meta)
		}
	}
}

// namespaceImportMeta finds the `import * as localName` binding the file
// re-exports locally and builds only that target's export map.
func (builder *exportBuilder) namespaceImportMeta(local *localExports, localName string) (*ExportMeta, bool) {
	if localName == "" {
		return nil, false
	}
	for _, binding := range local.Imports {
		if binding.NamespaceName != localName {
			continue
		}
		return &ExportMeta{Namespace: builder.exportMapForLink(binding.Link)}, true
	}
	return nil, false
}

func (builder *exportBuilder) exportMapForLink(link exportLink) *ExportMap {
	if link.View == moduleViewDefaultOnly {
		return defaultOnlyExportMap()
	}
	return builder.exportMapOf(link.Target)
}

func moduleExportName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	return rslint_utils.GetStaticPropertyName(node)
}
