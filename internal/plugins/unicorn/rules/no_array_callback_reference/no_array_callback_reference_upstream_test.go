// Ported from eslint-plugin-unicorn v76.0.0 test/no-array-callback-reference.js
// and docs/rules/no-array-callback-reference.md. All cases and fixture expansions are retained.
// Expectations were checked against the pinned ESLint rule, including snapshot cases.
package no_array_callback_reference_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_callback_reference"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArrayCallbackReferenceUpstreamJavaScript(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_callback_reference.NoArrayCallbackReferenceRule,
		[]rule_tester.ValidTestCase{
			{Code: "foo.every(element => fn(element))", FileName: "file.js"},
			{Code: "foo.filter(element => fn(element))", FileName: "file.js"},
			{Code: "foo.find(element => fn(element))", FileName: "file.js"},
			{Code: "foo.findIndex(element => fn(element))", FileName: "file.js"},
			{Code: "foo.findLast(element => fn(element))", FileName: "file.js"},
			{Code: "foo.findLastIndex(element => fn(element))", FileName: "file.js"},
			{Code: "foo.flatMap(element => fn(element))", FileName: "file.js"},
			{Code: "foo.forEach(element => fn(element))", FileName: "file.js"},
			{Code: "foo.map(element => fn(element))", FileName: "file.js"},
			{Code: "foo.some(element => fn(element))", FileName: "file.js"},
			{Code: "foo.reduce((accumulator, element) => fn(element))", FileName: "file.js"},
			{Code: "foo.reduceRight((accumulator, element) => fn(element))", FileName: "file.js"},
			{Code: "foo?.every(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.filter(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.find(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.findIndex(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.findLast(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.findLastIndex(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.flatMap(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.forEach(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.map(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.some(element => fn(element))", FileName: "file.js"},
			{Code: "foo?.reduce((accumulator, element) => fn(element))", FileName: "file.js"},
			{Code: "foo?.reduceRight((accumulator, element) => fn(element))", FileName: "file.js"},
			{Code: "this.every(fn)", FileName: "file.js"},
			{Code: "this.filter(fn)", FileName: "file.js"},
			{Code: "this.find(fn)", FileName: "file.js"},
			{Code: "this.findIndex(fn)", FileName: "file.js"},
			{Code: "this.findLast(fn)", FileName: "file.js"},
			{Code: "this.findLastIndex(fn)", FileName: "file.js"},
			{Code: "this.flatMap(fn)", FileName: "file.js"},
			{Code: "this.forEach(fn)", FileName: "file.js"},
			{Code: "this.map(fn)", FileName: "file.js"},
			{Code: "this.some(fn)", FileName: "file.js"},
			{Code: "this.reduce(fn)", FileName: "file.js"},
			{Code: "this.reduceRight(fn)", FileName: "file.js"},
			{Code: "foo.find(Boolean)", FileName: "file.js"},
			{Code: "foo.some(Boolean)", FileName: "file.js"},
			{Code: "foo.map(String)", FileName: "file.js"},
			{Code: "foo.map(Number)", FileName: "file.js"},
			{Code: "foo.map(BigInt)", FileName: "file.js"},
			{Code: "foo.map(Boolean)", FileName: "file.js"},
			{Code: "foo.map(Symbol)", FileName: "file.js"},
			{Code: "new foo.map(fn);", FileName: "file.js"},
			{Code: "map(fn);", FileName: "file.js"},
			{Code: "foo['map'](fn);", FileName: "file.js"},
			{Code: "foo[map](fn);", FileName: "file.js"},
			{Code: "foo.notListedMethod(fn);", FileName: "file.js"},
			{Code: "foo.map();", FileName: "file.js"},
			{Code: "foo.map(fn, extraArgument1, extraArgument2);", FileName: "file.js"},
			{Code: "foo.map(...argumentsArray)", FileName: "file.js"},
			{Code: "Promise.map(fn)", FileName: "file.js"},
			{Code: "Promise.forEach(fn)", FileName: "file.js"},
			{Code: "lodash.map(fn)", FileName: "file.js"},
			{Code: "underscore.map(fn)", FileName: "file.js"},
			{Code: "_.map(fn)", FileName: "file.js"},
			{Code: "Async.map(list, fn)", FileName: "file.js"},
			{Code: "async.map(list, fn)", FileName: "file.js"},
			{Code: "React.Children.forEach(children, fn)", FileName: "file.js"},
			{Code: "Children.forEach(children, fn)", FileName: "file.js"},
			{Code: "Vue.filter(name, fn)", FileName: "file.js"},
			{Code: "$(this).find(tooltip)", FileName: "file.js"},
			{Code: "$.map(realArray, function(value, index) {});", FileName: "file.js"},
			{Code: "$(this).filter(tooltip)", FileName: "file.js"},
			{Code: "jQuery(this).find(tooltip)", FileName: "file.js"},
			{Code: "jQuery.map(realArray, function(value, index) {});", FileName: "file.js"},
			{Code: "jQuery(this).filter(tooltip)", FileName: "file.js"},
			{Code: "Angular.forEach(list, fn)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"Angular"}}}},
			{Code: "P.map(list, fn)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"P"}}}},
			{Code: "myLib.utils.map(list, fn)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"myLib.utils"}}}},
			{Code: "myLib(args).map(fn)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"myLib"}}}},
			{Code: "Promise.map(list, fn)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"Angular"}}}},
			{Code: "lodash.map(list, fn)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"Angular"}}}},
			{Code: "foo.map([])", FileName: "file.js"},
			{Code: "foo.map([element])", FileName: "file.js"},
			{Code: "foo.map([...elements])", FileName: "file.js"},
			{Code: "foo.map(1 + fn)", FileName: "file.js"},
			{Code: "foo.map(\"length\" in fn)", FileName: "file.js"},
			{Code: "foo.map(fn instanceof Function)", FileName: "file.js"},
			{Code: "foo.map(class ClassCantUseAsFunction {})", FileName: "file.js"},
			{Code: "foo.map(0)", FileName: "file.js"},
			{Code: "foo.map(1)", FileName: "file.js"},
			{Code: "foo.map(0.1)", FileName: "file.js"},
			{Code: "foo.map(\"\")", FileName: "file.js"},
			{Code: "foo.map(\"string\")", FileName: "file.js"},
			{Code: "foo.map(/regex/)", FileName: "file.js"},
			{Code: "foo.map(null)", FileName: "file.js"},
			{Code: "foo.map(0n)", FileName: "file.js"},
			{Code: "foo.map(1n)", FileName: "file.js"},
			{Code: "foo.map(true)", FileName: "file.js"},
			{Code: "foo.map(false)", FileName: "file.js"},
			{Code: "foo.map({})", FileName: "file.js"},
			{Code: "foo.map(`templateLiteral`)", FileName: "file.js"},
			{Code: "foo.map(undefined)", FileName: "file.js"},
			{Code: "foo.map(- fn)", FileName: "file.js"},
			{Code: "foo.map(+ fn)", FileName: "file.js"},
			{Code: "foo.map(~ fn)", FileName: "file.js"},
			{Code: "foo.map(typeof fn)", FileName: "file.js"},
			{Code: "foo.map(void fn)", FileName: "file.js"},
			{Code: "foo.map(delete foo.fn)", FileName: "file.js"},
			{Code: "foo.map(++ fn)", FileName: "file.js"},
			{Code: "foo.map(-- fn)", FileName: "file.js"},
			{Code: "foo.map(a = fn)", FileName: "file.js"},
			{Code: "foo.map(fn())", FileName: "file.js"},
			{Code: "foo.map(new ClassReturnsFunction())", FileName: "file.js"},
			{Code: "foo.map(new Function())", FileName: "file.js"},
			{Code: "foo.map(fn``)", FileName: "file.js"},
			{Code: "foo.map(this)", FileName: "file.js"},
			{Code: "const query = {}; model.find(query)", FileName: "file.js"},
			{Code: "const taskName = \"task\"; service.find(taskName)", FileName: "file.js"},
			{Code: "const values = []; collection.map(values)", FileName: "file.js"},
			{Code: "const NotCallable = class {}; collection.map(NotCallable)", FileName: "file.js"},
			{Code: "const index = 1 + 1; collection.findIndex(index)", FileName: "file.js"},
			{Code: "const query = {}; const criteria = query; model.find(criteria)", FileName: "file.js"},
			{Code: "foo.map(() => {})", FileName: "file.js"},
			{Code: "foo.map(function() {})", FileName: "file.js"},
			{Code: "foo.map(function bar() {})", FileName: "file.js"},
			{Code: "(async () => await foo.every(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.filter(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.find(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.findIndex(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.findLast(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.findLastIndex(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.flatMap(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.forEach(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.map(bar))()", FileName: "file.js"},
			{Code: "(async () => await foo.some(bar))()", FileName: "file.js"},
			{Code: "foo.map(function (a) {}.bind(bar))", FileName: "file.js"},
			{Code: "async function foo() {\n\tconst clientId = 20\n\tconst client = await oidc.Client.find(clientId)\n}", FileName: "file.js"},
			{Code: "const results = collection\n\t.find({\n\t\t$and: [cursorQuery, params.query]\n\t}, {\n\t\tprojection: params.projection\n\t})\n\t.sort($sort)\n\t.limit(params.limit + 1)\n\t.toArray()", FileName: "file.js"},
			{Code: "const EventsStore = types.model('EventsStore', {\n\tevents: types.optional(types.map(Event), {}),\n})", FileName: "file.js"},
			{Code: "const collection = new Set(); collection.forEach(callback);", FileName: "file.js"},
			{Code: "const collection = new Map(); collection.forEach(callback);", FileName: "file.js"},
			{Code: "class Foo {} const collection = new Foo(); collection.map(callback);", FileName: "file.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "foo.every(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.every(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.every((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.every((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.every((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.filter(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.filter(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.filter((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.filter((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.filter((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.find(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.find(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.find((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.find((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.find((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.findIndex(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findIndex(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.findIndex((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.findIndex((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.findIndex((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.findLast(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findLast(…)`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.findLast((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.findLast((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.findLast((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.findLastIndex(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findLastIndex(…)`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.findLastIndex((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.findLastIndex((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.findLastIndex((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.flatMap(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.flatMap(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.flatMap((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.flatMap((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.flatMap((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.map(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.map((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.map((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.map((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.some(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.some(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.some((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo.some((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo.some((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.every(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.every(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.every((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.every((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.every((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.filter(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.filter(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.filter((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.filter((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.filter((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.find(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.find(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.find((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.find((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.find((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.findIndex(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findIndex(…)`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.findIndex((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.findIndex((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.findIndex((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.findLast(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findLast(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.findLast((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.findLast((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.findLast((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.findLastIndex(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findLastIndex(…)`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.findLastIndex((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.findLastIndex((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.findLastIndex((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.flatMap(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.flatMap(…)`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.flatMap((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.flatMap((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.flatMap((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.map(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.map((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.map((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.map((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo?.some(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.some(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo?.some((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "foo?.some((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "foo?.some((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.forEach(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.forEach(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.forEach((element) => { fn(element); })"},
					{MessageId: "replace-with-name", Output: "foo.forEach((element, index) => { fn(element, index); })"},
					{MessageId: "replace-with-name", Output: "foo.forEach((element, index, array) => { fn(element, index, array); })"},
				}},
			}},
			{Code: "const callback = value => value; array.map(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const callback = value => value; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const callback = value => value; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const callback = value => value; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "let query = {}; model.find(query);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `query` directly to `.find(…)`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "let query = {}; model.find((element) => query(element));"},
					{MessageId: "replace-with-name", Output: "let query = {}; model.find((element, index) => query(element, index));"},
					{MessageId: "replace-with-name", Output: "let query = {}; model.find((element, index, array) => query(element, index, array));"},
				}},
			}},
			{Code: "array.map(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.map((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "array.map((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "array.map((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "foo.map(element)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `element` directly to `.map(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.map((element_) => element(element_))"},
					{MessageId: "replace-with-name", Output: "foo.map((element_, index) => element(element_, index))"},
					{MessageId: "replace-with-name", Output: "foo.map((element_, index, array) => element(element_, index, array))"},
				}},
			}},
			{Code: "items.map(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "items.map((item) => fn(item))"},
					{MessageId: "replace-with-name", Output: "items.map((item, index) => fn(item, index))"},
					{MessageId: "replace-with-name", Output: "items.map((item, index, items) => fn(item, index, items))"},
				}},
			}},
			{Code: "items.map(item)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `item` directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "items.map((element) => item(element))"},
					{MessageId: "replace-with-name", Output: "items.map((element, index) => item(element, index))"},
					{MessageId: "replace-with-name", Output: "items.map((element, index, items) => item(element, index, items))"},
				}},
			}},
			{Code: "items.map(items)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `items` directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "items.map((item) => items(item))"},
					{MessageId: "replace-with-name", Output: "items.map((item, index) => items(item, index))"},
					{MessageId: "replace-with-name", Output: "items.map((item, index, array) => items(item, index, array))"},
				}},
			}},
			{Code: "items.map(index)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `index` directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "items.map((item) => index(item))"},
					{MessageId: "replace-with-name", Output: "items.map((item, index_) => index(item, index_))"},
					{MessageId: "replace-with-name", Output: "items.map((item, index_, items) => index(item, index_, items))"},
				}},
			}},
			{Code: "items.map(item.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "items.map((element) => item.fn(element))"},
					{MessageId: "replace-without-name", Output: "items.map((element, index) => item.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "items.map((element, index, array) => item.fn(element, index, array))"},
				}},
			}},
			{Code: "classes.map(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "classes.map((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "classes.map((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "classes.map((element, index, classes) => fn(element, index, classes))"},
				}},
			}},
			{Code: "indices.map(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "indices.map((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "indices.map((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "indices.map((element, index, indices) => fn(element, index, indices))"},
				}},
			}},
			{Code: "items.forEach(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.forEach(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "items.forEach((item) => { fn(item); })"},
					{MessageId: "replace-with-name", Output: "items.forEach((item, index) => { fn(item, index); })"},
					{MessageId: "replace-with-name", Output: "items.forEach((item, index, items) => { fn(item, index, items); })"},
				}},
			}},
			{Code: "items.reduce(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.reduce(…)`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "items.reduce((accumulator, item) => fn(accumulator, item))"},
					{MessageId: "replace-with-name", Output: "items.reduce((accumulator, item, index) => fn(accumulator, item, index))"},
					{MessageId: "replace-with-name", Output: "items.reduce((accumulator, item, index, items) => fn(accumulator, item, index, items))"},
				}},
			}},
			{Code: "items.reduceRight(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.reduceRight(…)`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "items.reduceRight((accumulator, item) => fn(accumulator, item))"},
					{MessageId: "replace-with-name", Output: "items.reduceRight((accumulator, item, index) => fn(accumulator, item, index))"},
					{MessageId: "replace-with-name", Output: "items.reduceRight((accumulator, item, index, items) => fn(accumulator, item, index, items))"},
				}},
			}},
			{Code: "items.reduce(accumulator)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `accumulator` directly to `.reduce(…)`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "items.reduce((accumulator_, item) => accumulator(accumulator_, item))"},
					{MessageId: "replace-with-name", Output: "items.reduce((accumulator_, item, index) => accumulator(accumulator_, item, index))"},
					{MessageId: "replace-with-name", Output: "items.reduce((accumulator_, item, index, items) => accumulator(accumulator_, item, index, items))"},
				}},
			}},
			{Code: "foo.reduce(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.reduce(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element) => fn(accumulator, element))"},
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element, index) => fn(accumulator, element, index))"},
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element, index, array) => fn(accumulator, element, index, array))"},
				}},
			}},
			{Code: "foo.reduceRight(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.reduceRight(…)`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element) => fn(accumulator, element))"},
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element, index) => fn(accumulator, element, index))"},
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element, index, array) => fn(accumulator, element, index, array))"},
				}},
			}},
			{Code: "foo.every(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.every(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.every((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.every((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.every((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.filter(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.filter(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.filter((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.filter((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.filter((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.find(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.find(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.find((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.find((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.find((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.findIndex(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findIndex(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.findIndex((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.findIndex((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.findIndex((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.findLast(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findLast(…)`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.findLast((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.findLast((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.findLast((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.findLastIndex(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.findLastIndex(…)`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.findLastIndex((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.findLastIndex((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.findLastIndex((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.flatMap(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.flatMap(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.flatMap((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.flatMap((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.flatMap((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.map(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.map((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.map((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.map((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.some(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.some(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.some((element) => fn(element), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.some((element, index) => fn(element, index), thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.some((element, index, array) => fn(element, index, array), thisArgument)"},
				}},
			}},
			{Code: "foo.forEach(fn, thisArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.forEach(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.forEach((element) => { fn(element); }, thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.forEach((element, index) => { fn(element, index); }, thisArgument)"},
					{MessageId: "replace-with-name", Output: "foo.forEach((element, index, array) => { fn(element, index, array); }, thisArgument)"},
				}},
			}},
			{Code: "foo.reduce(fn, initialValue)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.reduce(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element) => fn(accumulator, element), initialValue)"},
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element, index) => fn(accumulator, element, index), initialValue)"},
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element, index, array) => fn(accumulator, element, index, array), initialValue)"},
				}},
			}},
			{Code: "foo.reduceRight(fn, initialValue)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.reduceRight(…)`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element) => fn(accumulator, element), initialValue)"},
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element, index) => fn(accumulator, element, index), initialValue)"},
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element, index, array) => fn(accumulator, element, index, array), initialValue)"},
				}},
			}},
			{Code: "foo.reduce(Boolean, initialValue)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `Boolean` directly to `.reduce(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element) => Boolean(accumulator, element), initialValue)"},
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element, index) => Boolean(accumulator, element, index), initialValue)"},
					{MessageId: "replace-with-name", Output: "foo.reduce((accumulator, element, index, array) => Boolean(accumulator, element, index, array), initialValue)"},
				}},
			}},
			{Code: "foo.reduceRight(Boolean, initialValue)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `Boolean` directly to `.reduceRight(…)`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element) => Boolean(accumulator, element), initialValue)"},
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element, index) => Boolean(accumulator, element, index), initialValue)"},
					{MessageId: "replace-with-name", Output: "foo.reduceRight((accumulator, element, index, array) => Boolean(accumulator, element, index, array), initialValue)"},
				}},
			}},
			{Code: "foo.forEach(Boolean)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `Boolean` directly to `.forEach(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.forEach((element) => { Boolean(element); })"},
					{MessageId: "replace-with-name", Output: "foo.forEach((element, index) => { Boolean(element, index); })"},
					{MessageId: "replace-with-name", Output: "foo.forEach((element, index, array) => { Boolean(element, index, array); })"},
				}},
			}},
			{Code: "foo.every(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.every(…)`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.every((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.every((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.every((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.filter(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.filter(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.filter((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.filter((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.filter((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.find(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.find(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.find((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.find((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.find((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.findIndex(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.findIndex(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.findIndex((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.findIndex((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.findIndex((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.findLast(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.findLast(…)`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.findLast((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.findLast((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.findLast((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.findLastIndex(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.findLastIndex(…)`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.findLastIndex((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.findLastIndex((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.findLastIndex((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.flatMap(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.flatMap(…)`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.flatMap((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.flatMap((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.flatMap((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.map(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.map((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.map((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.map((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.some(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.some(…)`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.some((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.some((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.some((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.reduce(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.reduce(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.reduce((accumulator, element) => lib.fn(accumulator, element))"},
					{MessageId: "replace-without-name", Output: "foo.reduce((accumulator, element, index) => lib.fn(accumulator, element, index))"},
					{MessageId: "replace-without-name", Output: "foo.reduce((accumulator, element, index, array) => lib.fn(accumulator, element, index, array))"},
				}},
			}},
			{Code: "foo.reduceRight(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.reduceRight(…)`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.reduceRight((accumulator, element) => lib.fn(accumulator, element))"},
					{MessageId: "replace-without-name", Output: "foo.reduceRight((accumulator, element, index) => lib.fn(accumulator, element, index))"},
					{MessageId: "replace-without-name", Output: "foo.reduceRight((accumulator, element, index, array) => lib.fn(accumulator, element, index, array))"},
				}},
			}},
			{Code: "foo.map(a || b)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.map((element) => (a || b)(element))"},
					{MessageId: "replace-without-name", Output: "foo.map((element, index) => (a || b)(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.map((element, index, array) => (a || b)(element, index, array))"},
				}},
			}},
			{Code: "array.map(condition ? toFile : toBuffer);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `toFile` directly to `.map(…)`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.map(condition ? (element) => toFile(element) : toBuffer);"},
					{MessageId: "replace-with-name", Output: "array.map(condition ? (element, index) => toFile(element, index) : toBuffer);"},
					{MessageId: "replace-with-name", Output: "array.map(condition ? (element, index, array) => toFile(element, index, array) : toBuffer);"},
				}},
				{MessageId: "error-with-name", Message: "Do not pass function `toBuffer` directly to `.map(…)`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.map(condition ? toFile : (element) => toBuffer(element));"},
					{MessageId: "replace-with-name", Output: "array.map(condition ? toFile : (element, index) => toBuffer(element, index));"},
					{MessageId: "replace-with-name", Output: "array.map(condition ? toFile : (element, index, array) => toBuffer(element, index, array));"},
				}},
			}},
			{Code: "function * foo() { array.map(condition ? (yield toFile) : toBuffer); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 55},
				{MessageId: "error-with-name", Message: "Do not pass function `toBuffer` directly to `.map(…)`.", Line: 1, Column: 59, EndLine: 1, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "function * foo() { array.map(condition ? (yield toFile) : (element) => toBuffer(element)); }"},
					{MessageId: "replace-with-name", Output: "function * foo() { array.map(condition ? (yield toFile) : (element, index) => toBuffer(element, index)); }"},
					{MessageId: "replace-with-name", Output: "function * foo() { array.map(condition ? (yield toFile) : (element, index, array) => toBuffer(element, index, array)); }"},
				}},
			}},
			{Code: "async function foo() { array.map((await condition) ? toFile : toBuffer); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `toFile` directly to `.map(…)`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "async function foo() { array.map((await condition) ? (element) => toFile(element) : toBuffer); }"},
					{MessageId: "replace-with-name", Output: "async function foo() { array.map((await condition) ? (element, index) => toFile(element, index) : toBuffer); }"},
					{MessageId: "replace-with-name", Output: "async function foo() { array.map((await condition) ? (element, index, array) => toFile(element, index, array) : toBuffer); }"},
				}},
				{MessageId: "error-with-name", Message: "Do not pass function `toBuffer` directly to `.map(…)`.", Line: 1, Column: 63, EndLine: 1, EndColumn: 71, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "async function foo() { array.map((await condition) ? toFile : (element) => toBuffer(element)); }"},
					{MessageId: "replace-with-name", Output: "async function foo() { array.map((await condition) ? toFile : (element, index) => toBuffer(element, index)); }"},
					{MessageId: "replace-with-name", Output: "async function foo() { array.map((await condition) ? toFile : (element, index, array) => toBuffer(element, index, array)); }"},
				}},
			}},
			{Code: "bar.map(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "bar.map((element) => fn(element))"},
					{MessageId: "replace-with-name", Output: "bar.map((element, index) => fn(element, index))"},
					{MessageId: "replace-with-name", Output: "bar.map((element, index, array) => fn(element, index, array))"},
				}},
			}},
			{Code: "bar.reduce(fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.reduce(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "bar.reduce((accumulator, element) => fn(accumulator, element))"},
					{MessageId: "replace-with-name", Output: "bar.reduce((accumulator, element, index) => fn(accumulator, element, index))"},
					{MessageId: "replace-with-name", Output: "bar.reduce((accumulator, element, index, array) => fn(accumulator, element, index, array))"},
				}},
			}},
			{Code: "foo.map(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.map((element) => lib.fn(element))"},
					{MessageId: "replace-without-name", Output: "foo.map((element, index) => lib.fn(element, index))"},
					{MessageId: "replace-without-name", Output: "foo.map((element, index, array) => lib.fn(element, index, array))"},
				}},
			}},
			{Code: "foo.reduce(lib.fn)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.reduce(…)`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "foo.reduce((accumulator, element) => lib.fn(accumulator, element))"},
					{MessageId: "replace-without-name", Output: "foo.reduce((accumulator, element, index) => lib.fn(accumulator, element, index))"},
					{MessageId: "replace-without-name", Output: "foo.reduce((accumulator, element, index, array) => lib.fn(accumulator, element, index, array))"},
				}},
			}},
			{Code: "const fn = async () => {\n\tawait Promise.all(foo.map(toPromise));\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `toPromise` directly to `.map(…)`.", Line: 2, Column: 28, EndLine: 2, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const fn = async () => {\n\tawait Promise.all(foo.map((element) => toPromise(element)));\n}"},
					{MessageId: "replace-with-name", Output: "const fn = async () => {\n\tawait Promise.all(foo.map((element, index) => toPromise(element, index)));\n}"},
					{MessageId: "replace-with-name", Output: "const fn = async () => {\n\tawait Promise.all(foo.map((element, index, array) => toPromise(element, index, array)));\n}"},
				}},
			}},
			{Code: "async function fn() {\n\tfor await (const foo of bar.map(toPromise)) {}\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `toPromise` directly to `.map(…)`.", Line: 2, Column: 34, EndLine: 2, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "async function fn() {\n\tfor await (const foo of bar.map((element) => toPromise(element))) {}\n}"},
					{MessageId: "replace-with-name", Output: "async function fn() {\n\tfor await (const foo of bar.map((element, index) => toPromise(element, index))) {}\n}"},
					{MessageId: "replace-with-name", Output: "async function fn() {\n\tfor await (const foo of bar.map((element, index, array) => toPromise(element, index, array))) {}\n}"},
				}},
			}},
			{Code: "async function fn() {\n\tawait foo.reduce(foo, Promise.resolve())\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `foo` directly to `.reduce(…)`.", Line: 2, Column: 19, EndLine: 2, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "async function fn() {\n\tawait foo.reduce((accumulator, element) => foo(accumulator, element), Promise.resolve())\n}"},
					{MessageId: "replace-with-name", Output: "async function fn() {\n\tawait foo.reduce((accumulator, element, index) => foo(accumulator, element, index), Promise.resolve())\n}"},
					{MessageId: "replace-with-name", Output: "async function fn() {\n\tawait foo.reduce((accumulator, element, index, array) => foo(accumulator, element, index, array), Promise.resolve())\n}"},
				}},
			}},
			{Code: "const fn = (x, y) => x + y;\n[1, 2, 3].map(fn);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.map(…)`.", Line: 2, Column: 15, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const fn = (x, y) => x + y;\n[1, 2, 3].map((element) => fn(element));"},
					{MessageId: "replace-with-name", Output: "const fn = (x, y) => x + y;\n[1, 2, 3].map((element, index) => fn(element, index));"},
					{MessageId: "replace-with-name", Output: "const fn = (x, y) => x + y;\n[1, 2, 3].map((element, index, array) => fn(element, index, array));"},
				}},
			}},
			{Code: "Other.forEach(fn)", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"Angular"}}}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `fn` directly to `.forEach(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "Other.forEach((element) => { fn(element); })"},
					{MessageId: "replace-with-name", Output: "Other.forEach((element, index) => { fn(element, index); })"},
					{MessageId: "replace-with-name", Output: "Other.forEach((element, index, array) => { fn(element, index, array); })"},
				}},
			}},
		})
}

func TestNoArrayCallbackReferenceUpstreamTernaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_callback_reference.NoArrayCallbackReferenceRule,
		[]rule_tester.ValidTestCase{
			{Code: "foo.map(_ ? () => {} : _ ? () => {} : () => {})", FileName: "file.js"},
			{Code: "foo.reduce(_ ? () => {} : _ ? () => {} : () => {})", FileName: "file.js"},
			{Code: "foo.every(_ ? Boolean : _ ? Boolean : Boolean)", FileName: "file.js"},
			{Code: "foo.map(_ ? String : _ ? Number : Boolean)", FileName: "file.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "foo.map(\n\t_\n\t\t? String // This one should be ignored\n\t\t: callback\n);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 4, Column: 5, EndLine: 4, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.map(\n\t_\n\t\t? String // This one should be ignored\n\t\t: (element) => callback(element)\n);"},
					{MessageId: "replace-with-name", Output: "foo.map(\n\t_\n\t\t? String // This one should be ignored\n\t\t: (element, index) => callback(element, index)\n);"},
					{MessageId: "replace-with-name", Output: "foo.map(\n\t_\n\t\t? String // This one should be ignored\n\t\t: (element, index, array) => callback(element, index, array)\n);"},
				}},
			}},
			{Code: "foo.forEach(\n\t_\n\t\t? callbackA\n\t\t: _\n\t\t\t\t? callbackB\n\t\t\t\t: callbackC\n);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callbackA` directly to `.forEach(…)`.", Line: 3, Column: 5, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? (element) => { callbackA(element); }\n\t\t: _\n\t\t\t\t? callbackB\n\t\t\t\t: callbackC\n);"},
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? (element, index) => { callbackA(element, index); }\n\t\t: _\n\t\t\t\t? callbackB\n\t\t\t\t: callbackC\n);"},
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? (element, index, array) => { callbackA(element, index, array); }\n\t\t: _\n\t\t\t\t? callbackB\n\t\t\t\t: callbackC\n);"},
				}},
				{MessageId: "error-with-name", Message: "Do not pass function `callbackB` directly to `.forEach(…)`.", Line: 5, Column: 7, EndLine: 5, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? callbackA\n\t\t: _\n\t\t\t\t? (element) => { callbackB(element); }\n\t\t\t\t: callbackC\n);"},
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? callbackA\n\t\t: _\n\t\t\t\t? (element, index) => { callbackB(element, index); }\n\t\t\t\t: callbackC\n);"},
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? callbackA\n\t\t: _\n\t\t\t\t? (element, index, array) => { callbackB(element, index, array); }\n\t\t\t\t: callbackC\n);"},
				}},
				{MessageId: "error-with-name", Message: "Do not pass function `callbackC` directly to `.forEach(…)`.", Line: 6, Column: 7, EndLine: 6, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? callbackA\n\t\t: _\n\t\t\t\t? callbackB\n\t\t\t\t: (element) => { callbackC(element); }\n);"},
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? callbackA\n\t\t: _\n\t\t\t\t? callbackB\n\t\t\t\t: (element, index) => { callbackC(element, index); }\n);"},
					{MessageId: "replace-with-name", Output: "foo.forEach(\n\t_\n\t\t? callbackA\n\t\t: _\n\t\t\t\t? callbackB\n\t\t\t\t: (element, index, array) => { callbackC(element, index, array); }\n);"},
				}},
			}},
			{Code: "async function * foo () {\n\tfoo.map((0, bar));\n\tfoo.map(yield bar);\n\tfoo.map(yield* bar);\n\tfoo.map(() => bar);\n\tfoo.map(bar &&= baz);\n\tfoo.map(bar || baz);\n\tfoo.map(bar + bar);\n\tfoo.map(+ bar);\n\tfoo.map(++ bar);\n\tfoo.map(new Function(''));\n}", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 2, Column: 11, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "async function * foo () {\n\tfoo.map((element) => (0, bar)(element));\n\tfoo.map(yield bar);\n\tfoo.map(yield* bar);\n\tfoo.map(() => bar);\n\tfoo.map(bar &&= baz);\n\tfoo.map(bar || baz);\n\tfoo.map(bar + bar);\n\tfoo.map(+ bar);\n\tfoo.map(++ bar);\n\tfoo.map(new Function(''));\n}"},
					{MessageId: "replace-without-name", Output: "async function * foo () {\n\tfoo.map((element, index) => (0, bar)(element, index));\n\tfoo.map(yield bar);\n\tfoo.map(yield* bar);\n\tfoo.map(() => bar);\n\tfoo.map(bar &&= baz);\n\tfoo.map(bar || baz);\n\tfoo.map(bar + bar);\n\tfoo.map(+ bar);\n\tfoo.map(++ bar);\n\tfoo.map(new Function(''));\n}"},
					{MessageId: "replace-without-name", Output: "async function * foo () {\n\tfoo.map((element, index, array) => (0, bar)(element, index, array));\n\tfoo.map(yield bar);\n\tfoo.map(yield* bar);\n\tfoo.map(() => bar);\n\tfoo.map(bar &&= baz);\n\tfoo.map(bar || baz);\n\tfoo.map(bar + bar);\n\tfoo.map(+ bar);\n\tfoo.map(++ bar);\n\tfoo.map(new Function(''));\n}"},
				}},
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 3, Column: 10, EndLine: 3, EndColumn: 19},
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 4, Column: 10, EndLine: 4, EndColumn: 20},
				{MessageId: "error-without-name", Message: "Do not pass function directly to `.map(…)`.", Line: 7, Column: 10, EndLine: 7, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-without-name", Output: "async function * foo () {\n\tfoo.map((0, bar));\n\tfoo.map(yield bar);\n\tfoo.map(yield* bar);\n\tfoo.map(() => bar);\n\tfoo.map(bar &&= baz);\n\tfoo.map((element) => (bar || baz)(element));\n\tfoo.map(bar + bar);\n\tfoo.map(+ bar);\n\tfoo.map(++ bar);\n\tfoo.map(new Function(''));\n}"},
					{MessageId: "replace-without-name", Output: "async function * foo () {\n\tfoo.map((0, bar));\n\tfoo.map(yield bar);\n\tfoo.map(yield* bar);\n\tfoo.map(() => bar);\n\tfoo.map(bar &&= baz);\n\tfoo.map((element, index) => (bar || baz)(element, index));\n\tfoo.map(bar + bar);\n\tfoo.map(+ bar);\n\tfoo.map(++ bar);\n\tfoo.map(new Function(''));\n}"},
					{MessageId: "replace-without-name", Output: "async function * foo () {\n\tfoo.map((0, bar));\n\tfoo.map(yield bar);\n\tfoo.map(yield* bar);\n\tfoo.map(() => bar);\n\tfoo.map(bar &&= baz);\n\tfoo.map((element, index, array) => (bar || baz)(element, index, array));\n\tfoo.map(bar + bar);\n\tfoo.map(+ bar);\n\tfoo.map(++ bar);\n\tfoo.map(new Function(''));\n}"},
				}},
			}},
		})
}

