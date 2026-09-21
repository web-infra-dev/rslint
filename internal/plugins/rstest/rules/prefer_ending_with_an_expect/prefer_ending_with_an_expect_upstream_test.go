// Package prefer_ending_with_an_expect_test migrates the upstream suite from
// jest-community/eslint-plugin-jest@29.16.1
// src/rules/__tests__/prefer-ending-with-an-expect.test.ts (the rule has no
// counterpart in @vitest/eslint-plugin) into Rstest spellings. Rstest-specific
// behavior lives in prefer_ending_with_an_expect_extras_test.go.
//
// Dispositions that are not a plain Keep:
//
//   - `it("is weird", "because this should be a function", () => {})` is valid
//     upstream because jest only ever looks at the second argument. Rstest
//     accepts `(name, options, fn)`, and @rstest/core 0.11 runs the third
//     argument even when the second is not an options object, so the case
//     reverses to a report.
//   - `@jest/globals` imports are adapted to `@rstest/core`.
package prefer_ending_with_an_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_ending_with_an_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func assertNamesOption(names ...string) []interface{} {
	values := make([]interface{}, 0, len(names))
	for _, name := range names {
		values = append(values, name)
	}
	return []interface{}{map[string]interface{}{"assertFunctionNames": values}}
}

func additionalBlocksOption(names ...string) []interface{} {
	values := make([]interface{}, 0, len(names))
	for _, name := range names {
		values = append(values, name)
	}
	return []interface{}{map[string]interface{}{"additionalTestBlockFunctions": values}}
}

func mustEndWithExpectError(line int, column int, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "mustEndWithExpect",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

func runPreferEndingWithAnExpectRuleTester(
	t *testing.T,
	valid []rule_tester.ValidTestCase,
	invalid []rule_tester.InvalidTestCase,
) {
	t.Helper()
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_ending_with_an_expect.PreferEndingWithAnExpectRule,
		valid,
		invalid,
	)
}

