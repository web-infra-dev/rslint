package no_inner_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInnerDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInnerDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// Default mode ("functions") with blockScopedFunctions "allow" (default)
			{Code: `function doSomething() { }`},
			{Code: `function doSomething() { function somethingElse() { } }`},
			{Code: `(function() { function doSomething() { } }());`},
			{Code: `if (test) { var fn = function() { }; }`},
			{Code: `if (test) { var fn = function expr() { }; }`},
			{Code: `function decl() { var fn = function expr() { }; }`},
			{Code: `function decl(arg: any) { var fn; if (arg) { fn = function() { }; } }`},
			{Code: `if (test) var x = 42;`},
			{Code: `var x = 1;`},
			{Code: `var fn = function() { };`},
			{Code: `function foo() { if (test) { var x = 1; } }`},
			{Code: `if (test) { var foo; }`},
			{Code: `function doSomething() { while (test) { var foo; } }`},

			// Block-scoped functions in MODULE files (have import/export → strict → allowed)
			{Code: `export {}; if (foo) function f(){}`},
			{Code: `export {}; function bar() { if (foo) function f(){}; }`},
			{Code: `export {}; while (test) { function doSomething() { } }`},
			{Code: `export {}; do { function foo() {} } while (test)`},

			// Block-scoped functions with "use strict" directive → allowed
			{Code: `"use strict"; if (foo) function f(){}`},
			{Code: `"use strict"; function bar() { if (foo) function f(){}; }`},

			// Block-scoped functions inside function with "use strict" → allowed
			{Code: `function outer() { "use strict"; if (foo) function f(){}; }`},
			{Code: `function outer() { "use strict"; { function inner() {} } }`},

			// Block-scoped functions inside class body (implicit strict) → allowed
			{Code: `class C { method() { if(test) { function somethingElse() { } } } }`},
			{Code: `class C { method() { { function bar() { } } } }`},

			// Export declarations at module root
			{Code: `export function foo() {}`},
			{Code: `export default function() {}`},

			// "both" mode - valid placements
			{Code: `function doSomething() { }`, Options: []interface{}{"both"}},
			{Code: `function doSomething() { var x = 1; }`, Options: []interface{}{"both"}},
			{Code: `var x = 1;`, Options: []interface{}{"both"}},
			{Code: `var foo = 42;`, Options: []interface{}{"both"}},
			{Code: `var fn = function() { };`, Options: []interface{}{"both"}},
			{Code: `function foo() { var x = 1; }`, Options: []interface{}{"both"}},
			{Code: `(function() { var foo; }());`, Options: []interface{}{"both"}},

			// Arrow functions with "both" mode
			{Code: `export {}; foo(() => { function bar() { } });`},
			{Code: `var fn = () => {var foo;}`, Options: []interface{}{"both"}},
			{Code: `const doSomething = () => { var foo = 42; }`, Options: []interface{}{"both"}},

			// Class methods
			{Code: `var x = {doSomething() {function doSomethingElse() {}}}`},
			{Code: `var x = {doSomething() {var foo;}}`, Options: []interface{}{"both"}},
			{Code: `class C { method() { function foo() {} } }`, Options: []interface{}{"both"}},
			{Code: `class C { method() { var x; } }`, Options: []interface{}{"both"}},

			// Class static blocks
			{Code: `class C { static { function foo() {} } }`, Options: []interface{}{"both"}},
			{Code: `class C { static { var x; } }`, Options: []interface{}{"both"}},

			// let/const should never be flagged even in "both" mode
			{Code: `if (test) { let x = 1; }`, Options: []interface{}{"both"}},
			{Code: `if (test) { const x = 1; }`, Options: []interface{}{"both"}},

			// Export with "both" mode
			{Code: `export var foo: any;`, Options: []interface{}{"both"}},
			{Code: `export function bar() {}`, Options: []interface{}{"both"}},
			{Code: `export default function baz() {}`, Options: []interface{}{"both"}},

			// blockScopedFunctions "allow" in module/strict contexts
			{Code: `export {}; function foo() { { function bar() { } } }`,
				Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "allow"}}},

			// TypeScript-specific: declarations in namespace/module blocks
			{Code: `namespace Foo { function bar() {} }`},
			{Code: `namespace Foo { var x = 1; }`, Options: []interface{}{"both"}},
			{Code: `declare module 'foo' { function bar(): void; }`},
		},
		[]rule_tester.InvalidTestCase{
			// === blockScopedFunctions "allow" (default) in NON-STRICT script files ===
			// No import/export, no "use strict" → script mode → non-strict → reported
			{
				Code: `if (foo) function f(){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code: `function bar() { if (foo) function f(){}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code: `while (test) { function doSomething() { } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code: `do { function foo() {} } while (test)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code: `function doSomething() { do { function somethingElse() { } } while (test); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code: `(function() { if (test) { function doSomething() { } } }());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			// Bare block inside function (non-strict script)
			{
				Code: `function foo() { { function bar() { } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// === blockScopedFunctions "disallow" — always reports regardless of strict mode ===
			{
				Code:    `export {}; if (foo) function f(){}`,
				Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `export {}; function bar() { if (foo) function f(){}; }`,
				Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `export {}; while (test) { function doSomething() { } }`,
				Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `export {}; do { function foo() {} } while (test)`,
				Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `export {}; function doSomething() { do { function somethingElse() { } } while (test); }`,
				Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `export {}; (function() { if (test) { function doSomething() { } } }());`,
				Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// === "both" mode — var declarations in blocks ===
			{
				Code:    `if (foo) { var a; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `if (foo) var a;`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `function bar() { if (foo) var a; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `if (foo) { var fn = function(){} }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `while (test) { var foo; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `function doSomething() { if (test) { var foo = 42; } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `(function() { if (test) { var foo; } }());`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `const doSomething = () => { if (test) { var foo = 42; } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Class method - var in block
			{
				Code:    `class C { method() { if(test) { var foo; } } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Class static block - var in nested block
			{
				Code:    `class C { static { if (test) { var foo; } } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			// Class static block - function with blockScopedFunctions "disallow"
			{
				Code:    `class C { static { if (test) { function foo() {} } } }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			// Class static block - deeply nested var
			{
				Code:    `class C { static { if (test) { if (anotherTest) { var foo; } } } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// === "both" + blockScopedFunctions "disallow" ===
			{
				Code:    `if (foo) { var bar = 1; }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			{
				Code:    `export {}; if (test) { function doSomething() { } }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
			// Function in bare block inside function with blockScopedFunctions "disallow"
			{
				Code:    `export {}; function foo() { { function bar() { } } }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// TypeScript-specific: var in nested block inside namespace
			{
				Code:    `namespace Foo { if (test) { var x = 1; } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
		},
	)
}
