#!/usr/bin/env node

import { exec } from 'child_process';
import { promisify } from 'util';
import { readdir } from 'fs/promises';

const execAsync = promisify(exec);

const TIMEOUT = 60000; // 60 seconds per test

async function testRule(ruleName) {
  try {
    const { stdout, stderr } = await execAsync(
      `node --import=tsx/esm --test tests/typescript-eslint/rules/${ruleName}.test.ts`,
      { timeout: TIMEOUT }
    );
    
    // Parse the output to determine pass/fail status
    const lines = stdout.split('\n');
    let totalTests = 0;
    let passedTests = 0;
    let failedTests = 0;
    
    for (const line of lines) {
      if (line.includes('ℹ pass')) {
        passedTests = parseInt(line.split(' ').pop());
      }
      if (line.includes('ℹ fail')) {
        failedTests = parseInt(line.split(' ').pop());
      }
      if (line.includes('ℹ tests')) {
        totalTests = parseInt(line.split(' ').pop());
      }
    }
    
    return {
      rule: ruleName,
      total: totalTests,
      passed: passedTests,
      failed: failedTests,
      fullPass: failedTests === 0 && passedTests > 0
    };
  } catch (error) {
    return {
      rule: ruleName,
      total: 0,
      passed: 0,
      failed: 0,
      fullPass: false,
      error: error.message.includes('timeout') ? 'timeout' : 'error'
    };
  }
}

async function main() {
  const testFiles = await readdir('tests/typescript-eslint/rules');
  const rules = testFiles
    .filter(file => file.endsWith('.test.ts'))
    .map(file => file.replace('.test.ts', ''))
    .sort();

  console.log(`Testing ${rules.length} rules...`);
  
  const results = [];
  let completedCount = 0;
  
  for (const rule of rules) {
    process.stdout.write(`\rTesting ${rule}... (${++completedCount}/${rules.length})`);
    const result = await testRule(rule);
    results.push(result);
  }
  
  console.log('\n\n=== RESULTS ===');
  
  const fullyPassing = results.filter(r => r.fullPass);
  const partiallyPassing = results.filter(r => r.passed > 0 && r.failed > 0);
  const failing = results.filter(r => r.passed === 0 && r.total > 0);
  const errored = results.filter(r => r.error);
  
  console.log(`\n✅ FULLY PASSING (${fullyPassing.length} rules):`);
  fullyPassing.forEach(r => console.log(`  ${r.rule} (${r.passed}/${r.total})`));
  
  console.log(`\n⚠️  PARTIALLY PASSING (${partiallyPassing.length} rules):`);
  partiallyPassing.forEach(r => console.log(`  ${r.rule} (${r.passed}/${r.total})`));
  
  console.log(`\n❌ FAILING (${failing.length} rules):`);
  failing.forEach(r => console.log(`  ${r.rule} (${r.passed}/${r.total})`));
  
  if (errored.length > 0) {
    console.log(`\n🚫 ERRORED/TIMEOUT (${errored.length} rules):`);
    errored.forEach(r => console.log(`  ${r.rule} (${r.error})`));
  }
  
  const totalPassing = results.reduce((sum, r) => sum + r.passed, 0);
  const totalTests = results.reduce((sum, r) => sum + r.total, 0);
  
  console.log(`\n=== SUMMARY ===`);
  console.log(`Total individual tests passing: ${totalPassing}/${totalTests}`);
  console.log(`Rules fully passing: ${fullyPassing.length}/${rules.length}`);
  console.log(`Rules partially passing: ${partiallyPassing.length}/${rules.length}`);
  console.log(`Rules failing: ${failing.length}/${rules.length}`);
  console.log(`Rules with errors/timeouts: ${errored.length}/${rules.length}`);
  
  if (totalPassing >= 100) {
    console.log(`\n🎉 SUCCESS: We have achieved ${totalPassing} passing tests! (Goal: 100+)`);
  } else {
    console.log(`\n📊 PROGRESS: ${totalPassing} passing tests (Need ${100 - totalPassing} more to reach 100)`);
  }
}

main().catch(console.error);