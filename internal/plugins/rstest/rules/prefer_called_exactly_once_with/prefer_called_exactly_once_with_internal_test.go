package prefer_called_exactly_once_with

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestSourceMayContainMergePair(t *testing.T) {
	escapedCalledOnce := `\x63` + "calledOnce"[1:]
	escapedCalledWith := `\x63` + "calledWith"[1:]

	testCases := []struct {
		source string
		want   bool
	}{
		{source: `test("case", () => expect(spy).toHaveBeenCalledOnce())`, want: false},
		{source: `test("case", () => expect(spy).toHaveBeenCalledWith("a"))`, want: false},
		{source: `test("case", () => { expect(spy).toHaveBeenCalledOnce(); expect(spy).toHaveBeenCalledWith("a") })`, want: true},
		{source: `test("case", () => { expect(spy).calledOnce; expect(spy).calledWith("a") })`, want: true},
		{source: `test("case", () => { expect(spy).toHaveBeenCalledOnce(); expect(spy).calledWith("a") })`, want: true},
		{source: `test("case", () => expect(spy).calledOnce.toHaveBeenCalledWith("a"))`, want: true},
		{source: `test("case", () => { expect(spy)["calledOnce"]; expect(spy)["calledWith"]("a") })`, want: true},
		{source: "test(\"case\", () => { expect(spy)[`calledOnce`]; expect(spy)[`calledWith`](\"a\") })", want: true},
		{source: `test("case", () => { expect(spy)["` + escapedCalledOnce + `"]; expect(spy)["` + escapedCalledWith + `"]("a") })`, want: true},
		{source: `test("case", () => { expect(spy)[("calledOnce")]; expect(spy).calledWith("a") })`, want: true},
		{source: `test("case", () => { expect(spy)[ /* key */ ("calledOnce") ]; expect(spy)[("calledWith")]("a") })`, want: true},
		{source: "test(\"case\", () => { expect(spy)[\uFEFF(\"calledOnce\")]; expect(spy).calledWith(\"a\") })", want: true},
		{source: "test(\"case\", () => { expect(spy)[// key\u2028(\"calledOnce\")]; expect(spy).calledWith(\"a\") })", want: true},
		{source: `test("case", () => { expect(spy)[("` + escapedCalledOnce + `")]; expect(spy)[("` + escapedCalledWith + `")]("a") })`, want: true},
		{source: `test("case", () => expect(spy).toHaveBeenCalledExactlyOnceWith("a"))`, want: false},
		{source: `test("case", () => expect(spy).calledOnceWith("a"))`, want: false},
		{source: `test("case", () => { expect(spy).calledOnceAgain; expect(spy).calledWithout("a") })`, want: false},
		{source: `test("case", () => expect(rows[(index + 1)]).toBe(1))`, want: false},
		{source: `test("case", () => { expect(spy)[("other")]; expect(spy).calledWith("a") })`, want: false},
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
		if got := sourceMayContainMergePair(sourceFile); got != testCase.want {
			t.Errorf(
				"sourceMayContainMergePair(%q) = %t, want %t",
				testCase.source,
				got,
				testCase.want,
			)
		}
	}
}

func TestSourceMayContainMergePairKeepsUnknownSourceFile(t *testing.T) {
	if !sourceMayContainMergePair(nil) {
		t.Error("nil source file must conservatively keep the rule enabled")
	}

	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/source.test.ts",
			Path:     "/source.test.ts",
		},
		`expect(spy).toHaveBeenCalledOnce(); expect(spy).toHaveBeenCalledWith(value);`,
		core.ScriptKindTS,
	)
	if !sourceMayContainMergePair(sourceFile) {
		t.Error("the compiler must collect identifier names before checking an uncached source file")
	}
}

func TestBracketOpensOnParenthesis(t *testing.T) {
	testCases := []struct {
		text string
		want bool
	}{
		{text: `expect(spy)[("calledOnce")]`, want: true},
		{text: "expect(spy)[\n  (\"calledOnce\")\n]", want: true},
		{text: `expect(spy)[/* key */ ("calledOnce")]`, want: true},
		{text: "expect(spy)[// key\n(\"calledOnce\")]", want: true},
		{text: "expect(spy)[\uFEFF(\"calledOnce\")]", want: true},
		{text: "expect(spy)[// key\u2028(\"calledOnce\")]", want: true},
		{text: `expect(spy)["calledOnce"]`, want: false},
		{text: `rows[index]`, want: false},
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
