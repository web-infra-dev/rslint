#!/usr/bin/env node

const fs = require('node:fs');
const path = require('node:path');

const tuples = [
  'darwin-arm64',
  'darwin-x64',
  'linux-arm64-gnu',
  'linux-arm64-musl',
  'linux-x64-gnu',
  'linux-x64-musl',
  'win32-arm64-msvc',
  'win32-x64-msvc',
];

function filesIn(directory) {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const file = path.join(directory, entry.name);
    if (entry.isDirectory()) return filesIn(file);
    if (!entry.isFile())
      throw new Error(`Artifact must be a regular file: ${file}`);
    return [file];
  });
}

function moveArtifacts() {
  const files = filesIn('binaries');
  const names = new Set(tuples.map((tuple) => `rslint.${tuple}.node`));
  for (const file of files) {
    if (!names.has(path.basename(file))) {
      throw new Error(`Unexpected release artifact: ${file}`);
    }
  }
  // Validate the complete run before replacing any package outputs. Missing,
  // duplicate or stale process/service artifacts must never be hidden by a
  // previous local build.
  const artifacts = tuples.map((tuple) => {
    const filename = `rslint.${tuple}.node`;
    const matches = files.filter((file) => path.basename(file) === filename);
    if (matches.length !== 1) {
      throw new Error(
        `Expected exactly one ${filename}, found ${matches.length}`,
      );
    }
    const contents = fs.readFileSync(matches[0]);
    const header = contents.subarray(0, 4).toString('hex');
    const valid = tuple.startsWith('darwin-')
      ? ['cffaedfe', 'feedfacf'].includes(header)
      : tuple.startsWith('linux-')
        ? header === '7f454c46'
        : contents.subarray(0, 2).toString() === 'MZ';
    if (!valid) throw new Error(`Invalid native addon: ${matches[0]}`);
    const directory = path.join('npm', 'rslint', tuple);
    const metadata = JSON.parse(
      fs.readFileSync(path.join(directory, 'package.json'), 'utf8'),
    );
    if (
      metadata.name !== `@rslint/native-${tuple}` ||
      metadata.exports['.'] !== `./${filename}`
    ) {
      throw new Error(`Native package metadata does not match ${tuple}`);
    }
    for (const legacy of [
      'rslint',
      'rslint.exe',
      'rslint-tsrs-service',
      'rslint-tsrs-service.exe',
    ]) {
      if (fs.existsSync(path.join(directory, legacy))) {
        throw new Error(
          `Unexpected legacy executable in ${directory}: ${legacy}`,
        );
      }
    }
    return { source: matches[0], target: path.join(directory, filename) };
  });
  for (const { source, target } of artifacts) {
    fs.copyFileSync(source, target);
    console.log(`Copied ${source} to ${target}`);
  }
}

try {
  moveArtifacts();
} catch (error) {
  console.error(error.message);
  process.exit(1);
}
