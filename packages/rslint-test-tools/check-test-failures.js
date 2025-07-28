#!/usr/bin/env node

import { spawn } from 'child_process';
import { readdirSync } from 'fs';
import path from 'path';

const rulesDir = './tests/typescript-eslint/rules';
const testFiles = readdirSync(rulesDir).filter(f => f.endsWith('.test.ts'));

async function checkTest(testFile) {
  return new Promise((resolve) => {
    let output = '';
    const proc = spawn('node', [
      '--import=tsx/esm',
      '--test',
      path.join(rulesDir, testFile)
    ], {
      timeout: 30000
    });
    
    proc.stdout.on('data', (data) => {
      output += data.toString();
    });
    
    proc.stderr.on('data', (data) => {
      output += data.toString();
    });
    
    proc.on('close', (code) => {
      const passed = code === 0;
      const ruleName = testFile.replace('.test.ts', '');
      
      // Check for common failure patterns
      let failureType = 'unknown';
      if (output.includes('ruleCount": 0')) {
        failureType = 'missing-rule';
      } else if (output.includes('Expected no diagnostics for valid case')) {
        failureType = 'false-positive';
      } else if (output.includes('Expected diagnostics for invalid case')) {
        failureType = 'false-negative';
      } else if (output.includes('timeout')) {
        failureType = 'timeout';
      } else if (output.includes('TypeError') || output.includes('Cannot read properties')) {
        failureType = 'runtime-error';
      } else if (output.includes('syntactic errors')) {
        failureType = 'syntax-error';
      } else if (passed) {
        failureType = 'passed';
      }
      
      resolve({ ruleName, passed, failureType });
    });
    
    proc.on('error', () => {
      resolve({ ruleName: testFile.replace('.test.ts', ''), passed: false, failureType: 'error' });
    });
  });
}

async function main() {
  console.log('Checking all tests...\n');
  
  const results = [];
  
  // Check tests in batches to avoid overwhelming the system
  const batchSize = 5;
  for (let i = 0; i < testFiles.length; i += batchSize) {
    const batch = testFiles.slice(i, i + batchSize);
    const batchResults = await Promise.all(batch.map(checkTest));
    results.push(...batchResults);
    
    // Show progress
    console.log(`Progress: ${Math.min(i + batchSize, testFiles.length)}/${testFiles.length}`);
  }
  
  // Group by failure type
  const grouped = {};
  for (const result of results) {
    if (!grouped[result.failureType]) {
      grouped[result.failureType] = [];
    }
    grouped[result.failureType].push(result.ruleName);
  }
  
  // Display results
  console.log('\n=== Test Results Summary ===\n');
  
  for (const [type, rules] of Object.entries(grouped)) {
    console.log(`${type} (${rules.length}):`);
    rules.sort().forEach(rule => console.log(`  - ${rule}`));
    console.log();
  }
  
  console.log(`Total: ${results.length} tests`);
  console.log(`Passed: ${grouped.passed?.length || 0}`);
  console.log(`Failed: ${results.length - (grouped.passed?.length || 0)}`);
}

main().catch(console.error);