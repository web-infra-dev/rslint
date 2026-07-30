#!/usr/bin/env node

const fs = require('node:fs');
const path = require('node:path');
const { execFileSync } = require('node:child_process');

const REPO_ROOT = path.resolve(__dirname, '..');
const RULE_RELEASES_PATH = path.join(REPO_ROOT, 'website/rule-releases.json');
const CORE_RULES_DIR = path.join(REPO_ROOT, 'internal/rules');
const PLUGINS_DIR = path.join(REPO_ROOT, 'internal/plugins');
const STABLE_VERSION_RE = /^\d+\.\d+\.\d+$/;
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
  if (!pluginGroupByBlob.has(blob)) {
    pluginGroupByBlob.set(
      blob,
      getPluginGroup(getGitOutput(['cat-file', 'blob', blob]), plugin),
    );
  }
  return pluginGroupByBlob.get(blob);
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

    const coreRule = /^internal\/rules\/([^/]+)\//.exec(file);
    if (coreRule && coreRule[1] !== 'fixtures') {
      coreRules.add(coreRule[1]);
      continue;
    }

    const pluginRule = /^internal\/plugins\/([^/]+)\/rules\/([^/]+)\//.exec(
      file,
    );
    if (!pluginRule || pluginRule[2] === 'fixtures') continue;
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

function getRuleDirectories(directory) {
  if (!fs.existsSync(directory)) return [];
  return fs
    .readdirSync(directory, { withFileTypes: true })
    .filter(
      (entry) =>
        entry.isDirectory() &&
        entry.name !== 'fixtures' &&
        !entry.name.startsWith('.'),
    )
    .map((entry) => entry.name);
}

function getCurrentRuleIds() {
  const ruleIds = getRuleDirectories(CORE_RULES_DIR).map((rule) =>
    canonicalRuleId('eslint', rule),
  );

  for (const plugin of getRuleDirectories(PLUGINS_DIR)) {
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

function readRuleReleases() {
  return JSON.parse(fs.readFileSync(RULE_RELEASES_PATH, 'utf8'));
}

function writeRuleReleases(data) {
  fs.mkdirSync(path.dirname(RULE_RELEASES_PATH), { recursive: true });
  fs.writeFileSync(RULE_RELEASES_PATH, `${JSON.stringify(data, null, 2)}\n`);
}

function syncFullHistory() {
  const stableTags = getStableTags();
  if (stableTags.length === 0) {
    throw new Error('No stable release tags were found');
  }

  const assignedRules = new Set();
  const releases = stableTags.map(({ tag, version }) => {
    const rules = getRuleIdsAtRef(tag).filter(
      (rule) => !assignedRules.has(rule),
    );
    for (const rule of rules) assignedRules.add(rule);
    return { version, rules };
  });

  writeRuleReleases(releases);
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
  if (latestTag?.version === version) {
    console.log(
      `Skipped: package version v${version} matches the latest stable tag.`,
    );
    return false;
  }

  const previousTag = stableTags
    .filter((entry) => compareVersions(entry.version, version) < 0)
    .at(-1);
  if (!previousTag) {
    throw new Error(`No stable tag exists before v${version}`);
  }

  const releases = readRuleReleases();
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
      `Rule release JSON must end at v${previousTag.version} before syncing v${version}`,
    );
  }
  if (targetIndex !== -1 && targetIndex !== releases.length - 1) {
    throw new Error(`Refusing to rewrite historical version v${version}`);
  }

  const previousRules = new Set(getRuleIdsAtRef(previousTag.tag));
  const assignedRules = new Set(
    releases
      .filter((release) => compareVersions(release.version, version) < 0)
      .flatMap((release) => release.rules),
  );
  const rules = getCurrentRuleIds().filter(
    (rule) => !previousRules.has(rule) && !assignedRules.has(rule),
  );
  const release = { version, rules };

  if (targetIndex === -1) {
    releases.push(release);
  } else {
    releases[targetIndex] = release;
  }
  writeRuleReleases(releases);
  return true;
}

function printUsage() {
  console.log('\nUsage:');
  console.log('  pnpm sync:rule-releases       Incremental sync (default)');
  console.log('  pnpm sync:rule-releases full  Full historical sync');
}

function main() {
  const args = process.argv.slice(2);
  if (args.length > 1 || (args[0] && args[0] !== 'full')) {
    throw new Error('The only supported argument is "full"');
  }

  if (args[0] === 'full') {
    syncFullHistory();
    console.log(`Generated ${path.relative(REPO_ROOT, RULE_RELEASES_PATH)}.`);
  } else {
    const generated = syncCurrentVersion();
    if (generated) {
      console.log(`Generated ${path.relative(REPO_ROOT, RULE_RELEASES_PATH)}.`);
    }
  }
}

try {
  main();
} catch (error) {
  console.error(`Rule release sync failed: ${error.message}`);
  process.exitCode = 1;
} finally {
  printUsage();
}
