package no_anonymous_default_export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_anonymous_default_export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expected diagnostics and ranges were checked against eslint-plugin-import v2.32.0.
func TestNoAnonymousDefaultExportExtras(t *testing.T) {
	defaults := map[string]any{
		"allowArray":             false,
		"allowArrowFunction":     false,
		"allowAnonymousClass":    false,
		"allowAnonymousFunction": false,
		"allowCallExpression":    true,
		"allowNew":               false,
		"allowLiteral":           false,
		"allowObject":            false,
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &no_anonymous_default_export.NoAnonymousDefaultExportRule,
		[]rule_tester.ValidTestCase{
			// Empty options and explicit runtime defaults.
			{
				Code:    "export default factory();",
				Options: map[string]any{},
			},
			{
				Code:    "export default factory();",
				Options: defaults,
			},
			// Declarations outside default exports and named declarations remain allowed.
			{
				Code: "export function fn() {}\nexport class Named {}\nconst fnExpr = function() {}; const cls = class {}; const obj = {}; export { obj };",
			},
			{
				Code: "export default async function* named() {}",
			},
			{
				Code: "export default class Named extends Base {}",
			},
			{
				Code: "export * as default from \"pkg\";",
			},
			// Parenthesized functions/classes are expressions, not declarations.
			{
				Code: "export default (function() {});",
			},
			{
				Code: "export default (class {});",
			},
			// Other expression types are not literals or ordinary calls.
			{
				Code:    "export default -1;",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default +1;",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default undefined;",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default void 0;",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default 1 + 2;",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default (first, second);",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default flag ? [] : {};",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default tag`value`;",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default await factory();",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default obj.value;",
				Options: map[string]any{"allowCallExpression": false},
			},
			// Authored TypeScript wrappers and bodyless declarations remain visible to ESTree.
			{
				Code:    "export default ({} as object);",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default ({} satisfies object);",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default ([])!;",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default <number>123;",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default (factory() as unknown);",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code: "export = {};",
			},
			{
				Code: "export default interface Named { value: string }",
			},
			{
				Code: "export default function(): void;",
			},
			{
				Code: `declare module "pkg" { export default function(): void; }`,
			},
			{
				Code:    `export default factory<string>;`,
				Options: map[string]any{"allowCallExpression": false},
			},
			// Optional chains and dynamic imports are distinct ESTree node types.
			{
				Code:    "export default factory?.();",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default obj?.method();",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default obj.method?.();",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default (obj?.method());",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code:    "export default import(\"pkg\");",
				Options: map[string]any{"allowCallExpression": false},
			},
			{
				Code: "export default <Widget />;",
				Tsx:  true,
			},
			// Additional literal kinds and interpolated templates, both forbidden and allowed.
			{
				Code:    "export default true;",
				Options: map[string]any{"allowLiteral": true},
			},
			{
				Code:    "export default false;",
				Options: map[string]any{"allowLiteral": true},
			},
			{
				Code:    "export default null;",
				Options: map[string]any{"allowLiteral": true},
			},
			{
				Code:    "export default 123n;",
				Options: map[string]any{"allowLiteral": true},
			},
			{
				Code:    "export default /value/u;",
				Options: map[string]any{"allowLiteral": true},
			},
			{
				Code:    "export default `hello ${name}`;",
				Options: map[string]any{"allowLiteral": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ast.IsImportCall also matches import.defer(), unlike the pinned rule.
			{
				Code:    `export default import.defer("pkg");`,
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code:    `export default import.source("pkg");`,
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			// Default exports can occur inside ambient modules, not just at file scope.
			{
				Code: `declare module "pkg" { export default class {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code: `@decorate /* comment */ export default class {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 1, Column: 25, EndLine: 1, EndColumn: 48},
				},
			},
			// Parentheses, including JSDoc casts, can end an optional chain.
			{
				Code:    `export default (factory?.()).value();`,
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:     `export default (((/** @type {*} */ (factory?.method)))());`,
				FileName: "case.js",
				Options:  map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 59},
				},
			},
			// Preserve positions after a BOM, hashbang, CRLF, and astral characters.
			{
				Code:     "\ufeff#!/usr/bin/env node\r\nexport default [\"😀\"];\r\n",
				FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrayMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 23},
				},
			},
			// Empty options and explicit runtime defaults.
			{
				Code:    "export default [];",
				Options: map[string]any{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrayMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export default []",
				Options: defaults,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrayMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "export default () => {}",
				Options: defaults,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrowMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export default class {}",
				Options: defaults,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export default function() {}",
				Options: defaults,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: functionMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:    "export default 123",
				Options: defaults,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export default {}",
				Options: defaults,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: objectMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "export default new Foo()",
				Options: defaults,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: newMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// Parentheses are transparent for the expression kinds checked by the rule.
			{
				Code:    "export default ([]);",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrayMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "export default (() => {});",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrowMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    "export default ({});",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: objectMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "export default (123);",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    "export default (new Foo());",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: newMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "export default (factory());",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			// Additional literal kinds and interpolated templates, both forbidden and allowed.
			{
				Code: "export default true;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "export default false;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: "export default null;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "export default 123n;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "export default /value/u;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: "export default `hello ${name}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			// An outer ordinary call can end an optional chain; computed and typed calls still count.
			{
				Code:    "export default (obj?.method)();",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    "export default (factory?.())();",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    "export default obj[\"method\"]();",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    "export default factory<string>();",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    "export default import.meta.resolve(\"pkg\");",
				Options: map[string]any{"allowCallExpression": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: callMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
				},
			},
			// Additional anonymous declaration forms and JSX bodies.
			{
				Code: "export default new Foo;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: newMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "export default async function*() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: functionMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: "export default function<T>(value: T) { return value; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: functionMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 55},
				},
			},
			{
				Code: "export default abstract class {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code: "export default class extends Base { #value = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 50},
				},
			},
			{
				Code: "export default async () => <Widget />;",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: arrowMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
				},
			},
			// A decorator before export is outside the reported ExportDefaultDeclaration.
			{
				Code: "@decorator\nexport default class {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
				},
			},
			{
				Code: "export default @decorator class {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			// Complete export ranges include semicolons and use UTF-16 columns across lines.
			{
				Code: "'😀'; /* lead */ export default {\n  value: '中😀'\n}; // trailing",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: objectMessage, Line: 1, Column: 18, EndLine: 3, EndColumn: 3},
				},
			},
			{
				Code: "// leading\nexport /* comment */ default function() {\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: functionMessage, Line: 2, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code: "'😀'; export default class {\n  method() {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: classMessage, Line: 1, Column: 7, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code: "export default \"😀\";",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: literalMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// JavaScript JSDoc casts are transparent, like source parentheses.
			{
				Code:     "export default /** @type {object} */ ({});",
				FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: objectMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
				},
			},
			{
				Code:     "export default /** @satisfies {object} */ ({});",
				FileName: "case.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: objectMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
				},
			},
		},
	)
}
