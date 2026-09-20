// scripts/gen-rule-manifest.js
// Generate rule-manifest.json, initially all marked as none, can be improved for auto detection
const fs = require('fs');
const path = require('path');
const { getRuleDirectories } = require('./rule-paths');

const REPO_ROOT = path.join(__dirname, '..');

const MANIFEST_PATH = path.join(
  __dirname,
  '../website/generated/rule-manifest.json',
);

function getCoreRuleEntries(coreRulesDir) {
  // Collect rule directories from internal/rules/*
  if (!fs.existsSync(coreRulesDir)) return [];

  return fs
    .readdirSync(coreRulesDir, { withFileTypes: true })
    .filter((d) => d.isDirectory() && !d.name.startsWith('.'))
    .map((d) => ({ rule: d.name, group: 'eslint', pluginDir: null }));
}

function getPluginRuleEntries(pluginsDir) {
  // Collect rule directories from internal/plugins/{plugin}/rules/*
  if (!fs.existsSync(pluginsDir)) return [];
  const plugins = fs
    .readdirSync(pluginsDir, { withFileTypes: true })
    .filter((d) => d.isDirectory() && !d.name.startsWith('.'))
    .map((d) => d.name);
  const entries = [];
  const pluginNameCache = new Map();
  function getPluginDisplayName(plugin) {
    if (pluginNameCache.has(plugin)) return pluginNameCache.get(plugin);
    const pluginGo = path.join(pluginsDir, plugin, 'plugin.go');
    let display = plugin; // fallback
    if (fs.existsSync(pluginGo)) {
      try {
        const content = fs.readFileSync(pluginGo, 'utf-8');
        const m = content.match(/PLUGIN_NAME\s*=\s*"([^"]+)"/);
        if (m) display = m[1];
      } catch {
        // ignore, fall back to the directory name
      }
    }
    pluginNameCache.set(plugin, display);
    return display;
  }
  for (const plugin of plugins) {
    const rulesDir = path.join(pluginsDir, plugin, 'rules');
    if (!fs.existsSync(rulesDir) || !fs.statSync(rulesDir).isDirectory())
      continue;
    const pluginDisplayName = getPluginDisplayName(plugin);
    const ruleDirs = getRuleDirectories(rulesDir);
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

// Canonical manifest key. The whole file uses the `group` namespace so a
// lookup can never accidentally mix `group` and `testDir` spellings.
function ruleKey(group, rule) {
  return `${group}:${rule}`;
}

function getIncludedRuleTests(ruleEntries, configPath, testsBaseDir) {
  // Parse the Rstack test config's include list and associate each rule with its
  // enabled test files, including tests nested under a rule directory.
  const config = fs.readFileSync(configPath, 'utf-8');
  const groups = [...new Set(ruleEntries.map((entry) => entry.group))];
  const knownRules = new Set(
    ruleEntries.map(({ group, rule }) => ruleKey(group, rule)),
  );
  // Translate the on-disk test directory back to its manifest group once here,
  // so the returned map lives entirely in the `group` namespace.
  const testDirToGroup = new Map(groups.map((g) => [groupToTestDir(g), g]));
  // Fail loudly if a group's expected test directory is missing: otherwise
  // every rule in that group silently degrades to partial-test, and because
  // the manifest is gitignored and built at site-build time the flip would
  // never surface in a diff.
  for (const [testDir] of testDirToGroup) {
    const dir = path.join(testsBaseDir, testDir);
    if (!fs.existsSync(dir)) {
      throw new Error(
        `Expected test directory not found: ${path.relative(REPO_ROOT, dir)} ` +
          `(group "${testDirToGroup.get(testDir)}"). Update groupToTestDir or the test layout.`,
      );
    }
  }
  // Match uncommented flat and nested paths:
  //   ./tests/{testDir}/rules/{rule}.test.ts
  //   ./tests/{testDir}/rules/{rule}/{test-file}.test.ts
  const includeRegex =
    /^\s*'(\.(?:\/|\\)tests\/([\w-]+)\/rules\/([^']+)\.test\.ts)'/gm;
  const included = new Map();
  let match;
  while ((match = includeRegex.exec(config))) {
    const group = testDirToGroup.get(match[2]);
    if (!group) continue; // test dir with no matching plugin group
    const testPath = path.resolve(path.dirname(configPath), match[1]);
    // Prefer the complete nested rule name, then its owning rule directory
    // for suites split into several test files (e.g. rule/edge-cases.test.ts).
    let rule = match[3].replace(/-/g, '_');
    while (!knownRules.has(ruleKey(group, rule)) && rule.includes('/')) {
      rule = rule.slice(0, rule.lastIndexOf('/'));
    }
    if (!knownRules.has(ruleKey(group, rule))) continue;
    const key = ruleKey(group, rule);
    if (!included.has(key)) included.set(key, new Set());
    included.get(key).add(testPath);
  }
  // A non-empty include list that parses to zero entries means the regex or
  // the config quoting drifted (e.g. a formatter switched to double quotes).
  if (included.size === 0) {
    throw new Error(
      `No rule tests parsed from ${configPath}; ` +
        `the include-path regex is likely out of sync with the config format.`,
    );
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

function getSkipCases(testFile, repoRoot) {
  // Return skip cases as [{name, url}] for a single enabled test file.
  if (!fs.existsSync(testFile)) return [];
  const content = fs.readFileSync(testFile, 'utf-8');
  const relPath = path.relative(repoRoot, testFile).split(path.sep).join('/');
  const skipCases = getObjectSkipCases(content, relPath);
  // Also collect top-level it.skip('name', ...) / describe.skip('name', ...).
  skipCases.push(...getStatementLevelSkipCases(content, relPath));
  return skipCases;
}

function getDocPath(rule, pluginDir, repoRoot) {
  const base = pluginDir
    ? `internal/plugins/${pluginDir}/rules`
    : 'internal/rules';
  const relPath = `${base}/${rule}/${path.posix.basename(rule)}.md`;
  return fs.existsSync(path.join(repoRoot, relPath)) ? relPath : null;
}

function buildManifest(repoRoot = REPO_ROOT) {
  const ruleEntries = [
    ...getPluginRuleEntries(path.join(repoRoot, 'internal/plugins')),
    ...getCoreRuleEntries(path.join(repoRoot, 'internal/rules')),
  ];
  // Deduplicate by group + rule name, keeping first entry.
  const seen = new Map();
  for (const e of ruleEntries) {
    const key = ruleKey(e.group, e.rule);
    if (!seen.has(key)) seen.set(key, e);
  }
  const testWorkspace = path.join(repoRoot, 'packages/rslint-test-tools');
  const included = getIncludedRuleTests(
    ruleEntries,
    path.join(testWorkspace, 'rstack.config.mts'),
    path.join(testWorkspace, 'tests'),
  );
  const rules = Array.from(seen.values())
    .sort(
      (a, b) => a.rule.localeCompare(b.rule) || a.group.localeCompare(b.group),
    )
    .map((entry) => {
      const rule = entry.rule;
      let status = 'full';
      let failing_case = [];
      const group = entry.group;
      const testFiles = included.get(ruleKey(group, rule));
      if (!testFiles) {
        status = 'partial-test';
      } else {
        const skipCases = Array.from(testFiles).flatMap((file) =>
          getSkipCases(file, repoRoot),
        );
        if (skipCases.length > 0) {
          status = 'partial-impl';
          failing_case = skipCases;
        }
      }
      const docPath = getDocPath(rule, entry.pluginDir, repoRoot);
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
