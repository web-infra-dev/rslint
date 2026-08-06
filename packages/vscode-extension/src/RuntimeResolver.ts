import fs from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';

import { workspace, type TextDocument, type WorkspaceFolder } from 'vscode';

export interface ResolvedRuntime {
  /** Pool identity: workspace + physical core entry + PnP domain. */
  readonly key: string;
  readonly workspaceFolder: WorkspaceFolder;
  readonly entryPath: string;
  readonly packagePath: string;
  /** Process/config root for this module-resolution domain. */
  readonly workingDirectory: string;
  /** Active Yarn PnP hook, used to keep config discovery in this domain. */
  readonly pnpPath?: string;
  readonly version: string | undefined;
  readonly nodeArgs: readonly string[];
  /** Exact host files that can invalidate this generation. */
  readonly watchPaths: readonly string[];
  readonly source: 'configured' | 'node-modules' | 'pnp';
}

interface PackageManifest {
  name?: unknown;
  version?: unknown;
  exports?: unknown;
}

interface PnpApi {
  resolveRequest(
    request: string,
    issuer: string,
    options?: { considerBuiltins?: boolean },
  ): string | null;
}

interface PnpResolution {
  readonly domainFound: boolean;
  readonly runtime?: ResolvedRuntime;
}

interface PnpDomain {
  readonly pnpPath: string;
  readonly dataPath: string;
  readonly loaderPath: string | undefined;
  readonly generation: string;
  readonly nodeArgs: readonly string[];
  readonly watchPaths: readonly string[];
}

interface CachedPnpApi {
  readonly generation: string;
  readonly api: PnpApi;
}

interface NodeModulesTopology {
  readonly generation: string;
  readonly watchPaths: readonly string[];
}

interface RuntimePayloadTopology {
  readonly generation: string;
  readonly watchPaths: readonly string[];
}

// A generated PnP API retains the whole dependency graph. Keep hot domains
// fast without letting a workspace that churns nested .pnp.cjs files grow the
// extension host forever. Eviction affects performance only; resolution stays
// correct because the API is reconstructed from its on-disk generation.
const MAX_CACHED_PNP_APIS = 16;
const pnpApiCache = new Map<string, CachedPnpApi>();
const NODE_MODULES_TOPOLOGY_FILES = [
  'package-lock.json',
  'npm-shrinkwrap.json',
  'pnpm-lock.yaml',
  'pnpm-workspace.yaml',
  'yarn.lock',
  'bun.lock',
  'bun.lockb',
] as const;

function isPnpApi(value: unknown): value is PnpApi {
  return (
    value !== null &&
    typeof value === 'object' &&
    'resolveRequest' in value &&
    typeof value.resolveRequest === 'function'
  );
}

async function pathExists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch (error) {
    const code = (error as NodeJS.ErrnoException).code;
    if (code === 'ENOENT' || code === 'ENOTDIR') return false;
    throw error;
  }
}

async function canonicalPath(filePath: string): Promise<string> {
  try {
    return await fs.realpath(filePath);
  } catch (error) {
    const code = (error as NodeJS.ErrnoException).code;
    if (code !== 'ENOENT' && code !== 'ENOTDIR') throw error;
    // Yarn's zip paths are visible only after the PnP hook patches fs. Their
    // PnP-domain-qualified logical path is still a stable identity here.
    return path.normalize(filePath);
  }
}

