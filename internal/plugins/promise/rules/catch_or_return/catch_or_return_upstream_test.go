// TestCatchOrReturnUpstream migrates the full valid/invalid suite from upstream
// __tests__/catch-or-return.js 1:1. Position assertions cover line for every
// invalid case. rslint-specific lock-in cases live in catch_or_return_extras_test.go.
package catch_or_return_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/catch_or_return"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const catchMessage = "Expected catch() or return"
const doneMessage = "Expected done() or return"

func TestCatchOrReturnUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&catch_or_return.CatchOrReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- catch ----
			{Code: `frank().then(go).catch(doIt)`},
			{Code: `frank().then(go).then().then().then().catch(doIt)`},
			{Code: `frank().then(go).then().catch(function() { /* why bother */ })`},
			{Code: `frank.then(go).then(to).catch(jail)`},
			{Code: `Promise.resolve(frank).catch(jail)`},
			{Code: `Promise.resolve(frank)["catch"](jail)`},
			{Code: `frank.then(to).finally(fn).catch(jail)`},

			// ---- Cypress ----
			{Code: `cy.get(".myClass").then(go)`},
			{Code: `cy.get("button").click().then()`},

			// ---- arrow function use case ----
			{Code: "postJSON(\"/smajobber/api/reportJob.json\")\n\t.then(()=>this.setState())\n\t.catch(()=>this.setState())"},

			// ---- return ----
			{Code: `function a() { return frank().then(go) }`},
			{Code: `function a() { return frank().then(go).then().then().then() }`},
			{Code: `function a() { return frank().then(go).then()}`},
			{Code: `function a() { return frank.then(go).then(to) }`},

			// ---- allowThen - .then(null, fn) ----
			{Code: `frank().then(go).then(null, doIt)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(go).then().then().then().then(null, doIt)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(go).then().then(null, function() { /* why bother */ })`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank.then(go).then(to).then(null, jail)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(a).then(b).then(null, c)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(a).then(b).then().then().then(null, doIt)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(a).then(b).then(null, function() { /* why bother */ })`, Options: map[string]interface{}{"allowThen": true}},

			// ---- allowThen - .then(fn, fn) ----
			{Code: `frank().then(a, b)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(go).then(zam, doIt)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(a).then(b).then(c, d)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(go).then().then().then().then(wham, doIt)`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank().then(go).then().then(function() {}, function() { /* why bother */ })`, Options: map[string]interface{}{"allowThen": true}},
			{Code: `frank.then(go).then(to).then(pewPew, jail)`, Options: map[string]interface{}{"allowThen": true}},

			// ---- allowThenStrict - .then(null, fn) ----
			{Code: `frank().then(go).then(null, doIt)`, Options: map[string]interface{}{"allowThenStrict": true}},
			{Code: `frank().then(go).then().then().then().then(null, doIt)`, Options: map[string]interface{}{"allowThenStrict": true}},
			{Code: `frank().then(go).then().then(null, function() { /* why bother */ })`, Options: map[string]interface{}{"allowThenStrict": true}},
			{Code: `frank.then(go).then(to).then(null, jail)`, Options: map[string]interface{}{"allowThenStrict": true}},
			{Code: `frank().then(a).then(b).then(null, c)`, Options: map[string]interface{}{"allowThenStrict": true}},
			{Code: `frank().then(a).then(b).then().then().then(null, doIt)`, Options: map[string]interface{}{"allowThenStrict": true}},
			{Code: `frank().then(a).then(b).then(null, function() { /* why bother */ })`, Options: map[string]interface{}{"allowThenStrict": true}},

			// ---- allowFinally ----
			{Code: `frank().then(go).catch(doIt).finally(fn)`, Options: map[string]interface{}{"allowFinally": true}},
			{Code: `frank().then(go).then().then().then().catch(doIt).finally(fn)`, Options: map[string]interface{}{"allowFinally": true}},
			{Code: `frank().then(go).then().catch(function() { /* why bother */ }).finally(fn)`, Options: map[string]interface{}{"allowFinally": true}},

			// ---- terminationMethod=done ----
			{Code: `frank().then(go).done()`, Options: map[string]interface{}{"terminationMethod": "done"}},

			// ---- terminationMethod=[catch, done] ----
			{Code: `frank().then(go).catch()`, Options: map[string]interface{}{"terminationMethod": []interface{}{"catch", "done"}}},
			{Code: `frank().then(go).done()`, Options: map[string]interface{}{"terminationMethod": []interface{}{"catch", "done"}}},
			{Code: `frank().then(go).finally()`, Options: map[string]interface{}{"terminationMethod": []interface{}{"catch", "finally"}}},

			// ---- for coverage: non-promise expression statement ----
			{Code: `nonPromiseExpressionStatement();`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- catch failures ----
			{Code: `function callPromise(promise, cb) { promise.then(cb) }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `fetch("http://www.yahoo.com").then(console.log.bind(console))`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `a.then(function() { return "x"; }).then(function(y) { throw y; })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `Promise.resolve(frank)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `Promise.all([])`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `Promise.allSettled([])`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `Promise.any([])`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `Promise.race([])`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(to).catch(fn).then(foo)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().finally(fn)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(to).finally(fn)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).catch(doIt).finally(fn)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).then().then().then().catch(doIt).finally(fn)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).then().catch(function() { /* why bother */ }).finally(fn)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},

			// ---- return failures ----
			{Code: `function a() { frank().then(go) }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `function a() { frank().then(go).then().then().then() }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `function a() { frank().then(go).then()}`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `function a() { frank.then(go).then(to) }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},

			// ---- allowFinally=true failures ----
			{Code: `frank().then(go).catch(doIt).finally(fn).then(foo)`, Options: map[string]interface{}{"allowFinally": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).catch(doIt).finally(fn).foobar(foo)`, Options: map[string]interface{}{"allowFinally": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},

			// ---- terminationMethod=done failures ----
			{Code: `frank().then(go)`, Options: map[string]interface{}{"terminationMethod": "done"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: doneMessage, Line: 1}}},
			{Code: `frank().catch(go)`, Options: map[string]interface{}{"terminationMethod": "done"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: doneMessage, Line: 1}}},

			// ---- assume somePromise.ANYTHING() is a new promise ----
			{Code: `frank().catch(go).someOtherMethod()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},

			// ---- .then(null, fn) without allowThen ----
			{Code: `frank().then(a).then(b).then(null, c)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(a).then(b).then().then().then(null, doIt)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(a).then(b).then(null, function() { /* why bother */ })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},

			// ---- .then(fn, fn) without allowThen ----
			{Code: `frank().then(a, b)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).then(zam, doIt)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(a).then(b).then(c, d)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).then().then().then().then(wham, doIt)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).then().then(function() {}, function() { /* why bother */ })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank.then(go).then(to).then(pewPew, jail)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},

			// ---- .then(fn, fn) with allowThenStrict (non-null first arg still invalid) ----
			{Code: `frank().then(a, b)`, Options: map[string]interface{}{"allowThenStrict": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).then(zam, doIt)`, Options: map[string]interface{}{"allowThenStrict": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(a).then(b).then(c, d)`, Options: map[string]interface{}{"allowThenStrict": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).then().then().then().then(wham, doIt)`, Options: map[string]interface{}{"allowThenStrict": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank().then(go).then().then(function() {}, function() { /* why bother */ })`, Options: map[string]interface{}{"allowThenStrict": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
			{Code: `frank.then(go).then(to).then(pewPew, jail)`, Options: map[string]interface{}{"allowThenStrict": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "terminationMethod", Message: catchMessage, Line: 1}}},
		},
	)
}
