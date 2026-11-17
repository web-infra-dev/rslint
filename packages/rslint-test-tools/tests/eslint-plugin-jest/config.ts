import fs from 'node:fs';
import path from 'node:path';
import type { RuleTesterConfig } from '@rslint/rule-tester';

// Preserve the original projects and virtual TSX filename.
const compilerOptions = {
  jsx: 'react',
  target: 'es5',
  module: 'commonjs',
  strict: true,
  esModuleInterop: true,
  lib: ['es2015', 'es2017', 'esnext'],
};
const files: Record<string, string> = {
  'tsconfig.files.json': JSON.stringify({
    compilerOptions: { ...compilerOptions, allowJs: true },
    include: ['files/**/*'],
  }),
  'tsconfig.virtual.json': JSON.stringify({
    compilerOptions,
    include: ['src/virtual.tsx', 'src/virtual.ts'],
  }),
};

const fixtureDirectory = path.join(import.meta.dirname, 'fixtures/upstream');
function readFixtures(directory: string): void {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const file = path.join(directory, entry.name);
    if (entry.isDirectory()) readFixtures(file);
    else if (entry.isFile())
      files[
        path.relative(import.meta.dirname, file).replaceAll(path.sep, '/')
      ] = fs.readFileSync(file, 'utf8');
  }
}
readFixtures(fixtureDirectory);

export function fixtureConfig(archive?: string): RuleTesterConfig {
  const fixtureFiles: Record<string, string> = { ...files };
  if (archive) {
    const text = fs.readFileSync(
      path.resolve(import.meta.dirname, '../../../..', archive),
      'utf8',
    );
    const chunks = text.split(/^-- (.+) --\r?$/m);
    if (chunks.length < 3) throw new Error(`Missing Jest fixtures: ${archive}`);
    for (let i = 1; i < chunks.length; i += 2) {
      const filename = chunks[i].replaceAll('\\', '/');
      if (path.isAbsolute(filename) || filename.split('/').includes('..'))
        throw new Error(`Fixture path escapes its root: ${filename}`);
      fixtureFiles[filename] = chunks[i + 1].replace(/^\r?\n/, '');
    }
  }
  return {
    plugin: 'jest',
    filename: 'src/virtual.tsx',
    fixtureFiles,
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.files.json', './tsconfig.virtual.json'],
      },
    },
  };
}
