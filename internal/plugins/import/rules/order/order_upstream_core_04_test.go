// TestOrderUpstreamCore04 migrates the upstream core cases from import-js/eslint-plugin-import at 1a39fb8672c83f8781c3a97f02eee00a7a2f0253.
// It preserves upstream messages, source positions, options, and iterative autofix outputs.
// Framework-only parser/resolver settings use rslint native equivalents; tsgo lock-ins live in order_extras_*_test.go.
package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderUpstreamCore04(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Upstream case 0:60.
			{Code: "\n        import A from 'a';\n\n        import C from 'c';\n\n        import B from 'b';\n      ", Options: []any{map[string]any{"newlines-between": "always", "distinctGroup": true, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "a", "group": "external", "position": "before"}, map[string]any{"pattern": "b", "group": "external", "position": "after"}}}}},
			// Upstream case 0:61.
			{Code: "\n        import A from 'a';\n        import C from 'c';\n        import B from 'b';\n      ", Options: []any{map[string]any{"newlines-between": "always", "distinctGroup": false, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "a", "group": "external", "position": "before"}, map[string]any{"pattern": "b", "group": "external", "position": "after"}}}}},
			// Upstream case 0:62.
			{Code: "\n        import A from 'a';\n\n        import b from './b';\n        import B from './B';\n      ", Options: []any{map[string]any{"newlines-between": "always", "distinctGroup": false, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "a", "group": "external"}, map[string]any{"pattern": "b", "group": "internal", "position": "before"}}}}},
			// Upstream case 0:63.
			{Code: "\n        import A from \"baz\";\n        import B from \"Bar\";\n        import C from \"Foo\";\n\n        import D from \"..\";\n        import E from \"../\";\n        import F from \"../baz\";\n        import G from \"../Bar\";\n        import H from \"../Foo\";\n\n        import I from \".\";\n        import J from \"./baz\";\n        import K from \"./Bar\";\n        import L from \"./Foo\";\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"caseInsensitive": false, "order": "asc"}, "newlines-between": "always", "groups": []any{[]any{"builtin", "external", "internal", "unknown", "object", "type"}, "parent", []any{"sibling", "index"}}, "distinctGroup": false, "pathGroupsExcludedImportTypes": []any{}, "pathGroups": []any{map[string]any{"pattern": "./", "group": "sibling", "position": "before"}, map[string]any{"pattern": ".", "group": "sibling", "position": "before"}, map[string]any{"pattern": "..", "group": "parent", "position": "before"}, map[string]any{"pattern": "../", "group": "parent", "position": "before"}, map[string]any{"pattern": "[a-z]*", "group": "external", "position": "before"}, map[string]any{"pattern": "../[a-z]*", "group": "parent", "position": "before"}, map[string]any{"pattern": "./[a-z]*", "group": "sibling", "position": "before"}}}}},
			// Upstream case 0:64.
			{Code: "\n        import B from './B';\n        import b from './b';\n      ", Options: []any{map[string]any{"alphabetize": map[string]any{"order": "asc", "orderImportKind": "asc", "caseInsensitive": true}}}},
			// Upstream case 0:65.
			{Code: "\n        import { a, B as C, Z } from './Z';\n        const { D, n: c, Y } = require('./Z');\n        export { C, D };\n        export { A, B, C as default } from \"./Z\";\n\n        const { [\"ignore require-statements with non-identifier imports\"]: z, d } = require(\"./Z\");\n        exports = { [\"ignore exports statements with non-identifiers\"]: Z, D };\n      ", Options: []any{map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}}}},
			// Upstream case 0:66.
			{Code: "\n        const { b, A } = require('./Z');\n      ", Options: []any{map[string]any{"named": true, "alphabetize": map[string]any{"order": "desc"}}}},
			// Upstream case 0:67.
			{Code: "\n        import { A, B } from \"./Z\";\n        export { Z, A } from \"./Z\";\n        export { N, P } from \"./Z\";\n        const { X, Y } = require(\"./Z\");\n      ", Options: []any{map[string]any{"named": map[string]any{"require": true, "import": true, "export": false}}}},
			// Upstream case 0:68.
			{Code: "\n        import { B, A } from \"./Z\";\n        const { D, C } = require(\"./Z\");\n        export { B, A } from \"./Z\";\n      ", Options: []any{map[string]any{"named": map[string]any{"require": false, "import": false, "export": false}}}},
			// Upstream case 0:69.
			{Code: "\n        import { B, A, R } from \"foo\";\n        const { D, O, G } = require(\"tunes\");\n        export { B, A, Z } from \"foo\";\n      ", Options: []any{map[string]any{"named": map[string]any{"enabled": false}}}},
			// Upstream case 0:70.
			{Code: "\n        import { A as A, A as B, A as C } from \"./Z\";\n        const { a, a: b, a: c } = require(\"./Z\");\n      ", Options: []any{map[string]any{"named": true}}},
			// Upstream case 0:71.
			{Code: "\n        import { A, B, C } from \"./Z\";\n        exports = { A, B, C };\n        module.exports = { a: A, b: B, c: C };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:72.
			{Code: "\n        module.exports.A = { };\n        module.exports.A.B = { };\n        module.exports.B = { };\n        exports.C = { };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:73.
			{Code: "\n        var exports = null;\n        var module = null;\n        exports = { };\n        module = { };\n        module.exports = { };\n        module.exports.U = { };\n        module.exports.N = { };\n        module.exports.C = { };\n        exports.L = { };\n        exports.E = { };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}},
			// Upstream case 0:74.
			{Code: "\n        exports[\"B\"] = { };\n        exports[\"C\"] = { };\n        exports[\"A\"] = { };\n      ", Options: []any{map[string]any{"named": map[string]any{"cjsExports": true}, "alphabetize": map[string]any{"order": "asc"}}}},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream case 0:0.
			{Code: "\n        var async = require('async');\n        var fs = require('fs');\n      ", Output: []string{"\n        var fs = require('fs');\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:1.
			{Code: "\n        import a from './foo';\n        import b from '../bar';\n      ", Options: []any{map[string]any{"groups": []any{"builtin", "external", "internal", "type"}, "alphabetize": map[string]any{"order": "asc", "caseInsensitive": true}}}, Output: []string{"\n        import b from '../bar';\n        import a from './foo';\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`../bar` import should occur before import of `./foo`", Line: 3, Column: 9, EndLine: 3, EndColumn: 32}}},
			// Upstream case 0:2.
			{Code: "\n        var async = require('async');\n        var fs = require('fs'); \n      ", Output: []string{"\n        var fs = require('fs'); \n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:3.
			{Code: "\n        var async = require('async');\n        var fs = require('fs'); /* comment */\n      ", Output: []string{"\n        var fs = require('fs'); /* comment */\n        var async = require('async');\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// Upstream case 0:4.
			{Code: "\n        /* comment1 */  var async = require('async'); /* comment2 */\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n      ", Output: []string{"\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n        /* comment1 */  var async = require('async'); /* comment2 */\n      "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`fs` import should occur before import of `async`", Line: 3, Column: 34, EndLine: 3, EndColumn: 47}}},
		},
	)
}
