// scripts/gen-rule-manifest.js
// Generate rule-manifest.json, initially all marked as none, can be improved for auto detection
const fs = require('fs');
const path = require('path');

const RULES_DIR = path.join(__dirname, '../internal/rules');
const TEST_CONFIG_PATH = path.join(
  __dirname,
  '../packages/rslint-test-tools/rstest.config.mts',
);
const TESTS_DIR = path.join(
  __dirname,
  '../packages/rslint-test-tools/tests/typescript-eslint/rules',
);
const MANIFEST_PATH = path.join(
  __dirname,
  '../packages/rslint-test-tools/rule-manifest.json',
);

function getRuleDirs() {
  return fs
    .readdirSync(RULES_DIR, { withFileTypes: true })
    .filter(dirent => dirent.isDirectory() && !dirent.name.startsWith('.'))
    .map(dirent => dirent.name)
    .filter(name => name !== 'fixtures'); // exclude fixtures
}

function getIncludedRules() {
  // Parse rstest.config.mts include list, extract rule names
  const config = fs.readFileSync(TEST_CONFIG_PATH, 'utf-8');
  const includeRegex =
    /\.(?:\/|\\)tests\/(?:eslint-plugin-import|typescript-eslint)\/rules\/([\w-]+)\.test\.ts/g;
  const included = new Set();
  let match;
  while ((match = includeRegex.exec(config))) {
    const rule = match[1].replace(/-/g, '_');
    included.add(rule);
  }
  return included;
}

function getSkipCases(rule) {
  // Return skip cases as [{name, url}]
  const testFile = path.join(TESTS_DIR, `${rule.replace(/_/g, '-')}.test.ts`);
  if (!fs.existsSync(testFile)) return [];
  const content = fs.readFileSync(testFile, 'utf-8');
  const relPath = `packages/rslint-test-tools/tests/typescript-eslint/rules/${rule.replace(/_/g, '-')}.test.ts`;
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
  const rules = getRuleDirs()
    .sort()
    .map(rule => {
      let status = 'full';
      let failing_case = [];
      // Determine group: if rule dir exists in internal/plugins/xxx, group=xxx; else typescript-eslint
      let group = 'typescript-eslint';
      const pluginsDir = path.join(__dirname, '../internal/plugins');
      if (fs.existsSync(pluginsDir)) {
        const pluginGroups = fs
          .readdirSync(pluginsDir, { withFileTypes: true })
          .filter(dirent => dirent.isDirectory())
          .map(dirent => dirent.name);
        for (const plugin of pluginGroups) {
          const pluginRuleDir = path.join(pluginsDir, plugin, rule);
          if (
            fs.existsSync(pluginRuleDir) &&
            fs.statSync(pluginRuleDir).isDirectory()
          ) {
            group = plugin;
            break;
          }
        }
      }
      if (!included.has(rule)) {
        status = 'partial-test';
      } else {
        const skipCases = getSkipCases(rule);
        if (skipCases.length > 0) {
          status = 'partial-impl';
          failing_case = skipCases;
        }
      }
      return { name: rule, group, status, failing_case };
    });
  return { rules };
}

function main() {
  const manifest = buildManifest();
  fs.writeFileSync(MANIFEST_PATH, JSON.stringify(manifest, null, 2) + '\n');
  console.log('rule-manifest.json generated at', MANIFEST_PATH);
}

main();
