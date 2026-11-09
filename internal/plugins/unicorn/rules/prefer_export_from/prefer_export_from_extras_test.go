package prefer_export_from_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_export_from"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferExportFromExtras(t *testing.T) {
	root := fixtures.GetRootDir()
	const verbatimConfig = "tsconfig.verbatim.json"
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, verbatimConfig): `{"compilerOptions":{"module":"esnext","verbatimModuleSyntax":true}}`,
	})
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_export_from.PreferExportFromRule,
		[]rule_tester.ValidTestCase{
			{Code: `import {value} from 'm';
export {value as forwarded} from 'other';
export const cast = value as unknown;
export const nonNull = value!;
export const optional = value?.field;
export const computed = value['field'];
export const a = value, b = value;`},
			{Code: "import {value} from 'm';\nexport = value;"},
			{Code: "import {value} from 'm';\nexport const alias = value;\nexport {alias as another};"},
			{Code: "import {foo} from 'm';\n/***/\nexport const alias = foo;", FileName: "case.js"},
			{
				Code: "import Widget from 'm';\nconst element = <Widget />;\nexport {Widget};",
				Tsx:  true, Options: map[string]any{"checkUsedVariables": false},
			},
			{
				// The exported alias itself is the JSX factory.
				Code: "import {factory} from 'm';\nexport const React = factory;\nconst element = <div />;",
				Tsx:  true,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				// The shadowed parameter does not keep the import alive.
				Code:    "import {value} from 'm';\nfunction read(value) { return value; }\nexport {value};",
				Options: map[string]any{"checkUsedVariables": false},
				Output:  []string{"\nfunction read(value) { return value; }\n\nexport {value} from 'm';"},
				Errors:  []rule_tester.InvalidTestCaseError{reexportError("value", 3, 9, 3, 14)},
			},
			{
				Code:    "import {foo, bar} from 'm';\nuse(foo);\nexport {foo};",
				Options: map[string]any{},
				Output:  []string{"import {foo, bar} from 'm';\nuse(foo);\n\nexport {foo} from 'm';"},
				Errors:  []rule_tester.InvalidTestCaseError{reexportError("foo", 3, 9, 3, 12)},
			},
			{
				Code:    "import React from 'react';\nconst element = <div />;\nexport {React};",
				Tsx:     true,
				Options: map[string]any{"checkUsedVariables": true},
				Output:  []string{"import React from 'react';\nconst element = <div />;\n\nexport {default as React} from 'react';"},
				Errors:  []rule_tester.InvalidTestCaseError{reexportError("React", 3, 9, 3, 14)},
			},
			{
				Code: `import React from 'react';
function render(React: unknown) { return <div />; }
const element = <div />;
export {React};`,
				Tsx: true,
				Output: []string{`import React from 'react';
function render(React: unknown) { return <div />; }
const element = <div />;

export {default as React} from 'react';`},
				Errors: []rule_tester.InvalidTestCaseError{reexportError("React", 4, 9, 4, 14)},
			},
			{
				Code: `/** @jsx h */
/** @jsxFrag Fragment */
import {h, Fragment} from 'm';
const element = <>hello</>;
export {h, Fragment};`,
				FileName: "case.jsx",
				Output: []string{`/** @jsx h */
/** @jsxFrag Fragment */
import {h, Fragment} from 'm';
const element = <>hello</>;
export { Fragment};
export {h} from 'm';`, `/** @jsx h */
/** @jsxFrag Fragment */
import {h, Fragment} from 'm';
const element = <>hello</>;

export {h, Fragment} from 'm';`},
				Errors: []rule_tester.InvalidTestCaseError{reexportError("h", 5, 9, 5, 10), reexportError("Fragment", 5, 12, 5, 20)},
			},
			{
				Code:   "import type {Foo} from 'm';\ntype Current = Foo;\nexport type {Foo};",
				Output: []string{"import type {Foo} from 'm';\ntype Current = Foo;\n\nexport {type Foo} from 'm';"},
				Errors: []rule_tester.InvalidTestCaseError{reexportError("Foo", 3, 14, 3, 17)},
			},
			{
				Code:     "import {Foo} from 'm';\n/** @type {typeof Foo} */\nlet current;\nexport {Foo};",
				FileName: "case.js",
				Output:   []string{"import {Foo} from 'm';\n/** @type {typeof Foo} */\nlet current;\n\nexport {Foo} from 'm';"},
				Errors:   []rule_tester.InvalidTestCaseError{reexportError("Foo", 4, 9, 4, 12)},
			},
			{
				Code:     "import {type Foo} from 'm';\nexport type {Foo};\nexport type {Other} from 'm';",
				TSConfig: verbatimConfig,
				Output:   []string{"import {type Foo} from 'm';\n\nexport type {Other, Foo} from 'm';"},
				Errors:   []rule_tester.InvalidTestCaseError{reexportError("Foo", 2, 14, 2, 17)},
			},
			{
				Code:   "import {value} from 'm';\nexport {\n  value\n    as \"😀 alias\",\n};",
				Output: []string{"\n\nexport {value as \"😀 alias\"} from 'm';"},
				Errors: []rule_tester.InvalidTestCaseError{reexportError(`"😀 alias"`, 3, 3, 4, 18)},
			},
			{
				Code:   "import * as ns from 'm';\nexport {ns as \"default\"};",
				Output: []string{"\n\nexport * as \"default\" from 'm';"},
				Errors: []rule_tester.InvalidTestCaseError{reexportError(`"default"`, 2, 9, 2, 24)},
			},
			{
				Code:   "import {value} from 'm';\u2028export {value};",
				Output: []string{"\u2028\u2028export {value} from 'm';"},
				Errors: []rule_tester.InvalidTestCaseError{reexportError("value", 2, 9, 2, 14)},
			},
			{
				// Keep a syntactically valid default import after removing the namespace.
				Code:   "import main, * as ns from 'm';\nexport {ns};\nuse(main);",
				Output: []string{"import main from 'm';\n\nuse(main);\nexport * as ns from 'm';"},
				Errors: []rule_tester.InvalidTestCaseError{reexportError("ns", 2, 9, 2, 11)},
			},
			{
				// Equal attribute sets can be merged in either source order.
				Code: `import {foo} from 'm' with { type: 'unknown', mode: 'a' };
export {foo};
export {bar} from 'm' with { mode: 'a', type: 'unknown' };`,
				Output: []string{"\n\nexport {bar, foo} from 'm' with { mode: 'a', type: 'unknown' };"},
				Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 2, 9, 2, 12)},
			},
			{
				// Moving or merging this request would load the consumer first.
				Code: `import {value} from './setup.js';
import './consumer.js';
export {value};
export {other} from './setup.js';`,
				Output: []string{`export {value} from './setup.js';
import './consumer.js';

export {other} from './setup.js';`},
				Errors: []rule_tester.InvalidTestCaseError{reexportError("value", 3, 9, 3, 14)},
			},
			{
				Code:   "import {value} from 'm';\n(eval)('console.log(value)');\nexport {value};",
				Errors: []rule_tester.InvalidTestCaseError{reexportError("value", 3, 9, 3, 14)},
			},
			{
				// Optional eval is indirect and cannot access local bindings.
				Code:   "import {value} from 'm';\neval?.('0');\nexport {value};",
				Output: []string{"\neval?.('0');\n\nexport {value} from 'm';"},
				Errors: []rule_tester.InvalidTestCaseError{reexportError("value", 3, 9, 3, 14)},
			},
			{
				// Moving this re-export to EOF would take it out of the module.
				Code:   `declare module "ambient" { import {foo} from "m"; export {foo}; }`,
				Errors: []rule_tester.InvalidTestCaseError{reexportError("foo", 1, 59, 1, 62)},
			},
			{
				// A direct export would evaluate the dependency eagerly.
				Code:   "import defer * as ns from 'm';\nexport {ns};",
				Errors: []rule_tester.InvalidTestCaseError{reexportError("ns", 2, 9, 2, 11)},
			},
		})
}

