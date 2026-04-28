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

		// ====================================================================
		// Regression cases for property-name-position pollution of the
		// name-keyed unresolvedRefs fallback. When TypeChecker resolution
		// fails for an identifier in a property/label/attribute position
		// (e.g. `obj.name` on an `any`-typed receiver, ImportSpecifier /
		// ExportSpecifier `PropertyName` from unresolvable modules), the
		// fallback must not record those identifiers — otherwise a same-named
		// unused local is falsely treated as used.
		// ====================================================================

		// `_tail` renamed property in destructuring with rest sibling, under
		// a config that does NOT define varsIgnorePattern. argsIgnorePattern
		// must not apply to a `const` destructuring; the report should fire.
		{
			Code: `function f(params: any) {
  if (params.head !== undefined && params.tail !== undefined) {
    const { tail: _tail, ...rest } = params;
    return rest;
  }
  return params;
}
export { f };`,
			Options: map[string]interface{}{
				"args":              "none",
				"argsIgnorePattern": "^_",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 19},
			},
		},

		// Unused type parameter `T` on an exported interface (does not match
		// `^_|React`). Re-exporting the interface elsewhere must not mark the
		// type parameter as used.
		{
			Code: `export interface RpcResponse<T = unknown> {
  code?: number;
  message?: string;
}`,
			Options: map[string]interface{}{
				"args":                           "none",
				"varsIgnorePattern":              "^_|React",
				"argsIgnorePattern":              "^_",
				"destructuredArrayIgnorePattern": "^_",
				"caughtErrorsIgnorePattern":      "^_",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 30},
			},
		},

		// Object destructuring; `name` is unused, siblings are used. The
		// polluting `obj.name` accesses on an `any`-typed receiver in the
		// same file previously masked the report via the name fallback.
		{
			Code: `interface Trigger { id: string; projectId: string; name: string; rule: string; type: string; }
declare const externalAny: any;
export function pollute(triggers: Trigger[]): void {
  console.log(externalAny.name, externalAny.name);
  for (const trigger of triggers) {
    const { id, projectId, name, rule, type } = trigger;
    console.log(id, projectId, rule, type);
  }
}`,
			Options: map[string]interface{}{
				"args":                           "none",
				"varsIgnorePattern":              "^_|React",
				"argsIgnorePattern":              "^_",
				"destructuredArrayIgnorePattern": "^_",
				"caughtErrorsIgnorePattern":      "^_",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 6, Column: 28},
			},
		},

		// Same root cause as the previous case: destructured
		// `deployConfigSource` should be reported even when there are many
		// `obj.deployConfigSource` property accesses on `any`-typed values
		// in the same file.
		{
			Code: `type Meta = { id: string; region: string; appName: string; vRegion: string; deployConfigSource: string };
export function process(item: any) {
  const getMetaInfo = (i: any): Meta => i as Meta;
  const { id, region, appName, vRegion, deployConfigSource } = getMetaInfo(item);
  console.log(id, region, appName, vRegion);
}`,
			Options: map[string]interface{}{
				"args":                           "none",
				"varsIgnorePattern":              "^_|React",
				"argsIgnorePattern":              "^_",
				"destructuredArrayIgnorePattern": "^_",
				"caughtErrorsIgnorePattern":      "^_",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 4, Column: 41},
			},
		},

		// `const X = obj.X ?? ...` where `obj` is `any`. The same-named
		// PropertyAccess on the RHS must not be counted as a use of the LHS.
		{
			Code: `declare const project: any;
export function getCfg(): boolean {
  const isTikTokBiz = project.features?.isTikTokBiz ?? true;
  return true;
}`,
			Options: map[string]interface{}{
				"args":                           "none",
				"varsIgnorePattern":              "^_|React",
				"argsIgnorePattern":              "^_",
				"destructuredArrayIgnorePattern": "^_",
				"caughtErrorsIgnorePattern":      "^_",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// ImportSpecifier.PropertyName from an unresolvable module. The
		// `name` on the LHS of `as` references the module's exported
		// binding which is nil when the module cannot be resolved. It must
		// NOT be added to unresolvedRefs to falsely "use" the local `name`.
		{
			Code: `import { name as alias } from './does-not-exist';
function foo(): number {
  const name = 1;
  return alias as unknown as number;
}
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// ExportSpecifier.PropertyName in a re-export from an unresolvable
		// module — same fallback-pollution pattern as the import case above.
		{
			Code: `export { name as renamed } from './does-not-exist';
function foo(): number {
  const name = 1;
  return 0;
}
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// ExportSpecifier.Name in a non-renamed re-export from an unresolvable
		// module: `export { name } from 'mod'` — the `name` identifier is a
		// module-level binding reference, not a local one, so it must not
		// pollute the local `name` lookup. Symmetric to the renamed case.
		{
			Code: `export { name } from './does-not-exist';
function foo(): number {
  const name = 1;
  return 0;
}
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// Conversely: `export { name }` (NO `from`) — `name` IS a value
		// reference to the local `name`. The local must be marked as used
		// and the rule must not falsely report it.
		{
			Code: `function foo(): number {
  const name = 1;
  return name;
}
const name = 1;
export { name };
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{},
		},

		// Type-only re-export: `export type { name } from 'mod'` — same
		// rule as the value-only re-export above. The `name` must not be
		// treated as a re-export of the same-named local value declaration.
		{
			Code: `export type { name } from './does-not-exist';
function foo(): number {
  const name = 1;
  return 0;
}
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// Namespace re-export: `export * as name from 'mod'` — `name` here
		// is the namespace export name, not a reference to a local. Same
		// requirement as the named re-export forms.
		{
			Code: `export * as renamed from './does-not-exist';
function foo(): number {
  const renamed = 1;
  return 0;
}
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// Enum member name with a same-named unused local. EnumMember.Name is
		// a declaration name (handled by isDeclarationName); this guards
		// against future regressions in that classification.
		{
			Code: `export enum E { name = 1 }
function foo(): number {
  const name = 1;
  return E.name;
}
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// JSX namespaced attribute name (`<svg xml:lang="en">`). The
		// `xml`/`lang` parts of `xml:lang` are JsxNamespacedName components
		// that name an attribute, not in-scope value references. They must
		// not pollute same-named locals via unresolvedRefs.
		{
			Tsx: true,
			Code: `export function Foo() {
  const xml = 1;
  const lang = 2;
  return <svg xml:lang="en" />;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 9},
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// Import attribute (`import 'foo' with { type: 'json' }`): the
		// `type` key in the attribute clause is an attribute name, not a
		// value reference to any in-scope `type` local.
		{
			Code: `import './does-not-exist' with { type: 'json' };
function foo(): number {
  const type = 1;
  return 0;
}
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// Property signature in TypeLiteral / interface body — `name` here is
		// a property declaration, not a value reference. Same-named local
		// must still be reported.
		{
			Code: `type T = { name: string; age: number };
function foo(t: T): number {
  const name = 1;
  return t.age;
}
export { foo };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// Interface property signature — same as above but in an interface.
		{
			Code: `interface I { foo: string; bar: number }
function use(i: I): number {
  const foo = 1;
  return i.bar;
}
export { use };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 9},
			},
		},

		// Class property/method declarations — both are declaration names.
		// Same-named outer locals must still be reported.
		{
			Code: `class C {
  name: string = 'x';
  method(): number { return 1; }
}
const name = 1;
const method = 2;
new C();
export { C };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 5, Column: 7},
				{MessageId: "unusedVar", Line: 6, Column: 7},
			},
		},

		// Get/Set accessor names — declaration names.
		{
			Code: `class C {
  get name(): string { return 'x'; }
  set name(v: string) {}
}
const name = 1;
new C();
export { C };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 5, Column: 7},
			},
		},

		// Decorator with member access — `Foo.bar` decorator. `bar` is
		// PropertyAccess.Name (excluded) and must not pollute a local `bar`.
		{
			Code: `declare const Decor: any;
const bar = 1;
class C {
  @Decor.bar
  prop!: string;
}
new C();
export { C };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
			},
		},

		// JSX element with member-access tag (`<Foo.Bar />`) on an `any`
		// receiver. The `Bar` is a property name; same-named local must
		// remain reportable.
		{
			Tsx: true,
			Code: `declare const Foo: any;
const Bar = 1;
export const X = <Foo.Bar />;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
			},
		},

		// Nested PropertyAccess on `any` (`a.b.c.d`) with same-named
		// locals at every level. None of `b`, `c`, `d` are value refs.
		{
			Code: `declare const a: any;
const b = 1;
const c = 2;
const d = 3;
export const x = a.b.c.d;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 7},
				{MessageId: "unusedVar", Line: 4, Column: 7},
			},
		},

		// Nested object literal property names + property access — every
		// inner identifier in property-name position must not pollute
		// same-named locals.
		{
			Code: `const x = 1;
const y = 2;
const z = 3;
declare const inp: any;
const o = { x: inp.x, y: inp.y, z: inp.z };
export { o };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 7},
				{MessageId: "unusedVar", Line: 2, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},

		// Nested destructuring rename: outer pattern with renamed inner
		// pattern. The source-side property names (`a`, `b`) must not
		// pollute same-named locals.
		{
			Code: `declare const obj: any;
const a = 1;
const b = 2;
function f(): number {
  const { a: { b: alias } = {} } = obj;
  return alias as unknown as number;
}
export { f };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},

		// JSX nested member tag `<A.B.C />` — every Name beyond `A` is a
		// property reference and must not pollute same-named locals.
		{
			Tsx: true,
			Code: `declare const A: any;
const B = 1;
const C = 2;
export const X = <A.B.C />;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},

		// `typeof obj.x` on an `any`-typed receiver in a type position.
		// `x` is PropertyAccess.Name inside a TypeQuery; same-named
		// local must still be reported. (`obj` is additionally flagged as
		// only used as a type, which is the correct typescript-eslint
		// behavior — the test's purpose here is the `x` report.)
		{
			Code: `declare function getObj(): any;
const obj = getObj();
const x = 1;
type T = typeof obj.x;
const v: T = 1 as any;
export { v };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usedOnlyAsType", Line: 2, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},

		// Deeply chained PropertyAccess on `any` (`a.b.c.d.e`) — every
		// inner identifier (b/c/d/e) is at PropertyAccess.Name. None must
		// pollute its respective same-named local. Acts as a guard that
		// the exclusion is applied recursively, not just on the outermost
		// access.
		{
			Code: `declare const a: any;
const b = 1;
const c = 2;
const d = 3;
const e = 4;
console.log(a.b.c.d.e);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 7},
				{MessageId: "unusedVar", Line: 4, Column: 7},
				{MessageId: "unusedVar", Line: 5, Column: 7},
			},
		},

		// Optional-chain at multiple levels (`obj?.x?.y`) on `any` — same
		// guarantee as the regular chain above.
		{
			Code: `declare const obj: any;
const x = 1;
const y = 2;
console.log(obj?.x?.y);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}
