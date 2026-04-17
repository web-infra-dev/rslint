package accessor_pairs

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestAccessorPairsRule(t *testing.T) {
	bothOpts := map[string]interface{}{"setWithoutGet": true, "getWithoutSet": true}
	setOnlyOpts := map[string]interface{}{"setWithoutGet": true, "getWithoutSet": false}
	getOnlyOpts := map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true}
	noneOpts := map[string]interface{}{"setWithoutGet": false, "getWithoutSet": false}
	classBoth := map[string]interface{}{"setWithoutGet": true, "getWithoutSet": true, "enforceForClassMembers": true}
	classOff := map[string]interface{}{"setWithoutGet": true, "getWithoutSet": true, "enforceForClassMembers": false}
	tsOnly := map[string]interface{}{"enforceForTSTypes": true}
	tsGetAlso := map[string]interface{}{"enforceForTSTypes": true, "getWithoutSet": true}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&AccessorPairsRule,
		[]rule_tester.ValidTestCase{
			// Destructuring — not object literals semantically, not flagged
			{Code: `var { get: foo } = bar; ({ set: foo } = bar);`, Options: bothOpts},
			{Code: `var { set } = foo; ({ get } = foo);`, Options: bothOpts},

			// Default: getWithoutSet=false, so lone getter is valid
			{Code: `var o = { get a() {} }`},
			{Code: `var o = { get a() {} }`, Options: map[string]interface{}{}},

			// No accessors
			{Code: `var o = {};`, Options: bothOpts},
			{Code: `var o = { a: 1 };`, Options: bothOpts},
			{Code: `var o = { a };`, Options: bothOpts},
			{Code: `var o = { a: get };`, Options: bothOpts},
			{Code: `var o = { a: set };`, Options: bothOpts},
			{Code: `var o = { get: function(){} };`, Options: bothOpts},
			{Code: `var o = { set: function(foo){} };`, Options: bothOpts},
			{Code: `var o = { get };`, Options: bothOpts},
			{Code: `var o = { set };`, Options: bothOpts},
			{Code: `var o = { [get]: function() {} };`, Options: bothOpts},
			{Code: `var o = { [set]: function(foo) {} };`, Options: bothOpts},
			{Code: `var o = { get() {} };`, Options: bothOpts},
			{Code: `var o = { set(foo) {} };`, Options: bothOpts},

			// Both options disabled
			{Code: `var o = { get a() {} };`, Options: noneOpts},
			{Code: `var o = { set a(foo) {} };`, Options: noneOpts},
			{Code: `var o = { get a() {} };`, Options: setOnlyOpts},
			{Code: `var o = { set a(foo) {} };`, Options: getOnlyOpts},

			// Valid pairs
			{Code: `var o = { get a() {}, set a(foo) {} };`, Options: bothOpts},
			{Code: `var o = { set a(foo) {}, get a() {} };`, Options: bothOpts},
			{Code: `var o = { get 'a'() {}, set 'a'(foo) {} };`, Options: bothOpts},
			{Code: `var o = { get a() {}, set 'a'(foo) {} };`, Options: bothOpts},
			{Code: `var o = { get ['abc']() {}, set ['abc'](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [1e2]() {}, set 100(foo) {} };`, Options: bothOpts},
			{Code: "var o = { get abc() {}, set [`abc`](foo) {} };", Options: bothOpts},
			{Code: `var o = { get ['123']() {}, set 123(foo) {} };`, Options: bothOpts},

			// Valid pairs with computed expressions (token-level equality)
			{Code: `var o = { get [a]() {}, set [a](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [a]() {}, set [(a)](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [(a)]() {}, set [a](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [a]() {}, set [ a ](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [/*comment*/a/*comment*/]() {}, set [a](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [f()]() {}, set [f()](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [f(a)]() {}, set [f(a)](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [a + b]() {}, set [a + b](foo) {} };`, Options: bothOpts},
			{Code: "var o = { get [`${a}`]() {}, set [`${a}`](foo) {} };", Options: bothOpts},

			// Multiple valid pairs
			{Code: `var o = { get a() {}, set a(foo) {}, get b() {}, set b(bar) {} };`, Options: bothOpts},
			{Code: `var o = { get a() {}, set c(foo) {}, set a(bar) {}, get b() {}, get c() {}, set b(baz) {} };`, Options: bothOpts},

			// Valid pairs mixed with other elements
			{Code: `var o = { get a() {}, set a(foo) {}, b: bar };`, Options: bothOpts},
			{Code: `var o = { get a() {}, b, set a(foo) {} };`, Options: bothOpts},
			{Code: `var o = { get a() {}, ...b, set a(foo) {} };`, Options: bothOpts},
			{Code: `var o = { get a() {}, set a(foo) {}, ...a };`, Options: bothOpts},

			// Duplicate keys — each kind found → valid
			{Code: `var o = { get a() {}, get a() {}, set a(foo) {} };`, Options: bothOpts},
			{Code: `var o = { get a() {}, set a(foo) {}, get a() {} };`, Options: bothOpts},
			{Code: `var o = { get a() {}, get a() {} };`, Options: setOnlyOpts},
			{Code: `var o = { set a(foo) {}, set a(foo) {} };`, Options: getOnlyOpts},
			{Code: `var o = { get a() {}, set a(foo) {}, a };`, Options: bothOpts},
			{Code: `var o = { a, get a() {}, set a(foo) {} };`, Options: bothOpts},

			// Property descriptors — complete pairs or non-descriptor calls
			{Code: "var o = {a: 1};\n Object.defineProperty(o, 'b', \n{set: function(value) {\n val = value; \n},\n get: function() {\n return val; \n} \n});"},
			{Code: `var o = {set: function() {}}`},
			{Code: `Object.defineProperties(obj, {set: {value: function() {}}});`},
			{Code: `Object.create(null, {set: {value: function() {}}});`},
			{Code: `var o = {get: function() {}}`, Options: map[string]interface{}{"getWithoutSet": true}},
			{Code: `var o = {[set]: function() {}}`},
			{Code: `var set = 'value'; Object.defineProperty(obj, 'foo', {[set]: function(value) {}});`},

			// Classes — default (no errors without enforceForClassMembers or with options off)
			{Code: `class A { get a() {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class A { get #a() {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}},
			{Code: `class A { set a(foo) {} }`, Options: map[string]interface{}{"enforceForClassMembers": false}},
			{Code: `class A { get a() {} set b(foo) {} static get c() {} static set d(bar) {} }`, Options: classOff},
			{Code: `(class A { get a() {} set b(foo) {} static get c() {} static set d(bar) {} });`, Options: classOff},

			// Class — individual options disabled
			{Code: `class A { get a() {} }`, Options: map[string]interface{}{"setWithoutGet": true, "getWithoutSet": false, "enforceForClassMembers": true}},
			{Code: `class A { set a(foo) {} }`, Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true}},
			{Code: `class A { static get a() {} }`, Options: map[string]interface{}{"setWithoutGet": true, "getWithoutSet": false, "enforceForClassMembers": true}},
			{Code: `class A { static set a(foo) {} }`, Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true}},
			{Code: `A = class { set a(foo) {} };`, Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true}},
			{Code: `class A { get a() {} set b(foo) {} static get c() {} static set d(bar) {} }`, Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": false, "enforceForClassMembers": true}},

			// Class — no accessors
			{Code: `class A {}`, Options: classBoth},
			{Code: `(class {})`, Options: classBoth},
			{Code: `class A { constructor () {} }`, Options: classBoth},
			{Code: `class A { a() {} }`, Options: classBoth},
			{Code: `class A { static a() {} 'b'() {} }`, Options: classBoth},
			{Code: `class A { [a]() {} }`, Options: classBoth},
			{Code: `A = class { a() {} static a() {} b() {} static c() {} }`, Options: classBoth},

			// Class — valid pairs
			{Code: `class A { get a() {} set a(foo) {} }`, Options: classBoth},
			{Code: `class A { set a(foo) {} get a() {} }`, Options: classBoth},
			{Code: `class A { static get a() {} static set a(foo) {} }`, Options: classBoth},
			{Code: `class A { static set a(foo) {} static get a() {} }`, Options: classBoth},
			{Code: `(class { set a(foo) {} get a() {} });`, Options: classBoth},
			{Code: `class A { get 'a'() {} set ['a'](foo) {} }`, Options: classBoth},
			{Code: "class A { set [`a`](foo) {} get a() {} }", Options: classBoth},
			{Code: `class A { get 'a'() {} set a(foo) {} }`, Options: classBoth},
			{Code: `A = class { static get 1e2() {} static set [100](foo) {} };`, Options: classBoth},
			{Code: `class A { get [a]() {} set [a](foo) {} }`, Options: classBoth},
			{Code: `A = class { set [(f())](foo) {} get [(f())]() {} };`, Options: classBoth},
			{Code: `class A { static set [f(a)](foo) {} static get [f(a)]() {} }`, Options: classBoth},

			// Multiple valid pairs in same class
			{Code: `class A { get a() {} set b(foo) {} set a(bar) {} get b() {} }`, Options: classBoth},
			{Code: `class A { get a() {} set a(bar) {} b() {} set c(foo) {} get c() {} }`, Options: classBoth},
			{Code: `(class { get a() {} static set a(foo) {} set a(bar) {} static get a() {} });`, Options: classBoth},

			// Class — valid mixed with other members
			{Code: `class A { get a() {} b() {} set a(foo) {} }`, Options: classBoth},
			{Code: `class A { set a(foo) {} get a() {} b() {} }`, Options: classBoth},
			{Code: `class A { a() {} get b() {} c() {} set b(foo) {} d() {} }`, Options: classBoth},
			{Code: `class A { get a() {} set a(foo) {} static a() {} }`, Options: classBoth},
			{Code: `A = class { static get a() {} static b() {} static set a(foo) {} };`, Options: classBoth},
			{Code: `A = class { static set a(foo) {} static get a() {} a() {} };`, Options: classBoth},

			// Class — duplicate keys
			{Code: `class A { get a() {} get a() {} set a(foo) {} }`, Options: classBoth},
			{Code: `class A { get [a]() {} set [a](foo) {} set [a](foo) {} }`, Options: classBoth},
			{Code: "class A { get a() {} set 'a'(foo) {} get [`a`]() {} }", Options: classBoth},
			{Code: `A = class { get a() {} set a(foo) {} a() {} }`, Options: classBoth},
			{Code: `A = class { a() {} get a() {} set a(foo) {} }`, Options: classBoth},
			{Code: `class A { static set a(foo) {} static set a(foo) {} static get a() {} }`, Options: classBoth},
			{Code: `class A { static get a() {} static set a(foo) {} static get a() {} }`, Options: classBoth},
			{Code: `class A { static set a(foo) {} static get a() {} static a() {} }`, Options: classBoth},
			{Code: `class A { get a() {} a() {} set a(foo) {} }`, Options: classBoth},
			{Code: `class A { static set a(foo) {} static a() {} static get a() {} }`, Options: classBoth},

			// TS — default off, lone setter/getter in types is valid
			{Code: `interface I { get prop(): any }`},
			{Code: `type T = { set prop(value: any): void }`},
			{Code: `interface I { get prop(): any, set prop(value: any): void }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `type T = { get prop(): any, set prop(value: any): void }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `interface I { get prop(): any, between: true, set prop(value: any): void }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `interface I { set prop(value: any): void, get prop(): any }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `interface I { set prop(value: any): void, get 'prop'(): any }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `interface I {}`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `interface I { (...args): void }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `interface I { new(...args): unknown }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `interface I { prop: () => any }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `interface I { method(): any }`, Options: map[string]interface{}{"enforceForTSTypes": true}},
			{Code: `type T = { get prop(): any }`, Options: map[string]interface{}{"enforceForTSTypes": true}},

			// ---- Structural matching of computed keys ----

			// Whitespace inside the key expression is not part of the AST → match.
			{Code: `var o = { get [a+b]() {}, set [a + b](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [a  +  b]() {}, set [a+b](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [f ( a , b )]() {}, set [f(a,b)](foo) {} };`, Options: bothOpts},
			// Comments inside the key are skipped too.
			{Code: `var o = { get [/*x*/ a /*y*/ + b]() {}, set [a + b](foo) {} };`, Options: bothOpts},
			// Nested parens at every level.
			{Code: `var o = { get [((a + b))]() {}, set [a + b](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [(a) + (b)]() {}, set [a + b](foo) {} };`, Options: bothOpts},
			// Property / element access with same static name.
			{Code: `var o = { get [a.b]() {}, set [a.b](foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [a[0]]() {}, set [a[0]](foo) {} };`, Options: bothOpts},

			// ---- Numeric / bigint normalization on static keys ----

			{Code: `var o = { get 0x10() {}, set 16(foo) {} };`, Options: bothOpts},
			{Code: `var o = { get [0x10]() {}, set 16(foo) {} };`, Options: bothOpts},
			{Code: `var o = { get 1.0() {}, set 1(foo) {} };`, Options: bothOpts},
			// Normalized against an identifier name "Infinity" — NOT expected to match
			// (identifier vs numeric overflow is a subtle edge case and ESLint doesn't
			// guarantee it either, so we don't test the cross-form here).

			// ---- Private identifier forms their own equivalence class ----

			// `#a` and `'#a'` are separate keys — each forms its own unmatched group.
			// See invalid section for the negative assertion.
			{Code: `class A { get #a() {} set #a(foo) {} }`, Options: classBoth},
			{Code: `class A { get '#a'() {} set '#a'(foo) {} }`, Options: classBoth},

			// ---- Class: static and instance of the same name are independent pairs ----

			{Code: `class A { get a() {} set a(v) {} static get a() {} static set a(v) {} }`, Options: classBoth},

			// ---- Nested object literals & classes are processed independently ----

			{Code: `const x = { outer: { get a() {}, set a(v) {} } };`, Options: bothOpts},
			{Code: `class A { method() { return { get a() {}, set a(v) {} }; } }`, Options: classBoth},
			{Code: `class A { method() { return class B { get a() {} set a(v) {} }; } }`, Options: classBoth},

			// ---- Descriptor calls are NOT treated as descriptors in nested positions ----

			// `{set: ...}` is inside a conditional → not arg[2] of defineProperty.
			// The outer call is unrecognized but the inner object still exists as a
			// plain object literal (no accessor kinds → no report).
			{Code: `Object.defineProperty(o, 'k', cond ? {set: function(){}} : {});`, Options: bothOpts},
			// Wrapped in another call — not a direct arg of defineProperty.
			{Code: `Object.defineProperty(o, 'k', wrap({set: function(){}}));`, Options: bothOpts},
			// `{set: ...}` is the VALUE of a property assignment whose grandparent is
			// NOT a descriptor-map argument.
			{Code: `const x = { bag: { set: function(){} } };`, Options: bothOpts},

			// ---- Runtime descriptors passed via variables are opaque to the rule ----

			{Code: `const d = {set: function(v){}}; Object.defineProperty(o, 'k', d);`, Options: bothOpts},

			// ---- Shadowed Object shouldn't affect syntactic recognition ----
			// (We mirror ESLint: text-only match, no scope resolution. Documented.)
			{Code: `{ const Object = null; Object.defineProperty(o, 'k', {get() {}, set(v) {}}); }`, Options: bothOpts},

			// ---- Reserved-word property names ----

			// Bare reserved words land on Identifier → static name works.
			{Code: `var o = { get null() {}, set null(v) {} };`, Options: bothOpts},
			{Code: `var o = { get true() {}, set true(v) {} };`, Options: bothOpts},
			{Code: `var o = { get undefined() {}, set undefined(v) {} };`, Options: bothOpts},
			// [null] / [true] / [false] in computed form collapse to the same static name.
			{Code: `var o = { get null() {}, set [null](v) {} };`, Options: bothOpts},
			{Code: `var o = { get true() {}, set [true](v) {} };`, Options: bothOpts},
			{Code: `var o = { get false() {}, set [false](v) {} };`, Options: bothOpts},

			// ---- Descriptor: method shorthand / shorthand property / mix ----

			// Method shorthand forms a complete descriptor.
			{Code: `Object.defineProperty(o, 'k', {get() { return 1; }, set(v) {}});`, Options: bothOpts},
			// Shorthand property — `set` is a variable that holds the function.
			{Code: `var get, set; Object.defineProperty(o, 'k', {get, set});`, Options: bothOpts},
			// Mixed: one property-assignment + one method-shorthand.
			{Code: `Object.defineProperty(o, 'k', {get: function() {}, set(v) {}});`, Options: bothOpts},

			// ---- Bracket caller form ----

			// Well-known callers via element access with static string.
			{Code: "Reflect['defineProperty'](o, 'k', {get: function() {}, set: function(v) {}});", Options: bothOpts},
			{Code: "Object['defineProperties'](o, {k: {get: function() {}, set: function(v) {}}});", Options: bothOpts},
			{Code: "Object['create'](null, {k: {get: function() {}, set: function(v) {}}});", Options: bothOpts},
			// Optional chain + bracket: `Reflect?.['defineProperty']`.
			{Code: "Reflect?.['defineProperty'](o, 'k', {get: function() {}, set: function(v) {}});", Options: bothOpts},

			// ---- Regex computed key cross-form with string key ----

			// `[/a/]` and string `'/a/'` both normalize to static name "/a/".
			{Code: `var o = { get '/a/'() {}, set [/a/](v) {} };`, Options: bothOpts},

		},
		[]rule_tester.InvalidTestCase{
			// Default — setter without getter
			{
				Code: `var o = { set a(value) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11},
				},
			},
			{
				Code:    `var o = { set a(value) {} };`,
				Options: setOnlyOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11},
				},
			},
			{
				Code:    `var o = { set a(value) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11},
				},
			},
			{
				Code:    `var o = { get a() {} };`,
				Options: getOnlyOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
				},
			},
			{
				Code:    `var o = { get a() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
				},
			},
			{
				Code:    `var o = { get a() {} };`,
				Options: map[string]interface{}{"getWithoutSet": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
				},
			},

			// Various getter keys
			{
				Code:    `var o = { get abc() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			{
				Code:    `var o = { get 'abc'() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			{
				Code:    `var o = { get 123() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			{
				Code:    `var o = { get 1e2() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			{
				Code:    `var o = { get ['abc']() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			{
				Code:    "var o = { get [`abc`]() {} };",
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			{
				Code:    `var o = { get [abc]() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			{
				Code:    `var o = { get [f(abc)]() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			{
				Code:    `var o = { get [a + b]() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},

			// Various setter keys
			{
				Code:    `var o = { set [a + b](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},

			// Different keys
			{
				Code:    `var o = { get a() {}, set b(foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 23},
				},
			},
			{
				Code:    `var o = { set a(foo) {}, get b() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 26},
				},
			},
			{
				Code:    `var o = { get 1() {}, set b(foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 23},
				},
			},
			{
				Code:    `var o = { get [a]() {}, set [b](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 25},
				},
			},
			{
				Code:    `var o = { get [a]() {}, set a(foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 25},
				},
			},
			{
				Code:    `var o = { get a() {}, set [a](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 23},
				},
			},
			{
				Code:    `var o = { get [a + b]() {}, set [a - b](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 29},
				},
			},

			// Multiple invalid of same and different kinds
			{
				Code:    `var o = { get a() {}, get b() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 23},
				},
			},
			{
				Code:    `var o = { set a(foo) {}, set b(bar) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 26},
				},
			},

			// Separate object literals
			{
				Code:    `var o1 = { get a() {} }, o2 = { set a(foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 12},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 33},
				},
			},

			// Duplicate keys that remain unpaired
			{
				Code:    `var o = { get a() {}, get a() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 23},
				},
			},
			{
				Code:    `var o = { set a(foo) {}, set a(foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 26},
				},
			},

			// With spread / other properties
			{
				Code:    `var o = { a, get b() {}, c };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 14},
				},
			},
			{
				Code:    `var o = { get a() {}, ...b };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
				},
			},

			// Full location tests (Line + Column + EndLine + EndColumn)
			{
				Code:    `var o = { get b() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "var o = {\n  set [\n a](foo) {} };",
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 2, Column: 3, EndLine: 3, EndColumn: 4},
				},
			},

			// Property descriptors
			{
				Code: "var o = {d: 1};\n Object.defineProperty(o, 'c', \n{set: function(value) {\n val = value; \n} \n});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			{
				Code: `Reflect.defineProperty(obj, 'foo', {set: function(value) {}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			{
				Code: `Object.defineProperties(obj, {foo: {set: function(value) {}}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			{
				Code: `Object.create(null, {foo: {set: function(value) {}}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			// Optional chaining
			{
				Code: "var o = {d: 1};\n Object?.defineProperty(o, 'c', \n{set: function(value) {\n val = value; \n} \n});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			{
				Code: `Reflect?.defineProperty(obj, 'foo', {set: function(value) {}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			{
				Code: `(Object?.defineProperty)(o, 'c', {set: function(value) {}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},

			// Classes — default settings
			{
				Code: `class A { set a(foo) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass"},
				},
			},
			{
				Code:    `class A { get a() {} set b(foo) {} }`,
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass"},
				},
			},
			{
				Code:    `class A { get a() {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass"},
				},
			},
			{
				Code:    `class A { static get a() {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass"},
				},
			},
			{
				Code:    `class A { static set a(foo) {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass"},
				},
			},
			{
				Code:    `A = class { get a() {} set b(foo) {} };`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass"},
					{MessageId: "missingGetterInClass"},
				},
			},

			// Class — private identifiers
			{
				Code: `class A { set #a(foo) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass"},
				},
			},
			{
				Code: `class A { static set #a(foo) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass"},
				},
			},

			// Class — different keys with column positions
			{
				Code:    `class A { get a() {} set b(foo) {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 22},
				},
			},
			{
				Code:    `A = class { static get a() {} static set b(foo) {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 13},
					{MessageId: "missingGetterInClass", Line: 1, Column: 31},
				},
			},

			// Class — prototype and static with same key
			{
				Code:    `class A { get a() {} static set a(foo) {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 22},
				},
			},

			// Class — duplicates
			{
				Code:    `class A { get a() {} get a() {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingSetterInClass", Line: 1, Column: 22},
				},
			},
			{
				Code:    `class A { set a(foo) {} set a(foo) {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 25},
				},
			},

			// Class — full location
			{
				Code:    `class A { get a() {} };`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "A = class {\n  set [\n a](foo) {} };",
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 2, Column: 3, EndLine: 3, EndColumn: 4},
				},
			},
			{
				Code:    `class A { static get b() {} };`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11, EndLine: 1, EndColumn: 23},
				},
			},

			// Class — mixed ESLint cases
			{
				Code:    `class A { get a() {} } class B { set a(foo) {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 34},
				},
			},

			// TS types
			{
				Code: `({ set prop(value) {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},
			{
				Code:    `interface I { set prop(value: any): any }`,
				Options: tsOnly,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInType"},
				},
			},
			{
				Code:    `interface I { set prop(value: any): any, get other(): any }`,
				Options: tsOnly,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInType"},
				},
			},
			{
				Code:    `interface I { set prop(value: any): any, prop(): any }`,
				Options: tsOnly,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInType"},
				},
			},
			{
				Code:    `interface I { get prop(): any } interface J { set prop(value: any): void }`,
				Options: tsOnly,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInType"},
				},
			},
			{
				Code:    `type T = { set prop(value: any): any }`,
				Options: tsOnly,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInType"},
				},
			},
			{
				Code:    `function fn(): { set prop(value: any): any } { return null as any }`,
				Options: tsOnly,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInType"},
				},
			},
			{
				Code:    `type T = { get prop(): any }`,
				Options: tsGetAlso,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInType"},
				},
			},

			// ---- Structural mismatches (computed keys) ----

			// Same operands, different operator.
			{
				Code:    `var o = { get [a + b]() {}, set [a - b](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 29},
				},
			},
			// Optional chain vs regular access — structurally different.
			{
				Code:    `var o = { get [a.b]() {}, set [a?.b](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},
			// PropertyAccess vs ElementAccess — different Kind.
			{
				Code:    `var o = { get [a.b]() {}, set [a['b']](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},
			// Template tail text differs.
			{
				Code:    "var o = { get [`${a}`]() {}, set [`${a} `](foo) {} };",
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},
			// Raw numeric literal form differs (`0x1` vs `1`) — ESLint does a
			// token-level compare, and we route numeric literals through
			// scanner.GetSourceTextOfNodeFromSourceFile so the original source
			// form is preserved despite tsgo's parse-time normalization.
			{
				Code:    `var o = { get [a + 0x1]() {}, set [a + 1](v) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},
			// Arg count differs.
			{
				Code:    `var o = { get [f(a)]() {}, set [f(a, b)](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},

			// ---- Static vs dynamic/private cross-form (never match) ----

			// Static `a` vs dynamic `[a.b]` — different kinds.
			{
				Code:    `var o = { get a() {}, set [a.b](foo) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},
			// Private `#a` vs string `'#a'` — separate equivalence classes.
			{
				Code:    `class A { get #a() {} set '#a'(foo) {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 23},
				},
			},
			{
				Code:    `class A { get '#a'() {} set #a(foo) {} }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 25},
				},
			},

			// ---- Numeric normalization still matches across forms ----

			// Hex literal key paired with decimal identifier key.
			{
				Code:    `var o = { get 0x10() {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					// Key normalizes to "16" — report for missing setter on "16".
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},

			// ---- Nested ObjectLiteralExpression is checked independently ----

			{
				Code:    `const x = { outer: { get a() {} } };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
				},
			},
			// Outer class has a pair; nested class has only a setter.
			{
				Code:    `class Outer { get a() {} set a(v) {} method() { return class Inner { set a(v) {} }; } }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass"},
				},
			},
			// Object literal inside class method body.
			{
				Code:    `class A { m() { return { set a(v) {} }; } }`,
				Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},

			// ---- Descriptor-shape recognition ----

			// Spread in descriptor: treated syntactically, only named properties are
			// considered — so {...d, set: ...} still counts as "set without get".
			{
				Code:    `Object.defineProperty(o, 'k', {...d, set: function(v) {}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			// Doubly-nested descriptor maps: only the inner {set:...} is a descriptor.
			{
				Code:    `Object.defineProperties(obj, {foo: {set: function(v){}}, bar: {get: function(){}}});`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
					{MessageId: "missingSetterInPropertyDescriptor"},
				},
			},
			// Parenthesized Object.defineProperties callee with optional chain.
			{
				Code:    `(Object?.defineProperties)(obj, {foo: {set: function(v){}}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},

			// ---- Descriptor: method shorthand only reports correctly ----

			{
				Code: `Object.defineProperty(o, 'k', {set(v) {}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			{
				Code:    `Object.defineProperty(o, 'k', {get() { return 1; }});`,
				Options: map[string]interface{}{"setWithoutGet": true, "getWithoutSet": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInPropertyDescriptor"},
				},
			},
			// Shorthand property — only `set` is declared.
			{
				Code: `var set; Object.defineProperty(o, 'k', {set});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},

			// ---- Bracket caller form (reports) ----

			{
				Code: "Reflect['defineProperty'](o, 'k', {set: function(v) {}});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},
			{
				Code: "Object['defineProperties'](o, {k: {set: function(v) {}}});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor"},
				},
			},

			// ---- Reserved-word computed key mismatch ----

			// `[true]` vs `[false]` → different static names, both flagged.
			{
				Code:    `var o = { get [true]() {}, set [false](v) {} };`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral"},
					{MessageId: "missingGetterInObjectLiteral"},
				},
			},

			// =====================================================
			// Full parity pass with ESLint's test suite (rest of cases).
			// Grouped in the same order as lib/tests/lib/rules/accessor-pairs.js.
			// =====================================================

			// ---- Object literals: various setter keys ----

			{Code: `var o = { set abc(foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: `var o = { set 'abc'(foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: `var o = { set 123(foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: `var o = { set 1e2(foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: `var o = { set ['abc'](foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: "var o = { set [`abc`](foo) {} };", Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: `var o = { set [123](foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: `var o = { set [abc](foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: `var o = { set [f(abc)](foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral"}}},
			{Code: `var o = { get [123]() {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInObjectLiteral"}}},

			// ---- Different keys with numeric / whitespace / empty / null ----

			{
				Code: `var o = { get a() {}, set 1(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 23},
				},
			},
			{
				Code: `var o = { get a() {}, set 'a '(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 23},
				},
			},
			{
				Code: `var o = { get ' a'() {}, set 'a'(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 26},
				},
			},
			{
				Code: `var o = { get ''() {}, set ' '(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 24},
				},
			},
			{
				Code: `var o = { get ''() {}, set null(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 24},
				},
			},
			{
				Code: "var o = { get [`a`]() {}, set b(foo) {} };", Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 27},
				},
			},
			{
				Code: "var o = { get [`${0} `]() {}, set [`${0}`](foo) {} };", Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 31},
				},
			},

			// ---- Four-error combination ----

			{
				Code: `var o = { get a() {}, set b(foo) {}, set c(foo) {}, get d() {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 23},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 38},
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 53},
				},
			},

			// ---- Per object literal (reverse) ----

			{
				Code: `var o1 = { set a(foo) {} }, o2 = { get a() {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 12},
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 36},
				},
			},

			// ---- Valid+invalid combinations ----

			{
				Code: `var o = { get a() {}, get b() {}, set b(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11}},
			},
			{
				Code: `var o = { get b() {}, get a() {}, set b(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 23}},
			},
			{
				Code: `var o = { get b() {}, set b(foo) {}, get a() {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 38}},
			},
			{
				Code: `var o = { set a(foo) {}, get b() {}, set b(bar) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11}},
			},
			{
				Code: `var o = { get b() {}, set a(foo) {}, set b(bar) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 23}},
			},
			{
				Code: `var o = { get b() {}, set b(bar) {}, set a(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 38}},
			},
			{
				Code: `var o = { get v1() {}, set i1(foo) {}, get v2() {}, set v2(bar) {}, get i2() {}, set v1(baz) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 24},
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 69},
				},
			},

			// ---- Other elements don't affect ----

			{
				Code: `var o = { a, get b() {}, c, set d(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 14},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 29},
				},
			},
			{Code: `var o = { get a() {}, a:1 };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11}}},
			{Code: `var o = { a, get a() {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 14}}},
			{Code: `var o = { set a(foo) {}, a:1 };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11}}},
			{Code: `var o = { a, set a(foo) {} };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 14}}},
			{Code: `var o = { get a() {}, ...a };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 11}}},
			{Code: `var o = { set a(foo) {}, ...a };`, Options: bothOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 11}}},

			// ---- Property descriptors: full optional-chain + parenthesized matrix ----

			{Code: `Object?.defineProperties(obj, {foo: {set: function(v) {}}});`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInPropertyDescriptor"}}},
			{Code: `Object?.create(null, {foo: {set: function(v) {}}});`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInPropertyDescriptor"}}},
			{Code: `(Reflect?.defineProperty)(obj, 'foo', {set: function(v) {}});`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInPropertyDescriptor"}}},
			{Code: `(Object?.create)(null, {foo: {set: function(v) {}}});`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInPropertyDescriptor"}}},

			// ---- Classes: default / enforceForClassMembers:true echo ----

			{Code: `class A { set a(foo) {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `class A { set a(foo) {} }`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `A = class { get a() {} };`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},
			{Code: `class A { set a(value) {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `class A { static set a(value) {} }`, Options: map[string]interface{}{"enforceForClassMembers": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `A = class { set a(value) {} };`, Options: map[string]interface{}{"enforceForClassMembers": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `(class A { static set a(value) {} });`, Options: map[string]interface{}{"enforceForClassMembers": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},

			// ---- Class private-identifier forms ----

			{Code: `class A { set '#a'(foo) {} }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `class A { static set '#a'(foo) {} }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},

			// ---- Class: options don't affect each other ----

			{
				Code:    `class A { set a(value) {} }`,
				Options: map[string]interface{}{"setWithoutGet": true, "getWithoutSet": false, "enforceForClassMembers": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}},
			},
			{
				Code:    `A = class { static set a(value) {} };`,
				Options: classBoth,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}},
			},
			{
				Code:    `let foo = class A { get a() {} };`,
				Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}},
			},
			{
				Code:    `class A { static get a() {} };`,
				Options: classBoth,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}},
			},
			{
				Code:    `(class { get a() {} });`,
				Options: map[string]interface{}{"getWithoutSet": true, "enforceForClassMembers": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}},
			},
			{
				Code:    `class A { get '#a'() {} };`,
				Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}},
			},
			{
				Code:    `class A { get #a() {} };`,
				Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}},
			},
			{
				Code:    `class A { static get '#a'() {} };`,
				Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}},
			},
			{
				Code:    `class A { static get #a() {} };`,
				Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}},
			},

			// ---- Class: various kinds of keys ----

			{Code: `class A { get abc() {} }`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},
			{Code: `A = class { static set 'abc'(foo) {} };`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `(class { get 123() {} });`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},
			{Code: `class A { static get 1e2() {} }`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},
			{Code: `A = class { get ['abc']() {} };`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},
			{Code: "class A { set [`abc`](foo) {} }", Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `class A { static get [123]() {} }`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},
			{Code: `class A { get [abc]() {} }`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},
			{Code: `class A { static get [f(abc)]() {} }`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},
			{Code: `A = class { set [a + b](foo) {} };`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass"}}},
			{Code: `class A { get ['constructor']() {} }`, Options: classBoth, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass"}}},

			// ---- Class: different keys (names & positions) ----

			{
				Code: `A = class { set a(foo) {} get b() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 13},
					{MessageId: "missingSetterInClass", Line: 1, Column: 27},
				},
			},
			{
				Code: `class A { get a() {} set b(foo) {} }`, Options: map[string]interface{}{"setWithoutGet": false, "getWithoutSet": true, "enforceForClassMembers": true},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 11}},
			},
			{
				Code: `class A { get a() {} set b(foo) {} }`, Options: map[string]interface{}{"setWithoutGet": true, "getWithoutSet": false, "enforceForClassMembers": true},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass", Line: 1, Column: 22}},
			},
			{
				Code: `class A { get 'a '() {} set 'a'(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 25},
				},
			},
			{
				Code: `class A { get 'a'() {} set 1(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 24},
				},
			},
			{
				Code: `class A { get 1() {} set 2(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 22},
				},
			},
			{
				Code: `class A { get ''() {} set null(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 23},
				},
			},
			{
				Code: `class A { get a() {} set [a](foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 22},
				},
			},
			{
				Code: `class A { get [a]() {} set [b](foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 24},
				},
			},
			{
				Code: `class A { get [a]() {} set [a++](foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 24},
				},
			},
			{
				Code: `class A { get [a + b]() {} set [a - b](foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 28},
				},
			},

			// ---- Class: prototype vs static with same key ----

			{
				Code: `A = class { static get a() {} set a(foo) {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 13},
					{MessageId: "missingGetterInClass", Line: 1, Column: 31},
				},
			},
			{
				Code: `class A { set [a](foo) {} static get [a]() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 11},
					{MessageId: "missingSetterInClass", Line: 1, Column: 27},
				},
			},
			{
				Code: `class A { static set [a](foo) {} get [a]() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 11},
					{MessageId: "missingSetterInClass", Line: 1, Column: 34},
				},
			},

			// ---- Class: multiple invalid ----

			{
				Code: `A = class { get a() {} get [b]() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 13},
					{MessageId: "missingSetterInClass", Line: 1, Column: 24},
				},
			},
			{
				Code: `class A { get [a]() {} get [b]() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingSetterInClass", Line: 1, Column: 24},
				},
			},
			{
				Code: `A = class { set a(foo) {} set b(bar) {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 13},
					{MessageId: "missingGetterInClass", Line: 1, Column: 27},
				},
			},
			{
				Code: `class A { static get a() {} static get b() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingSetterInClass", Line: 1, Column: 29},
				},
			},
			{
				Code: `A = class { static set a(foo) {} static set b(bar) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 13},
					{MessageId: "missingGetterInClass", Line: 1, Column: 34},
				},
			},
			{
				Code: `class A { static get a() {} set b(foo) {} static set c(bar) {} get d() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 29},
					{MessageId: "missingGetterInClass", Line: 1, Column: 43},
					{MessageId: "missingSetterInClass", Line: 1, Column: 64},
				},
			},

			// ---- Class: per-class scoping ----

			{
				Code: `A = class { set a(foo) {} }, class { get a() {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 13},
					{MessageId: "missingSetterInClass", Line: 1, Column: 38},
				},
			},
			{
				Code: `A = class { get a() {} }, { set a(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 13},
					{MessageId: "missingGetterInObjectLiteral", Line: 1, Column: 29},
				},
			},
			{
				Code: `A = { get a() {} }, class { set a(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Line: 1, Column: 7},
					{MessageId: "missingGetterInClass", Line: 1, Column: 29},
				},
			},

			// ---- Class: valid+invalid combinations ----

			{
				Code: `class A { get a() {} get b() {} set b(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 11}},
			},
			{
				Code: `A = class { get b() {} get a() {} set b(foo) {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 24}},
			},
			{
				Code: `class A { set b(foo) {} get b() {} set a(bar) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass", Line: 1, Column: 36}},
			},
			{
				Code: `A = class { static get b() {} set a(foo) {} static set b(bar) {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass", Line: 1, Column: 31}},
			},
			{
				Code: `class A { static set a(foo) {} get b() {} set b(bar) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass", Line: 1, Column: 11}},
			},
			{
				Code: `class A { get b() {} static get a() {} set b(bar) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 22}},
			},
			{
				Code: `class A { static set b(foo) {} static get a() {} static get b() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 32}},
			},
			{
				Code: `class A { get [v1](){} static set i1(foo){} static set v2(bar){} get [i2](){} static get i3(){} set [v1](baz){} static get v2(){} set i4(quux){} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 24},
					{MessageId: "missingSetterInClass", Line: 1, Column: 66},
					{MessageId: "missingSetterInClass", Line: 1, Column: 79},
					{MessageId: "missingGetterInClass", Line: 1, Column: 131},
				},
			},

			// ---- Class: duplicate keys all reported ----

			{
				Code: `A = class { set a(foo) {} set a(foo) {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 13},
					{MessageId: "missingGetterInClass", Line: 1, Column: 27},
				},
			},
			{
				Code: `A = class { static get a() {} static get a() {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 13},
					{MessageId: "missingSetterInClass", Line: 1, Column: 31},
				},
			},
			{
				Code: `class A { set a(foo) {} set a(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Line: 1, Column: 11},
					{MessageId: "missingGetterInClass", Line: 1, Column: 25},
				},
			},

			// ---- Class: other elements don't affect ----

			{
				Code: `class A { a() {} get b() {} c() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 18}},
			},
			{
				Code: `A = class { a() {} get b() {} c() {} set d(foo) {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Line: 1, Column: 20},
					{MessageId: "missingGetterInClass", Line: 1, Column: 38},
				},
			},
			{
				Code: `class A { static a() {} get b() {} static c() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 25}},
			},
			{
				Code: `class A { a() {} get a() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 18}},
			},
			{
				Code: `A = class { static a() {} set a(foo) {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass", Line: 1, Column: 27}},
			},
			{
				Code: `class A { a() {} static get b() {} c() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 18}},
			},
			{
				Code: `A = class { static a() {} static set b(foo) {} static c() {} d() {} };`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass", Line: 1, Column: 27}},
			},
			{
				Code: `class A { a() {} static get a() {} a() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingSetterInClass", Line: 1, Column: 18}},
			},
			{
				Code: `class A { static set a(foo) {} static a() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingGetterInClass", Line: 1, Column: 11}},
			},

			// =====================================================
			// Message-text parity pass — verifies the full formatted
			// output byte-for-byte against ESLint's messages.
			// =====================================================

			// Object literal: identifier name.
			{
				Code: `var o = { set a(value) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Message: "Getter is not present for setter 'a'."},
				},
			},
			{
				Code: `var o = { get a() {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Message: "Setter is not present for getter 'a'."},
				},
			},
			// Object literal: numeric normalization.
			{
				Code: `var o = { set 1e2(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Message: "Getter is not present for setter '100'."},
				},
			},
			{
				Code: `var o = { get 0x10() {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Message: "Setter is not present for getter '16'."},
				},
			},
			// Object literal: string name.
			{
				Code: `var o = { set 'abc'(foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Message: "Getter is not present for setter 'abc'."},
				},
			},
			// Object literal: computed dynamic — no name.
			{
				Code: `var o = { get [abc]() {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInObjectLiteral", Message: "Setter is not present for getter."},
				},
			},
			{
				Code: `var o = { set [a + b](foo) {} };`, Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInObjectLiteral", Message: "Getter is not present for setter."},
				},
			},

			// Class: identifier.
			{
				Code: `class A { set a(foo) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Message: "Getter is not present for class setter 'a'."},
				},
			},
			{
				Code: `class A { get a() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Message: "Setter is not present for class getter 'a'."},
				},
			},
			// Class: static.
			{
				Code: `class A { static set a(foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Message: "Getter is not present for class static setter 'a'."},
				},
			},
			{
				Code: `class A { static get a() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Message: "Setter is not present for class static getter 'a'."},
				},
			},
			// Class: static + numeric normalization.
			{
				Code: `class A { static get 1e2() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Message: "Setter is not present for class static getter '100'."},
				},
			},
			// Class: private (no quotes, `#a`).
			{
				Code: `class A { set #a(foo) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Message: "Getter is not present for class private setter #a."},
				},
			},
			{
				Code: `class A { get #a() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Message: "Setter is not present for class private getter #a."},
				},
			},
			// Class: static private.
			{
				Code: `class A { static set #a(foo) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Message: "Getter is not present for class static private setter #a."},
				},
			},
			{
				Code: `class A { static get #a() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Message: "Setter is not present for class static private getter #a."},
				},
			},
			// Class: string that looks like private (WITH quotes, NOT private).
			{
				Code: `class A { set '#a'(foo) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Message: "Getter is not present for class setter '#a'."},
				},
			},
			{
				Code: `class A { static set '#a'(foo) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Message: "Getter is not present for class static setter '#a'."},
				},
			},
			// Class: computed dynamic — no name.
			{
				Code: `class A { set [a + b](foo) {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInClass", Message: "Getter is not present for class setter."},
				},
			},
			{
				Code: `class A { static get [f(x)]() {} }`, Options: classBoth,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInClass", Message: "Setter is not present for class static getter."},
				},
			},

			// Property descriptor — no name substitution, fixed text.
			{
				Code: `Object.defineProperty(o, 'k', {set: function(v) {}});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInPropertyDescriptor", Message: "Getter is not present in property descriptor."},
				},
			},
			{
				Code:    `Object.defineProperty(o, 'k', {get: function() {}});`,
				Options: bothOpts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInPropertyDescriptor", Message: "Setter is not present in property descriptor."},
				},
			},

			// TS: interface / type literal.
			{
				Code: `interface I { set prop(value: any): any }`, Options: map[string]interface{}{"enforceForTSTypes": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInType", Message: "Getter is not present for type setter 'prop'."},
				},
			},
			{
				Code: `type T = { get prop(): any }`, Options: map[string]interface{}{"enforceForTSTypes": true, "getWithoutSet": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSetterInType", Message: "Setter is not present for type getter 'prop'."},
				},
			},
			// TS computed key in an interface: ESLint emits `'null'` as the name
			// (a @typescript-eslint/parser quirk where `getStaticValue` coerces
			// an Identifier in a computed position to null → String(null)).
			// rslint reports with no name, which is semantically more correct.
			// Locked in here so future refactors can't silently change it.
			{
				Code:    `interface I { set [prop](value: any): any }`,
				Options: map[string]interface{}{"enforceForTSTypes": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingGetterInType", Message: "Getter is not present for type setter."},
				},
			},

			// ---- Ambient / abstract accessor without body ----

			// No body → `GetFunctionHeadLoc` falls through to `findOpenParenPos`,
			// which still finds the `(` of the parameter list. ESLint tests
			// don't exercise this, so we lock in our behavior here.
			{
				Code: `declare class A { set a(foo: number); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingGetterInClass",
						Message:   "Getter is not present for class setter 'a'.",
						// "set a" occupies columns 19–23; range ends at the `(` at col 24.
						Line: 1, Column: 19, EndLine: 1, EndColumn: 24,
					},
				},
			},
			{
				Code: `abstract class A { abstract set a(foo: number); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingGetterInClass",
						Message:   "Getter is not present for class setter 'a'.",
						// "abstract set a" occupies columns 20–33; range ends at `(` col 34.
						Line: 1, Column: 20, EndLine: 1, EndColumn: 34,
					},
				},
			},
		},
	)
}
