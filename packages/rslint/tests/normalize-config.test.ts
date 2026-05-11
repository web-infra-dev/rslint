import { describe, test, expect } from '@rstest/core';
import { normalizeConfig } from '../src/config-loader.js';
import { NATIVE_PLUGIN_PREFIXES } from '../src/define-config.js';

// normalizeConfig pivots on two valid shapes for `entry.plugins`:
//
//   - string[]   → JSON-config plugin allowlist (no live instance)
//   - object     → ESLint flat-config `{ prefix: pluginInstance }` (live)
//
// The object form must fold into `eslintPlugins` so plugin rules actually
// load on the worker side (functions don't survive the JSON wire). The
// previous behavior shipped the raw object through, where Go interpreted
// the missing string-array as zero allowed plugins and any plugin rules
// silently no-op'd.
//
// Tests below cover:
//   - object-form `plugins` becomes both the allowlist AND eslintPlugins
//   - array-form `plugins` stays as-is and produces no eslintPlugins
//   - explicit `eslintPlugins` wins over plugins-object extraction
//   - malformed `plugins` raises a loud error (no silent dropping)

const fakeUnicorn = {
  meta: { name: 'eslint-plugin-unicorn', version: '1.0.0' },
  rules: { 'no-null': { create: () => ({}) } },
};

