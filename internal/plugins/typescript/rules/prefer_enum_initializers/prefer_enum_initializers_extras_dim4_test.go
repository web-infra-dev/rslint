// TestPreferEnumInitializersExtrasDim4 walks the Dimension-4 universal edge
// shapes that the upstream test suite does not exercise. Each case carries an
// inline comment pointing at the specific row / tsgo AST quirk it covers so
// future refactors cannot silently regress them without breaking a named
// lock-in.
package prefer_enum_initializers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEnumInitializersExtrasDim4(t *testing.T) {
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
	}, []rule_tester.InvalidTestCase{
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
	})
}
