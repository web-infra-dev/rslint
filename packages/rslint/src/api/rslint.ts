/// <reference lib="esnext.disposable" preserve="true" />

/**
 * The `Rslint` class — the ESLint-style programmatic API (issue #1106).
 *
 * It is a thin JS facade over the low-level `lint()` IPC: it owns config
 * resolution (overrideConfig / overrideConfigFile / auto-discovery →
 * normalizeConfig — all in JS, Go `--api` never reads config from disk) and
 * reshapes the wire-level LintResponse into ESLint v10's `LintResult[]`
 * (per-file results, numeric severity, absolute paths, per-result output).
 *
 * The Go side is the single long-lived `--api` process held by the underlying
 * service; `close()` tears it down (mirrors RSLintService.close()).
 */
import path from 'node:path';
import { lstat, readFile, stat, writeFile } from 'node:fs/promises';
import { convertPathToPattern, glob, isDynamicPattern } from 'tinyglobby';
import picomatch from 'picomatch';

import { RSLintService } from '../service/service.js';
import { NodeRslintService } from '../internal/node.js';
import {
  collectPluginMeta,
  loadConfigFile,
  normalizeConfig,
} from '../config/config-loader.js';
import {
  coalesceCaseAliasedConfigs,
  discoverConfigs,
  filterConfigsByParentIgnores,
  findJSConfig,
  findNativeCaseAliasConfigPath,
  type ConfigEntry,
} from '../utils/config-discovery.js';
import {
  AncestorPathIndex,
  createCachedAncestorFinder,
  nativePathIdentity,
  RunPathResolver,
  type ResolvedFilesystemPath,
} from './path-identity.js';
import type { RslintConfigEntry } from '../config/define-config.js';
import type { Diagnostic, Fix, LintResponse } from '../types.js';

export interface RslintOptions {
  /** Base directory for config discovery and relative path resolution. */
  cwd?: string;
  /** Extra config appended after the resolved/discovered config (ESLint's overrideConfig). */
  overrideConfig?: RslintConfigEntry | RslintConfigEntry[] | null;
  /**
   * `string` — use this config file (no discovery).
   * `true`   — use only `overrideConfig` (no file, no discovery).
   * `null`/absent — auto-discover the nearest config (ESLint v10 semantics; no `false`).
   */
  overrideConfigFile?: string | true | null;
  /** Apply rule auto-fixes; results carry `output` (the JS side persists via outputFixes). */
  fix?: boolean;
  /**
   * In-memory file overlay (path → content) for fully in-memory linting (issue
   * #1106): put the `tsconfig.json` that `parserOptions.project` names plus any
   * dependency files here, then lint a buffer with `lintText`. Keys resolve
   * against `cwd` like a linted path (relative or absolute both work); a
   * same-path `lintText` code entry wins. Inside the tsconfig (`files`) and
   * `parserOptions.project`, use relative paths — the TS compiler resolves
   * those, and a bare POSIX-absolute path there has no drive letter on Windows,
   * so it won't match the overlay. rslint-only — ESLint has no in-memory file
   * map.
   *
   * Give the tsconfig explicit `files`, not a broad `include` glob: a glob is
   * expanded against the real filesystem and would scan `cwd` on disk.
   */
  virtualFiles?: Record<string, string>;
}

/** A single fix edit as a flat UTF-16 offset range + replacement text (ESLint shape). */
export interface LintMessageFix {
  range: [number, number];
  text: string;
}

export interface LintSuggestion {
  messageId?: string;
  data?: Record<string, string>;
  desc: string;
  fix: LintMessageFix;
}

export interface LintMessage {
  ruleId: string | null;
  severity: 1 | 2; // ESLint: 1 = warning, 2 = error
  message: string;
  messageId?: string;
  line: number; // 1-based
  column: number; // 1-based, UTF-16
  endLine?: number;
  endColumn?: number;
  fix?: LintMessageFix;
  suggestions?: LintSuggestion[];
}

export interface LintResult {
  filePath: string; // absolute, or the "<text>" sentinel for lintText with no filePath
  messages: LintMessage[];
  errorCount: number;
  warningCount: number;
  fixableErrorCount: number;
  fixableWarningCount: number;
  output?: string; // present only when fix:true changed the file
}

interface ResolvedConfig {
  config: Record<string, unknown>[];
  configPath?: string;
  configDirectory: string;
  pluginConfigDirectory: string;
  routingKey: string;
}

interface PluginLintHost {
  lint(request: unknown): Promise<unknown>;
  shutdown(): Promise<void>;
}

interface PluginLintSession {
  host: PluginLintHost;
  eslintPlugins: Array<{ prefix: string; ruleNames: string[] }>;
}

type CreatePluginLintHost = (
  configs: Array<{ configPath: string; configDirectory: string }>,
  onLog?: (record: { level: string; source: string; text: string }) => void,
) => Promise<PluginLintHost>;

let pluginHostFactoryPromise: Promise<CreatePluginLintHost> | undefined;

