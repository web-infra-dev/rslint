/// <reference lib="esnext.disposable" preserve="true" />

/**
 * The `Rslint` class — the ESLint-style programmatic API (issue #1106).
 *
 * It is a thin JS facade over the low-level `lint()` IPC. JavaScript expands
 * API globs; Go owns config discovery, ownership, ignore semantics, and final
 * lint-target admission. Reverse IPC asks this process to evaluate and
 * normalize only the JavaScript config candidates Go selected. The facade also
 * reshapes wire-level responses into ESLint v10's `LintResult[]` (per-file
 * results, numeric severity, absolute paths, output).
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
import { ConfigModuleHost, normalizeConfig } from '../config/config-loader.js';
import {
  nativePathIdentity,
  RunPathResolver,
  type ResolvedFilesystemPath,
} from './path-identity.js';
import {
  loadPluginHostFactory,
  PluginHostLifecycle,
  stageNativeConfigActivation,
  type PluginLintHost,
} from './native-config-activation.js';
import type { RslintConfigEntry } from '../config/define-config.js';
import type {
  Diagnostic,
  Fix,
  LintInboundHandlers,
  LintResponse,
} from '../types.js';

export interface RslintOptions {
  /** Base directory for config discovery and relative path resolution. */
  cwd?: string;
  /** Extra config appended after the resolved/discovered config (ESLint's overrideConfig). */
  overrideConfig?: RslintConfigEntry | RslintConfigEntry[] | null;
  /**
   * `string` — use this JS/TS config module (no discovery).
   * `true`   — use only `overrideConfig` (no file, no discovery).
   * `null`/absent — auto-discover the nearest config (ESLint v10 semantics; no `false`).
   */
  overrideConfigFile?: string | true | null;
  /** Apply rule auto-fixes; results carry `output` (the JS side persists via outputFixes). */
  fix?: boolean;
  /**
   * In-memory file overlay (path → content) for project inputs (issue #1106):
   * put the `tsconfig.json` that `parserOptions.project` names plus any
   * dependency files here, then lint a buffer with `lintText`. Keys resolve
   * against `cwd` like a linted path (relative or absolute both work); a
   * same-path `lintText` code entry wins. Inside the tsconfig (`files`) and
   * `parserOptions.project`, use relative paths — the TS compiler resolves
   * those, and a bare POSIX-absolute path there has no drive letter on Windows,
   * so it won't match the overlay. The overlay permits filesystem fallback and
   * is not a sandbox. rslint-only — ESLint has no in-memory file map.
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

interface PluginLintSession {
  host: PluginLintHost;
}

interface NativeConfigDiscoverySession {
  handlers: LintInboundHandlers;
  shutdown(): Promise<void>;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

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

export class Rslint {
  readonly #service: RSLintService;
  readonly #cwd: string;
  readonly #overrideConfig?: RslintConfigEntry | RslintConfigEntry[] | null;
  readonly #overrideConfigFile?: string | true | null;
  readonly #fix: boolean;
  readonly #virtualFiles?: Record<string, string>;
  readonly #pluginHosts = new PluginHostLifecycle();
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
    const pathResolver = new RunPathResolver();
    const resolvedFile =
      await pathResolver.resolveWithAncestorFallback(filePath);
    const overrideConfig = this.#getNormalizedOverrideConfig() ?? [];
    const usesNativeDiscovery = this.#overrideConfigFile !== true;
    const discoverySession = usesNativeDiscovery
      ? this.#createNativeConfigDiscoverySession()
      : null;
    try {
      const response = await this.#service.lint(
        {
          config: usesNativeDiscovery ? undefined : overrideConfig,
          configDiscovery: usesNativeDiscovery
            ? {
                mode:
                  typeof this.#overrideConfigFile === 'string'
                    ? 'explicit'
                    : 'auto',
                explicitConfigPath:
                  typeof this.#overrideConfigFile === 'string'
                    ? path.resolve(this.#cwd, this.#overrideConfigFile)
                    : undefined,
                explicitFiles: [true],
                overrideConfig,
              }
            : undefined,
          configDirectory: usesNativeDiscovery ? undefined : this.#cwd,
          workingDirectory: this.#cwd,
          files: [filePath],
          canonicalFiles: [resolvedFile.canonicalPath],
          // Overlay (in-memory tsconfig + deps) underlays the code buffer; a
          // same-path code entry wins so `lintText` always lints `code`.
          fileContents: { ...this.#resolveOverlay(), [filePath]: code },
          fix: this.#fix,
        },
        discoverySession?.handlers ?? {},
      );
      const results = this.#toLintResults(response, this.#cwd, [filePath], {
        [filePath]: code,
      });
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
      await discoverySession?.shutdown();
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

    // Preserve lexical aliases through discovery. Go owns the final canonical
    // target plan: aliases under one config are coalesced there, while aliases
    // governed by different configs must remain visible so it can reject the
    // ambiguous ownership instead of silently dropping one before discovery.
    const selectedFiles = [...plannedFiles].sort((left, right) =>
      nativePathIdentity.compare(left.lexicalPath, right.lexicalPath),
    );
    const overrideConfig = this.#getNormalizedOverrideConfig() ?? [];
    const usesNativeDiscovery = this.#overrideConfigFile !== true;
    const discoverySession = usesNativeDiscovery
      ? this.#createNativeConfigDiscoverySession()
      : null;
    try {
      const files = selectedFiles.map((file) => file.lexicalPath);
      const response = await this.#service.lint(
        {
          config: usesNativeDiscovery ? undefined : overrideConfig,
          configDiscovery: usesNativeDiscovery
            ? {
                mode:
                  typeof this.#overrideConfigFile === 'string'
                    ? 'explicit'
                    : 'auto',
                explicitConfigPath:
                  typeof this.#overrideConfigFile === 'string'
                    ? path.resolve(this.#cwd, this.#overrideConfigFile)
                    : undefined,
                directories: scanDirectories,
                explicitFiles: selectedFiles.map((file) => file.explicit),
                overrideConfig,
              }
            : undefined,
          configDirectory: usesNativeDiscovery ? undefined : this.#cwd,
          workingDirectory: this.#cwd,
          files,
          canonicalFiles: selectedFiles.map((file) => file.canonicalPath),
          // Overlay (e.g. an in-memory tsconfig) for the program over disk files.
          fileContents: this.#resolveOverlay(),
          fix: this.#fix,
        },
        discoverySession?.handlers ?? {},
      );
      const { contents, bomFiles } = await this.#readDiagnosticContents(
        response,
        this.#cwd,
      );
      // The multi-config native path returns absolute identities. Relative
      // values remain accepted for low-level/older implementations.
      const linted = response.lintedFiles
        ? response.lintedFiles.map((file) =>
            path.isAbsolute(file)
              ? path.normalize(file)
              : path.resolve(this.#cwd, file),
          )
        : files;
      return this.#toLintResults(
        response,
        this.#cwd,
        linted,
        contents,
        bomFiles,
      ).sort((a, b) => nativePathIdentity.compare(a.filePath, b.filePath));
    } finally {
      await discoverySession?.shutdown();
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

    // Await in-flight builds, then stop both published and post-prepare staged
    // workers before queuing the service exit behind any active lint request.
    await this.#pluginHosts.shutdownAll();
    try {
      await this.#service.close();
    } finally {
      // A second sweep closes a host whose activation crossed the first
      // snapshot but observed #closeRequested before publication.
      await this.#pluginHosts.shutdownAll();
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

  #createNativeConfigDiscoverySession(): NativeConfigDiscoverySession {
    const configHost = new ConfigModuleHost();
    const transactions = new Set<string>();
    let pluginSession: PluginLintSession | null = null;

    const shutdown = async (): Promise<void> => {
      if (pluginSession) {
        await this.#pluginHosts.shutdown(pluginSession.host);
      }
      pluginSession = null;
      for (const transactionId of transactions) {
        configHost.deleteSession(transactionId);
      }
      transactions.clear();
    };

    return {
      handlers: {
        loadConfigs: async (request) => {
          const response = await configHost.loadConfigs(request);
          transactions.add(request.transactionId);
          return response;
        },
        activateConfigs: async (request) => {
          const { activation, pluginHost } = await stageNativeConfigActivation(
            configHost,
            request,
            loadPluginHostFactory,
            (record) => {
              process.stderr.write(`[rslint:plugin] ${record.text}\n`);
            },
            () => this.#closeRequested,
            this.#pluginHosts,
          );
          if (pluginHost) {
            if (this.#closeRequested) {
              await this.#pluginHosts.shutdown(pluginHost);
              throw new Error('rslint service is closing');
            }
            this.#pluginHosts.publish(pluginHost);
            pluginSession = { host: pluginHost };
          }
          return activation;
        },
        pluginLint: async (request) => {
          if (!pluginSession) {
            throw new Error(
              'rslint API: pluginLint requested without an activated plugin host',
            );
          }
          return pluginSession.host.lint(request);
        },
      },
      shutdown,
    };
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
