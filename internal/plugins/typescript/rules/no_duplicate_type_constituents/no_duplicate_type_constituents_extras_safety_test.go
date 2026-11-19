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
	"github.com/web-infra-dev/rslint/internal/rules/no_debugger"
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
			Output: []string{"type T = string | number | (  | number);", "type T = string | number     ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 31, EndLine: 1, EndColumn: 37},
				{MessageId: "duplicate", Message: "Union type constituent is duplicated with number.", Line: 1, Column: 40, EndLine: 1, EndColumn: 46},
			},
		},
		{
			Code:   "type T = {a: string} & {b: number} & (& {a: string} & {b: number});",
			Output: []string{"type T = {a: string} & {b: number} & (  & {b: number});", "type T = {a: string} & {b: number}     ;"},
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

// Go's shared tester registers this rule as "test". Messages and ranges match
// v8.72.2; singleton fixes deliberately retain directive comments.
func TestNoDuplicateTypeConstituentsSuppressedFixes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateTypeConstituentsRule, nil, []rule_tester.InvalidTestCase{
		{Code: "type T = string | number | (\n// eslint-disable-next-line test\nstring |\nnumber\n);", Output: []string{"type T = string | number | (\n// eslint-disable-next-line test\nstring \n\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with number.", Line: 4, Column: 1, EndLine: 4, EndColumn: 7},
		}},
		{Code: "type T = string | number | (\nstring |\n// eslint-disable-next-line test\nnumber\n);", Output: []string{"type T = string | number | (\n \n// eslint-disable-next-line test\nnumber\n);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 2, Column: 1, EndLine: 2, EndColumn: 7},
		}},
		{Code: "type T = string | number | (|\n// eslint-disable-next-line test\nstring |\nnumber\n);", Output: []string{"type T = string | number | (|\n// eslint-disable-next-line test\nstring \n\n);", "type T = string | number  \n// eslint-disable-next-line test\n \n\n;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with number.", Line: 4, Column: 1, EndLine: 4, EndColumn: 7},
		}},
		{Code: "type T = string | number | (|\nstring |\n// eslint-disable-next-line test\nnumber\n);", Output: []string{"type T = string | number | (\n |\n// eslint-disable-next-line test\nnumber\n);", "type T = string | number  \n \n// eslint-disable-next-line test\n\n;"}, Errors: []rule_tester.InvalidTestCaseError{
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

func TestNoDuplicateTypeConstituentsUniqueSymbols(t *testing.T) {
	for _, code := range []string{
		"type Token = { readonly tag: unique symbol } & { readonly tag: unique symbol }; declare const tag: Token['tag']; const value: never = tag;",
		"type Token = { readonly tag: unique symbol } | { readonly tag: unique symbol };",
		"type Token = {inner: { readonly tag: unique symbol }} & {inner: { readonly tag: unique symbol }}; declare const tag: Token['inner']['tag']; const value: never = tag;",
		"type Token = (() => { readonly tag: unique symbol }) | (() => { readonly tag: unique symbol });",
	} {
		if diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll); len(diagnostics) != 0 {
			t.Fatalf("distinct unique-symbol declarations reported as duplicates: %v", diagnostics)
		}
	}
	for _, code := range []string{
		"declare const token: unique symbol; type T = {tag: typeof token} & {tag: typeof token};",
		"type Brand = { readonly tag: unique symbol }; type T = Brand & Brand;",
		"type Brand = { readonly tag: unique symbol }; type Alias = Brand; type T = Brand | Alias;",
		"declare const token: unique symbol; type Alias = typeof token; type T = Alias | typeof token;",
	} {
		diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll)
		if len(diagnostics) != 1 || diagnostics[0].Message.Id != "duplicate" {
			t.Fatalf("same declared symbol should remain fixable: %v", diagnostics)
		}
		fixed, _, changed := linter.ApplyRuleFixes(code, diagnostics)
		if !changed {
			t.Fatal("same-identity duplicate has no fix")
		}
		if after := lintSafetyFixture(t, fixed, rule.EditDemandAll); len(after) != 0 {
			t.Fatalf("symbol-reference fix did not converge: %v", after)
		}
	}
}

func TestNoDuplicateTypeConstituentsReservationIdentity(t *testing.T) {
	for _, code := range []string{
		"function f(x?: string | (| undefined) | undefined) {}",
		"function f(x?: string | (| (| undefined)) | undefined) {}",
		"function f(x?: string | (| undefined) | (| undefined) | undefined) {}",
	} {
		diagnostics := lintSafetyFixture(t, code, rule.EditDemandAll)
		if len(diagnostics) < 3 {
			t.Fatal("fixture must report both a group and its member")
		}
		for _, diagnostic := range diagnostics {
			if len(diagnostic.Fixes()) == 0 {
				t.Fatalf("reservation counted the same member twice: %v", diagnostic)
			}
			// An editor may apply any single quick fix independently.
			fixed, _, changed := linter.ApplyRuleFixes(code, []rule.RuleDiagnostic{diagnostic})
			if !changed {
				t.Fatal("quick fix did not apply")
			}
			lintSafetyFixture(t, fixed, rule.EditDemandNone)
		}
		if !reflect.DeepEqual(diagnostics[0].Fixes(), diagnostics[1].Fixes()) {
			t.Fatal("group and child should reserve the same edit")
		}
		fixed, _, changed := linter.ApplyRuleFixes(code, diagnostics)
		if !changed {
			t.Fatal("fix-all did not apply")
		}
		if after := lintSafetyFixture(t, fixed, rule.EditDemandAll); len(after) != 0 {
			t.Fatalf("fix-all left unnecessary constituents: %v", after)
		}
	}
}