async function readManifest(packagePath: string): Promise<PackageManifest> {
  const json = await fs.readFile(packagePath, 'utf8');
  const value: unknown = JSON.parse(json);
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${packagePath} does not contain a package manifest`);
  }
  return value as PackageManifest;
}

function editorRuntimeTarget(
  manifest: PackageManifest,
  packagePath: string,
): string {
  if (
    manifest.exports === null ||
    typeof manifest.exports !== 'object' ||
    Array.isArray(manifest.exports)
  ) {
    throw new Error(
      `${packagePath} does not export @rslint/core/editor-runtime`,
    );
  }
  const subpath = (manifest.exports as Record<string, unknown>)[
    './editor-runtime'
  ];
  const target =
    typeof subpath === 'string'
      ? subpath
      : subpath !== null &&
          typeof subpath === 'object' &&
          !Array.isArray(subpath) &&
          typeof (subpath as Record<string, unknown>).default === 'string'
        ? (subpath as Record<string, string>).default
        : undefined;
  if (!target || !target.startsWith('./')) {
    throw new Error(
      `${packagePath} must expose ./editor-runtime as a package-relative default target`,
    );
  }
  return target;
}

async function resolvePackageRuntime(
  packagePath: string,
  folder: WorkspaceFolder,
  source: 'configured' | 'node-modules',
): Promise<ResolvedRuntime> {
  const manifest = await readManifest(packagePath);
  if (manifest.name !== '@rslint/core') {
    throw new Error(
      `${source === 'configured' ? 'rslint.runtime.path must point to' : 'resolved package is not'} an @rslint/core package directory; found ${JSON.stringify(manifest.name)}`,
    );
  }
  const target = editorRuntimeTarget(manifest, packagePath);
  const logicalEntryPath = path.resolve(path.dirname(packagePath), target);
  await fs.access(logicalEntryPath);
  const [entryPath, canonicalPackagePath] = await Promise.all([
    fs.realpath(logicalEntryPath),
    fs.realpath(packagePath),
  ]);
  const logicalPackageDirectory = path.dirname(packagePath);
  const canonicalPackageDirectory = path.dirname(canonicalPackagePath);
  if (!isPathWithin(canonicalPackageDirectory, entryPath)) {
    throw new Error(
      `${packagePath} exposes an editor runtime outside its package directory`,
    );
  }
  const [entryGeneration, manifestGeneration, topology, payload] =
    await Promise.all([
      fileGeneration(entryPath),
      fileGeneration(canonicalPackagePath),
      nodeModulesTopology(canonicalPackagePath),
      runtimePayloadTopology(
        logicalPackageDirectory,
        canonicalPackageDirectory,
      ),
    ]);
  const generation = `${String(manifest.version)}:${entryGeneration}:${manifestGeneration}:${topology.generation}:${payload.generation}`;
  return {
    key: runtimeKey(folder, entryPath, generation),
    workspaceFolder: folder,
    entryPath,
    packagePath: canonicalPackagePath,
    workingDirectory: folder.uri.fsPath,
    version:
      typeof manifest.version === 'string' ? manifest.version : undefined,
    nodeArgs: [],
    watchPaths: [
      entryPath,
      canonicalPackagePath,
      ...topology.watchPaths,
      ...payload.watchPaths,
    ],
    source,
  };
}

function runtimeKey(
  folder: WorkspaceFolder,
  entryPath: string,
  generation: string,
  pnpPath?: string,
): string {
  return `${folder.uri.toString()}\0${entryPath}\0${pnpPath ?? ''}\0${generation}`;
}

function isPathWithin(root: string, candidate: string): boolean {
  const relative = path.relative(root, candidate);
  return (
    relative === '' ||
    (relative !== '..' &&
      !relative.startsWith(`..${path.sep}`) &&
      !path.isAbsolute(relative))
  );
}

async function fileGeneration(
  filePath: string,
  allowVirtual = false,
): Promise<string> {
  try {
    const stat = await fs.stat(filePath);
    return `${String(stat.dev)}:${String(stat.ino)}:${stat.size}:${stat.mtimeMs}:${stat.ctimeMs}`;
  } catch (error) {
    const code = (error as NodeJS.ErrnoException).code;
    if (allowVirtual && (code === 'ENOENT' || code === 'ENOTDIR')) {
      return 'virtual';
    }
    throw error;
  }
}

function zipArchivePath(logicalPath: string): string | undefined {
  const parsed = path.parse(logicalPath);
  const relativeParts = logicalPath
    .slice(parsed.root.length)
    .split(path.sep)
    .filter(Boolean);
  const archiveIndex = relativeParts.findIndex((part) =>
    part.toLowerCase().endsWith('.zip'),
  );
  return archiveIndex < 0
    ? undefined
    : path.join(parsed.root, ...relativeParts.slice(0, archiveIndex + 1));
}

async function virtualBackingFiles(
  logicalPaths: readonly string[],
): Promise<readonly string[]> {
  const backingFiles = new Set<string>();
  for (const logicalPath of logicalPaths) {
    if (await pathExists(logicalPath)) {
      backingFiles.add(await fs.realpath(logicalPath));
      continue;
    }
    const archivePath = zipArchivePath(logicalPath);
    if (archivePath && (await pathExists(archivePath))) {
      backingFiles.add(await fs.realpath(archivePath));
    }
  }
  return [...backingFiles];
}

/**
 * Capture the dependency domain Node will walk from the physical core package.
 * Lockfile/package-owner changes can replace transitive runtime dependencies
 * without changing core's own entry bytes; keeping them out of RuntimeKey
 * would incorrectly reuse a sidecar whose module cache still holds the old
 * graph. The nearest owner manifest and nearest lockfile boundary are enough;
 * document-local manifests must not split one shared root installation.
 */
async function nodeModulesTopology(
  canonicalPackagePath: string,
): Promise<NodeModulesTopology> {
  const packageDirectory = path.dirname(canonicalPackagePath);
  let current = path.dirname(packageDirectory);
  let ownerManifestFound = false;
  const parts: string[] = [];
  const watchPaths: string[] = [];

  for (;;) {
    if (!ownerManifestFound) {
      const ownerManifest = path.join(current, 'package.json');
      if (await pathExists(ownerManifest)) {
        const canonicalManifest = await fs.realpath(ownerManifest);
        parts.push(
          `owner:${canonicalManifest}:${await fileGeneration(canonicalManifest)}`,
        );
        watchPaths.push(canonicalManifest);
        ownerManifestFound = true;
      }
    }

    const topologyFiles = (
      await Promise.all(
        NODE_MODULES_TOPOLOGY_FILES.map(async (name) => {
          const candidate = path.join(current, name);
          return (await pathExists(candidate)) ? candidate : undefined;
        }),
      )
    ).filter((candidate): candidate is string => candidate !== undefined);
    if (topologyFiles.length > 0) {
      for (const topologyFile of topologyFiles) {
        const canonicalFile = await fs.realpath(topologyFile);
        parts.push(
          `lock:${canonicalFile}:${await fileGeneration(canonicalFile)}`,
        );
        watchPaths.push(canonicalFile);
      }
      break;
    }

    const parent = path.dirname(current);
    if (parent === current) break;
    current = parent;
  }

  return {
    generation: parts.length > 0 ? parts.join('|') : 'no-owner-topology',
    watchPaths,
  };
}

const ROOT_RUNTIME_PAYLOAD = /\.(?:cjs|js|json|mjs|node|wasm)$/i;

async function collectRuntimePayloadTree(
  directory: string,
  paths: Set<string>,
): Promise<void> {
  paths.add(directory);
  let entries;
  try {
    entries = await fs.readdir(directory, { withFileTypes: true });
  } catch (error) {
    const code = (error as NodeJS.ErrnoException).code;
    if (code === 'ENOENT' || code === 'ENOTDIR') return;
    throw error;
  }
  await Promise.all(
    entries.map(async (entry) => {
      const entryPath = path.join(directory, entry.name);
      paths.add(entryPath);
      if (entry.isDirectory()) {
        await collectRuntimePayloadTree(entryPath, paths);
      } else if (entry.isSymbolicLink()) {
        try {
          paths.add(await fs.realpath(entryPath));
        } catch (error) {
          const code = (error as NodeJS.ErrnoException).code;
          if (code !== 'ENOENT' && code !== 'ENOTDIR') throw error;
        }
      }
    }),
  );
}

/**
 * Fingerprint the complete executable payload shipped by core, not only its
 * public editor entry. Builds split that entry into shared chunks and load the
 * plugin host/worker lazily; a local rebuild can therefore change executable
 * bytes without touching editor-runtime.js or package.json. `bin` is included
 * for the same reason: it is part of the selected core implementation even
 * though the editor currently enters through the explicit runtime subpath.
 */
async function runtimePayloadTopology(
  logicalPackageDirectory: string,
  canonicalPackageDirectory: string,
): Promise<RuntimePayloadTopology> {
  const payloadPaths = new Set<string>([canonicalPackageDirectory]);
  const watchPaths = new Set<string>([
    logicalPackageDirectory,
    canonicalPackageDirectory,
  ]);
  const rootEntries = await fs.readdir(canonicalPackageDirectory, {
    withFileTypes: true,
  });
  for (const entry of rootEntries) {
    if (entry.isFile() && ROOT_RUNTIME_PAYLOAD.test(entry.name)) {
      const entryPath = path.join(canonicalPackageDirectory, entry.name);
      payloadPaths.add(entryPath);
      watchPaths.add(entryPath);
    }
  }
  await Promise.all(
    ['bin', 'dist'].map(async (name) =>
      collectRuntimePayloadTree(
        path.join(canonicalPackageDirectory, name),
        payloadPaths,
      ),
    ),
  );

  const records = await Promise.all(
    [...payloadPaths].sort().map(async (payloadPath) => {
      try {
        const canonical = await fs.realpath(payloadPath);
        watchPaths.add(payloadPath);
        watchPaths.add(canonical);
        return `${payloadPath}=>${canonical}:${await fileGeneration(canonical)}`;
      } catch (error) {
        const code = (error as NodeJS.ErrnoException).code;
        if (code === 'ENOENT' || code === 'ENOTDIR') {
          watchPaths.add(payloadPath);
          return `${payloadPath}:missing`;
        }
        throw error;
      }
    }),
  );
  return {
    generation: records.join('|'),
    watchPaths: [...watchPaths],
  };
}

async function physicalPnpRuntimePayloadTopology(
  packagePath: string,
): Promise<RuntimePayloadTopology> {
  if (!(await pathExists(packagePath))) {
    // ZipFS packages are fingerprinted and watched through their physical
    // archive below. Workspace/unplugged packages remain ordinary files and
    // need the same complete payload fingerprint as node_modules installs.
    return { generation: 'virtual-package-payload', watchPaths: [] };
  }
  const canonicalPackagePath = await fs.realpath(packagePath);
  return runtimePayloadTopology(
    path.dirname(packagePath),
    path.dirname(canonicalPackagePath),
  );
}

async function loadPnpApi(
  pnpPath: string,
  dataPath: string,
): Promise<{ api: PnpApi; generation: string }> {
  const generation = `${await fileGeneration(pnpPath)}:${await fileGeneration(dataPath, true)}`;
  const cached = pnpApiCache.get(pnpPath);
  if (cached?.generation === generation) {
    // Map insertion order gives us a tiny LRU without another index.
    pnpApiCache.delete(pnpPath);
    pnpApiCache.set(pnpPath, cached);
    return cached;
  }

  const loadPnp = createRequire(pnpPath);
  const pnpModulePath = loadPnp.resolve(pnpPath);
  // Do not leave project code in the shared extension-host require cache. Also
  // preserve a pre-existing entry in case another extension loaded this API.
  const previousModule = loadPnp.cache[pnpModulePath];
  delete loadPnp.cache[pnpModulePath];
  let value: unknown;
  try {
    value = loadPnp(pnpModulePath);
  } finally {
    delete loadPnp.cache[pnpModulePath];
    if (previousModule) loadPnp.cache[pnpModulePath] = previousModule;
  }
  if (!isPnpApi(value)) {
    throw new Error(`${pnpPath} does not expose a Yarn PnP API`);
  }

  const loaded: CachedPnpApi = { api: value, generation };
  pnpApiCache.delete(pnpPath);
  pnpApiCache.set(pnpPath, loaded);
  while (pnpApiCache.size > MAX_CACHED_PNP_APIS) {
    const oldest = pnpApiCache.keys().next().value as string | undefined;
    if (oldest === undefined) break;
    pnpApiCache.delete(oldest);
  }
  return loaded;
}

async function findPnpFile(startPath: string): Promise<string | undefined> {
  let current = path.dirname(startPath);
  for (;;) {
    for (const name of ['.pnp.cjs', '.pnp.js']) {
      const candidate = path.join(current, name);
      if (await pathExists(candidate)) return fs.realpath(candidate);
    }
    const parent = path.dirname(current);
    if (parent === current) return undefined;
    current = parent;
  }
}

async function describePnpDomain(
  documentPath: string,
): Promise<PnpDomain | undefined> {
  const pnpPath = await findPnpFile(documentPath);
  if (!pnpPath) return undefined;
  const pnpDirectory = path.dirname(pnpPath);
  const dataPath = path.join(pnpDirectory, '.pnp.data.json');
  const loaderCandidate = path.join(pnpDirectory, '.pnp.loader.mjs');
  const loaderPath = (await pathExists(loaderCandidate))
    ? await fs.realpath(loaderCandidate)
    : undefined;
  const generation = `${await fileGeneration(pnpPath)}:${await fileGeneration(dataPath, true)}:${loaderPath ? await fileGeneration(loaderPath) : 'no-loader'}`;
  return {
    pnpPath,
    dataPath,
    loaderPath,
    generation,
    nodeArgs: [
      '--require',
      pnpPath,
      ...(loaderPath ? ['--experimental-loader', loaderPath] : []),
    ],
    watchPaths: [pnpPath, dataPath, loaderCandidate],
  };
}

async function resolveConfiguredRuntime(
  configuredPath: string,
  folder: WorkspaceFolder,
  documentPath: string,
): Promise<ResolvedRuntime> {
  const packageDirectory = path.isAbsolute(configuredPath)
    ? configuredPath
    : path.resolve(folder.uri.fsPath, configuredPath);
  const packagePath = path.join(packageDirectory, 'package.json');
  const [runtime, pnpDomain] = await Promise.all([
    resolvePackageRuntime(packagePath, folder, 'configured'),
    describePnpDomain(documentPath),
  ]);
  if (!pnpDomain) return runtime;
  return {
    ...runtime,
    // A configured path overrides only core selection. Project config and
    // plugin imports still belong to the document issuer's PnP graph, so two
    // nested PnP domains must not share one Node process even with one core.
    key: `${runtime.key}\0configured-pnp:${pnpDomain.pnpPath}:${pnpDomain.generation}`,
    workingDirectory: path.dirname(pnpDomain.pnpPath),
    pnpPath: pnpDomain.pnpPath,
    nodeArgs: pnpDomain.nodeArgs,
    watchPaths: [...runtime.watchPaths, ...pnpDomain.watchPaths],
  };
}

async function resolveFromPnp(
  document: TextDocument,
  folder: WorkspaceFolder,
): Promise<PnpResolution> {
  const domain = await describePnpDomain(document.uri.fsPath);
  if (!domain) return { domainFound: false };
  const pnpPath = domain.pnpPath;
  const { api, generation: pnpGeneration } = await loadPnpApi(
    pnpPath,
    domain.dataPath,
  );
  const entry = api.resolveRequest(
    '@rslint/core/editor-runtime',
    document.uri.fsPath,
    { considerBuiltins: false },
  );
  const packagePath = api.resolveRequest(
    '@rslint/core/package.json',
    document.uri.fsPath,
    { considerBuiltins: false },
  );
  if (!entry && !packagePath) {
    // A PnP boundary owns dependency resolution for every issuer below it.
    // Falling through to an unrelated ancestor node_modules tree would violate
    // that graph and could silently run a core the project did not declare.
    return { domainFound: true };
  }
  if (!entry || !packagePath) {
    throw new Error(
      `${pnpPath} resolved an incomplete @rslint/core package (editor runtime: ${String(entry)}, manifest: ${String(packagePath)})`,
    );
  }
  const entryPath = await canonicalPath(entry);
  const [backingFiles, payload] = await Promise.all([
    virtualBackingFiles([entry, packagePath]),
    physicalPnpRuntimePayloadTopology(packagePath),
  ]);
  const backingGeneration = (
    await Promise.all(
      backingFiles.map(
        async (filePath) => `${filePath}:${await fileGeneration(filePath)}`,
      ),
    )
  ).join('|');
  const generation = `${pnpGeneration}:${domain.generation}:${entryPath}:${await fileGeneration(entry, true)}:${await fileGeneration(packagePath, true)}:${backingGeneration || 'no-native-backing'}:${payload.generation}`;
  return {
    domainFound: true,
    runtime: {
      key: runtimeKey(folder, entryPath, generation, pnpPath),
      workspaceFolder: folder,
      entryPath,
      packagePath,
      workingDirectory: path.dirname(pnpPath),
      pnpPath,
      version: undefined,
      nodeArgs: domain.nodeArgs,
      watchPaths: [
        ...domain.watchPaths,
        ...backingFiles,
        ...payload.watchPaths,
      ],
      source: 'pnp',
    },
  };
}

async function findNodeModulesPackage(
  startPath: string,
): Promise<string | undefined> {
  let current = path.dirname(startPath);
  for (;;) {
    const candidate = path.join(
      current,
      'node_modules/@rslint/core/package.json',
    );
    if (await pathExists(candidate)) return candidate;
    const parent = path.dirname(current);
    if (parent === current) return undefined;
    current = parent;
  }
}

function resolveFromNodeModules(
  document: TextDocument,
  folder: WorkspaceFolder,
): Promise<ResolvedRuntime | undefined> {
  return (async () => {
    const packagePath = await findNodeModulesPackage(document.uri.fsPath);
    return packagePath
      ? resolvePackageRuntime(packagePath, folder, 'node-modules')
      : undefined;
  })();
}

/** Resolve the core installation using the same issuer as the document. */
export async function resolveRuntimeForDocument(
  document: TextDocument,
): Promise<ResolvedRuntime | undefined> {
  const folder = workspace.getWorkspaceFolder(document.uri);
  if (!folder || document.uri.scheme !== 'file') return undefined;

  const configuredPath = workspace
    .getConfiguration('rslint', folder.uri)
    .get<string>('runtime.path', '')
    .trim();
  if (configuredPath) {
    return resolveConfiguredRuntime(
      configuredPath,
      folder,
      document.uri.fsPath,
    );
  }

  // A discovered PnP domain is authoritative for that issuer. PnP resolution
  // errors and a missing declaration both surface as that domain's result;
  // only issuers outside every PnP domain may walk a node_modules ancestry.
  const pnp = await resolveFromPnp(document, folder);
  return pnp.domainFound
    ? pnp.runtime
    : resolveFromNodeModules(document, folder);
}
