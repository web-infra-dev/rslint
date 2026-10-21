import path from 'node:path';
import {
  loadConfigFile,
  normalizeConfig,
} from '../config/config-file-loader.js';
import type { RslintConfig } from '../config/define-config.js';
import { getNativeBinding } from '../native/binding.js';

export type RustJsonValue =
  | null
  | boolean
  | number
  | string
  | RustJsonValue[]
  | { [key: string]: RustJsonValue };

/** Literal admitted files. There is no glob expansion or target discovery. */
export interface RustInputFile {
  path: string;
  /** An in-memory source overlay; omitted files are read by the Rust engine. */
  text?: string;
}

interface RustLintInputs {
  cwd?: string;
  files: readonly RustInputFile[];
  /** One explicit tsconfig; paths resolve from cwd. */
  project?: string;
  /** tsrs compiler option values for a source-only program, without project. */
  compilerOptions?: { [key: string]: RustJsonValue };
  /** Fix execution has not been migrated into this entry point yet. */
  fix?: false;
}

/**
 * Experimental direct Rust request, separate from the public Rslint API.
 * Only uniform config entries containing name/rules are supported. Scoped
 * configs, plugins, ignores and discovery reject instead of losing work.
 */
export interface RustLintOptions extends RustLintInputs {
  config: RslintConfig;
}

/** Select one config module explicitly; no ancestor search is performed. */
export interface RustLintConfigFileOptions extends RustLintInputs {
  configFile: string;
}

export interface RustLintDiagnostic {
  ruleId: string;
  messageId: string;
  message: string;
  severity: 1 | 2;
  line: number;
  column: number;
  endLine: number;
  endColumn: number;
  /** BOM-free JavaScript UTF-16 offsets, with an exclusive end. */
  range: [number, number];
  /** Proposed edit; fix application remains unsupported by this entry. */
  fix?: { range: [number, number]; text: string };
  suggestions?: Array<{
    messageId: string;
    desc: string;
    data?: Record<string, string>;
    fix: { range: [number, number]; text: string };
  }>;
}

export interface RustLintFileResult {
  filePath: string;
  messages: RustLintDiagnostic[];
  /** Includes all enabled rules even when the file has no diagnostics. */
  appliedRules: string[];
}

export interface RustLintResult {
  schemaVersion: 1;
  backend: 'tsrs-rust';
  compiler: { revision: string };
  files: RustLintFileResult[];
}

interface RustRuleConfig {
  name: string;
  severity: 1 | 2;
  options: RustJsonValue[];
}

interface RustLintRequest {
  schemaVersion: 1;
  cwd: string;
  files: RustInputFile[];
  rules: RustRuleConfig[];
  project?: string;
  compilerOptions?: { [key: string]: RustJsonValue };
}

interface RustNativeBinding {
  lintTsrsSync(request: string): string;
}

const INPUT_KEYS = new Set([
  'cwd',
  'files',
  'project',
  'compilerOptions',
  'fix',
]);

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function unsupported(message: string): never {
  throw new Error(`experimental Rust lint: ${message}`);
}

function validateUnicode(value: string, label: string): void {
  for (let index = 0; index < value.length; index++) {
    const code = value.charCodeAt(index);
    if (code >= 0xd800 && code <= 0xdbff) {
      const next = value.charCodeAt(++index);
      if (next >= 0xdc00 && next <= 0xdfff) continue;
      unsupported(
        `${label} contains an unpaired surrogate; this source encoding is not supported yet`,
      );
    }
    if (code >= 0xdc00 && code <= 0xdfff) {
      unsupported(
        `${label} contains an unpaired surrogate; this source encoding is not supported yet`,
      );
    }
  }
}

function validateKeys(
  value: Record<string, unknown>,
  allowed: ReadonlySet<string>,
  label: string,
): void {
  const prototype: unknown = Object.getPrototypeOf(value);
  if (prototype !== Object.prototype && prototype !== null) {
    unsupported(`${label} must be a plain object`);
  }
  for (const key of Reflect.ownKeys(value)) {
    if (typeof key !== 'string' || !allowed.has(key)) {
      unsupported(`${label} field ${String(key)} is not supported yet`);
    }
  }
}

/** JSON.stringify must not silently drop or coerce any configured value. */
function validateJson(
  value: unknown,
  label: string,
  active = new Set<object>(),
): asserts value is RustJsonValue {
  if (typeof value === 'string') {
    validateUnicode(value, label);
    return;
  }
  if (
    value === null ||
    typeof value === 'boolean' ||
    (typeof value === 'number' && Number.isFinite(value))
  ) {
    return;
  }
  if (typeof value !== 'object' || value === null) {
    unsupported(`${label} must contain only JSON values`);
  }
  if (active.has(value)) unsupported(`${label} must not be circular`);
  active.add(value);
  if (Array.isArray(value)) {
    for (let index = 0; index < value.length; index++) {
      validateJson(value[index], `${label}[${index}]`, active);
    }
  } else {
    const prototype: unknown = Object.getPrototypeOf(value);
    if (prototype !== Object.prototype && prototype !== null) {
      unsupported(`${label} must contain only plain JSON objects`);
    }
    for (const key of Reflect.ownKeys(value)) {
      if (typeof key !== 'string') {
        unsupported(`${label} must not contain symbol keys`);
      }
      validateUnicode(key, `${label} key`);
      validateJson(Reflect.get(value, key), `${label}.${key}`, active);
    }
  }
  active.delete(value);
}

