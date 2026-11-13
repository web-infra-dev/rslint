package consistent_json_file_read_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_json_file_read"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/consistent-json-file-read.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/consistent-json-file-read.md
func TestConsistentJSONFileReadUpstream(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		runCases(t, upstreamStringValid, upstreamStringInvalid, false)
	})
	t.Run("buffer", func(t *testing.T) {
		runCases(t, upstreamBufferValid, upstreamBufferInvalid, true)
	})
	t.Run("documented string", func(t *testing.T) {
		runCases(t, documentedStringValid, documentedStringInvalid, false)
	})
	t.Run("documented buffer", func(t *testing.T) {
		runCases(t, documentedBufferValid, documentedBufferInvalid, true)
	})
}

func runCases(t *testing.T, valid []string, invalid []rule_tester.InvalidTestCase, buffer bool) {
	t.Helper()
	validCases := make([]rule_tester.ValidTestCase, len(valid))
	for i, code := range valid {
		validCases[i] = rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
		if buffer {
			validCases[i].Options = []any{"buffer"}
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_json_file_read.ConsistentJSONFileReadRule, validCases, invalid)
}

func readCase(code string, output []string, buffer bool, line, column, endLine, endColumn int) rule_tester.InvalidTestCase {
	messageID := "consistent-json-file-read/string"
	message := "Prefer reading the JSON file as a string."
	var options []any
	if buffer {
		options = []any{"buffer"}
		messageID = "consistent-json-file-read/buffer"
		message = "Prefer reading the JSON file as a buffer."
	}
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "file.mjs", Options: options, Output: output,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID, Message: message, Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
		}},
	}
}

