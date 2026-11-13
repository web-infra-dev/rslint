package array_callback_return

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

// Existing local cases and additional regressions. Expectations are checked
// against ESLint v9.39.1; the complete upstream suite lives in the JS tests.
func TestArrayCallbackReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ArrayCallbackReturnRule,
		[]rule_tester.ValidTestCase{
			{Code: "foo.forEach(function(x) { return x; });"},
			{Code: "foo.forEach(function() { if (a === b) { return; } });"},
			{Code: "foo.forEach(function() { return; });"},
			{Code: "foo.forEach(x => x);"},
			{Code: "foo.map(function() { return true; });"},
			{Code: "foo.map(function() { return; });", Options: []any{map[string]any{"allowImplicit": true}}},
			{Code: "foo.filter(function bar() { return true; });"},
			{Code: "foo.filter(function(x) { if (x) { return true; } else { return false; } });"},
			{Code: "foo.find(function() { return true; });"},
			{Code: "foo.reduce(function(acc, val) { return acc + val; }, 0);"},
			{Code: "foo.some(function() { return true; });"},
			{Code: "foo.every(function() { return true; });"},
			{Code: "foo.sort(function(a, b) { return a - b; });"},
			{Code: "foo.toSorted(function(a, b) { return a - b; });"},
			{Code: "foo.flatMap(function(x) { return [x]; });"},
			{Code: "foo.map(x => x);"},
			{Code: "foo.map(x => { return x; });"},
			{Code: "foo.filter(x => x);"},
			{Code: "foo.every(x => x);"},
			{Code: "foo.some(x => x);"},
			{Code: "Array.from(x, function() { return true; });"},
			{Code: "Array.from(x, x => x);"},
			{Code: "foo.unknownMethod(function() {});"},
			{Code: "foo.unknownMethod(function() { return true; });"},
			{Code: "foo.map(function() { return function() { return true; }; });"},
			{Code: "foo.map(function() { return function() {}; });"},
			{Code: "foo.map(function() { return; });", Options: []any{map[string]any{"allowImplicit": true}}},
			{Code: "foo.map(function() { if (a === b) { return; } return; });", Options: []any{map[string]any{"allowImplicit": true}}},
			{Code: "foo.every(function() { return; });", Options: []any{map[string]any{"allowImplicit": true}}},
			{Code: "foo.filter(function() { return; });", Options: []any{map[string]any{"allowImplicit": true}}},
			{Code: "foo.every(function() { if (a) return; else return a; });", Options: []any{map[string]any{"allowImplicit": true}}},
			{Code: "foo.every(function() { if (a) return a; else return; });", Options: []any{map[string]any{"allowImplicit": true}}},
			{Code: "foo.map(() => { if (a) return 1; else return; });", Options: []any{map[string]any{"allowImplicit": true}}},
			{Code: "foo.forEach(function(x) { bar(x); });", Options: []any{map[string]any{"checkForEach": true}}},
			{Code: "foo.forEach(function(x) {});", Options: []any{map[string]any{"checkForEach": true}}},
			{Code: "foo.forEach(function() { return; });", Options: []any{map[string]any{"checkForEach": true}}},
			{Code: "foo.forEach(x => void bar(x));", Options: []any{map[string]any{"allowVoid": true, "checkForEach": true}}},
			{Code: "foo.forEach((x, i) => void (i % 2 && bar(x)));", Options: []any{map[string]any{"allowVoid": true, "checkForEach": true}}},
			{Code: "foo.map(function() { try { return true; } catch(e) { return false; } });"},
			{Code: "foo.map(async function() { return true; });"},
			{Code: "foo.map(async () => true);"},
			{Code: "foo.map(function*() { yield true; });"},
			{Code: "foo.map(function() { try { return true; } catch(e) {} });"},
			{Code: "foo.map(function* () {});"},
			{Code: "foo.map(function () { throw new Error(); });"},
			{Code: "foo.map(function () { while (true) {} });"},
			{Code: "foo.map(function () { try { return true; } catch (e) {} });"},
			{Code: "foo.map(() => function () {});"},
			{Code: "foo.map((function () {})());"},
			{Code: "foo.map(cb, function () {});"},
			{Code: "Array.from(function () {}, cb);"},
			{Code: "Array.from([], cb, function () {});"},
			{Code: "obj.Array.from([], function () {});"},
			{Code: "foo[\"m\" + \"ap\"](function () {});"},
			{Code: "foo[`m${\"ap\"}`](function () {});"},
			{Code: "foo.map((function () {}, cb));"},
			{Code: "foo.forEach(x => (void x));", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}}},
			{Code: "foo.forEach(x => { return (void x); });", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "Array.from(x, function() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: "foo.every(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.every() expects a return value from function.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "foo.filter(function foo() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.filter() expects a return value from function 'foo'.", Line: 1, Column: 12, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "foo.find(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.find() expects a return value from function.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "foo.map(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: "foo.reduce(function(acc) {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.reduce() expects a return value from function.", Line: 1, Column: 12, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: "foo.reduceRight(function(acc) {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.reduceRight() expects a return value from function.", Line: 1, Column: 17, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: "foo.some(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.some() expects a return value from function.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "foo.sort(function(a, b) {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.sort() expects a return value from function.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "foo.toSorted(function(a, b) {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.toSorted() expects a return value from function.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: "foo.flatMap(function(x) {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.flatMap() expects a return value from function.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "foo.findLast(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.findLast() expects a return value from function.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: "foo.findLastIndex(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.findLastIndex() expects a return value from function.", Line: 1, Column: 19, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: "foo.findIndex(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.findIndex() expects a return value from function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: "foo.map(() => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from arrow function.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "foo.filter(() => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.filter() expects a return value from arrow function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: "Array.from(x, () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: "foo.map(function() { if (a === b) { return true; } });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAtEnd", Message: "Array.prototype.map() expects a value to be returned at the end of function.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: "foo.map(() => { if (a === b) { return true; } });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAtEnd", Message: "Array.prototype.map() expects a value to be returned at the end of arrow function.", Line: 1, Column: 12, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "foo.filter(function bar() { if (a === b) { return true; } });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAtEnd", Message: "Array.prototype.filter() expects a value to be returned at the end of function 'bar'.", Line: 1, Column: 12, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "foo.filter(function() { if (a) { return true; } else if (b) { return false; } });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAtEnd", Message: "Array.prototype.filter() expects a value to be returned at the end of function.", Line: 1, Column: 12, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: "foo.map(function() { return; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: "foo.every(function() { return; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.every() expects a return value from function.", Line: 1, Column: 24, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: "foo.every(function() { if (a) return; else return a; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.every() expects a return value from function.", Line: 1, Column: 31, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code: "foo.every(function() { if (a) return a; else return; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.every() expects a return value from function.", Line: 1, Column: 46, EndLine: 1, EndColumn: 53},
				},
			},
			{
				Code: "foo.map(() => { if (a) return 1; else return; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.map() expects a return value from arrow function.", Line: 1, Column: 39, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code: "foo.every(function foo() { if (a) return; else return; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.every() expects a return value from function 'foo'.", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
					{
						MessageId: "expectedReturnValue", Message: "Array.prototype.every() expects a return value from function 'foo'.", Line: 1, Column: 48, EndLine: 1, EndColumn: 55},
				},
			},
			{
				Code: "foo.forEach(x => x);", Options: []any{map[string]any{"checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrapBraces", Output: "foo.forEach(x => {x});"}}},
				},
			},
			{
				Code: "foo.forEach(function(x) { return x; });", Options: []any{map[string]any{"checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from function.", Line: 1, Column: 27, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: "foo.forEach(function() { if (a === b) { return a; } });", Options: []any{map[string]any{"checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from function.", Line: 1, Column: 41, EndLine: 1, EndColumn: 50},
				},
			},
			{
				Code: "foo.forEach(() => bar());", Options: []any{map[string]any{"checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 16, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrapBraces", Output: "foo.forEach(() => {bar()});"}}},
				},
			},
			{
				Code: "foo.forEach(x => bar(x));", Options: []any{map[string]any{"allowVoid": true, "checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrapBraces", Output: "foo.forEach(x => {bar(x)});"}, {MessageId: "prependVoid", Output: "foo.forEach(x => void bar(x));"}}},
				},
			},
			{
				Code: "foo.map(x => x).filter(() => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.filter() expects a return value from arrow function.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: "foo.map(function() { return bar.filter(function() {}); });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.filter() expects a return value from function.", Line: 1, Column: 40, EndLine: 1, EndColumn: 48},
				},
			},
			{
				Code: "foo.map(function () { function* inner() { yield 1; } });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "foo.map(function () { if (a) return; else return 1; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 30, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code: "foo.map(function () { if (a) return; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 30, EndLine: 1, EndColumn: 37},
					{
						MessageId: "expectedAtEnd", Message: "Array.prototype.map() expects a value to be returned at the end of function.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "foo.map(function () { return; return; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
					{
						MessageId: "expectedReturnValue", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 31, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code: "foo.map(function () { process.exit(); });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "foo.map(function () { try { return x; } catch (e) {} });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAtEnd", Message: "Array.prototype.map() expects a value to be returned at the end of function.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "foo.map((function () {}));",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "foo.map((() => { return function () {}; })());",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: "NotArray.from([], function () {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from function.", Line: 1, Column: 19, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: "foo?.[\"map\"]?.((() => {}));",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from arrow function.", Line: 1, Column: 20, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: "foo.map(false && function () {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: "foo.map((function () {}) ? cb : other);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "foo.forEach(x => { if (x) return void x; return x; });", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 42, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prependVoid", Output: "foo.forEach(x => { if (x) return void x; return void x; });"}}},
				},
			},
			{
				Code: "foo.forEach(x => { if (x) return x; return y; });", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 27, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prependVoid", Output: "foo.forEach(x => { if (x) return void x; return y; });"}}},
					{
						MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 37, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prependVoid", Output: "foo.forEach(x => { if (x) return x; return void y; });"}}},
				},
			},
			{
				Code: "foo.forEach(x => x ? y : z);", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrapBraces", Output: "foo.forEach(x => {x ? y : z});"}, {MessageId: "prependVoid", Output: "foo.forEach(x => void (x ? y : z));"}}},
				},
			},
			{
				Code: "foo.forEach(x => x ** 2);", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrapBraces", Output: "foo.forEach(x => {x ** 2});"}, {MessageId: "prependVoid", Output: "foo.forEach(x => void (x ** 2));"}}},
				},
			},
			{
				Code: "foo.forEach(x => (x, y));", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrapBraces", Output: "foo.forEach(x => {(x, y)});"}, {MessageId: "prependVoid", Output: "foo.forEach(x => void (x, y));"}}},
				},
			},
			{
				Code: "foo.forEach(x => /* keep */ ((x + 1)));", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrapBraces", Output: "foo.forEach(x => /* keep */ {((x + 1))});"}, {MessageId: "prependVoid", Output: "foo.forEach(x => /* keep */ void ((x + 1)));"}}},
				},
			},
			{
				Code: "foo.forEach(x => { return/* keep */(x + 1); });", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 20, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prependVoid", Output: "foo.forEach(x => { return/* keep */void (x + 1); });"}}},
				},
			},
			{
				Code: "foo.forEach(x => { return(x + 1); });", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 20, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prependVoid", Output: "foo.forEach(x => { return void (x + 1); });"}}},
				},
			},
			{
				Code: "foo.forEach(x=>y?.());", Options: []any{map[string]any{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedNoReturnValue", Message: "Array.prototype.forEach() expects no useless return value from arrow function.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrapBraces", Output: "foo.forEach(x=>{y?.()});"}, {MessageId: "prependVoid", Output: "foo.forEach(x=>void y?.());"}}},
				},
			},
			{
				Code: "/* 😀 */ foo.map(\r\n function named () { return; });",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.map() expects a return value from function 'named'.", Line: 2, Column: 22, EndLine: 2, EndColumn: 29},
				},
			},
			{
				Code: "foo.map(function () { if (a) return; });", Options: []any{map[string]any{"allowImplicit": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedAtEnd", Message: "Array.prototype.map() expects a value to be returned at the end of function.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "foo.map(function () { return; });", Options: []any{map[string]any{"allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedReturnValue", Message: "Array.prototype.map() expects a return value from function.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: "Int8Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 26, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code: "Uint8Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: "Uint8ClampedArray[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 34, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: "Int16Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: "Uint16Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: "Int32Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: "Uint32Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: "Float32Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 29, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: "Float64Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 29, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: "BigInt64Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 30, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: "BigUint64Array[\"from\"]([], () => {});",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "expectedInside", Message: "Array.from() expects a return value from arrow function.", Line: 1, Column: 31, EndLine: 1, EndColumn: 33},
				},
			},
		},
	)
}
