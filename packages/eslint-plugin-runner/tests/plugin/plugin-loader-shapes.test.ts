/**
 * Plugin-loader input-shape tests.
 *
 * `eslintPlugins` / `plugins` map values can arrive at the loader in
 * several shapes depending on how the user wrote their imports and how
 * the plugin package is published. We must accept every shape that's
 * routinely produced by Node's ESM/CJS interop without silently
 * dropping the plugin's rules.
 *
 *   1. **Direct plugin object** — `import p from 'pkg'; plugins: { p }`
 *      `pluginObj === { meta?, rules: {...}, ... }` (most common).
 *
 *   2. **ESM namespace** — `import * as p from 'pkg'; plugins: { p }`
 *      For a pure-default-export ESM package, the namespace is
 *      `{ default: { rules: {...} } }`. Without unwrapping `.default`,
 *      `plugin.rules` is undefined and every rule under that prefix
 *      silently disappears.
 *
 *   3. **Mixed entries in one config** — some prefixes use direct
 *      shape, others use namespace shape; both must resolve correctly
 *      in a single config evaluation.
 *
 *   4. **Plain CJS** — `import p from 'cjs-pkg'` where the package
 *      does `module.exports = { rules: {...} }`. Node's ESM↔CJS
 *      interop returns the module.exports value as the default
 *      import — same shape as (1).
 *
 *   5. **Invalid values** — `null` / non-object plugin entries must
 *      be skipped silently (matching the loader's null guard) without
 *      destabilising the remaining valid entries in the same config.
 *
 * The original loader call site wrapped `pluginObj` as
 * `{ default: pluginObj }` before handing it to `unwrapPluginModule`.
 * That artificial wrap reduced the function to a no-op shape adapter
 * — it only ever peeled the synthetic outer layer, so a genuine
 * namespace-shaped plugin (case 2) came through with `.rules` still
 * nested under `.default` and was silently dropped from the rules map.
 * These tests pin the post-fix behaviour: every accepted shape must
 * surface its rules into the dispatch map with the proper prefix.
 */
import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';

import { loadPluginsFromConfigs } from '../../src/plugin/plugin-loader.js';

async function withTempDir<T>(
  files: Record<string, string>,
  fn: (dir: string) => Promise<T>,
): Promise<T> {
  const dir = await fs.mkdtemp(
    path.join(os.tmpdir(), 'rslint-plugin-loader-shapes-'),
  );
  try {
    for (const [rel, contents] of Object.entries(files)) {
      const abs = path.join(dir, rel);
      await fs.mkdir(path.dirname(abs), { recursive: true });
      await fs.writeFile(abs, contents);
    }
    return await fn(dir);
  } finally {
    await fs.rm(dir, { recursive: true, force: true });
  }
}

/** Minimal rule body — a real `create` so loader doesn't tree-shake. */
const RULE_DEMO_BODY = `{ create() { return {}; } }`;

