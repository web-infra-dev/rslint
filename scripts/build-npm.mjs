#!/usr/bin/env zx
import fs from 'fs';
$.verbose = true;

// build binary for following platforms
// #   darwin/amd64
// #   darwin/arm64
// #   linux/amd64
// #   linux/arm64
// #   windows/amd64
// #   windows/arm64

const platforms = [
  { os: 'darwin', arch: 'amd64', 'node-arch': 'x64' },
  { os: 'darwin', arch: 'arm64', 'node-arch': 'arm64' },
  { os: 'linux', arch: 'amd64', 'node-arch': 'x64' },
  { os: 'linux', arch: 'arm64', 'node-arch': 'arm64' },
  { os: 'windows', arch: 'amd64', 'node-arch': 'x64', 'node-os': 'win32' },
  { os: 'windows', arch: 'arm64', 'node-arch': 'arm64', 'node-os': 'win32' },
];

async function build_rslint() {
  for (const platform of platforms) {
    const nodeOs = platform['node-os'] || platform.os;
    const nodeArch = platform['node-arch'];
    const outputDir = `npm/rslint/${nodeOs}-${nodeArch}`;
    const ext = platform.os === 'windows' ? '.exe' : '';
    await $`GOOS=${platform.os} GOARCH=${platform.arch} go build -o ${outputDir}/rslint${ext} ./cmd/rslint`;
  }
}

async function build_tsgo() {
  for (const platform of platforms) {
    const nodeOs = platform['node-os'] || platform.os;
    const nodeArch = platform['node-arch'];
    const outputDir = `npm/tsgo/${nodeOs}-${nodeArch}`;
    const ext = platform.os === 'windows' ? '.exe' : '';
    await $`GOOS=${platform.os} GOARCH=${platform.arch} go build -o ${outputDir}/tsgo${ext} ./cmd/tsgo`;
  }
}

async function main() {
  const target = argv._[0];
  if (target === 'rslint') {
    await build_rslint();
  } else if (target === 'tsgo') {
    await build_tsgo();
  } else {
    // Build all by default
    await build_rslint();
    await build_tsgo();
  }
}

main();
