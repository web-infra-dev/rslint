// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-array-index-of.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/shared/simple-array-search-rule-tests.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-array-index-of.md
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { runInNewContext } from 'node:vm';

import { lint, type Diagnostic } from '@rslint/core/internal';

const ruleName = 'unicorn/prefer-array-index-of';
const searches = [
  ['findIndex', 'indexOf'],
  ['findLastIndex', 'lastIndexOf'],
] as const;

interface SearchCase {
  code: string;
  extension?: 'js' | 'ts';
  reports?: boolean;
  fixed?: string;
  suggested?: string;
  runtime?: 'unchanged' | { before: string; suggested: string };
}

const syntaxCases: SearchCase[] = [
  // Member and call shapes.
  { code: 'const FIND = foo.FIND' },
  { code: 'foo.FIND()' },
  { code: 'foo.FIND(function (x) {return x === 1;}, bar)' },
  { code: 'foo.FIND(...[function (x) {return x === 1;}])' },
  { code: 'new foo.FIND(x => x === 1)' },
  { code: 'FIND(x => x === 1)' },
  { code: 'foo["FIND"](x => x === 1)' },
  { code: 'foo[FIND](x => x === 1)' },
  { code: 'foo.not_FIND(x => x === 1)' },
  // Callback kind and parameters.
  { code: 'foo.FIND(myFunction)' },
  { code: 'foo.FIND((x, i) => x === i)' },
  { code: 'foo.FIND((x, i) => {return x === i})' },
  { code: 'foo.FIND(function(x, i) {return x === i})' },
  { code: 'foo.FIND(({x}) => x === 1)' },
  { code: 'foo.FIND(({x}) => {return x === 1})' },
  { code: 'foo.FIND(function({x}) {return x === 1})' },
  { code: 'foo.FIND(function * (x) {return x === 1})' },
  { code: 'foo.FIND(async (x) => x === 1)' },
  { code: 'foo.FIND(async (x) => {return x === 1})' },
  { code: 'foo.FIND(async function(x) {return x === 1})' },
  // Callback body.
  { code: 'foo.FIND(x => {return;})' },
  { code: 'foo.FIND(function (x) {return;})' },
  { code: 'foo.FIND((x = 1) => {return;})' },
  { code: 'foo.FIND(() => {return;})' },
  { code: 'foo.FIND(({x}) => {noop();return x === 1})' },
  { code: 'foo.FIND(({x}) => {bar(x === 1)})' },
  // Comparison and references.
  { code: 'foo.FIND(x => x - 1)' },
  { code: 'foo.FIND(x => {return x - 1})' },
  { code: 'foo.FIND(function (x){return x - 1})' },
  { code: 'foo.FIND(x => x !== 1)' },
  { code: 'foo.FIND(x => {return x !== 1})' },
  { code: 'foo.FIND(function (x){return x !== 1})' },
  { code: 'foo.FIND(x => 1 === 1.0)' },
  { code: 'foo.FIND(x => {return 1 === 1.0})' },
  { code: 'foo.FIND(function (x){return 1 === 1.0})' },
  { code: 'foo.FIND(x => x === x)' },
  { code: 'foo.FIND(x => y === 1)' },
  { code: 'foo.FIND(x => x + "foo" === "foo" + x)' },
  { code: 'foo.FIND(x => x === "foo" + x)' },
  { code: 'foo.FIND(x => x === (function (){return x === "1"})())' },
  { code: 'foo.FIND(function fn(x) {return x === fn(y)})' },
  // Function scope and receiver type.
  { code: 'foo.FIND(function(x) {return x === arguments.length})' },
  { code: 'foo.FIND(function(x) {return x === this.length})' },
  { code: 'foo.INDEX(0)' },
  {
    code: 'function f(foo: Set<number>) { return foo.FIND(x => x === 1); }',
    extension: 'ts',
  },
  // Upstream snapshot reports and edits.
  {
    code: 'values.FIND(x => x === "foo")',
    reports: true,
    fixed: 'values.INDEX("foo")',
  },
  {
    code: 'values?.FIND(x => x === "foo")',
    reports: true,
    fixed: 'values?.INDEX("foo")',
  },
  {
    code: 'values.FIND(x => "foo" === x)',
    reports: true,
    fixed: 'values.INDEX("foo")',
  },
  {
    code: 'values.FIND(x => {return x === "foo";})',
    reports: true,
    fixed: 'values.INDEX("foo")',
  },
  {
    code: 'values.FIND(function (x) {return x === "foo";})',
    reports: true,
    fixed: 'values.INDEX("foo")',
  },
  {
    code: '// 1\n(0, values)\n\t// 2\n\t./* 3 */FIND /* 3 */ (\n\t\t/* 4 */\n\t\tx /* 5 */ => /* 6 */ x /* 7 */ === /* 8 */ "foo" /* 9 */\n\t) /* 10 */',
    reports: true,
  },
  {
    code: 'foo.FIND(function (element) {\n\treturn element === bar.FIND(x => x === 1);\n});',
    reports: true,
    fixed:
      'foo.FIND(function (element) {\n\treturn element === bar.INDEX(1);\n});',
    suggested: 'foo.INDEX(bar.FIND(x => x === 1));',
  },
  {
    code: 'values.FIND(x => x === (0, "foo"))',
    reports: true,
    fixed: 'values.INDEX((0, "foo"))',
  },
  {
    code: 'values.FIND((x => x === (0, "foo")))',
    reports: true,
    fixed: 'values.INDEX((0, "foo"))',
  },
  {
    code: 'function fn() {\n\tfoo.FIND(x => x === arguments.length)\n}',
    reports: true,
    suggested: 'function fn() {\n\tfoo.INDEX(arguments.length)\n}',
  },
  {
    code: 'function fn() {\n\tfoo.FIND(x => x === this[1])\n}',
    reports: true,
    suggested: 'function fn() {\n\tfoo.INDEX(this[1])\n}',
  },
  {
    code: 'values.FIND(x => x === foo())',
    reports: true,
    suggested: 'values.INDEX(foo())',
  },
  {
    code: 'foo.FIND(function a(x) {\n\treturn x === (function (a) {\n\t\treturn a(this) === arguments[1]\n\t}).call(thisObject, anotherFunctionNamedA, secondArgument)\n})',
    reports: true,
    suggested:
      'foo.INDEX((function (a) {\n\t\treturn a(this) === arguments[1]\n\t}).call(thisObject, anotherFunctionNamedA, secondArgument))',
  },
  {
    code: 'values.FIND(item => item === target);',
    reports: true,
    fixed: 'values.INDEX(target);',
  },
  {
    code: 'values.FIND(item => item === target.id);',
    reports: true,
    suggested: 'values.INDEX(target.id);',
  },
  {
    code: 'values.FIND(item => item === target?.id);',
    reports: true,
    suggested: 'values.INDEX(target?.id);',
  },
  {
    code: 'values.FIND(item => item === 1 + 2);',
    reports: true,
    fixed: 'values.INDEX(1 + 2);',
  },
  {
    code: 'values.FIND(item => item === -2n);',
    reports: true,
    fixed: 'values.INDEX(-2n);',
  },
  {
    code: 'values.FIND(item => item === +1n);',
    reports: true,
    fixed: 'values.INDEX(+1n);',
  },
  {
    code: 'values.FIND(item => item === undefined);',
    reports: true,
    fixed: 'values.INDEX(undefined);',
  },
  {
    code: 'values.FIND(item => item === /item/u);',
    reports: true,
    fixed: 'values.INDEX(/item/u);',
  },
  {
    code: 'values.FIND(item => item === `item`);',
    reports: true,
    fixed: 'values.INDEX(`item`);',
  },
  {
    code: 'const target = 2; values.FIND(item => item === target);',
    reports: true,
    fixed: 'const target = 2; values.INDEX(target);',
  },
  {
    code: 'values.FIND(item => item === /* keep */ target.id);',
    reports: true,
  },
  // TypeScript syntax without a project.
  {
    code: 'const values = ["small", "medium", "large"] as const; declare const input: string; values.FIND(value => value === input);',
    extension: 'ts',
    reports: true,
    fixed:
      'const values = ["small", "medium", "large"] as const; declare const input: string; values.INDEX(input);',
  },
  {
    code: 'values.FIND(item => item === (target as number));',
    extension: 'ts',
    reports: true,
    fixed: 'values.INDEX((target as number));',
  },
  {
    code: 'values.FIND(item => item === (target satisfies number));',
    extension: 'ts',
    reports: true,
    fixed: 'values.INDEX((target satisfies number));',
  },
  {
    code: 'values.FIND(item => item === target!);',
    extension: 'ts',
    reports: true,
    fixed: 'values.INDEX(target!);',
  },
  {
    code: 'values.FIND(item => item === <number>target);',
    extension: 'ts',
    reports: true,
    fixed: 'values.INDEX(<number>target);',
  },
  {
    code: 'values.FIND(item => item === (-2 as number));',
    extension: 'ts',
    reports: true,
    fixed: 'values.INDEX((-2 as number));',
  },
  {
    code: 'values.FIND(item => item === (2 satisfies number));',
    extension: 'ts',
    reports: true,
    fixed: 'values.INDEX((2 satisfies number));',
  },
  {
    code: 'values.FIND(item => item === 2!);',
    extension: 'ts',
    reports: true,
    fixed: 'values.INDEX(2!);',
  },
  {
    code: 'values.FIND(item => item === <number>2);',
    extension: 'ts',
    reports: true,
    fixed: 'values.INDEX(<number>2);',
  },
  {
    code: 'values.FIND(item => item === (getTarget() as number));',
    extension: 'ts',
    reports: true,
    suggested: 'values.INDEX((getTarget() as number));',
  },
  {
    code: 'function foo() {\n\treturn (bar as string[]).FIND(x => x === "foo");\n}',
    extension: 'ts',
    reports: true,
    fixed: 'function foo() {\n\treturn (bar as string[]).INDEX("foo");\n}',
  },
  // Documentation examples.
  { code: "const index = foo.INDEX('foo');" },
  { code: 'const index = foo.FIND(x => x == undefined);' },
  { code: "const index = foo.FIND(x => x !== 'foo');" },
  { code: 'const index = foo.FIND((x, index) => x === index);' },
  { code: "const index = foo.FIND(x => (x === 'foo') && isValid());" },
  { code: "const index = foo.FIND(x => y === 'foo');" },
  { code: "const index = foo.FIND(x => y.x === 'foo');" },
  {
    code: 'const index = foo.FIND(x => {\n\tconst bar = getBar();\n\treturn x === bar;\n});',
  },
  {
    code: "const index = foo.FIND(x => x === 'foo');",
    reports: true,
    fixed: "const index = foo.INDEX('foo');",
  },
  {
    code: "const index = foo.FIND(x => 'foo' === x);",
    reports: true,
    fixed: "const index = foo.INDEX('foo');",
  },
  {
    code: "const index = foo.FIND(x => {\n\treturn x === 'foo';\n});",
    reports: true,
    fixed: "const index = foo.INDEX('foo');",
  },
  // Both index searches use strict equality with NaN.
  {
    code: 'foo.FIND(element => element === NaN);',
    reports: true,
    fixed: 'foo.INDEX(NaN);',
  },
];

