// TestOrderDimension4 exercises tsgo AST shapes that the ESTree upstream suite
// cannot express directly. Every applicable universal row is labelled below;
// declaration variants (class/function, async/generator, abstract/declare and
// overload signatures) are N/A because import/order only consumes module
// statements, require declarators, and CommonJS assignment expressions.
package order_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderDimension4(t *testing.T) {
	alphabetize := map[string]any{"alphabetize": map[string]any{"order": "asc"}}
	namedRequire := map[string]any{
		"named":       map[string]any{"enabled": true, "require": true},
		"alphabetize": map[string]any{"order": "asc"},
	}
	namedCJS := map[string]any{
		"named":       map[string]any{"enabled": true, "cjsExports": true},
		"alphabetize": map[string]any{"order": "asc"},
	}
	commentChain := strings.Repeat(" /* kept */", 101)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS non-null wrapper around require is intentionally opaque ----
			{Code: "const sibling = require('./z')!;\nconst fs = require('fs');", Options: alphabetize},
			// ---- Dimension 4: TS `as` wrapper around require is intentionally opaque ----
			{Code: "const sibling = require('./z') as unknown;\nconst fs = require('fs');", Options: alphabetize},
			// ---- Dimension 4: TS `satisfies` wrapper around require is intentionally opaque ----
			{Code: "const sibling = require('./z') satisfies unknown;\nconst fs = require('fs');", Options: alphabetize},
			// ---- Dimension 4: optional call and optional member chains are intentionally opaque ----
			{Code: "const sibling = require?.('./z');\nconst fs = require('fs');", Options: alphabetize},
			{Code: "const sibling = require('./z')?.value;\nconst fs = require('fs');", Options: alphabetize},

			// ---- Dimension 4: string, numeric, computed, rest, default, and nested binding keys bail out ----
			{Code: `const { "b": b, a } = require("pkg");`, Options: namedRequire},
			{Code: `const { 1: b, a } = require("pkg");`, Options: namedRequire},
			{Code: `const { [key]: b, a } = require("pkg");`, Options: namedRequire},
			{Code: `const { b, ...a } = require("pkg");`, Options: namedRequire},
			{Code: `const { b = 1, a } = require("pkg");`, Options: namedRequire},
			{Code: `const { b: { value }, a } = require("pkg");`, Options: namedRequire},

			// ---- Dimension 4: string, numeric, method, and spread CJS keys bail out ----
			{Code: `module.exports = { "b": b, a };`, Options: namedCJS},
			{Code: `module.exports = { 1: b, a };`, Options: namedCJS},
			{Code: `module.exports = { b() {}, a };`, Options: namedCJS},
			{Code: `module.exports = { ...b, a };`, Options: namedCJS},
			{Code: `module.exports = { b = fallback, a };`, Options: namedCJS},

			// ---- Dimension 4: nested require module ordering is outside Program/TSModuleBlock ----
			{Code: "function f() {\n  const sibling = require('./z');\n  const fs = require('fs');\n}", Options: alphabetize},

			// ---- Dimension 4: empty/dynamic/template/spread require arguments degrade without panic ----
			{Code: "const a = require();\nconst b = require(name);\nconst c = require(`pkg`);\nconst d = require(...args);"},
			// ---- Dimension 4: empty and comment-only inputs degrade without diagnostics ----
			{Code: ""},
			{Code: "// only trivia\n/* still only trivia */"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized require callee and expression remain transparent ----
			{
				Code:    "const b = (require)('./b');\nconst a = ((require('./a')));",
				Options: alphabetize,
				Output:  []string{"const a = ((require('./a')));\nconst b = (require)('./b');\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Dimension 4: property, string, numeric, template, and computed require access chains are transparent ----
			{
				Code: "const f = require('./f').value;\n" +
					"const e = require('./e')['value'];\n" +
					"const d = require('./d')[`value`];\n" +
					"const c = require('./c')[0];\n" +
					"const b = require('./b')[Symbol.iterator];\n" +
					"const a = require('./a');",
				Options: alphabetize,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order"}, {MessageId: "order"}, {MessageId: "order"}, {MessageId: "order"}, {MessageId: "order"},
				},
			},
			// ---- Dimension 4: identifier binding keys and aliases participate in named sorting ----
			{
				Code:    `const { b: bee, a } = require("pkg");`,
				Options: namedRequire,
				Output:  []string{`const { a, b: bee } = require("pkg");`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Dimension 4: computed identifier CJS keys match; literal keys do not ----
			{
				Code:    "exports[B] = 1;\nmodule.exports.A = 1;",
				Options: namedCJS,
				Output:  []string{"module.exports.A = 1;\nexports[B] = 1;\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Dimension 4: a computed Identifier key stays an Identifier in ESTree ----
			{
				Code:    `module.exports = { [key]: b, a };`,
				Options: namedCJS,
				Output:  []string{`module.exports = { a, [key]: b };`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 1, Column: 30}},
			},
			// ---- Dimension 4: named require listeners traverse nested function blocks ----
			{
				Code:    "function f() {\n  const { b, a } = require('pkg');\n}",
				Options: namedRequire,
				Output:  []string{"function f() {\n  const { a, b } = require('pkg');\n}"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Dimension 4: named CommonJS listeners traverse nested function blocks ----
			{
				Code:    "function f() {\n  exports.B = 1;\n  module.exports.A = 1;\n}",
				Options: namedCJS,
				Output:  []string{"function f() {\n  module.exports.A = 1;\n  exports.B = 1;\n}"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Dimension 4: each TypeScript module block has an independent import sequence ----
			{
				Code:    "declare namespace A {\n  import b = require('./b');\n  import a = require('./a');\n}",
				Options: alphabetize,
				Output:  []string{"declare namespace A {\n  import a = require('./a');\n  import b = require('./b');\n}"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Dimension 4: trailing comments, CRLF, and Unicode survive a module swap ----
			{
				Code:    "import β from 'β'; // β\r\nimport α from 'α'; // α\r\n",
				Options: alphabetize,
				Output:  []string{"import α from 'α'; // α\r\nimport β from 'β'; // β\r\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// Upstream only inspects 100 neighboring tokens/comments. The shared
			// comment index has no arbitrary cutoff, so every attached comment moves.
			{
				Code:    "import b from 'b';" + commentChain + "\nimport a from 'a';\n",
				Options: alphabetize,
				Output:  []string{"import a from 'a';\nimport b from 'b';" + commentChain + "\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
		},
	)
}
