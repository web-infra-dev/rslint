// TestConsistentEachForUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 consistent-each-for suite, minus the cases
// built on `suite`, which Rstest does not expose. Rstest-only shapes — the
// tagged-template form, aliases, `@rstest/playwright`, imported and shadowed
// registrations, and the option-parsing branches — live in
// consistent_each_for_extras_test.go.
package consistent_each_for

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentEachForUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentEachForRule,
		[]rule_tester.ValidTestCase{
			{Code: `test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`},
			{Code: `test.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`},
			{Code: `describe.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })`},
			{
				Code:    `test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "each"},
			},
			{
				Code:    `test.skip.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "each"},
			},
			{
				Code:    `test.only.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "each"},
			},
			{
				Code:    `test.concurrent.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "each"},
			},
			{
				Code:    `test.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "for"},
			},
			{
				Code:    `test.skip.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "for"},
			},
			{
				Code:    `test.only.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "for"},
			},
			{
				Code:    `it.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"it": "each"},
			},
			{
				Code:    `it.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"it": "for"},
			},
			{
				Code:    `describe.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })`,
				Options: map[string]any{"describe": "each"},
			},
			{
				Code:    `describe.skip.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })`,
				Options: map[string]any{"describe": "each"},
			},
			{
				Code:    `describe.for([1, 2, 3])("suite", ([n]) => { test("test", () => {}) })`,
				Options: map[string]any{"describe": "for"},
			},
			{
				Code: `test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })
describe.for([1, 2, 3])("suite", ([n]) => { test("test", () => {}) })`,
				Options: map[string]any{"test": "each", "describe": "for"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `test.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "each"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `test.each` over `test.for`",
						Line:      1,
						Column:    6,
					},
				},
			},
			{
				Code:    `test.skip.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "each"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `test.each` over `test.for`",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code:    `test.only.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "each"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `test.each` over `test.for`",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code:    `test.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "for"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `test.for` over `test.each`",
						Line:      1,
						Column:    6,
					},
				},
			},
			{
				Code:    `test.skip.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"test": "for"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `test.for` over `test.each`",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code:    `it.for([1, 2, 3])("test", ([n]) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"it": "each"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `it.each` over `it.for`",
						Line:      1,
						Column:    4,
					},
				},
			},
			{
				Code:    `it.each([1, 2, 3])("test", (n) => { expect(n).toBeDefined() })`,
				Options: map[string]any{"it": "for"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `it.for` over `it.each`",
						Line:      1,
						Column:    4,
					},
				},
			},
			{
				Code:    `describe.for([1, 2, 3])("suite", ([n]) => { test("test", () => {}) })`,
				Options: map[string]any{"describe": "each"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `describe.each` over `describe.for`",
						Line:      1,
						Column:    10,
					},
				},
			},
			{
				Code:    `describe.each([1, 2, 3])("suite", (n) => { test("test", () => {}) })`,
				Options: map[string]any{"describe": "for"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `describe.for` over `describe.each`",
						Line:      1,
						Column:    10,
					},
				},
			},
			{
				Code: `test.for([1, 2])("test1", ([n]) => {})
test.for([3, 4])("test2", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `test.each` over `test.for`",
						Line:      1,
						Column:    6,
					},
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `test.each` over `test.for`",
						Line:      2,
						Column:    6,
					},
				},
			},
		},
	)
}