describe('normalizeConfig — plugins folding', () => {
  test('object-form plugins folds into eslintPlugins AND allowlist', () => {
    const out = normalizeConfig([{ plugins: { uni: fakeUnicorn }, rules: {} }]);
    expect(out).toHaveLength(1);
    const entry = out[0];
    // Allowlist: keys of the object become the string array.
    expect(entry.plugins).toEqual(['uni']);
    // eslintPlugins: the live plugin gets extracted as a lean wire entry.
    expect(Array.isArray(entry.eslintPlugins)).toBe(true);
    const eps = entry.eslintPlugins as Array<{
      prefix: string;
      ruleNames: string[];
    }>;
    expect(eps).toHaveLength(1);
    expect(eps[0].prefix).toBe('uni');
    expect(eps[0].ruleNames).toEqual(['no-null']);
  });

  test('array-form plugins stays as the allowlist with no eslintPlugins', () => {
    const out = normalizeConfig([
      { plugins: ['@typescript-eslint', 'react'], rules: {} },
    ]);
    expect(out).toHaveLength(1);
    expect(out[0].plugins).toEqual(['@typescript-eslint', 'react']);
    expect(out[0].eslintPlugins).toBeUndefined();
  });

  test('explicit eslintPlugins takes precedence over plugins-object extraction', () => {
    // Both fields present. The explicit eslintPlugins (an alternate
    // plugin instance) is what should be extracted; the plugins object's
    // keys still feed the allowlist.
    const otherPlugin = {
      meta: { name: 'eslint-plugin-explicit', version: '9.9.9' },
      rules: { foo: { create: () => ({}) } },
    };
    const out = normalizeConfig([
      {
        plugins: { uni: fakeUnicorn },
        eslintPlugins: { uni: otherPlugin },
        rules: {},
      },
    ]);
    expect(out[0].plugins).toEqual(['uni']);
    const eps = out[0].eslintPlugins as Array<{
      prefix: string;
      ruleNames: string[];
    }>;
    expect(eps).toHaveLength(1);
    expect(eps[0].prefix).toBe('uni');
    // ruleNames comes from the EXPLICIT plugin (otherPlugin), not uni.
    expect(eps[0].ruleNames).toEqual(['foo']);
  });

  test('plugins as a non-string-array element rejects loudly', () => {
    expect(() =>
      normalizeConfig([
        { plugins: ['unicorn', { not: 'a string' }], rules: {} },
      ]),
    ).toThrow(/plugins\[1\]/);
  });

  test('plugins as a number is rejected', () => {
    expect(() => normalizeConfig([{ plugins: 42, rules: {} }])).toThrow(
      /plugins.*must be/,
    );
  });

  test('no plugins field leaves plugins undefined and no eslintPlugins', () => {
    const out = normalizeConfig([{ rules: {} }]);
    expect(out[0].plugins).toBeUndefined();
    expect(out[0].eslintPlugins).toBeUndefined();
  });

  test('non-object config entries warn via stderr (not console.warn)', () => {
    // M10: warnings for malformed config entries used to flow through
    // `console.warn`, which under VS Code's LSP host lands in the
    // dev console instead of the Rslint output channel — invisible to
    // users debugging their config. The fix routes the message
    // through `process.stderr.write` so it honors the host's stderr
    // wiring (LSP routes stderr → Rslint output channel; CLI prints
    // directly to the terminal stderr).
    const originalWarn = console.warn;
    let consoleWarnCalled = false;
    console.warn = () => {
      consoleWarnCalled = true;
    };
    const originalWrite = process.stderr.write.bind(process.stderr);
    let stderrBuffer = '';
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (process.stderr as any).write = (chunk: string | Uint8Array): boolean => {
      stderrBuffer +=
        typeof chunk === 'string' ? chunk : Buffer.from(chunk).toString();
      return true;
    };
    try {
      // Three entries: a valid one (sandwiched between two malformed),
      // a null, and a number. Both malformed should produce a stderr
      // warning each; the valid one survives normalize.
      const result = normalizeConfig([
        // Cast through `unknown` — `normalizeConfig` accepts `unknown[]`
        // at its public surface, but we want to verify behavior on
        // ill-typed elements without fighting the type system.
        null as unknown as Record<string, unknown>,
        { rules: { x: 'off' } },
        42 as unknown as Record<string, unknown>,
      ]);
      expect(result).toHaveLength(1);
      expect(consoleWarnCalled).toBe(false);
      // Both malformed entries produce a "not an object" line.
      const lines = stderrBuffer
        .split('\n')
        .filter((l) => l.includes('not an object'));
      expect(lines.length).toBe(2);
    } finally {
      console.warn = originalWarn;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (process.stderr as any).write = originalWrite;
    }
  });

  test('multi-entry config processes each entry independently', () => {
    const out = normalizeConfig([
      { plugins: { uni: fakeUnicorn }, rules: {} },
      { plugins: ['react'], rules: {} },
      { rules: {} },
    ]);
    expect(out).toHaveLength(3);
    expect(out[0].plugins).toEqual(['uni']);
    expect(out[0].eslintPlugins).toBeDefined();
    expect(out[1].plugins).toEqual(['react']);
    expect(out[1].eslintPlugins).toBeUndefined();
    expect(out[2].plugins).toBeUndefined();
    expect(out[2].eslintPlugins).toBeUndefined();
  });
});

