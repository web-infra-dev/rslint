import fs from 'node:fs/promises';
import path from 'node:path';
import { createRequire } from 'node:module';
import type { TextDocument, WorkspaceFolder } from 'vscode';
import type { ConfigModuleHost } from '@rslint/core/config-loader';
import type { createPluginLintHost } from '@rslint/core/eslint-plugin';

const CORE_PACKAGE_NAME = '@rslint/core';
const RESOLUTION_ANCHOR = '__rslint_vscode_resolver__.cjs';

type ConfigModuleHostConstructor = new () => ConfigModuleHost;
type PluginLintHostFactory = typeof createPluginLintHost;

interface ConfigLoaderModule {
  readonly ConfigModuleHost: ConfigModuleHostConstructor;
  readonly CONFIG_DISCOVERY_PROTOCOL_VERSION: number;
  readonly resolveRslintBinary: () => unknown;
}

interface PluginHostModule {
  readonly createPluginLintHost: PluginLintHostFactory;
}

interface CorePackageJson {
  readonly name: string;
  readonly version: string;
}

export interface CoreInstallation {
  /** Stable physical identity. Never merge packages by version text alone. */
  readonly identity: string;
  readonly packageDirectory: string;
  readonly version: string;
  readonly binaryPath: string;
  readonly protocolVersion: number;
  createConfigModuleHost(): ConfigModuleHost;
  createPluginLintHost: PluginLintHostFactory;
}

export interface ResolvedCoreRuntime {
  readonly key: string;
  readonly workspaceFolder: WorkspaceFolder;
  readonly installation: CoreInstallation;
}

