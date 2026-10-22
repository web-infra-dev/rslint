import { createHash } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { parseArgs } from 'node:util';
import {
  isJSConfigFile,
  isOutputFormat,
  type OutputFormat,
} from '../utils/args.js';
import { resolveRustConfig } from './config.js';
import type { RustLintDiagnostic } from './engine.js';
import {
  lintRustProjectsWithSourceSnapshotsSync,
  type RustProjectsResult,
} from './projects.js';

export interface RustCLIWriter {
  write(text: string): unknown;
  isTTY?: boolean;
}

export interface RustCLIOptions {
  cwd?: string;
  /** Includes wrapper startup in the default report's elapsed time. */
  startTime?: number;
  stdout?: RustCLIWriter;
  stderr?: RustCLIWriter;
}

const HELP = `🚀 Rslint - Rocket Speed Linter

Usage:
  rslint [OPTIONS] [files...]

Options:
  -c, --config PATH     Which JS/TS module config file to use.
  --format FORMAT       Output format: default | jsonline | github | gitlab
  --fix                 Automatically fix problems
  --no-color            Disable colored output
  --force-color         Force colored output
  --quiet               Report errors only
  --max-warnings Int    Number of warnings to trigger nonzero exit code
  --rule RULE           Rule override, e.g. 'no-console: error' (repeatable)
  --singleThreaded      Disable parallel compiler work
  -h, --help            Show help

The Rust migration currently requires --config. Automatic config discovery,
--init, --type-check, --type-check-only, --timing, --trace, and --cpuprof
are not implemented and return an error; no previous backend is invoked.
`;

function parseCLI(argv: readonly string[]) {
  return parseArgs({
    args: [...argv],
    allowPositionals: true,
    strict: true,
    options: {
      config: { type: 'string', short: 'c' },
      help: { type: 'boolean', short: 'h' },
      fix: { type: 'boolean' },
      format: { type: 'string' },
      quiet: { type: 'boolean' },
      'no-color': { type: 'boolean' },
      'force-color': { type: 'boolean' },
      'max-warnings': { type: 'string' },
      rule: { type: 'string', multiple: true },
      singleThreaded: { type: 'boolean' },
      'start-time': { type: 'string' },
      init: { type: 'boolean' },
      'type-check': { type: 'boolean' },
      'type-check-only': { type: 'boolean' },
      timing: { type: 'boolean' },
      trace: { type: 'string' },
      cpuprof: { type: 'string' },
    },
  });
}

function ruleOverrides(flags: readonly string[]) {
  const rules: Record<string, unknown> = Object.create(null);
  for (const flag of flags) {
    const colon = flag.indexOf(':');
    if (colon < 0)
      throw new Error(
        `invalid --rule format ${JSON.stringify(flag)}: expected "ruleName: severity"`,
      );
    const name = flag.slice(0, colon).trim();
    const value = flag.slice(colon + 1).trim();
    if (!name || !value)
      throw new Error(
        `invalid --rule format ${JSON.stringify(flag)}: empty ${name ? 'value' : 'rule name'}`,
      );
    rules[name] = value.startsWith('[') ? JSON.parse(value) : value;
  }
  return flags.length ? [{ rules }] : undefined;
}

function colorEnabled(
  values: ReturnType<typeof parseCLI>['values'],
  stdout: RustCLIWriter,
): boolean {
  if (values['no-color']) return false;
  if (values['force-color']) return true;
  if (process.env.FORCE_COLOR !== undefined)
    return ['', '1', 'true', '2', '3'].includes(process.env.FORCE_COLOR);
  if (process.env.NO_COLOR !== undefined || process.env.TERM === 'dumb')
    return false;
  return Boolean(process.env.GITHUB_ACTIONS || stdout.isTTY);
}

function colors(enabled: boolean) {
  const wrap = (open: string, close: string) => (text: string) =>
    enabled ? `\u001b[${open}m${text}\u001b[${close}m` : text;
  return {
    rule: wrap('92', '0'),
    file: wrap('36;3', '0'),
    start: wrap('36;1', '0'),
    error: wrap('31;1', '0'),
    success: wrap('32;1', '0'),
    dim: wrap('2', '0'),
    warn: wrap('33', '0'),
  };
}

type Colors = ReturnType<typeof colors>;
type SourceSnapshot = {
  bytes: Buffer;
  text: string;
  canonicalPath: string;
  device: number;
  inode: number;
};
type View = { filePath: string; message: RustLintDiagnostic; text: string };

