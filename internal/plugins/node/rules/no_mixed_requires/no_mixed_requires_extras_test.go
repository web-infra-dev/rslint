package no_mixed_requires_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expected diagnostics and ranges were checked against eslint-plugin-n v18.3.0.
func TestNoMixedRequiresExtras(t *testing.T) {
	runNoMixedRequiresTests(t,
		[]rule_tester.ValidTestCase{
			// A single declaration cannot mix declaration kinds or module groups.
			{Code: "const fs = require('fs').promises;", Options: []any{map[string]any{"grouping": true, "allowCall": true}}},
			// Match decoded identifiers and literal values, including line continuations.
			{Code: "const fs = r\\u0065quire('f\\\ns'), path = require('\\u0070ath');", Options: []any{true}},
			{Code: "const a = require(/** @type {string} */ ('fs')), b = require('path');", Options: []any{true}},
			// An authored TypeScript instantiation expression is not a bare callee.
			{Code: "const a = (require<string>)('fs'), b = 0;", FileName: "input.ts"},
			// Omitted and empty options use both false defaults.
			{Code: "const fs = require('fs'), helper = require('./helper');"},
			{Code: "const fs = require('fs'), helper = require('./helper');", Options: []any{map[string]any{}}},
			{Code: "const fs = require('fs'), helper = require('./helper');", Options: []any{map[string]any{"grouping": false, "allowCall": false}}},
			// Grouping only considers require initializers; mixed declarations take precedence.
			{Code: "var a, b = 0, c = load();", Options: []any{map[string]any{"grouping": true, "allowCall": true}}},
			{Code: "const a = require('fs'), b = require('path'), c = require('_http_agent'), d = require('sys');", Options: []any{true}},
			{Code: "const a = require('/root'), b = require('./b'), c = require('../c');", Options: []any{true}},
			{Code: "const a = require(''), b = require('pkg'), c = require('.../c'), d = require('..'), e = require('C:/d'), f = require('\\\\root');", Options: []any{true}},
			{Code: "const a = require(`fs`), b = require(/fs/), c = require(null), d = require(true), e = require(1n), f = require(...names);", Options: []any{true}},
			// Core names match the pinned list exactly, including decoded strings.
			{Code: "const a = require('\\x66s'), b = require(('path')), c = require('fs', 'extra');", Options: []any{true}},
			{Code: "const a = require('node:fs'), b = require('assert/strict'), c = require('test'), d = require('sqlite');", Options: []any{true}},
			// With allowCall, grouping uses the outer call argument rather than the loaded module.
			{Code: "const a = require('factory')('fs'), b = require('path');", Options: []any{map[string]any{"grouping": true, "allowCall": true}}},
			{Code: "const a = require('factory')('./a'), b = require('../b');", Options: []any{map[string]any{"grouping": true, "allowCall": true}}},
			{Code: "const a = require('factory')(), b = require();", Options: []any{map[string]any{"grouping": true, "allowCall": true}}},
			// Only direct require calls, members and permitted consecutive calls count.
			{Code: "var a = loader.require('fs'), b = new require('fs'), c = require`fs`, d = () => require('fs'), e = [require('fs')], f = ok ? require('fs') : null;"},
			{Code: "var a = require('fs').readFile(), b = 0;", Options: []any{map[string]any{"allowCall": true}}},
			{Code: "var a = (require('factory'))()().member, b = 0;"},
			{Code: "var a = (require)('fs'), b = ((require('path')))['join'];"},
			// Optional chains are ESTree ChainExpression nodes, including chains ended by parentheses.
			{Code: "var a = require?.('fs'), b = require('fs')?.readFile, c = (require?.('fs')).readFile, d = 0;"},
			{Code: "var a = (require?.('factory'))(), b = 0;", Options: []any{map[string]any{"allowCall": true}}},
			// Authored TypeScript wrappers remain visible, while type arguments do not hide a call.
			{Code: "const a = require('fs') as any, b = 0;", FileName: "input.ts"},
			{Code: "const a = <any>require('fs'), b = require('fs')!, c = (require as any)('fs'), d = require!('fs'), e = require('fs') satisfies any, f = 0;", FileName: "input.ts"},
		},
		[]rule_tester.InvalidTestCase{
			// Assertions are not string literals, even when their runtime value is a string.
			{Code: "const a = require(('fs' as string)), b = require('fs');", FileName: "input.ts", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 56)}},
			// Skipping an outer single declaration must not skip nested declarations.
			{Code: "const fn = () => { const a = require('fs'), b = 0; };", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 20, 1, 51)}},
			// Mixture takes precedence regardless of whether the ordinary variable comes first.
			{Code: "let a, b = require('fs'), c = require('pkg');", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 46)}},
			{Code: "const a = /** @satisfies {any} */ (require('fs')), b = 0;", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 58)}},
			{Code: "export /* before declare */ declare /* before const */ const a = require('fs'), b = 0;", FileName: "input.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 29, 1, 87)}},
			{Code: "let fs = require('fs'), filename;", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 34)}},
			{Code: "const fs = require('fs'), count = 0;", Options: []any{map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 37)}},
			{Code: "const fs = require('fs'), debug = require('debug')('app');", Options: []any{map[string]any{"allowCall": false}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 59)}},
			{Code: "var a = require('fs'), b = require('./b'), c = 0;", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 50)}},
			{Code: "const a = require('fs'), b = require('node:fs'), c = require('assert/strict');", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 79)}},
			{Code: "const a = require('fs'), b = require(`fs`);", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 44)}},
			{Code: "const a = require('fs')('package'), b = require('fs');", Options: []any{map[string]any{"grouping": true, "allowCall": true}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 55)}},
			{Code: "var a = (require('factory'))()().member, b = 0;", Options: []any{map[string]any{"allowCall": true}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 48)}},
			{Code: "const {readFile: read} = require('fs'), [first] = require('pkg'), rest = 0;", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 76)}},
			{Code: "var a = require('fs')[require('key')], b = 0;", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 46)}},
			{Code: "class C { #value; method() { const a = require('obj').#value, b = 0; } }", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 30, 1, 69)}},
			{Code: "var a = require('factory')?.(), b = require('fs');", Options: []any{map[string]any{"allowCall": true}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 51)}},
			{Code: "var a = require('fs')?.readFile, b = require('fs');", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 52)}},
			// The upstream rule deliberately does not resolve shadowed require bindings.
			{Code: "function f(require) { const a = require('x'), b = 0; }", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 23, 1, 53)}},
			// Range coverage: for headers exclude their separator, statements include semicolons.
			{Code: "for (let fs = require('fs'), i = 0; i < 1; i++) {}", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 6, 1, 35)}},
			{Code: "export const fs = require('fs'), i = 0;", FileName: "input.mjs", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 8, 1, 40)}},
			{Code: `/* 😀 */
const 文件 = require('fs'),
      图标 = '😀'; // trailing`, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 2, 1, 3, 17)}},
			{Code: "/* 😀 */ const fs = require('fs'), value = '😀';", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 10, 1, 49)}},
			{Code: `var a = require('a'), x = 0
var b = require('b'), y;`, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 28), noMixedRequiresError("noMixRequire", 2, 1, 2, 25)}},
			{Code: "var a = require('a'), x = 0 /* before semicolon */;", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 52)}},
			// JavaScript JSDoc casts and parentheses are transparent.
			{Code: "var a = /** @type {any} */ (require('fs')), b = 0;", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 51)}},
			{Code: "var a = (/** @type {any} */ (require))('fs'), b = 0;", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 53)}},
			{Code: "const a = require<string>('fs'), b = 0;", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 40)}},
			{Code: "declare const a = require('fs'), b = 0;", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 40)}},
			{Code: "export declare const a = require('fs'), b = 0;", FileName: "input.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 8, 1, 47)}},
			{Code: "using a = require('a'), b = resource;", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 38)}},
			{Code: "async function f() { await using a = require('a'), b = resource; }", FileName: "input.ts", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 22, 1, 65)}},
			// JSX expressions can contain declarations, but JSX values are ordinary initializers.
			{Code: "const element = <Widget load={() => { const a = require('a'), b = 0; }} />;", FileName: "input.tsx", Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 39, 1, 69)}},
		},
	)
}
