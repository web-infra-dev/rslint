#!/usr/bin/env node
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { parseArgs } from 'node:util';
import { pathToFileURL } from 'node:url';
import {
  REPO_ROOT,
  currentBuildInfo,
  readPackageVersion,
  readReleases,
} from './release-data.mjs';
import { getCurrentRuleIds } from './release-rules.mjs';

// A separate process lets the ordinary Go WASM runtime write its CLI output
// unchanged, and lets us detect nonzero exits and malformed/trailing output.
if (process.argv[2] === '--wasm-child') {
  await import(pathToFileURL(path.resolve(process.argv[4])).href);
  const go = new globalThis.Go();
  go.argv = ['rslint', '--version', '--json'];
  go.exit = (code) => {
    process.exitCode = code;
  };
  const { instance } = await WebAssembly.instantiate(
    fs.readFileSync(process.argv[3]),
    go.importObject,
  );
  await go.run(instance);
} else {
  const { values } = parseArgs({
    options: {
      binary: { type: 'string' },
      module: { type: 'string', default: 'packages/rslint/dist/build-info.js' },
      wasm: { type: 'string' },
      'rule-schemas': { type: 'string' },
      'wasm-runtime': {
        type: 'string',
        default: 'packages/rslint-wasm/wasm_exec.js',
      },
    },
  });
  if (!values.binary) throw new Error('--binary is required');
  const expected = currentBuildInfo(
    readReleases(),
    readPackageVersion(REPO_ROOT),
  );
  const { buildInfo } = await import(
    pathToFileURL(path.resolve(values.module)).href
  );
  assert.deepEqual(
    buildInfo,
    expected,
    'npm build info differs from the release record',
  );
  const native = execFileSync(
    path.resolve(values.binary),
    ['--version', '--json'],
    { encoding: 'utf8', timeout: 30_000 },
  );
  assert.deepEqual(
    JSON.parse(native),
    expected,
    'native build info differs from the release record',
  );
  if (values['rule-schemas']) {
    const schemas = JSON.parse(fs.readFileSync(values['rule-schemas'], 'utf8'));
    const recorded = getCurrentRuleIds(REPO_ROOT).map((id) => {
      const [group, name] = id.split(':');
      return group === 'eslint'
        ? name
        : `${group.replace(/^eslint-plugin-/, '')}/${name}`;
    });
    assert.deepEqual(
      recorded.sort(),
      schemas.map((entry) => entry.name).sort(),
      'release rule selection differs from the registered Go rule catalog',
    );
  }
  if (values.wasm) {
    const wasm = execFileSync(
      process.execPath,
      [process.argv[1], '--wasm-child', values.wasm, values['wasm-runtime']],
      { encoding: 'utf8', timeout: 30_000 },
    );
    assert.deepEqual(
      JSON.parse(wasm),
      expected,
      'WASM build info differs from the release record',
    );
  }
  console.log('Verified compiler bindings in built artifacts');
}
