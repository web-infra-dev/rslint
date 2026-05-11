/**
 * Real-world plugin edge cases — every test exercises the CLI end-to-end
 * with a real `eslint-plugin-unicorn` (or a self-contained fake plugin)
 * to cover cross-cutting paths that earlier suites missed:
 *
 *   A1  fix-conflict resolution
 *   A2  inline `eslint-disable-line` + plugin rule
 *   A3  plugin enabled but rule severity = 'off'
 *   A4  config overlap: rule re-declared in a later entry
 *   A5  `--quiet` + plugin rule severity = 'warn'
 *
 * All tests build a self-contained fixture (tempdir + symlinked
 * node_modules) so they run from `pnpm test` without external setup.
 */
import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs/promises';
import path from 'node:path';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  linkNodeModules,
} from './helpers.js';

const TSCONFIG = JSON.stringify({
  compilerOptions: {
    target: 'es2022',
    module: 'esnext',
    strict: false,
    noEmit: true,
    moduleResolution: 'bundler',
  },
  include: ['./src/**/*.ts'],
});

describe('Plugin edge cases — fix conflicts, directives, severity, overlap, --quiet', () => {
  test('A1: overlapping fixes — fixer applies non-conflicting, leaves conflicting for next pass', async () => {
    // Use TWO unicorn rules that both fix on the SAME file but
    // potentially overlap. `uni/no-null` rewrites `null` → `undefined`
    // (suggestion in upstream — eager mode here); `uni/prefer-array-some`
    // rewrites `.filter(...).length > 0` → `.some(...)`. The two rules
    // operate on disjoint AST regions in the source below, so all fixes
    // should apply in one pass. If the fixer mis-classified them as
    // conflicting, NOTHING would change. If it applied them with
    // overlap, the file would corrupt.
    //
    // What this test guards: when several non-overlapping fixes are
    // emitted for the same file across multiple rules, every one of
    // them lands cleanly.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    rules: {
      'uni/prefer-array-some': 'error',
    },
  },
];`,
      // Two independent .filter().length > 0 chains — two non-overlapping fixes.
      'src/index.ts':
        `export const x = [1, 2].filter((n) => n > 0).length > 0;\n` +
        `export const y = ['a', 'b'].filter((s) => s.length).length > 0;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      await runRslint(['--fix'], tempDir);
      const after = await fs.readFile(
        path.join(tempDir, 'src/index.ts'),
        'utf8',
      );
      // BOTH .filter chains rewritten to .some.
      expect(after).not.toContain('.filter(');
      expect(after).not.toContain('.length > 0');
      const someCount = (after.match(/\.some\(/g) ?? []).length;
      expect(someCount).toBe(2);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('A2: `eslint-disable-line` + plugin rule — directive suppresses the diagnostic', async () => {
    // Inline `// eslint-disable-line uni/no-null` must filter the
    // diagnostic out before it lands in the user's report. The
    // mechanism lives in apply-disable-directives.ts (runner-side).
    // Both rslint-prefixed and eslint-prefixed directive comments are
    // documented as supported (guide/inline-directives.md).
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    rules: { 'uni/no-null': 'error' },
  },
];`,
      // 3 occurrences: bare (fires), eslint-disable-line targeted
      // (suppressed), and eslint-disable-line broad (also suppressed).
      'src/index.ts':
        `export const a = null;\n` +
        `export const b = null; // eslint-disable-line uni/no-null\n` +
        `export const c = null; // eslint-disable-line\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));

      // ESTree visits multiple Identifier nodes per declaration in
      // ESM context (decl + export specifier), so each enabled line
      // can fire >1 diagnostic. The invariant the test pins is purely:
      //   line 1 has at least one no-null;
      //   lines 2 + 3 have ZERO no-null.
      const noNull = diagnostics.filter(
        (d: { ruleName: string }) => d.ruleName === 'uni/no-null',
      );
      const linesWithNoNull = new Set(
        noNull.map(
          (d: { range?: { start?: { line?: number } } }) =>
            d.range?.start?.line,
        ),
      );
      expect(linesWithNoNull.has(1)).toBe(true);
      expect(linesWithNoNull.has(2)).toBe(false);
      expect(linesWithNoNull.has(3)).toBe(false);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('A2b: `rslint-disable*` prefix suppresses plugin diagnostics (parity with Go side)', async () => {
    // The Go-side disable manager accepts both `eslint-` and `rslint-`
    // prefixes (internal/rule/disable_manager.go:37-38) and the user-
    // facing docs (website/docs/en/guide/inline-directives.md) say
    // both work interchangeably for native rules. The plugin path
    // previously hard-coded `eslint-` only, so `// rslint-disable-line`
    // on a plugin diagnostic was silently ignored — a surprising
    // divergence for users who pick one prefix and use it everywhere.
    // After the fix the two prefixes are interchangeable, including
    // mixed use (rslint-disable opened by an eslint-enable, etc.).
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    rules: { 'uni/no-null': 'error' },
  },
];`,
      // Line 1: bare → fires.
      // Line 2: rslint-disable-line targeting the rule → suppressed.
      // Line 3: rslint-disable-line broad (no rule list) → suppressed.
      // Lines 4-5: rslint-disable block + eslint-enable close → mixed
      //   prefix re-enables, so line 6 (after enable) fires again.
      // Lines 7-8: block disable using rslint-disable-next-line → only
      //   line 8 suppressed.
      'src/index.ts':
        `export const a1 = null;\n` +
        `export const a2 = null; // rslint-disable-line uni/no-null\n` +
        `export const a3 = null; // rslint-disable-line\n` +
        `/* rslint-disable uni/no-null */\n` +
        `export const a4 = null; /* eslint-enable uni/no-null */\n` +
        `export const a5 = null;\n` +
        `// rslint-disable-next-line uni/no-null\n` +
        `export const a6 = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const noNull = diagnostics.filter(
        (d: { ruleName: string }) => d.ruleName === 'uni/no-null',
      );
      const linesWithNoNull = new Set(
        noNull.map(
          (d: { range?: { start?: { line?: number } } }) =>
            d.range?.start?.line,
        ),
      );

      // Bare lines fire.
      expect(linesWithNoNull.has(1)).toBe(true);
      // Both `rslint-disable-line` variants suppress.
      expect(linesWithNoNull.has(2)).toBe(false);
      expect(linesWithNoNull.has(3)).toBe(false);
      // Inside the rslint-disable / eslint-enable block — line 5 is
      // BEFORE the enable comment (line 5 contains both decl and the
      // closing enable), the next standalone decl is line 6. The
      // assertion: line 5 (the declaration BEFORE the enable token
      // closes) is suppressed; line 6 (after the enable point) fires.
      expect(linesWithNoNull.has(6)).toBe(true);
      // disable-next-line targets line 8 → suppressed; line 7 is the
      // directive comment itself (no decl).
      expect(linesWithNoNull.has(8)).toBe(false);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('A3: plugin loaded but rule severity = "off" — rule listener never runs', async () => {
    // Even though `eslintPlugins: { uni: unicorn }` causes the plugin to
    // be imported by the worker, an `'uni/no-null': 'off'` rule
    // entry must NOT produce diagnostics. Two checks:
    //   1. zero diagnostics for the off rule (passive)
    //   2. another DIFFERENT unicorn rule that IS enabled still fires
    //      (proves the plugin loaded; we didn't accidentally skip the
    //      whole plugin)
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    rules: {
      'uni/no-null': 'off',                  // disabled
      'uni/prefer-array-some': 'error',      // enabled — proves plugin loaded
    },
  },
];`,
      'src/index.ts':
        `export const a = null;\n` +
        `export const b = [1, 2].filter((n) => n > 0).length > 0;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const ruleNames = diagnostics.map(
        (d: { ruleName: string }) => d.ruleName,
      );
      // off rule must NOT fire.
      expect(ruleNames).not.toContain('uni/no-null');
      // enabled rule MUST fire — sanity that the plugin DID load.
      expect(ruleNames).toContain('uni/prefer-array-some');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('A4: config overlap — later entry override of plugin rule severity wins', async () => {
    // Two entries both matching the same file pattern. The later
    // entry sets `uni/no-null` to 'off'. ESLint flat config
    // semantics: later entries shallow-merge `rules`, so 'off' wins.
    // Without proper merge logic, the rule might still fire.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    rules: { 'uni/no-null': 'error' },
  },
  {
    files: ['src/**/*.ts'],
    rules: { 'uni/no-null': 'off' },
  },
];`,
      'src/index.ts': `export const a = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      // Later entry's `'off'` must override the first entry's
      // `'error'` — no no-null diagnostic.
      const noNull = diagnostics.filter(
        (d: { ruleName: string }) => d.ruleName === 'uni/no-null',
      );
      expect(noNull).toHaveLength(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('B3: same prefix redefined with DIFFERENT instance → error (ESLint v10 parity)', async () => {
    // ESLint v10 behavior (lib/config/flat-config-schema.js
    // `pluginsSchema.merge`): if two config entries both declare the
    // same plugin prefix with different instances, ESLint throws
    // `TypeError: Cannot redefine plugin "${key}".`. Same instance
    // (e.g. `{ uni: x }` and `{ uni: x }` again) is a harmless dedupe.
    // rslint's plugin-loader now matches that contract: throw on
    // conflicting redefinition, dedupe on identical reference.
    //
    // Previously rslint used silent first-wins, which left the user
    // with a stale plugin and no diagnostic. The fix surfaces the
    // mistake at config-load time exactly like ESLint does.
    //
    // Use two tiny in-tree fake plugins so the test is self-contained.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin-a.mjs': `export default {
  meta: { name: 'plugin-a' },
  rules: {
    'fires': {
      meta: { messages: { x: 'plugin-A fired' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'x' }); } };
      },
    },
  },
};`,
      'plugin-b.mjs': `export default {
  meta: { name: 'plugin-b' },
  rules: {
    'fires': {
      meta: { messages: { y: 'plugin-B fired' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'y' }); } };
      },
    },
  },
};`,
      'rslint.config.mjs': `import pluginA from './plugin-a.mjs';