const typedCases: SearchCase[] = (
  [
    // wider string
    [
      'const values = ["small", "medium", "large"] as const; declare const input: string; values.SEARCH;',
      'value => value === input',
    ],
    // wider object
    [
      'declare const values: readonly {id: string; label: string}[]; declare const input: {id: string}; values.SEARCH;',
      'value => value === input',
    ],
    // reversed comparison
    [
      'const values = ["small", "medium", "large"] as const; declare const input: string; values.SEARCH;',
      'value => input === value',
    ],
    // generic unknown
    [
      'function check<Type>(value: unknown, values: readonly Type[]): number { return values.SEARCH; }',
      'option => option === value',
    ],
    // constrained literal receiver
    [
      'function check<Type extends readonly ["small", "large"]>(values: Type, input: string) { return values.SEARCH; }',
      'value => value === input',
    ],
    // optional value with block body
    [
      'declare const values: string[]; declare const input: string | undefined; values.SEARCH;',
      'value => {return value === input;}',
    ],
    // incompatible side effect
    [
      'const values = ["small", "medium", "large"] as const; declare function getInput(): string; values.SEARCH;',
      'value => value === getInput()',
    ],
    // widened callback parameter
    [
      'const values = ["small", "medium", "large"] as const; declare const input: string; values.SEARCH;',
      '(value: string) => value === input',
    ],
    // incompatible union receiver
    [
      'declare const values: string[] | number[]; declare const input: string; values.SEARCH;',
      'value => value === input',
    ],
    // incompatible optional receiver
    [
      'declare const values: readonly ["small"] | undefined; declare const input: string; values?.SEARCH;',
      'value => value === input',
    ],
    // compatible string
    [
      'declare const values: string[]; declare const input: string; values.SEARCH;',
      'value => value === input',
      'input',
    ],
    // compatible object with extra property
    [
      'declare const values: readonly {id: string; label: string}[]; declare const input: {id: string; label: string; extra: boolean}; values.SEARCH;',
      'value => value === input',
      'input',
    ],
    // compatible side effect
    [
      'declare const values: string[]; declare function getInput(): string; values.SEARCH;',
      'value => value === getInput()',
      'getInput()',
      'ts',
      true,
    ],
    // compatible literal
    [
      'const values = ["small", "medium", "large"] as const; values.SEARCH;',
      'value => value === "small"',
      '"small"',
    ],
    // compatible generic
    [
      'function check<Type>(input: Type, values: readonly Type[]) { return values.SEARCH; }',
      'value => value === input',
      'input',
    ],
    // compatible constrained receiver
    [
      'function check<Type extends readonly string[]>(values: Type, input: string) { return values.SEARCH; }',
      'value => value === input',
      'input',
    ],
    // compatible optional value
    [
      'declare const values: (string | undefined)[]; declare const input: string | undefined; values.SEARCH;',
      'value => value === input',
      'input',
    ],
    // compatible union receiver
    [
      'declare const values: string[] | (string | number)[]; declare const input: string; values.SEARCH;',
      'value => value === input',
      'input',
    ],
    // compatible optional receiver
    [
      'declare const values: string[] | undefined; declare const input: string; values?.SEARCH;',
      'value => value === input',
      'input',
    ],
    // compatible typed array
    [
      'declare const values: Uint8Array; declare const input: number; values.SEARCH;',
      'value => value === input',
      'input',
    ],
    // BigInt typed array with unknown search value
    [
      'declare const values: BigInt64Array; declare const input: unknown; values.SEARCH;',
      'value => value === input',
    ],
    // compatible BigInt typed array
    [
      'declare const values: BigInt64Array; declare const input: bigint; values.SEARCH;',
      'value => value === input',
      'input',
    ],
    // JSDoc wider string
    [
      '/** @type {readonly ("small" | "large")[]} */\nconst values = ["small", "large"];\n/** @param {string} input */\nfunction check(input) { return values.SEARCH; }',
      'value => value === input',
      undefined,
      'js',
    ],
    // JSDoc compatible string
    [
      '/** @type {string[]} */\nconst values = ["small", "large"];\n/** @param {string} input */\nfunction check(input) { return values.SEARCH; }',
      'value => value === input',
      'input',
      'js',
    ],
    // unknown receiver
    [
      'declare const values: any; declare const input: string; values.SEARCH;',
      '(value: unknown) => value === input',
      'input',
    ],
    // narrowed value
    [
      'function check(values: string[], input: string | undefined) { if (typeof input !== "string") { return; } return values.SEARCH; }',
      'value => value === input',
      'input',
    ],
  ] satisfies [string, string, string?, ('js' | 'ts')?, boolean?][]
).map(([template, callback, value, extension = 'ts', sideEffect = false]) => {
  const code = template.replace('SEARCH', `FIND(${callback})`);
  if (value === undefined) return { code, extension };
  const output = template.replace('SEARCH', `INDEX(${value})`);
  return {
    code,
    extension,
    reports: true,
    ...(sideEffect ? { suggested: output } : { fixed: output }),
  };
});

