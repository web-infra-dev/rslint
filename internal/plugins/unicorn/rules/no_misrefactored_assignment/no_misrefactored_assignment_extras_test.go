package no_misrefactored_assignment_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_misrefactored_assignment"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMisrefactoredAssignmentExtras(t *testing.T) {
	valid := append(jsCases(
		"getValue().x += getValue().x + c;",
		`class A { #x = 0; m() { this.#x += this["#x"] + b; } }`,
		"a += (a, b); a += (a || b); a += (a += b);",
		"obj.a += obj[String.raw`a`] + delta;",
		`o["\uD800"] += o["\uD801"] + x;`,
		// Unlike Unicorn, the rule does not resolve a number's toString() call as a key.
		`counters["1"] += counters[(1).toString()] + delta;`,
	), rule_tester.ValidTestCase{FileName: "case.ts", Code: "a += (a + b) as number;"})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_misrefactored_assignment.NoMisrefactoredAssignmentRule, valid,
		[]rule_tester.InvalidTestCase{
			// The diagnostic includes the target's parentheses; its message does not.
			wholeAssignment("case.js", "(a) += (a) + b;", "a", 1, 15, "(a) += b;"),
			wholeAssignment("case.js", "a.b += a?.b + c;", "a.b", 1, 16, "a.b += c;"),
			{
				FileName: "case.js",
				Code:     "class B extends A { m() { super.x += super.x + 1; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					assignmentError("super.x", 1, 27, 1, 49, "class B extends A { m() { super.x += 1; } }"),
				},
			},
			// Static keys use JS string conversion. String and template receivers
			// with the same decoded value also match.
			wholeAssignment("case.js", `obj["a" + "b"] += obj.ab + delta;`, "obj[\"a\" + \"b\"]", 1, 33, `obj["a" + "b"] += delta;`),
			wholeAssignment("case.js", "obj[1] += obj[1n] + delta;", "obj[1]", 1, 26, "obj[1] += delta;"),
			wholeAssignment("case.js", "\"a\".x += `a`.x + delta;", "\"a\".x", 1, 23, "\"a\".x += delta;"),
			wholeAssignment("case.ts", `a.x += a["x" as string] + b;`, "a.x", 1, 28, "a.x += b;"),
			wholeAssignment("case.js", `o["\u{1F984}"] += o["🦄"] + x;`, `o["\u{1F984}"]`, 1, 30, `o["\u{1F984}"] += x;`),
			wholeAssignment("case.js", `o.a += o["ab".at(0)] + x;`, "o.a", 1, 25, "o.a += x;"),
			wholeAssignment("case.js", `o.é += o["e\u0301".normalize()] + x;`, "o.é", 1, 36, "o.é += x;"),
			wholeAssignment("case.js", `o["1,2,3"] += o[[1, [2, 3]].flat()] + x;`, `o["1,2,3"]`, 1, 40, `o["1,2,3"] += x;`),
			wholeAssignment("case.js", `o.a += o[["a", "b"].slice(0, 1)] + x;`, "o.a", 1, 37, "o.a += x;"),

			// An authored assertion on the whole RHS is significant, while
			// wrappers on the duplicated operand are transparent.
			wholeAssignment("case.ts", "a += (a satisfies number) + b;", "a", 1, 30, "a += b;"),
			wholeAssignment("case.ts", "a.x += (a<number>).x + b;", "a.x", 1, 25, "a.x += b;"),
			wholeAssignment("case.ts", "a += a + (x!);", "a", 1, 14, "a += x!;"),
			wholeAssignment("case.ts", "a += a + (<number>x);", "a", 1, 21, "a += <number>x;"),
			wholeAssignment("case.ts", "a += a + (x<number>);", "a", 1, 21, "a += x<number>;"),
			wholeAssignment("case.ts", "total += total + (x satisfies number);", "total", 1, 38, "total += (x satisfies number);"),
			wholeAssignment("case.js", "a += /** @type {number} */ (a + b);", "a", 1, 35, "a += /** @type {number} */ (b);"),
			wholeAssignment("case.js", "a += a + /** @type {number} */ (b);", "a", 1, 35, ""),

			// Keep every required wrapper, and leave comments outside the
			// replacement range untouched.
			wholeAssignment("case.js", "a += a + ((/* stay */ b || c));", "a", 1, 31, "a += ((/* stay */ b || c));"),
			wholeAssignment("case.js", "a += (a + b /* stay */);", "a", 1, 24, "a += (b /* stay */);"),
			wholeAssignment("case.js", "a+=a+(/x/);", "a", 1, 11, "a+=/x/;"),
			{
				FileName: "case.js",
				Code:     "for (total += total + (key in o); ; ) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					assignmentError("total", 1, 6, 1, 33, "for (total += (key in o); ; ) {}"),
				},
			},
			{
				FileName: "case.tsx",
				Code:     "const node = <p>{(a += a + (<span />))}</p>;",
				Errors: []rule_tester.InvalidTestCaseError{
					assignmentError("a", 1, 19, 1, 38, "const node = <p>{(a += <span />)}</p>;"),
				},
			},
			{
				FileName: "case.js",
				Code:     "const title = \"🦄\";\nconst café = () => (\"🦄\", aé +=\n  aé + (b ?? c));",
				Errors: []rule_tester.InvalidTestCaseError{
					assignmentError("aé", 2, 27, 3, 16, "const title = \"🦄\";\nconst café = () => (\"🦄\", aé +=\n  (b ?? c));"),
				},
			},
		},
	)
}

func TestNoMisrefactoredAssignmentEditDemand(t *testing.T) {
	t.Parallel()
	const source = "a += a + getValue(/* keep */);\na += a + /* would be lost */ b;"
	compiler, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil || file == nil {
		t.Fatalf("creating test program: file %v, error %v", file != nil, err)
	}
	sourceProgram, err := program.NewFromBoundSources(compiler, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	native := no_misrefactored_assignment.NoMisrefactoredAssignmentRule
	lint := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram,
			File:    file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: native.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return native.Run(ctx, nil)
				}}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		})
		return diagnostics
	}

	baseline := lint(rule.EditDemandAll)
	if len(baseline) != 2 || baseline[0].Suggestions == nil || baseline[1].Suggestions != nil {
		t.Fatalf("expected two diagnostics and only the first to have a suggestion, got %+v", baseline)
	}
	if got := *baseline[0].Suggestions; len(got) != 1 || got[0].Message.Description != "Remove the duplicated `a`." {
		t.Fatalf("unexpected suggestion: %+v", got)
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		want := append([]rule.RuleDiagnostic(nil), baseline...)
		if demand&rule.EditDemandSuggestion == 0 {
			want[0].Suggestions = nil
		}
		got := lint(demand)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("demand %v: got %+v, want %+v", demand, got, want)
		}
		for _, diagnostic := range got {
			if diagnostic.FixesPtr != nil {
				t.Errorf("demand %v: unexpected autofix: %+v", demand, diagnostic.FixesPtr)
			}
		}
	}
}
