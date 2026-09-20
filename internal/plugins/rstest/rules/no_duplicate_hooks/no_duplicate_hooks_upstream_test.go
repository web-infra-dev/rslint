// TestNoDuplicateHooksUpstream migrates the complete eslint-plugin-jest
// v29.16.1 suite to Rstest API spellings. Rstest provenance and lexical-policy
// lock-ins live in no_duplicate_hooks_extras_test.go.
package no_duplicate_hooks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_duplicate_hooks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func duplicateHookError(hook string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noDuplicateHook",
		Message:   "Duplicate " + hook + " in describe block",
		Line:      line,
		Column:    column,
	}
}

func TestNoDuplicateHooksUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_duplicate_hooks.NoDuplicateHooksRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe("foo", () => {
  beforeEach(() => {})
  test("bar", () => { someFn() })
})`},
			{Code: `beforeEach(() => {})
test("bar", () => { someFn() })`},
			{Code: `describe("foo", () => {
  beforeAll(() => {})
  beforeEach(() => {})
  afterEach(() => {})
  afterAll(() => {})
  test("bar", () => { someFn() })
})`},
			{Code: `describe.skip("foo", () => { beforeEach(() => {}); beforeAll(() => {}) })
describe("bar", () => { beforeEach(() => {}); beforeAll(() => {}) })`},
			{Code: `describe("foo", () => {
  beforeEach(() => {})
  describe("inner", () => { beforeEach(() => {}) })
})`},
			{Code: `describe.each(['hello'])('%s', () => { beforeEach(() => {}) })`},
			{Code: `describe('outer', () => {
  describe.each(['hello'])('%s', () => { beforeEach(() => {}) })
  describe.each(['world'])('%s', () => { beforeEach(() => {}) })
})`},
			{Code: "describe.each``('%s', () => { beforeEach(() => {}) })"},
			{Code: "describe('outer', () => { describe.each``('%s', () => { beforeEach(() => {}) }); describe.each``('%s', () => { beforeEach(() => {}) }) })"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `describe("foo", () => {
  beforeEach(() => {})
  beforeEach(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 3)},
			},
			{
				Code: `describe.skip("foo", () => {
  beforeEach(() => {})
  beforeAll(() => {})
  beforeAll(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeAll", 4, 3)},
			},
			{
				Code: `describe.skip("foo", () => {
  afterEach(() => {})
  afterEach(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterEach", 3, 3)},
			},
			{
				Code: `import { afterEach } from '@rstest/core'
describe.skip("foo", () => {
  afterEach(() => {})
  afterEach(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterEach", 4, 3)},
			},
			{
				Code: `import { afterEach, afterEach as cleanup } from '@rstest/core'
describe.skip("foo", () => {
  afterEach(() => {})
  cleanup(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterEach", 4, 3)},
			},
			{
				Code: `describe.skip("foo", () => {
  afterAll(() => {})
  afterAll(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterAll", 3, 3)},
			},
			{
				Code: `afterAll(() => {})
afterAll(() => {})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("afterAll", 2, 1)},
			},
			{
				Code: `describe("foo", () => {
  beforeEach(() => {})
  beforeEach(() => {})
  beforeEach(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					duplicateHookError("beforeEach", 3, 3),
					duplicateHookError("beforeEach", 4, 3),
				},
			},
			{
				Code: `describe.skip("foo", () => {
  afterAll(() => {})
  afterAll(() => {})
  beforeAll(() => {})
  beforeAll(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					duplicateHookError("afterAll", 3, 3),
					duplicateHookError("beforeAll", 5, 3),
				},
			},
			{
				Code: `describe.skip("first", () => { beforeEach(() => {}) })
describe("second", () => {
  beforeEach(() => {})
  beforeEach(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 4, 3)},
			},
			{
				Code: `describe("outer", () => {
  beforeAll(() => {})
  describe("inner", () => {
    beforeEach(() => {})
    beforeEach(() => {})
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 5, 5)},
			},
			{
				Code: `describe.each(['hello'])('%s', () => {
  beforeEach(() => {})
  beforeEach(() => {})
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 3, 3)},
			},
			{
				Code: `describe('outer', () => {
  describe.each(['hello'])('%s', () => { beforeEach(() => {}) })
  describe.each(['world'])('%s', () => {
    beforeEach(() => {})
    beforeEach(() => {})
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 5, 5)},
			},
			{
				Code: `describe('outer', () => {
  describe.each(['world'])('%s', () => {
    describe('inner', () => {
      beforeEach(() => {})
      beforeEach(() => {})
    })
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 5, 7)},
			},
			{
				Code:   "describe.each``('%s', () => { beforeEach(() => {}); beforeEach(() => {}) })",
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 1, 53)},
			},
			{
				Code:   "describe('outer', () => { describe.each``('%s', () => { beforeEach(() => {}); beforeEach(() => {}) }) })",
				Errors: []rule_tester.InvalidTestCaseError{duplicateHookError("beforeEach", 1, 79)},
			},
		},
	)
}