// Upstream tests with the default option.
var upstreamStringValid = []string{
	`import {readFileSync} from "other";
JSON.parse(readFileSync(file));`,
	`function readFileSync() {}
JSON.parse(readFileSync(file));`,
	`import {readFile} from "node:fs/promises";
function parse(readFile) {
	return JSON.parse(readFile(file));
}`,
	`import {existsSync} from "node:fs";
JSON.parse(existsSync(file));`,
	"JSON.parse(await fs.readFile(file, \"utf8\"));",
	"JSON.parse(await fs.readFile(file, \"utf8\",));",
	"JSON.parse(await fs.readFile(file, \"UTF-8\"));",
	"JSON.parse(await fs.readFileSync(file, \"utf8\"));",
	"JSON.parse(fs.readFileSync(file, \"utf8\"));",
	"const CHARSET = \"UTF8\"; JSON.parse(await fs.readFile(file, CHARSET));",
	"const EIGHT = 8; JSON.parse(await fs.readFile(file, `utf${EIGHT}`));",
	"JSON.parse(await fs[\"readFile\"](file, \"utf8\"));",
	"JSON.parse(await fs.readFile(file, {encoding: \"utf8\"}));",
	"JSON.parse(await fs.readFile(file, \"gbk\"));",
	"JSON.parse(await fs.readFile(file, unknown));",
	"JSON.parse(await fs.readFile(...file));",
	"JSON.parse(await fs.readFile(file, 0));",
	"JSON.parse(await fs.readFile(file, \"utf8\", extraArgument));",
	"JSON.parse(await fs.readFile?.(file));",
	"JSON.parse(await fs?.readFile(file));",
	"JSON.parse(await fs.notReadFileMethod(file));",
	"JSON.parse?.(await fs.readFile(file));",
	"JSON?.parse(await fs.readFile(file));",
	"window.JSON.parse(await fs.readFile(file));",
	"JSON.stringify(await fs.readFile(file));",
	"NOT_JSON.parse(await fs.readFile(file));",
	"for (const string of []) JSON.parse(string);",
	"JSON.parse(await fs.readFile(file), extraArgument);",
	"JSON.parse(foo);",
	"JSON.parse();",
	"JSON.parse(await fs.readFile(file, {encoding: \"not-utf8\"}));",
	"JSON.parse(await fs.readFile(file, {flag: \"r\"}));",
	"JSON.parse(await fs.readFile(file, {encoding: null, flag: \"r\"}));",
	"JSON.parse(await fs.readFile(file, {encoding: \"utf8\", extraProperty: \"utf8\"}));",
	"const options = {}; Object.defineProperty(options); JSON.parse(await fs.readFile(file, options));",
	`const options = {encoding: null};
Object.defineProperty(options, 'encoding', {get() { return 'utf8'; }});
JSON.parse(await fs.readFile(file, options));`,
	`const options = {encoding: null};
Object.defineProperties(options, {encoding: {get() { return 'utf8'; }}});
JSON.parse(await fs.readFile(file, options));`,
	`const options = {encoding: null};
const alias = options;
Object.defineProperty(alias, 'encoding', {get() { return 'utf8'; }});
JSON.parse(await fs.readFile(file, options));`,
	`const options = {encoding: null};
const alias = options;
const secondAlias = alias;
Object.defineProperties(secondAlias, {encoding: {get() { return 'utf8'; }}});
JSON.parse(await fs.readFile(file, options));`,
	`const key = 'encoding';
const descriptorKey = 'get';
const options = {encoding: null};
Object.defineProperties(options, {[key]: {[descriptorKey]() { return 'utf8'; }}});
JSON.parse(await fs.readFile(file, options));`,
	`const options = {encoding: null};
Object.defineProperty(options, 'encoding', {['get']() { return 'utf8'; }});
JSON.parse(await fs.readFile(file, options));`,
	`const options = {encoding: null};
Object.defineProperty(options, 'encoding', {set() {}});
JSON.parse(await fs.readFile(file, options));`,
	"JSON.parse(await fs.readFile(file, {...encoding}));",
	"JSON.parse(await fs.readFile(file, {encoding: unknown}));",
	"const readingOptions = {flag: \"r\"};JSON.parse(await fs.readFile(file, readingOptions));",
	"const readingOptions = {encoding: null, flag: \"r\"};JSON.parse(await fs.readFile(file, readingOptions));",
	"const encoding = \"gbk\";JSON.parse(await fs.readFile(file, {encoding: encoding}));",
	"const readingOptions = {encoding: \"utf8\", extraProperty: undefined};JSON.parse(await fs.readFile(file, readingOptions));",
	`const {string} = await fs.readFile(file);
JSON.parse(string);`,
	`const string = fs.readFile(file, () => {});
JSON.parse(string);`,
	`const abortController = new AbortController();
const {signal} = abortController;
const promise = fs.readFile(file, {encoding: "utf8", signal});
if (foo) {
	JSON.parse(await promise);
} else {
	abortController.abort();
}`,
	`const string = await fs.readFile(file, "utf8");
console.log(string);
JSON.parse(string);`,
	"const string = await fs.readFile(file);\nJSON.parse(`[${string}]`);",
	`const foo = {};
foo.bar = await fs.readFile(file);
JSON.parse(foo.bar);`,
	`const foo = await fs.readFile(file);
const bar = await foo;
console.log(baz);
const baz = await bar;
JSON.parse(baz);`,
	`const foo = fs.readFile(file);
function fn1() {
	const foo = "{}";
	JSON.parse(foo);
}`,
	"JSON.parse(readFileSync(file));",
	// Rslint skips options objects that are modified elsewhere, including after the read.
	`const options = {encoding: null};
const buffer = fs.readFile(file, options);
Object.defineProperty(options, 'other', {set() {}});
JSON.parse(buffer);`,
}

