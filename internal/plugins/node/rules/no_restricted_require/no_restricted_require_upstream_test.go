package no_restricted_require

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"path/filepath"
	"testing"
)

// All cases from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-restricted-require.js
func TestNoRestrictedRequireUpstream(t *testing.T) {
	root := restrictedRequireRoot(t)
	runRestrictedRequireTests(t, root,
		[]rule_tester.ValidTestCase{
			// upstream valid 1
			{Code: "require(\"fs\")", Options: []any{[]any{"crypto"}}},
			// upstream valid 2
			{Code: "require(\"path\")", Options: []any{[]any{"crypto", "stream", "os"}}},
			// upstream valid 3
			{Code: "require(\"fs \")"},
			// upstream valid 4
			{Code: "require(2)", Options: []any{[]any{"crypto"}}},
			// upstream valid 5
			{Code: "require(foo)", Options: []any{[]any{"crypto"}}},
			// upstream valid 6
			{Code: "bar('crypto');", Options: []any{[]any{"crypto"}}},
			// upstream valid 7
			{Code: "require(\"foo/bar\");", Options: []any{[]any{"foo"}}},
			// upstream valid 8
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": []any{"foo", "bar"}}}}},
			// upstream valid 9
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": []any{"foo/c*"}}}}},
			// upstream valid 10
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": []any{"foo"}}, map[string]any{"name": []any{"foo/c*"}}}}},
			// upstream valid 11
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": []any{"foo"}}, map[string]any{"name": []any{"foo/*", "!foo/bar"}}}}},
			// upstream valid 12
			{Code: "require(\"os \")", Options: []any{[]any{"fs", "crypto ", "stream", "os"}}},
			// upstream valid 13
			{Code: "require(\"./foo\")", Options: []any{[]any{"foo"}}},
			// upstream valid 14
			{Code: "require(\"foo\")", Options: []any{[]any{"./foo"}}},
			// upstream valid 15
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": "@foo/bar"}}}},
			// upstream valid 16
			{Code: "require(\"../foo\");", FileName: "lib/sub/test.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "foo")}}}},
			// upstream valid 17
			{Code: "require(\"foo/bar/baz\")", Options: []any{[]any{"foo/*"}}},
		},
		[]rule_tester.InvalidTestCase{
			// upstream invalid 1
			{Code: "require(\"fs\")", Options: []any{[]any{"fs"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13}}},
			// upstream invalid 2
			{Code: "require(\"foo/bar\");", Options: []any{[]any{"foo/bar"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
			// upstream invalid 3
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": []any{"foo/bar"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
			// upstream invalid 4
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": []any{"foo/*"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
			// upstream invalid 5
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": []any{"foo/*"}}, map[string]any{"name": []any{"foo"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
			// upstream invalid 6
			{Code: "require(\"foo/bar\");", Options: []any{[]any{map[string]any{"name": []any{"foo/*", "!foo/baz"}}, map[string]any{"name": []any{"foo"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo/bar' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}}},
			// upstream invalid 7
			{Code: "require(\"foo\");", Options: []any{[]any{map[string]any{"name": "foo", "message": "Please use 'bar' module instead."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo' module is restricted from being used. Please use 'bar' module instead.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
			// upstream invalid 8
			{Code: "require(\"bar\");", Options: []any{[]any{"foo", map[string]any{"name": "bar", "message": "Please use 'baz' module instead."}, "baz"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'bar' module is restricted from being used. Please use 'baz' module instead.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},
			// upstream invalid 9
			{Code: "require(\"@foo/bar\");", Options: []any{[]any{map[string]any{"name": "@foo/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'@foo/bar' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 19}}},
			// upstream invalid 10
			{Code: "require(\"./foo/bar\");", Options: []any{[]any{map[string]any{"name": "./foo/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'./foo/bar' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 11
			{Code: "require(\"../foo\");", FileName: "lib/test.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "foo")}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../foo' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17}}},
			// upstream invalid 12
			{Code: "require(\"../../foo\");", FileName: "lib/sub/test.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "foo")}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../../foo' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
			// upstream invalid 13
			{Code: "require(\"../../foo\");", FileName: "lib/sub/test.js", Options: []any{[]any{map[string]any{"name": "**/foo"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../../foo' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 20}}},
		},
	)
}

