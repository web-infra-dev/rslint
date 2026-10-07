package rule

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

type usageOracleID struct {
	Pos   int    `json:"pos"`
	Name  string `json:"name"`
	Inner bool   `json:"inner"`
}

type usageOracleCase struct {
	Name       string            `json:"name"`
	Source     string            `json:"source"`
	TS         bool              `json:"ts"`
	SourceType string            `json:"sourceType"`
	Globals    map[string]string `json:"globals"`
	Probes     []struct {
		Pos   int             `json:"pos"`
		Name  string          `json:"name"`
		Found bool            `json:"found"`
		IDs   []usageOracleID `json:"ids"`
	} `json:"probes"`
}

func usageContext(t testing.TB, source string, ts bool, sourceType string, globals map[string]string) RuleContext {
	t.Helper()
	filename, kind := "/usage.jsx", core.ScriptKindJSX
	if ts {
		filename, kind = "/usage.tsx", core.ScriptKindTSX
	}
	options := LanguageOptions{SourceType: sourceType}
	sf, refs := newBoundRefStoreWithLanguageOptions(t, filename, kind, source, options)
	defaults, _, options := ResolveLanguageDefaults(filename, options)
	comments := NewCommentStore(sf)
	inline, declarations := ParseInlineGlobals(sf, comments)
	configured := make(map[string]utils.GlobalAccess, len(globals))
	for name, access := range globals {
		configured[name], _ = utils.NormalizeGlobalAccess(access)
	}
	return (RuleContext{SourceFile: sf, Refs: refs, LanguageOptions: options, Globals: NewGlobals(options, defaults, configured, inline, declarations)}).WithFileCache(NewFileCache())
}

func usageNodes(sf *ast.SourceFile) []*ast.Node {
	var result []*ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		result = append(result, node)
		node.ForEachChild(visit)
		return false
	}
	visit(sf.AsNode())
	return result
}

func usagePosition(sf *ast.SourceFile, node *ast.Node) int {
	return ecmascript.StringCodeUnitCount(sf.Text()[:utils.TrimNodeTextRange(sf, node).Pos()])
}

// The portable oracle records direct ESLint 10.12.0 calls with Espree and
// @typescript-eslint/parser 8.71.0, not conclusions inferred from references.
// Each probe clears eslintUsed first; cumulative/reversed runs below attack
// the Go cache independently of that upstream observation order.
func TestVariableUsageUpstream(t *testing.T) {
	data, err := os.ReadFile("testdata/variable_usage.json")
	if err != nil {
		t.Fatal(err)
	}
	testVariableUsageOracle(t, data)
}