var upstreamStringInvalid = []rule_tester.InvalidTestCase{
	readCase(`import {readFileSync} from "node:fs";
JSON.parse(readFileSync(file));`, []string{`import {readFileSync} from "node:fs";
JSON.parse(readFileSync(file, 'utf8'));`}, false, 2, 12, 2, 30),
	readCase(`import {readFile} from "node:fs/promises";
JSON.parse(await readFile(file));`, []string{`import {readFile} from "node:fs/promises";
JSON.parse(await readFile(file, 'utf8'));`}, false, 2, 18, 2, 32),
	readCase(`import {readFile} from "fs";
JSON.parse(await readFile(file));`, []string{`import {readFile} from "fs";
JSON.parse(await readFile(file, 'utf8'));`}, false, 2, 18, 2, 32),
	readCase(`import { readFileSync as read } from "node:fs";
JSON.parse(read(file));`, []string{`import { readFileSync as read } from "node:fs";
JSON.parse(read(file, 'utf8'));`}, false, 2, 12, 2, 22),
	readCase("JSON.parse(await fs.readFile(file));", []string{"JSON.parse(await fs.readFile(file, 'utf8'));"}, false, 1, 18, 1, 35),
	readCase("JSON.parse(await fs.readFile(file,));", []string{"JSON.parse(await fs.readFile(file, 'utf8'));"}, false, 1, 18, 1, 36),
	readCase("JSON.parse(await fs.readFile((file)));", []string{"JSON.parse(await fs.readFile((file), 'utf8'));"}, false, 1, 18, 1, 37),
	readCase("JSON.parse(await fs.readFile(((file))));", []string{"JSON.parse(await fs.readFile(((file)), 'utf8'));"}, false, 1, 18, 1, 39),
	readCase("JSON.parse(await fs.readFile(file, null));", []string{"JSON.parse(await fs.readFile(file, 'utf8'));"}, false, 1, 36, 1, 40),
	readCase("JSON.parse(await fs.readFile(file, undefined));", []string{"JSON.parse(await fs.readFile(file, 'utf8'));"}, false, 1, 36, 1, 45),
	readCase("JSON.parse(await fs.readFile(file, {encoding: null}));", []string{"JSON.parse(await fs.readFile(file, 'utf8'));"}, false, 1, 36, 1, 52),
	readCase("JSON.parse(await fs.readFile(file, {encoding: undefined}));", []string{"JSON.parse(await fs.readFile(file, 'utf8'));"}, false, 1, 36, 1, 57),
	readCase("const readingOptions = {encoding: null};JSON.parse(await fs.readFile(file, readingOptions));", []string{"const readingOptions = {encoding: null};JSON.parse(await fs.readFile(file, 'utf8'));"}, false, 1, 76, 1, 90),
	readCase("JSON.parse(await fs.readFileSync(file));", []string{"JSON.parse(await fs.readFileSync(file, 'utf8'));"}, false, 1, 18, 1, 39),
	readCase("JSON.parse(fs.readFileSync(file));", []string{"JSON.parse(fs.readFileSync(file, 'utf8'));"}, false, 1, 12, 1, 33),
	readCase(`const buffer = await fs.readFile(file);
JSON.parse(buffer);`, []string{`const buffer = await fs.readFile(file, 'utf8');
JSON.parse(buffer);`}, false, 1, 22, 1, 39),
	readCase(`let buffer = await fs.readFile(file);
JSON.parse(buffer);`, []string{`let buffer = await fs.readFile(file, 'utf8');
JSON.parse(buffer);`}, false, 1, 20, 1, 37),
	readCase(`const foo = await await fs.readFile(file);
const bar = await await foo;
const baz = await bar;
JSON.parse(baz);`, []string{`const foo = await await fs.readFile(file, 'utf8');
const bar = await await foo;
const baz = await bar;
JSON.parse(baz);`}, false, 1, 25, 1, 42),
	readCase(`var foo = await fs.readFile(file);
let bar = await foo;
const baz = await bar;
JSON.parse(baz);`, []string{`var foo = await fs.readFile(file, 'utf8');
let bar = await foo;
const baz = await bar;
JSON.parse(baz);`}, false, 1, 17, 1, 34),
	readCase(`const foo = fs.readFile(file);
async function fn1() {
	const bar = await foo;

	function fn2() {
		const baz = bar;
		JSON.parse(baz);
	}
}`, []string{`const foo = fs.readFile(file, 'utf8');
async function fn1() {
	const bar = await foo;

	function fn2() {
		const baz = bar;
		JSON.parse(baz);
	}
}`}, false, 1, 13, 1, 30),
	readCase(`const readingOptions = {encoding: null};
const buffer = fs.readFile(file, readingOptions);
function parse() {
	const readingOptions = {encoding: "utf8"};
	JSON.parse(buffer);
}`, []string{`const readingOptions = {encoding: null};
const buffer = fs.readFile(file, 'utf8');
function parse() {
	const readingOptions = {encoding: "utf8"};
	JSON.parse(buffer);
}`}, false, 2, 34, 2, 48),
	readCase(`const buffer = fs.readFile(file); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file); /* Should NOT report */
	JSON.parse(await foo);
}`, []string{`const buffer = fs.readFile(file, 'utf8'); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file); /* Should NOT report */
	JSON.parse(await foo);
}`}, false, 1, 16, 1, 33),
	readCase(`const buffer = fs.readFile(file); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file); /* Should NOT report */
	const baz = foo;
	for (;;) {
		const buffer = fs.readFile(file); /* Should NOT report */
		JSON.parse(await baz);
	}
}`, []string{`const buffer = fs.readFile(file, 'utf8'); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file); /* Should NOT report */
	const baz = foo;
	for (;;) {
		const buffer = fs.readFile(file); /* Should NOT report */
		JSON.parse(await baz);
	}
}`}, false, 1, 16, 1, 33),
	readCase(`const foo = fs.readFile(file);
function fn1() {
	JSON.parse(foo);

	function fn2() {
		const foo = "{}";
	}
}`, []string{`const foo = fs.readFile(file, 'utf8');
function fn1() {
	JSON.parse(foo);

	function fn2() {
		const foo = "{}";
	}
}`}, false, 1, 13, 1, 30),
	readCase(`const buffer = await NOT_A_FS_MODULE.readFile(file);
JSON.parse(buffer);`, []string{`const buffer = await NOT_A_FS_MODULE.readFile(file, 'utf8');
JSON.parse(buffer);`}, false, 1, 22, 1, 52),
	readCase("JSON.parse(await fs.readFile(file, {/* keep */ encoding: null}));", nil, false, 1, 36, 1, 63),
}

