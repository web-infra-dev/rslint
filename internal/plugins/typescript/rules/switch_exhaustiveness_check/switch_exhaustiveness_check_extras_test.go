// TestSwitchExhaustivenessCheckExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without
// breaking a named lock-in.
package switch_exhaustiveness_check

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const escapedEnumSuggestionOutput = `
enum Escaped {
  'a\\b' = 1,
  x = 2,
}
declare const value: Escaped;
switch (value) {
  case Escaped.x:
    break;
  case Escaped['a\\b']: { throw new Error('Not implemented yet: Escaped[\'a\\\\b\'] case') }
}
`

const loneSurrogateEnumSuggestionOutput = `
enum LoneSurrogate {
  '\uD800' = 1,
  x = 2,
}
declare const value: LoneSurrogate;
switch (value) {
  case LoneSurrogate.x:
    break;
  case LoneSurrogate['\uD800']: { throw new Error('Not implemented yet: LoneSurrogate[\'\\uD800\'] case') }
}
`

const qualifiedEscapedEnumSuggestionOutput = `
namespace A {
  export enum B {
    'x-y' = 1,
    z = 2,
  }
}
declare const value: A.B;
switch (value) {
  case A.B.z:
    break;
  case A.B['x-y']: { throw new Error('Not implemented yet: A.B[\'x-y\'] case') }
}
`

const importedEscapedEnumSuggestionOutput = `
import { QuotedEnum as Quoted } from './switch-exhaustiveness-check-quoted';
declare const value: Quoted;
switch (value) {
  case Quoted.z:
    break;
  case Quoted['x-y']: { throw new Error('Not implemented yet: Quoted[\'x-y\'] case') }
}
`

const importedUniqueSymbolSuggestionOutput = `
import * as values from './switch-exhaustiveness-check-quoted';
declare const value: typeof values.uniqueA | typeof values.uniqueB;
switch (value) {
  case values.uniqueA:
    break;
  case values.uniqueB: { throw new Error('Not implemented yet: values.uniqueB case') }
}
`

const objectPropertyUniqueSymbolSuggestionOutput = `
declare const values: {
  readonly a: unique symbol;
  readonly b: unique symbol;
};
declare const value: typeof values.a | typeof values.b;
switch (value) {
  case values.a:
    break;
  case values.b: { throw new Error('Not implemented yet: values.b case') }
}
`

const staticPropertyUniqueSymbolSuggestionOutput = `
declare class Values {
  static readonly a: unique symbol;
  static readonly b: unique symbol;
}
declare const value: typeof Values.a | typeof Values.b;
switch (value) {
  case Values.a:
    break;
  case Values.b: { throw new Error('Not implemented yet: Values.b case') }
}
`

const privateStaticUniqueSymbolSuggestionOutput = `
class Values {
  private static readonly a: unique symbol = Symbol();
  private static readonly b: unique symbol = Symbol();
  static test(value: typeof Values.a | typeof Values.b) {
    switch (value) {
      case Values.a:
        break;
      case Values.b: { throw new Error('Not implemented yet: Values.b case') }
    }
  }
}
`

const underscoredUniqueSymbolSuggestionOutput = `
declare const __x: unique symbol;
declare const other: unique symbol;
declare const value: typeof __x | typeof other;
switch (value) {
  case other:
    break;
  case __x: { throw new Error('Not implemented yet: __x case') }
}
`

func TestSwitchExhaustivenessCheckExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized discriminant (single level) ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch ((value)) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: parenthesized discriminant (multi-level) ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch (((value))) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: empty defaultCaseCommentPattern matches any comment ----
			// An explicitly configured empty pattern compiles to a regex that matches
			// every comment, so the trailing comment stands in for a `default:` clause.
			// Verified against ESLint: reports nothing here.
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  // whatever
}
`,
				Options: map[string]any{
					"considerDefaultExhaustiveForUnions": true,
					"defaultCaseCommentPattern":          "",
				},
			},
			// ---- Dimension 4: TS non-null assertion discriminant ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: Direction | undefined;
switch (value!) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: "as" type-expression discriminant ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: unknown;
switch (value as Direction) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: "satisfies" discriminant ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: 'north' | 'south';
switch (value satisfies Direction) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: optional-chain discriminant ----
			{Code: `
