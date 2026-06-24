#!/usr/bin/env node
/**
 * Search for ESLint rule documentation.
 *
 * Usage:
 *   node search_rule.mjs <rule-name>
 *
 * Examples:
 *   node search_rule.mjs no-console
 *   node search_rule.mjs @typescript-eslint/no-explicit-any
 *   node search_rule.mjs import/no-duplicates
 */

/**
 * Parse rule name to determine plugin and rule.
 * @param {string} ruleName
 * @returns {{plugin: string, rule: string}}
 */
function parseRuleName(ruleName) {
  ruleName = ruleName.trim();

  // @typescript-eslint/rule-name
  if (ruleName.startsWith('@typescript-eslint/')) {
    return {
      plugin: 'typescript-eslint',
      rule: ruleName.replace('@typescript-eslint/', ''),
    };
  }

  // plugin/rule-name format (e.g., import/no-duplicates)
  if (ruleName.includes('/')) {
    const [plugin, rule] = ruleName.split('/', 2);
    return { plugin, rule };
  }

  // Core ESLint rule
  return { plugin: 'eslint', rule: ruleName };
}

/**
 * Fetch URL content.
 * @param {string} url
 * @returns {Promise<string|null>}
 */
async function fetchUrl(url) {
  try {
    const response = await fetch(url, {
      headers: {
        'User-Agent': 'Mozilla/5.0 (compatible; ESLintRuleSearch/1.0)',
      },
      redirect: 'follow',
    });

    if (!response.ok) {
      return null;
    }

    return await response.text();
  } catch {
    return null;
  }
}

/**
 * Search ESLint core rules.
 * @param {string} ruleName
 * @returns {Promise<object|null>}
 */
async function searchEslintCore(ruleName) {
  const docUrl = `https://eslint.org/docs/latest/rules/${ruleName}`;
  const content = await fetchUrl(docUrl);

  if (!content) {
    return null;
  }

  // Check if page contains the rule name in title
  const titleMatch = content.match(/<title[^>]*>([^<]+)<\/title>/);
  if (!titleMatch) {
    return null;
  }

  const title = titleMatch[1];
  if (
    !title.toLowerCase().includes(ruleName) &&
    title.includes('Page Not Found')
  ) {
    return null;
  }

  return {
    found: true,
    plugin: 'eslint',
    rule: ruleName,
    docUrl,
    sourceUrl: `https://github.com/eslint/eslint/blob/main/lib/rules/${ruleName}.js`,
    testUrl: `https://github.com/eslint/eslint/blob/main/tests/lib/rules/${ruleName}.js`,
    title: title.replace(' - ESLint - Pluggable JavaScript Linter', '').trim(),
  };
}

/**
 * Search typescript-eslint rules.
 * @param {string} ruleName
 * @returns {Promise<object|null>}
 */
async function searchTypescriptEslint(ruleName) {
  const docUrl = `https://typescript-eslint.io/rules/${ruleName}`;
  const content = await fetchUrl(docUrl);

  if (!content) {
    return null;
  }

  // Check for valid rule page - handle <title data-rh="true"> format
  const titleMatch = content.match(/<title[^>]*>([^<]+)<\/title>/);
  if (!titleMatch) {
    return null;
  }

  const title = titleMatch[1];
  if (title.includes('Page Not Found') || title.includes('404')) {
    return null;
  }

  return {
    found: true,
    plugin: 'typescript-eslint',
    rule: ruleName,
    docUrl,
    sourceUrl: `https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/${ruleName}.ts`,
    testUrl: `https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/${ruleName}.test.ts`,
    title: title.replace(' | typescript-eslint', '').trim(),
  };
}

/**
 * Build result for plugin rule.
 * @param {string} owner
 * @param {string} repoName
 * @param {string} pluginName
 * @param {string} ruleName
 * @returns {object}
 */
