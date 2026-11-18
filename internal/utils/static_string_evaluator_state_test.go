package utils

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestStaticCallEscapeScanCost(t *testing.T) {
	for _, length := range []int{8, 16, 32, 64, 128} {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			code := "unknown" + strings.Repeat(".slice()", length) + ";"
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/chain.ts", Path: "/chain.ts"}, code, core.ScriptKindTS)
			resolver := &countingStaticReferenceResolver{}
			evaluator := NewStaticStringEvaluatorWithCallEscapes(nil, source, resolver)
			evaluator.ensureReferenceFlags()
			if resolver.calls > 64*(length+1) {
				t.Fatalf("%d resolutions exceed linear bound for %d calls", resolver.calls, length)
			}
			if evaluator.evaluation != nil {
				t.Fatal("reference-scan values must not escape into ordinary evaluation")
			}
		})
	}
}

func TestStaticReferenceScanCacheIsolation(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir.Dir, "file.ts")
	code := `
		const array = ["fill"];
		const other = ["p"];
		function unused() { array[method]("unused"); }
		const fill = "fill";
		array[fill]("x");
		const method = array[0];
		const value = other[0] + method;
	`
	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
	assert.NilError(t, err, "couldn't create program")
	source := program.GetSourceFile(filePath)
	assert.Assert(t, source != nil)
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	evaluator := NewStaticStringEvaluatorWithSourceFile(typeChecker, source)
	if _, known := evaluator.Eval(findVariableInitializer(t, source, "value")); known {
		t.Fatal("mutation-scan values must not escape into ordinary evaluation")
	}
}

func TestStaticMethodNameObservationOrder(t *testing.T) {
	for _, test := range []struct {
		name, operation, call string
		known                 bool
	}{
		{"preceding escape", `unknown(opts);`, `array[opts.current](0)`, false},
		{"following argument escape", ``, `array[opts.current](unknown(opts))`, true},
		{"following statement escape", ``, `array[opts.current](0); unknown(opts)`, true},
		{"escape inside key", ``, `array[(unknown(opts), opts.current)](0)`, false},
		{"unrelated escape", `unknown({current: "other"});`, `array[opts.current](0)`, true},
		{"captured primitive", `const key = opts.current; unknown(opts);`, `array[key](0)`, true},
		{"captured primitive before later arrow", `const key = opts.current;`, `array[key](() => unknown(opts))`, true},
		{"captured primitive before later function", `const key = opts.current;`, `array[key](function () { unknown(opts); })`, true},
		{"capture after invoked arrow", `(() => unknown(opts))(); const key = opts.current;`, `array[key](0)`, false},
		{"capture after hoisted call", `change(); const key = opts.current; function change() { unknown(opts); }`, `array[key](0)`, false},

		{"loop can repeat later escape", ``, `array[opts.current](unknown(opts))`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			code := `const opts = {current: "some"}; ` + test.operation + ` const result = ` + test.call + `;`
			if test.name == "loop can repeat later escape" {
				code = `const opts = {current: "some"}; for (;;) { const result = ` + test.call + `; }`
			}
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/method.ts", Path: "/method.ts"}, code, core.ScriptKindTS)
			binder.BindSourceFile(source)
			var access *ast.Node
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if ast.IsElementAccessExpression(node) && node.Expression().Text() == "array" {
					access = node
					return true
				}
				return node.ForEachChild(visit)
			}
			visit(source.AsNode())
			if access == nil {
				t.Fatal("missing method access")
			}
			resolver := &staticEvaluatorTestResolver{resolver: binder.NameResolver{Globals: source.Locals}}
			evaluator := NewStaticStringEvaluatorWithCallEscapes(nil, source, resolver)
			name, known := evaluator.EvalAccessExpressionNameAtLocation(access)
			if known != test.known || known && name != "some" {
				t.Fatalf("method=(%q,%v), want known=%v", name, known, test.known)
			}
			// Reusing the original evaluator must not retain the observation window.
			if test.name == "following argument escape" {
				if _, known := evaluator.EvalAccessExpressionName(access); known {
					t.Fatal("location view changed ordinary escape analysis")
				}
			}
		})
	}
}

func TestStaticMethodObservationScanCost(t *testing.T) {
	for _, length := range []int{8, 32, 128} {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			code := `const opts = {current: "some"}; ` + strings.Repeat("array[opts.current](0);\n", length)
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/methods.ts", Path: "/methods.ts"}, code, core.ScriptKindTS)
			resolver := &countingStaticReferenceResolver{}
			evaluator := NewStaticStringEvaluatorWithCallEscapes(nil, source, resolver)
			count := 0
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if ast.IsElementAccessExpression(node) {
					evaluator.EvalAccessExpressionNameAtLocation(node)
					count++
				}
				return node.ForEachChild(visit)
			}
			visit(source.AsNode())
			if count != length {
				t.Fatalf("observed %d methods, want %d", count, length)
			}
			if resolver.calls > 64*(length+1) {
				t.Fatalf("%d resolutions exceed the linear bound for %d methods", resolver.calls, length)
			}
		})
	}
}

func TestStaticPrimitiveCaptureScope(t *testing.T) {
	for _, test := range []struct {
		name, code string
		known      bool
	}{
		{"outer mutable object", `const opts={current:"some"}; function visit(){const key=opts.current; array[key](()=>unknown(opts));}`, false},
		{"local fresh object", `function visit(){const opts={current:"some"}; const key=opts.current; array[key](()=>unknown(opts));}`, true},
		{"alias of outer object", `const shared={current:"some"}; function visit(){const alias=shared; const key=alias.current; array[key](()=>unknown(alias));}`, false},
		{"literal primitive", `const opts={current:"some"}; function visit(){const key="some"; array[key](()=>unknown(opts));}`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/captures.ts", Path: "/captures.ts"}, test.code, core.ScriptKindTS)
			binder.BindSourceFile(source)
			var access *ast.Node
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if ast.IsElementAccessExpression(node) && node.Expression().Text() == "array" {
					access = node
					return true
				}
				return node.ForEachChild(visit)
			}
			visit(source.AsNode())
			if access == nil {
				t.Fatal("missing method access")
			}
			resolver := &staticEvaluatorTestResolver{resolver: binder.NameResolver{Globals: source.Locals}}
			evaluator := NewStaticStringEvaluatorWithCallEscapes(nil, source, resolver)
			method, known := evaluator.EvalAccessExpressionNameAtLocation(access)
			if known != test.known || known && method != "some" {
				t.Fatalf("method=(%q,%v), want known=%v", method, known, test.known)
			}
		})
	}
}
