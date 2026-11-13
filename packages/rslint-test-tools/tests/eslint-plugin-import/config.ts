import fs from 'node:fs';
import path from 'node:path';
import type { RuleTesterConfig } from '@rslint/rule-tester';

const directory = import.meta.dirname;
const compilerOptions = {
  jsx: 'preserve',
  target: 'es5',
  module: 'commonjs',
  strict: true,
  lib: ['es2015', 'es2017', 'esnext'],
  experimentalDecorators: true,
};
const files: Record<string, string> = {
  'tsconfig.files.json': JSON.stringify({
    compilerOptions: {
      ...compilerOptions,
      esModuleInterop: false,
      allowJs: true,
    },
    include: ['files/**/*', 'fixtures/**/*'],
  }),
  'tsconfig.virtual.json': JSON.stringify({
    compilerOptions: {
      ...compilerOptions,
      esModuleInterop: true,
      baseUrl: '.',
      paths: {
        '@/*': ['src/order-internal-modules/*'],
        buffer: ['src/order-internal-modules/api/service.ts'],
      },
    },
    include: ['src/virtual.tsx', 'src/virtual.ts'],
  }),
};
for (const name of ['files', 'fixtures', 'src']) {
  const walk = (relative: string): void => {
    const absolute = path.join(directory, relative);
    if (!fs.existsSync(absolute)) return;
    for (const entry of fs.readdirSync(absolute, { withFileTypes: true })) {
      const file = path.join(relative, entry.name);
      if (entry.isDirectory()) walk(file);
      else if (entry.isFile())
        files[file.replaceAll(path.sep, '/')] = fs.readFileSync(
          path.join(directory, file),
          'utf8',
        );
    }
  };
  walk(name);
}

// Keep the original filesystem condition explicit on each affected case.
export const caseSensitiveFileSystem = !fs.existsSync(
  path.join(directory, 'fixtures/named-rule/Named-Exports.js'),
);

export function fixtureConfig(archive?: string): RuleTesterConfig {
  const fixtureFiles = { ...files };
  if (archive) {
    const text = fs.readFileSync(
      path.resolve(directory, '../../../..', archive),
      'utf8',
    );
    const chunks = text.split(/^-- (.+) --\r?$/m);
    if (chunks.length < 3)
      throw new Error(`Missing Import fixtures: ${archive}`);
    for (let i = 1; i < chunks.length; i += 2) {
      const filename = chunks[i].replaceAll('\\', '/');
      if (filename.startsWith('/') || filename.split('/').includes('..'))
        throw new Error(`Fixture path escapes its root: ${filename}`);
      fixtureFiles[filename] = chunks[i + 1].replace(/^\r?\n/, '');
    }
  }
  return {
    plugin: 'import',
    filename: 'src/virtual.ts',
    fixtureFiles,
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.files.json', './tsconfig.virtual.json'],
      },
    },
  };
}

// Only the ancestor spelling changes in the caseSensitiveStrict examples.
export function fixtureUpperRoot(root: string): string {
  return path
    .join(path.dirname(root).toUpperCase(), path.basename(root))
    .replaceAll(path.sep, '/');
}

// Resolve fixed fixture placeholders without mutating options/settings tables.
export function resolveFixturePaths<T>(value: T, root: string): T {
  const resolve = (item: unknown): unknown => {
    if (typeof item === 'string')
      return item
        .replaceAll('<fixture-upper>', fixtureUpperRoot(root))
        .replaceAll('<fixture>', root.replaceAll(path.sep, '/'));
    if (Array.isArray(item)) return item.map(resolve);
    if (item && typeof item === 'object')
      return Object.fromEntries(
        Object.entries(item).map(([key, item]) => [key, resolve(item)]),
      );
    return item;
  };
  return resolve(value) as T;
}
