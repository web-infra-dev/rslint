package prefer_node_protocol_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_node_protocol"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferNodeProtocolUpstream migrates the full valid/invalid suite from
// upstream test/prefer-node-protocol.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in the
// prefer_node_protocol_extras_test.go file.
func TestPreferNodeProtocolUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_node_protocol.PreferNodeProtocolRule,
		[]rule_tester.ValidTestCase{
			// ---- import / export / dynamic import ----
			{Code: `import unicorn from "unicorn";`},
			{Code: `import fs from "./fs";`},
			{Code: `import fs from "unknown-builtin-module";`},
			{Code: `import fs from "node:fs";`},
			{Code: "async function foo() {\n\tconst fs = await import(fs);\n}"},
			{Code: "async function foo() {\n\tconst fs = await import(0);\n}"},
			{Code: "async function foo() {\n\tconst fs = await import(`fs`);\n}"},
			{Code: `import "punycode";`},      // Deprecated
			{Code: `import "node:punycode";`}, // Deprecated
			{Code: `import "punycode/";`},
			{Code: `import "fs/";`},
			// `test` is not a builtin module, `node:test` is
			{Code: `import "test";`},
			{Code: `import "node:test";`},
			// https://bun.sh/docs/runtime/bun-apis
			{Code: `import "bun";`},
			{Code: `import "bun:jsc";`},
			{Code: `import "bun:sqlite";`},

			// ---- require ----
			{Code: `const fs = require("node:fs");`},
			{Code: `const fs = require("node:fs/promises");`},
			{Code: `const fs = require(fs);`},
			{Code: `const fs = notRequire("fs");`},
			{Code: `const fs = foo.require("fs");`},
			{Code: `const fs = require.resolve("fs");`},
			{Code: "const fs = require(`fs`);"},
			{Code: `const fs = require?.("fs");`},
			{Code: `const fs = require("fs", extra);`},
			{Code: `const fs = require();`},
			{Code: `const fs = require(...["fs"]);`},
			{Code: `const fs = require("unicorn");`},

			// ---- process.getBuiltinModule ----
			{Code: `const fs = process.getBuiltinModule("node:fs")`},
			{Code: `const fs = process.getBuiltinModule?.("fs")`},
			{Code: `const fs = process?.getBuiltinModule("fs")`},
			{Code: `const fs = process.notGetBuiltinModule("fs")`},
			{Code: `const fs = notProcess.getBuiltinModule("fs")`},
			{Code: `const fs = process.getBuiltinModule("fs", extra)`},
			{Code: `const fs = process.getBuiltinModule(...["fs"])`},
			{Code: `const fs = process.getBuiltinModule()`},
			{Code: `const fs = process.getBuiltinModule("unicorn")`},
			// Not checking this to avoid false positive
			{Code: "import {getBuiltinModule} from 'node:process';\nconst fs = getBuiltinModule(\"fs\");"},

			// ---- TypeScript ----
			{Code: `const fs = require("node:fs") as "fs";`},
			{Code: `type fs = typeof import("node:fs");`},
			{Code: `type fs = typeof SomeType<"fs">;`},
			{Code: `type fs = typeof fs;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- import / export / dynamic import ----
			{
				Code:   `import fs from "fs";`,
				Output: []string{`import fs from "node:fs";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 16}},
			},
			{
				Code:   `export {promises} from "fs";`,
				Output: []string{`export {promises} from "node:fs";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 24}},
			},
			{
				Code:   "async function foo() {\n\tconst fs = await import('fs');\n}",
				Output: []string{"async function foo() {\n\tconst fs = await import('node:fs');\n}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 2, Column: 26}},
			},
			{
				Code:   `import fs from "fs/promises";`,
				Output: []string{`import fs from "node:fs/promises";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 16}},
			},
			{
				Code:   `export {default} from "fs/promises";`,
				Output: []string{`export {default} from "node:fs/promises";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 23}},
			},
			{
				Code:   "async function foo() {\n\tconst fs = await import('fs/promises');\n}",
				Output: []string{"async function foo() {\n\tconst fs = await import('node:fs/promises');\n}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 2, Column: 26}},
			},
			{
				Code:   `import {promises} from "fs";`,
				Output: []string{`import {promises} from "node:fs";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 24}},
			},
			{
				Code:   `export {default as promises} from "fs";`,
				Output: []string{`export {default as promises} from "node:fs";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 35}},
			},
			{
				Code:   `import {promises} from 'fs';`,
				Output: []string{`import {promises} from 'node:fs';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 24}},
			},
			{
				Code:   "async function foo() {\n\tconst fs = await import(\"fs/promises\");\n}",
				Output: []string{"async function foo() {\n\tconst fs = await import(\"node:fs/promises\");\n}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 2, Column: 26}},
			},
			{
				Code:   "async function foo() {\n\tconst fs = await import(/* escaped */\"\\u{66}s/promises\");\n}",
				Output: []string{"async function foo() {\n\tconst fs = await import(/* escaped */\"node:\\u{66}s/promises\");\n}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 2, Column: 39}},
			},
			{
				Code:   `import "buffer";`,
				Output: []string{`import "node:buffer";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 8}},
			},
			{
				Code:   `import "child_process";`,
				Output: []string{`import "node:child_process";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 8}},
			},
			{
				Code:   `import "timers/promises";`,
				Output: []string{`import "node:timers/promises";`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 8}},
			},

			// ---- require ----
			{
				Code:   `const {promises} = require("fs")`,
				Output: []string{`const {promises} = require("node:fs")`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 28}},
			},
			{
				Code:   `const fs = require('fs/promises')`,
				Output: []string{`const fs = require('node:fs/promises')`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 20}},
			},

			// ---- process.getBuiltinModule ----
			{
				Code:   `const fs = process.getBuiltinModule("fs")`,
				Output: []string{`const fs = process.getBuiltinModule("node:fs")`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 37}},
			},

			// ---- TypeScript ----
			{
				Code:   `const fs = require("fs") as typeof import("fs");`,
				Output: []string{`const fs = require("node:fs") as typeof import("node:fs");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-node-protocol", Line: 1, Column: 20},
					{MessageId: "prefer-node-protocol", Line: 1, Column: 43},
				},
			},
			{
				Code:   `type fs = import("fs");`,
				Output: []string{`type fs = import("node:fs");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 18}},
			},
			{
				Code:   `type fs = import("fs").fs<"fs">`,
				Output: []string{`type fs = import("node:fs").fs<"fs">`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 18}},
			},
			{
				Code:   `const fs = someFunc() as SomeType<typeof import("fs")>;`,
				Output: []string{`const fs = someFunc() as SomeType<typeof import("node:fs")>;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 49}},
			},
		},
	)
}
