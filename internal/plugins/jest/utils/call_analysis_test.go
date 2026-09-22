package utils

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func parseJestAnalysisFixture(source string) *ast.SourceFile {
	return parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/shared-jest-analysis.test.ts",
			Path:     "/shared-jest-analysis.test.ts",
		},
		source,
		core.ScriptKindTS,
	)
}

func jestAnalysisCalls(sourceFile *ast.SourceFile) []*ast.Node {
	var calls []*ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindCallExpression {
			calls = append(calls, node)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())
	return calls
}

func jestAnalysisCallWithEntries(
	t *testing.T,
	sourceFile *ast.SourceFile,
	names ...string,
) *ast.Node {
	t.Helper()
	for _, call := range jestAnalysisCalls(sourceFile) {
		entries := GetJestFnMemberEntries(call)
		if len(entries) != len(names) {
			continue
		}
		equal := true
		for index, name := range names {
			if entries[index].Name != name {
				equal = false
				break
			}
		}
		if equal {
			return call
		}
	}
	t.Fatalf("call %v not found", names)
	return nil
}

func TestGetJestCallAnalysisSharesWithinFileCache(t *testing.T) {
	sourceFile := parseJestAnalysisFixture(`test("case", () => {});`)
	cache := rule.NewFileCache()
	first := GetJestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
		Settings:   map[string]interface{}{"owner": "first"},
	}.WithFileCache(cache))
	second := GetJestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
		Settings:   map[string]interface{}{"owner": "second"},
	}.WithFileCache(cache))
	if first != second {
		t.Fatal("contexts for one file did not share their analysis")
	}
}

func TestGetJestCallAnalysisSeparatesFileCaches(t *testing.T) {
	sourceFile := parseJestAnalysisFixture(`test("case", () => {});`)
	first := GetJestCallAnalysis(
		rule.RuleContext{SourceFile: sourceFile}.WithFileCache(rule.NewFileCache()),
	)
	second := GetJestCallAnalysis(
		rule.RuleContext{SourceFile: sourceFile}.WithFileCache(rule.NewFileCache()),
	)
	if first == second {
		t.Fatal("different file caches shared an analysis")
	}
}

func TestGetJestCallAnalysisWithoutFileCache(t *testing.T) {
	ctx := rule.RuleContext{SourceFile: parseJestAnalysisFixture(`test("case", () => {});`)}
	first := GetJestCallAnalysis(ctx)
	second := GetJestCallAnalysis(ctx)
	if first == second {
		t.Fatal("context without a file cache unexpectedly promised sharing")
	}
}

func TestGetJestCallAnalysisDropsReporter(t *testing.T) {
	sourceFile := parseJestAnalysisFixture(`test("case", () => {});`)
	ctx := rule.RuleContext{
		SourceFile: sourceFile,
	}.WithFileCache(rule.NewFileCache()).WithReporter(
		"owner-rule",
		rule.SeverityWarning,
		func(rule.RuleDiagnostic) {},
	)
	analysis := GetJestCallAnalysis(ctx)

	defer func() {
		if got := recover(); got != "rule: uninitialized RuleContext reporter" {
			t.Fatalf("analysis reporter panic = %v", got)
		}
	}()
	analysis.ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{})
}

