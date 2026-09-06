#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';
import { parseArgs } from 'node:util';
import { fileURLToPath } from 'node:url';
import {
  REPO_ROOT,
  STABLE_VERSION_RE,
  compareVersions,
  checkHistoricalBindings,
  compilerCommit,
  currentBuildInfo,
  git,
  readPackageVersion,
  readReleases,
  resolveTypeScript,
  validateReleases,
  writeGeneratedBuildInfo,
} from './release-data.mjs';
import {
  getCurrentRuleIds,
  getRuleIdsAtRef,
  getStableTags,
} from './release-rules.mjs';

function newRules(root, releases, version) {
  const previous = releases
    .filter(
      (entry) =>
        STABLE_VERSION_RE.test(entry.version) &&
        compareVersions(entry.version, version) < 0,
    )
    .at(-1);
  if (!previous) throw new Error(`No stable release exists before ${version}`);
  const tag = `v${previous.version}`;
  try {
    git(root, ['rev-parse', '--verify', `refs/tags/${tag}^{commit}`]);
  } catch {
    git(root, ['fetch', '--no-tags', 'origin', 'tag', tag]);
  }
  const assigned = new Set([
    ...getRuleIdsAtRef(root, tag),
    ...releases
      .filter(
        (entry) =>
          STABLE_VERSION_RE.test(entry.version) &&
          compareVersions(entry.version, version) < 0,
      )
      .flatMap((entry) => entry.rules),
  ]);
  const current = getCurrentRuleIds(root);
  if (!current.length)
    throw new Error(
      'No current rules found; refusing to record an empty selection',
    );
  return current.filter((id) => !assigned.has(id));
}

export function rebuildRuleHistory(root, releases) {
  const tags = getStableTags(root);
  const versions = new Set(tags.map((entry) => entry.version));
  if (
    !tags.length ||
    releases.some(
      (entry) =>
        STABLE_VERSION_RE.test(entry.version) && !versions.has(entry.version),
    )
  ) {
    throw new Error('Full rule sync needs all stable tags; fetch tags first');
  }
  const assigned = new Set();
  const rebuilt = new Map(releases.map((entry) => [entry.version, entry]));
  for (const { tag, version } of tags) {
    const rules = getRuleIdsAtRef(root, tag).filter((id) => !assigned.has(id));
    for (const id of rules) assigned.add(id);
    // Preserve recorded compiler bindings and prerelease entries. Historical
    // releases without a binding remain unrecorded, even during a full sync.
    rebuilt.set(version, { ...rebuilt.get(version), version, rules });
  }
  return [...rebuilt.values()].sort((a, b) =>
    compareVersions(a.version, b.version),
  );
}

export function syncReleases({
  root = REPO_ROOT,
  release = false,
  full = false,
  check = false,
  base,
  verifyUpstream = false,
  typescriptRepository,
} = {}) {
  const version = readPackageVersion(root);
  let releases = readReleases(root);
  const commit = compilerCommit(root, check);
  if (check) {
    const info = currentBuildInfo(releases, version, release);
    if (info.typescript.commit !== commit)
      throw new Error(
        'Release metadata does not match the TypeScript submodule',
      );
    writeGeneratedBuildInfo(root, info, true);
    if (base) checkHistoricalBindings(root, releases, base);
    if (release || verifyUpstream) {
      const resolved = resolveTypeScript(root, commit, typescriptRepository);
      if (resolved.releaseVersion !== info.typescript.releaseVersion) {
        throw new Error(
          'Recorded TypeScript release version does not match the upstream commit',
        );
      }
    }
    if (release) {
      const expectedRules = newRules(root, releases, version);
      if (
        JSON.stringify(releases.at(-1).rules) !== JSON.stringify(expectedRules)
      ) {
        throw new Error(
          'Stale rules in current release; run pnpm sync:releases --release',
        );
      }
      for (const parent of ['packages', 'npm/rslint']) {
        for (const dir of fs.readdirSync(path.join(root, parent))) {
          if (parent === 'packages' && dir === 'tsgo') continue;
          const file = path.join(root, parent, dir, 'package.json');
          if (
            fs.existsSync(file) &&
            JSON.parse(fs.readFileSync(file, 'utf8')).version !== version
          ) {
            throw new Error(`Package version mismatch: ${file}`);
          }
        }
      }
    }
    return info;
  }

  if (full) releases = rebuildRuleHistory(root, releases);
  const typescript = resolveTypeScript(root, commit, typescriptRepository);
  releases = releases.filter((entry) => entry.version !== 'unreleased');
  if (release) {
    // Remote tags protect published versions even in a shallow checkout.
    if (git(root, ['ls-remote', '--tags', 'origin', `refs/tags/v${version}`])) {
      throw new Error(`Refusing to rewrite published version ${version}`);
    }
    const existing = releases.findIndex((entry) => entry.version === version);
    if (
      (existing !== -1 && !releases[existing].typescript) ||
      releases.some((entry) => compareVersions(entry.version, version) > 0)
    ) {
      throw new Error(`Refusing to rewrite historical version ${version}`);
    }
    const entry = {
      version,
      rules: newRules(root, releases, version),
      typescript,
    };
    if (existing === -1) releases.push(entry);
    else releases[existing] = entry;
  } else {
    releases.push({ version: 'unreleased', rules: [], typescript });
  }
  validateReleases(releases);
  const info = currentBuildInfo(releases, version, release);
  fs.writeFileSync(
    path.join(root, 'releases.json'),
    `${JSON.stringify(releases, null, 2)}\n`,
  );
  writeGeneratedBuildInfo(root, info);
  return info;
}

export function main(argv = process.argv.slice(2)) {
  const { values, positionals } = parseArgs({
    args: argv,
    allowPositionals: true,
    options: {
      check: { type: 'boolean' },
      release: { type: 'boolean' },
      full: { type: 'boolean' },
      base: { type: 'string' },
      'verify-upstream': { type: 'boolean' },
    },
  });
  if (
    positionals.length > 1 ||
    (positionals.length && positionals[0] !== 'full') ||
    (values.check && (values.full || positionals.length)) ||
    (!values.check && (values.base || values['verify-upstream']))
  ) {
    throw new Error('Usage: pnpm sync:releases [--release] [--check | --full]');
  }
  const info = syncReleases({
    ...values,
    verifyUpstream: values['verify-upstream'],
    full: values.full || positionals[0] === 'full',
  });
  console.log(
    `${values.check ? 'Verified' : 'Updated'} ${info.version}: TypeScript ${info.typescript.commit}`,
  );
}

if (
  process.argv[1] &&
  path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  try {
    main();
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
