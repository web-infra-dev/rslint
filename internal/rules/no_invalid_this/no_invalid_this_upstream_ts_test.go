// TestNoInvalidThisUpstreamTypeScript migrates the TypeScript-parser half
// of upstream ESLint core's
//
//	tests/lib/rules/no-invalid-this.js
//
// (the `ruleTesterTypeScript` suite, parsed with @typescript-eslint/parser
// — unlike the JS-parser "patterns" half in no_invalid_this_upstream_test.go,
// this section is not mode-multiplied by upstream itself).
//
// Position assertions cover line/column for every invalid case (upstream
// only asserts messageId).
//
// rslint-specific lock-in cases (Dimension 4 edge shapes, branch lock-ins,
// real-user issue shapes) live in no_invalid_this_extras_test.go.
package no_invalid_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidThisUpstreamTypeScript(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidThisRule,
		[]rule_tester.ValidTestCase{
			// ---- this-param on a method (Mocha-style callback) ----
			{Code: `
describe('foo', () => {
  it('does something', function (this: Mocha.Context) {
    this.timeout(100);
    // done
  });
});
    `},
			// ---- this-param + interface annotation ----
			{Code: `
interface SomeType {
  prop: string;
}
function foo(this: SomeType) {
  this.prop;
}
    `},
			// ---- this-param with a type reference ----
			{Code: `
function foo(this: prop) {
  this.propMethod();
}
    `},
			// ---- this-param as a non-first parameter position marker in a callback ----
			{Code: `
z(function (x, this: context) {
  console.log(x, this);
});
    `},
			// ---- JSDoc @this on parent ReturnStatement (gh#3287) ----
			{Code: `
function foo() {
  /** @this Obj*/ return function bar() {
    console.log(this);
    z(x => console.log(x, this));
  };
}
    `},
			// ---- anonymous FE assigned to uppercase var (ES5 constructor) ----
			{Code: `
var Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
    `},
			// ---- named FunctionDeclaration with uppercase name (ES5 constructor) ----
			{Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `},
			// ---- same, with empty options object (schema default) ----
			{Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `, Options: objectOption(map[string]interface{}{})},
			// ---- same, with capIsConstructor explicitly true (schema default) ----
			{Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `, Options: objectOption(map[string]interface{}{"capIsConstructor": true})},
			// ---- named FunctionExpression assigned to same-named uppercase var ----
			{Code: `
var Foo = function Foo() {
  console.log(this);
  z(x => console.log(x, this));
};
      `},
			// ---- class constructor ----
			{Code: `
class A {
  constructor() {
    console.log(this);
    z(x => console.log(x, this));
  }
}
      `},
			// ---- object literal property FunctionExpression ----
			{Code: `
var obj = {
  foo: function () {
    console.log(this);
    z(x => console.log(x, this));
  },
};
      `},
			// ---- object literal shorthand method ----
			{Code: `
var obj = {
  foo() {
    console.log(this);
    z(x => console.log(x, this));
  },
};
      `},
			// ---- logical-OR fallback assigned to property ----
			{Code: `
var obj = {
  foo:
    foo ||
    function () {
      console.log(this);
      z(x => console.log(x, this));
    },
};
      `},
			// ---- conditional expression assigned to property ----
			{Code: `
var obj = {
  foo: hasNative
    ? foo
    : function () {
        console.log(this);
        z(x => console.log(x, this));
      },
};
      `},
			// ---- IIFE return assigned to property ----
			{Code: `
var obj = {
  foo: (function () {
    return function () {
      console.log(this);
      z(x => console.log(x, this));
    };
  })(),
};
      `},
			// ---- Object.defineProperty value ----
			{Code: `
Object.defineProperty(obj, 'foo', {
  value: function () {
    console.log(this);
    z(x => console.log(x, this));
  },
});
      `},
			// ---- Object.defineProperties value ----
			{Code: `
Object.defineProperties(obj, {
  foo: {
    value: function () {
      console.log(this);
      z(x => console.log(x, this));
    },
  },
});
      `},
			// ---- member-expression assignment ----
			{Code: `
obj.foo = function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `},
			// ---- member-expression assignment, logical-OR fallback ----
			{Code: `
obj.foo =
  foo ||
  function () {
    console.log(this);
    z(x => console.log(x, this));
  };
      `},
			// ---- member-expression assignment, conditional expression ----
			{Code: `
obj.foo = foo
  ? bar
  : function () {
      console.log(this);
      z(x => console.log(x, this));
    };
      `},
			// ---- member-expression assignment, IIFE return ----
			{Code: `
obj.foo = (function () {
  return function () {
    console.log(this);
    z(x => console.log(x, this));
  };
})();
      `},
			// ---- member-expression assignment, arrow-IIFE return ----
			{Code: `
obj.foo = (() =>
  function () {
    console.log(this);
    z(x => console.log(x, this));
  })();
      `},
			// ---- .call(obj) ----
			{Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
}).call(obj);
        `},
			// ---- .bind(obj) ----
			{Code: `
var foo = function () {
  console.log(this);
  z(x => console.log(x, this));
}.bind(obj);
        `},
			// ---- Reflect.apply ----
			{Code: `
Reflect.apply(
  function () {
    console.log(this);
    z(x => console.log(x, this));
  },
  obj,
  [],
);
        `},
			// ---- .apply(obj) ----
			{Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
}).apply(obj);
        `},
			// ---- class instance method ----
			{Code: `
class A {
  foo() {
    console.log(this);
    z(x => console.log(x, this));
  }
}
        `},
			// ---- Array.from with thisArg ----
			{Code: `
Array.from(
  [],
  function () {
    console.log(this);
    z(x => console.log(x, this));
  },
  obj,
);
        `},
			// ---- Array.fromAsync with thisArg ----
			{Code: `
Array.fromAsync(
  [],
  function () {
    console.log(this);
    z(x => console.log(x, this));
  },
  obj,
);
        `},
			// ---- .every with thisArg ----
			{Code: `
foo.every(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
        `},
			// ---- .filter with thisArg ----
			{Code: `
foo.filter(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
        `},
			// ---- .find with thisArg ----
			{Code: `
foo.find(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
        `},
			// ---- .findIndex with thisArg ----
			{Code: `
foo.findIndex(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
        `},
			// ---- .forEach with thisArg ----
			{Code: `
foo.forEach(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
        `},
			// ---- .map with thisArg ----
			{Code: `
foo.map(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
        `},
			// ---- .some with thisArg ----
			{Code: `
foo.some(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
        `},
			// ---- own JSDoc @this tag ----
			{Code: `
/** @this Obj */ function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
        `},
			// ---- leading block comment @this on callback arg ----
			{Code: `
foo(
  /* @this Obj */ function () {
    console.log(this);
    z(x => console.log(x, this));
  },
);
        `},
			// ---- multi-line JSDoc with @this among other tags ----
			{Code: `
/**
 * @returns {void}
 * @this Obj
 */
function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
        `},
			// ---- assignment to uppercase identifier (ES5 constructor) ----
			{Code: `
Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
        `},
			// ---- parameter default, uppercase name (ES5 constructor) ----
			{Code: `
function foo(
  Ctor = function () {
    console.log(this);
    z(x => console.log(x, this));
  },
) {}
        `},
			// ---- destructuring default targeting a member expression ----
			{Code: `
[
  obj.method = function () {
    console.log(this);
    z(x => console.log(x, this));
  },
] = a;
        `},
			// ---- class field, non-computed, all valid regardless of expression form (accessor) ----
			{Code: `
          class A {
            a = 5;
            b = this.a;
            accessor c = this.a;
          }
          `},
			// ---- class field, accessor with arithmetic initializer ----
			{Code: `
          class A {
            a = 5;
            accessor b = this.a + 1;
          }
          `},
			// ---- class field, accessor initialized directly to this ----
			{Code: `
          class A {
            a = 5;
            accessor b = this;
          }
          `},
			// ---- class field referencing a sibling field ----
			{Code: `
          class A {
            b = 0;
            c = this.b;
          }
          `},
			// ---- class field, this inside a call argument ----
			{Code: `
          class A {
            b = new Array(this, 1, 2, 3);
          }
          `},
			// ---- class field, arrow function initializer ----
			{Code: `
          class A {
            b = () => {
            console.log(this);
            };
          }
          `},
			// ---- static method ----
			{Code: `
          class A {
            static foo() {
            console.log(this);
            z(x => console.log(x, this));
            }
          }
          `},
		},
		[]rule_tester.InvalidTestCase{
			// ---- interface + free function (no this-param) ----
			{
				Code: `export {};

    interface SomeType {
      prop: string;
    }
    function foo() {
      this.prop;
    }
          `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 7, Column: 7}},
			},
			// ---- class field, this inside a call argument passed a plain function (not arrow) ----
			{
				Code: `export {};

    class A {
      b = new Array(1, 2, function () {
        console.log(this);
        z(x => console.log(x, this));
      });
    }
          `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 5, Column: 21}, {MessageId: "unexpectedThis", Line: 6, Column: 31}},
			},
			// ---- class field arrow body containing a nested plain function ----
			{
				Code: `export {};

    class A {
      b = () => {
        function c() {
          console.log(this);
          z(x => console.log(x, this));
        }
      };
    }
          `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 6, Column: 23}, {MessageId: "unexpectedThis", Line: 7, Column: 33}},
			},
			// ---- computed accessor key referencing this (enclosing scope, not field value) ----
			{
				Code: `export {};

				function foo() {
  				class C {
    				accessor [this.a] = foo;
  				}
				}
          `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 5, Column: 19}},
			},
			// ---- computed accessor key this + valid accessor value this (only key errors) ----
			{
				Code: `export {};

				function foo() {
  				class C {
    				accessor [this.a] = this.b;
  				}
				}
          `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 5, Column: 19}},
			},
			// ---- non-computed accessor value this (valid) + computed accessor key this (invalid) ----
			{
				Code: `export {};

				function foo() {
  				class C {
    				accessor a = this.b;
    				accessor [this.c] = foo;
  				}
				}
          `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 6, Column: 19}},
			},
		},
	)
}
