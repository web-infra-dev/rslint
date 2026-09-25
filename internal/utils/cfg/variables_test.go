package cfg

import (
	"maps"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// Compare the dataflow with an independent forward search from each write.
// Random graphs include back edges, self loops, disconnected blocks, and
// repeated sites of the same assignment, as emitted for finally clauses.
func TestVariableLivenessPaths(t *testing.T) {
	random := rand.New(rand.NewPCG(42, 123))
	for _, variables := range []int{1, 2, 63, 64, 65, 127, 128, 129} {
		t.Run(strconv.Itoa(variables), func(t *testing.T) {
			for trial := range 100 {
				source := strings.Repeat("if (c) {}\n", trial%9)
				sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/graph.ts", Path: "/graph.ts"}, source, core.ScriptKindTS)
				// Use the builder to create blocks with their proper graph indices,
				// then replace the edges and events for this dataflow-only test.
				graph := Build(sf.AsNode(), Hooks[variableEvent]{})
				byVariable := make([][]*bool, variables)
				want := map[*bool]bool{new(bool): false} // no reachable site
				for _, blk := range graph.Blocks {
					blk.Reachable = random.IntN(5) != 0
					blk.Successors = nil
					for edge := random.IntN(4); edge > 0; edge-- {
						blk.Successors = append(blk.Successors, graph.Blocks[random.IntN(len(graph.Blocks))])
					}
					if !blk.Reachable {
						continue
					}
					for count := random.IntN(12); count > 0; count-- {
						variable := random.IntN(variables)
						e := variableEvent{variable: variable}
						if random.IntN(2) == 0 {
							previous := byVariable[variable]
							if len(previous) > 0 && random.IntN(3) == 0 {
								e.write = previous[random.IntN(len(previous))]
							} else {
								e.write = new(bool)
								byVariable[variable] = append(previous, e.write)
							}
							want[e.write] = true
							*e.write = true
						}
						blk.Events = append(blk.Events, e)
					}
				}
				for _, blk := range graph.Blocks {
					for index, e := range blk.Events {
						if e.write != nil && assignmentReachesRead(blk, index+1, e.variable) {
							want[e.write] = false
						}
					}
				}
				solveWriteLiveness(graph, variables)
				for node, wantDead := range want {
					if *node != wantDead {
						t.Fatalf("trial %d: dead=%v, want=%v", trial, *node, wantDead)
					}
				}
			}
		})
	}
}

func assignmentReachesRead(start *variableBlock, nextEvent int, variable int) bool {
	visited := make(map[*variableBlock]bool)
	var search func(*variableBlock, int) bool
	search = func(blk *variableBlock, index int) bool {
		for _, e := range blk.Events[index:] {
			if e.variable == variable {
				return e.write == nil
			}
		}
		for _, successor := range blk.Successors {
			if successor.Reachable && !visited[successor] {
				visited[successor] = true
				if search(successor, 0) {
					return true
				}
			}
		}
		return false
	}
	return search(start, nextEvent)
}

func TestBinderVariableLiveness(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		supported    bool
	}{
		{"straight line", `function f(g){let v=1;v=2;g(v);}`, true},
		{"branch", `function f(c,g){let v=0;if(c){v=1;}else{v=2;}g(v);v=3;}`, true},
		{"loop", `function f(c,g){let v=0;while(c()){g(v);v=1;}g(v);v=2;}`, true},
		{"compound rhs", `function f(g){let v=0;v+=(v=1);g(v);}`, true},
		{"logical rhs", `function f(g){let v=0;v||=(v=1);g(v);}`, false},
		{"conditional rhs", `function f(c,g){let v=0;v=c?v:(v=1);g(v);}`, true},
		{"optional call", `function f(g,o){let v=0;o?.m(v=1);g(v);}`, false},
		{"optional key", `function f(g,o){let v=0;o?.[v=1];g(v);}`, false},
		{"unreachable", `function f(g){let v=0;g(v);return;v=1;g(v);}`, true},
		{"typed write", `function f(g){let v=0;(v as number)=1;g(v);}`, false},
		{"typed reads", `function f(g){let v:number=0;g(v as number);v=1;g(v!);}`, true},
		{"unread", `function f(){let v=0;v=1;}`, true},
		{"parameter binding", `function f(v,g){g(v);v=1;}`, false},
		{"uninitialized binding", `function f(g){let v;g(v);v=1;}`, true},
		{"destructuring", `function f(o,g){let {a:v,b=v}=o;g(b);}`, false},
		{"destructuring rhs", `function f(o,g){let v=0;g(v);({[v=1]:x}=(v=2,o));g(v);}`, false},
		{"boolean condition", `function f(g){let v=0;if(true)v=1;g(v);}`, false},
		{"try", `function f(g){let v=0;try{v=1;}catch{}g(v);}`, false},
		{"finally", `function f(c,g){let v=0;g(v);try{if(c)return;}finally{v=1;}g(v);}`, false},
		{"type query", `function f(g){let v=0;g(v);v=1;type T=typeof v;}`, false},
		{"for increment after return", `function f(g){let v=0;for(v=1;(v=1);v++){return v;}g(v);}`, false},
		{"optional condition", `function f(c,g,o){let v=0;if(o?.m(v=1)){throw v;}else{for(v=1;c();v++){}}g(v);}`, false},
		{"logical loop condition", `function f(c,g){let v=0;do{v=(g(v),1);}while(c()&&(v=1));for(v=1;c()&&(v=1);v++){v||=(v=1);if(c())break;}g(v);}`, false},
		{"for of", `function f(g,xs){let v=0;for(v of xs)g(v);}`, false},
		{"switch", `function f(c,g){let v=0;switch(c){case 1:v=1;break;}g(v);}`, false},
		{"label", `function f(c,g){let v=0;outer:while(c){g(v);v=1;continue outer;}}`, false},
		{"nested function", `function f(g){let v=0;function nested(){}g(v);}`, false},
		{"generator", `function* f(g){let v=0;yield v;v=1;g(v);}`, false},
		{"async", `async function f(g){let v=0;await g(v);v=1;}`, false},
		{"arrow", `const f=g=>{let v=0;g(v);v=1;};`, false},
		{"source", `let v=0;g(v);v=1;`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, reads, writes := variableQueryFixture(t, tc.source)
			checkVariableQuery(t, root, reads, writes, tc.supported, tc.source)
		})
	}
}

