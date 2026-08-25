package prefer_ending_with_an_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_ending_with_an_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferEndingWithAnExpectUpstream migrates the full valid/invalid suite from upstream
// jest-community/eslint-plugin-jest/src/rules/__tests__/prefer-ending-with-an-expect.test.ts
// 1:1. Position assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in prefer_ending_with_an_expect_extras_test.go.
func TestPreferEndingWithAnExpectUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_ending_with_an_expect.PreferEndingWithAnExpectRule,
		[]rule_tester.ValidTestCase{
			{Code: `it.todo("will test something eventually")`},
			{Code: `test.todo("will test something eventually")`},
			{Code: `['x']();`},
			{Code: `it("is weird", "because this should be a function")`},
			{Code: `it("is weird", "because this should be a function", () => {})`},
			{Code: `it("should pass", () => expect(true).toBeDefined())`},
			{Code: `test("should pass", () => expect(true).toBeDefined())`},
			{Code: `it("should pass", myTest); function myTest() { expect(true).toBeDefined() }`},
			{Code: `
        test('should pass', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect`, `foo`}}}},
			{Code: `it("should return undefined",() => expectSaga(mySaga).returns());`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expectSaga`}}}},
			{Code: `test('verifies expect method call', () => expect$(123));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect\$`}}}},
			{Code: `test('verifies expect method call', () => new Foo().expect(123));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`Foo.expect`}}}},
			{Code: `
        test('verifies deep expect method call', () => {
          tester.foo().expect(123);
        });
      `, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`tester.foo.expect`}}}},
			{Code: `
        test('verifies chained expect method call', () => {
          doSomething();

          tester
            .foo()
            .bar()
            .expect(456);
        });
      `, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`tester.foo.bar.expect`}}}},
			{Code: `
        test("verifies the function call", () => {
          td.verify(someFunctionCall())
        })
      `, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`td.verify`}}}},
			{Code: `it("should pass", async () => expect(true).toBeDefined())`},
			{Code: `it("should pass", () => expect(true).toBeDefined())`, Options: []interface{}{map[string]interface{}{}}},
			{Code: `it("should pass", () => { expect(true).toBeDefined() })`},
			{Code: `it("should pass", function () { expect(true).toBeDefined() })`},
			{Code: `
      it('is a complete test', () => {
        const container = render(Greeter);

        expect(container).toBeDefined();

        container.setProp('name', 'Bob');

        expect(container.toHTML()).toContain('Hello Bob!');
      });
    `},
			{Code: `
        it('is a complete test', async () => {
          const container = render(Greeter);

          expect(container).toBeDefined();

          container.setProp('name', 'Bob');

          await expect(container.toHTML()).resolve.toContain('Hello Bob!');
        });
      `},
			{Code: `
        it('is a complete test', async function () {
          const container = render(Greeter);

          expect(container).toBeDefined();

          container.setProp('name', 'Bob');

          await expect(container.toHTML()).resolve.toContain('Hello Bob!');
        });
      `},
			{Code: `
        describe('GET /user', function () {
          it('responds with json', function (done) {
            doSomething();
            request(app).get('/user').expect('Content-Type', /json/).expect(200, done);
          });
        });
      `, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect`, `request.**.expect`}}}},
			{Code: "\n        each([\n          [2, 3],\n          [1, 3],\n        ]).test(\n          'the selection can change from %d to %d',\n          (firstSelection, secondSelection) => {\n            const container = render(MySelect, {\n              props: { options: [1, 2, 3], selected: firstSelection },\n            });\n\n            expect(container).toBeDefined();\n            expect(container.toHTML()).toContain(\n              `<option value=\"${firstSelection}\" selected>`\n            );\n\n            container.setProp('selected', secondSelection);\n\n            expect(container.toHTML()).not.toContain(\n              `<option value=\"${firstSelection}\" selected>`\n            );\n            expect(container.toHTML()).toContain(\n              `<option value=\"${secondSelection}\" selected>`\n            );\n          }\n        );\n      ", Options: []interface{}{map[string]interface{}{`additionalTestBlockFunctions`: []interface{}{`each.test`}}}},
			{Code: `test('should pass *', () => expect404ToBeLoaded());`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect*`}}}},
			{Code: `test('should pass *', () => expect.toHaveStatus404());`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect.**`}}}},
			{Code: `test('should pass', () => tester.foo().expect(123));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`tester.*.expect`}}}},
			{Code: `test('should pass **', () => tester.foo().expect(123));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`**`}}}},
			{Code: `test('should pass *', () => tester.foo().expect(123));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`*`}}}},
			{Code: `test('should pass', () => tester.foo().expect(123));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`tester.**`}}}},
			{Code: `test('should pass', () => tester.foo().expect(123));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`tester.*`}}}},
			{Code: `test('should pass', () => tester.foo().bar().expectIt(456));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`tester.**.expect*`}}}},
			{Code: `test('should pass', () => request.get().foo().expect(456));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`request.**.expect`}}}},
			{Code: `test('should pass', () => request.get().foo().expect(456));`, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`request.**.e*e*t`}}}},
			{Code: `
        import { test } from '@jest/globals';

        test('should pass', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect`, `foo`}}}},
			{Code: `
        import { test as checkThat } from '@jest/globals';

        checkThat('this passes', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect`, `foo`}}}},
			{Code: `
        const { test } = require('@jest/globals');

        test('verifies chained expect method call', () => {
          tester
            .foo()
            .bar()
            .expect(456);
        });
      `, Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`tester.foo.bar.expect`}}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `it("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `test("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code: `test.skip("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: `it("should fail", () => { somePromise.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `test("should fail", () => { foo(true).toBe(true); })`,
				Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `it("should also fail",() => expectSaga(mySaga).returns());`,
				Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should pass", () => somePromise().then(() => expect(true).toBeDefined()))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should pass", () => render(Greeter))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should pass", () => { render(Greeter) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should pass", function () { render(Greeter) })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should not pass", () => class {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should not pass", () => ([]))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should not pass", () => { const x = []; })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should not pass", function () { class Mx {} })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it('is a complete test', () => {
  const container = render(Greeter);

  expect(container).toBeDefined();

  container.setProp('name', 'Bob');
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it('is a complete test', async () => {
  const container = render(Greeter);

  await expect(container).toBeDefined();

  await container.setProp('name', 'Bob');
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `test('should fail', () => request.get().foo().expect(456));`,
				Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`request.*.expect`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail', () => request.get().foo().bar().expect(456));`,
				Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`request.foo**.expect`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail', () => tester.request(123));`,
				Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`request.*`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail', () => request(123));`,
				Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`request.*`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail', () => request(123));`,
				Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`request.**`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code: `import { test as checkThat } from '@jest/globals';

checkThat('this passes', () => {
  // ...
});`,
				Options: []interface{}{map[string]interface{}{`assertFunctionNames`: []interface{}{`expect`, `foo`}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 3, Column: 1, EndLine: 3, EndColumn: 10},
				},
			},
			{
				Code: `import { test as checkThat } from '@jest/globals';

checkThat.skip('this passes', () => {
  // ...
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `mustEndWithExpect`, Line: 3, Column: 1, EndLine: 3, EndColumn: 15},
				},
			},
		},
	)
}
