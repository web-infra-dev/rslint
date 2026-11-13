import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, buffer = false) => ({
  code,
  filename: 'file.mjs',
  options: buffer ? ['buffer'] : [],
});
const invalid = (code: string, buffer = false) => ({
  ...valid(code, buffer),
  errors: [
    {
      messageId: `consistent-json-file-read/${buffer ? 'buffer' : 'string'}`,
      message: `Prefer reading the JSON file as a ${buffer ? 'buffer' : 'string'}.`,
    },
  ],
});

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/consistent-json-file-read.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/consistent-json-file-read.md
ruleTester.run('consistent-json-file-read', null as never, {
  valid: [
    // Upstream tests with the default option.
    valid(`import {readFileSync} from "other";
JSON.parse(readFileSync(file));`),
    valid(`function readFileSync() {}
JSON.parse(readFileSync(file));`),
    valid(`import {readFile} from "node:fs/promises";
function parse(readFile) {
	return JSON.parse(readFile(file));
}`),
    valid(`import {existsSync} from "node:fs";
JSON.parse(existsSync(file));`),
    valid('JSON.parse(await fs.readFile(file, "utf8"));'),
    valid('JSON.parse(await fs.readFile(file, "utf8",));'),
    valid('JSON.parse(await fs.readFile(file, "UTF-8"));'),
    valid('JSON.parse(await fs.readFileSync(file, "utf8"));'),
    valid('JSON.parse(fs.readFileSync(file, "utf8"));'),
    valid(
      'const CHARSET = "UTF8"; JSON.parse(await fs.readFile(file, CHARSET));',
    ),
    valid(
      'const EIGHT = 8; JSON.parse(await fs.readFile(file, `utf${EIGHT}`));',
    ),
    valid('JSON.parse(await fs["readFile"](file, "utf8"));'),
    valid('JSON.parse(await fs.readFile(file, {encoding: "utf8"}));'),
    valid('JSON.parse(await fs.readFile(file, "gbk"));'),
    valid('JSON.parse(await fs.readFile(file, unknown));'),
    valid('JSON.parse(await fs.readFile(...file));'),
    valid('JSON.parse(await fs.readFile(file, 0));'),
    valid('JSON.parse(await fs.readFile(file, "utf8", extraArgument));'),
    valid('JSON.parse(await fs.readFile?.(file));'),
    valid('JSON.parse(await fs?.readFile(file));'),
    valid('JSON.parse(await fs.notReadFileMethod(file));'),
    valid('JSON.parse?.(await fs.readFile(file));'),
    valid('JSON?.parse(await fs.readFile(file));'),
    valid('window.JSON.parse(await fs.readFile(file));'),
    valid('JSON.stringify(await fs.readFile(file));'),
    valid('NOT_JSON.parse(await fs.readFile(file));'),
    valid('for (const string of []) JSON.parse(string);'),
    valid('JSON.parse(await fs.readFile(file), extraArgument);'),
    valid('JSON.parse(foo);'),
    valid('JSON.parse();'),
    valid('JSON.parse(await fs.readFile(file, {encoding: "not-utf8"}));'),
    valid('JSON.parse(await fs.readFile(file, {flag: "r"}));'),
    valid('JSON.parse(await fs.readFile(file, {encoding: null, flag: "r"}));'),
    valid(
      'JSON.parse(await fs.readFile(file, {encoding: "utf8", extraProperty: "utf8"}));',
    ),
    valid(
      'const options = {}; Object.defineProperty(options); JSON.parse(await fs.readFile(file, options));',
    ),
    valid(`const options = {encoding: null};
Object.defineProperty(options, 'encoding', {get() { return 'utf8'; }});
JSON.parse(await fs.readFile(file, options));`),
    valid(`const options = {encoding: null};
Object.defineProperties(options, {encoding: {get() { return 'utf8'; }}});
JSON.parse(await fs.readFile(file, options));`),
    valid(`const options = {encoding: null};
const alias = options;
Object.defineProperty(alias, 'encoding', {get() { return 'utf8'; }});
JSON.parse(await fs.readFile(file, options));`),
    valid(`const options = {encoding: null};
const alias = options;
const secondAlias = alias;
Object.defineProperties(secondAlias, {encoding: {get() { return 'utf8'; }}});
JSON.parse(await fs.readFile(file, options));`),
    valid(`const key = 'encoding';
const descriptorKey = 'get';
const options = {encoding: null};
Object.defineProperties(options, {[key]: {[descriptorKey]() { return 'utf8'; }}});
JSON.parse(await fs.readFile(file, options));`),
    valid(`const options = {encoding: null};
Object.defineProperty(options, 'encoding', {['get']() { return 'utf8'; }});
JSON.parse(await fs.readFile(file, options));`),
    valid(`const options = {encoding: null};
Object.defineProperty(options, 'encoding', {set() {}});
JSON.parse(await fs.readFile(file, options));`),
    valid('JSON.parse(await fs.readFile(file, {...encoding}));'),
    valid('JSON.parse(await fs.readFile(file, {encoding: unknown}));'),
    valid(
      'const readingOptions = {flag: "r"};JSON.parse(await fs.readFile(file, readingOptions));',
    ),
    valid(
      'const readingOptions = {encoding: null, flag: "r"};JSON.parse(await fs.readFile(file, readingOptions));',
    ),
    valid(
      'const encoding = "gbk";JSON.parse(await fs.readFile(file, {encoding: encoding}));',
    ),
    valid(
      'const readingOptions = {encoding: "utf8", extraProperty: undefined};JSON.parse(await fs.readFile(file, readingOptions));',
    ),
    valid(`const {string} = await fs.readFile(file);
JSON.parse(string);`),
    valid(`const string = fs.readFile(file, () => {});
JSON.parse(string);`),
    valid(`const abortController = new AbortController();
const {signal} = abortController;
const promise = fs.readFile(file, {encoding: "utf8", signal});
if (foo) {
	JSON.parse(await promise);
} else {
	abortController.abort();
}`),
    valid(`const string = await fs.readFile(file, "utf8");
console.log(string);
JSON.parse(string);`),
    valid(
      'const string = await fs.readFile(file);\nJSON.parse(`[${string}]`);',
    ),
    valid(`const foo = {};
foo.bar = await fs.readFile(file);
JSON.parse(foo.bar);`),
    valid(`const foo = await fs.readFile(file);
const bar = await foo;
console.log(baz);
const baz = await bar;
JSON.parse(baz);`),
    valid(`const foo = fs.readFile(file);
function fn1() {
	const foo = "{}";
	JSON.parse(foo);
}`),
    valid('JSON.parse(readFileSync(file));'),
    // Rslint skips options objects that are modified elsewhere, including after the read.
    valid(`const options = {encoding: null};
const buffer = fs.readFile(file, options);
Object.defineProperty(options, 'other', {set() {}});
JSON.parse(buffer);`),
    // Upstream tests with the buffer option.
    valid('JSON.parse(await fs.readFile(file));', true),
    valid('JSON.parse(await fs.readFileSync(file));', true),
    valid('JSON.parse(fs.readFileSync(file));', true),
    valid('JSON.parse(await fs.readFile(file, "buffer"));', true),
    valid('JSON.parse(await fs.readFile(file, "gbk"));', true),
    valid('JSON.parse(await fs.readFile(file, ));', true),
    valid('JSON.parse(await fs.readFile(file, null));', true),
    valid('JSON.parse(await fs.readFile(file, {encoding: null}));', true),
    valid('JSON.parse(await fs.readFile(file, unknown));', true),
    valid('JSON.parse(await fs.readFile(...file, "utf8"));', true),
    valid('JSON.parse(await fs.readFile(file, ..."utf8"));', true),
    valid('JSON.parse(await fs.readFile(file, 0));', true),
    valid('JSON.parse(await fs.readFile(file, "utf8", extraArgument));', true),
    valid('JSON.parse(await fs.readFile?.(file, "utf8"));', true),
    valid('JSON.parse(await fs?.readFile(file, "utf8"));', true),
    valid('JSON.parse(await fs.notReadFileMethod(file, "utf8"));', true),
    valid('JSON.parse?.(await fs.readFile(file, "utf8"));', true),
    valid('JSON?.parse(await fs.readFile(file, "utf8"));', true),
    valid('window.JSON.parse(await fs.readFile(file, "utf8"));', true),
    valid('JSON.stringify(await fs.readFile(file, "utf8"));', true),
    valid('NOT_JSON.parse(await fs.readFile(file, "utf8"));', true),
    valid('for (const string of []) JSON.parse(string);', true),
    valid('JSON.parse(await fs.readFile(file, "utf8"), extraArgument);', true),
    valid('JSON.parse(foo);', true),
    valid('JSON.parse();', true),
    valid('JSON.parse(await fs.readFile(file, {encoding: "not-utf8"}));', true),
    valid(
      'JSON.parse(await fs.readFile(file, {encoding: "utf8", extraProperty: "utf8"}));',
      true,
    ),
    valid('JSON.parse(await fs.readFile(file, {...encoding}));', true),
    valid('JSON.parse(await fs.readFile(file, {encoding: unknown}));', true),
    valid(
      'const readingOptions = {encoding: "utf8", flag: "r"};JSON.parse(await fs.readFile(file, readingOptions));',
      true,
    ),
    valid(
      'const encoding = "gbk";JSON.parse(await fs.readFile(file, {encoding: encoding}));',
      true,
    ),
    valid(
      'const readingOptions = {encoding: "utf8", extraProperty: undefined};JSON.parse(await fs.readFile(file, readingOptions));',
      true,
    ),
    valid(
      `const {string} = await fs.readFile(file, "utf8");
JSON.parse(string);`,
      true,
    ),
    valid(
      `const string = fs.readFile(file, () => {});
JSON.parse(string);`,
      true,
    ),
    valid(
      `const abortController = new AbortController();
const {signal} = abortController;
const promise = fs.readFile(file, {encoding: "utf8", signal});
if (foo) {
	JSON.parse(await promise);
} else {
	abortController.abort();
}`,
      true,
    ),
    valid(
      `const string = await fs.readFile(file, "utf8");
console.log(string);
JSON.parse(string);`,
      true,
    ),
    valid(
      'const string = await fs.readFile(file, "utf8");\nJSON.parse(`[${string}]`);',
      true,
    ),
    valid(
      `const foo = {};
foo.bar = await fs.readFile(file, "utf8");
JSON.parse(foo.bar);`,
      true,
    ),
    valid(
      `const foo = await fs.readFile(file, "utf8");
const bar = await foo;
console.log(baz);
const baz = await bar;
JSON.parse(baz);`,
      true,
    ),
    valid(
      `const foo = fs.readFile(file, "utf8");
function fn1() {
	const foo = "{}";
	JSON.parse(foo);
}`,
      true,
    ),
    valid(
      `const options = {encoding: 'utf8'};
Object.defineProperty(options, 'encoding', {...descriptor, get() { return 'utf8'; }});
JSON.parse(await fs.readFile(file, options));`,
      true,
    ),
    // Default option examples from the upstream documentation.
    valid(
      "const packageJson = JSON.parse(await fs.readFile('./package.json', 'utf8'));",
    ),
    valid(`const promise = fs.readFile('./package.json', {encoding: 'utf8', signal});
const packageJson = JSON.parse(await promise);`),
    // Buffer option examples from the upstream documentation.
    valid(
      "const packageJson = JSON.parse(await fs.readFile('./package.json'));",
      true,
    ),
  ],
  invalid: [
    // Upstream tests with the default option.
    invalid(`import {readFileSync} from "node:fs";
JSON.parse(readFileSync(file));`),
    invalid(`import {readFile} from "node:fs/promises";
JSON.parse(await readFile(file));`),
    invalid(`import {readFile} from "fs";
JSON.parse(await readFile(file));`),
    invalid(`import { readFileSync as read } from "node:fs";
JSON.parse(read(file));`),
    invalid('JSON.parse(await fs.readFile(file));'),
    invalid('JSON.parse(await fs.readFile(file,));'),
    invalid('JSON.parse(await fs.readFile((file)));'),
    invalid('JSON.parse(await fs.readFile(((file))));'),
    invalid('JSON.parse(await fs.readFile(file, null));'),
    invalid('JSON.parse(await fs.readFile(file, undefined));'),
    invalid('JSON.parse(await fs.readFile(file, {encoding: null}));'),
    invalid('JSON.parse(await fs.readFile(file, {encoding: undefined}));'),
    invalid(
      'const readingOptions = {encoding: null};JSON.parse(await fs.readFile(file, readingOptions));',
    ),
    invalid('JSON.parse(await fs.readFileSync(file));'),
    invalid('JSON.parse(fs.readFileSync(file));'),
    invalid(`const buffer = await fs.readFile(file);
JSON.parse(buffer);`),
    invalid(`let buffer = await fs.readFile(file);
JSON.parse(buffer);`),
    invalid(`const foo = await await fs.readFile(file);
const bar = await await foo;
const baz = await bar;
JSON.parse(baz);`),
    invalid(`var foo = await fs.readFile(file);
let bar = await foo;
const baz = await bar;
JSON.parse(baz);`),
    invalid(`const foo = fs.readFile(file);
async function fn1() {
	const bar = await foo;

	function fn2() {
		const baz = bar;
		JSON.parse(baz);
	}
}`),
    invalid(`const readingOptions = {encoding: null};
const buffer = fs.readFile(file, readingOptions);
function parse() {
	const readingOptions = {encoding: "utf8"};
	JSON.parse(buffer);
}`),
    invalid(`const buffer = fs.readFile(file); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file); /* Should NOT report */
	JSON.parse(await foo);
}`),
    invalid(`const buffer = fs.readFile(file); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file); /* Should NOT report */
	const baz = foo;
	for (;;) {
		const buffer = fs.readFile(file); /* Should NOT report */
		JSON.parse(await baz);
	}
}`),
    invalid(`const foo = fs.readFile(file);
function fn1() {
	JSON.parse(foo);

	function fn2() {
		const foo = "{}";
	}
}`),
    invalid(`const buffer = await NOT_A_FS_MODULE.readFile(file);
JSON.parse(buffer);`),
    invalid(
      'JSON.parse(await fs.readFile(file, {/* keep */ encoding: null}));',
    ),
    // Upstream tests with the buffer option.
    invalid('JSON.parse(await fs.readFile(file, "utf8"));', true),
    invalid('JSON.parse(await fs.readFile(file, "utf8",));', true),
    invalid('JSON.parse(await fs.readFile(file, "UTF-8"));', true),
    invalid('JSON.parse(await fs.readFileSync(file, "utf8"));', true),
    invalid('JSON.parse(fs.readFileSync(file, "utf8"));', true),
    invalid(
      'const CHARSET = "UTF8"; JSON.parse(await fs.readFile(file, CHARSET));',
      true,
    ),
    invalid(
      'const EIGHT = 8; JSON.parse(await fs.readFile(file, `utf${EIGHT}`));',
      true,
    ),
    invalid('JSON.parse(await fs["readFile"](file, "utf8"));', true),
    invalid('JSON.parse(await fs.readFile(file, {encoding: "utf8"}));', true),
    invalid(
      'const EIGHT = 8; JSON.parse(await fs.readFile(file, {encoding: `utf${EIGHT}`}));',
      true,
    ),
    invalid(
      'JSON.parse(await fs.readFile(file, {...({encoding: "utf8"})}));',
      true,
    ),
    invalid(
      'const encoding = "utf8";JSON.parse(await fs.readFile(file, {encoding}));',
      true,
    ),
    invalid(
      'const CHARSET = "utF-8", readingOptions = {encoding: CHARSET}; JSON.parse(await fs.readFile(file, readingOptions));',
      true,
    ),
    invalid(
      'const EIGHT = 8, ENCODING = "encoding"; JSON.parse(await fs.readFile(file, {[ENCODING]: `utf${EIGHT}`}));',
      true,
    ),
    invalid(
      `const string = await fs.readFile(file, "utf8");
JSON.parse(string);`,
      true,
    ),
    invalid(
      `const encoding = "utf8";
const string = fs.readFile(file, encoding);
function parse() {
	const encoding = undefined;
	JSON.parse(string);
}`,
      true,
    ),
    invalid(
      `let string = await fs.readFile(file, "utf8");
JSON.parse(string);`,
      true,
    ),
    invalid(
      `const foo = await await fs.readFile(file, "utf8");
const bar = await await foo;
const baz = await bar;
JSON.parse(baz);`,
      true,
    ),
    invalid(
      `var foo = await fs.readFile(file, "utf8");
let bar = await foo;
const baz = await bar;
JSON.parse(baz);`,
      true,
    ),
    invalid(
      `const foo = fs.readFile(file, "utf8");
async function fn1() {
	const bar = await foo;

	function fn2() {
		const baz = bar;
		JSON.parse(baz);
	}
}`,
      true,
    ),
    invalid(
      `const buffer = fs.readFile(file, "utf8"); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
	JSON.parse(await foo);
}`,
      true,
    ),
    invalid(
      `const buffer = fs.readFile(file, "utf8"); /* Should report */
const foo = buffer;
async function fn1() {
	const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
	const baz = foo;
	for (;;) {
		const buffer = fs.readFile(file, "utf8"); /* Should NOT report */
		JSON.parse(await baz);
	}
}`,
      true,
    ),
    invalid(
      `const foo = fs.readFile(file, "utf8");
function fn1() {
	JSON.parse(foo);

	function fn2() {
		const foo = "{}";
	}
}`,
      true,
    ),
    invalid(
      `const string = await NOT_A_FS_MODULE.readFile(file, "utf8");
JSON.parse(string);`,
      true,
    ),
    invalid(
      `import {readFileSync} from "node:fs";
JSON.parse(readFileSync(file, "utf8"));`,
      true,
    ),
    // Default option examples from the upstream documentation.
    invalid(
      "const packageJson = JSON.parse(await fs.readFile('./package.json'));",
    ),
    invalid(`const promise = fs.readFile('./package.json');
const packageJson = JSON.parse(await promise);`),
    // Buffer option examples from the upstream documentation.
    invalid(
      "const packageJson = JSON.parse(await fs.readFile('./package.json', 'utf8'));",
      true,
    ),
  ],
});
