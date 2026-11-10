package prefer_at

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferAtExtras(t *testing.T) {
	checkAll := map[string]any{"checkAllIndexAccess": true}
	defaults := map[string]any{"checkAllIndexAccess": false, "getLastElementFunctions": []any{}}
	testRule := PreferAtRule
	testRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return PreferAtRule.Run(ctx, options)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testRule,
		[]rule_tester.ValidTestCase{
			{Code: "items[0]; text.charAt(0);", FileName: "file.js", Options: defaults},
			{Code: "array[9007199254740992]; array[array.length - 9007199254740992];", FileName: "file.js", Options: checkAll},
			{Code: "string.charAt([-1]); string.charAt('-1'); string.charAt(-1n);", FileName: "file.js", Options: checkAll},
			{Code: `array[array["length"] - 1]; array[array?.length - 1];`, FileName: "file.js"},
			{Code: "(arr?.slice(-1))[0]; arr.slice(-1, 0)[0];", FileName: "file.js"},
			{Code: "/** @param {?NodeList} nodes */ function read(nodes) { return _.last(nodes); }", FileName: "file.js"},
			{Code: "/** @param {function(): NodeList} read */ function f(read) { return _.last(read()); }", FileName: "file.js"},
			{Code: "/** @param {number} count @param {function(): NodeList} read */ function f(read, count) { return _.last(read()); }", FileName: "file.js"},
			{Code: "/** @param read {function(number):\n * (NodeList | HTMLCollection)} */ function f(read) { return _.last(read(0)); }", FileName: "file.js"},
			{Code: "/** @param {(() => NodeList) | (() => HTMLCollection)} read */ function f(read) { return _.last(read()); }", FileName: "file.js"},
			{Code: "/** @param {(() => NodeList) & (() => HTMLCollection)} read */ function f(read) { return _.last(read()); }", FileName: "file.js"},
			{Code: "/** @param {(() => NodeList) & {tag: string}} read */ function f(read) { return _.last(read()); }", FileName: "file.js"},
			{Code: "/** @type {(() => NodeList) | (() => HTMLCollection)} */ const read = makeReader(); _.last(read());", FileName: "file.js"},
			{Code: "type NodeList = { length: number; [index: number]: Element }; function f(items: NodeList | null) { return items?.[items.length - 1]; }", FileName: "file.ts"},
			{Code: "interface Elements extends NodeListOf<Element> {} function f(items: Elements) { return _.last(items); }", FileName: "file.ts"},
			{Code: "const read = (((): NodeList => document.childNodes) satisfies () => NodeList)!; _.last(read());", FileName: "file.ts"},
			{Code: "function read() { const last = (arguments!)[arguments.length - 1]; return _.last(arguments as any); }", FileName: "file.ts"},
			{Code: "function read() { arguments.slice = Array.prototype.slice; return arguments.slice(-1)[0]; }", FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "/** @param {function(): NodeList} other */\nfunction f(read) { return _.last(read()); }", FileName: "file.js",
				Output: []string{"/** @param {function(): NodeList} other */\nfunction f(read) { return read().at(-1); }"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 2, 27, 2, 33, "_.last")},
			},
			{
				Code: "/** @param {(() => string[]) | (() => NodeList)} read */\nfunction f(read) { return _.last(read()); }", FileName: "file.js",
				Output: []string{"/** @param {(() => string[]) | (() => NodeList)} read */\nfunction f(read) { return read().at(-1); }"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 2, 27, 2, 33, "_.last")},
			},
			{
				Code: "/** @param {(() => string[]) & {(): NodeList}} read */\nfunction f(read) { return _.last(read()); }", FileName: "file.js",
				Output: []string{"/** @param {(() => string[]) & {(): NodeList}} read */\nfunction f(read) { return read().at(-1); }"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 2, 27, 2, 33, "_.last")},
			},
			{
				Code: "lodash.last(items);", FileName: "file.js", Options: defaults,
				Output: []string{"items.at(-1);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 1, 1, 1, 12, "lodash.last")},
			},
			{
				Code: "getLast(items); lodash.last(items); items[0];", FileName: "file.js",
				Options: map[string]any{"checkAllIndexAccess": true, "getLastElementFunctions": []any{"getLast"}},
				Output:  []string{"items.at(-1); items.at(-1); items.at(0);"},
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamError("get-last-function", 1, 1, 1, 8, "getLast"),
					upstreamError("get-last-function", 1, 17, 1, 28, "lodash.last"),
					upstreamError("index", 1, 43, 1, 44),
				},
			},
			{
				Code: "const node = <p>🍎{items[items.length -\n  1]}</p>;", FileName: "file.tsx",
				Output: []string{"const node = <p>🍎{items.at(-1)}</p>;"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("negative-index", 1, 26, 2, 4)},
			},
			{
				Code: `const items = store.items; store.items[store["it" + "ems"].length - 2 - 1];`, FileName: "file.js",
				Output: []string{"const items = store.items; store.items.at(- 2 - 1);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("negative-index", 1, 40, 1, 74)},
			},
			{
				Code: "arr[(arr.length as number) - 1];", FileName: "file.ts",
				Output: []string{"arr.at(-1);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("negative-index", 1, 5, 1, 31)},
			},
			{
				Code: "arr[((arr.length - 1) as number) - 2];", FileName: "file.ts",
				Output: []string{"arr.at(((- 1) as number) - 2);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("negative-index", 1, 5, 1, 37)},
			},
			{
				Code: "arr?.slice(-1).pop();", FileName: "file.js",
				Output: []string{"arr?.at(-1);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("slice", 1, 6, 1, 11)},
			},
			{
				Code: "const first = items[9007199254740991];", FileName: "file.js", Options: checkAll,
				Output: []string{"const first = items.at(9007199254740991);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("index", 1, 21, 1, 37)},
			},
			{
				Code: "function getNodes(): NodeList {return document.childNodes;} getNodes = getArray; _.last(getNodes());", FileName: "file.ts",
				Output: []string{"function getNodes(): NodeList {return document.childNodes;} getNodes = getArray; getNodes().at(-1);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 1, 82, 1, 88, "_.last")},
			},
			{
				Code: "async function getNodes(): NodeList {return document.childNodes;} _.last(getNodes());", FileName: "file.ts",
				Output: []string{"async function getNodes(): NodeList {return document.childNodes;} getNodes().at(-1);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 1, 67, 1, 73, "_.last")},
			},
			{
				Code: "function f(read: () => NodeList) { return _.last(read()); }", FileName: "file.ts",
				Output: []string{"function f(read: () => NodeList) { return read().at(-1); }"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 1, 43, 1, 49, "_.last")},
			},
			{
				Code: "function f(read = (): NodeList => document.childNodes) { return _.last(read()); }", FileName: "file.ts",
				Output: []string{"function f(read = (): NodeList => document.childNodes) { return read().at(-1); }"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 1, 65, 1, 71, "_.last")},
			},
			{
				Code: "const read = (((): NodeList => document.childNodes) as () => NodeList); _.last(read());", FileName: "file.ts",
				Output: []string{"const read = (((): NodeList => document.childNodes) as () => NodeList); read().at(-1);"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 1, 73, 1, 79, "_.last")},
			},
			{
				Code: "target[items[items.length - 1]]();", FileName: "file.js",
				Output: []string{"target[items.at(-1)]();"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("negative-index", 1, 14, 1, 30)},
			},
			{
				Code: "(/** @type {unknown} */ (items.slice(-1))).pop();", FileName: "file.js",
				Output: []string{"(/** @type {unknown} */ (items.at(-1)));"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("slice", 1, 32, 1, 37)},
			},
			{
				Code: "const constructor = new (items?.[items.length - 1]).Type();", FileName: "file.js",
				Output: []string{"const constructor = new (items?.at(-1)).Type();"},
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("negative-index", 1, 34, 1, 50)},
			},
			{
				Code: `import.meta.l\u0061st(items);`, FileName: "file.js",
				Options: map[string]any{"getLastElementFunctions": []any{"\uFEFF import.meta.last \u00A0"}},
				Output:  []string{"items.at(-1);"},
				Errors:  []rule_tester.InvalidTestCaseError{upstreamError("get-last-function", 1, 1, 1, 22, "import.meta.last")},
			},
			{
				// A plain object can provide numeric keys and a length without an at method.
				Code: "const items = { 0: 'first', length: 1 }; items[items.length - 1]; _.last(items);", FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamError("negative-index", 1, 48, 1, 64),
					upstreamError("get-last-function", 1, 67, 1, 73, "_.last"),
				},
			},
			{
				Code: "const items = { slice() { return ['first']; } }; items.slice(-1)[0];", FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("slice", 1, 56, 1, 61)},
			},
			{
				// The upstream fix produces an invalid bare decimal member access.
				Code: "1[1..length - 1];", FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{upstreamError("negative-index", 1, 3, 1, 16)},
			},
		},
	)
}

