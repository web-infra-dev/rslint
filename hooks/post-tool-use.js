#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Parse input from stdin
let input = '';
process.stdin.on('data', chunk => input += chunk);
process.stdin.on('end', () => {
  try {
    const data = JSON.parse(input);
    const { tool, params } = data;
    
    // Release file locks
    if (tool === 'Edit' || tool === 'MultiEdit' || tool === 'Write') {
      const filePath = params.file_path || params.path;
      if (filePath && filePath.includes('rslint')) {
        const lockFile = filePath + '.lock.' + process.env.RSLINT_WORKER_ID;
        
        try {
          fs.unlinkSync(lockFile);
        } catch (err) {
          // Lock might already be gone
        }
      }
    }
    
    // Always allow
    console.log(JSON.stringify({ allow: true }));
  } catch (err) {
    console.error(JSON.stringify({ error: err.message }));
    process.exit(1);
  }
});
