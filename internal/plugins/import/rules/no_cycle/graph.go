package no_cycle

import (
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// This file owns the rule's dependency graph: how one is built from the
// effective source set's module references, and every question the rule asks of it —
// including the search, which is a property of the graph rather than of the
// file that started it. no_cycle.go holds what is left: option parsing, the
// per-file entry point, and the message.

// graphKey identifies one shape of the dependency graph. The references
// themselves come from Program's generic module graph, so only what turns a
// reference into a graph edge belongs here. For searches using this graph,
// maxDepth bounds the search rather than changing its edges.
type graphKey struct {
	settings           string
	referenceKinds     program.ModuleReferenceKinds
	ignoreExternal     bool
	allowUnsafeDynamic bool
}

// moduleNode holds one file's references together with the edges they resolve
// to. Both are derived from the file's syntax, so the whole node is a property
// of the Program generation and is computed once per configuration-complete
// cache key.
type moduleNode struct {
	refs []program.ModuleReference
	// edge[i] is the node that reference i points at, or -1 when the
	// reference is type-only, unresolved, or excluded by ignoreExternal.
	edge []int32
	// expand[i] is edge[i] with the targets allowUnsafeDynamicCyclicDependency
	// withholds cleared. It is the edge set detectCycle walks; when the option
	// is off it is edge itself.
	expand []int32
}

type moduleGraph struct {
	nodes []moduleNode
	index map[*ast.SourceFile]int32
	// group[i] is i's strongly connected component over edge — over the whole
	// edge set, never over the smaller expand. A reference can only close a
	// cycle when its target reaches the importer again, which puts the two in
	// one group, so a file whose every target sits in another group needs no
	// search at all.
	//
	// Computing the groups over the larger of the two edge sets is what lets
	// detectCycle confine itself to one group under either configuration: a
	// route walked over expand is also a route over edge, so its every file
	// both reaches the importer and is reached from it, which is what sharing
	// its group means. A group can therefore hold files no expand route
	// reaches, but never miss one that a route does.
	group []int32

	// explored[i] becomes true only after node i and every node reachable from
	// it over edge have had their nodes[v] and group[v] finalized. Once
	// explored[i].Load() observes true, callers inspect i's reachable subgraph
	// without acquiring mu.
	explored []atomic.Bool

	referenceKinds     program.ModuleReferenceKinds
	allowUnsafeDynamic bool

	mu     sync.Mutex
	tarjan tarjanState
}

type tarjanState struct {
	visitIndex []int32
	lowLink    []int32
	onStack    []bool
	component  []int32
	pending    []groupFrame
	visited    int32
	groups     int32
}

type directTargetsKey struct {
	file  *ast.SourceFile
	kinds program.ModuleReferenceKinds
}

type directCyclesKey struct {
	file           *ast.SourceFile
	kinds          program.ModuleReferenceKinds
	settings       string
	ignoreExternal bool
}

type directCycle struct {
	reference   program.ModuleReference
	dynamicBack bool
}

// directCyclesFor answers maxDepth: 1 without constructing the whole graph.
// At that depth the search only visits the initial target, so a report needs
// a direct edge back to self, which also proves the two share a component.
// Cache both positive and empty answers to avoid rescanning a high-degree
// file on later lint passes. Keep source order and duplicates for reporting.
func directCyclesFor(ctx rule.RuleContext, sourceGraph program.ModuleGraph, settings *import_utils.ModuleSettings, opts ruleOptions, refs []program.ModuleReference) []directCycle {
	self := ctx.SourceFile
	key := directCyclesKey{file: self, kinds: opts.referenceKinds, settings: settings.Key(), ignoreExternal: opts.ignoreExternal}
	return rule.CachedByProgram(ctx, key, func() []directCycle {
		if fileIsExcluded(settings, opts, self) {
			return nil
		}
		var cycles []directCycle
		type backEdge struct{ dynamic, found bool }
		checked := make(map[*ast.SourceFile]backEdge)
		for i := range refs {
			ref := &refs[i]
			if ref.TypeOnly || ref.Target == nil || ref.Target == self || !import_utils.HasAuthoredModuleView(ref.Source()) {
				continue
			}
			back, known := checked[ref.Target]
			if !known {
				// An excluded target cannot close a cycle. Avoid collecting its
				// references, especially for large external export files.
				if !fileIsExcluded(settings, opts, ref.Target) {
					targets := directTargetsFor(ctx, sourceGraph, ref.Target, opts.referenceKinds)
					back.dynamic, back.found = targets[self]
				}
				checked[ref.Target] = back
			}
			if back.found {
				cycles = append(cycles, directCycle{reference: *ref, dynamicBack: back.dynamic})
			}
		}
		return cycles
	})
}

// directTargetsFor indexes a file only when a depth-one search reaches it.
// Sharing this set avoids repeatedly scanning a high-degree target for each
// of its importers. The value records whether any runtime reference to a
// target is dynamic: the unsafe option withholds that target's static edges
// too. This set does not apply settings or unsafe, so its key only needs the
// selected syntax kinds. Neither cache retains a Program through its value.
func directTargetsFor(ctx rule.RuleContext, sourceGraph program.ModuleGraph, file *ast.SourceFile, kinds program.ModuleReferenceKinds) map[*ast.SourceFile]bool {
	return rule.CachedByProgram(ctx, directTargetsKey{file: file, kinds: kinds}, func() map[*ast.SourceFile]bool {
		var targets map[*ast.SourceFile]bool
		refs := sourceGraph.References(file, kinds)
		for i := range refs {
			ref := &refs[i]
			if ref.TypeOnly || ref.Target == nil || !import_utils.HasAuthoredModuleView(ref.Source()) {
				continue
			}
			if targets == nil {
				targets = make(map[*ast.SourceFile]bool)
			}
			targets[ref.Target] = targets[ref.Target] || ref.Dynamic()
		}
		return targets
	})
}

// moduleGraphFor returns the Program generation's dependency graph for these
// options, initializing its file index on the first file that asks for it.
// Individual nodes and their strongly connected components are populated
// on demand by ensureExplored.
func moduleGraphFor(ctx rule.RuleContext, sourceGraph program.ModuleGraph, settings *import_utils.ModuleSettings, opts ruleOptions) *moduleGraph {
	key := graphKey{
		settings:           settings.Key(),
		referenceKinds:     opts.referenceKinds,
		ignoreExternal:     opts.ignoreExternal,
		allowUnsafeDynamic: opts.allowUnsafeDynamicCyclicDependency,
	}
	return rule.CachedByProgram(ctx, key, func() *moduleGraph {
		return newModuleGraph(sourceGraph, settings, opts)
	})
}

func newModuleGraph(sourceGraph program.ModuleGraph, settings *import_utils.ModuleSettings, opts ruleOptions) *moduleGraph {
	files := sourceGraph.Files()
	count := len(files)
	graph := &moduleGraph{
		nodes:              make([]moduleNode, count),
		index:              make(map[*ast.SourceFile]int32, count),
		group:              make([]int32, count),
		explored:           make([]atomic.Bool, count),
		referenceKinds:     opts.referenceKinds,
		allowUnsafeDynamic: opts.allowUnsafeDynamicCyclicDependency,
		tarjan: tarjanState{
			visitIndex: make([]int32, count),
			lowLink:    make([]int32, count),
			onStack:    make([]bool, count),
			pending:    make([]groupFrame, 0, 32),
			groups:     1,
		},
	}
	for i, file := range files {
		if fileIsExcluded(settings, opts, file) {
			continue
		}
		graph.index[file] = int32(i)
	}
	return graph
}

// ensureExplored populates the outgoing edges and finalizes the strongly
// connected components for self and every file reachable from it over edge.
// Because Tarjan closes every SCC completely before returning from its root,
// exploring only reachable roots yields the exact same component partition for
// all visited nodes while avoiding collecting or resolving unreachable files
// (such as default libraries or unreferenced type declarations).
func (graph *moduleGraph) ensureExplored(sourceGraph program.ModuleGraph, self int32) {
	if graph.explored[self].Load() {
		return
	}
	graph.mu.Lock()
	defer graph.mu.Unlock()
	if graph.explored[self].Load() {
		return
	}
	files := sourceGraph.Files()
	graph.tarjanVisit(self, func(nodeIdx int32) {
		graph.populateNode(sourceGraph, nodeIdx, files[nodeIdx])
	})
	if graph.tarjan.visited == int32(len(graph.index)) {
		graph.tarjan = tarjanState{}
	}
}

func (graph *moduleGraph) populateNode(sourceGraph program.ModuleGraph, nodeIdx int32, file *ast.SourceFile) {
	refs := sourceGraph.References(file, graph.referenceKinds)
	if len(refs) == 0 {
		return
	}
	var edge []int32
	for r := range refs {
		if refs[r].TypeOnly || refs[r].Target == nil || !import_utils.HasAuthoredModuleView(refs[r].Source()) {
			continue
		}
		target, ok := graph.index[refs[r].Target]
		if !ok {
			continue
		}
		if edge == nil {
			edge = make([]int32, len(refs))
			for j := range edge {
				edge[j] = -1
			}
		}
		edge[r] = target
	}
	if edge == nil {
		return
	}
	node := &graph.nodes[nodeIdx]
	node.refs = refs
	node.edge = edge
	node.expand = edge
	if graph.allowUnsafeDynamic {
		node.expand = withheldDynamicEdges(node)
	}
}

// fileIsExcluded reports whether every edge into file is dropped: its path is
// covered by `import/ignore`, or ignoreExternal is set and the path is external.
// A resolved reference's Path is its target's FileName, which tsgo requires to
// be absolute and nonempty. IsExternalPath only consults the specifier for an
// empty path, so every reference resolving to this file has the same answer.
func fileIsExcluded(settings *import_utils.ModuleSettings, opts ruleOptions, file *ast.SourceFile) bool {
	fileName := file.FileName()
	if settings.IsIgnoredPath(fileName) {
		return true
	}
	return opts.ignoreExternal && settings.IsExternalPath("", fileName)
}

// withheldDynamicEdges applies allowUnsafeDynamicCyclicDependency: a file's
// dynamically imported targets are not followed out of that file, whichever
// reference reaches them.
func withheldDynamicEdges(node *moduleNode) []int32 {
	var withheld map[int32]bool
	for r, target := range node.edge {
		if target < 0 || !node.refs[r].Dynamic() {
			continue
		}
		if withheld == nil {
			withheld = make(map[int32]bool)
		}
		withheld[target] = true
	}
	if withheld == nil {
		return node.edge
	}

	expand := make([]int32, len(node.edge))
	copy(expand, node.edge)
	for r, target := range expand {
		if target >= 0 && withheld[target] {
			expand[r] = -1
		}
	}
	return expand
}

// tarjanVisit assigns every file reachable from root its strongly connected
// component over edge, running Tarjan from an explicit stack so that a deep
// import chain cannot exhaust the goroutine stack. Component IDs start at 1 so
// group[v] == 0 unambiguously marks an unexplored node.
func (graph *moduleGraph) tarjanVisit(root int32, prepareNode func(node int32)) {
	if graph.tarjan.visitIndex == nil {
		return
	}
	t := &graph.tarjan
	if t.visitIndex[root] != 0 {
		return
	}
	t.pending = append(t.pending, groupFrame{node: root})

	for len(t.pending) > 0 {
		frame := &t.pending[len(t.pending)-1]
		node := frame.node
		if frame.edge == 0 {
			if prepareNode != nil {
				prepareNode(node)
			}
			t.visited++
			t.visitIndex[node] = t.visited
			t.lowLink[node] = t.visited
			t.component = append(t.component, node)
			t.onStack[node] = true
		}

		descended := false
		edges := graph.nodes[node].edge
		for int(frame.edge) < len(edges) {
			target := edges[frame.edge]
			frame.edge++
			if target < 0 {
				continue
			}
			if t.visitIndex[target] == 0 {
				t.pending = append(t.pending, groupFrame{node: target})
				descended = true
				break
			}
			if t.onStack[target] && t.visitIndex[target] < t.lowLink[node] {
				t.lowLink[node] = t.visitIndex[target]
			}
		}
		if descended {
			continue
		}

		if t.lowLink[node] == t.visitIndex[node] {
			start := len(t.component) - 1
			for t.component[start] != node {
				start--
			}
			members := t.component[start:]
			for _, member := range members {
				t.onStack[member] = false
				graph.group[member] = t.groups
			}
			// Every outgoing edge from members targets either another member of
			// this SCC or an SCC that closed strictly earlier, so publishing
			// explored after all group assignments makes the entire reachable
			// subgraph visible to lock-free readers.
			for _, member := range members {
				graph.explored[member].Store(true)
			}
			t.component = t.component[:start]
			t.groups++
		}

		t.pending = t.pending[:len(t.pending)-1]
		if len(t.pending) > 0 {
			parent := t.pending[len(t.pending)-1].node
			if t.lowLink[node] < t.lowLink[parent] {
				t.lowLink[parent] = t.lowLink[node]
			}
		}
	}
}

type groupFrame struct {
	node int32
	edge int32
}

// isReportCandidate reports whether reference r of file self is one a report
// could come from, which is also the only kind the rule searches from. It has
// to be an edge this configuration keeps, it has to point at another file —
// direct self imports belong to import/no-self-import, and a dynamic import is
// not a cycle when the configuration says so — and its target has to share
// self's group, because a reported reference has a route from its target back
// to self and so leaves the two mutually reachable.
func (graph *moduleGraph) isReportCandidate(opts ruleOptions, self int32, r int) bool {
	node := &graph.nodes[self]
	target := node.edge[r]
	if target < 0 || target == self {
		return false
	}
	if opts.allowUnsafeDynamicCyclicDependency && node.refs[r].Dynamic() {
		return false
	}
	return graph.group[target] == graph.group[self]
}

// hasCyclicCandidate reports whether any reference of self could be reported,
// which answers nearly every file in a healthy project from the group numbers
// alone, without walking the graph or allocating a traversal set.
func (graph *moduleGraph) hasCyclicCandidate(opts ruleOptions, self int32) bool {
	for r := range graph.nodes[self].refs {
		if graph.isReportCandidate(opts, self, r) {
			return true
		}
	}
	return false
}

// queuedModule is one breadth-first search entry. It keeps a parent link into
// the queue instead of a copy of the route so far, and names the reference it
// arrived through, so a route is materialized only for a path that closes a
// cycle.
type queuedModule struct {
	node    int32
	parent  int32
	viaNode int32
	viaRef  int32
	depth   int32
}

type routeStep struct {
	value string
	line  int
}

// detectCycle walks breadth-first from start looking for a way back to self
// and returns the route it arrived by. traversed is shared across one file's
// references, so a target an earlier reference already ruled out is not
// searched again — which stays faithful because a search that leaves self's
// group can never come back into it, so the references this file skips as
// non-candidates would not have reached anything these searches visit.
//
// The walk stays inside self's group: every file on a route back to self both
// reaches self and is reached from it, which is what that group means. That
// holds under allowUnsafeDynamicCyclicDependency too, even though the walk
// then follows expand rather than edge, because expand is a subset of the edge
// set the groups were computed over — see moduleGraph.group.
func (graph *moduleGraph) detectCycle(opts ruleOptions, self int32, traversed map[int32]bool, start int32) ([]routeStep, bool) {
	group := graph.group[self]

	queue := []queuedModule{{node: start, parent: -1, viaNode: -1, viaRef: -1}}
	for head := 0; head < len(queue); head++ {
		next := queue[head]
		if traversed[next.node] {
			continue
		}
		traversed[next.node] = true

		for r, target := range graph.nodes[next.node].expand {
			if target < 0 || traversed[target] {
				continue
			}
			if target == self {
				return graph.routeTo(queue, head), true
			}
			if int(next.depth)+1 >= opts.maxDepth {
				continue
			}
			if graph.group[target] != group {
				continue
			}
			queue = append(queue, queuedModule{
				node:    target,
				parent:  int32(head),
				viaNode: next.node,
				viaRef:  int32(r),
				depth:   next.depth + 1,
			})
		}
	}

	return nil, false
}

// routeTo follows the parent links back from a queue entry, materializing the
// specifier and source line of every reference the search passed through.
func (graph *moduleGraph) routeTo(queue []queuedModule, index int) []routeStep {
	depth := int(queue[index].depth)
	route := make([]routeStep, depth)
	for routeIndex := depth - 1; routeIndex >= 0; routeIndex-- {
		next := queue[index]
		edge := &graph.nodes[next.viaNode].refs[next.viaRef]
		route[routeIndex] = routeStep{
			value: edge.Text(),
			line:  sourceLine(edge.From, edge.Specifier),
		}
		index = int(next.parent)
	}
	return route
}

func sourceLine(sourceFile *ast.SourceFile, source *ast.Node) int {
	if sourceFile == nil || source == nil {
		return 1
	}
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, source.Pos())
	return line + 1
}
