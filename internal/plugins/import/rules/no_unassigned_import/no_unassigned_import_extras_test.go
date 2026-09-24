package no_unassigned_import_test

import (
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_unassigned_import"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnassignedImportExtras(t *testing.T) {
	root := fixtures.GetRootDir()
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &no_unassigned_import.NoUnassignedImportRule,
		[]rule_tester.ValidTestCase{
			// Only a standalone, non-optional require with one string literal
			// argument is checked. Templates are not ESTree Literals.
			{Code: "require(); require('foo', 'bar'); require(name); require(`foo`); require(1); require(...args);"},
			{Code: "loader.require('foo'); loader['require']('foo'); new require('foo');"},
			{Code: "require?.('foo'); (require?.('foo')); (require?.('foo'))('bar');"},
			{Code: "void require('foo'); await require('foo'); require('foo'), 0; value = require('foo');"},
			{Code: "const load = () => require('foo'); function loadMore() { return require('foo'); }"},
			{Code: "for (require('foo'); ok; require('bar')) {}"},
			{Code: "import('foo'); export {} from 'foo'; export * from 'foo';"},
			// Authored TypeScript wrappers remain visible to the upstream rule.
			{Code: "require('foo') as unknown; require('foo') satisfies unknown; require('foo')!; (require as Function)('foo'); require!('foo'); require('foo' as string);"},
			{Code: "import type { T } from 'foo'; import type * as NS from 'foo'; import type T from 'foo'; import value, {} from 'foo'; import value = require('foo');"},
			{Code: "import type from 'foo'; import type, {} from 'bar'; import from, {} from 'baz';"},
			{Code: "const element = <Component value={require('foo')} />;", Tsx: true},
			{Code: "class C { #require() {} method() { this.#require('foo'); } }"},
			{
				Code:    "import ''; require('');",
				Options: map[string]any{"allow": []any{""}},
			},
			{
				Code:     "import './nested/../style.css'; require(('./style.css'));",
				FileName: "src/app.ts",
				Options:  map[string]any{"allow": []any{"src/*.css"}},
			},
			{
				Code:     "import './style.css';",
				FileName: "src/app.ts",
				Options:  map[string]any{"allow": []any{tspath.ResolvePath(root.Dir, "src/*.css")}},
			},
			{
				Code:    "import " + strconv.Quote(tspath.ResolvePath(root.Dir, "styles/app.css")) + ";",
				Options: map[string]any{"allow": []any{"styles/*.css"}},
			},
			{
				Code:     "import './style.css';",
				FileName: "src/app.ts",
				Options:  map[string]any{"allow": []any{"./src/../src/*.css"}},
			},
			{
				Code:    "import 'styles/app.scss'; require('styles/reset.css');",
				Options: map[string]any{"allow": []any{"styles/*.{css,scss}"}},
			},
			{
				Code:    "import '@scope/polyfill'; import 'polyfill';",
				Options: map[string]any{"allow": []any{"@scope/@(polyfill|setup)", "!disallowed"}},
			},
			{
				Code:    "import './.setup/init.js';",
				Options: map[string]any{"allow": []any{"**/.setup/*.js"}},
			},
			{
				Code:    "import '\\u0070'; require('\\x70');",
				Options: map[string]any{"allow": []any{"p"}},
			},
			{
				// A glob question mark matches one UTF-16 code unit.
				Code:    "import '😀';",
				Options: map[string]any{"allow": []any{"??"}},
			},
			{
				Code:    "import type {} from 'types'; import {} from 'types';",
				Options: map[string]any{"allow": []any{"types"}},
			},
			{
				Code:    "import 'polyfill'; require('polyfill');",
				Options: map[string]any{"allow": []any{"unrelated", "polyfill"}, "devDependencies": false, "optionalDependencies": []any{}, "peerDependencies": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "r\\u0065quire('foo');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:     "/** @type {any} */ (require('foo'));",
				FileName: "cast.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code:     "(/** @type {any} */ (require))(/** @type {string} */ ('foo'));",
				FileName: "cast.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 62},
				},
			},
			{
				Code:    "import '😀';",
				Options: map[string]any{"allow": []any{"?"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "import 'a'; import 'b'; require('c');",
				Options: map[string]any{"allow": []any{"a", "c"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 13, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "import '/side-effects/polyfill.js';",
				Options: map[string]any{"allow": []any{"**/*.css"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code:    "import 'foo';",
				Options: map[string]any{"allow": []any{"!foo"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:     "import './styles';",
				FileName: "src/app.ts",
				Options:  map[string]any{"allow": []any{"src/styles/"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "import {} from 'foo';\nimport type {} from 'types';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
					{MessageId: "", Message: unassignedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 29},
				},
			},
			{
				Code: "/* 😀 */ import 'é';\n( /* comment */ require(\n  'foo'\n));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 21},
					{MessageId: "", Message: unassignedMessage, Line: 2, Column: 17, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "(require)(('foo'));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "function load(require) { require('foo'); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 26, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code: "class C { static { require<string>('foo'); } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code: "import 'data.json' with { type: 'json' };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code:    "import 'foo';",
				Options: map[string]any{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "require('foo');",
				Options: map[string]any{"allow": []any{}, "devDependencies": true, "optionalDependencies": false, "peerDependencies": []any{"**"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "import './.setup/init.js';",
				Options: map[string]any{"allow": []any{"**/*.js"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    "import 'styles/app.css';",
				Options: map[string]any{"allow": []any{"*.css", "#styles/**", "STYLES/**"}, "devDependencies": []any{"**"}, "optionalDependencies": true, "peerDependencies": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
		},
	)
}