func TestNoArrayCallbackReferenceUpstreamTypedReceivers(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_callback_reference.NoArrayCallbackReferenceRule,
		[]rule_tester.ValidTestCase{
			{Code: "interface SearchService {\n\tfind(callback: Function): unknown;\n}\ndeclare const callback: Function;\ndeclare const service: SearchService;\nservice.find(callback);", FileName: "file.ts"},
			{Code: "class SearchService {\n\tfind(taskName: string): unknown {\n\t\treturn taskName;\n\t}\n}\nconst service = new SearchService();\nconst taskName = 'task';\nservice.find(taskName);", FileName: "file.ts"},
			{Code: "declare const callback: Function;\nclass Collection {\n\tmap(callback: Function) {}\n}\nconst collection = new Collection();\ncollection.map(callback);", FileName: "file.ts"},
			{Code: "interface Model {\n\tfind(query: object): unknown;\n}\ndeclare const AccountModel: Model;\nconst query = {};\nAccountModel.find(query);", FileName: "file.ts"},
			{Code: "interface NgMocks {\n\tfind(component: unknown): unknown;\n}\ndeclare const ngMocks: NgMocks;\ndeclare const MyComponent: unknown;\nngMocks.find(MyComponent);", FileName: "file.ts"},
			{Code: "declare const callback: Function;\ndeclare const collection: string[] | {map(callback: Function): unknown};\ncollection.map(callback);", FileName: "file.ts"},
			{Code: "declare const callback: Function;\ndeclare const set: Set<string>;\nset.forEach(callback);", FileName: "file.ts"},
			{Code: "declare const callback: Function;\ndeclare const map: Map<string, string>;\nmap.forEach(callback);", FileName: "file.ts"},
			{Code: "declare const callback: Function;\ndeclare const service: {find(callback: Function): unknown} | undefined;\nservice?.find(callback);", FileName: "file.ts"},
			{Code: "export {};\ntype Array<T> = {map(callback: Function): unknown};\ndeclare const callback: Function;\ndeclare const collection: Array<string>;\ncollection.map(callback);", FileName: "file.ts"},
			{Code: "export {};\nclass Uint8Array {\n\tmap(callback: Function) {}\n}\ndeclare const callback: Function;\nconst collection = new Uint8Array();\ncollection.map(callback);", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "declare const callback: Function; declare const array: string[]; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 76, EndLine: 1, EndColumn: 84, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[]; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[]; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[]; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: readonly string[]; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 85, EndLine: 1, EndColumn: 93, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: readonly string[]; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: readonly string[]; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: readonly string[]; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: [string, string]; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 84, EndLine: 1, EndColumn: 92, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: [string, string]; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: [string, string]; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: [string, string]; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: Array<string>; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 81, EndLine: 1, EndColumn: 89, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: Array<string>; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: Array<string>; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: Array<string>; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: ReadonlyArray<string>; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 89, EndLine: 1, EndColumn: 97, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: ReadonlyArray<string>; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: ReadonlyArray<string>; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: ReadonlyArray<string>; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: Uint8Array; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 78, EndLine: 1, EndColumn: 86, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: Uint8Array; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: Uint8Array; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: Uint8Array; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: string[] | readonly number[]; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 96, EndLine: 1, EndColumn: 104, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | readonly number[]; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | readonly number[]; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | readonly number[]; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: string[] | Uint8Array; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 89, EndLine: 1, EndColumn: 97, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | Uint8Array; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | Uint8Array; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | Uint8Array; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: string[] & {foo: string}; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 92, EndLine: 1, EndColumn: 100, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] & {foo: string}; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] & {foo: string}; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] & {foo: string}; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: string[] | undefined; array?.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 89, EndLine: 1, EndColumn: 97, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | undefined; array?.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | undefined; array?.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: string[] | undefined; array?.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; function run<T extends string[]>(array: T) { array.map(callback); }", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 90, EndLine: 1, EndColumn: 98, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends string[]>(array: T) { array.map((element) => callback(element)); }"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends string[]>(array: T) { array.map((element, index) => callback(element, index)); }"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends string[]>(array: T) { array.map((element, index, array) => callback(element, index, array)); }"},
				}},
			}},
			{Code: "declare const callback: Function; function run<T extends readonly string[]>(array: T) { array.map(callback); }", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 99, EndLine: 1, EndColumn: 107, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends readonly string[]>(array: T) { array.map((element) => callback(element)); }"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends readonly string[]>(array: T) { array.map((element, index) => callback(element, index)); }"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends readonly string[]>(array: T) { array.map((element, index, array) => callback(element, index, array)); }"},
				}},
			}},
			{Code: "declare const callback: Function; class Strings extends Array<string> {} const array = new Strings(); array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 113, EndLine: 1, EndColumn: 121, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; class Strings extends Array<string> {} const array = new Strings(); array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; class Strings extends Array<string> {} const array = new Strings(); array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; class Strings extends Array<string> {} const array = new Strings(); array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; interface S extends ReadonlyArray<string> {} declare const array: S; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 114, EndLine: 1, EndColumn: 122, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; interface S extends ReadonlyArray<string> {} declare const array: S; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; interface S extends ReadonlyArray<string> {} declare const array: S; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; interface S extends ReadonlyArray<string> {} declare const array: S; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; function run<T extends Uint8Array>(array: T) { array.map(callback); }", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 92, EndLine: 1, EndColumn: 100, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends Uint8Array>(array: T) { array.map((element) => callback(element)); }"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends Uint8Array>(array: T) { array.map((element, index) => callback(element, index)); }"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; function run<T extends Uint8Array>(array: T) { array.map((element, index, array) => callback(element, index, array)); }"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: any; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 71, EndLine: 1, EndColumn: 79, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: any; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: any; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: any; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: unknown; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 75, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: unknown; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: unknown; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: unknown; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const callback: Function; declare const array: MissingType; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 79, EndLine: 1, EndColumn: 87, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: MissingType; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: MissingType; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const callback: Function; declare const array: MissingType; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
		})
}

