// TestNoInvalidThisUpstream migrates the full valid/invalid suite from
// upstream ESLint core's
//
//	tests/lib/rules/no-invalid-this.js
//
// (the JS-parser "patterns" half of the suite — each pattern is expanded
// into its NORMAL / "use strict" / module variants per the upstream
// pattern's own valid/invalid mode list; IMPLIED_STRICT is intentionally
// omitted since upstream classifies it identically to "use strict" in
// every pattern, so it adds no coverage). The TypeScript-parser half of
// upstream's suite (ruleTesterTypeScript) is migrated separately in
// no_invalid_this_upstream_ts_test.go.
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

// objectOption produces the array-wrapped option shape that exercises the
// rule's own options-array parsing.
func objectOption(opts map[string]interface{}) []interface{} {
	return []interface{}{opts}
}

func TestNoInvalidThisUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidThisRule,
		[]rule_tester.ValidTestCase{

			// ---- Global ----
			// script
			{Code: `console.log(this); z(x => console.log(x, this));`},
			// use strict
			{Code: `"use strict"; console.log(this); z(x => console.log(x, this));`},
			// script
			{Code: `() => { this }; this;`},
			// use strict
			{Code: `"use strict"; () => { this }; this;`},
			// script
			{Code: `this.eval('foo');`},
			// use strict
			{Code: `"use strict"; this.eval('foo');`},

			// ---- IIFE ----
			// script
			{Code: `(function() { console.log(this); z(x => console.log(x, this)); })();`},

			// ---- Just functions ----
			// script
			{Code: `function foo() { console.log(this); z(x => console.log(x, this)); }`},
			// script: test that the option doesn't reverse the logic and mistakenly allows lowercase functions
			{Code: `function foo() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": false})},
			// script
			{Code: `function Foo() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": false})},
			// script
			{Code: `var foo = (function() { console.log(this); z(x => console.log(x, this)); }).bar(obj);`},

			// ---- Functions in methods ----
			// script
			{Code: `var obj = {foo: function() { function foo() { console.log(this); z(x => console.log(x, this)); } foo(); }};`},
			// script
			{Code: `var obj = {foo() { function foo() { console.log(this); z(x => console.log(x, this)); } foo(); }};`},
			// script
			{Code: `var obj = {foo: function() { return function() { console.log(this); z(x => console.log(x, this)); }; }};`},
			// script
			{Code: `obj.foo = function() { return function() { console.log(this); z(x => console.log(x, this)); }; };`},

			// ---- Class Static methods ----
			// script
			{Code: `class A {static foo() { console.log(this); z(x => console.log(x, this)); }};`},
			// use strict
			{Code: `"use strict"; class A {static foo() { console.log(this); z(x => console.log(x, this)); }};`},
			// module
			{Code: `export {}; class A {static foo() { console.log(this); z(x => console.log(x, this)); }};`},

			// ---- Constructors ----
			// script
			{Code: `function Foo() { console.log(this); z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; function Foo() { console.log(this); z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; function Foo() { console.log(this); z(x => console.log(x, this)); }`},
			// script: test the default value in schema
			{Code: `function Foo() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{})},
			// use strict: test the default value in schema
			{Code: `"use strict"; function Foo() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{})},
			// module: test the default value in schema
			{Code: `export {}; function Foo() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{})},
			// script: test explicitly set option to the default value
			{Code: `function Foo() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": true})},
			// use strict: test explicitly set option to the default value
			{Code: `"use strict"; function Foo() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": true})},
			// module: test explicitly set option to the default value
			{Code: `export {}; function Foo() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": true})},
			// script
			{Code: `var Foo = function Foo() { console.log(this); z(x => console.log(x, this)); };`},
			// use strict
			{Code: `"use strict"; var Foo = function Foo() { console.log(this); z(x => console.log(x, this)); };`},
			// module
			{Code: `export {}; var Foo = function Foo() { console.log(this); z(x => console.log(x, this)); };`},
			// script
			{Code: `class A {constructor() { console.log(this); z(x => console.log(x, this)); }};`},
			// use strict
			{Code: `"use strict"; class A {constructor() { console.log(this); z(x => console.log(x, this)); }};`},
			// module
			{Code: `export {}; class A {constructor() { console.log(this); z(x => console.log(x, this)); }};`},

			// ---- On a property ----
			// script
			{Code: `var obj = {foo: function() { console.log(this); z(x => console.log(x, this)); }};`},
			// use strict
			{Code: `"use strict"; var obj = {foo: function() { console.log(this); z(x => console.log(x, this)); }};`},
			// module
			{Code: `export {}; var obj = {foo: function() { console.log(this); z(x => console.log(x, this)); }};`},
			// script
			{Code: `var obj = {foo() { console.log(this); z(x => console.log(x, this)); }};`},
			// use strict
			{Code: `"use strict"; var obj = {foo() { console.log(this); z(x => console.log(x, this)); }};`},
			// module
			{Code: `export {}; var obj = {foo() { console.log(this); z(x => console.log(x, this)); }};`},
			// script
			{Code: `var obj = {foo: foo || function() { console.log(this); z(x => console.log(x, this)); }};`},
			// use strict
			{Code: `"use strict"; var obj = {foo: foo || function() { console.log(this); z(x => console.log(x, this)); }};`},
			// module
			{Code: `export {}; var obj = {foo: foo || function() { console.log(this); z(x => console.log(x, this)); }};`},
			// script
			{Code: `var obj = {foo: hasNative ? foo : function() { console.log(this); z(x => console.log(x, this)); }};`},
			// use strict
			{Code: `"use strict"; var obj = {foo: hasNative ? foo : function() { console.log(this); z(x => console.log(x, this)); }};`},
			// module
			{Code: `export {}; var obj = {foo: hasNative ? foo : function() { console.log(this); z(x => console.log(x, this)); }};`},
			// script
			{Code: `var obj = {foo: (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })()};`},
			// use strict
			{Code: `"use strict"; var obj = {foo: (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })()};`},
			// module
			{Code: `export {}; var obj = {foo: (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })()};`},
			// script
			{Code: `Object.defineProperty(obj, "foo", {value: function() { console.log(this); z(x => console.log(x, this)); }})`},
			// use strict
			{Code: `"use strict"; Object.defineProperty(obj, "foo", {value: function() { console.log(this); z(x => console.log(x, this)); }})`},
			// module
			{Code: `export {}; Object.defineProperty(obj, "foo", {value: function() { console.log(this); z(x => console.log(x, this)); }})`},
			// script
			{Code: `Object.defineProperties(obj, {foo: {value: function() { console.log(this); z(x => console.log(x, this)); }}})`},
			// use strict
			{Code: `"use strict"; Object.defineProperties(obj, {foo: {value: function() { console.log(this); z(x => console.log(x, this)); }}})`},
			// module
			{Code: `export {}; Object.defineProperties(obj, {foo: {value: function() { console.log(this); z(x => console.log(x, this)); }}})`},

			// ---- Assigns to a property ----
			// script
			{Code: `obj.foo = function() { console.log(this); z(x => console.log(x, this)); };`},
			// use strict
			{Code: `"use strict"; obj.foo = function() { console.log(this); z(x => console.log(x, this)); };`},
			// module
			{Code: `export {}; obj.foo = function() { console.log(this); z(x => console.log(x, this)); };`},
			// script
			{Code: `obj.foo = foo || function() { console.log(this); z(x => console.log(x, this)); };`},
			// use strict
			{Code: `"use strict"; obj.foo = foo || function() { console.log(this); z(x => console.log(x, this)); };`},
			// module
			{Code: `export {}; obj.foo = foo || function() { console.log(this); z(x => console.log(x, this)); };`},
			// script
			{Code: `obj.foo = foo ? bar : function() { console.log(this); z(x => console.log(x, this)); };`},
			// use strict
			{Code: `"use strict"; obj.foo = foo ? bar : function() { console.log(this); z(x => console.log(x, this)); };`},
			// module
			{Code: `export {}; obj.foo = foo ? bar : function() { console.log(this); z(x => console.log(x, this)); };`},
			// script
			{Code: `obj.foo = (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })();`},
			// use strict
			{Code: `"use strict"; obj.foo = (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })();`},
			// module
			{Code: `export {}; obj.foo = (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })();`},
			// script
			{Code: `obj.foo = (() => function() { console.log(this); z(x => console.log(x, this)); })();`},
			// use strict
			{Code: `"use strict"; obj.foo = (() => function() { console.log(this); z(x => console.log(x, this)); })();`},
			// module
			{Code: `export {}; obj.foo = (() => function() { console.log(this); z(x => console.log(x, this)); })();`},
			// script
			{Code: `obj.foo = (function() { return () => { console.log(this); z(x => console.log(x, this)); }; })();`},
			// script
			{Code: `obj.foo = (() => () => { console.log(this); z(x => console.log(x, this)); })();`},
			// use strict
			{Code: `"use strict"; obj.foo = (() => () => { console.log(this); z(x => console.log(x, this)); })();`},
			// script
			{Code: `obj.foo = (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })?.();`},
			// use strict
			{Code: `"use strict"; obj.foo = (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })?.();`},
			// module
			{Code: `export {}; obj.foo = (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })?.();`},

			// ---- Class Instance Methods ----
			// script
			{Code: `class A {foo() { console.log(this); z(x => console.log(x, this)); }};`},
			// use strict
			{Code: `"use strict"; class A {foo() { console.log(this); z(x => console.log(x, this)); }};`},
			// module
			{Code: `export {}; class A {foo() { console.log(this); z(x => console.log(x, this)); }};`},

			// ---- Bind/Call/Apply ----
			// script
			{Code: `var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind(obj);`},
			// use strict
			{Code: `"use strict"; var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind(obj);`},
			// module
			{Code: `export {}; var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind(obj);`},
			// script
			{Code: `var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind(null);`},
			// script
			{Code: `(function() { console.log(this); z(x => console.log(x, this)); }).call(obj);`},
			// use strict
			{Code: `"use strict"; (function() { console.log(this); z(x => console.log(x, this)); }).call(obj);`},
			// module
			{Code: `export {}; (function() { console.log(this); z(x => console.log(x, this)); }).call(obj);`},
			// script
			{Code: `(function() { console.log(this); z(x => console.log(x, this)); }).call(undefined);`},
			// script
			{Code: `(function() { console.log(this); z(x => console.log(x, this)); }).apply(obj);`},
			// use strict
			{Code: `"use strict"; (function() { console.log(this); z(x => console.log(x, this)); }).apply(obj);`},
			// module
			{Code: `export {}; (function() { console.log(this); z(x => console.log(x, this)); }).apply(obj);`},
			// script
			{Code: `(function() { console.log(this); z(x => console.log(x, this)); }).apply(void 0);`},
			// script
			{Code: `Reflect.apply(function() { console.log(this); z(x => console.log(x, this)); }, obj, []);`},
			// use strict
			{Code: `"use strict"; Reflect.apply(function() { console.log(this); z(x => console.log(x, this)); }, obj, []);`},
			// module
			{Code: `export {}; Reflect.apply(function() { console.log(this); z(x => console.log(x, this)); }, obj, []);`},
			// script
			{Code: `var foo = function() { console.log(this); z(x => console.log(x, this)); }?.bind(obj);`},
			// use strict
			{Code: `"use strict"; var foo = function() { console.log(this); z(x => console.log(x, this)); }?.bind(obj);`},
			// module
			{Code: `export {}; var foo = function() { console.log(this); z(x => console.log(x, this)); }?.bind(obj);`},
			// script
			{Code: `var foo = (function() { console.log(this); z(x => console.log(x, this)); }?.bind)(obj);`},
			// use strict
			{Code: `"use strict"; var foo = (function() { console.log(this); z(x => console.log(x, this)); }?.bind)(obj);`},
			// module
			{Code: `export {}; var foo = (function() { console.log(this); z(x => console.log(x, this)); }?.bind)(obj);`},
			// script
			{Code: `var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind?.(obj);`},
			// use strict
			{Code: `"use strict"; var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind?.(obj);`},
			// module
			{Code: `export {}; var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind?.(obj);`},

			// ---- Array methods ----
			// script
			{Code: `Array.from([], function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `Array.fromAsync([], function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.every(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.filter(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.find(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.findIndex(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.findLast(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.findLastIndex(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.flatMap(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.forEach(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.map(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo.some(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `Array.from([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; Array.from([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; Array.from([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `Array.fromAsync([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; Array.fromAsync([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; Array.fromAsync([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.every(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.every(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.every(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.filter(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.filter(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.filter(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.find(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.find(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.find(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.findIndex(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.findIndex(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.findIndex(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.findLast(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.findLast(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.findLast(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.findLastIndex(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.findLastIndex(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.findLastIndex(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.flatMap(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.flatMap(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.flatMap(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.forEach(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.forEach(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.forEach(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.map(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.map(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.map(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.some(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo.some(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo.some(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo.forEach(function() { console.log(this); z(x => console.log(x, this)); }, null);`},
			// script
			{Code: `Array?.from([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; Array?.from([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; Array?.from([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `foo?.every(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; foo?.every(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; foo?.every(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `(Array?.from)([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; (Array?.from)([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; (Array?.from)([], function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// script
			{Code: `(foo?.every)(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// use strict
			{Code: `"use strict"; (foo?.every)(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},
			// module
			{Code: `export {}; (foo?.every)(function() { console.log(this); z(x => console.log(x, this)); }, obj);`},

			// ---- @this tag ----
			// script
			{Code: `/** @this Obj */ function foo() { console.log(this); z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; /** @this Obj */ function foo() { console.log(this); z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; /** @this Obj */ function foo() { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `/**
 * @returns {void}
 * @this Obj
 */
function foo() { console.log(this); z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; /**
 * @returns {void}
 * @this Obj
 */
function foo() { console.log(this); z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; /**
 * @returns {void}
 * @this Obj
 */
function foo() { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `/** @returns {void} */ function foo() { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `/** @this Obj */ foo(function() { console.log(this); z(x => console.log(x, this)); });`},
			// script
			{Code: `foo(/* @this Obj */ function() { console.log(this); z(x => console.log(x, this)); });`},
			// use strict
			{Code: `"use strict"; foo(/* @this Obj */ function() { console.log(this); z(x => console.log(x, this)); });`},
			// module
			{Code: `export {}; foo(/* @this Obj */ function() { console.log(this); z(x => console.log(x, this)); });`},

			// ---- gh#3287 ----
			// script
			{Code: `function foo() { /** @this Obj*/ return function bar() { console.log(this); z(x => console.log(x, this)); }; }`},
			// use strict
			{Code: `"use strict"; function foo() { /** @this Obj*/ return function bar() { console.log(this); z(x => console.log(x, this)); }; }`},
			// module
			{Code: `export {}; function foo() { /** @this Obj*/ return function bar() { console.log(this); z(x => console.log(x, this)); }; }`},

			// ---- gh#6824 ----
			// script
			{Code: `var Ctor = function() { console.log(this); z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; var Ctor = function() { console.log(this); z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; var Ctor = function() { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `var Ctor = function() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": false})},
			// script
			{Code: `var func = function() { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `var func = function() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": false})},
			// script
			{Code: `Ctor = function() { console.log(this); z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; Ctor = function() { console.log(this); z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; Ctor = function() { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `Ctor = function() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": false})},
			// script
			{Code: `func = function() { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `func = function() { console.log(this); z(x => console.log(x, this)); }`, Options: objectOption(map[string]interface{}{"capIsConstructor": false})},
			// script
			{Code: `function foo(Ctor = function() { console.log(this); z(x => console.log(x, this)); }) {}`},
			// use strict
			{Code: `"use strict"; function foo(Ctor = function() { console.log(this); z(x => console.log(x, this)); }) {}`},
			// module
			{Code: `export {}; function foo(Ctor = function() { console.log(this); z(x => console.log(x, this)); }) {}`},
			// script
			{Code: `function foo(func = function() { console.log(this); z(x => console.log(x, this)); }) {}`},
			// script
			{Code: `[obj.method = function() { console.log(this); z(x => console.log(x, this)); }] = a`},
			// use strict
			{Code: `"use strict"; [obj.method = function() { console.log(this); z(x => console.log(x, this)); }] = a`},
			// module
			{Code: `export {}; [obj.method = function() { console.log(this); z(x => console.log(x, this)); }] = a`},
			// script
			{Code: `[func = function() { console.log(this); z(x => console.log(x, this)); }] = a`},

			// ---- Logical assignments ----
			// script
			{Code: `obj.method &&= function () { console.log(this); z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; obj.method &&= function () { console.log(this); z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; obj.method &&= function () { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `obj.method ||= function () { console.log(this); z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; obj.method ||= function () { console.log(this); z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; obj.method ||= function () { console.log(this); z(x => console.log(x, this)); }`},
			// script
			{Code: `obj.method ??= function () { console.log(this); z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; obj.method ??= function () { console.log(this); z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; obj.method ??= function () { console.log(this); z(x => console.log(x, this)); }`},

			// ---- Class fields ----
			// script
			{Code: `class C { field = this }`},
			// use strict
			{Code: `"use strict"; class C { field = this }`},
			// module
			{Code: `export {}; class C { field = this }`},
			// script
			{Code: `class C { static field = this }`},
			// use strict
			{Code: `"use strict"; class C { static field = this }`},
			// module
			{Code: `export {}; class C { static field = this }`},
			// script
			{Code: `class C { field = console.log(this); }`},
			// use strict
			{Code: `"use strict"; class C { field = console.log(this); }`},
			// module
			{Code: `export {}; class C { field = console.log(this); }`},
			// script
			{Code: `class C { field = z(x => console.log(x, this)); }`},
			// use strict
			{Code: `"use strict"; class C { field = z(x => console.log(x, this)); }`},
			// module
			{Code: `export {}; class C { field = z(x => console.log(x, this)); }`},
			// script
			{Code: `class C { field = function () { console.log(this); z(x => console.log(x, this)); }; }`},
			// use strict
			{Code: `"use strict"; class C { field = function () { console.log(this); z(x => console.log(x, this)); }; }`},
			// module
			{Code: `export {}; class C { field = function () { console.log(this); z(x => console.log(x, this)); }; }`},
			// script
			{Code: `class C { #field = function () { console.log(this); z(x => console.log(x, this)); }; }`},
			// use strict
			{Code: `"use strict"; class C { #field = function () { console.log(this); z(x => console.log(x, this)); }; }`},
			// module
			{Code: `export {}; class C { #field = function () { console.log(this); z(x => console.log(x, this)); }; }`},
			// script: `this` is the top-level `this`
			{Code: `class C { [this.foo]; }`},
			// use strict: `this` is the top-level `this`
			{Code: `"use strict"; class C { [this.foo]; }`},
			// script
			{Code: `class C { foo = () => this; }`},
			// use strict
			{Code: `"use strict"; class C { foo = () => this; }`},
			// module
			{Code: `export {}; class C { foo = () => this; }`},
			// script
			{Code: `class C { foo = () => { this }; }`},
			// use strict
			{Code: `"use strict"; class C { foo = () => { this }; }`},
			// module
			{Code: `export {}; class C { foo = () => { this }; }`},

			// ---- Class static blocks ----
			// script
			{Code: `class C { static { this.x; } }`},
			// use strict
			{Code: `"use strict"; class C { static { this.x; } }`},
			// module
			{Code: `export {}; class C { static { this.x; } }`},
			// script
			{Code: `class C { static { () => { this.x; } } }`},
			// use strict
			{Code: `"use strict"; class C { static { () => { this.x; } } }`},
			// module
			{Code: `export {}; class C { static { () => { this.x; } } }`},
			// script
			{Code: `class C { static { class D { [this.x]; } } }`},
			// use strict
			{Code: `"use strict"; class C { static { class D { [this.x]; } } }`},
			// module
			{Code: `export {}; class C { static { class D { [this.x]; } } }`},
			// script
			{Code: `class C { static {} [this]; }`},
			// use strict
			{Code: `"use strict"; class C { static {} [this]; }`},
			// script
			{Code: `class C { static {} [this.x]; }`},
			// use strict
			{Code: `"use strict"; class C { static {} [this.x]; }`},
		},
		[]rule_tester.InvalidTestCase{

			// ---- Global ----
			// module
			{
				Code:   `export {}; console.log(this); z(x => console.log(x, this));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 24}, {MessageId: "unexpectedThis", Line: 1, Column: 53}},
			},
			// module
			{
				Code:   `export {}; () => { this }; this;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 20}, {MessageId: "unexpectedThis", Line: 1, Column: 28}},
			},
			// module
			{
				Code:   `export {}; this.eval('foo');`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 12}},
			},

			// ---- IIFE ----
			// use strict
			{
				Code:   `"use strict"; (function() { console.log(this); z(x => console.log(x, this)); })();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 41}, {MessageId: "unexpectedThis", Line: 1, Column: 70}},
			},
			// module
			{
				Code:   `export {}; (function() { console.log(this); z(x => console.log(x, this)); })();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 38}, {MessageId: "unexpectedThis", Line: 1, Column: 67}},
			},

			// ---- Just functions ----
			// use strict
			{
				Code:   `"use strict"; function foo() { console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 44}, {MessageId: "unexpectedThis", Line: 1, Column: 73}},
			},
			// module
			{
				Code:   `export {}; function foo() { console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 41}, {MessageId: "unexpectedThis", Line: 1, Column: 70}},
			},
			// use strict: test that the option doesn't reverse the logic and mistakenly allows lowercase functions
			{
				Code:    `"use strict"; function foo() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 44}, {MessageId: "unexpectedThis", Line: 1, Column: 73}},
			},
			// module: test that the option doesn't reverse the logic and mistakenly allows lowercase functions
			{
				Code:    `export {}; function foo() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 41}, {MessageId: "unexpectedThis", Line: 1, Column: 70}},
			},
			// use strict
			{
				Code:    `"use strict"; function Foo() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 44}, {MessageId: "unexpectedThis", Line: 1, Column: 73}},
			},
			// module
			{
				Code:    `export {}; function Foo() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 41}, {MessageId: "unexpectedThis", Line: 1, Column: 70}},
			},
			// script
			{
				Code:   `function foo() { "use strict"; console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 44}, {MessageId: "unexpectedThis", Line: 1, Column: 73}},
			},
			// use strict
			{
				Code:   `"use strict"; function foo() { "use strict"; console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 58}, {MessageId: "unexpectedThis", Line: 1, Column: 87}},
			},
			// module
			{
				Code:   `export {}; function foo() { "use strict"; console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 55}, {MessageId: "unexpectedThis", Line: 1, Column: 84}},
			},
			// script
			{
				Code:    `function Foo() { "use strict"; console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 44}, {MessageId: "unexpectedThis", Line: 1, Column: 73}},
			},
			// use strict
			{
				Code:    `"use strict"; function Foo() { "use strict"; console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 58}, {MessageId: "unexpectedThis", Line: 1, Column: 87}},
			},
			// module
			{
				Code:    `export {}; function Foo() { "use strict"; console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 55}, {MessageId: "unexpectedThis", Line: 1, Column: 84}},
			},
			// use strict
			{
				Code:   `"use strict"; var foo = (function() { console.log(this); z(x => console.log(x, this)); }).bar(obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 51}, {MessageId: "unexpectedThis", Line: 1, Column: 80}},
			},
			// module
			{
				Code:   `export {}; var foo = (function() { console.log(this); z(x => console.log(x, this)); }).bar(obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}, {MessageId: "unexpectedThis", Line: 1, Column: 77}},
			},

			// ---- Functions in methods ----
			// use strict
			{
				Code:   `"use strict"; var obj = {foo: function() { function foo() { console.log(this); z(x => console.log(x, this)); } foo(); }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 73}, {MessageId: "unexpectedThis", Line: 1, Column: 102}},
			},
			// module
			{
				Code:   `export {}; var obj = {foo: function() { function foo() { console.log(this); z(x => console.log(x, this)); } foo(); }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 70}, {MessageId: "unexpectedThis", Line: 1, Column: 99}},
			},
			// use strict
			{
				Code:   `"use strict"; var obj = {foo() { function foo() { console.log(this); z(x => console.log(x, this)); } foo(); }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 63}, {MessageId: "unexpectedThis", Line: 1, Column: 92}},
			},
			// module
			{
				Code:   `export {}; var obj = {foo() { function foo() { console.log(this); z(x => console.log(x, this)); } foo(); }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 60}, {MessageId: "unexpectedThis", Line: 1, Column: 89}},
			},
			// use strict
			{
				Code:   `"use strict"; var obj = {foo: function() { return function() { console.log(this); z(x => console.log(x, this)); }; }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 76}, {MessageId: "unexpectedThis", Line: 1, Column: 105}},
			},
			// module
			{
				Code:   `export {}; var obj = {foo: function() { return function() { console.log(this); z(x => console.log(x, this)); }; }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 73}, {MessageId: "unexpectedThis", Line: 1, Column: 102}},
			},
			// script
			{
				Code:   `var obj = {foo: function() { "use strict"; return function() { console.log(this); z(x => console.log(x, this)); }; }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 76}, {MessageId: "unexpectedThis", Line: 1, Column: 105}},
			},
			// use strict
			{
				Code:   `"use strict"; var obj = {foo: function() { "use strict"; return function() { console.log(this); z(x => console.log(x, this)); }; }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 90}, {MessageId: "unexpectedThis", Line: 1, Column: 119}},
			},
			// module
			{
				Code:   `export {}; var obj = {foo: function() { "use strict"; return function() { console.log(this); z(x => console.log(x, this)); }; }};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 87}, {MessageId: "unexpectedThis", Line: 1, Column: 116}},
			},
			// use strict
			{
				Code:   `"use strict"; obj.foo = function() { return function() { console.log(this); z(x => console.log(x, this)); }; };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 70}, {MessageId: "unexpectedThis", Line: 1, Column: 99}},
			},
			// module
			{
				Code:   `export {}; obj.foo = function() { return function() { console.log(this); z(x => console.log(x, this)); }; };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 67}, {MessageId: "unexpectedThis", Line: 1, Column: 96}},
			},
			// script
			{
				Code:   `obj.foo = function() { "use strict"; return function() { console.log(this); z(x => console.log(x, this)); }; };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 70}, {MessageId: "unexpectedThis", Line: 1, Column: 99}},
			},
			// use strict
			{
				Code:   `"use strict"; obj.foo = function() { "use strict"; return function() { console.log(this); z(x => console.log(x, this)); }; };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 84}, {MessageId: "unexpectedThis", Line: 1, Column: 113}},
			},
			// module
			{
				Code:   `export {}; obj.foo = function() { "use strict"; return function() { console.log(this); z(x => console.log(x, this)); }; };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 81}, {MessageId: "unexpectedThis", Line: 1, Column: 110}},
			},
			// script
			{
				Code:   `class A { foo() { return function() { console.log(this); z(x => console.log(x, this)); }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 51}, {MessageId: "unexpectedThis", Line: 1, Column: 80}},
			},
			// use strict
			{
				Code:   `"use strict"; class A { foo() { return function() { console.log(this); z(x => console.log(x, this)); }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 65}, {MessageId: "unexpectedThis", Line: 1, Column: 94}},
			},
			// module
			{
				Code:   `export {}; class A { foo() { return function() { console.log(this); z(x => console.log(x, this)); }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 62}, {MessageId: "unexpectedThis", Line: 1, Column: 91}},
			},

			// ---- Assigns to a property ----
			// use strict
			{
				Code:   `"use strict"; obj.foo = (function() { return () => { console.log(this); z(x => console.log(x, this)); }; })();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 66}, {MessageId: "unexpectedThis", Line: 1, Column: 95}},
			},
			// module
			{
				Code:   `export {}; obj.foo = (function() { return () => { console.log(this); z(x => console.log(x, this)); }; })();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 63}, {MessageId: "unexpectedThis", Line: 1, Column: 92}},
			},
			// module
			{
				Code:   `export {}; obj.foo = (() => () => { console.log(this); z(x => console.log(x, this)); })();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 49}, {MessageId: "unexpectedThis", Line: 1, Column: 78}},
			},

			// ---- Bind/Call/Apply ----
			// use strict
			{
				Code:   `"use strict"; var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind(null);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 50}, {MessageId: "unexpectedThis", Line: 1, Column: 79}},
			},
			// module
			{
				Code:   `export {}; var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind(null);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 47}, {MessageId: "unexpectedThis", Line: 1, Column: 76}},
			},
			// use strict
			{
				Code:   `"use strict"; (function() { console.log(this); z(x => console.log(x, this)); }).call(undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 41}, {MessageId: "unexpectedThis", Line: 1, Column: 70}},
			},
			// module
			{
				Code:   `export {}; (function() { console.log(this); z(x => console.log(x, this)); }).call(undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 38}, {MessageId: "unexpectedThis", Line: 1, Column: 67}},
			},
			// use strict
			{
				Code:   `"use strict"; (function() { console.log(this); z(x => console.log(x, this)); }).apply(void 0);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 41}, {MessageId: "unexpectedThis", Line: 1, Column: 70}},
			},
			// module
			{
				Code:   `export {}; (function() { console.log(this); z(x => console.log(x, this)); }).apply(void 0);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 38}, {MessageId: "unexpectedThis", Line: 1, Column: 67}},
			},

			// ---- Array methods ----
			// use strict
			{
				Code:   `"use strict"; Array.from([], function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 55}, {MessageId: "unexpectedThis", Line: 1, Column: 84}},
			},
			// module
			{
				Code:   `export {}; Array.from([], function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 52}, {MessageId: "unexpectedThis", Line: 1, Column: 81}},
			},
			// use strict
			{
				Code:   `"use strict"; Array.fromAsync([], function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 60}, {MessageId: "unexpectedThis", Line: 1, Column: 89}},
			},
			// module
			{
				Code:   `export {}; Array.fromAsync([], function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 57}, {MessageId: "unexpectedThis", Line: 1, Column: 86}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.every(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 50}, {MessageId: "unexpectedThis", Line: 1, Column: 79}},
			},
			// module
			{
				Code:   `export {}; foo.every(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 47}, {MessageId: "unexpectedThis", Line: 1, Column: 76}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.filter(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 51}, {MessageId: "unexpectedThis", Line: 1, Column: 80}},
			},
			// module
			{
				Code:   `export {}; foo.filter(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}, {MessageId: "unexpectedThis", Line: 1, Column: 77}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.find(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 49}, {MessageId: "unexpectedThis", Line: 1, Column: 78}},
			},
			// module
			{
				Code:   `export {}; foo.find(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 46}, {MessageId: "unexpectedThis", Line: 1, Column: 75}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.findIndex(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 54}, {MessageId: "unexpectedThis", Line: 1, Column: 83}},
			},
			// module
			{
				Code:   `export {}; foo.findIndex(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 51}, {MessageId: "unexpectedThis", Line: 1, Column: 80}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.findLast(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 53}, {MessageId: "unexpectedThis", Line: 1, Column: 82}},
			},
			// module
			{
				Code:   `export {}; foo.findLast(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 50}, {MessageId: "unexpectedThis", Line: 1, Column: 79}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.findLastIndex(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 58}, {MessageId: "unexpectedThis", Line: 1, Column: 87}},
			},
			// module
			{
				Code:   `export {}; foo.findLastIndex(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 55}, {MessageId: "unexpectedThis", Line: 1, Column: 84}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.flatMap(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 52}, {MessageId: "unexpectedThis", Line: 1, Column: 81}},
			},
			// module
			{
				Code:   `export {}; foo.flatMap(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 49}, {MessageId: "unexpectedThis", Line: 1, Column: 78}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.forEach(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 52}, {MessageId: "unexpectedThis", Line: 1, Column: 81}},
			},
			// module
			{
				Code:   `export {}; foo.forEach(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 49}, {MessageId: "unexpectedThis", Line: 1, Column: 78}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.map(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}, {MessageId: "unexpectedThis", Line: 1, Column: 77}},
			},
			// module
			{
				Code:   `export {}; foo.map(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 45}, {MessageId: "unexpectedThis", Line: 1, Column: 74}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.some(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 49}, {MessageId: "unexpectedThis", Line: 1, Column: 78}},
			},
			// module
			{
				Code:   `export {}; foo.some(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 46}, {MessageId: "unexpectedThis", Line: 1, Column: 75}},
			},
			// use strict
			{
				Code:   `"use strict"; foo.forEach(function() { console.log(this); z(x => console.log(x, this)); }, null);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 52}, {MessageId: "unexpectedThis", Line: 1, Column: 81}},
			},
			// module
			{
				Code:   `export {}; foo.forEach(function() { console.log(this); z(x => console.log(x, this)); }, null);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 49}, {MessageId: "unexpectedThis", Line: 1, Column: 78}},
			},

			// ---- @this tag ----
			// use strict
			{
				Code:   `"use strict"; /** @returns {void} */ function foo() { console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 67}, {MessageId: "unexpectedThis", Line: 1, Column: 96}},
			},
			// module
			{
				Code:   `export {}; /** @returns {void} */ function foo() { console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 64}, {MessageId: "unexpectedThis", Line: 1, Column: 93}},
			},
			// use strict
			{
				Code:   `"use strict"; /** @this Obj */ foo(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 61}, {MessageId: "unexpectedThis", Line: 1, Column: 90}},
			},
			// module
			{
				Code:   `export {}; /** @this Obj */ foo(function() { console.log(this); z(x => console.log(x, this)); });`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 58}, {MessageId: "unexpectedThis", Line: 1, Column: 87}},
			},

			// ---- gh#6824 ----
			// use strict
			{
				Code:    `"use strict"; var Ctor = function() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 51}, {MessageId: "unexpectedThis", Line: 1, Column: 80}},
			},
			// module
			{
				Code:    `export {}; var Ctor = function() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}, {MessageId: "unexpectedThis", Line: 1, Column: 77}},
			},
			// use strict
			{
				Code:   `"use strict"; var func = function() { console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 51}, {MessageId: "unexpectedThis", Line: 1, Column: 80}},
			},
			// module
			{
				Code:   `export {}; var func = function() { console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}, {MessageId: "unexpectedThis", Line: 1, Column: 77}},
			},
			// use strict
			{
				Code:    `"use strict"; var func = function() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 51}, {MessageId: "unexpectedThis", Line: 1, Column: 80}},
			},
			// module
			{
				Code:    `export {}; var func = function() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}, {MessageId: "unexpectedThis", Line: 1, Column: 77}},
			},
			// use strict
			{
				Code:    `"use strict"; Ctor = function() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 47}, {MessageId: "unexpectedThis", Line: 1, Column: 76}},
			},
			// module
			{
				Code:    `export {}; Ctor = function() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 44}, {MessageId: "unexpectedThis", Line: 1, Column: 73}},
			},
			// use strict
			{
				Code:   `"use strict"; func = function() { console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 47}, {MessageId: "unexpectedThis", Line: 1, Column: 76}},
			},
			// module
			{
				Code:   `export {}; func = function() { console.log(this); z(x => console.log(x, this)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 44}, {MessageId: "unexpectedThis", Line: 1, Column: 73}},
			},
			// use strict
			{
				Code:    `"use strict"; func = function() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 47}, {MessageId: "unexpectedThis", Line: 1, Column: 76}},
			},
			// module
			{
				Code:    `export {}; func = function() { console.log(this); z(x => console.log(x, this)); }`,
				Options: objectOption(map[string]interface{}{"capIsConstructor": false}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 44}, {MessageId: "unexpectedThis", Line: 1, Column: 73}},
			},
			// use strict
			{
				Code:   `"use strict"; function foo(func = function() { console.log(this); z(x => console.log(x, this)); }) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 60}, {MessageId: "unexpectedThis", Line: 1, Column: 89}},
			},
			// module
			{
				Code:   `export {}; function foo(func = function() { console.log(this); z(x => console.log(x, this)); }) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 57}, {MessageId: "unexpectedThis", Line: 1, Column: 86}},
			},
			// use strict
			{
				Code:   `"use strict"; [func = function() { console.log(this); z(x => console.log(x, this)); }] = a`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}, {MessageId: "unexpectedThis", Line: 1, Column: 77}},
			},
			// module
			{
				Code:   `export {}; [func = function() { console.log(this); z(x => console.log(x, this)); }] = a`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 45}, {MessageId: "unexpectedThis", Line: 1, Column: 74}},
			},

			// ---- Class fields ----
			// module: `this` is the top-level `this`
			{
				Code:   `export {}; class C { [this.foo]; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 23}},
			},

			// ---- Class static blocks ----
			// script
			{
				Code:   `class C { static { function foo() { this.x; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 37}},
			},
			// use strict
			{
				Code:   `"use strict"; class C { static { function foo() { this.x; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 51}},
			},
			// module
			{
				Code:   `export {}; class C { static { function foo() { this.x; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}},
			},
			// script
			{
				Code:   `class C { static { (function() { this.x; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 34}},
			},
			// use strict
			{
				Code:   `"use strict"; class C { static { (function() { this.x; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}},
			},
			// module
			{
				Code:   `export {}; class C { static { (function() { this.x; }); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 45}},
			},
			// script
			{
				Code:   `class C { static { (function() { this.x; })(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 34}},
			},
			// use strict
			{
				Code:   `"use strict"; class C { static { (function() { this.x; })(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 48}},
			},
			// module
			{
				Code:   `export {}; class C { static { (function() { this.x; })(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 45}},
			},
			// module
			{
				Code:   `export {}; class C { static {} [this]; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 33}},
			},
			// module
			{
				Code:   `export {}; class C { static {} [this.x]; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis", Line: 1, Column: 33}},
			},

			// ---- Framework gap: ecmaFeatures.globalReturn ----
			// rslint does not expose parserOptions.ecmaFeatures.globalReturn
			// (Node.js CommonJS-style top-level `return`), so top-level `this`
			// validity cannot be conditioned on it.
			{
				Code:   `console.log(this); z(x => console.log(x, this));`,
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis"}, {MessageId: "unexpectedThis"}},
			},
			{
				Code:   `return function() { console.log(this); z(x => console.log(x, this)); };`,
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis"}, {MessageId: "unexpectedThis"}},
			},

			// ---- Framework gap: ecmaVersion (ES3 "use strict" is inert) ----
			// rslint has no ecmaVersion setting; every "use strict" directive is
			// honored regardless of language-version target.
			{
				Code:   `function foo() { 'use strict'; this.eval(); }`,
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedThis"}},
			},
		},
	)
}
