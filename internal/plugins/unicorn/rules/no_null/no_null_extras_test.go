// TestNoNullExtras covers tsgo/ESTree shape differences, real-user cases, and
// branch lock-ins that are intentionally separate from no_null_upstream_test.go.
package no_null_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_null "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_null"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func typescriptReplacementCase(code string, options any, occurrences ...int) rule_tester.InvalidTestCase {
	testCase := replacementCase(code, options, occurrences...)
	testCase.FileName = "file.ts"
	return testCase
}

func typescriptMutableVariableCase(code, removedOutput string, occurrence int) rule_tester.InvalidTestCase {
	testCase := mutableVariableCaseAt(code, removedOutput, occurrence)
	testCase.FileName = "file.ts"
	return testCase
}

func TestNoNullExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_null.NoNullRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: null in TypeScript type positions is not a runtime literal ----
			{Code: `type MaybeString = string | null;`, FileName: "file.ts"},
			{Code: `interface Value { current: null }`, FileName: "file.ts"},

			// ---- Dimension 4: ESTree drops parentheses around strict-comparison operands ----
			{Code: `if (value === (((null)))) {}`, FileName: "file.js"},
			{Code: `if ((((null))) !== value) {}`, FileName: "file.js"},
			{Code: `if (value === null) {}`, FileName: "file.js", Options: []any{map[string]any{}}},

			// ---- Dimension 4: direct arguments remain direct through parentheses ----
			{Code: `fn(((null)))`, FileName: "file.js", Options: ignoreArguments},
			{Code: `new Box(((null)))`, FileName: "file.js", Options: ignoreArguments},

			// ---- Branch lock-ins: every built-in exception accepts ESTree-transparent grouping ----
			{Code: `((Object)).create(((null)))`, FileName: "file.js"},
			{Code: `((useRef))(((null)))`, FileName: "file.js"},
			{Code: `useRef?.(null)`, FileName: "file.js"},
			{Code: `((React)).useRef(((null)))`, FileName: "file.js"},
			{Code: `node?.insertBefore(other, ((null)))`, FileName: "file.js"},

			// The exceptions are syntactic, like upstream; local bindings with the
			// same spelling do not change the result.
			{Code: `const Object = factory; Object.create(null);`, FileName: "file.js"},
			{Code: `const useRef = custom; useRef(null);`, FileName: "file.js"},

			// ---- Real-user: #2633 / #1842 — APIs can require null arguments ----
			{Code: `shape.setMap(null)`, FileName: "file.js", Options: ignoreArguments},
			{Code: `markers[index].setMap(null)`, FileName: "file.js", Options: ignoreArguments},

			// ---- Real-user: #888 — React refs are initialized with null ----
			{Code: `const ref = useRef(null);`, FileName: "file.ts"},
			{Code: `const ref = React.useRef(null);`, FileName: "file.ts"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parentheses are transparent for reporting and edits ----
			replacementCase(`const value = (((null)));`, nil, 1),
			fixedCase(`if (value == (((null)))) {}`),
			returnCase(`function value() { return (null); }`, nil),

			// ---- Dimension 4: authored TypeScript wrappers remain visible upstream ----
			typescriptReplacementCase(`fn(null as unknown)`, ignoreArguments, 1),
			typescriptReplacementCase(`fn(null!)`, ignoreArguments, 1),
			typescriptReplacementCase(`if (value === (null as unknown)) {}`, nil, 1),
			typescriptReplacementCase(`if (value == (null as unknown)) {}`, nil, 1),

			// ---- Dimension 4: optional calls/members do not match exceptions that forbid them ----
			replacementCase(`React.useRef?.(null)`, nil, 1),
			replacementCase(`React?.useRef(null)`, nil, 1),
			replacementCase(`Object?.create(null)`, nil, 1),
			replacementCase(`node.insertBefore?.(other, null)`, nil, 1),

			// ---- Branch lock-ins: exception arity and direct-position gates ----
			replacementCase(`useRef(null, initialValue)`, nil, 1),
			replacementCase(`React.useRef(null, initialValue)`, nil, 1),
			replacementCase(`node.insertBefore(null, other)`, nil, 1),
			replacementCase(`Object.create(proto, null)`, nil, 1),

			// checkArguments only ignores a direct argument, not a nested null.
			replacementCase(`fn({value: null})`, ignoreArguments, 1),
			replacementCase(`new Box([null])`, ignoreArguments, 1),

			// ---- Dimension 4: ts-go CallExpression maps to ESTree ImportExpression ----
			replacementCase(`import(null)`, ignoreArguments, 1),

			// ---- Dimension 4: variable declaration shapes and TSESTree id ranges ----
			typescriptMutableVariableCase(
				`let value: string | null = null;`,
				`let value: string | null;`,
				2,
			),
			typescriptMutableVariableCase(
				`let {value}: {value: unknown} = null;`,
				`let {value}: {value: unknown};`,
				1,
			),
			mutableVariableCase(
				`for (let value = null; ; ) {}`,
				`for (let value; ; ) {}`,
			),
			mutableVariableCase(
				`using value = null;`,
				`using value;`,
			),
			mutableVariableCase(
				`async function f() { await using value = null; }`,
				`async function f() { await using value; }`,
			),

			// ---- Dimension 4: same-kind nesting reports each null independently ----
			replacementCase(`const first = null, second = null;`, nil, 1, 2),

			// An empty option object receives both schema defaults: direct arguments
			// are checked, while strict equality remains ignored.
			replacementCase(`fn(null)`, []any{map[string]any{}}, 1),

			// A null used as an optional-chain receiver is still a runtime literal.
			replacementCase(`const value = null?.property;`, nil, 1),

			// N/A: computed/private property-name equivalence classes do not apply;
			// the rule reports the null literal itself, not an enclosing property.
			// N/A: overload/ambient containers do not alter runtime-null handling;
			// their null occurrences are type nodes covered by the valid cases above.
		},
	)
}

