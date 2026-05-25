package no_restricted_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedImportsRule,
		// ========== Valid cases ==========
		[]rule_tester.ValidTestCase{
			// --- Basic: no options / non-matching ---
			{Code: `import os from "os";`},
			{Code: `import async from "async";`},
			{Code: `import os from "os";`, Options: []interface{}{"osx"}},
			{Code: `import fs from "fs";`, Options: []interface{}{"crypto"}},
			{Code: `import path from "path";`, Options: []interface{}{"crypto", "stream", "os"}},
			// Side-effect import not matching
			{Code: `import "foo"`, Options: []interface{}{"crypto"}},
			// Subpath import ≠ parent path (exact match semantics)
			{Code: `import "foo/bar";`, Options: []interface{}{"foo"}},

			// --- Paths option ---
			{Code: `import withPaths from "foo/bar";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{"foo", "bar"},
			}}},

			// --- Patterns option ---
			{Code: `import withPatterns from "foo/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{"foo/c*"},
			}}},

			// --- Relative / absolute paths not matching ---
			{Code: `import foo from 'foo';`, Options: []interface{}{"../foo"}},
			{Code: `import foo from 'foo';`, Options: []interface{}{map[string]interface{}{"paths": []interface{}{"../foo"}}}},
			{Code: `import foo from 'foo';`, Options: []interface{}{map[string]interface{}{"patterns": []interface{}{"../foo"}}}},
			{Code: `import foo from 'foo';`, Options: []interface{}{"/foo"}},
			{Code: `import relative from '../foo';`},
			{Code: `import relative from '../foo';`, Options: []interface{}{"../notFoo"}},
			{Code: `import absolute from '/foo';`},
			{Code: `import absolute from '/foo';`, Options: []interface{}{"/notFoo"}},

			// --- Gitignore negation ---
			{Code: `import withGitignores from "foo/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{"foo/*", "!foo/bar"},
			}}},
			{Code: `import withPatterns from "foo/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group":   []interface{}{"foo/*", "!foo/bar"},
					"message": "foo is forbidden, use bar instead",
				}},
			}}},

			// --- Case sensitive not matching ---
			{Code: `import x from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group":         []interface{}{"FOO"},
					"caseSensitive": true,
				}},
			}}},

			// --- importNames: allowed names not restricted ---
			{Code: `import AllowedObject from "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},
			// Default import treated as "default", NOT as identifier text
			{Code: `import DisallowedObject from "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},
			// Named import — source name differs from restricted
			{Code: `import { AllowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},
			// Alias: restriction checks source name (AllowedObject), not alias
			{Code: `import { AllowedObject as DisallowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},
			// Side-effect import with importNames → no specifiers → no error
			{Code: `import "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},
			// Namespace import from different module
			{Code: `import * as DisallowedObject from "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "bar", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},

			// --- Export: non-restricted ---
			{Code: `export * from "foo";`, Options: []interface{}{"bar"}},
			{Code: `export * from "foo";`, Options: []interface{}{map[string]interface{}{
				"name": "bar", "importNames": []interface{}{"DisallowedObject"},
			}}},

			// --- allowImportNames ---
			{Code: `import { AllowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "allowImportNames": []interface{}{"AllowedObject"},
				}},
			}}},
			{Code: `import { foo } from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"foo"}, "allowImportNames": []interface{}{"foo"},
				}},
			}}},
			{Code: `export { bar } from 'foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "allowImportNames": []interface{}{"bar"},
				}},
			}}},
			{Code: `export { bar } from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"foo"}, "allowImportNames": []interface{}{"bar"},
				}},
			}}},

			// --- Regex: non-matching ---
			{Code: `import x from "foo/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{"regex": "foo/baz"}},
			}}},
			{Code: `import x from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{"regex": "FOO", "caseSensitive": true}},
			}}},
			// Regex negative lookahead: "foo/bar" NOT matched by "foo/(?!bar)"
			{Code: `import x from "foo/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{"regex": "foo/(?!bar)"}},
			}}},

			// --- importNamePattern: default import NOT matched ---
			{Code: `import Foo from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
				}},
			}}},
			// importNamePattern: named import doesn't match
			{Code: `import { Bar } from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
				}},
			}}},
			// Aliased import: source name (Bar) doesn't match pattern ^Foo
			{Code: `import { Bar as Foo } from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
				}},
			}}},
			// Default import NOT matched by importNames in patterns (only named "Foo" matches, not default)
			{Code: `import Foo from '../../my/relative-module';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"**/my/relative-module"}, "importNames": []interface{}{"Foo"},
				}},
			}}},

			// --- allowImportNamePattern ---
			{Code: `import { Foo } from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"foo"}, "allowImportNamePattern": "^Foo",
				}},
			}}},

			// --- Pattern importNames + importNamePattern together: default import ---
			// Default import is "default", doesn't match importNames ["Foo"] or pattern "^Foo"
			{Code: `import Foo from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"importNames": []interface{}{"Foo"}, "group": []interface{}{"foo"}, "importNamePattern": "^Foo",
				}},
			}}},
			// Default + named: only named matches pattern, default doesn't
			{Code: `import Foo, { Baz as Bar } from '../../my/relative-module';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"**/my/relative-module"}, "importNamePattern": "^(Foo|Bar)",
				}},
			}}},

			// --- TypeScript: type import allowed ---
			{Code: `import type foo from 'import-foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{"name": "import-foo", "allowTypeImports": true}},
			}}},
			{Code: `import type { Bar } from 'import-foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{"name": "import-foo", "allowTypeImports": true}},
			}}},
			{Code: `export type { Bar } from 'import-foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{"name": "import-foo", "allowTypeImports": true}},
			}}},
			{Code: `import type { Bar } from 'import-foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"import-foo"}, "allowTypeImports": true,
				}},
			}}},
			// import type = require() allowed
			{Code: `import type foo = require("import-foo");`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{"name": "import-foo", "allowTypeImports": true}},
			}}},
			// all specifiers individually type → whole import is type-only
			{Code: `import { type A, type B } from "import-foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{"name": "import-foo", "allowTypeImports": true}},
			}}},
			// export with all specifiers individually type
			{Code: `export { type A, type B } from "import-foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{"name": "import-foo", "allowTypeImports": true}},
			}}},
			// import Bar = Foo.Bar (namespace import, no external module → ignored)
			{Code: `import Bar = Foo.Bar;`, Options: []interface{}{"Foo"}},
			// export type * allowed with allowTypeImports
			{Code: `export type * from "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
			}}},
			// import type = require() allowed with pattern
			{Code: `import type fs = require("fs");`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"f*"}, "allowTypeImports": true}},
			}}},
			// individual type specifier with importNames + allowTypeImports → allowed
			{Code: `import { type Bar } from 'import-foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "import-foo", "importNames": []interface{}{"Bar"}, "allowTypeImports": true,
				}},
			}}},
			{Code: `export { type Bar } from 'import-foo';`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "import-foo", "importNames": []interface{}{"Bar"}, "allowTypeImports": true,
				}},
			}}},
			// export { type bar, baz } where bar matches importNames + allowTypeImports: bar is type → skipped, baz doesn't match → no error
			{Code: `export { type bar, baz } from "foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"bar"}, "allowTypeImports": true,
				}},
			}}},
			// Pattern importNames + allowTypeImports: type specifier of matching name → skipped, non-matching name → no error
			{Code: `import { Bar, type Baz } from "import/private/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"import/private/*"}, "importNames": []interface{}{"Baz"}, "allowTypeImports": true,
				}},
			}}},
			// Pattern allowImportNames + allowTypeImports: Foo allowed, type Bar skipped
			{Code: `import { Foo, type Bar } from "import/private/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"import/private/*"}, "allowImportNames": []interface{}{"Foo"}, "allowTypeImports": true,
				}},
			}}},
			// Pattern allowImportNamePattern + allowTypeImports
			{Code: `import { Foo, type Bar } from "import/private/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"import/private/*"}, "allowImportNamePattern": "^Foo", "allowTypeImports": true,
				}},
			}}},
			// export { Baz, type Bar } with allowImportNames + allowTypeImports
			{Code: `export { Baz, type Bar } from "import/private/bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"import/private/*"}, "allowImportNames": []interface{}{"Baz"}, "allowTypeImports": true,
				}},
			}}},
			// export { bar, type baz } with path allowImportNames + allowTypeImports
			{Code: `export { bar, type baz } from "import-foo";`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "import-foo", "allowImportNames": []interface{}{"bar"}, "allowTypeImports": true,
				}},
			}}},
			// import = require() with path importNames: no specifiers → no importName error (valid)
			{Code: `import foo = require('foo');`, Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"foo"},
				}},
			}}},
			// @scoped package in pattern
			{Code: `import { foo } from '@app/api/enums';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{
					"group": []interface{}{"@app/api/*", "!@app/api/enums"},
				}},
			}}},
			// String literal specifier: 'AllowedObject' (not in restriction list)
			{Code: "import { 'AllowedObject' as bar } from \"foo\";", Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},
			// String literal ' ' (space) is not '' (empty)
			{Code: "import { ' ' as bar } from \"foo\";", Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{""},
				}},
			}}},
			// Aliased string literal: source name 'AllowedObject' doesn't match restriction 'DisallowedObject'
			{Code: "import { 'AllowedObject' as DisallowedObject } from \"foo\";", Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},
			// export string literal specifier: source name 'AllowedObject' not restricted
			{Code: "export { 'AllowedObject' } from \"foo\";", Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},
			// export aliased string literal: source name 'AllowedObject' not restricted
			{Code: "export { 'AllowedObject' as DisallowedObject } from \"foo\";", Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{map[string]interface{}{
					"name": "foo", "importNames": []interface{}{"DisallowedObject"},
				}},
			}}},

			// --- Edge cases: ./ in import source (pattern matching via **/ handles . segment) ---
			// unrooted pattern matches ./foo (** absorbs the . segment)
			{Code: `import foo from './foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{"bar"},
			}}},
			// rooted pattern does NOT match ./foo/bar (npm ignore behavior: no ./ normalization)
			{Code: `import foo from './foo/bar';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"foo/bar"}}},
			}}},
			// ./foo in pattern stays as-is (has /), does NOT match foo
			{Code: `import foo from 'foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"./foo"}}},
			}}},

			// --- Edge case: consecutive slashes — foo/* does NOT match foo//bar ---
			{Code: `import x from "foo//bar";`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"foo/*"}}},
			}}},

			// --- Edge case: whitespace in import source trimmed before matching ---
			// 'foo ' and ' foo' are trimmed to 'foo', so they DON'T match pattern 'bar'
			{Code: `import foo from 'foo ';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{"bar"},
			}}},
			{Code: `import foo from ' foo';`, Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{"bar"},
			}}},
		},

		// ========== Invalid cases ==========
		[]rule_tester.InvalidTestCase{
			// --- Basic string restriction ---
			{
				Code: `import "fs"`, Options: []interface{}{"fs"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			{
				Code: `import os from "os ";`, Options: []interface{}{"fs", "crypto ", "stream", "os"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			{
				Code: `import "foo/bar";`, Options: []interface{}{"foo/bar"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},

			// --- Path option ---
			{
				Code: `import withPaths from "foo/bar";`, Options: []interface{}{map[string]interface{}{"paths": []interface{}{"foo/bar"}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},

			// --- Pattern matching ---
			{
				Code: `import withPatterns from "foo/bar";`, Options: []interface{}{map[string]interface{}{"patterns": []interface{}{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			{
				Code: `import withPatterns from "foo/bar";`, Options: []interface{}{map[string]interface{}{"patterns": []interface{}{"bar"}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- Pattern group with custom message ---
			{
				Code: `import withPatterns from "foo/baz";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo/*", "!foo/bar"}, "message": "foo is forbidden, use foo/bar instead",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternWithCustomMessage", Line: 1, Column: 1}},
			},
			// Pattern group exact match
			{
				Code: `import x from "foo/bar";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"foo/bar"}}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- Case sensitivity ---
			{
				Code: `import x from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"FOO"}}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			// Explicit caseSensitive: false
			{
				Code: `import x from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"FOO"}, "caseSensitive": false,
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- Gitignore negation: negation doesn't match → still restricted ---
			{
				Code: `import x from "foo/bar";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"foo/*", "!foo/baz"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- Export * ---
			{
				Code: `export * from "fs";`, Options: []interface{}{"fs"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			// --- Export * as ns ---
			{
				Code: `export * as ns from "fs";`, Options: []interface{}{"fs"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			// --- Export named ---
			{
				Code: `export {a} from "fs";`, Options: []interface{}{"fs"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			// Export named with importNames + custom message
			{
				Code: `export {foo as b} from "fs";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "fs", "importNames": []interface{}{"foo"}, "message": "Don't import 'foo'.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage", Line: 1, Column: 9}},
			},
			// Export * as ns with importNames → error on * token (column 8)
			{
				Code: `export * as ns from "fs";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "fs", "importNames": []interface{}{"foo"}, "message": "Don't import 'foo'.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "everythingWithCustomMessage", Line: 1, Column: 8}},
			},

			// --- Custom messages ---
			{
				Code: `import x from "foo";`, Options: []interface{}{map[string]interface{}{
					"name": "foo", "message": "Please import from 'bar' instead.",
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "pathWithCustomMessage", Line: 1, Column: 1}},
			},
			{
				Code: `import x from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "message": "Please import from 'bar' instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "pathWithCustomMessage", Line: 1, Column: 1}},
			},

			// --- importNames: "default" ---
			{
				Code: `import DisallowedObject from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"default"},
						"message": "Please import the default import of 'foo' from /bar/ instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage"}},
			},
			// import foo, { bar } with importNames: ["default"] → only default reported
			{
				Code: `import foo, { bar } from 'mod';`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "mod", "importNames": []interface{}{"default"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importName"}},
			},
			// import foo, * as bar with importNames: ["default"] → BOTH default and * reported
			{
				Code: `import foo, * as bar from 'mod';`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "mod", "importNames": []interface{}{"default"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName"},
					{MessageId: "everything"},
				},
			},
			// import foo, { default as bar } with importNames: ["default"] → both default and named "default" reported
			{
				Code: `import foo, { default as bar } from 'mod';`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "mod", "importNames": []interface{}{"default"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName"},
					{MessageId: "importName"},
				},
			},

			// --- Star import with importNames ---
			{
				Code: `import * as All from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"DisallowedObject"},
						"message": "Please import 'DisallowedObject' from /bar/ instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "everythingWithCustomMessage"}},
			},
			// * import restricted as whole module (no importNames)
			{
				Code: `import * as bar from 'foo';`, Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},

			// --- Named import restrictions ---
			{
				Code: `import { DisallowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"DisallowedObject"},
						"message": "Please import 'DisallowedObject' from /bar/ instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10}},
			},
			// Aliased import — source name restricted
			{
				Code: `import { DisallowedObject as AllowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"DisallowedObject"},
						"message": "Please import 'DisallowedObject' from /bar/ instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10}},
			},
			// Only restricted name in mixed import gets error
			{
				Code: `import { AllowedObject, DisallowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"DisallowedObject"},
						"message": "Please import 'DisallowedObject' from /bar/ instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage", Line: 1, Column: 25}},
			},

			// --- Multiple restricted names → multiple errors ---
			{
				Code: `import { DisallowedObjectOne, DisallowedObjectTwo, AllowedObject } from "foo";`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "importNames": []interface{}{"DisallowedObjectOne", "DisallowedObjectTwo"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
					{MessageId: "importName", Line: 1, Column: 31},
				},
			},

			// --- Duplicate source name: import { a, a as b } → 2 errors ---
			{
				Code: `import { a, a as b } from 'mod';`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "mod", "importNames": []interface{}{"a"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
					{MessageId: "importName", Line: 1, Column: 13},
				},
			},
			// export { x as y, x as z } → 2 errors on same source name
			{
				Code: `export { x as y, x as z } from 'mod';`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "mod", "importNames": []interface{}{"x"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importName", Line: 1, Column: 10},
					{MessageId: "importName", Line: 1, Column: 18},
				},
			},

			// --- Multiple path entries for same module ---
			// Whole-module restriction + specific importName → 2 errors on { bar }
			{
				Code: `import { bar } from 'mod'`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "mod"},
						map[string]interface{}{"name": "mod", "importNames": []interface{}{"bar"}, "message": "Import bar from qux instead."},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
					{MessageId: "importNameWithCustomMessage"},
				},
			},
			// Multiple entries with different importNames → errors from each matching entry
			{
				Code: `import { foo, bar, baz } from 'mod'`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "mod", "importNames": []interface{}{"foo"}, "message": "Import foo from qux instead."},
						map[string]interface{}{"name": "mod", "importNames": []interface{}{"baz"}, "message": "Import baz from qux instead."},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage"},
					{MessageId: "importNameWithCustomMessage"},
				},
			},
			// * import with multiple path entries → error from each entry
			{
				Code: `import * as mod from 'mod'`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "mod", "importNames": []interface{}{"foo"}, "message": "Import foo from qux instead."},
						map[string]interface{}{"name": "mod", "importNames": []interface{}{"bar"}, "message": "Import bar from qux instead."},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "everythingWithCustomMessage"},
					{MessageId: "everythingWithCustomMessage"},
				},
			},

			// --- Real-world: React Native multiple restrictions ---
			{
				Code: `import { Image, Text, ScrollView } from 'react-native'`,
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "react-native", "importNames": []interface{}{"Text"}, "message": "import Text from ui/_components instead"},
						map[string]interface{}{"name": "react-native", "importNames": []interface{}{"ScrollView"}, "message": "import ScrollView from ui/_components instead"},
						map[string]interface{}{"name": "react-native", "importNames": []interface{}{"Image"}, "message": "import Image from ui/_components instead"},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importNameWithCustomMessage"},
					{MessageId: "importNameWithCustomMessage"},
					{MessageId: "importNameWithCustomMessage"},
				},
			},

			// --- Relative/absolute path restrictions ---
			{
				Code: `import relative from '../foo';`, Options: []interface{}{"../foo"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			{
				Code: `import absolute from '/foo';`, Options: []interface{}{"/foo"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			// Relative path matched by pattern
			{
				Code: `import x from '../foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"../foo"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			// Absolute path matched by pattern "foo" (matches /foo because "foo" component)
			{
				Code: `import x from '/foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"foo"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			// Hash import with escaped # pattern
			{
				Code: `import x from '#foo/bar';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"\\#foo"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- Regex ---
			{
				Code: `import x from "foo/baz";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"regex": "foo/baz", "message": "foo is forbidden, use bar instead",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternWithCustomMessage", Line: 1, Column: 1}},
			},
			{
				Code: `import x from "foo";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"regex": "FOO"}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			// Regex case sensitive matching
			{
				Code: `import x from 'FOO';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"regex": "FOO", "caseSensitive": true,
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			// Regex negative lookahead: "foo/baz" matched by "foo/(?!bar)"
			{
				Code: `import x from "foo/baz";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"regex": "foo/(?!bar)", "message": "foo is forbidden, use bar instead",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternWithCustomMessage", Line: 1, Column: 1}},
			},
			// Regex with importNamePattern
			{
				Code: `import { Foo } from '../../my/relative-module';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"regex": "my/relative-module", "importNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndImportName", Line: 1, Column: 10}},
			},

			// --- Pattern importNames ---
			{
				Code: `import { Foo } from '../../my/relative-module';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"**/my/relative-module"}, "importNames": []interface{}{"Foo"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndImportName", Line: 1, Column: 10}},
			},
			// Pattern importNames: multiple restricted in same import
			{
				Code: `import { Foo, Bar } from '../../my/relative-module';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"**/my/relative-module"}, "importNames": []interface{}{"Foo", "Bar"},
						"message": "Import from @/utils instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternAndImportNameWithCustomMessage", Line: 1, Column: 10},
					{MessageId: "patternAndImportNameWithCustomMessage", Line: 1, Column: 15},
				},
			},
			// Pattern importNames: "default" matches default import
			{
				Code: `import Foo from 'mod';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"mod"}, "importNames": []interface{}{"default"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndImportName"}},
			},
			// Pattern importNames: default + namespace → 2 errors
			{
				Code: `import def, * as ns from 'mod';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"mod"}, "importNames": []interface{}{"default"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternAndImportName"},
					{MessageId: "patternAndEverything"},
				},
			},

			// --- Pattern importNamePattern ---
			{
				Code: `import { Foo } from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndImportName", Line: 1, Column: 10}},
			},
			// Aliased: source name checked
			{
				Code: `import { Foo as Bar } from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndImportName", Line: 1, Column: 10}},
			},
			// importNamePattern does not match default imports
			{
				Code: `import Foo, { Bar } from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "importNamePattern": "^(Foo|Bar)",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndImportName", Line: 1, Column: 15}},
			},
			// Multiple matches with importNamePattern
			{
				Code: `import { Foo, Bar } from '../../my/relative-module';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"**/my/relative-module"}, "importNamePattern": "^(Foo|Bar)",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternAndImportName", Line: 1, Column: 10},
					{MessageId: "patternAndImportName", Line: 1, Column: 15},
				},
			},
			// export with importNamePattern
			{
				Code: `export { Foo } from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndImportName", Line: 1, Column: 10}},
			},
			// export * with importNamePattern → star reported at column 8 (the * token)
			{
				Code: `export * from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndEverythingWithRegexImportName", Line: 1, Column: 8}},
			},

			// --- importNames + importNamePattern together (OR logic) ---
			// importNames has "Foo", importNamePattern has "^Bar" → { Foo, Bar } both match
			{
				Code: `import { Foo, Bar } from '../../my/relative-module';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"importNames": []interface{}{"Foo"}, "group": []interface{}{"**/my/relative-module"}, "importNamePattern": "^Bar",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternAndImportName", Line: 1, Column: 10},
					{MessageId: "patternAndImportName", Line: 1, Column: 15},
				},
			},
			// With both importNames and importNamePattern, * import uses importNames (has priority)
			{
				Code: `import * as All from '../../my/relative-module';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"importNames": []interface{}{"Foo"}, "group": []interface{}{"**/my/relative-module"},
						"importNamePattern": "^Foo", "message": "Import from @/utils instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndEverythingWithCustomMessage"}},
			},

			// --- Star import with importNamePattern ---
			{
				Code: `import * as Foo from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndEverythingWithRegexImportName"}},
			},

			// --- Star import with pattern importNames ---
			{
				Code: `import * as All from 'foo-bar-baz';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"**"}, "importNames": []interface{}{"Foo", "Bar"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndEverything"}},
			},

			// --- Multiple pattern groups matching same import → separate errors ---
			{
				Code: `import * as All from "foo/bar";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{
						map[string]interface{}{"group": []interface{}{"foo/*"}, "allowImportNames": []interface{}{"Foo", "Bar"}},
						map[string]interface{}{"group": []interface{}{"*/bar"}, "allowImportNames": []interface{}{"Foo", "Bar"}, "message": "Good luck!"},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "everythingWithAllowImportNames"},
					{MessageId: "everythingWithAllowImportNamesAndCustomMessage"},
				},
			},

			// --- allowImportNames ---
			{
				Code: `import { AllowedObject, DisallowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "allowImportNames": []interface{}{"AllowedObject"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "allowedImportName", Line: 1, Column: 25}},
			},
			// allowImportNames: multiple allowed, one not
			{
				Code: `import { foo, bar, baz, qux } from "foo-bar-baz";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo-bar-baz"}, "allowImportNames": []interface{}{"foo", "bar", "baz"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "allowedImportName", Line: 1, Column: 25}},
			},
			// allowImportNames with star import
			{
				Code: `import * as AllowedObject from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "allowImportNames": []interface{}{"AllowedObject"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "everythingWithAllowImportNames"}},
			},
			// Pattern allowImportNames
			{
				Code: `import { AllowedObject, DisallowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "allowImportNames": []interface{}{"AllowedObject"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "allowedImportName", Line: 1, Column: 25}},
			},

			// --- allowImportNamePattern ---
			{
				Code: `import { Bar } from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "allowImportNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "allowedImportNamePattern", Line: 1, Column: 10}},
			},
			{
				Code: `export { Bar } from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "allowImportNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "allowedImportNamePattern", Line: 1, Column: 10}},
			},
			{
				Code: `import * as Foo from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "allowImportNamePattern": "^Foo",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "everythingWithAllowedImportNamePattern"}},
			},

			// --- TypeScript ---
			// Regular import fails even with allowTypeImports
			{
				Code: `import foo from 'import-foo';`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "import-foo", "allowTypeImports": true}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			// import = require()
			{
				Code: `import foo = require('import-foo');`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "import-foo"}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			// import = require() with pattern
			{
				Code: `import foo = require('import-foo');`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"import-foo"}}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			// Type import with restricted importNames still reported (no allowTypeImports)
			{
				Code: `import type { bar } from "mod";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "mod", "importNames": []interface{}{"bar"}, "message": "don't import 'bar' at all",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage", Line: 1, Column: 15}},
			},
			// Mixed type and regular imports: only regular reported with allowTypeImports
			{
				Code: `import { Bar, type Baz } from "import-foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "import-foo", "importNames": []interface{}{"Bar", "Baz"}, "allowTypeImports": true,
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importName", Line: 1, Column: 10}},
			},
			// export type * restricted (without allowTypeImports)
			{
				Code: `export type * from "foo";`, Options: []interface{}{"foo"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},
			// Pattern with allowTypeImports: regular import fails, type import passes
			{
				Code: `import { Bar } from 'import-foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"import-foo"}, "allowTypeImports": true,
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			// Pattern with allowTypeImports + importNames: regular specifier reported, type specifier skipped
			{
				Code: `import { Foo, type Bar } from "import/private/bar";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"import/private/*"}, "importNames": []interface{}{"Foo", "Bar"}, "allowTypeImports": true,
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "patternAndImportName", Line: 1, Column: 10},
				},
			},

			// --- Side-effect import with allowTypeImports: still restricted (side-effect can't be type-only) ---
			{
				Code: `import 'import-foo';`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "import-foo", "allowTypeImports": true}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},

			// --- import = require() with allowTypeImports: regular require still restricted ---
			{
				Code: `import foo = require('import-foo');`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "import-foo", "allowTypeImports": true, "message": "Please use import-bar instead.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "pathWithCustomMessage", Line: 1, Column: 1}},
			},

			// --- export { Bar, type Baz } with allowTypeImports: only Bar reported ---
			{
				Code: `export { Bar, type Baz } from 'import-foo';`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "import-foo", "importNames": []interface{}{"Bar", "Baz"}, "allowTypeImports": true,
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importName"}},
			},

			// --- export type { bar } with two path entries: one with allowTypeImports, one without ---
			// First entry allows type imports of "foo" → skips. Second entry restricts "bar" without allowTypeImports → reports.
			{
				Code: `export type { bar } from "mod";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{"name": "mod", "importNames": []interface{}{"foo"}, "allowTypeImports": true, "message": "import 'foo' only as type"},
						map[string]interface{}{"name": "mod", "importNames": []interface{}{"bar"}, "message": "don't import 'bar' at all"},
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage"}},
			},

			// --- export type * + export * with allowTypeImports: only regular export * reported ---
			{
				Code: "export type * from \"foo\";\nexport * from \"foo\";",
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "foo", "allowTypeImports": true}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 2, Column: 1}},
			},

			// --- export {} from "mod" (empty named export): whole path still restricted ---
			{
				Code: "export { } from \"mod\";\nexport type { } from \"mod\";",
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{"name": "mod", "allowTypeImports": true}},
				}},
				// export { } → no specifiers but not type-only → whole path restricted
				// export type { } → type-only → allowed
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},

			// --- allowTypeImports: false explicitly → type import still restricted ---
			{
				Code: "import { Foo } from 'restricted-path';\nimport type { Bar } from 'restricted-path';",
				Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "restricted-path", "allowTypeImports": false, "message": "This import is restricted.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "pathWithCustomMessage", Line: 1, Column: 1},
					{MessageId: "pathWithCustomMessage", Line: 2, Column: 1},
				},
			},

			// --- import = require() with pattern ---
			{
				Code: `import fs = require("fs");`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{"group": []interface{}{"f*"}}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- Type-only export matched by pattern (type-only but no allowTypeImports) ---
			{
				Code: `export type { InvalidTestCase } from '@typescript-eslint/utils/dist/ts-eslint';`,
				Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"@typescript-eslint/utils/dist/*"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- allowImportNames + allowTypeImports: regular import of non-allowed name restricted ---
			{
				Code: `import { baz } from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "allowImportNames": []interface{}{"bar"}, "allowTypeImports": true,
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "allowedImportName", Line: 1, Column: 10}},
			},

			// --- Both paths AND patterns matching same import → errors from both ---
			{
				Code: `import foo from 'import1/private/foo';`, Options: []interface{}{map[string]interface{}{
					"paths":    []interface{}{"import1/private/foo"},
					"patterns": []interface{}{"import1/private/*"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "path", Line: 1, Column: 1},
					{MessageId: "patterns", Line: 1, Column: 1},
				},
			},

			// --- @scoped package matched by pattern ---
			{
				Code: `import { foo } from '@app/api/bar';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"@app/api/*", "!@app/api/enums"},
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- @scoped with importNamePattern ---
			{
				Code: `import { Foo_Enum } from '@app/api';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"regex": "@app/api$", "importNamePattern": "_Enum$",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndImportName", Line: 1, Column: 10}},
			},

			// --- String literal specifier names ---
			// export { 'foo' as b } from "fs" with importNames: ["foo"]
			{
				Code: "export { 'foo' as b } from \"fs\";", Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "fs", "importNames": []interface{}{"foo"}, "message": "Don't import 'foo'.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10}},
			},
			// export { 'foo' } from "fs" (no alias, StringLiteral name)
			{
				Code: "export { 'foo' } from \"fs\";", Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "fs", "importNames": []interface{}{"foo"}, "message": "Don't import 'foo'.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10}},
			},
			// Empty string import name: export { '' } from "fs" with importNames: [""]
			{
				Code: "export { '' } from \"fs\";", Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "fs", "importNames": []interface{}{""}, "message": "Don't import ''.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "importNameWithCustomMessage", Line: 1, Column: 10}},
			},

			// --- Complex glob pattern ---
			// **/*-*-baz matching foo-bar-baz
			{
				Code: `import * as All from 'foo-bar-baz';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"**/*-*-baz"}, "importNames": []interface{}{"Foo", "Bar"},
						"message": "Use only 'Baz'.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndEverythingWithCustomMessage"}},
			},
			// pattern group with multiple exact paths
			{
				Code: `import x from "foo/baz";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group":   []interface{}{"foo/bar", "foo/baz"},
						"message": "some foo sub-imports are restricted",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternWithCustomMessage", Line: 1, Column: 1}},
			},

			// --- Edge cases: ./ in import source matched by unrooted pattern ---
			{
				Code: `import foo from './foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"foo"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			{
				Code: `import bar from './foo/bar';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"bar"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- Edge case: consecutive slashes matched by unrooted pattern ---
			{
				Code: `import x from "foo//bar";`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"foo"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			// foo/bar does NOT match foo//bar (rooted, strict)
			{
				Code: `import x from "foo//bar";`, Options: []interface{}{"foo//bar"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "path", Line: 1, Column: 1}},
			},

			// --- Edge case: whitespace in source trimmed → matches pattern ---
			{
				Code: `import foo from 'foo ';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"foo"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},
			{
				Code: `import foo from ' foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{"foo"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patterns", Line: 1, Column: 1}},
			},

			// --- Missing messageId coverage ---
			// allowedImportNameWithCustomMessage (path)
			{
				Code: `import { DisallowedObject } from "foo";`, Options: []interface{}{map[string]interface{}{
					"paths": []interface{}{map[string]interface{}{
						"name": "foo", "allowImportNames": []interface{}{"AllowedObject"},
						"message": "Only 'AllowedObject' is allowed.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "allowedImportNameWithCustomMessage"}},
			},
			// allowedImportNamePatternWithCustomMessage (pattern)
			{
				Code: `import { Bar } from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "allowImportNamePattern": "^Foo",
						"message": "Only Foo* allowed.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "allowedImportNamePatternWithCustomMessage"}},
			},
			// everythingWithAllowedImportNamePatternWithCustomMessage
			{
				Code: `import * as All from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "allowImportNamePattern": "^Allow",
						"message": "Only Allow* allowed.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "everythingWithAllowedImportNamePatternWithCustomMessage"}},
			},
			// patternAndEverythingWithRegexImportNameAndCustomMessage
			{
				Code: `import * as All from 'foo';`, Options: []interface{}{map[string]interface{}{
					"patterns": []interface{}{map[string]interface{}{
						"group": []interface{}{"foo"}, "importNamePattern": "^Foo",
						"message": "Don't import Foo*.",
					}},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternAndEverythingWithRegexImportNameAndCustomMessage"}},
			},
		},
	)
}
