// TestNoUnnecessaryTypeParametersExtras locks in branches and edge shapes
// that the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without
// breaking a named lock-in.
package no_unnecessary_type_parameters

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeParametersExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeParametersRule, []rule_tester.ValidTestCase{
		// ---- Dimension 1: AST node types ----
		{Code: `
// Async function: T used in param and (via await) the resolved return type.
async function identity<T>(arg: T): Promise<T> {
  return arg;
}
`},
		{Code: `
// Generator function: T used in the yielded value type and the parameter.
function* gen<T>(seed: T): Generator<T> {
  yield seed;
}
`},
		{Code: `
// Async generator.
async function* gen<T>(seed: T): AsyncGenerator<T> {
  yield seed;
}
`},

		// ---- Dimension 4: TSConstructSignatureDeclaration is never checked ----
		// Locks in upstream's selector list omission of
		// TSConstructSignatureDeclaration: even a type parameter used only
		// once in an interface's construct signature must never be reported,
		// because the rule's selectors never visit this node kind at all.
		{Code: `
interface I {
  new <T>(value: T): void;
}
`},

		// ---- Real-user: typescript-eslint/typescript-eslint#9961 (open) ----
		// Upstream's own class handling has a known, intentionally-unfixed
		// false positive: a class whose single member uses T only once is
		// still reported, even though (unlike a bare function) a reader might
		// expect per-member reasoning. rslint matches this for parity — see
		// the invalid cases below for the interface contrast this issue
		// documents (interface member usage is correctly NOT reported).
		{Code: `
interface ValueInterface<T> {
  value: T;
}
`},

		// ---- Real-user: typescript-eslint/typescript-eslint#11528 (not planned) ----
		// The maintainers confirmed this specific shape is intentional: TData
		// used as a direct generic argument of the *return type* (IError<TData>)
		// counts as "used multiple times" via the AST-level direct-generic-
		// argument shortcut, even though it's spelled only once. See the
		// invalid case below for the sibling shape from the same thread that
		// the maintainers confirmed SHOULD still be reported (TData used only
		// in the return type itself, with no other signature position).
		{Code: `
interface IError<TData> {
  error: true;
  data: TData;
}

declare const checkIsError: <TData>(anything: unknown) => anything is IError<TData>;

const extractErrorData = <TData,>(error?: unknown): IError<TData> | null => {
  return checkIsError<TData>(error) ? error : null;
};
`},

		// ---- Real-user: rspack normalization.ts inferred object-spread return ----
		// Spreading an optional generic value preserves T in the inferred return
		// type, so the parameter occurrence and inferred return occurrence relate
		// two signature positions. Upstream typescript-eslint 8.65.0 does not
		// report this shape.
		{
			Code:     `const cloneObject = <T>(value?: T) => ({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
function cloneObject<T>(value?: T) {
  return { ...value };
}
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code:     `const cloneObject = <T>(value?: T) => ({ tag: 1, ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code:     `const cloneObject = <T>(value?: T) => ({ ...value, tag: 1 });`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		// The legacy inference also flows through containers, branches, wrappers,
		// and generic return positions. These all resolve to types containing T in
		// TypeScript 5.9 with omitted strictness options.
		{
			Code: `
declare const flag: boolean;
const nested = <T>(value?: T) => ({ nested: { ...value } });
const array = <T>(value?: T) => [{ ...value }];
const tuple = <T>(value?: T) => [{ ...value }] as const;
const conditional = <T>(value?: T) => flag ? { ...value } : { ...value };
const logical = <T>(value?: T) => flag && { ...value };
const satisfied = <T>(value?: T) => ({ ...value }) satisfies object;
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
declare function identityObject<X>(value: X): X;
const promised = <T>(value?: T) => Promise.resolve({ ...value });
const identity = <T>(value?: T) => identityObject({ ...value });
const getter = <T>(value?: T) => ({ get clone() { return { ...value }; } });
function* generated<T>(value?: T) { yield { ...value }; }
const pair = <T, U>(left?: T, right?: U) => ({
  left: { ...left },
  right: [{ ...right }],
});
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
const local = <T>(value?: T) => {
  const clone = { ...value };
  return clone;
};
const shorthand = <T>(value?: T) => {
  const clone = { ...value };
  return { clone };
};
const spreadLocal = <T>(value?: T) => {
  const clone = { ...value };
  return { ...clone };
};
const asynchronous = async <T>(value?: T) => {
  const clone = { ...value };
  return clone;
};
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
declare function chooseRight<A, B>(left: A, right: B, echo: A): B;
class Box<X> { constructor(readonly value: X) {} }
const called = <T>(value?: T) => chooseRight({}, { ...value }, {});
const constructed = <T>(value?: T) => new Box({ ...value });
const returnedClass = <T>(value?: T) => class {
  clone() { return { ...value }; }
};
const returnedMethod = <T>(value?: T) => ({
  clone() { return { ...value }; },
});
function* generated<T>(value?: T) { yield* [{ ...value }]; }
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
declare const key: 'nested' | 'fixed';
const dynamic = <T>(value?: T) => ({ nested: { ...value }, fixed: 0 })[key];
const selected = <T>(value?: T) => {
  const result = { nested: { ...value }, fixed: 0 };
  return result.nested;
};
const destructured = <T>(value?: T) => {
  const { nested } = { nested: { ...value }, fixed: 0 };
  return nested;
};
function inferredIdentity<X>(input: X) { return input; }
const inferred = <T>(value?: T) => inferredIdentity({ ...value });
const mapped = <T>(value?: T) => [value].map(item => ({ ...item }));
declare function identityTag<X>(strings: TemplateStringsArray, value: X): X;
const tagged = <T>(value?: T) => identityTag` + "`" + `${{ ...value }}` + "`" + `;
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
declare function identityObject<X>(value: X): X;
declare function tupleIdentity<X>(...values: [X]): X;
const explicit = <T>(value?: T) => identityObject<T>({ ...value });
const spreadArgument = <T>(value?: T) => tupleIdentity(...[{ ...value }]);
const arrayDestructured = <T>(value?: T) => {
  const [clone] = [{ ...value }];
  return clone;
};
const arraySelected = <T>(value?: T) => [{ ...value }][0];
const spreadOverwrites = <T>(value?: T) =>
  ({ clone: 0, ...{ clone: { ...value } } }).clone;
const objectRest = <T>(value?: T) => {
  const { fixed, ...rest } = { fixed: 0, clone: { ...value } };
  return rest;
};
const localCall = <T>(value?: T) => {
  const getClone = () => ({ ...value });
  return getClone();
};
const immediateCall = <T>(value?: T) => (() => ({ ...value }))();
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
declare const flag: boolean;
const arrayRest = <T>(value?: T) => {
  const [, ...rest] = [0, { ...value }];
  return rest;
};
const objectRestSpreadWins = <T>(value?: T) => {
  const { ...rest } = { clone: 0, ...{ clone: { ...value } } };
  return rest;
};
const defaultBinding = <T>(value?: T) => {
  const { clone = { ...value } } = {};
  return clone;
};
const selectedConditional = <T>(value?: T) =>
  (flag ? { clone: { ...value } } : { clone: 0 }).clone;
const nestedObjectBinding = <T>(value?: T) => {
  const { outer: { clone } } = { outer: { clone: { ...value } } };
  return clone;
};
const nestedArrayBinding = <T>(value?: T) => {
  const [[clone]] = [[{ ...value }]];
  return clone;
};
interface StructuralInput<X> { value: X }
type StructuralAlias<X> = { value: X };
declare function structuralResult<X>(value: StructuralInput<X>): X;
declare function structuralAliasResult<X>(value: StructuralAlias<X>): X;
const structuralInference = <T>(value?: T) =>
  structuralResult({ value: { ...value } });
const structuralAliasInference = <T>(value?: T) =>
  structuralAliasResult({ value: { ...value } });
const inferredDefaultParameter = <T>(value?: T, clone = { ...value }) => 0;
const returnedFunctionDefault = <T>(value?: T) =>
  (clone = { ...value }) => 0;
const returnedMethodDefault = <T>(value?: T) => ({
  clone(input = { ...value }) { return input; },
});
const defaultOnlyAndReturned = <T>(clone = { ...({} as T) }) => clone;
const doubleReturnOnly = <T>() => ({
  first: { ...({} as T) },
  second: { ...({} as T) },
});
declare function mappedResult<X>(value: { [K in keyof X]: X[K] }): X;
declare function partialResult<X>(value: Partial<X>): X;
declare function fixedMappedResult<X>(value: { [K in 'value']: X }): X;
const mappedInference = <T>(value?: T) =>
  mappedResult({ clone: { ...value } });
const partialInference = <T>(value?: T) =>
  partialResult({ clone: { ...value } });
const fixedMappedInference = <T>(value?: T) =>
  fixedMappedResult({ value: { ...value } });
const constrainedObject = <T extends object>(value?: T) => ({ ...value });
const constrainedRecord = <T extends Record<string, unknown>>(value?: T) => ({ ...value });
const mixedLostReturn = <T>() => ({ direct: null as T, ...({} as T | undefined) });
const nestedDoubleReturnOnly = <T>() => {
  const inner = { first: { ...({} as T) }, second: { ...({} as T) } };
  return { ...inner };
};
const duplicateLocalReturnOnly = <T>() => {
  const clone = { ...({} as T | undefined) };
  return { first: clone, second: clone };
};
const duplicateSelectionReturnOnly = <T>() => {
  const result = { clone: { ...({} as T | undefined) } };
  return { first: result.clone, second: result.clone };
};
const wholeObjectSpreadWins = <T>(value?: T) =>
  ({ clone: 0, ...{ clone: { ...value } } });
class CloneMethod { clone() {} }
const classMethodDoesNotOverwrite = <T>(value?: T) =>
  ({ clone: { ...value }, ...new CloneMethod() });
const returnOnlyArray = <T>() => [{ ...({} as T | undefined) }];
const promisedReturnOnlyArray = <T>() => Promise.resolve([{ ...({} as T | undefined) }]);
declare function pairResult<A, B>(left: A, right: B): { left: A; right: B };
const pairCallReturnOnly = <T>() => pairResult(
  { ...({} as T | undefined) },
  { ...({} as T | undefined) },
);
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
const method = <T>(value?: T) => ({ ...({} as { method(): T }) });
const promise = <T>(value?: T) => ({ ...({} as Promise<T>) });
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		// Adversarial differential cases: legacy inference also preserves T
		// through nullable object constraints, object intersections, &&, distinct
		// union branches, and computed string index signatures.
		{
			Code: `
declare const flag: boolean;
declare const key: 'clone' | 'fixed';
const nullableConstraint = <T extends object | undefined>(value?: T) => ({ ...value });
const objectIntersection = <T>(value: (T & object) | undefined) => ({ ...value });
const logicalAnd = <T>(value?: T) => ({ ...(value && value) });
const transitiveConstraint = <U, T extends U>(value?: T, echo?: U) => ({ ...value });
const distinctBranches = <T>() => flag
  ? { left: { ...({} as T | undefined) } }
  : { right: { ...({} as T | undefined) } };
function distinctBlockReturns<T>() {
  if (flag) return { left: { ...({} as T | undefined) } };
  return { right: { ...({} as T | undefined) } };
}
const computedIndex = <T>() => ({ [key]: { ...({} as T | undefined) } });
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		{
			Code: `
type Mapped<X> = { [K in keyof X]: X[K] };
type StringMap<X> = { [key: string]: X };
const mapped = <T>(value?: T) => ({ ...({} as Mapped<T>) });
const indexed = <T>(value?: T) => ({ ...({} as StringMap<T>) });
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		// Object spread copies public class fields, so their value types remain
		// part of the inferred return surface.
		{
			Code: `
class PublicFields<X> {
  first!: X;
  second?: X;
}
const clone = <T>(value?: T) => ({ ...({} as PublicFields<T>) });
`,
			TSConfig: "tsconfig.default-strictness.json",
		},
		// Explicitly disabling strictness already gives tsgo the legacy inference;
		// the recovery path is only for the omitted-option default mismatch.
		{
			Code:     `const cloneObject = <T>(value?: T) => ({ nested: { ...value } });`,
			TSConfig: "tsconfig.unstrict.json",
		},

		// ---- Branch lock-in: isTypeParameterRepeatedInAST self-reference exclusion ----
		// Locks in upstream isTypeParameterRepeatedInAST() arm: a reference to
		// T inside T's own constraint clause doesn't count toward repetition,
		// so this needs the type-checker phase; T's constraint (Base<T>) and
		// its use as a direct generic argument there both correctly resolve
		// to "used multiple times" once the checker phase runs.
		{Code: `
interface Base<T> {
  value: T;
}
declare function recursive<T extends Base<T>>(x: T): T;
`},

		// ---- Branch lock-in: countTypeParameterUsage() KindConstructor arm ----
		// A class's constructor is a member visited with fromClass=true; the
		// special functionLikeType branch (visitSignature via
		// GetSignatureFromDeclaration) must fire for it rather than
		// GetTypeAtLocation, or the class's own T count comes out too low.
		{Code: `
class Box<T> {
  value: T;
  constructor(value: T) {
    this.value = value;
  }
}
`},

		// ---- Branch lock-in: collectTypeParameterUsageCounts() readonly array vs mutable array ----
		{Code: `declare function makeReadonlyArrayParam<T>(input: readonly T[]): T;`},
		{Code: `
// Mutable array used as a parameter (not a return type) counts as single-use,
// matching upstream's isReturnType-gated special case for Array/ReadonlyArray.
declare function identityArray<T>(input: T[]): T[];
`},

		// ---- Branch lock-in: generic type-alias reference (type.aliasTypeArguments) ----
		{Code: `
type Wrapper<T> = { value: T };
declare function wrap<T>(value: T): Wrapper<T>;
`},

		// ---- Dimension 4: private identifier / computed key on a class ----
		{Code: `
class Container<T> {
  #value: T;
  constructor(value: T) {
    this.#value = value;
  }
  get(): T {
    return this.#value;
  }
}
`},

		// ---- Dimension 4: polymorphic `this` type in a member signature ----
		// A `this` type carries the same type-parameter type flag as a real type
		// parameter, but its symbol is the declaring interface, so its first
		// declaration is not a type parameter declaration. It must be skipped
		// rather than blindly treated as one. Here T is used twice, so nothing
		// is reported; see the invalid cases for the single-use counterpart.
		{Code: `
interface Box<T> {
  value: T;
  take(value: T): void;
  self(): this;
}
declare function f<T>(b: Box<T>): void;
`},

		// ---- Branch lock-in: canHaveTypeArgumentsList() TypeQuery arm ----
		// `typeof g<T>` attaches its type arguments to a TypeQuery rather than a
		// TypeReference; T must still register as a direct generic-argument use.
		{Code: `
declare function g<U>(u: U): U;
declare function f<T>(x: typeof g<T>): void;
`},

		// ---- Branch lock-in: isDirectGenericTypeArgumentUsage() paren peeling ----
		// tsgo wraps `(T)` in a ParenthesizedType that ESLint's AST doesn't
		// have, so the node sitting in the outer type-argument list is that
		// wrapper rather than the reference itself. Both nesting depths must
		// still register as a direct generic-argument use.
		{Code: `class Boxes<T> extends Array<(T)> {}`},
		{Code: `
interface Holder<A> {
  value: A;
}
declare function hold<T>(x: Holder<((T))>): void;
`},

		// ---- Upstream v8.67: mapped-type `as` clause counts as a use ----
		// Upstream now visits MappedType.nameType, so a key remapped through T is
		// correctly recognized as a second signature position.
		{Code: `<T extends string>(t: T) => t as { [K in 'a' as T]: 0 };`},

		// ---- Deliberate divergence: every index signature counts, not just
		// string/number ----
		// An index signature instantiates its value type once per key, so T is
		// used many times over. ESLint can only read the string and number index
		// types, so it counts symbol-keyed and pattern-keyed signatures as no use
		// at all and reports T here, suggesting a fix that erases the type.
		{Code: `declare function symbolKeyed<T>(x: number): { [k: symbol]: T };`},
		{Code: "declare function patternKeyed<T>(x: number): { [k: `a${string}`]: T };"},
		{Code: `declare function symbolKeyedParam<T>(x: { [k: symbol]: T }): void;`},
	}, []rule_tester.InvalidTestCase{
		// ---- Real-user: inferred object-spread return strictness boundaries ----
		// With strict null checks enabled, TypeScript 5.9 infers `{}` here and
		// upstream reports T. The compatibility path must not suppress it.
		{
			Code: `const cloneObject = <T>(value?: T) => ({ ...value });`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the function signature.",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const cloneObject = (value?: unknown) => ({ ...value });`,
				}},
			}},
		},
		// An assertion on the spread operand erases T before return inference,
		// so upstream reports this even when strict null checks are disabled.
		{
			Code:     `const cloneObject = <T>(value?: T) => ({ ...(value as object) });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the function signature.",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const cloneObject = (value?: unknown) => ({ ...(value as object) });`,
				}},
			}},
		},
		// A generic used only by the inferred return still occupies just one
		// signature position and therefore remains reportable.
		{
			Code:     `const createObject = <T>() => ({ ...({} as T) });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the function signature.",
				Line:      1,
				Column:    23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const createObject = () => ({ ...({} as unknown) });`,
				}},
			}},
		},
		// A reachable spread is not enough on its own: if a call, sequence, or
		// property selection erases the spread result from the return type, T is
		// still used in only one signature position and must be reported.
		{
			Code: `declare function discardObject(value: object): number;
const discarded = <T>(value?: T) => discardObject({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare function discardObject(value: object): number;
const discarded = (value?: unknown) => discardObject({ ...value });`,
				}},
			}},
		},
		{
			Code: `declare function returnEmptyObject(value: object): {};
const discarded = <T>(value?: T) => returnEmptyObject({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare function returnEmptyObject(value: object): {};
const discarded = (value?: unknown) => returnEmptyObject({ ...value });`,
				}},
			}},
		},
		{
			Code:     `const discarded = <T>(value?: T) => ({ nested: { ...value }, fixed: 0 }).fixed;`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const discarded = (value?: unknown) => ({ nested: { ...value }, fixed: 0 }).fixed;`,
				}},
			}},
		},
		{
			Code:     `const explicit = <T>(value?: T): object => ({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const explicit = (value?: unknown): object => ({ ...value });`,
				}},
			}},
		},
		{
			Code: `const discarded = <T>(value?: T) => {
  const clone = { ...value };
  return 0;
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const discarded = (value?: unknown) => {
  const clone = { ...value };
  return 0;
};`,
				}},
			}},
		},
		{
			Code: `type EmptyPick<X> = Pick<X, never>;
const clone = <T>(value?: T) => ({ ...({} as EmptyPick<T>) });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `type EmptyPick<X> = Pick<X, never>;
const clone = (value?: unknown) => ({ ...({} as EmptyPick<unknown>) });`,
				}},
			}},
		},
		{
			Code: `type PhantomMapped<X> = { [K in never]: X };
const clone = <T>(value?: T) => ({ ...({} as PhantomMapped<T>) });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `type PhantomMapped<X> = { [K in never]: X };
const clone = (value?: unknown) => ({ ...({} as PhantomMapped<unknown>) });`,
				}},
			}},
		},
		{
			Code:     `const clone = <T>(value?: T) => ({ ...({} as () => T) });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = (value?: unknown) => ({ ...({} as () => unknown) });`,
				}},
			}},
		},
		// Explicit type arguments and NoInfer disconnect a spread argument from
		// the generic return, and a later property overwrite removes it from the
		// selected return surface.
		{
			Code: `declare function identityObject<X>(value: X): X;
const clone = <T>(value?: T) => identityObject<object>({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare function identityObject<X>(value: X): X;
const clone = (value?: unknown) => identityObject<object>({ ...value });`,
				}},
			}},
		},
		{
			Code: `declare function noInferIdentity<X = object>(value: NoInfer<X>): X;
const clone = <T>(value?: T) => noInferIdentity({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare function noInferIdentity<X = object>(value: NoInfer<X>): X;
const clone = (value?: unknown) => noInferIdentity({ ...value });`,
				}},
			}},
		},
		{
			Code: `const clone = <T>(value?: T) =>
  ({ clone: { ...value }, ...{ clone: 0 } }).clone;`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) =>
  ({ clone: { ...value }, ...{ clone: 0 } }).clone;`,
				}},
			}},
		},
		{
			Code: `const clone = <T>(value?: T) =>
  ({ clone: { ...value }, ...{ clone: 0 } });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) =>
  ({ clone: { ...value }, ...{ clone: 0 } });`,
				}},
			}},
		},
		{
			Code: `const clone = <T>(value?: T) => {
  const fixed = { clone: 0 };
  return { clone: { ...value }, ...fixed };
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) => {
  const fixed = { clone: 0 };
  return { clone: { ...value }, ...fixed };
};`,
				}},
			}},
		},
		{
			Code: `const clone = <T>(value?: T) => {
  const { generic, ...rest } = { generic: { ...value }, fixed: 0 };
  return rest;
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) => {
  const { generic, ...rest } = { generic: { ...value }, fixed: 0 };
  return rest;
};`,
				}},
			}},
		},
		{
			Code: `const clone = <T>(value?: T) => {
  const { ...rest } = { generic: { ...value }, ...{ generic: 0 } };
  return rest;
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) => {
  const { ...rest } = { generic: { ...value }, ...{ generic: 0 } };
  return rest;
};`,
				}},
			}},
		},
		{
			Code: `const clone = <T>(value?: T) => {
  const { generic = { ...value } }: { generic?: object } = {};
  return generic;
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) => {
  const { generic = { ...value } }: { generic?: object } = {};
  return generic;
};`,
				}},
			}},
		},
		{
			Code: `const clone = <T>(value?: T) => {
  const [fixed] = [0, { ...value }];
  return fixed;
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) => {
  const [fixed] = [0, { ...value }];
  return fixed;
};`,
				}},
			}},
		},
		{
			Code: `interface PhantomInput<X> { fixed: number }
declare function phantomResult<X>(value: PhantomInput<X>): X;
const clone = <T>(value?: T) => phantomResult({ fixed: 0, ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `interface PhantomInput<X> { fixed: number }
declare function phantomResult<X>(value: PhantomInput<X>): X;
const clone = (value?: unknown) => phantomResult({ fixed: 0, ...value });`,
				}},
			}},
		},
		{
			Code: `type PhantomInput<X> = { fixed: number };
declare function phantomResult<X>(value: PhantomInput<X>): X;
const clone = <T>(value?: T) => phantomResult({ fixed: 0, ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `type PhantomInput<X> = { fixed: number };
declare function phantomResult<X>(value: PhantomInput<X>): X;
const clone = (value?: unknown) => phantomResult({ fixed: 0, ...value });`,
				}},
			}},
		},
		{
			Code:     `const clone = <T>(value?: T, copy: object = { ...value }) => 0;`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = (value?: unknown, copy: object = { ...value }) => 0;`,
				}},
			}},
		},
		{
			Code: `const clone = <T>() => {
  const inner = { ...({} as T) };
  return { ...inner };
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = () => {
  const inner = { ...({} as unknown) };
  return { ...inner };
};`,
				}},
			}},
		},
		{
			Code: `declare function erasedMappedResult<X>(value: { [K in keyof X as never]: X[K] }): X;
const clone = <T>(value?: T) => erasedMappedResult({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare function erasedMappedResult<X>(value: { [K in keyof X as never]: X[K] }): X;
const clone = (value?: unknown) => erasedMappedResult({ ...value });`,
				}},
			}},
		},
		{
			Code:     `const clone = <T extends string>(value?: T) => ({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = (value?: string) => ({ ...value });`,
				}},
			}},
		},
		{
			Code:     `const clone = <T extends unknown>(value?: T) => ({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = (value?: unknown) => ({ ...value });`,
				}},
			}},
		},
		{
			Code:     `const clone = <T extends object>() => ({ ...({} as T | undefined) });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = () => ({ ...({} as object | undefined) });`,
				}},
			}},
		},
		{
			Code:     `const clone = <T>() => [{ ...({} as T | undefined) }] as const;`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = () => [{ ...({} as unknown | undefined) }] as const;`,
				}},
			}},
		},
		{
			Code:     `const clone = <T>() => ({ list: [{ ...({} as T | undefined) }] });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = () => ({ list: [{ ...({} as unknown | undefined) }] });`,
				}},
			}},
		},
		{
			Code: `declare const flag: boolean;
const clone = <T>() => flag
  ? { kind: 'a', ...({} as T | undefined) }
  : { kind: 'b', ...({} as T | undefined) };`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare const flag: boolean;
const clone = () => flag
  ? { kind: 'a', ...({} as unknown | undefined) }
  : { kind: 'b', ...({} as unknown | undefined) };`,
				}},
			}},
		},
		{
			Code: `declare function sameResult<X>(left: X, right: X): X;
const clone = <T>() => sameResult(
  { ...({} as T | undefined) },
  { ...({} as T | undefined) },
);`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare function sameResult<X>(left: X, right: X): X;
const clone = () => sameResult(
  { ...({} as unknown | undefined) },
  { ...({} as unknown | undefined) },
);`,
				}},
			}},
		},
		// Repeating the same generic spread on one object surface collapses to a
		// single T in TypeScript's inferred intersection, so it remains sole.
		{
			Code:     `const clone = <T>() => ({ ...({} as T | undefined), ...({} as T | undefined) });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = () => ({ ...({} as unknown | undefined), ...({} as unknown | undefined) });`,
				}},
			}},
		},
		// Distinct callee type parameters that are returned as a naked
		// intersection collapse when both arguments infer the same outer T.
		{
			Code:     `const clone = <T>() => Object.assign({ ...({} as T | undefined) }, { ...({} as T | undefined) });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = () => Object.assign({ ...({} as unknown | undefined) }, { ...({} as unknown | undefined) });`,
				}},
			}},
		},
		// Explicit annotations on locals and default parameters erase the spread
		// from the inferred signature even when that annotated value is returned.
		{
			Code: `const clone = <T>(value?: T) => {
  const copy: object = { ...value };
  return copy;
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) => {
  const copy: object = { ...value };
  return copy;
};`,
				}},
			}},
		},
		{
			Code:     `const clone = <T>(value?: T, copy: object = { ...value }) => copy;`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = (value?: unknown, copy: object = { ...value }) => copy;`,
				}},
			}},
		},
		// Primitive intersections and single computed properties do not preserve
		// multiple generic positions in the inferred return type.
		{
			Code:     `const clone = <T>(value: (T & string) | undefined) => ({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = (value: (unknown & string) | undefined) => ({ ...value });`,
				}},
			}},
		},
		{
			Code:     `const clone = <U, T extends U & string>(value?: T, echo?: U) => ({ ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const clone = <U>(value?: U & string, echo?: U) => ({ ...value });`,
				}},
			}},
		},
		{
			Code: `declare const key: 'clone';
const clone = <T>() => ({ [key]: { ...({} as T | undefined) } });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare const key: 'clone';
const clone = () => ({ [key]: { ...({} as unknown | undefined) } });`,
				}},
			}},
		},
		{
			Code: `const clone = <T>(value?: T) => {
  const { item }: { item: object } = { item: { ...value } };
  return item;
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const clone = (value?: unknown) => {
  const { item }: { item: object } = { item: { ...value } };
  return item;
};`,
				}},
			}},
		},
		// Class methods, accessors, and private/protected fields are absent from
		// the enumerable surface copied by object spread. Each class uses X twice
		// so the diagnostic isolates the enclosing function.
		{
			Code: `
class MethodSurface<X> {
  read(input: X): X { return input; }
}
const clone = <T>(value?: T) => ({ ...({} as MethodSurface<T>) });
`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class MethodSurface<X> {
  read(input: X): X { return input; }
}
const clone = (value?: unknown) => ({ ...({} as MethodSurface<unknown>) });
`,
				}},
			}},
		},
		{
			Code: `
class AccessorSurface<X> {
  get value(): X { throw new Error(); }
  set value(input: X) {}
}
const clone = <T>(value?: T) => ({ ...({} as AccessorSurface<T>) });
`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class AccessorSurface<X> {
  get value(): X { throw new Error(); }
  set value(input: X) {}
}
const clone = (value?: unknown) => ({ ...({} as AccessorSurface<unknown>) });
`,
				}},
			}},
		},
		{
			Code: `
class PrivateSurface<X> {
  private first!: X;
  private second!: X;
}
const clone = <T>(value?: T) => ({ ...({} as PrivateSurface<T>) });
`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class PrivateSurface<X> {
  private first!: X;
  private second!: X;
}
const clone = (value?: unknown) => ({ ...({} as PrivateSurface<unknown>) });
`,
				}},
			}},
		},
		{
			Code: `
class ProtectedSurface<X> {
  protected first!: X;
  protected second!: X;
}
const clone = <T>(value?: T) => ({ ...({} as ProtectedSurface<T>) });
`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      6,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class ProtectedSurface<X> {
  protected first!: X;
  protected second!: X;
}
const clone = (value?: unknown) => ({ ...({} as ProtectedSurface<unknown>) });
`,
				}},
			}},
		},
		{
			Code: `declare function chooseRight<A, B>(left: A, right: B, echo: A): B;
const discarded = <T>(value?: T) => chooseRight({ ...value }, {}, { ...value });`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare function chooseRight<A, B>(left: A, right: B, echo: A): B;
const discarded = (value?: unknown) => chooseRight({ ...value }, {}, { ...value });`,
				}},
			}},
		},
		{
			Code:     `const asserted = <T>(value?: T) => ({ ...value }) as object;`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const asserted = (value?: unknown) => ({ ...value }) as object;`,
				}},
			}},
		},
		{
			Code: `const discarded = <T>(value?: T) => {
  const result = { nested: { ...value }, fixed: 0 };
  return result.fixed;
};`,
			TSConfig: "tsconfig.default-strictness.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `const discarded = (value?: unknown) => {
  const result = { nested: { ...value }, fixed: 0 };
  return result.fixed;
};`,
				}},
			}},
		},

		// A comment after the separator belongs to the following type parameter.
		// Removing the first parameter must preserve it, matching upstream.
		{
			Code: `function f<T, /* keep */ U>(x: T, y: U): U { return y; }`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `function f</* keep */ U>(x: unknown, y: U): U { return y; }`,
				}},
			}},
		},
		{
			Code: "function f<T, // keep\n U>(x: T, y: U): U { return y; }",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    "function f<// keep\n U>(x: unknown, y: U): U { return y; }",
				}},
			}},
		},

		// ---- Branch lock-in: isDirectGenericTypeArgumentUsage() paren peeling ----
		// Peeling the ParenthesizedType must not lose the Array/ReadonlyArray
		// exclusion: `Array<(T)>` still defers to the type-checker phase, which
		// counts a mutable array parameter as a single use.
		{
			Code: `declare function collect<T>(x: Array<(T)>): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the function signature.",
				Line:      1,
				Column:    26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function collect(x: Array<(unknown)>): void;`,
				}},
			}},
		},

		// ---- Deliberate divergence: mapped-type `as` clause counts as a use ----
		// The counterpart to the valid case above: with the parameter list no
		// longer mentioning U, the `as` clause is U's only position, so it is
		// reported as used only once. ESLint reports it as never used.
		{
			Code: `declare function remap<T extends string, U extends string>(x: number): { [K in T as U]: 0 };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Message:   "Type parameter T is used only once in the function signature.",
					Line:      1,
					Column:    24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output:    `declare function remap<U extends string>(x: number): { [K in string as U]: 0 };`,
					}},
				},
				{
					MessageId: "sole",
					Message:   "Type parameter U is used only once in the function signature.",
					Line:      1,
					Column:    42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output:    `declare function remap<T extends string>(x: number): { [K in T as string]: 0 };`,
					}},
				},
			},
		},

		// ---- Dimension 4: polymorphic `this` type in a class member ----
		// Regression lock-in: `this` types reach the type-parameter branch of the
		// type walk with a class symbol, whose first declaration is a
		// ClassDeclaration. Reporting T here (used only once, in `value: T`)
		// matches ESLint; the `this` return contributes nothing.
		{
			Code: `
class Box<T> {
  value: T;
  self(): this {
    return this;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the class signature.",
				Line:      2,
				Column:    11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class Box {
  value: unknown;
  self(): this {
    return this;
  }
}
`,
				}},
			}},
		},
		// The same shape with an inferred (rather than annotated) `this` return.
		{
			Code: `
class Box<T> {
  value: T;
  self() {
    return this;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class Box {
  value: unknown;
  self() {
    return this;
  }
}
`,
				}},
			}},
		},

		// ---- Real-user: typescript-eslint/typescript-eslint#11528 (not planned) ----
		// The maintainers confirmed this sibling shape from the same thread
		// SHOULD be reported: TData appears only in the return type itself
		// (not as a generic argument of some other type there); a call inside
		// the body passing TData explicitly doesn't count, since body usage
		// is outside the declaring signature.
		{
			Code: `
interface IError<TData> {
  error: true;
  data: TData;
}

declare const checkIsError: <TData>(anything: unknown) => anything is IError<TData>;

const extractErrorData = <TData,>(error?: unknown): TData | null => {
  return checkIsError<TData>(error) ? error.data : null;
};
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      9,
				Column:    27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
interface IError<TData> {
  error: true;
  data: TData;
}

declare const checkIsError: <TData>(anything: unknown) => anything is IError<TData>;

const extractErrorData = (error?: unknown): unknown | null => {
  return checkIsError<unknown>(error) ? error.data : null;
};
`,
				}},
			}},
		},

		// ---- Dimension 4: TS non-null assertion receiver on a reference ----
		{
			Code: `declare function get<T>(): T!;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): unknown!;`,
				}},
			}},
		},

		// ---- Dimension 4: parenthesized receiver, single and multi-level ----
		{
			Code: `declare function get<T>(): (T);`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): (unknown);`,
				}},
			}},
		},
		{
			Code: `declare function get<T>(): ((T));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): ((unknown));`,
				}},
			}},
		},

		// ---- Real-user: typescript-eslint/typescript-eslint#9961 (open) ----
		// Class member usages (get accessor and method) still get reported
		// even though the constrained value is only ever read, matching
		// upstream's acknowledged-but-unfixed class-member false positive.
		{
			Code: `