func TestPreferAtTypeIntersections(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "function f(a: NodeList & unknown) { return a[a.length - 1]; }", FileName: "file.ts"},
		{Code: "function f<T extends any>(a: NodeList & T) { return a[a.length - 1]; }", FileName: "file.ts"},
		{Code: "type Nodes = NodeList & { children: Nodes }; function f(a: Nodes) { return a[a.length - 1]; }", FileName: "file.ts"},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, declaration := range []string{
		"function f(a: NodeList & any)",
		"function f(a: any & NodeList)",
		"type Alias = any; function f(a: NodeList & (unknown | Alias))",
		"type Cycle = Cycle | Cycle; function f(a: Cycle)",
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:     declaration + " {\n  return a[a.length - 1];\n}",
			FileName: "file.ts",
			Output:   []string{declaration + " {\n  return a.at(-1);\n}"},
			Errors:   []rule_tester.InvalidTestCaseError{upstreamError("negative-index", 2, 12, 2, 24)},
		})
	}
	for _, mode := range []string{"typed", "source-only"} {
		t.Run(mode, func(t *testing.T) {
			testRule := PreferAtRule
			if mode == "source-only" {
				testRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
					ctx.TypeChecker = nil
					return PreferAtRule.Run(ctx, options)
				}
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testRule, valid, invalid)
		})
	}
}

