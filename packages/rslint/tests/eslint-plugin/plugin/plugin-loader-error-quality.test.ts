import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';

import {
  loadPluginsFromConfigs,
  PluginLoaderError,
} from '../../../src/eslint-plugin/plugin/plugin-loader.js';

/**
 * Lock the quality of plugin-loader error messages so a future refactor
 * cannot silently drop the actionable parts the user needs to self-serve
 * a fix.
 *
 * What we pin:
 *
 *   1. PluginLoaderError carries the offending `configPath` field —
 *      not just baked into the message string. Programmatic callers
 *      (LSP, CLI) can surface it through their own UI.
 *   2. The string message includes BOTH the config file path AND the
 *      failing module specifier (e.g. `eslint-plugin-totally-nonexistent`).
 *      Node's native `Cannot find package 'X' imported from Y` already
 *      provides this; we just need to make sure our wrapping doesn't
 *      strip it.
 *   3. Both error paths (`failed to import config` for module-not-found
 *      vs. `Cannot redefine plugin` for prefix collisions) surface as
 *      the same `PluginLoaderError` shape — consistent for callers.
 */

async function withTempConfig(
  files: Record<string, string>,
  fn: (configPath: string) => Promise<void>,
): Promise<void> {
  const dir = await fs.mkdtemp(
    path.join(os.tmpdir(), 'rslint-plugin-loader-error-'),
  );
  try {
    for (const [rel, contents] of Object.entries(files)) {
      const abs = path.join(dir, rel);
      await fs.mkdir(path.dirname(abs), { recursive: true });
      await fs.writeFile(abs, contents);
    }
    await fn(path.join(dir, 'rslint.config.mjs'));
  } finally {
    await fs.rm(dir, { recursive: true, force: true });
  }
}

