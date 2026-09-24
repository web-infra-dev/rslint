package no_empty_named_blocks_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_empty_named_blocks"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyNamedBlocksExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `import { type Named } from 'mod';`},
		{Code: `import { 'named-export' as named } from 'mod';`},
		{Code: `import Default, * as Namespace from 'mod';`},
		{Code: `export {}; export {} from 'mod';`},
		{Code: `import('mod', { with: {} }); type Module = import('mod'); import mod = require('mod');`},
		{Code: `import Default from 'mod'; const view = <Default options={{}} />;`, Tsx: true},
		{Code: `/** @import {} from 'mod' */`, FileName: "jsdoc.js", TSConfig: "tsconfig.allow-js.json"},
		// Empty import attributes are not named import blocks. Upstream's
		// token search incorrectly reports these and can remove their braces.
		{Code: `import Default from 'mod' with {};`},
		{Code: `import * as Namespace from 'mod' with {};`},
		{Code: `import 'mod' with {};`},
		{Code: `import { Named } from 'mod' with {};`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, tc := range []struct{ code, output string }{
		{`import Default, {/* keep */} from 'mod';`, `import Default from 'mod';`},
		{`import Default /* keep */, /* removed */ {} /* after */ from 'mod';`, `import Default /* keep */ /* after */ from 'mod';`},
		// Keep the default binding separate from `from` in compact syntax.
		{`import Default,{}from'mod';`, `import Default from'mod';`},
		{`import \u0041,{}from'mod';`, `import \u0041 from'mod';`},
		{`import Default/**/,{}from'mod';`, `import Default/**/from'mod';`},
		{`import Default,{}/* keep */from'mod';`, `import Default/* keep */from'mod';`},
		{`import type Default, { /* empty */ } from 'mod';`, `import type Default from 'mod';`},
		// Contextual words are still default bindings. Upstream misclassifies
		// these as having no binding and offers broken side-effect suggestions.
		{`import type, {} from 'mod';`, `import type from 'mod';`},
		{`import from, {} from 'mod';`, `import from from 'mod';`},
		{`import Default, {} from 'mod' with {};`, `import Default from 'mod' with {};`},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: tc.code, Output: []string{tc.output}, Errors: emptyBlockError(1, 1, 1, len(tc.code)+1),
		})
	}
	for _, tc := range []struct{ code, output string }{
		{`import {/* keep */} from 'mod';`, `import 'mod';`},
		{`import/* keep */{}from'mod';`, `import/* keep */ 'mod';`},
		{`import{ }from'mod';`, `import'mod';`},
		{`import{}from 'mod';`, `import'mod';`},
		{`import {} from  'mod';`, `import  'mod';`},
		{`import {} from /* keep */'mod';`, `import /* keep */'mod';`},
		// Upstream removes the slash here because it finds whitespace later
		// in the gap. Only consume whitespace immediately after `from`.
		{`import {} from/* keep */ 'mod';`, `import /* keep */ 'mod';`},
		{`import {} from 'mod' with {};`, `import 'mod' with {};`},
		{`import {} from 'mod' with { type: 'json' };`, `import 'mod' with { type: 'json' };`},
		// Attribute keys must not count as imported bindings.
		{`import {} from 'mod' with { mode: 'custom' };`, `import 'mod' with { mode: 'custom' };`},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: tc.code, Errors: emptyBlockError(1, 1, 1, len(tc.code)+1, "", tc.output),
		})
	}
	invalid = append(invalid,
		rule_tester.InvalidTestCase{
			Code:   "import {} from\u00a0'模块';",
			Errors: emptyBlockError(1, 1, 1, 21, "", "import '模块';"),
		},
		rule_tester.InvalidTestCase{
			Code:   "/*😀*/ import 值, {} from '模块';",
			Output: []string{"/*😀*/ import 值 from '模块';"}, Errors: emptyBlockError(1, 8, 1, 31),
		},
		rule_tester.InvalidTestCase{
			Code:   "// leading\nimport {\n  // empty\n} from 'mod'; // trailing",
			Errors: emptyBlockError(2, 1, 4, 14, "// leading\n // trailing", "// leading\nimport 'mod'; // trailing"),
		},
		rule_tester.InvalidTestCase{
			Code:   "import\r\n{}\r\nfrom\r\n'mod';",
			Errors: emptyBlockError(1, 1, 4, 7, "", "import\r\n\n'mod';"),
		},
		rule_tester.InvalidTestCase{
			Code:   "import {} from\u2028'mod';",
			Errors: emptyBlockError(1, 1, 2, 7, "", "import 'mod';"),
		},
		rule_tester.InvalidTestCase{
			// Keep the BOM in edits; source columns exclude it.
			Code:   "\ufeff" + `import{}from'mod';`,
			Errors: emptyBlockError(1, 1, 1, 19, "\ufeff", "\ufeff"+`import 'mod';`),
		},
		rule_tester.InvalidTestCase{
			Code:   "import{}from// keep\n'mod';",
			Errors: emptyBlockError(1, 1, 2, 7, "", "import// keep\n'mod';"),
		},
		// The suggestion must use this declaration's `from`, not the first
		// matching token in the file as upstream does.
		rule_tester.InvalidTestCase{
			Code:   "import value from 'first';\nimport {} from 'second';",
			Errors: emptyBlockError(2, 1, 2, 25, "import value from 'first';\n", "import value from 'first';\nimport 'second';"),
		},
		rule_tester.InvalidTestCase{
			Code: "import {} from 'first';\nimport {} from 'second';",
			Errors: append(
				emptyBlockError(1, 1, 1, 24, "\nimport {} from 'second';", "import 'first';\nimport {} from 'second';"),
				emptyBlockError(2, 1, 2, 25, "import {} from 'first';\n", "import {} from 'first';\nimport 'second';")...,
			),
		},
		rule_tester.InvalidTestCase{
			Code: `import {} from 'mod';`, FileName: "empty.js", TSConfig: "tsconfig.allow-js.json",
			Errors: emptyBlockError(1, 1, 1, 22, "", `import 'mod';`),
		},
	)
	// A runtime import cannot retain the type-only resolution-mode attribute.
	for _, code := range []string{
		`import type {} from 'mod' with { 'resolution-mode': 'import' };`,
		`import type {} from 'mod' with { 'resolution-mode': 'require' };`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: code, Errors: emptyBlockError(1, 1, 1, len(code)+1, ""),
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_empty_named_blocks.NoEmptyNamedBlocksRule, valid, invalid)
}

