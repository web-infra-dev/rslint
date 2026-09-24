package enforce_node_protocol_usage_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	target "github.com/web-infra-dev/rslint/internal/plugins/import/rules/enforce_node_protocol_usage"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All cases from eslint-plugin-import v2.32.0, including both directions of the
// invalid templates (the upstream version guard accidentally skips always).
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/enforce-node-protocol-usage.js
func TestEnforceNodeProtocolUsageUpstream(t *testing.T) {
	settings := map[string]any{"import/node-version": "16.0.0"}
	valid := []rule_tester.ValidTestCase{
		{
			Code:    "import unicorn from \"unicorn\";",
			Options: []any{"always"},
		},
		{
			Code:    "import fs from \"./fs\";",
			Options: []any{"always"},
		},
		{
			Code:    "import fs from \"unknown-builtin-module\";",
			Options: []any{"always"},
		},
		{
			Code:    "import fs from \"node:fs\";",
			Options: []any{"always"},
		},
		{
			Code:    "\n        async function foo() {\n          const fs = await import(fs);\n        }\n      ",
			Options: []any{"always"},
		},
		{
			Code:    "\n        async function foo() {\n          const fs = await import(0);\n        }\n      ",
			Options: []any{"always"},
		},
		{
			Code:    "\n        async function foo() {\n          const fs = await import(`fs`);\n        }\n      ",
			Options: []any{"always"},
		},
		{
			Code:    "import \"punycode/\";",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = require(\"node:fs\");",
			Options: []any{"always"},
		},
		{
			Code:     "const fs = require(\"node:fs/promises\");",
			Options:  []any{"always"},
			Settings: settings,
		},
		{
			Code:    "const fs = require(fs);",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = notRequire(\"fs\");",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = foo.require(\"fs\");",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = require.resolve(\"fs\");",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = require(`fs`);",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = require?.(\"fs\");",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = require(\"fs\", extra);",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = require();",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = require(...[\"fs\"]);",
			Options: []any{"always"},
		},
		{
			Code:    "const fs = require(\"unicorn\");",
			Options: []any{"always"},
		},
		{
			Code:    "import fs from \"fs\";",
			Options: []any{"never"},
		},
		{
			Code:    "const fs = require(\"fs\");",
			Options: []any{"never"},
		},
		{
			Code:     "const fs = require(\"fs/promises\");",
			Options:  []any{"never"},
			Settings: settings,
		},
		{
			Code:    "import \"punycode/\";",
			Options: []any{"never"},
		},
		{
			Code:     "const fs = require(\"node:test\");",
			Options:  []any{"never"},
			Settings: settings,
		},
		{
			Code:    "\n            export class Thing {\n              constructor(public readonly name: string) {\n                  // Do nothing.\n              }\n\n              public sayHello(): void {\n                  console.log(`Hello, ${this.name}!`);\n              }\n            }\n          ",
			Options: []any{"always"},
		},
	}
	templates := []struct {
		code, output, name string
		line, column       int
		settings           map[string]any
	}{
		{
			code:   "import x from \"fs\";",
			output: "import x from \"node:fs\";",
			name:   "fs",
			line:   1, column: 15,
		},
		{
			code:   "export {promises} from \"fs\";",
			output: "export {promises} from \"node:fs\";",
			name:   "fs",
			line:   1, column: 24,
		},
		{
			code:   "\n        async function foo() {\n          const x = await import('fs');\n        }\n      ",
			output: "\n        async function foo() {\n          const x = await import('node:fs');\n        }\n      ",
			name:   "fs",
			line:   3, column: 34,
		},
		{
			code:   "import x from \"fs/promises\";",
			output: "import x from \"node:fs/promises\";",
			name:   "fs/promises",
			line:   1, column: 15,
		},
		{
			code:   "export {promises} from \"fs/promises\";",
			output: "export {promises} from \"node:fs/promises\";",
			name:   "fs/promises",
			line:   1, column: 24,
		},
		{
			code:   "\n        async function foo() {\n          const x = await import('fs/promises');\n        }\n      ",
			output: "\n        async function foo() {\n          const x = await import('node:fs/promises');\n        }\n      ",
			name:   "fs/promises",
			line:   3, column: 34,
		},
		{
			code:   "import x from \"buffer\";",
			output: "import x from \"node:buffer\";",
			name:   "buffer",
			line:   1, column: 15,
		},
		{
			code:   "export {promises} from \"buffer\";",
			output: "export {promises} from \"node:buffer\";",
			name:   "buffer",
			line:   1, column: 24,
		},
		{
			code:   "\n        async function foo() {\n          const x = await import('buffer');\n        }\n      ",
			output: "\n        async function foo() {\n          const x = await import('node:buffer');\n        }\n      ",
			name:   "buffer",
			line:   3, column: 34,
		},
		{
			code:   "import x from \"child_process\";",
			output: "import x from \"node:child_process\";",
			name:   "child_process",
			line:   1, column: 15,
		},
		{
			code:   "export {promises} from \"child_process\";",
			output: "export {promises} from \"node:child_process\";",
			name:   "child_process",
			line:   1, column: 24,
		},
		{
			code:   "\n        async function foo() {\n          const x = await import('child_process');\n        }\n      ",
			output: "\n        async function foo() {\n          const x = await import('node:child_process');\n        }\n      ",
			name:   "child_process",
			line:   3, column: 34,
		},
		{
			code:   "import x from \"timers/promises\";",
			output: "import x from \"node:timers/promises\";",
			name:   "timers/promises",
			line:   1, column: 15,
		},
		{
			code:   "export {promises} from \"timers/promises\";",
			output: "export {promises} from \"node:timers/promises\";",
			name:   "timers/promises",
			line:   1, column: 24,
		},
		{
			code:   "\n        async function foo() {\n          const x = await import('timers/promises');\n        }\n      ",
			output: "\n        async function foo() {\n          const x = await import('node:timers/promises');\n        }\n      ",
			name:   "timers/promises",
			line:   3, column: 34,
		},
		{
			code:   "import fs from \"fs/promises\";",
			output: "import fs from \"node:fs/promises\";",
			name:   "fs/promises",
			line:   1, column: 16,
			settings: settings,
		},
		{
			code:   "export {default} from \"fs/promises\";",
			output: "export {default} from \"node:fs/promises\";",
			name:   "fs/promises",
			line:   1, column: 23,
			settings: settings,
		},
		{
			code:   "\n      async function foo() {\n        const fs = await import('fs/promises');\n      }\n    ",
			output: "\n      async function foo() {\n        const fs = await import('node:fs/promises');\n      }\n    ",
			name:   "fs/promises",
			line:   3, column: 33,
			settings: settings,
		},
		{
			code:   "import {promises} from \"fs\";",
			output: "import {promises} from \"node:fs\";",
			name:   "fs",
			line:   1, column: 24,
		},
		{
			code:   "export {default as promises} from \"fs\";",
			output: "export {default as promises} from \"node:fs\";",
			name:   "fs",
			line:   1, column: 35,
		},
		{
			code:   "\n      async function foo() {\n        const fs = await import(\"fs/promises\");\n      }\n    ",
			output: "\n      async function foo() {\n        const fs = await import(\"node:fs/promises\");\n      }\n    ",
			name:   "fs/promises",
			line:   3, column: 33,
			settings: settings,
		},
		{
			code:   "import \"buffer\";",
			output: "import \"node:buffer\";",
			name:   "buffer",
			line:   1, column: 8,
		},
		{
			code:   "import \"child_process\";",
			output: "import \"node:child_process\";",
			name:   "child_process",
			line:   1, column: 8,
		},
		{
			code:   "import \"timers/promises\";",
			output: "import \"node:timers/promises\";",
			name:   "timers/promises",
			line:   1, column: 8,
			settings: settings,
		},
		{
			code:   "const {promises} = require(\"fs\")",
			output: "const {promises} = require(\"node:fs\")",
			name:   "fs",
			line:   1, column: 28,
		},
		{
			code:   "const fs = require(\"fs/promises\")",
			output: "const fs = require(\"node:fs/promises\")",
			name:   "fs/promises",
			line:   1, column: 20,
			settings: settings,
		},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, mode := range []string{"always", "never"} {
		for _, tc := range templates {
			code, output, endColumn, version := tc.code, tc.output, tc.column+len(tc.name)+2, tc.settings
			if mode == "never" {
				code, output, endColumn, version = output, code, endColumn+5, settings
			}
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: code, Output: []string{output}, Options: []any{mode}, Settings: version,
				Errors: []rule_tester.InvalidTestCaseError{protocolError(mode, tc.name, tc.line, tc.column, tc.line, endColumn)},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &target.EnforceNodeProtocolUsageRule, valid, invalid)
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/enforce-node-protocol-usage.md
// Each statement is a separate case: the documentation redeclares fs.
func TestEnforceNodeProtocolUsageDocumentation(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, mode := range []string{"always", "never"} {
		valid = append(valid, rule_tester.ValidTestCase{Code: "import * as test from 'node:test';", Options: []any{mode}})
		for _, tc := range []struct {
			code, output, name string
			column             int
		}{
			{"import fs from 'fs';", "import fs from 'node:fs';", "fs", 16},
			{"export { promises } from 'fs';", "export { promises } from 'node:fs';", "fs", 26},
			{"const fs = require('fs/promises');", "const fs = require('node:fs/promises');", "fs/promises", 20},
		} {
			code, output, endColumn := tc.code, tc.output, tc.column+len(tc.name)+2
			if mode == "never" {
				code, output, endColumn = output, code, endColumn+5
			}
			valid = append(valid, rule_tester.ValidTestCase{Code: output, Options: []any{mode}})
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: code, Output: []string{output}, Options: []any{mode},
				Errors: []rule_tester.InvalidTestCaseError{protocolError(mode, tc.name, 1, tc.column, 1, endColumn)},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &target.EnforceNodeProtocolUsageRule, valid, invalid)
}

func protocolError(mode, name string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	preferred, other := "node:"+name, name
	if mode == "never" {
		preferred, other = other, preferred
	}
	return rule_tester.InvalidTestCaseError{
		Message: "Prefer `" + preferred + "` over `" + other + "`.",
		Line:    line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}
