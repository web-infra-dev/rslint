import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  addUntypedPackage,
  normalizeOutput,
  TS_CONFIG,
  makeConfig,
} from './helpers';

// ---------------------------------------------------------------------------
// default format
// ---------------------------------------------------------------------------

describe('--type-check output snapshots (default format)', () => {
  test('type error only (TS2322)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': "const x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('lint error only with --type-check', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/no-explicit-any': 'error',
      }),
      'test.ts': 'let a: any = 1;\nconsole.log(a);\n',
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('both lint error and type error', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/no-explicit-any': 'error',
      }),
      'test.ts': "let a: any = 1;\nconst x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('no errors with --type-check', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': 'const x: number = 42;\n',
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('lint error without --type-check (baseline)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/no-explicit-any': 'error',
      }),
      'test.ts': 'let a: any = 1;\nconsole.log(a);\n',
    });
    try {
      const result = await runRslint([], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multi-line: MessageChain (TS7016 untyped module)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': "import foo from 'some-untyped-pkg';\nconsole.log(foo);\n",
    });
    await addUntypedPackage(tempDir, 'some-untyped-pkg');
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multi-line: deep MessageChain (nested interface mismatch)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': [
        'interface A { x: { y: { z: string } } }',
        'interface B { x: { y: { z: number } } }',
        'const a: A = {} as B;',
        '',
      ].join('\n'),
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multi-line: RelatedInformation (property type mismatch)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': [
        'interface Foo { bar: { baz: number } }',
        "const obj: Foo = { bar: { baz: 'wrong' } };",
        '',
      ].join('\n'),
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multi-line: RelatedInformation (missing property TS2741)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': [
        'interface C { required: string; also: number }',
        "const c: C = { required: 'hi' };",
        '',
      ].join('\n'),
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multi-line lint error (unbound-method)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/unbound-method': 'error',
      }),
      'test.ts': [
        'class Foo {',
        '  method() { return this; }',
        '}',
        'const foo = new Foo();',
        'const fn = foo.method;',
        'console.log(fn);',
        '',
      ].join('\n'),
    });
    try {
      const result = await runRslint([], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multi-line lint error + type error together', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/unbound-method': 'error',
      }),
      'test.ts': [
        'class Foo {',
        '  method() { return this; }',
        '}',
        'const foo = new Foo();',
        'const fn = foo.method;',
        "const x: number = 'hello';",
        '',
      ].join('\n'),
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--quiet suppresses warnings, keeps type errors', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/no-explicit-any': 'warn',
      }),
      'test.ts': "let a: any = 1;\nconst x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(['--type-check', '--quiet'], tempDir);
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

// ---------------------------------------------------------------------------
// jsonline format
// ---------------------------------------------------------------------------

describe('--type-check output snapshots (jsonline format)', () => {
  test('type error in jsonline format', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': "const x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(
        ['--type-check', '--format', 'jsonline'],
        tempDir,
      );
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('both lint and type error in jsonline format', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/no-explicit-any': 'error',
      }),
      'test.ts': "let a: any = 1;\nconst x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(
        ['--type-check', '--format', 'jsonline'],
        tempDir,
      );
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

// ---------------------------------------------------------------------------
// github format
// ---------------------------------------------------------------------------

describe('--type-check output snapshots (github format)', () => {
  test('type error in github format', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': "const x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(
        ['--type-check', '--format', 'github'],
        tempDir,
      );
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('both lint and type error in github format', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/no-explicit-any': 'error',
      }),
      'test.ts': "let a: any = 1;\nconst x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(
        ['--type-check', '--format', 'github'],
        tempDir,
      );
      expect(normalizeOutput(result.stdout, tempDir)).toMatchSnapshot();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});
