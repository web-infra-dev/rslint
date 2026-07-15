/// <reference lib="esnext.disposable" preserve="true" />

/**
 * The `Rslint` class — the ESLint-style programmatic API (issue #1106).
 *
 * It is a thin JS facade over the low-level `lint()` IPC. Go expands API
 * inputs and owns config discovery, ownership, ignore semantics, and final
 * lint-target admission. Reverse IPC asks this process to evaluate and
 * normalize only the JavaScript config candidates Go selected. The facade also
 * reshapes wire-level responses into ESLint v10's `LintResult[]` (per-file
 * results, numeric severity, absolute paths, output).
 *
 * The Go side is the single long-lived `--api` process held by the underlying
 * service; `close()` tears it down (mirrors RSLintService.close()).
 */
import path from 'node:path';
import { readFile, writeFile } from 'node:fs/promises';

import { RSLintService } from '../service/service.js';
import { NodeRslintService } from '../internal/node.js';
import {
  ConfigModuleHost,
  normalizeConfig,
  type ActivateConfigsRequest,
  type ActivateConfigsResponse,
  type ConfigPredicate,
} from '../config/config-loader.js';
import { nativePathIdentity } from './path-identity.js';
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
  line?: number; // 1-based
  column?: number; // 1-based, UTF-16
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

interface PluginLintHost {
  lint(request: unknown): Promise<unknown>;
  shutdown(): Promise<void>;
}

/** @internal Resource registry shared by activation and API close tests. */
export class PluginHostLifecycle {
  readonly #pendingBuilds = new Set<Promise<unknown>>();
  readonly #staged = new Set<PluginLintHost>();
  readonly #active = new Set<PluginLintHost>();
  readonly #shutdowns = new WeakMap<PluginLintHost, Promise<void>>();

  async trackBuild<T>(build: Promise<T>): Promise<T> {
    this.#pendingBuilds.add(build);
    void build.then(
      () => this.#pendingBuilds.delete(build),
      () => this.#pendingBuilds.delete(build),
    );
    const result = await build;
    return result;
  }

  stage(host: PluginLintHost): void {
    this.#staged.add(host);
  }

  publish(host: PluginLintHost): void {
    this.#staged.delete(host);
    this.#active.add(host);
  }

  async shutdown(host: PluginLintHost): Promise<void> {
    this.#staged.delete(host);
    this.#active.delete(host);
    let shutdown = this.#shutdowns.get(host);
    if (!shutdown) {
      shutdown = host.shutdown();
      this.#shutdowns.set(host, shutdown);
    }
    await shutdown;
  }