func TestNoArrayCallbackReferenceUpstreamTypeScriptSyntax(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_callback_reference.NoArrayCallbackReferenceRule,
		[]rule_tester.ValidTestCase{
			{Code: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.filter(isString);", FileName: "file.ts"},
			{Code: "const isString = (value: unknown): value is string => typeof value === 'string';\nfoo.filter(isString);", FileName: "file.ts"},
			{Code: "const isString = function (value: unknown): value is string {\n\treturn typeof value === 'string';\n};\nfoo.filter(isString);", FileName: "file.ts"},
			{Code: "const isString: (value: unknown) => value is string = value => typeof value === 'string';\nfoo.filter(isString);", FileName: "file.ts"},
			{Code: "function run(predicate: (value: unknown) => value is string) {\n\tfoo.filter(predicate);\n}", FileName: "file.ts"},
			{Code: "function runEvery(predicate: (value: unknown) => value is string) {\n\tfoo.every(predicate);\n}", FileName: "file.ts"},
			{Code: "function runFind(predicate: (value: unknown) => value is string) {\n\tfoo.find(predicate);\n}", FileName: "file.ts"},
			{Code: "function runFindLast(predicate: (value: unknown) => value is string) {\n\tfoo.findLast(predicate);\n}", FileName: "file.ts"},
			{Code: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nconst guard: (value: unknown) => value is string = isString;\nfoo.filter(guard);", FileName: "file.ts"},
			{Code: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.find(isString);", FileName: "file.ts"},
			{Code: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.every(isString);", FileName: "file.ts"},
			{Code: "import {isString} from './guards';\nfoo.filter(isString);", FileName: "file.ts"},
			{Code: "import {isString} from './guards';\nfoo.find(isString);", FileName: "file.ts"},
			{Code: "import {isString} from './guards';\nfoo.findLast(isString);", FileName: "file.ts"},
			{Code: "import {isString} from './guards';\nfoo.every(isString);", FileName: "file.ts"},
			{Code: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfunction isNumber(value: unknown): value is number {\n\treturn typeof value === 'number';\n}\nfoo.filter(condition ? isString : isNumber);", FileName: "file.ts"},
			{Code: "const query = {} as const;\nmodel.find(query);", FileName: "file.ts"},
			{Code: "const taskName = 'task' satisfies string;\nservice.find(taskName);", FileName: "file.ts"},
			{Code: "declare const collection: Set<string>; collection.forEach(callback);", FileName: "file.ts"},
			{Code: "declare const collection: Map<string, string>; collection.forEach(callback);", FileName: "file.ts"},
			{Code: "function run(collection: ReadonlySet<string>) { collection.forEach(callback); }", FileName: "file.ts"},
			{Code: "class Uint8Array {\n\tmap(callback: Function) {}\n}\nconst collection = new Uint8Array();\ncollection.map(callback);", FileName: "file.ts"},
			{Code: "declare const collection: string[] | Set<string>; collection.forEach(callback);", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "declare const array: string[]; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 42, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const array: string[]; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const array: string[]; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const array: string[]; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const array: readonly string[]; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 51, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const array: readonly string[]; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const array: readonly string[]; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const array: readonly string[]; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const array: Array<string>; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 47, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const array: Array<string>; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const array: Array<string>; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const array: Array<string>; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const array: Uint8Array; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const array: Uint8Array; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const array: Uint8Array; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const array: Uint8Array; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "const array = new Uint8Array(); array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const array = new Uint8Array(); array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const array = new Uint8Array(); array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const array = new Uint8Array(); array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const array: string[] | Uint8Array; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 55, EndLine: 1, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const array: string[] | Uint8Array; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const array: string[] | Uint8Array; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const array: string[] | Uint8Array; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "declare const array: string[] | undefined; array.map(callback);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "declare const array: string[] | undefined; array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "declare const array: string[] | undefined; array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "declare const array: string[] | undefined; array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "function isString(value: unknown): boolean {\n\treturn typeof value === 'string';\n}\nfoo.filter(isString);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `isString` directly to `.filter(…)`.", Line: 4, Column: 12, EndLine: 4, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): boolean {\n\treturn typeof value === 'string';\n}\nfoo.filter((element) => isString(element));"},
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): boolean {\n\treturn typeof value === 'string';\n}\nfoo.filter((element, index) => isString(element, index));"},
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): boolean {\n\treturn typeof value === 'string';\n}\nfoo.filter((element, index, array) => isString(element, index, array));"},
				}},
			}},
			{Code: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.map(isString);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `isString` directly to `.map(…)`.", Line: 4, Column: 9, EndLine: 4, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.map((element) => isString(element));"},
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.map((element, index) => isString(element, index));"},
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.map((element, index, array) => isString(element, index, array));"},
				}},
			}},
			{Code: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfunction isObject(value: unknown): boolean {\n\treturn typeof value === 'object';\n}\nfoo.filter(condition ? isString : isObject);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `isObject` directly to `.filter(…)`.", Line: 7, Column: 35, EndLine: 7, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfunction isObject(value: unknown): boolean {\n\treturn typeof value === 'object';\n}\nfoo.filter(condition ? isString : (element) => isObject(element));"},
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfunction isObject(value: unknown): boolean {\n\treturn typeof value === 'object';\n}\nfoo.filter(condition ? isString : (element, index) => isObject(element, index));"},
					{MessageId: "replace-with-name", Output: "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfunction isObject(value: unknown): boolean {\n\treturn typeof value === 'object';\n}\nfoo.filter(condition ? isString : (element, index, array) => isObject(element, index, array));"},
				}},
			}},
		})
}

