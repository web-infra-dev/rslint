// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
import path from 'node:path';
import { lint } from '@rslint/core/internal';

const ruleName = 'unicorn/prefer-dom-node-text-content';

// Use the IPC results directly to assert suggestion edits, which the Unicorn
// test wrapper does not compare.
async function lintSource(code: string, extension: 'js' | 'ts', fix = false) {
  const root = import.meta.dirname;
  const result = await lint({
    workingDirectory: root,
    configDirectory: root,
    config: [
      {
        plugins: ['unicorn'],
        languageOptions: { parserOptions: { projectService: false } },
        rules: { [ruleName]: 'error' },
      },
    ],
    fileContents: { [path.join(root, `input.${extension}`)]: code },
    fix,
  });
  expect(result.fileCount).toBe(1);
  expect(result.ruleCount).toBe(1);
  return result;
}

test.each([
  'innerText;',
  'node.textContent;',
  'node[innerText];',
  'innerText = true;',
  "node['innerText'];",
  'innerText.textContent',
  'const [innerText] = node;',
  '[innerText] = node;',
  'const {[innerText]: text} = node;',
  '({[innerText]: text} = node);',
  'const foo = {innerText}',
  'const foo = {innerText: text}',
  // Documentation examples.
  'const text = foo.textContent;',
  'const {textContent} = foo;',
  "foo.textContent = '🦄';",
])('accepts JS %s', async (code) => {
  expect((await lintSource(code, 'js')).diagnostics).toEqual([]);
});

test.each([
  'interface Value {innerText: string} declare const value: Value; value.innerText;',
  'interface Value {innerText: string} declare const value: Value; const {innerText} = value;',
  'interface Value {innerText: string} declare const value: Value; let innerText; ({innerText} = value);',
  'interface Value {innerText: string} function foo({innerText}: Value) { return innerText; }',
])('accepts source-only TS %s', async (code) => {
  expect((await lintSource(code, 'ts')).diagnostics).toEqual([]);
});

const rename = 'textContent';
const preserveBinding = 'textContent: innerText';
const invalid: [code: string, column: number, replacement: string][] = [
  ['node.innerText;', 6, rename],
  ['node?.innerText;', 7, rename],
  ["node.innerText = 'foo';", 6, rename],
  ['innerText.innerText;', 11, rename],
  ['const {innerText} = node;', 8, preserveBinding],
  ['const {innerText,} = node;', 8, preserveBinding],
  ['const {innerText: text} = node;', 8, rename],
  ['const {innerText = "default text"} = node;', 8, preserveBinding],
  ['const {innerText: text = "default text"} = node;', 8, rename],
  ['({innerText} = node);', 3, preserveBinding],
  ['({innerText: text} = node);', 3, rename],
  ['({innerText = "default text"} = node);', 3, preserveBinding],
  ['({innerText: text = "default text"} = node);', 3, rename],
  ['function foo({innerText}) {return innerText}', 15, preserveBinding],
  ['for (const [{innerText}] of elements);', 14, preserveBinding],
  // Documentation examples.
  ['const text = foo.innerText;', 18, rename],
  ['const {innerText} = foo;', 8, preserveBinding],
  ["foo.innerText = '🦄';", 5, rename],
];

test.each(invalid)(
  'suggests a replacement in %s',
  async (code, column, text) => {
    const result = await lintSource(code, 'js', true);
    const endColumn = column + 'innerText'.length;
    expect(result.output).toBeUndefined();
    expect(result.fixableErrorCount).toBe(0);
    expect(result.diagnostics).toHaveLength(1);
    const [diagnostic] = result.diagnostics;
    expect(diagnostic).toMatchObject({
      ruleName,
      messageId: 'error',
      message: 'Prefer `.textContent` over `.innerText`.',
      range: {
        start: { line: 1, column },
        end: { line: 1, column: endColumn },
      },
    });
    expect(diagnostic.fixes).toBeUndefined();
    expect(diagnostic.suggestions).toEqual([
      {
        messageId: 'suggestion',
        message: 'Switch to `.textContent`.',
        fixes: [{ startPos: column - 1, endPos: endColumn - 1, text }],
      },
    ]);
  },
);
