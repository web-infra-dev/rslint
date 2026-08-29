// TestNoExportsInScriptsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking
// a named lock-in.
package no_exports_in_scripts_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_exports_in_scripts "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_exports_in_scripts"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExportsInScriptsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_exports_in_scripts.NoExportsInScriptsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: a no-newline file whose only content is
			// "export" without a shebang still parses as a module and is
			// valid (e.g. the shebang gate must not over-fire on a file that
			// happens to be a single export line).
			{Code: `export const foo = 1;`},

			// ---- Dimension 1: empty file with only a shebang ----
			{Code: "#!/usr/bin/env node\n", FileName: "file.mjs"},

			// ---- Dimension 1: Windows-style CRLF line endings on a
			// shebang line still get treated as a script because the
			// scanner's GetShebang consumes the line and the first line
			// still begins with `#!`.
			{
				Code:     "#!/usr/bin/env node\r\nconst foo = 1;\r\nconsole.log(foo);\r\n",
				FileName: "file.mjs",
			},

			// ---- Dimension 1: shebang + CommonJS — no exports, no error ----
			{
				Code: "#!/usr/bin/env node\n" +
					"const fs = require('node:fs');\n" +
					"console.log(fs.readFileSync(__filename, 'utf8'));",
				FileName: "file.cjs",
			},

			// ---- Dimension 1: shebang + dynamic import (dynamic ESM-via-
			// import is not an export statement, so it stays valid in a
			// script). ----
			{
				Code: "#!/usr/bin/env node\n" +
					"const mod = await import('./some-module.mjs');\n" +
					"console.log(mod);",
				FileName: "file.mjs",
			},

			// ---- Dimension 4: TypeScript with shebang and no exports
			// (e.g. a TypeScript script that uses console) is valid. ----
			{
				Code:     "#!/usr/bin/env node\nconst x: number = 1;\nconsole.log(x);",
				FileName: "file.ts",
			},

			// ---- Dimension 4: an interface-only file under a shebang is
			// not reported — `export interface` is the form upstream
			// reports, and we mirror that; a bare `interface` is not an
			// export and never trips the listener. ----
			{
				Code:     "#!/usr/bin/env node\ninterface Foo { bar: string }\nconsole.log('hi');",
				FileName: "file.ts",
			},

			// ---- Dimension 4: function declaration without the `export`
			// modifier under a shebang is valid. The `reportIfExported`
			// helper only fires for the modifier-bearing variant. ----
			{
				Code:     "#!/usr/bin/env node\nfunction foo() { return 1; }\nconsole.log(foo());",
				FileName: "file.mjs",
			},

			// ---- Upstream parity: TypeScript `export =` is represented as
			// TSExportAssignment by typescript-estree, not an ESTree export
			// declaration, so unicorn v72 does not report it. ----
			{
				Code:     "#!/usr/bin/env node\nconst foo = 1;\nexport = foo;",
				FileName: "file.ts",
			},

			// The directive applies to the authored `export` line, matching
			// upstream's ExportNamedDeclaration wrapper rather than the
			// decorator at the start of the flattened tsgo class node.
			{
				Code:     "#!/usr/bin/env node\n@dec\n// eslint-disable-next-line test\nexport class C {}",
				FileName: "file.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 1: a CRLF shebang still trips the gate, the
			// export is reported on line 2. ----
			{
				Code:     "#!/usr/bin/env node\r\nexport const foo = 1;\r\n",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 22},
				},
			},

			// ---- Dimension 1: shebang + TypeScript `export type` (already
			// covered upstream, but with a different surrounding program to
			// confirm the listener fires regardless of other top-level
			// statements). ----
			{
				Code:     "#!/usr/bin/env node\nconst x = 1;\nexport type Foo = string;",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 3, Column: 1, EndLine: 3, EndColumn: 26},
				},
			},

			// ---- Branch lock-in: empty `export {}` is `KindExportDeclaration`
			// with no specifier list — the listener must still fire. ----
			{
				Code:     "#!/usr/bin/env node\nexport {}",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 10},
				},
			},

			// ---- Upstream parity: typescript-estree wraps an exported
			// import-equals declaration in ExportNamedDeclaration. ----
			{
				Code:     "#!/usr/bin/env node\nexport import Alias = require('pkg');",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 38},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport import Alias = Namespace.Value;",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 39},
				},
			},

			// ---- The same import-equals node can occur inside a namespace;
			// the listener must preserve upstream's nested-node behavior. ----
			{
				Code:     "#!/usr/bin/env node\nnamespace Outer {\n  export import Alias = Namespace.Value;\n}",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 3, Column: 3, EndLine: 3, EndColumn: 41},
				},
			},

			// ---- Real-user: a `#!` line that doesn't look like a typical
			// env-node shebang still triggers the rule. The rule doesn't
			// validate the shebang value, only that the first line begins
			// with `#!`. ----
			{
				Code:     "#!wat\nconst foo = 1;\nexport default foo;",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 3, Column: 1, EndLine: 3, EndColumn: 20},
				},
			},

			// ---- Branch lock-in: `export function` is a
			// KindFunctionDeclaration with ModifierFlagsExport — the
			// `reportIfExported` helper must still fire on it. ----
			{
				Code:     "#!/usr/bin/env node\nexport function foo() {}\n",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 25},
				},
			},

			// ---- Branch lock-in: `export class` is a KindClassDeclaration
			// with ModifierFlagsExport — must still fire. ----
			{
				Code:     "#!/usr/bin/env node\nexport class Foo {}\n",
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 20},
				},
			},

			// ---- Branch lock-in: `export enum` is a KindEnumDeclaration
			// with ModifierFlagsExport — must still fire. ----
			{
				Code:     "#!/usr/bin/env node\nexport enum Color { Red, Green, Blue }\n",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 39},
				},
			},

			// A directive before the decorator does not suppress the export
			// on the following line; the diagnostic begins at `export`.
			{
				Code:     "#!/usr/bin/env node\n// eslint-disable-next-line test\n@dec\nexport class C {}",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 4, Column: 1, EndLine: 4, EndColumn: 18},
				},
			},
			{
				Code:     "#!/usr/bin/env node\n@dec\nexport class C {}",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 3, Column: 1, EndLine: 3, EndColumn: 18},
				},
			},
			{
				Code:     "#!/usr/bin/env node\n@dec\nexport default class C {}",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 3, Column: 1, EndLine: 3, EndColumn: 26},
				},
			},
			{
				Code:     "#!/usr/bin/env node\nexport @dec class C {}",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 23},
				},
			},

			// ---- Real-user: a script that mixes a type export and an
			// interface export on the same line under a shebang — both must
			// be reported (this is also a lock-in that the listener fires
			// on each top-level export, not on the script-level entity). ----
			{
				Code:     "#!/usr/bin/env node\nexport type A = string; export interface B {}",
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
					{MessageId: messageID, Line: 2, Column: 25, EndLine: 2, EndColumn: 46},
				},
			},
		},
	)
}