function decodeSource(bytes: Buffer): string {
  if (bytes.length >= 2 && bytes[0] === 0xff && bytes[1] === 0xfe)
    return `\uFEFF${bytes.subarray(2).toString('utf16le')}`;
  if (bytes.length >= 2 && bytes[0] === 0xfe && bytes[1] === 0xff) {
    const body = Buffer.from(bytes.subarray(2));
    if (body.length % 2 !== 0)
      throw new Error('Invalid UTF-16 source byte length');
    return `\uFEFF${body.swap16().toString('utf16le')}`;
  }
  return bytes.toString('utf8');
}

function sourceSnapshot(filename: string): SourceSnapshot {
  const descriptor = fs.openSync(filename, 'r');
  try {
    const before = fs.fstatSync(descriptor);
    const bytes = fs.readFileSync(descriptor);
    const after = fs.fstatSync(descriptor);
    const canonicalPath = fs.realpathSync.native(filename);
    const current = fs.statSync(filename);
    if (
      !before.isFile() ||
      before.dev !== current.dev ||
      before.ino !== current.ino ||
      before.size !== after.size ||
      before.mtimeMs !== after.mtimeMs ||
      before.ctimeMs !== after.ctimeMs
    )
      throw new Error(`source changed while reading: ${filename}`);
    return {
      bytes,
      text: decodeSource(bytes),
      canonicalPath,
      device: before.dev,
      inode: before.ino,
    };
  } finally {
    fs.closeSync(descriptor);
  }
}

/** Whole-delta stale-source validation precedes every terminal write. */
function commitFixes(
  result: RustProjectsResult,
  originals: ReadonlyMap<string, SourceSnapshot>,
): void {
  const changes = new Map<string, { path: string; output: string }>();
  for (const file of result.files) {
    if (file.output === undefined) continue;
    const before = originals.get(file.filePath);
    if (!before)
      throw new Error(`Missing original fix source: ${file.filePath}`);
    const current = sourceSnapshot(file.filePath);
    if (
      current.canonicalPath !== before.canonicalPath ||
      current.device !== before.device ||
      current.inode !== before.inode ||
      !current.bytes.equals(before.bytes)
    )
      throw new Error(
        `fix target ${JSON.stringify(file.filePath)} changed after lint generation`,
      );
    const previous = changes.get(before.canonicalPath);
    if (previous && previous.output !== file.output)
      throw new Error(
        `conflicting fixes for aliases of ${JSON.stringify(file.filePath)}`,
      );
    // The previous CLI writes decoded source as UTF-8, including any source BOM.
    if (file.output !== before.text)
      changes.set(before.canonicalPath, {
        path: file.filePath,
        output: file.output,
      });
  }
  const failures: Error[] = [];
  for (const change of changes.values()) {
    try {
      fs.writeFileSync(change.path, change.output, {
        encoding: 'utf8',
        mode: 0o644,
      });
    } catch (cause) {
      failures.push(
        new Error(
          `write fixed file ${JSON.stringify(change.path)}: ${String(cause)}`,
          { cause },
        ),
      );
    }
  }
  if (failures.length)
    throw new AggregateError(
      failures,
      failures.map((error) => error.message).join('\n'),
    );
}

function escapeData(value: string): string {
  return value
    .replaceAll('%', '%25')
    .replaceAll('\r', '%0D')
    .replaceAll('\n', '%0A');
}

function escapeProperty(value: string): string {
  return escapeData(value).replaceAll(':', '%3A').replaceAll(',', '%2C');
}

// Go's encoding/json escapes these code points even when the output is not HTML.
function json(value: unknown): string {
  return JSON.stringify(value).replace(
    /[<>&\u2028\u2029]/g,
    (character) =>
      `\\u${character.charCodeAt(0).toString(16).padStart(4, '0')}`,
  );
}

