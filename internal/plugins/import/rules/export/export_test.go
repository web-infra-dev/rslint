package export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExportRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&export.ExportRule,
		[]rule_tester.ValidTestCase{
			// ---- Default exports and basic named exports (upstream Valid set) ----
			{Code: `import "./malformed.js"`},
			{Code: `var foo = "foo"; export default foo;`},
			{Code: `export var foo = "foo"; export var bar = "bar";`},
			{Code: `export var foo = "foo", bar = "bar";`},
			{Code: `export var { foo, bar } = object;`},
			{Code: `export var [ foo, bar ] = array;`},
			{Code: `let foo; export { foo, foo as bar }`},
			{Code: `let bar; export { bar }; export * from "./export-all"`},
			{Code: `export * from "./export-all"`},
			{Code: `export * from "./does-not-exist"`},
			{Code: `export default foo; export * from "./bar"`},
			{
				Code: `
import * as A from './named-export-collision/a';
import * as B from './named-export-collision/b';
export { A, B };
`,
			},
			{
				Code: `
export * as A from './named-export-collision/a';
export * as B from './named-export-collision/b';
`,
			},

			// ---- TS function overloads (default + impl) ----
			{
				Code: `
export default function foo(param: string): boolean;
export default function foo(param: string, param1: number): boolean;
export default function foo(param: string, param1?: number): boolean {
  return Boolean(param) && param1 !== undefined;
}
`,
			},
			{
				Code: `
export default function foo(param: string): boolean;
export default function foo(param: string, param1?: number): boolean {
  return Boolean(param) && param1 !== undefined;
}
`,
			},

			// ---- TS type / value name clash (different namespaces) ----
			{
				Code: `
export const Foo = 1;
export type Foo = number;
`,
			},
			{
				Code: `
export const Foo = 1;
export interface Foo {}
`,
			},

			// ---- TS named function overloads ----
			{
				Code: `
export function fff(a: string): void;
export function fff(a: number): void;
`,
			},
			{
				Code: `
export function fff(a: string): void;
export function fff(a: number): void;
export function fff(a: string | number): void {}
`,
			},

			// ---- Namespace scoping (each namespace is its own scope) ----
			{
				Code: `
export const Bar = 1;
export namespace Foo {
  export const Bar = 1;
}
`,
			},
			{
				Code: `
export type Bar = string;
export namespace Foo {
  export type Bar = string;
}
`,
			},
			{
				Code: `
export const Bar = 1;
export type Bar = string;
export namespace Foo {
  export const Bar = 1;
  export type Bar = string;
}
`,
			},
			{
				Code: `
export namespace Foo {
  export const Foo = 1;
  export namespace Bar {
    export const Foo = 2;
  }
  export namespace Baz {
    export const Foo = 3;
  }
}
`,
			},

			// ---- Namespace merging ----
			{
				Code: `
export class Foo {}
export namespace Foo {}
export namespace Foo {
  export class Bar {}
}
`,
			},
			{
				Code: `
export function Foo(): void;
export namespace Foo {}
`,
			},
			{
				Code: `
export function Foo(a: string): void;
export namespace Foo {}
`,
			},
			{
				Code: `
export function Foo(a: string): void;
export function Foo(a: number): void;
export namespace Foo {}
`,
			},
			{
				Code: `
export enum Foo {}
export namespace Foo {}
`,
			},

			// ---- Ambient modules (each `declare module "name"` is its own scope) ----
			{
				Code: `
declare module "a" {
  const Foo = 1;
  export { Foo as default };
}
declare module "b" {
  const Bar = 2;
  export { Bar as default };
}
`,
			},
			{
				Code: `
declare module "a" {
  const Foo = 1;
  export { Foo as default };
}
const Bar = 2;
export { Bar as default };
`,
			},

			// ---- Edge cases beyond upstream's suite ----
			// `export = X` (TS-only) is semantically distinct from `export default`.
			{Code: `export = 1; export default 2;`},
			// Empty named-export list — no entries, no spurious match.
			{Code: `export {};`},
			// `export {} from 'mod'` — empty re-export, also no entries.
			{Code: `export {} from "./export-all";`},
			// A non-exported namespace must not pollute the parent scope.
			{
				Code: `
export const Foo = 1;
namespace Foo {
  export const Bar = 2;
}
`,
			},
			// Specifier rename — different exported names don't collide.
			{Code: `let a, b; export { a as x, b as y };`},
			// `let foo; export { foo as default }; export default 1;` — different
			// kinds of default would collide; check the inverse here: a single
			// default-via-rename is fine on its own.
			{Code: `let foo; export { foo as default };`},
			// Generic function with overloads + namespace merging.
			{
				Code: `
export function Foo<T>(t: T): T;
export function Foo<T>(t: T, n: number): T;
export namespace Foo {
  export const tag = "value";
}
`,
			},
			// Async function and generator do not change export-name extraction.
			{Code: `export async function foo() {}; export function bar() {};`},
			{Code: `export function* gen1() {}; export function* gen2() {};`},
			// Abstract class + namespace merging is allowed (single abstract class).
			{
				Code: `
export abstract class Foo {
  abstract m(): void;
}
export namespace Foo {
  export const tag = "value";
}
`,
			},
			// Nested destructuring binding pattern — every leaf identifier is
			// extracted and the names don't accidentally collide.
			{Code: `let obj: any; export const { a: { b, c }, d: [e, ...rest] } = obj;`},
			// `export * from` chain (re-export-chain → export-all) — names from
			// the deepest module surface up; the rule walks the chain and the
			// chained module's `downstream` doesn't collide with locals.
			{Code: `export const local = 1; export * from "./re-export-chain";`},
			// `export * from` cycle — visited-set short-circuit prevents infinite
			// recursion; both files contribute their unique exports.
			{Code: `export * from "./re-export-cycle-a";`},
			// Type alias `Foo` (type bucket) and `export type { Foo } from`
			// (value bucket) live in different buckets per upstream's listener:
			// only TSTypeAliasDeclaration / TSInterfaceDeclaration get the
			// `type:` prefix; type-only re-exports keep the plain name.
			{
				Code: `
export type Foo = string;
export type { Foo } from "./export-all";
`,
				FileName: "main.ts",
			},
			// Same logic for interface vs `export type { Foo }`.
			{
				Code: `
export interface Foo {}
export type * as Foo from "./export-all";
`,
				FileName: "main.ts",
			},

			// ---- Anonymous default declarations ----
			{Code: `let x = 1; export default function() {}`},
			{Code: `let x = 1; export default class {}`},
			{Code: `let x = 1; export default function*() {}`},
			{Code: `let x = 1; export default async function() {}`},
			{Code: `let x = 1; export default async function*() {}`},
			{Code: `let x = 1; export default class extends Object {}`},

			// ---- `export { default as X } from 'mod'` re-exports the upstream
			//      default under a new value name, so it does NOT live in the
			//      "default" bucket — coexists fine with `export default 1`.
			{Code: `export default 1; export { default as Foo } from "./export-all";`, FileName: "main.ts"},

			// ---- TS `declare` function overloads ----
			{
				Code: `
export declare function foo(a: string): void;
export declare function foo(a: number): void;
export function foo(a: string | number): void {}
`,
			},
			// `declare class` is a real class declaration, not an overload —
			// pairs with namespace via merging exemption (single class).
			{
				Code: `
export declare class Foo {
  m(): void;
}
export namespace Foo {}
`,
			},

			// ---- `const enum` is parsed as EnumDeclaration; should merge with
			//      namespace just like a regular enum.
			{
				Code: `
export const enum Foo {
  A = 1,
}
export namespace Foo {}
`,
			},

			// (Same module re-exported twice is invalid — upstream's
			// ExportAllDeclaration listener calls `addNamed(name, node)` on
			// every occurrence, so each name appears in `Set<node>` once per
			// `export *` statement and the duplicate fires. See the matching
			// invalid case below.)

			// ---- export-all from multiple distinct modules with no name overlap.
			{
				Code: `
export * from "./export-all";
export * from "./named-export-collision/a";
`,
				FileName: "main.ts",
			},

			// ---- Deep namespace nesting (4 levels), each a separate scope.
			{
				Code: `
export namespace L1 {
  export namespace L2 {
    export namespace L3 {
      export namespace L4 {
        export const x = 1;
      }
      export const x = 2;
    }
    export const x = 3;
  }
  export const x = 4;
}
export const x = 5;
`,
			},

			// ---- Object rest / array rest in export const ----
			{Code: `let obj: any; export const { a, ...rest } = obj;`},
			{Code: `let arr: any[]; export const [first, ...others] = arr;`},

			// ---- `export type { T as U }` — exported name is `U`, not `T`.
			{
				Code: `
export const T = 1;
export type { T as U } from "./export-all";
`,
				FileName: "main.ts",
			},

			// ---- decorator + class — modifier flag still detected.
			{
				Code: `
declare const dec: any;
@dec
export class Foo {}
`,
			},

			// ---- Overload signature for class method does not affect
			//      module-level export tracking.
			{
				Code: `
export class Foo {
  m(): void;
  m(x: number): void;
  m(x?: number): void {}
}
`,
			},

			// ---- Imports do not register exports.
			{
				Code: `
import { foo } from "./export-all";
import * as star from "./export-all";
import def from "./export-all";
import "./export-all";
`,
				FileName: "main.ts",
			},

			// ---- Top-level `export {}` (empty named exports) is a TS marker
			//      to opt the file into module mode; never reports.
			{Code: `export {};`},
			{Code: `var foo = 1; export {};`},

			// ---- Re-export through a chain (a → b → c) with a unique local.
			{
				Code: `
export const localUnique = 1;
export * from "./re-export-chain";
`,
				FileName: "main.ts",
			},

			// ---- `export * as X from 'mod'` does NOT enter the duplicate map
			//      (upstream's ExportAllDeclaration listener returns early when
			//      `exported.name` is set). So neither of these is a duplicate.
			{
				Code: `
export * as A from "./named-export-collision/a";
export * as A from "./named-export-collision/b";
`,
				FileName: "main.ts",
			},
			{
				Code: `
export const Foo = 1;
export * as Foo from "./export-all";
`,
				FileName: "main.ts",
			},
			{
				Code: `
export const Foo = 1;
export type * as Foo from "./export-all";
`,
				FileName: "main.ts",
			},

			// ---- Default expression in various ES forms ----
			{Code: `export default () => 1;`},
			{Code: `export default { a: 1 };`},
			{Code: `let arr: any[]; export default [...arr];`},
			{Code: `class Base {} export default class extends Base {};`},
			{Code: `class Base {} export default class Foo extends Base {};`},
			{Code: `let x: any; export default x as any;`},
			{Code: `let x: any; export default x satisfies object;`},
			{Code: `let x: any; export default (x);`}, // parenthesized expression

			// ---- 3-bucket type mix WITHOUT collision (value+type buckets independent) ----
			// `interface Foo` is in `type:Foo`, `const Foo` is in `Foo` value bucket — neither bucket has size > 1.
			{
				Code: `
export const Foo = 1;
export interface Foo {}
`,
			},
			// type alias + interface in same `type:` bucket would collide; they
			// must be put in different *names* to stay valid.
			{
				Code: `
export type T1 = string;
export interface T2 {}
`,
			},

			// ---- type-only specifier list, multiple names ----
			{
				Code: `
export type { foo, baz } from "./export-all";
`,
				FileName: "main.ts",
			},
			// re-export under aliases — none collide since exported names differ.
			{
				Code: `
export type { foo as A, baz as B } from "./export-all";
`,
				FileName: "main.ts",
			},

			// ---- Empty source file ----
			{Code: ``},

			// ---- Only imports, no exports ----
			{
				Code: `
import { foo } from "./export-all";
`,
				FileName: "main.ts",
			},

			// ---- `export {}` then real exports — empty list is a no-op marker. ----
			{
				Code: `
export {};
export const x = 1;
export const y = 2;
`,
			},

			// ---- Arrow default + standalone named (no conflict) ----
			{Code: `export default () => 1; export const helper = () => 2;`},

			// ---- abstract anonymous default + non-default valid ----
			{Code: `export default abstract class { abstract m(): void };`},

			// ---- generic abstract class + namespace merging ----
			{
				Code: `
export abstract class Foo<T> {
  abstract m(t: T): T;
}
export namespace Foo {
  export const tag = 1;
}
`,
			},

			// ---- `declare class` overload + namespace merging ----
			{
				Code: `
export declare class Foo {
  m(a: string): void;
  m(a: number): void;
}
export namespace Foo {}
`,
			},

			// ---- Symbol-named computed key destructuring binding identifier
			//      still extracted (the value name `v` is what binds, not the key).
			{Code: `let obj: any; let k: any; export const { [k]: v } = obj;`},

			// ---- Two ambient modules with the same module name string ----
			// Each `declare module 'foo'` is its own scope; no cross-module
			// duplicate even when both declare a `default` export.
			{
				Code: `
declare module "foo" {
  const X = 1;
  export { X as default };
}
declare module "foo" {
  const Y = 2;
  export { Y as default };
}
`,
			},

			// ---- Two ambient modules; one has a duplicate, the other doesn't.
			// Locks in scope isolation: only the inner duplicate is detected.
			{
				Code: `
declare module "a" {
  export const X = 1;
}
declare module "b" {
  export const X = 1;
}
`,
			},

			// ---- declare global { ... } isolates exports just like ambient modules.
			{
				Code: `
declare global {
  interface Window {
    foo: number;
  }
}
export const Window = 1;
`,
			},

			// ---- Nested destructuring with default values — every leaf id is
			//      extracted, defaults don't change the binding name.
			{Code: `let obj: any; export const { a = 1, b: { c = 2 } = {} } = obj;`},

			// ---- Inline `type` modifier inside a specifier list — the spec
			//      still lives in the value bucket, never type bucket.
			{
				Code: `
export const Foo = 1;
export type Foo = string;
export { type Bar } from "./export-all";
`,
				FileName: "main.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Multiple defaults: overload + impl + re-export ----
			// Upstream's TS parser test: the body-less overload is stripped,
			// leaving the impl and the specifier as duplicates.
			{
				Code: `
export default function a(): void;
export default function a() {}
export { x as default };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 3, Column: 1},
					{MessageId: "multipleDefault", Line: 4, Column: 15},
				},
			},

			// ---- Type-twice clash ----
			{
				Code: `
export type Foo = string;
export type Foo = number;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 13},
					{MessageId: "multipleNamed", Line: 3, Column: 13},
				},
			},

			// ---- Interface-twice clash (upstream prefixes interface with type:) ----
			// Note: TS itself supports declaration-merging here; the rule's
			// stance is to surface this as a duplicate regardless.
			{
				Code: `
export interface Foo {}
export interface Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 18},
					{MessageId: "multipleNamed", Line: 3, Column: 18},
				},
			},

			// ---- Type-alias collides with interface (both type-bucketed) ----
			{
				Code: `
export type Foo = string;
export interface Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 13},
					{MessageId: "multipleNamed", Line: 3, Column: 18},
				},
			},

			// ---- Duplicates inside a namespace (separate scope) ----
			{
				Code: `
export const a = 1;
export namespace Foo {
  export const a = 2;
  export const a = 3;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 4, Column: 16},
					{MessageId: "multipleNamed", Line: 5, Column: 16},
				},
			},

			// ---- Multiple defaults inside an ambient module ----
			{
				Code: `
declare module "foo" {
  const Foo = 1;
  export default Foo;
  export default Foo;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 4, Column: 3},
					{MessageId: "multipleDefault", Line: 5, Column: 3},
				},
			},

			// ---- Nested namespaces with duplicate values per scope ----
			{
				Code: `
export namespace Foo {
  export namespace Bar {
    export const Foo = 1;
    export const Foo = 2;
  }
  export namespace Baz {
    export const Bar = 3;
    export const Bar = 4;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 4, Column: 18},
					{MessageId: "multipleNamed", Line: 5, Column: 18},
					{MessageId: "multipleNamed", Line: 8, Column: 18},
					{MessageId: "multipleNamed", Line: 9, Column: 18},
				},
			},

			// ---- 2× class + namespace: namespace is silenced, classes both report ----
			{
				Code: `
export class Foo {}
export class Foo {}
export namespace Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 14},
				},
			},

			// ---- 2× enum + namespace ----
			{
				Code: `
export enum Foo {}
export enum Foo {}
export namespace Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 13},
					{MessageId: "multipleNamed", Line: 3, Column: 13},
				},
			},

			// ---- enum + class + namespace ----
			{
				Code: `
export enum Foo {}
export class Foo {}
export namespace Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 13},
					{MessageId: "multipleNamed", Line: 3, Column: 14},
				},
			},

			// ---- const + class + namespace ----
			{
				Code: `
export const Foo = 'bar';
export class Foo {}
export namespace Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 14},
				},
			},

			// ---- function + class + namespace ----
			{
				Code: `
export function Foo() {}
export class Foo {}
export namespace Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 17},
					{MessageId: "multipleNamed", Line: 3, Column: 14},
				},
			},

			// ---- const + function + namespace ----
			{
				Code: `
export const Foo = 'bar';
export function Foo() {}
export namespace Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 17},
				},
			},

			// ---- const + namespace (no merging exemption) ----
			{
				Code: `
export const Foo = 'bar';
export namespace Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 18},
				},
			},

			// ---- Multiple top-level defaults; ambient module is its own scope ----
			{
				Code: `
declare module "a" {
  const Foo = 1;
  export { Foo as default };
}
const Bar = 2;
export { Bar as default };
const Baz = 3;
export { Baz as default };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 7, Column: 17},
					{MessageId: "multipleDefault", Line: 9, Column: 17},
				},
			},

			// ---- Re-exporting the same local under the same exported name twice ----
			{
				Code: `
let foo;
export { foo };
export { foo };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 10},
					{MessageId: "multipleNamed", Line: 4, Column: 10},
				},
			},

			// ---- Arbitrary module namespace identifier collides with identifier ----
			// Locks in: the rule looks up the exported NAME (not the local) so
			// `export { foo as "foo" }` collides with `export { foo }`.
			{
				Code: `
let foo;
export { foo };
export { foo as "foo" };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 10},
					{MessageId: "multipleNamed", Line: 4, Column: 17},
				},
			},

			// ---- Default expression + `export { x as default }` ----
			{
				Code: `
let x;
export default 1;
export { x as default };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 3, Column: 1},
					{MessageId: "multipleDefault", Line: 4, Column: 15},
				},
			},

			// ---- Two `export { foo as default }` specifiers ----
			{
				Code: `
let foo;
export { foo as default };
export { foo as default };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 3, Column: 17},
					{MessageId: "multipleDefault", Line: 4, Column: 17},
				},
			},

			// ---- export-all expansion: `export *` brings `foo` from the upstream
			//      module, colliding with the local `export { foo }` ----
			{
				Code:     `let foo; export { foo }; export * from "./export-all";`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 1, Column: 19},
					{MessageId: "multipleNamed", Line: 1, Column: 26},
				},
			},

			// ---- export-all expansion meets arbitrary string export name ----
			{
				Code:     `let foo; export { foo as "foo" }; export * from "./export-all";`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 1, Column: 26},
					{MessageId: "multipleNamed", Line: 1, Column: 35},
				},
			},

			// ---- export-all expansion: local `foo` declaration + `export *` ----
			{
				Code: `
export const foo = 1;
export * from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 1},
				},
			},

			// ---- "No named exports found" — upstream module exports only `default` ----
			{
				Code:     `export * from "./default-export-only";`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamedExports", Line: 1, Column: 15},
				},
			},

			// `export * as X` repeated against same identifier — upstream's
			// ExportAllDeclaration listener returns early when `exported.name`
			// is set (neither expands nor adds `X`), so duplicates of the
			// namespace-export name are not surfaced. Captured under valid in
			// the upper section.

			// ---- type-only re-export specifier duplicate ----
			// Per upstream, ExportSpecifier always lives in the value bucket
			// (no `type:` prefix), so two type-only re-exports of the same
			// name collide.
			{
				Code: `
export type { foo } from "./export-all";
export type { foo } from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 15},
					{MessageId: "multipleNamed", Line: 3, Column: 15},
				},
			},

			// ---- value + type-only re-export of the same name ----
			// `export const Foo` (value bucket) collides with
			// `export type { Foo } from 'mod'` because upstream's
			// ExportSpecifier listener does NOT prefix with `type:`.
			{
				Code: `
export const Foo = 1;
export type { Foo } from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 15},
				},
			},

			// `export const Foo + export * as Foo from` — upstream's
			// ExportAllDeclaration listener bails out early when
			// `exported.name` is set, so the namespace-export name does not
			// participate in the value bucket and no duplicate is reported.
			// Captured under valid in the upper section.

			// ---- Nested binding pattern duplicate ----
			// Locks in: each leaf identifier of the destructuring becomes its own
			// export entry; a later `export const b` collides with the inner `b`.
			{
				Code: `
let obj: any;
export const { a: { b }, c: [d] } = obj;
export const b = 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 21},
					{MessageId: "multipleNamed", Line: 4, Column: 14},
				},
			},

			// ---- Parse-error propagation: upstream module has a syntax error,
			//      `export * from` reports it at the source-string position.
			{
				Code:     `export * from "./malformed";`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "parseErrors",
						Message:   "Parse errors in imported module './malformed': A 'return' statement can only be used within a function body. (1:1)",
						Line:      1,
						Column:    15,
					},
				},
			},

			// ---- ns + class + function: NOT a merging combination
			//      (upstream allows ns+class OR ns+function but not both).
			{
				Code: `
export class Foo {}
export function Foo() {}
export namespace Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 17},
				},
			},

			// ---- Anonymous default class + named default re-export ----
			{
				Code: `
let x;
export default class {}
export { x as default };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 3, Column: 1},
					{MessageId: "multipleDefault", Line: 4, Column: 15},
				},
			},

			// ---- Anonymous default function + default expression ----
			{
				Code: `
export default function() {}
export default 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 2, Column: 1},
					{MessageId: "multipleDefault", Line: 3, Column: 1},
				},
			},

			// ---- Async generator default + default expression ----
			{
				Code: `
export default async function*() {}
export default 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 2, Column: 1},
					{MessageId: "multipleDefault", Line: 3, Column: 1},
				},
			},

			// ---- Two `export * from` of distinct modules with the same name ----
			{
				Code: `
export * from "./export-all";
export * from "./re-export-chain";
`,
				FileName: "main.ts",
				// `./export-all` exposes `foo`, `baz`; `./re-export-chain`
				// re-exports `./export-all` plus its own `downstream` — so
				// the chain pulls in `foo` and `baz` again, plus `downstream`.
				// Each `foo` / `baz` collision yields 2 reports.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 1},
					{MessageId: "multipleNamed", Line: 3, Column: 1},
					{MessageId: "multipleNamed", Line: 2, Column: 1},
					{MessageId: "multipleNamed", Line: 3, Column: 1},
				},
			},

			// ---- Deep BindingPattern + later collision ----
			{
				Code: `
let obj: any;
export const { a: { b: { c } } } = obj;
export const c = 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 26},
					{MessageId: "multipleNamed", Line: 4, Column: 14},
				},
			},

			// ---- Array binding pattern + later collision ----
			{
				Code: `
let arr: any[];
export const [, [, [a]]] = arr;
export const a = 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 21},
					{MessageId: "multipleNamed", Line: 4, Column: 14},
				},
			},

			// ---- Mixed Type + value via 3 forms: type alias + interface + const ----
			//      `type:Foo` bucket has 2 entries (type alias, interface), and
			//      `Foo` value bucket has 1 entry (const). The const is alone
			//      so it doesn't report; only the two `type:Foo` collide.
			{
				Code: `
export const Foo = 1;
export type Foo = string;
export interface Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 13},
					{MessageId: "multipleNamed", Line: 4, Column: 18},
				},
			},

			// ---- Same module re-exported via `export *` AND named import-export ----
			{
				Code: `
export * from "./export-all";
export const foo = 1;
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 1},
					{MessageId: "multipleNamed", Line: 3, Column: 14},
				},
			},

			// ---- `export {default as X} from 'mod'` produces named X — but
			//      a sibling `export const X` makes it collide on the value
			//      side. Locks in: re-export under alias does NOT track `default`.
			{
				Code: `
export const Foo = 1;
export { default as Foo } from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 21},
				},
			},

			// ---- Multiple `export { foo as default }` from re-exports ----
			{
				Code: `
export { foo as default } from "./export-all";
export { baz as default } from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 2, Column: 17},
					{MessageId: "multipleDefault", Line: 3, Column: 17},
				},
			},

			// ---- Default class + default class via re-export ----
			{
				Code: `
export default class Foo {}
export { default } from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 2, Column: 1},
					{MessageId: "multipleDefault", Line: 3, Column: 10},
				},
			},

			// ---- Triple duplicate (no namespace involved) ----
			{
				Code: `
export const Foo = 1;
export const Foo = 2;
export const Foo = 3;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 14},
					{MessageId: "multipleNamed", Line: 4, Column: 14},
				},
			},

			// ---- Function overload + impl + extra impl (the second impl is a
			//      duplicate, not an overload, since it has a body too).
			{
				Code: `
export function fff(a: string): void;
export function fff(a: number): void {}
export function fff(a: string | number): void {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 17},
					{MessageId: "multipleNamed", Line: 4, Column: 17},
				},
			},

			// ---- Variable + type alias of same name in same scope; `export const`
			//      lives in value bucket and `export type` in `type:` bucket,
			//      so a sibling re-export specifier of the same name from a
			//      module collides only with the value bucket.
			{
				Code: `
export type T = string;
export const T = 1;
export { T } from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 14},
					{MessageId: "multipleNamed", Line: 4, Column: 10},
				},
			},

			// ---- Nested namespace with same-name across LEVELS does NOT collide,
			//      but at the same LEVEL it does. Locks in scope isolation.
			{
				Code: `
export namespace Outer {
  export const x = 1;
  export namespace Inner {
    export const x = 2;
  }
  export const x = 3;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 16},
					{MessageId: "multipleNamed", Line: 7, Column: 16},
				},
			},

			// ---- `module A.B { ... }` dotted-namespace form: the inner block's
			//      duplicates report normally.
			{
				Code: `
export namespace A.B {
  export const x = 1;
  export const x = 2;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 16},
					{MessageId: "multipleNamed", Line: 4, Column: 16},
				},
			},

			// ---- Same module re-exported twice via `export *` ----
			// Upstream's listener calls `addNamed(name, node)` per occurrence;
			// each name (foo, baz) ends up with 2 entries → both `export *`
			// statements report each name once. With 2 names that's 4 errors.
			{
				Code: `
export * from "./export-all";
export * from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 1},
					{MessageId: "multipleNamed", Line: 3, Column: 1},
					{MessageId: "multipleNamed", Line: 2, Column: 1},
					{MessageId: "multipleNamed", Line: 3, Column: 1},
				},
			},

			// ---- Arrow default + default expression ----
			{
				Code: `
export default () => 1;
export default 2;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 2, Column: 1},
					{MessageId: "multipleDefault", Line: 3, Column: 1},
				},
			},

			// ---- Object literal default + named default re-export ----
			{
				Code: `
let x: any;
export default { a: 1 };
export { x as default };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 3, Column: 1},
					{MessageId: "multipleDefault", Line: 4, Column: 15},
				},
			},

			// ---- Abstract anonymous default + plain default ----
			{
				Code: `
export default abstract class { abstract m(): void };
export default 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 2, Column: 1},
					{MessageId: "multipleDefault", Line: 3, Column: 1},
				},
			},

			// ---- String-literal exported names colliding with each other ----
			// `let a, b; export { a as 'foo' }; export { b as 'foo' };` — both
			// resolve to the value-bucket name 'foo' and collide.
			{
				Code: `
let a, b;
export { a as "foo" };
export { b as "foo" };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 15},
					{MessageId: "multipleNamed", Line: 4, Column: 15},
				},
			},

			// ---- 3 different forms in the same value bucket: const + class +
			//      function. The interface lives in the type bucket; only the 3
			//      value-bucket entries collide.
			{
				Code: `
export const Foo = 1;
export class Foo {}
export function Foo() {}
export interface Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 14},
					{MessageId: "multipleNamed", Line: 4, Column: 17},
				},
			},

			// ---- Renaming makes a re-export collide with a local declaration ----
			{
				Code: `
let bar;
export const Foo = 1;
export { bar as Foo };
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 14},
					{MessageId: "multipleNamed", Line: 4, Column: 17},
				},
			},

			// ---- `export *` triggers "No named exports" alongside another local export ----
			// The local `x` is unique → not reported; only the upstream-default
			// import yields the noNamedExports diagnostic.
			{
				Code: `
export const x = 1;
export * from "./default-export-only";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamedExports", Line: 3, Column: 15},
				},
			},

			// ---- Two type aliases in same `type:` bucket via different forms
			//      (type alias + interface) ----
			{
				Code: `
export type Foo = string;
export interface Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 13},
					{MessageId: "multipleNamed", Line: 3, Column: 18},
				},
			},

			// ---- Anonymous class default + named anonymous-class default expression ----
			// (covers anonymous-default-of-different-shape entering the same
			//  default bucket)
			{
				Code: `
export default class extends Object {};
export default 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Line: 2, Column: 1},
					{MessageId: "multipleDefault", Line: 3, Column: 1},
				},
			},

			// ---- `export type * from 'mod'` brings names into the *value*
			//      bucket (no `type:` prefix), so a local `export const Foo`
			//      collides with `Foo` reachable through the type-only re-export.
			//
			// `./export-all` exposes `foo` and `baz`. The local `foo` collides
			// with the `foo` from the type-only star. `baz` from the star is
			// alone (no local) and stays silent.
			{
				Code: `
export const foo = 1;
export type * from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 14},
					{MessageId: "multipleNamed", Line: 3, Column: 1},
				},
			},

			// ---- Multiple `export *` lines that each hit the upstream's
			//      "default-only" emptiness — each line gets its own
			//      noNamedExports diagnostic (no dedup).
			{
				Code: `
export * from "./default-export-only";
export * from "./default-export-only";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamedExports", Line: 2, Column: 15},
					{MessageId: "noNamedExports", Line: 3, Column: 15},
				},
			},

			// ---- Namespace internal scope: const + class with the same name
			//      is NOT a merging combination, so it reports inside the
			//      namespace block (separate from the outer scope).
			{
				Code: `
export namespace Outer {
  export const X = 1;
  export class X {}
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 16},
					{MessageId: "multipleNamed", Line: 4, Column: 16},
				},
			},

			// ---- Two `export { type X }` (inline type modifier) of the same
			//      name — both go to the value bucket and collide.
			{
				Code: `
export { type foo } from "./export-all";
export { type foo } from "./export-all";
`,
				FileName: "main.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 2, Column: 15},
					{MessageId: "multipleNamed", Line: 3, Column: 15},
				},
			},

			// ---- Same ambient-module string declared twice with an internal
			//      duplicate inside one of them — the duplicate inside scope `a`
			//      reports; scope `b` is independent and stays clean.
			{
				Code: `
declare module "a" {
  export const X = 1;
  export const X = 2;
}
declare module "b" {
  export const X = 3;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Line: 3, Column: 16},
					{MessageId: "multipleNamed", Line: 4, Column: 16},
				},
			},

		},
	)
}
