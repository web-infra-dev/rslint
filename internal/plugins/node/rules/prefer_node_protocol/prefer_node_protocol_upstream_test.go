package prefer_node_protocol_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_node_protocol"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func runProtocolTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	for i := range valid {
		if valid[i].FileName == "" {
			valid[i].FileName = "input.js"
		}
		valid[i].LanguageOptions.SourceType = "module"
	}
	for i := range invalid {
		if invalid[i].FileName == "" {
			invalid[i].FileName = "input.js"
		}
		invalid[i].LanguageOptions.SourceType = "module"
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allowJs.json", t, &prefer_node_protocol.PreferNodeProtocolRule, valid, invalid)
}

// All 45 valid and 29 invalid cases, in upstream order:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-node-protocol.js
func TestPreferNodeProtocolUpstream(t *testing.T) {
	runProtocolTests(t, []rule_tester.ValidTestCase{
		// Imports
		{Code: "import nodePlugin from \"eslint-plugin-n\";"},
		{Code: "import fs from \"./fs\";"},
		{Code: "import fs from \"unknown-builtin-module\";"},
		{Code: "import fs from \"node:fs\";"},
		{Code: "\n            async function foo() {\n\t\t\t    const fs = await import(fs);\n\t\t    }\n        "},
		{Code: "\n            async function foo() {\n                const fs = await import(0);\n            }\n        "},
		{Code: "\n            async function foo() {\n                const fs = await import(`fs`);\n            }\n        "},
		{Code: "import \"punycode\";"},
		{Code: "import \"punycode/\";"},
		{Code: "import \"bun\";"},
		{Code: "import \"bun:jsc\";"},
		{Code: "import \"bun:sqlite\";"},
		{Code: "export {promises} from \"node:fs\";"},
		// require
		{Code: "const fs = require(\"node:fs\");"},
		{Code: "const fs = require(\"node:fs/promises\");"},
		{Code: "const fs = require(fs);"},
		{Code: "const fs = notRequire(\"fs\");"},
		{Code: "const fs = foo.require(\"fs\");"},
		{Code: "const fs = require.resolve(\"fs\");"},
		{Code: "const fs = require(`fs`);"},
		{Code: "const fs = require?.(\"fs\");"},
		{Code: "const fs = require(\"fs\", extra);"},
		{Code: "const fs = require();"},
		{Code: "const fs = require(...[\"fs\"]);"},
		{Code: "const fs = require(\"eslint-plugin-n\");"},
		// Unsupported Node versions
		{Code: "import fs from \"fs\";", Options: map[string]any{"version": ">=10.13.0"}},
		{Code: "import fs from \"fs\";", Options: map[string]any{"version": "12.19.1"}},
		{Code: "import fs from \"fs\";", Options: map[string]any{"version": "13.14.0"}},
		{Code: "import fs from \"fs\";", Options: map[string]any{"version": "14.13.0"}},
		{Code: "const fs = require(\"fs\");", Options: map[string]any{"version": "14.17.6"}},
		{Code: "const fs = require(\"fs\");", Options: map[string]any{"version": "15.14.0"}},
		// process.getBuiltinModule
		{Code: "const fs = process.getBuiltinModule(\"node:fs\");"},
		{Code: "const fs = globalThis.process.getBuiltinModule(\"node:fs\");"},
		{Code: "const fs = process.getBuiltinModule(\"node:fs/promises\");"},
		{Code: "const fs = process.getNotBuiltinModule(\"node:fs\", extra);"},
		{Code: "const fs = process.getBuiltinModule(fs);"},
		{Code: "const fs = process.getNotBuiltinModule(\"fs\");"},
		{Code: "const fs = process.foo.getNotBuiltinModule(\"fs\");"},
		{Code: "const fs = foo.process.getNotBuiltinModule(\"fs\");"},
		{Code: "const fs = process.getNotBuiltinModule.foo(\"fs\");"},
		{Code: "const fs = process.getNotBuiltinModule(`fs`);"},
		{Code: "const fs = process.getNotBuiltinModule();"},
		{Code: "const fs = process.getNotBuiltinModule(...[\"fs\"]);"},
		{Code: "const fs = process.getNotBuiltinModule(\"eslint-plugin-n\");"},
		{Code: "const fs = process.getBuiltinModule(\"node:fs\");", Options: map[string]any{"version": "12.19.1"}},
	}, []rule_tester.InvalidTestCase{
		// Imports
		{Code: "import fs from \"fs\";", Output: []string{"import fs from \"node:fs\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}}},
		{Code: "export {promises} from \"fs\";", Output: []string{"export {promises} from \"node:fs\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28}}},
		{Code: "\n                async function foo() {\n                    const fs = await import('fs');\n                }\n            ", Output: []string{"\n                async function foo() {\n                    const fs = await import('node:fs');\n                }\n            "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 3, Column: 45, EndLine: 3, EndColumn: 49}}},
		{Code: "import fs from \"fs/promises\";", Output: []string{"import fs from \"node:fs/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 29}}},
		{Code: "export {default} from \"fs/promises\";", Output: []string{"export {default} from \"node:fs/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 36}}},
		{Code: "\n                async function foo() {\n                    const fs = await import('fs/promises');\n                }\n            ", Output: []string{"\n                async function foo() {\n                    const fs = await import('node:fs/promises');\n                }\n            "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 3, Column: 45, EndLine: 3, EndColumn: 58}}},
		{Code: "import {promises} from \"fs\";", Output: []string{"import {promises} from \"node:fs\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28}}},
		{Code: "export {default as promises} from \"fs\";", Output: []string{"export {default as promises} from \"node:fs\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39}}},
		{Code: "import {promises} from 'fs';", Output: []string{"import {promises} from 'node:fs';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 28}}},
		{Code: "\n                async function foo() {\n                    const fs = await import(\"fs/promises\");\n                }\n            ", Output: []string{"\n                async function foo() {\n                    const fs = await import(\"node:fs/promises\");\n                }\n            "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 3, Column: 45, EndLine: 3, EndColumn: 58}}},
		{Code: "\n                async function foo() {\n                    const fs = await import(/* escaped */\"\\u{66}s/promises\");\n                }\n            ", Output: []string{"\n                async function foo() {\n                    const fs = await import(/* escaped */\"node:\\u{66}s/promises\");\n                }\n            "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 3, Column: 58, EndLine: 3, EndColumn: 76}}},
		{Code: "import \"buffer\";", Output: []string{"import \"node:buffer\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:buffer` over `buffer`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 16}}},
		{Code: "import \"child_process\";", Output: []string{"import \"node:child_process\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:child_process` over `child_process`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		{Code: "import \"timers/promises\";", Output: []string{"import \"node:timers/promises\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:timers/promises` over `timers/promises`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
		// require
		{Code: "const {promises} = require(\"fs\")", Output: []string{"const {promises} = require(\"node:fs\")"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 32}}},
		{Code: "const fs = require('fs/promises')", Output: []string{"const fs = require('node:fs/promises')"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 33}}},
		{Code: "\n                const express = require('express');\n                const fs = require('fs/promises');\n            ", Output: []string{"\n                const express = require('express');\n                const fs = require('node:fs/promises');\n            "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 3, Column: 36, EndLine: 3, EndColumn: 49}}},
		// Supported Node versions
		{Code: "import fs from \"fs\";", Options: map[string]any{"version": "12.20.0"}, Output: []string{"import fs from \"node:fs\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}}},
		{Code: "import fs from \"fs\";", Options: map[string]any{"version": "14.13.1"}, Output: []string{"import fs from \"node:fs\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}}},
		{Code: "const fs = require(\"fs\");", Options: map[string]any{"version": "14.18.0"}, Output: []string{"const fs = require(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24}}},
		{Code: "const fs = require(\"fs\");", Options: map[string]any{"version": "16.0.0"}, Output: []string{"const fs = require(\"node:fs\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24}}},
		{Code: "\n                const fs = require(\"fs\");\n                import buffer from 'buffer'\n            ", Options: map[string]any{"version": "12.20.0"}, Output: []string{"\n                const fs = require(\"fs\");\n                import buffer from 'node:buffer'\n            "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:buffer` over `buffer`.", Line: 3, Column: 36, EndLine: 3, EndColumn: 44}}},
		// Regression: upstream issue #431
		{Code: "import https from \"https\";", Output: []string{"import https from \"node:https\";"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:https` over `https`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}}},
		// process.getBuiltinModule
		{Code: "const {promises} = process.getBuiltinModule(\"fs\")", Output: []string{"const {promises} = process.getBuiltinModule(\"node:fs\")"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 49}}},
		{Code: "const {promises} = globalThis.process.getBuiltinModule(\"fs\")", Output: []string{"const {promises} = globalThis.process.getBuiltinModule(\"node:fs\")"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 56, EndLine: 1, EndColumn: 60}}},
		{Code: "const {promises} = process.getBuiltinModule(\"fs\", extra)", Output: []string{"const {promises} = process.getBuiltinModule(\"node:fs\", extra)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 49}}},
		{Code: "const fs = process.getBuiltinModule('fs/promises')", Output: []string{"const fs = process.getBuiltinModule('node:fs/promises')"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 50}}},
		{Code: "\n                const express = process.getBuiltinModule('express');\n                const fs = process.getBuiltinModule('fs/promises');\n            ", Output: []string{"\n                const express = process.getBuiltinModule('express');\n                const fs = process.getBuiltinModule('node:fs/promises');\n            "}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs/promises` over `fs/promises`.", Line: 3, Column: 53, EndLine: 3, EndColumn: 66}}},
		{Code: "const {promises} = process.getBuiltinModule(\"fs\")", Options: map[string]any{"version": "12.19.1"}, Output: []string{"const {promises} = process.getBuiltinModule(\"node:fs\")"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 49}}},
	})
}

// Every example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-node-protocol.md
func TestPreferNodeProtocolDocumentation(t *testing.T) {
	runProtocolTests(t, []rule_tester.ValidTestCase{
		{Code: "import fs from \"node:fs\""},
		{Code: "export { promises } from \"node:fs\""},
		{Code: "const fs = require(\"node:fs\")"},
		{Code: "const fs = process.getBuiltinModule(\"node:fs\")"},
	}, []rule_tester.InvalidTestCase{
		{Code: "import fs from \"fs\"", Output: []string{"import fs from \"node:fs\""}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20}}},
		{Code: "export { promises } from \"fs\"", Output: []string{"export { promises } from \"node:fs\""}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30}}},
		{Code: "const fs = require(\"fs\")", Output: []string{"const fs = require(\"node:fs\")"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24}}},
		{Code: "const fs = process.getBuiltinModule(\"fs\")", Output: []string{"const fs = process.getBuiltinModule(\"node:fs\")"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferNodeProtocol", Message: "Prefer `node:fs` over `fs`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 41}}},
	})
}