declare const obj: { prop: 'a' | 'b' } | undefined;
switch (obj?.prop) {
  case 'a':
    break;
  case 'b':
    break;
  case undefined:
    break;
}
`},
			// ---- Dimension 4: parenthesized case-test expression ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch (value) {
  case ('north'):
    break;
  case ('south'):
    break;
}
`},
			// ---- Dimension 4: fallthrough (grouped) case clauses cover every constituent ----
			{Code: `
type Direction = 'north' | 'south' | 'east' | 'west';
declare const value: Direction;
switch (value) {
  case 'north':
  case 'south':
    break;
  case 'east':
  case 'west':
    break;
}
`},
			// ---- Dimension 4: nested switch — inner switch's cases don't count toward the outer switch ----
			{Code: `
type Outer = 'a' | 'b';
type Inner = 'x' | 'y';
declare const outer: Outer;
declare const inner: Inner;
switch (outer) {
  case 'a':
    switch (inner) {
      case 'x':
        break;
      case 'y':
        break;
    }
    break;
  case 'b':
    break;
}
`},
			// ---- Real-user: literal case matching an intersection-branded union constituent behind a type alias (issue tracker pattern for branded types) ----
			{Code: `
type Brand<T, B extends string> = T & { readonly __brand: B };
type UserId = Brand<string, 'UserId'>;
type Status = 'active' | 'inactive';
declare const value: Status | UserId;
switch (value) {
  case 'active':
    break;
  case 'inactive':
    break;
  default:
    break;
}
`, Options: map[string]any{"allowDefaultCaseForExhaustiveSwitch": true}},
			// JavaScript trim removes U+FEFF around comment values. Go's
			// strings.TrimSpace does not, so this locks the ECMAScript helper.
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  // ` + "\uFEFF" + `no default` + "\uFEFF" + `
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": true},
			},
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  /* ` + "\uFEFF" + `no default` + "\uFEFF" + ` */
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": true},
			},
			// JavaScript trim does not remove U+0085 (NEL), unlike
			// strings.TrimSpace, so this is not an effective default comment.
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  case 'b':
    break;
  // ` + "\u0085" + `no default` + "\u0085" + `
}
`,
				Options: map[string]any{"allowDefaultCaseForExhaustiveSwitch": false},
			},
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  case 'b':
    break;
  /* ` + "\u0085" + `no default` + "\u0085" + ` */
}
`,
				Options: map[string]any{"allowDefaultCaseForExhaustiveSwitch": false},
			},
			// Configured patterns use real ECMAScript `u` semantics, including
			// Unicode properties that regexp2 rejects.
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  // 日本語
}
`,
				Options: map[string]any{
					"considerDefaultExhaustiveForUnions": true,
					"defaultCaseCommentPattern":          `^\p{Letter}+$`,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized discriminant reports the inner expression ----
			// ESTree has no parenthesized-expression node, so ESLint's `node.discriminant`
			// is always the unwrapped expression. Verified against ESLint: the diagnostic
			// covers `value` (4:10-4:15), not `(value)`.
			{
				Code: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch ((value)) {
  case 'north':
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    10,
						EndLine:   4,
						EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch ((value)) {
  case 'north':
    break;
  case "south": { throw new Error('Not implemented yet: "south" case') }
}
`,
							},
						},
					},
				},
			},
			// ---- Dimension 4: multi-level parenthesized discriminant unwraps fully ----
			// Verified against ESLint: the diagnostic covers the innermost `value`
			// (4:11-4:16).
			{
				Code: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch (((value))) {
  case 'north':
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    11,
						EndLine:   4,
						EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch (((value))) {
  case 'north':
    break;
  case "south": { throw new Error('Not implemented yet: "south" case') }
}
`,
							},
						},
					},
				},
			},
			// ---- Dimension 4: computed enum member name via multi-line template literal ----
			{
				Code: `
enum Enum {
  'a' = 1,
  ` + "[`key-with\n\n          new-line`]" + ` = 2,
}

declare const a: Enum;

switch (a) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum Enum {
  'a' = 1,
  ` + "[`key-with\n\n          new-line`]" + ` = 2,
}

declare const a: Enum;

switch (a) {
case Enum.a: { throw new Error('Not implemented yet: Enum.a case') }
case Enum['key-with\n\n          new-line']: { throw new Error('Not implemented yet: Enum[\'key-with\\n\\n          new-line\'] case') }
}
`,
							},
						},
					},
				},
			},
			// ---- Real-user (typescript-eslint#11842): a bare numeric literal `case 0:` does not
			// count as covering a numeric enum member with the same underlying value — TS treats
			// enum-member types and plain numeric-literal types as distinct type objects even when
			// their runtime values match, so the Set-based case-type lookup in getSwitchMetadata
			// (keyed by *checker.Type identity) does not consider it covered. This locks in the
			// checker-identity-based caseTypes.has(...) branch in getSwitchMetadata, matching
			// upstream's real reported behavior (closed as working-as-intended: use the qualified
			// enum member in the case test). ----
			{
				Code: `
enum DataType {
  First = 0,
  Second = 1,
}

function test(type: DataType) {
  switch (type) {
    case 0:
      break;
    case 1:
      break;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: DataType.First | DataType.Second",
						Line:      8,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum DataType {
  First = 0,
  Second = 1,
}

function test(type: DataType) {
  switch (type) {
    case 0:
      break;
    case 1:
      break;
    case DataType.First: { throw new Error('Not implemented yet: DataType.First case') }
    case DataType.Second: { throw new Error('Not implemented yet: DataType.Second case') }
  }
}
`,
							},
						},
					},
				},
			},
			// ---- Branch lock-in: checkSwitchUnnecessaryDefaultCase fires against a
			// comment-based default (not just a real `default:` clause). The switch below is
			// fully covered by explicit literal cases and the trailing comment matches the
			// (default) "no default" pattern, so getSwitchMetadata's defaultCase resolves via
			// getCommentDefaultCase rather than a real clause — this locks in that
			// checkSwitchUnnecessaryDefaultCase's `meta.defaultCase.exists()` check treats a
			// comment-based default the same as a real one, and that defaultCaseRef.reportRange
			// anchors the diagnostic on the comment when there is no clause node. ----
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  case 'b':
    break;
  // no default
}
`,
				Options: map[string]any{"allowDefaultCaseForExhaustiveSwitch": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase", Line: 8, Column: 3, EndLine: 8, EndColumn: 16},
				},
			},
			// ---- requiresQuoting walks UTF-16 code units: an enum member named with a
			// character outside the BMP is a surrogate pair there, and neither half is an
			// identifier character, so the suggestion must bracket-quote it. Verified
			// against ESLint: suggests `case Astral['𐐀']`. ----
			{
				Code: `
enum Astral {
  '𐐀' = 1,
  b = 2,
}

declare const value: Astral;

switch (value) {
  case Astral.b:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum Astral {
  '𐐀' = 1,
  b = 2,
}

declare const value: Astral;

switch (value) {
  case Astral.b:
    break;
  case Astral['𐐀']: { throw new Error('Not implemented yet: Astral[\'𐐀\'] case') }
}
`,
							},
						},
					},
				},
			},
			// ---- Branch lock-in: getCommentDefaultCase matches a block comment, not just a
			// line comment (the `/*...*/`-stripping branch in getCommentDefaultCase). ----
			{
				Code: `
declare const value: number;
switch (value) {
  case 0:
    break;
  case 1:
    break;
  /* no default */
}
`,
				Options: map[string]any{"requireDefaultForNonUnion": true},
				Errors:  nil,
			},
			// ES5 uses TypeScript's older Unicode identifier table. U+037F is
			// valid in modern targets but must be bracket-quoted for ES5 output.
			{
				Code: `
enum E {
  Ϳ = 1,
  x = 2,
}
declare const value: E;
switch (value) {
  case E.x:
    break;
}
`,
				TSConfig: "tsconfig.es5.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum E {
  Ϳ = 1,
  x = 2,
}
declare const value: E;
switch (value) {
  case E.x:
    break;
  case E['Ϳ']: { throw new Error('Not implemented yet: E[\'Ϳ\'] case') }
}
`,
							},
						},
					},
				},
			},
			// NEL remains in the comment value under ECMAScript trim, so it
			// cannot suppress a missing-union diagnostic.
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  // ` + "\u0085" + `no default` + "\u0085" + `
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  case "b": { throw new Error('Not implemented yet: "b" case') }
  // ` + "\u0085" + `no default` + "\u0085" + `
}
`,
							},
						},
					},
				},
			},
			// A FEFF-trimmed effective default is also the insertion anchor for
			// a requested missing-case suggestion.
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  // ` + "\uFEFF" + `no default` + "\uFEFF" + `
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  case "b": { throw new Error('Not implemented yet: "b" case') }
  // ` + "\uFEFF" + `no default` + "\uFEFF" + `
}
`,
							},
						},
					},
				},
			},
			// A raw backslash in an enum member name must be escaped again in
			// the generated bracket string; otherwise `\b` changes its value.
			{
				Code: `