func testVariableUsageOracle(t *testing.T, data []byte) {
	t.Helper()
	var fixture struct {
		Cases []usageOracleCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) == 0 {
		t.Fatal("empty usage oracle")
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(t *testing.T) {
			ctx := usageContext(t, test.Source, test.TS, test.SourceType, test.Globals)
			graph := scope.Build(ctx.SourceFile, scope.Options{})
			nodes := make(map[int]*ast.Node)
			for _, node := range usageNodes(ctx.SourceFile) {
				if node.Kind == ast.KindCallExpression || node.Kind == ast.KindTypeReference {
					nodes[usagePosition(ctx.SourceFile, node)] = node
				}
			}
			if len(test.Probes) == 0 {
				t.Fatal("case has no probes")
			}
			for _, cumulative := range []bool{false, true} {
				for _, reverse := range []bool{false, true} {
					ctx = ctx.WithFileCache(NewFileCache())
					expected := map[usageOracleID]bool{}
					for step := range test.Probes {
						index := step
						if reverse {
							index = len(test.Probes) - step - 1
						}
						probe := test.Probes[index]
						location := nodes[probe.Pos]
						if location == nil {
							t.Fatalf("no probe node at %d", probe.Pos)
						}
						if !cumulative {
							ctx = ctx.WithFileCache(NewFileCache())
							expected = map[usageOracleID]bool{}
						}
						reader := ctx.VariableUsage()
						if got := ctx.MarkVariableAsUsed(probe.Name, location); got != probe.Found {
							t.Fatalf("cumulative=%v reverse=%v %s@%d found=%v want=%v", cumulative, reverse, probe.Name, probe.Pos, got, probe.Found)
						}
						for _, id := range probe.IDs {
							expected[id] = true
						}
						for range 10 {
							if ctx.MarkVariableAsUsed(probe.Name, location) != probe.Found {
								t.Fatal("cached result changed")
							}
						}
						got := map[usageOracleID]bool{}
						expectedDeclarations := map[*ast.Node]bool{}
						for _, lexical := range graph.Scopes {
							for _, variable := range lexical.Vars {
								if variable.Anonymous || variable.ID == nil || variable.ID.Kind != ast.KindIdentifier {
									continue
								}
								id := usageOracleID{Pos: usagePosition(ctx.SourceFile, variable.ID), Name: variable.Name, Inner: variable.Kind == scope.DefClassInnerName}
								if reader.IsScopeBindingUsed(lexical, variable.Name) {
									got[id] = true
								}
								if expected[id] && (!id.Inner || variable.DefNode.Kind == ast.KindClassExpression) {
									expectedDeclarations[variable.ID] = true
								}
							}
						}
						if !reflect.DeepEqual(got, expected) {
							t.Fatalf("%s@%d marked=%v want=%v", probe.Name, probe.Pos, got, expected)
						}
						for _, node := range usageNodes(ctx.SourceFile) {
							if node.Kind == ast.KindIdentifier && reader.IsDeclarationUsed(node) != expectedDeclarations[node] {
								t.Fatalf("declaration projection at %d", usagePosition(ctx.SourceFile, node))
							}
						}
					}
				}
			}
		})
	}
}

func TestVariableUsageLifecycleAndLazyReads(t *testing.T) {
	ctx := usageContext(t, "const a=1; function f(){ return a; }", false, "module", nil)
	id := ctx.SourceFile.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name()
	reader := ctx.VariableUsage()
	if allocations := testing.AllocsPerRun(100, func() {
		if reader.IsDeclarationUsed(id) || reader.IsGlobalUsed("a") || reader.IsScopeBindingUsed(nil, "a") {
			t.Fatal("empty state has marks")
		}
	}); allocations != 0 {
		t.Fatalf("unused reader allocated %v times", allocations)
	}
	if ctx.fileCache.usage != nil || len(ctx.fileCache.values) != 0 {
		t.Fatal("reading materialized analysis or usage")
	}
	copied := ctx
	if !copied.MarkVariableAsUsed("a", nil) || !reader.IsDeclarationUsed(id) {
		t.Fatal("early reader did not see a later write through another context")
	}
	if ctx.VariableUsage().IsGlobalUsed("a") {
		t.Fatal("module local leaked to globals")
	}
	// A new owner, even with the same SourceFile and tsgo source caches, starts empty.
	next := ctx.WithFileCache(NewFileCache())
	if next.VariableUsage().IsDeclarationUsed(id) {
		t.Fatal("marks survived a new pass")
	}
	foreign := usageContext(t, "const a=2;", false, "module", nil)
	if ctx.MarkVariableAsUsed("a", foreign.SourceFile.AsNode()) || reader.IsDeclarationUsed(foreign.SourceFile.AsNode()) {
		t.Fatal("foreign source accepted")
	}
	if !reader.IsDeclarationUsed(id) || foreign.VariableUsage().IsDeclarationUsed(id) {
		t.Fatal("file isolation broken")
	}
	for _, invalid := range []*RuleContext{nil, {}, {SourceFile: ctx.SourceFile}} {
		if invalid.MarkVariableAsUsed("a", nil) || invalid.VariableUsage().IsDeclarationUsed(id) {
			t.Fatal("invalid context accepted")
		}
	}
	if ctx.MarkVariableAsUsed("", nil) {
		t.Fatal("empty name accepted")
	}
}

