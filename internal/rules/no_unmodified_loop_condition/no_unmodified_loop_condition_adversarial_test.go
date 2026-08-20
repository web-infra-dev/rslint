package no_unmodified_loop_condition

import (
	"fmt"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// referenceCollectIdentifierSymbols is the deliberately slower pre-optimization
// algorithm. Keeping it in tests gives the single-pass traversal an independent
// behavioral oracle for ancestor selection and dynamic-group filtering.
func referenceCollectIdentifierSymbols(condition *ast.Node, refs *rule.RefStore, checkConditionalExpressions bool) []identifierRef {
	dynamicGroups := map[*ast.Node]bool{}
	checkedDynamicGroups := map[*ast.Node]bool{}
	var result []identifierRef

	findGroup := func(identifier *ast.Node) (*ast.Node, bool) {
		var group *ast.Node
		for current := identifier; current != nil; current = current.Parent {
			if isConditionGroup(current, checkConditionalExpressions) {
				group = current
			}
			if current == condition {
				return group, true
			}
		}
		return nil, false
	}

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || isConditionBoundary(node) {
			return
		}
		if node.Kind == ast.KindIdentifier {
			symbol := refs.Resolve(node)
			if symbol == nil {
				return
			}
			group, ok := findGroup(node)
			if !ok {
				return
			}
			if group != nil {
				if !checkedDynamicGroups[group] {
					checkedDynamicGroups[group] = true
					dynamicGroups[group] = hasDynamicExpression(group)
				}
				if dynamicGroups[group] {
					return
				}
			}
			result = append(result, identifierRef{symbol: symbol, node: node, group: group})
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(condition)
	return result
}

// TestNoUnmodifiedLoopConditionOptimizationInvariants attacks the assumptions
// used by the single-pass collector: outermost groups are contiguous, logical
// operands split groups, function/class bodies are skipped, and enabling
// conditional-expression checking exposes the ternary branches independently.
func TestNoUnmodifiedLoopConditionOptimizationInvariants(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnmodifiedLoopConditionRule,
		[]rule_tester.ValidTestCase{
			{
				// Every reference belongs to the outer === group, even though it
				// contains nested binary expressions. Modifying c satisfies it.
				Code: `let a = 0, b = 1, c = 2, d = 3; while ((a < b) === (c < d)) { c++; }`,
			},
			{
				// The repeated a references occur in groups separated by logical
				// operands. The modification cache must not merge those groups.
				Code: `let a = 0, b = 1, c = 2; while ((a < b) && a && (a < c)) { a++; }`,
			},
			{
				// With the default option, the outer conditional remains one group.
				Code: `let choose = true, left = 1, right = 2, other = 3; while (choose ? left < right : other) { choose = false; left++; }`,
			},
			{
				// A dynamic node anywhere in a binary group suppresses the entire
				// group, including identifiers visited after the call expression.
				Code: `function sideEffect() { return 0; } let x = 0; while (sideEffect() < x) { }`,
			},
			{
				// The optimized report loop must still pass every diagnostic through
				// the ordinary line-disable path.
				Code: "let a = 0, b = 1;\n// eslint-disable-next-line test\nwhile (a && b) {}",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				// Dynamic expressions nested inside a skipped function expression
				// must not make the surrounding binary group dynamic.
				Code: `declare function sideEffect(): number; let x = 0; while (x < (() => sideEffect())) { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Message: "'x' is not modified in this loop."},
				},
			},
			{
				// Latest ESLint main behavior: the ternary itself is split with the
				// option enabled, while left < right remains a binary group.
				Code:    `let choose = true, left = 1, right = 2, other = 3; while (choose ? left < right : other) { choose = false; left++; }`,
				Options: map[string]any{"checkConditionalExpressions": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Message: "'other' is not modified in this loop."},
				},
			},
			{
				// A modified first group must not swallow later independent groups;
				// the remaining diagnostics must stay in source order.
				Code: `let a = 0, b = 1, c = 2, d = 3, e = 4; while ((a < b) && c && (d < e)) { a++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified", Message: "'c' is not modified in this loop."},
					{MessageId: "loopConditionNotModified", Message: "'d' is not modified in this loop."},
					{MessageId: "loopConditionNotModified", Message: "'e' is not modified in this loop."},
				},
			},
		},
	)
}

func TestSinglePassCollectorMatchesReference(t *testing.T) {
	expressions := []string{
		`x`,
		`!x`,
		`x < y`,
		`(x < y) && z`,
		`x || y < z`,
		`(x < y) === (z < w)`,
		`x ? y : z`,
		`x ? y < z : w`,
		`x = y < z`,
		`(x, y < z)`,
		`x < call(y)`,
		`call(x) < y`,
		`x < obj.y`,
		`x < new C()`,
		"x < tag`y`",
		`import(path) < x`,
		`x < (() => call(y))`,
		`x < class { field = y }`,
		`x && (y < z)`,
		`(x < y) ?? (z < w)`,
		`x++ < y`,
	}

	for _, checkConditionalExpressions := range []bool{false, true} {
		for _, expression := range expressions {
			name := fmt.Sprintf("conditional=%t/%s", checkConditionalExpressions, expression)
			t.Run(name, func(t *testing.T) {
				const fileName = "/collector-differential.ts"
				source := `export {};
declare function call(value: unknown): number;
declare class C {}
declare const tag: (parts: TemplateStringsArray) => number;
declare const obj: { y: number };
let x = 0, y = 1, z = 2, w = 3, path = "./module.js";
while (` + expression + `) {}`
				sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
					FileName: fileName,
					Path:     tspath.Path(fileName),
				}, source, core.ScriptKindTS)
				binder.BindSourceFile(sourceFile)
				_, refsInit := rule.ResolveLanguageDefaults(fileName)
				refs := rule.NewRefStore(sourceFile, &core.CompilerOptions{}, nil, refsInit)

				var condition *ast.Node
				var visit func(*ast.Node) bool
				visit = func(node *ast.Node) bool {
					if node.Kind == ast.KindWhileStatement {
						condition = node.AsWhileStatement().Expression
						return true
					}
					return node.ForEachChild(visit)
				}
				sourceFile.AsNode().ForEachChild(visit)
				if condition == nil {
					t.Fatal("fixture has no while condition")
				}

				want := referenceCollectIdentifierSymbols(condition, refs, checkConditionalExpressions)
				got := collectIdentifierSymbols(condition, refs, checkConditionalExpressions)
				if len(got) != len(want) {
					t.Fatalf("references = %d, want %d", len(got), len(want))
				}
				for index := range want {
					if got[index] != want[index] {
						t.Errorf("reference %d = %#v, want %#v", index, got[index], want[index])
					}
				}
			})
		}
	}
}
