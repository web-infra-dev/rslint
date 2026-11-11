import { describe, test, expect } from 'rstack/test';
import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { runCliProcess, type CliProcessResult } from './spawn-cli.js';

const RSLINT_BIN = require.resolve('@rslint/core/bin');

async function runRslint(
  args: string[],
  cwd?: string,
): Promise<CliProcessResult> {
  return runCliProcess(process.execPath, [RSLINT_BIN, ...args], {
    cwd: cwd || process.cwd(),
  });
}

async function createTempDir(files: Record<string, string>): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-disable-test-'));

  for (const [filePath, content] of Object.entries(files)) {
    const fullPath = path.join(tempDir, filePath);
    await fs.mkdir(path.dirname(fullPath), { recursive: true });
    await fs.writeFile(fullPath, content, 'utf8');
  }

  return tempDir;
}

async function cleanupTempDir(tempDir: string): Promise<void> {
  await fs.rm(tempDir, { recursive: true, force: true });
}

function makeConfig(rules: Record<string, unknown>) {
  return `export default [${JSON.stringify({
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules,
    plugins: ['@typescript-eslint'],
  })}];`;
}

const TSCONFIG = JSON.stringify({
  compilerOptions: {
    target: 'ES2020',
    module: 'ESNext',
    strict: true,
  },
  include: ['**/*.ts'],
});

const RULE = '@typescript-eslint/no-explicit-any';

