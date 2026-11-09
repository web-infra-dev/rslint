import path from 'node:path';
import { describe, expect, test } from 'rstack/test';
import { buildLintRequest } from '../src/request.ts';

const root = path.resolve('/fixture');
const config = {
  plugin: 'node',
  languageOptions: {
    sourceType: 'module' as const,
    globals: { a: 'readonly' as const },
    parserOptions: { project: ['./tsconfig.json'] },
  },
};

describe('buildLintRequest', () => {
  test('lints the case code as input.js by default', () => {
    const { request, filename } = buildLintRequest(config, 'node/x', root, {
      code: 'foo()',
    });
    expect(filename).toBe(path.join(root, 'input.js'));
    expect(request).toMatchObject({
      configDirectory: root,
      workingDirectory: root,
      fileContents: { [filename]: 'foo()' },
    });
    expect(request.config?.[0]).toMatchObject({
      plugins: ['node'],
      rules: { 'node/x': ['error'] },
    });
  });

  test('places the file relative to the fixture root', () => {
    const { filename } = buildLintRequest(config, 'node/x', root, {
      code: '',
      filename: 'src/a.ts',
    });
    expect(filename).toBe(path.join(root, 'src/a.ts'));
  });

  test('uses the tester filename as the default', () => {
    const { filename } = buildLintRequest(
      { ...config, filename: 'src/a.tsx' },
      'node/x',
      root,
      { code: '' },
    );
    expect(filename).toBe(path.join(root, 'src/a.tsx'));
  });

  test('appends options, resolving a function against the root', () => {
    const { request } = buildLintRequest(config, 'node/x', root, {
      code: '',
      options: (dir) => [{ dir }],
    });
    expect(request.config?.[0]).toMatchObject({
      rules: { 'node/x': ['error', { dir: root }] },
    });
  });

  test('merges case language options over the tester defaults', () => {
    const { request } = buildLintRequest(config, 'node/x', root, {
      code: '',
      languageOptions: {
        sourceType: 'commonjs',
        globals: { b: 'writable' },
        parserOptions: { projectService: true },
      },
      settings: { node: { version: '18' } },
    });
    expect(request.config?.[0]).toMatchObject({
      languageOptions: {
        sourceType: 'commonjs',
        globals: { a: 'readonly', b: 'writable' },
        parserOptions: { projectService: true, project: ['./tsconfig.json'] },
      },
      settings: { node: { version: '18' } },
    });
  });

  test('disables the project service unless a case enables it', () => {
    const { request } = buildLintRequest({ plugin: 'node' }, 'node/x', root, {
      code: '',
    });
    expect(request.config?.[0]).toMatchObject({
      languageOptions: { parserOptions: { projectService: false } },
    });
  });
});
