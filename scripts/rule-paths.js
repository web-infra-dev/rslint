const fs = require('node:fs');
const path = require('node:path');

function isRuleDirectory(name) {
  return !name.startsWith('.') && name !== 'fixtures' && name !== 'testdata';
}

// Rule-local slashes are directory boundaries: prefer_global/url/url.go
// implements prefer-global/url. Git paths and returned paths use forward slashes.
function ruleDirectoryFromSource(file) {
  const parts = file.split('/');
  const basename = parts.pop();
  if (!parts.length || !parts.every(isRuleDirectory)) return null;
  const name = parts.at(-1);
  return basename === `${name}.go` || basename === 'rule.go'
    ? parts.join('/')
    : null;
}

function getRuleDirectories(directory) {
  if (!fs.existsSync(directory)) return [];
  const rules = new Set();
  function walk(relative) {
    for (const entry of fs.readdirSync(path.join(directory, relative), {
      withFileTypes: true,
    })) {
      const file = relative ? `${relative}/${entry.name}` : entry.name;
      if (entry.isDirectory() && isRuleDirectory(entry.name)) {
        walk(file);
      } else if (entry.isFile()) {
        const rule = ruleDirectoryFromSource(file);
        if (rule !== null) rules.add(rule);
      }
    }
  }
  walk('');
  return [...rules].sort();
}

module.exports = { getRuleDirectories, ruleDirectoryFromSource };
