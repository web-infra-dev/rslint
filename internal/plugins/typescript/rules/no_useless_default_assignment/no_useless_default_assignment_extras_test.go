// TestNoUselessDefaultAssignmentExtras locks in branches and edge shapes
// that the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without
// breaking a named lock-in.
package no_useless_default_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessDefaultAssignmentExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessDefaultAssignmentRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: key forms — numeric-literal key (optional) ----
			// PropertyName as NumericLiteral exercises getStaticPropertyName's
			// KindNumericLiteral arm. Source has `0` defined as optional, so
			// the default IS meaningful.
			{Code: `
declare const obj: { 0?: string };
const { 0: x = 'default' } = obj;
      `},
			// ---- Dimension 4: key forms — symbol-via-computed key ----
			// Symbol-keyed properties don't survive getStaticPropertyName
			// (Symbol() isn't a Literal), so the rule cannot resolve the
			// property and skips silently. Lock in: no false positive.
			{Code: `
declare const sym: unique symbol;
declare const obj: { [sym]: string };
const { [sym]: x = 'default' } = obj;
      `},
			// ---- Dimension 4: declaration forms — setter parameter ----
			// SetAccessor is in isAnonymousFunctionLike but has no contextual
			// type unless implementing an interface; here it has none → skip.
			{Code: `
class C {
  set value(v: string = '') {}
}
      `},
			// ---- Dimension 4: declaration forms — generator parameter ----
			// FunctionExpression with `function*`. No contextual type → skip.
			{Code: `
const gen = function* (a: number = 5) {};
      `},
			// ---- Dimension 4: nested patterns — array inside array (optional) ----
			// Nested ArrayBindingPattern; outer source is a tuple-of-tuple
			// containing `number | undefined`, so the default IS meaningful.
			{Code: `
declare const x: [[number | undefined]];
const [[a = 1]] = x;
      `},
			// ---- Dimension 4: graceful degradation — rest element in pattern ----
			// Rest binding in destructuring; the rest element itself has no
			// initializer so the rule sees nothing. Lock in: doesn't crash.
			{Code: `
declare const arr: number[];
const [first, ...rest] = arr;
      `},
			// ---- Dimension 4: graceful degradation — empty destructuring ----
			{Code: `
const {} = { a: 1 };
const [] = [1];
      `},
			// ---- Dimension 4: graceful degradation — abstract / declare method ----
			// Bodyless members never carry a parameter initializer (TS rejects
			// it syntactically) — exhaustiveness audit, must not crash.
			{Code: `
abstract class C {
  abstract foo(a: string): void;
}
declare class D {
  bar(a: string): void;
}
      `},
			// ---- Real-user: bare arrow with no contextual type ----
			// If contextual type is null, upstream skips early. A standalone
			// arrow assigned to a `const` (no type annotation) has none.
			{Code: `
const f = (a: number = 5) => a;
      `},
			// ---- Real-user: array destructuring from non-tuple Array ----
			// Lock in upstream branch: when sourceType is Array<T>, the
			// destructuring's outer arm requires isTupleType to report; Array
			// is not a tuple, so no diagnostic on outer-level elements.
			{Code: `
declare const arr: Array<string | undefined>;
const [a = 'default'] = arr;
      `},
			// ---- Real-user: optional property in nested destructuring ----
			// `foo?: { bar?: string }` — both layers optional, default IS
			// meaningful at both levels.
			{Code: `
function f({ foo: { bar = '' } = {} }: { foo?: { bar?: string } }) {
  return bar;
}
      `},
			// ---- Locks in: `= void 0` is NOT recognized as `undefined` ----
			// Upstream uses `node.right.type === Identifier && node.right.name
			// === 'undefined'`. `void 0` is a VoidExpression / UnaryExpression
			// — not an Identifier — so the lexical check rejects it. Falls
			// through to the non-undefined arm; source `a?: string` is
			// optional so default IS meaningful → no report.
			{Code: `
function f({ a = void 0 }: { a?: string }) {
  return a;
}
      `},
			// ---- Locks in: `= obj.undefined` is NOT the literal `undefined` ----
			// PropertyAccessExpression, not Identifier. Optional source means
			// default IS meaningful.
			{Code: `
declare const obj: { undefined: string };
function f({ a = obj.undefined }: { a?: string }) {
  return a;
}
      `},
			// ---- Locks in: `= null` is NOT `= undefined` ----
			// `null` is a NullKeyword, not the `undefined` identifier. Source
			// `a?: string | null` admits both null and undefined; default IS
			// meaningful.
			{Code: `
function f({ a = null }: { a?: string | null }) {
  return a;
}
      `},
			// ---- Locks in: `Partial<T>` makes every property optional ----
			// Property `a` becomes `a?: string`; canBeUndefined returns true;
			// no report.
			{Code: `
function f({ a = 'd' }: Partial<{ a: string }>) {
  return a;
}
      `},
			// ---- Locks in: mapped type making properties optional ----
			// User-defined `MakeOptional` produces the same effect as Partial.
			{Code: `
type MakeOptional<T> = { [K in keyof T]?: T[K] };
function f({ a = 'd' }: MakeOptional<{ a: string }>) {
  return a;
}
      `},
			// ---- Locks in: `satisfies` clause on initializer (TS 4.9+) ----
			// SatisfiesExpression wraps the value; not the undefined identifier.
			// Falls through to non-undefined arm; optional source accepts the
			// default.
			{Code: `
function f({ a = ('d' satisfies string) }: { a?: string }) {
  return a;
}
      `},
			// ---- Locks in: catch-clause binding pattern (no source type) ----
			// Catch binding has no VariableDeclaration init in tsgo;
			// getSourceTypeForPattern returns nil → silent skip.
			{Code: `
try {
  throw new Error();
} catch ({ message = 'unknown' }) {
  console.log(message);
}
      `},
			// ---- Real-user: React-style functional component with optional prop ----
			{Code: `
interface CardProps {
  name: string;
  age?: number;
}
const Card = ({ name, age = 0 }: CardProps) => name + age;
      `},
			// ---- Real-user: async function with optional destructured option ----
			{Code: `
async function fetchUrl({ url, retries = 3 }: { url: string; retries?: number }) {
  return url + retries;
}
      `},
			// ---- Real-user: untyped Redux-style reducer (no contextual) ----
			// `const reducer = (...) => ...` has no contextual type on the
			// arrow; upstream skips. State is non-undefined, but without a
			// contextual signature we can't know about the parameter contract.
			{Code: `
type State = { count: number };
type Action = { type: string };
declare const initial: State;
const reducer = (state: State = initial, action: Action): State => state;
      `},
			// ---- Locks in: rest BindingElement does NOT trip the listener ----
			// `...rest` BindingElement has no Initializer (TS rejects rest
			// with default) — listener's `if Initializer == nil return` guards.
			{Code: `
declare const obj: { a: string; b: number; c: boolean };
const { a, ...rest } = obj;
      `},
			// ---- Locks in: array rest element with no default ----
			{Code: `
declare const tuple: [string, number, boolean];
const [first, ...rest] = tuple;
      `},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in IsUndefinedIdentifier paren-stripping for `(undefined)` ----
			// tsgo preserves ParenthesizedExpression while ESTree elides parens
			// at parse time. IsUndefinedIdentifier explicitly skips parens so
			// the lexical check matches upstream's `node.right.type === Identifier`.
			{
				Code: `
function foo({ a = (undefined) }: { a?: string }) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    20,
						EndColumn: 31,
					},
				},
				Output: []string{`
function foo({ a }: { a?: string }) {}
      `},
			},
			// ---- Locks in IsUndefinedIdentifier paren-stripping for `((undefined))` ----
			{
				Code: `
function foo({ a = ((undefined)) }: { a?: string }) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    20,
						EndColumn: 33,
					},
				},
				Output: []string{`
function foo({ a }: { a?: string }) {}
      `},
			},
			// ---- Locks in: `undefined as any` is NOT recognized as undefined ----
			// tsgo preserves the AsExpression wrapper; IsUndefinedIdentifier
			// strips parens only, so the lexical check fails (matching upstream's
			// `node.right.type === Identifier`). The rule then falls through to
			// the non-undefined arm, which detects the non-optional `a: string`
			// property and reports `uselessDefaultAssignment` instead.
			{
				Code: `
declare const obj: { a: string };
const { a = undefined as any } = obj;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    13,
						EndColumn: 29,
					},
				},
				Output: []string{`
declare const obj: { a: string };
const { a } = obj;
      `},
			},
			// ---- Locks in checkAssignmentPattern: BindingElement with renamed key
			// (`{ value: bar = '' }`) and non-optional property ----
			// Exercises getPropertyName picking up be.PropertyName (== 'value')
			// rather than be.Name (== 'bar'), and getTypeOfBindingElementProperty
			// resolving the source via VariableDeclaration init.
			{
				Code: `
declare const obj: { value: string };
const { value: bar = '' } = obj;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    22,
						EndColumn: 24,
					},
				},
				Output: []string{`
declare const obj: { value: string };
const { value: bar } = obj;
      `},
			},
			// ---- Locks in upstream getPropertyName(): NumericLiteral key ----
			// Source has `0: string` (non-optional). Property name "0" resolves
			// through getStaticPropertyName's NumericLiteral arm.
			{
				Code: `
declare const obj: { 0: string };
const { 0: x = 'default' } = obj;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    16,
						EndColumn: 25,
					},
				},
				Output: []string{`
declare const obj: { 0: string };
const { 0: x } = obj;
      `},
			},
			// ---- Locks in upstream hasPropertyInAllBranches: method-shorthand
			// property counts as "having the property" ----
			// `{ a() {} }` is MethodDeclaration in tsgo; ESLint treats it as
			// Property because `prop.type === Property`. Upstream's branch
			// matches it via getPropertyName(prop.key), so we should report.
			{
				Code: `
const { a = 'baz' } = cond ? { a() {} } : { a: 'bar' };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    13,
						EndColumn: 18,
					},
				},
				Output: []string{`
const { a } = cond ? { a() {} } : { a: 'bar' };
      `},
			},
			// ---- Locks in upstream hasPropertyInAllBranches: getter shorthand
			// counts as "having the property" ----
			{
				Code: `
const { a = 'baz' } = cond ? { get a() { return 'foo'; } } : { a: 'bar' };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    13,
						EndColumn: 18,
					},
				},
				Output: []string{`
const { a } = cond ? { get a() { return 'foo'; } } : { a: 'bar' };
      `},
			},
			// ---- Locks in upstream hasPropertyInAllBranches: shorthand
			// property assignment (`{ a }`) counts as the property ----
			{
				Code: `
declare const a: string;
const { a: x = 'baz' } = cond ? { a } : { a: 'bar' };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    16,
						EndColumn: 21,
					},
				},
				Output: []string{`
declare const a: string;
const { a: x } = cond ? { a } : { a: 'bar' };
      `},
			},
			// ---- Locks in canBeUndefined returning true for `any` parameter
			// with `= undefined` → preferOptionalSyntax ----
			{
				Code: `
function f(a: any = undefined) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferOptionalSyntax",
						Line:      2,
						Column:    21,
						EndColumn: 30,
					},
				},
				Output: []string{`
function f(a?: any) {}
      `},
			},
			// ---- Locks in canBeUndefined returning true for `unknown` parameter
			// with `= undefined` → preferOptionalSyntax ----
			{
				Code: `
function f(a: unknown = undefined) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferOptionalSyntax",
						Line:      2,
						Column:    25,
						EndColumn: 34,
					},
				},
				Output: []string{`
function f(a?: unknown) {}
      `},
			},
			// ---- Locks in checkAssignmentPattern: parameter with non-undefined
			// type annotation gets `= undefined` → uselessUndefined (NOT
			// preferOptionalSyntax) ----
			// `x: number` cannot be undefined, so the preferOptional arm
			// returns early. The `= undefined` is still useless.
			{
				Code: `
function f(x: number = undefined) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    24,
						EndColumn: 33,
					},
				},
				Output: []string{`
function f(x: number) {}
      `},
			},
			// ---- Locks in `this: T` parameter shifting paramIndex ----
			// `this:` lives in funcNode.Parameters() at index 0, but the
			// signature's parameters slice excludes it. getParameterType
			// detects the thisParameter via ast.GetThisParameter and decrements
			// the index, so BindingElement `bar` resolves to its real source.
			{
				Code: `
function f(this: void, { bar = 42 }: { bar: number }) {
  return bar;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    32,
						EndColumn: 34,
					},
				},
				Output: []string{`
function f(this: void, { bar }: { bar: number }) {
  return bar;
}
      `},
			},
			// ---- Locks in: VariableDeclaration init resolves source for
			// shorthand BindingElement → non-optional property reports ----
			{
				Code: `
declare const obj: { foo: string };
const { foo = undefined } = obj;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      3,
						Column:    15,
						EndColumn: 24,
					},
				},
				Output: []string{`
declare const obj: { foo: string };
const { foo } = obj;
      `},
			},
			// ---- Locks in: arrow inside .map gets contextual signature whose
			// parameter is non-optional `number` → useless default ----
			{
				Code: `
[1, 2, 3].map((x = 5) => x + 1);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    20,
						EndColumn: 21,
					},
				},
				Output: []string{`
[1, 2, 3].map((x) => x + 1);
      `},
			},
			// ---- Locks in: tuple element type resolution at non-zero index ----
			{
				Code: `
declare const tuple: [string, number];
const [a, b = 0] = tuple;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    15,
						EndColumn: 16,
					},
				},
				Output: []string{`
declare const tuple: [string, number];
const [a, b] = tuple;
      `},
			},
			// ---- Locks in preferOptionalSyntax when binding `left` is NOT a
			// simple Identifier (only the removeDefault fix runs; no `?`
			// insertion) ----
			// `{ a }` destructuring pattern cannot carry a `?`, so upstream
			// skips the insertion step.
			{
				Code: `
function f({ a }: { a: number } | undefined = undefined) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferOptionalSyntax",
						Line:      2,
						Column:    47,
						EndColumn: 56,
					},
				},
				Output: []string{`
function f({ a }: { a: number } | undefined) {}
      `},
			},
			// ---- Locks in three-level nested object destructuring ----
			// Tests recursive getSourceTypeForPattern → BindingElement →
			// ObjectBindingPattern → BindingElement → ObjectBindingPattern →
			// Parameter (chain length 5). Source property `c: string` is
			// non-optional, so the deepest default IS useless.
			{
				Code: `
function f({ a: { b: { c = '' } } }: { a: { b: { c: string } } }) {
  return c;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    28,
						EndColumn: 30,
					},
				},
				Output: []string{`
function f({ a: { b: { c } } }: { a: { b: { c: string } } }) {
  return c;
}
      `},
			},
			// ---- Locks in array-inside-object nested destructuring ----
			// Outer is ObjectBindingPattern, inner is ArrayBindingPattern;
			// getSourceTypeForPattern resolves through ObjectBindingPattern's
			// property `a: [string]` (tuple), then the inner array element type
			// is `string` (non-optional) → useless default.
			{
				Code: `
declare const obj: { a: [string] };
const { a: [x = ''] } = obj;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    17,
						EndColumn: 19,
					},
				},
				Output: []string{`
declare const obj: { a: [string] };
const { a: [x] } = obj;
      `},
			},
			// ---- Locks in object-inside-array nested destructuring ----
			// Outer is ArrayBindingPattern (tuple [{ x: string }]), inner is
			// ObjectBindingPattern. Property `x` is non-optional → useless.
			{
				Code: `
declare const tuple: [{ x: string }];
const [{ x = 'd' }] = tuple;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    14,
						EndColumn: 17,
					},
				},
				Output: []string{`
declare const tuple: [{ x: string }];
const [{ x }] = tuple;
      `},
			},
			// ---- Locks in static class member arrow with contextual type ----
			// Class field arrow that's assigned to a property of a typed
			// interface. The arrow gets the contextual signature from the
			// interface property.
			{
				Code: `
interface I {
  cb: (x: string) => void;
}
const obj: I = {
  cb: (x = 'd') => {},
};
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      6,
						Column:    12,
						EndColumn: 15,
					},
				},
				Output: []string{`
interface I {
  cb: (x: string) => void;
}
const obj: I = {
  cb: (x) => {},
};
      `},
			},
			// ---- Locks in getter/setter destructuring parameter ----
			// SetAccessor parameter with destructuring; the source type comes
			// from the parameter's type annotation. Property `x: string` is
			// non-optional → useless default.
			{
				Code: `
class C {
  set value({ x = '' }: { x: string }) {}
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    19,
						EndColumn: 21,
					},
				},
				Output: []string{`
class C {
  set value({ x }: { x: string }) {}
}
      `},
			},
			// ---- Locks in constructor parameter destructuring with type ----
			{
				Code: `
class C {
  constructor({ x = '' }: { x: string }) {}
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    21,
						EndColumn: 23,
					},
				},
				Output: []string{`
class C {
  constructor({ x }: { x: string }) {}
}
      `},
			},
			// ---- Locks in: `= undefined` on parameter with NO type annotation
			// → uselessUndefined (parameter) ----
			// Param.Type is nil, so the preferOptional arm is skipped. typeLabel
			// returns 'parameter'.
			{
				Code: `
function f(x = undefined) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      2,
						Column:    16,
						EndColumn: 25,
					},
				},
				Output: []string{`
function f(x) {}
      `},
			},
			// ---- Locks in: `= undefined` on tuple element ----
			// Array destructuring → typeLabel returns 'property' (matching
			// upstream's `parent === ArrayPattern → 'property'`).
			{
				Code: `
declare const tuple: [string];
const [x = undefined] = tuple;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessUndefined",
						Line:      3,
						Column:    12,
						EndColumn: 21,
					},
				},
				Output: []string{`
declare const tuple: [string];
const [x] = tuple;
      `},
			},
			// ---- Locks in: parameter with optional `?` modifier AND default
			// `= undefined` → preferOptionalSyntax with QuestionToken pushing
			// the removeDefault left-end past the `?` ----
			// TypeScript actually disallows `x?: T = U` syntactically (`?` and
			// initializer are mutually exclusive on parameters), so this isn't
			// a real-world case. Test is kept commented for completeness — if
			// TS ever loosens this, the QuestionToken branch in removeDefaultFix
			// is the right path.
			// ---- Locks in: `Required<Partial<{a: string}>>` strips optional ----
			// Partial makes `a` optional, Required strips that back to
			// non-optional. Property type lookup returns `string`; default
			// useless.
			{
				Code: `
function f({ a = 'd' }: Required<Partial<{ a: string }>>) {
  return a;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    18,
						EndColumn: 21,
					},
				},
				Output: []string{`
function f({ a }: Required<Partial<{ a: string }>>) {
  return a;
}
      `},
			},
			// ---- Locks in: intersection type — `a` is non-optional via one
			// constituent ----
			{
				Code: `
function f({ a = 'd' }: { a: string } & { b: number }) {
  return a;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    18,
						EndColumn: 21,
					},
				},
				Output: []string{`
function f({ a }: { a: string } & { b: number }) {
  return a;
}
      `},
			},
			// ---- Locks in: numeric-base normalization on BindingElement key ----
			// tsgo normalizes `0x1` to '1' via utils.NormalizeNumericLiteral,
			// matching the source type's numeric key `1`. Without this
			// normalization the property lookup would miss → silent skip
			// would be wrong.
			{
				Code: `
declare const obj: { 1: string };
const { 0x1: x = 'd' } = obj;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    18,
						EndColumn: 21,
					},
				},
				Output: []string{`
declare const obj: { 1: string };
const { 0x1: x } = obj;
      `},
			},
			// ---- Real-user: Redux-style reducer WITH contextual type ----
			// Annotating the const with `Reducer<S, A>` gives the arrow a
			// contextual signature whose `state: S` is non-optional → default
			// useless.
			{
				Code: `
type State = { count: number };
type Action = { type: string };
type Reducer<S, A> = (state: S, action: A) => S;
declare const initial: State;
const reducer: Reducer<State, Action> = (state = initial, action) => state;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      6,
						Column:    50,
						EndColumn: 57,
					},
				},
				Output: []string{`
type State = { count: number };
type Action = { type: string };
type Reducer<S, A> = (state: S, action: A) => S;
declare const initial: State;
const reducer: Reducer<State, Action> = (state, action) => state;
      `},
			},
			// ---- Real-user: `.reduce` callback param with default ----
			// `Array<T>.reduce<U>` callback is `(acc: U, cur: T, idx, arr) => U`.
			// With initial `0` → U is number, acc is non-optional. Default useless.
			{
				Code: `
declare const arr: number[];
const sum = arr.reduce((acc = 0, cur) => acc + cur, 0);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    31,
						EndColumn: 32,
					},
				},
				Output: []string{`
declare const arr: number[];
const sum = arr.reduce((acc, cur) => acc + cur, 0);
      `},
			},
			// ---- Real-user: `.filter` callback element parameter ----
			// `Array<T>.filter` callback element is `T` (non-optional).
			{
				Code: `
[1, 2, 3].filter((x = 0) => x > 0);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    23,
						EndColumn: 24,
					},
				},
				Output: []string{`
[1, 2, 3].filter((x) => x > 0);
      `},
			},
			// ---- Real-user: `.forEach` callback element parameter ----
			{
				Code: `
[1, 2, 3].forEach((x = 0) => console.log(x));
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    24,
						EndColumn: 25,
					},
				},
				Output: []string{`
[1, 2, 3].forEach((x) => console.log(x));
      `},
			},
			// ---- Real-user: curried arrow with outer-type contextual chain ----
			// `Curry = (a) => (b: number) => number`. Inner arrow's contextual
			// signature comes from the outer arrow's return type. Inner `b:
			// number` is non-optional → useless default.
			{
				Code: `
type Curry = (a: number) => (b: number) => number;
const c: Curry = (a) => (b = 5) => a + b;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    30,
						EndColumn: 31,
					},
				},
				Output: []string{`
type Curry = (a: number) => (b: number) => number;
const c: Curry = (a) => (b) => a + b;
      `},
			},
			// ---- Locks in: mixed-default multi-parameter destructuring ----
			// Only the first parameter (non-optional source) reports; the
			// second parameter is independent and has no default.
			{
				Code: `
function f({ a = '' }: { a: string }, { b }: { b: string }) {
  return a + b;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    18,
						EndColumn: 20,
					},
				},
				Output: []string{`
function f({ a }: { a: string }, { b }: { b: string }) {
  return a + b;
}
      `},
			},
			// ---- Locks in: rest binding with a default on a sibling ----
			// `...rest` BindingElement is skipped (no Initializer), but the
			// sibling `a = ''` still reports against the source's non-optional
			// `a: string`. Lock in: rest does NOT mask sibling reports.
			{
				Code: `
declare const obj: { a: string; b: number };
const { a = '', ...rest } = obj;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    13,
						EndColumn: 15,
					},
				},
				Output: []string{`
declare const obj: { a: string; b: number };
const { a, ...rest } = obj;
      `},
			},
			// ---- Real-user: string literal key with hyphens (kebab-case) ----
			// PropertyName as StringLiteral with hyphens — getStaticPropertyName
			// returns the raw text; matches source's string-keyed property.
			{
				Code: `
declare const obj: { 'foo-bar': string };
const { 'foo-bar': x = 'd' } = obj;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    24,
						EndColumn: 27,
					},
				},
				Output: []string{`
declare const obj: { 'foo-bar': string };
const { 'foo-bar': x } = obj;
      `},
			},
			// ---- Locks in: generic constraint with property ----
			// T constrained to `{ x: string }`; checker.getPropertyOfType(T,
			// 'x') resolves via constraint → string (non-optional). Lock in
			// that constraint-based property lookup works.
			{
				Code: `
function f<T extends { x: string }>({ x = '' }: T) {
  return x;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    43,
						EndColumn: 45,
					},
				},
				Output: []string{`
function f<T extends { x: string }>({ x }: T) {
  return x;
}
      `},
			},
			// ---- Locks in: type alias source ----
			// type alias dereference → object type → property lookup.
			{
				Code: `
type Opts = { name: string };
function f({ name = '' }: Opts) {
  return name;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      3,
						Column:    21,
						EndColumn: 23,
					},
				},
				Output: []string{`
type Opts = { name: string };
function f({ name }: Opts) {
  return name;
}
      `},
			},
			// ---- Locks in: class instance type source ----
			// Destructuring an instance of a class — checker resolves the
			// instance type's property `x: string` (non-optional).
			{
				Code: `
class C { x = ''; }
declare const c: C;
const { x = 'd' } = c;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      4,
						Column:    13,
						EndColumn: 16,
					},
				},
				Output: []string{`
class C { x = ''; }
declare const c: C;
const { x } = c;
      `},
			},
			// ---- Locks in: `= void 0` on a NON-optional source falls through
			// to the non-undefined arm and reports uselessDefaultAssignment ----
			// `void 0` is a UnaryExpression (not the `undefined` Identifier),
			// so the lexical undefined arm rejects it. The non-undefined arm
			// then runs against `a: string` (non-optional) → useless.
			// Verified against upstream `npx eslint`: emits "Default value is
			// useless because the property is not optional" at the same loc.
			{
				Code: `
function f({ a = void 0 }: { a: string }) {
  return a;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    18,
						EndColumn: 24,
					},
				},
				// No autofix output: removing the default would leave `{ a }`
				// without `= void 0`, which is correct, so a fix IS produced.
				Output: []string{`
function f({ a }: { a: string }) {
  return a;
}
      `},
			},
			// ---- Locks in: `= null` on `a: string | null` (NO undefined) ----
			// `string | null` does not include undefined → canBeUndefined
			// returns false → default useless. `null` is a NullKeyword, not
			// the undefined Identifier; lexical arm skips. Verified against
			// upstream.
			{
				Code: `
function f({ a = null }: { a: string | null }) {
  return a;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "uselessDefaultAssignment",
						Line:      2,
						Column:    18,
						EndColumn: 22,
					},
				},
				Output: []string{`
function f({ a }: { a: string | null }) {
  return a;
}
      `},
			},
		},
	)
}
