package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func parseAnalysisCacheFixture() *ast.SourceFile {
	return parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/shared-analysis.test.ts",
			Path:     "/shared-analysis.test.ts",
		},
		`test("case", () => {});`,
		core.ScriptKindTS,
	)
}

func TestRstestFirstIdentifierCacheMatchesUncached(t *testing.T) {
	for _, expression := range []string{
		`builder().step().step()`, `((builder)()).step()`,
		"tag`value`.step()", `(ignored(), builder).step()`,
		`builder[key()].step()`, `builder?.['step']?.()`,
		`import.meta.rstest.expect(1).toBe(1)`, `(() => {})().step()`,
		`(left && right).step()`, `(left ? a : b).step()`,
		`(builder as any).step()`, `builder!.step()`,
		`(builder satisfies Callable).step()`, `new Builder().step()`,
	} {
		t.Run(expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/chain.ts", Path: "/chain.ts",
			}, expression+";", core.ScriptKindTS)
			analysis := newRstestCallAnalysis(rule.RuleContext{SourceFile: source})
			var visit func(*ast.Node)
			visit = func(node *ast.Node) {
				want := testFramework.ResolveFirstIdentifier(node)
				if got := analysis.firstIdentifier(node); got != want {
					t.Fatalf("kind %v: root = %p, want %p", node.Kind, got, want)
				}
				if got, ok := analysis.firstIdentifiers[node]; node.Kind != ast.KindIdentifier && (!ok || got != want) {
					t.Fatalf("kind %v: cached root = (%p, %t), want (%p, true)", node.Kind, got, ok, want)
				}
				if got := analysis.firstIdentifier(node); got != want {
					t.Fatal("cached lookup changed the result")
				}
				node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
			}
			visit(source.Node.AsNode())
		})
	}
}

func TestRstestOuterCallMatchesTopMost(t *testing.T) {
	for _, expression := range []string{
		`expect(1).toBe(1).then(done)`, `((expect(1)).toBe)(1)`,
		`consume(expect(1).toBe(1))`, `object[expect(1).toBe(1)]()`,
		`expect(expect(1).toBe(1)).toBe(1)`, `expect(1).to.be.ok`,
		`expect?.(1)?.['toBe']?.(1)`, "expect(1).tag`value`",
		`(0, expect(1)).toBe(1)`, `(expect(1) as any).toBe(1)`,
		`expect(1)!.toBe(1)`, `(expect(1) satisfies Assertion).toBe(1)`,
	} {
		t.Run(expression, func(t *testing.T) {
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/chain.ts", Path: "/chain.ts",
			}, expression+";", core.ScriptKindTS)
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindCallExpression {
					if got, want := hasOuterCallOnCallee(node), FindTopMostCallExpression(node) != node; got != want {
						t.Fatalf("call at %d: outer = %t, want %t", node.Pos(), got, want)
					}
				}
				return node.ForEachChild(visit)
			}
			source.Node.ForEachChild(visit)
		})
	}
}

func TestRstestFirstIdentifierCachesWholeChain(t *testing.T) {
	for _, head := range []string{"builder()", "import.meta.rstest.expect(1)"} {
		source := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: "/chain.ts", Path: "/chain.ts",
		}, `test("case", () => { `+head+strings.Repeat(".step()", 2000)+`; });`, core.ScriptKindTS)
		analysis := newRstestCallAnalysis(rule.RuleContext{SourceFile: source})
		outer := analysis.calls[1].AsCallExpression().Expression
		want := testFramework.ResolveFirstIdentifier(outer)
		analysis.firstIdentifier(outer)
		count := len(analysis.firstIdentifiers)
		for _, call := range analysis.calls[1:] {
			expression := call.AsCallExpression().Expression
			if expression.Kind == ast.KindIdentifier {
				if got := analysis.firstIdentifier(expression); got != want {
					t.Fatalf("identifier root = %p, want %p", got, want)
				}
				continue
			}
			if got, ok := analysis.firstIdentifiers[expression]; !ok || got != want {
				t.Fatalf("child chain was not cached: (%p, %t), want (%p, true)", got, ok, want)
			}
			analysis.firstIdentifier(expression)
		}
		if len(analysis.firstIdentifiers) != count {
			t.Fatal("inner calls repeated chain-head resolution")
		}
	}
}

