// TestValidParamsUpstream migrates the full valid/invalid suite from upstream
// __tests__/valid-params.js 1:1. rslint-specific edge-shape and branch lock-in
// coverage lives in valid_params_extras_test.go.
package valid_params_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/valid_params"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidParamsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_params.ValidParamsRule,
		[]rule_tester.ValidTestCase{
			// ---- valid Promise.resolve() ----
			{Code: `Promise.resolve()`},
			{Code: `Promise.resolve(1)`},
			{Code: `Promise.resolve({})`},
			{Code: `Promise.resolve(referenceToSomething)`},

			// ---- valid Promise.reject() ----
			{Code: `Promise.reject()`},
			{Code: `Promise.reject(1)`},
			{Code: `Promise.reject({})`},
			{Code: `Promise.reject(referenceToSomething)`},
			{Code: `Promise.reject(Error())`},

			// ---- valid Promise.race() ----
			{Code: `Promise.race([])`},
			{Code: `Promise.race(iterable)`},
			{Code: `Promise.race([one, two, three])`},

			// ---- valid Promise.all() ----
			{Code: `Promise.all([])`},
			{Code: `Promise.all(iterable)`},
			{Code: `Promise.all([one, two, three])`},

			// ---- valid Promise.allSettled() ----
			{Code: `Promise.allSettled([])`},
			{Code: `Promise.allSettled(iterable)`},
			{Code: `Promise.allSettled([one, two, three])`},

			// ---- valid Promise.any() ----
			{Code: `Promise.any([])`},
			{Code: `Promise.any(iterable)`},
			{Code: `Promise.any([one, two, three])`},

			// ---- valid Promise.then() ----
			{Code: `somePromise().then(success)`},
			{Code: `somePromise().then(success, failure)`},
			{Code: `promiseReference.then(() => {})`},
			{Code: `promiseReference.then(() => {}, () => {})`},

			// ---- valid Promise.catch() ----
			{Code: `somePromise().catch(callback)`},
			{Code: `somePromise().catch(err => {})`},
			{Code: `promiseReference.catch(callback)`},
			{Code: `promiseReference.catch(err => {})`},

			// ---- valid Promise.finally() ----
			{Code: `somePromise().finally(callback)`},
			{Code: `somePromise().finally(() => {})`},
			{Code: `promiseReference.finally(callback)`},
			{Code: `promiseReference.finally(() => {})`},

			{
				Code: `
        somePromise.then(function() {
          return sth();
        }).catch(TypeError, function(e) {
          //
        }).catch(function(e) {
        });
      `,
				Options: []interface{}{map[string]interface{}{"exclude": []interface{}{"catch"}}},
			},

			// ---- integration test ----
			{Code: `
      Promise.all([
        Promise.resolve(1),
        Promise.resolve(2),
        Promise.reject(Error()),
      ])
        .then(console.log)
        .catch(console.error)
        .finally(console.log)
    `},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid Promise.resolve() ----
			{
				Code:   `Promise.resolve(1, 2)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneOptionalArgument", Message: "Promise.resolve() requires 0 or 1 arguments, but received 2", Line: 1, Column: 1}},
			},
			{
				Code:   `Promise.resolve({}, function() {}, 1, 2, 3)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneOptionalArgument", Message: "Promise.resolve() requires 0 or 1 arguments, but received 5", Line: 1, Column: 1}},
			},

			// ---- invalid Promise.reject() ----
			{
				Code:   `Promise.reject(1, 2, 3)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneOptionalArgument", Message: "Promise.reject() requires 0 or 1 arguments, but received 3", Line: 1, Column: 1}},
			},
			{
				Code:   `Promise.reject({}, function() {}, 1, 2)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneOptionalArgument", Message: "Promise.reject() requires 0 or 1 arguments, but received 4", Line: 1, Column: 1}},
			},

			// ---- invalid Promise.race() ----
			{
				Code:   `Promise.race(1, 2)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.race() requires 1 argument, but received 2", Line: 1, Column: 1}},
			},
			{
				Code:   `Promise.race({}, function() {}, 1, 2, 3)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.race() requires 1 argument, but received 5", Line: 1, Column: 1}},
			},

			// ---- invalid Promise.all() ----
			{
				Code:   `Promise.all(1, 2, 3)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.all() requires 1 argument, but received 3", Line: 1, Column: 1}},
			},
			{
				Code:   `Promise.all({}, function() {}, 1, 2)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.all() requires 1 argument, but received 4", Line: 1, Column: 1}},
			},

			// ---- invalid Promise.allSettled() ----
			{
				Code:   `Promise.allSettled(1, 2, 3)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.allSettled() requires 1 argument, but received 3", Line: 1, Column: 1}},
			},
			{
				Code:   `Promise.allSettled({}, function() {}, 1, 2)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.allSettled() requires 1 argument, but received 4", Line: 1, Column: 1}},
			},

			// ---- invalid Promise.any() ----
			{
				Code:   `Promise.any(1, 2, 3)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.any() requires 1 argument, but received 3", Line: 1, Column: 1}},
			},
			{
				Code:   `Promise.any({}, function() {}, 1, 2)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.any() requires 1 argument, but received 4", Line: 1, Column: 1}},
			},

			// ---- invalid Promise.then() ----
			{
				Code:   `somePromise().then()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 0", Line: 1, Column: 1}},
			},
			{
				Code:   `somePromise().then(() => {}, () => {}, () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 3", Line: 1, Column: 1}},
			},
			{
				Code:   `promiseReference.then()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 0", Line: 1, Column: 1}},
			},
			{
				Code:   `promiseReference.then(() => {}, () => {}, () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 3", Line: 1, Column: 1}},
			},

			// ---- invalid Promise.catch() ----
			{
				Code:   `somePromise().catch()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.catch() requires 1 argument, but received 0", Line: 1, Column: 1}},
			},
			{
				Code:   `somePromise().catch(() => {}, () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.catch() requires 1 argument, but received 2", Line: 1, Column: 1}},
			},
			{
				Code:   `promiseReference.catch()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.catch() requires 1 argument, but received 0", Line: 1, Column: 1}},
			},
			{
				Code:   `promiseReference.catch(() => {}, () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.catch() requires 1 argument, but received 2", Line: 1, Column: 1}},
			},

			// ---- invalid Promise.finally() ----
			{
				Code:   `somePromise().finally()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.finally() requires 1 argument, but received 0", Line: 1, Column: 1}},
			},
			{
				Code:   `somePromise().finally(() => {}, () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.finally() requires 1 argument, but received 2", Line: 1, Column: 1}},
			},
			{
				Code:   `promiseReference.finally()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.finally() requires 1 argument, but received 0", Line: 1, Column: 1}},
			},
			{
				Code:   `promiseReference.finally(() => {}, () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.finally() requires 1 argument, but received 2", Line: 1, Column: 1}},
			},
		},
	)
}
