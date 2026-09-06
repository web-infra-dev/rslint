import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { test } from 'node:test';
import { git, readReleases, writeGeneratedBuildInfo } from './release-data.mjs';
import { getCurrentRuleIds, getRuleIdsAtRef } from './release-rules.mjs';
import { syncReleases } from './sync-releases.mjs';

function write(root, file, data) {
  const target = path.join(root, file);
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.writeFileSync(
    target,
    typeof data === 'string' ? data : JSON.stringify(data),
  );
}

function init(root) {
  fs.mkdirSync(root, { recursive: true });
  git(root, ['init', '--quiet']);
  git(root, ['config', 'user.email', 'test@example.com']);
  git(root, ['config', 'user.name', 'Release Test']);
}

function commitAll(root) {
  git(root, ['add', '.']);
  git(root, [
    '-c',
    'commit.gpgsign=false',
    'commit',
    '--quiet',
    '-m',
    'fixture',
  ]);
  return git(root, ['rev-parse', 'HEAD']);
}

function fixture(t) {
  const base = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-releases-'));
  t.after(() => fs.rmSync(base, { recursive: true, force: true }));
  const upstream = path.join(base, 'compiler');
  init(upstream);
  write(upstream, 'tsc/go.mod', 'module example.com/compiler\n');
  const releasedCommit = commitAll(upstream);
  git(upstream, [
    '-c',
    'tag.gpgsign=false',
    'tag',
    '-a',
    'v7.0.2',
    '-m',
    'release',
  ]);
  write(upstream, 'tsc/compiler.go', 'package compiler\n');
  const compilerCommit = commitAll(upstream);

  const root = path.join(base, 'rslint');
  init(root);
  git(root, [
    '-c',
    'protocol.file.allow=always',
    'submodule',
    'add',
    '--quiet',
    upstream,
    'typescript-go',
  ]);
  write(root, 'internal/rules/old/old.go', 'package old\n');
  write(
    root,
    'internal/plugins/typescript/plugin.go',
    'const PLUGIN_NAME = "@typescript-eslint"\n',
  );
  write(
    root,
    'internal/plugins/typescript/rules/old_ts/old_ts.go',
    'package old_ts\n',
  );
  const setVersion = (version) => {
    write(root, 'package.json', { version });
    write(root, 'packages/rslint/package.json', { version });
    write(root, 'npm/rslint/test/package.json', { version });
  };
  setVersion('0.9.1');
  const original = [{ version: '0.9.1', rules: getCurrentRuleIds(root) }];
  write(root, 'releases.json', original);
  commitAll(root);
  git(root, ['-c', 'tag.gpgsign=false', 'tag', 'v0.9.1']);
  const origin = path.join(base, 'origin.git');
  git(base, ['clone', '--quiet', '--bare', root, origin]);
  git(root, ['remote', 'add', 'origin', origin]);
  const options = { root, typescriptRepository: upstream };
  return {
    base,
    root,
    origin,
    upstream,
    options,
    setVersion,
    original,
    compilerCommit,
    releasedCommit,
  };
}

test('development sync leaves old bindings unrecorded and generates checked metadata', (t) => {
  const f = fixture(t);
  const info = syncReleases(f.options);
  assert.deepEqual(info, {
    version: 'unreleased',
    typescript: { commit: f.compilerCommit, releaseVersion: null },
  });
  assert.deepEqual(readReleases(f.root).slice(0, -1), f.original);
  assert.deepEqual(syncReleases({ ...f.options, check: true }), info);
  assert.throws(
    () => syncReleases({ ...f.options, check: true, release: true }),
    /Missing current release/,
  );
});

test('canary does not consume first stable rule support, and empty-rule releases retain a binding', (t) => {
  const f = fixture(t);
  write(f.root, 'internal/rules/new/new.go', 'package added\n');
  f.setVersion('0.9.2-canary.1');
  syncReleases({ ...f.options, release: true });
  f.setVersion('0.9.2');
  syncReleases({ ...f.options, release: true });
  const history = readReleases(f.root);
  assert.deepEqual(
    history.slice(1).map((entry) => entry.rules),
    [['eslint:new'], ['eslint:new']],
  );
  assert.equal(history[1].typescript.commit, history[2].typescript.commit);
  syncReleases({ ...f.options, check: true, release: true });
  commitAll(f.root);
  git(f.root, ['-c', 'tag.gpgsign=false', 'tag', 'v0.9.2']);
  f.setVersion('0.9.3');
  syncReleases({ ...f.options, release: true });
  assert.deepEqual(readReleases(f.root).at(-1).rules, []);
  assert.equal(readReleases(f.root).at(-1).typescript.commit, f.compilerCommit);
});

