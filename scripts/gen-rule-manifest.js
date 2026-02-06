// scripts/gen-rule-manifest.js
// Generate rule-manifest.json, initially all marked as none, can be improved for auto detection
const fs = require('fs');
const path = require('path');

// Plugins root directory
const PLUGINS_DIR = path.join(__dirname, '../internal/plugins');
const CORE_RULES_DIR = path.join(__dirname, '../internal/rules');
const TEST_CONFIG_PATH = path.join(
  __dirname,
  '../packages/rslint-test-tools/rstest.config.mts',
);
const TESTS_BASE_DIR = path.join(
  __dirname,
  '../packages/rslint-test-tools/tests',
);
const MANIFEST_PATH = path.join(
  __dirname,
  '../website/generated/rule-manifest.json',
);

function getCoreRuleEntries() {
  // Collect rule directories from internal/rules/*
  if (!fs.existsSync(CORE_RULES_DIR)) return [];

  return fs
    .readdirSync(CORE_RULES_DIR, { withFileTypes: true })
    .filter(d => d.isDirectory() && !d.name.startsWith('.'))
    .map(d => ({ rule: d.name, group: 'eslint' }));
}

function getPluginRuleEntries() {
  // Collect rule directories from internal/plugins/{plugin}/rules/*
  if (!fs.existsSync(PLUGINS_DIR)) return [];
  const plugins = fs
    .readdirSync(PLUGINS_DIR, { withFileTypes: true })
    .filter(d => d.isDirectory() && !d.name.startsWith('.'))
    .map(d => d.name);
  const entries = [];
  const pluginNameCache = new Map();
  function getPluginDisplayName(plugin) {
    if (pluginNameCache.has(plugin)) return pluginNameCache.get(plugin);
    const pluginGo = path.join(PLUGINS_DIR, plugin, 'plugin.go');
    let display = plugin; // fallback
    if (fs.existsSync(pluginGo)) {
      try {
        const content = fs.readFileSync(pluginGo, 'utf-8');
        const m = content.match(/PLUGIN_NAME\s*=\s*"([^"]+)"/);
        if (m) display = m[1];
      } catch {}
    }
    pluginNameCache.set(plugin, display);
    return display;
  }
  for (const plugin of plugins) {
    const rulesDir = path.join(PLUGINS_DIR, plugin, 'rules');
    if (!fs.existsSync(rulesDir) || !fs.statSync(rulesDir).isDirectory())
      continue;
    const pluginDisplayName = getPluginDisplayName(plugin);
    const ruleDirs = fs
      .readdirSync(rulesDir, { withFileTypes: true })
      .filter(
        d =>
          d.isDirectory() && d.name !== 'fixtures' && !d.name.startsWith('.'),
      )
      .map(d => d.name);
    for (const rule of ruleDirs) {
      entries.push({ rule, group: pluginDisplayName });
    }
  }
  return entries;
}

// Map group name to test directory name: "@typescript-eslint" -> "typescript-eslint", etc.
function groupToTestDir(group) {
  return group.replace(/^@/, '');
}

function getIncludedRules() {
  // Parse rstest.config.mts include list, extract rule names from all test directories
  const config = fs.readFileSync(TEST_CONFIG_PATH, 'utf-8');
  // Match any uncommented test path: ./tests/{any-group}/rules/{rule}.test.ts
  const includeRegex =
    /^\s*'\.(?:\/|\\)tests\/([\w-]+)\/rules\/([\w-]+)\.test\.ts'/gm;
  const included = new Set();
  let match;
  while ((match = includeRegex.exec(config))) {
    const rule = match[2].replace(/-/g, '_');
    included.add(rule);
  }
  return included;
}

function getSkipCases(rule, group) {
  // Return skip cases as [{name, url}]
  const testDir = groupToTestDir(group);
  const testsDir = path.join(TESTS_BASE_DIR, testDir, 'rules');
  const testFile = path.join(testsDir, `${rule.replace(/_/g, '-')}.test.ts`);
  if (!fs.existsSync(testFile)) return [];
  const content = fs.readFileSync(testFile, 'utf-8');
  const relPath = `packages/rslint-test-tools/tests/${testDir}/rules/${rule.replace(/_/g, '-')}.test.ts`;
  // Get current commit hash
  let commit = process.env.GITHUB_SHA;
  if (!commit) {
    try {
      commit = require('child_process')
        .execSync('git rev-parse HEAD')
        .toString()
        .trim();
    } catch {
      commit = 'main';
    }
  }
  // url is changed to relative path + line number
  // Match { ... skip: true, name: 'xxx' } or it.skip('name', ...)
  const skipCases = [];
  // 1. Object case: { ..., skip: true, name: 'xxx' }
  const objCaseRegex =
    /\{[^}]*skip\s*:\s*true[^}]*name\s*:\s*['"]([^'"]+)['"][^}]*}/g;
  let m;
  while ((m = objCaseRegex.exec(content))) {
    const idx = m.index;
    const before = content.slice(0, idx);
    const line = before.split('\n').length;
    skipCases.push({
      name: m[1],
      url: `${relPath}#L${line}`,
    });
  }
  // 2. it.skip('name', ...)
  // 2. it.skip('name', ...)
  const itSkipRegex = /it\.skip\(['"]([^'"]+)['"]/g;
  while ((m = itSkipRegex.exec(content))) {
    const idx = m.index;
    const before = content.slice(0, idx);
    const line = before.split('\n').length;
    skipCases.push({
      name: m[1],
      url: `${relPath}#L${line}`,
    });
  }
  // 3. describe.skip('name', ...)
  // 3. describe.skip('name', ...)
  const describeSkipRegex = /describe\.skip\(['"]([^'"]+)['"]/g;
  while ((m = describeSkipRegex.exec(content))) {
    const idx = m.index;
    const before = content.slice(0, idx);
    const line = before.split('\n').length;
    skipCases.push({
      name: m[1],
      url: `${relPath}#L${line}`,
    });
  }
  return skipCases;
}

function buildManifest() {
  const included = getIncludedRules();
  const ruleEntries = [...getPluginRuleEntries(), ...getCoreRuleEntries()];
  const ruleSet = new Set(ruleEntries.map(e => e.rule));
  const ruleToGroup = new Map();
  for (const e of ruleEntries) ruleToGroup.set(e.rule, e.group);
  const rules = Array.from(ruleSet)
    .sort((a, b) => a.localeCompare(b))
    .map(rule => {
      let status = 'full';
      let failing_case = [];
      // Group now derived from PLUGIN_NAME constant; fallback to 'typescript-eslint'
      let group = ruleToGroup.get(rule) || 'typescript-eslint';
      if (!included.has(rule)) {
        status = 'partial-test';
      } else {
        const skipCases = getSkipCases(rule, group);
        if (skipCases.length > 0) {
          status = 'partial-impl';
          failing_case = skipCases;
        }
      }
      return { name: rule.replace(/_/g, '-'), group, status, failing_case };
    });
  return { rules };
}

function main() {
  const manifest = buildManifest();
  fs.mkdirSync(path.dirname(MANIFEST_PATH), { recursive: true });
  fs.writeFileSync(MANIFEST_PATH, JSON.stringify(manifest, null, 2) + '\n');
  console.log('rule-manifest.json generated at', MANIFEST_PATH);
}

main();