describe('plugin-loader input shapes', () => {
  test('direct plugin object — { p: { rules: {demo} } } registers p/demo', async () => {
    await withTempDir(
      {
        'rslint.config.mjs': `
          const plugin = { rules: { demo: ${RULE_DEMO_BODY} } };
          export default [{ eslintPlugins: { p: plugin } }];
        `,
      },
      async (dir) => {
        const out = await loadPluginsFromConfigs([
          {
            configPath: path.join(dir, 'rslint.config.mjs'),
            configDirectory: dir,
          },
        ]);
        const loaded = out.get(dir);
        expect(loaded).toBeDefined();
        expect(loaded!.rules.has('p/demo')).toBe(true);
        expect(loaded!.plugins).toHaveLength(1);
        expect(loaded!.plugins[0].prefix).toBe('p');
      },
    );
  });

  test('ESM namespace shape — { default: { rules: {demo} } } unwraps correctly', async () => {
    // Reproduces what `import * as p from './plugin.mjs'` produces for
    // a pure-default-export ESM plugin file. The config builds the
    // wrapper shape explicitly so the test does not depend on a
    // separate fixture file path.
    await withTempDir(
      {
        'rslint.config.mjs': `
          const plugin = { rules: { demo: ${RULE_DEMO_BODY} } };
          // Simulate the shape \`import * as p from 'pkg'\` produces.
          const namespaceLike = { default: plugin };
          export default [{ eslintPlugins: { p: namespaceLike } }];
        `,
      },
      async (dir) => {
        const out = await loadPluginsFromConfigs([
          {
            configPath: path.join(dir, 'rslint.config.mjs'),
            configDirectory: dir,
          },
        ]);
        const loaded = out.get(dir);
        expect(loaded).toBeDefined();
        // Pre-fix bug: wrapper passed through with .rules nested
        // under .default, so the rules map stayed empty.
        expect(loaded!.rules.has('p/demo')).toBe(true);
        expect(loaded!.plugins).toHaveLength(1);
        expect(loaded!.plugins[0].prefix).toBe('p');
      },
    );
  });

  test('mixed shapes in one config — direct + namespace coexist', async () => {
    // Important for partial-namespace-import configs where a user has
    // some plugins imported by default and others via `import * as`.
    // Both prefixes must load; one shape failing must not block the
    // other.
    await withTempDir(
      {
        'rslint.config.mjs': `
          const a = { rules: { ruleA: ${RULE_DEMO_BODY} } };
          const b = { rules: { ruleB: ${RULE_DEMO_BODY} } };
          export default [{
            eslintPlugins: {
              direct: a,
              wrapped: { default: b },
            },
          }];
        `,
      },
      async (dir) => {
        const out = await loadPluginsFromConfigs([
          {
            configPath: path.join(dir, 'rslint.config.mjs'),
            configDirectory: dir,
          },
        ]);
        const loaded = out.get(dir);
        expect(loaded).toBeDefined();
        expect(loaded!.rules.has('direct/ruleA')).toBe(true);
        expect(loaded!.rules.has('wrapped/ruleB')).toBe(true);
        expect(loaded!.plugins.map((p) => p.prefix).sort()).toEqual([
          'direct',
          'wrapped',
        ]);
      },
    );
  });

  test('object-form `plugins` field accepts the same shapes as `eslintPlugins`', async () => {
    // The loader scans BOTH `eslintPlugins` and an object-form
    // `plugins: { prefix: pluginValue }` (the latter is the
    // ESLint-flat-config shape). Same unwrap logic applies to both —
    // a regression in one path should be visible here.
    await withTempDir(
      {
        'rslint.config.mjs': `
          const plugin = { rules: { demo: ${RULE_DEMO_BODY} } };
          export default [{ plugins: { p: { default: plugin } } }];
        `,
      },
      async (dir) => {
        const out = await loadPluginsFromConfigs([
          {
            configPath: path.join(dir, 'rslint.config.mjs'),
            configDirectory: dir,
          },
        ]);
        const loaded = out.get(dir);
        expect(loaded).toBeDefined();
        expect(loaded!.rules.has('p/demo')).toBe(true);
      },
    );
  });

  test('null and non-object plugin values are silently skipped without affecting siblings', async () => {
    // `unwrapPluginModule(null)` returns null; the call site continues
    // past such entries instead of throwing or polluting the rules
    // map. A neighbouring valid entry must still load — verifies the
    // loop doesn't break early on invalid input.
    await withTempDir(
      {
        'rslint.config.mjs': `
          const valid = { rules: { okRule: ${RULE_DEMO_BODY} } };
          export default [{
            eslintPlugins: {
              broken: null,
              alsoBroken: 'not an object',
              valid,
            },
          }];
        `,
      },
      async (dir) => {
        const out = await loadPluginsFromConfigs([
          {
            configPath: path.join(dir, 'rslint.config.mjs'),
            configDirectory: dir,
          },
        ]);
        const loaded = out.get(dir);
        expect(loaded).toBeDefined();
        // valid entry survived
        expect(loaded!.rules.has('valid/okRule')).toBe(true);
        // broken entries did not register
        expect(loaded!.plugins.map((p) => p.prefix)).toEqual(['valid']);
      },
    );
  });

  test('namespace shape exposed via real `import * as` — fixture file path', async () => {
    // End-to-end variant: write a real plugin file with `export default`
    // and a config that uses `import * as`. This is the exact shape
    // Node's ESM loader produces in user-land — the test fails if any
    // future refactor of the unwrap logic regresses against the actual
    // runtime artefact (rather than the shape we hand-build above).
    await withTempDir(
      {
        'plugin.mjs': `
          export default { rules: { demo: ${RULE_DEMO_BODY} } };
        `,
        'rslint.config.mjs': `
          import * as p from './plugin.mjs';
          export default [{ eslintPlugins: { p } }];
        `,
      },
      async (dir) => {
        const out = await loadPluginsFromConfigs([
          {
            configPath: path.join(dir, 'rslint.config.mjs'),
            configDirectory: dir,
          },
        ]);
        const loaded = out.get(dir);
        expect(loaded).toBeDefined();
        expect(loaded!.rules.has('p/demo')).toBe(true);
      },
    );
  });
});

describe('loadPluginsFromConfigs merges configs sharing a directory', () => {
  test('two ConfigDescriptors with the same configDirectory → merged rules (not last-wins)', async () => {
    await withTempDir(
      {
        'plugin-pa.mjs': `export default { meta: { name: 'pa' }, rules: { 'no-foo': { create() { return {}; } } } };`,
        'plugin-pb.mjs': `export default { meta: { name: 'pb' }, rules: { 'no-bar': { create() { return {}; } } } };`,
        'config-a.mjs': `import p from './plugin-pa.mjs'; export default [{ eslintPlugins: { pa: p } }];`,
        'config-b.mjs': `import p from './plugin-pb.mjs'; export default [{ eslintPlugins: { pb: p } }];`,
      },
      async (dir) => {
        const map = await loadPluginsFromConfigs([
          { configPath: path.join(dir, 'config-a.mjs'), configDirectory: dir },
          { configPath: path.join(dir, 'config-b.mjs'), configDirectory: dir },
        ]);
        const loaded = map.get(dir);
        expect(loaded).toBeDefined();
        // Pre-fix the second config silently overwrote the first, so
        // `pa/no-foo` vanished. Post-fix both prefixes resolve.
        expect(loaded?.rules.has('pa/no-foo')).toBe(true);
        expect(loaded?.rules.has('pb/no-bar')).toBe(true);
      },
    );
  });
});