// The native runner has no Vue template parser.
const vueTemplate =
  '<template><div v-if="values.FIND(x => x === \'foo\')"></div></template>';

function runtimeCases(method: string): SearchCase[] {
  const cases: SearchCase[] = [
    {
      code: 'let calls = 0; function getValue() { calls++; return calls + 1; } const index = [1, 2].FIND(item => item === getValue()); JSON.stringify({index, calls});',
      reports: true,
      suggested:
        'let calls = 0; function getValue() { calls++; return calls + 1; } const index = [1, 2].INDEX(getValue()); JSON.stringify({index, calls});',
      runtime: {
        before:
          method === 'findIndex'
            ? '{"index":-1,"calls":2}'
            : '{"index":1,"calls":1}',
        suggested: '{"index":1,"calls":1}',
      },
    },
    {
      code: 'let value = 1; const index = [].FIND(item => item === (value += 1)); JSON.stringify({index, value});',
      reports: true,
      suggested:
        'let value = 1; const index = [].INDEX((value += 1)); JSON.stringify({index, value});',
      runtime: {
        before: '{"index":-1,"value":1}',
        suggested: '{"index":-1,"value":2}',
      },
    },
  ];
  const memberReads = ['object.id', 'object?.id', 'object.missing'];
  for (const value of [
    '0',
    '-0',
    '+2',
    '-(-2)',
    '1e999',
    '-1e999',
    '"item"',
    'true',
    'false',
    'null',
    '2n',
    'undefined',
    '`item`',
    '/item/u',
    'target',
    ...memberReads,
  ]) {
    for (const values of [
      '[]',
      '[0, 2, "item", true, false, null, 2n, 1e999, undefined]',
    ]) {
      const base = `const target = 2; const object = {id: "item"}; ${values}`;
      const output = `${base}.INDEX(${value});`;
      cases.push({
        code: `${base}.FIND(item => item === ${value});`,
        reports: true,
        ...(memberReads.includes(value)
          ? { suggested: output }
          : { fixed: output }),
        runtime: 'unchanged',
      });
    }
  }
  return cases;
}

