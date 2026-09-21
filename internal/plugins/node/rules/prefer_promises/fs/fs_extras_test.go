package fs_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional scope and AST cases checked against eslint-plugin-n v18.3.0.
func TestPreferPromisesFSExtras(t *testing.T) {
	runFSTests(t,
		[]rule_tester.ValidTestCase{
			// Promise-only imports and loaders.
			{Code: "const fs = require('fs/promises'); fs.readFile(); import other from 'node:fs/promises'; other.writeFile(); process.getBuiltinModule('fs/promises').access();"},
			// Synchronous, stream, file descriptor and legacy APIs.
			{Code: "const fs = require('node:fs'); fs.readFileSync(); fs.createWriteStream(); fs.watch(); fs.watchFile(); fs.exists(); fs.read(); fs.write(); fs.close(); fs.fstat(); fs.fsync(); fs.realpath.native();"},
			// Reads, exports, construction and indirect invocation.
			{Code: "import fs, {readFile} from 'fs'; fs.readFile; new readFile(); readFile.call(null); fs.readFile.apply(null, []); fs.readFile.bind(null)(); export {readFile} from 'fs'; export * from 'fs';"},
			// Other modules and local objects.
			{Code: "const fs = require('./fs'); fs.readFile(); const other = {readFile() {}}; other.readFile();"},
			// Shadowed and reassigned loader roots.
			{Code: "function f(require, process) { require('fs').readFile(); process.getBuiltinModule('fs').access(); } require('fs').stat(); require = other; process.getBuiltinModule('fs').rm(); process = other;"},
			// Disabled globals.
			{Code: "require('fs').readFile(); process.getBuiltinModule('fs').readFile();",
				Globals: map[string]any{"require": "off", "process": "off"}},
			// Dynamic module and property names.
			{Code: "const name = 'fs'; require(name).readFile(); const fs = require('fs'); const key = 'readFile'; fs[key](); require().access(); process.getBuiltinModule().access();"},
			// Private keys.
			{Code: "import fs from 'fs'; class C { #readFile; read() { fs.#readFile(); } }"},
			// Dynamic imports and rest patterns.
			{Code: "const fs = await import('fs'); fs.readFile(); const {...rest} = require('fs'); rest.readFile();"},
			// JSX tags and type references.
			{Code: "import fs from 'fs'; type Read = typeof fs.readFile; const el = <fs.readFile />;",
				FileName: "input.tsx"},
			// TypeScript import-equals.
			{Code: "import fs = require('fs'); fs.readFile();",
				FileName: "input.ts"},
			// Scoped process import.
			{Code: "import process from 'node:process'; process.getBuiltinModule('fs').readFile();"},
		},
		[]rule_tester.InvalidTestCase{
			// Inline module access, parentheses and JSDoc casts.
			{Code: "(require('fs').readFile)('file', cb); (/** @type {any} */ (require('node:fs'))).writeFile('file', data, cb);",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 1, 1, 37),
					fsError("writeFile", 1, 39, 1, 108),
				}},
			// Aliased loaders.
			{Code: "const load = require; load('fs').readFile(); const {getBuiltinModule: builtin} = process; builtin('node:fs').access();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 23, 1, 44),
					fsError("access", 1, 91, 1, 118),
				}},
			// Named ESM and namespace default aliases.
			{Code: "import {readFile as read} from 'node:fs'; read(); import * as fs from 'fs'; const {default: api} = fs; api.writeFile();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 43, 1, 49),
					fsError("writeFile", 1, 104, 1, 119),
				}},
			// Static computed properties and destructuring.
			{Code: "const fs = require('f' + 's'); fs['read' + 'File'](); fs[`writeFile`](); const {[`access`]: run} = fs; run();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 32, 1, 53),
					fsError("writeFile", 1, 55, 1, 72),
					fsError("access", 1, 104, 1, 109),
				}},
			// Optional loaders, members and calls.
			{Code: "require?.('fs')?.readFile?.(); process?.getBuiltinModule?.('node:fs')?.access?.(); (require('fs')?.stat)();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 1, 1, 30),
					fsError("access", 1, 32, 1, 82),
					fsError("stat", 1, 84, 1, 107),
				}},
			// Assignment and parameter defaults.
			{Code: "let readFile; ({readFile} = require('fs')); readFile(); function f({writeFile: run} = require('fs')) { run(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 45, 1, 55),
					fsError("writeFile", 1, 104, 1, 109),
				}},
			// Flow-insensitive module and method aliases.
			{Code: "let fs = require('fs'); const read = fs.readFile; fs = other; read(); fs.writeFile();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 63, 1, 69),
					fsError("writeFile", 1, 71, 1, 85),
				}},
			// Global object loaders.
			{Code: "global.require('fs').readFile(); globalThis.process.getBuiltinModule('fs').access();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 1, 1, 32),
					fsError("access", 1, 34, 1, 84),
				}},
			// Independent paths preserve duplicate diagnostics.
			{Code: "const fs = flag ? require('fs') : require('node:fs'); fs.readFile();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 55, 1, 68),
					fsError("readFile", 1, 55, 1, 68),
				}},
			// Nested calls and mixed loaders use source order.
			{Code: "import {readFile} from 'fs'; readFile(); process.getBuiltinModule('fs').access(); require('fs').writeFile(); const fs = require('fs'); fs.stat(fs.readdir());",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 30, 1, 40),
					fsError("access", 1, 42, 1, 81),
					fsError("writeFile", 1, 83, 1, 108),
					fsError("stat", 1, 136, 1, 157),
					fsError("readdir", 1, 144, 1, 156),
				}},
			// TypeScript assertions, non-null receivers and generic calls.
			{Code: "import fs from 'fs'; (fs as any).readFile(); fs!.writeFile(); (fs satisfies any).stat(); fs.readFile<string>();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 22, 1, 44),
					fsError("writeFile", 1, 46, 1, 61),
					fsError("stat", 1, 63, 1, 88),
					fsError("readFile", 1, 90, 1, 111),
				}},
			// JSX expression containers.
			{Code: "import fs from 'fs'; const el = <fs.readFile value={fs.readFile()}>{fs.writeFile()}</fs.readFile>;",
				FileName: "input.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 53, 1, 66),
					fsError("writeFile", 1, 69, 1, 83),
				}},
			// Multiline callback and UTF-16 ranges.
			{Code: "const fs = require('fs');\nconst text = '😀'; fs.readFile(\n  '路径',\n  (error, content) => {}\n);",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 2, 20, 5, 2),
				}},
			// Local shadowing preserves outer references.
			{Code: "const fs = require('fs'); function f(fs) { fs.readFile(); } fs.readFile();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 61, 1, 74),
				}},
			// Callback presence does not affect reporting.
			{Code: "const fs = require('fs'); fs.readFile('file'); fs.readFile('file', callback); await fs.readFile('file');",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 27, 1, 46),
					fsError("readFile", 1, 48, 1, 77),
					fsError("readFile", 1, 85, 1, 104),
				}},
		},
	)
}

