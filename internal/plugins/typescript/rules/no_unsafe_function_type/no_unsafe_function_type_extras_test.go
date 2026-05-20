package no_unsafe_function_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnsafeFunctionTypeExtras covers Layers 2 (edge-shape augmentation +
// real-user shapes) and 3 (branch lock-ins for the upstream source). These are
// cases upstream's own test suite does not exercise but that the port must
// keep aligned.
func TestNoUnsafeFunctionTypeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeFunctionTypeRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: identifier-name branch — name !== "Function" ----
		// Locks in upstream checkBannedTypes() arm: node.name === 'Function'
		// is false. Any other identifier passes silently.
		{Code: `let value: Foo;`},
		{Code: `let value: FunctionLike;`},
		{Code: `let value: MyFunction;`},

		// ---- Dimension 4: qualified type names (TypeName.Kind !== Identifier) ----
		// Locks in upstream checkBannedTypes() arm: node.type === Identifier
		// is false for `A.Function`, so it is not reported.
		{Code: `
      namespace A { export type Function = () => void; }
      let value: A.Function;
    `},

		// ---- Branch lock-in: isReferenceToGlobalFunction() — user-source defs in module scope ----
		// `export {}` forces module scope, so a local `type Function` no
		// longer collides with the global one. tsgo's checker then resolves
		// the reference to the user-source declaration, and this branch must
		// not report.
		{Code: `
      export {};
      type Function = () => void;
      let value: Function;
    `},
		{Code: `
      export {};
      class Function {}
      let value: Function;
    `},

		// ---- Branch lock-in: nested block scope shadows global Function ----
		// Mirrors the upstream valid case but with `class` rather than
		// `type`, exercising the value-binding side of the resolution.
		{Code: `
      {
        class Function {}
        let value: Function;
      }
    `},

		// ---- Dimension 4: nested scoping — local `type Function` inside an arrow body ----
		{Code: `
      const f = () => {
        type Function = (x: number) => number;
        const g: Function = x => x;
        return g;
      };
    `},

		// ---- Branch lock-in: class extends Function is OUT of scope ----
		// Upstream registers TSClassImplements + TSInterfaceHeritage, NOT a
		// "class extends" listener. tsgo's KindExpressionWithTypeArguments is
		// the same node kind for class extends, so the rule must filter on
		// HeritageClause.Token to avoid reporting here.
		{Code: `class Foo extends Function {}`},
		{Code: `const Cls = class extends Function {};`},

		// ---- Real-user: typeof Function in a value position ----
		// `typeof Function` is a TypeQuery, not a TypeReference; the rule
		// should not match value-position identifier nodes.
		{Code: `let value: typeof Function;`},

		// ---- Real-user: ExpressionWithTypeArguments in non-heritage positions ----
		// `Function` appearing inside `new`/`super` (value position) must not
		// be reported by the rule's heritage-clause branch.
		{Code: `const fn = Function;`},

		// ---- Dimension 4: nested declarations — outer interface extends local Function ----
		// `interface Foo extends Function` is illegal when Function is a local
		// `type`, since you cannot extend a type alias. Test the legal
		// shadowing variant via interface merging.
		{Code: `
      interface Function { foo(): void }
      interface Bar extends Function {}
    `},

		// ---- Branch lock-in: shadow inside a namespace body ----
		// Namespaces introduce their own scope; tsgo's checker should resolve
		// to the local Function rather than the global one.
		{Code: `
      namespace Inner {
        type Function = (x: number) => number;
        export let g: Function;
      }
    `},

		// ---- Branch lock-in: shadow inside a function body, multi-level ----
		// Verifies the resolution walks the actual lexical scope chain rather
		// than only the immediate parent block.
		{Code: `
      function outer() {
        function inner() {
          type Function = () => boolean;
          const flag: Function = () => true;
          return flag();
        }
        return inner();
      }
    `},

		// ---- Dimension 4: qualified name on the right of typeof ----
		// `typeof X.Y` is still a TypeQuery — the inner identifier is part of
		// an EntityName, never a TypeReference, so nothing fires.
		{Code: `let value: typeof globalThis.Function;`},

		// ---- Real-user: Function appearing only inside a value expression ----
		// `Function.prototype` references Function in value position. No
		// TypeReference / ExpressionWithTypeArguments listener should match.
		{Code: `const proto = Function.prototype;`},

		// ---- Real-user: ambient module declaration containing a re-exported Function ----
		// The local declaration shadows the lib.d.ts Function inside the module
		// body. The reference inside resolves to the local declaration.
		{Code: `
      declare module 'my-mod' {
        type Function = (x: number) => number;
        const value: Function;
      }
    `},

		// ---- Branch lock-in: nested namespace shadow ----
		// Two-level namespace nesting; the innermost scope's `Function` wins.
		{Code: `
      namespace Outer {
        export namespace Inner {
          export type Function = () => void;
          export let v: Function;
        }
      }
    `},

		// ---- Branch lock-in: same-file global augmentation ----
		// `declare global { interface Function {...} }` in the SAME file as the
		// reference is a user def in the current file's scope manager —
		// upstream silences the rule. This locks in
		// isReferenceToGlobalFunction's "current source file" check: a
		// declaration whose source file matches ctx.SourceFile counts as a
		// user def regardless of being in `declare global`. (Cross-file global
		// augmentation has the opposite outcome — see invalid cases.)
		{Code: `
      declare global {
        interface Function {
          customField: string;
        }
      }
      let value: Function;
      export {};
    `},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: generic instantiation on Function type ----
		// `Function<T>` is still a TypeReference with TypeName == Identifier
		// "Function". Locks in: the rule must not require the absence of
		// TypeArguments to fire.
		{
			Code: `let value: Function<number>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 12},
			},
		},

		// ---- Dimension 4: parameter / return / property type positions ----
		{
			Code: `function f(x: Function) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 15},
			},
		},
		{
			Code: `function f(): Function { return () => {}; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 15},
			},
		},
		{
			Code: `
        interface Holder {
          fn: Function;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 3, Column: 15},
			},
		},

		// ---- Dimension 4: parenthesized type wrapper ----
		// tsgo preserves KindParenthesizedType around the inner TypeReference.
		// The listener fires on the inner TypeReference regardless.
		{
			Code: `let value: (Function);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 13},
			},
		},

		// ---- Dimension 4: intersection / tuple / array combinator wrappers ----
		{
			Code: `let value: Function & object;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 12},
			},
		},
		{
			Code: `let value: [Function];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 13},
			},
		},
		{
			Code: `let value: Array<Function>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 18},
			},
		},

		// ---- Dimension 4: class declaration vs class expression heritage ----
		// Locks in: the heritage-clause filter recognizes both class forms.
		{
			Code: `const Cls = class implements Function {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 30},
			},
		},

		// ---- Branch lock-in: extends + implements with Function in implements list ----
		{
			Code: `
        class Bar {}
        class Foo extends Bar implements Function {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 3, Column: 42},
			},
		},

		// ---- Branch lock-in: multiple implements, Function is one of them ----
		{
			Code: `
        interface Other {}
        class Foo implements Other, Function {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 3, Column: 37},
			},
		},

		// ---- Dimension 4: same-kind nesting (interface in interface) ----
		// Both outer-uses of Function must be reported independently — the
		// listener visits each `ExpressionWithTypeArguments` separately.
		{
			Code: `
        interface A extends Function {}
        interface B extends Function {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 2, Column: 29},
				{MessageId: "bannedFunctionType", Line: 3, Column: 29},
			},
		},

		// ---- Dimension 4: nested type position inside generic ----
		{
			Code: `let value: Promise<Function>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 20},
			},
		},

		// ---- Real-user: function signature returning Function ----
		{
			Code: `type Factory = () => Function;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 22},
			},
		},

		// ---- Real-user: Function as a type-parameter constraint ----
		{
			Code: `function call<T extends Function>(fn: T) { fn(); }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 25},
			},
		},

		// ---- Dimension 4: type assertion / satisfies wrappers ----
		// Both put `Function` in a TypeNode position whose immediate parent
		// is a value expression (`as` / `satisfies`). The TypeReference must
		// still be visited.
		{
			Code: `const v = (() => {}) as Function;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 25},
			},
		},
		{
			Code: `const v = (() => {}) satisfies Function;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 32},
			},
		},

		// ---- Branch lock-in: conditional type's extends clause ----
		// `T extends Function` puts Function in a TypeNode position inside the
		// conditional. Both the check-type and the extends-type are visited.
		{
			Code: `type IsFn<T> = T extends Function ? true : false;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 26},
			},
		},
		{
			Code: `type IsFn<T> = T extends Function ? true : false; type Y = IsFn<() => void>;`,
			// Sanity wrapper: the rule fires once, not per usage of the alias.
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 26},
			},
		},

		// ---- Branch lock-in: mapped type's keyof / property type ----
		{
			Code: `type Keys = keyof Function;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 19},
			},
		},
		{
			Code: `type Record1 = { [k: string]: Function };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 31},
			},
		},

		// ---- Branch lock-in: deeper generic nesting ----
		{
			Code: `type T = Record<string, Function>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 25},
			},
		},
		{
			Code: `type T = Map<string, ReadonlyArray<Function>>;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 36},
			},
		},

		// ---- Branch lock-in: type-parameter default ----
		{
			Code: `type Box<T = Function> = { value: T };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 14},
			},
		},

		// ---- Branch lock-in: readonly tuple / readonly array ----
		{
			Code: `let value: readonly Function[];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 21},
			},
		},

		// ---- Branch lock-in: nested method / constructor / index signatures ----
		// All three are body-absent function-like forms inside an interface.
		// Each Function reference is an independent TypeReference visit.
		{
			Code: `
        interface API {
          on(handler: Function): void;
          new (cb: Function): unknown;
          [key: string]: Function;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 3, Column: 23},
				{MessageId: "bannedFunctionType", Line: 4, Column: 20},
				{MessageId: "bannedFunctionType", Line: 5, Column: 26},
			},
		},

		// ---- Branch lock-in: type predicate target ----
		// `x is Function` puts Function in the type-predicate's TypeNode slot.
		{
			Code: `function isF(x: unknown): x is Function { return typeof x === 'function'; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 32},
			},
		},

		// ---- Branch lock-in: abstract class implementing Function ----
		// Verifies abstract-modifier on the class doesn't suppress the
		// implements visit.
		{
			Code: `abstract class A implements Function { abstract m(): void; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 29},
			},
		},

		// ---- Branch lock-in: deeply nested function body containing a type alias ----
		// Locks in that the listener still fires inside a top-level function
		// body; the rule isn't accidentally module-scoped.
		{
			Code: `
        function outer() {
          type T = Function;
          return null as unknown as T;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 3, Column: 20},
			},
		},

		// ---- Dimension 4: top-level type alias with Function as the RHS ----
		// The simplest real-user shape — a direct type alias — must report.
		{
			Code: `type FnAlias = Function;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 16},
			},
		},

		// ---- Dimension 4: class members (field / method / getter / setter) ----
		// These are distinct AST shapes from interface members but resolve to
		// the same `KindTypeReference` listener. All four positions are real
		// user shapes worth locking in.
		{
			Code: `
        class C {
          handler: Function;
          run(cb: Function): Function {
            return cb;
          }
          get fn(): Function {
            return () => {};
          }
          set fn(v: Function) {}
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 3, Column: 20},
				{MessageId: "bannedFunctionType", Line: 4, Column: 19},
				{MessageId: "bannedFunctionType", Line: 4, Column: 30},
				{MessageId: "bannedFunctionType", Line: 7, Column: 21},
				{MessageId: "bannedFunctionType", Line: 10, Column: 21},
			},
		},

		// ---- Branch lock-in: explicit type argument in a call expression ----
		// `foo<Function>()` puts Function in `CallExpression.TypeArguments`,
		// which is still a TypeReference visit.
		{
			Code: `declare function foo<T>(): T; foo<Function>();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 35},
			},
		},

		// ---- Branch lock-in: explicit type argument in a `new` expression ----
		{
			Code: `class Box<T> { value!: T } new Box<Function>();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedFunctionType", Line: 1, Column: 36},
			},
		},
	})
}