let root: string;
beforeAll(async () => {
  root = await fs.realpath(
    await fs.mkdtemp(path.join(tmpdir(), 'rslint-array-index-of-')),
  );
  await fs.writeFile(
    path.join(root, 'tsconfig.json'),
    JSON.stringify({
      compilerOptions: {
        strict: true,
        allowJs: true,
        checkJs: true,
        target: 'ESNext',
        moduleDetection: 'force',
        types: [],
        noEmit: true,
      },
    }),
  );
});
afterAll(async () => {
  if (root) await fs.rm(root, { recursive: true, force: true });
});

function applyEdits(
  code: string,
  fixes: NonNullable<Diagnostic['fixes']>,
): string {
  return [...fixes]
    .sort((a, b) => b.startPos - a.startPos)
    .reduce(
      (text, edit) =>
        text.slice(0, edit.startPos) + edit.text + text.slice(edit.endPos),
      code,
    );
}

// The Unicorn wrapper checks diagnostics but does not apply edits or compare
// ranges. Use the same IPC entry point directly for these fixtures.
async function verifyCases(
  method: string,
  replacement: string,
  cases: SearchCase[],
  typed = false,
) {
  const expand = (text: string) =>
    text.replaceAll('FIND', method).replaceAll('INDEX', replacement);
  const files = cases.map(({ extension = 'js' }, i) =>
    path.join(root, `case-${i}.${extension}`),
  );
  const request = {
    workingDirectory: root,
    configDirectory: root,
    config: [
      {
        plugins: ['unicorn'],
        rules: { [ruleName]: 'error' },
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: typed ? ['./tsconfig.json'] : false,
          },
        },
      },
    ],
    fileContents: Object.fromEntries(
      cases.map(({ code }, i) => [files[i], expand(code)]),
    ),
  };
  const result = await lint(request);
  expect(result.fileCount).toBe(cases.length);
  expect(result.ruleCount).toBe(1);
  const prefix = `prefer-${replacement}-over-${method}/`;
  const message = `Use \`.${replacement}()\` instead of \`.${method}()\` when looking for the index of an item.`;
  const suggestionMessage = `Replace \`.${method}()\` with \`.${replacement}()\`.`;
  const fixed = cases.some(({ runtime }) => runtime !== undefined)
    ? await lint({ ...request, fix: true })
    : undefined;
  const fixedFiles = new Map(
    Object.entries(fixed?.output ?? {}).map(([file, output]) => [
      path.resolve(root, file),
      output,
    ]),
  );

  for (const [i, item] of cases.entries()) {
    const file = files[i];
    const code = expand(item.code);
    const diagnostics = result.diagnostics.filter(
      (d) => path.resolve(root, d.filePath) === file,
    );
    const ranges = item.reports
      ? Array.from(item.code.matchAll(/FIND/g), ({ index }) => {
          const before = expand(item.code.slice(0, index));
          const line = before.split('\n').length;
          const column = before.length - before.lastIndexOf('\n');
          return {
            start: { line, column },
            end: { line, column: column + method.length },
          };
        })
      : [];
    expect(
      diagnostics.map((d) => ({
        ruleName: d.ruleName,
        messageId: d.messageId,
        message: d.message,
        range: d.range,
      })),
      code,
    ).toEqual(
      ranges.map((range) => ({
        ruleName,
        messageId: prefix + 'error',
        message,
        range,
      })),
    );

    const fixes = diagnostics.flatMap((d) => d.fixes ?? []);
    expect(applyEdits(code, fixes), code).toBe(
      item.fixed ? expand(item.fixed) : code,
    );
    const suggestions = diagnostics.flatMap((d) => d.suggestions ?? []);
    const suggested = suggestions.map((s) => ({
      messageId: s.messageId,
      message: s.message,
      output: applyEdits(code, s.fixes ?? []),
    }));
    expect(suggested, code).toEqual(
      item.suggested
        ? [
            {
              messageId: prefix + 'suggestion',
              message: suggestionMessage,
              output: expand(item.suggested),
            },
          ]
        : [],
    );

    if (item.runtime === 'unchanged') {
      let output = fixedFiles.get(file);
      if (item.suggested) {
        expect(output, code).toBeUndefined();
        output = suggested[0].output;
      } else {
        expect(output, code).toBe(expand(item.fixed!));
      }
      expect(runInNewContext(output!)).toBe(runInNewContext(code));
    } else if (item.runtime) {
      expect(fixedFiles.has(file), code).toBe(false);
      expect(runInNewContext(code)).toBe(item.runtime.before);
      expect(runInNewContext(suggested[0].output)).toBe(item.runtime.suggested);
    }
  }
}

for (const [method, replacement] of searches) {
  describe(`${ruleName}: ${method}`, () => {
    test('syntax, snapshots, and examples', () =>
      verifyCases(method, replacement, syntaxCases));
    test('replacement parameter types', () =>
      verifyCases(method, replacement, typedCases, true));
    test('runtime behavior of edits', () =>
      verifyCases(method, replacement, runtimeCases(method)));
    test.skip(`Vue parser unavailable: ${vueTemplate.replace('FIND', method)}`, () => {});
  });
}
