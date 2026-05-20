// TestPreferEnumInitializersExtrasBranches locks in every reachable branch in
// the upstream rule body — including branches the upstream test suite itself
// does not exercise (index growth past single digit, declaration merging,
// per-block listener firing, three-consecutive uninit ordering). Each case
// carries `// Locks in upstream TSEnumDeclaration() arm <N>: <what>` so future
// refactors cannot silently flip behaviour without tripping a named lock-in.
package prefer_enum_initializers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEnumInitializersExtrasBranches(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferEnumInitializersRule, []rule_tester.ValidTestCase{
		// ---- Locks in declaration-merging valid path — two blocks of the same
		//      enum name, both fully initialized. Listener fires once per
		//      block; neither reports.
		{Code: `
enum Merged {
  A = 1,
  B = 2,
}
enum Merged {
  C = 3,
}
`},
	}, []rule_tester.InvalidTestCase{
		// ---- Locks in upstream TSEnumDeclaration() arm 1: first-member-index-0
		//      branch. Suggestion 1 uses `index` (0), suggestion 2 uses
		//      `index + 1` (1), suggestion 3 uses `'name'`.
		{
			Code: `
enum Direction {
  Up,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum Direction {
  Up = 0,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Direction {
  Up = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Direction {
  Up = 'Up',
}
`},
					},
				},
			},
		},

		// ---- Locks in upstream forEach index growth: third member uses
		//      `index = 2`, so suggestion 1 = `= 2`, suggestion 2 = `= 3`.
		//      Upstream's own tests only cover indices 0 and 1.
		{
			Code: `
enum E {
  A = 1,
  B = 2,
  C,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum E {
  A = 1,
  B = 2,
  C = 2,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum E {
  A = 1,
  B = 2,
  C = 3,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum E {
  A = 1,
  B = 2,
  C = 'C',
}
`},
					},
				},
			},
		},

		// ---- Locks in multi-digit index — `strconv.Itoa` must produce "10" /
		//      "11", not a single-digit fragment. Upstream's own tests cap at
		//      index 1; this is the first case here that exercises index ≥ 10.
		{
			Code: `
enum Big {
  M0 = 0,
  M1 = 1,
  M2 = 2,
  M3 = 3,
  M4 = 4,
  M5 = 5,
  M6 = 6,
  M7 = 7,
  M8 = 8,
  M9 = 9,
  M10,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      13,
					Column:    3,
					EndLine:   13,
					EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum Big {
  M0 = 0,
  M1 = 1,
  M2 = 2,
  M3 = 3,
  M4 = 4,
  M5 = 5,
  M6 = 6,
  M7 = 7,
  M8 = 8,
  M9 = 9,
  M10 = 10,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Big {
  M0 = 0,
  M1 = 1,
  M2 = 2,
  M3 = 3,
  M4 = 4,
  M5 = 5,
  M6 = 6,
  M7 = 7,
  M8 = 8,
  M9 = 9,
  M10 = 11,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Big {
  M0 = 0,
  M1 = 1,
  M2 = 2,
  M3 = 3,
  M4 = 4,
  M5 = 5,
  M6 = 6,
  M7 = 7,
  M8 = 8,
  M9 = 9,
  M10 = 'M10',
}
`},
					},
				},
			},
		},

		// ---- Locks in declaration merging — same enum name, two declaration
		//      blocks. Each block is its own KindEnumDeclaration AST node and
		//      its index counter starts at 0 PER BLOCK (matching upstream:
		//      `members` comes from `node.body`, scoped to this block, not the
		//      merged enum). So `B` in the second block reports index 0, not
		//      index 2. This is an upstream-shared quirk worth a lock-in.
		{
			Code: `
enum Merged {
  A = 1,
}
enum Merged {
  B,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      6,
					Column:    3,
					EndLine:   6,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum Merged {
  A = 1,
}
enum Merged {
  B = 0,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Merged {
  A = 1,
}
enum Merged {
  B = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Merged {
  A = 1,
}
enum Merged {
  B = 'B',
}
`},
					},
				},
			},
		},

		// ---- Locks in per-block-listener — multiple sibling enums in the same
		//      file. The fully-initialized ones stay silent; the broken one
		//      reports. Proves listeners don't bleed across declarations.
		{
			Code: `
enum First {
  A = 1,
  B = 2,
}
enum Second {
  C,
}
enum Third {
  D = 3,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      7,
					Column:    3,
					EndLine:   7,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum First {
  A = 1,
  B = 2,
}
enum Second {
  C = 0,
}
enum Third {
  D = 3,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum First {
  A = 1,
  B = 2,
}
enum Second {
  C = 1,
}
enum Third {
  D = 3,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum First {
  A = 1,
  B = 2,
}
enum Second {
  C = 'C',
}
enum Third {
  D = 3,
}
`},
					},
				},
			},
		},

		// ---- Locks in three-consecutive-uninit ordering. Members at indices
		//      0/1/2 produce 3 diagnostics IN SOURCE ORDER, each with its
		//      respective `index` and `index+1` suggestions. Upstream's tests
		//      only cap at 2 consecutive uninit members.
		{
			Code: `
enum Triple {
  A,
  B,
  C,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A = 0,
  B,
  C,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A = 1,
  B,
  C,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A = 'A',
  B,
  C,
}
`},
					},
				},
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A,
  B = 1,
  C,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A,
  B = 2,
  C,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A,
  B = 'B',
  C,
}
`},
					},
				},
				{
					MessageId: "defineInitializer",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A,
  B,
  C = 2,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A,
  B,
  C = 3,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Triple {
  A,
  B,
  C = 'C',
}
`},
					},
				},
			},
		},
	})
}
