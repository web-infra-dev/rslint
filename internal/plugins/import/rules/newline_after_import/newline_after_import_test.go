package newline_after_import_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/newline_after_import"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNewlineAfterImportRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&newline_after_import.NewlineAfterImportRule,
		[]rule_tester.ValidTestCase{
			// ===== Import declarations =====

			// Consecutive imports (no code after)
			{Code: "import path from 'path';\nimport foo from 'foo';\n"},
			// Inline consecutive imports
			{Code: "import path from 'path';import foo from 'foo';\n"},
			// Consecutive imports + blank line + code
			{Code: "import path from 'path';import foo from 'foo';\n\nvar bar = 42;"},
			// Single import + blank line + code
			{Code: "import foo from 'foo';\n\nvar bar = 'bar';"},
			// Import with no code after (end of file)
			{Code: "import foo from 'foo';"},
			// Empty file
			{Code: ""},
			// Various import forms
			{Code: "import * as foo from 'foo';\n\nvar x = 1;"},
			{Code: "import 'side-effect';\n\nvar x = 1;"},
			{Code: "import type { Foo } from 'foo';\n\nvar x = 1;"},
			// Import groups separated by code (each group has its own blank line)
			{Code: "import foo from 'foo';\n\nvar a = 123;\n\nimport { bar } from './bar-lib';"},
			// Import followed by export block
			{Code: "import stub from './stub';\n\nexport { stub }"},

			// ===== count option =====

			{Code: "import foo from 'foo';\n\n\nvar bar = 'bar';", Options: map[string]interface{}{"count": 2.0}},
			{Code: "import foo from 'foo';\n\n\n\n\nvar bar = 'bar';", Options: map[string]interface{}{"count": 4.0}},
			// More lines than count is OK when exactCount is false
			{Code: "import foo from 'foo';\n\n\nvar bar = 'bar';", Options: map[string]interface{}{"count": 1.0}},

			// ===== exactCount option =====

			{Code: "import foo from 'foo';\n\n\nvar bar = 'bar';", Options: map[string]interface{}{"count": 2.0, "exactCount": true}},
			{Code: "import foo from 'foo';\n\nvar bar = 'bar';", Options: map[string]interface{}{"count": 1.0, "exactCount": true}},

			// ===== considerComments option =====

			// Enough blank lines before comment
			{
				Code:    "import foo from 'foo';\n\n// Some random comment\nvar bar = 'bar';",
				Options: map[string]interface{}{"count": 1.0, "exactCount": true, "considerComments": true},
			},
			{
				Code:    "import foo from 'foo';\n\n\n// Some random comment\nvar bar = 'bar';",
				Options: map[string]interface{}{"count": 2.0, "exactCount": true, "considerComments": true},
			},
			// Without considerComments, comments don't affect gap calculation
			{
				Code:    "import foo from 'foo';\n\n// Some random comment\nvar bar = 'bar';",
				Options: map[string]interface{}{"count": 2.0, "exactCount": true},
			},
			{
				Code:    "import foo from 'foo';\n// Some random comment\nvar bar = 'bar';",
				Options: map[string]interface{}{"count": 1.0, "exactCount": true},
			},
			// Multiline block comment between import and code (without considerComments)
			{Code: "import path from 'path';\nimport foo from 'foo';\n/**\n * some multiline comment here\n * another line of comment\n**/\nvar bar = 42;"},
			// Inline import + multiline comment with considerComments
			{
				Code:    "import path from 'path';import foo from 'foo';\n\n/**\n * some multiline comment here\n * another line of comment\n**/\nvar bar = 42;",
				Options: map[string]interface{}{"considerComments": true},
			},
			// Single-line comment after consecutive imports (without considerComments)
			{Code: "import path from 'path';\nimport foo from 'foo';\n\n// Some random single line comment\nvar bar = 42;"},
			// Leading comment before import (should not affect)
			{Code: "/**\n * A leading comment\n */\nimport foo from 'foo';\n\n// Some random comment\nexport {foo};", Options: map[string]interface{}{"count": 2.0, "exactCount": true}},

			// ===== Require calls =====

			// Consecutive requires (no code after)
			{Code: "var path = require('path');\nvar foo = require('foo');\n"},
			// Single require (no code after)
			{Code: "require('foo');"},
			// Require + blank line + code
			{Code: "var foo = require('foo-module');\n\nvar foo = 'bar';"},
			// Require with count
			{Code: "var foo = require('foo-module');\n\n\nvar foo = 'bar';", Options: map[string]interface{}{"count": 2.0}},
			{Code: "var foo = require('foo-module');\n\n\n\n\nvar foo = 'bar';", Options: map[string]interface{}{"count": 4.0}},
			{Code: "var foo = require('foo-module');\n\n\n\n\nvar foo = 'bar';", Options: map[string]interface{}{"count": 4.0, "exactCount": true}},
			// Bare require + code
			{Code: "require('foo-module');\n\nvar foo = 'bar';"},
			// Require groups separated by code
			{Code: "var foo = require('foo-module');\n\nvar a = 123;\n\nvar bar = require('bar-lib');"},
			// Require with considerComments
			{
				Code:    "var foo = require('foo-module');\n\n\n// Some random comment\nvar foo = 'bar';",
				Options: map[string]interface{}{"count": 2.0, "considerComments": true},
			},
			{
				Code:    "var foo = require('foo-module');\n\n\n/**\n * Test comment\n */\nvar foo = 'bar';",
				Options: map[string]interface{}{"count": 2.0, "considerComments": true},
			},
			// exactCount + considerComments for require
			{
				Code:    "const foo = require('foo');\n\n\n// some random comment\nconst bar = function() {};",
				Options: map[string]interface{}{"count": 2.0, "exactCount": true, "considerComments": true},
			},
			// exactCount without considerComments — comment in gap doesn't matter
			{
				Code:    "var foo = require('foo-module');\n\n// Some random comment\n\n\nvar foo = 'bar';",
				Options: map[string]interface{}{"count": 4.0, "exactCount": true},
			},
			// exactCount + considerComments — comment in gap measured correctly
			{
				Code:    "var foo = require('foo-module');\n\n\n\n\n// Some random comment\nvar foo = 'bar';",
				Options: map[string]interface{}{"count": 4.0, "exactCount": true, "considerComments": true},
			},

			// ===== Require scope boundaries (should NOT be detected as top-level) =====

			// Require inside function
			{Code: "function x() { require('baz'); }"},
			// Require inside arrow function
			{Code: "const x = () => require('baz');\nvar y = 1;"},
			// Require inside nested function with logic
			{Code: "function a() {\n  var assign = Object.assign || require('object-assign');\n  var foo = require('foo');\n  var bar = 42;\n}"},
			// Require inside if block
			{Code: "if (true) {\n  var foo = require('foo');\n  foo();\n}"},
			// Require inside switch (case body is NOT a Block — require IS top-level per ESLint)
			{Code: "switch ('foo') { case 'bar': require('baz'); }"},
			// Require as argument inside function (arrow function scopes it)
			{Code: "const x = () => require('baz')\n    , y = () => require('bar')"},
			// Require in binary expression inside arrow function
			{Code: "const x = () => require('baz') && require('bar')"},
			// Require as standalone argument (no code after — valid)
			{Code: "a(require('b'), require('c'), require('d'));"},
			// Require inside object literal
			{Code: "var x = { foo: require('foo') };\nvar y = 1;"},
			// Complex switch inside function
			{Code: "function foo() {\n  switch (renderData.modalViewKey) {\n    case 'value':\n      var bar = require('bar');\n      return bar(renderData, options)\n    default:\n      return renderData.mainModalContent.clone()\n  }\n}"},
			// Arrow with block body
			{Code: "const x = () => { return require('baz'); };\nvar y = 1;"},
			// Nested arrow
			{Code: "const x = () => () => require('baz');\nvar y = 1;"},
			// Arrow in ternary (requires scoped inside arrows)
			{Code: "var foo = condition ? () => require('a') : () => require('b');\nvar bar = 42;"},

			// ===== Multi-line require + considerComments (window anchored to require call) =====

			// Multi-line require stmt: require call ends line 0, stmt ends line 1.
			// Comment on line 3 is within window [0, 2] (anchored to require call end).
			// Gap: commentLine(3) - stmtEndLine(1) = 2 = expected(2) → valid.
			{
				Code:    "var x = require('foo') +\n    123;\n\n// comment\nvar bar = 42;",
				Options: map[string]interface{}{"count": 1.0, "considerComments": true},
			},
			// Multi-line require stmt with count:2, enough gap before comment.
			{
				Code:    "var x = require('foo') +\n    123;\n\n\n// comment\nvar bar = 42;",
				Options: map[string]interface{}{"count": 2.0, "considerComments": true},
			},

			// ===== TSImportEqualsDeclaration =====

			{Code: "import { ReturnValue } from 'runner';\nimport runner = require('runner');"},
			{Code: "import runner = require('runner');\nimport { ReturnValue } from 'runner';"},
			// Mixed import, import=require, and another import
			{Code: "import { ReturnValue } from 'runner';\nimport runner = require('runner');\nimport { OtherValue } from 'other';"},
			// export import (should be skipped — it's a re-export, not an import)
			{Code: "export import a = obj;\nf(a);"},
			// Internal module import followed by export import
			{Code: "import { ns } from 'namespace';\nimport Bar = ns.baz.foo.Bar;\n\nexport import Foo = ns.baz.bar.Foo;"},

			// ===== Arrow + require in considerComments valid =====

			{
				Code:    "const x = () => require('baz')\n    , y = () => require('bar')\n\n// some comment here\n",
				Options: map[string]interface{}{"considerComments": true},
			},
			{
				Code:    "const x = () => require('baz') && require('bar')\n\n// Some random single line comment\nvar bar = 42;",
				Options: map[string]interface{}{"considerComments": true},
			},
			{
				Code:    "const x = () => require('baz') && require('bar')\n\n// Some random single line comment\nvar bar = 42;",
				Options: map[string]interface{}{"considerComments": true, "count": 1.0, "exactCount": true},
			},

			// ===== Decorator class =====

			// Decorated class with sufficient gap
			{Code: "import foo from 'foo';\n\n@SomeDecorator(foo)\nclass Foo {}"},
			// Export default decorated class with sufficient gap
			{Code: "import foo from 'foo';\n\n@SomeDecorator(foo)\nexport default class Test {}"},
		},
		[]rule_tester.InvalidTestCase{
			// ===== Basic import errors =====

			// Single import → code (no blank line)
			{
				Code:   "import foo from 'foo';\nexport default function() {};",
				Output: []string{"import foo from 'foo';\n\nexport default function() {};"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// count: 2 (too few lines)
			{
				Code:    "import foo from 'foo';\n\nexport default function() {};",
				Output:  []string{"import foo from 'foo';\n\n\nexport default function() {};"},
				Options: map[string]interface{}{"count": 2.0},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// Multiple imports, last one needs newline
			{
				Code:   "import path from 'path';\nimport foo from 'foo';\nvar bar = 42;",
				Output: []string{"import path from 'path';\nimport foo from 'foo';\n\nvar bar = 42;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 2, Column: 1,
				}},
			},
			// Inline imports on same line
			{
				Code:   "import path from 'path';import foo from 'foo';var bar = 42;",
				Output: []string{"import path from 'path';import foo from 'foo';\n\nvar bar = 42;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 25,
				}},
			},
			// Two import groups separated by code — both need newlines
			{
				Code:   "import foo from 'foo';\nvar a = 123;\n\nimport { bar } from './bar-lib';\nvar b=456;",
				Output: []string{"import foo from 'foo';\n\nvar a = 123;\n\nimport { bar } from './bar-lib';\n\nvar b=456;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "newlineAfterImport", Line: 1, Column: 1},
					{MessageId: "newlineAfterImport", Line: 4, Column: 1},
				},
			},
			// With count: 1 option explicitly
			{
				Code:    "import foo from 'foo';\nexport default function() {};",
				Output:  []string{"import foo from 'foo';\n\nexport default function() {};"},
				Options: map[string]interface{}{"count": 1.0},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},

			// ===== Require errors =====

			// Single require → code
			{
				Code:   "var foo = require('foo-module');\nvar something = 123;",
				Output: []string{"var foo = require('foo-module');\n\nvar something = 123;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 1, Column: 1,
				}},
			},
			// Consecutive requires → code
			{
				Code:   "var path = require('path');\nvar foo = require('foo');\nvar bar = 42;",
				Output: []string{"var path = require('path');\nvar foo = require('foo');\n\nvar bar = 42;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 2, Column: 1,
				}},
			},
			// Two require groups separated by non-require code
			{
				Code:   "var foo = require('foo-module');\nvar a = 123;\n\nvar bar = require('bar-lib');\nvar b=456;",
				Output: []string{"var foo = require('foo-module');\n\nvar a = 123;\n\nvar bar = require('bar-lib');\n\nvar b=456;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "newlineAfterRequire", Line: 1, Column: 1},
					{MessageId: "newlineAfterRequire", Line: 4, Column: 1},
				},
			},
			// Bare require followed by bare require followed by code
			{
				Code:   "var foo = require('foo-module');\nvar a = 123;\n\nrequire('bar-lib');\nvar b=456;",
				Output: []string{"var foo = require('foo-module');\n\nvar a = 123;\n\nrequire('bar-lib');\n\nvar b=456;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "newlineAfterRequire", Line: 1, Column: 1},
					{MessageId: "newlineAfterRequire", Line: 4, Column: 1},
				},
			},
			// Require in binary expression (not inside function — IS top-level)
			{
				Code:   "var assign = Object.assign || require('object-assign');\nvar foo = require('foo');\nvar bar = 42;",
				Output: []string{"var assign = Object.assign || require('object-assign');\nvar foo = require('foo');\n\nvar bar = 42;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 2, Column: 1,
				}},
			},
			// Mixed: require, then foo(require, require, require), then require, then code
			{
				Code:   "require('a');\nfoo(require('b'), require('c'), require('d'));\nrequire('d');\nvar foo = 'bar';",
				Output: []string{"require('a');\nfoo(require('b'), require('c'), require('d'));\nrequire('d');\n\nvar foo = 'bar';"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 3, Column: 1,
				}},
			},
			// Multi-line function call containing requires
			{
				Code:   "require('a');\nfoo(\nrequire('b'),\nrequire('c'),\nrequire('d')\n);\nvar foo = 'bar';",
				Output: []string{"require('a');\nfoo(\nrequire('b'),\nrequire('c'),\nrequire('d')\n);\n\nvar foo = 'bar';"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 6, Column: 1,
				}},
			},
			// Ternary with direct requires (NOT inside arrow) — IS top-level
			{
				Code:   "var foo = condition ? require('a') : require('b');\nvar bar = 42;",
				Output: []string{"var foo = condition ? require('a') : require('b');\n\nvar bar = 42;"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 1, Column: 1,
				}},
			},

			// ===== exactCount =====

			// Too few lines with exactCount (fixable)
			{
				Code:    "import foo from 'foo';\n\nexport default function() {};",
				Output:  []string{"import foo from 'foo';\n\n\nexport default function() {};"},
				Options: map[string]interface{}{"count": 2.0, "exactCount": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// Too many lines with exactCount (no fix)
			{
				Code:    "import foo from 'foo';\n\n\n\nexport default function() {};",
				Options: map[string]interface{}{"count": 2.0, "exactCount": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// Way too many lines with exactCount (no fix)
			{
				Code:    "import foo from 'foo';\n\n\n\n\nexport default function() {};",
				Options: map[string]interface{}{"count": 2.0, "exactCount": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// exactCount with count=1 (same line — no newline at all)
			{
				Code:    "import foo from 'foo';export default function() {};",
				Output:  []string{"import foo from 'foo';\n\nexport default function() {};"},
				Options: map[string]interface{}{"count": 1.0, "exactCount": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// exactCount too many for require (no fix)
			{
				Code:    "const foo = require('foo');\n\n\n\nconst bar = function() {};",
				Options: map[string]interface{}{"count": 2.0, "exactCount": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 1, Column: 1,
				}},
			},

			// ===== considerComments =====

			// Comment too close to import
			{
				Code:    "import path from 'path';\nimport foo from 'foo';\n// Some random single line comment\nvar bar = 42;",
				Output:  []string{"import path from 'path';\nimport foo from 'foo';\n\n// Some random single line comment\nvar bar = 42;"},
				Options: map[string]interface{}{"considerComments": true, "count": 1.0},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 2, Column: 1,
				}},
			},
			// Multiline block comment too close to import
			{
				Code:    "import path from 'path';\nimport foo from 'foo';\n/**\n * some multiline comment here\n * another line of comment\n**/\nvar bar = 42;",
				Output:  []string{"import path from 'path';\nimport foo from 'foo';\n\n/**\n * some multiline comment here\n * another line of comment\n**/\nvar bar = 42;"},
				Options: map[string]interface{}{"considerComments": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 2, Column: 1,
				}},
			},
			// considerComments with count: 2 (require + multiline comment)
			{
				Code:    "var foo = require('foo-module');\n\n/**\n * Test comment\n */\nvar foo = 'bar';",
				Output:  []string{"var foo = require('foo-module');\n\n\n/**\n * Test comment\n */\nvar foo = 'bar';"},
				Options: map[string]interface{}{"considerComments": true, "count": 2.0},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 1, Column: 1,
				}},
			},
			// considerComments + exactCount: too few lines before comment
			{
				Code:    "import foo from 'foo';\n// some random comment\nexport default function() {};",
				Output:  []string{"import foo from 'foo';\n\n\n// some random comment\nexport default function() {};"},
				Options: map[string]interface{}{"count": 2.0, "exactCount": true, "considerComments": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// considerComments + exactCount: too many lines before comment (no fix)
			{
				Code:    "import foo from 'foo';\n\n\n\n// some random comment\nexport default function() {};",
				Options: map[string]interface{}{"count": 2.0, "exactCount": true, "considerComments": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// Same-line trailing comment with considerComments
			{
				Code:    "import foo from 'foo';// some random comment\nexport default function() {};",
				Output:  []string{"import foo from 'foo';\n\n// some random comment\nexport default function() {};"},
				Options: map[string]interface{}{"count": 1.0, "exactCount": true, "considerComments": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},
			// Consecutive requires + comment with considerComments + count: 2
			{
				Code:    "var foo = require('foo-module');\nvar foo = require('foo-module');\n\n// Some random comment\nvar foo = 'bar';",
				Output:  []string{"var foo = require('foo-module');\nvar foo = require('foo-module');\n\n\n// Some random comment\nvar foo = 'bar';"},
				Options: map[string]interface{}{"considerComments": true, "count": 2.0},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 2, Column: 1,
				}},
			},
			// exactCount + considerComments: too many blank lines before comment (no fix, fallback)
			{
				Code:    "import foo from 'foo';\n\n\n// Some random single line comment\nvar bar = 42;",
				Options: map[string]interface{}{"considerComments": true, "count": 1.0, "exactCount": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},

			// ===== Multi-line import =====

			// Multi-line import, column reports at col 0 of end line
			{
				Code:    "\nimport { A, B, C, D } from\n'../path/to/my/module/in/very/far/directory'\n// some comment\nvar foo = 'bar';\n",
				Output:  []string{"\nimport { A, B, C, D } from\n'../path/to/my/module/in/very/far/directory'\n\n// some comment\nvar foo = 'bar';\n"},
				Options: map[string]interface{}{"considerComments": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 3, Column: 1,
				}},
			},

			// ===== Mixed import and require =====

			// Import followed by require (both need newlines)
			{
				Code:   "import foo from 'foo';\nvar bar = require('bar');\nvar baz = 42;",
				Output: []string{"import foo from 'foo';\n\nvar bar = require('bar');\n\nvar baz = 42;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "newlineAfterImport", Line: 1, Column: 1},
					{MessageId: "newlineAfterRequire", Line: 2, Column: 1},
				},
			},

			// ===== Multi-line require + considerComments =====

			// Multi-line require stmt: require call ends line 0, stmt ends line 1.
			// Comment on line 2 is within window [0, 2]. Gap: 2 - 1 = 1 < 2 → error.
			{
				Code:    "var x = require('foo') +\n    123;\n// comment\nvar bar = 42;",
				Output:  []string{"var x = require('foo') +\n    123;\n\n// comment\nvar bar = 42;"},
				Options: map[string]interface{}{"count": 1.0, "considerComments": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 2, Column: 1,
				}},
			},
			// Multi-line require stmt with count:2. Comment too close.
			{
				Code:    "var x = require('foo') +\n    123;\n\n// comment\nvar bar = 42;",
				Output:  []string{"var x = require('foo') +\n    123;\n\n\n// comment\nvar bar = 42;"},
				Options: map[string]interface{}{"count": 2.0, "considerComments": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterRequire", Line: 2, Column: 1,
				}},
			},

			// ===== Decorator class =====

			// Decorated class immediately after import (no blank line)
			{
				Code:   "import foo from 'foo';\n@SomeDecorator(foo)\nclass Foo {}",
				Output: []string{"import foo from 'foo';\n\n@SomeDecorator(foo)\nclass Foo {}"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "newlineAfterImport", Line: 1, Column: 1,
				}},
			},

			// ===== Comment between consecutive imports (with considerComments) =====

			// Both imports report: first because comment is too close, second because no blank line before code
			{
				Code:    "import path from 'path';\n// comment\nimport foo from 'foo';\nvar bar = 42;",
				Output:  []string{"import path from 'path';\n\n// comment\nimport foo from 'foo';\n\nvar bar = 42;"},
				Options: map[string]interface{}{"considerComments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "newlineAfterImport", Line: 1, Column: 1},
					{MessageId: "newlineAfterImport", Line: 3, Column: 1},
				},
			},
		},
	)
}