func TestJestCallAnalysisCachesSuccessMissAndReason(t *testing.T) {
	sourceFile := parseJestAnalysisFixture(`
helper();
test("case", () => expect(value).toBe(true));
`)
	analysis := GetJestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
	}.WithFileCache(rule.NewFileCache()))

	helper := jestAnalysisCallWithEntries(t, sourceFile, "helper")
	if analysis.ParseFnCall(helper) != nil {
		t.Fatal("ordinary call parsed as Jest")
	}
	if result, ok := analysis.fnCalls[helper]; !ok || result.parsed != nil || result.reason != "" {
		t.Fatalf("ordinary-call cache = %#v, present = %v", result, ok)
	}

	testCall := jestAnalysisCallWithEntries(t, sourceFile, "test")
	first := analysis.ParseTestCall(testCall)
	second := analysis.ParseTestCall(testCall)
	if first == nil || first != second {
		t.Fatal("successful test parse was not cached")
	}

	missingMatcherFile := parseJestAnalysisFixture(`expect(value);`)
	missingMatcherAnalysis := GetJestCallAnalysis(rule.RuleContext{
		SourceFile: missingMatcherFile,
	}.WithFileCache(rule.NewFileCache()))
	missingMatcher := jestAnalysisCallWithEntries(t, missingMatcherFile, "expect")
	if parsed, reason := missingMatcherAnalysis.ParseExpectCallWithReason(missingMatcher); parsed != nil || reason != ExpectParseReasonMatcherNotFound {
		t.Fatalf("bare expect = (%v, %q), want matcher-not-found", parsed, reason)
	}

	matcherNotCalledFile := parseJestAnalysisFixture(`expect(value).not;`)
	matcherNotCalledAnalysis := GetJestCallAnalysis(rule.RuleContext{
		SourceFile: matcherNotCalledFile,
	}.WithFileCache(rule.NewFileCache()))
	matcherNotCalled := jestAnalysisCallWithEntries(t, matcherNotCalledFile, "expect")
	if parsed, reason := matcherNotCalledAnalysis.ParseExpectCallWithReason(matcherNotCalled); parsed != nil || reason != ExpectParseReasonMatcherNotCalled {
		t.Fatalf("uncalled matcher = (%v, %q), want matcher-not-called", parsed, reason)
	}

	rejectedCases := []struct {
		source string
		want   string
	}{
		{source: `expect(1).not.not.each();`, want: ExpectParseReasonModifierUnknown},
		{source: "expect`value`();", want: ExpectParseReasonMatcherNotFound},
	}
	for _, test := range rejectedCases {
		t.Run(test.source, func(t *testing.T) {
			file := parseJestAnalysisFixture(test.source)
			analysis := GetJestCallAnalysis(rule.RuleContext{
				SourceFile: file,
			}.WithFileCache(rule.NewFileCache()))
			calls := jestAnalysisCalls(file)
			if len(calls) == 0 {
				t.Fatal("fixture contains no call")
			}
			call := calls[0]
			if analysis.ParseFnCall(call) != nil {
				t.Fatal("rejected syntax parsed as a successful Jest call")
			}
			if parsed, reason := analysis.ParseExpectCallWithReason(call); parsed != nil || reason != test.want {
				t.Fatalf("rejected expect = (%v, %q), want %q", parsed, reason, test.want)
			}
		})
	}
}

func TestJestCallAnalysisCollectsCallbacksLazily(t *testing.T) {
	sourceFile := parseJestAnalysisFixture(`
test("case", callback);
function callback() { expect(value).toBe(true); }
let uninitialized;
`)
	analysis := GetJestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
	}.WithFileCache(rule.NewFileCache()))
	if analysis.indexed || analysis.callbacksOK {
		t.Fatal("callback analysis was built eagerly")
	}

	callbacks := analysis.Callbacks()
	if !analysis.indexed || !analysis.callbacksOK {
		t.Fatal("callback analysis was not memoized")
	}
	if len(callbacks.Functions) != 1 {
		t.Fatalf("callback functions = %d, want 1", len(callbacks.Functions))
	}
	for function := range callbacks.Functions {
		if function.Kind != ast.KindFunctionDeclaration {
			t.Fatalf("callback kind = %v, want function declaration", function.Kind)
		}
	}
	second := analysis.Callbacks()
	if len(second.Functions) != len(callbacks.Functions) {
		t.Fatal("repeated callback lookup changed the cached result")
	}
}

func TestJestCallAnalysisUsesFileSettings(t *testing.T) {
	sourceFile := parseJestAnalysisFixture(`context("case", () => {});`)
	analysis := GetJestCallAnalysis(rule.RuleContext{
		SourceFile: sourceFile,
		Settings: map[string]interface{}{
			"jest": map[string]interface{}{
				"globalAliases": map[string]interface{}{
					"describe": []interface{}{"context"},
				},
			},
		},
	}.WithFileCache(rule.NewFileCache()))
	call := jestAnalysisCallWithEntries(t, sourceFile, "context")
	parsed := analysis.ParseFnCall(call)
	if parsed == nil || parsed.Kind != JestFnTypeDescribe || parsed.Name != "describe" {
		t.Fatalf("aliased global parse = %#v", parsed)
	}
}