// All cases from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-restricted-require.md
func TestNoRestrictedRequireDocumentation(t *testing.T) {
	root := restrictedRequireRoot(t)
	runRestrictedRequireTests(t, root,
		[]rule_tester.ValidTestCase{
			// documentation valid 1
			{Code: "const crypto = require('crypto');\nconst _ = require('lodash');", Options: []any{[]any{"fs", "cluster", "lodash/*"}}},
			// documentation valid 2
			{Code: "const pick = require('lodash/pick');", Options: []any{[]any{"fs", "cluster", map[string]any{"name": []any{"lodash/*", "!lodash/pick"}}}}},
			// documentation valid 3
			{Code: "require('foo-module2'); require('bar-module2');", Options: []any{[]any{"foo-module", "bar-module"}}},
			// documentation valid 4
			{Code: "require('baz-module/good');", Options: []any{[]any{map[string]any{"name": []any{"foo-module/private/*", "bar-module/*", "!baz-module/good"}, "message": "Please use xyz-module instead."}}}},
		},
		[]rule_tester.InvalidTestCase{
			// documentation invalid 1
			{Code: "const fs = require('fs');\nconst cluster = require('cluster');\nconst pick = require('lodash/pick');", Options: []any{[]any{"fs", "cluster", "lodash/*"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'fs' module is restricted from being used.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24}, {MessageId: "restricted", Message: "'cluster' module is restricted from being used.", Line: 2, Column: 25, EndLine: 2, EndColumn: 34}, {MessageId: "restricted", Message: "'lodash/pick' module is restricted from being used.", Line: 3, Column: 22, EndLine: 3, EndColumn: 35}}},
			// documentation invalid 2
			{Code: "require('foo-module'); require('bar-module');", Options: []any{[]any{"foo-module", "bar-module"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo-module' module is restricted from being used.", Line: 1, Column: 9, EndLine: 1, EndColumn: 21}, {MessageId: "restricted", Message: "'bar-module' module is restricted from being used.", Line: 1, Column: 32, EndLine: 1, EndColumn: 44}}},
			// documentation invalid 3
			{Code: "require('foo-module'); require('bar-module');", Options: []any{[]any{map[string]any{"name": "foo-module", "message": "Please use foo-module2 instead."}, map[string]any{"name": "bar-module", "message": "Please use bar-module2 instead."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'foo-module' module is restricted from being used. Please use foo-module2 instead.", Line: 1, Column: 9, EndLine: 1, EndColumn: 21}, {MessageId: "restricted", Message: "'bar-module' module is restricted from being used. Please use bar-module2 instead.", Line: 1, Column: 32, EndLine: 1, EndColumn: 44}}},
			// documentation invalid 4
			{Code: "require('lodash/pick');\nrequire('foo-module/private/a');\nrequire('bar-module/a');", Options: []any{[]any{map[string]any{"name": "lodash/*", "message": "Please use xyz-module instead."}, map[string]any{"name": []any{"foo-module/private/*", "bar-module/*", "!baz-module/good"}, "message": "Please use xyz-module instead."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'lodash/pick' module is restricted from being used. Please use xyz-module instead.", Line: 1, Column: 9, EndLine: 1, EndColumn: 22}, {MessageId: "restricted", Message: "'foo-module/private/a' module is restricted from being used. Please use xyz-module instead.", Line: 2, Column: 9, EndLine: 2, EndColumn: 31}, {MessageId: "restricted", Message: "'bar-module/a' module is restricted from being used. Please use xyz-module instead.", Line: 3, Column: 9, EndLine: 3, EndColumn: 23}}},
			// documentation invalid 5
			{Code: "require('../server/api.js');", FileName: "client/input.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "server/**"), "message": "Don't use server code from client code."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../server/api.js' module is restricted from being used. Don't use server code from client code.", Line: 1, Column: 9, EndLine: 1, EndColumn: 27}}},
			// documentation invalid 6
			{Code: "require('../client/view.js');", FileName: "server/input.js", Options: []any{[]any{map[string]any{"name": filepath.Join(root.Dir, "client/**"), "message": "Don't use client code from server code."}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Message: "'../client/view.js' module is restricted from being used. Don't use client code from server code.", Line: 1, Column: 9, EndLine: 1, EndColumn: 28}}},
		},
	)
}
