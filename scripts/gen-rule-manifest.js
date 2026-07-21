// scripts/gen-rule-manifest.js
// Generate rule-manifest.json, initially all marked as none, can be improved for auto detection
const fs = require('fs');
const path = require('path');

const REPO_ROOT = path.join(__dirname, '..');

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
    .filter((d) => d.isDirectory() && !d.name.startsWith('.'))
    .map((d) => ({ rule: d.name, group: 'eslint', pluginDir: null }));
}

function getPluginRuleEntries() {
  // Collect rule directories from internal/plugins/{plugin}/rules/*
  if (!fs.existsSync(PLUGINS_DIR)) return [];
  const plugins = fs
    .readdirSync(PLUGINS_DIR, { withFileTypes: true })
    .filter((d) => d.isDirectory() && !d.name.startsWith('.'))
    .map((d) => d.name);
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
        (d) =>
          d.isDirectory() && d.name !== 'fixtures' && !d.name.startsWith('.'),
      )
      .map((d) => d.name);
    for (const rule of ruleDirs) {
      entries.push({ rule, group: pluginDisplayName, pluginDir: plugin });
    }
  }
  return entries;
}

// Map group name to test directory name: "@typescript-eslint" -> "typescript-eslint", etc.
function groupToTestDir(group) {
  return group.replace(/^@/, '');
}

function getIncludedRuleTests() {
  // Parse rstest.config.mts include list and associate each rule with its
  // enabled test files, including tests nested under a rule directory.
  const config = fs.readFileSync(TEST_CONFIG_PATH, 'utf-8');
  // Match uncommented flat and nested paths:
  //   ./tests/{testDir}/rules/{rule}.test.ts
  //   ./tests/{testDir}/rules/{rule}/{test-file}.test.ts
  const includeRegex =
    /^\s*'(\.(?:\/|\\)tests\/([\w-]+)\/rules\/([\w-]+)(?:\/[^']+)?\.test\.ts)'/gm;
  const included = new Map();
  let match;
  while ((match = includeRegex.exec(config))) {
    const testPath = path.resolve(path.dirname(TEST_CONFIG_PATH), match[1]);
    const testDir = match[2];
    const rule = match[3].replace(/-/g, '_');
    // Key by test-dir + rule so same-named rules in different plugins stay distinct.
    const key = `${testDir}:${rule}`;
    if (!included.has(key)) included.set(key, new Set());
    included.get(key).add(testPath);
  }
  return included;
}

function isEscaped(content, index) {
  let backslashes = 0;
  for (let i = index - 1; i >= 0 && content[i] === '\\'; i--) {
    backslashes++;
  }
  return backslashes % 2 === 1;
}

function createParserState() {
  return {
    stack: [{ type: 'normal' }],
  };
}

function getCurrentContext(state) {
  return state.stack[state.stack.length - 1];
}

function isStatementLevel(state) {
  return state.stack.length === 1 && getCurrentContext(state).type === 'normal';
}

function advanceParserState(state, content, start, end) {
  for (let i = start; i < end; i++) {
    const ch = content[i];
    const next = content[i + 1];
    const escaped = isEscaped(content, i);
    const context = getCurrentContext(state);

    if (context.type === 'lineComment') {
      if (ch === '\n') {
        state.stack.pop();
      }
      continue;
    }
    if (context.type === 'blockComment') {
      if (ch === '*' && next === '/') {
        state.stack.pop();
        i++;
      }
      continue;
    }
    if (context.type === 'singleQuote') {
      if (ch === "'" && !escaped) {
        state.stack.pop();
      }
      continue;
    }
    if (context.type === 'doubleQuote') {
      if (ch === '"' && !escaped) {
        state.stack.pop();
      }
      continue;
    }
    if (context.type === 'templateLiteral') {
      if (ch === '`' && !escaped) {
        state.stack.pop();
      } else if (ch === '$' && next === '{' && !escaped) {
        state.stack.push({ type: 'templateExpression', braceDepth: 0 });
        i++;
      }
      continue;
    }
    if (context.type === 'templateExpression') {
      if (ch === '}' && !escaped) {
        if (context.braceDepth === 0) {
          state.stack.pop();
        } else {
          context.braceDepth--;
        }
        continue;
      }
      if (ch === '{' && !escaped) {
        context.braceDepth++;
        continue;
      }
    }

    if (ch === '/' && next === '/' && !escaped) {
      state.stack.push({ type: 'lineComment' });
      i++;
    } else if (ch === '/' && next === '*' && !escaped) {
      state.stack.push({ type: 'blockComment' });
      i++;
    } else if (ch === "'" && !escaped) {
      state.stack.push({ type: 'singleQuote' });
    } else if (ch === '"' && !escaped) {
      state.stack.push({ type: 'doubleQuote' });
    } else if (ch === '`' && !escaped) {
      state.stack.push({ type: 'templateLiteral' });
    }
  }
}

