// Additional AST and scope regressions verified against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-dom-node-append.js
package prefer_dom_node_append_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_dom_node_append"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

func TestPreferDomNodeAppendExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_dom_node_append.PreferDomNodeAppendRule, []rule_tester.ValidTestCase{
		{Code: "(node?.appendChild)(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "node?.appendChild?.(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "node.appendChild(undefined);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "node.appendChild(undefined);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"undefined": "off"}},
		{Code: "/* global undefined: off */ node.appendChild(undefined);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "node.appendChild(null);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "node.appendChild(void 0);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "node.appendChild(() => child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "node.appendChild({});", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "node.appendChild(`text`);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "class C { #appendChild() {} f() { this.#appendChild(child); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(node.appendChild as Function)(child);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "(node.appendChild)(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(node.append)(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(node.appendChild(child));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(node.append(child));"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = (node.appendChild(child));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(child).appendChild(grandchild);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node.appendChild(child).append(grandchild);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "for (node.appendChild(child); ready; node.appendChild(other)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (ready) node.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"if (ready) node.append(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "void node.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const f = () => node.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(child), node.appendChild(other);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(child?.element);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node.append(child?.element);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(getChild());", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node.append(getChild());"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node./* keep */ appendChild(/* child */ child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node./* keep */ append(/* child */ child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; node.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; node.append(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/** @type {Node} */ (node).appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"/** @type {Node} */ (node).append(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(node as Element).appendChild(child);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(node as Element).append(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(child as Node);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node.append(child as Node);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild<string>(child);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node.append<string>(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestPreferDomNodeAppendArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct{ source, output string }{{source: "node.appendChild(child);", output: "node.append(child);"},
		{source: "const foo = node.appendChild(child);", output: "const foo = node.appendChild(child);"}} {
		t.Run(testCase.source, func(t *testing.T) {
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			demands := []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll}
			diagnostics := make([]rule.RuleDiagnostic, len(demands))
			for index, demand := range demands {
				var found []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: prefer_dom_node_append.PreferDomNodeAppendRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return prefer_dom_node_append.PreferDomNodeAppendRule.Run(ctx, nil)
						}}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { found = append(found, d) }},
				})
				if len(found) != 1 {
					t.Fatalf("demand %d: expected one diagnostic, got %d", demand, len(found))
				}
				diagnostics[index] = found[0]
			}
			all := diagnostics[len(diagnostics)-1]
			for index, demand := range demands {
				diagnostic := diagnostics[index]
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.source != testCase.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected fix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed fix artifacts", demand)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d: autofix-only rule produced suggestions", demand)
				}
			}
			// Applying fixes can sort slices in place; identity checks are finished.
			for index, demand := range demands {
				wantFix := testCase.source != testCase.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				expected := testCase.source
				if wantFix {
					expected = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, diagnostics[index:index+1])
				if output != expected || fixed != wantFix {
					t.Errorf("demand %d: got %q, want %q", demand, output, expected)
				}
			}
		})
	}
}

func TestPreferDomNodeAppendReviewRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_dom_node_append.PreferDomNodeAppendRule, []rule_tester.ValidTestCase{
		{Code: "node.appendChild(void 0);", FileName: "review.js"},
		{Code: "node.appendChild((sideEffect(), null));", FileName: "review.js"},
		{Code: "node.appendChild(null as unknown as Node);", FileName: "review.ts"},
	}, []rule_tester.InvalidTestCase{
		{
			Code:     "function f(undefined) { node.appendChild(undefined); }",
			FileName: "review.js",
			Output:   []string{"function f(undefined) { node.append(undefined); }"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "prefer-dom-node-append",
			}},
		},
	})
}
