package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsPatterns(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// --- varsIgnorePattern ---
		{Code: `const _foo = 1;`, Options: map[string]interface{}{"varsIgnorePattern": "^_"}},
		// broad regex pattern
		{Code: `const x = 1;`, Options: map[string]interface{}{"varsIgnorePattern": "^."}},
		// character class
		{Code: `const ignoreFoo = 1;`, Options: map[string]interface{}{"varsIgnorePattern": "^ignore"}},
		// mixed: some match, some don't (matched var saved, used var ok)
		{Code: `const _a = 1; const b = 2; console.log(b);`, Options: map[string]interface{}{"varsIgnorePattern": "^_"}},
		// varsIgnorePattern saves class/using independently of dedicated options
		{Code: `class _Foo { static { console.log("init"); } }`, Options: map[string]interface{}{"varsIgnorePattern": "^_"}},
		{Code: `using _resource = {} as any;`, Options: map[string]interface{}{"varsIgnorePattern": "^_"}},

		// --- argsIgnorePattern ---
		{Code: `function foo(_bar) {} foo(1);`, Options: map[string]interface{}{"argsIgnorePattern": "^_"}},
		// parameter destructuring: argsIgnorePattern applies
		{Code: `function foo({ _a, b }: { _a: number; b: number }) { console.log(b); } foo({ _a: 1, b: 2 });`, Options: map[string]interface{}{"argsIgnorePattern": "^_"}},

		// --- caughtErrorsIgnorePattern ---
		{Code: `try {} catch (_err) {}`, Options: map[string]interface{}{"caughtErrorsIgnorePattern": "^_"}},
		// caughtErrors: 'all' + pattern matching
		{Code: `try {} catch (ignoreErr) {}`, Options: map[string]interface{}{"caughtErrors": "all", "caughtErrorsIgnorePattern": "^ignore"}},

		// --- caughtErrors: 'none' ---
		{Code: `try {} catch (e) {}`, Options: map[string]interface{}{"caughtErrors": "none"}},

		// --- args mode ---
		{Code: `function foo(bar) {} foo(1);`, Options: map[string]interface{}{"args": "none"}},
		// parameter destructuring: args "none" skips all
		{Code: `function foo({ a }: { a: number }) {} foo({ a: 1 });`, Options: map[string]interface{}{"args": "none"}},

		// --- pattern + args mode interactions ---
		// args "all" + argsIgnorePattern: matching param ignored even under "all"
		{Code: `export function foo(_a: number, b: string) { console.log(b); }`, Options: map[string]interface{}{"args": "all", "argsIgnorePattern": "^_"}},
		// args "none" + argsIgnorePattern: args not checked at all, pattern irrelevant
		{Code: `export function foo(unused: number) {}`, Options: map[string]interface{}{"args": "none", "argsIgnorePattern": "^_"}},

		// --- multiple patterns configured simultaneously ---
		{Code: `const _x = 1; export function foo(_a: number) {}`, Options: map[string]interface{}{"varsIgnorePattern": "^_", "argsIgnorePattern": "^_"}},
		{Code: `const _x = 1; try {} catch (_e) {}`, Options: map[string]interface{}{"varsIgnorePattern": "^_", "caughtErrorsIgnorePattern": "^_"}},
		// triple pattern
		{Code: `const _x = 1; export function foo(_a: number) {} try {} catch (_e) {}`, Options: map[string]interface{}{"varsIgnorePattern": "^_", "argsIgnorePattern": "^_", "caughtErrorsIgnorePattern": "^_"}},
		// mixed match/no-match: some args match, some don't (after-used, _a follows used b)
		{Code: `export function foo(b: string, _a: number) { console.log(b); }`, Options: map[string]interface{}{"argsIgnorePattern": "^_"}},

		// --- destructuredArrayIgnorePattern ---
		{Code: `const [_a, b] = [1, 2]; console.log(b);`, Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"}},
		{Code: `const [_a, _b] = [1, 2];`, Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"}},
		// nested array destructuring
		{Code: `const [[_a], b] = [[1], 2]; console.log(b);`, Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"}},
		// parameter array destructuring
		{Code: `function foo([_a, b]: number[]) { console.log(b); } foo([1, 2]);`, Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"}},
		// varsIgnorePattern alone also saves array destructured vars
		{Code: `const [_a] = [1];`, Options: map[string]interface{}{"varsIgnorePattern": "^_"}},
		// destructuredArrayIgnorePattern alone saves array destructured vars
		{Code: `const [_a] = [1];`, Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"}},
		// both patterns together
		{Code: `const [_a] = [1];`, Options: map[string]interface{}{"varsIgnorePattern": "^_", "destructuredArrayIgnorePattern": "^_"}},

		// --- ignoreRestSiblings ---
		{Code: `const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},
		// nested object destructuring with rest
		{Code: `const { a, b, ...rest } = { a: 1, b: 2, c: 3, d: 4 }; console.log(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},
		// function parameter destructuring with rest
		{Code: `function foo({ a, ...rest }: any) { console.log(rest); } foo({ a: 1, b: 2 });`, Options: map[string]interface{}{"ignoreRestSiblings": true}},
		// nested destructuring: rest sibling in inner object
		{Code: `
const { inner: { x, ...rest } } = { inner: { x: 1, y: 2, z: 3 } };
console.log(rest);
`, Options: map[string]interface{}{"ignoreRestSiblings": true}},
		// ignoreRestSiblings: false (default) — rest sibling NOT ignored, used
		{Code: `const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(foo, rest);`},

		// --- ignoreClassWithStaticInitBlock ---
		{Code: `class Foo { static { console.log("init"); } }`, Options: map[string]interface{}{"ignoreClassWithStaticInitBlock": true}},
		// multiple static blocks
		{Code: `class Foo { static { console.log("a"); } static { console.log("b"); } }`, Options: map[string]interface{}{"ignoreClassWithStaticInitBlock": true}},
		// static property + static block together
		{Code: `class Foo { static x = 1; static { console.log("init"); } }`, Options: map[string]interface{}{"ignoreClassWithStaticInitBlock": true}},
		// false + varsIgnorePattern can still save the class
		{Code: `class Foo { static {} }`, Options: map[string]interface{}{"ignoreClassWithStaticInitBlock": false, "varsIgnorePattern": "^Foo"}},

		// --- ignoreUsingDeclarations ---
		{Code: `using resource = {} as any;`, Options: map[string]interface{}{"ignoreUsingDeclarations": true}},
		{Code: `await using resource = {} as any;`, Options: map[string]interface{}{"ignoreUsingDeclarations": true}},

		// --- reportUsedIgnorePattern: used var that matches pattern → reported elsewhere, valid here means no standard error ---
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// --- varsIgnorePattern no match ---
		{
			Code:    `const foo = 1;`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},

		// --- reportUsedIgnorePattern ---
		{
			Code:    `const _foo = 1; console.log(_foo);`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 7}},
		},
		// reportUsedIgnorePattern applies to argsIgnorePattern too
		{
			Code:    `function foo(_x: number) { return _x; } foo(1);`,
			Options: map[string]interface{}{"argsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 14}},
		},
		// reportUsedIgnorePattern applies to caughtErrorsIgnorePattern too
		{
			Code:    `try { throw 1; } catch (_e) { console.log(_e); }`,
			Options: map[string]interface{}{"caughtErrorsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 25}},
		},
		// reportUsedIgnorePattern + destructuredArrayIgnorePattern
		{
			Code:    `const [_a] = [1]; console.log(_a);`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 8}},
		},
		// empty namespace + reportUsedIgnorePattern
		{
			Code: `
namespace _Foo {}
export const x = _Foo;
`,
			Options: map[string]interface{}{"reportUsedIgnorePattern": true, "varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 2, Column: 11}},
		},

		// --- pattern isolation: each pattern only applies to its own category ---
		// varsIgnorePattern should NOT apply to params
		{
			Code:    `function foo(_x: number) {} foo(1);`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 14}},
		},
		// varsIgnorePattern should NOT apply to catch
		{
			Code:    `try {} catch (_e) {}`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// argsIgnorePattern should NOT apply to vars
		{
			Code:    `const _x = 1;`,
			Options: map[string]interface{}{"argsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// argsIgnorePattern should NOT apply to catch
		{
			Code:    `try {} catch (_e) {}`,
			Options: map[string]interface{}{"argsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// caughtErrorsIgnorePattern should NOT apply to vars
		{
			Code:    `const _x = 1;`,
			Options: map[string]interface{}{"caughtErrorsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// caughtErrorsIgnorePattern should NOT apply to params
		{
			Code:    `export function foo(_x: number) {}`,
			Options: map[string]interface{}{"caughtErrorsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 21}},
		},
		// argsIgnorePattern no match
		{
			Code:    `export function foo(bar: number) {}`,
			Options: map[string]interface{}{"argsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 21}},
		},
		// caughtErrorsIgnorePattern no match
		{
			Code:    `try {} catch (err) {}`,
			Options: map[string]interface{}{"caughtErrorsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// multiple patterns: non-matching category still reported
		{
			Code:    `const _x = 1; try {} catch (_e) {}`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 29}},
		},
		// args "all" + argsIgnorePattern: non-matching params still reported
		{
			Code:    `export function foo(_a: number, bar: string) { console.log(_a); }`,
			Options: map[string]interface{}{"args": "all", "argsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 33}},
		},
		// mixed match/no-match: unmatched var still reported
		{
			Code:    `const _a = 1; const b = 2;`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 21}},
		},

		// --- caughtErrors: 'all' + caughtErrorsIgnorePattern: non-matching var ---
		{
			Code:    `try {} catch (err) {}`,
			Options: map[string]interface{}{"caughtErrors": "all", "caughtErrorsIgnorePattern": "^ignore"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// multiple catch blocks: first matches pattern, second doesn't
		{
			Code: `
try {} catch (ignoreErr) {}
try {} catch (err) {}
`,
			Options: map[string]interface{}{"caughtErrors": "all", "caughtErrorsIgnorePattern": "^ignore"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 15}},
		},
		// caughtErrors: 'all' without pattern — all unused catch vars reported
		{
			Code: `
try {} catch (e1) {}
try {} catch (e2) {}
`,
			Options: map[string]interface{}{"caughtErrors": "all"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 15},
				{MessageId: "unusedVar", Line: 3, Column: 15},
			},
		},

		// --- destructuredArrayIgnorePattern ---
		// pattern does not match → still reported
		{
			Code:    `const [foo] = [1];`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 8}},
		},
		// should NOT apply to plain params
		{
			Code:    `export function foo(_x: number) {}`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 21}},
		},
		// should NOT apply to catch
		{
			Code:    `try {} catch (_e) {}`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		// should NOT apply to object destructuring
		{
			Code:    `const { _a } = { _a: 1 };`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 9}},
		},
		// object destructuring inside array destructuring → NOT affected
		{
			Code:    `const [{ _a }] = [{ _a: 1 }];`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 10}},
		},
		// should NOT apply to non-destructured vars
		{
			Code:    `const _a = 1;`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// mixed array destructuring: unmatched elements still reported
		{
			Code: `
const array = ["a", "b", "c", "d", "e"];
const [a, _b, c] = array;
`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 8},
				{MessageId: "unusedVar", Line: 3, Column: 15},
			},
		},
		// destructuredArrayIgnorePattern + varsIgnorePattern together
		{
			Code: `
const array = ["a", "b", "c"];
const [a, _b, c] = array;
const fooArray = ["foo"];
const barArray = ["bar"];
const ignoreArray = ["ignore"];
`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_", "varsIgnorePattern": "ignore"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 8},
				{MessageId: "unusedVar", Line: 3, Column: 15},
				{MessageId: "unusedVar", Line: 4, Column: 7},
				{MessageId: "unusedVar", Line: 5, Column: 7},
			},
		},
		// object destructuring inside array: _a reported
		{
			Code: `
const array = [{}];
const [{ _a, foo }] = array;
console.log(foo);
`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 10}},
		},
		// parameter object-in-array destructuring: _a reported
		{
			Code: `
function foo([{ _a, bar }]: any) { bar; }
foo([]);
`,
			Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 17}},
		},

		// --- ignoreClassWithStaticInitBlock ---
		// class WITHOUT static block → still reported even when option is true
		{
			Code:    `class Foo { static x = 1; }`,
			Options: map[string]interface{}{"ignoreClassWithStaticInitBlock": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// class WITH static block → reported when option is false (default)
		{
			Code:   `class Foo { static { console.log("init"); } }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// class WITH static block → reported when option is explicitly false
		{
			Code:    `class Foo { static { console.log("init"); } }`,
			Options: map[string]interface{}{"ignoreClassWithStaticInitBlock": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// class with static method (not static block) → still reported
		{
			Code: `
class Foo {
  static bar() {}
}
`,
			Options: map[string]interface{}{"ignoreClassWithStaticInitBlock": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 7}},
		},
		// class with static field (not static block) → still reported
		{
			Code: `
class Foo {
  static bar = 1;
}
`,
			Options: map[string]interface{}{"ignoreClassWithStaticInitBlock": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Column: 7}},
		},

		// --- ignoreUsingDeclarations ---
		// using declaration → reported when option is false (default)
		{
			Code:   `using resource = {} as any;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// await using → reported when option is false (default)
		{
			Code:   `await using resource = {} as any;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 13}},
		},
		// using declaration → reported when option is explicitly false
		{
			Code:    `using resource = {} as any;`,
			Options: map[string]interface{}{"ignoreUsingDeclarations": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// const NOT affected by ignoreUsingDeclarations
		{
			Code:    `const foo = 1;`,
			Options: map[string]interface{}{"ignoreUsingDeclarations": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		// let NOT affected by ignoreUsingDeclarations
		{
			Code:    `let foo = 1;`,
			Options: map[string]interface{}{"ignoreUsingDeclarations": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 5}},
		},

		// --- ignoreRestSiblings: false (default) ---
		{
			Code:   `const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(rest);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 9}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}