async function loadPluginHostFactory(): Promise<CreatePluginLintHost> {
  pluginHostFactoryPromise ??= (async () => {
    // A package self-reference resolves to src under the test condition and to
    // dist/eslint-plugin in published builds. Keep it runtime-only: the library
    // declaration build deliberately excludes the worker implementation.
    const pluginEntry: string = '@rslint/core/eslint-plugin';
    const module: { createPluginLintHost: CreatePluginLintHost } = await import(
      /* webpackIgnore: true */ pluginEntry
    );
    return module.createPluginLintHost;
  })();
  const factory = await pluginHostFactoryPromise;
  return factory;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

interface LoadedConfigCandidate {
  status: 'loaded';
  configPath: string;
  configDirectory: string;
  baseConfig: Record<string, unknown>[];
  resolved: ResolvedConfig;
}

interface FailedConfigCandidate {
  status: 'failed';
  configPath: string;
  configDirectory: string;
  error: Error;
}

type ConfigCandidate = LoadedConfigCandidate | FailedConfigCandidate;

interface ClassifiedLintPatterns {
  literalFilePatterns: string[];
  literalDirectorySymlinks: string[];
  scanDirectories: string[];
}

type GlobPatternRole = 'match' | 'ignore' | 'skip';

// Keep this classification aligned with tinyglobby's processPatterns(): a
// leading `!(` starts a positive extglob, while `!pattern` is an exclusion.
function globPatternRole(pattern: string): GlobPatternRole {
  if (pattern[0] !== '!' || pattern[1] === '(') return 'match';
  if (pattern[1] !== '!' || pattern[2] === '(') return 'ignore';
  return 'skip';
}

interface PlannedLintFile extends ResolvedFilesystemPath {
  explicit: boolean;
}

const DEFAULT_LINT_GLOB_IGNORES = ['**/node_modules/**', '**/.git/**'];

function staticGlobRoot(pattern: string, cwd: string): string {
  const absolutePattern = path.resolve(cwd, pattern);
  const root = path.parse(absolutePattern).root;
  const segments = absolutePattern.slice(root.length).split(path.sep);
  let current = root;
  for (const segment of segments) {
    if (!segment || isDynamicPattern(segment)) break;
    current = path.join(current, segment);
  }
  return current || cwd;
}

function compactScanDirectories(directories: Iterable<string>): string[] {
  const byIdentity = new Map<string, string>();
  for (const directory of directories) {
    const normalized = path.normalize(directory);
    const key = nativePathIdentity.key(normalized);
    if (!byIdentity.has(key)) byIdentity.set(key, normalized);
  }
  const sorted = [...byIdentity.values()].sort(
    (a, b) => a.length - b.length || nativePathIdentity.compare(a, b),
  );
  const compact: string[] = [];
  for (const directory of sorted) {
    if (
      !compact.some((parent) =>
        nativePathIdentity.isSameOrChild(parent, directory),
      )
    ) {
      compact.push(directory);
    }
  }
  return compact;
}

function normalizeGlobPatternForCwd(pattern: string, cwd: string): string {
  let normalized = pattern;
  if (normalized.endsWith('/') || normalized.endsWith('\\')) {
    normalized = normalized.slice(0, -1);
  }
  if (path.isAbsolute(normalized)) {
    normalized = path.relative(cwd, normalized);
  }
  normalized = normalized.split(path.sep).join('/');
  return path.posix.normalize(normalized);
}

function matchesExcludedLiteral(
  filePath: string,
  patterns: readonly string[],
  cwd: string,
): boolean {
  const candidate = path.relative(cwd, filePath).split(path.sep).join('/');
  for (const rawPattern of patterns) {
    let pattern = rawPattern;
    if (pattern.startsWith('!')) {
      if (pattern[1] === '(') continue;
      if (pattern[1] === '!' && pattern[2] !== '(') continue;
      pattern = pattern.slice(1);
    }
    pattern = normalizeGlobPatternForCwd(pattern, cwd);
    const expandedPatterns = pattern.endsWith('*')
      ? [pattern]
      : [pattern, `${pattern}/**`];
    if (picomatch(expandedPatterns, { dot: true, nocase: false })(candidate)) {
      return true;
    }
  }
  return false;
}

async function classifyLintPatterns(
  patterns: string[],
  cwd: string,
): Promise<ClassifiedLintPatterns> {
  const literalFilePatterns: string[] = [];
  const literalDirectorySymlinks: string[] = [];
  const scanDirectories = new Set<string>();
  const excludedLiteralPatterns = [
    ...DEFAULT_LINT_GLOB_IGNORES,
    ...patterns.filter((pattern) => globPatternRole(pattern) === 'ignore'),
  ];

  for (const pattern of patterns) {
    if (globPatternRole(pattern) !== 'match') continue;
    if (isDynamicPattern(pattern)) {
      const scanRoot = staticGlobRoot(pattern, cwd);
      try {
        if ((await stat(scanRoot)).isDirectory()) {
          scanDirectories.add(scanRoot);
        }
      } catch {
        // A missing static root cannot contribute a matched target.
      }
      continue;
    }

    const absolute = path.resolve(cwd, pattern);
    try {
      const info = await stat(absolute);
      if (info.isDirectory()) {
        scanDirectories.add(absolute);
        if ((await lstat(absolute)).isSymbolicLink()) {
          literalDirectorySymlinks.push(absolute);
        }
      } else if (info.isFile()) {
        const absolutePattern = path.resolve(cwd, pattern);
        if (
          !matchesExcludedLiteral(absolutePattern, excludedLiteralPatterns, cwd)
        ) {
          literalFilePatterns.push(absolutePattern);
        }
      }
    } catch {
      // tinyglobby will omit a missing literal from the match set.
    }
  }

  return {
    literalFilePatterns,
    literalDirectorySymlinks,
    scanDirectories: compactScanDirectories(scanDirectories),
  };
}

function createCachedConfigFinder(): (startDirectory: string) => string | null {
  const find = createCachedAncestorFinder((directory) => {
    const configPath = findJSConfig(directory);
    return configPath ? path.normalize(configPath) : undefined;
  });
  return (startDirectory) => find(path.resolve(startDirectory)) ?? null;
}

export class Rslint {
  readonly #service: RSLintService;
  readonly #cwd: string;
  readonly #overrideConfig?: RslintConfigEntry | RslintConfigEntry[] | null;
  readonly #overrideConfigFile?: string | true | null;
  readonly #fix: boolean;
  readonly #virtualFiles?: Record<string, string>;
  readonly #activePluginHosts = new Set<PluginLintHost>();
  #normalizedOverrideConfig?: Record<string, unknown>[];
  #closeRequested = false;
  #closePromise?: Promise<void>;

  constructor(options: RslintOptions = {}) {
    this.#cwd = options.cwd ? path.resolve(options.cwd) : process.cwd();
    this.#overrideConfig = options.overrideConfig;
    this.#overrideConfigFile = options.overrideConfigFile;
    this.#fix = options.fix ?? false;
    this.#virtualFiles = options.virtualFiles;
    this.#service = new RSLintService(new NodeRslintService());
  }

  /**
   * Lint a string of code as if it lived at `filePath` (default a synthetic
   * `.ts` path).
   *
   * ESLint-alignment note: if `code` begins with a UTF-8 BOM, the reported
   * offsets (`fix.range`, `column`) are relative to the BOM-included input you
   * passed — self-consistent within that input (result `output`, line/column,
   * and re-applying `fix` all line up), but one unit ahead of ESLint v10, which
   * strips a leading BOM from its in-memory source. (The overlay keeps the BOM
   * and Go's offsets include it; lintFiles is unaffected because Go reads disk
   * files BOM-stripped.) Strip a leading `U+FEFF` first for ESLint-identical
   * offsets.
   */
  async lintText(
    code: string,
    options: { filePath?: string } = {},
  ): Promise<LintResult[]> {
    const filePath = path.resolve(this.#cwd, options.filePath ?? '__text__.ts');
    const lexicalDirectory = path.dirname(filePath);
    const pathResolver = new RunPathResolver();
    const resolvedFile =
      await pathResolver.resolveWithAncestorFallback(filePath);
    const canonicalDirectory = path.dirname(resolvedFile.canonicalPath);
    const resolved = await this.#resolveConfig(
      lexicalDirectory,
      canonicalDirectory,
    );
    const { config, configDirectory, pluginConfigDirectory } = resolved;
    const pluginSession = await this.#createPluginLintSession([resolved]);
    try {
      const response = await this.#service.lint(
        {
          config,
          eslintPlugins: pluginSession?.eslintPlugins,
          configDirectory,
          pluginConfigDirectory,
          workingDirectory: this.#cwd,
          files: [filePath],
          // Overlay (in-memory tsconfig + deps) underlays the code buffer; a
          // same-path code entry wins so `lintText` always lints `code`.
          fileContents: { ...this.#resolveOverlay(), [filePath]: code },
          fix: this.#fix,
        },
        pluginSession
          ? {
              pluginLint: async (request) => {
                const result = await pluginSession.host.lint(request);
                return result;
              },
            }
          : {},
      );
      const results = this.#toLintResults(
        response,
        configDirectory,
        [filePath],
        { [filePath]: code },
      );
      // ESLint's lintText returns exactly one result — for the linted buffer. An
      // in-memory overlay dependency file (pulled into the program and matching
      // the config) can carry its own diagnostics; keep only the linted file so
      // they neither leak a second result nor get written back by outputFixes.
      const primary = results.filter((r) =>
        nativePathIdentity.equals(r.filePath, filePath),
      );
      // ESLint: with no filePath, the result's path is the non-absolute "<text>"
      // sentinel, so outputFixes skips it. A user-supplied filePath stays
      // absolute, making outputFixes writing it back the caller's intent.
      if (options.filePath == null) {
        for (const r of primary) {
          if (nativePathIdentity.equals(r.filePath, filePath)) {
            r.filePath = '<text>';
          }
        }
      }
      return primary;
    } finally {
      await this.#shutdownPluginSession(pluginSession);
    }
  }

  /**
   * Lint files matched by glob pattern(s), resolved against `cwd`. Results are
   * ordered by the linted file's path (deterministic), not by glob-walk order.
   */
  async lintFiles(patterns: string | string[]): Promise<LintResult[]> {
    const globs = Array.isArray(patterns) ? patterns : [patterns];
    const { literalFilePatterns, literalDirectorySymlinks, scanDirectories } =
      await classifyLintPatterns(globs, this.#cwd);
    const globOptions = {
      cwd: this.#cwd,
      absolute: true,
      onlyFiles: true,
      dot: true,
      ignore: DEFAULT_LINT_GLOB_IGNORES,
    } as const;
    const matched = await glob(globs, {
      ...globOptions,
      followSymbolicLinks: false,
    });
    const directorySymlinkMatches =
      literalDirectorySymlinks.length === 0
        ? []
        : await glob(
            [
              ...literalDirectorySymlinks.map(
                (directory) => `${convertPathToPattern(directory)}/**/*`,
              ),
              ...globs.filter((pattern) => pattern.startsWith('!')),
            ],
            {
              ...globOptions,
              // The explicitly named symlink is the scan root; nested directory
              // symlinks remain excluded.
              followSymbolicLinks: false,
            },
          );
    // stat() already established each literal as a file. Keep its caller
    // spelling directly, including file symlinks, without a second glob crawl.
    const literalMatches = literalFilePatterns;

    const filesByIdentity = new Map<
      string,
      { filePath: string; explicit: boolean }
    >();
    for (const file of [...matched, ...directorySymlinkMatches]) {
      const normalized = path.normalize(file);
      const key = nativePathIdentity.key(normalized);
      if (!filesByIdentity.has(key)) {
        filesByIdentity.set(key, { filePath: normalized, explicit: false });
      }
    }
    for (const file of literalMatches) {
      const normalized = path.normalize(file);
      const key = nativePathIdentity.key(normalized);
      // Preserve the spelling of a literal target over an overlapping glob.
      filesByIdentity.set(key, { filePath: normalized, explicit: true });
    }

    const lexicalFiles = [...filesByIdentity.values()];
    if (lexicalFiles.length === 0) return [];

    const pathResolver = new RunPathResolver();
    const resolvedPaths = await pathResolver.resolveAll(
      lexicalFiles.map(({ filePath }) => filePath),
    );
    const plannedFiles: PlannedLintFile[] = resolvedPaths.map(
      (resolved, index) => ({
        ...resolved,
        explicit: lexicalFiles[index].explicit,
      }),
    );

    const configByFile = await this.#resolveLintFileConfigs(
      plannedFiles,
      scanDirectories,
      pathResolver,
    );
    const groups = new Map<
      string,
      {
        config: Record<string, unknown>[];
        configPath?: string;
        configDirectory: string;
        pluginConfigDirectory: string;
        routingKey: string;
        files: string[];
        canonicalFiles: string[];
      }
    >();

    const selectedByCanonicalPath = new Map<
      string,
      {
        file: PlannedLintFile;
        resolved: ResolvedConfig;
        ownerCanonicalKey: string;
      }
    >();
    for (const file of [...plannedFiles].sort((left, right) =>
      nativePathIdentity.compare(left.lexicalPath, right.lexicalPath),
    )) {
      const resolved = configByFile.get(file.lexicalKey);
      if (!resolved) continue;
      const owner = await pathResolver.resolve(resolved.pluginConfigDirectory);
      const existing = selectedByCanonicalPath.get(file.canonicalKey);
      if (existing) {
        if (existing.ownerCanonicalKey !== owner.canonicalKey) {
          throw new Error(
            `Lint target aliases ${JSON.stringify(existing.file.lexicalPath)} and ${JSON.stringify(file.lexicalPath)} ` +
              `resolve to the same file but are governed by different config directories ` +
              `${JSON.stringify(existing.resolved.pluginConfigDirectory)} and ${JSON.stringify(resolved.pluginConfigDirectory)}`,
          );
        }
        const preferFile =
          (file.explicit && !existing.file.explicit) ||
          (file.explicit === existing.file.explicit &&
            nativePathIdentity.compare(
              file.lexicalPath,
              existing.file.lexicalPath,
            ) < 0);
        if (preferFile) {
          selectedByCanonicalPath.set(file.canonicalKey, {
            file,
            resolved,
            ownerCanonicalKey: owner.canonicalKey,
          });
        }
        continue;
      }
      selectedByCanonicalPath.set(file.canonicalKey, {
        file,
        resolved,
        ownerCanonicalKey: owner.canonicalKey,
      });
    }

    for (const { file, resolved } of selectedByCanonicalPath.values()) {
      const {
        config,
        configPath,
        configDirectory,
        pluginConfigDirectory,
        routingKey,
      } = resolved;
      let group = groups.get(routingKey);
      if (!group) {
        group = {
          config,
          configPath,
          configDirectory,
          pluginConfigDirectory,
          routingKey,
          files: [],
          canonicalFiles: [],
        };
        groups.set(routingKey, group);
      }
      group.files.push(file.lexicalPath);
      group.canonicalFiles.push(file.canonicalPath);
    }

    const pluginSession = await this.#createPluginLintSession([
      ...groups.values(),
    ]);
    try {
      const results: LintResult[] = [];
      // Files are inserted in deterministic path order above. Keep that group
      // order instead of imposing a parent-first config execution dependency.
      for (const group of groups.values()) {
        const response = await this.#service.lint(
          {
            config: group.config,
            eslintPlugins: pluginSession?.eslintPlugins,
            configDirectory: group.configDirectory,
            pluginConfigDirectory: group.pluginConfigDirectory,
            workingDirectory: this.#cwd,
            files: group.files,
            canonicalFiles: group.canonicalFiles,
            // Overlay (e.g. an in-memory tsconfig) for the program over disk files.
            fileContents: this.#resolveOverlay(),
            fix: this.#fix,
          },
          pluginSession
            ? {
                pluginLint: async (request) => {
                  const result = await pluginSession.host.lint(request);
                  return result;
                },
              }
            : {},
        );
        const { contents, bomFiles } = await this.#readDiagnosticContents(
          response,
          group.configDirectory,
        );
        // Seed results from the files Go actually linted (config `ignores`
        // already excluded) rather than the glob matches. Go preserves each
        // selected target's path identity even when its Program uses a canonical
        // or symlinked source path. Fall back to the glob matches if an older
        // binary omits lintedFiles.
        const linted = response.lintedFiles
          ? response.lintedFiles.map((f) =>
              path.isAbsolute(f)
                ? path.normalize(f)
                : path.resolve(group.configDirectory, f),
            )
          : group.files;
        results.push(
          ...this.#toLintResults(
            response,
            group.configDirectory,
            linted,
            contents,
            bomFiles,
          ),
        );
      }
      return results.sort((a, b) =>
        nativePathIdentity.compare(a.filePath, b.filePath),
      );
    } finally {
      await this.#shutdownPluginSession(pluginSession);
    }
  }

  /**
   * Write the `output` of each result back to its file. Mirrors ESLint's
   * guards: only when `output` is a string and `filePath` is absolute (so a
   * lintText `<text>` result is skipped automatically).
   */
  static async outputFixes(results: LintResult[]): Promise<void> {
    await Promise.all(
      results.map(async (r) => {
        if (typeof r.output === 'string' && path.isAbsolute(r.filePath)) {
          await writeFile(r.filePath, r.output);
        }
      }),
    );
  }

  /** Tear down the long-lived Go `--api` process. */
  async close(): Promise<void> {
    this.#closePromise ??= this.#closeResources();
    await this.#closePromise;
  }

  async #closeResources(): Promise<void> {
    this.#closeRequested = true;
    const shutdownActiveHosts = async (): Promise<void> => {
      await Promise.allSettled(
        [...this.#activePluginHosts].map(async (host) => {
          await host.shutdown();
        }),
      );
      this.#activePluginHosts.clear();
    };

    // Unblock any Go lint request currently awaiting pluginLint before queuing
    // the service's exit request behind it. A second sweep catches a host whose
    // initialization raced the first snapshot; such a host also observes
    // #closeRequested and shuts itself down before becoming active.
    await shutdownActiveHosts();
    try {
      await this.#service.close();
    } finally {
      await shutdownActiveHosts();
    }
  }

  async [Symbol.asyncDispose](): Promise<void> {
    await this.close();
  }

  // ── internals ──────────────────────────────────────────────────────────

  // `virtualFiles` re-keyed by `path.resolve(cwd, key)` so relative and absolute
  // keys land where the config resolves on every OS (a bare `/x` would not match
  // a Windows `C:/x`).
  #resolveOverlay(): Record<string, string> | undefined {
    if (!this.#virtualFiles) return undefined;
    const resolved: Record<string, string> = {};
    for (const [p, content] of Object.entries(this.#virtualFiles)) {
      resolved[path.resolve(this.#cwd, p)] = content;
    }
    return resolved;
  }

  async #createPluginLintSession(
    resolvedConfigs: ReadonlyArray<ResolvedConfig>,
  ): Promise<PluginLintSession | null> {
    const configsByDescriptor = new Map<
      string,
      {
        configPath: string;
        configDirectory: string;
        entries: ReadonlyArray<unknown>;
      }
    >();
    for (const resolved of resolvedConfigs) {
      if (!resolved.configPath) continue;
      const key = `${nativePathIdentity.key(resolved.configPath)}\0${resolved.pluginConfigDirectory}`;
      if (!configsByDescriptor.has(key)) {
        configsByDescriptor.set(key, {
          configPath: resolved.configPath,
          configDirectory: resolved.pluginConfigDirectory,
          entries: resolved.config,
        });
      }
    }

    const { eslintPluginEntries, pluginConfigs } = collectPluginMeta([
      ...configsByDescriptor.values(),
    ]);
    if (pluginConfigs.length === 0) return null;

    const createPluginLintHost = await loadPluginHostFactory();
    const host = await createPluginLintHost(pluginConfigs, (record) => {
      process.stderr.write(`[rslint:plugin] ${record.text}\n`);
    });
    if (this.#closeRequested) {
      await host.shutdown();
      throw new Error('rslint service is closing');
    }
    this.#activePluginHosts.add(host);
    return { host, eslintPlugins: eslintPluginEntries };
  }

  async #shutdownPluginSession(
    session: PluginLintSession | null,
  ): Promise<void> {
    if (!session) return;
    try {
      await session.host.shutdown();
    } finally {
      this.#activePluginHosts.delete(session.host);
    }
  }

  async #loadConfigCandidate(
    configPath: string,
    configDirectory: string,
    candidatesByPath: Map<string, ConfigCandidate>,
  ): Promise<ConfigCandidate> {
    const normalizedPath = path.normalize(configPath);
    const key = nativePathIdentity.key(normalizedPath);
    const existing = candidatesByPath.get(key);
    if (existing) return existing;

    const normalizedDirectory = path.normalize(configDirectory);
    let baseConfig: Record<string, unknown>[];
    try {
      baseConfig = normalizeConfig(await loadConfigFile(normalizedPath));
    } catch (error) {
      const candidate: FailedConfigCandidate = {
        status: 'failed',
        configPath: normalizedPath,
        configDirectory: normalizedDirectory,
        error: error instanceof Error ? error : new Error(String(error)),
      };
      candidatesByPath.set(key, candidate);
      return candidate;
    }

    const candidate: LoadedConfigCandidate = {
      status: 'loaded',
      configPath: normalizedPath,
      configDirectory: normalizedDirectory,
      baseConfig,
      resolved: this.#composeConfig(
        baseConfig,
        normalizedDirectory,
        `config:${key}`,
        normalizedPath,
      ),
    };
    candidatesByPath.set(key, candidate);
    return candidate;
  }

  async #resolveLintFileConfigs(
    files: PlannedLintFile[],
    scanDirectories: string[],
    pathResolver: RunPathResolver,
  ): Promise<Map<string, ResolvedConfig | null>> {
    const result = new Map<string, ResolvedConfig | null>();

    // Explicit config modes use one config for every target. Go remains the
    // source of truth for that config's own files/ignores matching.
    if (this.#overrideConfigFile != null) {
      const resolved = await this.#resolveConfig(this.#cwd);
      for (const file of files) result.set(file.lexicalKey, resolved);
      return result;
    }

    const findConfigUp = createCachedConfigFinder();
    const discovered = new Map<
      string,
      { configPath: string; configDirectory: string }
    >();
    const addDiscovered = (
      configPath: string,
      configDirectory = path.dirname(configPath),
    ): void => {
      const normalizedPath = path.normalize(configPath);
      const key = nativePathIdentity.key(normalizedPath);
      if (!discovered.has(key)) {
        discovered.set(key, {
          configPath: normalizedPath,
          configDirectory: path.normalize(configDirectory),
        });
      }
    };

    if (scanDirectories.length > 0) {
      const scanned = await discoverConfigs(
        [],
        scanDirectories,
        this.#cwd,
        null,
      );
      for (const [configPath, configDirectory] of scanned) {
        addDiscovered(configPath, configDirectory);
      }
    }
    for (const file of files) {
      if (!file.explicit) continue;
      let configPath = findConfigUp(path.dirname(file.lexicalPath));
      if (!configPath && file.canonicalKey !== file.lexicalKey) {
        configPath = findConfigUp(path.dirname(file.canonicalPath));
      }
      if (configPath) addDiscovered(configPath);
    }

    if (discovered.size === 0) {
      const emptyConfig = this.#composeConfig(
        [],
        this.#cwd,
        `empty:${nativePathIdentity.key(this.#cwd)}`,
      );
      for (const file of files) result.set(file.lexicalKey, emptyConfig);
      return result;
    }

    const coalescedDiscovered = coalesceCaseAliasedConfigs(
      new Map(
        [...discovered.values()].map(({ configPath, configDirectory }) => [
          configPath,
          configDirectory,
        ]),
      ),
    ).configs;
    discovered.clear();
    for (const [configPath, configDirectory] of coalescedDiscovered) {
      addDiscovered(configPath, configDirectory);
    }
    const discoveredConfigs = new Map(
      [...discovered.values()].map(({ configPath, configDirectory }) => [
        configPath,
        configDirectory,
      ]),
    );

    const candidatesByPath = new Map<string, ConfigCandidate>();
    const loadCandidate = async (
      configPath: string,
      configDirectory: string,
    ): Promise<ConfigCandidate> =>
      this.#loadConfigCandidate(configPath, configDirectory, candidatesByPath);

    const failedQueue: FailedConfigCandidate[] = [];
    for (const { configPath, configDirectory } of discovered.values()) {
      const candidate = await loadCandidate(configPath, configDirectory);
      if (candidate.status === 'failed') failedQueue.push(candidate);
    }

    // Only failed configs need their undiscovered ancestors. This keeps an
    // explicit file target local while still providing normal ancestor
    // fallback, and config discovery remains proportional to unique dirs.
    for (let i = 0; i < failedQueue.length; i++) {
      const failed = failedQueue[i];
      const ancestorPath = findConfigUp(path.dirname(failed.configDirectory));
      if (!ancestorPath) continue;

      const normalizedPath = path.normalize(ancestorPath);
      const configDirectory = path.dirname(normalizedPath);
      const selectedPath =
        findNativeCaseAliasConfigPath(
          normalizedPath,
          configDirectory,
          discoveredConfigs,
        ) ?? normalizedPath;
      const key = nativePathIdentity.key(selectedPath);
      if (candidatesByPath.has(key)) continue;
      discoveredConfigs.set(selectedPath, configDirectory);
      const ancestor = await loadCandidate(selectedPath, configDirectory);
      if (ancestor.status === 'failed') failedQueue.push(ancestor);
    }

    const candidates = [...candidatesByPath.values()];
    const loadedCandidates = candidates.filter(
      (candidate): candidate is LoadedConfigCandidate =>
        candidate.status === 'loaded',
    );
    const failedCandidates = candidates.filter(
      (candidate): candidate is FailedConfigCandidate =>
        candidate.status === 'failed',
    );
    if (loadedCandidates.length === 0) {
      const first = failedCandidates[0];
      throw new Error(
        first
          ? `Failed to load config ${first.configPath}: ${first.error.message}`
          : 'No discovered rslint config could be loaded',
      );
    }
    for (const failed of failedCandidates) {
      console.warn(
        `[rslint] Skipping config ${failed.configPath}: ${failed.error.message}`,
      );
    }

    const candidateIndex = new AncestorPathIndex(
      candidates.map(
        (candidate) => [candidate.configDirectory, candidate] as const,
      ),
    );
    const candidateDirectories = await pathResolver.resolveAll(
      candidates.map((candidate) => candidate.configDirectory),
    );
    const canonicalCandidates = new Map<
      string,
      { canonicalDirectory: string; candidate: ConfigCandidate }
    >();
    for (let index = 0; index < candidates.length; index++) {
      const candidate = candidates[index];
      const directory = candidateDirectories[index];
      if (!canonicalCandidates.has(directory.canonicalKey)) {
        canonicalCandidates.set(directory.canonicalKey, {
          canonicalDirectory: directory.canonicalPath,
          candidate,
        });
      } else {
        const existing = canonicalCandidates.get(directory.canonicalKey);
        if (existing?.candidate !== candidate) {
          throw new Error(
            `Config directories ${JSON.stringify(existing?.candidate.configDirectory)} and ` +
              `${JSON.stringify(candidate.configDirectory)} resolve to the same filesystem location`,
          );
        }
      }
    }
    const canonicalCandidateIndex = new AncestorPathIndex(
      [...canonicalCandidates.values()].map(
        ({ canonicalDirectory, candidate }) =>
          [canonicalDirectory, candidate] as const,
      ),
    );
    const nearestCandidateForFile = (
      file: PlannedLintFile,
    ): ConfigCandidate | undefined =>
      candidateIndex.find(path.dirname(file.lexicalPath)) ??
      canonicalCandidateIndex.find(path.dirname(file.canonicalPath));
    const nearestLoadedCandidate = (
      candidate: ConfigCandidate | undefined,
    ): LoadedConfigCandidate | undefined => {
      let current = candidate;
      while (current) {
        if (current.status === 'loaded') return current;
        current = candidateIndex.findParent(current.configDirectory);
      }
      return undefined;
    };

    const configEntries: ConfigEntry[] = loadedCandidates.map((candidate) => ({
      configDirectory: candidate.configDirectory,
      // Parent-ignore discovery belongs to the authored file config. The API
      // override is evaluated later from cwd and must not move this boundary.
      entries: candidate.baseConfig as RslintConfigEntry[],
    }));

    const explicitConfigDirs = new Set<string>();
    for (const file of files) {
      if (!file.explicit) continue;
      const loaded = nearestLoadedCandidate(nearestCandidateForFile(file));
      if (loaded) explicitConfigDirs.add(loaded.configDirectory);
    }

    const directoryEntries = filterConfigsByParentIgnores(configEntries);
    const directoryConfigDirs = new Set(
      directoryEntries.map((entry) =>
        nativePathIdentity.key(entry.configDirectory),
      ),
    );
    const effectiveEntries = filterConfigsByParentIgnores(
      configEntries,
      explicitConfigDirs,
    );
    const effectiveConfigDirs = new Set(
      effectiveEntries.map((entry) =>
        nativePathIdentity.key(entry.configDirectory),
      ),
    );

    const emptyConfig = this.#composeConfig(
      [],
      this.#cwd,
      `empty:${nativePathIdentity.key(this.#cwd)}`,
    );
    for (const file of files) {
      let candidate = nearestCandidateForFile(file);
      if (!candidate) {
        result.set(file.lexicalKey, emptyConfig);
        continue;
      }

      const allowedDirs = file.explicit
        ? effectiveConfigDirs
        : directoryConfigDirs;
      while (candidate.status === 'failed') {
        candidate = candidateIndex.findParent(candidate.configDirectory);
        if (!candidate) break;
      }

      if (!candidate) {
        result.set(file.lexicalKey, null);
        continue;
      }

      result.set(
        file.lexicalKey,
        allowedDirs.has(nativePathIdentity.key(candidate.configDirectory))
          ? candidate.resolved
          : null,
      );
    }
    return result;
  }

  #getNormalizedOverrideConfig(): Record<string, unknown>[] | null {
    if (this.#overrideConfig == null) return null;
    if (this.#normalizedOverrideConfig) return this.#normalizedOverrideConfig;
    const override = Array.isArray(this.#overrideConfig)
      ? this.#overrideConfig
      : [this.#overrideConfig];
    for (const [index, entry] of override.entries()) {
      if (entry == null || typeof entry !== 'object' || Array.isArray(entry)) {
        continue;
      }
      const plugins = isRecord(entry) ? entry.plugins : undefined;
      if (
        plugins !== null &&
        typeof plugins === 'object' &&
        !Array.isArray(plugins)
      ) {
        throw new Error(
          `[rslint] overrideConfig entry at index ${index} uses object-form "plugins". ` +
            'Community ESLint plugins in overrideConfig are not supported because ' +
            'the plugin worker cannot re-import an in-memory plugin object. Move ' +
            'the plugin declaration to rslint.config.js (or .mjs/.cjs/.ts/.mts/.cts), or use ' +
            'array-form built-in plugins.',
        );
      }
    }
    this.#normalizedOverrideConfig = normalizeConfig(override);
    return this.#normalizedOverrideConfig;
  }

  #composeConfig(
    base: Record<string, unknown>[],
    configDirectory: string,
    routingKey: string,
    configPath?: string,
  ): ResolvedConfig {
    const override = this.#getNormalizedOverrideConfig();
    if (!override) {
      return {
        config: base,
        configPath,
        configDirectory,
        pluginConfigDirectory: configDirectory,
        routingKey,
      };
    }

    // ESLint resolves overrideConfig relative to the selected config's base.
    // Explicit overrideConfigFile already supplies cwd as configDirectory.
    return {
      config: [...base, ...override],
      configPath,
      configDirectory,
      pluginConfigDirectory: configDirectory,
      routingKey,
    };
  }

  async #resolveConfig(
    fromDir: string,
    canonicalFromDir = fromDir,
  ): Promise<ResolvedConfig> {
    let base: Record<string, unknown>[] = [];
    let configPath: string | undefined;
    let configDirectory = this.#cwd;
    let routingKey = `empty:${nativePathIdentity.key(this.#cwd)}`;

    if (this.#overrideConfigFile === true) {
      // Only overrideConfig — no file, no discovery.
      routingKey = `override:${nativePathIdentity.key(this.#cwd)}`;
    } else if (typeof this.#overrideConfigFile === 'string') {
      configPath = path.resolve(this.#cwd, this.#overrideConfigFile);
      base = normalizeConfig(await loadConfigFile(configPath));
      // Explicit overrideConfigFile follows CLI --config semantics: files,
      // ignores, and parserOptions.project resolve from invocation cwd.
      configDirectory = this.#cwd;
      routingKey = `config:${nativePathIdentity.key(configPath)}`;
    } else {
      return this.#resolveAutoConfig(fromDir, canonicalFromDir);
    }

    return this.#composeConfig(base, configDirectory, routingKey, configPath);
  }

  async #resolveAutoConfig(
    fromDir: string,
    canonicalFromDir: string,
  ): Promise<ResolvedConfig> {
    const findConfigUp = createCachedConfigFinder();
    let configPath = findConfigUp(fromDir);
    if (!configPath && !nativePathIdentity.equals(fromDir, canonicalFromDir)) {
      configPath = findConfigUp(canonicalFromDir);
    }
    if (!configPath) {
      return this.#composeConfig(
        [],
        this.#cwd,
        `empty:${nativePathIdentity.key(this.#cwd)}`,
      );
    }

    const candidatesByPath = new Map<string, ConfigCandidate>();
    let nearestFailure: FailedConfigCandidate | undefined;
    const failures: FailedConfigCandidate[] = [];
    while (configPath) {
      const candidate = await this.#loadConfigCandidate(
        configPath,
        path.dirname(configPath),
        candidatesByPath,
      );
      if (candidate.status === 'loaded') {
        for (const failed of failures) {
          console.warn(
            `[rslint] Skipping config ${failed.configPath}: ${failed.error.message}`,
          );
        }
        return candidate.resolved;
      }

      nearestFailure ??= candidate;
      failures.push(candidate);
      configPath = findConfigUp(path.dirname(candidate.configDirectory));
    }

    throw new Error(
      `Failed to load config ${nearestFailure!.configPath}: ${nearestFailure!.error.message}`,
    );
  }

  async #readDiagnosticContents(
    response: LintResponse,
    configDirectory: string,
  ): Promise<{ contents: Record<string, string>; bomFiles: Set<string> }> {
    // Read source for each file that produced a diagnostic so mergeFixes can
    // gap-fill multi-edit fixes (parity with lintText, which has the source
    // in-hand). Only diagnosed files are read; a lint with no findings reads
    // nothing.
    const contents: Record<string, string> = {};
    // Disk files whose bytes start with a UTF-8 BOM. Go reads them through a
    // decoder that strips the BOM, so its fix offsets AND Output are
    // BOM-stripped. We strip the BOM from the source fed to mergeFixes so the
    // gap-fill slices line up with those offsets — and `fix.range` therefore
    // stays BOM-stripped, matching ESLint v10 and the message line/column —
    // then re-prepend the BOM to Output (in toLintResults) so outputFixes writes
    // back the real file. (lintText is unaffected: its overlay keeps the BOM and
    // Go's offsets already include it, so no adjustment is needed there.)
    const bomFiles = new Set<string>();
    for (const d of response.diagnostics ?? []) {
      const abs = path.isAbsolute(d.filePath)
        ? path.normalize(d.filePath)
        : path.resolve(configDirectory, d.filePath);
      if (!(abs in contents)) {
        try {
          const raw = await readFile(abs, 'utf8');
          if (raw.charCodeAt(0) === 0xfeff) {
            bomFiles.add(nativePathIdentity.key(abs));
            contents[abs] = raw.slice(1); // BOM-stripped, matching Go's offsets
          } else {
            contents[abs] = raw;
          }
        } catch {
          // Unreadable (e.g. virtual/deleted) — mergeFixes degrades to first edit.
        }
      }
    }
    return { contents, bomFiles };
  }

  #toLintResults(
    response: LintResponse,
    configDirectory: string,
    files: string[],
    contents?: Record<string, string>,
    bomFiles?: Set<string>,
  ): LintResult[] {
    const toAbs = (p: string): string =>
      path.isAbsolute(p) ? path.normalize(p) : path.resolve(configDirectory, p);

    const contentByPath = new Map<string, string>();
    for (const [filePath, source] of Object.entries(contents ?? {})) {
      contentByPath.set(nativePathIdentity.key(filePath), source);
    }

    // Every linted file gets a result, even with zero messages (ESLint shape).
    // Identity keys coalesce path-casing variants on case-insensitive hosts.
    const byFile = new Map<
      string,
      { filePath: string; messages: LintMessage[] }
    >();
    for (const file of files) {
      const filePath = path.normalize(file);
      const key = nativePathIdentity.key(filePath);
      if (!byFile.has(key)) byFile.set(key, { filePath, messages: [] });
    }

    for (const d of response.diagnostics ?? []) {
      const abs = toAbs(d.filePath);
      const key = nativePathIdentity.key(abs);
      let bucket = byFile.get(key);
      if (!bucket) {
        bucket = { filePath: abs, messages: [] };
        byFile.set(key, bucket);
      }
      bucket.messages.push(toLintMessage(d, contentByPath.get(key)));
    }

    // Wire `output` is keyed by relative path; remap to absolute.
    const outputByPath = new Map<string, string>();
    for (const [rel, fixed] of Object.entries(response.output ?? {})) {
      outputByPath.set(nativePathIdentity.key(toAbs(rel)), fixed);
    }

    const results: LintResult[] = [];
    for (const [key, { filePath, messages }] of byFile) {
      let errorCount = 0;
      let warningCount = 0;
      let fixableErrorCount = 0;
      let fixableWarningCount = 0;
      for (const m of messages) {
        if (m.severity === 2) {
          errorCount++;
          if (m.fix) fixableErrorCount++;
        } else {
          warningCount++;
          if (m.fix) fixableWarningCount++;
        }
      }
      const result: LintResult = {
        filePath,
        messages,
        errorCount,
        warningCount,
        fixableErrorCount,
        fixableWarningCount,
      };
      const output = outputByPath.get(key);
      if (output !== undefined) {
        // Go's Output is BOM-stripped (ApplyRuleFixes runs over the decoded
        // SourceFile text); re-prepend the BOM for a disk file that had one so
        // outputFixes writes back a file identical to the original but for the
        // fix.
        result.output = bomFiles?.has(key) ? '\uFEFF' + output : output;
      }
      results.push(result);
    }
    return results;
  }
}

