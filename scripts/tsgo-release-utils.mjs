import fs from 'node:fs/promises';

export const TSGO_CRATE = {
  name: 'tsgo-client',
  manifestPath: 'crates/tsgo-client/Cargo.toml',
  lockPath: 'Cargo.lock',
};

export const TSGO_WRAPPER = {
  name: '@rslint/tsgo-server',
  manifestPath: 'packages/tsgo/package.json',
};

export const TSGO_PLATFORMS = [
  { key: 'darwin-arm64', os: 'darwin', cpu: 'arm64' },
  { key: 'darwin-x64', os: 'darwin', cpu: 'x64' },
  { key: 'linux-arm64', os: 'linux', cpu: 'arm64' },
  { key: 'linux-x64', os: 'linux', cpu: 'x64' },
  { key: 'win32-arm64', os: 'win32', cpu: 'arm64' },
  { key: 'win32-x64', os: 'win32', cpu: 'x64' },
].map((platform) => ({
  ...platform,
  name: `${TSGO_WRAPPER.name}-${platform.key}`,
  manifestPath: `npm/tsgo/${platform.key}/package.json`,
  executable: `./lib/tsgo${platform.os === 'win32' ? '.exe' : ''}`,
}));

export const TSGO_NPM_PACKAGES = [TSGO_WRAPPER, ...TSGO_PLATFORMS];

function parseTomlStringField(section, field, source) {
  const value = section.match(
    new RegExp(`^\\s*${field}\\s*=\\s*"([^"]+)"`, 'm'),
  )?.[1];
  if (!value) {
    throw new Error(`${field} not found in ${source}`);
  }
  return value;
}

function parseCargoManifestPackage(content) {
  const section = content.match(/\[package\]([\s\S]*?)(?=\n\[|$)/)?.[1];
  if (!section) {
    throw new Error(`[package] not found in ${TSGO_CRATE.manifestPath}`);
  }
  return {
    name: parseTomlStringField(section, 'name', TSGO_CRATE.manifestPath),
    version: parseTomlStringField(section, 'version', TSGO_CRATE.manifestPath),
  };
}

function parseCargoLockVersion(content, packageName) {
  const sections = content.split('[[package]]').slice(1);
  for (const section of sections) {
    const name = section.match(/^\s*name\s*=\s*"([^"]+)"/m)?.[1];
    if (name === packageName) {
      return parseTomlStringField(
        section,
        'version',
        `${TSGO_CRATE.lockPath} package ${packageName}`,
      );
    }
  }
  throw new Error(`${packageName} not found in ${TSGO_CRATE.lockPath}`);
}

function sameValues(actual, expected) {
  return (
    Array.isArray(actual) &&
    actual.length === expected.length &&
    actual.every((value, index) => value === expected[index])
  );
}

function compareSets(actual, expected) {
  return (
    actual.size === expected.size &&
    [...actual].every((value) => expected.has(value))
  );
}

export async function readTsgoReleaseState() {
  const [crateManifestRaw, cargoLockRaw, ...npmPackageRaws] = await Promise.all(
    [
      fs.readFile(TSGO_CRATE.manifestPath, 'utf8'),
      fs.readFile(TSGO_CRATE.lockPath, 'utf8'),
      ...TSGO_NPM_PACKAGES.map(({ manifestPath }) =>
        fs.readFile(manifestPath, 'utf8'),
      ),
    ],
  );

  const platformEntries = await fs.readdir('npm/tsgo', {
    withFileTypes: true,
  });

  return {
    crateManifestRaw,
    crate: parseCargoManifestPackage(crateManifestRaw),
    cargoLockRaw,
    cargoLockVersion: parseCargoLockVersion(cargoLockRaw, TSGO_CRATE.name),
    npmPackages: TSGO_NPM_PACKAGES.map((expected, index) => ({
      expected,
      manifest: JSON.parse(npmPackageRaws[index]),
    })),
    platformDirectories: platformEntries
      .filter((entry) => entry.isDirectory())
      .map((entry) => entry.name),
  };
}

