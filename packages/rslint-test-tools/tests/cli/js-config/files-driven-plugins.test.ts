import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs/promises';
import path from 'node:path';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  linkNodeModules,
} from './helpers.js';

// Mixed native + ESLint-plugin rules + gap files. Previously every
// gap-file test in this suite used the string-array `plugins:
// ['@typescript-eslint']` form (rule namespaces only, no in-process
// plugin code). That doesn't exercise the eslint-plugin compat path —
// no worker, no IPC reverse-RPC, no `eslintPlugins: {}` mapping. The
// real cross-cutting concern users hit is: native rules + a real
// `eslintPlugins: { unicorn: unicornPlugin }` + gap files. All three
// must produce diagnostics on both program-file AND gap-file.
describe('Files-driven lint: mixed native + eslintPlugins + gap files', () => {
  test('gap file gets BOTH native syntax rule AND eslint-plugin rule diagnostics', async () => {
    // tsconfig only includes src/; scripts/ is the gap-file directory.
    // The config enables ONE native rule (`@typescript-eslint/ban-
    // ts-comment` — fires on `// @ts-ignore`, doesn't need type info)
    // AND ONE eslint-plugin rule (`uni/no-null` — runs entirely
    // inside the Node worker against oxc-parser's AST). The plugin is
    // declared in object form via `eslintPlugins`, so the compat
    // dispatcher actually hands off to the worker pool.
    //
    // What this pins: in the mixed path (compatOnlyMode=false because
    // a native rule is present) the cmd.go fallback Program is built
    // for gap files (cmd.go:1405) AND the compat dispatcher is invoked
    // for those same files. Both signals must reach the diagnostic
    // stream. If either side regresses — e.g. fallback Program is
    // skipped, or DispatchCompat is gated only on program-file paths —
    // the test catches it.
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    // @ts-ignore — eslintPlugins is the in-process plugin form
    eslintPlugins: { uni: unicorn },
    plugins: ['@typescript-eslint'],
    rules: {
      // Native, syntax-only rule.
      '@typescript-eslint/ban-ts-comment': 'error',
      // ESLint-plugin rule routed through the worker pool.
      'uni/no-null': 'error',
    },
  },
];`,
      // In-tsconfig file: both rules should fire here.
      'src/index.ts': `// @ts-ignore\nconst a = null;\n`,
      // Gap file (not in tsconfig include): both rules should STILL
      // fire here — native via the fallback Program, plugin via the
      // worker. This is the cross-cutting invariant the test pins.
      'scripts/build.ts': `// @ts-ignore\nconst b = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));

      const srcRules = diagnostics
        .filter((d: any) => d.filePath.includes('src/index.ts'))
        .map((d: any) => d.ruleName);
      const gapRules = diagnostics
        .filter((d: any) => d.filePath.includes('scripts/build.ts'))
        .map((d: any) => d.ruleName);

      // src (in tsconfig): both rules fire.
      expect(srcRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(srcRules).toContain('uni/no-null');

      // gap (NOT in tsconfig): native syntax rule still fires via
      // fallback Program, AND eslint-plugin rule still fires via
      // worker compat dispatch.
      expect(gapRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(gapRules).toContain('uni/no-null');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--fix on mixed native + plugin + gap files rewrites every file', async () => {
    // Multi-pass --fix in the MIXED path takes the RunLinter branch
    // (H3's PERF TODO — only the all-plugin compatOnlyMode path uses
    // the lighter DispatchCompat fix-loop). This test guards that the
    // mixed fix-loop still re-lints + re-fixes BOTH program-files AND
    // gap-files until stable.
    //
    // Use uni/prefer-array-some — has an autofix (replaces
    // `.filter(...).length > 0` with `.some(...)`) — alongside a
    // native syntax rule that just reports (no fix needed). After
    // --fix the file content on disk must reflect the rewrite for
    // both program and gap files.
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: false },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    // @ts-ignore — eslintPlugins is the in-process plugin form
    eslintPlugins: { uni: unicorn },
    plugins: ['@typescript-eslint'],
    rules: {
      // Native: forces the mixed (non-compatOnly) path.
      '@typescript-eslint/ban-ts-comment': 'error',
      // ESLint-plugin: HAS an autofix.
      'uni/prefer-array-some': 'error',
    },
  },
];`,
      'src/index.ts': `export const x = [1, 2].filter((n) => n > 0).length > 0;\n`,
      'scripts/build.ts': `export const y = [1, 2].filter((n) => n > 0).length > 0;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      await runRslint(['--fix'], tempDir);
      const srcAfter = await fs.readFile(
        path.join(tempDir, 'src/index.ts'),
        'utf8',
      );
      const gapAfter = await fs.readFile(
        path.join(tempDir, 'scripts/build.ts'),
        'utf8',
      );

      // Both files rewritten: filter→some chain replaced.
      for (const [, content] of [
        ['src', srcAfter] as const,
        ['gap', gapAfter] as const,
      ]) {
        expect(content).not.toContain('.filter(');
        expect(content).not.toContain('.length > 0');
        expect(content).toContain('.some(');
        // No partial / no-op fix: the original literal `[1, 2]` stays.
        expect(content).toContain('[1, 2]');
      }
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('(a) eslintPlugins on a pure-JS project with NO tsconfig — plugin rule still fires', async () => {
    // No tsconfig.json at all. A plugin rule must still dispatch to the
    // worker (oxc-parser builds the AST; no TS Program is required). This
    // pins that the compat path does not silently depend on a TS Program
    // existing — a pure-JS repo using an ESLint plugin works.
    const tempDir = await createTempDir({
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['**/*.js'],
    // @ts-ignore — eslintPlugins is the in-process plugin form
    eslintPlugins: { uni: unicorn },
    rules: { 'uni/no-null': 'error' },
  },
];`,
      'index.js': `export const x = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const ruleNames = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l).ruleName);
      expect(ruleNames).toContain('uni/no-null');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('(b) eslintPlugins on a universal entry (NO `files` field) — plugin rule still fires', async () => {
    // A config entry that omits `files` is a universal entry. With
    // eslintPlugins on it the plugin rule must still dispatch. Omitting
    // `files` also means the compat-only fast path (which requires a
    // `files` field) is NOT taken; the entry goes through the normal path
    // and the plugin rule must still fire.
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    // No files field (universal entry).
    languageOptions: {
      parserOptions: { projectService: false, project: ['./tsconfig.json'] },
    },
    // @ts-ignore — eslintPlugins is the in-process plugin form
    eslintPlugins: { uni: unicorn },
    rules: { 'uni/no-null': 'error' },
  },
];`,
      'src/index.ts': `export const x = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const ruleNames = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l).ruleName);
      expect(ruleNames).toContain('uni/no-null');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});
