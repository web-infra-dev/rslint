#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

function findBinaries(dir = 'binaries', suffix) {
  const files = [];

  if (!fs.existsSync(dir)) {
    return files;
  }

  const entries = fs.readdirSync(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);

    if (entry.isDirectory()) {
      // Look for files ending with the suffix in subdirectories
      const subEntries = fs.readdirSync(fullPath, { withFileTypes: true });
      for (const subEntry of subEntries) {
        if (subEntry.isFile() && subEntry.name.includes(suffix)) {
          files.push(path.join(fullPath, subEntry.name));
        }
      }
    }
  }

  return files;
}

async function moveArtifacts() {
  console.log('Starting artifact move process...');

  try {
    // Find and move rslint binaries
    const rslintFiles = findBinaries('binaries', '-rslint');
    console.log(`Found ${rslintFiles.length} rslint binary files`);

    for (const file of rslintFiles) {
      console.log(`Processing ${file}`);
      const isWindows = file.includes('win32');
      const filename = path.basename(file);
      const dirname = filename.replace(/-rslint$/, '');
      const targetDir = path.join('npm', 'rslint', dirname);

      const targetFile = path.join(
        targetDir,
        isWindows ? 'rslint.exe' : 'rslint',
      );

      // Create target directory and copy file
      fs.mkdirSync(targetDir, { recursive: true });
      fs.copyFileSync(file, targetFile);
      fs.chmodSync(targetFile, 0o755); // Make executable

      console.log(`Copied ${file} to ${targetFile}`);
    }

    // Find and move tsgo binaries
    const tsgoFiles = findBinaries('binaries', '-tsgo');
    console.log(`Found ${tsgoFiles.length} tsgo binary files`);

    for (const file of tsgoFiles) {
      console.log(`Processing ${file}`);
      const isWindows = file.includes('win32');
      const filename = path.basename(file);
      const dirname = filename.replace(/-tsgo$/, '');
      const targetDir = path.join('npm', 'tsgo', dirname);

      const targetFile = path.join(
        targetDir,
        isWindows ? 'tsgo.exe' : 'tsgo',
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