// Native plugin prefix collision: rslint ports a curated set of plugin
// rules natively (Go side). When a user puts an `eslintPlugins` entry
// keyed by one of those native namespaces (e.g. `unicorn`,
// `@typescript-eslint`), the rule-name winner logic in
// `internal/config/rule_registry.go` makes the NATIVE rule win on every
// name collision — so most of the user's plugin object is silently
// shadowed. That mismatch between intent and behavior is the kind of
// "fake green" config bug a linter must reject up-front, not paper over.
//
// These tests pin:
//   1. Every reserved namespace throws with a clear, actionable message.
//   2. Multiple collisions in the same entry all surface together.
//   3. Object-form `plugins: { native: pluginObj }` also triggers the
//      check (it folds into eslintPlugins, so the same hazard applies).
//   4. The error message includes the rename guidance and the full
//      reserved list so the user can self-serve the fix.
//   5. A non-colliding prefix passes through cleanly (smoke check that
//      the validator isn't overzealous).
describe('normalizeConfig — native plugin prefix collision', () => {
  const fakePlugin = {
    meta: { name: 'fake', version: '1.0.0' },
    rules: { r: { create: () => ({}) } },
  };

  // Reuse the single source of truth so this never drifts from the gate
  // (which checks against NATIVE_PLUGIN_PREFIXES) or from the Go registry.
  const RESERVED_PREFIXES = NATIVE_PLUGIN_PREFIXES;

  for (const prefix of RESERVED_PREFIXES) {
    test(`reserved prefix "${prefix}" in eslintPlugins throws`, () => {
      expect(() =>
        normalizeConfig([
          { eslintPlugins: { [prefix]: fakePlugin }, rules: {} },
        ]),
      ).toThrow(new RegExp(`"${prefix.replace(/[/-]/g, '\\$&')}"`));
    });

    test(`reserved prefix "${prefix}" via object-form plugins also throws`, () => {
      // The object-form `plugins: {<prefix>: pluginObj}` is folded into
      // eslintPlugins by the loader, so the same collision hazard
      // applies — must be rejected by the same check.
      expect(() =>
        normalizeConfig([{ plugins: { [prefix]: fakePlugin }, rules: {} }]),
      ).toThrow(new RegExp(`"${prefix.replace(/[/-]/g, '\\$&')}"`));
    });
  }

  test('multiple collisions in one entry surface all of them in the message', () => {
    try {
      normalizeConfig([
        {
          eslintPlugins: {
            unicorn: fakePlugin,
            '@typescript-eslint': fakePlugin,
            react: fakePlugin,
          },
          rules: {},
        },
      ]);
      throw new Error('expected throw');
    } catch (err) {
      const msg = (err as Error).message;
      expect(msg).toContain('"unicorn"');
      expect(msg).toContain('"@typescript-eslint"');
      expect(msg).toContain('"react"');
    }
  });

  test('error message includes rename guidance + full reserved list', () => {
    try {
      normalizeConfig([{ eslintPlugins: { unicorn: fakePlugin }, rules: {} }]);
      throw new Error('expected throw');
    } catch (err) {
      const msg = (err as Error).message;
      // Actionable: tells the user EXACTLY how to fix it.
      expect(msg).toMatch(/rename the prefix/i);
      expect(msg).toMatch(/myUnicorn|myUnique|some-rule/);
      // Self-serve: dump the full reserved list so the user doesn't
      // have to dig through docs to know which other names are off-limits.
      for (const p of RESERVED_PREFIXES) {
        expect(msg).toContain(p);
      }
    }
  });

  test('non-colliding prefix passes through cleanly', () => {
    const out = normalizeConfig([
      { eslintPlugins: { myUnicorn: fakePlugin }, rules: {} },
    ]);
    expect(out).toHaveLength(1);
    const eps = out[0].eslintPlugins as Array<{
      prefix: string;
      ruleNames: string[];
    }>;
    expect(eps).toHaveLength(1);
    expect(eps[0].prefix).toBe('myUnicorn');
  });

  test('explicit eslintPlugins takes precedence — collision still detected', () => {
    // Even when explicit eslintPlugins overrides the plugins-object
    // extraction, a collision in the explicit map must still throw.
    expect(() =>
      normalizeConfig([
        {
          plugins: { myUnicorn: fakePlugin }, // safe
          eslintPlugins: { unicorn: fakePlugin }, // collision
          rules: {},
        },
      ]),
    ).toThrow(/"unicorn"/);
  });

  test('collision report quotes the offending index for multi-entry configs', () => {
    // Multi-entry config: entry index 2 has the collision. The error
    // message must name index 2, not index 0 — important for users
    // debugging large flat configs.
    try {
      normalizeConfig([
        { rules: {} },
        { eslintPlugins: { myA: fakePlugin }, rules: {} },
        { eslintPlugins: { unicorn: fakePlugin }, rules: {} },
      ]);
      throw new Error('expected throw');
    } catch (err) {
      const msg = (err as Error).message;
      expect(msg).toContain('index 2');
    }
  });
});
