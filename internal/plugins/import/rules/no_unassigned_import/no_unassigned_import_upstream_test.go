package no_unassigned_import_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_unassigned_import"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const unassignedMessage = "Imported module should be assigned"

// Examples from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-unassigned-import.md.
func TestNoUnassignedImportDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unassigned_import.NoUnassignedImportRule,
		[]rule_tester.ValidTestCase{
			{Code: "import _ from 'foo'"},
			{Code: "import _, {foo} from 'foo'"},
			{Code: "import _, {foo as bar} from 'foo'"},
			{Code: "import {foo as bar} from 'foo'"},
			{Code: "import * as _ from 'foo'"},
			// The documentation repeats the first require example twice.
			{Code: "const _ = require('foo')"},
			{Code: "const {foo} = require('foo')"},
			{Code: "const {foo: bar} = require('foo')"},
			{Code: "const [a, b] = require('foo')"},
			{Code: "bar(require('foo'))"},
			{Code: "require('foo').bar"},
			{Code: "require('foo').bar()"},
			{Code: "require('foo')()"},
			{
				Code:    "import './style.css'",
				Options: map[string]any{"allow": []any{"**/*.css"}},
			},
			{
				Code:    "import 'babel-register'",
				Options: map[string]any{"allow": []any{"babel-register"}},
			},
			{
				Code:     "import './styles/app.css'\nimport '../scripts/register.js'",
				FileName: "src/app.ts",
				Options:  map[string]any{"allow": []any{"src/styles/**", "**/scripts/*.js"}},
			},
			{
				// Listed under Fail in the upstream docs, but the actual rule
				// allows this path: ../styles resolves to PROJECT_ROOT/styles.
				Code:     "import '../styles/app.css'",
				FileName: "src/app.ts",
				Options:  map[string]any{"allow": []any{"styles/*.css"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "import 'should'\nrequire('should')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
					{MessageId: "", Message: unassignedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 18},
				},
			},
		},
	)
}

// All cases from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-unassigned-import.js.
// The fixture root is the test working directory; src/app.ts preserves the upstream
// src/app.js path relationships while using the fixture's TypeScript project.
// Empty message IDs, no fixes, and no suggestions are asserted by the Go tester.
func TestNoUnassignedImportUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unassigned_import.NoUnassignedImportRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "import _ from \"lodash\"",
			},
			{
				Code: "import _, {foo} from \"lodash\"",
			},
			{
				Code: "import _, {foo as bar} from \"lodash\"",
			},
			{
				Code: "import {foo as bar} from \"lodash\"",
			},
			{
				Code: "import * as _ from \"lodash\"",
			},
			{
				Code: "import _ from \"./\"",
			},
			{
				Code: "const _ = require(\"lodash\")",
			},
			{
				Code: "const {foo} = require(\"lodash\")",
			},
			{
				Code: "const {foo: bar} = require(\"lodash\")",
			},
			{
				Code: "const [a, b] = require(\"lodash\")",
			},
			{
				Code: "const _ = require(\"./\")",
			},
			{
				Code: "foo(require(\"lodash\"))",
			},
			{
				Code: "require(\"lodash\").foo",
			},
			{
				Code: "require(\"lodash\").foo()",
			},
			{
				Code: "require(\"lodash\")()",
			},
			{
				Code:    "import \"app.css\"",
				Options: []any{map[string]any{"allow": []any{"**/*.css"}}},
			},
			{
				Code:    "import \"app.css\";",
				Options: []any{map[string]any{"allow": []any{"*.css"}}},
			},
			{
				Code:    "import \"./app.css\"",
				Options: []any{map[string]any{"allow": []any{"**/*.css"}}},
			},
			{
				Code:    "import \"foo/bar\"",
				Options: []any{map[string]any{"allow": []any{"foo/**"}}},
			},
			{
				Code:    "import \"foo/bar\"",
				Options: []any{map[string]any{"allow": []any{"foo/bar"}}},
			},
			{
				Code:    "import \"../dir/app.css\"",
				Options: []any{map[string]any{"allow": []any{"**/*.css"}}},
			},
			{
				Code:    "import \"../dir/app.js\"",
				Options: []any{map[string]any{"allow": []any{"**/dir/**"}}},
			},
			{
				Code:    "require(\"./app.css\")",
				Options: []any{map[string]any{"allow": []any{"**/*.css"}}},
			},
			{
				Code:    "import \"babel-register\"",
				Options: []any{map[string]any{"allow": []any{"babel-register"}}},
			},
			{
				Code:     "import \"./styles/app.css\"",
				FileName: "src/app.ts",
				Options:  []any{map[string]any{"allow": []any{"src/styles/**"}}},
			},
			{
				Code:     "import \"../scripts/register.js\"",
				FileName: "src/app.ts",
				Options:  []any{map[string]any{"allow": []any{"src/styles/**", "**/scripts/*.js"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "import \"lodash\"",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: "require(\"lodash\")",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "import \"./app.css\"",
				Options: []any{map[string]any{"allow": []any{"**/*.js"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "import \"./app.css\"",
				Options: []any{map[string]any{"allow": []any{"**/dir/**"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "require(\"./app.css\")",
				Options: []any{map[string]any{"allow": []any{"**/*.js"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:     "import \"./styles/app.css\"",
				FileName: "src/app.ts",
				Options:  []any{map[string]any{"allow": []any{"styles/*.css"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: unassignedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
		},
	)
}
