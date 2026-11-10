package better_dom_traversing_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/better_dom_traversing"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBetterDomTraversingExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &better_dom_traversing.BetterDomTraversingRule, []rule_tester.ValidTestCase{
		// Only nonnegative, safe Number integers count as indexes.
		{
			Code: "element.children[-0]; element.children[0n]; element.children[9007199254740992]; element.children[1e999];", FileName: "probe.js",
		},
		// Rest targets and delete still write through the indexed access.
		{
			Code: "[...element.children[0]] = nodes; ({...element.childNodes[0]} = object); delete element?.root.children[0];", FileName: "probe.js",
		},
		// TypeScript assertion wrappers are transparent for write targets.
		{
			Code: "(element.children[0] as any) = node; delete element.children[0]!; [element.children[0] as any] = nodes; [element.children[0]!] = nodes;", FileName: "probe.ts",
		},
		// Dotted JSX tags are not runtime member expressions.
		{
			Code: "const A = <element.parentElement.parentElement />;", FileName: "probe.jsx",
		},
		// Only a heritage clause's entity name maps to ESTree member access.
		{
			Code: "type A = B.parentElement.parentElement; interface C extends Base<B.parentElement.parentElement> {}", FileName: "probe.ts",
		},
		{
			Code: "element[\"querySelector\"](\"a\").querySelector(\"b\"); element.querySelector(...[\"a\"]).querySelector(\"b\");", FileName: "probe.js",
		},
		{
			Code: "(node.querySelector(\"a\") as Element).querySelector(\"b\"); node.querySelector(\"a\" as string).querySelector(\"b\");", FileName: "probe.ts",
		},
	}, []rule_tester.InvalidTestCase{
		// Keep a declaration-like receiver grouped after export default.
		{
			Code: "export default (function(){}.node.querySelector(\"a\")).querySelector(\"b\");", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 16, 1, 73, "export default (function(){}.node).querySelector(\":scope a b\");"),
			},
		},
		{
			Code: "component?.props.children[0];", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("first-element-child", 1, 1, 1, 29, "component?.props.firstElementChild;"),
			},
		},
		{
			Code: "((element.querySelector))(\"a\").querySelector(\"b\");", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 1, 1, 50, "element.querySelector(\":scope a b\");"),
			},
		},
		{
			Code: "/** @type {Element} */ (element.querySelector(\"a\")).querySelector(\"b\");", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 24, 1, 71, "/** @type {Element} */ element.querySelector(\":scope a b\");"),
			},
		},
		{
			Code: "element.children[(0x0)]; element.children[1_000]; element.children[9007199254740991];", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("first-element-child", 1, 1, 1, 24, "element.firstElementChild; element.children[1_000]; element.children[9007199254740991];"),
				traversalError("query-selector", 1, 26, 1, 49, ""),
				traversalError("query-selector", 1, 51, 1, 85, ""),
			},
		},
		// tsgo accepts explicit octal prefixes and rejects legacy octal syntax.
		{
			Code: "element.children[0o10];", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("query-selector", 1, 1, 1, 23, ""),
			},
		},
		// Computed keys and default values in patterns are reads.
		{
			Code: "const {child = element.children[0]} = data; ({[element.children[0]]: child} = data);", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("first-element-child", 1, 16, 1, 35, "const {child = element.firstElementChild} = data; ({[element.children[0]]: child} = data);"),
				traversalError("first-element-child", 1, 48, 1, 67, "const {child = element.children[0]} = data; ({[element.firstElementChild]: child} = data);"),
			},
		},
		// An optional receiver and an optional collection have different AST boundaries.
		{
			Code: "element?.root.children[0]; (element?.root).children[0]; (element?.root.children)[0];", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("first-element-child", 1, 1, 1, 26, "element?.root.firstElementChild; (element?.root).children[0]; (element?.root.children)[0];"),
				traversalError("first-element-child", 1, 28, 1, 55, "element?.root.children[0]; (element?.root).firstElementChild; (element?.root.children)[0];"),
			},
		},
		// A grouped optional chain is a separate receiver.
		{
			Code: "(element?.root.children[0]).children[1]; element.children[0].children[1];", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("query-selector", 1, 1, 1, 40, ""),
				traversalError("first-element-child", 1, 2, 1, 27, "(element?.root.firstElementChild).children[1]; element.children[0].children[1];"),
				traversalError("query-selector", 1, 42, 1, 73, ""),
			},
		},
		// Parent traversal stops at the boundary of a parenthesized optional chain.
		{
			Code: "element?.root.parentElement.parentElement; (element?.root.parentElement).parentElement; (element?.root.parentElement.parentElement).parentElement;", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("closest", 1, 1, 1, 42, ""),
				traversalError("closest", 1, 90, 1, 131, ""),
			},
		},
		{
			Code: "class A implements B.parentElement.parentElement {}", FileName: "probe.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("closest", 1, 20, 1, 49, ""),
			},
		},
		{
			Code: "interface A extends B.parentElement.parentElement {}", FileName: "probe.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("closest", 1, 21, 1, 50, ""),
			},
		},
		{
			Code: "class A extends B.parentElement.parentElement {}", FileName: "probe.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("closest", 1, 17, 1, 46, ""),
			},
		},
		{
			Code: "fn?.(element.querySelector(\"a\").querySelector(\"b\"));", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 6, 1, 51, "fn?.(element.querySelector(\":scope a b\"));"),
			},
		},
		// An assertion interrupts the ESTree walk toward a later optional link.
		{
			Code: "element.querySelector(\"a\").querySelector(\"b\")!?.x; element?.querySelector(\"a\").querySelector(\"b\")!.x;", FileName: "probe.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 1, 1, 46, "element.querySelector(\":scope a b\")!?.x; element?.querySelector(\"a\").querySelector(\"b\")!.x;"),
			},
		},
		{
			Code: "element.querySelector<Element>(\"a\").querySelector<HTMLElement>(\"b\");", FileName: "probe.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 1, 1, 68, "element.querySelector(\":scope a b\");"),
			},
		},
		{
			Code: "window[\"document\"].querySelector(\"a\").querySelector(\"b\"); class C extends B { f() { return super.querySelector(\"a\").querySelector(\"b\"); } }", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 1, 1, 57, "window[\"document\"].querySelector(\":scope a b\"); class C extends B { f() { return super.querySelector(\"a\").querySelector(\"b\"); } }"),
				traversalError("merge-query-selector", 1, 92, 1, 135, "window[\"document\"].querySelector(\"a\").querySelector(\"b\"); class C extends B { f() { return super.querySelector(\":scope a b\"); } }"),
			},
		},
		// Selector values are decoded, then quoted in the first selector style.
		{
			Code: "const 例 = \"😀\"; document.querySelector(\"#\\u{1f600}\\u2028\").querySelector('button[title=\"x\"]');", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 17, 1, 94, "const 例 = \"😀\"; document.querySelector(\"#😀\\u{2028} button[title=\\\"x\\\"]\");"),
			},
		},
		// A lone surrogate needs an escape in the suggested literal.
		{
			Code: "element.querySelector(\"[title='\\ud800']\").querySelector(\"span\");", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 1, 1, 64, "element.querySelector(\":scope [title='\\u{d800}'] span\");"),
			},
		},
		{
			Code: `document.querySelector('\uD83D\uDE00').querySelector('span');`, FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 1, 1, 61, "document.querySelector('😀 span');"),
			},
		},
		// Keep surrounding comments and recognize comment-like text in a selector.
		{
			Code:     `/* keep */ document.querySelector("a[href^='https://example.test']").querySelector("span"); // keep`,
			FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 12, 1, 91, `/* keep */ document.querySelector("a[href^='https://example.test'] span"); // keep`),
			},
		},
		{
			Code: "(void 0).children[0]; (void 0).querySelector(\"a\").querySelector(\"b\");", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("first-element-child", 1, 1, 1, 21, "(void 0).firstElementChild; (void 0).querySelector(\"a\").querySelector(\"b\");"),
				traversalError("merge-query-selector", 1, 23, 1, 69, "(void 0).children[0]; (void 0).querySelector(\":scope a b\");"),
			},
		},
		// Keep keywords separate from the property or receiver introduced by a suggestion.
		{
			Code: "function f() { return(element.querySelector(\"a\")).querySelector(\"b\"); } const exists = element.children[0]in nodes;", FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 22, 1, 69, "function f() { return element.querySelector(\":scope a b\"); } const exists = element.children[0]in nodes;"),
				traversalError("first-element-child", 1, 88, 1, 107, "function f() { return(element.querySelector(\"a\")).querySelector(\"b\"); } const exists = element.firstElementChild in nodes;"),
			},
		},
		{
			Code: "const child = element.children[0]as Node;", FileName: "probe.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("first-element-child", 1, 15, 1, 34, "const child = element.firstElementChild as Node;"),
			},
		},
		{
			Code: `function read() { return(\u0072oot.querySelector('a')).querySelector('b'); }`, FileName: "probe.js",
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 25, 1, 74, `function read() { return \u0072oot.querySelector(':scope a b'); }`),
			},
		},
		// Keep let[0] grouped in a for initializer.
		{
			Code: "for ((let[0].querySelector(\"a\")).querySelector(\"b\");;);", FileName: "probe.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Errors: []rule_tester.InvalidTestCaseError{
				traversalError("merge-query-selector", 1, 6, 1, 52, "for ((let[0]).querySelector(\":scope a b\");;);"),
			},
		},
	})
}