// Upstream tests with the buffer option.
var upstreamBufferValid = []string{
	"JSON.parse(await fs.readFile(file));",
	"JSON.parse(await fs.readFileSync(file));",
	"JSON.parse(fs.readFileSync(file));",
	"JSON.parse(await fs.readFile(file, \"buffer\"));",
	"JSON.parse(await fs.readFile(file, \"gbk\"));",
	"JSON.parse(await fs.readFile(file, ));",
	"JSON.parse(await fs.readFile(file, null));",
	"JSON.parse(await fs.readFile(file, {encoding: null}));",
	"JSON.parse(await fs.readFile(file, unknown));",
	"JSON.parse(await fs.readFile(...file, \"utf8\"));",
	"JSON.parse(await fs.readFile(file, ...\"utf8\"));",
	"JSON.parse(await fs.readFile(file, 0));",
	"JSON.parse(await fs.readFile(file, \"utf8\", extraArgument));",
	"JSON.parse(await fs.readFile?.(file, \"utf8\"));",
	"JSON.parse(await fs?.readFile(file, \"utf8\"));",
	"JSON.parse(await fs.notReadFileMethod(file, \"utf8\"));",
	"JSON.parse?.(await fs.readFile(file, \"utf8\"));",
	"JSON?.parse(await fs.readFile(file, \"utf8\"));",
	"window.JSON.parse(await fs.readFile(file, \"utf8\"));",
	"JSON.stringify(await fs.readFile(file, \"utf8\"));",
	"NOT_JSON.parse(await fs.readFile(file, \"utf8\"));",
	"for (const string of []) JSON.parse(string);",
	"JSON.parse(await fs.readFile(file, \"utf8\"), extraArgument);",
	"JSON.parse(foo);",
	"JSON.parse();",
	"JSON.parse(await fs.readFile(file, {encoding: \"not-utf8\"}));",
	"JSON.parse(await fs.readFile(file, {encoding: \"utf8\", extraProperty: \"utf8\"}));",
	"JSON.parse(await fs.readFile(file, {...encoding}));",
	"JSON.parse(await fs.readFile(file, {encoding: unknown}));",
	"const readingOptions = {encoding: \"utf8\", flag: \"r\"};JSON.parse(await fs.readFile(file, readingOptions));",
	"const encoding = \"gbk\";JSON.parse(await fs.readFile(file, {encoding: encoding}));",
	"const readingOptions = {encoding: \"utf8\", extraProperty: undefined};JSON.parse(await fs.readFile(file, readingOptions));",
	`const {string} = await fs.readFile(file, "utf8");
JSON.parse(string);`,
	`const string = fs.readFile(file, () => {});
JSON.parse(string);`,
	`const abortController = new AbortController();
const {signal} = abortController;
const promise = fs.readFile(file, {encoding: "utf8", signal});
if (foo) {
	JSON.parse(await promise);
} else {
	abortController.abort();
}`,
	`const string = await fs.readFile(file, "utf8");
console.log(string);
JSON.parse(string);`,
	"const string = await fs.readFile(file, \"utf8\");\nJSON.parse(`[${string}]`);",
	`const foo = {};
foo.bar = await fs.readFile(file, "utf8");
JSON.parse(foo.bar);`,
	`const foo = await fs.readFile(file, "utf8");
const bar = await foo;
console.log(baz);
const baz = await bar;
JSON.parse(baz);`,
	`const foo = fs.readFile(file, "utf8");
function fn1() {
	const foo = "{}";
	JSON.parse(foo);
}`,
	`const options = {encoding: 'utf8'};
Object.defineProperty(options, 'encoding', {...descriptor, get() { return 'utf8'; }});
JSON.parse(await fs.readFile(file, options));`,
}