func TestNoEmptyNamedBlocksSchema(t *testing.T) {
	for _, options := range [][]any{{true}, {"always"}, {map[string]any{}}} {
		if err := no_empty_named_blocks.NoEmptyNamedBlocksRule.Schema.Validate(options); err == nil {
			t.Errorf("accepted options for a rule without options: %#v", options)
		}
	}
}

func TestNoEmptyNamedBlocksEditDemand(t *testing.T) {
	for _, tc := range []struct {
		code        string
		autofix     bool
		suggestions []string
	}{
		{code: `import Default, {} from 'mod';`, autofix: true},
		{code: `import {} from 'mod';`, suggestions: []string{"", `import 'mod';`}},
		{code: `import type{}from'mod';`, suggestions: []string{"", `import 'mod';`}},
		{code: `import type {} from 'mod' with { 'resolution-mode': 'require' };`, suggestions: []string{""}},
	} {
		t.Run(tc.code, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", Path: "/test.ts"}, tc.code, core.ScriptKindTS)
			var all rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
				var diagnostics []rule.RuleDiagnostic
				ctx := (rule.RuleContext{SourceFile: file}).WithDiagnosticConsumer(no_empty_named_blocks.NoEmptyNamedBlocksRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Demand: demand,
					Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
				})
				no_empty_named_blocks.NoEmptyNamedBlocksRule.Run(ctx, nil)[ast.KindImportDeclaration](file.Statements.Nodes[0])
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
				}
				got := diagnostics[0]
				if demand == rule.EditDemandAll {
					all = got
				}
				wantFix := tc.autofix && (demand == rule.EditDemandAll || demand == rule.EditDemandAutofix)
				wantSuggestions := !tc.autofix && (demand == rule.EditDemandAll || demand == rule.EditDemandSuggestion)
				if (got.FixesPtr != nil) != wantFix || (got.Suggestions != nil) != wantSuggestions {
					t.Fatalf("demand %d: incorrect edit availability", demand)
				}
				if wantFix && !reflect.DeepEqual(got.FixesPtr, all.FixesPtr) ||
					wantSuggestions && !reflect.DeepEqual(got.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: edits differ from all edits", demand)
				}
				if wantSuggestions {
					if len(*got.Suggestions) != len(tc.suggestions) {
						t.Fatalf("got %d suggestions", len(*got.Suggestions))
					}
					for index, wantOutput := range tc.suggestions {
						description := []string{"Remove unused import", "Remove empty import block"}[index]
						suggestion := (*got.Suggestions)[index]
						if suggestion.Message.Id != "" || suggestion.Message.Description != description {
							t.Fatalf("incorrect suggestion message: %#v", suggestion.Message)
						}
						output, _, _ := linter.ApplyRuleFixes(tc.code, []rule.RuleSuggestion{suggestion})
						if output != wantOutput {
							t.Fatalf("suggestion output %q, want %q", output, wantOutput)
						}
					}
				}
				want := all
				got.FixesPtr, want.FixesPtr = nil, nil
				got.Suggestions, want.Suggestions = nil, nil
				if !reflect.DeepEqual(got, want) || got.Message.Description != emptyBlockMessage ||
					got.Message.Id != "" || got.Range != core.NewTextRange(0, len(tc.code)) {
					t.Fatalf("demand %d: incorrect diagnostic identity", demand)
				}
			}
		})
	}
}
