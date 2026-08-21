package valid_expect_in_promise

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestSourceMayContainPromiseChain(t *testing.T) {
	testCases := []struct {
		source string
		want   bool
	}{
		{source: `test("case", () => expect(value).toBe(1))`, want: false},
		{source: `promise.then(value => expect(value).toBe(1))`, want: true},
		{source: `promise.catch(error => expect(error).toBeDefined())`, want: true},
		{source: `promise.finally(() => expect(cleanup).toHaveBeenCalled())`, want: true},
		{source: `promise["then"](value => expect(value).toBe(1))`, want: true},
		{source: "promise[`finally`](() => expect(cleanup).toHaveBeenCalled())", want: true},
		{source: `promise["\x74hen"](value => expect(value).toBe(1))`, want: true},
		{source: `const text = "then"`, want: false},
		{source: `const catchError = () => {}`, want: false},
		{source: `const text = "\x74hen"`, want: false},
		{source: `promise.then()`, want: false},
		{source: `promise.catch(onRejected, extra)`, want: false},
		// The `catch` of a try/catch is a keyword token, so it never reaches the
		// identifier table and must not open the walk.
		{source: `test("case", () => { try { run() } catch (error) { fail(error) } })`, want: false},
		// A backslash alone is not a bracket access; only an escaped bracket key
		// needs the walk.
		{source: `test("case", () => expect(text).toMatch(/\d+/))`, want: false},
		{source: `test("case", () => expect(text).toBe("a\tb"))`, want: false},
		// A bracket access keyed by an unrelated literal stays out.
		{source: `test("case", () => expect(map["other"]).toBe(1))`, want: false},
	}

	for _, testCase := range testCases {
		sourceFile := parser.ParseSourceFile(
			ast.SourceFileParseOptions{
				FileName: "/source.test.ts",
				Path:     "/source.test.ts",
			},
			testCase.source,
			core.ScriptKindTS,
		)
		if got := sourceMayContainPromiseChain(sourceFile); got != testCase.want {
			t.Errorf(
				"sourceMayContainPromiseChain(%q) = %t, want %t",
				testCase.source,
				got,
				testCase.want,
			)
		}
	}
}

func TestSourceMayContainPromiseChainKeepsUnknownSourceFile(t *testing.T) {
	if !sourceMayContainPromiseChain(nil) {
		t.Error("nil source file must conservatively keep the rule enabled")
	}
}