func TestNoNullEditDemand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		source          string
		expectsAutofix  bool
		expectsSuggests bool
	}{
		{name: "loose equality autofix", source: `if (value == null) {}`, expectsAutofix: true},
		{name: "replacement suggestion", source: `const value = null;`, expectsSuggests: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{
				rule.EditDemandNone,
				rule.EditDemandAutofix,
				rule.EditDemandSuggestion,
				rule.EditDemandAll,
			} {
				diagnostics[demand] = lintNoNullWithDemand(t, test.source, demand)
			}

			withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
				diagnostic.FixesPtr = nil
				diagnostic.Suggestions = nil
				return diagnostic
			}
			base := withoutEdits(diagnostics[rule.EditDemandNone])
			for demand, diagnostic := range diagnostics {
				if got := withoutEdits(diagnostic); !reflect.DeepEqual(got, base) {
					t.Errorf("demand %d changed diagnostic identity:\ngot  %#v\nwant %#v", demand, got, base)
				}
			}

			if diagnostics[rule.EditDemandNone].FixesPtr != nil ||
				diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
				t.Fatal("a non-autofix demand materialized fixes")
			}
			if diagnostics[rule.EditDemandNone].Suggestions != nil ||
				diagnostics[rule.EditDemandAutofix].Suggestions != nil {
				t.Fatal("a non-suggestion demand materialized suggestions")
			}

			autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
			allAutofixes := diagnostics[rule.EditDemandAll].FixesPtr
			if test.expectsAutofix {
				if autofixOnly == nil || allAutofixes == nil || !reflect.DeepEqual(*autofixOnly, *allAutofixes) {
					t.Fatal("autofix and all-edits demands produced different fixes")
				}
			} else if autofixOnly != nil || allAutofixes != nil {
				t.Fatal("suggestion-only diagnostic materialized autofixes")
			}

			suggestionOnly := diagnostics[rule.EditDemandSuggestion].Suggestions
			allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
			if test.expectsSuggests {
				if suggestionOnly == nil || allSuggestions == nil || !reflect.DeepEqual(*suggestionOnly, *allSuggestions) {
					t.Fatal("suggestion and all-edits demands produced different suggestions")
				}
			} else if suggestionOnly != nil || allSuggestions != nil {
				t.Fatal("autofix-only diagnostic materialized suggestions")
			}
		})
	}
}

func lintNoNullWithDemand(t testing.TB, source string, demand rule.EditDemand) rule.RuleDiagnostic {
	t.Helper()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts",
		Path:     "/edit-demand.ts",
	}, source, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	diagnostics := make([]rule.RuleDiagnostic, 0, 1)
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer(
		no_null.NoNullRule.Name,
		rule.SeverityError,
		rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	)

	listeners := no_null.NoNullRule.Run(ctx, nil)
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if len(diagnostics) != 1 {
		t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
	}
	return diagnostics[0]
}
