package no_restricted_import

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func restrictedImportRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-restricted-import")
	archive := txtarfs.MustParseFile(t, "testdata/extras.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string]string, len(names))
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

func TestNoRestrictedImportLiteralBackslash(t *testing.T) {
	if filepath.Separator != '/' {
		t.Skip("Backslash is a literal filename character only on POSIX hosts.")
	}
	root := restrictedImportRoot(t)
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoRestrictedImportRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     `import '../server/api.js';`,
				FileName: "client/input.js",
				Options:  []any{[]any{root.Dir + `/server\api.js`}},
			},
			{
				Code:     `import '..\\server\\not-found';`,
				FileName: "client/input.js",
				Options:  []any{[]any{root.Dir + `/server/not-found`}},
			},
		}, []rule_tester.InvalidTestCase{{
			Code:     `import '..\\server\\not-found';`,
			FileName: "client/input.js",
			Options:  []any{[]any{root.Dir + `/client/..\server\not-found`}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "restricted", Message: `'..\server\not-found' module is restricted from being used.`,
				Line: 1, Column: 8, EndLine: 1, EndColumn: 31,
			}},
		}})
}

func TestNoRestrictedImportSchema(t *testing.T) {
	for _, tc := range []struct {
		options []any
		valid   bool
	}{
		{nil, true},
		{[]any{[]any{}}, true},
		{[]any{[]any{"fs", map[string]any{"name": []any{}, "message": ""}}}, true},
		{[]any{"fs"}, false},
		{[]any{[]any{"fs"}, []any{}}, false},
		{[]any{[]any{false}}, false},
		{[]any{[]any{map[string]any{}}}, false},
		{[]any{[]any{map[string]any{"name": []any{42}}}}, false},
		{[]any{[]any{map[string]any{"name": "fs", "message": false}}}, false},
		{[]any{[]any{map[string]any{"name": "fs", "extra": true}}}, false},
	} {
		err := NoRestrictedImportRule.Schema.Validate(tc.options)
		if (err == nil) != tc.valid {
			t.Errorf("options %#v: valid = %v, error = %v", tc.options, tc.valid, err)
		}
	}
}

