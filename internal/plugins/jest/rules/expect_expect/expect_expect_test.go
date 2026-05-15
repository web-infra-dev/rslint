package expect_expect_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/expect_expect"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExpectExpectRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&expect_expect.ExpectExpectRule,
		[]rule_tester.ValidTestCase{
			{Code: `it.todo("will test something eventually")`},
			{Code: `test.todo("will test something eventually")`},
			{Code: `['x']();`},
			{Code: `it("should pass", () => expect(true).toBeDefined())`},
			{Code: `test("should pass", () => expect(true).toBeDefined())`},
			{Code: `it("should pass", () => somePromise().then(() => expect(true).toBeDefined()))`},
			{Code: `it("should pass", myTest); function myTest() { expect(true).toBeDefined() }`},
			{Code: `function myTest() { expect(true).toBeDefined() } it("should pass", myTest);`},
			{
				Code: `
        function sharedAssertion() {
          expect(true).toBeDefined();
        }

        it("first passes", sharedAssertion);
        test("second passes", sharedAssertion);
      `,
			},
			{
				Code: `
        test('should pass', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect", "foo"}}},
			},
			{
				Code:    `it("should return undefined",() => expectSaga(mySaga).returns());`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expectSaga"}}},
			},
			{
				Code:    `test('verifies expect method call', () => expect$(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{`expect\$`}}},
			},
			{
				Code:    `test('verifies expect method call', () => new Foo().expect(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"Foo.expect"}}},
			},
			{
				Code: `
        test('verifies deep expect method call', () => {
          tester.foo().expect(123);
        });
      `,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"tester.foo.expect"}}},
			},
			{
				Code: `
        test('verifies chained expect method call', () => {
          tester
            .foo()
            .bar()
            .expect(456);
        });
      `,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"tester.foo.bar.expect"}}},
			},
			{
				Code: `
        test("verifies the function call", () => {
          td.verify(someFunctionCall())
        })
      `,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"td.verify"}}},
			},
			{
				Code: `it("should pass", () => expect(true).toBeDefined())`,
				Options: []interface{}{
					map[string]interface{}{
						"assertFunctionNames":          nil,
						"additionalTestBlockFunctions": nil,
					},
				},
			},
			{
				Code: `
        theoretically('the number {input} is correctly translated to string', theories, theory => {
          const output = NumberToLongString(theory.input);
          expect(output).toBe(theory.expected);
        })
      `,
				Options: []interface{}{map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"theoretically"}}},
			},
			{
				Code:    `test('should pass *', () => expect404ToBeLoaded());`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect*"}}},
			},
			{
				Code:    `test('should pass *', () => expect.toHaveStatus404());`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect.**"}}},
			},
			{
				Code:    `test('should pass', () => tester.foo().expect(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"tester.*.expect"}}},
			},
			{
				Code:    `test('should pass **', () => tester.foo().expect(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"**"}}},
			},
			{
				Code:    `test('should pass *', () => tester.foo().expect(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"*"}}},
			},
			{
				Code:    `test('should pass', () => tester.foo().expect(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"tester.**"}}},
			},
			{
				Code:    `test('should pass', () => tester.foo().expect(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"tester.*"}}},
			},
			{
				Code:    `test('should pass', () => tester.foo().bar().expectIt(456));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"tester.**.expect*"}}},
			},
			{
				Code:    `test('should pass', () => request.get().foo().expect(456));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.**.expect"}}},
			},
			{
				Code:    `test('should pass', () => request.get().foo().expect(456));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.**.e*e*t"}}},
			},
			{
				Code: `test('should still lint with malformed pattern config', () => {
          expect(true).toBeDefined();
        });`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"(", "expect"}}},
			},
			{
				Code: `
        import { test } from '@jest/globals';

        test('should pass', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect", "foo"}}},
			},
			{
				Code: `
        import { test as checkThat } from '@jest/globals';

        checkThat('this passes', () => {
          expect(true).toBeDefined();
          foo(true).toBe(true);
        });
      `,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect", "foo"}}},
			},
			{
				Code: `
        const { test } = require('@jest/globals');

        test('verifies chained expect method call', () => {
          tester
            .foo()
            .bar()
            .expect(456);
        });
      `,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"tester.foo.bar.expect"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `it("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `it("should fail", myTest); function myTest() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code: `test("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code: `test.skip("should fail", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `afterEach(() => {});`,
				Options: []interface{}{map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"afterEach"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: `theoretically('the number {input} is correctly translated to string', theories, theory => {
          const output = NumberToLongString(theory.input);
        })`,
				Options: []interface{}{map[string]interface{}{"additionalTestBlockFunctions": []interface{}{"theoretically"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `it("should fail", () => { somePromise.then(() => {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `test("should fail", () => { foo(true).toBe(true); })`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `it("should also fail",() => expectSaga(mySaga).returns());`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `test('should fail', () => request.get().foo().expect(456));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.*.expect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail', () => request.get().foo().bar().expect(456));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.foo**.expect"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail', () => tester.request(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.*"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail', () => request(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.*"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail', () => request(123));`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"request.**"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    `test('should fail without crashing on malformed pattern config', () => {});`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"("}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code: `import { test as checkThat } from '@jest/globals';

        checkThat('this passes', () => {
          // ...
        });`,
				Options: []interface{}{map[string]interface{}{"assertFunctionNames": []interface{}{"expect", "foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 3, Column: 9, EndLine: 3, EndColumn: 18},
				},
			},
			{
				Code: `import { test as checkThat } from '@jest/globals';

        checkThat.skip('this passes', () => {
          // ...
        });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 3, Column: 9, EndLine: 3, EndColumn: 23},
				},
			},
			{
				Code: `it('outer', helper);
        function helper() {}

        describe('nested', () => {
          it('inner', helper);
          function helper() {
            expect(true).toBeDefined();
          }
        });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAssertions", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
				},
			},
		},
	)
}