func TestPreferExportFromEditDemand(t *testing.T) {
	for _, test := range []struct {
		name, code, output string
		options            any
		edit               rule.EditDemand
	}{
		{
			name:   "autofix",
			code:   "import {value} from 'm';\nexport {value};",
			output: "\n\nexport {value} from 'm';",
			edit:   rule.EditDemandAutofix,
		},
		{
			name:    "suggestion",
			code:    "import {unused, value} from 'm';\nexport {value};",
			output:  "import {unused, } from 'm';\n\nexport {value} from 'm';",
			options: map[string]any{"checkUsedVariables": false},
			edit:    rule.EditDemandSuggestion,
		},
		{
			name: "comment",
			code: "import {value} from 'm';\nexport {/* keep */ value};",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			compilerProgram, sourceFile, err := helper.CreateTestProgram(test.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			lintRule := &prefer_export_from.PreferExportFromRule
			options := rule_tester.ResolveTestCaseOptions(t, lintRule, test.options)
			var baseline rule.RuleDiagnostic
			var selectedEdits []rule.RuleFix
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: program.NewFromCompiler(compilerProgram), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name: lintRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners { return lintRule.Run(ctx, options) },
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
				}
				diagnostic := diagnostics[0]
				wantFix := demand&test.edit&rule.EditDemandAutofix != 0
				wantSuggestion := demand&test.edit&rule.EditDemandSuggestion != 0
				if (diagnostic.FixesPtr != nil) != wantFix || (diagnostic.Suggestions != nil) != wantSuggestion {
					t.Fatalf("demand %d: unexpected edits: %+v", demand, diagnostic)
				}
				var edits []rule.RuleFix
				if wantFix {
					edits = diagnostic.Fixes()
				}
				if wantSuggestion {
					suggestions := *diagnostic.Suggestions
					if len(suggestions) != 1 || suggestions[0].Message.Id != "suggestion" ||
						suggestions[0].Message.Description != "Switch to `export…from`." {
						t.Fatalf("demand %d: unexpected suggestions: %+v", demand, suggestions)
					}
					edits = suggestions[0].Fixes()
				}
				if wantFix || wantSuggestion {
					if demand == rule.EditDemandAll && !reflect.DeepEqual(edits, selectedEdits) {
						t.Fatal("requesting all edits changed the result")
					}
					selectedEdits = append([]rule.RuleFix(nil), edits...)
					output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleSuggestion{{FixesArr: edits}})
					if !fixed || output != test.output {
						t.Fatalf("demand %d: got %q, want %q", demand, output, test.output)
					}
				}
				diagnostic.FixesPtr = nil
				diagnostic.Suggestions = nil
				if demand == rule.EditDemandNone {
					baseline = diagnostic
				} else if !reflect.DeepEqual(diagnostic, baseline) {
					t.Fatalf("demand %d changed the diagnostic", demand)
				}
			}
		})
	}
}
