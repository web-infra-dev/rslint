// Expected reduce fixes also remove sparse-array holes.
// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-object-from-entries.js
package prefer_object_from_entries_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_object_from_entries"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferObjectFromEntriesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_object_from_entries.PreferObjectFromEntriesRule,
		[]rule_tester.ValidTestCase{
			// Upstream typed-array runtime regressions (lint and applied output).
			{Code: "JSON.stringify(new Uint8Array([1, 2]).reduce((object, value) => ({...object, [value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(new Uint8Array([1, 2]).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(new Float64Array([1, 2]).reduce((object, value) => ({...object, [value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(new Float64Array([1, 2]).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(new BigInt64Array([1n, 2n]).reduce((object, value) => ({...object, [value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(new BigInt64Array([1n, 2n]).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(Uint16Array.from([1, 2]).reduce((object, value) => ({...object, [value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(Uint16Array.from([1, 2]).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(Uint8Array.of(1, 2).reduce((object, value) => ({...object, [value]: Number(value)}), {}));", FileName: "case.js"},
			{Code: "JSON.stringify(Uint8Array.of(1, 2).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));", FileName: "case.js"},
			// test.snapshot: reduce and for-of.
			{Code: "new Uint8Array([1, 2]).reduce((object, value) => ({...object, [value]: value}), {});", FileName: "case.js"},
			{Code: "const values = new Int16Array([1, 2]); values.reduce((object, value) => Object.assign(object, {[value]: value}), {});", FileName: "case.js"},
			{Code: "Uint16Array.from([1, 2]).reduce((object, value) => ({...object, [value]: value}), {});", FileName: "case.js"},
			{Code: "Uint8Array.of(1, 2).reduce((object, value) => Object.assign(object, {[value]: value}), {});", FileName: "case.js"},
			{Code: "const values = Float32Array.of(...input); values.reduce((object, value) => Object.assign(object, {[value]: value}), {});", FileName: "case.js"},
			{Code: "const values = new Uint8Array(input); const alias = values; alias.reduce((object, value) => ({...object, [value]: value}), {});", FileName: "case.js"},
			{Code: "((new Uint8Array(values))).reduce((object, value) => ({...object, [value]: value}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), Object.create(null));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), Object.create(null),);", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), (( Object.create(null) )),);", FileName: "case.js"},
			{Code: "let object = Object.create(null);\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, __proto__: value}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object.assign(object, {__proto__: value}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, \"__proto__\": value}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object.assign(object, {'__proto__': value}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), {}, extraArgument);", FileName: "case.js"},
			{Code: "pairs.reduce({}, object => ({...object, key}));", FileName: "case.js"},
			{Code: "reduce(object => ({...object, key}), {});", FileName: "case.js"},
			{Code: "new reduce(object => ({...object, key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce?.(object => ({...object, key}), {});", FileName: "case.js"},
			{Code: "pairs?.reduce(object => ({...object, key}), {});", FileName: "case.js"},
			{Code: "pairs.notReduce(object => ({...object, key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), {notEmpty});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), []);", FileName: "case.js"},
			{Code: "pairs.reduce(...[(object => ({...object, key}))], {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), ...[{}]);", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), Object.create());", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), Object.create(null, extraArgument));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), Object.create?.(null));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), Object?.create(null));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), window.Object.create(null));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), Object.notCreate(null));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), NotObject.create(null));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), object.create(null));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), object.CREATE(null));", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key}), Object.create(\"null\"));", FileName: "case.js"},
			{Code: "pairs.reduce(callback, {})", FileName: "case.js"},
			{Code: "pairs.reduce(callback, Object.create(null))", FileName: "case.js"},
			{Code: "pairs.reduce(async function * () {}, {})", FileName: "case.js"},
			{Code: "pairs.reduce()", FileName: "case.js"},
			{Code: "pairs.reduce(callback, {}, extraArgument)", FileName: "case.js"},
			{Code: "pairs.reduce?.(callback, {})", FileName: "case.js"},
			{Code: "pairs?.reduce(callback, {})", FileName: "case.js"},
			{Code: "pairs.notReduce(callback, {})", FileName: "case.js"},
			{Code: "pairs[reduce](callback, {})", FileName: "case.js"},
			{Code: "pairs.reduce(...callback, {})", FileName: "case.js"},
			{Code: "pairs.reduce(function(object) {Object.assign(object, {key})}, {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key} + 1), {});", FileName: "case.js"},
			{Code: "pairs.reduce((object = {}) => ({...object, key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce((object) => ({...NotSameObject, key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key, anotherKey}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({keyFirst, ...object}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(async object => ({...object, key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(async object => await {...object, key}, {});", FileName: "case.js"},
			{Code: "pairs.reduce((...object) => ({...object, key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(({object}) => ({...object, key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, ...key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object.assign(NotSameObject, {key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object.assign(object, {}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object.assign(object, {...key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object.assign?.(object, {key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object?.assign(object, {key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object.notAssign(object, {key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => NotObject.assign(object, {key}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, object}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key: Object.keys(object)}), {});", FileName: "case.js"},
			{Code: "pairs.reduce((object, [key, value = object]) => ({...object, [key]: value}), {});", FileName: "case.js"},
			{Code: "pairs.reduce((object) => Object.assign(object, {object}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, key: function () { return object; }}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, method() {}}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => Object.assign(object, {async * method() {}}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, async method() {}}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, * method() {}}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, async * method() {}}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, get key() {}}), {});", FileName: "case.js"},
			{Code: "pairs.reduce(object => ({...object, set key(v) {}}), {});", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = doSomething(value);\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n\tcount++;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfoo();\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js"},
			{Code: "const object = {}, otherObject = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js"},
			{Code: "async function foo() {\n\tconst object = {};\n\tfor await (const [key, value] of pairs) {\n\t\tobject[key] = value;\n\t}\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] += value;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject.key = value;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, value] of pairs) {\n\totherObject[key] = value;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tthis.object[key] = value;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[value] = key;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (var [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, value] of object) {\n\tobject[key] = value;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [object, value] of pairs) {\n\tobject[object] = value;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key = defaultKey, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js"},
			{Code: "const object = {};\nfor (const [key, ...value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js"},
			{Code: "let object;\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js"},
			{Code: "const flattened = arrayOfObjects.reduce((flattened, next) => Object.assign(flattened, next), {});", FileName: "case.js"},
			// test.typescript.
			{Code: "function convert(values: Uint8Array) { return values.reduce((object, value) => ({...object, [value]: value}), {}); }", FileName: "case.ts"},
			{Code: "(values as Int32Array).reduce((object, value) => ({...object, [value]: value}), {});", FileName: "case.ts"},
			{Code: "(<Float32Array>values).reduce((object, value) => Object.assign(object, {[value]: value}), {});", FileName: "case.ts"},
			{Code: "function convert(values: Uint8Array) { return values!.reduce((object, value) => ({...object, [value]: value}), {}); }", FileName: "case.ts"},
			{Code: "(new Uint8Array(values) satisfies Uint8Array).reduce((object, value) => ({...object, [value]: value}), {});", FileName: "case.ts"},
			{Code: "type Bytes = Uint8Array; function convert(values: Bytes) { return values.reduce((object, value) => ({...object, [value]: value}), {}); }", FileName: "case.ts"},
			// test: typed parser services.
			{Code: "declare function getValues(): Uint8Array; getValues().subarray(0).reduce((object, value) => ({...object, [value]: value}), {});", FileName: "case.ts"},
			// test.snapshot: functions.
			{Code: "underscore.fromPairs(pairs)", FileName: "case.js"},
			{Code: "_.fromPairs", FileName: "case.js"},
			{Code: "_.fromPairs()", FileName: "case.js"},
			{Code: "new _.fromPairs(pairs)", FileName: "case.js"},
			{Code: "_.fromPairs(...[pairs])", FileName: "case.js"},
			{Code: "_?.fromPairs(pairs)", FileName: "case.js"},
			{Code: "_.foo(pairs)", FileName: "case.js", Options: []any{map[string]any{"functions": []any{"foo"}}}},
			{Code: "foo(pairs)", FileName: "case.js", Options: []any{map[string]any{"functions": []any{"utils.object.foo"}}}},
			{Code: "object.foo(pairs)", FileName: "case.js", Options: []any{map[string]any{"functions": []any{"utils.object.foo"}}}},
			// Documentation examples.
			{Code: "const object = Object.fromEntries(pairs);", FileName: "case.js"},
			{Code: "const object = pairs.reduce(\n\t(object, [key, value]) => ({...object, [key]: value}),\n\tObject.create(null)\n);", FileName: "case.js"},
			{Code: "const object = Object.fromEntries(pairs);", FileName: "case.js", Options: []any{map[string]any{"functions": []any{"getObjectFromKeyValue", "utils.fromPairs"}}}},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream typed-array runtime regressions (lint and applied output).
			{
				Code: "JSON.stringify([1, 2].reduce((object, value) => ({...object, [value]: value}), {}));", FileName: "case.js",
				Output: []string{"JSON.stringify(Object.fromEntries([1, 2].flatMap((value) => [[value, value]])));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: "JSON.stringify(new Array(1, 2).reduce((object, value) => ({...object, [value]: value}), {}));", FileName: "case.js",
				Output: []string{"JSON.stringify(Object.fromEntries(new Array(1, 2).flatMap((value) => [[value, value]])));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code: "JSON.stringify(Array.from([1, 2]).reduce((object, value) => ({...object, [value]: value}), {}));", FileName: "case.js",
				Output: []string{"JSON.stringify(Object.fromEntries(Array.from([1, 2]).flatMap((value) => [[value, value]])));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: "JSON.stringify(Array.of(1, 2).reduce((object, value) => ({...object, [value]: value}), {}));", FileName: "case.js",
				Output: []string{"JSON.stringify(Object.fromEntries(Array.of(1, 2).flatMap((value) => [[value, value]])));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 37},
				},
			},
			// test.snapshot: reduce and for-of.
			{
				Code: "pairs.reduce(object => ({...object, key}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', key]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, /* keep this */ key}), {});", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, key}), {},);", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', key]],));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, key,}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', key]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, key}), (( {} )));", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', key]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce( (( object => ({...object, key}) )) , {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap( (( () => [['key', key]] )) ));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce( (( (object) => ({...object, key}) )) , {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap( (( () => [['key', key]] )) ));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce( (( (object,) => ({...object, key}) )) , {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap( (( () => [['key', key]] )) ));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [((key))] : ((value))}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[((key)), ((value))]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "((\n\t(( pairs ))\n\t.reduce(\n\t\t((\n\t\t\t(object,) => ((\n\t\t\t\t((\n\t\t\t\t\tObject\n\t\t\t\t)).assign(\n\t\t\t\t\t((\n\t\t\t\t\t\tobject\n\t\t\t\t\t)),\n\t\t\t\t\t(({\n\t\t\t\t\t\t[ ((key)) ] : ((value)),\n\t\t\t\t\t}))\n\t\t\t\t)\n\t\t\t))\n\t\t)),\n\t\t({})\n\t)\n));", FileName: "case.js",
				Output: []string{"((\n\tObject.fromEntries((( pairs ))\n\t.flatMap(\n\t\t((\n\t\t\t() => \n\t\t\t\t[[((key)), ((value))]]\n\t\t\t\n\t\t))\n\t))\n));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, 0: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[0, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, true: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['true', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, 0n: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[0n, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, undefined: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['undefined', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, null: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['null', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, var: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['var', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, for: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['for', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, default: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['default', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, string: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['string', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, \"string\": value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[\"string\", value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [0]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[0, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [true]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[true, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [0n]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[0n, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [undefined]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[undefined, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [null]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[null, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [\"for\"]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[\"for\", value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [string]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[string, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, [\"string\"]: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[\"string\", value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => Object.assign(object, {key}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', key]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => Object.assign(object, {key,}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', key]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => Object.assign(object, {[key]: value,}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [[key, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce((object, element, index, array) => ({...object, key}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap((element, index, array) => [['key', key]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce((object, [key, value], index, array,) => ({...object, [key]: value + index + array.length}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(([key, value], index, array,) => [[key, value + index + array.length]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, key: function (object) { return object; }}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['key', function (object) { return object; }]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, method: async () => {}}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['method', async () => {}]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, method: async function * (){}}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['method', async function * (){}]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = ({});\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js",
				Output: []string{"const object = (Object.fromEntries(pairs));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {\n\t// Comment\n};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 4, Column: 1, EndLine: 6, EndColumn: 2},
				},
			},
			{
				Code: "const object = {};\n// Comment\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 3, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: "const object = {};\nfor (const [key, value] of pairs) {\n\t// Comment\n\tobject[key] = value;\n}", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n} // Comment", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {};\nfor (const [key, value] of ((pairs))) {\n\tobject[key] = value;\n}", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(((pairs)));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[(key)] = (value);\n}", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {};\nfor (let [key, value] of pairs) {\n\tobject[key] = value;\n}\nconsole.log(object);", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs);\nconsole.log(object);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n[1].forEach(foo);", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs)\n;\n[1].forEach(foo);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n(foo)();", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs)\n;\n(foo)();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n`foo`;", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs)\n;\n`foo`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n+foo;", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs)\n;\n+foo;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n/foo/.test(bar);", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs)\n;\n/foo/.test(bar);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\nfoo();", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs)\n;\nfoo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n[1].forEach(foo);", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs);\n[1].forEach(foo);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			// test.typescript.
			{
				Code: "array.reduce<Record<string, Data & {b?: string}>>((result, entry) => ({...result, [entry.id]: entry.data}), {});", FileName: "case.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			// test.snapshot: functions.
			{
				Code: "_.fromPairs(pairs)", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `_.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: "lodash.fromPairs(pairs)", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `lodash.fromPairs()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: "myFromPairsFunction(pairs)", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{"myFromPairsFunction"}}},
				Output:  []string{"Object.fromEntries(pairs)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `myFromPairsFunction()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: "utils.object.foo(pairs)", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{"utils.object.foo"}}},
				Output:  []string{"Object.fromEntries(pairs)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `utils.object.foo()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// test: prototype properties and comments.
			{
				Code: "pairs.reduce(object => ({...object, ['__proto__']: value}), {});", FileName: "case.js",
				Output: []string{"Object.fromEntries(pairs.flatMap(() => [['__proto__', value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, key}),/* keep */ {});", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "pairs.reduce(object => ({...object, key}), /* keep */ {});", FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: "const object = pairs.reduce((object, [key, value]) => ({...object, [key]: value}), {});", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs.flatMap(([key, value]) => [[key, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: "const object = pairs.reduce((object, element, index, array) => ({...object, [key]: value}), {});", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs.flatMap((element, index, array) => [[key, value]]));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
				},
			},
			// test.snapshot: unbraced loop.
			{
				Code: "const object = {};\nfor (const [key, value] of pairs)\n\tobject[key] = value;", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 3, EndColumn: 22},
				},
			},
			// Documentation examples.
			{
				Code: "const object = pairs.reduce(\n\t(object, [key, value]) => ({...object, [key]: value}),\n\t{}\n);", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs.flatMap(\n\t([key, value]) => [[key, value]]\n));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: "const object = pairs.reduce(\n\t(object, [key, value]) => Object.assign(object, {[key]: value}),\n\t{}\n);", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs.flatMap(\n\t([key, value]) => [[key, value]]\n));"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "reduce", Message: "Prefer `Object.fromEntries()` over `Array#reduce()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: "const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loop", Message: "Prefer `Object.fromEntries()` over a `for-of` loop.", Line: 2, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "const object = _.fromPairs(pairs);", FileName: "case.js",
				Output: []string{"const object = Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `_.fromPairs()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: "const object = utils.fromPairs(pairs);", FileName: "case.js",
				Options: []any{map[string]any{"functions": []any{"utils.fromPairs"}}},
				Output:  []string{"const object = Object.fromEntries(pairs);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Prefer `Object.fromEntries()` over `utils.fromPairs()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 31},
				},
			},
		},
	)
}
