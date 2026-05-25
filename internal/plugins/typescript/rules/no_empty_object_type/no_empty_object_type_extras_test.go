// TestNoEmptyObjectTypeExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package no_empty_object_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyObjectTypeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyObjectTypeRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: declaration forms — interface with body members ----
		// Locks in upstream TSInterfaceDeclaration() arm B: body.body.length !== 0
		{Code: `interface Foo { x: number }`},

		// ---- Dimension 4: declaration forms — interface with type parameters and body ----
		{Code: `interface Foo<T> { x: T }`},

		// ---- Dimension 4: nesting — TypeLiteral with members ----
		// Locks in upstream TSTypeLiteral() arm I: members.length truthy
		{Code: `let value: { a: number; b: string };`},

		// ---- Dimension 4: nesting — intersection guards empty `{}` ----
		// Locks in upstream TSTypeLiteral() arm J: parent === TSIntersectionType
		{Code: `type T = number & {};`},
		{Code: `type T = {} & number;`},
		{Code: `type T = number & {} & string;`},

		// ---- Dimension 4: nesting — Brand pattern (canonical use of {} in intersection) ----
		{Code: `type Brand<T, B> = T & { readonly __brand: B };`},

		// ---- Real-user: `extends Record<string, never>` empty-object pattern ----
		{Code: `
interface Base {
  x: number;
}
interface Empty extends Record<string, never> {}
`,
			Options: map[string]interface{}{"allowInterfaces": "with-single-extends"},
		},

		// ---- Real-user: React Props pattern via allowWithName ----
		{
			Code:    `interface MyComponentProps {}`,
			Options: map[string]interface{}{"allowWithName": ".*Props$"},
		},
		{
			Code:    `type MyComponentProps = {};`,
			Options: map[string]interface{}{"allowWithName": ".*Props$"},
		},

		// ---- AllowWithName: regex anchored at start (^Empty) ----
		{
			Code:    `interface EmptyState {}`,
			Options: map[string]interface{}{"allowWithName": "^Empty"},
		},

		// ---- AllowInterfaces: 'always' disables interface listener entirely ----
		// Locks in: top-level `allowInterfaces !== 'always'` guard. With 'always',
		// even `interface Foo extends Base {}` (1 extend) is allowed.
		{
			Code: `
interface Base { x: number }
interface Derived extends Base {}
`,
			Options: map[string]interface{}{"allowInterfaces": "always"},
		},

		// ---- AllowObjectTypes: 'always' disables type-literal listener entirely ----
		{
			Code:    `let x: {} = 0;`,
			Options: map[string]interface{}{"allowObjectTypes": "always"},
		},
		{
			Code:    `type T = {} | null;`,
			Options: map[string]interface{}{"allowObjectTypes": "always"},
		},

		// ---- Both listeners disabled simultaneously ----
		{
			Code: `
interface Foo {}
type Bar = {};
`,
			Options: map[string]interface{}{
				"allowInterfaces":  "always",
				"allowObjectTypes": "always",
			},
		},

		// ---- Mapped-type with members is not an empty TypeLiteral ----
		// The tsgo MappedType is a distinct Kind, so the listener is not invoked.
		{Code: `type T<K extends string> = { [P in K]: number };`},

		// ---- Dimension 4: nesting — intersection wins regardless of paren wrapping ----
		// Locks in upstream TSTypeLiteral() arm J: the IMMEDIATE parent is what
		// matters. Parens preserve intent: `({} & X)` still has `{}`'s parent as
		// IntersectionType inside the parens, so we still skip.
		{Code: `type T = ({} & X);`},
		{Code: `type T = (X & {});`},

		// ---- AllowInterfaces 'with-single-extends': multi-extends always allowed
		// regardless of option (arm D in upstream, independent of option) ----
		{
			Code: `
interface A { x: number }
interface B { y: number }
interface Both extends A, B {}
`,
			Options: map[string]interface{}{"allowInterfaces": "with-single-extends"},
		},

		// ---- AllowWithName: pattern with `$` anchor only matches at end ----
		{
			Code:    `interface UserProps {}`,
			Options: map[string]interface{}{"allowWithName": "Props$"},
		},

		// ---- Real-user: `Record<string, {}>` (canonical "any object" usage,
		// but `{}` should still be reported — author needs to migrate intent) ----
		// (Tested in invalid below.) — N/A here.

		// ---- Option JSON-array shape coverage ----
		// SKILL.md requires every option to be exercised in both the direct
		// `map[string]interface{}{...}` shape (covered by all options tests
		// above) AND the array-wrapped `[]interface{}{...}` shape (matches
		// the multi-element rule_tester / CLI invocation). These mirror
		// upstream-valid cases via the array shape so that the JSON path
		// through `utils.GetOptionsMap` is exercised end-to-end.
		{
			Code:    `interface Base {}`,
			Options: []interface{}{map[string]interface{}{"allowInterfaces": "always"}},
		},
		{
			Code:    `type Base = {};`,
			Options: []interface{}{map[string]interface{}{"allowObjectTypes": "always"}},
		},
		{
			Code:    `interface BaseProps {}`,
			Options: []interface{}{map[string]interface{}{"allowWithName": "Props$"}},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- PR review feedback lock-ins ----
		// These lock in three behaviors that gemini-code-assist[bot] raised
		// concerns about during PR review; the rule must keep matching
		// upstream typescript-eslint v8.45.0 in all three.

		// (1) Trailing comma in type parameters: `interface Foo<T,> {}`. tsgo's
		// `parseDelimitedList` consumes the comma and positions End() at `>`,
		// so the scanner-based `<>` capture in `typeParametersText` round-trips
		// the trailing comma into the suggestion verbatim. Lock that in.
		{
			Code: `interface Foo<T,> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterface", Output: `type Foo<T,> = object`},
						{MessageId: "replaceEmptyInterface", Output: `type Foo<T,> = unknown`},
					},
				},
			},
		},

		// (2) Comment between modifier and `interface` keyword must be
		// preserved in the suggestion. The fix replace-range starts at the
		// `interface` keyword (via scanner), so any trivia between modifiers
		// and `interface` falls outside the range and survives verbatim.
		{
			Code: `
namespace N {
  export /* hello */ interface Foo {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      3,
					Column:    32,
					EndLine:   3,
					EndColumn: 35,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterface",
							Output: `
namespace N {
  export /* hello */ type Foo = object
}
`,
						},
						{
							MessageId: "replaceEmptyInterface",
							Output: `
namespace N {
  export /* hello */ type Foo = unknown
}
`,
						},
					},
				},
			},
		},

		// (3) `string & ({})` — the parenthesized `{}`'s immediate parent is
		// `ParenthesizedType`, not `IntersectionType`. Upstream's
		// `node.parent.type === 'TSIntersectionType'` is an immediate-parent
		// check, so this case is REPORTED upstream — and so must we be. Lock
		// in the no-paren-unwrapping behavior.
		{
			Code: `type T = string & ({});`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T = string & (object);`},
						{MessageId: "replaceEmptyObjectType", Output: `type T = string & (unknown);`},
					},
				},
			},
		},

		// ---- Additional Dimension 4 lock-ins identified during self-review ----

		// `keyof {}` — TypeOperator parent, reports.
		{
			Code: `type T = keyof {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T = keyof object;`},
						{MessageId: "replaceEmptyObjectType", Output: `type T = keyof unknown;`},
					},
				},
			},
		},

		// `infer U extends {}` — `{}` is the constraint of an `infer` declaration
		// inside a conditional type. Parent is InferType.
		{
			Code: `type T<X> = X extends infer U extends {} ? U : never;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    39,
					EndLine:   1,
					EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T<X> = X extends infer U extends object ? U : never;`},
						{MessageId: "replaceEmptyObjectType", Output: `type T<X> = X extends infer U extends unknown ? U : never;`},
					},
				},
			},
		},

		// Mapped-type value `{}` — `{[P in K]: {}}`. The OUTER is a MappedType
		// (KindMappedType, not KindTypeLiteral), so it isn't visited; the inner
		// `{}` is a regular TypeLiteral and gets reported.
		{
			Code: `type T<K extends string> = { [P in K]: {} };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    40,
					EndLine:   1,
					EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T<K extends string> = { [P in K]: object };`},
						{MessageId: "replaceEmptyObjectType", Output: `type T<K extends string> = { [P in K]: unknown };`},
					},
				},
			},
		},

		// Namespace-qualified extends — `interface I extends X.Y {}`. The
		// extended-type capture must preserve the dotted form.
		{
			Code: `interface Foo extends Ns.Inner {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterfaceWithSuper", Output: `type Foo = Ns.Inner`},
					},
				},
			},
		},

		// Generic-namespace-qualified extends — `interface I extends X.Y<T> {}`.
		{
			Code: `interface Foo<T> extends Ns.Inner<T> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterfaceWithSuper", Output: `type Foo<T> = Ns.Inner<T>`},
					},
				},
			},
		},

		// ---- Dimension 1: type-literal in parameter / return / property positions ----
		{
			Code: `function f(x: {}) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `function f(x: object) {}`},
						{MessageId: "replaceEmptyObjectType", Output: `function f(x: unknown) {}`},
					},
				},
			},
		},
		{
			Code: `function f(): {} { return 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `function f(): object { return 1; }`},
						{MessageId: "replaceEmptyObjectType", Output: `function f(): unknown { return 1; }`},
					},
				},
			},
		},
		{
			Code: `const f = (): {} => 1;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `const f = (): object => 1;`},
						{MessageId: "replaceEmptyObjectType", Output: `const f = (): unknown => 1;`},
					},
				},
			},
		},
		{
			Code: `class C { x: {} = 0; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `class C { x: object = 0; }`},
						{MessageId: "replaceEmptyObjectType", Output: `class C { x: unknown = 0; }`},
					},
				},
			},
		},

		// ---- Dimension 1: nested generic / tuple ----
		{
			Code: `type T = Array<{}>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T = Array<object>;`},
						{MessageId: "replaceEmptyObjectType", Output: `type T = Array<unknown>;`},
					},
				},
			},
		},
		{
			Code: `type T = [{}];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T = [object];`},
						{MessageId: "replaceEmptyObjectType", Output: `type T = [unknown];`},
					},
				},
			},
		},

		// ---- Dimension 2: type-literal inside another (non-empty) type literal ----
		// The outer `{ x: {} }` has a member, so only the inner `{}` is reported.
		{
			Code: `type T = { x: {} };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T = { x: object };`},
						{MessageId: "replaceEmptyObjectType", Output: `type T = { x: unknown };`},
					},
				},
			},
		},

		// ---- Dimension 4: nesting — TypeLiteral inside union (NOT skipped) ----
		// Locks in upstream TSTypeLiteral() arm J inverse: parent is union, not
		// intersection — must still report.
		{
			Code: `type T = {} | string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T = object | string;`},
						{MessageId: "replaceEmptyObjectType", Output: `type T = unknown | string;`},
					},
				},
			},
		},
		// Both arms of a union containing two empty literals fire independently.
		{
			Code: `type T = {} | {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T = object | {};`},
						{MessageId: "replaceEmptyObjectType", Output: `type T = unknown | {};`},
					},
				},
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T = {} | object;`},
						{MessageId: "replaceEmptyObjectType", Output: `type T = {} | unknown;`},
					},
				},
			},
		},

		// ---- Dimension 4: declaration forms — generic constraint and default ----
		// Locks in: TSTypeLiteral inside a type-parameter constraint is still
		// reported (parent is TypeParameter, not Intersection / TypeAlias).
		{
			Code: `type T<X extends {}> = X;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T<X extends object> = X;`},
						{MessageId: "replaceEmptyObjectType", Output: `type T<X extends unknown> = X;`},
					},
				},
			},
		},
		// Default type parameter value
		{
			Code: `type T<X = {}> = X;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T<X = object> = X;`},
						{MessageId: "replaceEmptyObjectType", Output: `type T<X = unknown> = X;`},
					},
				},
			},
		},

		// ---- Locks in upstream TSInterfaceDeclaration() arm C inverse:
		// 1 extend with default 'never' still reports ----
		{
			Code: `interface Empty extends Base {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterfaceWithSuper", Output: `type Empty = Base`},
					},
				},
			},
		},

		// ---- Locks in upstream TSTypeLiteral() arm L inverse: parent is
		// TypeAliasDeclaration but allowWithName does NOT match ----
		{
			Code:    `type Foo = {};`,
			Options: map[string]interface{}{"allowWithName": "Bar"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type Foo = object;`},
						{MessageId: "replaceEmptyObjectType", Output: `type Foo = unknown;`},
					},
				},
			},
		},

		// ---- Locks in upstream TSTypeLiteral() arm M inverse: allowWithName
		// is set but parent is NOT TypeAliasDeclaration ----
		// `let x: {}` parent is a VariableDeclaration, not a TypeAliasDeclaration,
		// so allowWithName has no effect.
		{
			Code:    `let value: {};`,
			Options: map[string]interface{}{"allowWithName": "value"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `let value: object;`},
						{MessageId: "replaceEmptyObjectType", Output: `let value: unknown;`},
					},
				},
			},
		},

		// ---- Locks in: declaration-merged class drops suggestions on the
		// 0-extend interface too (E branch + F branch) ----
		{
			Code: `
interface Foo {}
class Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      2,
					Column:    11,
					EndLine:   2,
					EndColumn: 14,
				},
			},
		},

		// ---- Real-user: React.FC<{}> empty props ----
		{
			Code: `
type FC<P> = (props: P) => null;
const C: FC<{}> = () => null;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      3,
					Column:    13,
					EndLine:   3,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyObjectType",
							Output: `
type FC<P> = (props: P) => null;
const C: FC<object> = () => null;
`,
						},
						{
							MessageId: "replaceEmptyObjectType",
							Output: `
type FC<P> = (props: P) => null;
const C: FC<unknown> = () => null;
`,
						},
					},
				},
			},
		},

		// ---- Real-user: generic factory constraint commonly written as `<T extends {}>` ----
		{
			Code: `function clone<T extends {}>(input: T): T { return input; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `function clone<T extends object>(input: T): T { return input; }`},
						{MessageId: "replaceEmptyObjectType", Output: `function clone<T extends unknown>(input: T): T { return input; }`},
					},
				},
			},
		},

		// ---- Lock-in: interface with type parameter that has empty body ----
		// Reports on the interface itself; the inner `{}` in the parameter
		// constraint is also empty so we get two diagnostics.
		{
			Code: `interface Foo<T extends {}> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterface", Output: `type Foo<T extends {}> = object`},
						{MessageId: "replaceEmptyInterface", Output: `type Foo<T extends {}> = unknown`},
					},
				},
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    25,
					EndLine:   1,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `interface Foo<T extends object> {}`},
						{MessageId: "replaceEmptyObjectType", Output: `interface Foo<T extends unknown> {}`},
					},
				},
			},
		},

		// ---- Lock-in: empty interface with `export` modifier (no suggestion drops `export`) ----
		// Mirrors no_empty_interface's export-preservation handling — modifiers
		// are reconstructed in the replacement text.
		{
			Code: `
namespace N {
  export interface Foo {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      3,
					Column:    20,
					EndLine:   3,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterface",
							Output: `
namespace N {
  export type Foo = object
}
`,
						},
						{
							MessageId: "replaceEmptyInterface",
							Output: `
namespace N {
  export type Foo = unknown
}
`,
						},
					},
				},
			},
		},

		// ---- Dimension 1: function-type return — `let f: () => {}` ----
		// Locks in: TypeLiteral as a function-type return position. Parent is
		// FunctionType, not Intersection / TypeAlias.
		{
			Code: `let f: () => {} = null!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `let f: () => object = null!;`},
						{MessageId: "replaceEmptyObjectType", Output: `let f: () => unknown = null!;`},
					},
				},
			},
		},

		// ---- Dimension 1: class method param/return ----
		{
			Code: `class C { m(x: {}): {} { return x; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `class C { m(x: object): {} { return x; } }`},
						{MessageId: "replaceEmptyObjectType", Output: `class C { m(x: unknown): {} { return x; } }`},
					},
				},
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `class C { m(x: {}): object { return x; } }`},
						{MessageId: "replaceEmptyObjectType", Output: `class C { m(x: {}): unknown { return x; } }`},
					},
				},
			},
		},

		// ---- Dimension 4: parenthesized type literal ({}) — parens are
		// transparent; the rule reports the inner `{}` because parent is
		// ParenthesizedType, not Intersection ----
		{
			Code: `let x: ({}) = null!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `let x: (object) = null!;`},
						{MessageId: "replaceEmptyObjectType", Output: `let x: (unknown) = null!;`},
					},
				},
			},
		},

		// ---- Dimension 4: parenthesized type alias — allowWithName only checks
		// the IMMEDIATE parent (TSTypeAliasDeclaration), so a ParenthesizedType
		// wrapper makes the option ineffective ----
		{
			Code:    `type Foo = ({});`,
			Options: map[string]interface{}{"allowWithName": "Foo"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type Foo = (object);`},
						{MessageId: "replaceEmptyObjectType", Output: `type Foo = (unknown);`},
					},
				},
			},
		},

		// ---- Branch lock-in: generic with BOTH constraint and default — both
		// inner `{}` must fire independently and the interface itself fires too ----
		{
			Code: `interface I<T extends {} = {}> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterface", Output: `type I<T extends {} = {}> = object`},
						{MessageId: "replaceEmptyInterface", Output: `type I<T extends {} = {}> = unknown`},
					},
				},
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `interface I<T extends object = {}> {}`},
						{MessageId: "replaceEmptyObjectType", Output: `interface I<T extends unknown = {}> {}`},
					},
				},
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    28,
					EndLine:   1,
					EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `interface I<T extends {} = object> {}`},
						{MessageId: "replaceEmptyObjectType", Output: `interface I<T extends {} = unknown> {}`},
					},
				},
			},
		},

		// ---- Real-user: `Record<string, {}>` (canonical "any object" usage) ----
		{
			Code: `type Dict = Record<string, {}>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    28,
					EndLine:   1,
					EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type Dict = Record<string, object>;`},
						{MessageId: "replaceEmptyObjectType", Output: `type Dict = Record<string, unknown>;`},
					},
				},
			},
		},

		// ---- Lock-in: multi-modifier preservation ----
		// `declare interface` inside a declared namespace.
		{
			Code: `
declare namespace N {
  export interface Foo extends Base {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      3,
					Column:    20,
					EndLine:   3,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output: `
declare namespace N {
  export type Foo = Base
}
`,
						},
					},
				},
			},
		},

		// ---- Branch lock-in: 0-extend interface + `'with-single-extends'`
		// option still reports (the option exempts ONLY extend.length===1, not
		// extend.length===0). Locks in upstream TSInterfaceDeclaration() arm C
		// inverse against the 0-extend path: option does not bleed across
		// arm F. ----
		{
			Code:    `interface Foo {}`,
			Options: map[string]interface{}{"allowInterfaces": "with-single-extends"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterface", Output: `type Foo = object`},
						{MessageId: "replaceEmptyInterface", Output: `type Foo = unknown`},
					},
				},
			},
		},

		// ---- Branch lock-in: > 2 extends behaves identically to 2 extends
		// (arm D triggers on `extend.length > 1`, not `=== 2`) ----
		{
			Code: `interface Foo extends A, B, C { x: number }`,
			Errors: []rule_tester.InvalidTestCaseError{},
		},

		// ---- Dimension 4: TypeLiteral inside conditional-type branch ----
		// Parent is ConditionalType, not Intersection — should report.
		{
			Code: `type T<X> = X extends string ? {} : null;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    32,
					EndLine:   1,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type T<X> = X extends string ? object : null;`},
						{MessageId: "replaceEmptyObjectType", Output: `type T<X> = X extends string ? unknown : null;`},
					},
				},
			},
		},

		// ---- Dimension 4: TypeLiteral in `as` type-expression wrapper ----
		// Parent is AsExpression (tsgo Kind), reports.
		{
			Code: `const x = null as {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    19,
					EndLine:   1,
					EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `const x = null as object;`},
						{MessageId: "replaceEmptyObjectType", Output: `const x = null as unknown;`},
					},
				},
			},
		},

		// ---- Dimension 4: TypeLiteral in `satisfies` type-expression wrapper ----
		{
			Code: `const x = null satisfies {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `const x = null satisfies object;`},
						{MessageId: "replaceEmptyObjectType", Output: `const x = null satisfies unknown;`},
					},
				},
			},
		},

		// ---- Option JSON-array shape coverage (invalid side) ----
		// Mirrors an upstream-invalid case but passes options through the
		// `[]interface{}{...}` shape to exercise `utils.GetOptionsMap`'s
		// multi-element path end-to-end.
		{
			Code:    `type Base = {};`,
			Options: []interface{}{map[string]interface{}{"allowObjectTypes": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type Base = object;`},
						{MessageId: "replaceEmptyObjectType", Output: `type Base = unknown;`},
					},
				},
			},
		},
	})
}
