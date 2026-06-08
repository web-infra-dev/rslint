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

  // build:js is one rslib build (library surface + eslint-plugin worker); the
  // worker needs @rslint/native's index.d.ts from `napi build`.
  await $`pnpm run --filter @rslint/native build`;
  await $`pnpm run --filter @rslint/core build:js`;
  await $`pnpm run --filter rslint build`;

  for (const platform of platforms) {
    const os = platform['node-os'] || platform.os;
    const arch = platform['node-arch'] || platform.arch;

    console.log(`Start Publishing for ${os}-${arch}`);

    await $`rm -rf ./packages/vscode-extension/dist/rslint`;
    await $`rm -rf ./packages/vscode-extension/dist/rslint.exe`;
    const targetFilename = os == 'win32' ? 'rslint.exe' : 'rslint';
    await $`cp binaries/${os}-${arch}-rslint/${os}-${arch}-rslint ./packages/vscode-extension/dist/${targetFilename}`;

    await $`chmod +x ./packages/vscode-extension/dist/${targetFilename}`;

    // napi parser `.node` for the eslint-plugin worker. `build.js` already
    // staged the publish-runner's host `.node` as a fixed `rslint.node`;
    // overwrite it with THIS platform's `.node` so the worker's
    // `@rslint/native` shim loads the matching ABI. Source artifact comes from
    // the `napi-build` job (`native-<tuple>`, downloaded to `binaries/`); the
    // shim requires a constant `./rslint.node`, so one filename is overwritten
    // per iteration — no wrong-arch leak is possible. Linux ships gnu only
    // (musl is not a vsce target).
    const napiTuple =
      os === 'win32'
        ? `win32-${arch}-msvc`
        : os === 'linux'
          ? `linux-${arch}-gnu`
          : `${os}-${arch}`; // darwin-x64 / darwin-arm64
    const nativeDir =
      './packages/vscode-extension/dist/eslint-plugin/node_modules/@rslint/native';
    await $`rm -f ${nativeDir}/rslint.node`;
    await $`cp binaries/native-${napiTuple}/rslint.${napiTuple}.node ${nativeDir}/rslint.node`;

    await $`ls -lR ./packages/vscode-extension/dist`;
    const prereleaseFlag = prerelease ? ['--pre-release'] : [];

    await $`cd packages/vscode-extension && pnpm vsce package --target ${os}-${arch} ${prereleaseFlag}`;

    // Smoke-check the produced vsix: the eslint-plugin worker payload + its
    // nested native shim/.node must be present, or the packaged extension's
    // plugin host silently dies (the dev-only blind spot this whole change
    // fixes). `vsce ls` is unusable here (its npm dep walk breaks under pnpm),
    // so inspect the zip directly. Cross-build arch correctness is guaranteed
    // upstream by the `cp` from `native-${napiTuple}` (it errors if absent).
    const vsix = `packages/vscode-extension/rslint-${os}-${arch}-${version}.vsix`;
    const listing = (await $`unzip -Z1 ${vsix}`).stdout;
    const requiredEntries = [
      'extension/dist/eslint-plugin/index.js',
      'extension/dist/eslint-plugin/lint-worker.js',
      'extension/dist/eslint-plugin/package.json',
      'extension/dist/eslint-plugin/node_modules/@rslint/native/index.js',
      'extension/dist/eslint-plugin/node_modules/@rslint/native/rslint.node',
    ];
    const missing = requiredEntries.filter((e) => !listing.includes(e));
    if (missing.length > 0) {
      throw new Error(
        `vsix smoke check failed for ${os}-${arch}: missing ${missing.join(', ')}`,
      );
    }
    console.log(`vsix worker-payload smoke check passed for ${os}-${arch}`);

    // supports dry-run
    if (process.argv.includes('--dry-run')) {
      console.log(`Dry run: Skipping actual publish for ${os}-${arch}`);
      continue;
    }
    await $`cd packages/vscode-extension && pnpm ${marketplace} publish --packagePath ./rslint-${os}-${arch}-${version}.vsix ${prereleaseFlag}`;
    console.log(`Finish Publishing v${version} for ${os}-${arch}.`);
  }
}

publish_all();
