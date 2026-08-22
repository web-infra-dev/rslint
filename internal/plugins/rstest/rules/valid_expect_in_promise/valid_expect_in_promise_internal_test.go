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
		// A key behind parentheses is the one chain the identifier table cannot
		// answer, because the parser interns only a key that is a literal itself.
		{source: `promise[("then")](value => expect(value).toBe(1))`, want: true},
		{source: `promise[ /* c */ ("finally") ](() => expect(cleanup).toHaveBeenCalled())`, want: true},
		{source: `promise[("\x74hen")](value => expect(value).toBe(1))`, want: true},
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
		// A bracket paired with a member name the file only spells in text stays
		// out without a walk, because no bracket opens on a parenthesis.
		{source: `test("case", () => expect(map["other"]).toBe("then"))`, want: false},
		// A parenthesized bracket key that is not a chain costs one walk and
		// still stays out.
		{source: `test("case", () => expect(rows[(index + 1)]).toBe("then"))`, want: false},
		{source: `test("case", () => expect(rows[(index + 1)]).toBe(1))`, want: false},
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

func TestBracketOpensOnParenthesis(t *testing.T) {
	testCases := []struct {
		text string
		want bool
	}{
		{text: `promise[("then")](fn)`, want: true},
		{text: "promise[\n  (\"then\")\n](fn)", want: true},
		{text: `promise[/* key */ ("then")](fn)`, want: true},
		{text: "promise[// key\n(\"then\")](fn)", want: true},
		{text: `promise["then"](fn)`, want: false},
		{text: `rows[index](fn)`, want: false},
		{text: `rows[index] && other[(index)]`, want: true},
		{text: `no brackets at all`, want: false},
		{text: `trailing bracket [`, want: false},
		{text: `unterminated block comment [/*`, want: false},
		{text: `unterminated line comment [//`, want: false},
	}

	for _, testCase := range testCases {
		if got := bracketOpensOnParenthesis(testCase.text); got != testCase.want {
			t.Errorf(
				"bracketOpensOnParenthesis(%q) = %t, want %t",
				testCase.text,
				got,
				testCase.want,
			)
		}
	}
}
