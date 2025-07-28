#!/usr/bin/env zx
import fs from 'fs';

$.verbose = true;
async function publish_all() {
  const version = JSON.parse(fs.readFileSync('./packages/vscode-extension/package.json', 'utf-8')).version;
  await $`pnpm run --filter @rslint/core build`;
  await $`pnpm run --filter rslint build`;
  const platforms = [
    { os: 'darwin', arch: 'amd64', 'node-arch': 'x64' },
    { os: 'darwin', arch: 'arm64', 'node-arch': 'arm64' },
    { os: 'linux', arch: 'amd64', 'node-arch': 'x64' },
    { os: 'linux', arch: 'arm64', 'node-arch': 'arm64' },
    { os: 'windows', arch: 'amd64', 'node-arch': 'x64', 'node-os': 'win32' },
    { os: 'windows', arch: 'arm64', 'node-arch': 'arm64', 'node-os': 'win32' },
  ];
  for (const platform of platforms) {
   console.log(`Start Publishing for ${platform.os}-${platform.arch}`);
   await $`GOOS=${platform.os} GOARCH=${platform.arch} go build -o ./packages/vscode-extension/dist ./cmd/rslint`;
   const os = platform['node-os'] || platform.os;
   const arch = platform['node-arch'] || platform.arch;
   await $`cd packages/vscode-extension && vsce package --target ${os}-${arch}`;
   await $`cd packages/vscode-extension && vsce publish --packagePath ./rslint-${os}-${arch}-${version}.vsix `;
   console.log(`Finish Publishing for ${os}-${arch}`);
  }
}
publish_all()