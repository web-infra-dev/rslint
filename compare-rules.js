#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');

// Fetch TypeScript ESLint rules
console.log('Fetching TypeScript ESLint rules...');
const tsEslintRulesRaw = execSync(
  'curl -s "https://api.github.com/repos/typescript-eslint/typescript-eslint/contents/packages/eslint-plugin/src/rules"',
).toString();

// Parse the JSON response
const tsEslintData = JSON.parse(tsEslintRulesRaw);

// Extract rule names from filenames
const tsEslintRules = tsEslintData
  .filter(
    item =>
      item.type === 'file' &&
      item.name.endsWith('.ts') &&
      !item.name.includes('utils') &&
      !item.name.includes('index.ts') &&
      !item.name.includes('rules-requiring-type-'),
  )
  .map(item => {
    // Convert filename.ts to rule name (strip .ts and convert to kebab-case)
    return item.name.replace('.ts', '');
  })
  .sort();

// Get RSLint Go rule directories
console.log('Getting RSLint Go rules...');
const rslintRulesDirsRaw = execSync(
  'find internal/plugins/typescript/rules -type d -not -path "*/fixtures*" | grep -v "^internal/plugins/typescript/rules$"',
).toString();

// Extract rule names from directory paths
const rslintRules = rslintRulesDirsRaw
  .split('\n')
  .filter(Boolean)
  .map(dir => {
    // Extract the last part of the path (the rule name directory)
    const ruleName = dir.split('/').pop();
    return ruleName;
  })
  .sort();

// Convert kebab-case to snake_case for comparison
function kebabToSnake(str) {
  return str.replace(/-/g, '_');
}

// Convert snake_case to kebab-case for display
function snakeToKebab(str) {
  return str.replace(/_/g, '-');
}

// Compare the rules
const implementedRules = [];
const missingRules = [];

tsEslintRules.forEach(tsRule => {
  const equivalentGoRule = kebabToSnake(tsRule);

  if (rslintRules.includes(equivalentGoRule)) {
    implementedRules.push(tsRule);
  } else {
    missingRules.push(tsRule);
  }
});

// Output results
console.log(`\nTotal TypeScript ESLint rules: ${tsEslintRules.length}`);
console.log(`Total RSLint Go rules: ${rslintRules.length}`);
console.log(`Implemented rules: ${implementedRules.length}`);
console.log(`Missing rules: ${missingRules.length}`);

console.log('\nMissing rules that need to be ported:');
missingRules.forEach(rule => {
  console.log(
    `- ${rule}: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/${rule}.ts`,
  );
});

// Export the data to a JSON file for further analysis
const exportData = {
  tsEslintRules,
  rslintRules,
  implementedRules,
  missingRules,
  stats: {
    totalTsEslintRules: tsEslintRules.length,
    totalRslintRules: rslintRules.length,
    implementedRulesCount: implementedRules.length,
    missingRulesCount: missingRules.length,
    implementationPercentage: (
      (implementedRules.length / tsEslintRules.length) *
      100
    ).toFixed(2),
  },
};

fs.writeFileSync('rule-comparison.json', JSON.stringify(exportData, null, 2));
console.log('\nDetailed data exported to rule-comparison.json');
