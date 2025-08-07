#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { glob } = require('glob');

async function moveArtifacts() {
  console.log('Starting artifact move process...');

  try {
    // Find all rslint binary files
    const files = await glob('binaries/*/*-rslint');
    console.log(`Found ${files.length} rslint binary files`);

    for (const file of files) {
      console.log(`Processing ${file}`);

      const filename = path.basename(file);
      const dirname = filename.replace(/-rslint$/, '');
      const targetDir = path.join('npm', dirname);
      const targetFile = path.join(
        targetDir,
        process.platform === 'win32' ? 'rslint.exe' : 'rslint',
      );

      // Create target directory and copy file
      fs.mkdirSync(targetDir, { recursive: true });
      fs.copyFileSync(file, targetFile);
      fs.chmodSync(targetFile, 0o755); // Make executable

      console.log(`Copied ${file} to ${targetFile}`);
    }

    console.log('Artifact move process completed successfully!');
  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  }
}

moveArtifacts();
