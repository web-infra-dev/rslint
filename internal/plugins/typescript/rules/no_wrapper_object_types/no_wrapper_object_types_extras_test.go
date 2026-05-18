// TestNoWrapperObjectTypesExtras locks in branches and edge shapes that the
// upstream test suite does not exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package no_wrapper_object_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoWrapperObjectTypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoWrapperObjectTypesRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: identifier text outside the banned set ----
		// Locks in upstream checkBannedTypes() arm: classNames.has(typeName)
		// is false. Any other identifier passes silently.
		{Code: `let value: NumberOrString;`},
		{Code: `let value: ObjectLike;`},
		{Code: `let value: NotABoolean;`},

		// ---- Dimension 4: qualified type name (TypeName.Kind !== Identifier) ----
		// Locks in upstream checkBannedTypes() arm: node.type === Identifier
		// is false for `A.Number`, so the qualified form is not reported.
		{Code: `
      namespace A { export type Number = 0 | 1; }
      let value: A.Number;
    `},
		{Code: `
      namespace Outer {
        export namespace Inner {
          export type Symbol = 'a' | 'b';
        }
      }
      let value: Outer.Inner.Symbol;
    `},

		// ---- Branch lock-in: IsReferenceToGlobalIdentifier() — user-source def in module scope ----
		// `export {}` forces module scope, so a local `type Number` no longer
		// collides with the global one. tsgo's checker then resolves the
		// reference to the user-source declaration, and the rule must stay
		// silent.
		{Code: `
      export {};
      type Number = 0 | 1;
      let value: Number;
    `},
		{Code: `
      export {};
      class String {}
      let value: String;
    `},
		{Code: `
      export {};
      interface Object { foo: string; }
      let value: Object;
    `},

		// ---- Branch lock-in: nested block scope shadow ----
		// Mirrors the upstream "scope-introducing brace block" pattern from
		// `no-unsafe-function-type` but for the wrapper-object names.
		{Code: `
      {
        type Number = 0 | 1;
        let value: Number;
      }
    `},
		{Code: `
      {
        class Boolean {}
        let value: Boolean;
      }
    `},

		// ---- Branch lock-in: same-file declaration merging via interface ----
		// A local `interface Object` in the same file declarations augment
		// the global. IsReferenceToGlobalIdentifier's same-file check
		// silences the rule.
		{Code: `
      export {};
      interface Number {
        customField: string;
      }
      let value: Number;
    `},

		// ---- Branch lock-in: typeof X is a TypeQuery, not a TypeReference ----
		// Upstream does not listen to TSTypeQuery. The inner `Number` is part
		// of an EntityName, never a TypeReference, so nothing fires.
		{Code: `let value: typeof Number;`},
		{Code: `let value: typeof String;`},
		{Code: `let value: typeof globalThis.Number;`},

		// ---- Branch lock-in: `class extends X` is OUT of scope ----
		// Upstream registers TSClassImplements + TSInterfaceHeritage, NOT a
		// class-extends listener. tsgo's KindExpressionWithTypeArguments is
		// shared with `class extends`, so the rule must filter on
		// HeritageClause.Token to avoid reporting here. (Upstream's own valid
		// suite includes `class MyClass extends Number {}` for the same
		// reason.)
		{Code: `class Foo extends Number {}`},
		{Code: `class Foo extends Boolean {}`},
		{Code: `const Cls = class extends Object {};`},

		// ---- Dimension 4: identifier in value position ----
		// `new Number(0)` / `Number.MAX_VALUE` / `Symbol.iterator` reference
		// the wrappers as values, not as type references. Neither listener
		// fires.
		{Code: `const wrapped = new Number(0);`},
		{Code: `const big = Number.MAX_VALUE;`},
		{Code: `const sym = Symbol.iterator;`},

		// ---- Dimension 4: nested local shadow inside an arrow body ----
		{Code: `
      const f = () => {
        type Number = (x: number) => number;
        const g: Number = x => x;
        return g;
      };
    `},

		// ---- Branch lock-in: namespace body shadow ----
		// Namespaces introduce their own scope; tsgo's checker resolves to
		// the local Number rather than the global one.
		{Code: `
      namespace Inner {
        type Number = 0 | 1;
        export let g: Number;
      }
    `},

		// ---- Branch lock-in: same-file global augmentation ----
		// `declare global { interface Number {...} }` in the SAME file as the
		// reference is a user def in the current file's scope manager —
		// upstream silences the rule.
		{Code: `
      declare global {
        interface Number {
          customField: string;
        }
      }
      let value: Number;
      export {};
    `},

		// ---- Real-user: wrapper in JSDoc-style type-arg position via TypeQuery ----
		// `typeof v` where v: { Number: any } — inner identifier is a member
		// access on a TypeQuery, not a TypeReference.
		{Code: `
      declare const lib: { Number: number };
      let value: typeof lib.Number;
    `},

		// ---- Branch lock-in: type parameter named after a banned wrapper ----
		// `function f<Number>(x: Number)` — the inner Number resolves via
		// GetSymbolAtLocation to the type parameter, whose declaration lives
		// in the current source file. Verified across function / type alias /
		// class / interface containers so each type-parameter scope is
		// exercised.
		{Code: `function f<Number>(x: Number): Number { return x; }`},
		{Code: `type Box<Number> = { value: Number };`},
		{Code: `class C<Number> { value!: Number; }`},
		{Code: `interface I<Number> { value: Number }`},

		// ---- Branch lock-in: `infer X` introduces a fresh binding inside
		// the conditional's extends arm. Mirrors the upstream valid pattern
		// (`infer Function`) but with banned wrapper names. tsgo resolves
		// each reference to its infer-bound type parameter, so the rule
		// stays silent.
		{Code: `type Unwrap<T> = T extends Promise<infer Number> ? Number : never;`},
		{Code: `type Unwrap<T> = T extends infer String ? String : never;`},

		// ---- Branch lock-in: type parameter constraint mentions banned name
		// but resolves to the constraint's local declaration ----
		// `<Number extends number>` — the type parameter's constraint type
		// `number` is fine; the inner reference resolves to the type
		// parameter, not the lib `Number` interface.
		{Code: `function pick<Number extends number>(x: Number): Number { return x; }`},

		// ---- Branch lock-in: import-equals shadow ----
		// `import Number = require(...)` introduces a value/namespace
		// binding. utils.IsShadowed walks SourceFile statements and
		// recognizes KindImportEqualsDeclaration; the reference resolves
		// locally.
		{Code: `
      import Number = require('node:util');
      let value: typeof Number;
    `},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: parenthesized type wrapper ----
		// tsgo preserves KindParenthesizedType around the inner TypeReference.
		// The listener fires on the inner TypeReference regardless.
		{
			Code:   `let value: (Number);`,
			Output: []string{`let value: (number);`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 13},
			},
		},
		{
			Code:   `let value: ((String));`,
			Output: []string{`let value: ((string));`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 14},
			},
		},

		// ---- Dimension 4: array / tuple / generic wrappers ----
		{
			Code:   `let value: Number[];`,
			Output: []string{`let value: number[];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 12},
			},
		},
		{
			Code:   `let value: Array<Boolean>;`,
			Output: []string{`let value: Array<boolean>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 18},
			},
		},
		{
			Code:   `let value: ReadonlyArray<Symbol>;`,
			Output: []string{`let value: ReadonlyArray<symbol>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 26},
			},
		},

		// ---- Dimension 4: parameter / return / property type positions ----
		{
			Code:   `function f(x: Number) {}`,
			Output: []string{`function f(x: number) {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 15},
			},
		},
		{
			Code:   `function f(): String { return ''; }`,
			Output: []string{`function f(): string { return ''; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 15},
			},
		},
		{
			Code: `
        interface Holder {
          v: Object;
        }
      `,
			Output: []string{`
        interface Holder {
          v: object;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 3, Column: 14},
			},
		},

		// ---- Branch lock-in: class declaration vs class expression heritage ----
		// Locks in: the heritage-clause filter recognizes both class forms.
		{
			Code: `const Cls = class implements Object {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 30},
			},
		},

		// ---- Branch lock-in: extends + implements with banned name in implements only ----
		// `class extends Bar implements Number` — only the implements clause
		// reports; the (legitimate) extends clause does not.
		{
			Code: `
        class Bar {}
        class Foo extends Bar implements Number {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 3, Column: 42},
			},
		},

		// ---- Branch lock-in: multiple implements, banned name is one of them ----
		{
			Code: `
        interface Other {}
        class Foo implements Other, Number {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 3, Column: 37},
			},
		},

		// ---- Branch lock-in: implements lists every banned name in order ----
		// Each ExpressionWithTypeArguments is its own listener visit; the
		// rule reports every match independently and preserves source order.
		{
			Code: `class Foo implements BigInt, Boolean, Number, Object, String, Symbol {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 22},
				{MessageId: "bannedClassType", Line: 1, Column: 30},
				{MessageId: "bannedClassType", Line: 1, Column: 39},
				{MessageId: "bannedClassType", Line: 1, Column: 47},
				{MessageId: "bannedClassType", Line: 1, Column: 55},
				{MessageId: "bannedClassType", Line: 1, Column: 63},
			},
		},

		// ---- Dimension 4: same-kind nesting (interface in interface) ----
		// Both outer-uses of Number must be reported independently — the
		// listener visits each ExpressionWithTypeArguments separately.
		{
			Code: `
        interface A extends Number {}
        interface B extends Number {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 2, Column: 29},
				{MessageId: "bannedClassType", Line: 3, Column: 29},
			},
		},

		// ---- Branch lock-in: nested type position inside generic ----
		{
			Code:   `let value: Promise<Number>;`,
			Output: []string{`let value: Promise<number>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 20},
			},
		},
		{
			Code:   `let value: Map<String, Number>;`,
			Output: []string{`let value: Map<string, number>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 16},
				{MessageId: "bannedClassType", Line: 1, Column: 24},
			},
		},

		// ---- Real-user: factory returning wrapper type ----
		{
			Code:   `type Factory = () => String;`,
			Output: []string{`type Factory = () => string;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 22},
			},
		},

		// ---- Real-user: wrapper as a generic constraint ----
		{
			Code:   `function id<T extends String>(x: T): T { return x; }`,
			Output: []string{`function id<T extends string>(x: T): T { return x; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 23},
			},
		},

		// ---- Dimension 4: type assertion / satisfies wrappers ----
		// Both put a wrapper in a TypeNode position whose immediate parent
		// is a value expression. The TypeReference must still be visited.
		{
			Code:   `const v = 'x' as String;`,
			Output: []string{`const v = 'x' as string;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 18},
			},
		},
		{
			Code:   `const v = 0 satisfies Number;`,
			Output: []string{`const v = 0 satisfies number;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 23},
			},
		},

		// ---- Branch lock-in: conditional type's extends clause ----
		// `T extends Number` puts Number in a TypeNode position inside the
		// conditional. Both check-type and extends-type sides are visited.
		{
			Code:   `type IsNum<T> = T extends Number ? true : false;`,
			Output: []string{`type IsNum<T> = T extends number ? true : false;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 27},
			},
		},

		// ---- Branch lock-in: keyof / mapped type ----
		{
			Code:   `type Keys = keyof Number;`,
			Output: []string{`type Keys = keyof number;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 19},
			},
		},
		{
			Code:   `type Rec = { [k: string]: Object };`,
			Output: []string{`type Rec = { [k: string]: object };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 27},
			},
		},

		// ---- Branch lock-in: type-parameter default ----
		{
			Code:   `type Box<T = Number> = { value: T };`,
			Output: []string{`type Box<T = number> = { value: T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 14},
			},
		},

		// ---- Branch lock-in: readonly array of wrapper ----
		{
			Code:   `let value: readonly String[];`,
			Output: []string{`let value: readonly string[];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 21},
			},
		},

		// ---- Branch lock-in: nested class members ----
		// All four positions are real user shapes worth locking in: field,
		// method param, method return, getter return, setter param.
		{
			Code: `
        class C {
          handler: Number;
          run(s: String): Boolean {
            return false as Boolean;
          }
          get sym(): Symbol {
            return Symbol() as Symbol;
          }
          set sym(v: Symbol) {}
        }
      `,
			Output: []string{`
        class C {
          handler: number;
          run(s: string): boolean {
            return false as boolean;
          }
          get sym(): symbol {
            return Symbol() as symbol;
          }
          set sym(v: symbol) {}
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 3, Column: 20},
				{MessageId: "bannedClassType", Line: 4, Column: 18},
				{MessageId: "bannedClassType", Line: 4, Column: 27},
				{MessageId: "bannedClassType", Line: 5, Column: 29},
				{MessageId: "bannedClassType", Line: 7, Column: 22},
				{MessageId: "bannedClassType", Line: 8, Column: 32},
				{MessageId: "bannedClassType", Line: 10, Column: 22},
			},
		},

		// ---- Branch lock-in: explicit type argument in a call / new expression ----
		{
			Code:   `declare function foo<T>(): T; foo<Number>();`,
			Output: []string{`declare function foo<T>(): T; foo<number>();`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 35},
			},
		},
		{
			Code:   `class Box<T> { value!: T } new Box<Object>();`,
			Output: []string{`class Box<T> { value!: T } new Box<object>();`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 36},
			},
		},

		// ---- Dimension 4: top-level type alias with wrapper as RHS ----
		// The simplest real-user shape — a direct type alias.
		{
			Code:   `type Bag = Object;`,
			Output: []string{`type Bag = object;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 12},
			},
		},

		// ---- Branch lock-in: type-predicate target ----
		// `x is Object` puts Object in the type-predicate's TypeNode slot.
		{
			Code:   `function isObj(x: unknown): x is Object { return typeof x === 'object'; }`,
			Output: []string{`function isObj(x: unknown): x is object { return typeof x === 'object'; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 34},
			},
		},

		// ---- Branch lock-in: BigInt and Symbol round-trip ----
		// The two wrappers not covered by upstream's `Number/Symbol` union
		// case get their own simple round-trip lock-ins.
		{
			Code:   `let value: BigInt[];`,
			Output: []string{`let value: bigint[];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 12},
			},
		},
		{
			Code:   `let value: Symbol | undefined;`,
			Output: []string{`let value: symbol | undefined;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 12},
			},
		},

		// ---- Branch lock-in: same banned identifier appears more than once ----
		// Each occurrence is an independent listener visit; both fix sites
		// are applied in one rule_tester pass.
		{
			Code:   `type Pair = [Number, Number];`,
			Output: []string{`type Pair = [number, number];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 14},
				{MessageId: "bannedClassType", Line: 1, Column: 22},
			},
		},

		// ---- Real-user: abstract class implementing a wrapper ----
		// Verifies the abstract-modifier on the class doesn't suppress the
		// implements visit.
		{
			Code: `abstract class A implements Object { abstract m(): void; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 29},
			},
		},

		// ---- Branch lock-in: message data has both preferred and typeName ----
		// Locks in the message text on at least one case so future refactors
		// can't drop the `data` fields.
		{
			Code:   `let value: Object;`,
			Output: []string{`let value: object;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Message:   "Prefer using the primitive `object` as a type name, rather than the upper-cased `Object`.",
					Line:      1,
					Column:    12,
				},
			},
		},

		// ---- Dimension 4: function-type literal / constructor-type literal ----
		// Parameter and return type slots inside a function-type-literal
		// expression are TypeReference visits. Same for `new (...) => T`.
		{
			Code:   `let fn: (x: Number) => String;`,
			Output: []string{`let fn: (x: number) => string;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 13},
				{MessageId: "bannedClassType", Line: 1, Column: 24},
			},
		},
		{
			Code:   `let ctor: new (x: Boolean) => Object;`,
			Output: []string{`let ctor: new (x: boolean) => object;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 19},
				{MessageId: "bannedClassType", Line: 1, Column: 31},
			},
		},

		// ---- Branch lock-in: interface method / index / construct signatures ----
		// Each parameter and return type inside an interface member is its
		// own TypeReference visit. tsgo represents these via separate
		// signature kinds, so we verify the listener fans out across them.
		{
			Code: `
        interface API {
          on(handler: Number): void;
          new (cb: Boolean): String;
          [key: string]: Object;
        }
      `,
			Output: []string{`
        interface API {
          on(handler: number): void;
          new (cb: boolean): string;
          [key: string]: object;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 3, Column: 23},
				{MessageId: "bannedClassType", Line: 4, Column: 20},
				{MessageId: "bannedClassType", Line: 4, Column: 30},
				{MessageId: "bannedClassType", Line: 5, Column: 26},
			},
		},

		// ---- Dimension 4: labeled tuple elements + rest tuple ----
		// `[a: Number, b: String]` — labeled tuple element wraps the
		// TypeReference in a NamedTupleMember. The inner listener still
		// fires because the TypeReference itself is unchanged.
		{
			Code:   `let pair: [first: Number, second: String];`,
			Output: []string{`let pair: [first: number, second: string];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 19},
				{MessageId: "bannedClassType", Line: 1, Column: 35},
			},
		},
		{
			Code:   `let rest: [String, ...Number[]];`,
			Output: []string{`let rest: [string, ...number[]];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 12},
				{MessageId: "bannedClassType", Line: 1, Column: 23},
			},
		},

		// ---- Branch lock-in: optional / rest / default-typed parameters ----
		{
			Code:   `function f(a?: Number, ...rest: String[]) { return a; }`,
			Output: []string{`function f(a?: number, ...rest: string[]) { return a; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 16},
				{MessageId: "bannedClassType", Line: 1, Column: 33},
			},
		},

		// ---- Branch lock-in: deep nesting through stdlib generics ----
		// Locks in that the listener keeps firing at every depth — each
		// nested TypeReference is its own visit, not gated by some outer
		// "already reported" cache.
		{
			Code:   `let value: Promise<Map<String, ReadonlyArray<Awaited<Number>>>>;`,
			Output: []string{`let value: Promise<Map<string, ReadonlyArray<Awaited<number>>>>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 24},
				{MessageId: "bannedClassType", Line: 1, Column: 54},
			},
		},

		// ---- Branch lock-in: heritage clauses with type arguments ----
		// `class X implements Foo<T>` keeps `Foo` on the ExpressionWithTypeArguments
		// listener; type-argument list does not change the expression-side
		// match. Mirrors upstream's listener wiring exactly.
		{
			Code: `class X implements Number<string> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 20},
			},
		},
		{
			Code: `interface X extends Number<string> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 21},
			},
		},

		// ---- Branch lock-in: interface multi-extends, mix of clean and banned ----
		// Each ExpressionWithTypeArguments in the extends list is visited
		// independently; only the banned ones report.
		{
			Code: `
        interface Other {}
        interface X extends Other, Number, String {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 3, Column: 36},
				{MessageId: "bannedClassType", Line: 3, Column: 44},
			},
		},

		// ---- Branch lock-in: declare global wrapping a TypeReference ----
		// `declare global { interface I { x: Number } }` — the inner
		// TypeReference is a nested TypeReference inside a ModuleBlock.
		// The walker reaches it and reports.
		{
			Code: `
        declare global {
          interface External {
            value: Number;
          }
        }
        export {};
      `,
			Output: []string{`
        declare global {
          interface External {
            value: number;
          }
        }
        export {};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 4, Column: 20},
			},
		},

		// ---- Branch lock-in: a banned name appears as the constraint type ----
		// `<T extends Number>` — the constraint is a TypeReference. The
		// type parameter T does not introduce a Number shadow here.
		{
			Code:   `function g<T extends Number>(x: T): T { return x; }`,
			Output: []string{`function g<T extends number>(x: T): T { return x; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 22},
			},
		},

		// ---- Branch lock-in: mapped type with banned name in value side ----
		{
			Code:   `type Mapped<T> = { [K in keyof T]: Number };`,
			Output: []string{`type Mapped<T> = { [K in keyof T]: number };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 36},
			},
		},

		// ---- Branch lock-in: template literal type with wrapper in args ----
		// `Capitalize<String>` — utility types use the same TypeReference
		// path, regardless of being lib-provided string-manipulating types.
		{
			Code:   `type Cap = Capitalize<String>;`,
			Output: []string{`type Cap = Capitalize<string>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 23},
			},
		},

		// ---- Branch lock-in: chained as / satisfies wrappers ----
		// `x as Number as unknown` — only the inner `Number` is a
		// TypeReference; the outer `unknown` is fine.
		{
			Code:   `const v = (0 as Number) as unknown;`,
			Output: []string{`const v = (0 as number) as unknown;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 17},
			},
		},

		// ---- Branch lock-in: object method shorthand return type ----
		// `{ m(): Number { return 0 } }` — methods on object literals
		// (not classes) put the return-type TypeReference in a slightly
		// different parent chain. Verified independently.
		{
			Code:   `const o = { m(): Number { return 0 as any; } };`,
			Output: []string{`const o = { m(): number { return 0 as any; } };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 1, Column: 18},
			},
		},

		// ---- Real-user: arrow function as class-field property ----
		// `prop = (x: Number) => x` — class-field arrow puts parameters
		// in a separate AST shape than a regular method. The TypeReference
		// fires from inside the arrow's parameter list.
		{
			Code: `
        class K {
          prop = (x: Number): String => '' + x;
        }
      `,
			Output: []string{`
        class K {
          prop = (x: number): string => '' + x;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 3, Column: 22},
				{MessageId: "bannedClassType", Line: 3, Column: 31},
			},
		},

		// ---- Real-user: outer shadow does NOT cover inner reference ----
		// In the outer scope `let v: Number` reports because no shadow is in
		// scope. The nested block declares a local `type Number` — only
		// references inside that block are silenced. This locks in that the
		// lexical walk doesn't wrongly silence sibling references.
		{
			Code: `
        let outer: Number;
        {
          type Number = 0 | 1;
          let inner: Number;
        }
      `,
			Output: []string{`
        let outer: number;
        {
          type Number = 0 | 1;
          let inner: Number;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedClassType", Line: 2, Column: 20},
			},
		},
	})
}
