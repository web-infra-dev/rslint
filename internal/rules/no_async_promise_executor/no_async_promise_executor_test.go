package no_async_promise_executor

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSourceMayUsePromise(t *testing.T) {
	if !sourceMayUsePromise(nil) || !sourceMayUsePromise(&ast.SourceFile{}) {
		t.Fatal("missing parser metadata must conservatively keep listeners")
	}

	for _, testCase := range []struct {
		name string
		code string
		want bool
	}{
		{name: "unrelated code", code: `new Worker(() => {});`, want: false},
		{name: "promise without async text", code: `new Promise(resolve => resolve());`, want: false},
		{name: "async without promise identifier", code: `const work = async () => {};`, want: false},
		{name: "string only", code: `const marker = "Promise async";`, want: false},
		{name: "async executor", code: `new Promise(async () => {});`, want: true},
		{name: "escaped promise identifier", code: `new Prom\u0069se(async () => {});`, want: true},
		{name: "conservative property match", code: `service.Promise(async () => {});`, want: true},
		{name: "conservative async string match", code: `const marker = "async"; new Promise(resolve => resolve());`, want: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/source.ts",
				Path:     "/source.ts",
			}, testCase.code, core.ScriptKindTS)
			if got := sourceMayUsePromise(sourceFile); got != testCase.want {
				t.Fatalf("sourceMayUsePromise(%q) = %v, want %v", testCase.code, got, testCase.want)
			}
			listeners := NoAsyncPromiseExecutorRule.Run(rule.RuleContext{SourceFile: sourceFile}, nil)
			if got := len(listeners) != 0; got != testCase.want {
				t.Fatalf("listener presence for %q = %v, want %v", testCase.code, got, testCase.want)
			}
		})
	}
}

func TestNoAsyncPromiseExecutorRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAsyncPromiseExecutorRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Regular Promise executors (not async)
			{Code: `new Promise((resolve, reject) => {})`},
			{Code: `new Promise((resolve, reject) => {}, async function unrelated() {})`},
			{Code: `new Foo(async (resolve, reject) => {})`},
			{Code: `Promise(async (resolve, reject) => {})`},
			{Code: `new globalThis.Promise(async (resolve, reject) => {})`},
			{Code: `new Promise((async () => {}) as () => void)`},
			{Code: `const async = () => {}; new Promise(async)`},
			{Code: `new Promise(...[async () => {}])`},
			// Only global Promise references are checked.
			{Code: `/* global Promise:off */ new Promise(async (resolve, reject) => {})`},
			{Code: `new Promise(async (resolve, reject) => {})`, Globals: map[string]any{"Promise": "off"}},
			{Code: `let Promise; new Promise(async (resolve, reject) => {})`},
			{Code: `function f() { new Promise(async (resolve, reject) => {}); var Promise; }`},
			{Code: `function f(Promise) { new Promise(async (resolve, reject) => {}); }`},
			{Code: `function f({ Promise }) { new Promise(async () => {}); }`},
			{Code: `try {} catch (Promise) { new Promise(async () => {}); }`},
			{Code: `{ new Promise(async () => {}); let Promise; }`},
			{Code: `let Promise; new (Promise)(async () => {});`},
			{Code: `for (let Promise of values) new Promise(async () => {});`},
			{Code: `switch (value) { case 0: new Promise(async () => {}); break; case 1: let Promise; }`},
			{Code: `function f(Promise = globalThis.Promise) { new Promise(async () => {}); }`},
			{Code: `const f = function Promise() { new Promise(async () => {}); };`},
			{Code: `const C = class Promise { static { new Promise(async () => {}); } };`},
			{Code: `import Promise from "promise"; new Promise(async () => {});`},
			{Code: `import { Promise } from "promise"; new Promise(async () => {});`},
			{Code: `class Promise {} new Promise(async () => {});`},
			{Code: `function Promise() {} new Promise(async () => {});`},
			{Code: `namespace Promise {} new Promise(async () => {});`},
			{Code: `interface Promise {} new Promise(async () => {});`},
			{Code: `type Promise = {}; new Promise(async () => {});`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Async named function
			{
				Code: `new Promise(async function foo(resolve, reject) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    13,
					},
				},
			},

			// Async arrow function
			{
				Code: `new Promise(async (resolve, reject) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    13,
					},
				},
			},

			// Wrapped async arrow function
			{
				Code: `new Promise(((((async () => {})))))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    17,
					},
				},
			},
			{
				Code: `new (((Promise)))((((async function* () {}))))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    22,
					},
				},
			},
			{
				Code: `new Promise<unknown>(/* executor */ async () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "async", Line: 1, Column: 37, EndLine: 1, EndColumn: 42},
				},
			},

			// Parser identifier metadata normalizes escaped identifier text.
			{
				Code: `new Prom\u0069se(async () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    18,
					},
				},
			},
			{
				Code: `export {}; interface Promise {} new Promise(async () => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "async"},
				},
			},
			{
				Code: `{ type Promise = {}; new Promise(async () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "async"},
				},
			},
			{
				Code: `function f<Promise>() { new Promise(async () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "async"},
				},
			},
			{
				Code: `
function f(value = new Promise(async () => {})) {
  new Promise(async () => {});
  var Promise;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "async", Line: 2, Column: 32},
				},
			},

			// A shadowed nested scope must not poison the cached outer scope.
			{
				Code: `
new Promise(async () => {});
{
  let Promise;
  new Promise(async () => {});
}
new Promise(async () => {});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "async", Line: 2, Column: 13},
					{MessageId: "async", Line: 7, Column: 13},
				},
			},
		},
	)
}
