// TestConsistentEachForExtras covers what the migrated suite does not: the
// tagged-template form of both parameterized APIs, the source matrix
// (`@rstest/core` and `@rstest/playwright` imports, globals, aliases, shadowed
// and foreign names), the `test.describe` spelling whose resolved name is not
// `describe`, and every option-parsing branch.
package consistent_each_for

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func expectedError(message string, line int, column int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{
		{
			MessageId: "consistentMethod",
			Message:   message,
			Line:      line,
			Column:    column,
		},
	}
}

func TestConsistentEachForExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentEachForRule,
		[]rule_tester.ValidTestCase{
			// ---- nothing is configured: the rule is a no-op ----
			{Code: `test.for([1, 2])("adds", ([n]) => {})`},
			{Code: `test.for([1, 2])("adds", ([n]) => {})`, Options: map[string]any{}},
			// A key left out governs nothing, so `it` and `describe` stay free
			// while `test` is pinned.
			{
				Code:    `it.for([1, 2])("adds", ([n]) => {}); describe.for([1, 2])("group", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
			},

			// ---- a registration that is not parameterized is never reported ----
			{Code: `test("adds", () => {})`, Options: map[string]any{"test": "for"}},
			{Code: `test.skip("adds", () => {})`, Options: map[string]any{"test": "for"}},
			{Code: `describe("group", () => {})`, Options: map[string]any{"describe": "for"}},
			{Code: `beforeEach(() => {})`, Options: map[string]any{"test": "for"}},

			// ---- the name has to resolve to a Rstest registration ----
			{
				Code:    `import { test } from 'vitest'; test.for([1, 2])("adds", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
			},
			{
				Code:    `const test = { for: (cases: number[]) => (name: string, fn: () => void) => {} }; test.for([1, 2])("adds", () => {})`,
				Options: map[string]any{"test": "each"},
			},

			// A bare `.for` read without invoking it is not a registration, so
			// the alias it initializes resolves to nothing and the call made
			// through it is left alone.
			{
				Code:    `const cases = test.for;` + "\n" + `cases([1, 2])("adds", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
			},

			// ---- both parameterized forms, written the preferred way ----
			{
				Code:    "test.each`\n  a    | b\n  ${1} | ${2}\n`(\"adds $a and $b\", ({ a, b }) => {})",
				Options: map[string]any{"test": "each"},
			},
			{
				Code:    "test.for`\n  a    | b\n  ${1} | ${2}\n`(\"adds $a and $b\", ({ a, b }) => {})",
				Options: map[string]any{"test": "for"},
			},
			{
				Code:    `import { test } from '@rstest/playwright'; test.for([1, 2])("adds", ([n]) => {})`,
				Options: map[string]any{"test": "for"},
			},
			{
				Code:    `import { test } from '@rstest/playwright'; test.describe.each([1, 2])("group", (n) => {})`,
				Options: map[string]any{"describe": "each", "test": "for"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- the tagged-template form is the same registration ----
			{
				Code:    "test.each`\n  a    | b\n  ${1} | ${2}\n`(\"adds $a and $b\", ({ a, b }) => {})",
				Options: map[string]any{"test": "for"},
				Errors:  expectedError("Prefer using `test.for` over `test.each`", 1, 6),
			},
			{
				Code:    "describe.for`\n  a    | b\n  ${1} | ${2}\n`(\"group $a\", ({ a, b }) => {})",
				Options: map[string]any{"describe": "each"},
				Errors:  expectedError("Prefer using `describe.each` over `describe.for`", 1, 10),
			},

			// ---- the source matrix ----
			{
				Code:    `import { test } from '@rstest/core';` + "\n" + `test.for([1, 2])("adds", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
				Errors:  expectedError("Prefer using `test.each` over `test.for`", 2, 6),
			},
			{
				Code:    `import { test } from '@rstest/playwright';` + "\n" + `test.each([1, 2])("adds", (n) => {})`,
				Options: map[string]any{"test": "for"},
				Errors:  expectedError("Prefer using `test.for` over `test.each`", 2, 6),
			},
			// `test.describe` resolves to a suite while keeping the name `test`,
			// so the `describe` preference is what governs it.
			{
				Code:    `import { test } from '@rstest/playwright';` + "\n" + `test.describe.for([1, 2])("group", ([n]) => {})`,
				Options: map[string]any{"describe": "each"},
				Errors:  expectedError("Prefer using `describe.each` over `describe.for`", 2, 15),
			},

			// ---- modifiers between the registration and the accessor ----
			{
				Code:    `test.concurrent.for([1, 2])("adds", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
				Errors:  expectedError("Prefer using `test.each` over `test.for`", 1, 17),
			},
			{
				Code:    `test.skipIf(process.env.CI).for([1, 2])("adds", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
				Errors:  expectedError("Prefer using `test.each` over `test.for`", 1, 29),
			},
			{
				Code:    `test.extend({}).for([1, 2])("adds", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
				Errors:  expectedError("Prefer using `test.each` over `test.for`", 1, 17),
			},

			// ---- the accessor came in through an alias ----
			// The written call site has no `for` to point at, so the diagnostic
			// lands on the identifier that resolves to the parameterized API.
			{
				Code:    `const adds = test.for([1, 2]);` + "\n" + `adds("adds", ([n]) => {})`,
				Options: map[string]any{"test": "each"},
				Errors:  expectedError("Prefer using `test.each` over `test.for`", 2, 1),
			},

			// ---- each key governs its own registrations ----
			{
				Code: `test.for([1, 2])("adds", ([n]) => {})` + "\n" +
					`it.for([1, 2])("adds", ([n]) => {})` + "\n" +
					`describe.each([1, 2])("group", (n) => {})`,
				Options: map[string]any{"test": "each", "it": "each", "describe": "for"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `test.each` over `test.for`",
						Line:      1,
						Column:    6,
					},
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `it.each` over `it.for`",
						Line:      2,
						Column:    4,
					},
					{
						MessageId: "consistentMethod",
						Message:   "Prefer using `describe.for` over `describe.each`",
						Line:      3,
						Column:    10,
					},
				},
			},
		},
	)
}
