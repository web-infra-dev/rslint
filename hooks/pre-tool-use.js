#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

// Parse input from stdin
let input = '';
process.stdin.on('data', chunk => input += chunk);
process.stdin.on('end', async () => {
  try {
    const data = JSON.parse(input);
    const { tool, params } = data;
    
    // Lock files when they're being edited
    if (tool === 'Edit' || tool === 'MultiEdit' || tool === 'Write') {
      const filePath = params.file_path || params.path;
      if (filePath && filePath.includes('rslint')) {
        const lockFile = filePath + '.lock.' + process.env.RSLINT_WORKER_ID;
        const lockDir = path.dirname(lockFile);
        
        // Try to acquire lock
        let locked = false;
        for (let i = 0; i < 10; i++) {
          try {
            // Check for other locks
            const files = fs.readdirSync(lockDir).filter(f => 
              f.startsWith(path.basename(filePath) + '.lock.') && 
              f !== path.basename(lockFile)
            );
            
            if (files.length === 0) {
              // No other locks, create ours
              fs.writeFileSync(lockFile, process.env.RSLINT_WORKER_ID || 'main', { flag: 'wx' });
              locked = true;
              break;
            }
          } catch (err) {
            // Directory might not exist or lock already exists
          }
          
          if (!locked && i < 9) {
            // Wait before retry
            await new Promise(resolve => setTimeout(resolve, 500 + Math.random() * 500));
          }
        }
        
        if (!locked) {
          console.error(JSON.stringify({
            error: 'Could not acquire file lock',
            file: filePath,
            worker: process.env.RSLINT_WORKER_ID
          }));
          process.exit(1);
        }
      }
    }
    
    // Allow tool to proceed
    console.log(JSON.stringify({ allow: true }));
  } catch (err) {
    console.error(JSON.stringify({ error: err.message }));
    process.exit(1);
  }
});
