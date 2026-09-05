package cfg

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestPathAnalysis(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name          string
		code          string
		event         string
		wantReachable bool
		wantCyclic    bool
		wantEvery     bool
		wantEarly     bool
	}{
		{
			name:          "unconditional after branch",
			code:          `function f(x) { if (x) consume(); markHook(); }`,
			event:         "markHook",
			wantReachable: true,
			wantEvery:     true,
		},
		{
			name:          "conditional branch",
			code:          `function f(x) { if (x) markHook(); }`,
			event:         "markHook",
			wantReachable: true,
			wantEvery:     false,
		},
		{
			name:          "short circuit before early return",
			code:          `function f(x, y) { if (!x || y) return; markHook(); }`,
			event:         "markHook",
			wantReachable: true,
			wantEvery:     false,
			wantEarly:     true,
		},
		{
			name:          "loop condition",
			code:          `function f() { for (; markHook(); ) {} }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    false,
			wantEvery:     true,
		},
		{
			name:          "short circuit inside loop condition",
			code:          `function f(x) { for (; x && markHook(); ) {} }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    true,
			wantEvery:     false,
		},
		{
			name:          "empty for loop header",
			code:          `function f() { for (;;) { markHook(); } }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    false,
			wantEvery:     true,
		},
		{
			name:          "branch inside empty for loop",
			code:          `function f(x) { for (;;) { if (x) markHook(); } }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    true,
			wantEvery:     true,
		},
		{
			name:          "loop body",
			code:          `function f(x) { for (; x; ) { markHook(); } }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    true,
			wantEvery:     false,
		},
		{
			name: "switch break stops outer cycle",
			code: `function f(xs) {
				for (const x of xs) switch (x) {
					case 0: break;
					case 1: markHook(); break;
				}
			}`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    false,
			wantEvery:     false,
			wantEarly:     true,
		},
		{
			name: "switch return keeps outer cycle",
			code: `function f(xs) {
				for (const x of xs) switch (x) {
					case 0: return;
					case 1: markHook(); break;
				}
			}`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    true,
			wantEvery:     false,
			wantEarly:     true,
		},
		{
			name:          "uncaught throw is not a final path",
			code:          `function f(x) { if (x) throw x; markHook(); }`,
			event:         "markHook",
			wantReachable: true,
			wantEvery:     true,
		},
		{
			name:          "unreachable",
			code:          `function f() { return; markHook(); }`,
			event:         "markHook",
			wantReachable: false,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, locations := buildPathTestGraph(t, testCase.code)
			analysis := AnalyzePaths(graph)
			blocks := locations[testCase.event]
			if len(blocks) != 1 {
				t.Fatalf("event %q was laid out in %d blocks, want 1", testCase.event, len(blocks))
			}
			block := blocks[0]
			if got := block.Reachable; got != testCase.wantReachable {
				t.Fatalf("reachable = %v, want %v", got, testCase.wantReachable)
			}
			if !block.Reachable {
				return
			}
			if got := analysis.IsCyclic(block); got != testCase.wantCyclic {
				t.Errorf("cyclic = %v, want %v", got, testCase.wantCyclic)
			}
			if got := analysis.IsOnEveryFinalPath(block); got != testCase.wantEvery {
				t.Errorf("on every final path = %v, want %v", got, testCase.wantEvery)
			}
			shortestFinal, hasFinal := analysis.ShortestExitPathFromStart()
			shortestEvent, reachable := analysis.ShortestPathFromStart(block)
			gotEarly := hasFinal && reachable && shortestFinal < shortestEvent
			if analysis.IsExit(block) {
				gotEarly = hasFinal && reachable && shortestFinal <= shortestEvent
			}
			gotEarly = gotEarly && !analysis.IsOnEveryFinalPath(block)
			if gotEarly != testCase.wantEarly {
				t.Errorf("early final = %v (final=%d, event=%d), want %v", gotEarly, shortestFinal, shortestEvent, testCase.wantEarly)
			}
		})
	}
}

func TestPossibleThrowThroughFinallyIsExcluded(t *testing.T) {
	t.Parallel()
	graph, locations := buildPathTestGraph(t, `
		function f() {
			try {
				mayThrow();
				markInside();
			} finally {
				cleanup();
			}
			markAfter();
		}
	`)
	analysis := AnalyzePaths(graph)
	if !analysis.IsOnEveryFinalPath(locations["markInside"][0]) {
		t.Fatal("the thrown path that bypasses markInside should not count as a final path")
	}
	if !analysis.IsOnEveryFinalPath(locations["markAfter"][0]) {
		t.Fatal("a possible throw through finally should rejoin before markAfter")
	}
}

func TestYieldDoesNotHideLoopBackEdge(t *testing.T) {
	t.Parallel()
	graph, locations := buildPathTestGraph(t, `
		function* f(condition) {
			while (condition) {
				markHook();
				yield condition;
			}
		}
	`)
	block := locations["markHook"][0]
	if analysis := AnalyzePaths(graph); !analysis.IsCyclic(block) {
		t.Fatal("a resumable yield must not hide the loop back edge")
	}
}

func TestFinalPathCountSaturates(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		code string
		want bool
	}{
		{name: "return", code: `function f() { return; }`, want: true},
		{name: "throw", code: `function f() { throw value; }`, want: false},
		{name: "continue plus loop exit", code: `function f(x) { while (x) { continue; } }`, want: true},
		{name: "break plus loop exit", code: `function f(x) { while (x) { break; } }`, want: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, _ := buildPathTestGraph(t, testCase.code)
			if got := AnalyzePaths(graph).HasSingleFinalPath(); got != testCase.want {
				t.Fatalf("single final path = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestPathAnalysisRejectsProvisionallyReachableLayout(t *testing.T) {
	t.Parallel()
	graph, locations := buildPathTestGraph(t, `
		function f(condition) {
			for (; condition; markHook()) {
				return;
			}
		}
	`)
	block := locations["markHook"][0]
	if !block.Reachable {
		t.Fatal("the incrementor should be provisionally reachable while it is laid out")
	}
	if _, reachable := AnalyzePaths(graph).ShortestPathFromStart(block); reachable {
		t.Fatal("the incrementor should have no actual path from the graph entry")
	}
}

func buildPathTestGraph(t *testing.T, code string) (*Graph[string], map[string][]*Block[string]) {
	t.Helper()
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, code, core.ScriptKindTS)
	var root *ast.Node
	source.AsNode().ForEachChild(func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			root = node
			return true
		}
		return false
	})
	if root == nil {
		t.Fatal("test source has no function root")
	}

	graph := Build(root, Hooks[string]{
		Expression: func(builder *Builder[string], node *ast.Node) {
			if node.Kind != ast.KindCallExpression {
				return
			}
			callee := node.AsCallExpression().Expression
			if callee != nil && callee.Kind == ast.KindIdentifier {
				name := callee.AsIdentifier().Text
				if len(name) >= 4 && name[:4] == "mark" {
					builder.Emit(name)
				}
			}
		},
	})
	locations := make(map[string][]*Block[string])
	for _, block := range graph.Blocks {
		for _, event := range block.Events {
			locations[event] = append(locations[event], block)
		}
	}
	return graph, locations
}
