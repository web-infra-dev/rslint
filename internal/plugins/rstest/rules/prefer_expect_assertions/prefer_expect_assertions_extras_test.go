package prefer_expect_assertions_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_expect_assertions"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// missing expects haveExpectAssertions on call, the exact source text of a
// test registration that appears once in code. With a spelling, it also
// expects both suggestions, inserted right after the first occurrence of
// marker inside that call.
func missing(code, call, marker, spelling string) rule_tester.InvalidTestCaseError {
	start := strings.Index(code, call)
	line, column := position(code, start)
	endLine, endColumn := position(code, start+len(call))
	err := rule_tester.InvalidTestCaseError{MessageId: "haveExpectAssertions", Line: line, Column: column, EndLine: endLine, EndColumn: endColumn}
	if spelling == "" {
		return err
	}
	at := start + strings.Index(call, marker) + len(marker)
	insert := func(member string) string {
		return code[:at] + spelling + "." + member + "();" + code[at:]
	}
	err.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
		{MessageId: "suggestAddingHasAssertions", Output: insert("hasAssertions")},
		{MessageId: "suggestAddingAssertions", Output: insert("assertions")},
	}
	return err
}

// position converts a byte offset of ASCII source into a 1-based line and
// column.
func position(code string, offset int) (int, int) {
	line := 1 + strings.Count(code[:offset], "\n")
	return line, offset - strings.LastIndex(code[:offset], "\n")
}

func invalid(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, Errors: errors}
}

