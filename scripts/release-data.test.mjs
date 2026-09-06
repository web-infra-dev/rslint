import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  compareVersions,
  currentBuildInfo,
  generatedBuildInfo,
  releaseVersionAtCommit,
  validateReleases,
} from './release-data.mjs';

const commit = 'a'.repeat(40);
const other = 'b'.repeat(40);
const binding = { commit, releaseVersion: null };
const old = { version: '0.9.1', rules: ['eslint:old'] };
const current = {
  version: '0.9.2',
  rules: ['eslint:new'],
  typescript: binding,
};

test('orders stable and prerelease versions without numeric precision loss', () => {
  const versions = [
    '0.9.1',
    '0.9.2-alpha',
    '0.9.2-alpha.2',
    '0.9.2-alpha.10',
    '0.9.2-beta',
    '0.9.2-canary.abc',
    '0.9.2',
    '0.10.0',
    '999999999999999999999.0.0',
    'unreleased',
  ];
  assert.deepEqual([...versions].reverse().sort(compareVersions), versions);
  for (const version of versions)
    assert.equal(compareVersions(version, version), 0);
});

for (const version of [
  '',
  '1.2',
  '01.2.3',
  '1.2.3-01',
  '1.2.3-alpha.01',
  'v1.2.3',
  null,
]) {
  test(`rejects malformed version ${JSON.stringify(version)}`, () => {
    assert.throws(
      () => validateReleases([{ ...old, version }]),
      /Invalid release version/,
    );
  });
}

test('retains historical omissions and permits rules in both canary and stable', () => {
  const history = [old, { ...current, version: '0.9.2-canary.1' }, current];
  assert.equal(validateReleases(history), history);
});

for (const [name, entries, message] of [
  ['empty history', [], /non-empty/],
  ['duplicate versions', [old, old], /Duplicate release/],
  ['out of order', [current, old], /sorted/],
  [
    'duplicate rules',
    [{ ...old, rules: ['eslint:old', 'eslint:old'] }],
    /Duplicate rules/,
  ],
  [
    'duplicate stable introductions',
    [old, { ...current, rules: old.rules }],
    /multiple first stable/,
  ],
  ['malformed rule IDs', [{ ...old, rules: ['oops'] }], /Invalid rules/],
  [
    'missing later compiler',
    [current, { version: '0.9.3', rules: [] }],
    /Missing TypeScript/,
  ],
  [
    'missing development compiler',
    [old, { version: 'unreleased', rules: [] }],
    /Missing TypeScript/,
  ],
  [
    'short SHA',
    [{ ...current, typescript: { ...binding, commit: commit.slice(0, 12) } }],
    /full SHA/,
  ],
  [
    'missing release version',
    [{ ...current, typescript: { commit } }],
    /Invalid release version/,
  ],
]) {
  test(`rejects ${name}`, () =>
    assert.throws(() => validateReleases(entries), message));
}

test('selects an explicit development record and rejects it for publishing', () => {
  const history = [
    old,
    { version: 'unreleased', rules: [], typescript: binding },
  ];
  assert.deepEqual(currentBuildInfo(history, '0.9.1'), {
    version: 'unreleased',
    typescript: binding,
  });
  assert.throws(
    () => currentBuildInfo(history, '0.9.1', true),
    /Missing current release/,
  );
  assert.throws(
    () => currentBuildInfo([old], '0.9.1'),
    /Missing current release/,
  );
  assert.throws(
    () => currentBuildInfo([old, current], '0.9.3'),
    /Missing current release/,
  );
});

test('rejects coercible hashes and unsupported compiler version fields', () => {
  assert.throws(
    () =>
      validateReleases([
        { ...current, typescript: { ...binding, commit: [commit] } },
      ]),
    /full SHA/,
  );
  assert.throws(
    () =>
      validateReleases([
        { ...current, typescript: { ...binding, version: '7.1.0-dev' } },
      ]),
    /Unknown TypeScript/,
  );
  assert.throws(
    () => validateReleases([{ ...current, unexpected: true }]),
    /Unknown release/,
  );
});

test('build info contains only the selected version and compiler, with detached data', () => {
  const info = currentBuildInfo([old, current], '0.9.2', true);
  assert.deepEqual(info, { version: '0.9.2', typescript: binding });
  info.typescript.commit = other;
  assert.equal(current.typescript.commit, commit);
  const outputs = [...generatedBuildInfo(info).values()];
  assert.equal(outputs.length, 2);
  for (const source of outputs) {
    assert.ok(source.includes(other));
    assert.ok(!source.includes('eslint:old'));
  }
});

for (const [name, tags, expected] of [
  ['no release', '', null],
  ['lightweight tag', `${commit}\trefs/tags/v7.0.2`, '7.0.2'],
  [
    'annotated tag',
    `${other}\trefs/tags/v7.0.2\n${commit}\trefs/tags/v7.0.2^{}`,
    '7.0.2',
  ],
  [
    'annotation is not the compiler commit',
    `${commit}\trefs/tags/v7.0.2\n${other}\trefs/tags/v7.0.2^{}`,
    null,
  ],
  ['another commit is not a release match', `${other}\trefs/tags/v7.0.2`, null],
  ['non-release tag', `${commit}\trefs/tags/latest`, null],
  ['short hash', `${commit.slice(0, 12)}\trefs/tags/v7.0.2`, null],
  [
    'highest exact release',
    `${commit}\trefs/tags/v7.0.2\n${commit}\trefs/tags/v7.0.10\n${other}\trefs/tags/v8.0.0`,
    '7.0.10',
  ],
  ['prerelease tag', `${commit}\trefs/tags/v7.1.0-rc`, '7.1.0-rc'],
]) {
  test(`resolves ${name}`, () =>
    assert.equal(releaseVersionAtCommit(tags, commit), expected));
}
