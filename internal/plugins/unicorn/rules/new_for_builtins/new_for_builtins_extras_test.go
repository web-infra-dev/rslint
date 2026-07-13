package new_for_builtins_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/new_for_builtins"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNewForBuiltinsExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Inline comments describe the behavior each case
// protects so future refactors don't silently loosen the matching logic.
func TestNewForBuiltinsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&new_for_builtins.NewForBuiltinsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: optional calls that would require optional `new` are ignored ----
			jsValid("const value = (globalThis?.Array)();"),
			// ---- Dimension 4: dynamic element access is not a static builtin reference ----
			jsValid("const key = 'Array'; globalThis[key]();"),
			// ---- Dimension 4: numeric element keys do not match builtin names ----
			jsValid("globalThis[0]();"),
			// ---- Dimension 4: computed destructuring with a dynamic key does not create an alias ----
			jsValid(lines(
				"const key = 'Array';",
				"const {[key]: A} = globalThis;",
				"A();",
			)),
			// ---- Dimension 4: spread/rest object shapes do not crash alias collection ----
			jsValid("const {...rest} = globalThis; rest.Array();"),
			// ---- Dimension 4: empty destructuring patterns do not crash alias collection ----
			jsValid("const {} = globalThis;"),
			// ---- Dimension 4: nested function parameters shadow globals with the same name ----
			jsValid(lines(
				"function outer() {",
				"\tfunction inner(Array) {",
				"\t\treturn Array();",
				"\t}",
				"}",
			)),
			// ---- Dimension 4: alias scope does not leak out of a block ----
			jsValid(lines(
				"{",
				"\tconst {Array: A} = globalThis;",
				"}",
				"A();",
			)),
			// ---- Dimension 4: inner bindings shadow a tracked outer alias ----
			jsValid(lines(
				"const {Array: A} = globalThis;",
				"{",
				"\tconst A = function() {};",
				"\tA();",
				"}",
			)),
			// ---- Dimension 4: function parameters shadow a tracked outer alias ----
			jsValid(lines(
				"function outer() {",
				"\tconst {Array: A} = globalThis;",
				"\tfunction inner(A) {",
				"\t\treturn A();",
				"\t}",
				"}",
			)),
			// ---- Dimension 4: for-initializer alias scope does not leak after the loop ----
			jsValid(lines(
				"for (const {Array: A} = globalThis; false;) {}",
				"A();",
			)),
			// ---- Dimension 4: aliases from non-global objects are not tracked ----
			jsValid(lines(
				"const {Array: A} = maybeGlobal;",
				"A();",
			)),
			// ---- Dimension 4: local global-object names shadow the real globals ----
			jsValid("const globalThis = {Array() {}}; globalThis.Array();"),
			// ---- Dimension 4: bare WebAssembly namespace aliases are not tracked by upstream ----
			jsValid(lines(
				"const WA = WebAssembly;",
				"WA.Module(buffer);",
				"WA.JSTag();",
				"new WA();",
			)),
			// ---- Dimension 4: bare WebAssembly member aliases are not tracked by upstream ----
			jsValid(lines(
				"const Module = WebAssembly.Module;",
				"Module(buffer);",
			)),
			// N/A: private identifiers are not valid property names on global builtin references.
			// N/A: declaration/container forms do not apply; this rule targets CallExpression and NewExpression.
			// N/A: overload, abstract, and declare body-absent forms do not apply to call/new expressions.

			// ---- Real-user: #122 immutable Map import must shadow the builtin ----
			jsValid(lines(
				"import {Map as ImmutableMap} from 'immutable';",
				"const m = ImmutableMap();",
			)),
			// ---- Real-user: #917 Object identity guard remains allowed in nested expression ----
			jsValid("const isObject = value => value !== null && Object(value) === value;"),

			// Locks in upstream enforceNewExpression() arm 1: optional chain early return.
			jsValid("const value = globalThis?.Map();"),
			// Locks in upstream enforceCallExpression() arm 1: String wrapper reports without fix.
			jsValid("const value = String('test');"),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver and callee wrappers are transparent ----
			enforceInvalid("const value = (globalThis).Array();", "(globalThis).Array()", "Array"),
			// ---- Dimension 4: multi-level parenthesized receivers are transparent ----
			enforceInvalid("const value = ((globalThis)).Array();", "((globalThis)).Array()", "Array"),
			// ---- Dimension 4: TS type-expression wrappers on receivers are transparent ----
			tsEnforceInvalid("const value = (globalThis as any).Array();", "(globalThis as any).Array()", "Array"),
			// ---- Dimension 4: TS non-null wrappers on receivers are transparent ----
			tsEnforceInvalid("const value = (globalThis!).Array();", "(globalThis!).Array()", "Array"),
			// ---- Dimension 4: TS satisfies wrappers on receivers are transparent ----
			tsEnforceInvalid(
				"const value = (globalThis satisfies typeof globalThis).Array();",
				"(globalThis satisfies typeof globalThis).Array()",
				"Array",
			),
			// ---- Dimension 4: static string element access matches dotted access ----
			enforceInvalid("const value = globalThis['Array']();", "globalThis['Array']()", "Array"),
			// ---- Dimension 4: static template element access matches dotted access ----
			enforceInvalid("const value = globalThis[`Array`]();", "globalThis[`Array`]()", "Array"),
			// ---- Dimension 4: static element access composes across namespace paths ----
			enforceInvalid(
				"const value = globalThis['Intl']['DateTimeFormat']();",
				"globalThis['Intl']['DateTimeFormat']()",
				"Intl.DateTimeFormat",
			),
			// ---- Dimension 4: static element access to Date keeps Date's special replacement ----
			dateFixInvalid("const value = globalThis['Date']();", "globalThis['Date']()", "const value = String(new Date());"),
			// ---- Dimension 4: sibling block declarations do not shadow a later outer global reference ----
			enforceInvalid(lines(
				"function outer() {",
				"\t{ const Array = function() {}; }",
				"\treturn Array();",
				"}",
			), "Array()", "Array"),
			// ---- Dimension 4: static element access can pass through TS assertions ----
			tsEnforceInvalid("const value = globalThis[('Array' as const)]();", "globalThis[('Array' as const)]()", "Array"),
			// ---- Dimension 4: static computed destructuring keys create aliases ----
			enforceInvalid(lines(
				"const {['Array']: A} = globalThis;",
				"A();",
			), "A()", "Array"),
			// ---- Dimension 4: static template computed destructuring keys create aliases ----
			enforceInvalid(lines(
				"const {[`Array`]: A} = globalThis;",
				"A();",
			), "A()", "Array"),
			// ---- Dimension 4: destructuring aliases remain tracked with default values ----
			enforceInvalid(lines(
				"const {Array: A = function() {}} = globalThis;",
				"A();",
			), "A()", "Array"),
			// ---- Dimension 4: function-scoped destructuring aliases behave like globals ----
			enforceInvalid(lines(
				"function outer() {",
				"\tconst {Array: A} = globalThis;",
				"\treturn A();",
				"}",
			), "A()", "Array"),
			// ---- Dimension 4: nested object destructuring preserves the full builtin path ----
			enforceInvalid(
				"const {Intl: {DateTimeFormat: Format}} = globalThis; Format();",
				"Format()",
				"Intl.DateTimeFormat",
			),
			// ---- Dimension 4: aliases can be chained before the call site ----
			enforceInvalid(lines(
				"const {Array: A} = globalThis;",
				"const B = A;",
				"B();",
			), "B()", "Array"),
			// ---- Dimension 4: alias declarations after a use still match upstream scope tracing ----
			enforceInvalid(lines(
				"A();",
				"const {Array: A} = globalThis;",
			), "A()", "Array"),
			// ---- Dimension 4: bare Intl namespace aliases are tracked by upstream ----
			enforceInvalid(lines(
				"const I = Intl;",
				"I.DateTimeFormat();",
			), "I.DateTimeFormat()", "Intl.DateTimeFormat"),
			// ---- Dimension 4: namespace aliases from global objects are tracked ----
			enforceInvalid(lines(
				"const I = globalThis.Intl;",
				"I.DateTimeFormat();",
			), "I.DateTimeFormat()", "Intl.DateTimeFormat"),
			// ---- Dimension 4: bare Temporal namespace aliases are tracked by upstream ----
			enforceInvalid(lines(
				"const T = Temporal;",
				"T.PlainDate(2024, 1, 1);",
			), "T.PlainDate(2024, 1, 1)", "Temporal.PlainDate"),
			// ---- Dimension 4: disallowed Temporal children are tracked through namespace aliases ----
			disallowCallOrNewInvalid(lines(
				"const T = Temporal;",
				"T.Now?.();",
			), "T.Now?.()", "Temporal.Now"),
			// ---- Dimension 4: aliasing a disallowed Temporal member is tracked ----
			disallowCallOrNewInvalid(lines(
				"const Now = Temporal.Now;",
				"Now();",
			), "Now()", "Temporal.Now"),
			// ---- Dimension 4: global-object WebAssembly aliases report direct invalid calls ----
			disallowCallOrNewInvalid(lines(
				"const WA = globalThis.WebAssembly;",
				"WA();",
			), "WA()", "WebAssembly"),
			// ---- Dimension 4: namespace aliases from global objects report disallowed child constructors ----
			disallowCallOrNewInvalid(lines(
				"const {WebAssembly: WA} = globalThis;",
				"new WA.JSTag();",
			), "new WA.JSTag()", "WebAssembly.JSTag"),
			// ---- Dimension 4: member aliases from global object namespaces are tracked ----
			enforceInvalid(lines(
				"const Module = globalThis.WebAssembly.Module;",
				"Module(buffer);",
			), "Module(buffer)", "WebAssembly.Module"),
			// ---- Dimension 4: aliases declared in a for initializer are usable inside the loop ----
			enforceInvalid(lines(
				"for (const {Array: A} = globalThis; false;) {",
				"\tA();",
				"}",
			), "A()", "Array"),
			// ---- Dimension 4: var aliases declared inside a block are usable in that block ----
			enforceInvalid(lines(
				"if (condition) {",
				"\tvar A = Array;",
				"\tA();",
				"}",
			), "A()", "Array"),
			// ---- Dimension 4: var aliases declared inside a block are hoisted to the outer scope ----
			enforceInvalid(lines(
				"if (condition) {",
				"\tvar A = Array;",
				"}",
				"A();",
			), "A()", "Array"),

			// ---- Real-user: #901 new String('test') reports but has no autofix ----
			disallowNoFixInvalid("const str = new String('test');", "new String('test')", "String"),
			// ---- Real-user: #1835 namespace object call is not fixable ----
			disallowCallOrNewInvalid("const tag = WebAssembly.JSTag();", "WebAssembly.JSTag()", "WebAssembly.JSTag"),

			// Locks in upstream enforceNewExpression() arm 2: Object loose equality is not exempt.
			enforceInvalid("const isObject = value => Object(value) != value;", "Object(value)", "Object"),
			// Locks in upstream enforceNewExpression() arm 3: Date with comments uses a suggestion.
			dateSuggestionInvalid("const value = Date(/* now */);", "Date(/* now */)", "const value = String(new Date());"),
			// Locks in upstream enforceCallExpression() arm 2: Symbol is autofixable.
			disallowInvalid("const symbol = new Symbol('x');", "new Symbol('x')", "Symbol", "const symbol = Symbol('x');"),
			// Locks in the same multiline remove-`new` fixer branch for BigInt.
			disallowInvalid(lines(
				"const big = new // bigint",
				"\tBigInt(1);",
			), "new // bigint\n\tBigInt(1)", "BigInt", lines(
				"const big = // bigint",
				"\tBigInt(1);",
			)),
			// Locks in upstream disallowCallOrNewExpression() construct path.
			disallowCallOrNewInvalid("const invalid = new WebAssembly();", "new WebAssembly()", "WebAssembly"),
			// Locks in upstream GlobalReferenceTracker alias path from globalThis destructuring.
			disallowNoFixInvalid(lines(
				"const {Number: RenamedNumber} = globalThis;",
				"new RenamedNumber(1);",
			), "new RenamedNumber(1)", "Number"),
			// Locks in fixSpaceAroundKeyword-equivalent behavior for keyword fusion.
			enforceInvalid(lines(
				"function f() {",
				"\treturn(globalThis).Array();",
				"}",
			), "(globalThis).Array()", "Array"),
		},
	)
}

func tsEnforceInvalid(code string, target string, name string) rule_tester.InvalidTestCase {
	testCase := enforceInvalid(code, target, name)
	testCase.FileName = "file.ts"
	return testCase
}