func TestVariableUsageReusesAnalysisAcrossOptions(t *testing.T) {
	for _, before := range []bool{false, true} {
		t.Run(fmt.Sprintf("graph-first=%v", before), func(t *testing.T) {
			ctx := usageContext(t, "let a; class C { field = a; }", false, "module", nil)
			analysis := ctx.ScopeAnalysis()
			var filtered *scope.Manager
			if before {
				filtered = analysis.References(map[string]struct{}{"a": {}})
			}
			if !ctx.MarkVariableAsUsed("a", nil) {
				t.Fatal("missing binding")
			}
			if before && analysis.PeekDeclarations() != filtered {
				t.Fatal("rebuilt existing declaration graph")
			}
			complete := analysis.References(nil)
			refs := append([]*scope.Reference(nil), complete.References...)
			reader := ctx.VariableUsage()
			if !reader.IsScopeBindingUsed(complete.Global, "a") {
				t.Fatal("graph variants lost identity")
			}
			if !ctx.MarkVariableAsUsed("a", nil) || !reflect.DeepEqual(refs, complete.References) {
				t.Fatal("usage changed normal references")
			}
			if reader.IsScopeBindingUsed(complete.Global, "C") {
				t.Fatal("unmarked declaration used")
			}
		})
	}
}

func TestVariableUsageGlobalsAndImplicitBindings(t *testing.T) {
	for _, sourceType := range []string{"module", "script", "commonjs"} {
		t.Run(sourceType, func(t *testing.T) {
			ctx := usageContext(t, "/* global a, hidden:off */ let a; function f(){let a;} const arrow=()=>0;", false, sourceType, nil)
			reader := ctx.VariableUsage()
			if !ctx.MarkVariableAsUsed("a", nil) || reader.IsGlobalUsed("a") != (sourceType == "script") {
				t.Fatal("global/local identity mixed")
			}
			if ctx.MarkVariableAsUsed("hidden", nil) {
				t.Fatal("disabled global accepted")
			}
			if !ctx.MarkVariableAsUsed("Math", nil) || !reader.IsGlobalUsed("Math") {
				t.Fatal("missing implicit global")
			}
			if got := ctx.MarkVariableAsUsed("arguments", nil); got != (sourceType == "commonjs") {
				t.Fatalf("wrapper arguments=%v", got)
			}
			for _, node := range usageNodes(ctx.SourceFile) {
				if node.Kind == ast.KindFunctionDeclaration {
					if !ctx.MarkVariableAsUsed("arguments", node) || !reader.IsScopeBindingUsed(ctx.ScopeAnalysis().Declarations().Acquire(node), "arguments") {
						t.Fatal("function arguments not marked")
					}
				}
			}
		})
	}
}

func TestVariableUsageAbsentNameKeepsScopeLazy(t *testing.T) {
	ctx := usageContext(t, "const local=1;", false, "module", nil)
	for range 100 {
		if ctx.MarkVariableAsUsed("absent", nil) {
			t.Fatal("missing name marked")
		}
	}
	if ctx.ScopeAnalysis().PeekDeclarations() != nil {
		t.Fatal("absent name built a scope tree")
	}
	if !ctx.MarkVariableAsUsed("Array", nil) || !ctx.VariableUsage().IsGlobalUsed("Array") {
		t.Fatal("text absence hid a global")
	}
	if ctx.ScopeAnalysis().PeekDeclarations() != nil {
		t.Fatal("absent global name built a scope tree")
	}
}

var variableUsageBenchmarkSink bool

