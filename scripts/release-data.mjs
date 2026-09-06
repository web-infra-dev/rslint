import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { isDeepStrictEqual } from 'node:util';

export const REPO_ROOT = path.resolve(
  fileURLToPath(new URL('..', import.meta.url)),
);
export const TYPESCRIPT_REPOSITORY =
  'https://github.com/microsoft/TypeScript.git';
export const STABLE_VERSION_RE = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;
const VERSION_RE =
  /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([\da-zA-Z-]+(?:\.[\da-zA-Z-]+)*))?$/;
const COMMIT_RE = /^[0-9a-f]{40}$/;

export function git(root, args) {
  return execFileSync('git', args, {
    cwd: root,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
    timeout: 60_000,
  }).trim();
}

function parseVersion(version) {
  const match = typeof version === 'string' && VERSION_RE.exec(version);
  if (!match || match[4]?.split('.').some((part) => /^0\d+$/.test(part))) {
    throw new Error(`Invalid release version: ${version}`);
  }
  return { core: match.slice(1, 4), prerelease: match[4]?.split('.') ?? [] };
}

function comparePart(a, b) {
  const aNumeric = /^\d+$/.test(a);
  const bNumeric = /^\d+$/.test(b);
  if (aNumeric && bNumeric && a.length !== b.length) return a.length - b.length;
  if (aNumeric !== bNumeric) return aNumeric ? -1 : 1;
  return a < b ? -1 : a > b ? 1 : 0;
}

export function compareVersions(a, b) {
  if (a === 'unreleased' || b === 'unreleased') {
    return a === b ? 0 : a === 'unreleased' ? 1 : -1;
  }
  const left = parseVersion(a);
  const right = parseVersion(b);
  for (let i = 0; i < 3; i++) {
    const order = comparePart(left.core[i], right.core[i]);
    if (order) return order;
  }
  if (!left.prerelease.length || !right.prerelease.length) {
    return Number(!left.prerelease.length) - Number(!right.prerelease.length);
  }
  for (
    let i = 0;
    i < Math.max(left.prerelease.length, right.prerelease.length);
    i++
  ) {
    if (left.prerelease[i] === undefined) return -1;
    if (right.prerelease[i] === undefined) return 1;
    const order = comparePart(left.prerelease[i], right.prerelease[i]);
    if (order) return order;
  }
  return 0;
}

export function validateReleases(releases) {
  if (!Array.isArray(releases) || releases.length === 0) {
    throw new Error('Release history must be a non-empty array');
  }
  const versions = new Set();
  const stableRules = new Set();
  let previous;
  let metadataStarted = false;
  for (const entry of releases) {
    if (!entry || typeof entry !== 'object')
      throw new Error('Invalid release record');
    const { version, rules, typescript } = entry;
    if (
      Object.keys(entry).some(
        (key) => !['version', 'rules', 'typescript'].includes(key),
      )
    ) {
      throw new Error(`Unknown release metadata field for ${version}`);
    }
    if (version !== 'unreleased') parseVersion(version);
    if (versions.has(version)) throw new Error(`Duplicate release: ${version}`);
    if (previous && compareVersions(previous, version) >= 0) {
      throw new Error('Release history must be sorted by version');
    }
    versions.add(version);
    previous = version;
    if (
      !Array.isArray(rules) ||
      rules.some(
        (id) => typeof id !== 'string' || !/^[^\s:]+:[^\s:]+$/.test(id),
      )
    ) {
      throw new Error(`Invalid rules for ${version}`);
    }
    if (new Set(rules).size !== rules.length)
      throw new Error(`Duplicate rules in ${version}`);
    if (STABLE_VERSION_RE.test(version)) {
      for (const id of rules) {
        if (stableRules.has(id))
          throw new Error(`Rule ${id} has multiple first stable releases`);
        stableRules.add(id);
      }
    }
    if (typescript === undefined) {
      if (metadataStarted || version === 'unreleased') {
        throw new Error(`Missing TypeScript binding for ${version}`);
      }
      continue;
    }
    metadataStarted = true;
    if (
      !typescript ||
      typeof typescript.commit !== 'string' ||
      !COMMIT_RE.test(typescript.commit)
    ) {
      throw new Error(`TypeScript commit for ${version} must be a full SHA`);
    }
    if (
      Object.keys(typescript).some(
        (key) => !['commit', 'releaseVersion'].includes(key),
      )
    ) {
      throw new Error(`Unknown TypeScript metadata field for ${version}`);
    }
    if (typescript.releaseVersion !== null)
      parseVersion(typescript.releaseVersion);
  }
  return releases;
}

export function readReleases(root = REPO_ROOT) {
  return validateReleases(
    JSON.parse(fs.readFileSync(path.join(root, 'releases.json'), 'utf8')),
  );
}

export function readPackageVersion(root) {
  const { version } = JSON.parse(
    fs.readFileSync(path.join(root, 'package.json'), 'utf8'),
  );
  parseVersion(version);
  return version;
}

