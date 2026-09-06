import fs from 'node:fs';
import path from 'node:path';
import { git, compareVersions } from './release-data.mjs';
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

function canonicalRuleId(group, rule) {
  return `${group}:${rule.replace(/_/g, '-')}`;
}

function uniqueSorted(values) {
  return [...new Set(values)].sort();
}

export function getStableTags(root) {
  return git(root, ['tag', '--list', 'v*'])
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

function getPluginGroupAtBlob(root, blob, plugin) {
  if (!blob) return getPluginGroup(undefined, plugin);
  const key = `${plugin}:${blob}`;
  if (!pluginGroupByBlob.has(key)) {
    pluginGroupByBlob.set(
      key,
      getPluginGroup(git(root, ['cat-file', 'blob', blob]), plugin),
    );
  }
  return pluginGroupByBlob.get(key);
}

export function getRuleIdsAtRef(root, ref) {
  const tree = git(root, [
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
    const group = getPluginGroupAtBlob(root, pluginBlobs.get(plugin), plugin);
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
        !entry.name.startsWith('.') &&
        !['fixtures', 'testdata'].includes(entry.name),
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

export function getCurrentRuleIds(root) {
  const coreRulesDir = path.join(root, 'internal/rules');
  const pluginsDir = path.join(root, 'internal/plugins');
  const ruleIds = getRuleDirectories(coreRulesDir).map((rule) =>
    canonicalRuleId('eslint', rule),
  );

  for (const plugin of getDirectories(pluginsDir)) {
    const rulesDirectory = path.join(pluginsDir, plugin, 'rules');
    const rules = getRuleDirectories(rulesDirectory);
    if (rules.length === 0) continue;

    const pluginFile = path.join(pluginsDir, plugin, 'plugin.go');
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
