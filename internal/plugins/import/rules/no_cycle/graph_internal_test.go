package no_cycle

import (
	"reflect"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// The rule's fixtures can only describe graphs someone is willing to write out
// as files, which in practice means short chains back to the linted file. The
// component numbering they exercise is a fraction of what the rule depends on,
// so the shapes that decide whether it is right — a component reached but not
// joined, a cross edge into a finished component, components sharing a member —
// are given to computeGroups directly here, as adjacency lists.
//
// Only the partition is contractual. Which number a component gets is an
// artifact of the order the traversal happens to finish in, and nothing reads
// the numbers for anything but equality.

// graphFor builds the part of a moduleGraph that tarjanVisit reads. -1 stands
// for a reference that resolves to nothing the rule follows.
func graphFor(adjacency [][]int32) *moduleGraph {
	count := len(adjacency)
	graph := &moduleGraph{
		nodes:    make([]moduleNode, count),
		group:    make([]int32, count),
		explored: make([]atomic.Bool, count),
		tarjan: tarjanState{
			visitIndex: make([]int32, count),
			lowLink:    make([]int32, count),
			onStack:    make([]bool, count),
			pending:    make([]groupFrame, 0, 32),
			groups:     1,
		},
	}
	for node, edges := range adjacency {
		graph.nodes[node].edge = edges
	}
	return graph
}

func computeGroups(graph *moduleGraph) {
	for root := range len(graph.nodes) {
		if graph.tarjan.visitIndex[root] == 0 {
			graph.tarjanVisit(int32(root), nil)
		}
	}
}

// componentsOf reads a numbering back as the partition it describes: each
// component's members ascending, the components ordered by their first member.
func componentsOf(group []int32) [][]int {
	byNumber := make(map[int32][]int, len(group))
	for node, number := range group {
		byNumber[number] = append(byNumber[number], node)
	}
	components := make([][]int, 0, len(byNumber))
	for _, members := range byNumber {
		components = append(components, members)
	}
	slices.SortFunc(components, func(a, b []int) int { return a[0] - b[0] })
	return components
}

func TestComputeGroups(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		adjacency [][]int32
		want      [][]int
	}{
		{
			name:      "a graph with no files has no components",
			adjacency: [][]int32{},
			want:      [][]int{},
		},
		{
			name:      "files that import nothing each stand alone",
			adjacency: [][]int32{nil, nil, nil},
			want:      [][]int{{0}, {1}, {2}},
		},
		{
			name:      "references the rule does not follow are not edges",
			adjacency: [][]int32{{-1, -1}, {-1}},
			want:      [][]int{{0}, {1}},
		},
		{
			name:      "a file importing itself is still one component",
			adjacency: [][]int32{{0}},
			want:      [][]int{{0}},
		},
		{
			name:      "a chain joins nothing",
			adjacency: [][]int32{{1}, {2}, nil},
			want:      [][]int{{0}, {1}, {2}},
		},
		{
			name:      "two files importing each other share a component",
			adjacency: [][]int32{{1}, {0}},
			want:      [][]int{{0, 1}},
		},
		{
			name:      "a three file cycle shares a component",
			adjacency: [][]int32{{1}, {2}, {0}},
			want:      [][]int{{0, 1, 2}},
		},
		{
			// Two cycles through one shared file are one component, not two:
			// every member reaches every other through it.
			name:      "cycles sharing a file are one component",
			adjacency: [][]int32{{1}, {0, 2}, {1}},
			want:      [][]int{{0, 1, 2}},
		},
		{
			// The regression this shape exists for: 3 is visited after the
			// component {1,2} has already been closed and popped, and its edge
			// into 1 must not pull 3 — and through it the root 0 — into that
			// component. Reaching a component is not joining it, which is the
			// whole premise of the rule's pruning.
			name:      "an edge into a closed component does not join it",
			adjacency: [][]int32{{1, 3}, {2}, {1}, {1}},
			want:      [][]int{{0}, {1, 2}, {3}},
		},
		{
			name:      "components joined by a one way edge stay apart",
			adjacency: [][]int32{{1}, {0, 2}, {3}, {2}},
			want:      [][]int{{0, 1}, {2, 3}},
		},
		{
			// Nothing connects these, so the traversal reaches the second only
			// by starting a new root.
			name:      "components in separate roots are found",
			adjacency: [][]int32{{1}, {0}, {3}, {2}},
			want:      [][]int{{0, 1}, {2, 3}},
		},
		{
			// Tarjan's own worked example, renumbered from zero: a graph with
			// nested back edges, cross edges between components, and a self
			// loop, all at once.
			name: "the classic worked example",
			adjacency: [][]int32{
				{1},
				{2},
				{0},
				{1, 2, 4},
				{3, 5},
				{2, 6},
				{5},
				{4, 6, 7},
			},
			want: [][]int{{0, 1, 2}, {3, 4}, {5, 6}, {7}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			graph := graphFor(tt.adjacency)
			computeGroups(graph)

			if len(graph.group) != len(tt.adjacency) {
				t.Fatalf("numbered %d files, want %d", len(graph.group), len(tt.adjacency))
			}
			got := componentsOf(graph.group)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("components = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComputeGroupsHandlesADeepChain locks in that the depth of an import chain
// is bounded by the heap rather than by the goroutine stack, which is why the
// traversal carries its own stack. The chain is far longer than any real
// project's, and a recursive traversal of it is not something this test can
// prove impossible — it fixes the answer for a shape nobody would write out as
// fixtures.
func TestComputeGroupsHandlesADeepChain(t *testing.T) {
	t.Parallel()

	const files = 100_000
	// One backing array for every edge list, so the chain costs one allocation
	// rather than one per file.
	backing := make([]int32, files)
	adjacency := make([][]int32, files)
	for node := range files - 1 {
		backing[node] = int32(node + 1)
		adjacency[node] = backing[node : node+1]
	}

	graph := graphFor(adjacency)
	computeGroups(graph)

	seen := make(map[int32]bool, files)
	for _, number := range graph.group {
		seen[number] = true
	}
	if len(seen) != files {
		t.Fatalf("a chain produced %d components, want %d", len(seen), files)
	}
}

// TestWithheldDynamicEdges locks in what allowUnsafeDynamicCyclicDependency
// withholds: a target, not a reference. A file that both imports a module
// dynamically and imports it outright stops reaching it either way, which is
// what upstream does and what the whole-file wording of the option means.
func TestWithheldDynamicEdges(t *testing.T) {
	t.Parallel()

	dynamic := program.ModuleReference{Kind: program.ModuleReferenceDynamicImport}
	static := program.ModuleReference{Kind: program.ModuleReferenceImport}

	t.Run("a file with no dynamic import keeps its own edges", func(t *testing.T) {
		t.Parallel()

		node := &moduleNode{refs: []program.ModuleReference{static, static}, edge: []int32{1, 2}}
		expand := withheldDynamicEdges(node)
		if &expand[0] != &node.edge[0] {
			t.Fatal("expected the edge slice itself, not a copy")
		}
	})

	t.Run("a dynamically imported target is withheld", func(t *testing.T) {
		t.Parallel()

		node := &moduleNode{refs: []program.ModuleReference{dynamic, static}, edge: []int32{1, 2}}
		if got := withheldDynamicEdges(node); !reflect.DeepEqual(got, []int32{-1, 2}) {
			t.Fatalf("expand = %v, want [-1 2]", got)
		}
	})

	t.Run("a static reference to a withheld target is withheld too", func(t *testing.T) {
		t.Parallel()

		node := &moduleNode{refs: []program.ModuleReference{static, dynamic, static}, edge: []int32{1, 1, 2}}
		if got := withheldDynamicEdges(node); !reflect.DeepEqual(got, []int32{-1, -1, 2}) {
			t.Fatalf("expand = %v, want [-1 -1 2]", got)
		}
	})

	t.Run("an unfollowed dynamic reference withholds nothing", func(t *testing.T) {
		t.Parallel()

		node := &moduleNode{refs: []program.ModuleReference{dynamic, static}, edge: []int32{-1, 2}}
		if got := withheldDynamicEdges(node); !reflect.DeepEqual(got, []int32{-1, 2}) {
			t.Fatalf("expand = %v, want [-1 2]", got)
		}
	})
}

func TestTarjanVisitIncrementalOrderIndependence(t *testing.T) {
	t.Parallel()

	// Classic Tarjan graph plus a disconnected component {8, 9} that should
	// remain unvisited until explicitly requested.
	adjacency := [][]int32{
		{1},
		{2},
		{0},
		{1, 2, 4},
		{3, 5},
		{2, 6},
		{5},
		{4, 6, 7},
		{9},
		{8},
	}
	wantAll := [][]int{{0, 1, 2}, {3, 4}, {5, 6}, {7}, {8, 9}}

	for name, order := range map[string][]int32{
		"reverse":      {7, 6, 5, 4, 3, 2, 1, 0},
		"middle-first": {3, 7, 0, 5},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			graph := graphFor(adjacency)
			for _, root := range order {
				graph.tarjanVisit(root, nil)
			}
			for node := range 8 {
				if !graph.explored[node].Load() || graph.group[node] == 0 {
					t.Fatalf("reachable node %d was not finalized (explored=%t, group=%d)", node, graph.explored[node].Load(), graph.group[node])
				}
			}
			// Disconnected nodes 8 and 9 must remain unexplored until requested.
			for _, node := range []int32{8, 9} {
				if graph.tarjan.visitIndex[node] != 0 || graph.explored[node].Load() || graph.group[node] != 0 {
					t.Fatalf("disconnected node %d was visited prematurely (visitIndex=%d, explored=%t, group=%d)", node, graph.tarjan.visitIndex[node], graph.explored[node].Load(), graph.group[node])
				}
			}
			graph.tarjanVisit(8, nil)
			for _, node := range []int32{8, 9} {
				if !graph.explored[node].Load() || graph.group[node] == 0 {
					t.Fatalf("node %d was not finalized after visiting root 8 (explored=%t, group=%d)", node, graph.explored[node].Load(), graph.group[node])
				}
			}
			if got := componentsOf(graph.group); !reflect.DeepEqual(got, wantAll) {
				t.Fatalf("incremental components = %v, want %v", got, wantAll)
			}
		})
	}
}

func TestEnsureExploredReleasesTarjanScratchWhenComplete(t *testing.T) {
	t.Parallel()

	const root = "/no-cycle-tarjan-release"
	fileA := tspath.ResolvePath(root, "a.ts")
	fileB := tspath.ResolvePath(root, "b.ts")
	fileC := tspath.ResolvePath(root, "c.ts")
	fileIgnored := tspath.ResolvePath(root, "ignored.ts")

	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		fileA:       `import "./b"; export const a = 1;`,
		fileB:       `import "./a"; export const b = 1;`,
		fileC:       `export const c = 1;`,
		fileIgnored: `export const ignored = 1;`,
	})
	p, err := program.NewFromRoots(program.RootOptions{
		RootFileNames: []string{fileA, fileB, fileC, fileIgnored},
		Host:          utils.CreateCompilerHost(root, fs),
		CompilerOptions: &core.CompilerOptions{
			NoLib:            core.TSTrue,
			Module:           core.ModuleKindESNext,
			ModuleResolution: core.ModuleResolutionKindBundler,
		},
		SingleThreaded: true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}

	settings := import_utils.SettingsFor((rule.RuleContext{
		Settings: map[string]interface{}{
			"import/ignore": []interface{}{`/ignored\.ts$`},
		},
	}).WithProgram(p))
	opts := ruleOptions{
		maxDepth:       unlimitedDepth,
		referenceKinds: program.ESModuleReferences,
	}
	sourceGraph := p.ModuleGraph()
	graph := newModuleGraph(sourceGraph, settings, opts)

	if len(graph.index) != 3 {
		t.Fatalf("included node count = %d, want 3", len(graph.index))
	}
	idxA := graph.index[p.GetSourceFile(fileA)]
	idxB := graph.index[p.GetSourceFile(fileB)]
	idxC := graph.index[p.GetSourceFile(fileC)]

	// Exploring component {a, b} leaves c unexplored, so scratch buffers remain live.
	graph.ensureExplored(sourceGraph, idxA)
	if graph.tarjan.visitIndex == nil {
		t.Fatal("tarjan scratch was released while included node c was still unexplored")
	}
	if !graph.explored[idxA].Load() || !graph.explored[idxB].Load() || graph.explored[idxC].Load() {
		t.Fatalf("unexpected explored flags: a=%t b=%t c=%t", graph.explored[idxA].Load(), graph.explored[idxB].Load(), graph.explored[idxC].Load())
	}

	// Exploring the final included node c releases all 5 Tarjan scratch slices,
	// even though ignored.ts is excluded from graph.index.
	graph.ensureExplored(sourceGraph, idxC)
	if graph.tarjan.visitIndex != nil ||
		graph.tarjan.lowLink != nil ||
		graph.tarjan.onStack != nil ||
		graph.tarjan.component != nil ||
		graph.tarjan.pending != nil {
		t.Fatalf("tarjan scratch slices were not released after exploring all included nodes: %+v", graph.tarjan)
	}
	if graph.group[idxA] == 0 || graph.group[idxA] != graph.group[idxB] || graph.group[idxC] == 0 || graph.group[idxC] == graph.group[idxA] {
		t.Fatalf("unexpected SCC groups: a=%d b=%d c=%d", graph.group[idxA], graph.group[idxB], graph.group[idxC])
	}

	// Subsequent ensureExplored and direct tarjanVisit calls remain safe no-ops.
	for _, idx := range []int32{idxA, idxB, idxC} {
		graph.ensureExplored(sourceGraph, idx)
		graph.tarjanVisit(idx, nil)
	}
}