func TestPreferExpectAssertionsExtras(t *testing.T) {
	const rstestImport = "import { beforeAll, beforeEach, afterEach, afterAll, describe, expect, test } from '@rstest/core';\n"

	// ---- Hooks: Rstest resets assertion state before beforeEach and checks it
	// before afterEach, so only beforeEach sets a requirement for the test. ----
	afterEachOnly := rstestImport + "describe('s', () => { afterEach(() => { expect.hasAssertions(); }); test('t', () => {}); });"
	beforeAllOnly := rstestImport + "beforeAll(() => { expect.hasAssertions(); }); test('t', () => {});"
	afterAllOnly := rstestImport + "afterAll(() => expect.hasAssertions()); test('t', () => {});"
	timerInHook := rstestImport + "beforeEach(() => { setTimeout(() => expect.hasAssertions()); }); test('t', () => {});"
	// Rstest's hasAssertions reads its receiver, so a detached reference throws
	// a TypeError when the hook calls it instead of setting a requirement.
	detachedReference := rstestImport + "beforeEach(expect.hasAssertions); test('t', () => {});"
	siblingSuite := rstestImport + "describe('a', () => { beforeEach(() => expect.hasAssertions()); }); describe('b', () => { test('t', () => {}); });"
	reassignedHook := rstestImport + "let setup = () => expect.hasAssertions(); setup = () => {}; beforeEach(setup); test('t', () => {});"
	foreignHook := rstestImport + "import { beforeEach as vitestBeforeEach } from 'vitest'; vitestBeforeEach(() => expect.hasAssertions()); test('t', () => {});"
	shadowedHookContext := rstestImport + "beforeEach((ctx) => { const other = { expect }; other.expect.hasAssertions(); }); test('t', () => {});"
	argumentInHook := rstestImport + "beforeEach(() => expect.hasAssertions(1)); test('t', () => {});"
	disallowInHook := rstestImport + "beforeEach(() => { expect.hasAssertions(); }); test('t', () => {});"

	// ---- Provenance ----
	renamed := "import { expect as check, test as scenario } from '@rstest/core';\nscenario('t', () => { run(); });"
	namespace := "import * as rs from '@rstest/core';\nrs.test('t', () => { run(); });"
	commonJS := "const { expect: check, test } = require('@rstest/core');\ntest('t', () => { run(); });"
	wholeRequire := "const rs = require('rstack/test');\nrs.test('t', () => { run(); });"
	noExpectImport := "import { test } from '@rstest/core';\ntest('t', () => { run(); });"
	noExpectImportContext := "import { test } from '@rstest/core';\ntest('t', (context) => { run(); });"
	chaiExpect := "import { expect } from 'chai';\nimport { test } from '@rstest/core';\ntest('t', () => { expect.hasAssertions(); });"
	shadowedByInner := rstestImport + "test('t', () => { run(); function expect() {} });"
	shadowedByOuter := rstestImport + "describe('s', (expect) => { test('t', () => { run(); }); });"

	// ---- Callback shapes ----
	optionsOverload := rstestImport + "test('t', { timeout: 100 }, () => { run(); });"
	eachRow := rstestImport + "test.each([1])('t %s', (context) => { run(context); });"
	forContext := rstestImport + "test.for([1])('t %s', (row, context) => { run(row); });"
	concurrent := rstestImport + "test.concurrent('t', async (context) => { await run(); });"

	// ---- First statement ----
	nestedBranch := rstestImport + "test('t', () => { if (ready) { expect.assertions(1); } run(); });"
	directive := rstestImport + "test('t', function () { 'use strict'; run(); });"
	matcherNamedAssertions := rstestImport + "test('t', () => { expect(value).assertions(1); });"

	// ---- Arguments ----
	bracketDisallowed := rstestImport + "test('t', () => { expect['hasAssertions'](); });"

	// ---- Options ----
	testInLoop := rstestImport + "for (const value of values) { test('t', () => { expect(value).toBe(1); }); }"
	whileLoop := rstestImport + "test('t', () => { while (next()) { expect(current()).toBe(1); } });"
	methodCallback := rstestImport + "test('t', () => { register({ handle() { expect(1).toBe(1); } }); });"
	leak := rstestImport + "test('a', () => { expect.assertions(1); expect(1).toBe(1); });\ntest('b', async () => { await run(); });"

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_expect_assertions.PreferExpectAssertionsRule,
		[]rule_tester.ValidTestCase{
			// ---- beforeEach covers every test of its suite, in any position,
			// through inline, concise, named and TestContext callbacks. ----
			{Code: rstestImport + "beforeEach(() => { expect.hasAssertions(); }); test('t', () => {});"},
			{Code: rstestImport + "beforeEach(() => expect.assertions(1)); test('t', () => {});"},
			{Code: rstestImport + "describe('s', () => { test('t', () => {}); beforeEach(() => { expect.hasAssertions(); }); });"},
			{Code: rstestImport + "describe('s', () => { beforeEach(() => expect.hasAssertions()); describe('inner', () => { test('t', () => {}); }); });"},
			{Code: rstestImport + "function requireAssertions() { expect.hasAssertions(); } beforeEach(requireAssertions); test('t', () => {});"},
			{Code: rstestImport + "test('t', () => {}); const setup = () => { expect.hasAssertions(); }; beforeEach(setup);"},
			{Code: rstestImport + "beforeEach(({ expect }) => { expect.hasAssertions(); }); test('t', () => {});"},
			{Code: rstestImport + "beforeEach(({ expect: local }) => local.hasAssertions()); test('t', () => {});"},
			{Code: rstestImport + "beforeEach((context) => { context.expect.hasAssertions(); }); test('t', () => {});"},
			{Code: rstestImport + "describe.each([1])('s %s', () => { beforeEach(() => expect.hasAssertions()); test('t', () => {}); });"},
			{Code: rstestImport + "beforeEach(((() => { expect.hasAssertions(); }) as () => void)); test('t', () => {});"},

			// ---- Provenance ----
			{Code: "import { expect as check, test as scenario } from '@rstest/core';\nscenario('t', () => { check.assertions(1); });"},
			{Code: "import * as rs from '@rstest/core';\nrs.test('t', () => { rs.expect.hasAssertions(); });"},
			{Code: "const { expect, test } = require('@rstest/core');\ntest('t', () => { expect.hasAssertions(); });"},
			{Code: "import { expect, test } from 'rstack/test';\ntest('t', () => { expect.hasAssertions(); });"},
			{Code: "import { test } from '@rstest/core';\ntest('t', () => { import.meta.rstest.expect.hasAssertions(); });"},
			{Code: "import { test } from '@rstest/core';\ntest.for([1])('t', (row, { expect }) => { expect.hasAssertions(); });"},
			{Code: "import { expect, test } from 'vitest';\ntest('t', () => {});"},
			{Code: "import { expect } from '@rstest/core';\nconst test = (name, fn) => fn();\ntest('t', () => {});"},

			// ---- Callback shapes ----
			{Code: rstestImport + "test('t', { timeout: 100 }, () => { expect.hasAssertions(); });"},
			{Code: rstestImport + "test('t', () => { expect.hasAssertions(); }, 100);"},
			{Code: rstestImport + "test.todo('t');"},
			{Code: rstestImport + "const run = () => {}; test('t', run);"},
			{Code: rstestImport + "test.each`a\n${1}`('t', () => { expect.assertions(0); });"},

			// ---- First statement ----
			{Code: rstestImport + "test('t', function () { 'use strict'; expect.hasAssertions(); });"},
			{Code: rstestImport + "test('t', async () => { await expect.hasAssertions(); });"},
			{Code: rstestImport + "test('t', () => { (expect.hasAssertions)(); });"},
			{Code: rstestImport + "test('t', () => { expect['assertions'](1); });"},

			// ---- Arguments that are integer number literals ----
			{Code: rstestImport + "test('t', () => { expect.assertions(1.0); });"},
			{Code: rstestImport + "test('t', () => { expect.assertions(0x1); });"},
			{Code: rstestImport + "test('t', () => { expect.assertions(1e3); });"},
			{Code: rstestImport + "test('t', () => { expect.assertions(1_000); });"},
			{Code: rstestImport + "test('t', () => { expect.assertions((2)); });"},

			// ---- Options ----
			{Code: testInLoop, Options: []interface{}{map[string]interface{}{"onlyFunctionsWithExpectInLoop": true}}},
			{Code: rstestImport + "test('t', () => { expect.hasAssertions(); })", Options: []interface{}{map[string]interface{}{"disallowHasAssertions": false}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Hooks that cannot cover a test ----
			invalid(afterEachOnly, missing(afterEachOnly, "test('t', () => {})", "test('t', () => {", "expect")),
			invalid(beforeAllOnly, missing(beforeAllOnly, "test('t', () => {})", "test('t', () => {", "expect")),
			invalid(afterAllOnly, missing(afterAllOnly, "test('t', () => {})", "test('t', () => {", "expect")),
			invalid(timerInHook, missing(timerInHook, "test('t', () => {})", "test('t', () => {", "expect")),
			invalid(detachedReference, missing(detachedReference, "test('t', () => {})", "test('t', () => {", "expect")),
			invalid(siblingSuite, missing(siblingSuite, "test('t', () => {})", "test('t', () => {", "expect")),
			invalid(reassignedHook, missing(reassignedHook, "test('t', () => {})", "test('t', () => {", "expect")),
			invalid(foreignHook, missing(foreignHook, "test('t', () => {})", "test('t', () => {", "expect")),
			invalid(shadowedHookContext, missing(shadowedHookContext, "test('t', () => {})", "test('t', () => {", "expect")),
			{
				Code: argumentInHook,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "hasAssertionsTakesNoArguments", Line: 2, Column: 25, EndLine: 2, EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestRemovingExtraArguments",
						Output:    strings.Replace(argumentInHook, "hasAssertions(1)", "hasAssertions()", 1),
					}},
				}},
			},
			{
				Code:    disallowInHook,
				Options: []interface{}{map[string]interface{}{"disallowHasAssertions": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferAssertionsOverHasAssertions", Line: 2, Column: 27, EndLine: 2, EndColumn: 40,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestReplacingWithAssertions",
						Output:    strings.Replace(disallowInHook, "hasAssertions", "assertions", 1),
					}},
				}},
			},

			// ---- Suggestions spell expect the way the file reaches it ----
			invalid(renamed, missing(renamed, "scenario('t', () => { run(); })", "() => {", "check")),
			invalid(namespace, missing(namespace, "rs.test('t', () => { run(); })", "() => {", "rs.expect")),
			invalid(commonJS, missing(commonJS, "test('t', () => { run(); })", "() => {", "check")),
			invalid(wholeRequire, missing(wholeRequire, "rs.test('t', () => { run(); })", "() => {", "rs.expect")),
			// The file imports from Rstest but not expect, so a bare expect may
			// not exist at runtime.
			invalid(noExpectImport, missing(noExpectImport, "test('t', () => { run(); })", "", "")),
			invalid(noExpectImportContext, missing(noExpectImportContext, "test('t', (context) => { run(); })", "(context) => {", "context.expect")),
			invalid(chaiExpect, missing(chaiExpect, "test('t', () => { expect.hasAssertions(); })", "", "")),
			invalid(shadowedByInner, missing(shadowedByInner, "test('t', () => { run(); function expect() {} })", "", "")),
			invalid(shadowedByOuter, missing(shadowedByOuter, "test('t', () => { run(); })", "", "")),

			// ---- Callback shapes ----
			invalid(optionsOverload, missing(optionsOverload, "test('t', { timeout: 100 }, () => { run(); })", "() => {", "expect")),
			// `.each` spreads the row into the parameters, so `context` is a row.
			invalid(eachRow, missing(eachRow, "test.each([1])('t %s', (context) => { run(context); })", "(context) => {", "expect")),
			invalid(forContext, missing(forContext, "test.for([1])('t %s', (row, context) => { run(row); })", "(row, context) => {", "expect")),
			invalid(concurrent, missing(concurrent, "test.concurrent('t', async (context) => { await run(); })", "async (context) => {", "context.expect")),

			// ---- First statement ----
			invalid(nestedBranch, missing(nestedBranch, "test('t', () => { if (ready) { expect.assertions(1); } run(); })", "() => {", "expect")),
			invalid(directive, rule_tester.InvalidTestCaseError{
				MessageId: "haveExpectAssertions", Line: 2, Column: 1, EndLine: 2, EndColumn: 48,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestAddingHasAssertions", Output: strings.Replace(directive, "'use strict';", "'use strict';expect.hasAssertions();", 1)},
					{MessageId: "suggestAddingAssertions", Output: strings.Replace(directive, "'use strict';", "'use strict';expect.assertions();", 1)},
				},
			}),
			invalid(matcherNamedAssertions, missing(matcherNamedAssertions, "test('t', () => { expect(value).assertions(1); })", "() => {", "expect")),

			// ---- Arguments ----
			{
				Code: rstestImport + "test('t', () => { expect.assertions(1.5); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assertionsRequiresNumberArgument", Line: 2, Column: 37, EndLine: 2, EndColumn: 40},
				},
			},
			{
				Code: rstestImport + "test('t', () => { expect.assertions(1n); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assertionsRequiresNumberArgument", Line: 2, Column: 37, EndLine: 2, EndColumn: 39},
				},
			},
			{
				Code: rstestImport + "test('t', () => { expect.assertions(-1); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assertionsRequiresNumberArgument", Line: 2, Column: 37, EndLine: 2, EndColumn: 39},
				},
			},
			{
				Code: rstestImport + "test('t', () => { expect.assertions(1 as number); });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assertionsRequiresNumberArgument", Line: 2, Column: 37, EndLine: 2, EndColumn: 48},
				},
			},
			{
				Code:    bracketDisallowed,
				Options: []interface{}{map[string]interface{}{"disallowHasAssertions": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferAssertionsOverHasAssertions", Line: 2, Column: 26, EndLine: 2, EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestReplacingWithAssertions",
						Output:    strings.Replace(bracketDisallowed, "'hasAssertions'", "'assertions'", 1),
					}},
				}},
			},

			// ---- Options ----
			{
				Code:    whileLoop,
				Options: []interface{}{map[string]interface{}{"onlyFunctionsWithExpectInLoop": true}},
				Errors:  []rule_tester.InvalidTestCaseError{missing(whileLoop, "test('t', () => { while (next()) { expect(current()).toBe(1); } })", "() => {", "expect")},
			},
			{
				Code:    methodCallback,
				Options: []interface{}{map[string]interface{}{"onlyFunctionsWithExpectInCallback": true}},
				Errors:  []rule_tester.InvalidTestCaseError{missing(methodCallback, "test('t', () => { register({ handle() { expect(1).toBe(1); } }); })", "() => {", "expect")},
			},
			// A test that satisfies the rule does not exempt the next test that the
			// options select.
			{
				Code:    leak,
				Options: []interface{}{map[string]interface{}{"onlyFunctionsWithAsyncKeyword": true}},
				Errors:  []rule_tester.InvalidTestCaseError{missing(leak, "test('b', async () => { await run(); })", "async () => {", "expect")},
			},
		},
	)
}
