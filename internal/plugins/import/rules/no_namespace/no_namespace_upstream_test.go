package no_namespace_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_namespace"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func namespaceError(line, column, endLine, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "",
		Message:   "Unexpected namespace import.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}}
}

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-namespace.js
func TestNoNamespaceUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_namespace.NoNamespaceRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { a, b } from 'foo';`},
			{Code: `import { a, b } from './foo';`},
			{Code: `import bar from 'bar';`},
			{Code: `import bar from './bar';`},
			{Code: `import * as bar from './ignored-module.ext';`, Options: map[string]any{"ignore": []any{"*.ext"}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `import * as foo from 'foo';`, Errors: namespaceError(1, 8, 1, 16)},
			{Code: `import defaultExport, * as foo from 'foo';`, Errors: namespaceError(1, 23, 1, 31)},
			{Code: `import * as foo from './foo';`, Errors: namespaceError(1, 8, 1, 16)},
			// FIX_TESTS (ESLint 5+).
			{
				Code: `import * as foo from './foo';
      florp(foo.bar);
      florp(foo['baz']);`,
				Output: []string{`import { bar, baz } from './foo';
      florp(bar);
      florp(baz);`},
				Errors: namespaceError(1, 8, 1, 16),
			},
			{
				Code: `import * as foo from './foo';
      const bar = 'name conflict';
      const baz = 'name conflict';
      const foo_baz = 'name conflict';
      florp(foo.bar);
      florp(foo['baz']);`,
				Output: []string{`import { bar as foo_bar, baz as foo_baz_1 } from './foo';
      const bar = 'name conflict';
      const baz = 'name conflict';
      const foo_baz = 'name conflict';
      florp(foo_bar);
      florp(foo_baz_1);`},
				Errors: namespaceError(1, 8, 1, 16),
			},
			{
				Code: `import * as foo from './foo';
      function func(arg) {
        florp(foo.func);
        florp(foo['arg']);
      }`,
				Output: []string{`import { func as foo_func, arg as foo_arg } from './foo';
      function func(arg) {
        florp(foo_func);
        florp(foo_arg);
      }`},
				Errors: namespaceError(1, 8, 1, 16),
			},
		})
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-namespace.md
func TestNoNamespaceUpstreamDocs(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_namespace.NoNamespaceRule,
		[]rule_tester.ValidTestCase{
			{Code: `import defaultExport from './foo'`},
			{Code: `import { a, b }  from './bar'`},
			{Code: `import defaultExport, { a, b }  from './foobar'`},
			{
				Code:    "/* eslint import/no-namespace: [\"error\", {ignore: ['*.ext']}] */\nimport * as bar from './ignored-module.ext';",
				Options: map[string]any{"ignore": []any{"*.ext"}},
			},
		}, []rule_tester.InvalidTestCase{
			{Code: `import * as foo from 'foo';`, Errors: namespaceError(1, 8, 1, 16)},
			{Code: `import defaultExport, * as foo from 'foo';`, Errors: namespaceError(1, 23, 1, 31)},
		})
}
