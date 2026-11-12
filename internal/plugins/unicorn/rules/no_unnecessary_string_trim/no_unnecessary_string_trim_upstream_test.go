package no_unnecessary_string_trim_test

// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-unnecessary-string-trim.js

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_string_trim"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryStringTrimUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `import {z} from "zod"; z.string().trim().startsWith("asdf")`, FileName: "file.js"},
		{Code: `import {z as schema} from "zod"; schema.string().trim().startsWith("asdf")`, FileName: "file.js"},
		{Code: `import {"z" as schema} from "zod"; schema.string().trim().startsWith("asdf")`, FileName: "file.js"},
		{Code: `import * as z from "zod"; z.string().trim().endsWith("asdf")`, FileName: "file.js"},
		{Code: `import {z} from "zod"; (z.string() as unknown).trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `import {z} from "zod"; (z.string() satisfies unknown).trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `import {z} from "zod"; z.string()!.trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `import {z} from "zod"; (<unknown>z.string()).trim().startsWith("-");`, FileName: "file.ts"},
		// Rslint skips these upstream invalid cases to preserve valid Zod schemas.
		{Code: `import z from "zod"; z.string().trim().startsWith("-")`, FileName: "file.js"},
		{Code: `import {z} from "zod"; z["string"]().trim().startsWith("-")`, FileName: "file.js"},
		{Code: `import {z} from "zod"; z.string("argument").trim().startsWith("-")`, FileName: "file.js"},
		{Code: `import z = require("zod"); z.string().trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `foo.trimStart().startsWith("-")`, FileName: "file.js"},
		{Code: `foo.trimEnd().endsWith("-")`, FileName: "file.js"},
		{Code: `foo.trim().includes("-")`, FileName: "file.js"},
		{Code: `foo.trim().startsWith("-", 1)`, FileName: "file.js"},
		{Code: `foo.trim().endsWith("-", 1)`, FileName: "file.js"},
		{Code: `foo.trim(" ").startsWith("-")`, FileName: "file.js"},
		{Code: `foo.trim().startsWith(prefix)`, FileName: "file.js"},
		{Code: `foo.trim().endsWith(suffix)`, FileName: "file.js"},
		{Code: `foo.trim().startsWith("foo ")`, FileName: "file.js"},
		{Code: "foo.trim().startsWith(`foo `)", FileName: "file.js"},
		{Code: `foo.trim().endsWith(" foo")`, FileName: "file.js"},
		{Code: "foo.trim().endsWith(` foo`)", FileName: "file.js"},
		{Code: `const modes = new Set(['foo']); modes.clear(); value.trim().startsWith(modes.size ? 'x' : 'x ')`, FileName: "file.js"},
		{Code: `const modes = new Set(['foo']); modes.clear(); value.trim().endsWith((modes.size && 'x') || suffix)`, FileName: "file.js"},
		{Code: `const object = {value: true}; Object.defineProperty(object, 'value', {get() { return false; }}); value.trim().startsWith(object.value ? 'x' : suffix)`, FileName: "file.js"},
		{Code: `const modes = new Set(['foo']); modes.clear(); value.trim().startsWith((modes.size ? 'x' : 'x ') as string)`, FileName: "file.ts"},
		{Code: `const prefix = "foo "; foo.trim().startsWith(prefix)`, FileName: "file.js"},
		{Code: `const suffix = " foo"; foo.trim().endsWith(suffix)`, FileName: "file.js"},
		{Code: `foo.trim().startsWith("foo" + " ")`, FileName: "file.js"},
		{Code: `foo.trim().startsWith(/foo/)`, FileName: "file.js"},
		{Code: `foo.trim().endsWith(Symbol("foo"))`, FileName: "file.js"},
		{Code: `const value = {trim() { return "ok"; }}; value.trim().startsWith("o")`, FileName: "file.js"},
		{Code: `const value = []; value.trim().startsWith("-")`, FileName: "file.js"},
		{Code: `const value = 1; value.trim().startsWith("-")`, FileName: "file.js"},
		{Code: `foo.trim().startsWith("foo " as const)`, FileName: "file.ts"},
		{Code: `foo.trim().startsWith(...argumentsArray)`, FileName: "file.js"},
		{Code: `foo.trim().endsWith(...argumentsArray)`, FileName: "file.js"},
		{Code: `foo.trim().startsWith("-", ...positions)`, FileName: "file.js"},
		{Code: `foo.trim().endsWith("-", ...positions)`, FileName: "file.js"},
		{Code: `foo.trim().startsWith?.("-")`, FileName: "file.js"},
		{Code: `foo.trim().endsWith?.("-")`, FileName: "file.js"},
		{Code: `foo.trim?.().startsWith("-")`, FileName: "file.js"},
		{Code: `foo.trim?.().endsWith("-")`, FileName: "file.js"},
		{Code: `foo?.trim().startsWith("-")`, FileName: "file.js"},
		{Code: `foo.trim()?.startsWith("-")`, FileName: "file.js"},
		{Code: `foo?.trim()?.startsWith("-")`, FileName: "file.js"},
		{Code: `foo?.trim().endsWith("-")`, FileName: "file.js"},
		{Code: `foo.trim()?.endsWith("-")`, FileName: "file.js"},
		{Code: `foo?.trim()?.endsWith("-")`, FileName: "file.js"},
		{Code: `foo?.bar.trim().startsWith("-")`, FileName: "file.js"},
		{Code: `(foo?.bar).trim().startsWith("-")`, FileName: "file.js"},
		{Code: `(foo?.bar).baz.trim().startsWith("-")`, FileName: "file.js"},
		{Code: `foo.trim()["startsWith"]("-")`, FileName: "file.js"},
		{Code: `foo["trim"]().startsWith("-")`, FileName: "file.js"},
		{Code: `trim().startsWith("-")`, FileName: "file.js"},
		{Code: `foo.trim.startsWith("-")`, FileName: "file.js"},
		{Code: `new foo.trim().startsWith("-")`, FileName: "file.js"},
		{Code: `function foo(value: number[]) { value.trim().startsWith("-"); }`, FileName: "file.ts"},

		// Examples from the rule documentation.
		{Code: `const result = value.trimStart().startsWith('-');`, FileName: "file.js"},
		{Code: `const result = value.trimEnd().endsWith('-');`, FileName: "file.js"},
		{Code: `const result = value.trim().startsWith('-', 1);`, FileName: "file.js"},
		{Code: `const result = value.trim().startsWith('foo ');`, FileName: "file.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		invalidTrimCase(`foo.trim().startsWith("-")`, "trimStart", 1, 5, "file.js"),
		invalidTrimCase(`foo.trim().endsWith("-")`, "trimEnd", 1, 5, "file.js"),
		invalidTrimCase(`foo.trim().startsWith()`, "trimStart", 1, 5, "file.js"),
		invalidTrimCase(`foo.trim().endsWith()`, "trimEnd", 1, 5, "file.js"),
		invalidTrimCase(`" foo ".trim().startsWith("f")`, "trimStart", 1, 9, "file.js"),
		invalidTrimCase(`" foo ".trim().endsWith("o")`, "trimEnd", 1, 9, "file.js"),
		invalidTrimCase(`foo.trim().startsWith(" foo")`, "trimStart", 1, 5, "file.js"),
		invalidTrimCase(`foo.trim().endsWith("foo ")`, "trimEnd", 1, 5, "file.js"),
		invalidTrimCase("foo.trim().startsWith(`foo`)", "trimStart", 1, 5, "file.js"),
		invalidTrimCase("foo.trim().endsWith(`foo`)", "trimEnd", 1, 5, "file.js"),
		invalidTrimCase(`const prefix = "foo"; foo.trim().startsWith(prefix)`, "trimStart", 1, 27, "file.js"),
		invalidTrimCase(`const suffix = "foo"; foo.trim().endsWith(suffix)`, "trimEnd", 1, 27, "file.js"),
		invalidTrimCase(`foo.trim().startsWith("foo" + "bar")`, "trimStart", 1, 5, "file.js"),
		invalidTrimCase(`import {z} from "other-package"; z.string().trim().startsWith("-")`, "trimStart", 1, 45, "file.js"),
		invalidTrimCase(`import * as z from "other-package"; z.string().trim().startsWith("-")`, "trimStart", 1, 48, "file.js"),
		invalidTrimCase(`import {z} from "zod/mini"; z.string().trim().startsWith("-")`, "trimStart", 1, 40, "file.js"),
		invalidTrimCase(`import {z} from "zod"; z.number().trim().startsWith("-")`, "trimStart", 1, 35, "file.js"),
		invalidTrimCase(`import {z} from "zod"; function foo(z) { z.string().trim().startsWith("-"); }`, "trimStart", 1, 53, "file.js"),
		invalidTrimCase(`import type {z} from "zod"; z.string().trim().startsWith("-");`, "trimStart", 1, 40, "file.ts"),
		invalidTrimCase(`import type * as z from "zod"; z.string().trim().endsWith("-");`, "trimEnd", 1, 43, "file.ts"),
		invalidTrimCase(`foo.trim().startsWith(undefined)`, "trimStart", 1, 5, "file.js"),
		invalidTrimCase(`foo.trim().endsWith(void 0)`, "trimEnd", 1, 5, "file.js"),
		invalidTrimCase(`const search = undefined; foo.trim().startsWith(search)`, "trimStart", 1, 31, "file.js"),
		invalidTrimCase(`foo.trim().startsWith(123)`, "trimStart", 1, 5, "file.js"),
		invalidTrimCase(`foo.trim().endsWith(false)`, "trimEnd", 1, 5, "file.js"),
		invalidTrimCase(`foo.trim().startsWith(null)`, "trimStart", 1, 5, "file.js"),
		invalidTrimCase(`foo.trim().endsWith(1n)`, "trimEnd", 1, 5, "file.js"),
		invalidTrimCase(`(foo).trim().startsWith("-")`, "trimStart", 1, 7, "file.js"),
		invalidTrimCase(`(foo.trim()).startsWith("-")`, "trimStart", 1, 6, "file.js"),
		invalidTrimCase("if (foo.trim().startsWith(\"-\")) {\n\tbar();\n}", "trimStart", 1, 9, "file.js"),
		invalidTrimCase("foo\n\t// comment\n\t.trim/* comment */()\n\t.startsWith(\"-\")", "trimStart", 3, 3, "file.js"),
		invalidTrimCase(`function foo(value: string) { value.trim().endsWith("-"); }`, "trimEnd", 1, 37, "file.ts"),
		invalidTrimCase(`function foo(value: string) { value.trim().startsWith("-" as const); }`, "trimStart", 1, 37, "file.ts"),
		invalidTrimCase(`function foo(value: string) { value.trim().endsWith("-" satisfies string); }`, "trimEnd", 1, 37, "file.ts"),
		invalidTrimCase(`import {string as z} from "zod"; z.string().trim().startsWith("-")`, "trimStart", 1, 45, "file.js"),
		invalidTrimCase(`z.string().trim().startsWith("-")`, "trimStart", 1, 12, "file.js"),

		// Examples from the rule documentation.
		invalidTrimCase(`const result = value.trim().startsWith('-');`, "trimStart", 1, 22, "file.js"),
		invalidTrimCase(`const result = value.trim().endsWith('-');`, "trimEnd", 1, 22, "file.js"),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unnecessary_string_trim.NoUnnecessaryStringTrimRule, valid, invalid)
}

func invalidTrimCase(code, replacement string, line, column int, fileName string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: fileName,
		Output:   []string{strings.Replace(code, ".trim", "."+replacement, 1)},
		Errors:   []rule_tester.InvalidTestCaseError{trimError(replacement, line, column, column+4)},
	}
}

func trimError(replacement string, line, column, endColumn int) rule_tester.InvalidTestCaseError {
	method := "startsWith"
	if replacement == "trimEnd" {
		method = "endsWith"
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: "no-unnecessary-string-trim",
		Message:   fmt.Sprintf("Prefer `String#%s()` before `String#%s()`.", replacement, method),
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}