  async shutdownAll(): Promise<void> {
    while (this.#pendingBuilds.size > 0) {
      await Promise.allSettled([...this.#pendingBuilds]);
    }
    const shutdowns: Promise<void>[] = [];
    for (const host of new Set([...this.#staged, ...this.#active])) {
      shutdowns.push(this.shutdown(host));
    }
    await Promise.allSettled(shutdowns);
  }
}

interface PluginLintSession {
  host: PluginLintHost;
}

interface NativeConfigDiscoverySession {
  handlers: LintInboundHandlers;
  shutdown(): Promise<void>;
}

type CreatePluginLintHost = (
  configs: Array<{ configPath: string; configDirectory: string }>,
  onLog?: (record: { level: string; source: string; text: string }) => void,
) => Promise<PluginLintHost>;

function isPluginHostFactoryModule(
  value: unknown,
): value is { createPluginLintHost: CreatePluginLintHost } {
  return isRecord(value) && typeof value.createPluginLintHost === 'function';
}

function isPluginLintHost(value: unknown): value is PluginLintHost {
  return (
    isRecord(value) &&
    typeof value.lint === 'function' &&
    typeof value.shutdown === 'function'
  );
}

let pluginHostFactoryPromise: Promise<CreatePluginLintHost> | undefined;

async function loadPluginHostFactory(): Promise<CreatePluginLintHost> {
  pluginHostFactoryPromise ??= (async () => {
    // A package self-reference resolves to src under the test condition and to
    // dist/eslint-plugin in published builds. Keep it runtime-only: the library
    // declaration build deliberately excludes the worker implementation.
    const pluginEntry: string = '@rslint/core/eslint-plugin';
    const module: unknown = await import(/* webpackIgnore: true */ pluginEntry);
    if (!isPluginHostFactoryModule(module)) {
      throw new Error(
        'rslint ESLint-plugin entry does not export createPluginLintHost',
      );
    }
    return module.createPluginLintHost;
  })();
  const factory = await pluginHostFactoryPromise;
  return factory;
}

/**
 * Stage one native-discovery activation without exposing a worker imported
 * from config bytes that differ from Go's normalized entries.
 *
 * @internal Exported from this source module for lifecycle regression tests;
 * it is intentionally not re-exported from the package root.
 */
export async function stageNativeConfigActivation(
  configHost: ConfigModuleHost,
  request: ActivateConfigsRequest,
  getPluginHostFactory: () => Promise<CreatePluginLintHost>,
  onLog: (record: { level: string; source: string; text: string }) => void,
  isClosing: () => boolean,
  lifecycle?: PluginHostLifecycle,
): Promise<{
  activation: ActivateConfigsResponse;
  pluginHost: PluginLintHost | null;
}> {
  let pluginHost: PluginLintHost | null = null;
  try {
    const activation = await configHost.activateConfigs(
      request,
      undefined,
      async (candidate) => {
        if (candidate.pluginConfigs.length === 0) return;
        if (isClosing()) throw new Error('rslint service is closing');
        const createPluginLintHost = await getPluginHostFactory();
        if (isClosing()) throw new Error('rslint service is closing');
        const build = (async () => {
          const host = await createPluginLintHost(
            candidate.pluginConfigs,
            onLog,
          );
          lifecycle?.stage(host);
          return host;
        })();
        pluginHost = await (lifecycle?.trackBuild(build) ?? build);
        if (isClosing()) throw new Error('rslint service is closing');
      },
    );
    // close() can start while the post-prepare fingerprint read is pending.
    if (isClosing()) throw new Error('rslint service is closing');
    return { activation, pluginHost };
  } catch (error) {
    try {
      const createdHost: unknown = pluginHost;
      if (isPluginLintHost(createdHost)) {
        await (lifecycle?.shutdown(createdHost) ?? createdHost.shutdown());
      }
    } catch {
      // Preserve the activation error: the source mismatch is what Go must
      // receive, while the host has still been asked to terminate.
    }
    throw error;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

export class Rslint {
  readonly #service: RSLintService;
  readonly #cwd: string;
  readonly #overrideConfig?: RslintConfigEntry | RslintConfigEntry[] | null;
  readonly #overrideConfigFile?: string | true | null;
  readonly #fix: boolean;
  readonly #virtualFiles?: Record<string, string>;
  readonly #pluginHosts = new PluginHostLifecycle();
  readonly #configHost = new ConfigModuleHost();
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
    const { entries: overrideConfig, predicates: overridePredicates } =
      this.#normalizeOverrideConfigForRequest();
    const discoveryMode =
      this.#overrideConfigFile === true
        ? 'inline'
        : typeof this.#overrideConfigFile === 'string'
          ? 'explicit'
          : 'auto';
    const discoverySession =
      this.#createNativeConfigDiscoverySession(overridePredicates);
    try {
      const response = await this.#service.lint(
        {
          configDiscovery: {
            mode: discoveryMode,
            inputs: [filePath],
            explicitConfigPath:
              typeof this.#overrideConfigFile === 'string'
                ? path.resolve(this.#cwd, this.#overrideConfigFile)
                : undefined,
            overrideConfig,
          },
          workingDirectory: this.#cwd,
          // Overlay (in-memory tsconfig + deps) underlays the code buffer; a
          // same-path code entry wins so `lintText` always lints `code`.
          fileContents: { ...this.#resolveOverlay(), [filePath]: code },
          fix: this.#fix,
        },
        discoverySession.handlers,
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
      await discoverySession.shutdown();
    }
  }

  /** Lint files matched by glob pattern(s), resolved against `cwd`. */
  async lintFiles(patterns: string | string[]): Promise<LintResult[]> {
    let globs: string[];
    if (patterns === '' || (Array.isArray(patterns) && patterns.length === 0)) {
      globs = ['.'];
    } else {
      if (
        (typeof patterns !== 'string' && !Array.isArray(patterns)) ||
        (typeof patterns === 'string' && patterns.length === 0) ||
        (Array.isArray(patterns) &&
          patterns.some(
            (pattern) => typeof pattern !== 'string' || pattern.length === 0,
          ))
      ) {
        throw new Error(
          "'patterns' must be a non-empty string or an array of non-empty strings",
        );
      }
      globs = Array.isArray(patterns) ? patterns : [patterns];
    }
    const { entries: overrideConfig, predicates: overridePredicates } =
      this.#normalizeOverrideConfigForRequest();
    const discoveryMode =
      this.#overrideConfigFile === true
        ? 'inline'
        : typeof this.#overrideConfigFile === 'string'
          ? 'explicit'
          : 'auto';
    const discoverySession =
      this.#createNativeConfigDiscoverySession(overridePredicates);
    try {
      const response = await this.#service.lint(
        {
          configDiscovery: {
            mode: discoveryMode,
            inputs: globs,
            explicitConfigPath:
              typeof this.#overrideConfigFile === 'string'
                ? path.resolve(this.#cwd, this.#overrideConfigFile)
                : undefined,
            overrideConfig,
          },
          workingDirectory: this.#cwd,
          // Overlay (e.g. an in-memory tsconfig) for the program over disk files.
          fileContents: this.#resolveOverlay(),
          fix: this.#fix,
        },
        discoverySession.handlers,
      );
      const { contents, bomFiles } = await this.#readDiagnosticContents(
        response,
        this.#cwd,
      );
      const linted = response.resultFiles.map((file) =>
        path.isAbsolute(file)
          ? path.normalize(file)
          : path.resolve(this.#cwd, file),
      );
      return this.#toLintResults(
        response,
        this.#cwd,
        linted,
        contents,
        bomFiles,
      );
    } finally {
      await discoverySession.shutdown();
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

  #createNativeConfigDiscoverySession(
    externalPredicates: ReadonlyMap<string, ConfigPredicate>,
  ): NativeConfigDiscoverySession {
    const configHost = this.#configHost;
    const transactions = new Set<string>();
    let pluginSession: PluginLintSession | null = null;

    const shutdown = async (): Promise<void> => {
      try {
        if (pluginSession) {
          await this.#pluginHosts.shutdown(pluginSession.host);
        }
      } finally {
        pluginSession = null;
        for (const transactionId of transactions) {
          configHost.deleteSession(transactionId);
        }
        transactions.clear();
      }
    };

    return {
      handlers: {
        loadConfigs: async (request) => {
          // Register before module evaluation: a sibling discovery branch can
          // fail while this import is still pending, causing the outer lint to
          // enter shutdown before loadConfigs settles. The host allocates its
          // session before awaiting the import, so shutdown can safely detach
          // that state even when JavaScript module evaluation itself cannot be
          // cancelled.
          transactions.add(request.transactionId);
          return configHost.loadConfigs(request);
        },
        evaluateConfigPredicates: async (request) => {
          transactions.add(request.transactionId);
          return configHost.evaluateConfigPredicates(
            request,
            undefined,
            externalPredicates,
          );
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

  #normalizeOverrideConfigForRequest(): {
    entries: Record<string, unknown>[];
    predicates: ReadonlyMap<string, ConfigPredicate>;
  } {
    if (this.#overrideConfig == null) {
      return { entries: [], predicates: new Map() };
    }
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
    const predicates = new Map<string, ConfigPredicate>();
    let predicateSequence = 0;
    const entries = normalizeConfig(override, (predicate) => {
      const predicateId = `override:predicate-${String(++predicateSequence).padStart(6, '0')}`;
      predicates.set(predicateId, predicate);
      return predicateId;
    });
    return { entries, predicates };
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

    for (const warning of response.fileWarnings ?? []) {
      const abs = toAbs(warning.filePath);
      const key = nativePathIdentity.key(abs);
      let bucket = byFile.get(key);
      if (!bucket) {
        bucket = { filePath: abs, messages: [] };
        byFile.set(key, bucket);
      }
      bucket.messages.push({
        ruleId: null,
        severity: 1,
        message: warning.message,
      });
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
