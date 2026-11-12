import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import type { RuleTesterConfig } from '@rslint/rule-tester';

const compilerOptions = {
  jsx: 'preserve',
  target: 'esnext',
  module: 'commonjs',
  strict: true,
  esModuleInterop: true,
  lib: ['es2015', 'es2017', 'esnext', 'DOM'],
};

// Preserve the original suite's TypeScript projects and virtual TSX filename.
export const config: RuleTesterConfig = {
  plugin: 'rstest',
  filename: 'src/virtual.tsx',
  fixtureFiles: {
    'tsconfig.files.json': JSON.stringify({
      compilerOptions: { ...compilerOptions, allowJs: true },
      include: ['files/**/*'],
    }),
    'tsconfig.virtual.json': JSON.stringify({
      compilerOptions,
      include: ['src/virtual.tsx', 'src/virtual.ts'],
    }),
    'files/unbound-method.ts': 'export {};\n',
  },
  languageOptions: {
    parserOptions: {
      projectService: false,
      project: ['./tsconfig.files.json', './tsconfig.virtual.json'],
    },
  },
};

// Multi-rule and fix-convergence tests use the same project as RuleTester.
export async function withFixture<T>(
  run: (root: string) => Promise<T>,
): Promise<T> {
  const root = await fs.realpath(
    await fs.mkdtemp(path.join(os.tmpdir(), 'rslint-rstest-')),
  );
  try {
    for (const [name, content] of Object.entries(config.fixtureFiles ?? {})) {
      const file = path.join(root, name);
      await fs.mkdir(path.dirname(file), { recursive: true });
      await fs.writeFile(file, content);
    }
    return await run(root);
  } finally {
    await fs.rm(root, {
      recursive: true,
      force: true,
      maxRetries: 20,
      retryDelay: 1_000,
    });
  }
}
