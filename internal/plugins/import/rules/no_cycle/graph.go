package no_cycle

import (
	"github.com/microsoft/typescript-go/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// graphKey identifies one shape of the dependency graph. The references
// themselves come from the shared module index, so only what turns a
// reference into a graph edge belongs here. maxDepth does not: it bounds the
// search, not the graph, so configurations that differ only in maxDepth share
// one graph.
type graphKey struct {
	settings           string
	syntax             rule.ModuleSyntax
	ignoreExternal     bool
	allowUnsafeDynamic bool
}

// moduleNode holds one file's references together with the edges they resolve
// to. Both are derived from the file's syntax, so the whole node is a property
// of the Program and is computed once per lint run.
type moduleNode struct {
	refs []rule.ModuleEdge
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
	// group[i] is i's strongly connected component over edge. A reference can
	// only close a cycle when its target reaches the importer again, which
	// puts the two in one group, so a file whose every target sits in another
	// group needs no search at all.
	group []int32
}

// moduleGraphFor returns the Program's dependency graph for these options,
// building it on the first file of the run that asks for it.
func moduleGraphFor(ctx rule.RuleContext, opts ruleOptions) *moduleGraph {
	settings := import_utils.SettingsFor(ctx)
	key := graphKey{
		settings:           settings.Key(),
		syntax:             opts.syntax,
		ignoreExternal:     opts.ignoreExternal,
		allowUnsafeDynamic: opts.allowUnsafeDynamicCyclicDependency,
	}
	return import_utils.CachedByProgram(ctx.Program, key, func() *moduleGraph {
		return buildModuleGraph(ctx, settings, opts)
	})
}

func buildModuleGraph(ctx rule.RuleContext, settings *import_utils.ModuleSettings, opts ruleOptions) *moduleGraph {
	files := ctx.Modules.Files()
	graph := &moduleGraph{
		nodes: make([]moduleNode, len(files)),
		index: make(map[*ast.SourceFile]int32, len(files)),
	}
	for i, file := range files {
		graph.index[file] = int32(i)
	}

	for i, file := range files {
		// An edge into an excluded file is never traversable, so no search can
		// enter it and no reference of its own can close a cycle back to it.
		// Its outgoing references are therefore never needed, which keeps
		// ignored — and, under ignoreExternal, node_modules — files out of the
		// collection work entirely.
		if fileIsExcluded(settings, opts, file) {
			continue
		}
		refs := ctx.Modules.Edges(file, opts.syntax)
		if len(refs) == 0 {
			continue
		}
		node := &graph.nodes[i]
		node.refs = refs
		node.edge = make([]int32, len(refs))
		for r := range refs {
			node.edge[r] = -1
			if !referenceIsTraversable(settings, opts, refs[r]) {
				continue
			}
			if target, ok := graph.index[refs[r].Target]; ok {
				node.edge[r] = target
			}
		}
		node.expand = node.edge
		if opts.allowUnsafeDynamicCyclicDependency {
			node.expand = withheldDynamicEdges(node)
		}
	}

	graph.computeGroups()
	return graph
}

// fileIsExcluded reports whether referenceIsTraversable drops every edge into
// file: its path is covered by `import/ignore`, or ignoreExternal is set and
// the path is external. IsExternalPath only consults its specifier when the
// resolved path is empty, so passing none matches what it answers for any
// edge resolving to this file.
func fileIsExcluded(settings *import_utils.ModuleSettings, opts ruleOptions, file *ast.SourceFile) bool {
	fileName := file.FileName()
	if settings.IsIgnoredPath(fileName) {
		return true
	}
	return opts.ignoreExternal && settings.IsExternalPath("", fileName)
}

// referenceIsTraversable reports whether an edge is one the rule follows: it
// has to survive into the emitted JavaScript, name a file the Program loaded,
// and be neither ignored by `import/ignore` nor set aside by ignoreExternal.
func referenceIsTraversable(settings *import_utils.ModuleSettings, opts ruleOptions, edge rule.ModuleEdge) bool {
	if edge.TypeOnly || edge.Target == nil {
		return false
	}
	if settings.IsIgnoredPath(edge.Target.FileName()) {
		return false
	}
	return !shouldIgnoreExternal(settings, opts, edge)
}

func shouldIgnoreExternal(settings *import_utils.ModuleSettings, opts ruleOptions, edge rule.ModuleEdge) bool {
	if !opts.ignoreExternal {
		return false
	}
	return settings.IsExternalPath(edge.Text(), moduleReferencePath(edge))
}

func moduleReferencePath(edge rule.ModuleEdge) string {
	if edge.Target != nil {
		return edge.Target.FileName()
	}
	return edge.ResolvedPath
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

// computeGroups assigns every file its strongly connected component over
// edge, running Tarjan from an explicit stack so that a deep import chain
// cannot exhaust the goroutine stack.
func (graph *moduleGraph) computeGroups() {
	count := len(graph.nodes)
	graph.group = make([]int32, count)

	visitIndex := make([]int32, count)
	lowLink := make([]int32, count)
	onStack := make([]bool, count)
	component := make([]int32, 0, count)
	pending := make([]groupFrame, 0, 32)

	var visited, groups int32
	for root := range count {
		if visitIndex[root] != 0 {
			continue
		}
		pending = append(pending, groupFrame{node: int32(root)})

		for len(pending) > 0 {
			frame := &pending[len(pending)-1]
			node := frame.node
			if frame.edge == 0 {
				visited++
				visitIndex[node] = visited
				lowLink[node] = visited
				component = append(component, node)
				onStack[node] = true
			}

			descended := false
			edges := graph.nodes[node].edge
			for int(frame.edge) < len(edges) {
				target := edges[frame.edge]
				frame.edge++
				if target < 0 {
					continue
				}
				if visitIndex[target] == 0 {
					pending = append(pending, groupFrame{node: target})
					descended = true
					break
				}
				if onStack[target] && visitIndex[target] < lowLink[node] {
					lowLink[node] = visitIndex[target]
				}
			}
			if descended {
				continue
			}

			if lowLink[node] == visitIndex[node] {
				for {
					member := component[len(component)-1]
					component = component[:len(component)-1]
					onStack[member] = false
					graph.group[member] = groups
					if member == node {
						break
					}
				}
				groups++
			}

			pending = pending[:len(pending)-1]
			if len(pending) > 0 {
				parent := pending[len(pending)-1].node
				if lowLink[node] < lowLink[parent] {
					lowLink[parent] = lowLink[node]
				}
			}
		}
	}
}

type groupFrame struct {
	node int32
	edge int32
}
