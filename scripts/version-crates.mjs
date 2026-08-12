#!/usr/bin/env zx

import fs from 'node:fs/promises';
import { argv, chalk } from 'zx';
import {
  TSGO_CRATE,
  TSGO_NPM_PACKAGES,
  readTsgoReleaseState,
  replaceCargoLockVersion,
  replaceCargoManifestVersion,
  validateTsgoReleaseState,
} from './tsgo-release-utils.mjs';

// Version the complete crate release line: tsgo-client and all related
// @rslint/tsgo-server npm packages.

const bumpType = argv._[0];

if (!bumpType || !['major', 'minor', 'patch'].includes(bumpType)) {
  console.error(chalk.red('❌ Usage: pnpm version:crates <major|minor|patch>'));
  console.error(
    chalk.gray(
      '   Canary versions are unsupported because crates.io releases cannot be deleted.',
    ),
  );
  process.exit(1);
}

function bumpVersion(version, type) {
  const match = version.match(/^(\d+)\.(\d+)\.(\d+)$/);
  if (!match) {
    throw new Error(`unsupported current tsgo version: ${version}`);
  }
  const [, major, minor, patch] = match.map(Number);
  switch (type) {
    case 'major':
      return `${major + 1}.0.0`;
    case 'minor':
      return `${major}.${minor + 1}.0`;
    case 'patch':
      return `${major}.${minor}.${patch + 1}`;
    default:
      throw new Error(`invalid bump type: ${type}`);
  }
}

async function main() {
  // Validate every manifest and Cargo.lock before writing anything. This keeps
  // an accidental pre-existing mismatch from being silently normalized.
  const state = await readTsgoReleaseState();
  const oldVersion = validateTsgoReleaseState(state);
  const newVersion = bumpVersion(oldVersion, bumpType);

  const crateManifest = replaceCargoManifestVersion(
    state.crateManifestRaw,
    newVersion,
  );
  const cargoLock = replaceCargoLockVersion(state.cargoLockRaw, newVersion);
  const npmManifests = state.npmPackages.map(({ expected, manifest }) => ({
    manifestPath: expected.manifestPath,
    content: `${JSON.stringify({ ...manifest, version: newVersion }, null, 2)}\n`,
    name: manifest.name,
  }));

  await Promise.all([
    fs.writeFile(TSGO_CRATE.manifestPath, crateManifest),
    fs.writeFile(TSGO_CRATE.lockPath, cargoLock),
    ...npmManifests.map(({ manifestPath, content }) =>
      fs.writeFile(manifestPath, content),
    ),
  ]);

  console.log(
    chalk.green(`✅ ${TSGO_CRATE.name}: ${oldVersion} → ${newVersion}`),
  );
  for (const { name } of npmManifests) {
    console.log(chalk.green(`✅ ${name}: ${oldVersion} → ${newVersion}`));
  }

  console.log(chalk.blue('\nNext steps:'));
  console.log(chalk.gray('  1. Commit the shared tsgo version bump'));
  console.log(
    chalk.gray(
      '  2. Run the "📦 Release tsgo" workflow for npm and crates.io together',
    ),
  );
}

main().catch((error) => {
  console.error(chalk.red('❌ Failed to version tsgo:'), error.message);
  process.exit(1);
});
