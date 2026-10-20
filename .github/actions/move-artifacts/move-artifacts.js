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

function requireArtifact(files, filename) {
  const matches = files.filter(
    (file) => path.basename(file) === filename && fs.existsSync(file),
  );
  if (matches.length !== 1) {
    throw new Error(
      `Expected exactly one ${filename} artifact, found ${matches.length}`,
    );
  }
  const file = matches[0];
  const stat = fs.statSync(file);
  if (!stat.isFile() || stat.size === 0) {
    throw new Error(`Empty or invalid artifact: ${file}`);
  }
  return file;
}

async function moveArtifacts() {
  console.log('Starting artifact move process...');

  try {
    // Move Go CLI binaries into the @rslint/native-{tuple} subpackages (flat,
    // alongside the .node). Artifacts are named `{platform}-rslint`. Linux
    // binaries are built with cgo disabled, so one build serves both glibc and
    // musl — it's copied into both the -gnu and -musl tuple dirs.
    const platformToTuples = {
      'darwin-arm64': ['darwin-arm64'],
      'darwin-x64': ['darwin-x64'],
      'linux-arm64': ['linux-arm64-gnu', 'linux-arm64-musl'],
      'linux-x64': ['linux-x64-gnu', 'linux-x64-musl'],
      'win32-arm64': ['win32-arm64-msvc'],
      'win32-x64': ['win32-x64-msvc'],
    };

    const rslintFiles = findBinaries('binaries', '-rslint');
    console.log(`Found ${rslintFiles.length} rslint binary files`);
    const nativeFiles = findBinaries('binaries', 'rslint.').filter((file) =>
      file.endsWith('.node'),
    );
    // Validate this run's complete input before touching package outputs.
    // A previous local build must not conceal a missing or duplicate artifact.
    const goArtifacts = Object.entries(platformToTuples).map(
      ([platform, tuples]) => ({
        file: requireArtifact(rslintFiles, `${platform}-rslint`),
        tuples,
        binName: platform.startsWith('win32-') ? 'rslint.exe' : 'rslint',
      }),
    );
    const nodeArtifacts = Object.values(platformToTuples)
      .flat()
      .map((tuple) => ({
        file: requireArtifact(nativeFiles, `rslint.${tuple}.node`),
        tuple,
      }));
    const ruleSchemasSrc = requireArtifact(
      [path.join('binaries', 'rule-schemas', 'rule-schemas.json')],
      'rule-schemas.json',
    );

    for (const { file, tuples, binName } of goArtifacts) {
      console.log(`Processing ${file}`);
      for (const tuple of tuples) {
        const targetDir = path.join('npm', 'rslint', tuple);
        const targetFile = path.join(targetDir, binName);
        fs.mkdirSync(targetDir, { recursive: true });
        fs.copyFileSync(file, targetFile);
        fs.chmodSync(targetFile, 0o755); // Make executable
        console.log(`Copied ${file} to ${targetFile}`);
      }
    }

    // Move napi `.node` parser binaries into the @rslint/native-{tuple}
    // subpackages (flat, alongside the Go binary). Artifacts are named
    // `rslint.{tuple}.node`; target is `npm/rslint/{tuple}/` (libc suffix kept
    // so gnu/musl stay separate). Not chmod'd — a `.node` is dlopen'd.
    console.log(`Found ${nodeArtifacts.length} napi .node files`);

    for (const { file, tuple } of nodeArtifacts) {
      console.log(`Processing ${file}`);
      const filename = path.basename(file); // rslint.linux-x64-gnu.node
      const targetDir = path.join('npm', 'rslint', tuple);
      const targetFile = path.join(targetDir, filename);

      fs.mkdirSync(targetDir, { recursive: true });
      fs.copyFileSync(file, targetFile);

      console.log(`Copied ${file} to ${targetFile}`);
    }

    // Move the rule options JSON Schema dump (a single, platform-independent
    // artifact named `rule-schemas`, uploaded once from the `build` job's
    // linux-amd64 leg) to the fixed path
    // the generate-rule-option-types rslib plugin reads — see
    // packages/rslint/rstack.config.ts's onAfterBuild hook, which runs during
    // publish-npm's `build:js` step.
    const ruleSchemasTarget = path.join(
      'packages',
      'rslint',
      'rule-schemas.json',
    );
    fs.copyFileSync(ruleSchemasSrc, ruleSchemasTarget);
    console.log(`Copied ${ruleSchemasSrc} to ${ruleSchemasTarget}`);

    console.log('Artifact move process completed successfully!');
  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  }
}

moveArtifacts();
