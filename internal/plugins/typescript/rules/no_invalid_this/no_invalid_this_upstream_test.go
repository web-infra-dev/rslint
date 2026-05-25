// TestNoInvalidThisUpstream migrates the full valid/invalid suite from
// upstream typescript-eslint's
//
//	packages/eslint-plugin/tests/rules/no-invalid-this.test.ts
//
// 1:1. Position assertions cover line/column for every invalid case
// (the upstream typescript-eslint suite only asserts messageId, so this
// layer adds line/column to satisfy the rslint requirement that invalid
// cases pin position).
//
// rslint-specific lock-in cases (Dimension 4 edge shapes, branch lock-ins,
// real-user issue shapes) live in no_invalid_this_extras_test.go.
package no_invalid_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// objectOption produces the array-wrapped option shape that exercises
// utils.GetOptionsMap's JSON path. Passing a typed struct directly would
// short-circuit the JSON round-trip and leave the CLI-facing wiring untested.
func objectOption(opts map[string]interface{}) []interface{} {
	return []interface{}{opts}
}

func TestNoInvalidThisUpstream(t *testing.T) {
	// `unexpected` is the standard 1-error shape used by single-`this` cases.
	unexpected := func(line, col int) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{
			{MessageId: "unexpectedThis", Line: line, Column: col},
		}
	}
	// `unexpected2` is the standard 2-error shape: outer `this` + the
	// trailing `z(x => console.log(x, this));` arrow `this`. Used by the
	// vast majority of upstream invalid cases.
	unexpected2 := func(line1, col1, line2, col2 int) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{
			{MessageId: "unexpectedThis", Line: line1, Column: col1},
			{MessageId: "unexpectedThis", Line: line2, Column: col2},
		}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidThisRule,
		[]rule_tester.ValidTestCase{
			// ---- TypeScript-specific: `this` parameter ----
			{Code: `
describe('foo', () => {
  it('does something', function (this: Mocha.Context) {
    this.timeout(100);
    // done
  });
});
    `},
			{Code: `
      interface SomeType {
        prop: string;
      }
      function foo(this: SomeType) {
        this.prop;
      }
    `},
			{Code: `
function foo(this: prop) {
  this.propMethod();
}
    `},
			{Code: `
z(function (x, this: context) {
  console.log(x, this);
});
    `},

			// ---- @this JSDoc tag on nested function (eslint #3287) ----
			{Code: `
function foo() {
  /** @this Obj*/ return function bar() {
    console.log(this);
    z(x => console.log(x, this));
  };
}
    `},

			// ---- uppercase-named variable initializer (eslint #6824) ----
			{Code: `
var Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
    `},

			// ---- Constructors (capIsConstructor: true is the default) ----
			{Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `},
			{
				Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `,
				Options: objectOption(map[string]interface{}{}),
			},
			{
				Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `,
				Options: objectOption(map[string]interface{}{"capIsConstructor": true}),
			},
			{Code: `
var Foo = function Foo() {
  console.log(this);
  z(x => console.log(x, this));
};
      `},
			{Code: `
class A {
  constructor() {
    console.log(this);
    z(x => console.log(x, this));
  }
}
      `},

			// ---- On a property ----
			{Code: `
var obj = {
  foo: function () {
    console.log(this);
    z(x => console.log(x, this));
  },
};
      `},
			{Code: `
var obj = {
  foo() {
    console.log(this);
    z(x => console.log(x, this));
  },
};
      `},
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
			{Code: `
Object.defineProperty(obj, 'foo', {
  value: function () {
    console.log(this);
    z(x => console.log(x, this));
  },
});
      `},
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

			// ---- Assigns to a property ----
			{Code: `
obj.foo = function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `},
			{Code: `
obj.foo =
  foo ||
  function () {
    console.log(this);
    z(x => console.log(x, this));
  };
      `},
			{Code: `
obj.foo = foo
  ? bar
  : function () {
      console.log(this);
      z(x => console.log(x, this));
    };
      `},
			{Code: `
obj.foo = (function () {
  return function () {
    console.log(this);
    z(x => console.log(x, this));
  };
})();
      `},
			{Code: `
obj.foo = (() =>
  function () {
    console.log(this);
    z(x => console.log(x, this));
  })();
      `},

			// ---- Bind/Call/Apply ----
			{Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
}).call(obj);
    `},
			{Code: `
var foo = function () {
  console.log(this);
  z(x => console.log(x, this));
}.bind(obj);
    `},
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
			{Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
}).apply(obj);
    `},

			// ---- Class Instance Methods ----
			{Code: `
