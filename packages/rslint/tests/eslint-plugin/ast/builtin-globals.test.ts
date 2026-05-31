/**
 * Built-in globals drift guard — pins the vendored `BUILTIN_GLOBAL_NAMES`
 * against the `globals` package's `.builtin` set.
 *
 * scope-factory seeds these as readonly globals (so `no-undef` & co. don't
 * false-positive on `Array` / `Object` / `Promise` / ...). We vendor the 62
 * names instead of `import 'globals'`, because the package ships all 58
 * environments (~84KB of JSON) but only `.builtin` (~1KB) is ever read —
 * bundling the whole thing into the worker wasted 98.7%. This guard fails if
 * the upstream set changes (e.g. a new ES global lands), prompting a vendored
 * update so the worker keeps seeding the right names without re-adding the dep.
 */
import { describe, expect, test } from '@rstest/core';
import globals from 'globals';
import { BUILTIN_GLOBAL_NAMES } from '../../../src/eslint-plugin/ast/builtin-globals.js';

describe('BUILTIN_GLOBAL_NAMES vendored set', () => {
  test('matches globals.builtin (drift guard)', () => {
    const upstream = Object.keys(
      (globals as unknown as { builtin: Record<string, unknown> }).builtin,
    );
    expect([...BUILTIN_GLOBAL_NAMES].sort()).toEqual([...upstream].sort());
  });
});