export function validateTsgoReleaseState(state) {
  const errors = [];
  const version = state.crate.version;

  if (state.crate.name !== TSGO_CRATE.name) {
    errors.push(
      `${TSGO_CRATE.manifestPath} has name ${state.crate.name}, expected ${TSGO_CRATE.name}`,
    );
  }
  if (state.cargoLockVersion !== version) {
    errors.push(
      `${TSGO_CRATE.lockPath} has ${TSGO_CRATE.name}@${state.cargoLockVersion}, expected ${version}`,
    );
  }

  const expectedPlatformDirectories = new Set(
    TSGO_PLATFORMS.map(({ key }) => key),
  );
  const actualPlatformDirectories = new Set(state.platformDirectories);
  if (!compareSets(actualPlatformDirectories, expectedPlatformDirectories)) {
    errors.push(
      `npm/tsgo platform directories are ${[...actualPlatformDirectories].sort().join(', ')}, expected ${[...expectedPlatformDirectories].sort().join(', ')}`,
    );
  }

  for (const { expected, manifest } of state.npmPackages) {
    if (manifest.name !== expected.name) {
      errors.push(
        `${expected.manifestPath} has name ${manifest.name}, expected ${expected.name}`,
      );
    }
    if (manifest.version !== version) {
      errors.push(
        `${manifest.name ?? expected.manifestPath} is ${manifest.version}, but ${TSGO_CRATE.name} is ${version}`,
      );
    }
    if (manifest.publishConfig?.access !== 'public') {
      errors.push(`${expected.name} must publish with public access`);
    }
  }

  const wrapper = state.npmPackages[0].manifest;
  const optionalDependencies = wrapper.optionalDependencies ?? {};
  const actualOptionalDependencies = new Set(Object.keys(optionalDependencies));
  const expectedOptionalDependencies = new Set(
    TSGO_PLATFORMS.map(({ name }) => name),
  );
  if (!compareSets(actualOptionalDependencies, expectedOptionalDependencies)) {
    errors.push(
      `${TSGO_WRAPPER.name} optional dependencies do not exactly match its platform packages`,
    );
  }
  for (const name of expectedOptionalDependencies) {
    if (optionalDependencies[name] !== 'workspace:*') {
      errors.push(
        `${TSGO_WRAPPER.name} must depend on ${name} through workspace:*`,
      );
    }
  }
  if (wrapper.bin?.tsgo !== './bin/tsgo.cjs') {
    errors.push(`${TSGO_WRAPPER.name} must expose ./bin/tsgo.cjs as tsgo`);
  }
  if (!sameValues(wrapper.files, ['bin/tsgo.cjs'])) {
    errors.push(`${TSGO_WRAPPER.name} must publish only bin/tsgo.cjs`);
  }

  for (let index = 0; index < TSGO_PLATFORMS.length; index++) {
    const expected = TSGO_PLATFORMS[index];
    const manifest = state.npmPackages[index + 1].manifest;
    if (!sameValues(manifest.os, [expected.os])) {
      errors.push(`${expected.name} must target os=${expected.os}`);
    }
    if (!sameValues(manifest.cpu, [expected.cpu])) {
      errors.push(`${expected.name} must target cpu=${expected.cpu}`);
    }
    if (!sameValues(manifest.files, ['lib'])) {
      errors.push(`${expected.name} must publish only its lib directory`);
    }
    if (
      !sameValues(manifest.publishConfig?.executableFiles, [
        expected.executable,
      ])
    ) {
      errors.push(
        `${expected.name} must mark ${expected.executable} executable`,
      );
    }
  }

  if (errors.length > 0) {
    throw new Error(`Invalid tsgo release state:\n- ${errors.join('\n- ')}`);
  }

  return version;
}

export function replaceCargoManifestVersion(content, newVersion) {
  let inPackage = false;
  let replacements = 0;
  const lines = content.split('\n').map((line) => {
    if (/^\s*\[/.test(line)) {
      inPackage = line.trim() === '[package]';
      return line;
    }
    if (!inPackage || !/^\s*version\s*=/.test(line)) {
      return line;
    }
    replacements++;
    return line.replace(
      /^(\s*version\s*=\s*")[^"]+(".*)$/,
      `$1${newVersion}$2`,
    );
  });

  if (replacements !== 1) {
    throw new Error(
      `expected one [package] version in ${TSGO_CRATE.manifestPath}, found ${replacements}`,
    );
  }
  return lines.join('\n');
}

export function replaceCargoLockVersion(content, newVersion) {
  const sections = content.split('[[package]]');
  let replacements = 0;
  for (let index = 1; index < sections.length; index++) {
    const name = sections[index].match(/^\s*name\s*=\s*"([^"]+)"/m)?.[1];
    if (name !== TSGO_CRATE.name) {
      continue;
    }
    sections[index] = sections[index].replace(
      /^(\s*version\s*=\s*")[^"]+(".*)$/m,
      (_, prefix, suffix) => {
        replacements++;
        return `${prefix}${newVersion}${suffix}`;
      },
    );
  }

  if (replacements !== 1) {
    throw new Error(
      `expected one ${TSGO_CRATE.name} version in ${TSGO_CRATE.lockPath}, found ${replacements}`,
    );
  }
  return sections.join('[[package]]');
}