function getStatementLevelSkipCases(content, relPath) {
  // Match top-level it.skip/describe.skip only. Ignore Jest API calls embedded in
  // RuleTester fixture strings (code/output properties or template literals).
  const skipCases = [];
  const lines = content.split('\n');
  const stmtSkipRegex =
    /(?:^|[^\w$])((?:it|describe)\.skip\s*\(['"]([^'"]+)['"])/g;
  const state = createParserState();
  let offset = 0;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const lineNum = i + 1;
    let cursor = offset;
    let match;

    stmtSkipRegex.lastIndex = 0;
    while ((match = stmtSkipRegex.exec(line))) {
      const prefixLength = match[0].length - match[1].length;
      const matchIndex = match.index + prefixLength;
      const absoluteMatchIndex = offset + matchIndex;

      advanceParserState(state, content, cursor, absoluteMatchIndex);
      if (isStatementLevel(state)) {
        skipCases.push({
          name: match[2],
          url: `${relPath}#L${lineNum}`,
        });
      }
      cursor = absoluteMatchIndex;
    }

    advanceParserState(state, content, cursor, offset + line.length);
    if (getCurrentContext(state).type === 'lineComment') {
      state.stack.pop();
    }
    offset += line.length + 1;
  }

  return skipCases;
}

function getObjectSkipCases(content, relPath) {
  // Match statement-level `skip: true` test-case objects, whether or not they
  // carry a `name` property; unnamed cases get a line-based placeholder.
  const skipCases = [];
  const skipRegex = /\bskip\s*:\s*true\b/g;
  const state = createParserState();
  let cursor = 0;
  let match;

  while ((match = skipRegex.exec(content))) {
    advanceParserState(state, content, cursor, match.index);
    if (isStatementLevel(state)) {
      const line = content.slice(0, match.index).split('\n').length;
      skipCases.push({
        name: `Skipped test case at line ${line}`,
        url: `${relPath}#L${line}`,
      });
    }
    cursor = skipRegex.lastIndex;
  }

  return skipCases;
}

function getSkipCases(testFile) {
  // Return skip cases as [{name, url}] for a single enabled test file.
  if (!fs.existsSync(testFile)) return [];
  const content = fs.readFileSync(testFile, 'utf-8');
  const relPath = path.relative(REPO_ROOT, testFile).split(path.sep).join('/');
  const skipCases = getObjectSkipCases(content, relPath);
  // Also collect top-level it.skip('name', ...) / describe.skip('name', ...).
  skipCases.push(...getStatementLevelSkipCases(content, relPath));
  return skipCases;
}

function getDocPath(rule, pluginDir) {
  let mdFile;
  let relPath;
  if (pluginDir) {
    mdFile = path.join(PLUGINS_DIR, pluginDir, 'rules', rule, `${rule}.md`);
    relPath = `internal/plugins/${pluginDir}/rules/${rule}/${rule}.md`;
  } else {
    mdFile = path.join(CORE_RULES_DIR, rule, `${rule}.md`);
    relPath = `internal/rules/${rule}/${rule}.md`;
  }
  return fs.existsSync(mdFile) ? relPath : null;
}

function buildManifest() {
  const included = getIncludedRuleTests();
  const ruleEntries = [...getPluginRuleEntries(), ...getCoreRuleEntries()];
  // Deduplicate by group + rule name, keeping first entry.
  const seen = new Map();
  for (const e of ruleEntries) {
    const key = `${e.group}:${e.rule}`;
    if (!seen.has(key)) seen.set(key, e);
  }
  const rules = Array.from(seen.values())
    .sort(
      (a, b) => a.rule.localeCompare(b.rule) || a.group.localeCompare(b.group),
    )
    .map((entry) => {
      const rule = entry.rule;
      let status = 'full';
      let failing_case = [];
      const group = entry.group;
      const testFiles = included.get(`${groupToTestDir(group)}:${rule}`);
      if (!testFiles) {
        status = 'partial-test';
      } else {
        const skipCases = Array.from(testFiles).flatMap(getSkipCases);
        if (skipCases.length > 0) {
          status = 'partial-impl';
          failing_case = skipCases;
        }
      }
      const docPath = getDocPath(rule, entry.pluginDir);
      return {
        name: rule.replace(/_/g, '-'),
        group,
        status,
        failing_case,
        docPath,
      };
    });
  return { rules };
}

function main() {
  const manifest = buildManifest();
  fs.mkdirSync(path.dirname(MANIFEST_PATH), { recursive: true });
  fs.writeFileSync(MANIFEST_PATH, JSON.stringify(manifest, null, 2) + '\n');
  console.log('rule-manifest.json generated at', MANIFEST_PATH);
}

if (require.main === module) {
  main();
}

module.exports = {
  buildManifest,
  getStatementLevelSkipCases,
};
