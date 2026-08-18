// TestOrderNestedAndSyntaxEdges and TestOrderDimension4 exercise tsgo-native
// nesting, TypeScript wrappers, Unicode, trivia, and CommonJS shapes that are
// not isolated by the migrated compatibility cases.
//
// Declaration variants (class/function, async/generator, abstract/declare,
// overload signatures, and private keys) are not rule inputs: import/order
// consumes module statements, require declarators, and CommonJS assignments.
package order_test

import (
	"strings"
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
			// eslint-plugin-import v2.32.0 assumes parent and sibling paths occupy
			// separate groups. If configuration merges them, alphabetizing preserves
			// their source order rather than comparing the two relative roots.
			{
				Code: "import sibling from './z';\nimport parent from '../a';",
				Options: map[string]any{
					"groups":      []any{[]any{"parent", "sibling"}},
					"alphabetize": map[string]any{"order": "asc"},
				},
			},
			// Case-insensitive alphabetizing uses JavaScript's full default case
			// conversion. U+0130 expands to i + combining dot rather than Go's
			// simple one-rune lowercase mapping.
			{
				Code: "import first from 'i\u0307a';\nimport second from '\u0130b';",
				Options: map[string]any{
					"alphabetize": map[string]any{"order": "asc", "caseInsensitive": true},
				},
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
			// An identifier property in module[exports] is accepted, while a
			// string-literal property is deliberately not a CommonJS target.
			{Code: `module["exports"] = { b, a };`, Options: namedCJS},
			// JavaScript trim semantics treat FEFF as whitespace on an otherwise
			// empty line, so this already separates the two import groups.
			{
				Code:    "import fs from 'fs';\n\uFEFF\nimport sibling from './sibling';",
				Options: map[string]any{"newlines-between": "always"},
			},
			// A chained require measures the CallExpression itself; a line break in
			// the surrounding member chain does not create a multiline island.
			{
				Code: "const first = require('first')\n  .value;\n" +
					"const second = require('second');",
				Options: map[string]any{
					"newlines-between":   "always-and-inside-groups",
					"consolidateIslands": "inside-groups",
				},
			},
			// Newline counting follows the require CallExpression even when its
			// enclosing statement continues on a later line.
			{
				Code: "const fs = require('fs')\n\n  .promises;\n" +
					"const sibling = require('./sibling');",
				Options: map[string]any{"newlines-between": "always"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// The line-range fixer stops only at LF. ECMAScript still counts U+2028
			// as whitespace, so removing a blank Unicode-separated line
			// consumes both separators and leaves the semicolons adjacent.
			{
				Code:    "import a from 'a';\u2028\u2028import b from 'b';",
				Options: map[string]any{"newlines-between": "never"},
				Output:  []string{"import a from 'a';import b from 'b';"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "groupNewline", Line: 1, Column: 1}},
			},
			// A dotted namespace is represented by a chain of ModuleDeclarations;
			// only the terminal ModuleBlock owns the imports.
			{
				Code: "declare namespace Company.Platform.API {\n" +
					"  import z = require('./z');\n" +
					"  import a = require('./a');\n" +
					"}",
				Options: alphabetize,
				Output: []string{"declare namespace Company.Platform.API {\n" +
					"  import a = require('./a');\n" +
					"  import z = require('./z');\n" +
					"}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 3}},
			},
			// Every TypeScript namespace owns an independent statement list. Three
			// levels lock in recursive collection, source-ordered diagnostics, and
			// fixes that never escape their immediate module block.
			{
				Code: "declare namespace Level1 {\n" +
					"  import z1 = require('./z1');\n" +
					"  import a1 = require('./a1');\n" +
					"  namespace Level2 {\n" +
					"    import z2 = require('./z2');\n" +
					"    import a2 = require('./a2');\n" +
					"    namespace Level3 {\n" +
					"      import z3 = require('./z3');\n" +
					"      import a3 = require('./a3');\n" +
					"    }\n" +
					"  }\n" +
					"}",
				Options: alphabetize,
				Output: []string{"declare namespace Level1 {\n" +
					"  import a1 = require('./a1');\n" +
					"  import z1 = require('./z1');\n" +
					"  namespace Level2 {\n" +
					"    import a2 = require('./a2');\n" +
					"    import z2 = require('./z2');\n" +
					"    namespace Level3 {\n" +
					"      import a3 = require('./a3');\n" +
					"      import z3 = require('./z3');\n" +
					"    }\n" +
					"  }\n" +
					"}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Line: 3},
					{MessageId: "order", Line: 6},
					{MessageId: "order", Line: 9},
				},
			},
			// Named require listeners use the linter traversal, so deeply nested
			// switch/loop/try/catch/finally statement lists are all reached and
			// fixed independently.
			{
				Code: "function load(kind) {\n" +
					"  try {\n" +
					"    switch (kind) {\n" +
					"    case 1:\n" +
					"      for (const value of values) {\n" +
					"        const { z, a } = require('one');\n" +
					"      }\n" +
					"    }\n" +
					"  } catch (error) {\n" +
					"    const { y, b } = require('two');\n" +
					"  } finally {\n" +
					"    do {\n" +
					"      const { x, c } = require('three');\n" +
					"    } while (ready);\n" +
					"  }\n" +
					"}",
				Options: namedRequire,
				Output: []string{"function load(kind) {\n" +
					"  try {\n" +
					"    switch (kind) {\n" +
					"    case 1:\n" +
					"      for (const value of values) {\n" +
					"        const { a, z } = require('one');\n" +
					"      }\n" +
					"    }\n" +
					"  } catch (error) {\n" +
					"    const { b, y } = require('two');\n" +
					"  } finally {\n" +
					"    do {\n" +
					"      const { c, x } = require('three');\n" +
					"    } while (ready);\n" +
					"  }\n" +
					"}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Line: 6},
					{MessageId: "order", Line: 10},
					{MessageId: "order", Line: 13},
				},
			},
			// try, catch, finally, and class static blocks all expose ordinary
			// statement-list parents; CommonJS sequences stay local to each one.
			{
				Code: "try {\n" +
					"  exports.Z = 1;\n" +
					"  exports.A = 1;\n" +
					"} catch {\n" +
					"  module.exports.Y = 1;\n" +
					"  module.exports.B = 1;\n" +
					"} finally {\n" +
					"  exports.X = 1;\n" +
					"  exports.C = 1;\n" +
					"}\n" +
					"class Registry {\n" +
					"  static {\n" +
					"    exports.W = 1;\n" +
					"    module.exports.D = 1;\n" +
					"  }\n" +
					"}",
				Options: namedCJS,
				Output: []string{"try {\n" +
					"  exports.A = 1;\n" +
					"  exports.Z = 1;\n" +
					"} catch {\n" +
					"  module.exports.B = 1;\n" +
					"  module.exports.Y = 1;\n" +
					"} finally {\n" +
					"  exports.C = 1;\n" +
					"  exports.X = 1;\n" +
					"}\n" +
					"class Registry {\n" +
					"  static {\n" +
					"    module.exports.D = 1;\n" +
					"    exports.W = 1;\n" +
					"  }\n" +
					"}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Line: 3},
					{MessageId: "order", Line: 6},
					{MessageId: "order", Line: 9},
					{MessageId: "order", Line: 14},
				},
			},
			// Only the identifier's current lexical scope is checked. An outer
			// declaration therefore does not suppress CommonJS sorting here.
			{
				Code:    "const module = {};\nfunction outer() {\n  if (ready) {\n    (() => {\n      module.exports = { b, a };\n      module.exports.Z = 1;\n      module.exports.A = 1;\n    })();\n  }\n}",
				Options: namedCJS,
				Output:  []string{"const module = {};\nfunction outer() {\n  if (ready) {\n    (() => {\n      module.exports = { a, b };\n      module.exports.A = 1;\n      module.exports.Z = 1;\n    })();\n  }\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Line: 5, Column: 29},
					{MessageId: "order", Line: 7, Column: 7},
				},
			},
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
			// Computed identifier properties, parenthesized RHS objects, and
			// parenthesized identifier values remain sortable in tsgo's AST.
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
			// Parentheses around a require argument are transparent and do not stop
			// module-level ordering or whole-statement fixes.
			{
				Code:    "const b = require(((\"./b\")));\nconst a = require((\"./a\"));",
				Options: alphabetize,
				Output:  []string{"const a = require((\"./a\"));\nconst b = require(((\"./b\")));\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 11}},
			},
			// String-named imports have no identifier name; the diagnostic renders
			// it as "undefined", so aliases decide
			// their alphabetical order. String export aliases likewise sort by the
			// local Identifier rather than the exported literal.
			{
				Code: "import { \"a\" as z, \"z\" as a } from './pkg';\n" +
					"export { z as \"zed\", a as \"aye\" };",
				Options: map[string]any{
					"named":       map[string]any{"import": true, "export": true},
					"alphabetize": map[string]any{"order": "asc"},
				},
				Output: []string{"import { \"z\" as a, \"a\" as z } from './pkg';\n" +
					"export { a as \"aye\", z as \"zed\" };"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Message: "`undefined as a` import should occur before import of `undefined as z`", Line: 1, Column: 20, EndLine: 1, EndColumn: 28},
					{MessageId: "order", Message: "`a` export should occur before export of `z`", Line: 2, Column: 22, EndLine: 2, EndColumn: 32},
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
			// Nested module and outer-source diagnostics remain in document order
			// even when the outer import sequence begins before the namespace and
			// ends after it. The namespace is also an autofix barrier for the outer
			// sequence, while its own import-equals declarations remain fixable.
			{
				Code: "import outerZ = require('./z');\n" +
					"declare namespace N {\n" +
					"  import innerZ = require('./z');\n" +
					"  import innerA = require('./a');\n" +
					"}\n" +
					"import outerA = require('./a');",
				Options: alphabetize,
				Output: []string{"import outerZ = require('./z');\n" +
					"declare namespace N {\n" +
					"  import innerA = require('./a');\n" +
					"  import innerZ = require('./z');\n" +
					"}\n" +
					"import outerA = require('./a');"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Line: 4, Column: 3},
					{MessageId: "order", Line: 6, Column: 1},
				},
			},
		},
	)
}

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

			// ---- Dimension 4: string, numeric, rest, default, and nested binding keys bail out ----
			{Code: `const { "b": b, a } = require("pkg");`, Options: namedRequire},
			{Code: `const { 1: b, a } = require("pkg");`, Options: namedRequire},
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
			// Type arguments are call metadata and do not hide a direct require
			// from module or named-specifier ordering.
			{
				Code: "const sibling = require<unknown>('./z');\n" +
					"const fs = require<unknown>('fs');",
				Options: alphabetize,
				Output: []string{"const fs = require<unknown>('fs');\n" +
					"const sibling = require<unknown>('./z');\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2}},
			},
			{
				Code:    `const { b, a } = require<Record<string, unknown>>("pkg");`,
				Options: namedRequire,
				Output:  []string{`const { a, b } = require<Record<string, unknown>>("pkg");`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 1, Column: 12}},
			},
			// Compound and logical assignments are named CommonJS exports too.
			{
				Code:    "exports.Z += 1;\nmodule.exports.A ??= 1;",
				Options: namedCJS,
				Output:  []string{"module.exports.A ??= 1;\nexports.Z += 1;\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			{
				Code:    `module.exports ||= { b, a };`,
				Options: namedCJS,
				Output:  []string{`module.exports ||= { a, b };`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 1, Column: 25}},
			},
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
			// ---- Dimension 4: a computed identifier binding key remains sortable ----
			{
				Code:    `const { [key]: b, a } = require("pkg");`,
				Options: namedRequire,
				Output:  []string{`const { a, [key]: b } = require("pkg");`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 1, Column: 19}},
			},
			// ---- Dimension 4: computed identifier CJS keys match; literal keys do not ----
			{
				Code:    "exports[B] = 1;\nmodule.exports.A = 1;",
				Options: namedCJS,
				Output:  []string{"module.exports.A = 1;\nexports[B] = 1;\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Dimension 4: a computed identifier object key remains sortable ----
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
			// ---- Dimension 4: source ranges preserve the first-line indentation boundary ----
			{
				Code:    "  import b from 'b';\nimport a from 'a';\n",
				Options: alphabetize,
				Output:  []string{" import a from 'a';\n import b from 'b';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// Movable line ranges inspect at most 100 adjacent comments. The complete
			// slice between imports still retains the 101st during this forward swap.
			{
				Code:    "import b from 'b';" + commentChain + "\nimport a from 'a';\n",
				Options: alphabetize,
				Output:  []string{"import a from 'a';\nimport b from 'b';" + commentChain + "\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
		},
	)
}
