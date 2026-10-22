import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, describe, expect, test } from 'rstack/test';
import { admitRustFiles } from '../src/rust/admission.js';

const roots: string[] = [];
function fixture(files: Record<string, string>): string {
  const root = fs.realpathSync(
    fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rust-admission-')),
  );
  roots.push(root);
  for (const [name, text] of Object.entries(files)) {
    fs.mkdirSync(path.dirname(path.join(root, name)), { recursive: true });
    fs.writeFileSync(path.join(root, name), text);
  }
  return root;
}
function names(root: string, config: unknown, patterns?: string[]): string[] {
  return admitRustFiles({ cwd: root, config, patterns }).map((file) =>
    path.relative(root, file.path).split(path.sep).join('/'),
  );
}
afterEach(() => {
  for (const root of roots.splice(0))
    fs.rmSync(root, { recursive: true, force: true });
});

describe('Rust explicit flat-config admission', () => {
  test('shares path coordinates across nested ignores and config matches within one invocation', () => {
    const root = fixture({
      '.gitignore': '*.generated.ts\n',
      'src/.gitignore': '!deep/keep.generated.ts\n',
      'src/deep/a.ts': '',
      'src/deep/b.ts': '',
      'src/deep/keep.generated.ts': '',
      'src/deep/skip.generated.ts': '',
    });
    const config = Array.from({ length: 12 }, () => ({
      files: ['**/*.ts'],
      ignores: ['missing/**', 'other/**'],
      rules: { 'no-debugger': 'error' },
    }));
    const original = path.relative;
    const calls = new Map<string, number>();
    let files: ReturnType<typeof admitRustFiles>;
    path.relative = (from, to) => {
      const key = JSON.stringify([from, to]);
      calls.set(key, (calls.get(key) ?? 0) + 1);
      return original(from, to);
    };
    try {
      files = admitRustFiles({ cwd: root, config });
    } finally {
      path.relative = original;
    }
    expect(files.map((file) => original(root, file.path))).toEqual([
      path.join('src', 'deep', 'a.ts'),
      path.join('src', 'deep', 'b.ts'),
      path.join('src', 'deep', 'keep.generated.ts'),
    ]);
    expect(
      files.every((file) => file.config.rules[0]?.name === 'no-debugger'),
    ).toBe(true);
    expect(calls.size).toBeGreaterThan(0);
    expect(Math.max(...calls.values())).toBe(1);
  });

  test('does not retain path or ignore state across admissions when a leaf alias changes', () => {
    const root = fixture({ 'first/a.ts': '', 'second/a.ts': '' });
    const alias = path.join(root, 'alias.ts');
    fs.symlinkSync(path.join(root, 'first/a.ts'), alias);
    const options = { cwd: root, config: [{}], patterns: ['alias.ts'] };
    const [first] = admitRustFiles(options);
    expect(first.path).toBe(alias);
    expect(first.canonicalPath).toBe(path.join(root, 'first/a.ts'));
    fs.unlinkSync(alias);
    fs.symlinkSync(path.join(root, 'second/a.ts'), alias);
    const [second] = admitRustFiles(options);
    expect(second.path).toBe(alias);
    expect(second.canonicalPath).toBe(path.join(root, 'second/a.ts'));
    fs.writeFileSync(path.join(root, '.gitignore'), 'alias.ts\n');
    expect(admitRustFiles(options)).toEqual([]);
    fs.unlinkSync(path.join(root, '.gitignore'));
    expect(admitRustFiles(options).map((file) => file.path)).toEqual([alias]);
  });
  test('preserves distinct module and inline owners without imposing basePath scope', () => {
    const root = fixture({ 'owner/a.ts': '', 'a.ts': '' });
    const config = [
      { files: ['a.ts'], rules: { 'no-debugger': 'error' } },
      { files: ['a.ts'], rules: { 'no-empty': 'warn' } },
    ];
    const options = {
      cwd: root,
      config,
      configEntryDirectories: [path.join(root, 'owner'), root],
    };
    const files = admitRustFiles(options);
    expect(
      files.find((file) => file.path === path.join(root, 'a.ts'))?.config.rules,
    ).toEqual([{ name: 'no-empty', severity: 1, options: [] }]);
    expect(
      files.find((file) => file.path === path.join(root, 'owner/a.ts'))?.config
        .rules,
    ).toEqual([{ name: 'no-debugger', severity: 2, options: [] }]);
    for (const configEntryDirectories of [[], ['relative', root]])
      expect(() =>
        admitRustFiles({ ...options, configEntryDirectories }),
      ).toThrow('configEntryDirectories');
  });
  test('admits all eight default extensions and diagnostic-free files, excluding dependency directories', () => {
    const root = fixture({
      'a.js': '',
      'b.mjs': '',
      'c.cjs': '',
      'd.jsx': '',
      'e.ts': '',
      'f.tsx': '',
      'g.mts': '',
      'h.cts': '',
      '.hidden.ts': '',
      'other.json': '',
      'node_modules/leak.ts': '',
      '.git/leak.ts': '',
      'upper.TS': '',
    });
    expect(names(root, [{}])).toEqual([
      '.hidden.ts',
      'a.js',
      'b.mjs',
      'c.cjs',
      'd.jsx',
      'e.ts',
      'f.tsx',
      'g.mts',
      'h.cts',
      'upper.TS',
    ]);
    expect(
      admitRustFiles({
        cwd: root,
        config: [{ rules: { 'no-debugger': 'error' } }],
      }).find((file) => file.path.endsWith('upper.TS'))?.config.rules,
    ).toEqual([]);
    expect(names(root, [{ files: ['**/*.TS'] }])).toContain('upper.TS');
  });

  test('distinguishes global ignores, entry-local ignores, and authored empty rule maps', () => {
    const root = fixture({ 'skip/a.ts': '', 'keep/a.ts': '', 'local.ts': '' });
    const config = [
      { ignores: ['skip/**'] },
      { rules: { 'no-debugger': 'error' } },
      { ignores: ['local.ts'], rules: { 'no-debugger': 'off' } },
    ];
    const files = admitRustFiles({ cwd: root, config });
    expect(files.map((file) => path.relative(root, file.path))).toEqual([
      'keep/a.ts',
      'local.ts',
    ]);
    expect(files[0].config.rules).toEqual([]);
    expect(files[1].config.rules).toEqual([
      { name: 'no-debugger', severity: 2, options: [] },
    ]);
    expect(names(root, [{ ignores: ['skip/**'], rules: {} }])).toContain(
      'skip/a.ts',
    );
  });

  test('keeps positive directory blocks absolute but allows file-level re-inclusion', () => {
    const root = fixture({
      'blocked/a.ts': '',
      'covered/a.ts': '',
      'covered/b.ts': '',
    });
    expect(
      names(root, [
        {
          ignores: [
            'blocked/**',
            '!blocked/a.ts',
            'covered/**/*',
            '!covered/a.ts',
          ],
        },
        {},
      ]),
    ).toEqual(['covered/a.ts']);
  });

  test('uses sequential nested gitignore rules and refuses child resurrection through an ignored parent', () => {
    const root = fixture({
      '.gitignore': 'blocked/\n*.generated.ts\n',
      'blocked/.gitignore': '!a.ts\n',
      'blocked/a.ts': '',
      'src/.gitignore': '!keep.generated.ts\n',
      'src/keep.generated.ts': '',
      'src/drop.generated.ts': '',
      'src/plain.ts': '',
    });
    expect(names(root, [{}])).toEqual([
      'src/keep.generated.ts',
      'src/plain.ts',
    ]);
    expect(names(root, [{}], ['blocked/a.ts'])).toEqual([]);
  });

  test('global negation reopens Git traversal without implicitly including every descendant', () => {
    const root = fixture({
      '.gitignore': 'blocked/\n',
      'blocked/a.ts': '',
      'blocked/b.ts': '',
    });
    expect(
      names(root, [{ ignores: ['!blocked', '!blocked/a.ts'] }, {}]),
    ).toEqual(['blocked/a.ts']);
    expect(names(root, [{ ignores: ['!blocked/**/*'] }, {}])).toEqual([]);
  });

  test('Git contents-only globs do not block the directory node itself', () => {
    const root = fixture({
      '.gitignore': 'covered/**\n!covered/a.ts\n',
      'covered/a.ts': '',
      'covered/b.ts': '',
    });
    expect(names(root, [{}])).toEqual(['covered/a.ts']);
  });

  test('keeps Git braces, parentheses, quoted metacharacters, leading spaces and trailing spaces literal', () => {
    const root = fixture({
      '.gitignore':
        'literal-{a,b}.ts\na(b).ts\nquoted\\*.ts\n leading.ts\ntrailing\\ .ts  \n[!a].ts\n',
      'literal-a.ts': '',
      'literal-{a,b}.ts': '',
      'ab.ts': '',
      'a(b).ts': '',
      'quoted*.ts': '',
      'quotedX.ts': '',
      ' leading.ts': '',
      'leading.ts': '',
      'trailing .ts': '',
      'a.ts': '',
      'b.ts': '',
    });
    expect(names(root, [{}])).toEqual([
      'a.ts',
      'ab.ts',
      'leading.ts',
      'literal-a.ts',
      'quotedX.ts',
    ]);
  });

  test('matches nested AND selectors, negation, brace expansion and inherited preset arrays', () => {
    const root = fixture({
      'src/a.ts': '',
      'src/a.test.ts': '',
      'src/b.tsx': '',
      'other/a.ts': '',
    });
    const files = admitRustFiles({
      cwd: root,
      config: [
        [{ rules: { 'no-debugger': 'warn' } }],
        {
          files: [['src/**/*.{ts,tsx}', '!**/*.test.ts']],
          rules: { 'no-debugger': 'error' },
        },
      ],
    });
    expect(
      Object.fromEntries(
        files.map((file) => [
          path.relative(root, file.path),
          file.config.rules[0].severity,
        ]),
      ),
    ).toEqual({
      'other/a.ts': 1,
      'src/a.test.ts': 1,
      'src/a.ts': 2,
      'src/b.tsx': 2,
    });
  });

  test('merges severity-only overrides without dropping options and deep-merges settings/globals', () => {
    const root = fixture({ 'a.ts': '' });
    const [file] = admitRustFiles({
      cwd: root,
      config: [
        {
          rules: { example: ['error', { allow: ['a'] }] },
          settings: { x: { a: 1, b: [1] } },
          languageOptions: { globals: { first: 'readonly' } },
        },
        {
          rules: { example: ['warn'] },
          settings: { x: { b: [2], c: true } },
          languageOptions: { globals: { second: 'writable' } },
        },
      ],
    });
    expect(file.config.rules).toEqual([
      { name: 'example', severity: 1, options: [{ allow: ['a'] }] },
    ]);
    expect(file.config.settings).toEqual({ x: { a: 1, b: [2], c: true } });
    expect(file.config.languageOptions).toEqual({
      globals: { first: 'readonly', second: 'writable' },
    });
  });

  test('preserves module owner and basePath separately, including project override origin', () => {
    const root = fixture({
      'src/a.ts': '',
      'other/b.ts': '',
      'tools/config.js': '',
    });
    const owner = path.join(root, 'tools');
    const files = admitRustFiles({
      cwd: root,
      configDirectory: owner,
      config: [
        {
          basePath: 'src',
          languageOptions: { parserOptions: { project: ['a.json'] } },
        },
        {
          basePath: '.',
          languageOptions: {
            parserOptions: { projectService: true, project: false },
          },
        },
        {
          files: ['../other/**'],
          languageOptions: { parserOptions: { projectService: false } },
        },
      ],
    });
    expect(
      files.find((file) => file.path.endsWith('/src/a.ts'))?.config
        .projectPolicy,
    ).toEqual({
      serviceRootDirectory: owner,
      projectDisabled: false,
      defaultProjectDisabled: false,
    });
    expect(
      files.find((file) => file.path.endsWith('/other/b.ts'))?.config
        .projectPolicy,
    ).toEqual({ projectDisabled: true, defaultProjectDisabled: true });
    const [selected] = admitRustFiles({
      cwd: root,
      patterns: ['src/a.ts'],
      config: [
        {
          basePath: 'src',
          languageOptions: { parserOptions: { project: ['a.json'] } },
        },
        { languageOptions: { parserOptions: { projectService: false } } },
      ],
    });
    expect(selected.config.projectPolicy.explicitProject).toEqual({
      patterns: ['a.json'],
      baseDirectory: path.join(root, 'src'),
    });
  });

  test('keeps configured type rules even when project service is disabled; planner must decide checker availability', () => {
    const root = fixture({ 'src/a.ts': '', 'fixture/a.ts': '' });
    const files = admitRustFiles({
      cwd: root,
      config: [
        {
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-unsafe-call': 'error' },
          languageOptions: { parserOptions: { projectService: true } },
        },
        {
          files: ['fixture/**'],
          languageOptions: { parserOptions: { projectService: false } },
        },
      ],
    });
    expect(files.every((file) => file.config.rules.length === 1)).toBe(true);
    expect(
      files.find((file) => file.path.endsWith('/fixture/a.ts'))?.config
        .projectPolicy.defaultProjectDisabled,
    ).toBe(true);
  });

  test('deduplicates overlapping targets, follows explicit directory aliases, and preserves leaf-symlink names', () => {
    const root = fixture({ 'src/a.ts': '', 'src/skip.ts': '' });
    fs.symlinkSync('src', path.join(root, 'alias'), 'dir');
    fs.symlinkSync('src/skip.ts', path.join(root, 'named.ts'));
    const config = [
      { basePath: 'src', files: ['a.ts'], rules: { 'no-debugger': 'error' } },
      { rules: { 'no-ternary': 'warn' } },
    ];
    expect(names(root, config)).toEqual([
      'named.ts',
      'src/a.ts',
      'src/skip.ts',
    ]);
    expect(names(root, config, ['src', 'src/a.ts', '**/a.ts'])).toEqual([
      'src/a.ts',
      'src/skip.ts',
    ]);
    const [alias] = admitRustFiles({
      cwd: root,
      config,
      patterns: ['alias/a.ts'],
    });
    expect(alias.path).toBe(path.join(root, 'alias/a.ts'));
    expect(alias.canonicalPath).toBe(path.join(root, 'src/a.ts'));
    expect(alias.config.rules.map((rule) => rule.name)).toEqual([
      'no-debugger',
      'no-ternary',
    ]);
    expect(
      admitRustFiles({
        cwd: root,
        config,
        patterns: ['named.ts'],
      })[0].config.rules.map((rule) => rule.name),
    ).toEqual(['no-ternary']);
  });

  test('rejects unknown behavior instead of silently ignoring configuration or falling back', () => {
    const root = fixture({ 'a.ts': '' });
    for (const config of [
      [{ processor: {} }],
      [{ languageOptions: { parserOptions: { projectService: {} } } }],
      [{ languageOptions: { parserOptions: { project: true } } }],
      [
        {
          languageOptions: {
            parserOptions: { projectService: true, project: [] },
          },
        },
      ],
      [{ plugins: { community: { rules: {} } } }],
      [{ languageOptions: { parserOptions: { tsconfigRootDir: 'relative' } } }],
      [{ files: ['@(src|lib)/*.ts'] }],
    ]) {
      expect(() => admitRustFiles({ cwd: root, config })).toThrow();
    }
    expect(() =>
      admitRustFiles({ cwd: root, config: [], patterns: [] }),
    ).toThrow('non-empty');
    expect(() =>
      admitRustFiles({ cwd: root, config: [], patterns: ['../outside'] }),
    ).toThrow('outside cwd');
  });
});
