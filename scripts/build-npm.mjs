#!/usr/bin/env zx
import 'zx/globals';
import { join } from 'node:path';
import { platform, arch } from 'node:os';

$.verbose = true;
$.cwd = join(__dirname, '..');

/**
 * @typedef {Object} PlatformConfig
 * @property {string} GOOS - Go operating system
 * @property {string} GOARCH - Go architecture
 * @property {string} ext - File extension for the binary
 */

/**
 * @typedef {Object} PlatformConfigWithBinPath
 * @property {string} GOOS - Go operating system
 * @property {string} GOARCH - Go architecture
 * @property {string} ext - File extension for the binary
 * @property {string} binPath - Full binary path with platform-specific naming
 */

/**
 * @type {Record<string, Record<string, PlatformConfig>>}
 */
const platformMap = {
  darwin: {
    x64: { GOOS: 'darwin', GOARCH: 'amd64', ext: '' },
    arm64: { GOOS: 'darwin', GOARCH: 'arm64', ext: '' },
  },
  linux: {
    x64: { GOOS: 'linux', GOARCH: 'amd64', ext: '' },
    arm64: { GOOS: 'linux', GOARCH: 'arm64', ext: '' },
  },
  win32: {
    x64: { GOOS: 'windows', GOARCH: 'amd64', ext: '.exe' },
    arm64: { GOOS: 'windows', GOARCH: 'arm64', ext: '.exe' },
  },
};

/**
 * Get platform configuration for building Go binaries
 * @param {string} [platform=process.platform()] - The target platform
 * @param {string} [arch=os.arch()] - The target architecture
 * @returns {PlatformConfigWithBinPath} Platform configuration with binary path
 * @throws {Error} When platform/architecture combination is not supported
 */
function getPlatformConfig(platform = platform(), arch = arch()) {
  const config = platformMap[platform]?.[arch];
  if (!config) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }

  const { GOOS, GOARCH, ext } = config;

  return {
    GOOS,
    GOARCH,
    ext,
  };
}

async function buildForTarget(targetPlatform, targetArch) {
  const { GOOS, GOARCH, ext } = getPlatformConfig(targetPlatform, targetArch);

  console.log(`Building for ${GOOS}/${GOARCH}...`);
  await $`GOOS=${GOOS} GOARCH=${GOARCH} go build -o npm/${targetPlatform}-${targetArch}/rslint${ext} ./cmd/rslint`;
}

const args = process.argv.slice(2);
const command = args[1];

switch (command) {
  case 'current':
    buildForTarget(platform(), arch());
    break;

  case 'all':
    console.log('Building for all supported platforms...');
    await Promise.all(
      Object.entries(platformMap).flatMap(([platform, architectures]) =>
        Object.keys(architectures).map(async arch => {
          return buildForTarget(platform, arch).catch(error => {
            console.error(
              `Failed to build ${platform}-${arch}:`,
              error.message,
            );
          });
        }),
      ),
    );
    break;

  case 'darwin-x64':
  case 'darwin-arm64':
  case 'linux-x64':
  case 'linux-arm64':
  case 'win32-x64':
  case 'win32-arm64':
    const [targetPlatform, targetArch] = command.split('-');
    buildForTarget(targetPlatform, targetArch);
    break;

  default:
    console.log('Usage: node build-platform.js <command>');
    console.log('Commands:');
    console.log('  current           - Build for current platform');
    console.log('  all              - Build for all platforms');
    console.log('  darwin-x64       - Build for macOS Intel');
    console.log('  darwin-arm64     - Build for macOS Apple Silicon');
    console.log('  linux-x64        - Build for Linux x64');
    console.log('  linux-arm64      - Build for Linux ARM64');
    console.log('  win32-x64        - Build for Windows x64');
    console.log('  win32-arm64      - Build for Windows ARM64');
    process.exit(1);
}