test('a shallow checkout without tags cannot rewrite a published version', (t) => {
  const f = fixture(t);
  const clone = path.join(f.base, 'shallow');
  git(f.base, [
    'clone',
    '--quiet',
    '--depth=1',
    '--no-tags',
    pathToFileURL(f.origin).href,
    clone,
  ]);
  git(clone, [
    '-c',
    'protocol.file.allow=always',
    'submodule',
    'update',
    '--init',
    '--quiet',
  ]);
  assert.equal(git(clone, ['tag', '--list']), '');
  assert.throws(
    () => syncReleases({ ...f.options, root: clone, release: true }),
    /published version/,
  );
  assert.deepEqual(readReleases(clone), f.original);
  // A new release can fetch just the preceding tag needed for its rule diff.
  write(clone, 'package.json', { version: '0.9.2' });
  syncReleases({ ...f.options, root: clone, release: true });
  assert.equal(readReleases(clone).at(-1).typescript.commit, f.compilerCommit);
});

test('full sync preserves stable and canary bindings without backfilling old releases', (t) => {
  const f = fixture(t);
  f.setVersion('0.9.2-canary.1');
  syncReleases({ ...f.options, release: true });
  f.setVersion('0.9.2');
  syncReleases({ ...f.options, release: true });
  commitAll(f.root);
  git(f.root, ['-c', 'tag.gpgsign=false', 'tag', 'v0.9.2']);
  const before = readReleases(f.root);
  syncReleases({ ...f.options, full: true });
  assert.deepEqual(readReleases(f.root).slice(0, -1), before);
  git(f.root, ['tag', '-d', 'v0.9.1']);
  assert.throws(
    () => syncReleases({ ...f.options, full: true }),
    /all stable tags/,
  );
});

test('failed upstream lookups do not erase metadata or pretend there is no release', (t) => {
  const f = fixture(t);
  syncReleases(f.options);
  const before = fs.readFileSync(path.join(f.root, 'releases.json'), 'utf8');
  assert.throws(() =>
    syncReleases({
      ...f.options,
      typescriptRepository: path.join(f.base, 'missing'),
    }),
  );
  assert.equal(
    fs.readFileSync(path.join(f.root, 'releases.json'), 'utf8'),
    before,
  );
});

test('dirty, unstaged, uninitialized, and mismatched compiler sources are rejected', (t) => {
  const f = fixture(t);
  syncReleases(f.options);
  const compiler = path.join(f.root, 'typescript-go');
  write(compiler, 'tsc/compiler.go', 'package modified\n');
  assert.throws(() => syncReleases({ ...f.options, check: true }), /dirty/);
  git(compiler, ['restore', '.']);
  write(compiler, 'tsc/extra.go', 'package extra\n');
  assert.throws(() => syncReleases({ ...f.options, check: true }), /dirty/);
  fs.unlinkSync(path.join(compiler, 'tsc/extra.go'));
  git(compiler, ['checkout', '--quiet', f.releasedCommit]);
  assert.throws(() => syncReleases({ ...f.options, check: true }), /gitlink/);
  git(f.root, ['add', 'typescript-go']);
  assert.throws(
    () => syncReleases({ ...f.options, check: true }),
    /does not match the TypeScript/,
  );
  syncReleases(f.options);
  assert.equal(readReleases(f.root).at(-1).typescript.releaseVersion, '7.0.2');
  fs.rmSync(compiler, { recursive: true });
  fs.mkdirSync(compiler);
  assert.throws(() => syncReleases(f.options), /not initialized/);
});

test('stale generated outputs and false upstream release claims fail checks', (t) => {
  const f = fixture(t);
  f.setVersion('0.9.2');
  const info = syncReleases({ ...f.options, release: true });
  for (const file of [
    'internal/buildinfo/generated.go',
    'packages/rslint/src/build-info.generated.ts',
  ]) {
    write(f.root, file, 'stale');
    assert.throws(
      () => syncReleases({ ...f.options, check: true }),
      /Stale generated/,
    );
    writeGeneratedBuildInfo(f.root, info);
  }
  const history = readReleases(f.root);
  history.at(-1).typescript.releaseVersion = '7.0.2';
  write(f.root, 'releases.json', history);
  writeGeneratedBuildInfo(f.root, {
    ...info,
    typescript: history.at(-1).typescript,
  });
  assert.throws(
    () => syncReleases({ ...f.options, check: true, release: true }),
    /does not match the upstream/,
  );
});

