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
  test('preserves UTF-16 regexp messages across the API boundary', async () => {
    const filename = path.resolve(cwd, 'src/surrogate.js');
    const cases = [
      { pattern: '[\ud800-\udfff]', flags: 'u', reason: '' },
      { pattern: '[a-\ufffd]', flags: 'u', reason: '' },
      {
        pattern: '.',
        flags: '\ud800',
        message: "Invalid flags supplied to RegExp constructor '\ud800'.",
      },
      {
        pattern: '[\udfff-a]',
        flags: 'u',
        reason: 'Range out of order in character class',
      },
      {
        pattern: '[\udfff-a]',
        flags: 'v',
        reason: 'Range out of order in character class',
      },
      {
        pattern: '[\ufffd-a]',
        flags: 'u',
        reason: 'Range out of order in character class',
      },
      {
        pattern: '(?<\ud800>a)',
        flags: 'u',
        reason: 'Invalid capture group name',
      },
    ];
    for (const { pattern, flags, reason, message } of cases) {
      const code = `new RegExp(${JSON.stringify(pattern)}, ${JSON.stringify(flags)});`;
      for (const fix of [false, true]) {
        const result = await lint({
          configDirectory: cwd,
          workingDirectory: cwd,
          config: [
            {
              languageOptions: { parserOptions: { projectService: false } },
              rules: { 'no-invalid-regexp': 'error' },
            },
          ],
          fileContents: { [filename]: code },
          fix,
        });
        expect(result.fileCount).toBe(1);
        expect(result.ruleCount).toBe(1);
        expect(result.errorCount).toBe(reason || message ? 1 : 0);
        expect(result.warningCount).toBe(0);
        expect(result.fixableErrorCount).toBe(0);
        expect(result.fixableWarningCount).toBe(0);
        expect(result.diagnostics).toEqual(
          reason || message
            ? [
                {
                  ruleName: 'no-invalid-regexp',
                  messageId: 'regexMessage',
                  message:
                    message ??
                    `Invalid regular expression: /${pattern}/${flags}: ${reason}.`,
                  filePath: 'src/surrogate.js',
                  severity: 'error',
                  range: {
                    start: { line: 1, column: 1 },
                    end: { line: 1, column: code.length },
                  },
                },
              ]
            : [],
        );
        expect(result.output ?? {}).toEqual({});
      }
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