class A {
  foo() {
    console.log(this);
    z(x => console.log(x, this));
  }
}
    `},

			// ---- Class Properties (regular fields) ----
			{Code: `
class A {
  b = 0;
  c = this.b;
}
    `},
			{Code: `
class A {
  b = new Array(this, 1, 2, 3);
}
    `},
			{Code: `
class A {
  b = () => {
    console.log(this);
  };
}
    `},

			// ---- Array methods with thisArg ----
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
			{Code: `
foo.every(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
    `},
			{Code: `
foo.filter(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
    `},
			{Code: `
foo.find(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
    `},
			{Code: `
foo.findIndex(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
    `},
			{Code: `
foo.forEach(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
    `},
			{Code: `
foo.map(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
    `},
			{Code: `
foo.some(function () {
  console.log(this);
  z(x => console.log(x, this));
}, obj);
    `},

			// ---- @this tag ----
			{Code: `
/** @this Obj */ function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
    `},
			{Code: `
foo(
  /* @this Obj */ function () {
    console.log(this);
    z(x => console.log(x, this));
  },
);
    `},
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

			// ---- Uppercase assignment / parameter default / destructuring ----
			{Code: `
Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
    `},
			{Code: `
function foo(
  Ctor = function () {
    console.log(this);
    z(x => console.log(x, this));
  },
) {}
    `},
			{Code: `
[
  obj.method = function () {
    console.log(this);
    z(x => console.log(x, this));
  },
] = a;
    `},

			// ---- Static methods & auto-accessor class properties ----
			{Code: `
class A {
  static foo() {
    console.log(this);
    z(x => console.log(x, this));
  }
}
    `},
			{Code: `
class A {
  a = 5;
  b = this.a;
  accessor c = this.a;
}
    `},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Interface + free function ----
			{
				Code: `
interface SomeType {
  prop: string;
}
function foo() {
  this.prop;
}
      `,
				Errors: unexpected(6, 3),
			},

			// ---- Global (top-level `this`) ----
			{
				Code: `
console.log(this);
z(x => console.log(x, this));
      `,
				Errors: unexpected2(2, 13, 3, 23),
			},
			{
				Code: `
console.log(this);
z(x => console.log(x, this));
      `,
				Errors: unexpected2(2, 13, 3, 23),
				// Upstream sets parserOptions.ecmaFeatures.globalReturn:true.
				// rslint does not expose ecmaFeatures; the module-default
				// behavior gives the same diagnostics so the case still
				// locks in the expected output. See no_invalid_this.md
				// "Differences from ESLint" for the broader divergence.
			},

			// ---- IIFE ----
			{
				Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
})();
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},

			// ---- Just functions ----
			{
				Code: `
function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
function foo() {
  'use strict';
  console.log(this);
  z(x => console.log(x, this));
}
      `,
				Errors: unexpected2(4, 15, 5, 25),
			},
			{
				Code: `
function Foo() {
  'use strict';
  console.log(this);
  z(x => console.log(x, this));
}
      `,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  unexpected2(4, 15, 5, 25),
			},
			{
				// SKIP: rslint does not support ESLint's parserOptions.ecmaFeatures.globalReturn,
				// without which a top-level `return` is a parse error.
				Skip: true,
				Code: `
return function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
var foo = function () {
  console.log(this);
  z(x => console.log(x, this));
}.bar(obj);
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},

			// ---- Functions in methods ----
			{
				Code: `
var obj = {
  foo: function () {
    function foo() {
      console.log(this);
      z(x => console.log(x, this));
    }
    foo();
  },
};
      `,
				Errors: unexpected2(5, 19, 6, 29),
			},
			{
				Code: `
var obj = {
  foo() {
    function foo() {
      console.log(this);
      z(x => console.log(x, this));
    }
    foo();
  },
};
      `,
				Errors: unexpected2(5, 19, 6, 29),
			},
			{
				Code: `
var obj = {
  foo: function () {
    return function () {
      console.log(this);
      z(x => console.log(x, this));
    };
  },
};
      `,
				Errors: unexpected2(5, 19, 6, 29),
			},
			{
				Code: `
var obj = {
  foo: function () {
    'use strict';
    return function () {
      console.log(this);
      z(x => console.log(x, this));
    };
  },
};
      `,
				Errors: unexpected2(6, 19, 7, 29),
			},
			{
				Code: `
obj.foo = function () {
  return function () {
    console.log(this);
    z(x => console.log(x, this));
  };
};
      `,
				Errors: unexpected2(4, 17, 5, 27),
			},
			{
				Code: `
obj.foo = function () {
  'use strict';
  return function () {
    console.log(this);
    z(x => console.log(x, this));
  };
};
      `,
				Errors: unexpected2(5, 17, 6, 27),
			},

			// ---- Class Methods (function returned from method) ----
			{
				Code: `
class A {
  foo() {
    return function () {
      console.log(this);
      z(x => console.log(x, this));
    };
  }
}
      `,
				Errors: unexpected2(5, 19, 6, 29),
			},

			// ---- Class Properties (free function in field initializer) ----
			{
				Code: `
class A {
  b = new Array(1, 2, function () {
    console.log(this);
    z(x => console.log(x, this));
  });
}
      `,
				Errors: unexpected2(4, 17, 5, 27),
			},
			{
				Code: `
class A {
  b = () => {
    function c() {
      console.log(this);
      z(x => console.log(x, this));
    }
  };
}
      `,
				Errors: unexpected2(5, 19, 6, 29),
			},

			// ---- Class Static methods (arrows returned from IIFE assigned to obj) ----
			{
				Code: `
obj.foo = (function () {
  return () => {
    console.log(this);
    z(x => console.log(x, this));
  };
})();
      `,
				Errors: unexpected2(4, 17, 5, 27),
			},
			{
				Code: `
obj.foo = (() => () => {
  console.log(this);
  z(x => console.log(x, this));
})();
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},

			// ---- Bind/Call/Apply with null/undefined thisArg ----
			{
				Code: `
var foo = function () {
  console.log(this);
  z(x => console.log(x, this));
}.bind(null);
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
}).call(undefined);
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
}).apply(void 0);
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},

			// ---- Array methods without thisArg ----
			{
				Code: `
Array.from([], function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
foo.every(function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
foo.filter(function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
foo.find(function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
foo.findIndex(function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
foo.forEach(function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
foo.map(function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
foo.some(function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},

			// ---- Array method with null thisArg ----
			{
				Code: `
foo.forEach(function () {
  console.log(this);
  z(x => console.log(x, this));
}, null);
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},

			// ---- @this tag absent ----
			{
				Code: `
/** @returns {void} */ function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				// @this on the outer CallExpression doesn't apply to the function argument.
				Code: `
/** @this Obj */ foo(function () {
  console.log(this);
  z(x => console.log(x, this));
});
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},

			// ---- capIsConstructor:false disables uppercase recognition ----
			{
				Code: `
var Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
var func = function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
var func = function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  unexpected2(3, 15, 4, 25),
			},

			{
				Code: `
Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
func = function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `,
				Errors: unexpected2(3, 15, 4, 25),
			},
			{
				Code: `
func = function () {
  console.log(this);
  z(x => console.log(x, this));
};
      `,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  unexpected2(3, 15, 4, 25),
			},

			{
				Code: `
function foo(
  func = function () {
    console.log(this);
    z(x => console.log(x, this));
  },
) {}
      `,
				Errors: unexpected2(4, 17, 5, 27),
			},

			{
				Code: `
[
  func = function () {
    console.log(this);
    z(x => console.log(x, this));
  },
] = a;
      `,
				Errors: unexpected2(4, 17, 5, 27),
			},
		},
	)
}
