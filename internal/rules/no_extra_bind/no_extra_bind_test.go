package no_extra_bind

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraBindRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraBindRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// --- Bind access patterns ---
			{Code: `var a = function(b) { return b }.bind(c, d)`},
			{Code: `var a = function(b: any) { return b }.bind(...c)`},
			{Code: `var a = function() { return 1; }.bind()`},
			{Code: `var a = function() { return 1; }[bind](b)`},
			{Code: `var a = (() => { return b }).bind(c, d)`},

			// --- Not .bind() ---
			{Code: `var a = f.bind(a)`},
			{Code: `f.bind(a)`},
			{Code: `(function() { this.b; }).call(c)`},
			{Code: `(function() { return 1; }).apply(c)`},
			{Code: `var a = function() { this.b }()`},
			{Code: `var a = function() { this.b }.foo()`},
			{Code: `var a = function() { return 1; }`},

			// --- Function uses this directly ---
			{Code: `var a = function() { this.b }.bind(c)`},
			{Code: `var a = function() { return this; }.bind(c)`},
			{Code: `var a = function() { this.b; return 1; }.bind(c)`},

			// --- Arrow captures this from outer scope ---
			{Code: `var a = function() { return () => this; }.bind(b)`},
			{Code: `var a = function() { var f = () => this }.bind(c)`},
			{Code: `var a = function() { var f = () => () => this }.bind(c)`},

			// --- Nested bind where outer uses this ---
			{Code: `(function() { (function() { this.b }.bind(this)) }.bind(c))`},

			// --- This in function + also in class method (direct this counts) ---
			{Code: `var a = function() { this.x; class Foo { bar() { this.y } } }.bind(c)`},

			// --- Computed property name: this belongs to outer scope ---
			{Code: `var a = function() { var o = { [this.key]() {} } }.bind(c)`},
			{Code: `var a = function() { class Foo { [this.key]() {} } }.bind(c)`},
			{Code: `var a = function() { var o = { [this.key]: 1 } }.bind(c)`},
			{Code: `var a = function() { class Foo { x = this.y } }.bind(c)`},
			{Code: `var a = function(a = this) { return a }.bind(c)`},
			{Code: `var a = function() { var o = { get [this.key]() { return 1 } } }.bind(c)`},
			{Code: `var a = function() { var o = { set [this.key](v) {} } }.bind(c)`},

			// --- this in extends clause belongs to outer scope ---
			{Code: `var a = function() { class Foo extends this.Base {} }.bind(c)`},

			// --- this in arrow default parameter inherits outer this ---
			{Code: `var a = function() { var f = (x = this) => x }.bind(c)`},

			// --- Triple nested arrow chain inherits this ---
			{Code: `var a = function() { var f = () => () => () => this }.bind(c)`},

			// --- this in computed property name inside arrow (arrow transparent) ---
			{Code: `var a = function() { var f = () => { class Foo { [this.key]() {} } } }.bind(c)`},

			// --- this in class static initializer block (not a this-scope boundary) ---
			{Code: `var a = function() { class Foo { static { this.x = 1 } } }.bind(c)`},

			// --- this in try/catch/finally (no scope boundary) ---
			{Code: `var a = function() { try {} catch(e) { this.x } }.bind(c)`},
			{Code: `var a = function() { try {} finally { this.x } }.bind(c)`},

			// --- this in extends + static field together ---
			{Code: `var a = function() { class Foo extends this.Base { static x = this } }.bind(c)`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// === Basic: function doesn't use this ===
			{
				Code:   `var a = function() { return 1; }.bind(b)`,
				Output: []string{`var a = function() { return 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { return 1; }.bind(this)`,
				Output: []string{`var a = function() { return 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Arrow function: .bind() is always unnecessary ===
			{
				Code:   `var a = (() => { return 1; }).bind(b)`,
				Output: []string{`var a = (() => { return 1; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = (() => { this.b }).bind(c)`,
				Output: []string{`var a = (() => { this.b })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = (() => { return this; }).bind(b)`,
				Output: []string{`var a = (() => { return this; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = (() => 1).bind(c)`,
				Output: []string{`var a = (() => 1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === this in nested function does NOT count for outer ===
			{
				Code:   `var a = function() { (function(){ this.c }) }.bind(b)`,
				Output: []string{`var a = function() { (function(){ this.c }) }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { function inner() { this.b; } }.bind(c)`,
				Output: []string{`var a = function() { function inner() { this.b; } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { function c(){ this.d } }.bind(b)`,
				Output: []string{`var a = function() { function c(){ this.d } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Bind access patterns ===
			{
				Code:   `var a = function() { return 1; }['bind'](b)`,
				Output: []string{`var a = function() { return 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   "var a = function() { return 1; }[`bind`](b)",
				Output: []string{`var a = function() { return 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = (function() { return 1; }.bind)(this)`,
				Output: []string{`var a = (function() { return 1; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = (function() { return 1; }).bind(this)`,
				Output: []string{`var a = (function() { return 1; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Parenthesized: multiple levels ===
			{
				Code:   `var a = ((function() { return 1; })).bind(c)`,
				Output: []string{`var a = ((function() { return 1; }))`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = ((function() { return 1; }.bind))(c)`,
				Output: []string{`var a = ((function() { return 1; }))`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = ((function() { return 1; }).bind)(c)`,
				Output: []string{`var a = ((function() { return 1; }))`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = (function() { return 1; }['bind'])(c)`,
				Output: []string{`var a = (function() { return 1; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Deeply nested ===
			{
				Code:   `var a = function() { (function(){ (function(){ this.d }.bind(c)) }) }.bind(b)`,
				Output: []string{`var a = function() { (function(){ (function(){ this.d }.bind(c)) }) }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { function a() { function b() { this.x } } }.bind(c)`,
				Output: []string{`var a = function() { function a() { function b() { this.x } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === this in function inside arrow inside bound function ===
			{
				Code:   `var a = function() { var f = () => { return function() { this.x } } }.bind(c)`,
				Output: []string{`var a = function() { var f = () => { return function() { this.x } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// === this in arrow inside inner function ===
			{
				Code:   `var a = function() { function inner() { var f = () => this } }.bind(c)`,
				Output: []string{`var a = function() { function inner() { var f = () => this } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Async / Generator ===
			{
				Code:   `var a = (async function() { return 1; }).bind(c)`,
				Output: []string{`var a = (async function() { return 1; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = (function*() { yield 1; }).bind(c)`,
				Output: []string{`var a = (function*() { yield 1; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = (async () => { return 1; }).bind(c)`,
				Output: []string{`var a = (async () => { return 1; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === this-scoping: class methods/accessors/constructors isolate this ===
			{
				Code:   `var a = function() { class Foo { bar() { this.x } } }.bind(c)`,
				Output: []string{`var a = function() { class Foo { bar() { this.x } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { var o = { foo() { this.x } } }.bind(c)`,
				Output: []string{`var a = function() { var o = { foo() { this.x } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { var o = { get foo() { return this.x } } }.bind(c)`,
				Output: []string{`var a = function() { var o = { get foo() { return this.x } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { var o = { set foo(v) { this.x = v } } }.bind(c)`,
				Output: []string{`var a = function() { var o = { set foo(v) { this.x = v } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { class Foo { constructor() { this.x = 1 } } }.bind(c)`,
				Output: []string{`var a = function() { class Foo { constructor() { this.x = 1 } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() { class Foo { bar() { var f = () => this } } }.bind(c)`,
				Output: []string{`var a = function() { class Foo { bar() { var f = () => this } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === this in default param of inner function ===
			{
				Code:   `var a = function() { function inner(a = this) {} }.bind(c)`,
				Output: []string{`var a = function() { function inner(a = this) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// === this in property value (FunctionExpression isolates) ===
			{
				Code:   `var a = function() { var o = { foo: function() { this.x } } }.bind(c)`,
				Output: []string{`var a = function() { var o = { foo: function() { this.x } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// === this in computed name of nested method ===
			{
				Code:   `var a = function() { var o = { foo() { var p = { [this.key]() {} } } } }.bind(c)`,
				Output: []string{`var a = function() { var o = { foo() { var p = { [this.key]() {} } } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Static method isolates this ===
			{
				Code:   `var a = function() { class Foo { static bar() { this.x } } }.bind(c)`,
				Output: []string{`var a = function() { class Foo { static bar() { this.x } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// === Class expression: method isolates this ===
			{
				Code:   `var a = function() { var C = class { method() { this.x } } }.bind(c)`,
				Output: []string{`var a = function() { var C = class { method() { this.x } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// === Method inside arrow (arrow transparent, method isolates) ===
			{
				Code:   `var a = function() { var f = () => ({ method() { this.x } }) }.bind(c)`,
				Output: []string{`var a = function() { var f = () => ({ method() { this.x } }) }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// === Nested class constructors both isolate ===
			{
				Code:   `var a = function() { class Foo { constructor() { class Bar { constructor() { this.x } } } } }.bind(c)`,
				Output: []string{`var a = function() { class Foo { constructor() { class Bar { constructor() { this.x } } } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// === Deeply nested arrows inside method (all arrows transparent, method isolates) ===
			{
				Code:   `var a = function() { class Foo { bar() { var f = () => { var g = () => this } } } }.bind(c)`,
				Output: []string{`var a = function() { class Foo { bar() { var f = () => { var g = () => this } } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			// === this in computed name of nested class method (belongs to outer method scope) ===
			{
				Code:   `var a = function() { class Outer { method() { class Inner { [this.key]() {} } } } }.bind(c)`,
				Output: []string{`var a = function() { class Outer { method() { class Inner { [this.key]() {} } } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Comment preservation in autofix ===
			{
				Code:   `var a = function() {}/**/.bind(b)`,
				Output: []string{`var a = function() {}/**/`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   "var a = function() {} // comment\n.bind(b)",
				Output: []string{"var a = function() {} // comment\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() {} /* a */ /* b */ .bind(b)`,
				Output: []string{`var a = function() {} /* a */ /* b */ `},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Named function expression ===
			{
				Code:   `var a = (function foo() { return 1; }).bind(c)`,
				Output: []string{`var a = (function foo() { return 1; })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Deep alternating nesting: function -> method -> arrow -> function -> constructor ===
			{
				Code:   `var a = function() { class Foo { method() { var f = () => { var g = function() { class Bar { constructor() { this.x } } } } } } }.bind(c)`,
				Output: []string{`var a = function() { class Foo { method() { var f = () => { var g = function() { class Bar { constructor() { this.x } } } } } } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Multiple bind calls: both reported and fixed ===
			{
				Code:   `var a = function() { var b = function() { return 1 }.bind(d) }.bind(c)`,
				Output: []string{`var a = function() { var b = function() { return 1 } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}, {MessageId: "unexpected"}},
			},

			// === Autofix with null arg ===
			{
				Code:   `var a = function() {}.bind(null)`,
				Output: []string{`var a = function() {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},

			// === Side-effect arguments: no autofix ===
			{
				Code:   `var a = function() {}.bind(b++)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() {}.bind(b())`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() {}.bind(b.c)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   "var a = function() {}.bind(`${b}`)",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() {}.bind([])`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
			{
				Code:   `var a = function() {}.bind({})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected"}},
			},
		},
	)
}
