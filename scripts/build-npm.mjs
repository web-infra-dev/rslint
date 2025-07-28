#!/usr/bin/env zx
import fs from 'fs';
$.verbose = true;

// build binary for following platforms
// #   darwin/amd64
// #   darwin/arm64
// #   linux/amd64
// #   linux/arm64
// #   windows/amd64
async function build_all() {
  const platforms = [
    { os: 'darwin', arch: 'amd64', 'node-arch': 'x64' },
    { os: 'darwin', arch: 'arm64', 'node-arch': 'arm64' },
    { os: 'linux', arch: 'amd64', 'node-arch': 'x64' },
    { os: 'linux', arch: 'arm64', 'node-arch': 'arm64' },
    { os: 'windows', arch: 'amd64', 'node-arch': 'x64', 'node-os': 'win32' },
  ];
  for (const platform of platforms) {
    await $`GOOS=${platform.os} GOARCH=${platform.arch} go build -o npm/${platform['node-os'] || platform.os}-${platform['node-arch']}/rslint ./cmd/rslint`;
  }
}
async function main() {
  await build_all();
}

main();
