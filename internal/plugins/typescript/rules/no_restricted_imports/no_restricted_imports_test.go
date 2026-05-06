package no_restricted_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Test suite mirrors typescript-eslint's tests/rules/no-restricted-imports.test.ts
// (https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-restricted-imports.test.ts).
//
// The TypeScript variant only adds the `allowTypeImports` schema option; every other
// behavior comes from the base ESLint rule. The base rule's full test suite already
// lives in internal/rules/no_restricted_imports/no_restricted_imports_test.go, so this
// file focuses on the TS-specific surface — type-only imports/exports, inline type
// specifiers, and `import = require(...)` forms.
func TestNoRestrictedImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedImportsRule,
		[]rule_tester.ValidTestCase{
			// ---- Baseline: no options or non-matching configuration ----
			{Code: `import foo from 'foo';`},
			{Code: `import foo = require('foo');`},
			{Code: `import 'foo';`},
			{Code: `import foo from 'foo';`, Options: []interface{}{"import1", "import2"}},
			{Code: `import foo = require('foo');`, Options: []interface{}{"import1", "import2"}},
			{Code: `export { foo } from 'foo';`, Options: []interface{}{"import1", "import2"}},
			{Code: `import foo from 'foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{"import1", "import2"},
			}}},
			{Code: `export { foo } from 'foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{"import1", "import2"},
			}}},
			{Code: `import 'foo';`, Options: []interface{}{"import1", "import2"}},
			{
				Code: `import foo from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths":    []interface{}{"import1", "import2"},
					"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
				}},
			},
			{
				Code: `export { foo } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths":    []interface{}{"import1", "import2"},
					"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
				}},
			},
			{
				Code: `import foo from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "import-foo", "message": "Please use import-bar instead."},
						map[string]interface{}{"name": "import-baz", "message": "Please use import-quux instead."},
					},
				}},
			},
			{
				Code: `export { foo } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "import-foo", "message": "Please use import-bar instead."},
						map[string]interface{}{"name": "import-baz", "message": "Please use import-quux instead."},
					},
				}},
			},
			{
				Code: `import foo from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"importNames": []interface{}{"Bar"},
							"message":     "Please use Bar from /import-bar/baz/ instead.",
							"name":        "import-foo",
						},
					},
				}},
			},
			{
				Code: `export { foo } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"importNames": []interface{}{"Bar"},
							"message":     "Please use Bar from /import-bar/baz/ instead.",
							"name":        "import-foo",
						},
					},
				}},
			},
			{
				Code: `import foo from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{"group": []interface{}{"import1/private/*"}, "message": "usage of import1 private modules not allowed."},
						map[string]interface{}{"group": []interface{}{"import2/*", "!import2/good"}, "message": "import2 is deprecated, except the modules in import2/good."},
					},
				}},
			},
			{
				Code: `export { foo } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{"group": []interface{}{"import1/private/*"}, "message": "usage of import1 private modules not allowed."},
						map[string]interface{}{"group": []interface{}{"import2/*", "!import2/good"}, "message": "import2 is deprecated, except the modules in import2/good."},
					},
				}},
			},
			{
				// import = require where importNames is set but binding is the default-style id, not a member access.
				Code: `import foo = require('foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"importNames": []interface{}{"foo"},
							"message":     "Please use Bar from /import-bar/baz/ instead.",
							"name":        "foo",
						},
					},
				}},
			},

			// ---- allowTypeImports — paths ----
			{
				// Whole declaration is `import type` — exempted via path-level allowTypeImports.
				Code: `import type foo from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"message":          "Please use import-bar instead.",
							"name":             "import-foo",
						},
					},
				}},
			},
			{
				// `import type x = require(...)` — IsTypeOnly on KindImportEqualsDeclaration.
				Code: `import type _ = require('import-foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"message":          "Please use import-bar instead.",
							"name":             "import-foo",
						},
					},
				}},
			},
			{
				Code: `import type { Bar } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar"},
							"message":          "Please use Bar from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
			},
			{
				// `import { type Bar }` — only the specifier is type-only; allowTypeImports on the path skips the named-import check.
				Code: `import { type Bar } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar"},
							"message":          "Please use Bar from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
			},
			{
				Code: `export type { Bar } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar"},
							"message":          "Please use Bar from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
			},
			{
				Code: `export { type Bar } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar"},
							"message":          "Please use Bar from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
			},

			// ---- allowTypeImports — patterns ----
			{
				Code: `import type foo from 'import1/private/bar';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"group":            []interface{}{"import1/private/*"},
							"message":          "usage of import1 private modules not allowed.",
						},
					},
				}},
			},
			{
				Code: `export type { foo } from 'import1/private/bar';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"group":            []interface{}{"import1/private/*"},
							"message":          "usage of import1 private modules not allowed.",
						},
					},
				}},
			},
			{
				// export * is not type-only — but the source doesn't match anyway.
				Code:    `export * from 'foo';`,
				Options: []interface{}{"import1"},
			},
			{
				Code: `import type { MyType } from './types';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"group":            []interface{}{"fail"},
							"message":          "Please do not load from 'fail'.",
						},
					},
				}},
			},
			{
				Code: `
import type { foo } from 'import1/private/bar';
import type { foo } from 'import2/private/bar';
      `,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"group":            []interface{}{"import1/private/*"},
							"message":          "usage of import1 private modules not allowed.",
						},
						map[string]interface{}{
							"allowTypeImports": true,
							"group":            []interface{}{"import2/private/*"},
							"message":          "usage of import2 private modules not allowed.",
						},
					},
				}},
			},
			{
				// Regex matcher with allowTypeImports.
				Code: `
import type { foo } from 'import1/private/bar';
import type { foo } from 'import2/private/bar';
      `,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"message":          "usage of import1 private modules not allowed.",
							"regex":            "import1/.*",
						},
						map[string]interface{}{
							"allowTypeImports": true,
							"message":          "usage of import2 private modules not allowed.",
							"regex":            "import2/.*",
						},
					},
				}},
			},
			{
				// Case-sensitive regex doesn't match (regex looks for [A-Z]+, source is lowercase).
				Code: `import { foo } from 'import1/private';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"caseSensitive":    true,
							"message":          "usage of import1 private modules not allowed.",
							"regex":            "import1/[A-Z]+",
						},
					},
				}},
			},

			// ---- Empty option containers — should not crash ----
			{Code: `import foo from 'foo';`, Options: []interface{}{}},
			{Code: `import foo from 'foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{},
			}}},
			{Code: `import foo from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{},
			}}},
			{Code: `import foo from 'foo';`, Options: []interface{}{map[string]interface{}{
				"paths":    []interface{}{},
				"patterns": []interface{}{},
			}}},

			// ============================================================
			// Extra coverage — tsgo / TS-specific edge cases
			// ============================================================

			// ---- Source-level short-circuit (typescript-eslint divergence from base) ----
			// Two duplicate path entries for the same source, one with allowTypeImports=true.
			// Upstream wrapper: ANY allow-type entry → skip whole type-only declaration.
			// rslint core alone would still report from the strict entry; our wrapper
			// applies the source-level short-circuit so the result matches upstream.
			{
				Code: `import type x from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "foo", "allowTypeImports": true},
						map[string]interface{}{"name": "foo"},
					},
				}},
			},
			{
				// Same short-circuit when one path allows and one pattern matches strictly.
				Code: `import type x from 'lodash';`,
				Options: []interface{}{map[string]interface{}{
					"paths":    []interface{}{map[string]interface{}{"name": "lodash", "allowTypeImports": true}},
					"patterns": []interface{}{"lodash"},
				}},
			},
			{
				// Pattern allow + duplicate strict pattern → still skipped on whole-type-only.
				Code: `import type x from 'lib/internal/x';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{"group": []interface{}{"lib/internal/*"}, "allowTypeImports": true},
						map[string]interface{}{"group": []interface{}{"lib/internal/*"}, "message": "no internals"},
					},
				}},
			},
			{
				// Inline type-only specifier short-circuit through path matching.
				Code: `import { type Bar, type Baz } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "lib", "allowTypeImports": true, "importNames": []interface{}{"Bar", "Baz"}},
					},
				}},
			},

			// ---- Whitespace in source — base trims, so '  foo  ' matches 'foo' ----
			// Whole-type-only import of '  foo  ' trims to 'foo', matches the
			// allow-type path → short-circuit skip.
			{
				Code: `import type x from '  foo  ';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
			},

			// ---- Case-sensitive pattern: uppercase source NOT matched by lowercase glob ----
			{
				Code: `import x from 'FOO';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group":         []interface{}{"foo"},
						"caseSensitive": true,
					}},
				}},
			},

			// ---- Type-only namespace import / export ----
			{
				Code: `import type * as ns from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
			},
			{
				Code: `export type * from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
			},
			{
				Code: `export type * as ns from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
			},

			// ---- Empty named-import / -export braces — 0 specifiers ----
			{
				// `import {} from 'bar'` — no specifiers; not whole-type-only because every() is vacuously true
				// only when length > 0. With 0 specifiers, upstream considers it NOT all-type-only and passes
				// to base. Source 'bar' is NOT in restrictions, so still valid.
				Code:    `import {} from 'bar';`,
				Options: []interface{}{"foo"},
			},
			{
				// `import type {} from 'foo'` — whole-type-only via the clause-level `type`.
				Code: `import type {} from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
			},

			// ---- Default + named, all type-only inline ----
			{
				// Default specifier exists → NOT whole-type-only by upstream's rule (every() requires ImportSpecifier).
				// So this is a value import overall. But the path doesn't match → no error.
				Code:    `import x, { type Y } from 'lib';`,
				Options: []interface{}{"other"},
			},

			// ---- Scoped packages and deep paths ----
			{
				Code: `import type x from '@scope/pkg';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "@scope/pkg", "allowTypeImports": true}},
				}},
			},
			{
				Code: `import type x from '@scope/pkg/sub/path';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"@scope/pkg/**"}, "allowTypeImports": true}},
				}},
			},

			// ---- Relative / absolute paths ----
			{
				Code: `import type x from './local/types';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"./local/*"}, "allowTypeImports": true}},
				}},
			},
			{
				Code: `import type x from '../shared/types';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"../shared/*"}, "allowTypeImports": true}},
				}},
			},

			// ---- Source with hash / query — exact match ----
			{
				Code: `import type x from './foo?raw';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "./foo?raw", "allowTypeImports": true}},
				}},
			},

			// ---- Comments inside import — should not affect source extraction ----
			{
				Code: `import /* foo */ type /* bar */ x /* baz */ from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
			},
			{
				Code: `
import type {
  // a leading comment
  Bar,
  // another
  type Baz,
} from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "allowTypeImports": true, "importNames": []interface{}{"Bar", "Baz"},
					}},
				}},
			},

			// ---- Dynamic import — not an ImportDeclaration, must NOT be checked ----
			{
				Code:    `const m = import('foo');`,
				Options: []interface{}{"foo"},
			},
			{
				Code:    `async function f() { return await import('foo'); }`,
				Options: []interface{}{"foo"},
			},

			// ---- Triple-slash reference — not an ImportDeclaration ----
			{
				Code:    `/// <reference path="foo" />`,
				Options: []interface{}{"foo"},
			},

			// ---- Module declaration is a different node — must NOT be checked ----
			{
				Code:    `declare module 'foo' { export const x: number; }`,
				Options: []interface{}{"foo"},
			},

			// ---- Import attributes (TS 5.3+ `with`, TS 4.5+ `assert`) — sources still match ----
			{
				Code: `import type x from 'foo' with { type: 'json' };`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
			},

			// ---- Real-world: lodash type imports from a value-restricted module ----
			{
				Code: `import type { DebouncedFunc } from 'lodash';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name":             "lodash",
						"allowTypeImports": true,
						"message":          "Use individual lodash submodules at runtime; types are fine.",
					}},
				}},
			},
			// ---- Real-world: type-only re-export from a private barrel ----
			{
				Code: `export type { Internal } from './internal';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group":            []interface{}{"./internal", "./internal/**"},
						"allowTypeImports": true,
						"message":          "Internal modules — types only.",
					}},
				}},
			},

			// ---- Synthesized default specifier: type-only import-equals exempted via short-circuit ----
			{
				// `import type x = require('foo')` is whole-type-only AND source matches an
				// allow-type path → short-circuit skips the entire declaration regardless of
				// the importNames=['default'] match it would otherwise trigger.
				Code: `import type x = require('foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"name":             "foo",
							"importNames":      []interface{}{"default"},
							"allowTypeImports": true,
						},
					},
				}},
			},

			// ============================================================
			// tsgo-specific implementation differences — extra coverage
			// ============================================================

			// ---- String literal source resolution ----
			// tsgo's GetStaticStringValue resolves escape sequences: 'foo' → 'foo'.
			// Upstream's `node.source.value` is also the resolved value (not the raw text).
			// Restriction "foo" only matches the resolved 'foo' source — locked into a valid
			// case where the restriction is something else, so escape resolution doesn't
			// accidentally flip a value test.
			{
				Code:    "import x from '\\u0066oo';",
				Options: []interface{}{"bar"},
			},

			// ---- ImportEquals with qualified name (NOT external) — must not match ----
			// `import x = M.N` is a TypeScript namespace alias, not a require(). The listener
			// must not fire path/pattern checks (no source string to compare).
			{
				Code:    `namespace M { export const N = 1; } import x = M.N;`,
				Options: []interface{}{"M", "M.N", "x"},
			},
			{
				Code:    `namespace M {} import x = M;`,
				Options: []interface{}{"M"},
			},

			// ---- Trailing slash on restriction name does NOT match source without slash ----
			{
				Code:    `import x from 'foo';`,
				Options: []interface{}{"foo/"},
			},

			// ---- name='foo/bar' must NOT match source 'foo' (exact match required for paths) ----
			{
				Code:    `import x from 'foo';`,
				Options: []interface{}{"foo/bar"},
			},

			// ---- name='foo' must NOT match source 'foo/bar' (exact match required) ----
			{
				Code:    `import x from 'foo/bar';`,
				Options: []interface{}{"foo"},
			},

			// ---- importNames: ['*'] applies only to namespace imports, not named ----
			{
				Code: `import { x } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"*"},
					}},
				}},
			},

			// ---- allowImportNames: ['default'] permits the default import ----
			{
				Code: `import x from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "allowImportNames": []interface{}{"default"},
					}},
				}},
			},

			// ---- name with special chars (scoped + dot + dash + underscore) ----
			{
				Code:    `import x from 'baz';`,
				Options: []interface{}{"@my-org/pkg.beta_v1"},
			},

			// ---- import inside an ambient module declaration: still subscribed ----
			// ImportEqualsDeclaration listener fires regardless of nesting; the source
			// 'unrelated' isn't restricted → no error.
			{
				Code: `
declare module 'foo' {
  import x = require('unrelated');
  export const y: typeof x;
}`,
				Options: []interface{}{"restricted"},
			},

			// ---- `as` rename: importNames matches the SOURCE name, not the local binding ----
			{
				// Source name is `Foo`; local is `Bar`. importNames=['Bar'] should NOT match.
				Code: `import { Foo as Bar } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "importNames": []interface{}{"Bar"},
					}},
				}},
			},
			{
				// allowImportNames includes the SOURCE name `Foo` → permitted.
				Code: `import { Foo as Bar } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "allowImportNames": []interface{}{"Foo"},
					}},
				}},
			},

			// ---- `as` rename + inline type-only + allowTypeImports → exempted ----
			{
				// `type Foo as Bar` is a type-only specifier; with allowTypeImports it's skipped.
				Code: `import { type Foo as Bar } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "allowTypeImports": true, "importNames": []interface{}{"Foo"},
					}},
				}},
			},
			{
				// Multiple type-only renames, all exempted via allowTypeImports.
				Code: `import { type A as X, type B as Y } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "allowTypeImports": true, "importNames": []interface{}{"A", "B"},
					}},
				}},
			},
			{
				// All-type-only with rename → whole-import-type-only short-circuit applies.
				Code: `import type { A as X, B as Y } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "allowTypeImports": true, "importNames": []interface{}{"A", "B"},
					}},
				}},
			},

			// ---- export rename + allowTypeImports ----
			{
				// `export { type Foo as Bar }` source name is `Foo`, type-only specifier exempted.
				Code: `export { type Foo as Bar } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "allowTypeImports": true, "importNames": []interface{}{"Foo"},
					}},
				}},
			},
			{
				// `export type { Foo as Bar }` whole-type-only export → short-circuit.
				Code: `export type { Foo as Bar } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "allowTypeImports": true, "importNames": []interface{}{"Foo"},
					}},
				}},
			},

			// ---- string-literal source name in rename: `import { 'A' as B } from 'lib'` ----
			// tsgo parses the string literal property; source name is the literal value.
			{
				Code: `import { 'A' as B } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "importNames": []interface{}{"B"},
					}},
				}},
			},

			{
				// `allowImportNames` includes the synthesized 'default' → no report.
				Code: `import x = require('foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"name":             "foo",
							"allowImportNames": []interface{}{"default"},
						},
					},
				}},
			},
			{
				// `importNames` does NOT include 'default' → no report on default specifier.
				Code: `import x = require('foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"name":        "foo",
							"importNames": []interface{}{"namedX"},
						},
					},
				}},
			},
			{
				// allowImportNamePattern matches 'default' → no report.
				Code: `import x = require('lib/foo');`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"group":                  []interface{}{"lib/*"},
							"allowImportNamePattern": "^def",
						},
					},
				}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Plain string-array options: import / export / import = require ----
			{
				Code:    `import foo from 'import1';`,
				Options: []interface{}{"import1", "import2"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},
			{
				Code:    `import foo = require('import1');`,
				Options: []interface{}{"import1", "import2"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},
			{
				Code:    `export { foo } from 'import1';`,
				Options: []interface{}{"import1", "import2"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},
			{
				Code: `import foo from 'import1';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{"import1", "import2"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},
			{
				Code: `export { foo } from 'import1';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{"import1", "import2"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- Patterns ----
			{
				Code: `import foo from 'import1/private/foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths":    []interface{}{"import1", "import2"},
					"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patterns", Line: 1, Column: 1},
				},
			},
			{
				Code: `export { foo } from 'import1/private/foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths":    []interface{}{"import1", "import2"},
					"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patterns", Line: 1, Column: 1},
				},
			},

			// ---- Custom-message paths ----
			{
				Code: `import foo from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "import-foo", "message": "Please use import-bar instead."},
						map[string]interface{}{"name": "import-baz", "message": "Please use import-quux instead."},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "pathWithCustomMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `export { foo } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "import-foo", "message": "Please use import-bar instead."},
						map[string]interface{}{"name": "import-baz", "message": "Please use import-quux instead."},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "pathWithCustomMessage", Line: 1, Column: 1},
				},
			},

			// ---- importNames + custom message ----
			{
				Code: `import { Bar } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"importNames": []interface{}{"Bar"},
							"message":     "Please use Bar from /import-bar/baz/ instead.",
							"name":        "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10},
				},
			},
			{
				Code: `export { Bar } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"importNames": []interface{}{"Bar"},
							"message":     "Please use Bar from /import-bar/baz/ instead.",
							"name":        "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10},
				},
			},

			// ---- Pattern + custom message ----
			{
				Code: `import foo from 'import1/private/foo';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{"group": []interface{}{"import1/private/*"}, "message": "usage of import1 private modules not allowed."},
						map[string]interface{}{"group": []interface{}{"import2/*", "!import2/good"}, "message": "import2 is deprecated, except the modules in import2/good."},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternWithCustomMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `export { foo } from 'import1/private/foo';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{"group": []interface{}{"import1/private/*"}, "message": "usage of import1 private modules not allowed."},
						map[string]interface{}{"group": []interface{}{"import2/*", "!import2/good"}, "message": "import2 is deprecated, except the modules in import2/good."},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternWithCustomMessage", Line: 1, Column: 1},
				},
			},

			// ---- Side-effect imports ----
			{
				// Side-effect import is restricted by path (no specifiers).
				Code: `import 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "import-foo"},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},
			{
				// Side-effect import is NOT type-only — allowTypeImports does not exempt it.
				Code: `import 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"allowTypeImports": true, "name": "import-foo"},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- allowTypeImports=true but the import is value-only → must report ----
			{
				Code: `import foo from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"message":          "Please use import-bar instead.",
							"name":             "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "pathWithCustomMessage", Line: 1, Column: 1},
				},
			},
			{
				// `import x = require(...)` is a runtime import, not type-only.
				Code: `import foo = require('import-foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"message":          "Please use import-bar instead.",
							"name":             "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "pathWithCustomMessage", Line: 1, Column: 1},
				},
			},
			{
				// importNames with allowTypeImports — value `Bar` is restricted (not type-only).
				Code: `import { Bar } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar"},
							"message":          "Please use Bar from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10},
				},
			},
			{
				Code: `export { Bar } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar"},
							"message":          "Please use Bar from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10},
				},
			},

			// ---- Patterns + allowTypeImports — value-only must still report ----
			{
				Code: `import foo from 'import1/private/bar';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"group":            []interface{}{"import1/private/*"},
							"message":          "usage of import1 private modules not allowed.",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternWithCustomMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `export { foo } from 'import1/private/bar';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"group":            []interface{}{"import1/private/*"},
							"message":          "usage of import1 private modules not allowed.",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternWithCustomMessage", Line: 1, Column: 1},
				},
			},
			{
				// regex pattern + allowTypeImports — value export must still report.
				Code: `export { foo } from 'import1/private/bar';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"message":          "usage of import1 private modules not allowed.",
							"regex":            "import1/.*",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternWithCustomMessage", Line: 1, Column: 1},
				},
			},
			{
				// case-sensitive regex matches lowercase regex character class — value import.
				Code: `import { foo } from 'import1/private-package';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"caseSensitive":    true,
							"message":          "usage of import1 private modules not allowed.",
							"regex":            "import1/private-[a-z]*",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternWithCustomMessage", Line: 1, Column: 1},
				},
			},

			// ---- export * from ... ----
			{
				Code:    `export * from 'import1';`,
				Options: []interface{}{"import1"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- Pattern matches a deep path ----
			{
				Code: `import type { InvalidTestCase } from '@typescript-eslint/utils/dist/ts-eslint';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"@typescript-eslint/utils/dist/*"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patterns", Line: 1, Column: 1},
				},
			},

			// ---- Mixed value + inline-type specifiers ----
			{
				// `Bar` is a value (restricted); `type Baz` is type-only and exempted by allowTypeImports.
				// Expected: 1 error (only on Bar).
				Code: `import { Bar, type Baz } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar", "Baz"},
							"message":          "Please use Bar and Baz from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10},
				},
			},
			{
				// allowTypeImports=false: both Bar and Baz are reported.
				Code: `import { Bar, type Baz } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": false,
							"importNames":      []interface{}{"Bar", "Baz"},
							"message":          "Please use Bar and Baz from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10},
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 15},
				},
			},
			{
				Code: `export { Bar, type Baz } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar", "Baz"},
							"message":          "Please use Bar and Baz from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10},
				},
			},
			{
				Code: `export { Bar, type Baz } from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"allowTypeImports": false,
							"importNames":      []interface{}{"Bar", "Baz"},
							"message":          "Please use Bar and Baz from /import-bar/baz/ instead.",
							"name":             "import-foo",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10},
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 15},
				},
			},

			// ============================================================
			// Extra coverage — invalid edge cases
			// ============================================================

			// ---- Whole type-only on a source NOT covered by allow-type → must report ----
			// Two paths: one allows type imports for 'allowed', one strict for 'foo'.
			// Type-only import of 'foo' is NOT in the allow-type set → still reported.
			{
				Code: `import type x from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "allowed", "allowTypeImports": true},
						map[string]interface{}{"name": "foo"},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},
			{
				// Pattern allow + path strict for unrelated source → strict path still hits.
				Code: `import x from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths":    []interface{}{"foo"},
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"bar/**"}, "allowTypeImports": true}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- Whitespace-padded source — base trims before matching ----
			{
				Code:    `import x from '  foo  ';`,
				Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- Type-only namespace / export-all is NOT short-circuited at source level ----
			// Upstream's wrapper passes ExportAllDeclaration directly to base. A type-only
			// `export type *` from a strictly-restricted source must still be reported.
			{
				Code:    `export type * from 'foo';`,
				Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},
			{
				Code:    `export type * as ns from 'foo';`,
				Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- Mixed default + inline-type with allowTypeImports ----
			// Default specifier means "not whole-type-only" → passed to base.
			// Base reports default (path entry has importNames covering it? No — importNames=['Y'])
			// Actually default name is "default", not "Y", so the importName check skips default.
			// But the path entry has only importNames, not a name-only restriction → no error on default.
			// The `type Y` specifier IS in importNames AND is type-only AND allowTypeImports=true → skipped.
			// Result: 0 errors.
			// To produce a report, we use a path with NO importNames so the default path is restricted.
			{
				Code: `import x, { type Y } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
				// Default specifier is a value import — `allowTypeImports` only exempts whole-type-only.
				// Since `import x, { type Y }` is mixed, it's NOT whole-type-only and the path-level
				// restriction applies → 1 error on the whole declaration.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- Empty named-imports brace with default still has a runtime side-effect ----
			{
				Code:    `import x, {} from 'foo';`,
				Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- import-equals (CJS) with non-string require argument is silently ignored ----
			// `import x = ns.y` is NOT an ExternalModuleReference → must not match.
			// Sanity: a real-world value import-equals from a restricted source still reports.
			{
				Code:    `import foo = require('restricted');`,
				Options: []interface{}{"restricted"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},
			{
				// Whitespace-padded require() source — typescript-eslint's wrapper trims
				// (synthesizes an ImportDeclaration). We match that behavior.
				Code:    `import foo = require('  restricted  ');`,
				Options: []interface{}{"restricted"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- import-equals synthesizes a default specifier so that importNames /
			// allowImportNames / importNamePattern apply to the local binding.
			// Upstream typescript-eslint's wrapper rewrites `import x = require('foo')`
			// into `import x from 'foo'` with a single ImportDefaultSpecifier; ESLint
			// base treats import-equals as having no specifiers (so this would not
			// fire there). We follow upstream typescript-eslint.
			{
				Code: `import x = require('foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "foo", "importNames": []interface{}{"default"}},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 8},
				},
			},
			{
				Code: `import x = require('foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"name":             "foo",
							"importNames":      []interface{}{"default"},
							"message":          "use named imports",
							"allowTypeImports": true,
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					// Not whole-type-only (no `import type`), so allowTypeImports doesn't apply.
					{MessageId: "importNameWithCustomMessage", Line: 1, Column: 8},
				},
			},
			{
				// allowImportNames=['namedX']: 'default' (synthesized) NOT in allow-list → report.
				Code: `import x = require('foo');`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"name":             "foo",
							"allowImportNames": []interface{}{"namedX"},
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "allowedImportName", Line: 1, Column: 8},
				},
			},
			{
				// importNamePattern matches 'default' → report from pattern matcher.
				Code: `import x = require('lib/foo');`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{
							"group":             []interface{}{"lib/*"},
							"importNamePattern": "^def",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternAndImportName", Line: 1, Column: 8},
				},
			},
			// ============================================================
			// tsgo-specific implementation differences — extra invalid coverage
			// ============================================================

			// ---- Source value resolution: escape sequences DO match after resolution ----
			// `'foo'` resolves to 'foo' which matches the restriction.
			{
				Code:    "import x from '\\u0066oo';",
				Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- importNames: ['*'] restricts the namespace import ----
			{
				Code: `import * as ns from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"*"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					// Namespace specifier hits the path's importNames=['*'] match → "everything" message.
					// Reported on the * specifier loc.
					{MessageId: "everything", Line: 1, Column: 8},
				},
			},

			// ---- Source-order preservation in error reports ----
			// Specifiers reported in declaration order, not config order.
			{
				Code: `import { B, A, C } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"A", "B", "C"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					// B at col 10, A at col 13, C at col 16 — source order preserved by orderedImportNames.
					{MessageId: "importName", Line: 1, Column: 10},
					{MessageId: "importName", Line: 1, Column: 13},
					{MessageId: "importName", Line: 1, Column: 16},
				},
			},

			// ---- Both a path AND a pattern matching same source: independent reports ----
			{
				Code: `import x from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths":    []interface{}{"foo"},
					"patterns": []interface{}{"foo"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
					{MessageId: "patterns", Line: 1, Column: 1},
				},
			},

			// ---- Group + regex on the same pattern: regex takes precedence (matches schema's oneOf intent) ----
			// regex 'bar.*' matches 'bar/baz' but group ['foo*'] would not — verify regex wins.
			{
				Code: `import x from 'bar/baz';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo*"},
						"regex": "bar.*",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patterns", Line: 1, Column: 1},
				},
			},

			// ---- Multiple imports of the same restricted source: each reported separately ----
			{
				Code: `
import a from 'foo';
import b from 'foo';
import c = require('foo');`,
				Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 2, Column: 1},
					{MessageId: "path", Line: 3, Column: 1},
					{MessageId: "path", Line: 4, Column: 1},
				},
			},

			// ---- Mixed import / export-named / import-equals from same source: 3 reports ----
			{
				Code: `
import x from 'foo';
export { x } from 'foo';
import y = require('foo');`,
				Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 2, Column: 1},
					{MessageId: "path", Line: 3, Column: 1},
					{MessageId: "path", Line: 4, Column: 1},
				},
			},

			// ---- Special-char path matching (scoped + dot + dash + underscore) ----
			{
				Code:    `import x from '@my-org/pkg.beta_v1';`,
				Options: []interface{}{"@my-org/pkg.beta_v1"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- Deep glob: `**/leaf` matches any depth ----
			{
				Code: `import x from 'a/b/c/d/leaf';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"**/leaf"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patterns", Line: 1, Column: 1},
				},
			},

			// ---- Side-effect import inside the source-level short-circuit candidate set ----
			// `import 'foo'` has 0 specifiers → not whole-type-only by upstream's spec —
			// short-circuit must NOT skip → path-level check fires.
			{
				Code: `import 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "foo", "allowTypeImports": true},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- ImportEquals + importNamePattern matches synthetic 'default' ----
			{
				Code: `import x = require('lib/foo');`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group":             []interface{}{"lib/*"},
						"importNamePattern": "^default$",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternAndImportName", Line: 1, Column: 8},
				},
			},

			// ---- ImportEquals + allowImportNamePattern — synthetic 'default' NOT matching → report ----
			{
				Code: `import x = require('lib/foo');`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group":                  []interface{}{"lib/*"},
						"allowImportNamePattern": "^named",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "allowedImportNamePattern", Line: 1, Column: 8},
				},
			},

			// ---- import inside ambient module declaration with restricted source ----
			{
				Code: `
declare module 'wrapper' {
  import x = require('restricted');
  export const y: typeof x;
}`,
				Options: []interface{}{"restricted"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 3, Column: 3},
				},
			},

			// ---- Per-specifier reporting position for `import { Foo as Bar }` matching 'Foo' ----
			// Source name is 'Foo', so importNames=['Foo'] matches and reports on the
			// element location (covers the whole `Foo as Bar`).
			{
				Code: `import { Foo as Bar } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "importNames": []interface{}{"Foo"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
				},
			},

			// ---- Re-export of default: `export { default } from 'foo'` ----
			{
				Code: `export { default } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"default"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
				},
			},

			// ---- `export { default as Foo } from 'foo'`: source name is 'default' ----
			{
				Code: `export { default as Foo } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"default"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
				},
			},

			// ---- Type-only inline specifier with importNames + NO allowTypeImports ----
			// Per-specifier allowTypeImports defaults to false → type-only specifier
			// is reported just like a value specifier.
			{
				Code: `import { type Bar } from 'foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"Bar"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
				},
			},

			// ---- `as` rename + allowImportNames: SOURCE name not in allow → report ----
			{
				// Source name `Foo` is NOT in allowImportNames=['Bar'] → reports allowedImportName.
				Code: `import { Foo as Bar } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "allowImportNames": []interface{}{"Bar"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "allowedImportName", Line: 1, Column: 10},
				},
			},

			// ---- `as` rename in export + allowTypeImports + value specifier → report ----
			{
				// `export { Foo as Bar }` is value (no `type` modifier on the specifier or clause).
				// Source name `Foo` matches importNames; allowTypeImports doesn't apply (not type-only).
				Code: `export { Foo as Bar } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name":             "lib",
						"allowTypeImports": true,
						"importNames":      []interface{}{"Foo"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
				},
			},

			// ---- Mixed value+type rename specifiers with importNames matching source names ----
			{
				// `import { A as X, type B as Y }` — A is value, B is type-only.
				// allowTypeImports=true → only A reported (source name 'A').
				Code: `import { A as X, type B as Y } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name":             "lib",
						"allowTypeImports": true,
						"importNames":      []interface{}{"A", "B"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
				},
			},
			{
				// Same code, allowTypeImports=false → both A and B reported.
				Code: `import { A as X, type B as Y } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name":             "lib",
						"allowTypeImports": false,
						"importNames":      []interface{}{"A", "B"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
					{MessageId: "importName", Line: 1, Column: 18},
				},
			},

			// ---- string-literal source name in rename matches importNames ----
			{
				// `import { 'A' as B }` — source name is the literal value 'A'.
				Code: `import { 'A' as B } from 'lib';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "lib", "importNames": []interface{}{"A"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
				},
			},

			// ---- Comments and multiline don't break source detection ----
			{
				Code: `
import {
  Bar,
  type Baz,
} from 'import-foo';`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"name":             "import-foo",
							"allowTypeImports": true,
							"importNames":      []interface{}{"Bar", "Baz"},
							"message":          "Use lib instead.",
						},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					// Bar (column 3 of line 3) is reported; type Baz is exempted.
					{MessageId: "importNameWithCustomMessage", Line: 3, Column: 3},
				},
			},

			// ---- Negation pattern — restricted parent, allowed child, restricted grandchild ----
			{
				Code: `import x from 'lib/internal/restricted';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group":            []interface{}{"lib/internal/*", "!lib/internal/allowed"},
						"allowTypeImports": true,
						"message":          "Internal modules — types only.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternWithCustomMessage", Line: 1, Column: 1},
				},
			},

			// ---- Case-insensitive default vs. case-sensitive override ----
			{
				// Default insensitivity: `FOO` matches glob `foo` because gitignore matchers default to insensitive.
				Code: `import x from 'FOO';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"foo"}}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patterns", Line: 1, Column: 1},
				},
			},
			// ---- Value-only namespace import on a restricted path ----
			{
				Code:    `import * as ns from 'foo';`,
				Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
				},
			},

			// ---- Real-world: type from internal barrel, value of a different export still reports ----
			{
				Code: `import { x, type T } from './internal';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group":            []interface{}{"./internal", "./internal/**"},
						"allowTypeImports": true,
						"message":          "Internal modules — types only.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					// Whole import is mixed (x is value), so source short-circuit doesn't apply.
					// Pattern matches → reportPathForPatterns. Per group.allowTypeImports, type-only specifiers are skipped.
					// `x` triggers the patterns message; `type T` is exempted.
					{MessageId: "patternWithCustomMessage", Line: 1, Column: 1},
				},
			},
		},
	)
}
