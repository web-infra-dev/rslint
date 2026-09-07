package prefer_await_to_callbacks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/prefer_await_to_callbacks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Ported from eslint-plugin-promise v7.3.0:
// https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/__tests__/prefer-await-to-callbacks.js
// https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/docs/rules/prefer-await-to-callbacks.md
func TestPreferAwaitToCallbacksUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_await_to_callbacks.PreferAwaitToCallbacksRule,
		[]rule_tester.ValidTestCase{
			// Upstream tests.
			{Code: `async function hi() { await thing().catch(err => console.log(err)) }`, Tsx: true},
			{Code: `async function hi() { await thing().then() }`, Tsx: true},
			{Code: `async function hi() { await thing().catch() }`, Tsx: true},
			{Code: `dbConn.on("error", err => { console.error(err) })`, Tsx: true},
			{Code: `dbConn.once("error", err => { console.error(err) })`, Tsx: true},
			{Code: `heart(something => {})`, Tsx: true},
			{Code: `getErrors().map(error => responseTo(error))`, Tsx: true},
			{Code: `errors.filter(err => err.status === 402)`, Tsx: true},
			{Code: `errors.some(err => err.message.includes("Yo"))`, Tsx: true},
			{Code: `errors.every(err => err.status === 402)`, Tsx: true},
			{Code: `errors.filter(err => console.log(err))`, Tsx: true},
			{Code: `const error = errors.find(err => err.stack.includes("file.js"))`, Tsx: true},
			{Code: `this.myErrors.forEach(function(error) { log(error); })`, Tsx: true},
			{Code: `find(errors, function(err) { return  err.type === "CoolError" })`, Tsx: true},
			{Code: `map(errors, function(error) { return  err.type === "CoolError" })`, Tsx: true},
			{Code: `_.find(errors, function(error) { return  err.type === "CoolError" })`, Tsx: true},
			{Code: `_.map(errors, function(err) { return  err.type === "CoolError" })`, Tsx: true},
			// Upstream documentation (the yield example is wrapped in a generator).
			{Code: `await doSomething(arg)`, Tsx: true},
			{Code: `async function doSomethingElse() {}`, Tsx: true},
			{Code: `function* values() { yield yieldValue(err => {}) }`, Tsx: true},
			{Code: `eventEmitter.on('error', err => {})`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream tests.
			{Code: `heart(function(err) {})`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 7, EndLine: 1, EndColumn: 23},
				}},
			{Code: `heart(err => {})`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 7, EndLine: 1, EndColumn: 16},
				}},
			{Code: `heart("ball", function(err) {})`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 15, EndLine: 1, EndColumn: 31},
				}},
			{Code: `function getData(id, callback) {}`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 22, EndLine: 1, EndColumn: 30},
				}},
			{Code: `const getData = (cb) => {}`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				}},
			{Code: `var x = function (x, cb) {}`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 22, EndLine: 1, EndColumn: 24},
				}},
			{Code: `cb()`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				}},
			{Code: `callback()`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				}},
			{Code: `heart(error => {})`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 7, EndLine: 1, EndColumn: 18},
				}},
			{Code: `async.map(files, fs.stat, function(err, results) { if (err) throw err; });`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 27, EndLine: 1, EndColumn: 73},
				}},
			{Code: `_.map(files, fs.stat, function(err, results) { if (err) throw err; });`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 23, EndLine: 1, EndColumn: 69},
				}},
			{Code: `map(files, fs.stat, function(err, results) { if (err) throw err; });`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 21, EndLine: 1, EndColumn: 67},
				}},
			{Code: `map(function(err, results) { if (err) throw err; });`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 5, EndLine: 1, EndColumn: 51},
				}},
			{Code: `customMap(errors, (err) => err.message)`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 19, EndLine: 1, EndColumn: 39},
				}},
			// Upstream documentation (the yield example is wrapped in a generator).
			{Code: `cb()
callback()
doSomething(arg, (err) => {})
function doSomethingElse(cb) {}`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 3, Column: 18, EndLine: 3, EndColumn: 29},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 4, Column: 26, EndLine: 4, EndColumn: 28},
				}},
		},
	)
}
