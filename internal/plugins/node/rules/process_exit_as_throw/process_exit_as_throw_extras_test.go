package process_exit_as_throw_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/process_exit_as_throw"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/rules_of_hooks"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/array_callback_return"
	"github.com/web-infra-dev/rslint/internal/rules/consistent_return"
	"github.com/web-infra-dev/rslint/internal/rules/constructor_super"
	"github.com/web-infra-dev/rslint/internal/rules/getter_return"
	"github.com/web-infra-dev/rslint/internal/rules/no_fallthrough"
	"github.com/web-infra-dev/rslint/internal/rules/no_this_before_super"
	"github.com/web-infra-dev/rslint/internal/rules/no_unreachable"
	"github.com/web-infra-dev/rslint/internal/rules/no_unreachable_loop"
	"github.com/web-infra-dev/rslint/internal/rules/no_useless_return"
)

// Reference-checked shapes: direct/non-computed matching, transparent parentheses
// and JSDoc, optional/short-circuit continuations, nested roots, JSX, catch/finally,
// grouped/hoisted statements and UTF-16 ranges. Authored TS wrappers stay visible.
func TestProcessExitAsThrowExpressions(t *testing.T) {
	consumer := withProcessExit(no_unreachable.NoUnreachableRule)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
		[]rule_tester.ValidTestCase{
			{Code: "process.exit?.(process.exit(1)); after();"},
			{Code: "process[\"exit\"](1);\nbar();"},
			{Code: "process[exit](1);\nbar();"},
			{Code: "Process.exit(1);\nbar();"},
			{Code: "process.Exit(1);\nbar();"},
			{Code: "new process.exit();\nbar();"},
			{Code: "getProcess().exit();\nbar();"},
			{Code: "foo.process.exit();\nbar();"},
			{Code: "process?.exit(1);\nbar();"},
			{Code: "process.exit?.(1);\nbar();"},
			{Code: "(process?.exit)(1);\nbar();"},
			{Code: "(process?.exit)?.(1);\nbar();"},
			{Code: "flag && process.exit(1);\nbar();"},
			{Code: "flag || process.exit(1);\nbar();"},
			{Code: "flag ?? process.exit(1);\nbar();"},
			{Code: "flag ? process.exit(1) : foo();\nbar();"},
			{Code: "const f = () => process.exit(1);\nbar();"},
			{Code: "class C { #exit() {} f(process) { process.#exit(1); bar(); } };\nbar();"},
			{Code: "(process as any).exit(1);\nbar();"},
			{Code: "(process.exit as any)(1);\nbar();"},
			{Code: "process!.exit(1);\nbar();"},
			{Code: "process.exit!();\nbar();"},
			{Code: "(process.exit satisfies Function)(1);\nbar();"},
		}, []rule_tester.InvalidTestCase{
			// A leading function must not select its body as the program's CFG.
			{Code: "function f() {}\nfoo();\nprocess.exit(1);\nbar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(4, 1, 4, 7)}},
			// The shared CFG preserves the throwing iterable's completion. The
			// upstream ESLint 7 analyzer instead revives this for-of body.
			{Code: "for (const x of process.exit()) { after(); }", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 33, 1, 45)}},
			{Code: "while (process.exit()) { after(); }", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 24, 1, 36)}},
			{Code: "for (process.exit();;) { after(); }", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 24, 1, 36)}},
			{Code: "switch (process.exit()) { case 1: after(); break; default: other(); }", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 35, 1, 50), unreachableAt(1, 60, 1, 68)}},
			{Code: "function f({[process.exit()]: x}) { after(); }", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 35, 1, 47)}},
			{Code: "if (flag) process.exit(); else { for(process.exit();;) { after(); } }", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 56, 1, 68)}},
			{Code: "function f() { try { process.exit(); } finally { return; } after(); }", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 60, 1, 68)}},
			{Code: "process.exit();\nbar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "(process).exit(1);\nbar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "(process.exit)(1);\nbar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "flag ? process.exit(1) : process.exit(2);\nbar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "foo(process.exit(1));\nbar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "(foo(), process.exit(1));\nbar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "await process.exit(1);\nbar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "const element = <process.exit code={process.exit(1)} />;\nbar();", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "/** @type {any} */ (process).exit(1);\nbar();", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(2, 1, 2, 7)}},
			{Code: "if (process.exit(1)) { bar(); }", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 22, 1, 32)}},
			{Code: "function f() { process.exit(1); bar(); } baz();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 33, 1, 39)}},
			{Code: "try { process.exit(1); bar(); } catch(e) { recover(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 24, 1, 30)}},
			{Code: "try { process.exit(1); } finally { cleanup(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 49, 1, 57)}},
			{Code: "if (x) process.exit(1); else process.exit(2); after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 47, 1, 55)}},
			{Code: "process.exit(1); var x; function f() {} const y = 1; bar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 41, 1, 60)}},
			{Code: "\"😀\"; process.exit(\n 1\n); bar();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(3, 4, 3, 10)}},
		})
}

func TestProcessExitAsThrowConsumers(t *testing.T) {
	t.Run("no diagnostics of its own", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &process_exit_as_throw.ProcessExitAsThrowRule,
			[]rule_tester.ValidTestCase{{Code: "process.exit(1); after();"}}, nil)
	})
	t.Run("no-fallthrough", func(t *testing.T) {
		consumer := withProcessExit(no_fallthrough.NoFallthroughRule)
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
			[]rule_tester.ValidTestCase{
				{Code: "switch(x) { case 1: process.exit(1); case 2: work(); }"},
				{Code: "switch(x) { case 1: try { process.exit(1); } finally { cleanup(); } case 2: work(); }"},
				// Documented difference: dispatch exits before reaching the leading default.
				{Code: "switch (x) { default: work(); case process.exit(): more(); }"},
			}, []rule_tester.InvalidTestCase{{
				Code:   "switch(x) {\ncase 1: flag && process.exit(1);\ncase 2: work();\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "case", Message: "Expected a 'break' statement before 'case'.", Line: 3, Column: 1, EndLine: 3, EndColumn: 16}},
			}})
	})
	t.Run("getter-return", func(t *testing.T) {
		consumer := withProcessExit(getter_return.GetterReturnRule)
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
			[]rule_tester.ValidTestCase{{Code: "const obj = { get value() { process.exit(1); } };"}}, nil)
	})
	t.Run("no-useless-return", func(t *testing.T) {
		consumer := withProcessExit(no_useless_return.NoUselessReturnRule)
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
			[]rule_tester.ValidTestCase{{Code: "function f() { process.exit(1); return; }"}}, nil)
	})
	t.Run("no-unreachable-loop", func(t *testing.T) {
		consumer := withProcessExit(no_unreachable_loop.NoUnreachableLoopRule)
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer, nil,
			[]rule_tester.InvalidTestCase{{Code: "function f() { while(x) { process.exit(1); } }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalid", Message: "Invalid loop. Its body allows only one iteration.", Line: 1, Column: 16, EndLine: 1, EndColumn: 45}},
			}})
	})
	t.Run("consistent-return with shadowing and finally", func(t *testing.T) {
		consumer := withProcessExit(consistent_return.ConsistentReturnRule)
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
			[]rule_tester.ValidTestCase{
				{Code: "function f(process) { if (x) return 1; process.exit(1); }"},
				{Code: "function f() { if (x) return 1; try { process.exit(1); } finally { cleanup(); } }"},
			}, nil)
	})
	t.Run("rules-of-hooks excludes thrown paths", func(t *testing.T) {
		consumer := withProcessExit(rules_of_hooks.RulesOfHooksRule)
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
			[]rule_tester.ValidTestCase{{Code: "function Component() { if (bad) process.exit(1); useState(); }"}}, nil)
	})
}

