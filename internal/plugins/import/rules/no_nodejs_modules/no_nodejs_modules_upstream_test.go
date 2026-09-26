package no_nodejs_modules_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_nodejs_modules"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Every case from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-nodejs-modules.js
// Includes every conditional node: case, supported by rslint's builtin set.
func TestNoNodejsModulesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_nodejs_modules.NoNodejsModulesRule,
		[]rule_tester.ValidTestCase{
			{Code: "import _ from \"lodash\""},
			{Code: "import find from \"lodash.find\""},
			{Code: "import foo from \"./foo\""},
			{Code: "import foo from \"../foo\""},
			{Code: "import foo from \"foo\""},
			{Code: "import foo from \"./\""},
			{Code: "import foo from \"@scope/foo\""},
			{Code: "var _ = require(\"lodash\")"},
			{Code: "var find = require(\"lodash.find\")"},
			{Code: "var foo = require(\"./foo\")"},
			{Code: "var foo = require(\"../foo\")"},
			{Code: "var foo = require(\"foo\")"},
			{Code: "var foo = require(\"./\")"},
			{Code: "var foo = require(\"@scope/foo\")"},
			{Code: "import events from \"events\"",
				Options: []any{map[string]any{"allow": []any{"events"}}},
			},
			{Code: "import path from \"path\"",
				Options: []any{map[string]any{"allow": []any{"path"}}},
			},
			{Code: "var events = require(\"events\")",
				Options: []any{map[string]any{"allow": []any{"events"}}},
			},
			{Code: "var path = require(\"path\")",
				Options: []any{map[string]any{"allow": []any{"path"}}},
			},
			{Code: "import path from \"path\";import events from \"events\"",
				Options: []any{map[string]any{"allow": []any{"path", "events"}}},
			},
			{Code: "import events from \"node:events\"",
				Options: []any{map[string]any{"allow": []any{"node:events"}}},
			},
			{Code: "var events = require(\"node:events\")",
				Options: []any{map[string]any{"allow": []any{"node:events"}}},
			},
			{Code: "import path from \"node:path\"",
				Options: []any{map[string]any{"allow": []any{"node:path"}}},
			},
			{Code: "var path = require(\"node:path\")",
				Options: []any{map[string]any{"allow": []any{"node:path"}}},
			},
			{Code: "import path from \"node:path\";import events from \"node:events\"",
				Options: []any{map[string]any{"allow": []any{"node:path", "node:events"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "import path from \"path\"",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"path\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{Code: "import fs from \"fs\"",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{Code: "var path = require(\"path\")",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"path\"", Line: 1, Column: 12, EndLine: 1, EndColumn: 27},
				},
			},
			{Code: "var fs = require(\"fs\")",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 10, EndLine: 1, EndColumn: 23},
				},
			},
			{Code: "import fs from \"fs\"",
				Options: []any{map[string]any{"allow": []any{"path"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{Code: "import path from \"node:path\"",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:path\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				},
			},
			{Code: "var path = require(\"node:path\")",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:path\"", Line: 1, Column: 12, EndLine: 1, EndColumn: 32},
				},
			},
			{Code: "import fs from \"node:fs\"",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{Code: "var fs = require(\"node:fs\")",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 1, Column: 10, EndLine: 1, EndColumn: 28},
				},
			},
			{Code: "import fs from \"node:fs\"",
				Options: []any{map[string]any{"allow": []any{"node:path"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"node:fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
		},
	)
}

// Examples from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-nodejs-modules.md
func TestNoNodejsModulesDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_nodejs_modules.NoNodejsModulesRule,
		[]rule_tester.ValidTestCase{
			{Code: "import _ from 'lodash';"},
			{Code: "import foo from 'foo';"},
			{Code: "import foo from './foo';"},
			{Code: "var _ = require('lodash');"},
			{Code: "var foo = require('foo');"},
			{Code: "var foo = require('./foo');"},
			{Code: "import path from 'path';",
				Options: []any{map[string]any{"allow": []any{"path"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "import fs from 'fs';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{Code: "import path from 'path';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"path\"", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{Code: "var fs = require('fs');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"fs\"", Line: 1, Column: 10, EndLine: 1, EndColumn: 23},
				},
			},
			{Code: "var path = require('path');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Do not import Node.js builtin module \"path\"", Line: 1, Column: 12, EndLine: 1, EndColumn: 27},
				},
			},
		},
	)
}
