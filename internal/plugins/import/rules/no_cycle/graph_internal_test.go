package no_cycle

import (
	"reflect"
	"slices"
	"testing"

	"github.com/web-infra-dev/rslint/internal/program"
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

// graphFor builds the part of a moduleGraph that computeGroups reads. -1 stands
// for a reference that resolves to nothing the rule follows.
func graphFor(adjacency [][]int32) *moduleGraph {
	graph := &moduleGraph{nodes: make([]moduleNode, len(adjacency))}
	for node, edges := range adjacency {
		graph.nodes[node].edge = edges
	}
	return graph
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
			graph.computeGroups()

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
	graph.computeGroups()

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
