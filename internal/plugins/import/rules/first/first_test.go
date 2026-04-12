package first_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/first"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFirstRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&first.FirstRule,
		[]rule_tester.ValidTestCase{
			// ── Import statement variants ───────────────────────────────────
			// Named imports
			{Code: `import { x } from './foo'; import { y } from './bar'; export { x, y }`},
			// Default import
			{Code: "import a from 'a';\nimport b from 'b';"},
			// Namespace import
			{Code: "import * as ns from 'foo';\nimport { x } from 'bar';"},
			// Side-effect import (no bindings)
			{Code: "import 'foo';\nimport 'bar';"},
			// Mixed import styles
			{Code: "import a from 'a';\nimport { b } from 'b';\nimport * as c from 'c';\nimport 'd';"},
			// Type-only import (TypeScript)
			{Code: "import type { Foo } from './foo';\nimport { bar } from './bar';"},
			// ImportEqualsDeclaration (external module)
			{Code: "import y = require('bar');\nimport { x } from 'foo';"},
			// ImportEqualsDeclaration with multiple imports
			{Code: "import a = require('a');\nimport b = require('b');\nimport { c } from 'c';"},

			// ── Directive handling ──────────────────────────────────────────
			// Single directive before imports
			{Code: "'use strict';\nimport { x } from 'foo';"},
			// Double-quoted directive before imports
			{Code: "\"use strict\";\nimport { x } from 'foo';"},
			// Multiple directives before imports
			{Code: "'use strict';\n'use asm';\nimport { x } from 'foo';"},
			// Directive only, no imports
			{Code: "'use strict';"},

			// ── absolute-first option ──────────────────────────────────────
			// Absolute then relative — ok
			{Code: `import { x } from 'foo'; import { y } from './bar'`},
			// Relative then absolute without option — ok
			{Code: `import { x } from './foo'; import { y } from 'bar'`},
			// disable-absolute-first explicitly
			{Code: `import { x } from './foo'; import { y } from 'bar'`, Options: []interface{}{"disable-absolute-first"}},
			// Only absolutes — ok with absolute-first
			{Code: "import a from 'a';\nimport b from 'b';", Options: []interface{}{"absolute-first"}},
			// Only relatives — ok with absolute-first
			{Code: "import a from './a';\nimport b from './b';", Options: []interface{}{"absolute-first"}},
			// Absolute then relative — ok with absolute-first
			{Code: "import a from 'a';\nimport b from './b';", Options: []interface{}{"absolute-first"}},
			// Scoped package — treated as absolute
			{Code: "import a from '@scope/pkg';\nimport b from './b';", Options: []interface{}{"absolute-first"}},
			// Parent relative import (../) — also relative
			{Code: "import a from 'a';\nimport b from '../b';", Options: []interface{}{"absolute-first"}},

			// ── Edge cases ─────────────────────────────────────────────────
			// Empty file
			{Code: ""},
			// Single import only
			{Code: "import a from 'a';"},
			// Single non-import only (no imports = no error)
			{Code: "var a = 1;"},
			// Dynamic import expression in a var declaration is NOT an import statement
			{Code: "import { x } from 'bar';\nconst a = import('foo');"},
			// Re-export followed by more imports is fine (re-export is non-import code
			// but there is no import AFTER it in this case)
			{Code: "import { x } from 'foo';\nexport { y } from 'bar';"},
			// Import with import attributes
			{Code: "import data from './data.json' with { type: 'json' };\nimport { x } from 'foo';"},
			// absolute-first: rel → abs → rel — second relative doesn't re-trigger
			{Code: "import a from 'abs';\nimport b from './rel1';\nimport c from './rel2';", Options: []interface{}{"absolute-first"}},
			// Internal module reference (import x = Namespace.Y) at top is valid
			{Code: "import x = require('foo');\nimport y from 'bar';"},
			// Empty import bindings
			{Code: "import {} from 'foo';\nimport a from 'bar';"},
			// Multiple legal imports — lastLegalImp tracks the last one
			{Code: "import a from 'a';\nimport b from 'b';\nimport c from 'c';\nvar x = 1;"},
			// Directives then code, no imports — no errors
			{Code: "'use strict';\nvar a = 1;"},
			// export * from is non-import, but no import follows — valid
			{Code: "import a from 'a';\nexport * from 'b';"},
		},
		[]rule_tester.InvalidTestCase{
			// ── Basic misplaced import detection ───────────────────────────
			// 0: Single misplaced import after export
			{
				Code: "import { x } from './foo';\nexport { x };\nimport { y } from './bar';",
				Output: []string{
					"import { x } from './foo';\nimport { y } from './bar';\nexport { x };",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
				},
			},
			// 1: Two misplaced imports after export
			{
				Code: "import { x } from './foo';\nexport { x };\nimport { y } from './bar';\nimport { z } from './baz';",
				Output: []string{
					"import { x } from './foo';\nimport { y } from './bar';\nimport { z } from './baz';\nexport { x };",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
					{MessageId: "first", Line: 4, Column: 1},
				},
			},
			// 2: Import after variable declaration, no previous legal import
			{
				Code: "var a = 1;\nimport { y } from './bar';",
				Output: []string{
					"import { y } from './bar';\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 3: Import immediately after non-import with no whitespace
			{
				Code: "if (true) { console.log(1) }import a from 'b'",
				Output: []string{
					"import a from 'b'\nif (true) { console.log(1) }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 1, Column: 29},
				},
			},

			// ── Directive handling ──────────────────────────────────────────
			// 4: Directive after first import is NOT special — treated as non-import
			{
				Code: "import { x } from 'foo';\n'use directive';\nimport { y } from 'bar';",
				Output: []string{
					"import { x } from 'foo';\nimport { y } from 'bar';\n'use directive';",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
				},
			},

			// ── Import statement variants (misplaced) ──────────────────────
			// 5: Default import after code
			{
				Code: "var a = 1;\nimport x from './foo';",
				Output: []string{
					"import x from './foo';\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 6: Namespace import after code
			{
				Code: "var a = 1;\nimport * as ns from './foo';",
				Output: []string{
					"import * as ns from './foo';\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 7: Side-effect import after code
			{
				Code: "var a = 1;\nimport './foo';",
				Output: []string{
					"import './foo';\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 8: Type-only import after code
			{
				Code: "var a = 1;\nimport type { Foo } from './foo';",
				Output: []string{
					"import type { Foo } from './foo';\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 9: ImportEqualsDeclaration after code
			{
				Code: "var a = 1;\nimport x = require('./foo');",
				Output: []string{
					"import x = require('./foo');\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},

			// ── Interleaved imports and code ───────────────────────────────
			// 10: Multiple non-import statements between imports
			{
				Code: "import a from 'a';\nvar x = 1;\nvar y = 2;\nimport b from 'b';",
				Output: []string{
					"import a from 'a';\nimport b from 'b';\nvar x = 1;\nvar y = 2;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 4, Column: 1},
				},
			},
			// 11: Import after function declaration
			{
				Code: "import a from 'a';\nfunction foo() {}\nimport b from 'b';",
				Output: []string{
					"import a from 'a';\nimport b from 'b';\nfunction foo() {}",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
				},
			},
			// 12: Import after class declaration
			{
				Code: "import a from 'a';\nclass Foo {}\nimport b from 'b';",
				Output: []string{
					"import a from 'a';\nimport b from 'b';\nclass Foo {}",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
				},
			},
			// 13: Import after if statement
			{
				Code: "import a from 'a';\nif (true) { a(); }\nimport b from 'b';",
				Output: []string{
					"import a from 'a';\nimport b from 'b';\nif (true) { a(); }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
				},
			},
			// 14: Alternating import/code/import/code/import
			{
				Code: "import a from 'a';\nvar x = 1;\nimport b from 'b';\nvar y = 2;\nimport c from 'c';",
				Output: []string{
					"import a from 'a';\nimport b from 'b';\nimport c from 'c';\nvar x = 1;\nvar y = 2;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
					{MessageId: "first", Line: 5, Column: 1},
				},
			},
			// 15: Multiple misplaced imports, no legal import at top
			{
				Code: "var x = 1;\nimport a from 'a';\nimport b from 'b';",
				Output: []string{
					"import a from 'a';\nimport b from 'b';\nvar x = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
					{MessageId: "first", Line: 3, Column: 1},
				},
			},

			// ── absolute-first option ──────────────────────────────────────
			// 16: Relative then absolute with absolute-first
			{
				Code:    "import { x } from './foo'; import { y } from 'bar'",
				Options: []interface{}{"absolute-first"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "absolute", Line: 1, Column: 46},
				},
			},
			// 17: absolute-first with ImportEqualsDeclaration
			{
				Code:    "import { x } from './foo';\nimport y = require('bar');",
				Options: []interface{}{"absolute-first"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "absolute", Line: 2, Column: 12},
				},
			},
			// 18: absolute-first — multiple absolute imports after relative
			{
				Code:    "import a from './a';\nimport b from 'b';\nimport c from 'c';",
				Options: []interface{}{"absolute-first"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "absolute", Line: 2, Column: 15},
					{MessageId: "absolute", Line: 3, Column: 15},
				},
			},
			// 19: absolute-first — relative then scoped package
			{
				Code:    "import a from './a';\nimport b from '@scope/pkg';",
				Options: []interface{}{"absolute-first"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "absolute", Line: 2, Column: 15},
				},
			},
			// 20: absolute-first — parent relative (..) is still relative
			{
				Code:    "import a from '../a';\nimport b from 'b';\nimport c from './c';",
				Options: []interface{}{"absolute-first"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "absolute", Line: 2, Column: 15},
				},
			},

			// ── Combined: misplaced + absolute-first ───────────────────────
			// 21: Both absolute-first and misplaced errors fire simultaneously
			{
				Code:    "import a from './a';\nvar x = 1;\nimport b from 'b';",
				Options: []interface{}{"absolute-first"},
				Output: []string{
					"import a from './a';\nimport b from 'b';\nvar x = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "absolute", Line: 3, Column: 15},
					{MessageId: "first", Line: 3, Column: 1},
				},
			},

			// ── Non-obvious non-import statements ──────────────────────────
			// 22: Dynamic import() in variable declaration is NOT an import statement
			{
				Code: "const a = import('foo');\nimport { x } from 'bar';",
				Output: []string{
					"import { x } from 'bar';\nconst a = import('foo');",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 23: Re-export (export...from) is NOT an import — import after it is misplaced
			{
				Code: "import { x } from 'foo';\nexport { y } from 'bar';\nimport { z } from 'baz';",
				Output: []string{
					"import { x } from 'foo';\nimport { z } from 'baz';\nexport { y } from 'bar';",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
				},
			},

			// ── First-expression edge cases ────────────────────────────────
			// 24: Template literal as first expression is NOT a directive
			{
				Code: "`use strict`;\nimport a from 'a';",
				Output: []string{
					"import a from 'a';\n`use strict`;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 25: Empty statement (;) is NOT a directive — counts as non-import
			{
				Code: ";\nimport a from 'a';",
				Output: []string{
					"import a from 'a';\n;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 26: export default before import
			{
				Code: "export default 42;\nimport a from 'a';",
				Output: []string{
					"import a from 'a';\nexport default 42;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},

			// ── Directive + misplaced import interaction ───────────────────
			// 27: Directive at top, code, then import — import moves before body[0] (the directive)
			// because there is no lastLegalImp, fix inserts before body[0].
			{
				Code: "'use strict';\nvar a = 1;\nimport b from 'b';",
				Output: []string{
					"import b from 'b';\n'use strict';\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
				},
			},

			// ── absolute-first edge cases ──────────────────────────────────
			// 28: absolute-first: rel → abs → rel — abs reported, second rel not
			{
				Code:    "import a from './a';\nimport b from 'b';\nimport c from './c';",
				Options: []interface{}{"absolute-first"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "absolute", Line: 2, Column: 15},
				},
			},

			// ── Autofix edge cases ─────────────────────────────────────────
			// ── export * from as non-import ────────────────────────────────
			// 29: export * re-export is non-import — import after it is misplaced
			{
				Code: "import a from 'a';\nexport * from 'b';\nimport c from 'c';",
				Output: []string{
					"import a from 'a';\nimport c from 'c';\nexport * from 'b';",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 3, Column: 1},
				},
			},

			// ── Empty import bindings ──────────────────────────────────────
			// 30: Empty import {} after code
			{
				Code: "var a = 1;\nimport {} from 'foo';",
				Output: []string{
					"import {} from 'foo';\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},

			// ── Import attributes ──────────────────────────────────────────
			// 31: Misplaced import with `with` clause — fix must move entire statement
			{
				Code: "var a = 1;\nimport data from './data.json' with { type: 'json' };",
				Output: []string{
					"import data from './data.json' with { type: 'json' };\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},

			// ── Comments between statements ────────────────────────────────
			// 32: Comment between code and misplaced import — preserved in fix
			{
				Code: "import a from 'a';\nvar x = 1;\n// this import is misplaced\nimport b from 'b';",
				Output: []string{
					"import a from 'a';\n// this import is misplaced\nimport b from 'b';\nvar x = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 4, Column: 1},
				},
			},

			// ── Multiple legal imports then misplaced ──────────────────────
			// 33: Three legal imports, code, then misplaced — fix inserts after 3rd legal
			{
				Code: "import a from 'a';\nimport b from 'b';\nimport c from 'c';\nvar x = 1;\nimport d from 'd';",
				Output: []string{
					"import a from 'a';\nimport b from 'b';\nimport c from 'c';\nimport d from 'd';\nvar x = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 5, Column: 1},
				},
			},

			// ── Autofix edge cases ─────────────────────────────────────────
			// 34: Moving import with complex multiline code between
			{
				Code: "import a from 'a';\nconst obj = {\n  key: 'value',\n};\nimport b from 'b';",
				Output: []string{
					"import a from 'a';\nimport b from 'b';\nconst obj = {\n  key: 'value',\n};",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 5, Column: 1},
				},
			},
			// 35: Moving ImportEqualsDeclaration
			{
				Code: "var a = 1;\nimport x = require('foo');",
				Output: []string{
					"import x = require('foo');\nvar a = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},

			// ── Multiline import ───────────────────────────────────────────
			// 36: Multiline destructured import misplaced — entire statement moves
			{
				Code: "var x = 1;\nimport {\n  a,\n  b,\n  c\n} from 'foo';",
				Output: []string{
					"import {\n  a,\n  b,\n  c\n} from 'foo';\nvar x = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 37: Multiline comment before misplaced import — comment moves with it
			{
				Code: "import a from 'a';\nvar x = 1;\n/**\n * Module B\n */\nimport b from 'b';",
				Output: []string{
					"import a from 'a';\n/**\n * Module B\n */\nimport b from 'b';\nvar x = 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 6, Column: 1},
				},
			},

			// ── Multiple non-imports before first import ───────────────────
			// 38: Several non-import statements then import — all code stays, import moves to top
			{
				Code: "var a = 1;\nvar b = 2;\nvar c = 3;\nimport x from 'x';",
				Output: []string{
					"import x from 'x';\nvar a = 1;\nvar b = 2;\nvar c = 3;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 4, Column: 1},
				},
			},

			// ── Reference checking (shouldSort) ───────────────────────────
			// The first misplaced import always gets a fix (lastSortNodesIndex
			// starts at 0).  The reference check only prevents BATCH fixing:
			// once shouldSort becomes false, later imports are excluded from the
			// current fix pass and handled in subsequent passes.

			// 39: Imported variable referenced before import — only first import
			// is fixed per pass.  Three fix passes are required.
			{
				Code: "var a = 1;\nimport { y } from './bar';\nif (true) { x() };\nimport { x } from './foo';\nimport { z } from './baz';",
				Output: []string{
					// Pass 1: import { y } moved to top (y not referenced before it)
					"import { y } from './bar';\nvar a = 1;\nif (true) { x() };\nimport { x } from './foo';\nimport { z } from './baz';",
					// Pass 2: import { x } moved (first misplaced always fixable)
					"import { y } from './bar';\nimport { x } from './foo';\nvar a = 1;\nif (true) { x() };\nimport { z } from './baz';",
					// Pass 3: import { z } moved
					"import { y } from './bar';\nimport { x } from './foo';\nimport { z } from './baz';\nvar a = 1;\nif (true) { x() };",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
					{MessageId: "first", Line: 4, Column: 1},
					{MessageId: "first", Line: 5, Column: 1},
				},
			},
			// 40: No reference before import — all imports are fixable in one pass
			{
				Code: "var a = 1;\nimport { y } from './bar';\nvar b = 2;\nimport { z } from './baz';",
				Output: []string{
					"import { y } from './bar';\nimport { z } from './baz';\nvar a = 1;\nvar b = 2;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
					{MessageId: "first", Line: 4, Column: 1},
				},
			},
			// 41: Side-effect import (no bindings) — no references to check, always sortable
			{
				Code: "var a = 1;\nimport './side-effect';\nvar b = 2;\nimport { x } from './foo';",
				Output: []string{
					"import './side-effect';\nimport { x } from './foo';\nvar a = 1;\nvar b = 2;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
					{MessageId: "first", Line: 4, Column: 1},
				},
			},
			// 42: Default import referenced before import — first import always
			// gets fix; second import fixed in pass 2.
			{
				Code: "console.log(foo);\nimport foo from './foo';\nimport bar from './bar';",
				Output: []string{
					// Pass 1: import foo moved (first misplaced always fixable),
					// import bar has no fix (shouldSort=false).
					"import foo from './foo';\nconsole.log(foo);\nimport bar from './bar';",
					// Pass 2: import bar moved after import foo.
					"import foo from './foo';\nimport bar from './bar';\nconsole.log(foo);",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
					{MessageId: "first", Line: 3, Column: 1},
				},
			},
			// 43: Namespace import referenced before import — still fixed (first always fixable)
			{
				Code: "ns.doSomething();\nimport * as ns from './mod';",
				Output: []string{
					"import * as ns from './mod';\nns.doSomething();",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 44: Reference in nested function — still fixed (first always fixable)
			{
				Code: "function setup() { x(); }\nimport { x } from './foo';",
				Output: []string{
					"import { x } from './foo';\nfunction setup() { x(); }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 45: Type annotation with same name does NOT suppress fix (symbol-based check)
			{
				Code: "const x: Foo = {} as Foo;\nimport type { Foo } from './foo';",
				Output: []string{
					"import type { Foo } from './foo';\nconst x: Foo = {} as Foo;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 46: Declaration name with same name does NOT suppress fix
			{
				Code: "function foo() {}\nimport { foo } from './mod';",
				Output: []string{
					"import { foo } from './mod';\nfunction foo() {}",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 47: Renamed import — only local name matters, not original name
			{
				Code: "console.log(bar);\nimport { foo as bar } from './mod';",
				Output: []string{
					// bar IS referenced → but first import always fixable
					"import { foo as bar } from './mod';\nconsole.log(bar);",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 48: Multiple bindings, one referenced — suppresses fix for all
			{
				Code: "var v = 1;\nimport { a } from './a';\nconsole.log(b);\nimport { b, c } from './bc';",
				Output: []string{
					// Pass 1: import { a } moved (a not referenced),
					// import { b, c } not moved (b is referenced before it)
					"import { a } from './a';\nvar v = 1;\nconsole.log(b);\nimport { b, c } from './bc';",
					// Pass 2: import { b, c } moved (first misplaced always fixable)
					"import { a } from './a';\nimport { b, c } from './bc';\nvar v = 1;\nconsole.log(b);",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
					{MessageId: "first", Line: 4, Column: 1},
				},
			},
			// 49: Default + named import — default name referenced
			{
				Code: "foo();\nimport foo, { bar } from './mod';",
				Output: []string{
					// First always fixable despite reference
					"import foo, { bar } from './mod';\nfoo();",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
			// 50: Reference in various expression positions
			{
				Code: "var v = 1;\nimport { a } from './a';\nvar r = x + 1;\nimport { x } from './x';",
				Output: []string{
					// a not referenced → sortable; x referenced in binary expr → shouldSort=false
					"import { a } from './a';\nvar v = 1;\nvar r = x + 1;\nimport { x } from './x';",
					// pass 2
					"import { a } from './a';\nimport { x } from './x';\nvar v = 1;\nvar r = x + 1;",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
					{MessageId: "first", Line: 4, Column: 1},
				},
			},
			// 51: Parameter shadows imported name — should NOT suppress fix
			{
				Code: "function bar(x: number) { return x; }\nimport { x } from './mod';",
				Output: []string{
					"import { x } from './mod';\nfunction bar(x: number) { return x; }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "first", Line: 2, Column: 1},
				},
			},
		},
	)
}
