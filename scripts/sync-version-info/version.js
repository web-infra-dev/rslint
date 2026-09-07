const TYPESCRIPT_REPOSITORY = 'https://github.com/microsoft/TypeScript.git';

function normalizeVersion(version) {
  return String(version).replace(/^v/, '');
}

function compareVersions(a, b) {
  const aParts = normalizeVersion(a).split('.').map(Number);
  const bParts = normalizeVersion(b).split('.').map(Number);
  for (let i = 0; i < 3; i++) {
    if (aParts[i] !== bParts[i]) return aParts[i] - bParts[i];
  }
  return 0;
}

function getTypeScriptBinding(getGitOutput) {
  // Read the pinned revision, including a staged compiler update. This also
  // works when the submodule has not been initialized in this checkout.
  const entry = getGitOutput(['ls-files', '--stage', '--', 'typescript-go']);
  const commit = /^160000 ([0-9a-f]{40}) 0\ttypescript-go$/.exec(entry)?.[1];
  if (!commit) throw new Error('No TypeScript submodule commit is recorded');
  try {
    getGitOutput([
      'diff',
      '--exit-code',
      '--ignore-submodules=dirty',
      '--',
      'typescript-go',
    ]);
  } catch {
    throw new Error('Stage the TypeScript submodule revision before syncing');
  }

  // Query upstream tags so a shallow submodule does not hide release tags.
  // Annotated tags identify their commit through the peeled ref (^{}).
  const tags = getGitOutput([
    'ls-remote',
    '--tags',
    TYPESCRIPT_REPOSITORY,
    'refs/tags/v*',
  ]);
  const refs = new Map();
  for (const line of tags.split('\n')) {
    const match =
      /^([0-9a-f]{40})\s+refs\/tags\/v(\d+\.\d+\.\d+(?:-[\w.-]+)?)(\^\{\})?$/.exec(
        line,
      );
    if (!match) continue;
    const [, sha, version, peeled] = match;
    const ref = refs.get(version) ?? {};
    ref[peeled ? 'commit' : 'object'] = sha;
    refs.set(version, ref);
  }
  const versions = [...refs]
    .filter(([, ref]) => (ref.commit ?? ref.object) === commit)
    .map(([version]) => version)
    .sort(
      (a, b) =>
        compareVersions(a.split('-')[0], b.split('-')[0]) ||
        Number(!a.includes('-')) - Number(!b.includes('-')) ||
        a.localeCompare(b, 'en', { numeric: true }),
    );
  return { commit, releaseVersion: versions.at(-1) ?? null };
}

module.exports = { normalizeVersion, compareVersions, getTypeScriptBinding };
