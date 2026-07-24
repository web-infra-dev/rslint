package triple_slash_reference

import (
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestTripleSlashReferenceRuleExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&TripleSlashReferenceRule,
		[]rule_tester.ValidTestCase{
			{
				// getCommentsBefore(Program) does not include comments after the
				// first syntax token.
				Code:    "const value = 1;\n/// <reference path=\"foo\" />\n",
				Options: map[string]interface{}{"path": "never"},
			},
			{
				Code: `
//// <reference path="four-slashes" />
/* /// <reference path="block" /> */
const stringValue = '/// <reference path="string" />';
const templateValue = ` + "`/// <reference path=\"template\" />`" + `;
const taggedValue = tag` + "`/// <reference path=\"tagged\" />`" + `;
`,
				Options: map[string]interface{}{"path": "never"},
			},
			{
				Code: `
/// <Reference path="uppercase-tag" />
/// <reference PATH="uppercase-kind" />
/// <reference other="first" path="not-first" />
`,
				Options: map[string]interface{}{"path": "never"},
			},
			{
				Code: `
/// <reference types="foo" />
import foo = ns.foo;
`,
				Options: map[string]interface{}{"types": "prefer-import"},
			},
			{
				Code: `
/// <reference types="foo" />
import('foo');
require('foo');
export * from 'foo';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
			},
			{
				Code: `
/// <reference types=" foo " />
import 'foo';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
			},
			{
				Code: `
/// <reference types="f\u006fo" />
import 'foo';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
			},
			{
				Code: `
/// <reference path="foo" />
/// <reference types="foo" />
/// <reference lib="foo" />
import 'foo';
`,
				Options: map[string]interface{}{
					"lib":   "always",
					"path":  "always",
					"types": "always",
				},
			},
			{
				Code: `
// eslint-disable-next-line
/// <reference path="disabled" />
`,
				Options: map[string]interface{}{"path": "never"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				// Defaults are lib=always, path=never, types=prefer-import.
				Code: `
/// <reference path="path-default" />
/// <reference lib="lib-default" />
/// <reference types="same" />
/// <reference types="different" />
import 'same';
`,
				Errors: []rule_tester.InvalidTestCaseError{
					tripleSlashReferenceError(2, 1, "path-default"),
					tripleSlashReferenceError(4, 1, "same"),
				},
			},
			{
				// A partial object is merged with the upstream defaults.
				Code: `
/// <reference path="path-default" />
/// <reference lib="lib-never" />
/// <reference types="types-default" />
import 'types-default';
`,
				Options: map[string]interface{}{"lib": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					tripleSlashReferenceError(2, 1, "path-default"),
					tripleSlashReferenceError(3, 1, "lib-never"),
					tripleSlashReferenceError(4, 1, "types-default"),
				},
			},
			{
				Code: `
/* license */
// ordinary leading comment
/// <reference path="foo" />
const value = 1;
`,
				Options: map[string]interface{}{"path": "never"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(4, 1, "foo")},
			},
			{
				Code:    "\ufeff/// <reference path=\"bom\" />\n",
				Options: map[string]interface{}{"path": "never"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(1, 2, "bom")},
			},
			{
				Code:    "#!/usr/bin/env node\n/// <reference path=\"shebang\" />\n",
				Options: map[string]interface{}{"path": "never"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(2, 1, "shebang")},
			},
			{
				Code: "\t/// <reference path=\"foo\" />\r\nconst value = 1;\r\n",
				Options: map[string]interface{}{
					"path": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "tripleSlashReference",
					Message:   "Do not use a triple slash reference for foo, use `import` style instead.",
					Line:      1,
					Column:    2,
					EndLine:   1,
					EndColumn: 30,
				}},
			},
			{
				Code: "///\v<reference\u00a0path\u000b=\u000c\"unicode-space\"\n",
				Options: map[string]interface{}{
					"path": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					tripleSlashReferenceError(1, 1, "unicode-space"),
				},
			},
			{
				// The upstream capture is greedy and the regexp is not anchored
				// at the end of the comment.
				Code:    `/// <reference path="foo" types="bar" /> trailing`,
				Options: map[string]interface{}{"path": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					tripleSlashReferenceError(1, 1, `foo" types="bar`),
				},
			},
			{
				Code: `
/// <reference types="foo" />
import type { Foo } from 'foo';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(2, 1, "foo")},
			},
			{
				Code: `
/// <reference types="foo" />
import 'foo';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(2, 1, "foo")},
			},
			{
				Code: `
/// <reference types=" foo " />
import ' foo ';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(2, 1, " foo ")},
			},
			{
				Code: `
/// <reference types="foo" />
import 'f\u006fo';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
				Errors:  []rule_tester.InvalidTestCaseError{tripleSlashReferenceError(2, 1, "foo")},
			},
			{
				Code: `
/// <reference types="foo" />
/// <reference types="foo" />
import first from 'foo';
import second from 'foo';
`,
				Options: map[string]interface{}{"types": "prefer-import"},
				Errors: []rule_tester.InvalidTestCaseError{
					tripleSlashReferenceError(2, 1, "foo"),
					tripleSlashReferenceError(3, 1, "foo"),
					tripleSlashReferenceError(2, 1, "foo"),
					tripleSlashReferenceError(3, 1, "foo"),
				},
			},
		},
	)
}

