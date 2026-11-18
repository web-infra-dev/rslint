package no_duplicate_type_constituents

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Unlike an output-only assertion, this checks that every fix pass preserves
// valid syntax and the assignments in the fixture. noImplicitAny=false makes
// mapped types without a value annotation legal, as in the review reproduction.
func lintSafetyFixture(t testing.TB, code string, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	root := fixtures.GetRootDir()
	file := tspath.ResolvePath(root.Dir, "review.ts")
	config := tspath.ResolvePath(root.Dir, "review-tsconfig.json")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{
		file:   code,
		config: `{"compilerOptions":{"strict":true,"noImplicitAny":false},"files":["review.ts"]}`,
	})
	program, err := utils.CreateProgram(true, fs, root.Dir, config, utils.CreateCompilerHost(root.Dir, fs))
	if err != nil {
		t.Fatal(err)
	}
	source := program.GetSourceFile(file)
	if diagnostics := program.GetSyntacticDiagnostics(t.Context(), source); len(diagnostics) != 0 {
		t.Fatalf("syntax errors after fix: %v\n%s", diagnostics, code)
	}
	if diagnostics := program.GetSemanticDiagnostics(t.Context(), source); len(diagnostics) != 0 {
		t.Fatalf("type errors after fix: %v\n%s", diagnostics, code)
	}
	checker, release := program.GetTypeChecker(t.Context())
	defer release()
	var diagnostics []rule.RuleDiagnostic
	ctx := (rule.RuleContext{SourceFile: source, TypeChecker: checker}).
		WithProgram(lintprogram.NewFromCompiler(program)).
		WithDiagnosticConsumer("no-duplicate-type-constituents", rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		})
	listeners := NoDuplicateTypeConstituentsRule.Run(ctx, nil)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		node.ForEachChild(visit)
		return false
	}
	visit(source.AsNode())
	return diagnostics
}

func TestNoDuplicateTypeConstituentsFixSafety(t *testing.T) {
	cases := []string{
		"type T = string | number | (| string | number); const value: T = 1;",
		"type T = {a: string} & {b: number} & (& {a: string} & {b: number}); const value: T = {a: '', b: 1};",
		"function f(x?: | undefined | undefined) {}",
		"function f(x?: | undefined) {}",
		"function f(x?: (| undefined | (| undefined | undefined))) {}",
		"function f(x?: string | (| undefined | string)) {}",
		"function f(x?: string | number | (| undefined | string | number)) {}",
	}
	for _, operator := range []string{"|", "&"} {
		for _, count := range []int{2, 3, 5} {
			var members []string
			for i := range count {
				members = append(members, fmt.Sprintf("{p%d?: string}", i))
			}
			joined := strings.Join(members, " "+operator+" ")
			for _, leading := range []string{"", operator} {
				for _, depth := range []int{1, 3} {
					group := strings.Repeat("(", depth) + leading + " /*keep*/ " + joined + strings.Repeat(")", depth)
					cases = append(cases, "type T = "+joined+" "+operator+" "+group+"; const value: T = {};")
				}
			}
		}
	}
	for i, original := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			code := original
			for pass := range 10 {
				diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll)
				if pass == 0 {
					if len(diagnostics) == 0 {
						t.Fatal("fixture did not exercise the rule")
					}
					withoutEdits := lintSafetyFixture(t, code, rule.EditDemandNone)
					if len(diagnostics) != len(withoutEdits) {
						t.Fatal("edit demand changed diagnostics")
					}
					for j, diagnostic := range diagnostics {
						other := withoutEdits[j]
						if !reflect.DeepEqual(diagnostic.Message, other.Message) || diagnostic.Range != other.Range {
							t.Fatal("edit demand changed a message or range")
						}
						if len(other.Fixes()) != 0 {
							t.Fatal("diagnostic-only consumer received fixes")
						}
					}
				}
				fixed, _, changed := linter.ApplyRuleFixes(code, diagnostics)
				if !changed {
					if strings.HasPrefix(original, "type T") && len(diagnostics) != 0 {
						t.Fatalf("duplicate group survived autofix: %s", code)
					}
					if strings.Contains(original, "/*keep*/") && !strings.Contains(code, "/*keep*/") {
						t.Fatal("comment removed")
					}
					return
				}
				if fixed == code {
					t.Fatal("fix did not make progress")
				}
				code = fixed
			}
			t.Fatal("autofix did not converge")
		})
	}
}

func TestNoDuplicateTypeConstituentsMappedFieldRoles(t *testing.T) {
	for _, code := range []string{
		"type T = { [K in 'a' as 'b'] } | { [K in 'a']: 'b' }; const value: T = { a: 'b' };",
		"type T = { [K in 'a']: 'b' } | { [K in 'a' as 'b'] }; const value: T = { a: 'b' };",
		"type T = { [K in 'a' as 'b']?: 'c' } | { [K in 'a']?: 'b' }; const value: T = { a: 'b' };",
	} {
		if diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll); len(diagnostics) != 0 {
			t.Fatalf("distinct mapped key/value fields reported as duplicates: %v", diagnostics)
		}
	}
}

func BenchmarkNoDuplicateTypeConstituentsLargeUnion(b *testing.B) {
	code := syntaxBenchmarkSource(2500, "literal")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if diagnostics := lintSafetyFixture(b, code, rule.EditDemandAll); len(diagnostics) != 0 {
			b.Fatal("distinct literals reported as duplicates")
		}
	}
}