func TestPreferAtSharedTypeAliases(t *testing.T) {
	const depth = 32
	var source strings.Builder
	source.WriteString("type T0 = NodeList;\n")
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&source, "type T%d = T%d | T%d;\n", i, i-1, i-1)
	}
	fmt.Fprintf(&source, "function f(a: T%d) { return a[a.length - 1] ?? _.last(a); }", depth)
	testRule := PreferAtRule
	testRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return PreferAtRule.Run(ctx, options)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testRule,
		[]rule_tester.ValidTestCase{{Code: source.String(), FileName: "file.ts"}}, nil,
	)
}

func TestPreferAtEditDemand(t *testing.T) {
	const source = "items[items.length - 1]; text.charAt(text.length - 1);"
	program, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     PreferAtRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferAtRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	all := run(rule.EditDemandAll)
	if len(all[0].Fixes()) == 0 || all[0].Suggestions != nil || all[1].FixesPtr != nil ||
		all[1].Suggestions == nil || len(*all[1].Suggestions) != 1 {
		t.Fatalf("expected one autofix and one suggestion: %+v", all)
	}
	suggestion := (*all[1].Suggestions)[0]
	if suggestion.Message.Id != "use-at" || suggestion.Message.Description != "Use `.at(…)`." {
		t.Errorf("unexpected suggestion message: %+v", suggestion.Message)
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		for i, diagnostic := range run(demand) {
			expected := all[i]
			if demand != rule.EditDemandAutofix {
				expected.FixesPtr = nil
			}
			if demand != rule.EditDemandSuggestion {
				expected.Suggestions = nil
			}
			if diagnostic.Range != expected.Range || !reflect.DeepEqual(diagnostic.Message, expected.Message) ||
				!reflect.DeepEqual(diagnostic.FixesPtr, expected.FixesPtr) || !reflect.DeepEqual(diagnostic.Suggestions, expected.Suggestions) {
				t.Errorf("demand %d changed diagnostic %d: %+v, want %+v", demand, i, diagnostic, expected)
			}
		}
	}
}
