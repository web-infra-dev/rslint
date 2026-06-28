#!/usr/bin/env zx

import { fs, path, chalk } from 'zx';

// Version bump for publishable Rust crates, kept separate from
// scripts/version.mjs: crates are versioned and published to crates.io
// (via .github/workflows/release-crates.yml) on their own line, independent
// of the unified npm package version.
//
// canary is intentionally unsupported here — a crates.io library gains nothing
// from prerelease versions (default `^` ranges never resolve to them) and
// published versions can never be deleted, only yanked.

// Crates bumped by this script. rslint-native is excluded (publish = false).
const PUBLISHABLE_CRATES = ['crates/tsgo-client/Cargo.toml'];
const CARGO_LOCK = 'Cargo.lock';

const bumpType = process.argv[3];

if (!bumpType || !['major', 'minor', 'patch'].includes(bumpType)) {
  console.error(
    chalk.red('❌ Usage: zx scripts/version-crates.mjs <major|minor|patch>'),
  );
  console.error(
    chalk.gray(
      '   canary is not supported for crates (independent crates.io versioning).',
    ),
  );
  process.exit(1);
}

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

// Replace `version = "..."` in the [package] section only, leaving version
// fields under [dependencies] etc. untouched.
function bumpManifest(content, type) {
  const lines = content.split('\n');
  let inPackage = false;
  let name = null;
  let oldVersion = null;
  let newVersion = null;

  for (let i = 0; i < lines.length; i++) {
    if (/^\s*\[/.test(lines[i])) {
      inPackage = lines[i].trim() === '[package]';
      continue;
    }
    if (!inPackage) continue;

    const nameMatch = lines[i].match(/^\s*name\s*=\s*"([^"]+)"/);
    if (nameMatch) name = nameMatch[1];

    if (newVersion === null) {
      const m = lines[i].match(/^(\s*version\s*=\s*")([^"]+)(".*)$/);
      if (m) {
        oldVersion = m[2];
        newVersion = bumpVersion(m[2], type);
        lines[i] = `${m[1]}${newVersion}${m[3]}`;
      }
    }
  }

  if (newVersion === null) {
    throw new Error('no [package] version field found');
  }
  return { content: lines.join('\n'), name, oldVersion, newVersion };
}

// Sync a local workspace member's version in Cargo.lock so
// `cargo publish --locked` stays happy. Local members carry no checksum, so
// only the version line under their `name = "..."` entry needs touching.
function bumpLock(content, crateName, newVersion) {
  const lines = content.split('\n');
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].trim() !== `name = "${crateName}"`) continue;
    for (let j = i + 1; j < lines.length; j++) {
      if (/^\s*\[\[package\]\]/.test(lines[j])) break;
      const m = lines[j].match(/^(\s*version\s*=\s*")([^"]+)(".*)$/);
      if (m) {
        lines[j] = `${m[1]}${newVersion}${m[3]}`;
        return { content: lines.join('\n'), updated: true };
      }
    }
  }
  return { content, updated: false };
}

async function main() {
  const lockPath = path.join(process.cwd(), CARGO_LOCK);
  const hasLock = await fs.pathExists(lockPath);
  let lockContent = hasLock ? await fs.readFile(lockPath, 'utf8') : '';
  let lockDirty = false;

  for (const rel of PUBLISHABLE_CRATES) {
    const manifestPath = path.join(process.cwd(), rel);
    const raw = await fs.readFile(manifestPath, 'utf8');
    const { content, name, oldVersion, newVersion } = bumpManifest(
      raw,
      bumpType,
    );

    await fs.writeFile(manifestPath, content);
    console.log(chalk.green(`✅ ${name}: ${oldVersion} → ${newVersion}`));
    console.log(chalk.gray(`   ${rel}`));

    if (hasLock) {
      const { content: nextLock, updated } = bumpLock(
        lockContent,
        name,
        newVersion,
      );
      if (updated) {
        lockContent = nextLock;
        lockDirty = true;
      } else {
        console.log(
          chalk.yellow(
            `⚠️  ${name} not found in Cargo.lock — run \`cargo update -p ${name}\` to sync`,
          ),
        );
      }
    }
  }

  if (lockDirty) {
    await fs.writeFile(lockPath, lockContent);
    console.log(chalk.green('✅ Cargo.lock synced'));
  }

  console.log(chalk.blue('\n🔧 Next steps:'));
  console.log(
    chalk.gray('  1. Commit the version bump (Cargo.toml + Cargo.lock)'),
  );
  console.log(
    chalk.gray(
      '  2. Publish via the "📦 Release crates" workflow (Actions → workflow_dispatch)',
    ),
  );
}

main().catch((err) => {
  console.error(chalk.red('❌ Error during crate version bump:'), err.message);
  process.exit(1);
});
