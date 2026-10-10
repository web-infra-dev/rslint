// TestImportStyleUpstream migrates every semantic case from
// eslint-plugin-unicorn v77.0.0 test/import-style.js. Snapshot-only duplicates
// are retained in addition to the main matrix so the upstream grouping remains
// auditable. Rslint-specific AST and diagnostic lock-ins live in the extras file.
package import_style_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	import_style "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/import_style"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func styleOptions() []any {
	return []any{map[string]any{
		"checkExportFrom": true,
		"styles": map[string]any{
			"unassigned": map[string]any{"unassigned": true, "named": false},
			"default":    map[string]any{"default": true, "named": false},
			"namespace":  map[string]any{"namespace": true, "named": false},
			"named":      map[string]any{"named": true},
		},
	}}
}

func bannedOptions(checkExport bool, extendDefaults bool) []any {
	value := map[string]any{
		"styles": map[string]any{"banned": map[string]any{
			"unassigned": false, "default": false, "namespace": false, "named": false,
		}},
		"extendDefaultStyles": extendDefaults,
	}
	if checkExport {
		value["checkExportFrom"] = true
	}
	return []any{value}
}

func withStyles(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js", Options: styleOptions()}
}

func lineColumn(code, needle string) (int, int) {
	index := strings.Index(code, needle)
	if index < 0 {
		panic("missing diagnostic start needle: " + needle)
	}
	line := strings.Count(code[:index], "\n") + 1
	lineStart := strings.LastIndex(code[:index], "\n") + 1
	return line, utf8.RuneCountInString(code[lineStart:index]) + 1
}

func styleInvalid(code, start, allowed, module string, options []any) rule_tester.InvalidTestCase {
	line, column := lineColumn(code, start)
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "file.js", Options: options,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId:   "importStyle",
			Message:     "Use " + allowed + " import for module `" + module + "`.",
			Line:        line,
			Column:      column,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
		}},
	}
}

func defaultStyleInvalid(code, start, allowed, module string) rule_tester.InvalidTestCase {
	return styleInvalid(code, start, allowed, module, styleOptions())
}

func bannedInvalid(code, start string, options []any, fileName string) rule_tester.InvalidTestCase {
	line, column := lineColumn(code, start)
	return rule_tester.InvalidTestCase{
		Code: code, FileName: fileName, Options: options,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "importStyleBanned",
			Message:   "All import styles are disabled for module `banned`. Use the `no-restricted-imports` rule to disallow a module.",
			Line:      line, Column: column,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
		}},
	}
}

func TestImportStyleUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{}
	for _, code := range []string{
		`require('unassigned')`, `const {} = require('unassigned')`, `import 'unassigned'`,
		`import {} from 'unassigned'`, `import('unassigned')`, `export {} from 'unassigned'`,
		`const x = require('default')`, `const {default: x} = require('default')`, `const [] = require("default")`,
		`import x from 'default'`, "async () => {\n\tconst {default: x} = await import('default');\n}", `export {default} from 'default'`,
		`const x = require('namespace')`, `const [] = require("namespace")`, `import * as x from 'namespace'`,
		"async () => {\n\tconst x = await import('namespace');\n}", `export * from 'namespace'`,
		`const {x} = require('named')`, `const {...rest} = require("named")`, `const {x: y} = require('named')`,
		`import {x} from 'named'`, `import {x as y} from 'named'`,
		"async () => {\n\tconst {x} = await import('named');\n}", "async () => {\n\tconst {x: y} = await import('named');\n}",
		`export {x} from 'named'`, `export {x as y} from 'named'`,
		`const foo = 1; export {foo}`, `export const foo = 1;`, `export function foo() {}`,
	} {
		valid = append(valid, withStyles(code))
	}

	for _, code := range []string{
		`import {inspect} from 'util'`, `import {inspect} from 'node:util'`,
		`const {inspect} = require('util')`, `const {inspect} = require('node:util')`,
		`import chalk from 'chalk'`, `import {default as chalk} from 'chalk'`,
		`export {promisify, callbackify} from 'util'`, `export {promisify, callbackify} from 'node:util'`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: "file.js"})
	}

	valid = append(valid,
		rule_tester.ValidTestCase{Code: `require('chalk')`, FileName: "file.js", Options: []any{map[string]any{"styles": map[string]any{}, "extendDefaultStyles": false}}},
		// Upstream documentation example: adding named keeps path's default style too.
		rule_tester.ValidTestCase{Code: `import {join} from 'path'`, FileName: "file.js", Options: []any{map[string]any{"styles": map[string]any{"path": map[string]any{"named": true}}}}},
		rule_tester.ValidTestCase{Code: `import 'chalk'`, FileName: "file.js", Options: []any{map[string]any{"checkImport": false}}},
		rule_tester.ValidTestCase{Code: "async () => {\n\tconst {red} = await import('chalk');\n}", FileName: "file.js", Options: []any{map[string]any{"checkDynamicImport": false}}},
		rule_tester.ValidTestCase{Code: `import('chalk')`, FileName: "file.js", Options: []any{map[string]any{"checkDynamicImport": false}}},
		rule_tester.ValidTestCase{Code: `require('chalk')`, FileName: "file.js", Options: []any{map[string]any{"checkRequire": false}}},
		rule_tester.ValidTestCase{Code: `const {red} = require('chalk')`, FileName: "file.js", Options: []any{map[string]any{"checkRequire": false}}},
		rule_tester.ValidTestCase{Code: `import util, {inspect} from 'named-or-default'`, FileName: "file.js", Options: []any{map[string]any{"styles": map[string]any{"named-or-default": map[string]any{"named": true, "default": true}}}}},
	)
	for _, code := range []string{
		`import fs from 'node:fs'`, `import * as fs from 'node:fs'`, `import {readFile} from 'node:fs'`,
		`import fsPromises from 'node:fs/promises'`, `import path from 'node:path'`, `import {inspect} from 'node:util'`,
		"async () => {\n\tconst {inspect} = await import('node:util');\n}", `import fs from 'fs'`,
		`import unknown from 'node:unknown'`, `const fs = require('node:fs')`, `import('node:unknown')`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: "file.js"})
	}
	valid = append(valid,
		rule_tester.ValidTestCase{Code: `import * as fs from 'node:fs'`, FileName: "file.js", Options: []any{map[string]any{"styles": map[string]any{"fs": map[string]any{"namespace": true}}}}},
	)
	for _, code := range []string{
		`require(1, 2, 3)`, `require(variable)`, `const x = require(variable)`,
		`const x = require('unassigned').x`, "async () => {\n\tconst {red} = await import(variable);\n}",
	} {
		valid = append(valid, withStyles(code))
	}
	valid = append(valid,
		rule_tester.ValidTestCase{Code: "\n\timport util from \"node:util\";\n\timport * as util2 from \"node:util\";\n\timport {foo} from \"node:util\";\n", FileName: "file.js", Options: []any{map[string]any{"styles": map[string]any{"util": false}}}},
		rule_tester.ValidTestCase{Code: "\n\timport util from \"node:util\";\n\timport * as util2 from \"node:util\";\n\timport {foo} from \"node:util\";\n", FileName: "file.js", Options: []any{map[string]any{"styles": map[string]any{"util": map[string]any{"named": false}}}}},
	)

	invalid := []rule_tester.InvalidTestCase{}
	appendGroup := func(allowed, module string, cases [][2]string) {
		for _, item := range cases {
			invalid = append(invalid, defaultStyleInvalid(item[0], item[1], allowed, module))
		}
	}
	appendGroup("unassigned", "unassigned", [][2]string{
		{`const {x} = require('unassigned')`, `{x}`}, {`const {default: x} = require('unassigned')`, `{default`},
		{`import x from 'unassigned'`, `import`}, {"async () => {\n\tconst {default: x} = await import('unassigned');\n}", `{default`},
		{`const x = require('unassigned')`, `x =`}, {`import * as x from 'unassigned'`, `import`},
		{"async () => {\n\tconst x = await import('unassigned');\n}", `x =`}, {`const {x: y} = require('unassigned')`, `{x:`},
		{`import {x} from 'unassigned'`, `import`}, {`import {x as y} from 'unassigned'`, `import`},
		{"async () => {\n\tconst {x} = await import('unassigned');\n}", `{x}`},
		{"async () => {\n\tconst {x: y} = await import('unassigned');\n}", `{x:`},
		{`const {...rest} = require("unassigned")`, `{...rest}`}, {`const [] = require("unassigned")`, `[] =`},
		{`export * from 'unassigned'`, `export`}, {`export {x} from 'unassigned'`, `export`},
		{`export {x as y} from 'unassigned'`, `export`}, {`export {default} from 'unassigned'`, `export`},
	})
	appendGroup("default", "default", [][2]string{
		{`require('default')`, `require`}, {`const {} = require('default')`, `{} =`},
		{`const {...rest} = require("default")`, `{...rest}`}, {`import 'default'`, `import`},
		{`import {} from 'default'`, `import`}, {`import('default')`, `import`},
		{`import * as x from 'default'`, `import`}, {"async () => {\n\tconst x = await import('default');\n}", `x =`},
		{`const {x} = require('default')`, `{x}`}, {`const {x: y} = require('default')`, `{x:`},
		{`import {x} from 'default'`, `import`}, {`import {x as y} from 'default'`, `import`},
		{"async () => {\n\tconst {x} = await import('default');\n}", `{x}`},
		{"async () => {\n\tconst {x: y} = await import('default');\n}", `{x:`},
		{`export * from 'default'`, `export`}, {`export {x} from 'default'`, `export`},
		{`export {x as y} from 'default'`, `export`},
	})
	appendGroup("namespace", "namespace", [][2]string{
		{`require('namespace')`, `require`}, {`const {} = require('namespace')`, `{} =`},
		{`import 'namespace'`, `import`}, {`import {} from 'namespace'`, `import`}, {`import('namespace')`, `import`},
		{`const {default: x} = require('namespace')`, `{default`}, {`const {...rest} = require("namespace")`, `{...rest}`},
		{`import x from 'namespace'`, `import`}, {`const {x} = require('namespace')`, `{x}`},
		{`const {x: y} = require('namespace')`, `{x:`}, {`import {x} from 'namespace'`, `import`},
		{`import {x as y} from 'namespace'`, `import`}, {"async () => {\n\tconst {x} = await import('namespace');\n}", `{x}`},
		{"async () => {\n\tconst {x: y} = await import('namespace');\n}", `{x:`},
		{`export {x} from 'namespace'`, `export`}, {`export {x as y} from 'namespace'`, `export`},
		{`export {default} from 'namespace'`, `export`},
	})
	appendGroup("named", "named", [][2]string{
		{`require('named')`, `require`}, {`const {} = require('named')`, `{} =`}, {`const [] = require("named")`, `[] =`},
		{`import 'named'`, `import`}, {`import {} from 'named'`, `import`}, {`import('named')`, `import`},
		{`const x = require('named')`, `x =`}, {`const {default: x} = require('named')`, `{default`},
		{`import x from 'named'`, `import`}, {"async () => {\n\tconst {default: x} = await import('named');\n}", `{default`},
		{"async () => {\n\tconst [x] = await import('named');\n}", `[x]`}, {`import * as x from 'named'`, `import`},
		{"async () => {\n\tconst x = await import('named');\n}", `x =`}, {`export * from 'named'`, `export`},
		{`export {default} from 'named'`, `export`}, {`import util, {inspect} from 'named'`, `import`},
	})
	invalid = append(invalid,
		defaultStyleInvalid(`import util, {inspect} from 'default'`, `import`, "default", "default"),
		styleInvalid(`import * as path from 'node:path'`, `import`, "default", "node:path", nil),
		styleInvalid(`import util from 'node:util'`, `import`, "named", "node:util", nil),
		styleInvalid(`import * as util from 'node:util'`, `import`, "named", "node:util", nil),
		styleInvalid(`import('node:util')`, `import`, "named", "node:util", nil),
		styleInvalid("async () => {\n\tconst util = await import('node:util');\n}", `util =`, "named", "node:util", nil),
		styleInvalid(`export * from 'node:util'`, `export`, "named", "node:util", styleOptions()),
		styleInvalid(`import * as fs from 'node:fs'`, `import`, "default", "node:fs", []any{map[string]any{"styles": map[string]any{"fs": map[string]any{"default": true}}}}),
	)
	for _, item := range [][4]string{
		{`import util from 'util'`, `import`, "named", "util"}, {`import * as util from 'util'`, `import`, "named", "util"},
		{`import util from 'node:util'`, `import`, "named", "node:util"}, {`const util = require('util')`, `util =`, "named", "util"},
		{`const util = require('node:util')`, `util =`, "named", "node:util"}, {`require('util')`, `require`, "named", "util"},
		{`require('node:util')`, `require`, "named", "node:util"}, {`require('ut' + 'il')`, `require`, "named", "util"},
		{`require('node:' + 'util')`, `require`, "named", "node:util"}, {`import {red} from 'chalk'`, `import`, "default", "chalk"},
		{`import {red as green} from 'chalk'`, `import`, "default", "chalk"},
		{"async () => {\n\tconst {red} = await import('chalk');\n}", `{red}`, "default", "chalk"},
	} {
		invalid = append(invalid, styleInvalid(item[0], item[1], item[2], item[3], nil))
	}
	invalid = append(invalid,
		styleInvalid(`require('no-unassigned')`, `require`, "named, namespace, or default", "no-unassigned", []any{map[string]any{"styles": map[string]any{"no-unassigned": map[string]any{"named": true, "namespace": true, "default": true}}}}),
		styleInvalid("\n\timport * as util from \"node:util\";\n", `import`, "named or default", "node:util", []any{map[string]any{"styles": map[string]any{"util": map[string]any{"default": true}}}}),
		styleInvalid("\n\timport {promisify} from \"node:util\";\n", `import`, "default", "node:util", []any{map[string]any{"styles": map[string]any{"util": map[string]any{"default": true, "named": false}}}}),
	)

	// Upstream TypeScript group.
	valid = append(valid,
		rule_tester.ValidTestCase{Code: `import type chalk from 'chalk'`, FileName: "file.ts", Tsx: false, Options: []any{map[string]any{"extendDefaultStyles": false, "styles": map[string]any{"chalk": map[string]any{"named": true}, "named": map[string]any{"default": true}}}}},
		rule_tester.ValidTestCase{Code: `import type {x} from 'named'`, FileName: "file.ts", Tsx: false, Options: []any{map[string]any{"extendDefaultStyles": false, "styles": map[string]any{"chalk": map[string]any{"named": true}, "named": map[string]any{"default": true}}}}},
		rule_tester.ValidTestCase{Code: `import type {ChalkInstance} from 'chalk'`, FileName: "file.ts", Tsx: false},
		rule_tester.ValidTestCase{Code: `import {type ChalkInstance} from 'chalk'`, FileName: "file.ts", Tsx: false},
		rule_tester.ValidTestCase{Code: `import chalk, {type ChalkInstance} from 'chalk'`, FileName: "file.ts", Tsx: false},
		rule_tester.ValidTestCase{Code: `let a`, FileName: "file.js"}, // upstream snapshot valid
	)
	tsInvalid := styleInvalid(`import {type ChalkInstance, red} from 'chalk'`, `import`, "default", "chalk", nil)
	tsInvalid.FileName, tsInvalid.Tsx = "file.ts", false
	invalid = append(invalid, tsInvalid)
	for _, code := range []string{`import type {Foo} from 'banned'`, `import {type Foo} from 'banned'`} {
		invalid = append(invalid, bannedInvalid(code, `import`, bannedOptions(false, false), "file.ts"))
	}

	// Upstream snapshot group, including deliberate duplicates of the main suite.
	for _, item := range [][4]string{
		{`import util from 'util'`, `import`, "named", "util"}, {`import * as util from 'util'`, `import`, "named", "util"},
		{`import util from 'node:util'`, `import`, "named", "node:util"}, {`const util = require('util')`, `util =`, "named", "util"},
		{`const util = require('node:util')`, `util =`, "named", "node:util"}, {`require('util')`, `require`, "named", "util"},
		{`require('node:util')`, `require`, "named", "node:util"}, {`import {red} from 'chalk'`, `import`, "default", "chalk"},
		{`import {red as green} from 'chalk'`, `import`, "default", "chalk"},
		{"async () => {\n\tconst {red} = await import('chalk');\n}", `{red}`, "default", "chalk"},
	} {
		invalid = append(invalid, styleInvalid(item[0], item[1], item[2], item[3], nil))
	}
	for _, item := range [][2]string{
		{`import 'banned'`, `import`}, {`import foo from 'banned'`, `import`},
		{`import * as foo from 'banned'`, `import`}, {`import {foo} from 'banned'`, `import`},
		{"async () => {\n\tconst foo = await import('banned');\n}", `foo =`}, {`import('banned')`, `import`},
		{`const foo = require('banned')`, `foo =`}, {`require('banned')`, `require`},
	} {
		invalid = append(invalid, bannedInvalid(item[0], item[1], bannedOptions(false, false), "file.js"))
	}
	invalid = append(invalid,
		bannedInvalid(`export {foo} from 'banned'`, `export`, bannedOptions(true, false), "file.js"),
		bannedInvalid(`export * from 'banned'`, `export`, bannedOptions(true, false), "file.js"),
		bannedInvalid(`import 'banned'`, `import`, bannedOptions(false, true), "file.js"),
	)

	// Upstream no-argument require group.
	for _, code := range []string{`require();`, `const a = require();`, `const {a} = require();`, `const [a] = require();`, `export const a = require();`} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: "file.js"})
	}
	invalid = append(invalid,
		styleInvalid(`require("chalk");`, `require`, "default", "chalk", nil),
		styleInvalid(`const {red} = require("chalk");`, `{red}`, "default", "chalk", nil),
	)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &import_style.ImportStyleRule, valid, invalid)
}