// Expectations checked against eslint-plugin-n v18.3.0 with the TypeScript parser.
func TestNoRestrictedImportExtras(t *testing.T) {
	root := restrictedImportRoot(t)
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoRestrictedImportRule,
		[]rule_tester.ValidTestCase{
			// empty restriction list
			{Code: "import 'fs'; import('./missing');", FileName: "input.js", Options: []any{[]any{}}},
			// empty pattern group
			{Code: "import 'foo';", FileName: "input.js", Options: []any{[]any{map[string]any{"name": []any{}}}}},
			// unmatched and negative-only groups
			{Code: "import 'foo';", FileName: "input.js", Options: []any{[]any{map[string]any{"name": []any{"!foo", "!!foo"}}, map[string]any{"name": []any{"bar", "!foo"}}}}},
			// negative pattern removes an earlier match
			{Code: "import 'foo';", FileName: "input.js", Options: []any{[]any{map[string]any{"name": []any{"*", "!foo"}}}}},
			// non-literal arguments and TypeScript wrappers are ignored
			{Code: "import(name); import(`fs`); import('f' + 's'); import('fs' as string); import('fs'!); import('fs' satisfies string); import(-1);\ntype T = import('fs').T; import x = require('fs');\nrequire('fs'); obj?.import('fs'); obj['import']('fs'); class X { #import() { return 'fs'; } }", FileName: "client/input.ts", Options: []any{[]any{"**"}}},
			// local exports have no source
			{Code: "const value = 1; export { value }; export default value;", FileName: "input.js", Options: []any{[]any{"**"}}},
			// query and fragment stay in the name
			{Code: "import 'pkg?raw#part'; import 'pkg#part';", FileName: "input.js", Options: []any{[]any{"pkg"}}},
			// a single star stays within one segment
			{Code: "import 'pkg/a/b';", FileName: "input.js", Options: []any{[]any{"pkg/*"}}},
			// non-extended syntax does not expand
			{Code: "import 'pkg/file1.js'; import 'pkg/a.js'; import 'foo';", FileName: "input.js", Options: []any{[]any{"pkg/file?.js", "pkg/[ab].js", "pkg/{a,b}.js", "@(foo|bar)", "!(bar)"}}},
			// directory imports retain lexical paths
			{Code: "import '../server';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/server/index.js", "{{root}}", root.Dir)}}},
			// unresolved packages and builtins have no absolute path
			{Code: "import 'missing-package'; import 'fs'; import 'node:fs'; import 'https://example.com/mod.js';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/**", "{{root}}", root.Dir)}}},
			// absolute and name patterns share ordering
			{Code: "import '../server/api.js';", FileName: "client/input.js", Options: []any{[]any{map[string]any{"name": []any{"../server/*", strings.ReplaceAll("!{{root}}/server/api.js", "{{root}}", root.Dir)}}}}},
			// missing map target has no path
			{Code: "import '#missing';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/missing.js", "{{root}}", root.Dir)}}},
			// absolute restriction retains resource query
			{Code: "import '../server/api.js?raw#part';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/server/api.js", "{{root}}", root.Dir)}}},
		},
		[]rule_tester.InvalidTestCase{
			// negation is ordered
			{Code: "import 'foo';", FileName: "input.js", Options: []any{[]any{map[string]any{"name": []any{"!foo", "*"}, "message": "Restricted again."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo' module is restricted from being used. Restricted again.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// a later positive re-adds the target
			{Code: "import 'foo';", FileName: "input.js", Options: []any{[]any{map[string]any{"name": []any{"*", "!foo", "foo"}, "message": "Restricted again."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo' module is restricted from being used. Restricted again.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// separate definitions cannot cancel one another
			{Code: "import 'foo';", FileName: "input.js", Options: []any{[]any{"foo", "!foo"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// first matching definition supplies the message
			{Code: "import 'foo';", FileName: "input.js", Options: []any{[]any{map[string]any{"name": "foo", "message": "First."}, map[string]any{"name": "*", "message": "Second."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo' module is restricted from being used. First.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// empty and whitespace custom messages
			{Code: "import 'foo'; import 'bar';", FileName: "input.js", Options: []any{[]any{map[string]any{"name": "foo", "message": ""}, map[string]any{"name": "bar", "message": " \t"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}, {MessageId: "restricted", Message: "'bar' module is restricted from being used.  \t", Line: 1, Column: 22, EndLine: 1, EndColumn: 27}}},
			// builtins and node protocol are separate names
			{Code: "import 'node:fs'; import 'fs/promises'; import 'fs';", FileName: "input.js", Options: []any{[]any{"node:fs", "fs/promises"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'node:fs' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}, {MessageId: "restricted", Message: "'fs/promises' module is restricted from being used.", Line: 1, Column: 26, EndLine: 1, EndColumn: 39}}},
			// builtin internal modules are included
			{Code: "export * from '_http_agent';", FileName: "input.js", Options: []any{[]any{"_http_agent"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'_http_agent' module is restricted from being used.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28}}},
			// literal import and export forms
			{Code: "import * as ns from 'pkg';\nexport * as other from 'pkg';\nexport { value as renamed } from 'pkg';\nimport 'pkg' with { type: 'json' };", FileName: "input.js", Options: []any{[]any{"pkg"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 1, Column: 21, EndLine: 1, EndColumn: 26}, {MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 2, Column: 24, EndLine: 2, EndColumn: 29}, {MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 3, Column: 34, EndLine: 3, EndColumn: 39}, {MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 4, Column: 8, EndLine: 4, EndColumn: 13}}},
			// type imports and re-exports
			{Code: "import type {Value} from 'pkg';\nexport type {Value} from 'pkg';\nexport type * from 'pkg';\nimport { type Value as Alias } from 'pkg';", FileName: "client/input.ts", Options: []any{[]any{"pkg"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 1, Column: 26, EndLine: 1, EndColumn: 31}, {MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 2, Column: 26, EndLine: 2, EndColumn: 31}, {MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 3, Column: 20, EndLine: 3, EndColumn: 25}, {MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 4, Column: 37, EndLine: 4, EndColumn: 42}}},
			// literal dynamic imports through parentheses
			{Code: "import((('fs')), { with: { type: 'json' } });", FileName: "input.js", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 10, EndLine: 1, EndColumn: 14}}},
			// non-string literals are stringified
			{Code: "import(42); import(0x2an); import(true); import(false); import(null); import(/foo/gi);", FileName: "input.js", Options: []any{[]any{"**"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'42' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 10}, {MessageId: "restricted", Message: "'42' module is restricted from being used.", Line: 1, Column: 20, EndLine: 1, EndColumn: 25}, {MessageId: "restricted", Message: "'true' module is restricted from being used.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39}, {MessageId: "restricted", Message: "'false' module is restricted from being used.", Line: 1, Column: 49, EndLine: 1, EndColumn: 54}, {MessageId: "restricted", Message: "'null' module is restricted from being used.", Line: 1, Column: 64, EndLine: 1, EndColumn: 68}, {MessageId: "restricted", Message: "'/foo/gi' module is restricted from being used.", Line: 1, Column: 78, EndLine: 1, EndColumn: 85}}},
			// JSX expression still visits dynamic imports
			{Code: "const view = <div>{import('fs')}</div>;", FileName: "client/input.tsx", Options: []any{[]any{"fs"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31}}},
			// escaped source and loader suffix
			{Code: "import 'f\\u0073!loader?raw'; export * from 'fs!other'; import('fs!third');", FileName: "input.js", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 44, EndLine: 1, EndColumn: 54}, {MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 63, EndLine: 1, EndColumn: 73}}},
			// queries and URL sources match literally
			{Code: "import 'pkg?raw#part'; import 'https://example.com/mod.js'; import 'data:text/javascript,export{}';", FileName: "input.js", Options: []any{[]any{"pkg?raw#part", "https://example.com/*", "data:*"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg?raw#part' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}, {MessageId: "restricted", Message: "'https://example.com/mod.js' module is restricted from being used.", Line: 1, Column: 31, EndLine: 1, EndColumn: 59}}},
			// empty module name
			{Code: "import '';", FileName: "input.js", Options: []any{[]any{""}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 10}}},
			// globstar matches parent and empty path segments
			{Code: "import '../private/a'; import './private/a'; import 'private/'; import '';", FileName: "input.js", Options: []any{[]any{"**"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../private/a' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}, {MessageId: "restricted", Message: "'./private/a' module is restricted from being used.", Line: 1, Column: 31, EndLine: 1, EndColumn: 44}, {MessageId: "restricted", Message: "'private/' module is restricted from being used.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63}, {MessageId: "restricted", Message: "'' module is restricted from being used.", Line: 1, Column: 72, EndLine: 1, EndColumn: 74}}},
			// globstar spans directories and hidden names
			{Code: "import 'pkg/a/b'; import 'pkg/.hidden';", FileName: "input.js", Options: []any{[]any{"pkg/**"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg/a/b' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}, {MessageId: "restricted", Message: "'pkg/.hidden' module is restricted from being used.", Line: 1, Column: 26, EndLine: 1, EndColumn: 39}}},
			// glob syntax is non-extended
			{Code: "import 'pkg/file?.js'; import 'pkg/[ab].js'; import 'pkg/{a,b}.js'; import '@(foo|bar)';", FileName: "input.js", Options: []any{[]any{"pkg/file?.js", "pkg/[ab].js", "pkg/{a,b}.js", "@(foo|bar)"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg/file?.js' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}, {MessageId: "restricted", Message: "'pkg/[ab].js' module is restricted from being used.", Line: 1, Column: 31, EndLine: 1, EndColumn: 44}, {MessageId: "restricted", Message: "'pkg/{a,b}.js' module is restricted from being used.", Line: 1, Column: 53, EndLine: 1, EndColumn: 67}, {MessageId: "restricted", Message: "'@(foo|bar)' module is restricted from being used.", Line: 1, Column: 76, EndLine: 1, EndColumn: 88}}},
			// multiline literal and UTF-16 ranges
			{Code: "const emoji = '💡'; import('模块/💡');\nimport('multi\\\nline');", FileName: "input.js", Options: []any{[]any{"模块/*", "multiline"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'模块/💡' module is restricted from being used.", Line: 1, Column: 28, EndLine: 1, EndColumn: 35}, {MessageId: "restricted", Message: "'multiline' module is restricted from being used.", Line: 2, Column: 8, EndLine: 3, EndColumn: 6}}},
			// unresolved relative path fallback
			{Code: "import '../server/not-found';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/server/**", "{{root}}", root.Dir)}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../server/not-found' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}}},
			// dot and parent directory fallback
			{Code: "import '.'; import '..';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/client", "{{root}}", root.Dir), strings.ReplaceAll("{{root}}", "{{root}}", root.Dir)}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'.' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}, {MessageId: "restricted", Message: "'..' module is restricted from being used.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24}}},
			// relative files use resolved extensions
			{Code: "import '../server/api';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/server/api.js", "{{root}}", root.Dir)}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../server/api' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
			// resolved package and package subpath
			{Code: "import 'pkg'; export * from 'pkg/sub';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/node_modules/pkg/**", "{{root}}", root.Dir)}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}, {MessageId: "restricted", Message: "'pkg/sub' module is restricted from being used.", Line: 1, Column: 29, EndLine: 1, EndColumn: 38}}},
			// name restrictions are independent of resolution
			{Code: "import 'missing-package';", FileName: "client/input.js", Options: []any{[]any{"missing-package"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'missing-package' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
			// negative path only removes a matching target
			{Code: "import '../server/api.js';", FileName: "client/input.js", Options: []any{[]any{map[string]any{"name": []any{"../server/*", strings.ReplaceAll("!{{root}}/server/other.js", "{{root}}", root.Dir)}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../server/api.js' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
			// positive path after name exclusion
			{Code: "import '../server/api.js';", FileName: "client/input.js", Options: []any{[]any{map[string]any{"name": []any{"**", "!../server/*", strings.ReplaceAll("{{root}}/server/api.js", "{{root}}", root.Dir)}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../server/api.js' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
			// package import maps
			{Code: "import '#entry';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/server/**", "{{root}}", root.Dir)}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'#entry' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// TypeScript aliases and extension mapping
			{Code: "import {value} from 'alias/mapped.js';", FileName: "client/input.ts", Options: []any{[]any{strings.ReplaceAll("{{root}}/server/mapped.ts", "{{root}}", root.Dir)}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'alias/mapped.js' module is restricted from being used.", Line: 1, Column: 21, EndLine: 1, EndColumn: 38}}},
			// type and runtime conditions choose different paths
			{Code: "import type {Value} from 'pkg/types'; import 'pkg/types'; export type * from 'pkg/types';", FileName: "client/input.ts", Options: []any{[]any{strings.ReplaceAll("{{root}}/node_modules/pkg/types.d.ts", "{{root}}", root.Dir)}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg/types' module is restricted from being used.", Line: 1, Column: 26, EndLine: 1, EndColumn: 37}, {MessageId: "restricted", Message: "'pkg/types' module is restricted from being used.", Line: 1, Column: 78, EndLine: 1, EndColumn: 89}}},
			// shared extension settings
			{Code: "import '../server/custom';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/server/custom.ext", "{{root}}", root.Dir)}}, Settings: map[string]any{"node": map[string]any{"tryExtensions": []any{".ext"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../server/custom' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
			// shared additional base directories
			{Code: "import 'extra';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/vendor/node_modules/extra/**", "{{root}}", root.Dir)}}, Settings: map[string]any{"node": map[string]any{"resolvePaths": []any{strings.ReplaceAll("{{root}}/vendor", "{{root}}", root.Dir)}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'extra' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}}},
			// shared module search directories
			{Code: "import 'custom';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/custom_modules/custom/**", "{{root}}", root.Dir)}}, Settings: map[string]any{"node": map[string]any{"resolverConfig": map[string]any{"modules": []any{"custom_modules"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'custom' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// absolute restriction matches full resource query
			{Code: "import '../server/api.js?raw#part';", FileName: "client/input.js", Options: []any{[]any{strings.ReplaceAll("{{root}}/server/api.js?raw#part", "{{root}}", root.Dir)}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../server/api.js?raw#part' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 35}}},
		},
	)
}

// Additional literal and AST expectations checked against eslint-plugin-n v18.3.0.
func TestNoRestrictedImportLiteralEdges(t *testing.T) {
	root := restrictedImportRoot(t)
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoRestrictedImportRule,
		[]rule_tester.ValidTestCase{ // literal names do not ignore final line terminators
			{Code: "import('fs\\n'); import('fs\\r'); import('fs\\r\\n'); import('fs\\u2028'); import('fs\\u2029');", FileName: "client/edges.ts", Options: []any{[]any{"fs"}}},
			// module matching is case sensitive
			{Code: "import 'FS'; import 'Node:fs';", FileName: "client/edges.ts", Options: []any{[]any{"fs", "node:fs"}}},
			// parenthesized type assertions are not literals
			{Code: "import(('fs' as const)); import((<string>'fs'));", FileName: "client/edges.ts", Options: []any{[]any{"fs"}}}},
		[]rule_tester.InvalidTestCase{ // escaped surrogate pair equals its literal spelling
			{Code: "import '\\uD83D\\uDCA1';", FileName: "client/edges.ts", Options: []any{[]any{"💡"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'💡' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
			// glob containing a supplementary character
			{Code: "import 'pkg/\\uD83D\\uDCA1';", FileName: "client/edges.ts", Options: []any{[]any{"*/💡"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'pkg/💡' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
			// regexp flags are canonicalized
			{Code: "import(/foo/ig); import(/[/]/yig);", FileName: "client/edges.ts", Options: []any{[]any{"**"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'/foo/gi' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}, {MessageId: "restricted", Message: "'/[/]/giy' module is restricted from being used.", Line: 1, Column: 25, EndLine: 1, EndColumn: 33}}},
			// large numbers and BigInts preserve JavaScript values
			{Code: "import(0x1000000000000001); import(1e21); import(1e309); import(0xffffffffffffffffffffn); import(1_000n);", FileName: "client/edges.ts", Options: []any{[]any{"**"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'1152921504606847000' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}, {MessageId: "restricted", Message: "'1e+21' module is restricted from being used.", Line: 1, Column: 36, EndLine: 1, EndColumn: 40}, {MessageId: "restricted", Message: "'Infinity' module is restricted from being used.", Line: 1, Column: 50, EndLine: 1, EndColumn: 55}, {MessageId: "restricted", Message: "'1208925819614629174706175' module is restricted from being used.", Line: 1, Column: 65, EndLine: 1, EndColumn: 88}, {MessageId: "restricted", Message: "'1000' module is restricted from being used.", Line: 1, Column: 98, EndLine: 1, EndColumn: 104}}},
			// module names preserve Unicode whitespace and null
			{Code: "import '\\u00a0pkg\\u00a0'; import('pkg\\0suffix');", FileName: "client/edges.ts", Options: []any{[]any{" pkg ", "pkg\u0000suffix"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "' pkg ' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}, {MessageId: "restricted", Message: "'pkg\u0000suffix' module is restricted from being used.", Line: 1, Column: 34, EndLine: 1, EndColumn: 47}}},
			// a loader-only name becomes empty
			{Code: "import '!loader';", FileName: "client/edges.ts", Options: []any{[]any{""}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// nested dynamic imports inspect only literal sources
			{Code: "const nested = import(import('fs'));", FileName: "client/edges.ts", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 30, EndLine: 1, EndColumn: 34}}}},
	)
}