describe('rslint-disable comment directives', () => {
  // -------------------------------------------------------------------
  // rslint-disable-next-line
  // -------------------------------------------------------------------

  test('rslint-disable-next-line suppresses diagnostic on next line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// rslint-disable-next-line @typescript-eslint/no-explicit-any',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // rslint-disable-line
  // -------------------------------------------------------------------

  test('rslint-disable-line suppresses diagnostic on same line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        'const x: any = 1; // rslint-disable-line @typescript-eslint/no-explicit-any',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // rslint-disable + rslint-enable (block range)
  // -------------------------------------------------------------------

  test('rslint-disable + rslint-enable suppresses diagnostics inside range', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        'const b: any = 2;',
        '/* rslint-enable @typescript-eslint/no-explicit-any */',
        'export { a, b };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // rslint-disable without enable (rest of file)
  // -------------------------------------------------------------------

  test('rslint-disable without enable suppresses rest of file', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        'const b: any = 2;',
        'export { a, b };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // wildcard (no rule name)
  // -------------------------------------------------------------------

  test('rslint-disable without rule name suppresses all rules', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable */',
        'const a: any = 1;',
        '/* rslint-enable */',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // mixed prefixes: rslint-disable + eslint-enable
  // -------------------------------------------------------------------

  test('rslint-disable + eslint-enable mixed prefixes work', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
        'export { a };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-disable + rslint-enable mixed prefixes work', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        '/* rslint-enable @typescript-eslint/no-explicit-any */',
        'export { a };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // negative tests
  // -------------------------------------------------------------------

  test('rslint-disable-next-line does not suppress other lines', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// rslint-disable-next-line @typescript-eslint/no-explicit-any',
        'const x: any = 1;',
        'const y: any = 2;', // not suppressed
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('rslint-disable + rslint-enable does NOT suppress diagnostics outside range', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        '/* rslint-enable @typescript-eslint/no-explicit-any */',
        'const b: any = 2;', // outside range — should trigger
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // block comment style for line directives
  // -------------------------------------------------------------------

  test('block comment style works for rslint-disable-next-line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable-next-line @typescript-eslint/no-explicit-any */',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // rslint-disable-next-line with -- description
  // -------------------------------------------------------------------

  test('rslint-disable-next-line with -- description suppresses diagnostic', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// rslint-disable-next-line @typescript-eslint/no-explicit-any -- needed for legacy API',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('eslint-disable comment directives', () => {
  // -------------------------------------------------------------------
  // eslint-disable-next-line
  // -------------------------------------------------------------------

  test('eslint-disable-next-line suppresses diagnostic on next line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// eslint-disable-next-line @typescript-eslint/no-explicit-any',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-disable-next-line does not suppress other lines', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// eslint-disable-next-line @typescript-eslint/no-explicit-any',
        'const x: any = 1;',
        'const y: any = 2;', // not suppressed
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // eslint-disable-line
  // -------------------------------------------------------------------

  test('eslint-disable-line suppresses diagnostic on same line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        'const x: any = 1; // eslint-disable-line @typescript-eslint/no-explicit-any',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // eslint-disable + eslint-enable (block range) — core bug fix
  // -------------------------------------------------------------------

  test('eslint-disable + eslint-enable suppresses diagnostics inside range', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        'const b: any = 2;',
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
        'export { a, b };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-disable + eslint-enable does NOT suppress diagnostics outside range', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
        'const b: any = 2;', // outside range — should trigger
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // eslint-disable without enable (rest of file)
  // -------------------------------------------------------------------

  test('eslint-disable without enable suppresses rest of file', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        'const b: any = 2;',
        'const c: any = 3;',
        'export { a, b, c };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // wildcard (no rule name)
  // -------------------------------------------------------------------

  test('eslint-disable without rule name suppresses all rules', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable */',
        'const a: any = 1;',
        '/* eslint-enable */',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('wildcard eslint-disable-next-line suppresses all rules on next line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': ['// eslint-disable-next-line', 'const x: any = 1;'].join(
        '\n',
      ),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // block comment style for line directives
  // -------------------------------------------------------------------

  test('block comment style works for disable-next-line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable-next-line @typescript-eslint/no-explicit-any */',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // multiple disable/enable ranges
  // -------------------------------------------------------------------

  // -------------------------------------------------------------------
  // disable-next-line with -- description
  // -------------------------------------------------------------------

  test('eslint-disable-next-line with -- description suppresses diagnostic', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// eslint-disable-next-line @typescript-eslint/no-explicit-any -- needed for legacy API',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // disable-next-line inside function arguments (leading trivia fix)
  // -------------------------------------------------------------------

  test('eslint-disable-next-line works inside function call arguments', async () => {
    const ASSERTION_RULE = '@typescript-eslint/consistent-type-assertions';
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({
        [ASSERTION_RULE]: ['error', { assertionStyle: 'never' }],
      }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        'type Foo = Record<string, unknown>;',
        'const result = [1, 2, 3].reduce(',
        '  (acc, _item) => acc,',
        '  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions',
        '  {} as Foo,',
        ');',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      // Check the specific error format, not just the rule name string,
      // since the rule name also appears in the disable comment context lines
      expect(result.stdout).not.toContain(
        'consistent-type-assertions  \u2014 [error]',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-disable-next-line works inside array literals', async () => {
    const ASSERTION_RULE = '@typescript-eslint/consistent-type-assertions';
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({
        [ASSERTION_RULE]: ['error', { assertionStyle: 'never' }],
      }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        'type Foo = Record<string, unknown>;',
        'const arr = [',
        '  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions',
        '  {} as Foo,',
        '];',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.stdout).not.toContain(
        'consistent-type-assertions  \u2014 [error]',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // multiple disable/enable ranges
  // -------------------------------------------------------------------

  test('multiple disable/enable ranges work independently', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'export const a: any = 1;', // suppressed
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
        'export const b = 2;', // no violation
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'export const c: any = 3;', // suppressed
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('unused disable directives', () => {
  for (const prefix of ['eslint', 'rslint']) {
    for (const { name, source, rules, fixed } of [
      {
        name: 'line disable suppresses native and plugin diagnostics',
        source: `// ${prefix}-disable no-debugger, custom/check\ndebugger;\n`,
        rules: [],
      },
      {
        name: 'line enable ends block suppression',
        source: `/* ${prefix}-disable no-debugger, custom/check */\n// ${prefix}-enable no-debugger, custom/check\ndebugger;\n`,
        rules: ['', 'custom/check', 'no-debugger'],
        fixed: ' \n \ndebugger;\n',
      },
      {
        name: 'unused line disable is reported and removed',
        source: `// ${prefix}-disable no-debugger, custom/check\nconst x = 1;\n`,
        rules: [''],
        fixed: ' \nconst x = 1;\n',
      },
    ]) {
      test(`${prefix}: ${name}`, async () => {
        const tempDir = await createTempDir({
          'test.ts': source,
          'rslint.config.mjs': `export default [{
            files: ['**/*.ts'],
            linterOptions: { reportUnusedDisableDirectives: 'warn' },
            plugins: { custom: { rules: { check: {
              meta: { schema: [] },
              create(context) { return { DebuggerStatement(node) { context.report({ node, message: 'plugin debugger' }); } }; }
            } } } },
            rules: { 'no-debugger': 'error', 'custom/check': 'error' }
          }];`,
        });
        try {
          for (const args of [[], ['--fix']]) {
            const result = await runRslint(
              [...args, '--format', 'jsonline', 'test.ts'],
              tempDir,
            );
            const expectedRules = args.length
              ? rules.filter((rule) => rule !== '')
              : rules;
            expect(result.exitCode).toBe(
              expectedRules.some((rule) => rule !== '') ? 1 : 0,
            );
            const reportedRules = result.stdout
              .split('\n')
              .filter(Boolean)
              .map(
                (line) => (JSON.parse(line) as { ruleName: string }).ruleName,
              )
              .sort();
            expect(reportedRules).toEqual(expectedRules);
            expect(
              await fs.readFile(path.join(tempDir, 'test.ts'), 'utf8'),
            ).toBe(args.length ? (fixed ?? source) : source);
          }
        } finally {
          await cleanupTempDir(tempDir);
        }
      });
    }
  }

  test('uses mixed native/plugin usage and fixes only unused entries', async () => {
    const source =
      '/* eslint-disable no-debugger, custom/check, no-console -- reason */\ndebugger;\n';
    const tempDir = await createTempDir({
      'test.ts': source,
      'rslint.config.mjs': `export default [{
        files: ['**/*.ts'],
        linterOptions: { reportUnusedDisableDirectives: 'warn' },
        plugins: { custom: { rules: { check: {
          meta: { schema: [] },
          create(context) { return { DebuggerStatement(node) { context.report({ node, message: 'plugin debugger' }); } }; }
        } } } },
        rules: { 'no-debugger': 'error', 'custom/check': 'error', 'no-console': 'off' }
      }];`,
    });
    try {
      const report = await runRslint(['test.ts'], tempDir);
      expect(report.exitCode).toBe(0);
      expect(report.stdout).toContain(
        "Unused eslint-disable directive (no problems were reported from 'no-console').",
      );
      expect(report.stdout).not.toContain('plugin debugger');
      const fixed = await runRslint(['--fix', 'test.ts'], tempDir);
      expect(fixed.exitCode).toBe(0);
      expect(await fs.readFile(path.join(tempDir, 'test.ts'), 'utf8')).toBe(
        '/* eslint-disable no-debugger, custom/check -- reason */\ndebugger;\n',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('per-file settings apply to zero-rule files and CLI severity overrides them', async () => {
    const tempDir = await createTempDir({
      'on.ts': '/* rslint-disable */\nconst x = 1;\n',
      'off.ts': '/* rslint-disable */\nconst x = 1;\n',
      'rslint.config.mjs': `export default [
        { files: ['**/*.ts'], linterOptions: { reportUnusedDisableDirectives: 'error' } },
        { files: ['off.ts'], linterOptions: { reportUnusedDisableDirectives: false } }
      ];`,
    });
    try {
      expect((await runRslint(['on.ts'], tempDir)).exitCode).toBe(1);
      expect((await runRslint(['off.ts'], tempDir)).exitCode).toBe(0);
      expect(
        (
          await runRslint(
            [
              '--report-unused-disable-directives-severity=warn',
              '--report-unused-disable-directives',
              'off.ts',
            ],
            tempDir,
          )
        ).exitCode,
      ).toBe(0);
      expect(
        (
          await runRslint(
            ['--report-unused-disable-directives=false', 'on.ts'],
            tempDir,
          )
        ).stdout,
      ).not.toContain('Unused eslint-disable');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

test('unused comment fixes preserve BOM, Unicode, and CRLF', async () => {
  const tempDir = await createTempDir({
    'test.ts': '\ufeff/* eslint-disable */\r\nconst note = "中文😀";\r\n',
    'rslint.config.mjs': `export default [{ files: ['**/*.ts'], linterOptions: { reportUnusedDisableDirectives: true } }];`,
  });
  try {
    const result = await runRslint(['--fix', 'test.ts'], tempDir);
    expect(result.exitCode).toBe(0);
    expect(await fs.readFile(path.join(tempDir, 'test.ts'), 'utf8')).toBe(
      '\ufeff \r\nconst note = "中文😀";\r\n',
    );
  } finally {
    await cleanupTempDir(tempDir);
  }
});

test('unused directive settings follow each config directory', async () => {
  const tempDir = await createTempDir({
    'a/input.ts': '/* eslint-disable */\nconst x = 1;\n',
    'b/input.ts': '/* eslint-disable */\nconst x = 1;\n',
    'a/rslint.config.mjs': `export default [{ files: ['**/*.ts'], linterOptions: { reportUnusedDisableDirectives: 'error' } }];`,
    'b/rslint.config.mjs': `export default [{ files: ['**/*.ts'], linterOptions: { reportUnusedDisableDirectives: 'off' } }];`,
  });
  try {
    const report = await runRslint(['a/input.ts', 'b/input.ts'], tempDir);
    expect(report.exitCode).toBe(1);
    expect(
      report.stdout.match(/Unused eslint-disable directive/g),
    ).toHaveLength(1);
    const overridden = await runRslint(
      ['--report-unused-disable-directives', 'a/input.ts', 'b/input.ts'],
      tempDir,
    );
    expect(overridden.exitCode).toBe(1);
    expect(
      overridden.stdout.match(/Unused eslint-disable directive/g),
    ).toHaveLength(2);
  } finally {
    await cleanupTempDir(tempDir);
  }
});