export class CoreNotFoundError extends Error {
  public constructor(searchDirectory: string, cause?: unknown) {
    super(
      `Could not resolve ${CORE_PACKAGE_NAME} from ${searchDirectory}. Install it in the project or configure rslint.corePath.`,
      { cause },
    );
    this.name = 'CoreNotFoundError';
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isConfigLoaderModule(value: unknown): value is ConfigLoaderModule {
  return (
    isRecord(value) &&
    typeof value.ConfigModuleHost === 'function' &&
    Number.isInteger(value.CONFIG_DISCOVERY_PROTOCOL_VERSION) &&
    typeof value.resolveRslintBinary === 'function'
  );
}

function isPluginHostModule(value: unknown): value is PluginHostModule {
  return isRecord(value) && typeof value.createPluginLintHost === 'function';
}

function normalizeIdentity(filePath: string): string {
  const normalized = path.normalize(filePath);
  return process.platform === 'win32' ? normalized.toLowerCase() : normalized;
}

function runtimeKey(
  workspaceFolder: WorkspaceFolder,
  installation: CoreInstallation,
): string {
  return `${workspaceFolder.uri.toString()}\0${installation.identity}`;
}

function resolverFrom(directory: string): NodeJS.Require {
  return createRequire(path.join(directory, RESOLUTION_ANCHOR));
}

export function resolveCorePackageDirectory(
  searchDirectory: string,
  configuredDirectory?: string,
): string {
  if (configuredDirectory) return path.resolve(configuredDirectory);
  try {
    return path.dirname(
      resolverFrom(searchDirectory).resolve(
        `${CORE_PACKAGE_NAME}/package.json`,
      ),
    );
  } catch (error) {
    throw new CoreNotFoundError(searchDirectory, error);
  }
}

async function readPackageJson(
  packageDirectory: string,
): Promise<CorePackageJson> {
  const packageJsonPath = path.join(packageDirectory, 'package.json');
  let parsed: unknown;
  try {
    parsed = JSON.parse(await fs.readFile(packageJsonPath, 'utf8'));
  } catch (error) {
    throw new Error(`Could not read ${packageJsonPath}`, { cause: error });
  }
  if (
    !isRecord(parsed) ||
    parsed.name !== CORE_PACKAGE_NAME ||
    typeof parsed.version !== 'string' ||
    parsed.version.length === 0
  ) {
    throw new Error(
      `${packageJsonPath} is not a valid ${CORE_PACKAGE_NAME} package`,
    );
  }
  return { name: parsed.name, version: parsed.version };
}

function resolveExport(packageDirectory: string, subpath: string): string {
  try {
    return resolverFrom(packageDirectory).resolve(
      `${CORE_PACKAGE_NAME}/${subpath}`,
    );
  } catch (error) {
    throw new Error(
      `${CORE_PACKAGE_NAME} does not provide the required ./${subpath} export`,
      { cause: error },
    );
  }
}

function loadModule(modulePath: string): unknown {
  // VS Code 1.131 embeds Node 22, whose require(esm) support lets the CommonJS
  // extension bundle and its CommonJS test build use one exact loading path.
  // The selected core entries contain no top-level await.
  return resolverFrom(path.dirname(modulePath))(modulePath) as unknown;
}

async function loadInstallation(
  packageDirectory: string,
  physicalDirectory: string,
): Promise<CoreInstallation> {
  const packageJson = await readPackageJson(packageDirectory);
  const configLoaderPath = resolveExport(packageDirectory, 'config-loader');
  const pluginHostPath = resolveExport(packageDirectory, 'eslint-plugin');
  const configLoaderModule = loadModule(configLoaderPath);
  if (!isConfigLoaderModule(configLoaderModule)) {
    throw new Error(
      `${CORE_PACKAGE_NAME}/config-loader has an incompatible module shape`,
    );
  }

  // Runtime modules cross a dynamic package boundary, so validate their exact
  // callable shape before retaining anything from the selected installation.
  const binaryPath = configLoaderModule.resolveRslintBinary();
  if (typeof binaryPath !== 'string' || binaryPath.length === 0) {
    throw new Error(
      `${CORE_PACKAGE_NAME}/config-loader returned an invalid binary path`,
    );
  }
  const binaryStat = await fs.stat(binaryPath).catch((error: unknown) => {
    throw new Error(`Rslint binary does not exist at ${binaryPath}`, {
      cause: error,
    });
  });
  if (!binaryStat.isFile()) {
    throw new Error(`Rslint binary is not a file: ${binaryPath}`);
  }

  let pluginFactoryPromise: Promise<PluginLintHostFactory> | undefined;
  const getPluginFactory = async (): Promise<PluginLintHostFactory> => {
    pluginFactoryPromise ??= Promise.resolve().then(() => {
      const module = loadModule(pluginHostPath);
      if (!isPluginHostModule(module)) {
        throw new Error(
          `${CORE_PACKAGE_NAME}/eslint-plugin has an incompatible module shape`,
        );
      }
      return module.createPluginLintHost;
    });
    try {
      return await pluginFactoryPromise;
    } catch (error) {
      pluginFactoryPromise = undefined;
      throw error;
    }
  };

  return {
    identity: normalizeIdentity(physicalDirectory),
    packageDirectory,
    version: packageJson.version,
    binaryPath,
    protocolVersion: configLoaderModule.CONFIG_DISCOVERY_PROTOCOL_VERSION,
    createConfigModuleHost: () => new configLoaderModule.ConfigModuleHost(),
    createPluginLintHost: async (...args) => {
      const factory = await getPluginFactory();
      return factory(...args);
    },
  };
}

/** Resolves and loads only project-local core installations; no PnP or fallback. */
export class CoreResolver {
  private readonly physicalDirectories = new Map<string, Promise<string>>();
  private readonly installations = new Map<string, Promise<CoreInstallation>>();

  public clear(): void {
    this.physicalDirectories.clear();
    this.installations.clear();
  }

  public async resolve(
    document: TextDocument,
    workspaceFolder: WorkspaceFolder,
    configuredPath?: string,
  ): Promise<ResolvedCoreRuntime> {
    const configuredDirectory = configuredPath?.trim()
      ? path.resolve(workspaceFolder.uri.fsPath, configuredPath.trim())
      : undefined;
    const packageDirectory = resolveCorePackageDirectory(
      path.dirname(document.uri.fsPath),
      configuredDirectory,
    );
    const directoryKey = normalizeIdentity(packageDirectory);
    let physicalDirectoryPromise = this.physicalDirectories.get(directoryKey);
    if (!physicalDirectoryPromise) {
      physicalDirectoryPromise = fs
        .realpath(packageDirectory)
        .then(normalizeIdentity, (error: unknown) => {
          throw new Error(
            `Could not access ${CORE_PACKAGE_NAME} at ${packageDirectory}`,
            { cause: error },
          );
        });
      this.physicalDirectories.set(directoryKey, physicalDirectoryPromise);
      void physicalDirectoryPromise.catch(() => {
        if (
          this.physicalDirectories.get(directoryKey) ===
          physicalDirectoryPromise
        ) {
          this.physicalDirectories.delete(directoryKey);
        }
      });
    }
    const physicalDirectory = await physicalDirectoryPromise;
    let installationPromise = this.installations.get(physicalDirectory);
    if (!installationPromise) {
      installationPromise = loadInstallation(
        packageDirectory,
        physicalDirectory,
      );
      this.installations.set(physicalDirectory, installationPromise);
      void installationPromise.catch(() => {
        if (this.installations.get(physicalDirectory) === installationPromise) {
          this.installations.delete(physicalDirectory);
        }
      });
    }
    const installation = await installationPromise;
    return {
      key: runtimeKey(workspaceFolder, installation),
      workspaceFolder,
      installation,
    };
  }
}