describe('plugin-loader error message quality', () => {
  test('missing plugin import → message includes BOTH config path AND package specifier', async () => {
    await withTempConfig(
      {
        'rslint.config.mjs': `import broken from 'eslint-plugin-totally-nonexistent-${Date.now()}';
export default [{ plugins: { foo: broken } }];
`,
      },
      async (configPath) => {
        let caught: unknown = null;
        try {
          await loadPluginsFromConfigs([
            { configPath, configDirectory: path.dirname(configPath) },
          ]);
        } catch (err) {
          caught = err;
        }
        // Must throw the typed error, not bare Error.
        expect(caught).toBeInstanceOf(PluginLoaderError);
        const e = caught as PluginLoaderError;

        // Structured field: callers should not have to regex out the
        // config path — `configPath` is a first-class property.
        expect(e.configPath).toBe(configPath);

        // String message must surface BOTH the config file (so the
        // user knows WHICH config has the problem in multi-config
        // monorepos) AND the failing specifier (so they know WHICH
        // plugin to npm-install). Node's own `Cannot find package`
        // gives us the latter for free; we just must not strip it.
        expect(e.message).toContain(configPath);
        expect(e.message).toMatch(/eslint-plugin-totally-nonexistent/);
        // Sanity: doesn't degrade to a useless generic phrasing.
        expect(e.message).not.toBe('failed to load plugin');
      },
    );
  });

  test('same prefix re-declared with different instance → PluginLoaderError with prefix-named message', async () => {
    await withTempConfig(
      {
        'plugin-a.mjs': `export default { meta: { name: 'a' }, rules: { r: { meta: {}, create() { return {}; } } } };`,
        'plugin-b.mjs': `export default { meta: { name: 'b' }, rules: { r: { meta: {}, create() { return {}; } } } };`,
        'rslint.config.mjs': `import a from './plugin-a.mjs';
import b from './plugin-b.mjs';
export default [
  { plugins: { sharedPrefix: a } },
  { plugins: { sharedPrefix: b } },
];
`,
      },
      async (configPath) => {
        let caught: unknown = null;
        try {
          await loadPluginsFromConfigs([
            { configPath, configDirectory: path.dirname(configPath) },
          ]);
        } catch (err) {
          caught = err;
        }
        expect(caught).toBeInstanceOf(PluginLoaderError);
        const e = caught as PluginLoaderError;
        // The prefix in the conflict is the load-failure signal the
        // user needs to find which key to rename.
        expect(e.message).toMatch(/sharedPrefix/);
        // ESLint v10 phrasing — same as upstream so the user can
        // search the web for the exact error and get useful hits.
        expect(e.message).toMatch(/Cannot redefine plugin/i);
        expect(e.configPath).toBe(configPath);
      },
    );
  });

  test('same prefix re-declared with IDENTICAL instance → dedup, no error', async () => {
    // Counterpart to the conflict case. ESLint v10's merge function
    // does not throw when both entries point at the EXACT SAME plugin
    // object (identity-equal). rslint must mirror that.
    await withTempConfig(
      {
        'plugin.mjs': `export default { meta: { name: 'shared' }, rules: { r: { meta: {}, create() { return {}; } } } };`,
        'rslint.config.mjs': `import p from './plugin.mjs';
export default [
  { plugins: { sharedPrefix: p } },
  { plugins: { sharedPrefix: p } },
];
`,
      },
      async (configPath) => {
        // Must NOT throw — identity dedup is the happy path.
        const map = await loadPluginsFromConfigs([
          { configPath, configDirectory: path.dirname(configPath) },
        ]);
        const loaded = map.get(path.dirname(configPath));
        expect(loaded).toBeDefined();
        // Just one plugin entry under the prefix; rule resolves once.
        expect(loaded!.plugins).toHaveLength(1);
        expect(loaded!.plugins[0].prefix).toBe('sharedPrefix');
      },
    );
  });

  test('different config files cannot redefine a prefix under one routing key even with disjoint rules', async () => {
    await withTempConfig(
      {
        'plugin-a.mjs': `export default { rules: { alpha: { meta: {}, create() { return {}; } } } };`,
        'plugin-b.mjs': `export default { rules: { beta: { meta: {}, create() { return {}; } } } };`,
        'rslint.config.mjs': `import plugin from './plugin-a.mjs'; export default [{ plugins: { p: plugin } }];`,
        'second.config.mjs': `import plugin from './plugin-b.mjs'; export default [{ plugins: { p: plugin } }];`,
      },
      async (configPath) => {
        const secondConfigPath = path.join(
          path.dirname(configPath),
          'second.config.mjs',
        );
        let caught: unknown;
        try {
          await loadPluginsFromConfigs([
            { configPath, configDirectory: '/shared-routing-key' },
            {
              configPath: secondConfigPath,
              configDirectory: '/shared-routing-key',
            },
          ]);
        } catch (error) {
          caught = error;
        }
        expect(caught).toBeInstanceOf(PluginLoaderError);
        expect((caught as PluginLoaderError).configPath).toBe(secondConfigPath);
        expect((caught as Error).message).toMatch(
          /Cannot redefine plugin "p".*shared-routing-key/,
        );
      },
    );
  });

  test('different config files dedupe the same plugin instance under one routing key', async () => {
    await withTempConfig(
      {
        'plugin.mjs': `export default { rules: { check: { meta: {}, create() { return {}; } } } };`,
        'rslint.config.mjs': `import plugin from './plugin.mjs'; export default [{ plugins: { p: plugin } }];`,
        'second.config.mjs': `import plugin from './plugin.mjs'; export default [{ plugins: { p: plugin } }];`,
      },
      async (configPath) => {
        const secondConfigPath = path.join(
          path.dirname(configPath),
          'second.config.mjs',
        );
        const map = await loadPluginsFromConfigs([
          { configPath, configDirectory: '/shared-routing-key' },
          {
            configPath: secondConfigPath,
            configDirectory: '/shared-routing-key',
          },
        ]);

        const loaded = map.get('/shared-routing-key');
        expect(loaded).toBeDefined();
        expect(loaded!.plugins).toHaveLength(1);
        expect([...loaded!.rules.keys()]).toEqual(['p/check']);
      },
    );
  });

  test.each(['cjs', 'ts', 'mts', 'cts'])(
    'loads object-form plugins from a .%s config through the worker loader',
    async (extension) => {
      const configSource =
        extension === 'cjs'
          ? `const plugin = require('./plugin.cjs');
module.exports = [{
  files: ['src/**/*.ts'],
  plugins: { probe: plugin },
}];`
          : extension === 'cts'
            ? `const plugin = require('./plugin.cjs');
const config: Array<Record<string, unknown>> = [{
  files: ['src/**/*.ts'],
  plugins: { probe: plugin },
}];
module.exports = config;`
            : `import plugin from './plugin.cjs';
export default [{
  files: ['src/**/*.ts'],
  plugins: { probe: plugin as unknown as Record<string, unknown> },
}];`;

      await withTempConfig(
        {
          'plugin.cjs': `module.exports = {
  meta: { name: 'ts-config-plugin' },
  rules: { fires: { meta: { messages: { x: 'ok' } }, create() { return {}; } } },
};`,
          [`rslint.config.${extension}`]: configSource,
        },
        async (configMjsPath) => {
          const configPath = configMjsPath.replace(/\.mjs$/, `.${extension}`);
          const loaded = await loadPluginsFromConfigs([
            {
              configPath,
              configDirectory: path.dirname(configPath),
            },
          ]);
          const pluginSet = loaded.get(path.dirname(configPath));
          expect(pluginSet).toBeDefined();
          expect(pluginSet!.plugins).toHaveLength(1);
          expect(pluginSet!.plugins[0].prefix).toBe('probe');
          expect(pluginSet!.rules.has('probe/fires')).toBe(true);
        },
      );
    },
  );
});