var upstreamBufferInvalid = []rule_tester.InvalidTestCase{
	readCase("JSON.parse(await fs.readFile(file, \"utf8\"));", []string{"JSON.parse(await fs.readFile(file));"}, true, 1, 36, 1, 42),
	readCase("JSON.parse(await fs.readFile(file, \"utf8\",));", []string{"JSON.parse(await fs.readFile(file,));"}, true, 1, 36, 1, 42),
	readCase("JSON.parse(await fs.readFile(file, \"UTF-8\"));", []string{"JSON.parse(await fs.readFile(file));"}, true, 1, 36, 1, 43),
	readCase("JSON.parse(await fs.readFileSync(file, \"utf8\"));", []string{"JSON.parse(await fs.readFileSync(file));"}, true, 1, 40, 1, 46),
	readCase("JSON.parse(fs.readFileSync(file, \"utf8\"));", []string{"JSON.parse(fs.readFileSync(file));"}, true, 1, 34, 1, 40),
	readCase("const CHARSET = \"UTF8\"; JSON.parse(await fs.readFile(file, CHARSET));", []string{"const CHARSET = \"UTF8\"; JSON.parse(await fs.readFile(file));"}, true, 1, 60, 1, 67),
	readCase("const EIGHT = 8; JSON.parse(await fs.readFile(file, `utf${EIGHT}`));", []string{"const EIGHT = 8; JSON.parse(await fs.readFile(file));"}, true, 1, 53, 1, 66),
	readCase("JSON.parse(await fs[\"readFile\"](file, \"utf8\"));", []string{"JSON.parse(await fs[\"readFile\"](file));"}, true, 1, 39, 1, 45),
	readCase("JSON.parse(await fs.readFile(file, {encoding: \"utf8\"}));", []string{"JSON.parse(await fs.readFile(file));"}, true, 1, 36, 1, 54),
	readCase("const EIGHT = 8; JSON.parse(await fs.readFile(file, {encoding: `utf${EIGHT}`}));", []string{"const EIGHT = 8; JSON.parse(await fs.readFile(file));"}, true, 1, 53, 1, 78),
	readCase("JSON.parse(await fs.readFile(file, {...({encoding: \"utf8\"})}));", []string{"JSON.parse(await fs.readFile(file));"}, true, 1, 36, 1, 61),
	readCase("const encoding = \"utf8\";JSON.parse(await fs.readFile(file, {encoding}));", []string{"const encoding = \"utf8\";JSON.parse(await fs.readFile(file));"}, true, 1, 60, 1, 70),
	readCase("const CHARSET = \"utF-8\", readingOptions = {encoding: CHARSET}; JSON.parse(await fs.readFile(file, readingOptions));", []string{"const CHARSET = \"utF-8\", readingOptions = {encoding: CHARSET}; JSON.parse(await fs.readFile(file));"}, true, 1, 99, 1, 113),
	readCase("const EIGHT = 8, ENCODING = \"encoding\"; JSON.parse(await fs.readFile(file, {[ENCODING]: `utf${EIGHT}`}));", []string{"const EIGHT = 8, ENCODING = \"encoding\"; JSON.parse(await fs.readFile(file));"}, true, 1, 76, 1, 103),
	readCase(`const string = await fs.readFile(file, "utf8");
JSON.parse(string);`, []string{`const string = await fs.readFile(file);
JSON.parse(string);`}, true, 1, 40, 1, 46),
	readCase(`const encoding = "utf8";
const string = fs.readFile(file, encoding);
function parse() {
	const encoding = undefined;
	JSON.parse(string);
}`, []string{`const encoding = "utf8";
const string = fs.readFile(file);
function parse() {
	const encoding = undefined;
	JSON.parse(string);
}`}, true, 2, 34, 2, 42),
	readCase(`let string = await fs.readFile(file, "utf8");
JSON.parse(string);`, []string{`let string = await fs.readFile(file);
JSON.parse(string);`}, true, 1, 38, 1, 44),
	readCase(`const foo = await await fs.readFile(file, "utf8");
const bar = await await foo;
const baz = await bar;
JSON.parse(baz);`, []string{`const foo = await await fs.readFile(file);
const bar = await await foo;
const baz = await bar;
JSON.parse(baz);`}, true, 1, 43, 1, 49),
	readCase(`var foo = await fs.readFile(file, "utf8");
let bar = await foo;
const baz = await bar;
JSON.parse(baz);`, []string{`var foo = await fs.readFile(file);
let bar = await foo;
const baz = await bar;
JSON.parse(baz);`}, true, 1, 35, 1, 41),
	readCase(`const foo = fs.readFile(file, "utf8");
async function fn1() {
	const bar = await foo;

	function fn2() {
		const baz = bar;
		JSON.parse(baz);
	}
}`, []string{`const foo = fs.readFile(file);
async function fn1() {
	const bar = await foo;

	function fn2() {
		const baz = bar;
		JSON.parse(baz);
	}
}`}, true, 1, 31, 1, 37),
	readCase(`const buffer = fs.readFile(file, "utf8"); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
	JSON.parse(await foo);
}`, []string{`const buffer = fs.readFile(file); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
	JSON.parse(await foo);
}`}, true, 1, 34, 1, 40),
	readCase(`const buffer = fs.readFile(file, "utf8"); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
	const baz = foo;
	for (;;) {
		const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
		JSON.parse(await baz);
	}
}`, []string{`const buffer = fs.readFile(file); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
	const baz = foo;
	for (;;) {
		const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
		JSON.parse(await baz);
	}
}`}, true, 1, 34, 1, 40),
	readCase(`const foo = fs.readFile(file, "utf8");
function fn1() {
	JSON.parse(foo);

	function fn2() {
		const foo = "{}";
	}
}`, []string{`const foo = fs.readFile(file);
function fn1() {
	JSON.parse(foo);

	function fn2() {
		const foo = "{}";
	}
}`}, true, 1, 31, 1, 37),
	readCase(`const string = await NOT_A_FS_MODULE.readFile(file, "utf8");
JSON.parse(string);`, []string{`const string = await NOT_A_FS_MODULE.readFile(file);
JSON.parse(string);`}, true, 1, 53, 1, 59),
	readCase(`import {readFileSync} from "node:fs";
JSON.parse(readFileSync(file, "utf8"));`, []string{`import {readFileSync} from "node:fs";
JSON.parse(readFileSync(file));`}, true, 2, 31, 2, 37),
}