class ValueClassGetter<T> {
  get value(): T {
    return null!;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the class signature.",
				Line:      2,
				Column:    24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class ValueClassGetter {
  get value(): unknown {
    return null!;
  }
}
`,
				}},
			}},
		},
		{
			Code: `
class ValueClassMethod<T> {
  getValue(): T {
    return null!;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class ValueClassMethod {
  getValue(): unknown {
    return null!;
  }
}
`,
				}},
			}},
		},

		// ---- Dimension 2: deeply nested (3+ levels) function scoping ----
		// Only the innermost `identity`'s own T is reported; the outer two
		// same-named T's are each used twice in their own signatures (param +
		// return) and must not "bleed" into each other's counts.
		{
			Code: `
function outer<T>(a: T): T {
  function middle<T>(b: T): T {
    function identity<T>(c: T): void {}
    return b;
  }
  return a;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the function signature.",
				Line:      4,
				Column:    23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function outer<T>(a: T): T {
  function middle<T>(b: T): T {
    function identity(c: unknown): void {}
    return b;
  }
  return a;
}
`,
				}},
			}},
		},

		// ---- Dimension 4: graceful degradation — overload signature (body-absent) ----
		{
			Code: `
declare class Store {
  get<T>(key: string): T;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
declare class Store {
  get(key: string): unknown;
}
`,
				}},
			}},
		},

		// ---- Branch lock-in: buildReplaceWithConstraintFixes() union constraint
		// needs parens when substituted into an intersection position ----
		{
			Code: `
function f<T extends string | number>(x: T & { tag: true }): void {}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function f(x: (string | number) & { tag: true }): void {}
`,
				}},
			}},
		},

		// ---- Branch lock-in: buildReplaceWithConstraintFixes() union constraint
		// needs parens when substituted into an indexed-access position ----
		{
			Code: `
type Lookup = { a: 1; b: 2 };
function f<T extends 'a' | 'b'>(x: Lookup[T]): void {}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
type Lookup = { a: 1; b: 2 };
function f(x: Lookup[('a' | 'b')]): void {}
`,
				}},
			}},
		},

		// ---- Deliberate divergence: the constraint is parenthesized whenever
		// the type grammar demands it, not only for unions, intersections and
		// conditionals ----
		// A function, constructor or type-operator constraint binds looser than
		// the array, indexed-access and `keyof` positions it can land in, so it
		// needs the same treatment. ESLint checks only the three complex kinds
		// and emits `() => void[]`, `readonly string[][]` and
		// `keyof string | number`, each of which denotes a different type than
		// the code being replaced.
		{
			Code: `declare function callbacks<T extends () => void>(x: T[]): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function callbacks(x: (() => void)[]): void;`,
				}},
			}},
		},
		{
			Code: `declare function ctors<T extends new () => void>(x: T[]): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function ctors(x: (new () => void)[]): void;`,
				}},
			}},
		},
		{
			Code: `declare function frozen<T extends readonly string[]>(x: T[]): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function frozen(x: (readonly string[])[]): void;`,
				}},
			}},
		},
		{
			Code: `declare function keys<T extends string | number>(x: keyof T): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function keys(x: keyof (string | number)): void;`,
				}},
			}},
		},
		{
			Code: `declare function check<T extends () => void>(x: T extends object ? 1 : 2): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function check(x: (() => void) extends object ? 1 : 2): void;`,
				}},
			}},
		},
		{
			// A type query is already a primary type, so the array position
			// takes it as written.
			Code: `declare const seed: string;
declare function seeded<T extends typeof seed>(x: T[]): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `declare const seed: string;
declare function seeded(x: typeof seed[]): void;`,
				}},
			}},
		},