/** Reshape a wire Diagnostic into an ESLint LintMessage. */
function toLintMessage(d: Diagnostic, sourceText?: string): LintMessage {
  const message: LintMessage = {
    ruleId: d.ruleName || null,
    severity: d.severity === 'error' ? 2 : 1,
    message: d.message,
    line: d.range.start.line,
    column: d.range.start.column,
    endLine: d.range.end.line,
    endColumn: d.range.end.column,
  };
  // ESLint omits messageId when a rule reports a raw message; Go sends "" then.
  if (d.messageId) message.messageId = d.messageId;
  const fix = mergeFixes(d.fixes, sourceText);
  if (fix) message.fix = fix;
  if (d.suggestions && d.suggestions.length > 0) {
    message.suggestions = d.suggestions.map((s) => {
      const sFix = mergeFixes(s.fixes, sourceText);
      return {
        messageId: s.messageId,
        ...(s.data ? { data: s.data } : {}),
        desc: s.message,
        // A suggestion always carries a fix; fall back to an empty edit if a
        // rule somehow emitted none (keeps the ESLint shape non-optional).
        fix: sFix ?? { range: [0, 0], text: '' },
      };
    });
  }
  return message;
}

/**
 * Collapse rslint's per-diagnostic fix edits (possibly several) into ESLint's
 * single `{ range, text }`. A lone edit maps directly; multiple edits merge
 * into one span, filling gaps from sourceText (ESLint's mergeFixes). Without
 * sourceText (e.g. a diagnosed file whose source could not be read), a
 * multi-edit fix degrades to its first edit rather than guessing across a gap.
 *
 * Offsets are flat UTF-16, in the same BOM-stripped space as Go's fix ranges
 * (matching ESLint v10, whose `fix.range` is relative to BOM-stripped source).
 * The caller passes a BOM-stripped sourceText for disk files so the gap-fill
 * slices line up; the BOM is re-applied only to the per-file Output.
 */
function mergeFixes(
  fixes: Fix[] | undefined,
  sourceText?: string,
): LintMessageFix | undefined {
  if (!fixes || fixes.length === 0) return undefined;
  if (fixes.length === 1) {
    return { range: [fixes[0].startPos, fixes[0].endPos], text: fixes[0].text };
  }
  const sorted = [...fixes].sort(
    (a, b) => a.startPos - b.startPos || a.endPos - b.endPos,
  );
  const start = sorted[0].startPos;
  const end = sorted[sorted.length - 1].endPos;
  if (sourceText === undefined) {
    return {
      range: [sorted[0].startPos, sorted[0].endPos],
      text: sorted[0].text,
    };
  }
  let text = '';
  let lastPos = start;
  for (const f of sorted) {
    // Skip an edit overlapping the previous one (ESLint drops overlapping fixes
    // rather than emitting corrupt merged text); rslint rules emit
    // non-overlapping edits per diagnostic, so this is a guard, not a path.
    if (f.startPos < lastPos) continue;
    text += sourceText.slice(lastPos, f.startPos) + f.text;
    lastPos = f.endPos;
  }
  return { range: [start, end], text };
}
