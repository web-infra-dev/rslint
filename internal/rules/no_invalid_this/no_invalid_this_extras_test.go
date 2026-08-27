// TestNoInvalidThisExtras locks in branches and edge shapes that the
// upstream ESLint core test suite doesn't exercise. Each case carries an
// inline comment pointing at the specific branch / Dimension 4 row / tsgo
// AST quirk it covers, so future refactors can't silently regress them
// without breaking a named lock-in.
package no_invalid_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidThisExtras(t *testing.T) {
	unexpected := func(line, col int) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{
			{MessageId: "unexpectedThis", Line: line, Column: col},
		}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidThisRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: multi-level parenthesized receiver on .call/.bind/.apply ----
			{Code: `
export {};
((function () {
  this;
})).call(obj);
    `},

			// ---- Dimension 4: ElementAccessExpression form of .call/.bind/.apply (not just dotted) ----
			// Locks in the `KindElementAccessExpression` arm of isDefaultThisBinding,
			// which no upstream test reaches (upstream only uses dotted access).
			{Code: `
export {};
(function () {
  this;
})['call'](obj);
    `},
			{Code: "export {};\n(function () {\n  this;\n})[`bind`](obj);\n"},

			// ---- Dimension 4: ElementAccessExpression form of array-thisArg methods ----
			// Locks in the ElementAccessExpression arm of isMethodWhichHasThisArg.
			{Code: `
export {};
foo['forEach'](function () {
  this;
}, obj);
    `},

			// ---- Dimension 4: generator function / async function / async generator ----
			// hasThisParameter / isStrictFunction / isDefaultThisBinding must not
			// care about the async/generator modifiers.
			{Code: `
export {};
(function* () {
  this;
}).call(obj);
    `},
			{Code: `
export {};
(async function () {
  this;
}).call(obj);
    `},
			{Code: `
export {};
(async function* () {
  this;
}).call(obj);
    `},
			{Code: `
export {};
function Foo(this: T) {
  this;
}
    `},

			// ---- Dimension 4: class expression (not just class declaration) ----
			{Code: `
export {};
const C = class {
  foo() {
    this;
  }
};
    `},
			{Code: `
export {};
const C = class Named {
  static foo() {
    this;
  }
};
    `},

			// ---- Dimension 4: PrivateIdentifier method (not just private field) ----
			{Code: `
export {};
class A {
  #foo() {
    this;
  }
}
    `},
			{Code: `
export {};
class A {
  get #foo() {
    return this;
  }
  set #foo(v) {
    this;
  }
}
    `},

			// ---- Dimension 4: graceful degradation — body-absent forms must not crash ----
			// Overload signatures, abstract methods, and ambient declarations have no
			// body — hasThisParameter/isStrictFunction must tolerate a nil Body().
			{Code: `
export {};
declare function foo(x: number): void;
    `},
			{Code: `
export {};
abstract class A {
  abstract foo(): void;
}
    `},
			{Code: `
export {};
function overloaded(x: number): void;
function overloaded(x: string): void;
function overloaded(x: unknown): void {
  console.log(x);
}
    `},
			{Code: `
export {};
class A {
  declare foo: string;
}
    `},

			// ---- Dimension 4: graceful degradation — spread/rest must not crash or mask siblings ----
			{Code: `
export {};
var obj = {
  ...spread,
  foo: function () {
    this;
  },
};
    `},
			{Code: `
export {};
function foo({ ...rest }) {
  console.log(rest);
}
    `},
			{Code: `
export {};
var obj = { foo: function () { this; }.bind(obj), ...rest };
    `},

			// ---- Dimension 4: empty bodies ----
			{Code: `
export {};
class A {
  foo() {}
}
    `},
			{Code: `
export {};
(function () {}).call(obj);
    `},

			// ---- Locks in isDefaultThisBinding ShorthandPropertyAssignment arm: uppercase name (ES5 constructor) ----
			// Not reached by any upstream test — only the array-destructuring
			// AssignmentExpression form (`[Foo = function(){}] = a`) is migrated
			// upstream; this is the OBJECT shorthand-default form.
			{Code: `
export {};
var { Foo = function () { this; } } = obj;
    `},

			// ---- Locks in isDefaultThisBinding KindBindingElement arm: declaration-context default, uppercase ----
			// Not reached by any upstream test — upstream's parameter-default case
			// uses simple Identifier parameters (KindParameter), never a
			// destructuring BindingElement default.
			{Code: `
export {};
var [Foo = function () { this; }] = a;
    `},
			{Code: `
export {};
function foo([Foo = function () { this; }]) {}
    `},

			// ---- Decorator on a method: `this` resolves to the enclosing (valid) scope, not the method's own frame ----
			// No upstream test exercises decorators (a TS-only / stage-3 syntax);
			// this locks in decoratorOfClassMemberAncestor's non-computed-key peek.
			{Code: `
export {};
function outer(this: Ctx) {
  class C {
    @deco(this)
    foo() {}
  }
}
    `},
			// ---- Decorator on a computed-key method: no peek needed (push already deferred) ----
			{Code: `
export {};
function outer(this: Ctx) {
  class C {
    @deco(this)
    [computedName]() {}
  }
}
    `},

			// ---- Decorator on a class field: same enclosing-scope resolution as a method decorator ----
			// A field's own frame is always valid, so only the peek makes this
			// case follow `outer`'s (valid, this-param) frame.
			{Code: `
export {};
function outer(this: Ctx) {
  class C {
    @deco(this)
    x = 1;
  }
}
    `},
			// ---- Decorator on an `accessor` field (tsgo folds AccessorProperty onto PropertyDeclaration) ----
			{Code: `
export {};
function outer(this: Ctx) {
  class C {
    @deco(this)
    accessor x = 1;
  }
}
    `},
			// ---- Decorator on a computed-key field: no peek needed (push already deferred) ----
			{Code: `
export {};
function outer(this: Ctx) {
  class C {
    @deco(this)
    [computedName] = 1;
  }
}
    `},
			// ---- Field initializer keeps its own always-valid frame ----
			// Contrast case: only the decorator position peeks past the field.
			{Code: `
export {};
function outer() {
  "use strict";
  class C {
    x = this;
  }
}
    `},

			// ---- Dimension 4: arrow function's OWN "use strict" directive does not affect strictness ----
			// Strictness is decided by the nearest non-arrow container (here, the
			// sloppy-mode outer function), never by an inner arrow's own directive —
			// arrows never push a frame or gate on their own body.
			{Code: `
function foo() {
  var f = () => {
    "use strict";
    console.log(this);
  };
}
    `, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},

			// ---- Real-user (eslint/eslint#14534): typed arrow class field, private/public modifiers ----
			// https://github.com/eslint/eslint/issues/14534
			{Code: `
export {};
class TestClass {
  private mText: string = "test";

  public Baz = (): void => {
    console.log(this.mText);
  }

  public Foo(): void {
    console.log(this.mText);
  }
}
    `},

			// ---- Real-user (eslint/eslint#13894): field initialized by a factory-method call, method returning an arrow ----
			// https://github.com/eslint/eslint/issues/13894
			{Code: `
export {};
class BufferedLog {
  private pendingLogs: Array<PendingLog> = [];

  debug = this.addLogFunction(LogLevel.Debug);

  private addLogFunction(level: LogLevel): (message: string) => void {
    return (message: string): void => {
      this.pendingLogs.push(new PendingLog(level, message));
    };
  }
}
    `},

			// ---- Source-type, JSDoc, callback, and field parity regressions ----
			{
				Code:            `export {}; this; function f(){ this; }`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            `@dec(function(){ this; }) class C {}`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            `function foo() { "use strict"; this; }`,
				FileName:        "legacy.js",
				TSConfig:        "tsconfig.allowJs.json",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3, SourceType: "script"},
			},
			{Code: `export {}; function f(){ /* @thisX */ function g(){ this; } }`},
			{Code: `export /* @this */ function foo(){ this; }`},
			{Code: `export default /* @this */ function foo(){ this; }`},
			{Code: "export {}; function f(){ /** \u00a0@this */ function g(){ this; } }"},
			{Code: "export {}; function f(){ /** \ufeff@this */ function g(){ this; } }"},
			{Code: `export {}; /** @this */ const x = [function(){ this; }];`},
			{Code: `export {}; /** @this */ const x = !function(){ this; };`},
			{Code: `export {}; /** @this */ const [x = function(){ this; }] = [];`},
			{Code: `export {}; Uint8Array.from([], function () { this; }, obj);`},
			{Code: `export {}; BigInt64Array['from']([], function () { this; }, obj);`},
			{Code: `class C { accessor x = function(){ this; } }`},
			{Code: `/** @this */ foo(function(){ this; } as any);`},
			{Code: `/** @this */ foo(function(){ this; } satisfies Function);`},
			{Code: `/** @this */ foo(function(){ this; }!);`},
			{Code: `const outer = /** @this */ function(){ return function(){ this; }; };`},
			{Code: `const o = { /** @this */ p: [function(){ this; }] };`},
			{Code: `const outer = /** @this */ () => function(){ this; };`},
			{Code: `const { /** @this */ p = [function(){ this; }] } = obj;`},
			{Code: `class C { /** @this */ x = [function(){ this; }]; }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: ElementAccessExpression .call with null receiver ----
			{
				Code: `
export {};
(function () {
  this;
})['call'](null);
    `,
				Errors: unexpected(4, 3),
			},

			// ---- Dimension 4: generator / async function without a this-binding call ----
			{
				Code: `
export {};
(function* () {
  this;
});
    `,
				Errors: unexpected(4, 3),
			},
			{
				Code: `
export {};
(async function () {
  this;
});
    `,
				Errors: unexpected(4, 3),
			},

			// ---- Dimension 4: class expression method returning a standalone function ----
			{
				Code: `
export {};
const C = class {
  foo() {
    return function () {
      this;
    };
  }
};
    `,
				Errors: unexpected(6, 7),
			},

			// ---- Dimension 4: PrivateIdentifier method, standalone (not bound) ----
			{
				Code: `
export {};
class A {
  #foo() {
    return function () {
      this;
    };
  }
}
    `,
				Errors: unexpected(6, 7),
			},

			// ---- Locks in Reflect.apply arg-count guard: too few args falls through to default-bound ----
			{
				Code: `
