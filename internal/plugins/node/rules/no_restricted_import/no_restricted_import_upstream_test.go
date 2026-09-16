package no_restricted_import

import (
	"path/filepath"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All cases from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-restricted-import.js
func TestNoRestrictedImportUpstream(t *testing.T) {
	root := restrictedImportRoot(t)
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoRestrictedImportRule,
		[]rule_tester.ValidTestCase{
			// upstream valid 1
			{Code: "import \"fs\"", Options: []any{[]any{"crypto"}}},
			// upstream valid 2
			{Code: "import \"path\"", Options: []any{[]any{"crypto", "stream", "os"}}},
			// upstream valid 3
			{Code: "import \"fs \""},
			// upstream valid 4
			{Code: "import \"foo/bar\";", Options: []any{[]any{"foo"}}},
			// upstream valid 5
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": []any{"foo", "bar"}}}}},
			// upstream valid 6
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": []any{"foo/c*"}}}}},
			// upstream valid 7
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": []any{"foo"}}, map[string]any{"name": []any{"foo/c*"}}}}},
			// upstream valid 8
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": []any{"foo"}}, map[string]any{"name": []any{"foo/*", "!foo/bar"}}}}},
			// upstream valid 9
			{Code: "import \"os \"", Options: []any{[]any{"fs", "crypto ", "stream", "os"}}},
			// upstream valid 10
			{Code: "import \"./foo\"", Options: []any{[]any{"foo"}}},
			// upstream valid 11
			{Code: "import \"foo\"", Options: []any{[]any{"./foo"}}},
			// upstream valid 12
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": "@foo/bar"}}}},
			// upstream valid 13
			{Code: "import \"../foo\";", FileName: "lib/sub/test.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "foo")}}}},
			// upstream valid 14
			{Code: "import(fs)", Options: []any{[]any{"fs"}}},
		},
		[]rule_tester.InvalidTestCase{
			// upstream invalid 1
			{Code: "import \"fs\"", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
			// upstream invalid 2
			{Code: "import fs from \"fs\"", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 3
			{Code: "import {} from \"fs\"", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 4
			{Code: "export * from \"fs\"", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19}}},
			// upstream invalid 5
			{Code: "export {} from \"fs\"", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 6
			{Code: "import \"foo/bar\";", Options: []any{[]any{"foo/bar"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 7
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": []any{"foo/bar"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 8
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": []any{"foo/*"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 9
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": []any{"foo/*"}}, map[string]any{"name": []any{"foo"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 10
			{Code: "import \"foo/bar\";", Options: []any{[]any{map[string]any{"name": []any{"foo/*", "!foo/baz"}}, map[string]any{"name": []any{"foo"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 11
			{Code: "import \"foo\";", Options: []any{[]any{map[string]any{"name": "foo", "message": "Please use 'bar' module instead."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo' module is restricted from being used. Please use 'bar' module instead.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 12
			{Code: "import \"bar\";", Options: []any{[]any{"foo", map[string]any{"name": "bar", "message": "Please use 'baz' module instead."}, "baz"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'bar' module is restricted from being used. Please use 'baz' module instead.", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 13
			{Code: "import \"@foo/bar\";", Options: []any{[]any{map[string]any{"name": "@foo/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'@foo/bar' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 18}}},
			// upstream invalid 14
			{Code: "import \"./foo/bar\";", Options: []any{[]any{map[string]any{"name": "./foo/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./foo/bar' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// upstream invalid 15
			{Code: "import \"../foo\";", FileName: "lib/test.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "foo")}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../foo' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
			// upstream invalid 16
			{Code: "import \"../../foo\";", FileName: "lib/sub/test.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "foo")}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../../foo' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19}}},
			// upstream invalid 17
			{Code: "import(\"fs\")", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
		},
	)
}

// All cases from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-restricted-import.md
func TestNoRestrictedImportDocumentation(t *testing.T) {
	root := restrictedImportRoot(t)
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoRestrictedImportRule,
		[]rule_tester.ValidTestCase{
			// documentation valid 1
			{Code: "import crypto from 'crypto';\nimport _ from 'lodash';", Options: []any{[]any{"fs", "cluster", "lodash/*"}}},
			// documentation valid 2
			{Code: "import pick from 'lodash/pick';", Options: []any{[]any{"fs", "cluster", map[string]any{"name": []any{"lodash/*", "!lodash/pick"}}}}},
			// documentation valid 3
			{Code: "import 'foo-module2'; import 'bar-module2';", Options: []any{[]any{"foo-module", "bar-module"}}},
			// documentation valid 4
			{Code: "import 'baz-module/good';", Options: []any{[]any{map[string]any{"name": []any{"foo-module/private/*", "bar-module/*", "!baz-module/good"}, "message": "Please use xyz-module instead."}}}},
		},
		[]rule_tester.InvalidTestCase{
			// documentation invalid 1
			{Code: "import fs from 'fs';\nimport cluster from 'cluster';\nimport pick from 'lodash/pick';", Options: []any{[]any{"fs", "cluster", "lodash/*"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}, {MessageId: "restricted", Message: "'cluster' module is restricted from being used.", Line: 2, Column: 21, EndLine: 2, EndColumn: 30}, {MessageId: "restricted", Message: "'lodash/pick' module is restricted from being used.", Line: 3, Column: 18, EndLine: 3, EndColumn: 31}}},
			// documentation invalid 2
			{Code: "import 'foo-module'; import 'bar-module';", Options: []any{[]any{"foo-module", "bar-module"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo-module' module is restricted from being used.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}, {MessageId: "restricted", Message: "'bar-module' module is restricted from being used.", Line: 1, Column: 29, EndLine: 1, EndColumn: 41}}},
			// documentation invalid 3
			{Code: "import 'foo-module'; import 'bar-module';", Options: []any{[]any{map[string]any{"name": "foo-module", "message": "Please use foo-module2 instead."}, map[string]any{"name": "bar-module", "message": "Please use bar-module2 instead."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo-module' module is restricted from being used. Please use foo-module2 instead.", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}, {MessageId: "restricted", Message: "'bar-module' module is restricted from being used. Please use bar-module2 instead.", Line: 1, Column: 29, EndLine: 1, EndColumn: 41}}},
			// documentation invalid 4
			{Code: "import 'lodash/pick';\nimport 'foo-module/private/a';\nimport 'bar-module/a';", Options: []any{[]any{map[string]any{"name": "lodash/*", "message": "Please use xyz-module instead."}, map[string]any{"name": []any{"foo-module/private/*", "bar-module/*", "!baz-module/good"}, "message": "Please use xyz-module instead."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'lodash/pick' module is restricted from being used. Please use xyz-module instead.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}, {MessageId: "restricted", Message: "'foo-module/private/a' module is restricted from being used. Please use xyz-module instead.", Line: 2, Column: 8, EndLine: 2, EndColumn: 30}, {MessageId: "restricted", Message: "'bar-module/a' module is restricted from being used. Please use xyz-module instead.", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			// documentation invalid 5
			{Code: "import '../server/api.js';", FileName: "client/input.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "server/**"), "message": "Don't use server code from client code."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../server/api.js' module is restricted from being used. Don't use server code from client code.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
			// documentation invalid 6
			{Code: "import '../client/view.js';", FileName: "server/input.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "client/**"), "message": "Don't use client code from server code."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../client/view.js' module is restricted from being used. Don't use client code from server code.", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},
		},
	)
}
