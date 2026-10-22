/// <reference lib="esnext.disposable" preserve="true" />

/**
 * The ESLint-style programmatic API. Configuration is evaluated in JavaScript,
 * then the in-process Rust addon owns parsing, type queries, rules and fixes.
 * Lint operations never start a service process and never write input files.
 */
import path from 'node:path';
import { writeFile } from 'node:fs/promises';
import type {
  RslintConfig,
  RslintConfigEntry,
} from '../config/define-config.js';
import { resolveRustConfig } from '../rust/config.js';
import { lintRustProjectsSync, toPublicLintResults } from '../rust/projects.js';
import { nativePathIdentity } from './path-identity.js';

export interface RslintOptions {
  /**
   * Working directory for targets and discovery. It is also the authored path
   * base for relative fields in inline `overrideConfig` entries that omit
   * `basePath`.
   */
  cwd?: string;
  /**
   * Extra config appended to the selected config array. An authored `basePath`
   * inherits that array's base (the discovered module directory in automatic
   * mode, or `cwd` for an explicit/override-only array), matching ESLint.
   * Relative fields in entries without `basePath` retain Rslint's existing
   * `cwd`-relative behavior.
   */
  overrideConfig?: RslintConfigEntry | RslintConfig | null;
  /**
   * `string` — use this JS/TS config module (no discovery); an authored
   *            `basePath` resolves from `cwd`, matching ESLint's explicit
   *            config behavior. Entries without `basePath` retain rslint's
   *            existing module-directory-relative behavior.
   * `true`   — use only `overrideConfig` (no file, no discovery).
   * `null`/absent — auto-discover the nearest config (ESLint v10 semantics; no `false`).
   */
  overrideConfigFile?: string | true | null;
  /** Apply rule auto-fixes; results carry `output` (the JS side persists via outputFixes). */
  fix?: boolean;
  /**
   * In-memory file overlay (path → content) for project inputs (issue #1106):
   * put a config selected by `parserOptions.project` or discovered by
   * `projectService`, plus dependency files, here before calling `lintText`.
   * Automatic project discovery uses the same overlay as source loading. Keys resolve
   * against `cwd` like a linted path (relative or absolute both work); a
   * same-path `lintText` code entry wins. `parserOptions.project` resolves from
   * the config entry's effective base, while paths inside the tsconfig resolve
   * from that tsconfig's directory. Prefer relative paths for portability: a
   * bare POSIX-absolute path has no drive letter on Windows. The overlay permits
   * filesystem fallback and is not a sandbox. rslint-only — ESLint has no
   * in-memory file map.
   *
   * Give the tsconfig explicit `files`, not a broad `include` glob: a glob is
   * expanded against the real filesystem from the tsconfig's directory.
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
  output?: string; // present when fix:true applied at least one fix to the file
}

export class Rslint {
  readonly #cwd: string;
  readonly #overrideConfig?: RslintConfigEntry | RslintConfig | null;
  readonly #overrideConfigFile?: string | true | null;
  readonly #fix: boolean;
  readonly #virtualFiles?: Record<string, string>;
  #requestQueue: Promise<void> = Promise.resolve();
  #closeRequested = false;
  #closePromise?: Promise<void>;

  constructor(options: RslintOptions = {}) {
    this.#cwd = options.cwd ? path.resolve(options.cwd) : process.cwd();
    this.#overrideConfig = options.overrideConfig;
    this.#overrideConfigFile = options.overrideConfigFile;
    this.#fix = options.fix ?? false;
    this.#virtualFiles = options.virtualFiles;
  }

  /**
   * Lint an in-memory buffer, including a path which has never existed on disk.
   * Only this buffer is a lint target; overlay dependencies supply type data.
   * The Rust source reader strips a leading BOM for diagnostic/fix coordinates
   * and preserves it in complete fixed output unless a fix removes it.
   */
  async lintText(
    code: string,
    options: { filePath?: string } = {},
  ): Promise<LintResult[]> {
    return this.#enqueue(async () => {
      if (typeof code !== 'string')
        throw new Error('lintText code must be a string');
      const filePath = path.resolve(
        this.#cwd,
        options.filePath ?? '__text__.ts',
      );
      const sources = this.#resolveOverlay();
      const config = await this.#resolveConfig();
      const result = lintRustProjectsSync({
        ...config,
        buffer: { path: filePath, text: code },
        sources,
        fix: this.#fix,
      });
      // Keep the requested lexical spelling even if its selected Program uses
      // a physical alias. Dependency sources must never leak additional results.
      const primary = toPublicLintResults(result).filter((file) =>
        nativePathIdentity.equals(file.filePath, filePath),
      );
      if (primary.length !== result.files.length)
        throw new Error(
          'Rust lintText returned a result for a dependency file',
        );
      if (options.filePath == null) {
        for (const file of primary) file.filePath = '<text>';
      }
      return primary;
    });
  }

  /** Lint selected disk files, preserving zero-diagnostic files and stable order. */
  async lintFiles(patterns: string | string[]): Promise<LintResult[]> {
    return this.#enqueue(async () => {
      const targets = Array.isArray(patterns) ? [...patterns] : [patterns];
      if (targets.length === 0) return [];
      if (targets.some((target) => typeof target !== 'string' || !target))
        throw new Error('lintFiles patterns must be non-empty strings');
      const sources = this.#resolveOverlay();
      const config = await this.#resolveConfig();
      return toPublicLintResults(
        lintRustProjectsSync({
          ...config,
          patterns: targets,
          sources,
          fix: this.#fix,
        }),
      ).sort((left, right) =>
        nativePathIdentity.compare(left.filePath, right.filePath),
      );
    });
  }

  /** Persist only explicit absolute result paths, never the synthetic <text>. */
  static async outputFixes(results: LintResult[]): Promise<void> {
    await Promise.all(
      results.map(async (result) => {
        if (
          typeof result.output === 'string' &&
          path.isAbsolute(result.filePath)
        ) {
          await writeFile(result.filePath, result.output);
        }
      }),
    );
  }

  /** Finish already accepted requests; subsequent lint calls reject. */
  async close(): Promise<void> {
    this.#closeRequested = true;
    this.#closePromise ??= this.#requestQueue;
    await this.#closePromise;
  }

  async [Symbol.asyncDispose](): Promise<void> {
    await this.close();
  }

  #enqueue<T>(operation: () => Promise<T>): Promise<T> {
    if (this.#closeRequested)
      return Promise.reject(new Error('rslint service is closing'));
    // Match the public API's request ordering. A failure belongs to its caller
    // and must not poison later requests or cause an unhandled close rejection.
    const pending = this.#requestQueue.then(operation);
    this.#requestQueue = pending.then(
      () => undefined,
      () => undefined,
    );
    return pending;
  }

  #resolveConfig() {
    return resolveRustConfig({
      cwd: this.#cwd,
      overrideConfig: this.#overrideConfig,
      overrideConfigFile: this.#overrideConfigFile,
    });
  }

  #resolveOverlay(): Array<{ path: string; text: string }> {
    const resolved = new Map<string, string>();
    for (const [filename, text] of Object.entries(this.#virtualFiles ?? {}))
      resolved.set(path.resolve(this.#cwd, filename), text);
    return [...resolved].map(([filename, text]) => ({ path: filename, text }));
  }
}