function buildPluginResult(owner, repoName, pluginName, ruleName) {
  const baseUrl = `https://github.com/${owner}/${repoName}`;

  return {
    found: true,
    plugin: pluginName,
    rule: ruleName,
    docUrl: `${baseUrl}#rules`,
    sourceUrl: `${baseUrl}/tree/main/src/rules`,
    repoUrl: baseUrl,
    title: `${pluginName}/${ruleName}`,
  };
}

/**
 * Search for ESLint plugin on GitHub.
 * @param {string} pluginName
 * @param {string} ruleName
 * @returns {Promise<object|null>}
 */
async function searchGithubPlugin(pluginName, ruleName) {
  const repoPatterns = [
    `eslint-plugin-${pluginName}`,
    `eslint-plugin-${pluginName.replace(/-/g, '')}`,
  ];

  const commonOwners = [
    'import-js',
    'jsx-eslint',
    'eslint-community',
    'mysticatea',
  ];

  for (const repoName of repoPatterns) {
    // Try common organizations/users
    for (const owner of commonOwners) {
      const apiUrl = `https://api.github.com/repos/${owner}/${repoName}`;
      const content = await fetchUrl(apiUrl);

      if (content) {
        try {
          const repoData = JSON.parse(content);
          if (repoData.id) {
            return buildPluginResult(owner, repoName, pluginName, ruleName);
          }
        } catch {
          continue;
        }
      }
    }

    // Search GitHub
    const searchUrl = `https://api.github.com/search/repositories?q=${encodeURIComponent(repoName)}+eslint+plugin&sort=stars&per_page=5`;
    const content = await fetchUrl(searchUrl);

    if (content) {
      try {
        const searchData = JSON.parse(content);
        if (searchData.items?.length) {
          for (const item of searchData.items) {
            if (
              item.name.toLowerCase().includes(repoName.toLowerCase()) ||
              item.name.toLowerCase().includes(pluginName.toLowerCase())
            ) {
              return buildPluginResult(
                item.owner.login,
                item.name,
                pluginName,
                ruleName,
              );
            }
          }
        }
      } catch {
        continue;
      }
    }
  }

  return null;
}

/**
 * Search for ESLint rule documentation.
 * @param {string} ruleName
 * @returns {Promise<object>}
 */
async function searchRule(ruleName) {
  const { plugin, rule } = parseRuleName(ruleName);

  let result = null;

  if (plugin === 'eslint') {
    result = await searchEslintCore(rule);
  } else if (plugin === 'typescript-eslint') {
    result = await searchTypescriptEslint(rule);
  } else {
    result = await searchGithubPlugin(plugin, rule);
  }

  if (result) {
    return result;
  }

  return {
    found: false,
    plugin,
    rule,
    error: `Could not find rule '${ruleName}'. Please provide the documentation URL manually.`,
  };
}

/**
 * Main function.
 */
async function main() {
  const args = process.argv.slice(2);

  if (args.length < 1) {
    console.log('Usage: node search_rule.mjs <rule-name>');
    console.log('\nExamples:');
    console.log('  node search_rule.mjs no-console');
    console.log('  node search_rule.mjs @typescript-eslint/no-explicit-any');
    console.log('  node search_rule.mjs import/no-duplicates');
    process.exit(1);
  }

  const ruleName = args[0];
  const result = await searchRule(ruleName);

  console.log(JSON.stringify(result, null, 2));

  if (result.found) {
    console.log('\n' + '='.repeat(60));
    console.log(`Rule: ${result.title || result.rule}`);
    console.log(`Plugin: ${result.plugin}`);
    console.log(`Documentation: ${result.docUrl}`);
    if (result.sourceUrl) {
      console.log(`Source Code: ${result.sourceUrl}`);
    }
    if (result.testUrl) {
      console.log(`Test File: ${result.testUrl}`);
    }
    console.log('='.repeat(60));
    process.exit(0);
  } else {
    console.error(`\nError: ${result.error}`);
    process.exit(1);
  }
}

main();