func TestTripleSlashReferenceMatcherParity(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantModules []string
	}{
		{
			name:        "zero whitespace and no closing tag",
			source:      `/// <reference` + `types="zero-space"`,
			wantModules: []string{"zero-space"},
		},
		{
			name:        "single quotes",
			source:      `/// <reference path='single-quote'`,
			wantModules: []string{"single-quote"},
		},
		{
			name:        "pipe quotes",
			source:      `/// <reference path=|pipe-quote|`,
			wantModules: []string{"pipe-quote"},
		},
		{
			name:        "mismatched quotes",
			source:      `/// <reference path="mismatched'`,
			wantModules: []string{"mismatched"},
		},
		{
			name:        "empty module",
			source:      `/// <reference path=""`,
			wantModules: []string{""},
		},
		{
			name:        "greedy module capture",
			source:      `/// <reference path="foo" types="bar" />`,
			wantModules: []string{`foo" types="bar`},
		},
		{
			name:        "ECMAScript whitespace",
			source:      "///\v<reference\u00a0path\u000b=\u000c\"unicode-space\"",
			wantModules: []string{"unicode-space"},
		},
		{
			name:   "unquoted module",
			source: `/// <reference path=unquoted />`,
		},
		{
			name:   "missing closing quote",
			source: `/// <reference path="unterminated />`,
		},
		{
			name:   "case sensitive tag",
			source: `/// <Reference path="uppercase" />`,
		},
		{
			name:   "case sensitive kind",
			source: `/// <reference PATH="uppercase" />`,
		},
		{
			name:   "kind is not first attribute",
			source: `/// <reference other="first" path="not-first" />`,
		},
	}

	options := []any{map[string]interface{}{
		"lib":   "never",
		"path":  "never",
		"types": "never",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diagnostics := runTripleSlashReferenceDirect(test.source, options)
			modules := make([]string, 0, len(diagnostics))
			for _, diagnostic := range diagnostics {
				modules = append(modules, diagnostic.Message.Data["module"])
			}
			if !slices.Equal(modules, test.wantModules) {
				t.Errorf("reported modules = %q, want %q", modules, test.wantModules)
			}
		})
	}
}

func TestTripleSlashReferenceDiagnosticDataAndRange(t *testing.T) {
	source := `/// <reference path="foo" />`
	diagnostics := runTripleSlashReferenceDirect(
		source,
		[]any{map[string]interface{}{"path": "never"}},
	)

	if len(diagnostics) != 1 {
		t.Fatalf("got %d diagnostics, want 1", len(diagnostics))
	}
	diagnostic := diagnostics[0]
	if diagnostic.Range.Pos() != 0 || diagnostic.Range.End() != len(source) {
		t.Errorf("diagnostic range = [%d,%d), want [0,%d)", diagnostic.Range.Pos(), diagnostic.Range.End(), len(source))
	}
	if diagnostic.Message.Description != "Do not use a triple slash reference for foo, use `import` style instead." {
		t.Errorf("diagnostic message = %q", diagnostic.Message.Description)
	}
	if diagnostic.Message.Data["module"] != "foo" {
		t.Errorf("diagnostic module data = %q, want %q", diagnostic.Message.Data["module"], "foo")
	}
}

func TestTripleSlashReferenceDisableDirectives(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		options []any
	}{
		{
			name: "immediate never report",
			source: "// eslint-disable-next-line @typescript-eslint/triple-slash-reference\n" +
				"/// <reference path=\"disabled\" />\n",
			options: []any{map[string]interface{}{"path": "never"}},
		},
		{
			name: "deferred prefer-import report",
			source: "// eslint-disable-next-line @typescript-eslint/triple-slash-reference\n" +
				"/// <reference types=\"disabled\" />\n" +
				"import 'disabled';\n",
			options: []any{map[string]interface{}{"types": "prefer-import"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if diagnostics := runTripleSlashReferenceDirect(test.source, test.options); len(diagnostics) != 0 {
				t.Errorf("got %d diagnostics, want none: %v", len(diagnostics), diagnostics)
			}
		})
	}
}

func runTripleSlashReferenceDirect(source string, options []any) []rule.RuleDiagnostic {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	diagnostics := make([]rule.RuleDiagnostic, 0, 1)
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithReporter(
		"@typescript-eslint/triple-slash-reference",
		rule.SeverityError,
		func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	)

	listeners := TripleSlashReferenceRule.Run(ctx, options)
	for _, statement := range sourceFile.Statements.Nodes {
		if listener := listeners[statement.Kind]; listener != nil {
			listener(statement)
		}
	}
	return diagnostics
}
