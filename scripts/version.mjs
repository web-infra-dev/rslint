#!/usr/bin/env zx

import { $, fs, path, glob, chalk } from 'zx';

// Validate argument
const bumpType = process.argv[3];
const canaryMode = process.argv[4]; // Optional: 'commit' or 'timestamp' (default)

if (!bumpType || !['major', 'minor', 'patch', 'canary'].includes(bumpType)) {
  console.error(
    chalk.red(
      '❌ Usage: zx scripts/version.mjs <major|minor|patch|canary> [commit|timestamp]',
    ),
  );
  console.error(
    chalk.gray('   For canary: timestamp (default) or commit hash'),
  );
  process.exit(1);
}

console.log(chalk.blue(`🚀 Bumping all package versions: ${bumpType}`));

/**
 * Bump semantic version
 * @param {string} version - Current version (e.g., "1.2.3")
 * @param {string} type - Bump type ("major", "minor", "patch", "canary")
 * @param {string} mode - For canary: "commit" or "timestamp" (default)
 * @returns {Promise<string>} - New version
 */
async function bumpVersion(version, type, mode = 'timestamp') {
  // Remove existing prerelease identifiers for base version calculation
  const baseVersion = version.split('-')[0];
  const [major, minor, patch] = baseVersion.split('.').map(Number);

  switch (type) {
    case 'major':
      return `${major + 1}.0.0`;
    case 'minor':
      return `${major}.${minor + 1}.0`;
    case 'patch':
      return `${major}.${minor}.${patch + 1}`;
    case 'canary': {
      // For canary, we bump patch and add canary suffix
      let identifier;
      if (mode === 'commit') {
        try {
          // Get short commit hash
          const commitHash = await $`git rev-parse --short HEAD`;
          identifier = commitHash.stdout.trim();
        } catch (error) {
          console.warn(
            chalk.yellow('⚠️  Failed to get commit hash, using timestamp'),
          );
          identifier = Math.floor(Date.now() / 1000);
        }
      } else {
        // Use timestamp (default)
        identifier = Math.floor(Date.now() / 1000);
      }
      return `${major}.${minor}.${patch + 1}-canary.${identifier}`;
    }
    default:
      throw new Error(`Invalid bump type: ${type}`);
  }
}

/**
 * Update package.json version
 * @param {string} packagePath - Path to package.json
 * @param {string} newVersion - New version to set
 */
async function updatePackageVersion(packagePath, newVersion) {
  const packageJson = JSON.parse(await fs.readFile(packagePath, 'utf8'));
  const oldVersion = packageJson.version;

  packageJson.version = newVersion;

  await fs.writeFile(packagePath, JSON.stringify(packageJson, null, 2) + '\n');

  console.log(
    chalk.green(
      `✅ ${path.basename(path.dirname(packagePath))}: ${oldVersion} → ${newVersion}`,
    ),
  );

  return { oldVersion, newVersion, name: packageJson.name };
}

/**
 * Update workspace dependencies to use new versions
 * @param {string} packagePath - Path to package.json
 * @param {Object} versionMap - Map of package names to new versions
 */
async function updateWorkspaceDependencies(packagePath, versionMap) {
  const packageJson = JSON.parse(await fs.readFile(packagePath, 'utf8'));
  let updated = false;

  // Update dependencies
  for (const depType of [
    'dependencies',
    'devDependencies',
    'peerDependencies',
    'optionalDependencies',
  ]) {
    const deps = packageJson[depType];
    if (!deps) continue;

    for (const [depName, depVersion] of Object.entries(deps)) {
      // Update workspace dependencies
      if (depVersion.startsWith('workspace:') && versionMap[depName]) {
        // Keep workspace protocol but update for reference
        console.log(
          chalk.yellow(
            `🔄 Workspace dep in ${path.basename(path.dirname(packagePath))}: ${depName}`,
          ),
        );
        updated = true;
      }
    }
  }

  if (updated) {
    await fs.writeFile(
      packagePath,
      JSON.stringify(packageJson, null, 2) + '\n',
    );
  }
}

