#!/usr/bin/env zx
import fs from 'node:fs';
import { $, argv } from 'zx';

const marketplace = argv.marketplace || 'vsce';
const prerelease = argv.prerelease || false;

$.verbose = true;

async function publish() {
  const version = JSON.parse(
    fs.readFileSync('./packages/vscode-extension/package.json', 'utf8'),
  ).version;
  const fileName = `rslint-${version}.vsix`;
  const packagePath = `packages/vscode-extension/${fileName}`;
  const prereleaseFlag = prerelease ? ['--pre-release'] : [];

  await $`pnpm --filter rslint build`;
  fs.rmSync(packagePath, { force: true });
  await $`cd packages/vscode-extension && pnpm vsce package --out ${fileName} ${prereleaseFlag}`;

  const listing = (await $`unzip -Z1 ${packagePath}`).stdout.split('\n');
  if (!listing.includes('extension/dist/main.js')) {
    throw new Error('vsix smoke check failed: dist/main.js is missing');
  }
  const forbidden = listing.filter(
    (entry) =>
      entry.startsWith('extension/dist/eslint-plugin/') ||
      entry === 'extension/dist/rslint' ||
      entry === 'extension/dist/rslint.exe' ||
      entry.endsWith('.node'),
  );
  if (forbidden.length > 0) {
    throw new Error(
      `vsix smoke check failed: bundled runtime payload: ${forbidden.join(', ')}`,
    );
  }
  console.log('universal vsix smoke check passed');

  if (process.argv.includes('--dry-run')) {
    console.log('Dry run: skipping marketplace publish');
    return;
  }
  await $`cd packages/vscode-extension && pnpm ${marketplace} publish --packagePath ./${fileName} ${prereleaseFlag}`;
}

publish();