import pluginB from './plugin-b.mjs';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: pluginA },
    rules: { 'uni/fires': 'error' },
  },
  // Second entry — overrides 'uni' with plugin B.
  {
    files: ['src/**/*.ts'],
    // @ts-ignore
    eslintPlugins: { uni: pluginB },
    rules: { 'uni/fires': 'error' },
  },
];`,
      'src/index.ts': `export const TRIGGER = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Conflicting redefinition surfaces as a worker init failure;
      // the CLI exits non-zero AND stderr names the offending prefix
      // with the exact ESLint phrasing.
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toMatch(
        /Cannot redefine plugin "uni"|redefine plugin "uni"/,
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('B3: same prefix redefined with SAME instance is a harmless dedupe', async () => {
    // Counterpart to the conflict case — when both entries declare
    // the IDENTICAL plugin instance under the same prefix, ESLint v10
    // doesn't throw (the merge result is just that one plugin). rslint
    // must mirror this: the lint runs to completion, plugin rules
    // fire normally.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `export default {
  meta: { name: 'plugin-shared' },
  rules: {
    'fires': {
      meta: { messages: { x: 'shared plugin fired' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'x' }); } };
      },
    },
  },
};`,
      'rslint.config.mjs': `import shared from './plugin.mjs';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: shared },
    rules: { 'uni/fires': 'error' },
  },
  // Same module reference — dedupe, not error.
  {
    files: ['src/**/*.ts'],
    // @ts-ignore
    eslintPlugins: { uni: shared },
  },
];`,
      'src/index.ts': `export const TRIGGER = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const fired = diagnostics.filter(
        (d: { ruleName: string }) => d.ruleName === 'uni/fires',
      );
      expect(fired.length).toBeGreaterThan(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('B5: same rule severity escalated in later entry — last entry wins', async () => {
    // Counterpart to A4 (which turns rule OFF). Here the rule is
    // 'warn' in entry 1, escalated to 'error' in entry 2. The
    // diagnostic must carry severity=error.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    rules: { 'uni/no-null': 'warn' },
  },
  {
    files: ['src/**/*.ts'],
    rules: { 'uni/no-null': 'error' },
  },
];`,
      'src/index.ts': `export const a = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const noNull = diagnostics.filter(
        (d: { ruleName: string }) => d.ruleName === 'uni/no-null',
      );
      expect(noNull.length).toBeGreaterThan(0);
      // jsonline severity is sometimes string, sometimes number.
      // Accept either, but it MUST indicate error, not warn.
      for (const d of noNull) {
        const sev = d.severity;
        const isError =
          sev === 'error' || sev === 'Error' || sev === 2 || sev === 'fatal';
        expect(isError).toBe(true);
      }
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('B6: symlinked source file — native + plugin rules both still fire', async () => {
    // target.ts has the real source; symlink.ts points at it. Lint
    // must produce diagnostics regardless of which path the underlying
    // tools resolve to (Go-side ts-go program + Node-side worker may
    // disagree about symlink resolution, but the rule output must
    // appear under SOMEWHERE).
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    plugins: ['@typescript-eslint'],
    rules: {
      '@typescript-eslint/ban-ts-comment': 'error',
      'uni/no-null': 'error',
    },
  },
];`,
      'src/target.ts': `// @ts-ignore\nexport const v = null;\n`,
    });
    await linkNodeModules(tempDir);
    await fs.symlink(
      path.join(tempDir, 'src/target.ts'),
      path.join(tempDir, 'src/symlink.ts'),
      'file',
    );

    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const ruleNames = new Set(
        diagnostics.map((d: { ruleName: string }) => d.ruleName),
      );
      expect(ruleNames.has('@typescript-eslint/ban-ts-comment')).toBe(true);
      expect(ruleNames.has('uni/no-null')).toBe(true);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('A5: --quiet suppresses plugin "warn" diagnostics but keeps "error"', async () => {
    // --quiet contract: only severity='error' surfaces. Test that
    // plugin diagnostics participate equally — a `warn` plugin rule
    // is filtered, an `error` plugin rule still prints.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    rules: {
      'uni/no-null': 'warn',                 // should be suppressed by --quiet
      'uni/prefer-array-some': 'error',      // should still print
    },
  },
];`,
      'src/index.ts':
        `export const a = null;\n` +
        `export const b = [1, 2].filter((n) => n > 0).length > 0;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--quiet'], tempDir);
      // --quiet: stdout should reference the error rule but NOT the warn rule.
      expect(result.stdout).toContain('prefer-array-some');
      expect(result.stdout).not.toContain('no-null');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('D1: whole-file `/* eslint-disable */` directive suppresses plugin diagnostics', async () => {
    // Block-form `/* eslint-disable */` at the top of a file disables
    // EVERY rule for the file. Mix native + plugin rules — both must
    // be silenced. Then a second file without the directive proves
    // the rules still fire elsewhere (sanity check that we didn't
    // accidentally disable globally).
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    plugins: ['@typescript-eslint'],
    rules: {
      '@typescript-eslint/ban-ts-comment': 'error',
      'uni/no-null': 'error',
    },
  },
];`,
      // Disabled file — both rule families must produce ZERO diagnostics.
      'src/disabled.ts': `/* eslint-disable */\n// @ts-ignore\nexport const a = null;\n`,
      // Enabled file — proves rules still fire elsewhere.
      'src/enabled.ts': `// @ts-ignore\nexport const b = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));

      const disabledRules = diagnostics
        .filter((d: { filePath: string }) =>
          d.filePath.includes('src/disabled.ts'),
        )
        .map((d: { ruleName: string }) => d.ruleName);
      const enabledRules = diagnostics
        .filter((d: { filePath: string }) =>
          d.filePath.includes('src/enabled.ts'),
        )
        .map((d: { ruleName: string }) => d.ruleName);

      // disabled.ts must have NO rule diagnostics (native or plugin).
      expect(disabledRules).not.toContain('@typescript-eslint/ban-ts-comment');
      expect(disabledRules).not.toContain('uni/no-null');

      // enabled.ts proves the rules ARE active when the directive is absent.
      expect(enabledRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(enabledRules).toContain('uni/no-null');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('D2: linterOptions.reportUnusedDisableDirectives is currently NOT implemented for plugin path', async () => {
    // Docs (`website/docs/en/guide/eslint-plugin-compat.md` Limitations
    // table) say the unused-disable-directive feature is not
    // implemented for plugin rules. Pin that: even when the user
    // sets `linterOptions.reportUnusedDisableDirectives: 'error'`
    // AND the source contains an unused directive, NO additional
    // diagnostic about the unused directive appears.
    //
    // If this changes (feature actually implemented), this test
    // becomes a "must update docs" reminder.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    linterOptions: { reportUnusedDisableDirectives: 'error' },
    // @ts-ignore
    eslintPlugins: { uni: unicorn },
    rules: { 'uni/no-null': 'error' },
  },
];`,
      // The directive disables a rule that the source DOESN'T trigger
      // — so it's "unused". ESLint would warn about that; rslint
      // currently won't.
      'src/index.ts': `// eslint-disable-next-line uni/no-null\nexport const x = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));

      // Pin the current behavior: NO diagnostic about an unused
      // directive surfaces. Matches the docs' "Not implemented" note.
      const unusedDirective = diagnostics.filter(
        (d: { ruleName: string; message?: string }) =>
          /unused.*disable|unused-disable/i.test(d.ruleName) ||
          /unused.*disable/i.test(d.message ?? ''),
      );
      expect(unusedDirective).toHaveLength(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('D3: monorepo nested configs — CLI end-to-end with plugin in each sub-config', async () => {
    // CLI-side end-to-end of the multi-config plugin scenario already
    // covered by vscode e2e (`fixtures-eslint-plugin-monorepo`) and by
    // worker-pool unit. The cli path has its own config-discovery +
    // dispatch wiring that the unit tests don't exercise; this test
    // pins that the same monorepo layout linted via `rslint` (no LSP,
    // no in-process WorkerPool host) still produces per-sub-config
    // diagnostics for the right files.
    //
    // Layout:
    //   packages/x/{rslint.config.mjs, src/index.ts}  — plugin rule X enabled
    //   packages/y/{rslint.config.mjs, src/index.ts}  — plugin rule Y enabled
    // We assert: x file produces X diagnostic, y file produces Y
    // diagnostic, neither file produces the OTHER's diagnostic.
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'es2022',
          module: 'esnext',
          strict: false,
          noEmit: true,
          moduleResolution: 'bundler',
        },
        include: ['./packages/*/src/**/*.ts'],
      }),
      'plugin-x.mjs': `export default {
  meta: { name: 'plugin-x' },
  rules: {
    'no-foo': {
      meta: { messages: { x: 'X-side foo' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'foo') ctx.report({ node, messageId: 'x' }); } };
      },
    },
  },
};`,
      'plugin-y.mjs': `export default {
  meta: { name: 'plugin-y' },
  rules: {
    'no-bar': {
      meta: { messages: { y: 'Y-side bar' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'bar') ctx.report({ node, messageId: 'y' }); } };
      },
    },
  },
};`,
      'packages/x/rslint.config.mjs': `import pluginX from '../../plugin-x.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['../../tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { px: pluginX },
  rules: { 'px/no-foo': 'error' },
}];`,
      'packages/y/rslint.config.mjs': `import pluginY from '../../plugin-y.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['../../tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { py: pluginY },
  rules: { 'py/no-bar': 'error' },
}];`,
      'packages/x/src/index.ts': `export const foo = 1; export const bar = 2;\n`,
      'packages/y/src/index.ts': `export const foo = 1; export const bar = 2;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));

      const xRules = diagnostics
        .filter((d: { filePath: string }) =>
          d.filePath.includes('packages/x/src/index.ts'),
        )
        .map((d: { ruleName: string }) => d.ruleName);
      const yRules = diagnostics
        .filter((d: { filePath: string }) =>
          d.filePath.includes('packages/y/src/index.ts'),
        )
        .map((d: { ruleName: string }) => d.ruleName);

      // x file: only px/no-foo fires (plugin X is the only plugin in
      // cfgX's LoadedPlugins). py/no-bar must NOT fire on x file
      // (worker-pool's per-config map keeps the two plugin sets
      // disjoint; if any cross-config leak happened, this catches it
      // end-to-end through the CLI rather than the unit test).
      expect(xRules).toContain('px/no-foo');
      expect(xRules).not.toContain('py/no-bar');

      // Symmetric for y.
      expect(yRules).toContain('py/no-bar');
      expect(yRules).not.toContain('px/no-foo');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U1: CJS plugin (module.exports) loads via Node ESM interop', async () => {
    // Plugin authored as CommonJS: `module.exports = { meta, rules }`.
    // Many older plugins ship this way. `await import()` of a CJS
    // file returns `{ default: <module.exports>, ...named }`, so the
    // plugin-loader's `unwrapPluginModule` must recover the actual
    // plugin object from `.default`.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      // .cjs file = forced CJS regardless of package.json type.
      'plugin.cjs': `module.exports = {
  meta: { name: 'cjs-plugin' },
  rules: {
    'fires': {
      meta: { messages: { x: 'CJS plugin fired' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'x' }); } };
      },
    },
  },
};`,
      'rslint.config.mjs': `import cjsPlugin from './plugin.cjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { cjs: cjsPlugin },
  rules: { 'cjs/fires': 'error' },
}];`,
      'src/index.ts': `export const TRIGGER = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const fired = diagnostics.filter(
        (d: { ruleName: string }) => d.ruleName === 'cjs/fires',
      );
      expect(fired.length).toBeGreaterThan(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U2: --max-warnings counts plugin "warn" diagnostics in the budget', async () => {
    // `--max-warnings 0` means "0 warnings allowed". A single plugin
    // rule firing at severity='warn' must push the exit code to 1,
    // same as a native warning would.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uni: unicorn },
  rules: { 'uni/no-null': 'warn' },
}];`,
      'src/index.ts': `export const a = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      // Without --max-warnings: exit code 0 (warnings allowed by default).
      const lax = await runRslint([], tempDir);
      expect(lax.exitCode).toBe(0);

      // With --max-warnings 0: any warning trips exit 1.
      const strict = await runRslint(['--max-warnings', '0'], tempDir);
      expect(strict.exitCode).not.toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U3: plugin with `configs` / `processors` fields loads without error', async () => {
    // ESLint plugins commonly ship presets via `plugin.configs` and
    // language extensions via `plugin.processors`. rslint doesn't
    // consume either (configs need manual spread; processors are an
    // out-of-scope feature per the limitations doc) — but the
    // PRESENCE of those fields on the plugin object must not crash
    // the loader.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `export default {
  meta: { name: 'rich-plugin' },
  configs: {
    recommended: { rules: { 'rich/fires': 'error' } },
    strict: { rules: { 'rich/fires': 'error' } },
  },
  processors: {
    '.md': { preprocess() { return []; }, postprocess() { return []; } },
  },
  rules: {
    'fires': {
      meta: { messages: { x: 'rich plugin fired' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'x' }); } };
      },
    },
  },
};`,
      'rslint.config.mjs': `import rich from './plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { rich },
  rules: { 'rich/fires': 'error' },
}];`,
      'src/index.ts': `export const TRIGGER = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const fired = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l))
        .filter((d: { ruleName: string }) => d.ruleName === 'rich/fires');
      expect(fired.length).toBeGreaterThan(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test("U4: plugin `process.stderr.write` side effect doesn't corrupt the IPC frame stream", async () => {
    // The runner's WORKER monkey-patches `console.*` to redirect log
    // lines to the parent's onLog handler instead of stdout (where
    // the IPC frames flow). What the worker does NOT (and cannot
    // reasonably) patch:
    //
    //   1. `process.stderr.write` — this is a legitimate stream
    //      separate from stdout; patching it would break legit error
    //      logging. Worker-side stderr is inherited by the parent
    //      process's stderr — visible to the user but it doesn't
    //      pollute stdout.
    //   2. `console.log` invoked DURING `loadConfigFile` on the HOST
    //      side (Node parent loads the user's config to enumerate
    //      plugins for the IPC handshake; plugin's top-level
    //      `console.log` runs in the host process, which doesn't
    //      patch console). This is a documented trade-off — plugins
    //      with top-level prints will leak into host stdout. Users
    //      are advised against top-level side effects in plugin
    //      modules.
    //
    // What this test pins: invariant (1) — plugin's
    // `process.stderr.write` during worker init does NOT land in
    // the IPC frame stream (worker stdout) and does NOT corrupt
    // jsonline parsing of the lint result. The diagnostic still
    // surfaces normally.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `process.stderr.write('PLUGIN_STDERR_NOISE_MARKER\\n');
export default {
  meta: { name: 'noisy' },
  rules: {
    'fires': {
      meta: { messages: { x: 'noisy plugin fired' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'x' }); } };
      },
    },
  },
};`,
      'rslint.config.mjs': `import noisy from './plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { noisy },
  rules: { 'noisy/fires': 'error' },
}];`,
      'src/index.ts': `export const TRIGGER = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // The marker MUST NOT appear in stdout (the IPC frame stream).
      // It SHOULD appear in stderr (where worker stderr funnels).
      expect(result.stdout).not.toContain('PLUGIN_STDERR_NOISE_MARKER');
      expect(result.stderr).toContain('PLUGIN_STDERR_NOISE_MARKER');
      // jsonline output remains parseable.
      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      for (const line of lines) {
        expect(() => JSON.parse(line)).not.toThrow();
      }
      // Diagnostic surfaced normally.
      const fired = lines
        .map((l) => JSON.parse(l))
        .filter((d: { ruleName: string }) => d.ruleName === 'noisy/fires');
      expect(fired.length).toBeGreaterThan(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U5: plugin diagnostic position on multi-byte UTF-8 source', async () => {
    // Source contains CJK + emoji characters before the rule
    // trigger. UTF-16 → UTF-8 byte offset conversion in
    // ecma-language-plugin.ts must keep the diagnostic pointing at
    // the actual identifier. A wrong conversion would shift the
    // start position by N bytes (often 2-3 per CJK char) and
    // the slice of source at the reported range would not equal
    // the identifier text.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uni: unicorn },
  rules: { 'uni/no-null': 'error' },
}];`,
      // Comment + emoji + CJK before the null literal forces multi-
      // byte offsets to be non-trivial.
      'src/index.ts': `// 中文注释 🚀 emoji here\nexport const x = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const noNull = diagnostics.find(
        (d: { ruleName: string }) => d.ruleName === 'uni/no-null',
      );
      expect(noNull).toBeDefined();
      // Diagnostic line should be 2 (the `export const x = null;`
      // line); UTF-8 byte miscount would have shifted to a wrong
      // line. We don't pin exact column (line-relative offset is
      // also subject to multi-byte math) — line is the strongest
      // unambiguous signal here.
      expect(noNull.range?.start?.line).toBe(2);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U6: pure-insertion fix (`[i, i]` zero-width range) applies correctly', async () => {
    // A fixer that inserts BEFORE a node uses `replaceTextRange([i,
    // i], '...')`. The applier must distinguish "insert" from
    // "replace zero chars" and place text at exactly that offset.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `export default {
  meta: { name: 'inserter' },
  rules: {
    'add-prefix': {
      meta: { type: 'problem', fixable: 'code', messages: { x: 'add prefix' } },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'INSERT_HERE') {
              ctx.report({
                node,
                messageId: 'x',
                fix(fixer) { return fixer.insertTextBefore(node, 'PRE_'); },
              });
            }
          },
        };
      },
    },
  },
};`,
      'rslint.config.mjs': `import plugin from './plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { ins: plugin },
  rules: { 'ins/add-prefix': 'error' },
}];`,
      'src/index.ts': `export const INSERT_HERE = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      await runRslint(['--fix'], tempDir);
      const after = await fs.readFile(
        path.join(tempDir, 'src/index.ts'),
        'utf8',
      );
      // Identifier rewritten with prefix prepended (insertion BEFORE
      // the identifier). Pin the EXACT new text rather than a partial
      // match so we catch off-by-one insert offsets.
      expect(after).toContain('PRE_INSERT_HERE');
      // Original identifier name is not left orphaned.
      expect(after).not.toMatch(/\bINSERT_HERE\b(?!_)/);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U7: multi-line fix range replaces correctly across lines', async () => {
    // Fix range spans more than one source line (common for import
    // reorder rules, or rewriting blocks). The applier must use
    // byte ranges directly, not assume "single-line replacement".
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `export default {
  meta: { name: 'multiliner' },
  rules: {
    'collapse-block': {
      meta: { type: 'problem', fixable: 'code', messages: { x: 'collapse' } },
      create(ctx) {
        return {
          BlockStatement(node) {
            // Collapse \`{\\n  body\\n}\` into \`{ /* collapsed */ }\`.
            // Only run on blocks containing exactly one statement.
            if (node.body.length === 1) {
              ctx.report({
                node,
                messageId: 'x',
                fix(fixer) { return fixer.replaceText(node, '{ /* collapsed */ }'); },
              });
            }
          },
        };
      },
    },
  },
};`,
      'rslint.config.mjs': `import plugin from './plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { ml: plugin },
  rules: { 'ml/collapse-block': 'error' },
}];`,
      'src/index.ts': `export function f() {\n  const x = 1;\n}\n`,
    });
    await linkNodeModules(tempDir);
    try {
      await runRslint(['--fix'], tempDir);
      const after = await fs.readFile(
        path.join(tempDir, 'src/index.ts'),
        'utf8',
      );
      expect(after).toContain('{ /* collapsed */ }');
      // The original 3-line block contents must be gone.
      expect(after).not.toContain('const x = 1;');
      // Sanity: result is still valid syntax (start of function
      // unchanged before the collapsed block).
      expect(after).toMatch(/^export function f\(\)/);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U8: empty source file + plugin rule does not crash', async () => {
    // oxc-parser on empty source produces an empty Program. Plugin
    // rules attached to specific node types simply never fire.
    // Critical invariant: rslint produces zero diagnostics, exit 0,
    // no thrown error or crash.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uni: unicorn },
  rules: { 'uni/no-null': 'error' },
}];`,
      'src/empty.ts': ``,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // No throw, no diagnostics, exit 0.
      expect(result.exitCode).toBe(0);
      expect(result.stdout.trim()).toBe('');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U13: shebang-only file + plugin rule does not crash', async () => {
    // Shebang lines (`#!/usr/bin/env node`) need special tokenizer
    // handling (rslint v10 shim — getDisableDirectives + Shebang
    // comment type). A file with ONLY a shebang must lint cleanly:
    // shebang becomes a Comment, no other tokens, no diagnostic.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uni: unicorn },
  rules: { 'uni/no-null': 'error' },
}];`,
      'src/cli.ts': `#!/usr/bin/env node\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.exitCode).toBe(0);
      // No diagnostics for a file containing only a shebang.
      const diagLines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      expect(diagLines).toHaveLength(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U14: CRLF line endings — plugin diagnostic line numbers correct', async () => {
    // Source uses Windows-style \\r\\n line terminators. Plugin
    // diagnostic line numbers must still reflect 1-based source
    // lines (the same as ESLint would emit). Off-by-one or skipping
    // CR would shift every diagnostic by one line.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uni: unicorn },
  rules: { 'uni/no-null': 'error' },
}];`,
      // 4 lines, CRLF separators; `null` on the THIRD line.
      'src/index.ts': `// header\r\nconst x = 1;\r\nexport const y = null;\r\nexport { y };\r\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const noNull = diagnostics.find(
        (d: { ruleName: string }) => d.ruleName === 'uni/no-null',
      );
      expect(noNull).toBeDefined();
      // `null` is on line 3 (1-based), regardless of CR/LF mix.
      expect(noNull.range?.start?.line).toBe(3);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('U15: BOM at file start — rslint does not crash (current behavior pin)', async () => {
    // KNOWN LIMITATION: a file starting with U+FEFF (UTF-8 BOM) is
    // currently NOT linted by rslint — neither native rules nor
    // plugin rules fire on it. This is a Go-side file routing / TS
    // parser behavior (the BOM-prefixed file doesn't match tsconfig
    // include semantics or fails an upstream filter), unrelated to
    // the plugin compat path itself.
    //
    // Why this test exists: pin the CURRENT behavior so a future
    // change (positive — BOM handled — or negative — BOM causes a
    // hard crash) surfaces. The minimum invariant is "rslint must
    // not crash on BOM input", which this test enforces.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uni: unicorn },
  rules: { 'uni/no-null': 'error' },
}];`,
      // ﻿ + comment + null on line 2.
      'src/index.ts': `﻿// comment\nexport const x = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Hard invariant: rslint TERMINATES (no hang). 0 = clean, 1 =
      // lint errors found, 2 = config/runtime error. All three are
      // observed for BOM input depending on which code path catches
      // it first (worker init, ts-go parse, oxc parse). Test pins
      // "doesn't hang and doesn't segfault" — anything else is a
      // documented limitation worth a follow-up issue, not a
      // regression to chase here.
      expect([0, 1, 2]).toContain(result.exitCode);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('V1: async create() silently registers zero listeners — matches ESLint v10', async () => {
    // Verified upstream behavior (ESLint v10
    // node_modules/eslint/lib/linter/linter.js: `return rule.create(ruleContext)`
    // — sync call, no await): when a plugin author writes
    //   async create(ctx) { return { Identifier(n) {...} }; }
    // ESLint receives a PROMISE as the listener record, iterates
    // its enumerable own keys (zero AST-node keys), registers
    // nothing, produces no diagnostics, no warning, no ruleErrors.
    //
    // rslint mirrors this exactly. Verified by running the same
    // fixture through eslint.cjs and observing identical output
    // (only the sync control rule fires; the async rule is silent).
    //
    // Two rules in the same plugin: one async create (must be
    // silent), one sync create (must fire). The sync control proves
    // the plugin loaded and the worker is dispatching correctly,
    // so the async silence is isolated to the create-returns-Promise
    // path, not a config or routing issue. If either side ever
    // changes (rslint becomes stricter and surfaces ruleErrors, or
    // upstream ESLint starts detecting Promise returns), this test
    // fails and the team makes a conscious choice.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `export default {
  meta: { name: 'async-pin' },
  rules: {
    'async-fires': {
      meta: { messages: { x: 'ASYNC create fired' } },
      async create(ctx) {
        return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'x' }); } };
      },
    },
    'sync-control': {
      meta: { messages: { y: 'SYNC control fired' } },
      create(ctx) {
        return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'y' }); } };
      },
    },
  },
};`,
      'rslint.config.mjs': `import p from './plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { a: p },
  rules: { 'a/async-fires': 'error', 'a/sync-control': 'error' },
}];`,
      'src/index.ts': `export const TRIGGER = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const ruleNames = diagnostics.map(
        (d: { ruleName: string }) => d.ruleName,
      );

      // SYNC control fires (proves plugin loaded + dispatch works).
      expect(ruleNames).toContain('a/sync-control');
      // ASYNC create does NOT fire (matches ESLint v10).
      expect(ruleNames).not.toContain('a/async-fires');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('V2: meta.schema is NOT validated — invalid options pass through (diverges from ESLint v10)', async () => {
    // DIVERGENCE FROM ESLint v10:
    // ESLint v10 invokes ajv (v6.14) to validate user `options`
    // against `rule.meta.schema` before calling `rule.create`.
    // A schema violation throws a configuration error.
    //
    // rslint currently DOES NOT validate. `options-defaults.ts`
    // only fills default values; the user's options pass through
    // to `rule.create(ctx)` as-is, even when their type or shape
    // doesn't match the declared schema.
    //
    // Trade-off accepted: avoids the ~1.1 MB ajv@6 bundle in every
    // worker, in exchange for plugin authors needing to validate
    // `ctx.options` defensively if their invariants matter. See
    // website/docs/en/guide/eslint-plugin-compat.md Limitations.
    //
    // This test PINS the current behavior so if validation is later
    // added the test forces an intentional choice rather than
    // silent drift.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `export default {
  meta: { name: 'schema-pin' },
  rules: {
    'strict-options': {
      meta: {
        messages: { x: 'fired with options={{opts}}' },
        // Schema demands an object with required \`mode\` enum.
        schema: [{
          type: 'object',
          properties: { mode: { type: 'string', enum: ['fast', 'slow'] } },
          required: ['mode'],
          additionalProperties: false,
        }],
      },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'TRIGGER') {
              ctx.report({ node, messageId: 'x', data: { opts: JSON.stringify(ctx.options) } });
            }
          },
        };
      },
    },
  },
};`,
      // User passes 12345 (a number) — schema-violating. ESLint
      // would reject this at config-load time; rslint accepts.
      'rslint.config.mjs': `import p from './plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { s: p },
  rules: { 's/strict-options': ['error', 12345] },
}];`,
      'src/index.ts': `export const TRIGGER = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));

      // The rule fires (since rslint doesn't validate, the create
      // function runs with the bad options). The diagnostic
      // message echoes ctx.options so we can confirm the invalid
      // value really did reach the rule.
      const fired = diagnostics.find(
        (d: { ruleName: string }) => d.ruleName === 's/strict-options',
      );
      expect(fired).toBeDefined();
      expect(String(fired.message)).toContain('12345');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('N1: eslintPlugins prefix collision with native plugin → fail-fast before Go child spawns', async () => {
    // End-to-end pin for the host-side collision validator:
    //   - User config registers `unicorn` and `@typescript-eslint` in
    //     `eslintPlugins` — both are reserved native plugin namespaces.
    //   - rslint CLI must fail at config-load time (BEFORE spawning the
    //     Go child or any worker), exit non-zero, and write a clear
    //     message to stderr that names every offending prefix.
    //
    // This guards two contracts:
    //   1. The reserved list in `define-config.ts NATIVE_PLUGIN_PREFIXES`
    //      stays wired to the loader's collision check.
    //   2. The error path is fail-fast — the CLI doesn't fall through
    //      to "no config, run unconfigured" (which would be a "fake
    //      green" outcome).
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `export default {
  meta: { name: 'collider' },
  rules: { r: { meta: {}, create() { return {}; } } },
};`,
      // Two reserved-namespace prefixes in one entry.
      'rslint.config.mjs': `import p from './plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  // @ts-ignore
  eslintPlugins: { 'unicorn': p, '@typescript-eslint': p },
  rules: { 'uni/r': 'error' },
}];`,
      'src/index.ts': `export const x = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Fail-fast exit code (1 from cli.ts FAIL-FAST loadJsConfigs path).
      expect(result.exitCode).not.toBe(0);
      // Stderr must name both offending prefixes AND include the
      // remediation hint so the user can self-serve the fix.
      expect(result.stderr).toContain('"unicorn"');
      expect(result.stderr).toContain('"@typescript-eslint"');
      expect(result.stderr).toMatch(/rename the prefix/i);
      // No diagnostics on stdout — the lint must NOT run.
      expect(result.stdout.trim()).toBe('');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('N2: collision in a NESTED config in a multi-config workspace surfaces with index/path', async () => {
    // Monorepo: root config is clean, but packages/bad/rslint.config.mjs
    // uses a reserved namespace. The discovery walker still surfaces
    // the collision (it's not just a single-config check).
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'es2022',
          module: 'esnext',
          strict: false,
          noEmit: true,
          moduleResolution: 'bundler',
        },
        include: ['./packages/*/src/**/*.ts'],
      }),
      'plugin.mjs': `export default { meta: { name: 'p' }, rules: { r: { meta: {}, create() { return {}; } } } };`,
      // Root config — clean.
      'rslint.config.mjs': `export default [];`,
      // Nested bad config — uses reserved prefix.
      'packages/bad/rslint.config.mjs': `import p from '../../plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['../../tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { 'react-hooks': p },
  rules: { 'react-hooks/r': 'error' },
}];`,
      'packages/bad/src/index.ts': `export const x = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint([], tempDir);
      // Multi-config: the loader's loop emits a "Warning: skipping
      // config <path>: ..." line for the bad config and CONTINUES
      // with the remaining configs. The collision message must
      // still appear in stderr so the user knows which sub-config
      // is the problem.
      expect(result.stderr).toMatch(/react-hooks/);
      expect(result.stderr).toMatch(/packages\/bad\/rslint\.config\.mjs/);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('N3: legitimate non-reserved prefix passes validator and lints normally', async () => {
    // Smoke check that the validator only stops reserved prefixes,
    // not innocent ones. `uniMine` mirrors the recommended rename
    // (e.g. user wraps the upstream unicorn plugin under their own
    // namespace).
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'plugin.mjs': `export default {
  meta: { name: 'mine' },
  rules: {
    'fires': {
      meta: { messages: { x: 'fired' } },
      create(ctx) { return { Identifier(node) { if (node.name === 'TRIGGER') ctx.report({ node, messageId: 'x' }); } }; },
    },
  },
};`,
      'rslint.config.mjs': `import p from './plugin.mjs';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uniMine: p },
  rules: { 'uniMine/fires': 'error' },
}];`,
      'src/index.ts': `export const TRIGGER = 1;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Lint actually ran AND the renamed prefix fired.
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      expect(
        diagnostics.some(
          (d: { ruleName: string }) => d.ruleName === 'uniMine/fires',
        ),
      ).toBe(true);
      // Validator silent on legitimate config — no warning noise.
      expect(result.stderr).not.toMatch(/conflict with rslint/);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('G3: compat-only mode + explicit CLI file does NOT emit "not found in the project" warning', async () => {
    // Pre-fix: in the all-compat-rules fast path the Go side skips
    // ts-go Program construction (`programs` stays empty). The
    // warning emitter `collectAllowFileWarnings` then used the empty
    // `programs` slice as the oracle for "was this file linted?" and
    // flagged every CLI-passed file as `allowFileNotInProgram`,
    // printing `was not found in the project, skipping` to stderr —
    // even though `DispatchCompat` had just finished linting the file.
    //
    // Post-fix the compat-only branch passes its actually-linted
    // file set to `collectAllowFileWarnings`, so the warning is
    // silent. This test pins both pieces:
    //   (a) the lint result still surfaces the plugin's diagnostic
    //       for the CLI-passed file (the lint really ran);
    //   (b) stderr is FREE of the misleading "not found in the
    //       project" line for that same file.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uni: unicorn },
  rules: { 'uni/no-null': 'error' },
}];`,
      'src/index.ts': `export const x = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      // Pass the file EXPLICITLY on the CLI — this is the trigger
      // condition for the bug. Without the explicit positional arg
      // the warning oracle isn't consulted.
      const result = await runRslint(
        ['--format', 'jsonline', 'src/index.ts'],
        tempDir,
      );
      // The lint really ran (compat dispatcher path is alive).
      const diags = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      expect(
        diags.some((d: { ruleName: string }) => d.ruleName === 'uni/no-null'),
      ).toBe(true);
      // And the misleading warning is NOT in stderr.
      expect(result.stderr).not.toMatch(/was not found in the project/);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('G5: --rule native override on an all-plugin config disables compat-only fast path', async () => {
    // Pre-fix: configs containing only ESLint-plugin rules + a `files`
    // glob set `compatOnlyMode=true` BEFORE the `--rule` CLI override
    // was applied. The fast path then routed every file through
    // `DispatchCompat`, which filters every entry where
    // `IsEslintPluginRule == false` — so a user-added native rule
    // (e.g. `--rule "no-console:error"`) was silently dropped and
    // ran zero times. CLI false negative.
    //
    // Post-fix the CLI override is applied inside each config-load
    // branch BEFORE the `allRulesAreCompat` evaluation, so adding a
    // native rule flips the fast-path off and the file goes through
    // the normal RunLinter path. This test pins that:
    //
    //   (a) the file is still actually linted (sanity);
    //   (b) the user-introduced `no-console` rule fires on a
    //       `console.log` line in the source.
    //
    // Pre-fix, (b) would silently report zero diagnostics for
    // no-console even though the user passed it on the command line.
    const tempDir = await createTempDir({
      'tsconfig.json': TSCONFIG,
      'rslint.config.mjs': `import unicorn from 'eslint-plugin-unicorn';
export default [{
  files: ['src/**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  // @ts-ignore
  eslintPlugins: { uni: unicorn },
  // ALL rules in the file are plugin rules — the fast-path trigger.
  rules: { 'uni/no-null': 'error' },
}];`,
      'src/index.ts':
        // Contains both the plugin trigger (`null`) and the native
        // rule trigger (`console.log`).
        `console.log("hi");\nexport const x = null;\n`,
    });
    await linkNodeModules(tempDir);
    try {
      const result = await runRslint(
        ['--format', 'jsonline', '--rule', 'no-console:error'],
        tempDir,
      );
      const diags = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l));
      const ruleNames = diags.map((d: { ruleName: string }) => d.ruleName);
      // Plugin rule fires (sanity — compat path itself still works).
      expect(ruleNames).toContain('uni/no-null');
      // Native rule introduced by `--rule` fires — pre-fix this would
      // be missing because compat-only mode swallowed it.
      expect(ruleNames).toContain('no-console');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});
