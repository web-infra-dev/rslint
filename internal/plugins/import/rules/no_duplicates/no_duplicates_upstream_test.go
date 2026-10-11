package no_duplicates_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_duplicates"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicatesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_duplicates.NoDuplicatesRule,
		[]rule_tester.ValidTestCase{
			// --- Different modules ---
			{Code: `import { x } from './foo'; import { y } from './bar'`},
			{Code: `import foo from "module-a"; import { bar } from "module-b"`},

			// --- Namespace + named from same module (cannot be merged into one line) ---
			{Code: `import * as ns from './foo'; import {y} from './foo'`},
			{Code: `import {y} from './foo'; import * as ns from './foo'`},

			// --- Type import + value import (separate categories when not prefer-inline) ---
			{Code: `import type { x } from './foo'; import y from './foo'`},
			{Code: `import type x from './foo'; import type y from './bar'`},
			{Code: `import type {x} from './foo'; import type {y} from './bar'`},
			// Type default + type named from same module → different categories
			{Code: `import type x from './foo'; import type {y} from './foo'`},
			// Empty type import + regular import from different modules
			{Code: "import type {} from './module';\nimport {} from './module2';"},

			// --- considerQueryString option ---
			{
				Code:    `import x from './bar?optionX'; import y from './bar?optionY';`,
				Options: map[string]interface{}{"considerQueryString": true},
			},

			// --- Inline type specifier + value import (not prefer-inline) ---
			{Code: `import { type x } from './foo'; import y from './foo'`},
			{Code: `import { type x } from './foo'; import { y } from './foo'`},
			// Inline type + type import from different module
			{Code: `import { type x } from './foo'; import type y from 'bar'`},

			// --- Single import (no duplicate) ---
			{Code: `import { x } from './foo'`},
			{Code: `import './foo'`},
			{Code: `import type { x } from './foo'`},

			// --- declare module scoping ---
			// Top-level + declare module imports are separate scopes → not duplicates
			{Code: "import type { Identifier } from 'module';\n\ndeclare module 'module2' {\n  import type { Identifier } from 'module';\n}"},
			// Two different declare module blocks → separate scopes
			{Code: "declare module 'a' {\n  import { x } from 'bar';\n}\ndeclare module 'b' {\n  import { x } from 'bar';\n}"},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Basic merge scenarios
			// ============================================================

			// 0: Two named imports
			{
				Code:   `import { x } from './foo'; import { y } from './foo'`,
				Output: []string{`import { x , y } from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 19},
					{MessageId: "noDuplicates", Line: 1, Column: 46},
				},
			},
			// 1: Three-way merge
			{
				Code:   `import {x} from './foo'; import {y} from './foo'; import { z } from './foo'`,
				Output: []string{`import {x,y, z } from './foo';  `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 42},
					{MessageId: "noDuplicates", Line: 1, Column: 69},
				},
			},
			// 2: Side-effect + named
			{
				Code:   `import './foo'; import {x} from './foo'`,
				Output: []string{`import {x} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 8},
					{MessageId: "noDuplicates", Line: 1, Column: 33},
				},
			},
			// 3: Side-effect + default
			{
				Code:   `import './foo'; import def from './foo'`,
				Output: []string{`import def from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 8},
					{MessageId: "noDuplicates", Line: 1, Column: 33},
				},
			},
			// 4: Default + named
			{
				Code:   `import def from './foo'; import {x} from './foo'`,
				Output: []string{`import def, {x} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 42},
				},
			},
			// 5: Named + default (reverse order)
			{
				Code:   `import {x} from './foo'; import def from './foo'`,
				Output: []string{`import def, {x} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 42},
				},
			},
			// 6: Duplicate side-effect imports
			{
				Code:   `import './foo'; import './foo'`,
				Output: []string{`import './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 8},
					{MessageId: "noDuplicates", Line: 1, Column: 24},
				},
			},
			// 7: Named + empty braces
			{
				Code:   `import {x} from './foo'; import {} from './foo'`,
				Output: []string{`import {x} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 41},
				},
			},
			// 8: Empty braces + named (reverse)
			{
				Code:   `import { } from './foo'; import {x} from './foo'`,
				Output: []string{`import { x} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 42},
				},
			},
			// 9: Side-effect + default,named combo
			{
				Code:   `import './foo'; import def, {x} from './foo'`,
				Output: []string{`import def, {x} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 8},
					{MessageId: "noDuplicates", Line: 1, Column: 38},
				},
			},
			// 10: Named + default,named combo
			{
				Code:   `import {x} from './foo'; import def, {y} from './foo'`,
				Output: []string{`import def, {x,y} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 47},
				},
			},

			// ============================================================
			// Autofix bail scenarios
			// ============================================================

			// 11: Different default names → no autofix
			{
				Code: `import foo from 'non-existent'; import bar from 'non-existent';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 49},
				},
			},
			// 12: Namespace imports cannot be merged
			{
				Code: `import * as ns1 from './foo'; import * as ns2 from './foo'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 52},
				},
			},

			// ============================================================
			// Comment-related autofix bail
			// ============================================================

			// 13: Comment before second import → bail on second (rest has comment)
			{
				Code: "import {x} from './foo'\n// comment\nimport {y} from './foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 3, Column: 17},
				},
			},
			// 14: Trailing comment on first import → bail on all (first has comment)
			{
				Code: "import {x} from './foo' // line comment\nimport {y} from './foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 2, Column: 17},
				},
			},
			// 15: Trailing comment on second import → bail on second
			{
				Code: "import {x} from './foo'\nimport {y} from './foo' // line comment",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 2, Column: 17},
				},
			},
			// 16: Block comment before second import → bail
			{
				Code: "import {x} from './foo'\n/* comment */ import {y} from './foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 2, Column: 31},
				},
			},
			// 17: Comment inside non-specifiers: `import/* c */{y}`
			{
				Code: "import {x} from './foo'\nimport/* comment */{y} from './foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 2, Column: 29},
				},
			},
			// 18: Comment between `from` and path: `import{y}from/* c */'./foo'`
			{
				Code: "import {x} from './foo'\nimport{y}from/* comment */'./foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 2, Column: 27},
				},
			},
			// 19: Comment between `from` and module path on separate lines → bail on first
			{
				Code: "import {x} from\n// some-tool-disable-next-line\n'./foo'\nimport {y} from './foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 3, Column: 1},
					{MessageId: "noDuplicates", Line: 4, Column: 17},
				},
			},
			// 20: Comment after all imports does NOT bail autofix
			{
				Code:   "import {x} from './foo'\nimport {y} from './foo'\n// some-tool-disable-next-line",
				Output: []string{"import {x,y} from './foo'\n// some-tool-disable-next-line"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 2, Column: 17},
				},
			},
			// 20: Comment separated by blank line from import does NOT bail
			{
				Code:   "import {x} from './foo'\n// comment\n\nimport {y} from './foo'",
				Output: []string{"import {x,y} from './foo'\n// comment\n\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 4, Column: 17},
				},
			},

			// ============================================================
			// Whitespace edge cases
			// ============================================================

			// 21: No space after import keyword
			{
				Code:   `import'./foo'; import {x} from './foo'`,
				Output: []string{`import {x} from'./foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 7},
					{MessageId: "noDuplicates", Line: 1, Column: 32},
				},
			},
			// 22: No space before braces
			{
				Code:   `import{x} from './foo'; import def from './foo'`,
				Output: []string{`import def,{x} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 16},
					{MessageId: "noDuplicates", Line: 1, Column: 41},
				},
			},

			// ============================================================
			// Multiline imports
			// ============================================================

			// 23: Multiline with trailing newline removal (#2027)
			{
				Code:   "import { Foo } from './foo';\nimport { Bar } from './foo';\nexport const value = {}",
				Output: []string{"import { Foo , Bar } from './foo';\nexport const value = {}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 21},
					{MessageId: "noDuplicates", Line: 2, Column: 21},
				},
			},
			// 24: Multiline with default import merge (#2027)
			{
				Code:   "import { Foo } from './foo';\nimport Bar from './foo';\nexport const value = {}",
				Output: []string{"import Bar, { Foo } from './foo';\nexport const value = {}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 21},
					{MessageId: "noDuplicates", Line: 2, Column: 17},
				},
			},
			// 25: Multi-line named imports with trailing commas
			{
				Code:   "import {A1,} from 'foo';\nimport {B1,} from 'foo';\nimport {C1,} from 'foo';",
				Output: []string{"import {A1,B1,C1} from 'foo';\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 19},
					{MessageId: "noDuplicates", Line: 2, Column: 19},
					{MessageId: "noDuplicates", Line: 3, Column: 19},
				},
			},

			// ============================================================
			// Namespace mixed scenarios
			// ============================================================

			// 26: Namespace + named + named: partial merge (named merged, ns untouched)
			{
				Code:   `import * as ns from './foo'; import {x} from './foo'; import {y} from './foo'`,
				Output: []string{`import * as ns from './foo'; import {x,y} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 46},
					{MessageId: "noDuplicates", Line: 1, Column: 71},
				},
			},
			// 27: Named + namespace + named + side-effect: merge named+side-effect, skip ns
			// imported map has: {x}, {y}, './foo' (side-effect) → 3 errors
			{
				Code:   "import {x} from './foo'; import * as ns from './foo'; import {y} from './foo'; import './foo'",
				Output: []string{"import {x,y} from './foo'; import * as ns from './foo';  "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 71},
					{MessageId: "noDuplicates", Line: 1, Column: 87},
				},
			},

			// ============================================================
			// Query strings
			// ============================================================

			// 28: Without considerQueryString, query strings stripped
			{
				Code: `import x from './bar?optionX'; import y from './bar?optionY';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 15},
					{MessageId: "noDuplicates", Line: 1, Column: 46},
				},
			},

			// ============================================================
			// TypeScript type-only imports
			// ============================================================

			// 29: Duplicate type-only named imports
			{
				Code:   `import type {x} from './foo'; import type {y} from './foo'`,
				Output: []string{`import type {x,y} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 52},
				},
			},
			// 30: Duplicate type-only default imports (different names → no autofix)
			{
				Code: `import type x from './foo'; import type y from './foo'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 20},
					{MessageId: "noDuplicates", Line: 1, Column: 48},
				},
			},
			// 31: Duplicate type-only default imports with SAME name → autofix removes dup
			{
				Code:   `import type x from './foo'; import type x from './foo'`,
				Output: []string{`import type x from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 20},
					{MessageId: "noDuplicates", Line: 1, Column: 48},
				},
			},
			// 32: Inline type + type-only without prefer-inline → grouped as namedTypes
			{
				Code:   `import {type x} from './foo'; import type {y} from './foo'`,
				Output: []string{`import {type x,y} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 52},
				},
			},
			// 33: Inline type imports without prefer-inline
			{
				Code:   `import {type x} from './foo'; import {type y} from './foo'`,
				Output: []string{`import {type x,type y} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 52},
				},
			},

			// ============================================================
			// prefer-inline option
			// ============================================================

			// 34: Value + type import
			{
				Code:    `import {AValue} from './foo'; import type {AType} from './foo'`,
				Options: map[string]interface{}{"prefer-inline": true},
				Output:  []string{`import {AValue,type AType} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 56},
				},
			},
			// 35: Inline type + type-only import
			{
				Code:    `import {type x} from 'foo'; import type {y} from 'foo'`,
				Options: map[string]interface{}{"prefer-inline": true},
				Output:  []string{`import {type x,type y} from 'foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 50},
				},
			},
			// 36: Type-only + inline type (reverse)
			{
				Code:    `import type {x} from 'foo'; import {type y} from 'foo'`,
				Options: map[string]interface{}{"prefer-inline": true},
				Output:  []string{`import {type x,type y} from 'foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 50},
				},
			},
			// 37: Both inline type
			{
				Code:    `import {type x} from './foo'; import {type y} from './foo'`,
				Options: map[string]interface{}{"prefer-inline": true},
				Output:  []string{`import {type x,type y} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 52},
				},
			},
			// 38: Mixed value + inline type + type import
			{
				Code:    `import {AValue, type x, BValue} from './foo'; import {type y} from './foo'`,
				Options: map[string]interface{}{"prefer-inline": true},
				Output:  []string{`import {AValue, type x, BValue,type y} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 38},
					{MessageId: "noDuplicates", Line: 1, Column: 68},
				},
			},
			// 39: prefer-inline: fix should not corrupt imports named 'from' (#3224)
			{
				Code:    `import type {Observable} from './foo'; import {from} from './foo'`,
				Options: map[string]interface{}{"prefer-inline": true},
				Output:  []string{`import {type Observable,from} from './foo'; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 31},
					{MessageId: "noDuplicates", Line: 1, Column: 59},
				},
			},

			// ============================================================
			// 4+ duplicate imports
			// ============================================================

			// 40: Four named imports from same module
			{
				Code:   "import {a} from './foo'; import {b} from './foo'; import {c} from './foo'; import {d} from './foo'",
				Output: []string{"import {a,b,c,d} from './foo';   "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 42},
					{MessageId: "noDuplicates", Line: 1, Column: 67},
					{MessageId: "noDuplicates", Line: 1, Column: 92},
				},
			},

			// ============================================================
			// Identifier deduplication
			// ============================================================

			// 41: Duplicate identifier across imports → deduplicated in merge
			{
				Code:   "import {a,b} from './foo'; import { b, c } from './foo'",
				Output: []string{"import {a,b, c } from './foo'; "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 19},
					{MessageId: "noDuplicates", Line: 1, Column: 49},
				},
			},

			// ============================================================
			// Renamed specifiers
			// ============================================================

			// 42: Renamed specifiers preserved during merge
			{
				Code:   "import { x as a } from './foo'; import { y as b } from './foo'",
				Output: []string{"import { x as a , y as b } from './foo'; "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 24},
					{MessageId: "noDuplicates", Line: 1, Column: 56},
				},
			},

			// ============================================================
			// Mixed scenarios with default + namespace + named
			// ============================================================

			// 43: Side-effect + default + named all in `imported` → 3-way merge
			{
				Code:   "import './foo'; import def from './foo'; import {x} from './foo'",
				Output: []string{"import def, {x} from './foo';  "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 8},
					{MessageId: "noDuplicates", Line: 1, Column: 33},
					{MessageId: "noDuplicates", Line: 1, Column: 58},
				},
			},

			// ============================================================
			// prefer-inline: type-only first + value second
			// ============================================================

			// 44: prefer-inline: first is type-only named, second is value named
			{
				Code:    "import type {A} from './foo'; import {B} from './foo'",
				Options: map[string]interface{}{"prefer-inline": true},
				Output:  []string{"import {type A,B} from './foo'; "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 47},
				},
			},

			// 45: prefer-inline: first is value named, second is type-only with multiple specifiers
			{
				Code:    "import {A} from './foo'; import type {B, C} from './foo'",
				Options: map[string]interface{}{"prefer-inline": true},
				Output:  []string{"import {A,type B,type C} from './foo'; "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 50},
				},
			},

			// ============================================================
			// Double-quoted and backtick module specifiers
			// ============================================================

			// ============================================================
			// Comments inside braces (preserved during merge)
			// ============================================================

			// 46: Block comment inside braces is preserved during merge
			{
				Code:   "import { x /* comment */ } from './foo'; import {y} from './foo'",
				Output: []string{"import { x /* comment */ ,y} from './foo'; "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 33},
					{MessageId: "noDuplicates", Line: 1, Column: 58},
				},
			},
			// 47: 4-way with empty braces + comment-containing braces + named
			{
				Code:   "import {x} from './foo'; import {} from './foo'; import {/*c*/} from './foo'; import {y} from './foo'",
				Output: []string{"import {x/*c*/,y} from './foo';   "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates"},
					{MessageId: "noDuplicates"},
					{MessageId: "noDuplicates"},
					{MessageId: "noDuplicates"},
				},
			},

			// ============================================================
			// Renamed specifiers
			// ============================================================

			// 49: Renamed specifiers preserved during merge
			{
				Code:   "import { x as a } from './foo'; import { y as b } from './foo'",
				Output: []string{"import { x as a , y as b } from './foo'; "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 24},
					{MessageId: "noDuplicates", Line: 1, Column: 56},
				},
			},
			// 50: Renamed specifier NOT deduplicated with plain specifier
			// `{x}` and `{x as y}` are different specifier strings → both kept
			{
				Code:   "import {x} from './foo'; import { x as y } from './foo'",
				Output: []string{"import {x, x as y } from './foo'; "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 49},
				},
			},

			// ============================================================
			// Both imports have default + named
			// ============================================================

			// 51: Same default + different named → merge named into first
			{
				Code:   "import def, {x} from './foo'; import def, {y} from './foo'",
				Output: []string{"import def, {x,y} from './foo'; "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 22},
					{MessageId: "noDuplicates", Line: 1, Column: 52},
				},
			},
			// 52: Different defaults + named → no autofix (conflicting defaults)
			{
				Code: "import def1, {x} from './foo'; import def2, {y} from './foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 23},
					{MessageId: "noDuplicates", Line: 1, Column: 54},
				},
			},

			// ============================================================
			// Mixed scenarios
			// ============================================================

			// 53: Side-effect + default + named → 3-way merge
			{
				Code:   "import './foo'; import def from './foo'; import {x} from './foo'",
				Output: []string{"import def, {x} from './foo';  "},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 8},
					{MessageId: "noDuplicates", Line: 1, Column: 33},
					{MessageId: "noDuplicates", Line: 1, Column: 58},
				},
			},

			// ============================================================
			// Double-quoted module paths
			// ============================================================

			// 54: Double-quoted module paths
			{
				Code:   `import {x} from "./foo"; import {y} from "./foo"`,
				Output: []string{`import {x,y} from "./foo"; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 42},
				},
			},

			// ============================================================
			// Type namespace imports
			// ============================================================

			// 55: Duplicate `import type * as ns` → reported, no autofix (namespace can't merge)
			{
				Code: "import type * as ns1 from './foo'; import type * as ns2 from './foo'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 27},
					{MessageId: "noDuplicates", Line: 1, Column: 62},
				},
			},

			// ============================================================
			// Multiple duplicate groups (tests deterministic ordering)
			// ============================================================

			// ============================================================
			// declare module scope
			// ============================================================

			// 57: Duplicates INSIDE a declare module block → reported + merged
			{
				Code:   "declare module 'foo' {\n  import { x } from 'bar';\n  import { y } from 'bar';\n}",
				Output: []string{"declare module 'foo' {\n  import { x , y } from 'bar';\n  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 2, Column: 21},
					{MessageId: "noDuplicates", Line: 3, Column: 21},
				},
			},

			// 58: Two different modules each with duplicates → errors in document order
			{
				Code: "import {a} from './foo'; import {b} from './bar'; import {c} from './foo'; import {d} from './bar'",
				// Fixes are applied in two passes: first ./foo merge, then ./bar merge.
				Output: []string{
					"import {a,c} from './foo'; import {b} from './bar';  import {d} from './bar'",
					"import {a,c} from './foo'; import {b,d} from './bar';  ",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					// ./foo group first (appears earlier in source)
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 67},
					// ./bar group second
					{MessageId: "noDuplicates", Line: 1, Column: 42},
					{MessageId: "noDuplicates", Line: 1, Column: 92},
				},
			},
		},
	)
}
