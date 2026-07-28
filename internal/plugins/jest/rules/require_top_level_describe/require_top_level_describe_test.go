package require_top_level_describe_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/require_top_level_describe"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireTopLevelDescribeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_top_level_describe.RequireTopLevelDescribeRule,
		[]rule_tester.ValidTestCase{
			{Code: "it.each()"},
			{Code: `describe("test suite", () => { test("my test") });`},
			{Code: `describe("test suite", () => { it("my test") });`},
			{Code: `
      describe("test suite", () => {
        beforeEach("a", () => {});
        describe("b", () => {});
        test("c", () => {})
      });
    `},
			{Code: `describe("test suite", () => { beforeAll("my beforeAll") });`},
			{Code: `describe("test suite", () => { afterEach("my afterEach") });`},
			{Code: `describe("test suite", () => { afterAll("my afterAll") });`},
			{Code: `
      describe("test suite", () => {
        it("my test", () => {})
        describe("another test suite", () => {
        });
        test("my other test", () => {})
      });
    `},
			{Code: "foo()"},
			{Code: `describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });`},
			{Code: `
      describe('%s', () => {
        it('is fine', () => {
          //
        });
      });

      describe.each('world')('%s', () => {
        it.each([1, 2, 3])('%n', () => {
          //
        });
      });
    `},
			{Code: `
      describe.each('hello')('%s', () => {
        it('is fine', () => {
          //
        });
      });

      describe.each('world')('%s', () => {
        it.each([1, 2, 3])('%n', () => {
          //
        });
      });
    `},
			{Code: `
        import { jest } from '@jest/globals';

        jest.doMock('my-module');
      `},
			{Code: `jest.doMock("my-module")`},
			{Code: `
        describe('one', () => {});
        describe('two', () => {});
        describe('three', () => {});
      `},
			{
				Code: `
          describe('one', () => {
            describe('two', () => {});
            describe('three', () => {});
          });
        `,
				Options: []interface{}{
					map[string]interface{}{"maxNumberOfTopLevelDescribes": 1},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `beforeEach("my test", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook"},
				},
			},
			{
				Code: `
        test("my test", () => {})
        describe("test suite", () => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTestCase"},
				},
			},
			{
				Code: `
        test("my test", () => {})
        describe("test suite", () => {
          it("test", () => {})
        });
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTestCase"},
				},
			},
			{
				Code: `
        describe("test suite", () => {});
        afterAll("my test", () => {})
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook"},
				},
			},
			{
				Code: `
        import { describe, afterAll as onceEverythingIsDone } from '@jest/globals';

        describe("test suite", () => {});
        onceEverythingIsDone("my test", () => {})
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook"},
				},
			},
			{
				Code: `
        import { 'describe' as describe, afterAll as onceEverythingIsDone } from '@jest/globals';

        describe("test suite", () => {});
        onceEverythingIsDone("my test", () => {})
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook"},
				},
			},
			{
				Code: `it.skip('test', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTestCase"},
				},
			},
			{
				Code: `it.each([1, 2, 3])('%n', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTestCase"},
				},
			},
			{
				Code: `it.skip.each([1, 2, 3])('%n', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTestCase"},
				},
			},
			{
				Code: "it.skip.each``('%n', () => {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTestCase"},
				},
			},
			{
				Code: "it.each``('%n', () => {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTestCase"},
				},
			},
			{
				Code: `describe('one', () => {});
describe('two', () => {});
describe('three', () => {});`,
				Options: []interface{}{
					map[string]interface{}{"maxNumberOfTopLevelDescribes": 2},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyDescribes", Line: 3},
				},
			},
			{
				Code: `describe('one', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
});
describe('two', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
  describe('three (nested)', () => {});
});
describe('three', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
  describe('three (nested)', () => {});
});`,
				Options: []interface{}{
					map[string]interface{}{"maxNumberOfTopLevelDescribes": 2},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyDescribes", Line: 10},
				},
			},
			{
				Code: `import {
  describe as describe1,
  describe as describe2,
  describe as describe3,
} from '@jest/globals';

describe1('one', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
});
describe2('two', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
  describe('three (nested)', () => {});
});
describe3('three', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
  describe('three (nested)', () => {});
});`,
				Options: []interface{}{
					map[string]interface{}{"maxNumberOfTopLevelDescribes": 2},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyDescribes", Line: 16},
				},
			},
			{
				Code: `describe('one', () => {});
describe('two', () => {});
describe('three', () => {});`,
				Options: []interface{}{
					map[string]interface{}{"maxNumberOfTopLevelDescribes": 1},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tooManyDescribes", Line: 2},
					{MessageId: "tooManyDescribes", Line: 3},
				},
			},
		},
	)
}
