import { describe, test, expect } from '@rstest/core';
import {
  selectPluginSource,
  unwrapPluginModule,
} from '../src/plugin-source.js';

// Single-source contract: the object-form `plugins` is the ONLY live
// community-plugin source; the array form (native-name whitelist) carries no
// live objects.
describe('selectPluginSource', () => {
  const plugin = { rules: { foo: {} } };

  test('object-form plugins → returned verbatim as the live source', () => {
    expect(selectPluginSource({ plugins: { uc: plugin } })).toEqual({
      uc: plugin,
    });
  });

  test('array-form plugins (native whitelist) → null (no live objects)', () => {
    expect(
      selectPluginSource({ plugins: ['@typescript-eslint', 'unicorn'] }),
    ).toBeNull();
  });

  test('empty object plugins {} → empty map (a downstream no-op, not an error)', () => {
    expect(selectPluginSource({ plugins: {} })).toEqual({});
  });

  test('absent plugins → null', () => {
    expect(selectPluginSource({ rules: {} })).toBeNull();
  });

  test('null / non-object entry / non-object plugins → null', () => {
    expect(selectPluginSource(null)).toBeNull();
    expect(selectPluginSource('nope')).toBeNull();
    expect(selectPluginSource({ plugins: 'nope' })).toBeNull();
  });
});

describe('unwrapPluginModule', () => {
  test('prefers `.default` (ESM importing a CJS plugin)', () => {
    const plugin = { rules: {} };
    expect(unwrapPluginModule({ default: plugin })).toBe(plugin);
  });

  test('falls back to the module itself when there is no `.default`', () => {
    const plugin = { rules: {} };
    expect(unwrapPluginModule(plugin)).toBe(plugin);
  });

  test('falls back to the module itself when `.default` is not an object', () => {
    // A `{ default: <non-object> }` is not an ESM-interop plugin wrapper, so
    // the module itself is returned rather than the non-object default.
    const mod = { default: 'not-an-object', rules: {} };
    expect(unwrapPluginModule(mod)).toBe(mod);
  });

  test('null / non-object → null', () => {
    expect(unwrapPluginModule(null)).toBeNull();
    expect(unwrapPluginModule('x')).toBeNull();
  });
});
