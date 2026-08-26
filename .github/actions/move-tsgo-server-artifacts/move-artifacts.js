#!/usr/bin/env node

const fs = require('node:fs');
const path = require('node:path');

const platforms = [
  'darwin-arm64',
  'darwin-x64',
  'linux-arm64',
  'linux-x64',
  'win32-arm64',
  'win32-x64',
];

function requireFile(file) {
  if (!fs.statSync(file, { throwIfNoEntry: false })?.isFile()) {
    throw new Error(`required artifact file not found: ${file}`);
  }
}

function requireDirectory(dir) {
  if (!fs.statSync(dir, { throwIfNoEntry: false })?.isDirectory()) {
    throw new Error(`required artifact directory not found: ${dir}`);
  }
}

function copyTypeScriptLibs(sourceDir, targetDir) {
  const libFiles = fs
    .readdirSync(sourceDir, { withFileTypes: true })
    .filter(
      (entry) =>
        entry.isFile() && entry.name !== 'tsgo' && entry.name !== 'tsgo.exe',
    )
    .map((entry) => entry.name);

  if (!libFiles.includes('lib.d.ts')) {
    throw new Error(`lib.d.ts not found in ${sourceDir}`);
  }

  for (const file of libFiles) {
    fs.copyFileSync(path.join(sourceDir, file), path.join(targetDir, file));
  }
}

function moveArtifacts() {
  const libsSource = path.join('binaries', 'tsgo-server-libs');
  requireDirectory(libsSource);

  for (const platform of platforms) {
    const artifactName = `${platform}-tsgo-server`;
    const binarySource = path.join('binaries', artifactName, artifactName);
    requireFile(binarySource);

    const targetDir = path.join('npm', 'tsgo', platform, 'lib');
    const binaryName = platform.startsWith('win32-') ? 'tsgo.exe' : 'tsgo';
    const binaryTarget = path.join(targetDir, binaryName);

    fs.mkdirSync(targetDir, { recursive: true });
    fs.copyFileSync(binarySource, binaryTarget);
    fs.chmodSync(binaryTarget, 0o755);
    copyTypeScriptLibs(libsSource, targetDir);

    console.log(`Assembled ${targetDir}`);
  }
}

try {
  moveArtifacts();
} catch (error) {
  console.error(error.message);
  process.exit(1);
}
