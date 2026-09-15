package unicornutil_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNewExpressionToCallFixesOperandGrouping(t *testing.T) {
	for _, test := range []struct {
		name   string
		code   string
		suffix string
		output string
	}{
		{
			name:   "return member",
			code:   "function f() { return new // value\nBuffer([1]).length; }",
			suffix: ".from",
			output: "function f() { return ( // value\nBuffer.from([1]).length); }",
		},
		{
			name:   "throw call chain",
			code:   "function f() { throw new // value\nBuffer(1).toString(); }",
			suffix: ".alloc",
			output: "function f() { throw ( // value\nBuffer.alloc(1).toString()); }",
		},
		{
			name:   "yield binary expression",
			code:   "function* f() { yield new // value\nBuffer(1).length + 1; }",
			suffix: ".alloc",
			output: "function* f() { yield ( // value\nBuffer.alloc(1).length + 1); }",
		},
		{
			name:   "suggestion suffix",
			code:   "function f() { return new // value\nBuffer(input).length; }",
			suffix: ".from",
			output: "function f() { return ( // value\nBuffer.from(input).length); }",
		},
		{
			name:   "existing operand parentheses",
			code:   "function f() { return (new // value\nBuffer(1)).length; }",
			suffix: ".alloc",
			output: "function f() { return (// value\nBuffer.alloc(1)).length; }",
		},
		{
			name:   "preceding unary operator",
			code:   "function f() { return !new // value\nBuffer(1); }",
			suffix: ".alloc",
			output: "function f() { return !// value\nBuffer.alloc(1); }",
		},
		{
			name:   "new-for-builtins return member",
			code:   "function f() { return new // value\nSymbol('x').description; }",
			output: "function f() { return ( // value\nSymbol('x').description); }",
		},
		{
			name:   "new-for-builtins throw call",
			code:   "function f() { throw new // value\nSymbol('x').toString(); }",
			output: "function f() { throw ( // value\nSymbol('x').toString()); }",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.js"}, test.code, core.ScriptKindJS)
			var constructor *ast.Node
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindNewExpression && constructor == nil {
					constructor = node
				}
				return node.ForEachChild(func(child *ast.Node) bool {
					child.Parent = node
					return visit(child)
				})
			}
			visit(sourceFile.AsNode())
			if constructor == nil {
				t.Fatal("expected a constructor")
			}
			newExpression := constructor.AsNewExpression()
			fixes := unicornutil.NewExpressionToCallFixes(sourceFile, constructor,
				utils.TrimNodeTextRange(sourceFile, constructor), newExpression, newExpression.Expression, test.suffix)
			if len(fixes) == 0 {
				t.Fatal("expected constructor-to-call fixes")
			}
			output, _, _ := linter.ApplyRuleFixes(test.code, []rule.RuleSuggestion{{FixesArr: fixes}})
			if output != test.output {
				t.Fatalf("output = %q, want %q", output, test.output)
			}
		})
	}
}
