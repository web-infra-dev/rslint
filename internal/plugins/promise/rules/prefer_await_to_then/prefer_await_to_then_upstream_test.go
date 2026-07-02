package prefer_await_to_then_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/prefer_await_to_then"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const msg = "Prefer await to then()/catch()/finally()."

func TestPreferAwaitToThenUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_await_to_then.PreferAwaitToThenRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: `async function hi() { await thing() }`},
			{Code: `async function hi() { await thing().then() }`},
			{Code: `async function hi() { await thing().catch() }`},
			{Code: `async function hi() { await thing().finally() }`},
			// Cypress
			{Code: `function hi() { cy.get(".myClass").then(go) }`},
			{Code: `function hi() { cy.get("button").click().then() }`},
			{Code: `function * hi() { yield thing().then() }`},
			{Code: `a = async () => (await something())`},
			{Code: "a = async () => {\n  try { await something() } catch (error) { somethingElse() }\n}"},
			{Code: `something().then(async () => await somethingElse())`},
			{Code: `function foo() { hey.somethingElse(x => {}) }`},
			{Code: "const isThenable = (obj) => {\n  return obj && typeof obj.then === 'function';\n};"},
			{Code: "function isThenable(obj) {\n  return obj && typeof obj.then === 'function';\n}"},
			{Code: "class Foo {\n  constructor () {\n    doSomething.then(abc);\n  }\n}"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `function foo() { hey.then(x => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			// Errors reported outer-to-inner (call chain traversal order)
			{
				Code: `function foo() { hey.then(function() { }).then() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 43, EndLine: 1, EndColumn: 47},
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `function foo() { hey.then(function() { }).then(x).catch() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 51, EndLine: 1, EndColumn: 56},
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 43, EndLine: 1, EndColumn: 47},
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `async function a() { hey.then(function() { }).then(function() { }) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 47, EndLine: 1, EndColumn: 51},
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 26, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: `function foo() { hey.catch(x => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 22, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: `function foo() { hey.finally(x => {}) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// strict mode: await-wrapped .then() is still reported
			{
				Code:    `async function hi() { await thing().then() }`,
				Options: map[string]interface{}{"strict": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 37, EndLine: 1, EndColumn: 41},
				},
			},
			// strict mode: constructor .then() is still reported
			{
				Code:    "class Foo {\n  constructor () {\n    doSomething.then(abc);\n  }\n}",
				Options: map[string]interface{}{"strict": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 3, Column: 17, EndLine: 3, EndColumn: 21},
				},
			},
			// strict mode: await-wrapped .catch() is still reported
			{
				Code:    `async function hi() { await thing().catch() }`,
				Options: map[string]interface{}{"strict": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 37, EndLine: 1, EndColumn: 42},
				},
			},
			// strict mode: await-wrapped .finally() is still reported
			{
				Code:    `async function hi() { await thing().finally() }`,
				Options: map[string]interface{}{"strict": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
				},
			},
			// strict mode: yield-wrapped .then() is still reported
			{
				Code:    `function * hi() { yield thing().then() }`,
				Options: map[string]interface{}{"strict": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferAwaitToCallback", Message: msg, Line: 1, Column: 33, EndLine: 1, EndColumn: 37},
				},
			},
		},
	)
}