func TestPreferEndingWithAnExpectUpstream(t *testing.T) {
	runPreferEndingWithAnExpectRuleTester(
		t,
		[]rule_tester.ValidTestCase{
			{Code: `it.todo("will test something eventually")`},
			{Code: `test.todo("will test something eventually")`},
			{Code: `['x']();`},
			{Code: `it("is weird", "because this should be a function")`},
			{Code: `it("should pass", () => expect(true).toBeDefined())`},
			{Code: `test("should pass", () => expect(true).toBeDefined())`},
			{Code: `it("should pass", myTest); function myTest() { expect(true).toBeDefined() }`},
			{
				Code: `test('should pass', () => {
  expect(true).toBeDefined();
  foo(true).toBe(true);
});`,
				Options: assertNamesOption("expect", "foo"),
			},
			{
				Code:    `it("should return undefined",() => expectSaga(mySaga).returns());`,
				Options: assertNamesOption("expectSaga"),
			},
			{
				Code:    `test('verifies expect method call', () => expect$(123));`,
				Options: assertNamesOption(`expect\$`),
			},
			{
				Code:    `test('verifies expect method call', () => new Foo().expect(123));`,
				Options: assertNamesOption("Foo.expect"),
			},
			{
				Code: `test('verifies deep expect method call', () => {
  tester.foo().expect(123);
});`,
				Options: assertNamesOption("tester.foo.expect"),
			},
			{
				Code: `test('verifies chained expect method call', () => {
  doSomething();

  tester
    .foo()
    .bar()
    .expect(456);
});`,
				Options: assertNamesOption("tester.foo.bar.expect"),
			},
			{
				Code: `test("verifies the function call", () => {
  td.verify(someFunctionCall())
})`,
				Options: assertNamesOption("td.verify"),
			},
			{Code: `it("should pass", async () => expect(true).toBeDefined())`},
			{
				Code:    `it("should pass", () => expect(true).toBeDefined())`,
				Options: []interface{}{map[string]interface{}{}},
			},
			{Code: `it("should pass", () => { expect(true).toBeDefined() })`},
			{Code: `it("should pass", function () { expect(true).toBeDefined() })`},
			{Code: `it('is a complete test', () => {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');

  expect(container.toHTML()).toContain('Hello Bob!');
});`},
			{Code: `it('is a complete test', async () => {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');

  await expect(container.toHTML()).resolves.toContain('Hello Bob!');
});`},
			{Code: `it('is a complete test', async function () {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');

  await expect(container.toHTML()).resolves.toContain('Hello Bob!');
});`},
			{
				Code: `describe('GET /user', function () {
  it('responds with json', function (done) {
    doSomething();
    request(app).get('/user').expect('Content-Type', /json/).expect(200, done);
  });
});`,
				Options: assertNamesOption("expect", "request.**.expect"),
			},
			{
				Code:    "each([\n  [2, 3],\n  [1, 3],\n]).test(\n  'the selection can change from %d to %d',\n  (firstSelection, secondSelection) => {\n    const container = render(MySelect, {\n      props: { options: [1, 2, 3], selected: firstSelection },\n    });\n\n    expect(container).toBeDefined();\n    expect(container.toHTML()).toContain(\n      `<option value=\"${firstSelection}\" selected>`\n    );\n\n    container.setProp('selected', secondSelection);\n\n    expect(container.toHTML()).not.toContain(\n      `<option value=\"${firstSelection}\" selected>`\n    );\n    expect(container.toHTML()).toContain(\n      `<option value=\"${secondSelection}\" selected>`\n    );\n  }\n);",
				Options: additionalBlocksOption("each.test"),
			},

			// ---- wildcards ----
			{Code: `test('should pass *', () => expect404ToBeLoaded());`, Options: assertNamesOption("expect*")},
			{Code: `test('should pass *', () => expect.toHaveStatus404());`, Options: assertNamesOption("expect.**")},
			{Code: `test('should pass', () => tester.foo().expect(123));`, Options: assertNamesOption("tester.*.expect")},
			{Code: `test('should pass **', () => tester.foo().expect(123));`, Options: assertNamesOption("**")},
			{Code: `test('should pass *', () => tester.foo().expect(123));`, Options: assertNamesOption("*")},
			{Code: `test('should pass', () => tester.foo().expect(123));`, Options: assertNamesOption("tester.**")},
			{Code: `test('should pass', () => tester.foo().expect(123));`, Options: assertNamesOption("tester.*")},
			{Code: `test('should pass', () => tester.foo().bar().expectIt(456));`, Options: assertNamesOption("tester.**.expect*")},
			{Code: `test('should pass', () => request.get().foo().expect(456));`, Options: assertNamesOption("request.**.expect")},
			{Code: `test('should pass', () => request.get().foo().expect(456));`, Options: assertNamesOption("request.**.e*e*t")},

			// ---- aliases (adapted from @jest/globals) ----
			{
				Code: `import { test } from '@rstest/core';

test('should pass', () => {
  expect(true).toBeDefined();
  foo(true).toBe(true);
});`,
				Options: assertNamesOption("expect", "foo"),
			},
			{
				Code: `import { test as checkThat } from '@rstest/core';

checkThat('this passes', () => {
  expect(true).toBeDefined();
  foo(true).toBe(true);
});`,
				Options: assertNamesOption("expect", "foo"),
			},
			{
				Code: `const { test } = require('@rstest/core');

test('verifies chained expect method call', () => {
  tester
    .foo()
    .bar()
    .expect(456);
});`,
				Options: assertNamesOption("tester.foo.bar.expect"),
			},
		},
		[]rule_tester.InvalidTestCase{
			// Reversed from upstream: Rstest runs the third argument, so the
			// callback is checked instead of ignored.
			{
				Code:   `it("is weird", "because this should be a function", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `test("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:   `test.skip("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 10)},
			},
			{
				Code:   `it("should fail", () => { somePromise.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:    `test("should fail", () => { foo(true).toBe(true); })`,
				Options: assertNamesOption("expect"),
				Errors:  []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:    `it("should also fail",() => expectSaga(mySaga).returns());`,
				Options: assertNamesOption("expect"),
				Errors:  []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should pass", () => somePromise().then(() => expect(true).toBeDefined()))`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should pass", () => render(Greeter))`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should pass", () => { render(Greeter) })`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should pass", function () { render(Greeter) })`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should not pass", () => class {})`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should not pass", () => ([]))`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should not pass", () => { const x = []; })`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code:   `it("should not pass", function () { class Mx {} })`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code: `it('is a complete test', () => {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');
});`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},
			{
				Code: `it('is a complete test', async () => {
  const container = render(Greeter);

  await expect(container).toBeDefined();

  await container.setProp('name', 'Bob');
});`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 3)},
			},

			// ---- wildcards ----
			{
				Code:    `test('should fail', () => request.get().foo().expect(456));`,
				Options: assertNamesOption("request.*.expect"),
				Errors:  []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:    `test('should fail', () => request.get().foo().bar().expect(456));`,
				Options: assertNamesOption("request.foo**.expect"),
				Errors:  []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:    `test('should fail', () => tester.request(123));`,
				Options: assertNamesOption("request.*"),
				Errors:  []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:    `test('should fail', () => request(123));`,
				Options: assertNamesOption("request.*"),
				Errors:  []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},
			{
				Code:    `test('should fail', () => request(123));`,
				Options: assertNamesOption("request.**"),
				Errors:  []rule_tester.InvalidTestCaseError{mustEndWithExpectError(1, 1, 5)},
			},

			// ---- aliases (adapted from @jest/globals) ----
			{
				Code: `import { test as checkThat } from '@rstest/core';

checkThat('this passes', () => {
  // ...
});`,
				Options: assertNamesOption("expect", "foo"),
				Errors:  []rule_tester.InvalidTestCaseError{mustEndWithExpectError(3, 1, 10)},
			},
			{
				Code: `import { test as checkThat } from '@rstest/core';

checkThat.skip('this passes', () => {
  // ...
});`,
				Errors: []rule_tester.InvalidTestCaseError{mustEndWithExpectError(3, 1, 15)},
			},
		},
	)
}