function compileRules(config: unknown): RustRuleConfig[] {
  // Validate authored keys first: normalizeConfig intentionally strips some
  // unsupported keys, whereas this staged engine must reject all of them.
  if (Array.isArray(config)) {
    for (const [index, entry] of config.flat(1).entries()) {
      if (isRecord(entry)) {
        validateKeys(entry, new Set(['name', 'rules']), `config[${index}]`);
      }
    }
  }
  const merged = new Map<
    string,
    { severity: 0 | 1 | 2; options: RustJsonValue[] }
  >();
  for (const entry of normalizeConfig(config)) {
    if (!isRecord(entry.rules)) continue;
    for (const [name, setting] of Object.entries(entry.rules)) {
      validateUnicode(name, 'rule name');
      const parts: unknown[] = Array.isArray(setting) ? setting : [setting];
      const level = parts[0];
      const severity =
        level === 'off' || level === 0
          ? 0
          : level === 'warn' || level === 1
            ? 1
            : 2;
      // Flat-config severity-only overrides retain earlier rule options.
      const options =
        parts.length > 1 ? parts.slice(1) : (merged.get(name)?.options ?? []);
      validateJson(options, `rule ${JSON.stringify(name)} options`);
      merged.set(name, { severity, options });
    }
  }
  return [...merged].flatMap(([name, rule]) =>
    rule.severity === 0
      ? []
      : [{ name, severity: rule.severity, options: rule.options }],
  );
}

function buildRequest(options: RustLintOptions): RustLintRequest {
  if (!isRecord(options)) unsupported('options must be an object');
  validateKeys(options, new Set([...INPUT_KEYS, 'config']), 'options');
  if (options.fix !== undefined && options.fix !== false) {
    unsupported('fix is not supported yet');
  }
  if (options.cwd !== undefined && typeof options.cwd !== 'string') {
    unsupported('cwd must be a string');
  }
  const cwd = path.resolve(options.cwd ?? process.cwd());
  validateUnicode(cwd, 'cwd');
  if (!Array.isArray(options.files) || options.files.length === 0) {
    unsupported('files must explicitly include every admitted input');
  }
  const files = options.files.map((file, index) => {
    if (!isRecord(file)) unsupported(`files[${index}] must be an object`);
    validateKeys(file, new Set(['path', 'text']), `files[${index}]`);
    if (typeof file.path !== 'string' || file.path.length === 0) {
      unsupported(`files[${index}].path must be a non-empty literal path`);
    }
    if (file.text !== undefined && typeof file.text !== 'string') {
      unsupported(`files[${index}].text must be a string`);
    }
    validateUnicode(file.path, `files[${index}].path`);
    if (file.text !== undefined)
      validateUnicode(file.text, `files[${index}].text`);
    return {
      path: path.resolve(cwd, file.path),
      ...(file.text === undefined ? {} : { text: file.text }),
    };
  });
  if (
    options.project !== undefined &&
    (typeof options.project !== 'string' || options.project.length === 0)
  ) {
    unsupported('project must be one explicit tsconfig path');
  }
  if (options.project !== undefined && options.compilerOptions !== undefined) {
    unsupported('compilerOptions cannot override an explicit project');
  }
  if (options.project !== undefined)
    validateUnicode(options.project, 'project');
  if (options.compilerOptions !== undefined) {
    if (!isRecord(options.compilerOptions)) {
      unsupported('compilerOptions must be an object');
    }
    validateJson(options.compilerOptions, 'compilerOptions');
  }
  return {
    schemaVersion: 1,
    cwd,
    files,
    rules: compileRules(options.config),
    ...(options.project === undefined
      ? {}
      : { project: path.resolve(cwd, options.project) }),
    ...(options.compilerOptions === undefined
      ? {}
      : { compilerOptions: options.compilerOptions }),
  };
}

function integer(value: unknown, minimum: number): value is number {
  return (
    typeof value === 'number' && Number.isSafeInteger(value) && value >= minimum
  );
}

function isFix(value: unknown): boolean {
  return (
    isRecord(value) &&
    typeof value.text === 'string' &&
    Array.isArray(value.range) &&
    value.range.length === 2 &&
    integer(value.range[0], -1) &&
    // -1 removes the BOM. Merging that edit with a later edit can extend
    // the end beyond zero, but no replacement may end before the source.
    integer(value.range[1], Math.max(0, value.range[0]))
  );
}

