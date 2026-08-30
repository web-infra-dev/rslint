package utils_test

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestGetFunctionHeadLocConstructorWithDecoratorFactory pins down that the
// reported range for a nameless member (constructor) whose class has a
// decorator factory starts *after* the decorators and ends at the
// parameter-list `(` — not at the decorator-factory `(`. Without skipping
// the decorators when computing the parameter-search origin, the scan would
// match the `(` inside `@Dec()` and produce an inverted range.
func TestGetFunctionHeadLocConstructorWithDecoratorFactory(t *testing.T) {
	code := "declare function Dec(): any;\nclass A {\n  @Dec()\n  constructor(x: number) {}\n}\n"

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	_, sourceFile, err := helper.CreateTestProgram(code, "a.ts", "tsconfig.json")
	if err != nil {
		t.Fatalf("CreateTestProgram: %v", err)
	}

	var ctor *ast.Node
	var walk func(n *ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		if n.Kind == ast.KindConstructor {
			ctor = n
			return true
		}
		stop := false
		n.ForEachChild(func(c *ast.Node) bool {
			if walk(c) {
				stop = true
				return true
			}
			return false
		})
		return stop
	}
	walk(sourceFile.AsNode())

	if ctor == nil {
		t.Fatal("constructor not found in parsed source")
	}

	r := utils.GetFunctionHeadLoc(sourceFile, ctor)
	if r.Pos() >= r.End() {
		t.Fatalf("inverted range: pos=%d end=%d", r.Pos(), r.End())
	}

	got := sourceFile.Text()[r.Pos():r.End()]
	// "constructor" is 11 characters, the head range must cover just the
	// keyword up to (but not including) the parameter `(`.
	if strings.TrimSpace(got) != "constructor" {
		t.Fatalf("unexpected head range text: %q (pos=%d end=%d)", got, r.Pos(), r.End())
	}
}

func TestGetFunctionHeadLocTypeAccessorComputedKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "interface parenthesized key", code: `interface I { set [('a')](value: string); }`, want: `set [`},
		{name: "type literal call key", code: `type T = { get [foo()](): string; };`, want: `get [foo`},
		{name: "optional call key", code: `interface I { set [foo?.()](value: string); }`, want: `set [foo?.`},
		{name: "comment fake paren", code: `interface I { set [/* ( */ 'a'](value: string); }`, want: `set [/* ( */ 'a']`},
		{name: "string fake paren", code: `interface I { set ['('](value: string); }`, want: `set ['(']`},
		{name: "regexp fake paren", code: `interface I { set [/a(b)/](value: string); }`, want: `set [/a(b)/]`},
		{name: "template text fake paren", code: "interface I { set [`(`](value: string); }", want: "set [`(`]"},
		{name: "template expression call", code: "interface I { set [`${foo()}`](value: string); }", want: "set [`${foo"},
		{name: "multiline call key", code: "interface I {\n  set [\n    foo()\n  ](value: string);\n}", want: "set [\n    foo"},
		{name: "declare class keeps complete key", code: `declare class C { set [('a')](value: string); }`, want: `set [('a')]`},
		{name: "class keeps complete key", code: `class C { set [('a')](value: string) {} }`, want: `set [('a')]`},
		{name: "object keeps complete key", code: `({ set [('a')](value: string) {} })`, want: `set [('a')]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/file.ts",
				Path:     "/file.ts",
			}, tt.code, core.ScriptKindTS)

			var accessor *ast.Node
			var walk func(*ast.Node) bool
			walk = func(node *ast.Node) bool {
				if node == nil {
					return false
				}
				if node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
					accessor = node
					return true
				}
				return node.ForEachChild(walk)
			}
			walk(sourceFile.AsNode())
			if accessor == nil {
				t.Fatal("accessor not found")
			}

			range_ := utils.GetFunctionHeadLoc(sourceFile, accessor)
			if range_.Pos() > range_.End() {
				t.Fatalf("inverted range: [%d,%d)", range_.Pos(), range_.End())
			}
			if got := strings.TrimLeft(sourceFile.Text()[range_.Pos():range_.End()], " \t"); got != tt.want {
				t.Fatalf("head text = %q, want %q", got, tt.want)
			}
		})
	}
}