function machineOutput(
  format: Exclude<OutputFormat, 'default'>,
  views: View[],
): string {
  if (format === 'jsonline')
    return views
      .map(
        ({ filePath, message: m }) =>
          json({
            ruleName: m.ruleId,
            message: m.message,
            filePath,
            range: {
              start: { line: m.line, column: m.column },
              end: { line: m.endLine, column: m.endColumn },
            },
            severity: m.severity === 2 ? 'error' : 'warn',
          }) + '\n',
      )
      .join('');
  if (format === 'github')
    return views
      .map(
        ({ filePath, message: m }) =>
          `::${m.severity === 2 ? 'error' : 'warning'} file=${escapeProperty(filePath)},line=${m.line},endLine=${m.endLine},col=${m.column},endColumn=${m.endColumn},title=${escapeProperty(m.ruleId)}::${escapeData(m.message)}\n`,
      )
      .join('');
  const salts = new Map<string, number>();
  const emitted = new Set<string>();
  return (
    json(
      views.map(({ filePath, message: m }) => {
        const input = `${filePath}:${m.ruleId}:${m.message}:${m.line}:${m.column}:${m.endLine}:${m.endColumn}`;
        let salt = salts.get(input) ?? 0;
        let fingerprint: string;
        do {
          fingerprint = createHash('md5')
            .update(salt ? `${input}:${salt}` : input)
            .digest('hex');
          salt++;
        } while (emitted.has(fingerprint));
        salts.set(input, salt);
        emitted.add(fingerprint);
        return {
          description: m.message,
          check_name: m.ruleId,
          fingerprint,
          severity: m.severity === 2 ? 'major' : 'minor',
          location: {
            path: filePath,
            lines: { begin: m.line, end: m.endLine },
            positions: {
              begin: { line: m.line, column: m.column },
              end: { line: m.endLine, column: m.endColumn },
            },
          },
        };
      }),
    ) + '\n'
  );
}

function lineStarts(text: string): number[] {
  const starts = [0];
  for (let index = 0; index < text.length; index++) {
    const code = text.charCodeAt(index);
    if (code === 13 && text.charCodeAt(index + 1) === 10) index++;
    if (code === 10 || code === 13 || code === 0x2028 || code === 0x2029)
      starts.push(index + 1);
  }
  return starts;
}

/** The frozen formatter advances displayed lines at LF, but locations use all JS line breaks. */
function codeFrame(view: View, color: Colors): string {
  const m = view.message;
  const text = view.text.replace(/^\uFEFF/, '');
  const bytes = Buffer.from(text);
  const lines = lineStarts(text).map((offset) =>
    Buffer.byteLength(text.slice(0, offset)),
  );
  const start = Buffer.byteLength(text.slice(0, m.range[0]));
  const end = Buffer.byteLength(text.slice(0, m.range[1]));
  const first = Math.max(m.line - 2, 0);
  const last = Math.min(m.endLine, lines.length - 1);
  const boxEnd = last === lines.length - 1 ? bytes.length : lines[last + 1] - 1;
  const widths = new Array<number>(last - first + 1).fill(0);
  const ends = new Array<number>(widths.length).fill(0);
  let indent = Number.MAX_SAFE_INTEGER;
  let line = first;
  let hasIndent = false;
  let lastNonSpace = -1;
  let offset = lines[first];
  const isSpace = (character: string) =>
    /^[\t\n\v\f\r \u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]$/.test(
      character,
    );
  for (const character of bytes.subarray(offset, boxEnd).toString('utf8')) {
    const current = offset;
    offset += Buffer.byteLength(character);
    if (character === '\n') {
      if (line !== last) {
        hasIndent = false;
        ends[line - first] = Math.max(lastNonSpace - lines[line], 0);
        lastNonSpace = -1;
        line++;
      }
      continue;
    }
    if (!hasIndent && !isSpace(character)) {
      hasIndent = true;
      widths[line - first] = Math.max(current - lines[line], 0);
      indent = Math.min(indent, widths[line - first]);
    }
    if (hasIndent && !isSpace(character)) lastNonSpace = offset;
  }
  if (line === last)
    ends[line - first] = Math.max(lastNonSpace - lines[line], 0);
  if (indent === Number.MAX_SAFE_INTEGER) indent = 0;
  const severity = m.severity === 2 ? color.error : color.warn;
  const continuation = `    ${color.dim('│')}${' '.repeat(Buffer.byteLength(m.ruleId) + 1)}`;
  let output = ` ${color.rule(` ${m.ruleId} `)} — ${severity(`[${m.severity === 2 ? 'error' : 'warn'}] `)}${m.message.replaceAll('\n', `\n${continuation}`)}\n  ${color.dim('╭─┴──────────(')} ${color.file(`${view.filePath}:${m.line}:${m.column}`)} ${color.dim(')─────')}\n`;
  let highlighted = false;
  const lastNumber = String(last + 1);
  for (let current = first; current <= last; current++) {
    if (last - first >= 4 && first + 1 < current && current < last - 1) {
      output += `  ${color.dim('│ ')}${color.dim('.'.repeat(lastNumber.length))}${color.dim(' │')}\n`;
      current = last - 1;
      highlighted = start < lines[current] && end >= lines[current];
    }
    const lineStart = lines[current] + indent;
    const lineEnd = Math.max(lines[current] + ends[current - first], lineStart);
    let underlineStart = Math.max(
      lineStart,
      lines[current] + widths[current - first],
    );
    let underlineEnd = underlineStart;
    if (highlighted) underlineEnd = lineEnd;
    else if (
      lines[current] <= start &&
      (current === lines.length - 1 || start < lines[current + 1])
    ) {
      underlineStart = Math.min(Math.max(lineStart, start), lineEnd);
      underlineEnd = lineEnd;
      highlighted = true;
    }
    if (
      lines[current] <= end &&
      (current === lines.length - 1 || end < lines[current + 1])
    ) {
      underlineEnd = Math.min(Math.max(underlineStart, end), lineEnd);
      highlighted = false;
    }
    const number = String(current + 1);
    output += `  ${color.dim('│ ')}${current !== last && number.length < lastNumber.length ? ' ' : ''}${color.dim(number)}${color.dim(' │')}  `;
    const slice = (begin: number, finish: number) =>
      bytes.subarray(begin, finish).toString('utf8');
    output +=
      underlineStart !== underlineEnd
        ? slice(lineStart, underlineStart) +
          severity(slice(underlineStart, underlineEnd)) +
          slice(underlineEnd, lineEnd)
        : slice(lineStart, lineEnd);
    output += '\n';
  }
  return `${output}  ${color.dim('╰────────────────────────────────')}\n\n`;
}