		// ---- Branch lock-in: no constraint at all replaces with "unknown" ----
		{
			Code: `function bare<T>(x: T) {}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `function bare(x: unknown) {}`,
				}},
			}},
		},
	})
}

// TestNoUnnecessaryTypeParametersEditDemand exercises the four edit-demand
// modes for a representative diagnostic (a sole-use suggestion), asserting
// diagnostic count/message/range are unaffected by demand while the
// suggestion artifact only materializes when requested (the rule has no
// autofix, only a suggestion).
func TestNoUnnecessaryTypeParametersEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"declare function take<T>(param: T): void;\n",
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:             NoUnnecessaryTypeParametersRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: NoUnnecessaryTypeParametersRule.RequiresTypeInfo,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoUnnecessaryTypeParametersRule.Run(ctx, nil)
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
		return diagnostic
	}

	want := withoutEdits(diagnostics[rule.EditDemandAll][0])
	for demand, got := range diagnostics {
		if got := withoutEdits(got[0]); !reflect.DeepEqual(got, want) {
			t.Errorf("diagnostic changed for demand %d:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
		if got[0].FixesPtr != nil {
			t.Errorf("demand %d unexpectedly has autofixes (rule only has suggestions)", demand)
		}
	}

	suggestionOnly := diagnostics[rule.EditDemandSuggestion][0].Suggestions
	allEdits := diagnostics[rule.EditDemandAll][0].Suggestions
	if suggestionOnly == nil || !reflect.DeepEqual(suggestionOnly, allEdits) {
		t.Fatalf("suggestions differ between suggestion-only and all-edits demand")
	}
	if len(*suggestionOnly) != 1 || (*suggestionOnly)[0].Message.Id != "replaceUsagesWithConstraint" {
		t.Fatalf("suggestions = %#v, want 1 replaceUsagesWithConstraint suggestion", *suggestionOnly)
	}
	if diagnostics[rule.EditDemandNone][0].Suggestions != nil ||
		diagnostics[rule.EditDemandAutofix][0].Suggestions != nil {
		t.Errorf("suggestions attached without suggestion demand")
	}
}
