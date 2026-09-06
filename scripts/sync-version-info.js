#!/usr/bin/env node

const fs = require('node:fs');
const path = require('node:path');
const { execFileSync } = require('node:child_process');

const REPO_ROOT = path.resolve(__dirname, '..');
const RELEASES_PATH = path.join(REPO_ROOT, 'website/releases.json');
const CORE_RULES_DIR = path.join(REPO_ROOT, 'internal/rules');
const PLUGINS_DIR = path.join(REPO_ROOT, 'internal/plugins');
const STABLE_VERSION_RE = /^\d+\.\d+\.\d+$/;
const TYPESCRIPT_REPOSITORY = 'https://github.com/microsoft/TypeScript.git';
const PLUGIN_GROUP_FALLBACKS = new Map([
  ['import', 'eslint-plugin-import'],
  ['jest', 'eslint-plugin-jest'],
  ['jsx_a11y', 'eslint-plugin-jsx-a11y'],
  ['promise', 'eslint-plugin-promise'],
  ['react', 'eslint-plugin-react'],
  ['react_hooks', 'eslint-plugin-react-hooks'],
  ['typescript', '@typescript-eslint'],
  ['unicorn', 'eslint-plugin-unicorn'],
]);
const pluginGroupByBlob = new Map();

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

function canonicalRuleId(group, rule) {
  return `${group}:${rule.replace(/_/g, '-')}`;
}

function uniqueSorted(values) {
  return [...new Set(values)].sort();
}

function getGitOutput(args) {
  return execFileSync('git', args, {
    cwd: REPO_ROOT,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  }).trim();
}

function getStableTags() {
  return getGitOutput(['tag', '--list', 'v*'])
    .split('\n')
    .filter(Boolean)
    .map((tag) => ({ tag, version: normalizeVersion(tag) }))
    .filter(({ version }) => STABLE_VERSION_RE.test(version))
    .sort((a, b) => compareVersions(a.version, b.version));
}

function getPluginGroup(content, plugin) {
  const match = content?.match(/PLUGIN_NAME\s*=\s*"([^"]+)"/);
  return match?.[1] || PLUGIN_GROUP_FALLBACKS.get(plugin) || plugin;
}

function getPluginGroupAtBlob(blob, plugin) {
  if (!blob) return getPluginGroup(undefined, plugin);
  const key = `${plugin}:${blob}`;
  if (!pluginGroupByBlob.has(key)) {
    pluginGroupByBlob.set(
      key,
      getPluginGroup(getGitOutput(['cat-file', 'blob', blob]), plugin),
    );
  }
  return pluginGroupByBlob.get(key);
}

function getRuleIdsAtRef(ref) {
  const tree = getGitOutput([
    'ls-tree',
    '-r',
    ref,
    '--',
    'internal/rules',
    'internal/plugins',
  ]);
  if (!tree) return [];

  const coreRules = new Set();
  const pluginRules = new Map();
  const pluginBlobs = new Map();

  for (const line of tree.split('\n')) {
    const entry = /^\d+\s+\w+\s+([0-9a-f]+)\t(.+)$/.exec(line);
    if (!entry) continue;
    const [, blob, file] = entry;

    const pluginFile = /^internal\/plugins\/([^/]+)\/plugin\.go$/.exec(file);
    if (pluginFile) pluginBlobs.set(pluginFile[1], blob);

    const coreRule = /^internal\/rules\/([^/]+)\/([^/]+)\.go$/.exec(file);
    if (
      coreRule &&
      !['fixtures', 'testdata'].includes(coreRule[1]) &&
      (coreRule[1] === coreRule[2] || coreRule[2] === 'rule')
    ) {
      coreRules.add(coreRule[1]);
      continue;
    }

    const pluginRule =
      /^internal\/plugins\/([^/]+)\/rules\/([^/]+)\/([^/]+)\.go$/.exec(file);
    if (
      !pluginRule ||
      ['fixtures', 'testdata'].includes(pluginRule[2]) ||
      (pluginRule[2] !== pluginRule[3] && pluginRule[3] !== 'rule')
    )
      continue;
    const [, plugin, rule] = pluginRule;
    if (!pluginRules.has(plugin)) pluginRules.set(plugin, new Set());
    pluginRules.get(plugin).add(rule);
  }

  // Before v0.1.11, TypeScript rules lived in internal/rules.
  const coreGroup = pluginRules.has('typescript')
    ? 'eslint'
    : '@typescript-eslint';
  const ruleIds = [...coreRules].map((rule) =>
    canonicalRuleId(coreGroup, rule),
  );

  for (const [plugin, rules] of pluginRules) {
    const group = getPluginGroupAtBlob(pluginBlobs.get(plugin), plugin);
    for (const rule of rules) {
      ruleIds.push(canonicalRuleId(group, rule));
    }
  }

  return uniqueSorted(ruleIds);
}

function getDirectories(directory) {
  if (!fs.existsSync(directory)) return [];
  return fs
    .readdirSync(directory, { withFileTypes: true })
    .filter(
      (entry) =>
        entry.isDirectory() &&
        !['fixtures', 'testdata'].includes(entry.name) &&
        !entry.name.startsWith('.'),
    )
    .map((entry) => entry.name);
}

function getRuleDirectories(directory) {
  return getDirectories(directory).filter((name) =>
    [name, 'rule'].some((file) =>
      fs.existsSync(path.join(directory, name, `${file}.go`)),
    ),
  );
}

