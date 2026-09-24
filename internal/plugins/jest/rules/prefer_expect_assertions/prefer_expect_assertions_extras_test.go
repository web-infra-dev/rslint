package prefer_expect_assertions_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_expect_assertions"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// missing expects haveExpectAssertions on call, the exact source text of a
// test registration that appears once in code, with both suggestions inserted
// right after the first occurrence of marker inside that call. An empty
// marker expects no suggestions.
func missing(code, call, marker string) rule_tester.InvalidTestCaseError {
	start := strings.Index(code, call)
	line, column := position(code, start)
	endLine, endColumn := position(code, start+len(call))
	err := rule_tester.InvalidTestCaseError{MessageId: "haveExpectAssertions", Line: line, Column: column, EndLine: endLine, EndColumn: endColumn}
	if marker == "" {
		return err
	}
	at := start + strings.Index(call, marker) + len(marker)
	insert := func(member string) string {
		return code[:at] + "expect." + member + "();" + code[at:]
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

func invalid(code string, options any, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, Options: options, Errors: errors}
}

func option(name string) []interface{} {
	return []interface{}{map[string]interface{}{name: true}}
}

func TestPreferExpectAssertionsExtras(t *testing.T) {
	matcherNamedAssertions := `it('t', () => { expect(value).assertions(1); });`
	nestedBranch := `it('t', () => { if (ready) { expect.assertions(1); } run(); });`
	logicalFirst := `it('t', () => { strict && expect.hasAssertions(); run(); });`
	directive := `it('t', function () { 'use strict'; run(); });`
	directiveWithoutSemicolon := "it('t', () => {\n  'use strict'\n  run();\n});"
	reassignedHook := `let setup = () => expect.hasAssertions(); setup = () => {}; beforeEach(setup); it('t', () => {});`
	foreignReference := `beforeEach(helpers.hasAssertions); it('t', () => {});`
	shadowedReference := `function run(expect) { beforeEach(expect.hasAssertions); it('t', () => {}); }`
	beforeAllOnly := `beforeAll(() => { expect.hasAssertions(); }); it('t', () => {});`
	// A declaration the hook does not make on every run leaves the test
	// unprotected.
	branchInHook := `afterEach(() => { if (strict) { expect.hasAssertions(); } }); it('t', () => {});`
	logicalInHook := `beforeEach(() => strict && expect.hasAssertions()); it('t', () => {});`
	returnBeforeHookDeclaration := `beforeEach(() => { if (skip) return; expect.hasAssertions(); }); it('t', () => {});`
	// A hook written in a helper registers into the suites that call the
	// helper, not into every suite of the file.
	helperHook := "function installAssertions() { beforeEach(() => expect.hasAssertions()); }\ndescribe('A', () => { installAssertions(); test('a', () => {}); });\ndescribe('B', () => { test('b', () => {}); });"
	sharedTests := "function commonTests() { it('case', () => {}); }\ndescribe('A', () => { beforeEach(() => expect.hasAssertions()); commonTests(); });\ndescribe('B', () => { commonTests(); });"
	sharedSuiteBody := "const body = () => { it('case', () => {}); };\ndescribe('A', () => { afterEach(() => expect.hasAssertions()); describe('inner', body); });\ndescribe('B', () => { describe('inner', body); });"
	unusedHelper := "describe('A', () => { function setup() { beforeEach(() => expect.hasAssertions()); } });\ndescribe('B', () => { it('b', () => {}); });"
	aliasedHelper := "function setup() { beforeEach(() => expect.hasAssertions()); }\nconst install = setup;\ndescribe('A', () => { install(); });\ndescribe('B', () => { it('b', () => {}); });"
	optionalCallInHook := `beforeEach(() => maybe?.(expect.hasAssertions())); it('t', () => {});`
	bindingDefaultInHook := `afterEach(() => { const { x = expect.hasAssertions() } = { x: 1 }; }); it('t', () => {});`
	optionalCallFirst := `it('t', () => { maybe?.(expect.hasAssertions()); run(); });`
	whileLoop := `it('t', () => { while (next()) { expect(current()).toBe(1); } });`
	methodCallback := `it('t', () => { register({ handle() { expect(1).toBe(1); } }); });`
	leak := "it('a', () => { expect.assertions(1); expect(1).toBe(1); });\nit('b', async () => { await run(); });"

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_expect_assertions.PreferExpectAssertionsRule,
		[]rule_tester.ValidTestCase{
			// ---- Hooks ----
			// Upstream keeps this case commented out with a todo saying it should be
			// valid. Jest's hasAssertions reads the global matcher state rather than
			// its receiver, so a hook can call the detached method.
			{Code: `describe('my tests', () => {
  beforeEach(expect.hasAssertions);

  it("is a number that is greater than four", () => {
    expect(number).toBeGreaterThan(4);
  });

  it("returns numbers that are greater than four", () => {
    for (const number of getNumbers()) {
      expect(number).toBeGreaterThan(4);
    }
  });
});

it("returns numbers that are greater than five", () => {
  expect(number).toBeGreaterThan(5);
});`, Options: option("onlyFunctionsWithExpectInLoop")},
			{Code: `function requireAssertions() { expect.hasAssertions(); } afterEach(requireAssertions); it('t', () => {});`},
			{Code: `it('t', () => {}); const setup = () => expect.assertions(1); beforeEach(setup);`},
			{Code: `import { beforeEach as setup, expect as check, it } from '@jest/globals'; setup(() => check.hasAssertions()); it('t', () => {});`},
			{Code: `import { expect as check, it } from '@jest/globals'; beforeEach(check.hasAssertions); it('t', () => {});`},
			{Code: `beforeEach(() => { const server = start(); expect.hasAssertions(); }); it('t', () => {});`},
			{Code: "describe('A', suiteBody);\nfunction suiteBody() { afterEach(() => expect.hasAssertions()); it('a', () => {}); }"},
			{Code: "function installAssertions() { beforeEach(() => expect.hasAssertions()); }\ndescribe('A', () => { installAssertions(); describe('inner', () => { it('a', () => {}); }); });"},
			{Code: "function commonTests() { it('case', () => {}); }\ndescribe('A', () => { beforeEach(() => expect.hasAssertions()); commonTests(); });\ndescribe('B', () => { afterEach(() => expect.hasAssertions()); commonTests(); });"},
			{Code: "function setup() { afterEach(() => expect.hasAssertions()); }\nconst install = setup;\ndescribe('A', () => { install(); it('a', () => {}); });"},

			// ---- First statement ----
			{Code: `it('t', function () { 'use strict'; expect.hasAssertions(); });`},
			{Code: `it('t', async () => { await expect.hasAssertions(); });`},
			{Code: `it('t', () => { expect['assertions'](1); });`},

			// ---- Arguments that are integer number literals ----
			{Code: `it('t', () => { expect.assertions(1.0); });`},
			{Code: `it('t', () => { expect.assertions(0x1); });`},
			{Code: `it('t', () => { expect.assertions(1e3); });`},
			{Code: `it('t', () => { expect.assertions(1_000); });`},

			// ---- Options ----
			// The loop encloses the registration, not the assertion.
			{Code: `for (const value of values) { it('t', () => { expect(value).toBe(1); }); }`, Options: option("onlyFunctionsWithExpectInLoop")},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Hooks ----
			invalid(reassignedHook, nil, missing(reassignedHook, "it('t', () => {})", "() => {")),
			invalid(foreignReference, nil, missing(foreignReference, "it('t', () => {})", "() => {")),
			invalid(shadowedReference, nil, missing(shadowedReference, "it('t', () => {})", "() => {")),
			invalid(beforeAllOnly, nil, missing(beforeAllOnly, "it('t', () => {})", "() => {")),
			invalid(branchInHook, nil, missing(branchInHook, "it('t', () => {})", "it('t', () => {")),
			invalid(logicalInHook, nil, missing(logicalInHook, "it('t', () => {})", "it('t', () => {")),
			invalid(returnBeforeHookDeclaration, nil, missing(returnBeforeHookDeclaration, "it('t', () => {})", "it('t', () => {")),
			invalid(helperHook, nil, missing(helperHook, "test('b', () => {})", "test('b', () => {")),
			invalid(sharedTests, nil, missing(sharedTests, "it('case', () => {})", "it('case', () => {")),
			invalid(sharedSuiteBody, nil, missing(sharedSuiteBody, "it('case', () => {})", "it('case', () => {")),
			invalid(unusedHelper, nil, missing(unusedHelper, "it('b', () => {})", "it('b', () => {")),
			invalid(aliasedHelper, nil, missing(aliasedHelper, "it('b', () => {})", "it('b', () => {")),
			invalid(optionalCallInHook, nil, missing(optionalCallInHook, "it('t', () => {})", "it('t', () => {")),
			invalid(bindingDefaultInHook, nil, missing(bindingDefaultInHook, "it('t', () => {})", "it('t', () => {")),

			// ---- First statement ----
			// A matcher named assertions reads it off expect's result and declares
			// no assertion count.
			invalid(matcherNamedAssertions, nil, missing(matcherNamedAssertions, matcherNamedAssertions[:len(matcherNamedAssertions)-1], "() => {")),
			// A call inside a branch only declares a count on some paths.
			invalid(nestedBranch, nil, missing(nestedBranch, nestedBranch[:len(nestedBranch)-1], "() => {")),
			// So does one on the right of a short-circuiting operator.
			invalid(logicalFirst, nil, missing(logicalFirst, logicalFirst[:len(logicalFirst)-1], "() => {")),
			invalid(optionalCallFirst, nil, missing(optionalCallFirst, optionalCallFirst[:len(optionalCallFirst)-1], "() => {")),
			invalid(directive, nil, rule_tester.InvalidTestCaseError{
				MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 46,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestAddingHasAssertions", Output: strings.Replace(directive, "'use strict';", "'use strict';expect.hasAssertions();", 1)},
					{MessageId: "suggestAddingAssertions", Output: strings.Replace(directive, "'use strict';", "'use strict';expect.assertions();", 1)},
				},
			}),

			// A directive written without a semicolon gets one, or the inserted
			// statement would run into it.
			invalid(directiveWithoutSemicolon, nil, rule_tester.InvalidTestCaseError{
				MessageId: "haveExpectAssertions", Line: 1, Column: 1, EndLine: 4, EndColumn: 3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestAddingHasAssertions", Output: strings.Replace(directiveWithoutSemicolon, "'use strict'", "'use strict';expect.hasAssertions();", 1)},
					{MessageId: "suggestAddingAssertions", Output: strings.Replace(directiveWithoutSemicolon, "'use strict'", "'use strict';expect.assertions();", 1)},
				},
			}),

			// ---- Arguments ----
			invalid(`it('t', () => { expect.assertions(1.5); });`, nil,
				rule_tester.InvalidTestCaseError{MessageId: "assertionsRequiresNumberArgument", Line: 1, Column: 35, EndLine: 1, EndColumn: 38}),
			invalid(`it('t', () => { expect.assertions(1n); });`, nil,
				rule_tester.InvalidTestCaseError{MessageId: "assertionsRequiresNumberArgument", Line: 1, Column: 35, EndLine: 1, EndColumn: 37}),
			invalid(`it('t', () => { expect.assertions(-1); });`, nil,
				rule_tester.InvalidTestCaseError{MessageId: "assertionsRequiresNumberArgument", Line: 1, Column: 35, EndLine: 1, EndColumn: 37}),

			// ---- Options ----
			invalid(whileLoop, option("onlyFunctionsWithExpectInLoop"), missing(whileLoop, whileLoop[:len(whileLoop)-1], "() => {")),
			invalid(methodCallback, option("onlyFunctionsWithExpectInCallback"), missing(methodCallback, methodCallback[:len(methodCallback)-1], "() => {")),
			// A test that satisfies the rule does not exempt the next test that the
			// options select.
			invalid(leak, option("onlyFunctionsWithAsyncKeyword"), missing(leak, "it('b', async () => { await run(); })", "async () => {")),
		},
	)
}
