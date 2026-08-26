import { beforeAll, describe, expect, test } from '@rstest/core';
import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import { createRequire } from 'node:module';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import upstreamGlobals from 'globals';
import upstreamPackage from 'globals/package.json';
import { RSLINT_GLOBAL_SETS } from '../src/config/globals/presets.js';
import { UPSTREAM_GLOBAL_SET_NAMES } from '../src/config/globals/upstream.js';

const DIST_ROOT = path.resolve(__dirname, '../dist');
const DIST_INDEX = path.join(DIST_ROOT, 'index.js');
const DIST_DATA = path.join(DIST_ROOT, 'globals');
const THIRD_PARTY_NOTICES = path.resolve(
  __dirname,
  '../THIRD-PARTY-NOTICES.md',
);
const require = createRequire(import.meta.url);
const upstreamLicense = path.join(
  path.dirname(require.resolve('globals/package.json')),
  'license',
);
type GlobalsCatalog = typeof import('../src/index.js').globals;

let globals: GlobalsCatalog;

beforeAll(async () => {
  const builtModule: unknown = await import(pathToFileURL(DIST_INDEX).href);
  if (
    builtModule === null ||
    typeof builtModule !== 'object' ||
    !('globals' in builtModule)
  ) {
    throw new TypeError('The built @rslint/core entry has no globals export');
  }
  globals = builtModule.globals as GlobalsCatalog;
});

describe('built-in globals catalog', () => {
  test('matches the pinned globals package exactly', () => {
    expect(UPSTREAM_GLOBAL_SET_NAMES).toEqual(Object.keys(upstreamGlobals));
    expect(Object.keys(globals)).toEqual([
      ...Object.keys(upstreamGlobals),
      ...Object.keys(RSLINT_GLOBAL_SETS),
    ]);
    for (const name of UPSTREAM_GLOBAL_SET_NAMES) {
      expect(globals[name]).toEqual(upstreamGlobals[name]);
    }
    for (const [name, globalSet] of Object.entries(RSLINT_GLOBAL_SETS)) {
      expect(Reflect.get(globals, name)).toEqual(globalSet);
    }
  });

  test('preserves representative access levels and environment boundaries', () => {
    expect(globals.browser.window).toBe(false);
    expect(globals.browser.location).toBe(true);
    expect(globals.node.process).toBe(false);
    expect(globals.node.require).toBe(false);
    expect(globals.nodeBuiltin.process).toBe(false);
    expect(Object.hasOwn(globals.nodeBuiltin, 'require')).toBe(false);
    expect(globals.greasemonkey.GM_cookie).toBe(false);
    expect(globals['react-native'].__DEV__).toBe(false);
  });

  test('includes language catalogs for parity with the upstream API', () => {
    expect(globals.builtin.Array).toBe(false);
    expect(globals.es5.JSON).toBe(false);
    expect(globals.es2026.Array).toBe(false);
    expect(globals.es2027.DisposableStack).toBe(false);
  });

  test('emits every pinned set as an exact standalone asset', () => {
    const names = Object.keys(upstreamGlobals) as Array<
      keyof typeof upstreamGlobals
    >;
    const emittedFiles = fs.readdirSync(DIST_DATA).sort();
    expect(emittedFiles.filter((file) => file.endsWith('.json'))).toEqual(
      names.map((name) => `${name}.json`).sort(),
    );
    expect(emittedFiles).toContain('index.d.ts');
    for (const name of names) {
      const emitted: unknown = JSON.parse(
        fs.readFileSync(path.join(DIST_DATA, `${name}.json`), 'utf8'),
      );
      expect(emitted).toEqual(upstreamGlobals[name]);
    }
  });

  test('built JavaScript and declarations have no globals package dependency', () => {
    const builtJavaScript = fs.readFileSync(DIST_INDEX, 'utf8');
    const builtDeclarations = fs.readFileSync(
      path.join(DIST_ROOT, 'index.d.ts'),
      'utf8',
    );
    const globalsDeclarations = fs.readFileSync(
      path.join(DIST_DATA, 'index.d.ts'),
      'utf8',
    );
    expect(builtJavaScript).not.toContain('globals/globals.json');
    expect(builtJavaScript).not.toContain('AbsoluteOrientationSensor');
    expect(builtDeclarations).not.toContain("require('globals')");
    expect(builtDeclarations).not.toContain('import type {');
    expect(builtDeclarations).toContain("import('./globals/index.js').Globals");
    expect(builtDeclarations).not.toContain('type GlobalsBrowser =');
    expect(globalsDeclarations).toContain(
      `// Types bundled from globals ${upstreamPackage.version}; see THIRD-PARTY-NOTICES.md.`,
    );
    expect(globalsDeclarations).toContain("readonly 'window': false;");
    expect(globalsDeclarations).toContain('export type { Globals };');
    expect(globalsDeclarations).not.toContain('export = globals;');
  });

  test('keeps the pinned upstream license in the published notice', () => {
    const notices = fs.readFileSync(THIRD_PARTY_NOTICES, 'utf8');
    const license = fs.readFileSync(upstreamLicense, 'utf8');
    const normalizeWhitespace = (value: string): string =>
      value.replaceAll(/\s+/gu, ' ').trim();

    expect(notices).toContain(
      `software and data from \`globals\` ${upstreamPackage.version}`,
    );
    expect(normalizeWhitespace(notices)).toContain(
      normalizeWhitespace(license),
    );
  });

  test('the built root loads and caches only accessed environment assets', () => {
    const entryUrl = pathToFileURL(DIST_INDEX).href;
    const expectedSetCount =
      Object.keys(upstreamGlobals).length +
      Object.keys(RSLINT_GLOBAL_SETS).length;
    const script = String.raw`
      import { createRequire } from 'node:module';
      const require = createRequire(import.meta.url);
      const loaded = () => Object.keys(require.cache)
        .filter((file) => file.replaceAll('\\', '/').includes('/globals/'))
        .map((file) => file.replaceAll('\\', '/').split('/').at(-1))
        .sort();
      const assert = (condition, message) => {
        if (!condition) throw new Error(message);
      };

      const { globals } = await import(${JSON.stringify(entryUrl)});
      assert(loaded().length === 0, 'root import loaded a globals asset');
      assert(Object.keys(globals).length === ${expectedSetCount}, 'catalog keys are incomplete');
      assert(loaded().length === 0, 'enumerating keys loaded a globals asset');

      const browser = globals.browser;
      assert(browser.window === false, 'browser asset has the wrong data');
      assert(JSON.stringify(loaded()) === '["browser.json"]', 'browser access loaded another set');
      assert(globals.browser === browser, 'browser map was not cached');

      assert(globals.node.process === false, 'node asset has the wrong data');
      assert(JSON.stringify(loaded()) === '["browser.json","node.json"]', 'node access loaded another set');
      console.log('LAZY_GLOBALS_OK');
    `;
    const result = spawnSync(
      process.execPath,
      ['--input-type=module', '-e', script],
      {
        cwd: DIST_ROOT,
        encoding: 'utf8',
      },
    );
    expect(result.status, result.stderr).toBe(0);
    expect(result.stdout.trim()).toBe('LAZY_GLOBALS_OK');
  });
});