func TestProcessExitAsThrowIndependentAnalyzers(t *testing.T) {
	// These existing rules do not consume the shared CFG. Keep the upstream
	// expectations visible until their own analyses support call completion;
	// this is a documented framework gap, not a process.exit matching exception.
	for _, testCase := range []struct {
		consumer rule.Rule
		code     string
	}{
		{array_callback_return.ArrayCallbackReturnRule, "items.map(() => { process.exit(1); });"},
		{constructor_super.ConstructorSuperRule, "class C extends Base { constructor() { process.exit(1); } }"},
		{no_this_before_super.NoThisBeforeSuperRule, "class C extends Base { constructor() { process.exit(1); this.value = 1; } }"},
	} {
		t.Run(testCase.consumer.Name, func(t *testing.T) {
			consumer := withProcessExit(testCase.consumer)
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
				[]rule_tester.ValidTestCase{
					{Code: testCase.code, Skip: true},
					{Code: strings.Replace(testCase.code, "process.exit(1);", "return process.exit(1);", 1)},
				}, nil)
		})
	}
}

// Additional reference-checked evaluation order, escaped identifiers, class
// members and ranges spanning nested statements. Differences are marked below.
func TestProcessExitAsThrowEvaluationAndRanges(t *testing.T) {
	consumer := withProcessExit(no_unreachable.NoUnreachableRule)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
		[]rule_tester.ValidTestCase{
			{Code: "foo?.(process.exit()); after();"},
			{Code: "value &&= process.exit(); after();"},
			{Code: "const { value = process.exit() } = obj; after();"},
			// Class members keep separate code paths; see the documented ESLint 7 difference.
			{Code: "class C { value = process.exit() }; after();"},
			{Code: "class C { static value = process.exit() }; after();"},
			{Code: "function f(x = process.exit()) { work(); } after();"},
		}, []rule_tester.InvalidTestCase{
			// Match the decoded identifier names, independent of their source spelling.
			{Code: "\\u0070\\u0072\\u006f\\u0063\\u0065\\u0073\\u0073.\\u0065\\u0078\\u0069\\u0074(1); after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 73, 1, 81)}},
			{Code: "(process.exit())?.value; after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 26, 1, 34)}},
			{Code: "(foo?.bar)(process.exit()); after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 29, 1, 37)}},
			{Code: "const {[process.exit()]: value} = obj; after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 40, 1, 48)}},
			{Code: "const value = { [process.exit()]() {} }; after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 42, 1, 50)}},
			{Code: "class C { [process.exit()]() {} }; after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 36, 1, 44)}},
			{Code: "class C extends process.exit() {}; after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 36, 1, 44)}},
			// The static block exit affects its body, not the enclosing class declaration.
			{Code: "class C { static { process.exit(); work(); } }; after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 36, 1, 43)}},
			{Code: "const value = /** @satisfies {any} */ (process).exit(); after();", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 57, 1, 65)}},
			{Code: "function f({[process.exit()]: x} = {}) { work(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 40, 1, 51)}},
			{Code: "if (process.exit()) {} else {} after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 21, 1, 23), unreachableAt(1, 29, 1, 40)}},
			{Code: "if (process.exit()) work(); else other(); after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 21, 1, 28), unreachableAt(1, 34, 1, 51)}},
			{Code: "while (process.exit()) work(); after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 24, 1, 40)}},
			{Code: "for (;process.exit();) { work(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 24, 1, 44)}},
			// Keep throwing loop sources unreachable; ESLint 7 revives these bodies.
			{Code: "for (const x in process.exit()) { work(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 33, 1, 53)}},
			{Code: "for (const x of process.exit()) { work(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 33, 1, 53)}},
			// Dispatch must evaluate the case test before entering the leading default.
			{Code: "switch (x) { default: work(); case process.exit(): more(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 23, 1, 30), unreachableAt(1, 52, 1, 59), unreachableAt(1, 62, 1, 70)}},
			{Code: "try { work(); } catch ({[process.exit()]: error}) { recover(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 51, 1, 65)}},
			{Code: "process.exit(); function f({[process.exit()]:x}) { work(); } after();", Errors: []rule_tester.InvalidTestCaseError{unreachableAt(1, 50, 1, 70)}},
		})
}
