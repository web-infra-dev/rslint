package no_misused_promises

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoMisusedPromisesArgumentTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedPromisesRule,
		[]rule_tester.ValidTestCase{
			{Code: `declare function invoke(): void; invoke();`},
			{Code: `declare function invoke(x: string): void; invoke("value");`},
			// Multiple void overloads followed by a thenable overload must still allow async callbacks.
			{Code: `
interface MultiOverload {
  (a: string, cb: () => void): void;
  (a: number, cb: () => void): void;
  (a: string | number, cb: () => Promise<void>): void;
}
declare const multi: MultiOverload;
multi('x', async () => {});
`},
			// Optional/excess signature parameters beyond len(args) are ignored.
			{Code: `
declare function withOptional(a: string, b?: () => void, c?: () => void): void;
declare const s: string;
withOptional(s);
`},
			// Overloaded rest-array parameters where one overload accepts Promise-returning callbacks.
			{Code: `
interface RestOverloads {
  (...cbs: Array<() => void>): void;
  (...cbs: Array<() => Promise<void>>): void;
}
declare const runAll: RestOverloads;
runAll(async () => {}, async () => {});
`},
			// Overloaded tuple rest parameters where one overload accepts Promise-returning callbacks.
			{Code: `
interface TupleRestOverloads {
  (...args: [string, () => void]): void;
  (...args: [string, () => Promise<void>]): void;
}
declare const tupleRestOk: TupleRestOverloads;
tupleRestOk('x', async () => {});
`},
			// Overloaded constructors with zero arguments.
			{Code: `
declare class Runner {
  constructor();
  constructor(cb?: () => void);
}
new Runner();
new Runner;
`},
			// Overloaded constructors with arguments where one overload accepts Promise-returning callbacks.
			{Code: `
declare class AsyncRunner {
  constructor(cb: () => void);
  constructor(cb: () => Promise<void>);
}
new AsyncRunner(async () => {});
`},
			// Generic call instantiated to accept Promise-returning function.
			{Code: `
declare function wrap<T>(cb: T): void;
wrap<() => Promise<void>>(async () => {});
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
export {};
declare global { interface String { (): Promise<void>; } }
declare function invoke(callback: () => void): void;
invoke("value");
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "voidReturnArgument"}},
			},
			// Mixed positions across overloads: arg 0 accepts Promise, arg 1 only accepts void.
			{
				Code: `
interface Mixed {
  (a: () => void, b: () => void): void;
  (a: () => Promise<void>, b: () => void): void;
}
declare const mixed: Mixed;
mixed(async () => {}, async () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnArgument", Line: 7, Column: 23},
				},
			},
			// Multiple void overloads and no thenable overload still report once.
			{
				Code: `
interface AllVoid {
  (a: string, cb: () => void): void;
  (a: number, cb: () => void): void;
}
declare const allVoid: AllVoid;
allVoid('x', async () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnArgument", Line: 7, Column: 14},
				},
			},
			// Rest-array parameter checked across multiple arguments when no overload accepts Promise.
			{
				Code: `
interface VoidRest {
  (...cbs: Array<() => void>): void;
  (a: number, ...cbs: Array<() => void>): void;
}
declare const voidRest: VoidRest;
voidRest(async () => {}, () => {}, async () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnArgument", Line: 7, Column: 10},
					{MessageId: "voidReturnArgument", Line: 7, Column: 36},
				},
			},
			// Tuple rest parameter expecting a void-returning callback.
			{
				Code: `
declare function tupleRestVoid(...args: [string, () => void]): void;
tupleRestVoid('x', async () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnArgument", Line: 3, Column: 20},
				},
			},
			// Overloaded constructors with arguments where all matching overloads expect void.
			{
				Code: `
declare class VoidRunner {
  constructor(cb: () => void);
  constructor(cb: () => void, label: string);
}
new VoidRunner(async () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnArgument", Line: 6, Column: 16},
				},
			},
			// Generic call instantiated to void exercises the post-loop contextual type query.
			{
				Code: `
declare function wrap<T>(cb: T): void;
wrap<() => void>(async () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnArgument", Line: 3, Column: 18},
				},
			},
		})
}

func TestNoMisusedPromisesContextualTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedPromisesRule,
		[]rule_tester.ValidTestCase{
			// JSX: empty expression container and context-independent non-callable expressions.
			{
				Code: `
interface Props {
  a?: string;
  b?: number;
  c?: boolean;
  d?: string;
  e?: string;
  f?: boolean;
  g?: string;
  onClick?: () => void;
}
declare function Comp(props: Props): null;
declare const val: string;
declare const obj: { prop: string };
declare const flag: boolean;
<Comp a={"x"} b={1} c={true} d={val} e={obj.prop} f={!flag} g={flag ? "a" : "b"} onClick={/* empty */} />;
`,
				Tsx: true,
			},
			// JSX: context-sensitive callback returning void in generic component.
			{
				Code: `
declare function GenericComp<T>(props: { item: T; onSelect: (x: T) => void }): null;
<GenericComp item={{ run: () => Promise.resolve() }} onSelect={(x) => { void x.run(); }} />;
`,
				Tsx: true,
			},
			// Object properties: context-independent values & discriminated-union sync callback & narrowed shorthand.
			{
				Code: `
type Handler =
  | { kind: 'async'; cb: (x: { run(): Promise<void> }) => void }
  | { kind: 'sync'; cb: (x: { run(): number }) => void };
const syncHandler: Handler = { kind: 'sync', cb: (x) => x.run() };

declare const obj: { prop: string };
declare const flag: boolean;
const data: { a: number; b: string; c: string; d: string } = {
  a: 1,
  b: "x",
  c: obj.prop,
  d: flag ? "a" : "b",
};

declare let fn: (() => Promise<void>) | string;
if (typeof fn === 'string') {
  const narrowed: { fn: (() => void) | string } = { fn };
}
`,
			},
			// Return statements: context-independent non-callable values in async/annotated/unannotated functions.
			{
				Code: `
interface User { id: number }
async function getUser(ok: boolean, user: User): Promise<User | null> {
  if (!ok) {
    return null;
  }
  return user;
}
function unannotated(x: string) {
  return x;
}
`,
			},
			// Context-independent identifier referring to a contextually typed callback parameter (sync).
			{
				Code: `
type Task =
  | { kind: 'async'; run: (x: () => Promise<void>) => { cb: () => void } }
  | { kind: 'sync'; run: (x: () => number) => { cb: () => void } };
const syncTask: Task = { kind: 'sync', run: (x) => ({ cb: x }) };

declare function withSpec<T>(spec: { item: T; make: (x: T) => { cb: () => void } }): void;
withSpec({ item: () => 1, make: (x) => ({ cb: x }) });
`,
			},
			// Unannotated getter paired with annotated setter returning a sync function.
			{
				Code: `
class SyncBox {
  set handler(fn: () => void) {}
  get handler() {
    return () => {};
  }
}
`,
			},
			// Unannotated object literals / method maps and destructuring assignments have no contextual type.
			{
				Code: `
declare const api: { save(value: string): number; asyncSave(value: string): Promise<void> };
async function asyncFn() {}
const handlers = {
  save: api.save,
  asyncSave: api.asyncSave,
  asyncFn,
  method() { return 1; },
  async asyncMethod() {},
};
let target: unknown;
({ save: target } = handlers);
`,
			},
			// JS (checkJs): unannotated method map without JSDoc @type.
			{
				Code: `
async function asyncFn() {}
const handlers = { save: asyncFn, asyncFn };
`,
				FileName: "jsdoc.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			// Typed array of synchronous methods and Record<string, unknown> method references.
			{
				Code: `
declare const api: { save(value: string): number; asyncSave(value: string): Promise<void> };
const commands: Array<{ run(): void }> = [
  { run(): void {} },
  { run() { return 1; } },
];
const handlers: Record<string, unknown> = {
  save: api.save,
  asyncSave: api.asyncSave,
};
`,
			},
			// strictNullChecks: false — unknown does not gain call signatures from Object.
			{
				Code: `
export {};
declare global {
  interface Object {
    (): void;
  }
}
declare const api: { asyncSave(): Promise<void> };
const handlers: Record<string, unknown> = {
  save: api.asyncSave,
};
`,
				TSConfig: "tsconfig.strict-with-null-checks-off.json",
			},
			// Unannotated string/primitive literal tables (including `as const`) and
			// unannotated single-property literal objects have no contextual type,
			// even when global String is augmented to return a Promise.
			{
				Code: `
export {};
declare global {
  interface String {
    (): Promise<void>;
  }
}
export const labels = { save: "Save", cancel: "Cancel" };
export const constLabels = { save: "Save", cancel: "Cancel" } as const;
export const singleProp = { save: "Save" };
export const singlePropConst = { save: "Save" } as const;
export const items = [{ a: "1" }, { a: "2" }];
export const constItems = [{ a: "1" }, { a: "2" }] as const;
`,
			},
			// Typed tables/arrays with plain literal properties and JS unannotated cast
			// without a property-level JSDoc @type remain valid.
			{
				Code: `
export const typedItems: Array<{ a: string }> = [{ a: "1" }, { a: "2" }];
export const typedRecord: Record<string, string> = { a: "1", b: "2" };
export const typedUnknown: Record<string, unknown> = { a: "1", b: 2, c: true, d: null, e: /x/, f: void 0 };
`,
			},
			{
				Code: `
async function asyncFn() {}
const unannotatedCast = {
  a: /** @type {() => Promise<void>} */ (1),
};
const plainLiteralWithType = {
  /** @type {() => void} */
  a: "plain",
};
// Property-level JSDoc @type on a shorthand property also types the
// shorthand declaration itself (checkShorthandPropertyAssignment returns
// node.Type()), so GetTypeAtLocation(node.Name()) is () => void rather than
// () => Promise<void> (matching upstream/main).
const multiPropShorthandWithPropType = {
  label: "ok",
  /** @type {() => void} */
  asyncFn,
};
`,
				FileName: "jsdoc.js",
				TSConfig: "tsconfig.allowJs.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			// JSX: context-sensitive callback inferred from sibling generic prop.
			{
				Code: `
declare function GenericComp<T>(props: { item: T; onSelect: (x: T) => void }): null;
<GenericComp item={{ run: () => Promise.resolve() }} onSelect={(x) => x.run()} />;
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnAttribute", Line: 3},
				},
			},
			// JSX: context-independent expressions returning thenables (property access, ternary, nullish coalescing).
			{
				Code: `
interface Props { onClick: () => void }
declare function Comp(props: Props): null;
declare const handlers: { onSave: () => Promise<void>; onOther: () => number; maybe?: () => Promise<void> };
declare const cond: boolean;
<Comp onClick={handlers.onSave} />;
<Comp onClick={cond ? handlers.onSave : handlers.onOther} />;
<Comp onClick={handlers.maybe ?? handlers.onOther} />;
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnAttribute", Line: 6},
					{MessageId: "voidReturnAttribute", Line: 7},
					{MessageId: "voidReturnAttribute", Line: 8},
				},
			},
			// JSX: primitive prototype augmentation remains checked.
			{
				Code: `
export {};
declare global { interface String { (): Promise<void>; } }
declare function Comp(props: { onClick: () => void }): null;
<Comp onClick={"value"} />;
`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnAttribute", Line: 5},
				},
			},
			// Object properties: context-sensitive callback in discriminated union, conditional, and generic wrapper.
			{
				Code: `
type Handler =
  | { kind: 'async'; cb: (x: { run(): Promise<void> }) => void }
  | { kind: 'sync'; cb: (x: { run(): number }) => void };
declare function wrap<F>(fn: F): F;
declare const cond: boolean;

const h1: Handler = { kind: 'async', cb: (x) => x.run() };
const h2: { cb: (x: { run(): Promise<void> }) => void } = {
  cb: cond ? (x) => x.run() : () => {},
};
const h3: { cb: (x: { run(): Promise<void> }) => void } = {
  cb: wrap((x) => x.run()),
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 8},
					{MessageId: "voidReturnProperty", Line: 10},
					{MessageId: "voidReturnProperty", Line: 13},
				},
			},
			// Object properties: context-independent expressions returning thenables & primitive prototype augmentation.
			{
				Code: `
export {};
declare global { interface String { (): Promise<void>; } }
declare const handlers: { onSave?: () => Promise<void> };
const obj: { a: () => void; b: () => void } = {
  a: handlers.onSave!,
  b: "value",
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 6},
					{MessageId: "voidReturnProperty", Line: 7},
				},
			},
			// Return statements: context-sensitive callback, context-independent identifier, and primitive augmentation.
			{
				Code: `
export {};
declare global { interface String { (): Promise<void>; } }
declare const asyncFn: () => Promise<void>;
function f1(): (x: { run(): Promise<void> }) => void {
  return (x) => x.run();
}
const f2: () => () => void = () => {
  return asyncFn;
};
function f3(): () => void {
  return "value";
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnReturnValue", Line: 6},
					{MessageId: "voidReturnReturnValue", Line: 9},
					{MessageId: "voidReturnReturnValue", Line: 12},
				},
			},
			// Context-independent identifier referring to a contextually typed callback parameter (discriminated union & generic object literal).
			{
				Code: `
type Task =
  | { kind: 'async'; run: (x: () => Promise<void>) => { cb: () => void } }
  | { kind: 'sync'; run: (x: () => number) => { cb: () => void } };
const asyncTask: Task = { kind: 'async', run: (x) => ({ cb: x }) };

declare function withSpec<T>(spec: { item: T; make: (x: T) => { cb: () => void } }): void;
withSpec({ item: async () => {}, make: (x) => ({ cb: x }) });
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 5},
					{MessageId: "voidReturnProperty", Line: 8},
				},
			},
			// Unannotated getter paired with annotated setter returning an async function.
			{
				Code: `
class AsyncBox {
  set handler(fn: () => void) {}
  get handler() {
    return async () => {};
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnReturnValue", Line: 5},
				},
			},
			// Unannotated outer object with nested typed object literal and nested typed properties.
			{
				Code: `
declare const api: { save(): number; asyncSave(): Promise<void> };
const outer = {
  before: api.asyncSave,
  inner: ({ save: async () => {} } satisfies { save: () => void }),
  after: api.asyncSave,
};
const typed: { outer: { inner: () => void } } = {
  outer: { inner: async () => {} },
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 5},
					{MessageId: "voidReturnProperty", Line: 9},
				},
			},
			// Typed object literal with unknown/unannotated nested object literals only reports void-typed properties.
			{
				Code: `
const x: { a: { b: () => void }; c: unknown } = {
  a: { b: async () => {} },
  c: { d: async () => {} },
};
const nested: { first: () => void; make: () => unknown; last: () => void } = {
  first: async () => {},
  make: () => {
    const local = { asyncSave: async () => {} };
    return local;
  },
  last: async () => {},
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 3},
					{MessageId: "voidReturnProperty", Line: 7},
					{MessageId: "voidReturnProperty", Line: 12},
				},
			},
			// JS (checkJs): JSDoc @type on object literal and directly on properties of an unannotated object literal.
			{
				Code: `
async function asyncFn() {}
const typedObj = /** @type {{ save: () => void; asyncFn: () => void }} */ ({
  save: async () => {},
  asyncFn,
});
const typedProps = {
  /** @type {() => void} */
  save: asyncFn,
  /** @type {() => void} */
  inline: async () => {},
};
`,
				FileName: "jsdoc.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 4},
					{MessageId: "voidReturnProperty", Line: 5},
					{MessageId: "voidReturnProperty", Line: 9},
					{MessageId: "voidReturnProperty", Line: 11},
				},
			},
			// Typed array with async method and Record<string, () => void> method reference.
			{
				Code: `
declare const api: { save(value: string): number; asyncSave(value: string): Promise<void> };
const commands: Array<{ run(): void }> = [
  { run(): void {} },
  { async run() {} },
];
const handlers: Record<string, () => void> = {
  save: api.save,
  asyncSave: api.asyncSave,
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 5},
					{MessageId: "voidReturnProperty", Line: 9},
				},
			},
			// Augmented String call signatures on typed properties remain checked
			// (including under strictNullChecks: false).
			{
				Code: `
export {};
declare global {
  interface String {
    (): void;
  }
}
declare const api: { asyncSave(): Promise<void> };
const handlers: Record<string, string> = {
  save: api.asyncSave,
};
`,
				TSConfig: "tsconfig.strict-with-null-checks-off.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 10},
				},
			},
			// JSDoc @type cast on a literal in a typed object literal must not be
			// skipped by isGuaranteedLiteralLike.
			{
				Code: `
const castLiteral = /** @type {{ a: () => void }} */ ({
  a: /** @type {() => Promise<void>} */ (1),
});
`,
				FileName: "jsdoc.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 3},
				},
			},
			{
				Code: `
const outer: { a: unknown; m(): void } = {
  a: { x: 1 },
  async m() {},
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 4},
				},
			},
			// Typed single-property objects, `as const` assertions, and repeated
			// augmented string/template literals remain reported (primitive literal
			// cache must not skip augmented primitive kinds).
			{
				Code: `
export {};
declare global {
  interface String {
    (): Promise<void>;
  }
}
const okNum: { n: number } = { n: 1 };
const typedConst: { save: () => void } = { save: "Save" } as const;
const typedItems: Array<{ a: () => void }> = [
  { a: "first" },
  { a: ` + "`second`" + ` },
];
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 9},
					{MessageId: "voidReturnProperty", Line: 11},
					{MessageId: "voidReturnProperty", Line: 12},
				},
			},
			// JS (checkJs): property-level JSDoc @type on literal, non-literal,
			// and object-method properties in single- and multi-property
			// unannotated object literals (including after a prior sibling caches
			// the enclosing object's nil contextual type), plus typed-object
			// shorthand after an excess/unmatched property.
			{
				Code: `
async function asyncFn() {}
const singlePropLiteral = {
  /** @type {() => void} */
  a: /** @type {() => Promise<void>} */ (1),
};
const multiPropLiteral = {
  label: "ok",
  /** @type {() => void} */
  b: /** @type {() => Promise<void>} */ ("value"),
};
const singlePropNonLiteral = {
  /** @type {() => void} */
  save: asyncFn,
};
const multiPropNonLiteral = {
  first: asyncFn,
  /** @type {() => void} */
  save: asyncFn,
};
const singlePropTypedObjShorthand = /** @type {{ asyncFn: () => void }} */ ({
  asyncFn,
});
const multiPropTypedObjShorthand = /** @type {{ asyncFn: () => void }} */ ({
  extra: asyncFn,
  asyncFn,
});
const multiPropObjectMethods = {
  first: asyncFn,
  /** @type {() => () => void} */
  fullSig() {
    return asyncFn;
  },
  /** @returns {() => void} */
  returnsTag() {
    return asyncFn;
  },
};
`,
				FileName: "jsdoc.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 5},
					{MessageId: "voidReturnProperty", Line: 10},
					{MessageId: "voidReturnProperty", Line: 14},
					{MessageId: "voidReturnProperty", Line: 19},
					{MessageId: "voidReturnProperty", Line: 22},
					{MessageId: "voidReturnProperty", Line: 26},
					{MessageId: "voidReturnReturnValue", Line: 32},
					{MessageId: "voidReturnReturnValue", Line: 36},
				},
			},
			// Typed object literal with an excess property preceding a void-typed
			// property must not cache the enclosing object as having nil context.
			{
				Code: `
declare const extra: string;
const withExcess: { b: () => void } = {
  a: extra,
  b: async () => {},
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 5},
				},
			},
			// Typed tuple arrays and sparse numeric-keyed objects where element 0
			// has nil contextual type must not cache the array as untyped for later
			// void-returning elements.
			{
				Code: `
declare const asyncFn: () => Promise<void>;
const tupleUnknownThenVoid: [unknown, { cb: () => void }] = [
  { label: "first" },
  { cb: asyncFn },
];
const sparseNumeric: { 1: { cb: () => void } } = [
  { a: 1 },
  { cb: asyncFn },
];
const typedArrayOneProp: Array<{ cb: () => void }> = [
  { cb: asyncFn },
  { cb: asyncFn },
];
const typedArrayLiteralThenRef: Array<{ label: string; cb: () => void }> = [
  { label: "Save", cb: asyncFn },
  { label: "Close", cb: asyncFn },
];
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 5},
					{MessageId: "voidReturnProperty", Line: 9},
					{MessageId: "voidReturnProperty", Line: 12},
					{MessageId: "voidReturnProperty", Line: 13},
					{MessageId: "voidReturnProperty", Line: 16},
					{MessageId: "voidReturnProperty", Line: 17},
				},
			},
			// Malformed `static { return }` inside an array inside an object
			// literal method must not corrupt array/object LIFO context stacks.
			{
				Code: `
declare const asyncFn: () => Promise<void>;
const outer = [
  {
    m() {
      const inner = [
        { a: 1 },
        class {
          static {
            return asyncFn;
          }
        },
      ];
      return inner;
    },
    p: 1,
  },
  ({ cb: asyncFn } satisfies { cb: () => void }),
];
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 18},
				},
			},
			// JS (checkJs): untyped array with JSDoc @type cast on a later object
			// element and property-level JSDoc @type inside a later sibling object
			// must still be checked after element 0 caches the array as untyped.
			{
				Code: `
async function asyncFn() {}
const mixedUntypedArray = [
  { label: "first", cb: asyncFn },
  { label: "second", cb: asyncFn },
  /** @type {{ cb: () => void }} */ ({ cb: asyncFn }),
  {
    label: "fourth",
    /** @type {() => void} */
    cb: asyncFn,
  },
];
`,
				FileName: "jsdoc.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 6},
					{MessageId: "voidReturnProperty", Line: 10},
				},
			},
			// Deeply nested untyped objects containing both untyped arrays and a
			// typed leaf object preserve strict O(1) LIFO stack discipline.
			{
				Code: `
declare const asyncFn: () => Promise<void>;
const deep = {
  l1: "a",
  n1: {
    l2: "b",
    rows: [
      { path: "p1", handle: asyncFn },
      { path: "p2", handle: asyncFn },
    ],
    typedLeaf: ({ cb: asyncFn } satisfies { cb: () => void }),
    afterLeaf: asyncFn,
  },
  afterN1: asyncFn,
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 11},
				},
			},
		})
}

func runNoMisusedPromisesMultiFile(
	t *testing.T,
	entryFileName string,
	options *core.CompilerOptions,
	files map[string]string,
) []rule.RuleDiagnostic {
	t.Helper()

	root := fixtures.GetRootDir()
	overlays := make(map[string]string, len(files))
	rootNames := make([]string, 0, len(files))
	for fileName, code := range files {
		filePath := tspath.ResolvePath(root.Dir, fileName)
		overlays[filePath] = code
		rootNames = append(rootNames, filePath)
	}
	fs := utils.NewOverlayVFS(root.FS, overlays)
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(
		false,
		options,
		rootNames,
		host,
	)
	if err != nil {
		t.Fatalf("create lenient program: %v", err)
	}
	entryPath := tspath.ResolvePath(root.Dir, entryFileName)
	sourceFile := program.GetSourceFile(entryPath)
	if sourceFile == nil {
		t.Fatalf("lenient program did not contain %s", entryFileName)
	}

	var diagnostics []rule.RuleDiagnostic
	programs := []*lintprogram.Program{lintprogram.NewFromCompiler(program)}
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{sourceFile.FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:             NoMisusedPromisesRule.Name,
				Severity:         rule.SeverityError,
				RequiresTypeInfo: true,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoMisusedPromisesRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	_, err = linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandNone,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	if err != nil {
		t.Fatalf("run linter: %v", err)
	}
	return diagnostics
}

func TestNoMisusedPromisesJSDocLiteralAndWithStatement(t *testing.T) {
	t.Parallel()

	const ambient = `
interface String {
  (): Promise<void>;
}
declare function takeVoid(cb: () => void): void;
declare const scope: any;
`
	const jsCode = `
async function asyncFn() {}
with (scope) {
  takeVoid("inside-with");
}
const afterWith = /** @type {{ cb: () => void }} */ ({
  cb: "outside-with",
});
const literalAfterLiteral = {
  label: "ok",
  /** @type {() => void} */
  a: "value-after-literal",
};
const literalAfterNonLiteral = {
  first: asyncFn,
  /** @type {() => void} */
  b: "value-after-non-literal",
};
const singlePropLiteral = {
  /** @type {() => void} */
  c: "value-single",
};
const nestedTypedInsideUntyped = {
  a: "plain",
  inner: /** @type {{ cb: () => void }} */ ({
    cb: "nested-typed",
  }),
  /** @type {() => void} */
  afterInner: "after-nested-typed",
};
const flatUntypedSiblings = [
  { id: "row1", cb: asyncFn },
  { id: "row2", cb: asyncFn },
];
const typedTableAfterUntyped = /** @type {{ save: () => void }} */ ({
  save: "typed-after-untyped",
});
`

	diagnostics := runNoMisusedPromisesMultiFile(
		t,
		"check.js",
		&core.CompilerOptions{
			AllowJs: core.TSTrue,
			CheckJs: core.TSTrue,
			Strict:  core.TSTrue,
		},
		map[string]string{
			"ambient.d.ts": ambient,
			"check.js":     jsCode,
		},
	)

	wantSnippets := []string{
		`"outside-with"`,
		`"value-after-literal"`,
		`"value-after-non-literal"`,
		`"value-single"`,
		`"nested-typed"`,
		`"after-nested-typed"`,
		`"typed-after-untyped"`,
	}
	if len(diagnostics) != len(wantSnippets) {
		t.Fatalf("got %d diagnostics, want %d: %#v", len(diagnostics), len(wantSnippets), diagnostics)
	}
	for i, want := range wantSnippets {
		d := diagnostics[i]
		if d.Message.Id != "voidReturnProperty" {
			t.Errorf("diagnostic[%d]: got messageId %q, want %q", i, d.Message.Id, "voidReturnProperty")
		}
		got := jsCode[d.Range.Pos():d.Range.End()]
		if strings.TrimSpace(got) != want {
			t.Errorf("diagnostic[%d]: got snippet %q, want %q", i, got, want)
		}
	}
}

func TestNoMisusedPromisesReturnsOnlyObjectMethodStack(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedPromisesRule,
		[]rule_tester.ValidTestCase{
			{
				Options: []any{map[string]any{"checksVoidReturn": map[string]any{"properties": false, "returns": true}}},
				Code: `
declare const asyncFn: () => Promise<void>;
const untyped = [
  { m1() { return asyncFn; }, m2() { return asyncFn; } },
  { m1() { return asyncFn; }, m2() { return asyncFn; } },
];
`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Options: []any{map[string]any{"checksVoidReturn": map[string]any{"properties": false, "returns": true}}},
				Code: `
declare const asyncFn: () => Promise<void>;
const untyped = {
  m1() { return asyncFn; },
  nested: (() => {
    const typed: { run(): () => void } = {
      run() { return asyncFn; },
    };
    return typed;
  })(),
  m2() { return asyncFn; },
};
const afterTyped: { run(): () => void } = {
  run() { return asyncFn; },
};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnReturnValue", Line: 7, Column: 22},
					{MessageId: "voidReturnReturnValue", Line: 14, Column: 18},
				},
			},
		},
	)
}

