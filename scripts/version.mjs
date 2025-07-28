#!/usr/bin/env zx

import { $, fs, path, glob, chalk } from 'zx';

// Validate argument
const bumpType = process.argv[3];
if (!bumpType || !['major', 'minor', 'patch'].includes(bumpType)) {
  console.error(
    chalk.red('❌ Usage: zx scripts/version.mjs <major|minor|patch>'),
  );
  process.exit(1);
}

console.log(chalk.blue(`🚀 Bumping all package versions: ${bumpType}`));

/**
 * Bump semantic version
 * @param {string} version - Current version (e.g., "1.2.3")
 * @param {string} type - Bump type ("major", "minor", "patch")
 * @returns {string} - New version
 */
function bumpVersion(version, type) {
  const [major, minor, patch] = version.split('.').map(Number);

  switch (type) {
    case 'major':
      return `${major + 1}.0.0`;
    case 'minor':
      return `${major}.${minor + 1}.0`;
    case 'patch':
      return `${major}.${minor}.${patch + 1}`;
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
    "optionalDependencies",
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
    const npmPackagePaths = await glob('npm/*/package.json', {
      cwd: process.cwd(),
    });

    const allPackagePaths = [
      rootPackagePath,
      ...workspacePackagePaths.map(p => path.join(process.cwd(), p)),
      ...npmPackagePaths.map(p => path.join(process.cwd(), p)),
    ];

    console.log(
      chalk.blue(`📦 Found ${allPackagePaths.length} packages to update`),
    );
    console.log(chalk.gray(`  - Root: 1`));
    console.log(
      chalk.gray(`  - Workspace packages: ${workspacePackagePaths.length}`),
    );
    console.log(chalk.gray(`  - NPM packages: ${npmPackagePaths.length}`));

    // Check current versions to find the highest one for unification
    const currentVersions = [];
    for (const packagePath of allPackagePaths) {
      const packageJson = JSON.parse(await fs.readFile(packagePath, 'utf8'));
      currentVersions.push(packageJson.version);
    }

    // Find the highest current version
    const highestVersion = currentVersions
      .map(v => v.split('.').map(Number))
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
    const newVersion = bumpVersion(highestVersion, bumpType);
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
    console.log(chalk.gray('  4. Commit changes and create release'));
  } catch (error) {
    console.error(chalk.red('❌ Error during version bump:'), error.message);
    process.exit(1);
  }
}

main();