export {};
Reflect.apply(function () { this; }, obj);
    `,
				Errors: unexpected(3, 29),
			},

			// ---- Locks in isDefaultThisBinding ShorthandPropertyAssignment arm: lowercase name (not a constructor) ----
			{
				Code: `
export {};
var { func = function () { this; } } = obj;
    `,
				Errors: unexpected(3, 28),
			},

			// ---- Locks in isDefaultThisBinding KindBindingElement arm: declaration-context default, lowercase ----
			{
				Code: `
export {};
var [func = function () { this; }] = a;
    `,
				Errors: unexpected(3, 27),
			},
			{
				Code: `
export {};
function foo([func = function () { this; }]) {}
    `,
				Errors: unexpected(3, 36),
			},

			// ---- Decorator NOT on the method itself: `this` inside the method body is still the method's own (VALID) — contrast case ----
			// Confirms the decorator peek doesn't leak into the method body.
			{
				Code: `
export {};
function outer(this: Ctx) {
  class C {
    @deco(this)
    foo() {
      return function () {
        this;
      };
    }
  }
}
    `,
				Errors: unexpected(8, 9),
			},

			// ---- Decorator's `this` genuinely resolves to the enclosing scope, not the method's always-valid frame ----
			// `outer` here is a plain strict function (no this-param, lowercase
			// name) so its own `this` is default-bound/INVALID. If the peek
			// were broken and fell back to the method's own frame (always
			// valid), this would wrongly report as valid — this is the
			// discriminating half of the "Decorator on a method" valid case.
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(this)
    foo() {}
  }
}
    `,
				Errors: unexpected(6, 11),
			},
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(this)
    [computedName]() {}
  }
}
    `,
				Errors: unexpected(6, 11),
			},

			// ---- Decorator on a class field resolves to the enclosing INVALID scope ----
			// The discriminating half of the "Decorator on a class field" valid
			// case: without the peek the field's always-valid frame would
			// swallow the report.
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(this)
    x = 1;
  }
}
    `,
				Errors: unexpected(6, 11),
			},
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(this)
    accessor x = 1;
  }
}
    `,
				Errors: unexpected(6, 11),
			},
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(this)
    [computedName] = 1;
  }
}
    `,
				Errors: unexpected(6, 11),
			},

			// ---- Nested member boundary inside a decorator does not restore the decorated member's frame ----
			// The computed key of `I`'s member is evaluated while the
			// decorator runs, i.e. in `outer`'s scope — but `I`'s own member
			// has not pushed its frame yet (computed keys defer), so the walk
			// must pass straight through it and still peek past `C.foo`.
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(class I { [this]() {} })
    foo() {}
  }
}
    `,
				Errors: unexpected(6, 22),
			},
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(class I { [this] = 1 })
    foo() {}
  }
}
    `,
				Errors: unexpected(6, 22),
			},
			// Object-literal accessor key, same shape.
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco({ get [this]() {} })
    foo() {}
  }
}
    `,
				Errors: unexpected(6, 18),
			},

			// ---- Each nested decorator hop peeks past its own member ----
			// `this` sits two decorated members deep, so a single peek would
			// land on `I.bar`'s always-valid frame instead of `outer`'s.
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(class I { @deco2(this) bar() {} })
    foo() {}
  }
}
    `,
				Errors: unexpected(6, 28),
			},
			{
				Code: `
export {};
function outer() {
  "use strict";
  class C {
    @deco(class I { @deco2(class J { [this]() {} }) bar() {} })
    foo() {}
  }
}
    `,
				Errors: unexpected(6, 39),
			},

			// ---- A class decorator hosts no member frame of its own ----
			{
				Code: `
export {};
function outer() {
  "use strict";
  @deco(class I { [this]() {} })
  class C {}
}
    `,
				Errors: unexpected(5, 20),
			},

			// ---- Source-type, JSDoc, strictness, and decorator parity regressions ----
			{
				Code:            `this; function f(){ this; }`,
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedThis", Line: 1, Column: 1},
					{MessageId: "unexpectedThis", Line: 1, Column: 21},
				},
			},
			{Code: `export {}; /* @this */ const x = [function(){ this; }];`, Errors: unexpected(1, 47)},
			{Code: "export {}; /** @this */\n\nconst x = [function(){ this; }];", Errors: unexpected(3, 24)},
			{Code: `export {}; function outer() { class C { @dec((@d(this) class I {})) m(){} } }`, Errors: unexpected(1, 50)},
			{
				Code:            `export {}; function foo(){ this; }`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3},
				Errors:          unexpected(1, 28),
			},
			{
				Code:            `class C { m(){ function foo(){ this; } } }`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3},
				Errors:          unexpected(1, 32),
			},
			{Code: `enum E { X = (function(){ this; }, 1) }`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: unexpected(1, 27)},
			{Code: `class C { @dec(function(){ this; }) m(){} }`, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: unexpected(1, 28)},
			{Code: `/** @this */ function outer(){ return function(){ this; }; }`, Errors: unexpected(1, 51)},
			{Code: `foo(/** @this */ (function(){ this; } as any));`, Errors: unexpected(1, 31)},
			{Code: `foo(/** @this */ (function(){ this; } satisfies Function));`, Errors: unexpected(1, 31)},
			{Code: `type T = typeof this;`, Errors: unexpected(1, 17)},
			{Code: `type T = typeof this.x;`, Errors: unexpected(1, 17)},
			{Code: `class C { x: typeof this; }`, Errors: unexpected(1, 21)},
			{Code: `class C { accessor x: typeof this = 1; }`, Errors: unexpected(1, 30)},
			{
				Code:            `function f(x: number){ 'use strict'; this; }`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 3, SourceType: "script"},
				Errors:          unexpected(1, 38),
			},
		},
	)
}