func TestBinderVariableLivenessMissingFlow(t *testing.T) {
	const source = `function f(g){let v=0;v=1;g(v);}`
	for _, tc := range []struct {
		name string
		flow *ast.FlowNode
	}{
		{"unbound read", nil},
		{"unsupported flow", &ast.FlowNode{Flags: ast.FlowFlagsReduceLabel}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, reads, writes := variableQueryFixture(t, source)
			for read := range reads {
				read.AsIdentifier().FlowNode = tc.flow
			}
			checkVariableQuery(t, root, reads, writes, false, source)
		})
	}
}

func TestVariableLivenessRepeatedWrites(t *testing.T) {
	for _, tc := range []struct {
		source    string
		supported bool
	}{
		{`function f(g){let v=0;v=1;g(v);}`, true},
		{`function f(g){let v=0;try{}finally{v=1;}g(v);}`, false},
	} {
		root, reads, writes := variableQueryFixture(t, tc.source)
		writes = append(writes, writes...)
		originalReads, originalWrites := maps.Clone(reads), slices.Clone(writes)
		checkVariableQuery(t, root, reads, writes, tc.supported, tc.source)
		if !maps.Equal(reads, originalReads) || !slices.Equal(writes, originalWrites) {
			t.Fatal("query changed its caller's references")
		}
	}
}

// The CFG remains the reference for binder reuse, independently of the rule's
// symbol-selection and reporting policy. Generated programs cover intertwined
// assignments, nested conditions, loop back edges and abrupt exits.
func TestBinderVariableLivenessGenerated(t *testing.T) {
	random := rand.New(rand.NewPCG(73, 19))

	conditions := []string{"c()", "v > 0", "c() ? v : (v=2)", "(v=1)", "!v"}
	statements := []string{"v=1;", "g(v);", "v=(v+1);", "v+=(v=1);", "v=c()?v:(v=1);", "g(v++);", "v=(g(v),1);", "return v;", "throw v;"}
	var program func(int) string
	program = func(depth int) string {
		if depth == 0 {
			return statements[random.IntN(len(statements))]
		}
		condition := conditions[random.IntN(len(conditions))]
		switch random.IntN(6) {
		case 0:
			return program(depth-1) + program(depth-1)
		case 1:
			return "if(" + condition + "){" + program(depth-1) + "}else{" + program(depth-1) + "}"
		case 2:
			return "while(" + condition + "){" + program(depth-1) + "if(c())break;}"
		case 3:
			return "{v=1;while(" + condition + "){" + program(depth-1) + "v++;if(c())continue;if(c())break;}}"
		case 4:
			return "do{" + program(depth-1) + "}while(" + condition + ");"
		default:
			return statements[random.IntN(len(statements))]
		}
	}
	for trial := range 1200 {
		source := "function f(c,g,o){let v=0;" + program(3) + "g(v);v=3;}"
		root, reads, writes := variableQueryFixture(t, source)
		t.Run(strconv.Itoa(trial), func(t *testing.T) { checkVariableQuery(t, root, reads, writes, true, source) })
	}
}

