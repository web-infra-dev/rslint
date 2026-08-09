package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func TestRstestCallAnalysisCandidateNames(t *testing.T) {
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/candidate.test.ts",
			Path:     "/candidate.test.ts",
		},
		`import { test as importedTest, expect as importedExpect } from "@rstest/core";
const testAlias = importedTest;
const testCase = testAlias;
const expectAlias = importedExpect;
const check = expectAlias;`,
		core.ScriptKindTS,
	)
	analysis := NewRstestCallAnalysis(rule.RuleContext{SourceFile: sourceFile})
	for _, name := range []string{"importedTest", "testAlias", "testCase"} {
		if analysis.candidates[name]&rstestCandidateFn == 0 {
			t.Errorf("registration candidate %q was not propagated", name)
		}
		if analysis.candidates[name]&rstestCandidateTest == 0 {
			t.Errorf("test candidate %q was not propagated", name)
		}
	}
	for _, name := range []string{"importedExpect", "expectAlias", "check"} {
		if analysis.candidates[name]&rstestCandidateExpect == 0 {
			t.Errorf("expect candidate %q was not propagated", name)
		}
	}
}

func TestRstestCallAnalysisParseFnCallComputesCacheMiss(t *testing.T) {
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/callbacks.test.ts",
			Path:     "/callbacks.test.ts",
		},
		`test("late", () => {});
describe("late suite", () => {});
beforeEach(() => {});`,
		core.ScriptKindTS,
	)
	analysis := NewRstestCallAnalysis(rule.RuleContext{SourceFile: sourceFile})
	lateCalls := make([]*ast.Node, 0, 3)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			lateCalls = append(lateCalls, node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.Node.ForEachChild(visit)
	if len(lateCalls) != 3 {
		t.Fatalf("found %d late calls, want 3", len(lateCalls))
	}
	if len(analysis.fnCalls) != 0 {
		t.Fatalf("analysis eagerly parsed %d calls, want 0", len(analysis.fnCalls))
	}

	wantKinds := []RstestFnType{
		RstestFnTypeTest,
		RstestFnTypeDescribe,
		RstestFnTypeHook,
	}
	for index, lateCall := range lateCalls {
		parsed := analysis.ParseFnCall(lateCall)
		if parsed == nil || parsed.Kind != wantKinds[index] {
			t.Fatalf("cache miss %d parsed as %#v, want %q", index, parsed, wantKinds[index])
		}
		if cached := analysis.ParseFnCall(lateCall); cached != parsed {
			t.Fatalf("second parse %d did not reuse the cached result", index)
		}
	}
}

func TestRstestCallAnalysisCallbacksWidenExpectCandidates(t *testing.T) {
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/context.test.ts",
			Path:     "/context.test.ts",
		},
		`test("context", ({ expect: check }) => {
  check(1).toBe(1);
});`,
		core.ScriptKindTS,
	)
	analysis := NewRstestCallAnalysis(rule.RuleContext{SourceFile: sourceFile})
	var checkCall *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			root := testFramework.ResolveFirstIdentifier(node.AsCallExpression().Expression)
			if root != nil && root.Kind == ast.KindIdentifier &&
				root.AsIdentifier().Text == "check" {
				checkCall = node
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.Node.ForEachChild(visit)
	if checkCall == nil {
		t.Fatal("context expect call not found")
	}
	if analysis.isExpectCandidate(checkCall) {
		t.Fatal("context expect was a candidate before callbacks were collected")
	}
	analysis.Callbacks()
	if !analysis.isExpectCandidate(checkCall) {
		t.Fatal("callbacks did not add the context expect candidate")
	}
}

func TestRstestCallAnalysisCallbacksSkipFilesWithoutTests(t *testing.T) {
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/expect-only.test.ts",
			Path:     "/expect-only.test.ts",
		},
		`expect(value).toBe(1); helper();`,
		core.ScriptKindTS,
	)
	analysis := NewRstestCallAnalysis(rule.RuleContext{SourceFile: sourceFile})
	if analysis.hasTests {
		t.Fatal("expect-only file unexpectedly has a test candidate")
	}
	callbacks := analysis.Callbacks()
	if len(analysis.fnCalls) != 0 {
		t.Fatalf("callbacks parsed %d calls without a test candidate", len(analysis.fnCalls))
	}
	if len(callbacks.Functions) != 0 ||
		len(callbacks.ContextReceivers) != 0 ||
		len(callbacks.ContextExpectNames) != 0 {
		t.Fatalf("callbacks = %#v, want empty maps", callbacks)
	}
}

func TestRstestCallAnalysisParseExpectCallFillsIdentityCache(t *testing.T) {
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/expect-cache.test.ts",
			Path:     "/expect-cache.test.ts",
		},
		`expect(value).toBe(1);`,
		core.ScriptKindTS,
	)
	analysis := NewRstestCallAnalysis(rule.RuleContext{SourceFile: sourceFile})
	var expectCall *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression &&
			FindTopMostCallExpression(node) == node {
			expectCall = node
			return true
		}
		return node.ForEachChild(visit)
	}
	sourceFile.Node.ForEachChild(visit)
	if expectCall == nil {
		t.Fatal("expect call not found")
	}
	if parsed := analysis.ParseExpectCall(expectCall); parsed == nil {
		t.Fatal("full expect parse returned nil")
	}
	if result, ok := analysis.isExpect[expectCall]; !ok || !result {
		t.Fatalf("identity cache = (%t, %t), want (true, true)", result, ok)
	}
}

func TestRstestCandidateSeedsMatchDirectRegistrationNames(t *testing.T) {
	actual := cloneRstestCandidateSeeds()
	if len(actual) != len(rstestDirectAPIStates)+1 {
		t.Fatalf(
			"seed count = %d, want %d",
			len(actual),
			len(rstestDirectAPIStates)+1,
		)
	}
	for name, states := range rstestDirectAPIStates {
		kind := actual[name]
		if kind&rstestCandidateFn == 0 {
			t.Errorf("registration %q is missing from seeds", name)
		}
		for profile, state := range states {
			if directRstestAPIState(rstestAPIProfile(profile), name) != state {
				t.Errorf(
					"profile %d registration %q = %d, want %d",
					profile,
					name,
					directRstestAPIState(rstestAPIProfile(profile), name),
					state,
				)
			}
		}
	}
	if actual["expect"] != rstestCandidateExpect {
		t.Errorf("expect seed = %b, want %b", actual["expect"], rstestCandidateExpect)
	}
	for _, hook := range testFramework.HooksOrder {
		if actual[hook]&rstestCandidateFn == 0 {
			t.Errorf("hook %q is missing from registration seeds", hook)
		}
	}
}