// Additional language and configuration cases checked against eslint-plugin-n v18.3.0.
func TestPreferPromisesFSLanguageEdges(t *testing.T) {
	runFSTests(t,
		[]rule_tester.ValidTestCase{
			// Tagged templates and type queries are not calls.
			{Code: "import fs from 'fs'; fs.readFile`file`; type F = ReturnType<typeof fs.readFile>; interface I extends fs.readFile {}",
				FileName: "input.ts",
			},
			// CommonJS wrapper shadow.
			{Code: "var require; require('fs').readFile();",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// Inline globals override configured globals.
			{Code: "/* global require: off, process: off */\nrequire('fs').readFile(); process.getBuiltinModule('fs').writeFile();"},
		},
		[]rule_tester.InvalidTestCase{
			// Escaped module names and identifiers.
			{Code: "const fs = require('\\x66s'); fs.re\\u0061dFile();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 30, 1, 48),
				},
			},
			// Computed assertions and generic method aliases.
			{Code: "import fs from 'fs'; fs['readFile' as const](); const read = fs.readFile<string>; read(); (<any>fs).writeFile();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 22, 1, 47),
					fsError("readFile", 1, 83, 1, 89),
					fsError("writeFile", 1, 91, 1, 112),
				},
			},
			// Class evaluation and parameter properties.
			{Code: "import fs from 'fs'; class C extends fs.readFile() { [fs.readdir()]() {} constructor(private io = require('fs')) { super(); io.stat(); this.io.stat(); } }",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 38, 1, 51),
					fsError("readdir", 1, 55, 1, 67),
					fsError("stat", 1, 125, 1, 134),
				},
			},
			// Type-only imports used at runtime.
			{Code: "import type {readFile} from 'fs'; readFile();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 35, 1, 45),
				},
			},
			// JSDoc imports do not shadow runtime require.
			{Code: "/** @import {require} from 'other' */\nrequire('fs').readFile();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 2, 1, 2, 25),
				},
			},
			// Resource declarations preserve module aliases.
			{Code: "using fs = require('fs'); fs.readFile();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 27, 1, 40),
				},
			},
			// Assignment defaults and alias cycles.
			{Code: "let read; ({readFile: read = fallback} = require('fs')); read(); let fs = require('fs'); let alias = fs; fs = alias; alias.writeFile();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 58, 1, 64),
					fsError("writeFile", 1, 118, 1, 135),
				},
			},
			// CommonJS loader.
			{Code: "const fs = require('fs'); fs.readFile();",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 27, 1, 40),
				},
			},
			// Script loader.
			{Code: "var fs = require('fs'); fs.readFile();",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 25, 1, 38),
				},
			},
			// Disable and re-enable directives.
			// RuleTester registers the rule as "test".
			{Code: "const fs = require('fs');\n// eslint-disable-next-line test\nfs.readFile();\n/* eslint-disable test */\nfs.writeFile();\n/* eslint-enable test */\nfs.access();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 7, 1, 7, 12),
				},
			},
			// Different methods converging on one call keep upstream order.
			{Code: "import {writeFile} from 'fs'; const fs = require('fs'); const run = flag ? writeFile : fs.readFile; run();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 101, 1, 106),
					fsError("writeFile", 1, 101, 1, 106),
				},
			},
			// Conditional, logical and sequence module values.
			{Code: "(flag ? require('fs') : other).readFile(); (require('fs') || other).writeFile(); (0, require('fs')).access();",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 1, 1, 42),
					fsError("writeFile", 1, 44, 1, 80),
					fsError("access", 1, 82, 1, 109),
				},
			},
			// Decorators and enum initializers.
			{Code: "import fs from 'fs'; @fs.readFile() class C {} enum E { value = fs.stat() }",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 23, 1, 36),
					fsError("stat", 1, 65, 1, 74),
				},
			},
			// Type declarations do not shadow value aliases.
			{Code: "import fs from 'fs'; type fs = {}; fs.readFile();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 36, 1, 49),
				},
			},
		},
	)
}