function isSuggestion(value: unknown): boolean {
  return (
    isRecord(value) &&
    typeof value.messageId === 'string' &&
    typeof value.desc === 'string' &&
    isFix(value.fix) &&
    (!('data' in value) ||
      (isRecord(value.data) &&
        Object.values(value.data).every((entry) => typeof entry === 'string')))
  );
}

function validateResult(
  value: unknown,
  request: RustLintRequest,
): asserts value is RustLintResult {
  const fail = (): never =>
    unsupported('native addon returned an invalid or incomplete lint result');
  if (
    !isRecord(value) ||
    value.schemaVersion !== 1 ||
    value.backend !== 'tsrs-rust' ||
    !isRecord(value.compiler) ||
    typeof value.compiler.revision !== 'string' ||
    !/^[a-f\d]{40}$/.test(value.compiler.revision) ||
    !Array.isArray(value.files) ||
    value.files.length !== request.files.length
  ) {
    return fail();
  }
  const enabled = new Map(
    request.rules.map((rule) => [rule.name, rule.severity]),
  );
  for (const [index, file] of value.files.entries()) {
    if (
      !isRecord(file) ||
      typeof file.filePath !== 'string' ||
      path.normalize(file.filePath) !==
        path.normalize(request.files[index].path) ||
      !Array.isArray(file.messages) ||
      !Array.isArray(file.appliedRules) ||
      file.appliedRules.length !== request.rules.length ||
      file.appliedRules.some(
        (name, ruleIndex) => name !== request.rules[ruleIndex].name,
      )
    ) {
      return fail();
    }
    for (const message of file.messages) {
      if (
        !isRecord(message) ||
        typeof message.ruleId !== 'string' ||
        !enabled.has(message.ruleId) ||
        message.severity !== enabled.get(message.ruleId) ||
        typeof message.messageId !== 'string' ||
        typeof message.message !== 'string' ||
        !integer(message.line, 1) ||
        !integer(message.column, 1) ||
        !integer(message.endLine, message.line) ||
        !integer(message.endColumn, 1) ||
        !Array.isArray(message.range) ||
        message.range.length !== 2 ||
        !integer(message.range[0], 0) ||
        !integer(message.range[1], message.range[0]) ||
        ('fix' in message && !isFix(message.fix)) ||
        ('suggestions' in message &&
          (!Array.isArray(message.suggestions) ||
            !message.suggestions.every(isSuggestion)))
      ) {
        return fail();
      }
    }
  }
}

function isRustBinding(value: unknown): value is RustNativeBinding {
  return isRecord(value) && typeof value.lintTsrsSync === 'function';
}

/**
 * Execute Rust linting directly through the platform addon. This synchronous
 * experimental entry blocks its calling thread; use a worker for interactive
 * consumers. It does not invoke the existing public CLI or another process.
 */
export function lintRustFilesSync(options: RustLintOptions): RustLintResult {
  const request = buildRequest(options);
  const binding: unknown = getNativeBinding();
  if (!isRustBinding(binding)) {
    unsupported(
      'native addon has no lintTsrsSync export; rebuild the Rust addon',
    );
  }
  const response: unknown = JSON.parse(
    binding.lintTsrsSync(JSON.stringify(request)),
  );
  validateResult(response, request);
  return response;
}

/**
 * Evaluate one explicitly selected config module, then call the same direct
 * synchronous engine. The Promise only covers config loading; lint computation
 * still blocks the calling thread. No automatic config or file discovery runs.
 */
export async function lintRustFiles(
  options: RustLintOptions | RustLintConfigFileOptions,
): Promise<RustLintResult> {
  if (!isRecord(options)) unsupported('options must be an object');
  if ('config' in options) return lintRustFilesSync(options as RustLintOptions);
  validateKeys(options, new Set([...INPUT_KEYS, 'configFile']), 'options');
  if (
    typeof options.configFile !== 'string' ||
    options.configFile.length === 0
  ) {
    unsupported('select config or one explicit configFile');
  }
  if (options.cwd !== undefined && typeof options.cwd !== 'string') {
    unsupported('cwd must be a string');
  }
  const configFile = path.resolve(
    options.cwd ?? process.cwd(),
    options.configFile,
  );
  validateUnicode(configFile, 'configFile');
  if (!/\.(?:[cm]?[jt]s)$/.test(configFile)) {
    unsupported('configFile must be a JavaScript or TypeScript module');
  }
  const { configFile: _configFile, ...inputs } = options;
  const config: unknown = await loadConfigFile(configFile);
  // buildRequest owns the same authored-config validation for both entry points.
  return lintRustFilesSync({ ...inputs, config } as RustLintOptions);
}