// Run both rules through the production dispatcher so inline directives and
// their effect on another rule are observed on every rebuilt source snapshot.
func lintDirectiveFixture(t *testing.T, code string) []rule.RuleDiagnostic {
	t.Helper()
	return lintDirectiveFixtureRules(t, code, []*rule.Rule{&NoDuplicateTypeConstituentsRule, &no_debugger.NoDebuggerRule})
}

func lintDirectiveFixtureRules(t testing.TB, code string, rules []*rule.Rule) []rule.RuleDiagnostic {
	t.Helper()
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, source, err := helper.CreateTestProgram(code, "directives.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics := program.GetSyntacticDiagnostics(t.Context(), source); len(diagnostics) != 0 {
		t.Fatalf("invalid fixed syntax: %v\n%s", diagnostics, code)
	}
	plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{lintprogram.NewFromCompiler(program)},
		TargetsByProgram: [][]string{{source.FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			var configured []rule.ConfiguredRule
			for _, r := range rules {
				configured = append(configured, rule.ConfiguredRule{Name: r.Name, Environment: &rule.RuleEnvironment{}, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, nil) }})
			}
			return configured
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []rule.RuleDiagnostic
	_, err = linter.RunLinter(linter.RunLinterOptions{SingleThreaded: true, LintPlan: plan, Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }}})
	if err != nil {
		t.Fatal(err)
	}
	return diagnostics
}

func TestNoDuplicateTypeConstituentsDirectivePreservation(t *testing.T) {
	for _, prefix := range []string{"eslint", "rslint"} {
		for _, action := range []string{"enable", "disable"} {
			for _, member := range []string{
				"(| /* DIRECTIVE */ {a: string})",
				"(| (| /* DIRECTIVE */ {a: string}))",
				"(/* DIRECTIVE */ {a: string})",
			} {
				directive := prefix + "-" + action + " no-debugger"
				code := "/* " + prefix + "-disable no-debugger */\ntype T = {a: string} | " + strings.ReplaceAll(member, "DIRECTIVE", directive) + ";\ndebugger;"
				if action == "disable" {
					code = strings.Replace(code, prefix+"-disable", prefix+"-enable", 1)
				}
				wantDebugger := 0
				if action == "enable" {
					wantDebugger = 1
				}
				for pass := range 10 {
					diagnostics := lintDirectiveFixture(t, code)
					debuggerCount := 0
					for _, diagnostic := range diagnostics {
						if diagnostic.RuleName == "no-debugger" {
							debuggerCount++
						}
					}
					if debuggerCount != wantDebugger {
						t.Fatalf("fix changed directive effect: %s", code)
					}
					fixed, _, changed := linter.ApplyRuleFixes(code, diagnostics)
					if !changed {
						if pass == 0 {
							t.Fatal("singleton duplicate was not fixed")
						}
						if !strings.Contains(code, directive) {
							t.Fatal("directive was deleted")
						}
						break
					}
					code = fixed
					if pass == 9 {
						t.Fatal("directive-preserving fixes did not converge")
					}
				}
			}
		}
	}
	// A directive inside the actual type body is not removable safely. This
	// guard also covers duplicates that are not singleton groups.
	code := "/* eslint-disable no-debugger */\ntype T = {a: string} | { /* eslint-enable no-debugger */ a: string };\ndebugger;"
	diagnostics := lintDirectiveFixture(t, code)
	if len(diagnostics) != 2 {
		t.Fatalf("expected duplicate and debugger diagnostics: %v", diagnostics)
	}
	if _, _, changed := linter.ApplyRuleFixes(code, diagnostics); changed {
		t.Fatal("unsafe directive-bearing body received a fix")
	}
	// Directive-looking text inside a string is not a comment.
	code = "/* eslint-disable no-debugger */\ntype T = {a: '/* eslint-enable no-debugger */'} | {a: '/* eslint-enable no-debugger */'};\ndebugger;"
	diagnostics = lintDirectiveFixture(t, code)
	if len(diagnostics) != 1 {
		t.Fatalf("unexpected diagnostics for directive-looking string: %v", diagnostics)
	}
	fixed, _, changed := linter.ApplyRuleFixes(code, diagnostics)
	if !changed {
		t.Fatal("ordinary string prevented a safe fix")
	}
	if after := lintDirectiveFixture(t, fixed); len(after) != 0 {
		t.Fatalf("string fix changed directive behavior: %v", after)
	}
}