// Initial diagnostics match v8.72.2; the safe fix passes intentionally differ
// from upstream's invalid empty-group output.
func TestNoDuplicateTypeConstituentsLeadingSeparatorOutputs(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateTypeConstituentsRule, nil, []rule_tester.InvalidTestCase{
		{
			Code:   "type T = string | number | (| string | number);",
			Output: []string{"type T = string | number | (  | number);", "type T = string | number    ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 31, EndLine: 1, EndColumn: 37},
				{MessageId: "duplicate", Message: "Union type constituent is duplicated with number.", Line: 1, Column: 40, EndLine: 1, EndColumn: 46},
			},
		},
		{
			Code:   "type T = {a: string} & {b: number} & (& {a: string} & {b: number});",
			Output: []string{"type T = {a: string} & {b: number} & (  & {b: number});", "type T = {a: string} & {b: number}    ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with {a: string}.", Line: 1, Column: 41, EndLine: 1, EndColumn: 52},
				{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with {b: number}.", Line: 1, Column: 55, EndLine: 1, EndColumn: 66},
			},
		},
	})
}

func TestNoDuplicateTypeConstituentsAncestorTypes(t *testing.T) {
	for _, code := range []string{
		"type T = (string | never) | number; const value: T = 'keep';",
		"type T = (string | any) | boolean; const value: T = {keep: true};",
		"type T = ({a: string} & unknown) & {b: number}; const value: T = {a: '', b: 1};",
	} {
		if diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll); len(diagnostics) != 0 {
			t.Fatalf("a containing group is not a duplicate occurrence: %v", diagnostics)
		}
	}
	// Skipping an ancestor must still find an earlier sibling with an alias of
	// the same type, and preserve that sibling as the diagnostic's previous node.
	code := "type A = string; type B = string; type T = (A | B) | number; const value: T = 'keep';"
	diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll)
	if len(diagnostics) != 1 || diagnostics[0].Message.Description != "Union type constituent is duplicated with A." {
		t.Fatalf("expected only B to duplicate A: %v", diagnostics)
	}
	fixed, _, changed := linter.ApplyRuleFixes(code, diagnostics)
	if !changed {
		t.Fatal("alias duplicate has no fix")
	}
	if after := lintSafetyFixture(t, fixed, rule.EditDemandAll); len(after) != 0 {
		t.Fatalf("alias fix did not converge: %v", after)
	}
}

// Go's shared tester registers this rule as "test". These cases use that name
// in directives, and preserve the v8.72.2 messages, ranges and fix passes.
func TestNoDuplicateTypeConstituentsSuppressedFixes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateTypeConstituentsRule, nil, []rule_tester.InvalidTestCase{
		{Code: "type T = string | number | (\n// eslint-disable-next-line test\nstring |\nnumber\n);", Output: []string{"type T = string | number | (\n// eslint-disable-next-line test\nstring \n\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with number.", Line: 4, Column: 1, EndLine: 4, EndColumn: 7},
		}},
		{Code: "type T = string | number | (\nstring |\n// eslint-disable-next-line test\nnumber\n);", Output: []string{"type T = string | number | (\n \n// eslint-disable-next-line test\nnumber\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 2, Column: 1, EndLine: 2, EndColumn: 7},
		}},
		{Code: "type T = string | number | (|\n// eslint-disable-next-line test\nstring |\nnumber\n);", Output: []string{"type T = string | number | (|\n// eslint-disable-next-line test\nstring \n\n);", "type T = string | number   \n\n;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with number.", Line: 4, Column: 1, EndLine: 4, EndColumn: 7},
		}},
		{Code: "type T = string | number | (|\nstring |\n// eslint-disable-next-line test\nnumber\n);", Output: []string{"type T = string | number | (\n |\n// eslint-disable-next-line test\nnumber\n);", "type T = string | number  \n \n;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 2, Column: 1, EndLine: 2, EndColumn: 7},
		}},
		{Code: "type T = {a: string} & {b: number} & (\n// eslint-disable-next-line test\n{a: string} &\n{b: number}\n);", Output: []string{"type T = {a: string} & {b: number} & (\n// eslint-disable-next-line test\n{a: string} \n\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with {b: number}.", Line: 4, Column: 1, EndLine: 4, EndColumn: 12},
		}},
		{Code: "type T = {a: string} & {b: number} & (\n{a: string} &\n// eslint-disable-next-line test\n{b: number}\n);", Output: []string{"type T = {a: string} & {b: number} & (\n \n// eslint-disable-next-line test\n{b: number}\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with {a: string}.", Line: 2, Column: 1, EndLine: 2, EndColumn: 12},
		}},
		{Code: "type T = {a: string} & {b: number} & (&\n// eslint-disable-next-line test\n{a: string} &\n{b: number}\n);", Output: []string{"type T = {a: string} & {b: number} & (&\n// eslint-disable-next-line test\n{a: string} \n\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with {b: number}.", Line: 4, Column: 1, EndLine: 4, EndColumn: 12},
		}},
		{Code: "type T = {a: string} & {b: number} & (&\n{a: string} &\n// eslint-disable-next-line test\n{b: number}\n);", Output: []string{"type T = {a: string} & {b: number} & (\n &\n// eslint-disable-next-line test\n{b: number}\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with {a: string}.", Line: 2, Column: 1, EndLine: 2, EndColumn: 12},
		}},
		{Code: "function f(x?: (\n// eslint-disable-next-line test\nundefined |\nundefined\n)) {}", Output: []string{"function f(x?: (\n// eslint-disable-next-line test\nundefined \n\n)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with undefined.", Line: 4, Column: 1, EndLine: 4, EndColumn: 10},
		}},
		{Code: "function f(x?: (\nundefined |\n// eslint-disable-next-line test\nundefined\n)) {}", Output: []string{"function f(x?: (\n \n// eslint-disable-next-line test\nundefined\n)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 2, Column: 1, EndLine: 2, EndColumn: 10},
		}},
	})
}