enum Escaped {
  'a\\b' = 1,
  x = 2,
}
declare const value: Escaped;
switch (value) {
  case Escaped.x:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: escapedEnumSuggestionOutput},
						},
					},
				},
			},
			// tsgo stores a lone UTF-16 surrogate as CESU-8 in Symbol.Name. The
			// shared string escaper must restore a valid `\uXXXX` source escape.
			{
				Code: `
enum LoneSurrogate {
  '\uD800' = 1,
  x = 2,
}
declare const value: LoneSurrogate;
switch (value) {
  case LoneSurrogate.x:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: loneSurrogateEnumSuggestionOutput},
						},
					},
				},
			},
			// The enum container must remain qualified when a string-named
			// member needs bracket notation. Upstream incorrectly emits bare B.
			{
				Code: `
namespace A {
  export enum B {
    'x-y' = 1,
    z = 2,
  }
}
declare const value: A.B;
switch (value) {
  case A.B.z:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: qualifiedEscapedEnumSuggestionOutput},
						},
					},
				},
			},
			// SymbolToStringEx must honor the import alias visible at the switch,
			// rather than printing the original export name or module internals.
			{
				FileName: "imported-escaped-enum.ts",
				Code: `
import { QuotedEnum as Quoted } from './switch-exhaustiveness-check-quoted';
declare const value: Quoted;
switch (value) {
  case Quoted.z:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: importedEscapedEnumSuggestionOutput},
						},
					},
				},
			},
			// Unique-symbol suggestions also need a runtime expression resolved
			// in the current scope; the raw symbol name is unbound after a
			// namespace import.
			{
				FileName: "imported-unique-symbol.ts",
				Code: `
import * as values from './switch-exhaustiveness-check-quoted';
declare const value: typeof values.uniqueA | typeof values.uniqueB;
switch (value) {
  case values.uniqueA:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: typeof uniqueB",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: importedUniqueSymbolSuggestionOutput},
						},
					},
				},
			},
			// Properties of an accessible object or class may carry unique-symbol
			// types without a named namespace parent. Keep the full value path.
			{
				Code: `
declare const values: {
  readonly a: unique symbol;
  readonly b: unique symbol;
};
declare const value: typeof values.a | typeof values.b;
switch (value) {
  case values.a:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: typeof b",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: objectPropertyUniqueSymbolSuggestionOutput},
						},
					},
				},
			},
			{
				Code: `
declare class Values {
  static readonly a: unique symbol;
  static readonly b: unique symbol;
}
declare const value: typeof Values.a | typeof Values.b;
switch (value) {
  case Values.a:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: typeof b",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: staticPropertyUniqueSymbolSuggestionOutput},
						},
					},
				},
			},
			{
				Code: `
class Values {
  private static readonly a: unique symbol = Symbol();
  private static readonly b: unique symbol = Symbol();
  static test(value: typeof Values.a | typeof Values.b) {
    switch (value) {
      case Values.a:
        break;
    }
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: typeof b",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: privateStaticUniqueSymbolSuggestionOutput},
						},
					},
				},
			},
			// A bare name printed for a type-only property is not a runtime path:
			// unrelated values with the same name must not make it look accessible.
			{
				Code: `
type Values = {
  readonly a: unique symbol;
  readonly b: unique symbol;
};
declare const value: Values['a'] | Values['b'];
const a = 0;
const b = 1;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			// A checker-formatted qualified path is not sufficient proof by itself:
			// an unrelated local value can shadow the declaration that owns the
			// missing property. Suppress the semantically invalid suggestion.
			{
				Code: `
export {};
declare const values: {
  readonly a: unique symbol;
  readonly b: unique symbol;
};
type Value = typeof values.a | typeof values.b;
function test(value: Value) {
  const values = { a: 1, b: 2 };
  switch (value) {
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			// The same identity check applies to an enum container hidden by a
			// local object with the same qualified spelling.
			{
				Code: `
export {};
namespace N {
  export enum E { A, B }
}
type Value = N.E;
function test(value: Value) {
  const N = { E: { A: 'a', B: 'b' } };
  switch (value) {
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			// TypeScript's escapedName turns a source `__x` into `___x`; retain
			// the actual bound identifier in both the diagnostic and suggestion.
			{
				Code: `
declare const __x: unique symbol;
declare const other: unique symbol;
declare const value: typeof __x | typeof other;
switch (value) {
  case other:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: typeof __x",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: underscoredUniqueSymbolSuggestionOutput},
						},
					},
				},
			},
			// A type-only import has no runtime binding that a generated case can
			// safely reference. Keep the diagnostic, but suppress the suggestion.
			{
				FileName: "type-only-unique-symbol.ts",
				Code: `
import type { uniqueA, uniqueB } from './switch-exhaustiveness-check-quoted';
declare const value: typeof uniqueA | typeof uniqueB;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			{
				FileName: "type-only-enum.ts",
				Code: `
import type { QuotedEnum as Quoted } from './switch-exhaustiveness-check-quoted';
declare const value: Quoted;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			// Type-only status is transitive through barrel re-exports even when
			// the importing file uses an ordinary import declaration.
			{
				FileName: "transitive-type-only-unique-symbol.ts",
				Code: `
import { uniqueA, uniqueB } from './switch-exhaustiveness-check-type-only-barrel';
declare const value: typeof uniqueA | typeof uniqueB;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			{
				FileName: "transitive-type-only-enum.ts",
				Code: `
import { QuotedEnum as Quoted } from './switch-exhaustiveness-check-type-only-barrel';
declare const value: Quoted;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			{
				FileName: "namespace-transitive-type-only-unique-symbol.ts",
				Code: `
import * as barrel from './switch-exhaustiveness-check-type-only-barrel';
declare const value: typeof barrel.uniqueA | typeof barrel.uniqueB;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			{
				FileName: "namespace-transitive-type-only-enum.ts",
				Code: `
import * as barrel from './switch-exhaustiveness-check-type-only-barrel';
declare const value: barrel.QuotedEnum;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			{
				FileName: "multi-barrel-type-only-enum.ts",
				Code: `
import { QuotedEnum as Quoted } from './switch-exhaustiveness-check-type-only-barrel-2';
declare const value: Quoted;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			{
				FileName: "star-type-only-unique-symbol.ts",
				Code: `
import { uniqueA, uniqueB } from './switch-exhaustiveness-check-type-only-star';
declare const value: typeof uniqueA | typeof uniqueB;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
		},
	)

	t.Run("EditDemand", testSwitchExhaustivenessCheckEditDemand)
	t.Run("SymbolSuggestionDemandIndependence", testSymbolSuggestionDemandIndependence)
	t.Run("CustomPatternFuseIsDemandIndependent", testCustomPatternFuseIsDemandIndependent)
	t.Run("MissingCaseOrderTracksCheckerCreationOrder", testMissingCaseOrderTracksCheckerCreationOrder)
	t.Run("EscapedEnumSuggestionTypeChecks", func(t *testing.T) {
		for name, output := range map[string]string{
			"backslash":        escapedEnumSuggestionOutput,
			"imported-symbol":  importedUniqueSymbolSuggestionOutput,
			"lone-surrogate":   loneSurrogateEnumSuggestionOutput,
			"object-property":  objectPropertyUniqueSymbolSuggestionOutput,
			"private-static":   privateStaticUniqueSymbolSuggestionOutput,
			"qualified":        qualifiedEscapedEnumSuggestionOutput,
			"static-property":  staticPropertyUniqueSymbolSuggestionOutput,
			"underscored":      underscoredUniqueSymbolSuggestionOutput,
			"imported-aliased": importedEscapedEnumSuggestionOutput,
		} {
			t.Run(name, func(t *testing.T) {
				program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(
					output,
					name+"-enum-suggestion.ts",
					"tsconfig.json",
				)
				if err != nil {
					t.Fatal(err)
				}
				if diagnostics := program.GetSyntacticDiagnostics(context.Background(), sourceFile); len(diagnostics) != 0 {
					t.Fatalf("suggestion has syntactic diagnostics: %#v", diagnostics)
				}
				if diagnostics := program.GetSemanticDiagnostics(context.Background(), sourceFile); len(diagnostics) != 0 {
					t.Fatalf("suggestion has semantic diagnostics: %#v", diagnostics)
				}
			})
		}
	})
	t.Run("TypeOnlyInputsTypeCheck", func(t *testing.T) {
		for name, source := range map[string]string{
			"direct-unique": `
import type { uniqueA, uniqueB } from './switch-exhaustiveness-check-quoted';
declare const value: typeof uniqueA | typeof uniqueB;
switch (value) {}
`,
			"transitive-enum": `
import { QuotedEnum as Quoted } from './switch-exhaustiveness-check-type-only-barrel';
declare const value: Quoted;
switch (value) {}
`,
			"transitive-unique": `
import { uniqueA, uniqueB } from './switch-exhaustiveness-check-type-only-barrel';
declare const value: typeof uniqueA | typeof uniqueB;
switch (value) {}
`,
			"namespace-transitive-enum": `
import * as barrel from './switch-exhaustiveness-check-type-only-barrel';
declare const value: barrel.QuotedEnum;
switch (value) {}
`,
			"namespace-transitive-unique": `
import * as barrel from './switch-exhaustiveness-check-type-only-barrel';
declare const value: typeof barrel.uniqueA | typeof barrel.uniqueB;
switch (value) {}
`,
			"multi-barrel-enum": `
import { QuotedEnum as Quoted } from './switch-exhaustiveness-check-type-only-barrel-2';
declare const value: Quoted;
switch (value) {}
`,
			"star-unique": `
import { uniqueA, uniqueB } from './switch-exhaustiveness-check-type-only-star';
declare const value: typeof uniqueA | typeof uniqueB;
switch (value) {}
`,
			"unrelated-shadow": `
type Values = {
  readonly a: unique symbol;
  readonly b: unique symbol;
};
declare const value: Values['a'] | Values['b'];
const a = 0;
const b = 1;
switch (value) {}
`,
			"qualified-property-shadow": `
export {};
declare const values: {
  readonly a: unique symbol;
  readonly b: unique symbol;
};
type Value = typeof values.a | typeof values.b;
function test(value: Value) {
  const values = { a: 1, b: 2 };
  switch (value) {}
}
`,
			"qualified-enum-shadow": `
export {};
namespace N {
  export enum E { A, B }
}
type Value = N.E;
function test(value: Value) {
  const N = { E: { A: 'a', B: 'b' } };
  switch (value) {}
}
`,
		} {
			t.Run(name, func(t *testing.T) {
				program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(
					source,
					name+"-type-only-input.ts",
					"tsconfig.json",
				)
				if err != nil {
					t.Fatal(err)
				}
				if diagnostics := program.GetSemanticDiagnostics(context.Background(), sourceFile); len(diagnostics) != 0 {
					t.Fatalf("input has semantic diagnostics: %#v", diagnostics)
				}
			})
		}
	})
}

func testSymbolSuggestionDemandIndependence(t *testing.T) {
	const source = `
import { QuotedEnum as Quoted } from './switch-exhaustiveness-check-quoted';
declare const first: Quoted;
switch (first) {
  case Quoted.z:
    break;
}

declare const values: {
  readonly a: unique symbol;
  readonly b: unique symbol;
};
declare const objectValue: typeof values.a | typeof values.b;
switch (objectValue) {
  case values.a:
    break;
}

declare const pre: 'early';
type B1 = { b1: 1 };
type B2 = { b2: 2 };
type I1 = 'late' & B1;
type I2 = typeof pre & B2;
declare const second: I1 | I2;
switch (second) {}
`
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		program, sourceFile, err := helper.CreateTestProgram(source, "symbol-demand-independence.ts", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      lintprogram.NewFromCompiler(program),
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:             SwitchExhaustivenessCheckRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: true,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return SwitchExhaustivenessCheckRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostic.FixesPtr = nil
					diagnostic.Suggestions = nil
					diagnostic.SourceFile = nil
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		return diagnostics
	}

	withoutSuggestions := run(rule.EditDemandNone)
	withSuggestions := run(rule.EditDemandSuggestion)
	if !reflect.DeepEqual(withSuggestions, withoutSuggestions) {
		t.Fatalf("diagnostics changed after accessible-symbol suggestion queries:\nnone: %#v\nsuggestion: %#v", withoutSuggestions, withSuggestions)
	}
	if len(withoutSuggestions) != 3 || withoutSuggestions[2].Message.Data["missingBranches"] != `"late" | "early"` {
		t.Fatalf("diagnostics = %#v, want enum, object-property, and ordered intersection diagnostics", withoutSuggestions)
	}
}

func testMissingCaseOrderTracksCheckerCreationOrder(t *testing.T) {
	lintMissingBranches := func(t *testing.T, source, fileName string, prewarm bool) string {
		t.Helper()
		program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(
			source,
			fileName,
			"tsconfig.json",
		)
		if err != nil {
			t.Fatal(err)
		}
		if prewarm {
			if diagnostics := program.GetSemanticDiagnostics(context.Background(), sourceFile); len(diagnostics) != 0 {
				t.Fatalf("prewarm produced semantic diagnostics: %#v", diagnostics)
			}
		}

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      lintprogram.NewFromCompiler(program),
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:             SwitchExhaustivenessCheckRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: true,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return SwitchExhaustivenessCheckRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: rule.EditDemandNone,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("diagnostics = %#v, want one missing-case diagnostic", diagnostics)
		}
		return diagnostics[0].Message.Data["missingBranches"]
	}

	const source = `
enum Tone {
  Z = 'z',
  A = 'a',
}
type Mixed = false | true | 1 | 2 | 'later' | 'earlier' | Tone.Z | Tone.A;
declare const value: Mixed;
switch (value) {
  case false:
    break;
}
`
	tests := []struct {
		name    string
		prewarm bool
		want    string
	}{
		{name: "cold", want: `true | 1 | 2 | "later" | "earlier" | Tone.Z | Tone.A`},
		{name: "prewarmed", prewarm: true, want: `true | Tone.Z | Tone.A | 1 | 2 | "later" | "earlier"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Recreate the Program several times to prove deterministic output
			// for a given checker lifecycle. TypeScript itself orders these union
			// constituents by creation ID, so a semantic-diagnostics prewarm moves
			// the enum types earlier in both TypeScript 5.9 and tsgo.
			for range 3 {
				if got := lintMissingBranches(t, source, "missing-case-order-"+test.name+".ts", test.prewarm); got != test.want {
					t.Fatalf("missingBranches = %q, want %q", got, test.want)
				}
			}
		})
	}

	// TypeScript orders the outer union by its constituent type IDs but retains
	// insertion order within each intersection. Sorting all flattened leaves
	// would incorrectly move "early" ahead of its earlier-created parent.
	const intersectionSource = `
declare const pre: 'early';
type B1 = { b1: 1 };
type B2 = { b2: 2 };
type I1 = 'late' & B1;
type I2 = typeof pre & B2;
type U = I1 | I2;
declare const value: U;
switch (value) {}
`
	for _, prewarm := range []bool{false, true} {
		name := "cold"
		if prewarm {
			name = "prewarmed"
		}
		t.Run("intersection/"+name, func(t *testing.T) {
			for range 3 {
				if got := lintMissingBranches(t, intersectionSource, "missing-intersection-order-"+name+".ts", prewarm); got != `"late" | "early"` {
					t.Fatalf(`missingBranches = %q, want %q`, got, `"late" | "early"`)
				}
			}
		})
	}
}

func testCustomPatternFuseIsDemandIndependent(t *testing.T) {
	attackComment := strings.Repeat("a", 24) + "!"
	source := `
declare const first: 'a' | 'b';
switch (first) {
  case 'a':
    break;
  // ` + attackComment + `
}

declare const second: 'a' | 'b';
switch (second) {
  case 'a':
    break;
  case 'b':
    break;
  // intentional
}
`
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	options := []any{map[string]any{
		"allowDefaultCaseForExhaustiveSwitch": false,
		"defaultCaseCommentPattern":           `^(a+)+$|^intentional$`,
	}}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		program, sourceFile, err := helper.CreateTestProgram(source, "custom-pattern-fuse.ts", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      lintprogram.NewFromCompiler(program),
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:             SwitchExhaustivenessCheckRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: true,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return SwitchExhaustivenessCheckRule.Run(ctx, options)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostic.FixesPtr = nil
					diagnostic.Suggestions = nil
					diagnostic.SourceFile = nil
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		return diagnostics
	}

	withoutSuggestions := run(rule.EditDemandNone)
	withSuggestions := run(rule.EditDemandSuggestion)
	if !reflect.DeepEqual(withSuggestions, withoutSuggestions) {
		t.Fatalf("diagnostics changed with suggestion demand:\nnone: %#v\nsuggestion: %#v", withoutSuggestions, withSuggestions)
	}
	if len(withoutSuggestions) != 1 || withoutSuggestions[0].Message.Id != "switchIsNotExhaustive" {
		t.Fatalf("diagnostics = %#v, want only the first switch's missing-case diagnostic", withoutSuggestions)
	}
}

func testSwitchExhaustivenessCheckEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	const source = "type T = 1 | 2;\n" +
		"function test(value: T): number {\n" +
		"  switch (value) {\n" +
		"    case 1:\n" +
		"      return 1;\n" +
		"  }\n" +
		"}\n"

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		// Suggestion rendering asks the checker for accessible runtime symbols.
		// Use a fresh Program for every demand so those queries cannot prewarm a
		// later run and hide demand-dependent diagnostics.
		program, sourceFile, err := helper.CreateTestProgram(source, "edit-demand.ts", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		sourceProgram := lintprogram.NewFromCompiler(program)

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      sourceProgram,
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:             SwitchExhaustivenessCheckRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: SwitchExhaustivenessCheckRule.RequiresTypeInfo,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return SwitchExhaustivenessCheckRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnostics := map[rule.EditDemand][]rule.RuleDiagnostic{
		rule.EditDemandNone:       run(rule.EditDemandNone),
		rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
		rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
		rule.EditDemandAll:        run(rule.EditDemandAll),
	}
	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		diagnostic.SourceFile = nil
		return diagnostic
	}

	want := withoutEdits(diagnostics[rule.EditDemandAll][0])
	for demand, demandDiagnostics := range diagnostics {
		if got := withoutEdits(demandDiagnostics[0]); !reflect.DeepEqual(got, want) {
			t.Errorf("diagnostic changed for demand %d:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
		if demandDiagnostics[0].FixesPtr != nil {
			t.Errorf("demand %d unexpectedly has autofixes (rule only emits suggestions)", demand)
		}
	}

	suggestionOnly := diagnostics[rule.EditDemandSuggestion][0].Suggestions
	allEdits := diagnostics[rule.EditDemandAll][0].Suggestions
	if suggestionOnly == nil || !reflect.DeepEqual(suggestionOnly, allEdits) {
		t.Fatalf("suggestions differ between suggestion and all-edits demand")
	}
	if len(*suggestionOnly) != 1 || (*suggestionOnly)[0].Message.Id != "addMissingCases" {
		t.Fatalf("suggestions = %#v, want a single addMissingCases suggestion", *suggestionOnly)
	}

	if diagnostics[rule.EditDemandNone][0].Suggestions != nil ||
		diagnostics[rule.EditDemandAutofix][0].Suggestions != nil {
		t.Errorf("suggestions attached without suggestion demand")
	}
}
