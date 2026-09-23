package no_unresolved_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_unresolved"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Every semantic case from eslint-plugin-import v2.32.0, including SYNTAX_CASES.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-unresolved.js
// Parser variants share the native parser. Webpack groups remain as explained
// skips because the native runtime cannot execute JavaScript resolver plugins.
// Babel's export-default-from proposal is not supported by the tsgo parser.
func TestNoUnresolvedUpstream(t *testing.T) {
	t.Run("no-unresolved (node)", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import \"./malformed.js\""},
				{Code: "import foo from \"./bar\";", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import bar from './bar.js';", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import {someThing} from './test-module';", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import fs from 'fs';", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import('fs');", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import('fs');", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import * as foo from \"a\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "export { foo } from \"./bar\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "export * from \"./bar\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "let foo; export { foo }", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "export * as bar from \"./bar\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "export bar from \"./bar\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Skip: true}, // Babel export-default-from proposal.

				{Code: "import foo from \"./jsx/MyUnCoolComponent.jsx\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "var foo = require(\"./bar\")", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(\"./bar\")", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(\"./does-not-exist\")", Options: []any{map[string]any{"commonjs": false}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(\"./does-not-exist\")", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require([\"./bar\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "define([\"./bar\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require([\"./does-not-exist\"], function (bar) {})", Options: []any{map[string]any{"amd": false}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "define([\"require\", \"exports\", \"module\"], function (r, e, m) { })", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require([\"./does-not-exist\"])", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "define([\"./does-not-exist\"], function (bar) {})", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(\"./does-not-exist\", \"another arg\")", Options: []any{map[string]any{"commonjs": true, "amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "proxyquire(\"./does-not-exist\")", Options: []any{map[string]any{"commonjs": true, "amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "(function() {})(\"./does-not-exist\")", Options: []any{map[string]any{"commonjs": true, "amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "define([0, foo], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(0)", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(foo)", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import reallyfake from \"./reallyfake/module\"", Settings: map[string]any{"import/ignore": []any{"^\\./fake/"}, "import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './reallyfake/module'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 45}}},
				{Code: "import bar from './baz';", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 24}}},
				{Code: "import bar from './empty-folder';", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './empty-folder'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 33}}},
				{Code: "import { DEEP } from 'in-alternate-root';", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'in-alternate-root'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 41}}},
				{Code: "import('in-alternate-root').then(function({DEEP}) {});", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'in-alternate-root'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
				{Code: "export { foo } from \"./does-not-exist\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 39}}},
				{Code: "export * from \"./does-not-exist\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 33}}},
				{Code: "import('in-alternate-root').then(function({DEEP}) {});", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'in-alternate-root'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
				{Code: "export * as bar from \"./does-not-exist\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 40}}},
				{Code: "export bar from \"./does-not-exist\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Skip: true, // Babel export-default-from proposal.
					Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 35}}},
				{Code: "var bar = require(\"./baz\")", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}}},
				{Code: "require(\"./baz\")", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
				{Code: "require([\"./baz\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
				{Code: "define([\"./baz\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
				{Code: "define([\"./baz\", \"./bar\", \"./does-not-exist\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}, {Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 45}}},
			})
	})
	t.Run("issue #333 (node)", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import foo from \"./bar.json\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import foo from \"./bar\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import foo from \"./bar.json\"", Settings: map[string]any{"import/extensions": []any{".js"}, "import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import foo from \"./bar\"", Settings: map[string]any{"import/extensions": []any{".js"}, "import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import bar from \"./foo.json\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './foo.json'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 29}}},
			})
	})
	t.Run("case sensitivity", func(t *testing.T) {
		root := unresolvedRoot(t, true)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import foo from \"./jsx/MyUncoolComponent.jsx\"", Options: []any{map[string]any{"caseSensitive": false}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import foo from \"./jsx/MyUncoolComponent.jsx\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Casing of ./jsx/MyUncoolComponent.jsx does not match the underlying filesystem.", Line: 1, Column: 17, EndLine: 1, EndColumn: 46}}},
				{Code: "import foo from \"./jsx/MyUncoolComponent.jsx\"", Options: []any{map[string]any{"caseSensitive": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Casing of ./jsx/MyUncoolComponent.jsx does not match the underlying filesystem.", Line: 1, Column: 17, EndLine: 1, EndColumn: 46}}},
			})
	})
	t.Run("case sensitivity strict", func(t *testing.T) {
		root := unresolvedRoot(t, true)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import foo from \"/NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx\"", Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import foo from \"/NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx\"", Options: []any{map[string]any{"caseSensitiveStrict": true}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Casing of /NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx does not match the underlying filesystem.", Line: 1, Column: 17, EndLine: 1, EndColumn: 59}}},
				{Code: "import foo from \"/NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx\"", Options: []any{map[string]any{"caseSensitiveStrict": true, "caseSensitive": false}}, Settings: map[string]any{"import/resolver": "node", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Casing of /NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx does not match the underlying filesystem.", Line: 1, Column: 17, EndLine: 1, EndColumn: 59}}},
			})
	})
	t.Run("no-unresolved (webpack)", func(t *testing.T) {
		t.Skip("Webpack resolver plugins require JavaScript execution; the identical node cases run above")
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import \"./malformed.js\""},
				{Code: "import foo from \"./bar\";", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import bar from './bar.js';", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import {someThing} from './test-module';", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import fs from 'fs';", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import('fs');", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import('fs');", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import * as foo from \"a\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "export { foo } from \"./bar\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "export * from \"./bar\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "let foo; export { foo }", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "export * as bar from \"./bar\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "export bar from \"./bar\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Skip: true}, // Babel export-default-from proposal.

				{Code: "import foo from \"./jsx/MyUnCoolComponent.jsx\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "var foo = require(\"./bar\")", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(\"./bar\")", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(\"./does-not-exist\")", Options: []any{map[string]any{"commonjs": false}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(\"./does-not-exist\")", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require([\"./bar\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "define([\"./bar\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require([\"./does-not-exist\"], function (bar) {})", Options: []any{map[string]any{"amd": false}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "define([\"require\", \"exports\", \"module\"], function (r, e, m) { })", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require([\"./does-not-exist\"])", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "define([\"./does-not-exist\"], function (bar) {})", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(\"./does-not-exist\", \"another arg\")", Options: []any{map[string]any{"commonjs": true, "amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "proxyquire(\"./does-not-exist\")", Options: []any{map[string]any{"commonjs": true, "amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "(function() {})(\"./does-not-exist\")", Options: []any{map[string]any{"commonjs": true, "amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "define([0, foo], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(0)", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "require(foo)", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import reallyfake from \"./reallyfake/module\"", Settings: map[string]any{"import/ignore": []any{"^\\./fake/"}, "import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './reallyfake/module'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 45}}},
				{Code: "import bar from './baz';", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 24}}},
				{Code: "import bar from './empty-folder';", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './empty-folder'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 33}}},
				{Code: "import { DEEP } from 'in-alternate-root';", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'in-alternate-root'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 41}}},
				{Code: "import('in-alternate-root').then(function({DEEP}) {});", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'in-alternate-root'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
				{Code: "export { foo } from \"./does-not-exist\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 39}}},
				{Code: "export * from \"./does-not-exist\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 33}}},
				{Code: "import('in-alternate-root').then(function({DEEP}) {});", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'in-alternate-root'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
				{Code: "export * as bar from \"./does-not-exist\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 40}}},
				{Code: "export bar from \"./does-not-exist\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Skip: true, // Babel export-default-from proposal.
					Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 35}}},
				{Code: "var bar = require(\"./baz\")", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}}},
				{Code: "require(\"./baz\")", Options: []any{map[string]any{"commonjs": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
				{Code: "require([\"./baz\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
				{Code: "define([\"./baz\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
				{Code: "define([\"./baz\", \"./bar\", \"./does-not-exist\"], function (bar) {})", Options: []any{map[string]any{"amd": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './baz'.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}, {Message: "Unable to resolve path to module './does-not-exist'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 45}}},
			})
	})
	t.Run("issue #333 (webpack)", func(t *testing.T) {
		t.Skip("Webpack resolver plugins require JavaScript execution; the identical node cases run above")
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import foo from \"./bar.json\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import foo from \"./bar\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import foo from \"./bar.json\"", Settings: map[string]any{"import/extensions": []any{".js"}, "import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
				{Code: "import foo from \"./bar\"", Settings: map[string]any{"import/extensions": []any{".js"}, "import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import bar from \"./foo.json\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './foo.json'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 29}}},
			})
	})
	t.Run("case sensitivity", func(t *testing.T) {
		t.Skip("Webpack resolver plugins require JavaScript execution; the identical node cases run above")
		root := unresolvedRoot(t, true)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import foo from \"./jsx/MyUncoolComponent.jsx\"", Options: []any{map[string]any{"caseSensitive": false}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import foo from \"./jsx/MyUncoolComponent.jsx\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Casing of ./jsx/MyUncoolComponent.jsx does not match the underlying filesystem.", Line: 1, Column: 17, EndLine: 1, EndColumn: 46}}},
				{Code: "import foo from \"./jsx/MyUncoolComponent.jsx\"", Options: []any{map[string]any{"caseSensitive": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Casing of ./jsx/MyUncoolComponent.jsx does not match the underlying filesystem.", Line: 1, Column: 17, EndLine: 1, EndColumn: 46}}},
			})
	})
	t.Run("case sensitivity strict", func(t *testing.T) {
		t.Skip("Webpack resolver plugins require JavaScript execution; the identical node cases run above")
		root := unresolvedRoot(t, true)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import foo from \"/NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx\"", Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import foo from \"/NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx\"", Options: []any{map[string]any{"caseSensitiveStrict": true}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Casing of /NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx does not match the underlying filesystem.", Line: 1, Column: 17, EndLine: 1, EndColumn: 59}}},
				{Code: "import foo from \"/NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx\"", Options: []any{map[string]any{"caseSensitiveStrict": true, "caseSensitive": false}}, Settings: map[string]any{"import/resolver": "webpack", "import/cache": map[string]any{"lifetime": 0}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Casing of /NO-UNRESOLVED/jsx/MyUnCoolComponent.jsx does not match the underlying filesystem.", Line: 1, Column: 17, EndLine: 1, EndColumn: 59}}},
			})
	})
	t.Run("no-unresolved (import/resolve legacy)", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import { DEEP } from 'in-alternate-root';", Settings: map[string]any{"import/resolve": map[string]any{"paths": []any{"/no-unresolved/alternate-root"}}}},
				{Code: "import { DEEP } from 'in-alternate-root'; import { bar } from 'src-bar';", Settings: map[string]any{"import/resolve": map[string]any{"paths": []any{"src-root", "alternate-root"}}}},
				{Code: "import * as foo from \"jsx-module/foo\"", Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".jsx"}}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import * as foo from \"jsx-module/foo\"", Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'jsx-module/foo'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 38}}},
			})
	})
	t.Run("no-unresolved (webpack-specific)", func(t *testing.T) {
		t.Skip("Webpack resolver plugins require JavaScript execution; the identical node cases run above")
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import * as foo from \"jsx-module/foo\"", Settings: map[string]any{"import/resolver": "webpack"}},
				{Code: "import * as foo from \"some-loader?with=args!jsx-module/foo\"", Settings: map[string]any{"import/resolver": "webpack"}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import * as foo from \"jsx-module/foo\"", Settings: map[string]any{"import/resolver": map[string]any{"webpack": map[string]any{"config": "webpack.empty.config.js"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'jsx-module/foo'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 38}}},
			})
	})
	t.Run("no-unresolved ignore list", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import \"./malformed.js\"", Options: []any{map[string]any{"ignore": []any{".png$", ".gif$"}}}},
				{Code: "import \"./test.giffy\"", Options: []any{map[string]any{"ignore": []any{".png$", ".gif$"}}}},
				{Code: "import \"./test.gif\"", Options: []any{map[string]any{"ignore": []any{".png$", ".gif$"}}}},
				{Code: "import \"./test.png\"", Options: []any{map[string]any{"ignore": []any{".png$", ".gif$"}}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import \"./test.gif\"", Options: []any{map[string]any{"ignore": []any{".png$"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './test.gif'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
				{Code: "import \"./test.png\"", Options: []any{map[string]any{"ignore": []any{".gif$"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './test.png'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			})
	})
	t.Run("no-unresolved unknown resolver", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{},
			[]rule_tester.InvalidTestCase{
				{Code: "import \"./malformed.js\"", Settings: map[string]any{"import/resolver": "doesnt-exist"}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Resolve error: unable to load resolver \"doesnt-exist\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 1}, {Message: "Unable to resolve path to module './malformed.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
				{Code: "import \"./malformed.js\"; import \"./fake.js\"", Settings: map[string]any{"import/resolver": "doesnt-exist"}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Resolve error: unable to load resolver \"doesnt-exist\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 1}, {Message: "Unable to resolve path to module './malformed.js'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 24}, {Message: "Unable to resolve path to module './fake.js'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 44}}},
			})
	})
	t.Run("no-unresolved electron", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import \"electron\"", Settings: map[string]any{"import/core-modules": []any{"electron"}}},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import \"electron\"", Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module 'electron'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			})
	})
	t.Run("no-unresolved syntax verification", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "for (let { foo, bar } of baz) {}"},
				{Code: "for (let [ foo, bar ] of baz) {}"},
				{Code: "const { x, y } = bar"},
				{Code: "const { x, y, ...z } = bar"},
				{Code: "let x; export { x }"},
				{Code: "let x; export { x as y }"},
				{Code: "export const x = null"},
				{Code: "export var x = null"},
				{Code: "export let x = null"},
				{Code: "export default x"},
				{Code: "export default class x {}"},
				{Code: "import json from \"./data.json\"", Settings: map[string]any{"import/extensions": []any{".js"}}},
				{Code: "import foo from \"./foobar.json\";", Settings: map[string]any{"import/extensions": []any{".js"}}},
				{Code: "import foo from \"./foobar\";", Settings: map[string]any{"import/extensions": []any{".js"}}},
				{Code: "import { foo } from \"./issue-370-commonjs-namespace/bar\"", Settings: map[string]any{"import/ignore": []any{"foo"}}},
				{Code: "export * from \"./issue-370-commonjs-namespace/bar\"", Settings: map[string]any{"import/ignore": []any{"foo"}}},
				{Code: "import * as a from \"./commonjs-namespace/a\"; a.b"},
				{Code: "import { foo } from \"./ignore.invalid.extension\""},
			},
			[]rule_tester.InvalidTestCase{})
	})
	t.Run("import() with built-in parser", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import('fs');"},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import(\"./does-not-exist-l0w9ssmcqy9\").then(() => {})", Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module './does-not-exist-l0w9ssmcqy9'.", Line: 1, Column: 8, EndLine: 1, EndColumn: 38}}},
			})
	})
	t.Run("typescript: no-unresolved ignore type-only", func(t *testing.T) {
		root := unresolvedRoot(t, false)
		rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unresolved.NoUnresolvedRule,
			[]rule_tester.ValidTestCase{
				{Code: "import type { JSONSchema7Type } from \"@types/json-schema\";"},
				{Code: "export type { JSONSchema7Type } from \"@types/json-schema\";"},
			},
			[]rule_tester.InvalidTestCase{
				{Code: "import { JSONSchema7Type } from \"@types/json-schema\";", Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module '@types/json-schema'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 53}}},
				{Code: "export { JSONSchema7Type } from \"@types/json-schema\";", Errors: []rule_tester.InvalidTestCaseError{{Message: "Unable to resolve path to module '@types/json-schema'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 53}}},
			})
	})
}

// Documentation examples use ./foo and ./mod as missing modules. Absolute
// casing examples are covered by the portable upstream strict-mode cases.
func TestNoUnresolvedDocumentation(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, item := range []struct {
		code    string
		options map[string]any
		names   []string
	}{
		{`import x from './foo'`, nil, []string{"./foo"}},
		{`const { default: x } = require('./foo'); require(0); require(['x', 'y'], function(x, y) {});`, map[string]any{"commonjs": true}, []string{"./foo"}},
		{`define(['./foo'], function(foo) {}); require(['./foo'], function(foo) {}); const { default: x } = require('./foo');`, map[string]any{"amd": true}, []string{"./foo", "./foo"}},
		{`const { default: x } = require('./foo'); define(['./foo'], function(foo) {}); require(['./foo'], function(foo) {});`, map[string]any{"amd": true, "commonjs": true}, []string{"./foo", "./foo", "./foo"}},
		{`import { x } from './mod'; import coolImg from '../../img/coolImg.img';`, map[string]any{"ignore": []any{`\.img$`}}, []string{"./mod"}},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: item.code, Options: item.options, Errors: unresolvedErrors(item.code, item.names...)})
	}
	rule_tester.RunRuleTester(unresolvedRoot(t, false), "tsconfig.json", t, &no_unresolved.NoUnresolvedRule, nil, invalid)
}
