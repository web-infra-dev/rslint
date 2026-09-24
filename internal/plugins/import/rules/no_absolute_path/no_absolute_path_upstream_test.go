package no_absolute_path

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const absolutePathMessage = "Do not import modules using an absolute path"

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-absolute-path.js
// and the examples in docs/rules/no-absolute-path.md at the same tag.
// Upstream uses a literal message without a message ID or suggestions.
func TestNoAbsolutePathUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &NoAbsolutePathRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "import _ from \"lodash\"",
			},
			{
				Code: "import find from \"lodash.find\"",
			},
			{
				Code: "import foo from \"./foo\"",
			},
			{
				Code: "import foo from \"../foo\"",
			},
			{
				Code: "import foo from \"foo\"",
			},
			{
				Code: "import foo from \"./\"",
			},
			{
				Code: "import foo from \"@scope/foo\"",
			},
			{
				Code: "var _ = require(\"lodash\")",
			},
			{
				Code: "var find = require(\"lodash.find\")",
			},
			{
				Code: "var foo = require(\"./foo\")",
			},
			{
				Code: "var foo = require(\"../foo\")",
			},
			{
				Code: "var foo = require(\"foo\")",
			},
			{
				Code: "var foo = require(\"./\")",
			},
			{
				Code: "var foo = require(\"@scope/foo\")",
			},
			{
				Code: "import events from \"events\"",
			},
			{
				Code: "import path from \"path\"",
			},
			{
				Code: "var events = require(\"events\")",
			},
			{
				Code: "var path = require(\"path\")",
			},
			{
				Code: "import path from \"path\";import events from \"events\"",
			},
			{
				Code:    "import path from \"path\"",
				Options: []any{map[string]any{"amd": true}},
			},
			{
				Code: "require([\"/some/path\"], function (f) { /* ... */ })",
			},
			{
				Code: "define([\"/some/path\"], function (f) { /* ... */ })",
			},
			{
				Code:    "require([\"./some/path\"], function (f) { /* ... */ })",
				Options: []any{map[string]any{"amd": true}},
			},
			{
				Code:    "define([\"./some/path\"], function (f) { /* ... */ })",
				Options: []any{map[string]any{"amd": true}},
			},
			// Documentation pass examples.
			{
				Code: "import _ from 'lodash';",
			},
			{
				Code: "import foo from 'foo';",
			},
			{
				Code: "import foo from './foo';",
			},
			{
				Code: "var _ = require('lodash');",
			},
			{
				Code: "var foo = require('foo');",
			},
			{
				Code: "var foo = require('./foo');",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "import f from \"/foo\"",
				FileName: "/foo/bar/index.js",
				Output:   []string{"import f from \"..\""},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:     "import f from \"/foo/bar/baz.js\"",
				FileName: "/foo/bar/index.js",
				Output:   []string{"import f from \"./baz.js\""},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:     "import f from \"/foo/path\"",
				FileName: "/foo/bar/index.js",
				Output:   []string{"import f from \"../path\""},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:     "import f from \"/some/path\"",
				FileName: "/foo/bar/index.js",
				Output:   []string{"import f from \"../../some/path\""},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:     "import f from \"/some/path\"",
				FileName: "/foo/bar/index.js",
				Options:  []any{map[string]any{"amd": true}},
				Output:   []string{"import f from \"../../some/path\""},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:     "var f = require(\"/foo\")",
				FileName: "/foo/bar/index.js",
				Output:   []string{"var f = require(\"..\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:     "var f = require(\"/foo/path\")",
				FileName: "/foo/bar/index.js",
				Output:   []string{"var f = require(\"../path\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:     "var f = require(\"/some/path\")",
				FileName: "/foo/bar/index.js",
				Output:   []string{"var f = require(\"../../some/path\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:     "var f = require(\"/some/path\")",
				FileName: "/foo/bar/index.js",
				Options:  []any{map[string]any{"amd": true}},
				Output:   []string{"var f = require(\"../../some/path\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:     "require([\"/some/path\"], function (f) { /* ... */ })",
				FileName: "/foo/bar/index.js",
				Options:  []any{map[string]any{"amd": true}},
				Output:   []string{"require([\"../../some/path\"], function (f) { /* ... */ })"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:     "define([\"/some/path\"], function (f) { /* ... */ })",
				FileName: "/foo/bar/index.js",
				Options:  []any{map[string]any{"amd": true}},
				Output:   []string{"define([\"../../some/path\"], function (f) { /* ... */ })"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 21},
				},
			},
			// Documentation fail examples and AMD configuration.
			{
				Code:     "import f from '/foo';",
				FileName: "/foo/bar/index.js",
				Output:   []string{"import f from \"..\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:     "import f from '/some/path';",
				FileName: "/foo/bar/index.js",
				Output:   []string{"import f from \"../../some/path\";"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:     "var f = require('/foo');",
				FileName: "/foo/bar/index.js",
				Output:   []string{"var f = require(\"..\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:     "var f = require('/some/path');",
				FileName: "/foo/bar/index.js",
				Output:   []string{"var f = require(\"../../some/path\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:     "define(['/foo'], function (foo) { /*...*/ })\nrequire(['/foo'], function (foo) { /*...*/ })\nconst foo = require('/foo')",
				FileName: "/foo/bar/index.js",
				Options:  []any{map[string]any{"commonjs": false, "amd": true}},
				Output:   []string{"define([\"..\"], function (foo) { /*...*/ })\nrequire([\"..\"], function (foo) { /*...*/ })\nconst foo = require('/foo')"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: absolutePathMessage, Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
					{MessageId: "", Message: absolutePathMessage, Line: 2, Column: 10, EndLine: 2, EndColumn: 16},
				},
			},
		})
}
