package utils_test

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
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
