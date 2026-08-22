package prefer_todo_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_todo"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTodoExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_todo.PreferTodoRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: overload matrix - ambiguous two-arg second parameter.
			{Code: `declare const options: any; test("case", options);`},
			// ---- Dimension 4: overload matrix - ambiguous three-arg second parameter.
			{Code: `declare const options: any; test("case", options, () => {});`},
			{Code: `declare const callback: any; test("case", callback);`},
			{Code: `declare const callback: any; test("case", callback, 1000);`},
			{Code: `test("case", { timeout: 1000 }, callback);`},
			{Code: `test("case", () => 1);`},
			{Code: `test("case", async () => await work());`},
			{Code: `test("case", function* () { yield 1; });`},
			// ---- Dimension 4: modifier matrix / fail-closed.
			{Code: `test.todo("case");`},
			{Code: `test.fails("case");`},
			{Code: `test.fail("case");`, FileName: "file.ts"},
			{Code: `test.only("case");`},
			{Code: `test.concurrent("case");`},
			{Code: `test.sequential("case");`},
			{Code: `test.runIf(condition)("case");`},
			{Code: `test.skipIf(condition)("case");`},
			{Code: `test.each([1])("case");`},
			{Code: `test.for([1])("case");`},
			{Code: `test.extend(fixtures)("case");`},
			// ---- Same-file alias fail-closed rows from the spec.
			{Code: `const failing = test.fails; failing("case");`},
			{Code: `const eachCase = test.each([1]); eachCase("case");`},
			{Code: `const todoCase = test.todo; todoCase("case");`},
			{Code: `const extended = test.extend(fixtures); extended("case");`},
			{Code: `const focused = test.only; focused.skip("case", () => {});`},
			{Code: `const failing = test.fails; failing.skip("case", () => {});`},
			{Code: `const ext = test.extend(fixtures); ext.skip("case", () => {});`},
			// ---- API sources and reverse sources.
			{Code: `import { test } from "node:test"; test("case");`},
			{Code: `import { test } from "vitest"; test("case");`},
			{Code: `import { test } from "@playwright/test"; test("case");`},
			{Code: `const foreign = { test() {} }; foreign.test("case");`},
			{Code: `const foreign = { test() {} }; const case_ = foreign.test; case_("case");`},
			{Code: `import * as foreign from "node:test"; foreign.test("case");`},
			{Code: `import * as pw from "@playwright/test"; pw.test("case");`},
			// Whole-module require objects have mutable API properties. Even a const
			// binding cannot prove that `.test` still refers to Rstest at this call.
			{Code: `const core = require("@rstest/core"); core.test("case");`},
			{Code: `const core = require("@rstest/core"); core.test = custom; core.test("case");`},
			{Code: `const playwright = require("@rstest/playwright"); playwright.test("case");`},
			{Code: `const test = createRunner(); test("case");`},
			// The computed member is not statically known, so the run mode is
			// unknown too: it may be `.only` or `.each`, neither of which this rule
			// reports. Nothing is reported rather than guess at the mode.
			{Code: `test[skip]("case", () => {});`},
			// ---- `import.meta` at the root of a longer chain is a different object.
			{Code: `let api = import.meta.rstest; api = other; api.test("case");`},
			{Code: `import.meta.env.rstest.test("case");`},
			{Code: `import.meta.glob("x").rstest.test("case");`},
			// ---- A destructure has to read the namespace's own `test`/`it`.
			{Code: `const { runner: { test } } = import.meta.rstest; test("case");`},
			{Code: `const { test = fallback } = import.meta.rstest; test("case");`},
			{Code: `const { ...rest } = import.meta.rstest; rest.test("case");`},
			{Code: `const { skip: test } = import.meta.rstest; test("case");`},
			{Code: `let { test } = import.meta.rstest; test("case");`},
			{Code: `const core = require("@rstest/core"); const { test } = core; test("case");`},
			// Destructuring an ES namespace import binding is not resolved by the
			// shared parser, so no rstest rule sees this as a registration.
			{Code: `import * as rstest from "@rstest/core";
const { test } = rstest;
test("case");`},
			// ---- Malformed / graceful degradation.
			{Code: `test("case", () => {}, () => {});`},
			{Code: `test("case", {}, () => {}, 1);`},
			{Code: `test?.("case");`},
			{Code: `(test?.skip)("case", () => {});`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- A registration the parser has proved is reported even when no
			// rewrite delivers a todo. `rstest/no-disabled-tests` already tells the
			// user these tests have no body; staying silent here would be the only
			// rule in the plugin that does not.
			{
				Code:   `const skipped = test.skip; skipped("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 28}},
			},
			{
				Code:   `test.skip.skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			{
				Code:   `test["skip"].skip("case", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			// ---- source forms: `import.meta.rstest` object destructuring.
			{
				Code: `const { test } = import.meta.rstest;
test("case");`,
				Output: []string{`const { test } = import.meta.rstest;
test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 2, Column: 1}},
			},
			{
				Code: `const { test } = import.meta.rstest;
test("case", () => {});`,
				Output: []string{`const { test } = import.meta.rstest;
test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 2, Column: 1}},
			},
			{
				Code: `const { it } = import.meta.rstest;
it("case");`,
				Output: []string{`const { it } = import.meta.rstest;
it.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 2, Column: 1}},
			},
			{
				Code: `const { test: case_ } = import.meta.rstest;
case_("case");`,
				Output: []string{`const { test: case_ } = import.meta.rstest;
case_.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 2, Column: 1}},
			},
			{
				Code: `const { test } = import.meta.rstest;
test.skip("case", () => {});`,
				Output: []string{`const { test } = import.meta.rstest;
test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 2, Column: 1}},
			},
			// ---- overload matrix: one-arg missing.
			{
				Code:   `test("case");`,
				Output: []string{`test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 1, Column: 1}},
			},
			// ---- overload matrix: missing callback with object-literal options retained.
			{
				Code:   `test("case", { timeout: 1000 });`,
				Output: []string{`test.todo("case", { timeout: 1000 });`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 1, Column: 1}},
			},
			// ---- overload matrix: empty callback in function-first overload deletes callback+timeout.
			{
				Code:   `test("case", () => {}, 1_000);`,
				Output: []string{`test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			{
				Code:   `test("case", async () => {}, 1_000);`,
				Output: []string{`test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			{
				Code:   `test("case", function* () {}, 1_000);`,
				Output: []string{`test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			// ---- overload matrix: options overload preserves options.
			{
				Code:   `test("case", { retry: 2, meta }, () => {});`,
				Output: []string{`test.todo("case", { retry: 2, meta });`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			// ---- comment-only body is still empty.
			{
				Code:   `test("case", () => { /* TODO */ });`,
				Output: []string{`test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			// ---- source forms: import rename.
			{
				Code: `import { test as case_ } from "@rstest/core";
case_("case");`,
				Output: []string{`import { test as case_ } from "@rstest/core";
case_.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 2, Column: 1}},
			},
			// ---- source forms: require destructuring.
			{
				Code: `const { test } = require("@rstest/core");
test("case");`,
				Output: []string{`const { test } = require("@rstest/core");
test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 2, Column: 1}},
			},
			// ---- source forms: namespace/module object.
			{
				Code: `import * as rstest from "@rstest/core";
rstest.test("case");`,
				Output: []string{`import * as rstest from "@rstest/core";
rstest.test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 2, Column: 1}},
			},
			// ---- source forms: import.meta.rstest object alias.
			{
				Code: `const api = import.meta.rstest;
api.test("case");`,
				Output: []string{`const api = import.meta.rstest;
api.test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 2, Column: 1}},
			},
			// ---- same-file plain alias allowed only when every hop is bare.
			{
				Code: `const base = test;
const case_ = base;
case_("case");`,
				Output: []string{`const base = test;
const case_ = base;
case_.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 3, Column: 1}},
			},
			// ---- skip accessor replacement keeps quotes/trivia.
			{
				Code:   `test['skip']("case", () => {});`,
				Output: []string{`test['todo']("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			{
				Code:   `test["skip"]("case", () => {});`,
				Output: []string{`test["todo"]("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			{
				Code:   "test[`skip`](\"case\", () => {});",
				Output: []string{"test[`todo`](\"case\");"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			{
				Code: `import { test as case_ } from "@rstest/core";
case_.skip("case", () => {});`,
				Output: []string{`import { test as case_ } from "@rstest/core";
case_.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 2, Column: 1}},
			},
			{
				Code: `const api = import.meta.rstest;
api.test.skip("case", () => {});`,
				Output: []string{`const api = import.meta.rstest;
api.test.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 2, Column: 1}},
			},
			{
				Code: `test
  . /* keep */ skip("case", () => {});`,
				Output: []string{`test
  . /* keep */ todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 1, Column: 1}},
			},
			// ---- parenthesized callee remains parseable.
			{
				Code:   `(test)("case");`,
				Output: []string{`(test).todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 1, Column: 1}},
			},
			// ---- real-user shape: options overload from migrated suites.
			{
				Code: `describe("suite", () => {
  test("needs body", { timeout: 100 });
});`,
				Output: []string{`describe("suite", () => {
  test.todo("needs body", { timeout: 100 });
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unimplementedTest", Line: 2, Column: 3}},
			},
			// ---- real-user shape: alias imported as namespace helper.
			{
				Code: `import * as core from "@rstest/core";
const case_ = core.test;
case_("case", () => {});`,
				Output: []string{`import * as core from "@rstest/core";
const case_ = core.test;
case_.todo("case");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTest", Line: 3, Column: 1}},
			},
		},
	)
}

// TestPreferTodoEditDemand exercises Dimension 3: diagnostic identity stays
// fixed across edit-demand modes, and the rule materializes only autofixes.
func TestPreferTodoEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`test("bare");
test.skip("skip", () => {});
test("options", { timeout: 100 }, () => {});`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      lintprogram.NewFromCompiler(program),
			File:         sourceFile.FileName(),
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     prefer_todo.PreferTodoRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_todo.PreferTodoRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: diagnostics = %d, want 3", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index := range allEdits {
		want := withoutEdits(allEdits[index])
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got := withoutEdits(diagnostics[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d diagnostic %d changed:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
	}

	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, suggestionOnly} {
		for index, diagnostic := range diagnostics {
			if diagnostic.FixesPtr != nil {
				t.Errorf("non-autofix diagnostic %d materialized fixes", index)
			}
			if diagnostic.Suggestions != nil {
				t.Errorf("diagnostic %d unexpectedly materialized suggestions", index)
			}
		}
	}
	for index, diagnostic := range autofixOnly {
		if diagnostic.FixesPtr == nil {
			t.Errorf("autofix diagnostic %d did not materialize fixes", index)
			continue
		}
		if !reflect.DeepEqual(*diagnostic.FixesPtr, *allEdits[index].FixesPtr) {
			t.Errorf("autofix and all-edits diagnostic %d produced different fixes", index)
		}
		if diagnostic.Suggestions != nil {
			t.Errorf("autofix diagnostic %d unexpectedly materialized suggestions", index)
		}
	}
}
