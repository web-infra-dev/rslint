package fs_test

import (
	"maps"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_promises/fs"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func fsError(name string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "preferPromises", Message: "Use 'fs.promises." + name + "()' instead.",
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func runFSTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	nodeGlobals := func(overrides map[string]any) map[string]any {
		globals := map[string]any{"require": "readonly", "process": "readonly", "global": "readonly"}
		maps.Copy(globals, overrides)
		return globals
	}
	for i := range valid {
		if valid[i].FileName == "" {
			valid[i].FileName = "input.js"
		}
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions.SourceType = "module"
		}
		valid[i].Globals = nodeGlobals(valid[i].Globals)
	}
	for i := range invalid {
		if invalid[i].FileName == "" {
			invalid[i].FileName = "input.js"
		}
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions.SourceType = "module"
		}
		invalid[i].Globals = nodeGlobals(invalid[i].Globals)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allowJs.json", t, &fs.PreferPromisesFSRule, valid, invalid)
}

// Every test from eslint-plugin-n v18.3.0, with exact upstream diagnostics.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-promises/fs.js
func TestPreferPromisesFSUpstream(t *testing.T) {
	runFSTests(t,
		[]rule_tester.ValidTestCase{
			// Upstream valid 1.
			{Code: "const fs = require('fs'); fs.createReadStream()"},
			// Upstream valid 2.
			{Code: "const fs = require('fs'); fs.accessSync()"},
			// Upstream valid 3.
			{Code: "const fs = require('fs'); fs.promises.access()"},
			// Upstream valid 4.
			{Code: "const fs = require('node:fs'); fs.promises.access()"},
			// Upstream valid 5.
			{Code: "const {promises} = require('fs'); promises.access()"},
			// Upstream valid 6.
			{Code: "const {promises: fs} = require('fs'); fs.access()"},
			// Upstream valid 7.
			{Code: "const {promises: {access}} = require('fs'); access()"},
			// Upstream valid 8.
			{Code: "import fs from 'fs'; fs.promises.access()"},
			// Upstream valid 9.
			{Code: "import fs from 'node:fs'; fs.promises.access()"},
			// Upstream valid 10.
			{Code: "import * as fs from 'fs'; fs.promises.access()"},
			// Upstream valid 11.
			{Code: "import {promises} from 'fs'; promises.access()"},
			// Upstream valid 12.
			{Code: "import {promises as fs} from 'fs'; fs.access()"},
			// Upstream valid 13.
			{Code: "const fs = process.getBuiltinModule('fs'); fs.promises.access()"},
			// Upstream valid 14.
			{Code: "const fs = process.getBuiltinModule('node:fs'); fs.promises.access()"},
			// Upstream valid 15.
			{Code: "const {promises} = process.getBuiltinModule('fs'); promises.access()"},
			// Upstream valid 16.
			{Code: "const {promises: fs} = process.getBuiltinModule('fs'); fs.access()"},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream invalid 1.
			{Code: "const fs = require('fs'); fs.access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 27, 1, 38),
				}},
			// Upstream invalid 2.
			{Code: "const fs = require('node:fs'); fs.access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 32, 1, 43),
				}},
			// Upstream invalid 3.
			{Code: "const {access} = require('fs'); access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 33, 1, 41),
				}},
			// Upstream invalid 4.
			{Code: "import fs from 'fs'; fs.access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 22, 1, 33),
				}},
			// Upstream invalid 5.
			{Code: "import fs from 'node:fs'; fs.access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 27, 1, 38),
				}},
			// Upstream invalid 6.
			{Code: "import * as fs from 'fs'; fs.access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 27, 1, 38),
				}},
			// Upstream invalid 7.
			{Code: "import {access} from 'fs'; access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 28, 1, 36),
				}},
			// Upstream invalid 8.
			{Code: "const fs = process.getBuiltinModule('fs'); fs.access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 44, 1, 55),
				}},
			// Upstream invalid 9.
			{Code: "const fs = process.getBuiltinModule('node:fs'); fs.access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 49, 1, 60),
				}},
			// Upstream invalid 10.
			{Code: "const {access} = process.getBuiltinModule('fs'); access()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("access", 1, 50, 1, 58),
				}},
			// Upstream: other FS members (invalid 11–39).
			{Code: "const fs = require('fs'); fs.copyFile()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("copyFile", 1, 27, 1, 40),
				}},
			// Upstream invalid 12.
			{Code: "const fs = require('fs'); fs.open()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("open", 1, 27, 1, 36),
				}},
			// Upstream invalid 13.
			{Code: "const fs = require('fs'); fs.rename()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("rename", 1, 27, 1, 38),
				}},
			// Upstream invalid 14.
			{Code: "const fs = require('fs'); fs.truncate()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("truncate", 1, 27, 1, 40),
				}},
			// Upstream invalid 15.
			{Code: "const fs = require('fs'); fs.rmdir()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("rmdir", 1, 27, 1, 37),
				}},
			// Upstream invalid 16.
			{Code: "const fs = require('fs'); fs.mkdir()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("mkdir", 1, 27, 1, 37),
				}},
			// Upstream invalid 17.
			{Code: "const fs = require('fs'); fs.readdir()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readdir", 1, 27, 1, 39),
				}},
			// Upstream invalid 18.
			{Code: "const fs = require('fs');fs.readlink()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readlink", 1, 26, 1, 39),
				}},
			// Upstream invalid 19.
			{Code: "const fs = require('fs'); fs.symlink()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("symlink", 1, 27, 1, 39),
				}},
			// Upstream invalid 20.
			{Code: "const fs = require('fs'); fs.lstat()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("lstat", 1, 27, 1, 37),
				}},
			// Upstream invalid 21.
			{Code: "const fs = require('fs'); fs.stat()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("stat", 1, 27, 1, 36),
				}},
			// Upstream invalid 22.
			{Code: "const fs = require('fs'); fs.link()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("link", 1, 27, 1, 36),
				}},
			// Upstream invalid 23.
			{Code: "const fs = require('fs'); fs.unlink()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("unlink", 1, 27, 1, 38),
				}},
			// Upstream invalid 24.
			{Code: "const fs = require('fs'); fs.chmod()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("chmod", 1, 27, 1, 37),
				}},
			// Upstream invalid 25.
			{Code: "const fs = require('fs'); fs.lchmod()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("lchmod", 1, 27, 1, 38),
				}},
			// Upstream invalid 26.
			{Code: "const fs = require('fs'); fs.lchown()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("lchown", 1, 27, 1, 38),
				}},
			// Upstream invalid 27.
			{Code: "const fs = require('fs'); fs.chown()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("chown", 1, 27, 1, 37),
				}},
			// Upstream invalid 28.
			{Code: "const fs = require('fs'); fs.utimes()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("utimes", 1, 27, 1, 38),
				}},
			// Upstream invalid 29.
			{Code: "const fs = require('fs'); fs.realpath()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("realpath", 1, 27, 1, 40),
				}},
			// Upstream invalid 30.
			{Code: "const fs = require('fs'); fs.mkdtemp()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("mkdtemp", 1, 27, 1, 39),
				}},
			// Upstream invalid 31.
			{Code: "const fs = require('fs'); fs.writeFile()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("writeFile", 1, 27, 1, 41),
				}},
			// Upstream invalid 32.
			{Code: "const fs = require('fs'); fs.appendFile()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("appendFile", 1, 27, 1, 42),
				}},
			// Upstream invalid 33.
			{Code: "const fs = require('fs'); fs.readFile()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 1, 27, 1, 40),
				}},
			// Upstream invalid 34.
			{Code: "const fs = require('fs'); fs.cp()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("cp", 1, 27, 1, 34),
				}},
			// Upstream invalid 35.
			{Code: "const fs = require('fs'); fs.glob()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("glob", 1, 27, 1, 36),
				}},
			// Upstream invalid 36.
			{Code: "const fs = require('fs'); fs.lutimes()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("lutimes", 1, 27, 1, 39),
				}},
			// Upstream invalid 37.
			{Code: "const fs = require('fs'); fs.opendir()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("opendir", 1, 27, 1, 39),
				}},
			// Upstream invalid 38.
			{Code: "const fs = require('fs'); fs.rm()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("rm", 1, 27, 1, 34),
				}},
			// Upstream invalid 39.
			{Code: "const fs = require('fs'); fs.statfs()",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("statfs", 1, 27, 1, 38),
				}},
		},
	)
}

// All four executable documentation examples; rule-enabling comments are omitted.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-promises/fs.md
func TestPreferPromisesFSDocumentation(t *testing.T) {
	runFSTests(t,
		[]rule_tester.ValidTestCase{
			// Documentation example 3.
			{Code: "const { promises: fs } = require(\"fs\")\n\nasync function readData(filePath) {\n    const content = await fs.readFile(filePath, \"utf8\")\n    //...\n}"},
			// Documentation example 4.
			{Code: "import { promises as fs } from \"fs\"\n\nasync function readData(filePath) {\n    const content = await fs.readFile(filePath, \"utf8\")\n    //...\n}"},
		},
		[]rule_tester.InvalidTestCase{
			// Documentation example 1.
			{Code: "const fs = require(\"fs\")\n\nfunction readData(filePath) {\n    fs.readFile(filePath, \"utf8\", (error, content) => {\n        //...\n    })\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 4, 5, 6, 7),
				}},
			// Documentation example 2.
			{Code: "import fs from \"fs\"\n\nfunction readData(filePath) {\n    fs.readFile(filePath, \"utf8\", (error, content) => {\n        //...\n    })\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					fsError("readFile", 4, 5, 6, 7),
				}},
		},
	)
}
