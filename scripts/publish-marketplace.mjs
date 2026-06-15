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

  // build:js is one rslib build (library surface + eslint-plugin worker). The
  // worker's native loader resolves the platform .node at runtime, so no napi
  // build is needed here — the per-platform .node comes from the napi-build job
  // (downloaded to binaries/) and is staged per-target in the loop below.
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

    // napi parser `.node` for the eslint-plugin worker. `build.js` staged the
    // publish-runner's HOST platform package; replace it with THIS target's
    // package (`@rslint/native-<napiTuple>`), which the worker's loader resolves
    // at runtime. Source artifact comes from the `napi-build` job
    // (`native-<tuple>`, downloaded to `binaries/`). Linux ships gnu only (musl
    // is not a vsce target; the loader picks gnu under the glibc Electron host).
    const napiTuple =
      os === 'win32'
        ? `win32-${arch}-msvc`
        : os === 'linux'
          ? `linux-${arch}-gnu`
          : `${os}-${arch}`; // darwin-x64 / darwin-arm64
    const nativeRoot =
      './packages/vscode-extension/dist/eslint-plugin/node_modules/@rslint';
    const targetPkgDir = `${nativeRoot}/native-${napiTuple}`;
    // Drop any prior iteration's package (and build.js's host one) — only the
    // current target's package may ship in this vsix.
    await $`rm -rf ${nativeRoot}/native-*`;
    await $`mkdir -p ${targetPkgDir}`;
    fs.writeFileSync(
      `${targetPkgDir}/package.json`,
      `${JSON.stringify(
        {
          name: `@rslint/native-${napiTuple}`,
          exports: { '.': `./rslint.${napiTuple}.node` },
        },
        null,
        2,
      )}\n`,
    );
    await $`cp binaries/native-${napiTuple}/rslint.${napiTuple}.node ${targetPkgDir}/rslint.${napiTuple}.node`;

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
      `extension/dist/eslint-plugin/node_modules/@rslint/native-${napiTuple}/rslint.${napiTuple}.node`,
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
