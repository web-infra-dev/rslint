import { describe, test, expect } from '@rstest/core';
import {
  selectPluginSource,
  configEntryHasEslintPlugins,
} from '../src/plugin-source.js';

// selectPluginSource is the SINGLE source of the host↔worker plugin
// precedence contract: @rslint/core's normalizeConfig extracts rule names
// from it for Go's placeholder registry, and the runner's plugin-loader
// extracts live instances from it. These lock every branch so the two
// consumers can't drift in precedence.
describe('selectPluginSource — host↔worker precedence contract', () => {
  test('explicit eslintPlugins wins over object-form plugins', () => {
    const ep = { '@scope/a': {} };
    const plugins = { b: {} };
    expect(selectPluginSource({ eslintPlugins: ep, plugins })).toBe(ep);
  });

  test('object-form plugins is the fallback when eslintPlugins is absent', () => {
    const plugins = { b: {} };
    expect(selectPluginSource({ plugins })).toBe(plugins);
  });

  test('a malformed eslintPlugins (array) falls through to plugins, not throw', () => {
    const plugins = { b: {} };
    // Lenient by design: the worker re-imports raw config and must not
    // crash; the host layers its own fail-fast throw separately.
    expect(selectPluginSource({ eslintPlugins: [], plugins })).toBe(plugins);
  });

  test('null eslintPlugins falls through to plugins', () => {
    const plugins = { b: {} };
    expect(selectPluginSource({ eslintPlugins: null, plugins })).toBe(plugins);
  });

  test('array-form plugins does not qualify as an object source', () => {
    expect(selectPluginSource({ plugins: ['unicorn'] })).toBeNull();
  });

  test('neither field present → null', () => {
    expect(selectPluginSource({})).toBeNull();
  });
});

// configEntryHasEslintPlugins reads the NORMALIZED shape (eslintPlugins is
// a wire array). The CLI host and the LSP host both gate worker-descriptor
// creation on it, so lock the array/empty/absent/wrong-shape branches.
describe('configEntryHasEslintPlugins — normalized-entry plugin gate', () => {
  test('non-empty eslintPlugins array → true', () => {
    expect(
      configEntryHasEslintPlugins({
        eslintPlugins: [{ prefix: 'p', ruleNames: [] }],
      }),
    ).toBe(true);
  });

  test('empty eslintPlugins array → false', () => {
    expect(configEntryHasEslintPlugins({ eslintPlugins: [] })).toBe(false);
  });

  test('absent eslintPlugins → false', () => {
    expect(configEntryHasEslintPlugins({})).toBe(false);
  });

  test('raw object-form eslintPlugins (not the normalized array) → false', () => {
    expect(configEntryHasEslintPlugins({ eslintPlugins: { p: {} } })).toBe(
      false,
    );
  });
});
