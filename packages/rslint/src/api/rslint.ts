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
import { readFile, writeFile } from 'node:fs/promises';
import { glob } from 'tinyglobby';

import { RSLintService } from '../service/service.js';
import { NodeRslintService } from '../internal/node.js';
import { loadConfigFile, normalizeConfig } from '../config/config-loader.js';
import { findJSConfigUp } from '../utils/config-discovery.js';
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

export class Rslint {
  readonly #service: RSLintService;
  readonly #cwd: string;
  readonly #overrideConfig?: RslintConfigEntry | RslintConfigEntry[] | null;
  readonly #overrideConfigFile?: string | true | null;
  readonly #fix: boolean;
  readonly #virtualFiles?: Record<string, string>;

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
    const { config, configDirectory } = await this.#resolveConfig(
      path.dirname(filePath),
    );
    const response = await this.#service.lint({
      config,
      configDirectory,
      workingDirectory: this.#cwd,
      files: [filePath],
      // Overlay (in-memory tsconfig + deps) underlays the code buffer; a
      // same-path code entry wins so `lintText` always lints `code`.
      fileContents: { ...this.#resolveOverlay(), [filePath]: code },
      fix: this.#fix,
    });
    const results = this.#toLintResults(response, configDirectory, [filePath], {
      [filePath]: code,
    });
    // ESLint's lintText returns exactly one result — for the linted buffer. An
    // in-memory overlay dependency file (pulled into the program and matching
    // the config) can carry its own diagnostics; keep only the linted file so
    // they neither leak a second result nor get written back by outputFixes.
    const primary = results.filter((r) => r.filePath === filePath);
    // ESLint: with no filePath, the result's path is the non-absolute "<text>"
    // sentinel, so outputFixes skips it (we lint at a synthetic absolute path
    // so Go can build a program, then relabel). A user-supplied filePath stays
    // absolute — outputFixes writing it back is then the caller's intent.
    if (options.filePath == null) {
      for (const r of primary) {
        if (r.filePath === filePath) r.filePath = '<text>';
      }
    }
    return primary;
  }

  /**
   * Lint files matched by glob pattern(s), resolved against `cwd`. Results are
   * ordered by the linted file's path (deterministic), not by glob-walk order.
   */
  async lintFiles(patterns: string | string[]): Promise<LintResult[]> {
    const globs = Array.isArray(patterns) ? patterns : [patterns];
    const matched = await glob(globs, {
      cwd: this.#cwd,
      absolute: true,
      onlyFiles: true,
    });
    const files = matched.map((f) => path.normalize(f));
    if (files.length === 0) return [];
    const { config, configDirectory } = await this.#resolveConfig(this.#cwd);
    const response = await this.#service.lint({
      config,
      configDirectory,
      workingDirectory: this.#cwd,
      files,
      // Overlay (e.g. an in-memory tsconfig) for the program over disk files.
      fileContents: this.#resolveOverlay(),
      fix: this.#fix,
    });
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
            bomFiles.add(abs);
            contents[abs] = raw.slice(1); // BOM-stripped, matching Go's offsets
          } else {
            contents[abs] = raw;
          }
        } catch {
          // Unreadable (e.g. virtual/deleted) — mergeFixes degrades to first edit.
        }
      }
    }
    // Seed results from the files Go actually linted (config `ignores` already
    // excluded, paths program-canonical) rather than the glob matches: an
    // ignored match produces no phantom empty result, and a symlinked glob path
    // can't duplicate a result whose diagnostic is keyed to the program path.
    // Fall back to the glob matches if an older binary omits lintedFiles.
    const linted = response.lintedFiles
      ? response.lintedFiles.map((f) =>
          path.isAbsolute(f)
            ? path.normalize(f)
            : path.resolve(configDirectory, f),
        )
      : files;
    return this.#toLintResults(
      response,
      configDirectory,
      linted,
      contents,
      bomFiles,
    );
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
    await this.#service.close();
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

  async #resolveConfig(fromDir: string): Promise<{
    config: Record<string, unknown>[];
    configDirectory: string;
  }> {
    let base: Record<string, unknown>[] = [];
    let configDirectory = this.#cwd;

    if (this.#overrideConfigFile === true) {
      // Only overrideConfig — no file, no discovery.
    } else if (typeof this.#overrideConfigFile === 'string') {
      const configPath = path.resolve(this.#cwd, this.#overrideConfigFile);
      base = normalizeConfig(await loadConfigFile(configPath));
      configDirectory = path.dirname(configPath);
    } else {
      // null / absent: auto-discover the nearest config from fromDir upward.
      const configPath = findJSConfigUp(fromDir);
      if (configPath) {
        base = normalizeConfig(await loadConfigFile(configPath));
        configDirectory = path.dirname(configPath);
      }
    }

    if (this.#overrideConfig != null) {
      const override = Array.isArray(this.#overrideConfig)
        ? this.#overrideConfig
        : [this.#overrideConfig];
      base = [...base, ...normalizeConfig(override)];
    }

    return { config: base, configDirectory };
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

    // Every linted file gets a result, even with zero messages (ESLint shape).
    const byFile = new Map<string, LintMessage[]>();
    for (const f of files) byFile.set(path.normalize(f), []);

    for (const d of response.diagnostics ?? []) {
      const abs = toAbs(d.filePath);
      let messages = byFile.get(abs);
      if (!messages) {
        messages = [];
        byFile.set(abs, messages);
      }
      messages.push(toLintMessage(d, contents?.[abs]));
    }

    // Wire `output` is keyed by relative path; remap to absolute.
    const outputByAbs = new Map<string, string>();
    for (const [rel, fixed] of Object.entries(response.output ?? {})) {
      outputByAbs.set(toAbs(rel), fixed);
    }

    const results: LintResult[] = [];
    for (const [filePath, messages] of byFile) {
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
      const output = outputByAbs.get(filePath);
      if (output !== undefined) {
        // Go's Output is BOM-stripped (ApplyRuleFixes runs over the decoded
        // SourceFile text); re-prepend the BOM for a disk file that had one so
        // outputFixes writes back a file identical to the original but for the
        // fix.
        result.output = bomFiles?.has(filePath) ? '\uFEFF' + output : output;
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