func BenchmarkVariableUsage(b *testing.B) {
	ctx := usageContext(b, "const React={};", false, "module", nil)
	id := ctx.SourceFile.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name()
	b.Run("unused", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			pass := ctx.WithFileCache(NewFileCache())
			variableUsageClosureSink[0] = captureUsageBenchmarkContext(pass)
			variableUsageClosureSink[1] = captureUsageBenchmarkContext(pass)
			variableUsageBenchmarkSink = pass.VariableUsage().IsDeclarationUsed(id)
		}
	})
	b.Run("read", func(b *testing.B) {
		ctx.MarkVariableAsUsed("React", nil)
		usage := ctx.VariableUsage()
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			variableUsageBenchmarkSink = usage.IsDeclarationUsed(id)
		}
	})
	for _, depth := range []int{0, 32} {
		source := "const React={};" + strings.Repeat("{", depth) + "probe();" + strings.Repeat("}", depth)
		ctx := usageContext(b, source, false, "module", nil)
		var at *ast.Node
		for _, node := range usageNodes(ctx.SourceFile) {
			if node.Kind == ast.KindCallExpression {
				at = node
			}
		}
		for _, name := range []string{"React", "missing"} {
			for _, warm := range []bool{false, true} {
				b.Run(fmt.Sprintf("depth=%d/%s/warm=%v", depth, name, warm), func(b *testing.B) {
					pass := ctx.WithFileCache(NewFileCache())
					pass.MarkVariableAsUsed(name, at)
					b.ReportAllocs()
					b.ResetTimer()
					for b.Loop() {
						if !warm {
							pass = ctx.WithFileCache(NewFileCache())
						}
						variableUsageBenchmarkSink = pass.MarkVariableAsUsed(name, at)
					}
				})
			}
		}
	}
}

func TestVariableUsageScopeQueriesDoNotSearchParents(t *testing.T) {
	ctx := usageContext(t, "let x; function f(){ return x; }", false, "module", nil)
	function := ctx.SourceFile.Statements.Nodes[1]
	if !ctx.MarkVariableAsUsed("x", function) {
		t.Fatal("missing outer binding")
	}
	graph := ctx.ScopeAnalysis().Declarations()
	reader := ctx.VariableUsage()
	if reader.IsScopeBindingUsed(graph.Acquire(function), "x") || !reader.IsScopeBindingUsed(graph.Global, "x") {
		t.Fatal("cached parent lookup was mistaken for an inner binding")
	}
	foreign := usageContext(t, "let x;", false, "module", nil)
	if reader.IsScopeBindingUsed(foreign.ScopeAnalysis().Declarations().Global, "x") {
		t.Fatal("foreign graph matched")
	}
	ctx.Refs = nil
	if !ctx.MarkVariableAsUsed("f", nil) {
		t.Fatal("marking required reference analysis")
	}
}

var variableUsageClosureSink [2]func() bool

// cspell:ignore noinline
//
//go:noinline
func captureUsageBenchmarkContext(ctx RuleContext) func() bool {
	return func() bool { return ctx.SourceFile != nil }
}

func BenchmarkVariableUsageManyBindings(b *testing.B) {
	var source strings.Builder
	for i := range 512 {
		fmt.Fprintf(&source, "{ let x%d; probe(); }", i)
	}
	ctx := usageContext(b, source.String(), false, "module", nil)
	var nodes []*ast.Node
	var names []string
	for _, node := range usageNodes(ctx.SourceFile) {
		if node.Kind == ast.KindCallExpression {
			names = append(names, fmt.Sprintf("x%d", len(nodes)))
			nodes = append(nodes, node)
		}
	}
	for _, warm := range []bool{false, true} {
		b.Run(fmt.Sprintf("warm=%v", warm), func(b *testing.B) {
			pass := ctx.WithFileCache(NewFileCache())
			query := func() {
				for i, node := range nodes {
					variableUsageBenchmarkSink = pass.MarkVariableAsUsed(names[i], node)
				}
			}
			query()
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if !warm {
					pass = ctx.WithFileCache(NewFileCache())
				}
				query()
			}
		})
	}
}

func TestVariableUsageParallelPassesShareOnlySourceFacts(t *testing.T) {
	t.Parallel()
	ctx := usageContext(t, "let a; let b;", false, "module", nil)
	for i := range 16 {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			pass := ctx.WithFileCache(NewFileCache())
			name, other := "a", "b"
			if i%2 != 0 {
				name, other = other, name
			}
			reader := pass.VariableUsage()
			if !pass.MarkVariableAsUsed(name, nil) {
				t.Fatal("binding missing")
			}
			global := pass.ScopeAnalysis().Declarations().Global
			if !reader.IsScopeBindingUsed(global, name) || reader.IsScopeBindingUsed(global, other) {
				t.Fatal("independent pass state leaked")
			}
		})
	}
}
