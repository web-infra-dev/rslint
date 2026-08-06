#!/usr/bin/env zx
import fs from 'node:fs';
import { $, argv } from 'zx';

const marketplace = argv.marketplace || 'vsce';
const prerelease = process.argv.includes('--prerelease');
const skipPackage = process.argv.includes('--skip-package');
const dryRun = process.argv.includes('--dry-run');

if (marketplace !== 'vsce' && marketplace !== 'ovsx') {
  throw new Error(`Unsupported extension marketplace: ${String(marketplace)}`);
}

$.verbose = true;

async function publish() {
  const manifest = JSON.parse(
    fs.readFileSync('./packages/vscode-extension/package.json', 'utf8'),
  );
  const version = manifest.version;
  const fileName = `rslint-${version}.vsix`;
  const packagePath = `packages/vscode-extension/${fileName}`;
  const prereleaseFlag = prerelease ? ['--pre-release'] : [];

  if (!skipPackage) {
    await $`pnpm run --filter rslint build`;
    await $`cd packages/vscode-extension && pnpm vsce package --out ${fileName} ${prereleaseFlag}`;
  } else if (!fs.existsSync(packagePath)) {
    throw new Error(
      `--skip-package requested but ${packagePath} does not exist`,
    );
  }

  const listing = (await $`unzip -Z1 ${packagePath}`).stdout.split('\n');
  const required = ['extension/package.json', 'extension/dist/main.js'];
  const forbiddenPrefixes = [
    'extension/dist/rslint',
    'extension/dist/eslint-plugin/',
    'extension/node_modules/@rslint/',
  ];
  const missing = required.filter((entry) => !listing.includes(entry));
  const forbidden = listing.filter(
    (entry) =>
      forbiddenPrefixes.some((prefix) => entry.startsWith(prefix)) ||
      /\/editor-runtime\.js$/.test(entry) ||
      entry.endsWith('.node') ||
      /\/rslint(?:\.exe)?$/.test(entry) ||
      /\/lint-worker\.js$/.test(entry) ||
      entry.includes('/eslint-plugin/'),
  );
  if (missing.length > 0 || forbidden.length > 0) {
    throw new Error(
      `universal VSIX smoke check failed: missing=[${missing.join(', ')}], forbidden=[${forbidden.join(', ')}]`,
    );
  }
  console.log(`Universal VSIX smoke check passed: ${packagePath}`);

  if (dryRun) {
    console.log('Dry run: skipping marketplace publish');
    return;
  }
  await $`cd packages/vscode-extension && pnpm ${marketplace} publish --packagePath ${fileName} ${prereleaseFlag}`;
  console.log(`Published rslint v${version} to ${marketplace}`);
}

await publish();