function getCurrentRuleIds() {
  const ruleIds = getRuleDirectories(CORE_RULES_DIR).map((rule) =>
    canonicalRuleId('eslint', rule),
  );

  for (const plugin of getDirectories(PLUGINS_DIR)) {
    const rulesDirectory = path.join(PLUGINS_DIR, plugin, 'rules');
    const rules = getRuleDirectories(rulesDirectory);
    if (rules.length === 0) continue;

    const pluginFile = path.join(PLUGINS_DIR, plugin, 'plugin.go');
    const content = fs.existsSync(pluginFile)
      ? fs.readFileSync(pluginFile, 'utf8')
      : undefined;
    const group = getPluginGroup(content, plugin);
    for (const rule of rules) {
      ruleIds.push(canonicalRuleId(group, rule));
    }
  }

  return uniqueSorted(ruleIds);
}

function readReleases() {
  return JSON.parse(fs.readFileSync(RELEASES_PATH, 'utf8'));
}

function writeReleases(data) {
  fs.mkdirSync(path.dirname(RELEASES_PATH), { recursive: true });
  fs.writeFileSync(RELEASES_PATH, `${JSON.stringify(data, null, 2)}\n`);
}

function getTypeScriptBinding() {
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

function syncFullHistory() {
  const stableTags = getStableTags();
  if (stableTags.length === 0) {
    throw new Error('No stable release tags were found');
  }
  const recorded = new Map(
    readReleases().map((entry) => [entry.version, entry]),
  );
  const versions = new Set(stableTags.map(({ version }) => version));
  if (
    [...recorded.keys()].some(
      (version) => version !== 'main' && !versions.has(version),
    )
  ) {
    throw new Error(
      'Fetch all recorded stable release tags before a full sync',
    );
  }

  const assignedRules = new Set();
  const releases = stableTags.map(({ tag, version }) => {
    const rules = getRuleIdsAtRef(tag).filter(
      (rule) => !assignedRules.has(rule),
    );
    for (const rule of rules) assignedRules.add(rule);
    // Rebuild rule history without adding or changing compiler bindings.
    return { ...recorded.get(version), version, rules };
  });
  if (recorded.has('main')) releases.push(recorded.get('main'));

  writeReleases(releases);
}

function syncCurrentVersion() {
  const version = normalizeVersion(
    JSON.parse(fs.readFileSync(path.join(REPO_ROOT, 'package.json'), 'utf8'))
      .version,
  );
  if (!STABLE_VERSION_RE.test(version)) {
    throw new Error(`Package version "${version}" is not a stable version`);
  }

  const stableTags = getStableTags();
  const latestTag = stableTags.at(-1);
  if (latestTag && compareVersions(version, latestTag.version) < 0) {
    throw new Error(
      `Package version v${version} is older than ${latestTag.tag}`,
    );
  }
  const published = latestTag?.version === version;
  const previousTag = latestTag;
  if (!previousTag) {
    throw new Error(`No stable tag exists before v${version}`);
  }

  const releases = readReleases().filter((entry) => entry.version !== 'main');
  const latestRelease = releases.at(-1);
  const targetIndex = releases.findIndex(
    (release) => release.version === version,
  );
  if (
    !latestRelease ||
    (latestRelease.version !== previousTag.version &&
      latestRelease.version !== version)
  ) {
    throw new Error(
      `Release JSON must end at v${previousTag.version} before syncing v${version}`,
    );
  }
  if (targetIndex !== -1 && targetIndex !== releases.length - 1) {
    throw new Error(`Refusing to rewrite historical version v${version}`);
  }
  if (!published && targetIndex !== -1 && !releases[targetIndex].typescript) {
    throw new Error(
      `Refusing to add a compiler binding to historical version v${version}; fetch release tags first`,
    );
  }

  const previousRules = new Set(getRuleIdsAtRef(previousTag.tag));
  const assignedRules = new Set(
    releases
      .filter((release) => compareVersions(release.version, version) < 0)
      .flatMap((release) => release.rules),
  );
  const currentRules = getCurrentRuleIds();
  if (!currentRules.length) throw new Error('No current rules were found');
  const rules = currentRules.filter(
    (rule) => !previousRules.has(rule) && !assignedRules.has(rule),
  );
  const typescript = getTypeScriptBinding();
  if (!published) {
    const release = { version, rules, typescript };
    if (targetIndex === -1) releases.push(release);
    else releases[targetIndex] = release;
  }
  releases.push({ version: 'main', rules: published ? rules : [], typescript });
  writeReleases(releases);
}

function printUsage() {
  console.log('\nUsage:');
  console.log(
    '  pnpm sync:version-info       Sync main and the current stable release',
  );
  console.log(
    '  pnpm sync:version-info full  Rebuild rule history, preserving TypeScript bindings',
  );
}

function main() {
  const args = process.argv.slice(2);
  if (args.length > 1 || (args[0] && args[0] !== 'full')) {
    throw new Error('The only supported argument is "full"');
  }

  if (args[0] === 'full') {
    syncFullHistory();
  } else {
    syncCurrentVersion();
  }
  console.log(`Generated ${path.relative(REPO_ROOT, RELEASES_PATH)}.`);
}

try {
  main();
} catch (error) {
  console.error(`Release sync failed: ${error.message}`);
  process.exitCode = 1;
} finally {
  printUsage();
}
