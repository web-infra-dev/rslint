package class_methods_use_this

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestClassMethodsUseThisRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ClassMethodsUseThisRule, []rule_tester.ValidTestCase{
		// Static methods should be ignored
		{Code: `
class A {
  static foo() {
    console.log('foo');
  }
}
		`},

		// Constructor should be ignored
		{Code: `
class A {
  constructor() {
    console.log('constructor');
  }
}
		`},

		// Methods that use 'this' should be valid
		{Code: `
class A {
  foo() {
    return this.bar;
  }
}
		`},

		// Methods that use 'super' should be valid
		{Code: `
class A {
  foo() {
    return super.foo();
  }
}
		`},

		// Abstract methods should be ignored
		{Code: `
abstract class A {
  abstract foo(): void;
}
		`},

		// Methods in classes that implement interfaces (when ignoreClassesThatImplementAnInterface is true)
		{
			Code: `
interface I {
  foo(): void;
}
class A implements I {
  foo() {
    console.log('foo');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"ignoreClassesThatImplementAnInterface": true,
			}},
		},

		// Public fields in classes that implement interfaces (when ignoreClassesThatImplementAnInterface is "public-fields")
		{
			Code: `
interface I {
  foo(): void;
}
class A implements I {
  foo() {
    console.log('foo');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"ignoreClassesThatImplementAnInterface": "public-fields",
			}},
		},

		// Private methods in classes that implement interfaces should still be checked by default
		{Code: `
interface I {
  foo(): void;
}
class A implements I {
  private bar() {
    return this.baz;
  }
  foo() {
    return this.bar();
  }
}
		`},

		// Methods with override modifier (when ignoreOverrideMethods is true)
		{
			Code: `
class Base {
  foo() {
    return this.bar;
  }
}
class A extends Base {
  override foo() {
    console.log('overridden');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"ignoreOverrideMethods": true,
			}},
		},

		// Methods in except list
		{
			Code: `
class A {
  foo() {
    console.log('foo');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"exceptMethods": []interface{}{"foo"},
			}},
		},

		// Private methods in except list
		{
			Code: `
class A {
  #foo() {
    console.log('private foo');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"exceptMethods": []interface{}{"#foo"},
			}},
		},

		// Property initializers (when enforceForClassFields is false)
		{
			Code: `
class A {
  foo = () => {
    console.log('foo');
  };
}
			`,
			Options: []interface{}{map[string]interface{}{
				"enforceForClassFields": false,
			}},
		},

		// Getter/setter properties (when enforceForClassFields is false)
		{
			Code: `
class A {
  get foo() {
    return 'foo';
  }
  set foo(value) {
    console.log(value);
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"enforceForClassFields": false,
			}},
		},

		// Property initializers that use 'this'
		{Code: `
class A {
  foo = () => {
    return this.bar;
  };
}
		`},

		// Computed property names
		{Code: `
class A {
  [methodName]() {
    console.log('computed');
  }
}
		`},

		// Function declarations have their own 'this' context
		{Code: `
class A {
  foo() {
    function bar() {
      console.log('bar');
    }
    return this.baz;
  }
}
		`},

		// Static blocks
		{Code: `
class A {
  static {
    console.log('static block');
  }
}
		`},

		// Methods that use 'this' in nested functions should be valid (this usage counts)
		{Code: `
class A {
  foo() {
    const self = this;
    function bar() {
      return self.baz;
    }
    return bar();
  }
}
		`},
	}, []rule_tester.InvalidTestCase{
		// Methods that don't use 'this' should be flagged
		{
			Code: `
class A {
  foo() {
    console.log('foo');
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      3,
					Column:    3,
				},
			},
		},

		// Function expressions that don't use 'this'
		{
			Code: `
class A {
  foo = function() {
    console.log('foo');
  };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      3,
					Column:    9,
				},
			},
		},

		// Arrow functions in property initializers that don't use 'this'
		{
			Code: `
class A {
  foo = () => {
    console.log('foo');
  };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      3,
					Column:    9,
				},
			},
		},

		// Getter that doesn't use 'this'
		{
			Code: `
class A {
  get foo() {
    return 'constant';
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      3,
					Column:    7,
				},
			},
		},

		// Setter that doesn't use 'this'
		{
			Code: `
class A {
  set foo(value) {
    console.log(value);
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      3,
					Column:    7,
				},
			},
		},

		// Multiple methods without 'this'
		{
			Code: `
class A {
  foo() {
    console.log('foo');
  }
  bar() {
    console.log('bar');
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      3,
					Column:    3,
				},
				{
					MessageId: "missingThis",
					Line:      6,
					Column:    3,
				},
			},
		},

		// Method in class that implements interface but with ignoreClassesThatImplementAnInterface: "public-fields" and private method
		{
			Code: `
interface I {
  foo(): void;
}
class A implements I {
  private bar() {
    console.log('bar');
  }
  foo() {
    return this.bar();
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"ignoreClassesThatImplementAnInterface": "public-fields",
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      6,
					Column:    11,
				},
			},
		},

		// Method without override modifier
		{
			Code: `
class Base {
  foo() {
    return this.bar;
  }
}
class A extends Base {
  foo() {
    console.log('overridden');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"ignoreOverrideMethods": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      8,
					Column:    3,
				},
			},
		},

		// Method not in except list
		{
			Code: `
class A {
  foo() {
    console.log('foo');
  }
  bar() {
    console.log('bar');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"exceptMethods": []interface{}{"foo"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      6,
					Column:    3,
				},
			},
		},

		// Private method not in except list
		{
			Code: `
class A {
  #foo() {
    console.log('private foo');
  }
  #bar() {
    console.log('private bar');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"exceptMethods": []interface{}{"#foo"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      6,
					Column:    3,
				},
			},
		},

		// Nested class methods
		{
			Code: `
class A {
  foo() {
    class B {
      bar() {
        console.log('nested');
      }
    }
    return this.baz;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      5,
					Column:      7,
				},
			},
		},

		// Method that calls other methods but doesn't use 'this'
		{
			Code: `
class A {
  foo() {
    this.bar();
    return someGlobalFunction();
  }
  bar() {
    console.log('bar');
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      7,
					Column:    3,
				},
			},
		},

		// Class expression
		{
			Code: `
const A = class {
  foo() {
    console.log('foo');
  }
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      3,
					Column:    3,
				},
			},
		},

		// Property initializer with enforceForClassFields disabled but still checking other methods
		{
			Code: `
class A {
  prop = () => {
    console.log('arrow');
  };
  method() {
    console.log('method');
  }
}
			`,
			Options: []interface{}{map[string]interface{}{
				"enforceForClassFields": false,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingThis",
					Line:      6,
					Column:    3,
				},
			},
		},
	})
}