func checkVariableQuery(t *testing.T, root *ast.Node, reads map[*ast.Node]int, writes []VariableWrite, supported bool, source string) {
	t.Helper()
	want := make(map[*ast.Node]bool)
	for _, node := range deadWritesFromCFG(root, reads, writes, 1) {
		want[node] = true
	}
	fromBinder, gotSupport := deadWritesFromBinder(root, reads, writes)
	dead := make(map[*ast.Node]bool)
	for _, node := range fromBinder {
		dead[node] = true
	}
	if gotSupport != supported {
		t.Fatalf("binder supported=%v, want=%v\n%s", gotSupport, supported, source)
	}
	if gotSupport && !maps.Equal(dead, want) {
		t.Fatalf("binder dead positions=%v, CFG=%v\n%s", writePositions(dead), writePositions(want), source)
	}
	if !gotSupport && len(dead) != 0 {
		t.Fatal("failed binder query left partial results")
	}
	got := make(map[*ast.Node]bool)
	for _, node := range DeadWrites(root, reads, writes, 1) {
		if got[node] {
			t.Fatal("duplicate dead write")
		}
		got[node] = true
	}
	if !maps.Equal(got, want) {
		t.Fatalf("DeadWrites positions=%v, CFG=%v\n%s", writePositions(got), writePositions(want), source)
	}
}

func writePositions(nodes map[*ast.Node]bool) []int {
	positions := make([]int, 0, len(nodes))
	for node := range nodes {
		positions = append(positions, node.Pos())
	}
	return positions
}

// Resolve a single fixture variable with the same binder and shared reference
// predicates used by consumers. Read/write selection is outside DeadWrites.
func variableQueryFixture(t *testing.T, source string) (*ast.Node, map[*ast.Node]int, []VariableWrite) {
	t.Helper()
	sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/variable.ts", Path: "/variable.ts"}, source, core.ScriptKindTS)
	binder.BindSourceFile(sf)
	var nodes []*ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		nodes = append(nodes, n)
		n.ForEachChild(func(c *ast.Node) bool { walk(c); return false })
	}
	walk(sf.AsNode())
	var symbol *ast.Symbol
	var root *ast.Node
	for _, n := range nodes {
		if (n.Kind == ast.KindVariableDeclaration || n.Kind == ast.KindParameter || n.Kind == ast.KindBindingElement) && n.Name().Kind == ast.KindIdentifier && n.Name().Text() == "v" {
			symbol = n.Symbol()
			root = RootOf(n.Name())
			break
		}
	}
	if symbol == nil || root == nil {
		t.Fatal("fixture has no declared variable v")
	}
	resolver := binder.NameResolver{CompilerOptions: &core.CompilerOptions{}, Globals: sf.Locals}
	reads := make(map[*ast.Node]int)
	for _, n := range nodes {
		if n.Kind != ast.KindIdentifier || !scope.IsReferenceIdentifier(n) || RootOf(n) != root || resolver.Resolve(n, n.Text(), ast.SymbolFlagsValue, nil, false, false) != symbol {
			continue
		}
		read := !utils.IsWriteReference(n)
		current := n
		for current.Parent != nil && ast.IsOuterExpression(current.Parent, ast.OEKParentheses|ast.OEKAssertions) {
			current = current.Parent
		}
		parent := current.Parent
		if parent != nil {
			if parent.Kind == ast.KindBinaryExpression {
				binary := parent.AsBinaryExpression()
				read = read || (binary.Left == current && binary.OperatorToken.Kind != ast.KindEqualsToken && ast.IsAssignmentOperator(binary.OperatorToken.Kind))
			}
			if parent.Kind == ast.KindPostfixUnaryExpression {
				read = true
			}
			if parent.Kind == ast.KindPrefixUnaryExpression {
				op := parent.AsPrefixUnaryExpression().Operator
				read = read || op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken
			}
		}
		if read {
			reads[n] = 0
		}
	}
	writes := make(map[*ast.Node]int)
	Build(root, Hooks[struct{}]{Write: func(_ *Builder[struct{}], n *ast.Node) {
		if _, seen := writes[n]; seen || n.Kind != ast.KindIdentifier || n.Text() != "v" {
			return
		}
		var target *ast.Symbol
		if n.Parent.Name() == n && (n.Parent.Kind == ast.KindVariableDeclaration || n.Parent.Kind == ast.KindBindingElement || n.Parent.Kind == ast.KindParameter) {
			target = n.Parent.Symbol()
		} else {
			target = resolver.Resolve(n, n.Text(), ast.SymbolFlagsValue, nil, false, false)
		}
		if target == symbol {
			writes[n] = 0
		}
	}})
	if len(writes) == 0 {
		t.Fatalf("fixture has no writes: %s", source)
	}
	var assignments []VariableWrite
	for node, variable := range writes {
		assignments = append(assignments, VariableWrite{Node: node, Variable: variable})
	}
	return root, reads, assignments
}