func TestNoArrayCallbackReferenceUpstreamDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_callback_reference.NoArrayCallbackReferenceRule,
		[]rule_tester.ValidTestCase{
			{Code: "const unicorn = x => x + 1;\n\nexport default unicorn;", FileName: "file.js"},
			{Code: "const unicorn = (x, y) => x + (y ? y : 1);\n\nexport default unicorn;", FileName: "file.js"},
			{Code: "import unicorn from 'unicorn';\n\n[1, 2, 3].map(x => unicorn(x));\n//=> [2, 3, 4]", FileName: "file.js"},
			{Code: "const foo = array.map(element => callback(element));", FileName: "file.js"},
			{Code: "const foo = array.map(Boolean);", FileName: "file.js"},
			{Code: "array.forEach(element => {\n\tcallback(element);\n});", FileName: "file.js"},
			{Code: "const foo = array.every(element => callback(element));", FileName: "file.js"},
			{Code: "const foo = array.filter(element => callback(element));", FileName: "file.js"},
			{Code: "const foo = array.filter(Boolean);", FileName: "file.js"},
			{Code: "const foo = array.find(element => callback(element));", FileName: "file.js"},
			{Code: "const index = array.findIndex(element => callback(element));", FileName: "file.js"},
			{Code: "const foo = array.some(element => callback(element));", FileName: "file.js"},
			{Code: "const foo = array.reduce(\n\t(accumulator, element) => accumulator + callback(element),\n\t0\n);", FileName: "file.js"},
			{Code: "const foo = array.reduceRight(\n\t(accumulator, element) => [\n\t\t...accumulator,\n\t\tcallback(element)\n\t],\n\t[]\n);", FileName: "file.js"},
			{Code: "const foo = array.flatMap(element => callback(element));", FileName: "file.js"},
			// The documentation marks this factory call as incorrect, but the pinned implementation explicitly ignores calls.
			{Code: "array.forEach(someFunction({foo: 'bar'}));", FileName: "file.js"},
			// The documentation marks this factory call as incorrect, but the pinned implementation explicitly ignores calls.
			{Code: "const callback = someFunction({foo: 'bar'});\n\narray.forEach(element => {\n\tcallback(element);\n});", FileName: "file.js"},
			{Code: "array.forEach(function (element) {\n\tcallback(element, this);\n}, thisArgument);", FileName: "file.js"},
			{Code: "function readFile(filename) {\n\treturn fs.readFile(filename, 'utf8');\n}\n\nPromise.map(filenames, readFile);", FileName: "file.js"},
			{Code: "/* eslint unicorn/no-array-callback-reference: [\"error\", {\"ignore\": [\"Angular\"]}] */\nAngular.forEach(list, fn); // Passes", FileName: "file.js", Options: []any{map[string]any{"ignore": []any{"Angular"}}}},
		}, []rule_tester.InvalidTestCase{
			{Code: "import unicorn from 'unicorn';\n\n[1, 2, 3].map(unicorn);\n//=> [2, 3, 4]", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `unicorn` directly to `.map(…)`.", Line: 3, Column: 15, EndLine: 3, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "import unicorn from 'unicorn';\n\n[1, 2, 3].map((element) => unicorn(element));\n//=> [2, 3, 4]"},
					{MessageId: "replace-with-name", Output: "import unicorn from 'unicorn';\n\n[1, 2, 3].map((element, index) => unicorn(element, index));\n//=> [2, 3, 4]"},
					{MessageId: "replace-with-name", Output: "import unicorn from 'unicorn';\n\n[1, 2, 3].map((element, index, array) => unicorn(element, index, array));\n//=> [2, 3, 4]"},
				}},
			}},
			{Code: "import unicorn from 'unicorn';\n\n[1, 2, 3].map(unicorn);\n//=> [2, 3, 5]", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `unicorn` directly to `.map(…)`.", Line: 3, Column: 15, EndLine: 3, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "import unicorn from 'unicorn';\n\n[1, 2, 3].map((element) => unicorn(element));\n//=> [2, 3, 5]"},
					{MessageId: "replace-with-name", Output: "import unicorn from 'unicorn';\n\n[1, 2, 3].map((element, index) => unicorn(element, index));\n//=> [2, 3, 5]"},
					{MessageId: "replace-with-name", Output: "import unicorn from 'unicorn';\n\n[1, 2, 3].map((element, index, array) => unicorn(element, index, array));\n//=> [2, 3, 5]"},
				}},
			}},
			{Code: "const foo = array.map(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.map(…)`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const foo = array.map((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const foo = array.map((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const foo = array.map((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "array.forEach(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.forEach(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.forEach((element) => { callback(element); });"},
					{MessageId: "replace-with-name", Output: "array.forEach((element, index) => { callback(element, index); });"},
					{MessageId: "replace-with-name", Output: "array.forEach((element, index, array) => { callback(element, index, array); });"},
				}},
			}},
			{Code: "const foo = array.every(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.every(…)`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const foo = array.every((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const foo = array.every((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const foo = array.every((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "const foo = array.filter(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.filter(…)`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const foo = array.filter((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const foo = array.filter((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const foo = array.filter((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "const foo = array.find(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.find(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const foo = array.find((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const foo = array.find((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const foo = array.find((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "const index = array.findIndex(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.findIndex(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const index = array.findIndex((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const index = array.findIndex((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const index = array.findIndex((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "const foo = array.some(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.some(…)`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const foo = array.some((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const foo = array.some((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const foo = array.some((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "const foo = array.reduce(callback, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.reduce(…)`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const foo = array.reduce((accumulator, element) => callback(accumulator, element), 0);"},
					{MessageId: "replace-with-name", Output: "const foo = array.reduce((accumulator, element, index) => callback(accumulator, element, index), 0);"},
					{MessageId: "replace-with-name", Output: "const foo = array.reduce((accumulator, element, index, array) => callback(accumulator, element, index, array), 0);"},
				}},
			}},
			{Code: "const foo = array.reduceRight(callback, []);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.reduceRight(…)`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const foo = array.reduceRight((accumulator, element) => callback(accumulator, element), []);"},
					{MessageId: "replace-with-name", Output: "const foo = array.reduceRight((accumulator, element, index) => callback(accumulator, element, index), []);"},
					{MessageId: "replace-with-name", Output: "const foo = array.reduceRight((accumulator, element, index, array) => callback(accumulator, element, index, array), []);"},
				}},
			}},
			{Code: "const foo = array.flatMap(callback);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.flatMap(…)`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "const foo = array.flatMap((element) => callback(element));"},
					{MessageId: "replace-with-name", Output: "const foo = array.flatMap((element, index) => callback(element, index));"},
					{MessageId: "replace-with-name", Output: "const foo = array.flatMap((element, index, array) => callback(element, index, array));"},
				}},
			}},
			{Code: "array.forEach(callback, thisArgument);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error-with-name", Message: "Do not pass function `callback` directly to `.forEach(…)`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "replace-with-name", Output: "array.forEach((element) => { callback(element); }, thisArgument);"},
					{MessageId: "replace-with-name", Output: "array.forEach((element, index) => { callback(element, index); }, thisArgument);"},
					{MessageId: "replace-with-name", Output: "array.forEach((element, index, array) => { callback(element, index, array); }, thisArgument);"},
				}},
			}},
		})
}