func TestNoMisusedPromisesLazyContextExpirationAndWrappers(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedPromisesRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
type Row = { path: string; handle: () => void };
declare const path: string;
declare const syncHandle: () => void;
const factories: Array<() => Row> = [
  () => ({ path, handle: syncHandle }),
  () => { return { path, handle: syncHandle }; },
];
function makeRow(): Row {
  return { path, handle: syncHandle };
}
const constItems = [{ label: "Save" } as const, { label: "Close" } as const];
const coords = [[10, 20], [30, 40], [50, 60]];
const nested = [
  [{ a: 1 }, { b: 2 }],
  { c: async () => {}, d: async () => {} },
];
const outerWithInnerArrayInMethod = [
  {
    m() {
      const inner = [{ a: 1 }, { b: 2 }];
      return inner;
    },
    label: "Save",
  },
  {
    m() {
      const inner = [{ a: 1 }, { b: 2 }];
      return inner;
    },
    label: "Close",
  },
];
`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
type Row = { path: string; handle: () => void };
declare const path: string;
declare const asyncHandle: () => Promise<void>;
const factories: Array<() => Row> = [
  () => ({ path, handle: asyncHandle }),
  () => { return { path, handle: asyncHandle }; },
];
function makeRow(): Row {
  return { path, handle: asyncHandle };
}
const mixed = [
  { a: 1 },
  { path, handle: asyncHandle } as Row,
  ({ path, handle: asyncHandle } as const) as Row,
  { b: 2 } as const,
];
const typedConst: Array<{ readonly handle: () => void }> = [
  { handle: () => {} } as const,
  { handle: asyncHandle } as const,
];
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 6, Column: 26},
					{MessageId: "voidReturnProperty", Line: 7, Column: 34},
					{MessageId: "voidReturnProperty", Line: 10, Column: 26},
					{MessageId: "voidReturnProperty", Line: 14, Column: 19},
					{MessageId: "voidReturnProperty", Line: 15, Column: 20},
					{MessageId: "voidReturnProperty", Line: 20, Column: 13},
				},
			},
			{
				Code: `
type Outer = { nested: { a: number; b: number }; cb: () => void };
declare const asyncHandle: () => Promise<void>;
const outer: Outer = {
  nested: { a: 1, b: 2 },
  cb: asyncHandle,
};
const arr: Array<{ cb: () => void } | Array<{ a: number; b: number }>> = [
  [{ a: 1, b: 2 }, { a: 3, b: 4 }],
  { cb: asyncHandle },
];
const typedOuterWithInnerArray: Array<{ m(): Array<{ a: number }>; cb: () => void }> = [
  {
    m() {
      const inner = [{ a: 1 }, { a: 2 }];
      return inner;
    },
    cb: asyncHandle,
  },
  {
    m() {
      const inner = [{ a: 3 }, { a: 4 }];
      return inner;
    },
    cb: asyncHandle,
  },
];
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "voidReturnProperty", Line: 6, Column: 7},
					{MessageId: "voidReturnProperty", Line: 10, Column: 9},
					{MessageId: "voidReturnProperty", Line: 18, Column: 9},
					{MessageId: "voidReturnProperty", Line: 25, Column: 9},
				},
			},
		},
	)
}
