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
    // Move Go CLI binaries into the @rslint/native-{tuple} subpackages (flat,
    // alongside the .node). Artifacts are named `{platform}-rslint`. Go
    // binaries are statically linked, so one linux build serves both glibc and
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

    for (const file of rslintFiles) {
      console.log(`Processing ${file}`);
      const isWindows = file.includes('win32');
      const filename = path.basename(file);
      const platform = filename.replace(/-rslint$/, ''); // e.g. linux-x64
      const tuples = platformToTuples[platform];
      if (!tuples) {
        console.log(`Warning: no tuple mapping for ${platform}, skipping`);
        continue;
      }
      const binName = isWindows ? 'rslint.exe' : 'rslint';

      for (const tuple of tuples) {
        const targetDir = path.join('npm', 'rslint', tuple);
        const targetFile = path.join(targetDir, binName);
        fs.mkdirSync(targetDir, { recursive: true });
        fs.copyFileSync(file, targetFile);
        fs.chmodSync(targetFile, 0o755); // Make executable
        console.log(`Copied ${file} to ${targetFile}`);
      }
    }

    // Find and move tsgo binaries to lib directory
    const tsgoFiles = findBinaries('binaries', '-tsgo');
    console.log(`Found ${tsgoFiles.length} tsgo binary files`);

    for (const file of tsgoFiles) {
      // Skip tsgo-built directories
      if (file.includes('-tsgo-built')) {
        continue;
      }
      console.log(`Processing ${file}`);
      const isWindows = file.includes('win32');
      const filename = path.basename(file);
      const dirname = filename.replace(/-tsgo$/, '');
      const targetDir = path.join('npm', 'tsgo', dirname, 'lib');

      const targetFile = path.join(targetDir, isWindows ? 'tsgo.exe' : 'tsgo');

      // Create target directory and copy file
      fs.mkdirSync(targetDir, { recursive: true });
      fs.copyFileSync(file, targetFile);
      fs.chmodSync(targetFile, 0o755); // Make executable

      console.log(`Copied ${file} to ${targetFile}`);
    }

    // Copy typescript-go built files (lib files) to tsgo platform packages
    // Files are downloaded from platform-specific artifacts to binaries/{platform}-tsgo-built/
    const tsgoPlatforms = [
      'darwin-arm64',
      'darwin-x64',
      'linux-arm64',
      'linux-x64',
      'win32-arm64',
      'win32-x64',
    ];

    for (const platform of tsgoPlatforms) {
      const tsgoBuiltSource = path.join(
        'binaries',
        `${platform}-tsgo-built`,
        'local',
      );

      if (!fs.existsSync(tsgoBuiltSource)) {
        console.log(
          `Warning: typescript-go built source not found at ${tsgoBuiltSource}`,
        );
        continue;
      }

      // Get all files from the built/local directory
      const libFiles = fs.readdirSync(tsgoBuiltSource);
      console.log(`Found ${libFiles.length} lib files for ${platform}`);

      const targetLibDir = path.join('npm', 'tsgo', platform, 'lib');
      fs.mkdirSync(targetLibDir, { recursive: true });

      for (const file of libFiles) {
        // Skip tsgo binary from typescript-go build, we use our own ./cmd/tsgo build
        if (file === 'tsgo' || file === 'tsgo.exe') {
          continue;
        }
        const srcFile = path.join(tsgoBuiltSource, file);
        const destFile = path.join(targetLibDir, file);
        const stat = fs.statSync(srcFile);
        if (stat.isFile()) {
          fs.copyFileSync(srcFile, destFile);
        }
      }
      console.log(`Copied built files to ${targetLibDir}`);
    }

    // Move napi `.node` parser binaries into the @rslint/native-{tuple}
    // subpackages (flat, alongside the Go binary). Artifacts are named
    // `rslint.{tuple}.node`; target is `npm/rslint/{tuple}/` (libc suffix kept
    // so gnu/musl stay separate). Not chmod'd — a `.node` is dlopen'd.
    const nodeFiles = findBinaries('binaries', 'rslint.');
    console.log(`Found ${nodeFiles.length} napi .node files`);

    for (const file of nodeFiles) {
      console.log(`Processing ${file}`);
      const filename = path.basename(file); // rslint.linux-x64-gnu.node
      const tuple = filename.replace(/^rslint\./, '').replace(/\.node$/, ''); // linux-x64-gnu
      const targetDir = path.join('npm', 'rslint', tuple);
      const targetFile = path.join(targetDir, filename);

      fs.mkdirSync(targetDir, { recursive: true });
      fs.copyFileSync(file, targetFile);

      console.log(`Copied ${file} to ${targetFile}`);
    }

    // Move the rule options JSON Schema dump (a single, platform-independent
    // artifact named `rule-schemas`, uploaded once from the `build` job's
    // linux-amd64 leg) to the fixed path
    // scripts/generate-rule-option-types.mjs reads — see
    // packages/rslint/rslib.config.ts's onAfterBuild hook, which runs during
    // publish-npm's `build:js` step.
    const ruleSchemasSrc = path.join(
      'binaries',
      'rule-schemas',
      'rule-schemas.json',
    );
    if (fs.existsSync(ruleSchemasSrc)) {
      const ruleSchemasTarget = path.join(
        'packages',
        'rslint',
        'rule-schemas.json',
      );
      fs.copyFileSync(ruleSchemasSrc, ruleSchemasTarget);
      console.log(`Copied ${ruleSchemasSrc} to ${ruleSchemasTarget}`);
    } else {
      console.log(`Warning: rule schemas dump not found at ${ruleSchemasSrc}`);
    }

    console.log('Artifact move process completed successfully!');
  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  }
}

moveArtifacts();
