import { lint } from '@rslint/core/internal';
import { describe, test, expect } from 'rstack/test';
import path from 'node:path';
import { RemoteSourceFile } from '@rslint/api';

// rslint's JavaScript API takes a config OBJECT (Go never reads config from disk, and
// there is no separate ruleOptions surface). A single-rule config keeps each
// test scoped to the rule under test.
const cfg = (project, rule) => [
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { project: [project] } },
    rules: { [rule]: 'error' },
    plugins: ['@typescript-eslint'],
  },
];

describe('lint api', async (t) => {
  let cwd = path.resolve(import.meta.dirname, '../fixtures');
  test('virtual file support', async (t) => {
    let config = cfg(
      './tsconfig.virtual.json',
      '@typescript-eslint/no-unsafe-member-access',
    );
    let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
    // Use virtual file contents instead of reading from disk
    const diags = await lint({
      config,
      configDirectory: cwd,
      workingDirectory: cwd,
      fileContents: {
        [virtual_entry]: `
                    let a:any = 10;
                    a.b =10;
                `,
      },
    });

    expect(diags).toMatchSnapshot();
  });
  test('diag snapshot', async (t) => {
    let config = cfg(
      './tsconfig.json',
      '@typescript-eslint/no-unsafe-member-access',
    );
    const diags = await lint({
      config,
      configDirectory: cwd,
      workingDirectory: cwd,
    });
    expect(diags).toMatchSnapshot();
  });

  test('explicit files filter limits lint scope', async () => {
    const config = cfg(
      './tsconfig.json',
      '@typescript-eslint/no-unsafe-member-access',
    );
    const targetFile = path.resolve(cwd, 'src/index.ts');
    const diags = await lint({
      config,
      files: [targetFile],
      configDirectory: cwd,
      workingDirectory: cwd,
    });

    expect(diags.fileCount).toBe(1);
    expect(diags.diagnostics.length).toBe(2);
    expect(new Set(diags.diagnostics.map((diag) => diag.filePath))).toEqual(
      new Set(['src/index.ts']),
    );
  });

  test('uses project types for character comparisons', async () => {
    const ruleName = 'unicorn/no-invalid-character-comparison';
    const result = await lint({
      config: [
        {
          ...cfg('./tsconfig.virtual.json', ruleName)[0],
          plugins: ['unicorn'],
        },
      ],
      configDirectory: cwd,
      workingDirectory: cwd,
      fileContents: {
        [path.resolve(cwd, 'src/virtual.ts')]: `
export {};
declare const read: () => string;
declare const text: {value: string};
interface Read {(): string}
declare const callback: Read;
read().charAt(0) === "ab";
text.value.at(0) !== "ab";
callback()[0] === "ab";
`,
      },
    });
    expect(result.fileCount).toBe(1);
    expect(result.ruleCount).toBe(1);
    expect(
      result.diagnostics.map(({ messageId, message }) => ({
        messageId,
        message,
      })),
    ).toEqual(
      ['false', 'true', 'false'].map((value) => ({
        messageId: 'no-invalid-character-comparison',
        message: `A single character can never equal the multi-character string \`ab\`, so this comparison is always ${value}.`,
      })),
    );
  });

  test('preserves surrogate code units in diagnostic messages', async () => {
    const cases = [
      {
        code: String.raw`"x"[0] === "\uD800a";`,
        value: '\ud800a',
        result: 'false',
      },
      {
        code: String.raw`"x"[0] === "a\uDFFF";`,
        value: 'a\udfff',
        result: 'false',
      },
      {
        code: String.raw`"x"[0] !== "\uD800\uDC00";`,
        value: '\ud800\udc00',
        result: 'true',
      },
    ];
    const ruleName = 'unicorn/no-invalid-character-comparison';
    const result = await lint({
      config: [{ plugins: ['unicorn'], rules: { [ruleName]: 'error' } }],
      configDirectory: cwd,
      workingDirectory: cwd,
      fileContents: {
        [path.resolve(cwd, 'src/virtual.ts')]: cases
          .map(({ code }) => code)
          .join('\n'),
      },
    });
    expect(result.diagnostics).toHaveLength(cases.length);
    for (const [
      index,
      { code, value, result: comparisonResult },
    ] of cases.entries()) {
      expect(result.diagnostics[index]).toMatchObject({
        ruleName,
        messageId: 'no-invalid-character-comparison',
        message: `A single character can never equal the multi-character string \`${value}\`, so this comparison is always ${comparisonResult}.`,
        range: {
          start: { line: index + 1, column: 1 },
          end: { line: index + 1, column: code.length },
        },
      });
    }
  });
});

describe('lint fix:true', async (t) => {
  let cwd = path.resolve(import.meta.dirname, '../fixtures');

  test('fix:true returns fixed output in-band', async (t) => {
    let config = cfg(
      './tsconfig.virtual.json',
      '@typescript-eslint/array-type',
    );
    let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
    const fileContent = `let a: Array<string> = [];`;

    const diags = await lint({
      config,
      configDirectory: cwd,
      workingDirectory: cwd,
      fileContents: {
        [virtual_entry]: fileContent,
      },
      fix: true,
      includeEncodedSourceFiles: true,
    });

    // fix:true applies fixes in-band and returns the fixed source per file in
    // `output` (array-type rewrites `Array<string>` → `string[]`); the fix is
    // not written to disk.
    expect(diags.output).toBeDefined();
    expect(diags.diagnostics).toEqual([]);
    expect(diags.errorCount).toBe(0);
    expect(diags.fixableErrorCount).toBe(0);
    const encoded = diags.encodedSourceFiles['src/virtual.ts'];
    const buffer = Uint8Array.from(atob(encoded), (c) => c.charCodeAt(0));
    expect(new RemoteSourceFile(buffer, new TextDecoder()).text).toBe(
      'let a: string[] = [];',
    );
    expect({
      input: fileContent,
      output: diags.output,
    }).toMatchSnapshot();
  });
});

describe('encoded source files', async (t) => {
  let cwd = path.resolve(import.meta.dirname, '../fixtures');
  let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
  const fileContent = `let x: string = "hello";const y = x as string;`;
  test('encoded source files', async (t) => {
    const diags = await lint({
      config: cfg(
        './tsconfig.virtual.json',
        '@typescript-eslint/no-unnecessary-type-assertion',
      ),
      configDirectory: cwd,
      workingDirectory: cwd,
      includeEncodedSourceFiles: true,
      fileContents: {
        [virtual_entry]: fileContent,
      },
    });
    const content = diags.encodedSourceFiles['src/virtual.ts'];
    // decode content from base64 to uint8array
    const buffer = Uint8Array.from(atob(content), (c) => c.charCodeAt(0));

    const sourceFile = new RemoteSourceFile(buffer, new TextDecoder());

    const source = sourceFile.text;
    expect(source).toBe(fileContent);
  });
});
