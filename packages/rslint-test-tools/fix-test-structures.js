#!/usr/bin/env node

import { readFileSync, writeFileSync } from 'fs';
import { readdirSync } from 'fs';
import path from 'path';

const rulesDir = './tests/typescript-eslint/rules';
const testFiles = readdirSync(rulesDir).filter(f => f.endsWith('.test.ts'));

let fixedCount = 0;

for (const testFile of testFiles) {
  const filePath = path.join(rulesDir, testFile);
  let content = readFileSync(filePath, 'utf-8');
  let originalContent = content;
  
  // Fix 1: Remove duplicate rule name parameter in ruleTester.run
  // Pattern: ruleTester.run('rule-name', 'rule-name', {
  const doubleNamePattern = /ruleTester\.run\('([^']+)',\s*'[^']+',\s*\{/g;
  content = content.replace(doubleNamePattern, "ruleTester.run('$1', {");
  
  // Fix 2: Fix null parameter in ruleTester.run
  // Pattern: ruleTester.run('rule-name', null, {
  const nullPattern = /ruleTester\.run\('([^']+)',\s*null,\s*\{/g;
  content = content.replace(nullPattern, "ruleTester.run('$1', {");
  
  if (content !== originalContent) {
    writeFileSync(filePath, content);
    console.log(`Fixed: ${testFile}`);
    fixedCount++;
  }
}

console.log(`\nFixed ${fixedCount} test files`);