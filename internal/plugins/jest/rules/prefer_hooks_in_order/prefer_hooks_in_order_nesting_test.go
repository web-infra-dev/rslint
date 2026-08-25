package prefer_hooks_in_order_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_hooks_in_order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func reorderHooksError(currentHook, previousHook string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "reorderHooks",
		Message: fmt.Sprintf(
			"`%s` hooks should be before any `%s` hooks",
			currentHook,
			previousHook,
		),
		Line:   line,
		Column: column,
	}
}

// TestPreferHooksInOrderNesting covers the shapes where a hook call is nested
// inside another hook call's subtree. eslint-plugin-jest tracks a single
// "inside a hook" boolean, so on these shapes it both ends the surrounding run
// early and compares nested hooks against an index that leaked in from the
// enclosing run; every case below asserts the full message so the hook a
// report is attributed to is pinned, not just its position.
func TestPreferHooksInOrderNesting(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_hooks_in_order.PreferHooksInOrderRule,
		[]rule_tester.ValidTestCase{
			// A correctly ordered run inside a hook callback is its own run: it
			// is not compared against the hook that encloses it.
			{Code: `
afterAll(() => {
  beforeEach(() => {});
  afterEach(() => {});
});`},
			{Code: `
beforeAll(() => {
  beforeAll(() => {});
  afterAll(() => {});
});`},
		},
		[]rule_tester.InvalidTestCase{
			// An inverted run nested inside a hook callback is reported on its
			// own terms, against the hook that precedes it at its own level.
			{
				Code: `
afterAll(() => {
  describe('inner', () => {
    afterEach(() => {});
    beforeEach(() => {});
  });
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterEach", 5, 5),
				},
			},
			{
				Code: `
beforeAll(() => {
  afterAll(() => {});
  beforeEach(() => {});
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeEach", "afterAll", 4, 3),
				},
			},
			{
				Code: `
afterAll(() => {
  beforeEach(() => {
    afterEach(() => {});
  });
  beforeAll(() => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "beforeEach", 6, 3),
				},
			},
			// A hook callback does not end the run that surrounds it, whatever
			// the callback contains.
			{
				Code: `
afterAll(() => {
  beforeEach(() => {});
  it('inner', () => {});
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 6, 1),
				},
			},
			{
				Code: `
afterAll(() => {
  describe('inner', () => {
    beforeAll(() => {});
  });
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 7, 1),
				},
			},
			{
				Code: `
afterAll(() => {
  function helper() {
    beforeAll(() => {});
  }
  helper();
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 8, 1),
				},
			},
			{
				Code: `
afterAll(() => {
  beforeEach(() => {});
  afterEach(() => {});
  doSomething();
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 7, 1),
				},
			},
			// The surrounding run keeps its own previous hook, so the report is
			// attributed to the hook that actually precedes it at that level.
			{
				Code: `
afterEach(() => {
  beforeEach(() => {});
  afterAll(() => {});
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterEach", 6, 1),
				},
			},
			{
				Code: `
afterAll(() => {
  describe('d', () => {
    beforeEach(() => {});
  });
  afterEach(() => {});
});
beforeAll(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					reorderHooksError("beforeAll", "afterAll", 8, 1),
				},
			},
		},
	)
}
