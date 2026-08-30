// TestNoExportsInScriptsUpstream migrates the full valid/invalid suite from
// upstream test/no-exports-in-scripts.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live
// in no_exports_in_scripts_extras_test.go.
package no_exports_in_scripts_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_exports_in_scripts "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_exports_in_scripts"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "no-exports-in-scripts"

func TestNoExportsInScriptsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_exports_in_scripts.NoExportsInScriptsRule,
		[]rule_tester.ValidTestCase{
			// ---- No shebang: the file is a module, exports are allowed ----
			{Code: `export const foo = 1;`, FileName: "file.mjs"},
			{Code: `export default foo;`, FileName: "file.mjs"},
			{Code: `export * from "./foo.js";`, FileName: "file.mjs"},
			{Code: "const foo = 1;\nexport {foo};", FileName: "file.mjs"},
			{Code: `export {};`, FileName: "file.mjs"},

			// ---- TypeScript: type-only exports without a shebang are valid ----
			{Code: `export type Foo = string;`, FileName: "file.ts"},
			{Code: `export interface Foo {}`, FileName: "file.ts"},

			// ---- Shebang present, no export statements ----
			{
				Code: "#!/usr/bin/env node\n" +
					"import process from 'node:process';\n" +
					"\n" +
					"console.log(process.argv);",
				FileName: "file.mjs",
			},
			{
				Code: "#!/usr/bin/env node\n" +
					"const foo = 1;\n" +
					"module.exports = foo;",
				FileName: "file.cjs",
			},

			// ---- Pseudo-shebang in a comment does not count ----
			{
				Code:     "// #!/usr/bin/env node\nexport const foo = 1;",
				FileName: "file.mjs",
			},

			// ---- Pseudo-shebang in a string does not count ----
			{
				Code:     "console.log('#!/usr/bin/env node');\nexport const foo = 1;",
				FileName: "file.mjs",
			},
		},
		[]rule_tester.InvalidTestCase{
			// All invalid cases: file starts with `#!...` and contains an
			// export statement on a subsequent line. Line/Column point at
			// the start of the export (after the leading newline trivia),
			// EndColumn is one past the last character of the export.
			{
				Code:     "#!/usr/bin/env node\nexport const foo = 1;",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 22},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport default foo;",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 20},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport * from './foo.js';",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport * as foo from './foo.js';",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 33},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nconst foo = 1;\nexport {foo};",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 3, Column: 1, EndLine: 3, EndColumn: 14},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport {foo} from './foo.js';",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 30},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport {};",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport const foo = 1;\nexport const bar = 2;",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 22},
					{MessageId: messageID, Line: 3, Column: 1, EndLine: 3, EndColumn: 22},
				},
			},
			// ---- TypeScript: type-only exports under a shebang ----
			{
				Code:     "#!/usr/bin/env node\nexport type Foo = string;",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport interface Foo {}",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
				},
			},
		},
	)
}