function duration(startTime: number): string {
  const milliseconds = Math.max(0, Math.round(Date.now() - startTime));
  if (milliseconds < 1) return '<1ms';
  if (milliseconds < 1000) return `${milliseconds}ms`;
  const seconds = milliseconds / 1000;
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const tail = `${Math.round((seconds % 60) * 1000) / 1000}s`;
  return minutes < 60
    ? `${minutes}m${tail}`
    : `${Math.floor(minutes / 60)}h${minutes % 60}m${tail}`;
}

function plural(count: number, noun: string): string {
  return `${count} ${noun}${count === 1 ? '' : 's'}`;
}

function status(
  label: 'start' | 'success' | 'error',
  message: string,
  color: Colors,
  details = '',
): string {
  return `${color[label](label)}${' '.repeat(8 - label.length)}${message}${details ? ` ${color.dim(details)}` : ''}\n`;
}

/** argv excludes the executable pair. The caller owns process.exitCode and stream flushing. */
export async function runRustCLI(
  argv: readonly string[],
  options: RustCLIOptions = {},
): Promise<number> {
  const stdout = options.stdout ?? process.stdout;
  const stderr = options.stderr ?? process.stderr;
  const startTime = options.startTime ?? Date.now();
  const cwd = path.resolve(options.cwd ?? process.cwd());
  let parsed: ReturnType<typeof parseCLI>;
  let format: OutputFormat;
  let maxWarnings = -1;
  let overrides: ReturnType<typeof ruleOverrides>;
  try {
    parsed = parseCLI(argv);
    if (parsed.values.help) {
      stdout.write(HELP);
      return 0;
    }
    for (const name of [
      'init',
      'type-check',
      'type-check-only',
      'timing',
      'trace',
      'cpuprof',
    ] as const)
      if (parsed.values[name] !== undefined)
        throw new Error(
          `--${name} is not implemented by the Rust lint entry point`,
        );
    const requested = parsed.values.format ?? 'default';
    if (!isOutputFormat(requested))
      throw new Error(
        `invalid output format ${JSON.stringify(requested)} (expected default, jsonline, github, or gitlab)`,
      );
    format = requested;
    if (parsed.values['max-warnings'] !== undefined) {
      const value = parsed.values['max-warnings'];
      if (!/^[+-]?\d+$/.test(value) || !Number.isSafeInteger(Number(value)))
        throw new Error(
          `invalid value ${JSON.stringify(value)} for flag -max-warnings`,
        );
      maxWarnings = Number(value);
    }
    overrides = ruleOverrides(parsed.values.rule ?? []);
  } catch (error) {
    stderr.write(
      `error: ${error instanceof Error ? error.message : String(error)}\n`,
    );
    return 2;
  }
  const color = colors(colorEnabled(parsed.values, stdout));
  let started = false;
  try {
    const selected = parsed.values.config;
    if (selected) {
      const filename = path.resolve(cwd, selected);
      if (!fs.existsSync(filename)) {
        stderr.write(`Error: config file not found: ${filename}\n`);
        return 1;
      }
      if (!isJSConfigFile(filename)) {
        stderr.write(
          `Error: unsupported config file: ${filename}\nRslint configs must be JavaScript or TypeScript modules.${/\.jsonc?$/.test(filename) ? ' Run `rslint --init` to migrate a legacy JSON config.' : ''}\n`,
        );
        return 1;
      }
    }
    const config = await resolveRustConfig({
      cwd,
      overrideConfigFile: selected ?? null,
      overrideConfig: overrides,
    });
    const request = {
      ...config,
      ...(parsed.positionals.length ? { patterns: parsed.positionals } : {}),
    };
    if (format === 'default') {
      stdout.write(status('start', 'Linting...', color));
      started = true;
    }
    const originals = new Map<string, SourceSnapshot>();
    const result = lintRustProjectsWithSourceSnapshotsSync(
      {
        ...request,
        sourceOnlyCurrentDirectory: selected ? config.configDirectory : cwd,
        fix: parsed.values.fix ?? false,
        singleThreaded: parsed.values.singleThreaded ?? false,
      },
      (filename) => {
        const source = sourceSnapshot(filename);
        originals.set(filename, source);
        return source.text;
      },
    );
    if (
      result.files.length !== originals.size ||
      result.files.some((file) => !originals.has(file.filePath))
    )
      throw new Error('admitted files changed during lint generation');
    const views: View[] = result.files
      .flatMap((file) =>
        file.messages.map((message) => ({
          filePath: path.relative(cwd, file.filePath).split(path.sep).join('/'),
          message,
          text: file.output ?? originals.get(file.filePath)!.text,
        })),
      )
      .sort(
        (left, right) =>
          Buffer.compare(
            Buffer.from(left.filePath),
            Buffer.from(right.filePath),
          ) || left.message.range[0] - right.message.range[0],
      );
    const errors = views.filter((view) => view.message.severity === 2).length;
    const warnings = views.filter((view) => view.message.severity === 1).length;
    const warningLimit =
      errors === 0 && maxWarnings >= 0 && warnings > maxWarnings;
    const visible = parsed.values.quiet
      ? views.filter((view) => view.message.severity === 2)
      : views;
    // Render before committing: formatter failures must not leave changed sources.
    const rendered =
      format === 'default'
        ? `${visible.length ? '\n' : ''}${visible.map((view) => codeFrame(view, color)).join('')}`
        : machineOutput(format, visible);
    if (parsed.values.fix) commitFixes(result, originals);
    if (
      format !== 'default' &&
      parsed.positionals.length &&
      !result.files.length
    )
      return 0;
    stdout.write(rendered);
    if (format === 'default') {
      const { fixedIssues, threads } = result.summary;
      const fixes = fixedIssues
        ? ` after applying ${plural(fixedIssues, 'fix').replace(/fixs$/, 'fixes')}`
        : '';
      let message: string;
      if (errors) {
        const findings = [
          `${color.error(String(errors))} error${errors === 1 ? '' : 's'}`,
        ];
        if (warnings)
          findings.push(
            `${color.warn(String(warnings))} warning${warnings === 1 ? '' : 's'}`,
          );
        message = `Lint failed with ${findings.join(' and ')}${fixes} in ${duration(startTime)}`;
      } else if (warningLimit)
        message = `Lint failed${fixes} in ${duration(startTime)}: ${color.warn(String(warnings))} warning${warnings === 1 ? '' : 's'} exceeded the configured limit of ${maxWarnings}`;
      else if (!result.files.length && !warnings) {
        stdout.write(
          status(
            'success',
            `No files to lint in ${duration(startTime)}`,
            color,
          ),
        );
        return 0;
      } else
        message = `Lint passed${warnings ? ` with ${color.warn(String(warnings))} warning${warnings === 1 ? '' : 's'}` : ''}${fixes} in ${duration(startTime)}`;
      const ruleCount = new Set(
        result.files.flatMap((file) => file.appliedRules),
      ).size;
      stdout.write(
        status(
          errors || warningLimit ? 'error' : 'success',
          message,
          color,
          `(${plural(result.files.length, 'file')}, ${plural(ruleCount, 'rule')}, ${plural(threads, 'thread')})`,
        ),
      );
    } else if (warningLimit)
      stderr.write(
        `Rslint found too many warnings (maximum: ${maxWarnings}).\n`,
      );
    return errors || warningLimit ? 1 : 0;
  } catch (error) {
    const reason = error instanceof Error ? error.message : String(error);
    if (started)
      stdout.write(
        status(
          'error',
          `Linting failed in ${duration(startTime)}: ${reason}`,
          color,
        ),
      );
    stderr.write(`error: ${reason}\n`);
    return 1;
  }
}
