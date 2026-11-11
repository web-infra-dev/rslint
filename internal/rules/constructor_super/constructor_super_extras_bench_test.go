package constructor_super

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Cold inputs measure the first callback on a newly parsed/bound AST. Warm
// inputs expose reusable AST facts; neither benchmark includes parse/bind time.
func BenchmarkConstructorSuper(b *testing.B) {
	for _, size := range []int{20, 1000} {
		for _, derived := range []bool{false, true} {
			for _, cold := range []bool{false, true} {
				source := "class A"
				if derived {
					source += " extends B"
				}
				source += "{constructor(){"
				if derived {
					source += "super();"
				}
				source += strings.Repeat("this.x=1;", size) + "}}"
				b.Run(fmt.Sprintf("statements=%d/derived=%t/cold=%t", size, derived, cold), func(b *testing.B) { benchmarkConstructor(b, source, cold) })
			}
		}
	}
	for _, depth := range []int{24, 64} {
		source := "class A extends B{constructor(){" + strings.Repeat("while(a){", depth) + "super();" + strings.Repeat("}", depth) + "}}"
		b.Run(fmt.Sprintf("nested-loops=%d", depth), func(b *testing.B) { benchmarkConstructor(b, source, false) })
	}
}
func benchmarkConstructor(b *testing.B, source string, cold bool) {
	prepare := func() (func(*ast.Node), *ast.Node) {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/benchmark.js", Path: "/benchmark.js"}, source, core.ScriptKindJS)
		binder.BindSourceFile(file)
		var constructor *ast.Node
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n.Kind == ast.KindConstructor {
				constructor = n
				return
			}
			n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
		if constructor == nil {
			b.Fatal("benchmark input has no constructor")
		}
		context := (rule.RuleContext{SourceFile: file}).WithReporter("constructor-super", rule.SeverityError, func(rule.RuleDiagnostic) {})
		return ConstructorSuperRule.Run(context, nil)[ast.KindConstructor], constructor
	}
	callback, node := prepare()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if cold {
			b.StopTimer()
			callback, node = prepare()
			b.StartTimer()
		}
		callback(node)
	}
}