// Default option examples from the upstream documentation.
var documentedStringValid = []string{
	"const packageJson = JSON.parse(await fs.readFile('./package.json', 'utf8'));",
	`const promise = fs.readFile('./package.json', {encoding: 'utf8', signal});
const packageJson = JSON.parse(await promise);`,
}

var documentedStringInvalid = []rule_tester.InvalidTestCase{
	readCase("const packageJson = JSON.parse(await fs.readFile('./package.json'));", []string{"const packageJson = JSON.parse(await fs.readFile('./package.json', 'utf8'));"}, false, 1, 38, 1, 67),
	readCase(`const promise = fs.readFile('./package.json');
const packageJson = JSON.parse(await promise);`, []string{`const promise = fs.readFile('./package.json', 'utf8');
const packageJson = JSON.parse(await promise);`}, false, 1, 17, 1, 46),
}

// Buffer option examples from the upstream documentation.
var documentedBufferValid = []string{
	"const packageJson = JSON.parse(await fs.readFile('./package.json'));",
}

var documentedBufferInvalid = []rule_tester.InvalidTestCase{
	readCase("const packageJson = JSON.parse(await fs.readFile('./package.json', 'utf8'));", []string{"const packageJson = JSON.parse(await fs.readFile('./package.json'));"}, true, 1, 68, 1, 74),
}
