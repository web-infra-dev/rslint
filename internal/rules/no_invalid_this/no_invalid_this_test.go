package no_invalid_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoInvalidThisRule(t *testing.T) {
	defaultOptions := map[string]interface{}{"capIsConstructor": true}
	capIsConstructorFalse := map[string]interface{}{"capIsConstructor": false}
	emptyOptions := map[string]interface{}{}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInvalidThisRule, []rule_tester.ValidTestCase{
		// TypeScript-specific: this parameter
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

		// @this JSDoc tag
		{Code: `
function foo() {
  /** @this Obj*/ return function bar() {
    console.log(this);
    z(x => console.log(x, this));
  };
}
		`},

		// Constructors (capIsConstructor: true)
		{Code: `
var Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
		`},
		{Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
		`},
		{Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
		`, Options: emptyOptions},
		{Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
		`, Options: defaultOptions},
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

		// Object methods
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

		// Property assignment
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

		// Bind/Call/Apply
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

		// Class methods
		{Code: `
class A {
  foo() {
    console.log(this);
    z(x => console.log(x, this));
  }
}
		`},

		// Class properties
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

		// Array methods with thisArg
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

		// @this JSDoc tag
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

		// Assignment patterns
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

		// Static methods
		{Code: `
class A {
  static foo() {
    console.log(this);
    z(x => console.log(x, this));
  }
}
		`},

		// Accessor properties
		{Code: `
class A {
  a = 5;
  b = this.a;
  accessor c = this.a;
}
		`},
	}, []rule_tester.InvalidTestCase{
		// Global scope
		{
			Code: `
interface SomeType {
  prop: string;
}
function foo() {
  this.prop;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 6, Column: 3},
			},
		},
		{
			Code: `
console.log(this);
z(x => console.log(x, this));
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 2, Column: 13},
				{MessageId: "unexpectedThis", Line: 3, Column: 23},
			},
		},

		// IIFE
		{
			Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
})();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},

		// Regular functions
		{
			Code: `
function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
			`,
			Options: capIsConstructorFalse,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
function Foo() {
  console.log(this);
  z(x => console.log(x, this));
}
			`,
			Options: capIsConstructorFalse,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
function foo() {
  'use strict';
  console.log(this);
  z(x => console.log(x, this));
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 4, Column: 15},
				{MessageId: "unexpectedThis", Line: 5, Column: 25},
			},
		},
		{
			Code: `
function Foo() {
  'use strict';
  console.log(this);
  z(x => console.log(x, this));
}
			`,
			Options: capIsConstructorFalse,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 4, Column: 15},
				{MessageId: "unexpectedThis", Line: 5, Column: 25},
			},
		},
		{
			Code: `
var foo = function () {
  console.log(this);
  z(x => console.log(x, this));
}.bar(obj);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},

		// Nested functions in methods
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 5, Column: 19},
				{MessageId: "unexpectedThis", Line: 6, Column: 29},
			},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 5, Column: 19},
				{MessageId: "unexpectedThis", Line: 6, Column: 29},
			},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 5, Column: 19},
				{MessageId: "unexpectedThis", Line: 6, Column: 29},
			},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 6, Column: 19},
				{MessageId: "unexpectedThis", Line: 7, Column: 29},
			},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 4, Column: 17},
				{MessageId: "unexpectedThis", Line: 5, Column: 27},
			},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 5, Column: 17},
				{MessageId: "unexpectedThis", Line: 6, Column: 27},
			},
		},

		// Class methods with nested functions
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 5, Column: 19},
				{MessageId: "unexpectedThis", Line: 6, Column: 29},
			},
		},

		// Class properties with nested functions
		{
			Code: `
class A {
  b = new Array(1, 2, function () {
    console.log(this);
    z(x => console.log(x, this));
  });
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 4, Column: 17},
				{MessageId: "unexpectedThis", Line: 5, Column: 27},
			},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 5, Column: 19},
				{MessageId: "unexpectedThis", Line: 6, Column: 29},
			},
		},

		// Arrow functions returning arrow functions
		{
			Code: `
obj.foo = (function () {
  return () => {
    console.log(this);
    z(x => console.log(x, this));
  };
})();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 4, Column: 17},
				{MessageId: "unexpectedThis", Line: 5, Column: 27},
			},
		},
		{
			Code: `
obj.foo = (() => () => {
  console.log(this);
  z(x => console.log(x, this));
})();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},

		// Bind/Call/Apply with null/undefined
		{
			Code: `
var foo = function () {
  console.log(this);
  z(x => console.log(x, this));
}.bind(null);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
}).call(undefined);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
(function () {
  console.log(this);
  z(x => console.log(x, this));
}).apply(void 0);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},

		// Array methods without thisArg
		{
			Code: `
Array.from([], function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
foo.every(function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
foo.filter(function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
foo.find(function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
foo.findIndex(function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
foo.forEach(function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
foo.map(function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
foo.some(function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
foo.forEach(function () {
  console.log(this);
  z(x => console.log(x, this));
}, null);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},

		// Functions without @this tag
		{
			Code: `
/** @returns {void} */ function foo() {
  console.log(this);
  z(x => console.log(x, this));
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
/** @this Obj */ foo(function () {
  console.log(this);
  z(x => console.log(x, this));
});
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},

		// Variable assignments with capIsConstructor: false
		{
			Code: `
var Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
			`,
			Options: capIsConstructorFalse,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
var func = function () {
  console.log(this);
  z(x => console.log(x, this));
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
var func = function () {
  console.log(this);
  z(x => console.log(x, this));
};
			`,
			Options: capIsConstructorFalse,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
Ctor = function () {
  console.log(this);
  z(x => console.log(x, this));
};
			`,
			Options: capIsConstructorFalse,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
func = function () {
  console.log(this);
  z(x => console.log(x, this));
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
		},
		{
			Code: `
func = function () {
  console.log(this);
  z(x => console.log(x, this));
};
			`,
			Options: capIsConstructorFalse,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 3, Column: 15},
				{MessageId: "unexpectedThis", Line: 4, Column: 25},
			},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 4, Column: 17},
				{MessageId: "unexpectedThis", Line: 5, Column: 27},
			},
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
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedThis", Line: 4, Column: 17},
				{MessageId: "unexpectedThis", Line: 5, Column: 27},
			},
		},
	})
}