async function main() {
  try {
    // Find all package.json files in the workspace
    const rootPackagePath = path.join(process.cwd(), 'package.json');
    const workspacePackagePaths = await glob('packages/*/package.json', {
      cwd: process.cwd(),
    });
    const npmRslintPackagePaths = await glob('npm/rslint/*/package.json', {
      cwd: process.cwd(),
    });
    const npmTsgoPackagePaths = await glob('npm/tsgo/*/package.json', {
      cwd: process.cwd(),
    });

    const allPackagePaths = [
      rootPackagePath,
      ...workspacePackagePaths.map(p => path.join(process.cwd(), p)),
      ...npmRslintPackagePaths.map(p => path.join(process.cwd(), p)),
      ...npmTsgoPackagePaths.map(p => path.join(process.cwd(), p)),
    ];

    console.log(
      chalk.blue(`📦 Found ${allPackagePaths.length} packages to update`),
    );
    console.log(chalk.gray(`  - Root: 1`));
    console.log(
      chalk.gray(`  - Workspace packages: ${workspacePackagePaths.length}`),
    );
    console.log(
      chalk.gray(`  - NPM rslint packages: ${npmRslintPackagePaths.length}`),
    );
    console.log(
      chalk.gray(`  - NPM tsgo packages: ${npmTsgoPackagePaths.length}`),
    );

    // Check current versions to find the highest one for unification
    const currentVersions = [];
    for (const packagePath of allPackagePaths) {
      const packageJson = JSON.parse(await fs.readFile(packagePath, 'utf8'));
      currentVersions.push(packageJson.version);
    }

    // Find the highest current version
    const highestVersion = currentVersions
      .map(v => {
        // Extract base version (remove prerelease identifiers)
        const baseVersion = v.split('-')[0];
        return baseVersion.split('.').map(Number);
      })
      .reduce((max, current) => {
        for (let i = 0; i < 3; i++) {
          if (current[i] > max[i]) return current;
          if (current[i] < max[i]) return max;
        }
        return max;
      })
      .join('.');

    console.log(
      chalk.yellow(
        `🔍 Current versions found: ${[...new Set(currentVersions)].join(', ')}`,
      ),
    );
    console.log(
      chalk.yellow(`📌 Unifying to highest version: ${highestVersion}`),
    );

    // Calculate the new version from the highest current version
    const newVersion = await bumpVersion(highestVersion, bumpType, canaryMode);
    console.log(chalk.green(`🎯 Target version: ${newVersion}`));

    // First pass: bump all versions to the new unified version
    const versionMap = {};
    const updates = [];

    for (const packagePath of allPackagePaths) {
      const packageJson = JSON.parse(await fs.readFile(packagePath, 'utf8'));
      const oldVersion = packageJson.version;

      const result = await updatePackageVersion(packagePath, newVersion);
      updates.push(result);

      if (packageJson.name) {
        versionMap[packageJson.name] = newVersion;
      }
    }

    // Second pass: update workspace dependencies
    console.log(chalk.blue('\n🔗 Updating workspace dependencies...'));

    for (const packagePath of allPackagePaths) {
      await updateWorkspaceDependencies(packagePath, versionMap);
    }

    // Summary
    console.log(chalk.green('\n✨ Version bump completed!'));
    console.log(chalk.gray('📋 Summary:'));

    updates.forEach(({ name, oldVersion, newVersion }) => {
      const displayName = name || 'root';
      console.log(
        chalk.gray(`  ${displayName}: ${oldVersion} → ${newVersion}`),
      );
    });

    // Additional steps
    console.log(chalk.blue('\n🔧 Next steps:'));
    console.log(chalk.gray('  1. Run: pnpm install (to update lockfile)'));
    console.log(
      chalk.gray('  2. Run: pnpm run build (to build with new versions)'),
    );
    console.log(
      chalk.gray('  3. Run: pnpm run test (to verify everything works)'),
    );

    if (bumpType === 'canary') {
      console.log(
        chalk.gray('  4. Publish canary: pnpm run release --tag canary'),
      );
      console.log(
        chalk.yellow(
          '\n⚠️  Canary versions should be published with the "canary" tag',
        ),
      );
      console.log(
        chalk.yellow('   Or use: pnpm run release:canary (automated)'),
      );
    } else {
      console.log(chalk.gray('  4. Commit changes and create release'));
    }
  } catch (error) {
    console.error(chalk.red('❌ Error during version bump:'), error.message);
    process.exit(1);
  }
}

main();
