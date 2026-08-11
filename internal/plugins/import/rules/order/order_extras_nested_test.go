package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderNestedAndSyntaxEdges(t *testing.T) {
	alphabetize := map[string]any{"alphabetize": map[string]any{"order": "asc"}}
	namedRequire := map[string]any{
		"named":       map[string]any{"require": true},
		"alphabetize": map[string]any{"order": "asc"},
	}
	namedCJS := map[string]any{
		"named":       map[string]any{"cjsExports": true},
		"alphabetize": map[string]any{"order": "asc"},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Case-insensitive alphabetizing uses JavaScript's full default case
			// conversion. U+0130 expands to i + combining dot rather than Go's
			// simple one-rune lowercase mapping.
			{
				Code: "import first from 'i\u0307a';\nimport second from '\u0130b';",
				Options: map[string]any{
					"alphabetize": map[string]any{"order": "asc", "caseInsensitive": true},
				},
			},
			// An outer lexical declaration shadows the CommonJS receiver in all
			// descendant scopes, even when the assignment is several blocks deep.
			{
				Code:    "const module = {};\nfunction outer() {\n  if (ready) {\n    (() => {\n      module.exports = { b, a };\n      module.exports.Z = 1;\n      module.exports.A = 1;\n    })();\n  }\n}",
				Options: namedCJS,
			},
			// Parameter shadowing applies independently to the bare exports form.
			{
				Code:    "function write(exports) {\n  exports.B = 1;\n  exports.A = 1;\n}",
				Options: namedCJS,
			},
			// Direct parents, rather than the nearest statement-list ancestor,
			// isolate export sequences in unbraced control-flow bodies.
			{
				Code: "if (one) exports.Z = 1;\n" +
					"if (two) exports.A = 1;\n" +
					"while (three) module.exports.Y = 1;\n" +
					"while (four) module.exports.B = 1;",
				Options: namedCJS,
			},
			// Each switch clause is also an independent direct-parent sequence.
			{
				Code:    "switch (kind) {\ncase 1:\n  exports.Z = 1;\n  break;\ncase 2:\n  exports.A = 1;\n}",
				Options: namedCJS,
			},
			// Upstream accepts an Identifier property in module[exports], but a
			// string-literal property is deliberately not a CommonJS target.
			{Code: `module["exports"] = { b, a };`, Options: namedCJS},
			// JavaScript trim semantics treat FEFF as whitespace on an otherwise
			// empty line, so this already separates the two import groups.
			{
				Code:    "import fs from 'fs';\n\uFEFF\nimport sibling from './sibling';",
				Options: map[string]any{"newlines-between": "always"},
			},
			// A chained require keeps the CallExpression as its upstream multiline
			// node; a line break in the surrounding member chain does not turn the
			// import itself into a multiline island.
			{
				Code: "const first = require('first')\n  .value;\n" +
					"const second = require('second');",
				Options: map[string]any{
					"newlines-between":   "always-and-inside-groups",
					"consolidateIslands": "inside-groups",
				},
			},
			// Newline counting follows the require CallExpression location, as in
			// ESTree, even when its enclosing statement continues on a later line.
			{
				Code: "const fs = require('fs')\n\n  .promises;\n" +
					"const sibling = require('./sibling');",
				Options: map[string]any{"newlines-between": "always"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// A direct require initializer is promoted to its whole variable
			// declaration, so a split declaration is a multiline island.
			{
				Code: "const first =\n  require('first');\n" +
					"const second = require('second');",
				Options: map[string]any{
					"newlines-between":   "always-and-inside-groups",
					"consolidateIslands": "inside-groups",
				},
				Output: []string{"const first =\n  require('first');\n\n" +
					"const second = require('second');"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consolidate", Line: 2, Column: 3}},
			},
			// Separate nested statement containers are alphabetized independently;
			// no assignment may move across either block boundary.
			{
				Code: "if (one) {\n  exports.Z = 1;\n  exports.A = 1;\n}\n" +
					"if (two) {\n  exports.Y = 1;\n  exports.B = 1;\n}",
				Options: namedCJS,
				Output: []string{"if (one) {\n  exports.A = 1;\n  exports.Z = 1;\n}\n" +
					"if (two) {\n  exports.B = 1;\n  exports.Y = 1;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Line: 3, Column: 3},
					{MessageId: "order", Line: 7, Column: 3},
				},
			},
			// Computed Identifier properties, parenthesized RHS objects, and
			// parenthesized Identifier values all retain the upstream ESTree shape.
			{
				Code:    `module[exports] = ({ [b]: (bee), a });`,
				Options: namedCJS,
				Output:  []string{`module[exports] = ({ a, [b]: (bee) });`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 1, Column: 34}},
			},
			// The linter's shared traversal reaches require destructuring through
			// function, loop, and block nesting; parentheses stay transparent.
			{
				Code:    "const run = () => {\n  while (ready) {\n    const { z, a } = require(((\"pkg\")));\n    break;\n  }\n};",
				Options: namedRequire,
				Output:  []string{"const run = () => {\n  while (ready) {\n    const { a, z } = require(((\"pkg\")));\n    break;\n  }\n};"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 3, Column: 16, EndLine: 3, EndColumn: 17}},
			},
			// Parentheses around a require argument are absent from ESTree and do
			// not stop module-level ordering or whole-statement fixes.
			{
				Code:    "const b = require(((\"./b\")));\nconst a = require((\"./a\"));",
				Options: alphabetize,
				Output:  []string{"const a = require((\"./a\"));\nconst b = require(((\"./b\")));\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 11}},
			},
			// Static string names are handled through the shared property-name
			// helper instead of assuming every specifier name is an Identifier.
			{
				Code: "import { \"z\" as z, \"a\" as a } from './pkg';\n" +
					"export { z as \"zed\", a as \"aye\" };",
				Options: map[string]any{
					"named":       map[string]any{"import": true, "export": true},
					"alphabetize": map[string]any{"order": "asc"},
				},
				Output: []string{"import { \"a\" as a, \"z\" as z } from './pkg';\n" +
					"export { a as \"aye\", z as \"zed\" };"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Line: 1},
					{MessageId: "order", Line: 2},
				},
			},
			// Named fixes use ECMAScript trim semantics. FEFF is whitespace in
			// JavaScript (but not in unicode.IsSpace) and stays trailing trivia.
			{
				Code:    "import { b, a\uFEFF } from 'pkg';",
				Options: map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}},
				Output:  []string{"import { a, b\uFEFF } from 'pkg';"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 1, Column: 13, EndLine: 1, EndColumn: 14}},
			},
			// JavaScript relational comparison is UTF-16 based: U+10000 starts
			// with D800 and therefore sorts before BMP U+E000.
			{
				Code:    "import bmp from '\uE000';\nimport astral from '\U00010000';",
				Options: alphabetize,
				Output:  []string{"import astral from '\U00010000';\nimport bmp from '\uE000';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// tsgo preserves a lone surrogate escape as WTF-8; comparison restores
			// its original UTF-16 code unit instead of treating it as RuneError.
			{
				Code:    "import bmp from '\uE000';\nimport lone from '\\uD800';",
				Options: alphabetize,
				Output:  []string{"import lone from '\\uD800';\nimport bmp from '\uE000';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
		},
	)
}
