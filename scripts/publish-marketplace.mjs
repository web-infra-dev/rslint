#!/usr/bin/env zx
import fs from 'fs';
import { argv } from 'zx';

const marketplace = argv.marketplace || 'vsce';
const prerelease = argv.prerelease || false;

$.verbose = true;

async function publish_all() {
  const version = JSON.parse(
    fs.readFileSync('./packages/vscode-extension/package.json', 'utf-8'),
  ).version;

  const platforms = [
    { os: 'darwin', arch: 'amd64', 'node-arch': 'x64' },
    { os: 'darwin', arch: 'arm64', 'node-arch': 'arm64' },
    { os: 'linux', arch: 'amd64', 'node-arch': 'x64' },
    { os: 'linux', arch: 'arm64', 'node-arch': 'arm64' },
    { os: 'windows', arch: 'amd64', 'node-arch': 'x64', 'node-os': 'win32' },
    { os: 'windows', arch: 'arm64', 'node-arch': 'arm64', 'node-os': 'win32' },
  ];

  await $`pnpm run --filter @rslint/core build:js`;
  await $`pnpm run --filter rslint build`;

  for (const platform of platforms) {
    const os = platform['node-os'] || platform.os;
    const arch = platform['node-arch'] || platform.arch;

    console.log(`Start Publishing for ${os}-${arch}`);

    await $`rm -rf ./packages/vscode-extension/dist/rslint`;
    await $`rm -rf ./packages/vscode-extension/dist/rslint.exe`;
    await $`cp binaries/${os}-${arch}-rslint/${os}-${arch}-rslint ./packages/vscode-extension/dist/rslint`;

    await $`cd packages/vscode-extension && pnpm ${marketplace} package --target ${os}-${arch}`;
    await $`ls -R packages/vscode-extension`;

    // supports dry-run
    if (process.argv.includes('--dry-run')) {
      console.log(`Dry run: Skipping actual publish for ${os}-${arch}`);
      continue;
    }
    await $`cd packages/vscode-extension && echo pnpm ${marketplace} publish --packagePath ./rslint-${os}-${arch}-${version}.vsix ${prerelease ? '--pre-release' : ''}`;
    console.log(`Finish Publishing v${version} for ${os}-${arch}.`);
  }
}

publish_all();
