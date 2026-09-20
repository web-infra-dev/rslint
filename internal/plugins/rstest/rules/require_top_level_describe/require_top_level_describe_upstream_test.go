// TestRequireTopLevelDescribeUpstream ports the complete eslint-plugin-jest
// v29.16.0 and @vitest/eslint-plugin 1.6.27 suites to Rstest API spellings.
// Rstest-specific syntax and regressions live in
// require_top_level_describe_extras_test.go.
package require_top_level_describe_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_top_level_describe"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func maxDescribesOption(maxAllowed int) []any {
	return []any{map[string]any{"maxNumberOfTopLevelDescribes": maxAllowed}}
}

func unexpectedTestCaseError(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unexpectedTestCase",
		Message:   "All test cases must be wrapped in a describe block",
		Line:      line,
		Column:    column,
	}
}

func unexpectedHookError(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unexpectedHook",
		Message:   "All hooks must be wrapped in a describe block",
		Line:      line,
		Column:    column,
	}
}

func tooManyDescribesError(maxAllowed, line, column int) rule_tester.InvalidTestCaseError {
	plural := "s"
	if maxAllowed == 1 {
		plural = ""
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: "tooManyDescribes",
		Message:   fmt.Sprintf("There should not be more than %d describe%s at the top level", maxAllowed, plural),
		Line:      line,
		Column:    column,
	}
}

func runRequireTopLevelDescribeRuleTester(
	t *testing.T,
	valid []rule_tester.ValidTestCase,
	invalid []rule_tester.InvalidTestCase,
) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_top_level_describe.RequireTopLevelDescribeRule,
		valid,
		invalid,
	)
}

func TestRequireTopLevelDescribeUpstream(t *testing.T) {
	runRequireTopLevelDescribeRuleTester(
		t,
		[]rule_tester.ValidTestCase{
			{Code: `it.each()`},
			{Code: `import { it } from '@rstest/core';
it.extend({})`},
			{Code: `describe("test suite", () => { test("my test") });`},
			{Code: `describe("test suite", () => { it("my test") });`},
			{Code: `describe("test suite", () => {
  beforeEach("a", () => {});
  describe("b", () => {});
  test("c", () => {})
});`},
			{Code: `describe("test suite", () => { beforeAll("my beforeAll") });`},
			{Code: `describe("test suite", () => { afterEach("my afterEach") });`},
			{Code: `describe("test suite", () => { afterAll("my afterAll") });`},
			{Code: `describe("test suite", () => {
  it("my test", () => {})
  describe("another test suite", () => {
  });
  test("my other test", () => {})
});`},
			{Code: `foo()`},
			{Code: `describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });`},
			{Code: `describe('%s', () => {
  it('is fine', () => {
    //
  });
});

describe.each('world')('%s', () => {
  it.each([1, 2, 3])('%n', () => {
    //
  });
});`},
			{Code: `describe.each('hello')('%s', () => {
  it('is fine', () => {
    //
  });
});

describe.each('world')('%s', () => {
  it.each([1, 2, 3])('%n', () => {
    //
  });
});`},
			// The upstream jest suite proves a top-level framework namespace
			// call is not a registration; `rs.mock` is the Rstest spelling.
			{Code: `import { rs } from '@rstest/core';

rs.mock('./my-module');`},
			{Code: `rs.mock("./my-module")`},
			{Code: `describe('one', () => {});
describe('two', () => {});
describe('three', () => {});`},
			{
				Code: `describe('one', () => {
  describe('two', () => {});
  describe('three', () => {});
});`,
				Options: maxDescribesOption(1),
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `beforeEach("my test", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedHookError(1, 1)},
			},
			{
				Code: `test("my test", () => {})
describe("test suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code: `test("my test", () => {})
describe("test suite", () => {
  it("test", () => {})
});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code: `describe("test suite", () => {});
afterAll("my test", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedHookError(2, 1)},
			},
			{
				Code: `import { describe, afterAll as onceEverythingIsDone } from '@rstest/core';

describe("test suite", () => {});
onceEverythingIsDone("my test", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedHookError(4, 1)},
			},
			{
				Code: `import { 'describe' as describe, afterAll as onceEverythingIsDone } from '@rstest/core';

describe("test suite", () => {});
onceEverythingIsDone("my test", () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedHookError(4, 1)},
			},
			{
				Code:   `it.skip('test', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code:   `it.each([1, 2, 3])('%n', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code:   `it.skip.each([1, 2, 3])('%n', () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code:   "it.skip.each``('%n', () => {});",
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code:   "it.each``('%n', () => {});",
				Errors: []rule_tester.InvalidTestCaseError{unexpectedTestCaseError(1, 1)},
			},
			{
				Code: `describe('one', () => {});
describe('two', () => {});
describe('three', () => {});`,
				Options: maxDescribesOption(2),
				Errors:  []rule_tester.InvalidTestCaseError{tooManyDescribesError(2, 3, 1)},
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
				Options: maxDescribesOption(2),
				Errors:  []rule_tester.InvalidTestCaseError{tooManyDescribesError(2, 10, 1)},
			},
			{
				Code: `import {
  describe as describe1,
  describe as describe2,
  describe as describe3,
} from '@rstest/core';

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
				Options: maxDescribesOption(2),
				Errors:  []rule_tester.InvalidTestCaseError{tooManyDescribesError(2, 16, 1)},
			},
			{
				Code: `describe('one', () => {});
describe('two', () => {});
describe('three', () => {});`,
				Options: maxDescribesOption(1),
				Errors: []rule_tester.InvalidTestCaseError{
					tooManyDescribesError(1, 2, 1),
					tooManyDescribesError(1, 3, 1),
				},
			},
		},
	)
}