func TestBetterDomTraversingEditDemand(t *testing.T) {
	const source = "element.childNodes[0]; element.children[0]; element.querySelector('a').querySelector('b'); element.children[1];"
	compiler, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	program := lintprogram.NewFromCompiler(compiler)
	lint := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program,
			File:    file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     better_dom_traversing.BetterDomTraversingRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return better_dom_traversing.BetterDomTraversingRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		return diagnostics
	}
	all := lint(rule.EditDemandAll)
	if len(all) != 4 {
		t.Fatalf("all edits: got %d diagnostics, want 4", len(all))
	}
	for index, diagnostic := range all {
		if (diagnostic.Suggestions != nil) != (index < 3) {
			t.Fatalf("all edits: unexpected suggestions for %q", diagnostic.Message.Id)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		got := lint(demand)
		if len(got) != len(all) {
			t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(got), len(all))
		}
		for index, diagnostic := range got {
			baseline := all[index]
			if !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Range != baseline.Range ||
				diagnostic.RuleName != baseline.RuleName || diagnostic.Severity != baseline.Severity {
				t.Errorf("demand %d: changed diagnostic %d", demand, index)
			}
			if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d: unexpected autofix for %q", demand, diagnostic.Message.Id)
			}
			var suggestions *[]rule.RuleSuggestion
			if demand&rule.EditDemandSuggestion != 0 {
				suggestions = baseline.Suggestions
			}
			if !reflect.DeepEqual(diagnostic.Suggestions, suggestions) {
				t.Errorf("demand %d: unexpected suggestions for %q", demand, diagnostic.Message.Id)
			}
		}
	}
}
