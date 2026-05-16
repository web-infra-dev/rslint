// TestNoInvalidThisExtras locks in branches and edge shapes that the
// upstream typescript-eslint test suite doesn't exercise. Each case carries
// an inline comment pointing at the specific branch / Dimension 4 row /
// tsgo AST quirk it covers, so future refactors can't silently regress
// them without breaking a named lock-in.
package no_invalid_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
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
			// ---- Dimension 4: paren-wrapped receiver on .bind/.call/.apply ----
			// Parens are elided in ESTree so multi-level wrapping must not
			// break the `.call`/`.bind`/`.apply` recognition.
			{Code: `
((function () {
  this;
})).call(obj);
    `},
			// ---- Dimension 4: paren chain around obj.foo = (paren-wrapped) ----
			{Code: `
obj.foo = ((function () {
  this;
}));
    `},
			// ---- Dimension 4: element access for `.bind` ----
			// `obj['bind'](thisArg)` should be recognized just like `obj.bind(thisArg)`.
			{Code: `
(function () {
  this;
})['bind'](obj);
    `},
			// ---- Dimension 4: element access for `.call` ----
			{Code: `
(function () {
  this;
})['call'](obj);
    `},
			// ---- Dimension 4: element access for `.apply` ----
			{Code: `
(function () {
  this;
})['apply'](obj);
    `},
			// ---- Dimension 4: array method recognized through element access ----
			// `arr['forEach'](fn, thisArg)` mirrors `arr.forEach(fn, thisArg)`.
			{Code: `
foo['forEach'](function () {
  this;
}, obj);
    `},

			// ---- Dimension 4: string-literal property key on object literal ----
			{Code: `
var obj = {
  'foo': function () {
    this;
  },
};
    `},
			// ---- Dimension 4: numeric-literal property key on object literal ----
			{Code: `
var obj = {
  0: function () {
    this;
  },
};
    `},
			// ---- Dimension 4: computed property key on object literal ----
			{Code: `
var obj = {
  [key]: function () {
    this;
  },
};
    `},

			// ---- Dimension 4: class expression with method ----
			{Code: `
const X = class {
  foo() {
    this;
  }
};
    `},
			// ---- Dimension 4: class expression with class field arrow ----
			{Code: `
const X = class {
  foo = () => {
    this;
  };
};
    `},
			// ---- Dimension 4: class field function expression ----
			{Code: `
class A {
  foo = function () {
    this;
  };
}
    `},
			// ---- Dimension 4: class with computed method key ----
			{Code: `
class A {
  ['foo']() {
    this;
  }
}
    `},
			// ---- Dimension 4: computed-key method body still gets a valid frame ----
			// Even when the key is computed, `this` inside the body is the
			// instance — confirm the deferred push lands before the body is
			// visited and not too late.
			{Code: `
class A {
  [Symbol.iterator]() {
    this;
  }
}
    `},
			// ---- Dimension 4: computed-key class field body ----
			{Code: `
class A {
  [Symbol.iterator] = function () {
    this;
  };
}
    `},
			// ---- Dimension 4: object-literal method with computed key & body using `this` ----
			// Confirms the computed-key deferral works on object literals too,
			// where the method body's `this` is the object.
			{Code: `
var obj = {
  ['foo']() {
    this;
  },
};
    `},
			// ---- Wrapper-bug lock-in: `this` in computed key of class field is masked ----
			// Upstream wrapper pushes `valid=true` on `PropertyDefinition` /
			// `AccessorProperty` entry, which fires BEFORE the computed key
			// is visited. The wrapper's `ThisExpression` listener then
			// short-circuits and never delegates to baseRule, so the
			// otherwise-valid report at top-level is silently masked. rslint
			// reproduces this behavior verbatim for byte-level alignment with
			// `@typescript-eslint/no-invalid-this`. A separate Layer-3 test
			// (the method/accessor counterpart above) confirms that
			// non-field computed keys still report correctly.
			{Code: `
class A {
  [this.foo] = 1;
}
    `},
			{Code: `
class A {
  accessor [this.foo] = 1;
}
    `},
			// ---- Wrapper-bug lock-in: `this` in decorator on a field is masked ----
			// PropertyDefinition / AccessorProperty push happens on entry,
			// so decorators on these (visited after entry) see the field's
			// frame and the report is silently swallowed. Methods don't
			// share this masking (see Layer-3 method-decorator tests).
			{Code: `
class A {
  @deco(this)
  foo = 1;
}
    `},
			{Code: `
class A {
  @deco(this)
  accessor foo = 1;
}
    `},
			// ---- Dimension 4: private method ----
			{Code: `
class A {
  #foo() {
    this;
  }
}
    `},
			// ---- Dimension 4: static method on class expression ----
			{Code: `
const X = class {
  static foo() {
    this;
  }
};
    `},
			// ---- Dimension 4: async method ----
			{Code: `
class A {
  async foo() {
    this;
  }
}
    `},
			// ---- Dimension 4: generator method ----
			{Code: `
class A {
  *foo() {
    this;
  }
}
    `},
			// ---- Dimension 4: async generator method ----
			{Code: `
class A {
  async *foo() {
    this;
  }
}
    `},
			// ---- Dimension 4: getter/setter on class ----
			{Code: `
class A {
  get foo() {
    return this;
  }
  set foo(v) {
    this.v = v;
  }
}
    `},
			// ---- Dimension 4: empty class body ----
			{Code: `
class A {}
    `},
			// ---- Dimension 4: empty function body ----
			{Code: `
function foo() {}
    `},
			// ---- Dimension 4: abstract bodyless method does not crash ----
			// Bodyless members never visit `this`, but ensure the visitor
			// still pops correctly when they appear adjacent to bodied ones.
			{Code: `
abstract class A {
  abstract foo(): void;
  bar() {
    this;
  }
}
    `},
			// ---- Dimension 4: overload signatures ----
			{Code: `
function foo(x: number): void;
function foo(x: string): void;
function foo(this: any, x: number | string): void {
  this;
}
    `},
			// ---- Dimension 4: class-in-class (inner method's `this` is inner instance) ----
			{Code: `
class A {
  foo() {
    class B {
      bar() {
        this;
      }
    }
  }
}
    `},
			// ---- Dimension 4: class static block ----
			{Code: `
class A {
  static {
    this;
  }
}
    `},

			// ---- Branch lock-in: LogicalExpression `&&` (upstream only tests `||`) ----
			// Locks in upstream isDefaultThisBinding() LogicalExpression arm
			// with `&&` operator — walker should treat as transparent.
			{Code: `
var obj = {
  foo: maybeFn && function () {
    this;
  },
};
    `},
			// ---- Branch lock-in: LogicalExpression `??` (nullish coalescing) ----
			{Code: `
var obj = {
  foo: maybeFn ?? function () {
    this;
  },
};
    `},
			// ---- Branch lock-in: nested `||` chains ----
			{Code: `
obj.foo = a || b || function () {
  this;
};
    `},

			// ---- Branch lock-in: BindingElement with uppercase name & default function ----
			// Locks in upstream AssignmentPattern arm for declaration-style
			// destructuring `var { Ctor = function(){} } = obj` (tsgo:
			// BindingElement with `name=Ctor`, `initializer=function`).
			{Code: `
var { Ctor = function () {
  this;
} } = obj;
    `},
			// ---- Branch lock-in: BindingElement in array pattern ----
			{Code: `
var [Ctor = function () {
  this;
}] = arr;
    `},
			// ---- Branch lock-in: ShorthandPropertyAssignment with default function ----
			// `({Ctor = function(){}} = obj)` — assignment-style destructuring.
			// In tsgo the inner element is a ShorthandPropertyAssignment with
			// `ObjectAssignmentInitializer`, distinct from BindingElement.
			{Code: `
({ Ctor = function () {
  this;
} } = obj);
    `},

			// ---- Branch lock-in: Reflect.apply with non-null thisArg ----
			// Already covered upstream but lock in element-access form.
			{Code: `
Reflect['apply'](
  function () {
    this;
  },
  obj,
  [],
);
    `},
			// ---- Branch lock-in: Array.fromAsync with thisArg ----
			{Code: `
Array.fromAsync(
  iter,
  function () {
    this;
  },
  obj,
);
    `},

			// ---- Branch lock-in: findLast / findLastIndex / flatMap (less-tested array methods) ----
			{Code: `
foo.findLast(function () {
  this;
}, obj);
    `},
			{Code: `
foo.findLastIndex(function () {
  this;
}, obj);
    `},
			{Code: `
foo.flatMap(function () {
  this;
}, obj);
    `},

			// ---- Real-user: `this` inside the value position of a class field whose initializer is a paren-wrapped function expression ----
			// Mirrors the "implicit method" patterns React class components use.
			{Code: `
class C {
  handler = (function () {
    this;
  });
}
    `},
			// ---- Real-user: `this` in an arrow callback inside a method ----
			// Direct hit on lexical-`this` semantics; broken if arrows accidentally push.
			{Code: `
class C {
  foo() {
    setTimeout(() => {
      this;
    });
  }
}
    `},
			// ---- Real-user: callback wrapped with a typed @this-style JSDoc on the parent statement ----
			{Code: `
/** @this Mocha.Context */
function setup() {
  this.timeout(100);
}
    `},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: `this` inside a computed method key (class) ----
			// At class-evaluation time `this` is the surrounding scope, NOT
			// the method's instance. The wrapper's FunctionExpression
			// listener fires on the FE child (visited AFTER the key in
			// ESTree), so the key's `this` is delegated to baseRule, which
			// reports at top level.
			{
				Code: `
class A {
  [this.foo]() {}
}
      `,
				Errors: unexpected(3, 4),
			},
			// ---- Dimension 4: `this` inside a computed accessor key ----
			// Same reasoning: get/set accessors are method-like, so the
			// wrapper masks the key's `this` only via baseRule (which reports).
			{
				Code: `
class A {
  get [this.foo]() {
    return 1;
  }
}
      `,
				Errors: unexpected(3, 8),
			},
			// ---- Dimension 4: `this` inside a computed object-literal method key ----
			{
				Code: `
var obj = {
  [this.foo]() {},
};
      `,
				Errors: unexpected(3, 4),
			},
			// ---- Dimension 4: `this` in decorator on a method ----
			// Decorators on method-likes run at class-evaluation time, so
			// the wrapper's `FunctionExpression` push (which fires on FE
			// entry, AFTER decorators in ESTree) doesn't mask them.
			// rslint compensates for tsgo's MethodDeclaration-is-the-FE
			// collapse by peeking one frame deeper at `this` inside such
			// a decorator.
			{
				Code: `
class A {
  @deco(this)
  foo() {}
}
      `,
				Errors: unexpected(3, 9),
			},
			{
				Code: `
class A {
  @deco(this)
  get foo() {
    return 1;
  }
}
      `,
				Errors: unexpected(3, 9),
			},
			// ---- Branch lock-in: TS expression wrappers are opaque to the walker ----
			// Upstream's `isDefaultThisBinding` has no case for
			// `TSAsExpression` / `TSSatisfiesExpression` / `TSNonNullExpression`,
			// so a function wrapped in any of them falls through to the
			// default branch (default-bound, `this` reported). Locked in
			// for all three so future "be smart about wrappers" refactors
			// can't silently flip semantics.
			{
				Code: `
(function () {
  this;
}!).call(obj);
      `,
				Errors: unexpected(3, 3),
			},
			{
				Code: `
(function () {
  this;
} as () => void).call(obj);
      `,
				Errors: unexpected(3, 3),
			},
			{
				Code: `
(function () {
  this;
} satisfies () => void).call(obj);
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Dimension 4: arrow inside decorator on method ----
			// Arrow's `this` is lexical → captures decorator's scope (outer).
			// The walker peeks past the arrow when deciding which stack
			// frame applies.
			{
				Code: `
class A {
  @deco(() => this)
  foo() {}
}
      `,
				Errors: unexpected(3, 15),
			},
			// ---- Branch lock-in: auto-accessor with function-expression initializer ----
			// ESLint's walker has no `AccessorProperty` case, so a function
			// inside `accessor x = function(){}` falls through to default
			// (default-bound, `this` is invalid). Regular fields stay valid
			// because the walker DOES recognize `PropertyDefinition`.
			{
				Code: `
class A {
  accessor foo = function () {
    this;
  };
}
      `,
				Errors: unexpected(4, 5),
			},
			// ---- Dimension 4: free function nested inside a method body ----
			// Locks in that the method's frame doesn't shadow the inner
			// function frame — the inner free function's `this` is reported.
			{
				Code: `
class A {
  foo() {
    function bar() {
      this;
    }
  }
}
      `,
				Errors: unexpected(5, 7),
			},
			// ---- Dimension 4: free function nested inside an arrow body ----
			// Arrow doesn't push, but the inner function does. The inner
			// function is in default-binding context (the arrow body).
			{
				Code: `
const f = () => {
  function bar() {
    this;
  }
};
      `,
				Errors: unexpected(4, 5),
			},
			// ---- Dimension 4: async generator at top level ----
			{
				Code: `
async function* gen() {
  this;
}
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Dimension 4: arrow in static block — own `this` is class, but a nested function loses it ----
			{
				Code: `
class A {
  static {
    function inner() {
      this;
    }
  }
}
      `,
				Errors: unexpected(5, 7),
			},
			// ---- Dimension 4: function declared in module top level (no enclosing class) ----
			{
				Code: `
function foo() {
  this;
}
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Dimension 4: `.call` with `null` first arg (mirror of upstream `.bind(null)`) ----
			{
				Code: `
(function () {
  this;
}).call(null);
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Dimension 4: `.apply` with no args ----
			{
				Code: `
(function () {
  this;
}).apply();
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Dimension 4: element-access `.call` with null thisArg ----
			{
				Code: `
(function () {
  this;
})['call'](null);
      `,
				Errors: unexpected(3, 3),
			},

			// ---- Branch lock-in: BinaryExpression non-= operator (e.g. +) ----
			// Function in a non-assignment binary context — default binding.
			// Locks in upstream isDefaultThisBinding default branch.
			{
				Code: `
var x = 1 + function () {
  return this;
}();
      `,
				Errors: unexpected(3, 10),
			},
			// ---- Branch lock-in: comma operator's right operand ----
			// tsgo collapses SequenceExpression onto BinaryExpression(',').
			// ESLint's walker has no SequenceExpression case so it falls
			// through to `default: return true`; we mirror that and the
			// function is treated as default-bound.
			{
				Code: `
obj.foo = (0, function () {
  this;
});
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Branch lock-in: `.call` without arguments ----
			{
				Code: `
(function () {
  this;
}).call();
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Branch lock-in: Reflect.apply with too few args ----
			{
				Code: `
Reflect.apply(function () {
  this;
});
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Branch lock-in: Array.from with too few args ----
			{
				Code: `
Array.from([], function () {
  this;
});
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Branch lock-in: array method with thisArg=undefined ----
			{
				Code: `
foo.forEach(function () {
  this;
}, undefined);
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Branch lock-in: non-callee member access `.call` (not invoked) ----
			// `(function(){...}).call` without immediate invocation isn't .call binding.
			{
				Code: `
var x = (function () {
  return this;
}).call;
      `,
				Errors: unexpected(3, 10),
			},
			// ---- Branch lock-in: BindingElement with lowercase name & default function ----
			{
				Code: `
var { ctor = function () {
  this;
} } = obj;
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Branch lock-in: ShorthandPropertyAssignment with lowercase name & default function ----
			{
				Code: `
({ ctor = function () {
  this;
} } = obj);
      `,
				Errors: unexpected(3, 3),
			},

			// ---- Real-user: `this` inside callback to an unknown method ----
			// `arr.unknown(fn)` — `unknown` isn't in the array-method-with-thisArg
			// allow list, so the callback's `this` is default-bound.
			{
				Code: `
foo.reduce(function () {
  this;
});
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Real-user: chained method call resembles array method but isn't ----
			{
				Code: `
foo.flatten(function () {
  this;
}, obj);
      `,
				Errors: unexpected(3, 3),
			},
			// ---- Real-user: `this` inside the value of a regular variable (no uppercase) ----
			{
				Code: `
var func = function bar() {
  this;
};
      `,
				Errors: unexpected(3, 3),
			},
		},
	)
}
