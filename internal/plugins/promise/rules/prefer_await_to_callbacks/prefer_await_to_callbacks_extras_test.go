package prefer_await_to_callbacks_test

import (
	"path/filepath"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/prefer_await_to_callbacks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Additional AST and branch coverage, checked against eslint-plugin-promise v7.3.0.
func TestPreferAwaitToCallbacksExtras(t *testing.T) {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		filepath.Join(root.Dir, "tsconfig.callbacks.json"): `{"extends":"./tsconfig.json","compilerOptions":{"allowJs":true}}`,
	})
	rule_tester.RunRuleTester(root, "tsconfig.callbacks.json", t, &prefer_await_to_callbacks.PreferAwaitToCallbacksRule,
		[]rule_tester.ValidTestCase{
			// Only bare callback names in the final parameter position match.
			{Code: `function work(cb, done) {}
const expression = function(callback, done) {};
const arrow = (cb, done) => {};`, Tsx: true},
			{Code: `function work(cb = noop) {}
const expression = function(...callback) {};
const arrow = ({cb}) => {};
function destructure([callback]) {}`, Tsx: true},
			{Code: `function work(CB) {}
CB(); Callback(); object.cb(); object.callback(); new callback();`, Tsx: true},
			// Only the final inline callback and its first bare error parameter match.
			{Code: `heart(); heart(handler); heart(...handlers); heart(err => {}, value);`, Tsx: true},
			{Code: `heart(() => {}); heart((value, err) => {}); heart(({err}) => {}); heart(([error]) => {}); heart((err = null) => {}); heart((...error) => {});`, Tsx: true},
			{Code: "new Task(err => {}); tag`${err => {}}`;", Tsx: true},
			// Method exemptions use property names and exact argument counts.
			{Code: `lodash.map(errors, err => {}); underscore.find(errors, error => {}); _.forEach(errors, err => {});`, Tsx: true},
			{Code: `db.on(err => {}); db.once(1, 2, function(error) {}); errors.map(err => {}, receiver);`, Tsx: true},
			{Code: `db[on](err => {}); db[(once)](error => {}); errors[map](err => {}); lodash[(filter)](errors, error => {});`, Tsx: true},
			{Code: `class Events { #on() {} #once() {} #map() {} listen() { this.#on(err => {}); this.#once(error => {}); this.#map(err => {}); } }`, Tsx: true},
			// Parentheses, optional chains and authored TypeScript wrappers.
			{Code: `(errors.map)((err => {})); (lodash).map(errors, err => {}); errors?.map?.(err => {}); db?.[on]?.(err => {});`, Tsx: true},
			{Code: `cb!(); (cb as Function)(); (callback satisfies Function)(); heart((err => {}) as Function); heart((err => {})!); heart((err => {}) satisfies Function);`, Tsx: true},
			// Await and yield exemptions inspect every ancestor, including nested functions.
			{Code: `async function work() { await wrap(() => { heart(err => {}); }); }
function* values() { yield* wrap(heart(error => {})); }`, Tsx: true},
			// ESTree method functions and TypeScript declaration/parameter boundaries.
			{Code: `declare function work(callback: Function): void;
abstract class Task { abstract method(cb: Function): void; }
interface Worker { method(callback: Function): void; (cb: Function): void; }
type Handler = (callback: Function) => void;`, Tsx: true},
			{Code: `class Task { constructor(private cb: Function) {} }
function work(cb: Function): void;
function work(done: Function) {}`, Tsx: true},
			{Code: `heart(function(this: Receiver, err: Error) {});`, Tsx: true},
			// JSX expressions and JavaScript JSDoc casts retain runtime meaning.
			{Code: `const view = <cb callback={handler} />;`, Tsx: true},
			{Code: `/** @type {Function} */ (errors.map)(err => {});
/** @this {Receiver} */
function work() {}`, FileName: "casts.js"},
			// Import expressions, synthetic JSDoc parameters and source trivia.
			{Code: `import(err => {}); import("module", error => {});`, Tsx: true},
			{Code: `class Task { constructor(public readonly callback: Function) {} }`, Tsx: true},
			{Code: `/** @this {Receiver} */
function work() {}
heart(/** @this {Receiver} */ function() {});`, FileName: "jsdoc-parameters.js"},
			{Code: `(callback<string>)(); heart((err => {}) as ((err: Error) => void));`, Tsx: true},
			// RunRuleTester registers the configured rule as "test".
			{Code: `/* eslint-disable test */
cb(); heart(err => {}); function work(callback) {}
/* eslint-enable test */
// eslint-disable-next-line test
callback();`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Only bare callback names in the final parameter position match.
			{Code: `const expression = function named(callback) {};
const arrow = callback => {};
async function work(cb) {}
function* generator(callback) {}`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 35, EndLine: 1, EndColumn: 43},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 15, EndLine: 2, EndColumn: 23},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 3, Column: 21, EndLine: 3, EndColumn: 23},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 4, Column: 21, EndLine: 4, EndColumn: 29},
				}},
			{Code: `function work(callback, cb) {}`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 25, EndLine: 1, EndColumn: 27},
				}},
			{Code: `function outer(cb) { cb(); }`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 16, EndLine: 1, EndColumn: 18},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				}},
			// Only the final inline callback and its first bare error parameter match.
			{Code: `heart(async function named(error) {}); heart(function* named(err) {}); heart(async (err) => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 7, EndLine: 1, EndColumn: 37},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 46, EndLine: 1, EndColumn: 69},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 78, EndLine: 1, EndColumn: 95},
				}},
			{Code: `heart((err, callback) => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 7, EndLine: 1, EndColumn: 28},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21},
				}},
			{Code: `cb(err => {}); callback(function(error) {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 16, EndLine: 1, EndColumn: 44},
				}},
			// Method exemptions use property names and exact argument counts.
			{Code: `errors.map(errors, err => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 20, EndLine: 1, EndColumn: 29},
				}},
			{Code: `errors.reduce(err => {}); on(err => {}); custom.map(1, 2, err => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 30, EndLine: 1, EndColumn: 39},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 59, EndLine: 1, EndColumn: 68},
				}},
			{Code: `db["on"](err => {}); errors["map"](err => {}); db[method](error => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 36, EndLine: 1, EndColumn: 45},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 59, EndLine: 1, EndColumn: 70},
				}},
			{Code: `class Events { #listen() {} listen() { this.#listen(err => {}); } }`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 53, EndLine: 1, EndColumn: 62},
				}},
			{Code: `db.on("ready", cb => {}); errors.map(callback => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 16, EndLine: 1, EndColumn: 18},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 38, EndLine: 1, EndColumn: 46},
				}},
			// Parentheses, optional chains and authored TypeScript wrappers.
			{Code: `((cb))(); callback?.(); cb<string>();`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 11, EndLine: 1, EndColumn: 23},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 25, EndLine: 1, EndColumn: 37},
				}},
			{Code: `((heart))(((err) => {})); heart?.(function(error) {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 12, EndLine: 1, EndColumn: 23},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 35, EndLine: 1, EndColumn: 53},
				}},
			{Code: `(errors?.map)(err => {}); (db?.on)(err => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 36, EndLine: 1, EndColumn: 45},
				}},
			{Code: `(errors.map)!(err => {}); (lodash as any).map(errors, err => {}); db[(on as string)](error => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 55, EndLine: 1, EndColumn: 64},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 86, EndLine: 1, EndColumn: 97},
				}},
			// Await and yield exemptions inspect every ancestor, including nested functions.
			{Code: `async function work() { await cb(); await heart((err, callback) => {}); }
function* values() { yield callback(); }`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 55, EndLine: 1, EndColumn: 63},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 28, EndLine: 2, EndColumn: 38},
				}},
			{Code: `heart(async err => { await load(); });`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 7, EndLine: 1, EndColumn: 37},
				}},
			// ESTree method functions and TypeScript declaration/parameter boundaries.
			{Code: `const object = { method(cb) {}, set value(callback) {}, get value() { return 1; } };
class Task { constructor(cb) {} method(callback) {} set value(cb) {} get value() { return 1; } }`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 25, EndLine: 1, EndColumn: 27},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 43, EndLine: 1, EndColumn: 51},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 26, EndLine: 2, EndColumn: 28},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 40, EndLine: 2, EndColumn: 48},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 63, EndLine: 2, EndColumn: 65},
				}},
			{Code: `class Task { #method(cb) {} static async method(callback) {} *generator(cb) {} }
const object = { [key](callback) {} };`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 22, EndLine: 1, EndColumn: 24},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 49, EndLine: 1, EndColumn: 57},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 73, EndLine: 1, EndColumn: 75},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 24, EndLine: 2, EndColumn: 32},
				}},
			{Code: `function work(callback?: (err: Error) => void) {}
const arrow = (cb: Function) => {};
heart((err: Error) => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 15, EndLine: 1, EndColumn: 46},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 16, EndLine: 2, EndColumn: 28},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 3, Column: 7, EndLine: 3, EndColumn: 25},
				}},
			{Code: `class Task { method(@decorate cb: Function) {} }`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 31, EndLine: 1, EndColumn: 43},
				}},
			// JSX expressions and JavaScript JSDoc casts retain runtime meaning.
			{Code: `const view = <Box value={cb()} onReady={(callback) => {}}>{heart(err => <Item />)}</Box>;`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 42, EndLine: 1, EndColumn: 50},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 66, EndLine: 1, EndColumn: 81},
				}},
			{Code: `/** @type {Function} */ (cb)();
heart(/** @type {Function} */ (err => {}));
/** @param {Function} callback */
function work(callback) {}`, FileName: "casts.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 32, EndLine: 2, EndColumn: 41},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 4, Column: 15, EndLine: 4, EndColumn: 23},
				}},
			// Complete diagnostic ranges include multiline callbacks and UTF-16 columns.
			{Code: `const emoji = "😀"; cb(
  value
);
heart(
  async (error) => {
    log(error);
  }
);
function work(
  value,
  callback
) {}`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 21, EndLine: 3, EndColumn: 2},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 5, Column: 3, EndLine: 7, EndColumn: 4},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 11, Column: 3, EndLine: 11, EndColumn: 11},
				}},
			{Code: `const café = "😀"; function work(c\u0062) {}
call((\u0065rr) => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 34, EndLine: 1, EndColumn: 41},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 6, EndLine: 2, EndColumn: 22},
				}},
			// Import expressions, synthetic JSDoc parameters and source trivia.
			{Code: `import.defer("module", err => {}); import.source("module", error => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 24, EndLine: 1, EndColumn: 33},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 60, EndLine: 1, EndColumn: 71},
				}},
			{Code: `class Task extends Base { constructor() { super(err => {}); } }`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 49, EndLine: 1, EndColumn: 58},
				}},
			{Code: `const obj = { [((cb))()]() {} };`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24},
				}},
			{Code: `class Task { constructor(@decorate callback: Function) {} }`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 36, EndLine: 1, EndColumn: 54},
				}},
			{Code: `function work(callback?) {}
function other(cb /* comment */) {}
cb /* comment */ ();`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 16, EndLine: 2, EndColumn: 18},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 3, Column: 1, EndLine: 3, EndColumn: 20},
				}},
			{Code: `heart(/** @this {Receiver} */ function(err) {});
/** @this {Receiver} */
function work(callback) {}`, FileName: "jsdoc-parameters.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 1, Column: 31, EndLine: 1, EndColumn: 47},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 3, Column: 15, EndLine: 3, EndColumn: 23},
				}},
			{Code: `/** @param {Function} [callback=noop] */
function work(callback) {}
heart(/** @param {Error} [err=null] */ function(err) {});`, FileName: "jsdoc-parameters.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 2, Column: 15, EndLine: 2, EndColumn: 23},
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 3, Column: 40, EndLine: 3, EndColumn: 56},
				}},
			// RunRuleTester registers the configured rule as "test".
			{Code: `// eslint-disable-next-line test
cb();
callback(); // eslint-disable-line test
heart(err => {});`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error", Message: "Avoid callbacks. Prefer Async/Await.", Line: 4, Column: 7, EndLine: 4, EndColumn: 16},
				}},
		},
	)
}