export function checkHistoricalBindings(root, releases, base) {
  const ref = git(root, [
    'rev-parse',
    '--verify',
    '--end-of-options',
    `${base}^{commit}`,
  ]);
  const paths = git(root, [
    'ls-tree',
    '--name-only',
    ref,
    '--',
    'releases.json',
    'website/rule-releases.json',
  ]).split('\n');
  const file = paths.includes('releases.json')
    ? 'releases.json'
    : 'website/rule-releases.json';
  const previous = validateReleases(
    JSON.parse(git(root, ['show', `${ref}:${file}`])),
  );
  const byVersion = new Map(releases.map((entry) => [entry.version, entry]));
  for (const entry of previous) {
    if (entry.version === 'unreleased') continue;
    const current = byVersion.get(entry.version);
    if (!current)
      throw new Error(`Removed historical release ${entry.version}`);
    if (!isDeepStrictEqual(entry.typescript, current.typescript)) {
      throw new Error(
        `Changed historical compiler binding for ${entry.version}`,
      );
    }
  }
}

export function currentBuildInfo(releases, packageVersion, release = false) {
  validateReleases(releases);
  const entry = releases.at(-1);
  if (
    !entry.typescript ||
    (entry.version !== packageVersion &&
      (release || entry.version !== 'unreleased'))
  ) {
    throw new Error(
      `Missing current release metadata for ${packageVersion}; run pnpm sync:releases${release ? ' --release' : ''}`,
    );
  }
  return { version: entry.version, typescript: { ...entry.typescript } };
}

export function compilerCommit(root, checkGitlink = false) {
  const compilerRoot = path.join(root, 'typescript-go');
  if (
    fs.realpathSync(git(compilerRoot, ['rev-parse', '--show-toplevel'])) !==
    fs.realpathSync(compilerRoot)
  ) {
    throw new Error('TypeScript submodule is not initialized');
  }
  const commit = git(compilerRoot, ['rev-parse', 'HEAD']);
  if (!COMMIT_RE.test(commit)) throw new Error('Invalid TypeScript checkout');
  if (
    git(compilerRoot, ['status', '--porcelain', '--untracked-files=normal'])
  ) {
    throw new Error(
      'TypeScript checkout is dirty; its SHA would not identify the compiled sources',
    );
  }
  if (checkGitlink) {
    const index = git(root, ['ls-files', '--stage', '--', 'typescript-go']);
    if (index !== `160000 ${commit} 0\ttypescript-go`) {
      throw new Error(
        'TypeScript checkout does not match the staged submodule gitlink',
      );
    }
  }
  return commit;
}

// ls-remote includes both annotated tag objects and their peeled commits.
// Only an exact commit match counts; an ancestor's tag is never a binding.
export function releaseVersionAtCommit(tags, commit) {
  const refs = new Map();
  for (const line of tags.split('\n')) {
    const match = /^([0-9a-f]{40})\s+refs\/tags\/v(.+?)(\^\{\})?$/.exec(line);
    if (!match) continue;
    const [, sha, version, peeled] = match;
    try {
      parseVersion(version);
    } catch {
      continue;
    }
    const ref = refs.get(version) ?? {};
    ref[peeled ? 'commit' : 'object'] = sha;
    refs.set(version, ref);
  }
  return (
    [...refs]
      .filter(([, ref]) => (ref.commit ?? ref.object) === commit)
      .map(([version]) => version)
      .sort(compareVersions)
      .at(-1) ?? null
  );
}

export function resolveTypeScript(
  root,
  commit,
  repository = TYPESCRIPT_REPOSITORY,
) {
  // A network error must fail, rather than being recorded as "no release".
  const tags = git(root, ['ls-remote', '--tags', repository, 'refs/tags/v*']);
  return { commit, releaseVersion: releaseVersionAtCommit(tags, commit) };
}

export function generatedBuildInfo(info) {
  const json = JSON.stringify(info);
  return new Map([
    [
      'internal/buildinfo/generated.go',
      `// Code generated by scripts/sync-releases.mjs. DO NOT EDIT.\n\npackage buildinfo\n\nconst currentJSON = ${JSON.stringify(json)}\n`,
    ],
    [
      'packages/rslint/src/build-info.generated.ts',
      `// Generated by scripts/sync-releases.mjs. Do not edit.\nexport const generatedBuildInfo = ${JSON.stringify(info, null, 2)} as const;\n`,
    ],
  ]);
}

export function writeGeneratedBuildInfo(root, info, check = false) {
  for (const [relative, content] of generatedBuildInfo(info)) {
    const file = path.join(root, relative);
    if (check) {
      if (
        !fs.existsSync(file) ||
        fs.readFileSync(file, 'utf8').replace(/\r\n/g, '\n') !== content
      ) {
        throw new Error(
          `Stale generated metadata: ${relative}; run pnpm sync:releases`,
        );
      }
    } else {
      fs.mkdirSync(path.dirname(file), { recursive: true });
      fs.writeFileSync(file, content);
    }
  }
}
