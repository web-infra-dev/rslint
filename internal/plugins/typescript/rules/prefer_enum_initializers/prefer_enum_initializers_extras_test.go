// TestPreferEnumInitializersExtras locks in branches, edge shapes, and
// real-user code patterns that the upstream test suite does not exercise.
// Cases are grouped by category via the `// ---- <kind>: <what> ----` marker
// comments so a reviewer can grep by category, but they all live in one file
// (the rule is simple enough that the per-area file split prescribed by
// PORT_RULE.md for >80-case extras suites would be overkill here).
package prefer_enum_initializers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEnumInitializersExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferEnumInitializersRule, []rule_tester.ValidTestCase{

		// ---- Dimension 4: empty container — `enum X {}` has zero members, the
		//      forEach loop body never runs, no diagnostic.
		{Code: `enum Empty {}`},
		{Code: `const enum EmptyConst {}`},
		{Code: `declare enum EmptyDeclare {}`},

		// ---- Dimension 4: const enum — same listener path, every member has
		//      an initializer so the report branch is unreached.
		{Code: `
const enum Status {
  Open = 1,
  Close = 2,
}
`},

		// ---- Dimension 4: declare-ambient enum — body is allowed in declare,
		//      explicit initializers keep the rule silent.
		{Code: `
declare enum AmbientOK {
  A = 0,
  B = 1,
}
`},

		// ---- Dimension 4: quoted-string member name — initialized, no report.
		{Code: `
enum Quoted {
  'Up' = 1,
  'Down' = 2,
}
`},

		// ---- Dimension 4: mixed-shape value expressions on every member —
		//      computed-key initializers, template literals, unary minus.
		//      Rule only cares about initializer-presence, not its shape.
		{Code: `
enum Mixed {
  A = -1,
  B = ` + "`" + `b` + "`" + `,
  C = 1 << 2,
}
`},

		// ---- Dimension 4: `undefined` keyword as initializer — `member.Initializer`
		//      is a non-nil Identifier node, so the gate ` != nil` correctly skips.
		//      Locks in that "no initializer" is NOT the same as "initializer is undefined".
		{Code: `
enum HasUndefined {
  A = undefined,
  B = undefined,
}
`},

		// ---- Dimension 4: `null` literal as initializer — same as above but with
		//      a NullKeyword AST node. Gate must still treat it as initialized.
		{Code: `
enum HasNull {
  A = null,
  B = null,
}
`},

		// ---- Dimension 4: numeric `0` initializer — semantically equivalent to
		//      the default for the first member but still counts as explicit.
		{Code: `
enum ExplicitZero {
  A = 0,
  B = 1,
}
`},

		// ---- Dimension 4: leading JSDoc / block comment between brace and
		//      first member — must not affect listener gating. Comment is
		//      leading trivia of the member node.
		{Code: `
enum Documented {
  /** doc for A */
  A = 1,
  /* leading */ B = 2,
}
`},

		// ---- Dimension 2: enum inside a namespace — listener still fires; all
		//      members initialized, so it stays valid.
		{Code: `
namespace NS {
  export enum Inner {
    A = 1,
    B = 2,
  }
}
`},

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

		// ---- Real-user: enum in a function body — listener fires regardless of
		//      enclosing function scope.
		{Code: `
function makeEnum() {
  enum Local {
    X = 1,
    Y = 2,
  }
  return Local;
}
`},

		// ---- Real-user: triply-nested namespaces — listener still fires at
		//      arbitrary depth.
		{Code: `
namespace Outer {
  export namespace Middle {
    export namespace Inner {
      export enum Deep {
        X = 1,
        Y = 2,
      }
    }
  }
}
`},
	}, []rule_tester.InvalidTestCase{

		// ---- Dimension 4: `declare` (ambient) enum — listener fires on
		//      KindEnumDeclaration regardless of declare modifier; uninitialized
		//      members are still reported. Locked in here because the question
		//      "should declare enum be silently skipped?" recurs (it sounds
		//      reasonable for `.d.ts` ergonomics) but upstream does NOT skip:
		//      empirically `npx eslint` on `declare enum X { Up, Down }` with
		//      @typescript-eslint/prefer-enum-initializers produces 2 errors.
		//      Skipping `declare` would silently diverge from upstream.
		{
			Code: `
declare enum AmbientUninit {
  Up,
  Down,
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
declare enum AmbientUninit {
  Up = 0,
  Down,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
declare enum AmbientUninit {
  Up = 1,
  Down,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
declare enum AmbientUninit {
  Up = 'Up',
  Down,
}
`},
					},
				},
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
declare enum AmbientUninit {
  Up,
  Down = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
declare enum AmbientUninit {
  Up,
  Down = 2,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
declare enum AmbientUninit {
  Up,
  Down = 'Down',
}
`},
					},
				},
			},
		},

		// ---- Dimension 4: const enum — same listener fires on KindEnumDeclaration
		//      regardless of `const` modifier; uninitialized member still reported.
		{
			Code: `
const enum Bits {
  Read,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
const enum Bits {
  Read = 0,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
const enum Bits {
  Read = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
const enum Bits {
  Read = 'Read',
}
`},
					},
				},
			},
		},

		// ---- Dimension 4: quoted-string member name — upstream uses
		//      sourceCode.getText(member) which returns the raw source text
		//      INCLUDING the surrounding quotes. The replacement therefore
		//      reproduces the quotes verbatim. This locks in the
		//      character-fidelity contract; the alternative "strip quotes for
		//      the suggestion" would diverge from upstream.
		{
			Code: `
enum Q {
  'Up',
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum Q {
  'Up' = 0,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Q {
  'Up' = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Q {
  'Up' = '` + `'Up'` + `',
}
`},
					},
				},
			},
		},

		// ---- Dimension 2: enum inside a namespace — listener should still fire
		//      and report the uninitialized inner member, untouched by outer
		//      container kind.
		{
			Code: `
namespace NS {
  export enum Inner {
    A,
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    5,
					EndLine:   4,
					EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
namespace NS {
  export enum Inner {
    A = 0,
  }
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
namespace NS {
  export enum Inner {
    A = 1,
  }
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
namespace NS {
  export enum Inner {
    A = 'A',
  }
}
`},
					},
				},
			},
		},

		// ---- Dimension 4: same-kind nesting boundary — outer enum has every
		//      member initialized but a sibling inner enum (inside a namespace
		//      sharing the file) has none. Only the inner one should report,
		//      proving the listener does not "bleed" across declarations.
		{
			Code: `
enum Outer {
  A = 0,
  B = 1,
}
namespace M {
  export enum Inner {
    X,
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      8,
					Column:    5,
					EndLine:   8,
					EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum Outer {
  A = 0,
  B = 1,
}
namespace M {
  export enum Inner {
    X = 0,
  }
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Outer {
  A = 0,
  B = 1,
}
namespace M {
  export enum Inner {
    X = 1,
  }
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Outer {
  A = 0,
  B = 1,
}
namespace M {
  export enum Inner {
    X = 'X',
  }
}
`},
					},
				},
			},
		},

		// ---- Dimension 4: graceful degradation — a member that is the ONLY
		//      member of an enum still gets the standard 3-suggestion set, and
		//      the report node's text range covers exactly the identifier.
		{
			Code: `enum Solo { Alone }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `enum Solo { Alone = 0 }`},
						{MessageId: "defineInitializerSuggestion", Output: `enum Solo { Alone = 1 }`},
						{MessageId: "defineInitializerSuggestion", Output: `enum Solo { Alone = 'Alone' }`},
					},
				},
			},
		},

		// ---- Dimension 4: leading block comment between brace and member.
		//      tsgo's node.Pos() includes the comment as leading trivia;
		//      `utils.TrimmedNodeText` must skip it so `name == "Up"`. The fix
		//      then replaces ONLY `Up`, leaving the comment intact in the
		//      output. This is the equivalent of upstream's `getText` reading
		//      from `range[0]` which is after the comment by construction.
		{
			Code: `
enum WithComment {
  /* leading */ Up,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Column:    17,
					EndLine:   3,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithComment {
  /* leading */ Up = 0,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithComment {
  /* leading */ Up = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithComment {
  /* leading */ Up = 'Up',
}
`},
					},
				},
			},
		},

		// ---- Dimension 4: `undefined` keyword as initializer locked in on the
		//      INVALID side — `member.Initializer` is non-nil for `Up`, so it's
		//      skipped; `Down` (no initializer) reports at index 1. Proves the
		//      gate does NOT confuse "= undefined" with "no initializer".
		{
			Code: `
enum WithUndefined {
  Up = undefined,
  Down,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithUndefined {
  Up = undefined,
  Down = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithUndefined {
  Up = undefined,
  Down = 2,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithUndefined {
  Up = undefined,
  Down = 'Down',
}
`},
					},
				},
			},
		},

		// ---- Dimension 4: `null` literal as initializer locked in on the
		//      INVALID side — `member.Initializer` is non-nil NullKeyword, gate
		//      skips `Up`; `Down` reports.
		{
			Code: `
enum WithNull {
  Up = null,
  Down,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithNull {
  Up = null,
  Down = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithNull {
  Up = null,
  Down = 2,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum WithNull {
  Up = null,
  Down = 'Down',
}
`},
					},
				},
			},
		},

		// ---- Dimension 4: non-ASCII identifier — UTF-16 column accounting on
		//      the diagnostic range, and TrimmedNodeText must return the raw
		//      bytes (Go string slice over the source file). The suggestion
		//      reproduces the same name verbatim, so `中文 = '中文'` round-trips
		//      cleanly.
		{
			Code: `
enum Unicode {
  中文,
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
enum Unicode {
  中文 = 0,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Unicode {
  中文 = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum Unicode {
  中文 = '中文',
}
`},
					},
				},
			},
		},

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

		// ---- Real-user: mixed initialized/uninitialized members at arbitrary
		//      indices — the implicit-numbering pitfall the rule documentation
		//      cites as motivation. Only the uninitialized members report.
		{
			Code: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted,
  NoContent,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted = 2,
  NoContent,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted = 3,
  NoContent,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted = 'Accepted',
  NoContent,
}
`},
					},
				},
				{
					MessageId: "defineInitializer",
					Line:      6,
					Column:    3,
					EndLine:   6,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted,
  NoContent = 3,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted,
  NoContent = 4,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted,
  NoContent = 'NoContent',
}
`},
					},
				},
			},
		},

		// ---- Real-user: exported enum — common shape in declaration files.
		//      Listener fires regardless of `export` modifier.
		{
			Code: `
export enum LogLevel {
  Trace,
  Debug,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace = 0,
  Debug,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace = 1,
  Debug,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace = 'Trace',
  Debug,
}
`},
					},
				},
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace,
  Debug = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace,
  Debug = 2,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace,
  Debug = 'Debug',
}
`},
					},
				},
			},
		},

		// ---- Real-user: enum inside a function body — listener fires
		//      regardless of enclosing scope.
		{
			Code: `
function makeEnum() {
  enum Local {
    X,
  }
  return Local;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    5,
					EndLine:   4,
					EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
function makeEnum() {
  enum Local {
    X = 0,
  }
  return Local;
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
function makeEnum() {
  enum Local {
    X = 1,
  }
  return Local;
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
function makeEnum() {
  enum Local {
    X = 'X',
  }
  return Local;
}
`},
					},
				},
			},
		},

		// ---- Real-user: triply-nested namespaces with a broken inner enum —
		//      listener must fire at arbitrary depth, no early-exit.
		{
			Code: `
namespace A {
  export namespace B {
    export namespace C {
      export enum Deep {
        X,
      }
    }
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      6,
					Column:    9,
					EndLine:   6,
					EndColumn: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
namespace A {
  export namespace B {
    export namespace C {
      export enum Deep {
        X = 0,
      }
    }
  }
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
namespace A {
  export namespace B {
    export namespace C {
      export enum Deep {
        X = 1,
      }
    }
  }
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
namespace A {
  export namespace B {
    export namespace C {
      export enum Deep {
        X = 'X',
      }
    }
  }
}
`},
					},
				},
			},
		},
	})
}
