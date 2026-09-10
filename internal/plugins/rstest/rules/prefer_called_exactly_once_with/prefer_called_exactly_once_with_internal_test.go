package prefer_called_exactly_once_with

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
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

func TestOnlyInertStatementIntervals(t *testing.T) {
	tests := []struct {
		name   string
		middle string
		want   bool
	}{
		{name: "adjacent", want: true},
		{name: "comments", middle: "// keep\n/* keep */", want: true},
		{name: "declarations", middle: "const unrelated = 1; let value; var other = 2;", want: true},
		{name: "assertions", middle: "expect(other).toBe(1); expect(third).toEqual({ value: 2 });", want: true},
		{name: "first barrier", middle: "spy.mockClear(); const unrelated = 1;"},
		{name: "last barrier", middle: "const unrelated = 1; spy.mockClear();"},
		{name: "empty statement", middle: ";"},
		{name: "nested barrier", middle: "if (flag) { spy.mockReset(); }"},
		{name: "reassignment", middle: "spy = other;"},
		{name: "target shadow", middle: "var spy;"},
		{name: "argument shadow", middle: "var expected;"},
		{name: "matcher shadow", middle: "var toHaveBeenCalledWith;"},
		{name: "unknown call without checker", middle: "console.log('checkpoint');"},
		{name: "await", middle: "await expect(other).resolves.toBe(1);"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := "before();\nexpect(spy).toHaveBeenCalledOnce();\n" + tt.middle + "\nexpect(spy).toHaveBeenCalledWith(expected);\nafter();"
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/case.ts", Path: "/case.ts"}, source, core.ScriptKindTS)
			ctx := rule.RuleContext{SourceFile: sourceFile}
			analysis := rstestUtils.GetRstestCallAnalysis(ctx)
			statements := sourceFile.Statements.Nodes
			firstIndex, secondIndex := 1, len(statements)-2
			first := mergeCandidateForStatement(analysis, sourceFile, statements[firstIndex])
			second := mergeCandidateForStatement(analysis, sourceFile, statements[secondIndex])
			if first == nil || second == nil {
				t.Fatal("expected two merge candidates")
			}
			first.statementIndex, second.statementIndex = firstIndex, secondIndex
			for _, pair := range [][2]*mergeCandidate{{first, second}, {second, first}} {
				if got := onlyInertStatementsBetween(ctx, analysis, statements, pair[0], pair[1]); got != tt.want {
					t.Errorf("interval (%d, %d) = %t, want %t", pair[0].statementIndex, pair[1].statementIndex, got, tt.want)
				}
			}
		})
	}
}

func TestCheckBlockStatementIndices(t *testing.T) {
	source := `before();
expect(first).toHaveBeenCalledOnce();
expect(first).toHaveBeenCalledWith(1);
between();
expect(second).toHaveBeenCalledWith(2);
const unrelated = 0;
expect(second).toHaveBeenCalledOnce();
expect(third).calledOnce.and.calledWith(3);
expect(blocked).toHaveBeenCalledOnce();
blocked.mockReset();
expect(blocked).toHaveBeenCalledWith(4);
after();`
	for _, nested := range []bool{false, true} {
		t.Run(fmt.Sprintf("nested=%t", nested), func(t *testing.T) {
			code := source
			if nested {
				code = "{\n" + source + "\n}"
			}
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/case.ts", Path: "/case.ts"}, code, core.ScriptKindTS)
			var diagnostics []rule.RuleDiagnostic
			ctx := (rule.RuleContext{SourceFile: sourceFile}).WithReporter(PreferCalledExactlyOnceWithRule.Name, rule.SeverityError, func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			})
			node := sourceFile.AsNode()
			if nested {
				node = sourceFile.Statements.Nodes[0]
			}
			checkBlock(ctx, rstestUtils.GetRstestCallAnalysis(ctx), node)
			if len(diagnostics) != 3 {
				t.Fatalf("got %d diagnostics, want 3", len(diagnostics))
			}
			for i, matcher := range []string{"toHaveBeenCalledWith", "toHaveBeenCalledOnce", "calledWith"} {
				diagnostic := diagnostics[i]
				if text := code[diagnostic.Range.Pos():diagnostic.Range.End()]; text != matcher {
					t.Errorf("diagnostic %d: got %q, want %q", i, text, matcher)
				}
				wantFixes := 2
				if i == 2 {
					wantFixes = 0
				}
				if len(diagnostic.Fixes()) != wantFixes {
					t.Errorf("diagnostic %d: got %d fixes, want %d", i, len(diagnostic.Fixes()), wantFixes)
				}
			}
		})
	}
}

func BenchmarkAdjacentPairs(b *testing.B) {
	for _, pairs := range []int{500, 1000, 2000} {
		b.Run(strconv.Itoa(pairs), func(b *testing.B) {
			var source strings.Builder
			for i := range pairs {
				fmt.Fprintf(&source, "expect(spy%d).toHaveBeenCalledOnce();\nexpect(spy%d).toHaveBeenCalledWith(1);\n", i, i)
			}
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/case.ts", Path: "/case.ts"}, source.String(), core.ScriptKindTS)
			b.ReportAllocs()
			b.SetBytes(int64(source.Len()))
			for b.Loop() {
				reports := 0
				ctx := (rule.RuleContext{SourceFile: sourceFile}).WithDiagnosticConsumer(PreferCalledExactlyOnceWithRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Report: func(rule.RuleDiagnostic) { reports++ },
				})
				PreferCalledExactlyOnceWithRule.Run(ctx, nil)
				if reports != pairs {
					b.Fatalf("got %d reports, want %d", reports, pairs)
				}
			}
		})
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
