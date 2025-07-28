#!/usr/bin/env node

import { spawn } from 'child_process';
import { readdirSync } from 'fs';
import path from 'path';

const rulesDir = './tests/typescript-eslint/rules';
const testFiles = readdirSync(rulesDir).filter(f => f.endsWith('.test.ts'));

async function updateSingleSnapshot(testFile) {
  return new Promise((resolve, reject) => {
    console.log(`Updating snapshot for ${testFile}...`);
    const proc = spawn('node', [
      '--import=tsx/esm',
      '--test',
      '--test-timeout', '30000',
      '--test-update-snapshots',
      '--test-force-exit',
      path.join(rulesDir, testFile)
    ], {
      stdio: 'inherit'
    });
    
    proc.on('close', (code) => {
      if (code === 0) {
        console.log(`✓ Updated ${testFile}`);
        resolve();
      } else {
        console.log(`✗ Failed ${testFile}`);
        resolve(); // Continue even on failure
      }
    });
    
    proc.on('error', (err) => {
      console.error(`Error updating ${testFile}:`, err);
      resolve();
    });
  });
}

async function main() {
  console.log(`Found ${testFiles.length} test files to update`);
  
  // Update snapshots one by one
  for (const testFile of testFiles) {
    await updateSingleSnapshot(testFile);
  }
  
  console.log('Snapshot update complete!');
}

main().catch(console.error);