func BenchmarkRstestCallAnalysisChains(b *testing.B) {
	for _, kind := range []string{"ordinary", "extend", "expect"} {
		for _, size := range []int{500, 1000, 2000} {
			b.Run(fmt.Sprintf("%s/%d", kind, size), func(b *testing.B) {
				var code string
				switch kind {
				case "ordinary":
					code = `test("case", () => { builder()` + strings.Repeat(".step()", size) + `; });`
				case "extend":
					code = `test` + strings.Repeat(".extend({})", size) + `("case", () => {});`
				case "expect":
					code = `test("case", () => { expect(1)` + strings.Repeat(`.a("number")`, size) + `; });`
				}
				source := parser.ParseSourceFile(ast.SourceFileParseOptions{
					FileName: "/chain.ts", Path: "/chain.ts",
				}, code, core.ScriptKindTS)
				b.ReportAllocs()
				b.SetBytes(int64(len(code)))
				b.ResetTimer()
				for b.Loop() {
					analysis := newRstestCallAnalysis(rule.RuleContext{SourceFile: source})
					for _, call := range analysis.calls {
						analysis.ParseFnCall(call)
						analysis.ParseExpectCall(call)
						analysis.IsExpectCall(call)
					}
				}
			})
		}
	}
}

func TestGetRstestCallAnalysisSharesWithinFileCache(t *testing.T) {
	sourceFile := parseAnalysisCacheFixture()
	cache := rule.NewFileCache()
	first := GetRstestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
		Settings:   map[string]interface{}{"owner": "first"},
	}.WithFileCache(cache))
	second := GetRstestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
		Settings:   map[string]interface{}{"owner": "second"},
	}.WithFileCache(cache))
	if first != second {
		t.Fatal("contexts for one file did not share their analysis")
	}
	if first.ctx.Settings != nil {
		t.Fatal("shared analysis retained rule-specific context fields")
	}
}

func TestGetRstestCallAnalysisSeparatesFileCaches(t *testing.T) {
	sourceFile := parseAnalysisCacheFixture()
	first := GetRstestCallAnalysis(
		rule.RuleContext{SourceFile: sourceFile}.WithFileCache(rule.NewFileCache()),
	)
	second := GetRstestCallAnalysis(
		rule.RuleContext{SourceFile: sourceFile}.WithFileCache(rule.NewFileCache()),
	)
	if first == second {
		t.Fatal("different file caches shared an analysis")
	}
}

func TestGetRstestCallAnalysisWithoutFileCache(t *testing.T) {
	ctx := rule.RuleContext{SourceFile: parseAnalysisCacheFixture()}
	first := GetRstestCallAnalysis(ctx)
	second := GetRstestCallAnalysis(ctx)
	if first == second {
		t.Fatal("context without a file cache unexpectedly promised sharing")
	}
}

func TestGetRstestCallAnalysisDropsReporter(t *testing.T) {
	sourceFile := parseAnalysisCacheFixture()
	ctx := rule.RuleContext{
		SourceFile: sourceFile,
	}.WithFileCache(rule.NewFileCache()).WithReporter(
		"owner-rule",
		rule.SeverityWarning,
		func(rule.RuleDiagnostic) {},
	)
	analysis := GetRstestCallAnalysis(ctx)

	defer func() {
		if got := recover(); got != "rule: uninitialized RuleContext reporter" {
			t.Fatalf("analysis reporter panic = %v", got)
		}
	}()
	analysis.ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{})
}