test('release checks reject a platform package from a different version', (t) => {
  const f = fixture(t);
  f.setVersion('0.9.2');
  syncReleases({ ...f.options, release: true });
  write(f.root, 'npm/rslint/test/package.json', { version: '0.9.1' });
  assert.throws(
    () => syncReleases({ ...f.options, check: true, release: true }),
    /Package version mismatch/,
  );
});

test('CI rejects hand-edited history, backfills, and deleted release records', (t) => {
  const f = fixture(t);
  const info = syncReleases(f.options);
  syncReleases({ ...f.options, check: true, base: 'HEAD' });
  const history = readReleases(f.root);
  history[0].typescript = info.typescript;
  write(f.root, 'releases.json', history);
  assert.throws(
    () => syncReleases({ ...f.options, check: true, base: 'HEAD' }),
    /Changed historical/,
  );
  write(f.root, 'releases.json', history.slice(1));
  assert.throws(
    () => syncReleases({ ...f.options, check: true, base: 'HEAD' }),
    /Removed historical/,
  );
});

test('release checks catch rules added after the version bump', (t) => {
  const f = fixture(t);
  f.setVersion('0.9.2');
  syncReleases({ ...f.options, release: true });
  write(f.root, 'internal/rules/late/late.go', 'package late\n');
  assert.throws(
    () => syncReleases({ ...f.options, check: true, release: true }),
    /Stale rules/,
  );
  syncReleases({ ...f.options, release: true });
  syncReleases({ ...f.options, check: true, release: true });
});

test('rule history excludes helper directories, testdata, and documentation-only rules', (t) => {
  const f = fixture(t);
  write(
    f.root,
    'internal/plugins/typescript/rules/fixtures/fixtures.go',
    'package fixtures\n',
  );
  write(
    f.root,
    'internal/rules/accessorutil/accessor_key.go',
    'package accessorutil\n',
  );
  write(f.root, 'internal/rules/testdata/example.go', 'package example\n');
  write(f.root, 'internal/rules/future/future.md', '# Future rule\n');
  write(
    f.root,
    'internal/plugins/typescript/rules/helpers/shared.go',
    'package helpers\n',
  );
  commitAll(f.root);
  assert.deepEqual(getCurrentRuleIds(f.root), f.original[0].rules);
  assert.deepEqual(getRuleIdsAtRef(f.root, 'HEAD'), f.original[0].rules);
});

test('rule.go implementations under plugin directories retain their namespace', (t) => {
  const f = fixture(t);
  write(
    f.root,
    'internal/plugins/rstest/plugin.go',
    'const PLUGIN_NAME = "rstest"\n',
  );
  write(
    f.root,
    'internal/plugins/rstest/rules/example/rule.go',
    'package example\n',
  );
  commitAll(f.root);
  const expected = [...f.original[0].rules, 'rstest:example'].sort();
  assert.deepEqual(getCurrentRuleIds(f.root), expected);
  assert.deepEqual(getRuleIdsAtRef(f.root, 'HEAD'), expected);
});

test('identical plugin files retain distinct fallback namespaces in history', (t) => {
  const f = fixture(t);
  for (const plugin of ['import', 'jest']) {
    write(f.root, `internal/plugins/${plugin}/plugin.go`, 'package plugin\n');
    write(
      f.root,
      `internal/plugins/${plugin}/rules/example/example.go`,
      'package example\n',
    );
  }
  commitAll(f.root);
  assert.deepEqual(getRuleIdsAtRef(f.root, 'HEAD'), getCurrentRuleIds(f.root));
});

test('PR checks verify release claims and accept platform checkout line endings', (t) => {
  const f = fixture(t);
  const info = syncReleases(f.options);
  for (const file of [
    'internal/buildinfo/generated.go',
    'packages/rslint/src/build-info.generated.ts',
  ]) {
    const text = fs.readFileSync(path.join(f.root, file), 'utf8');
    write(f.root, file, text.replace(/\n/g, '\r\n'));
  }
  syncReleases({ ...f.options, check: true, verifyUpstream: true });
  const history = readReleases(f.root);
  history.at(-1).typescript.releaseVersion = '7.0.2';
  write(f.root, 'releases.json', history);
  writeGeneratedBuildInfo(f.root, {
    ...info,
    typescript: history.at(-1).typescript,
  });
  assert.throws(
    () => syncReleases({ ...f.options, check: true, verifyUpstream: true }),
    /does not match the upstream/,
  );
});
