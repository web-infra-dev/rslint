// cspell:words nonegate noext
import fs from 'node:fs';
import path from 'node:path';
import picomatch from 'picomatch';
import type { LintMessage, LintResult } from '../api/rslint.js';
import { getNativeBinding } from '../native/binding.js';
import {
  admitRustFiles,
  type AdmissionOptions,
  type AdmittedProjectPolicy,
} from './admission.js';
import {
  isRecord,
  validateJson,
  validateKeys,
  validateRustResult,
  validateUnicode,
  type RustLintFileResult,
  type RustLintResult,
  type RustRuleConfig,
} from './engine.js';

export type RustProjectPolicy =
  | { kind: 'sourceOnly' }
  | { kind: 'service'; rootDirectory: string }
  | { kind: 'explicit'; configFiles: string[] };

export interface RustProjectsOptions extends AdmissionOptions {
  /** Source-only host cwd, independent of admission and configured projects. */
  sourceOnlyCurrentDirectory?: string;
  fix?: boolean;
  singleThreaded?: boolean;
  /** In-memory source overlays, including dependencies and project configs. */
  sources?: ReadonlyArray<{ path: string; text: string }>;
  /** One in-memory lintText target; it takes precedence over dependency overlays. */
  buffer?: { path: string; text: string };
}

export interface RustProjectFileResult extends RustLintFileResult {
  /** Actual capability of this file's selected Program, not filename inference. */
  hasTypeInformation: boolean;
}

export interface RustProjectsResult extends RustLintResult {
  files: RustProjectFileResult[];
}

interface ProjectInput {
  path: string;
  text?: string;
  policy: RustProjectPolicy;
  rules: RustRuleConfig[];
}

interface ProjectBinding {
  lintProjectsTsrsSync(request: string): string;
  ruleCapabilitiesTsrsSync(request: string): string;
}

function projectError(message: string): never {
  throw new Error(`experimental Rust project lint: ${message}`);
}

function bindingHasProjects(value: unknown): value is ProjectBinding {
  return (
    isRecord(value) &&
    typeof value.lintProjectsTsrsSync === 'function' &&
    typeof value.ruleCapabilitiesTsrsSync === 'function'
  );
}

function isFile(filename: string): boolean {
  try {
    return fs.statSync(filename).isFile();
  } catch (error) {
    if (isRecord(error) && error.code === 'ENOENT') return false;
    throw error;
  }
}

function canonicalPath(filename: string): string {
  try {
    return fs.realpathSync.native(filename);
  } catch (error) {
    if (!isRecord(error) || error.code !== 'ENOENT') throw error;
    const parent = path.dirname(filename);
    return parent === filename
      ? filename
      : path.join(canonicalPath(parent), path.basename(filename));
  }
}

/** The authored base is literal, even when a directory name contains glob syntax. */
function projectGlob(base: string, pattern: string): string[] {
  if (/[?*+@!]\(/.test(pattern))
    projectError(`project extglob is not supported: ${pattern}`);
  const normalized = path.posix.normalize(pattern.replaceAll('\\', '/'));
  const first = normalized.search(/[*?\[]/);
  const prefix = normalized.slice(0, first);
  const separator = prefix.lastIndexOf('/');
  const authoredRoot = separator < 0 ? '.' : prefix.slice(0, separator) || '/';
  const root = path.resolve(base, authoredRoot);
  const absolute = path.resolve(base, normalized);
  const relative = path.relative(root, absolute).split(path.sep).join('/');
  const match = picomatch(
    relative.replace(/[()|]/g, (part) => `[${part}]`),
    {
      dot: true,
      nonegate: true,
      noext: true,
      regex: false,
      literalBrackets: false,
      posix: true,
      strictBrackets: true,
    },
  );
  const found: string[] = [];
  function visit(directory: string, ancestors: ReadonlySet<string>): void {
    let physical: string;
    let children: fs.Dirent[];
    try {
      physical = fs.realpathSync.native(directory);
      if (ancestors.has(physical)) return;
      children = fs.readdirSync(directory, { withFileTypes: true });
    } catch (error) {
      if (
        isRecord(error) &&
        (error.code === 'ENOENT' || error.code === 'ENOTDIR')
      )
        return;
      throw error;
    }
    const next = new Set([...ancestors, physical]);
    for (const child of children) {
      const filename = path.join(directory, child.name);
      let directoryEntry = child.isDirectory();
      let fileEntry = child.isFile();
      if (child.isSymbolicLink()) {
        try {
          const stat = fs.statSync(filename);
          directoryEntry = stat.isDirectory();
          fileEntry = stat.isFile();
        } catch (error) {
          if (isRecord(error) && error.code === 'ENOENT') continue;
          throw error;
        }
      }
      if (directoryEntry) visit(filename, next);
      else if (
        fileEntry &&
        match(path.relative(root, filename).split(path.sep).join('/'))
      )
        found.push(filename);
    }
  }
  visit(root, new Set());
  // Go's sort.Strings compares UTF-8 bytes; locale sorting changes precedence.
  return found.sort((left, right) =>
    Buffer.compare(Buffer.from(left), Buffer.from(right)),
  );
}

function resolvePolicy(
  policy: AdmittedProjectPolicy,
  overlays: ReadonlyMap<string, string>,
): RustProjectPolicy {
  if (policy.serviceRootDirectory)
    return { kind: 'service', rootDirectory: policy.serviceRootDirectory };
  if (policy.projectDisabled || !policy.explicitProject)
    return { kind: 'sourceOnly' };
  const { patterns, baseDirectory } = policy.explicitProject;
  const base = policy.tsconfigRootDirOverride ?? baseDirectory;
  const selected = new Set<string>();
  for (const pattern of patterns) {
    validateUnicode(pattern, 'project pattern');
    if (/[*?\[]/.test(pattern)) {
      const matches = projectGlob(base, pattern);
      if (matches.length === 0)
        projectError(
          `glob pattern ${JSON.stringify(pattern)} matched no files`,
        );
      for (const match of matches) selected.add(match);
    } else {
      const filename = path.resolve(base, pattern);
      if (!overlays.has(filename) && !isFile(filename))
        projectError(`tsconfig file ${JSON.stringify(filename)} doesn't exist`);
      selected.add(filename);
    }
  }
  return { kind: 'explicit', configFiles: [...selected] };
}

/**
 * Evaluate configuration before calling this entry. Admission and per-file
 * project policies are computed from that configuration, never a benchmark
 * manifest. The addon runs in this process and fixes remain in memory.
 */
export function lintRustProjectsSync(
  options: RustProjectsOptions,
): RustProjectsResult {
  return executeRustProjects(options);
}

/**
 * Internal CLI boundary: capture disk text after validated admission, before
 * constructing the native generation. The reader receives only admitted path
 * strings; it cannot supply a replacement file/config/rule list. Public APIs
 * keep using lintRustProjectsSync and its normal overlay/buffer validation.
 */
export function lintRustProjectsWithSourceSnapshotsSync(
  options: Omit<RustProjectsOptions, 'sources' | 'buffer'>,
  readSource: (filename: string) => string,
): RustProjectsResult {
  if (typeof readSource !== 'function')
    projectError('source snapshot reader must be a function');
  return executeRustProjects(options, readSource);
}

function executeRustProjects(
  options: RustProjectsOptions,
  readSource?: (filename: string) => string,
): RustProjectsResult {
  if (!isRecord(options)) projectError('options must be an object');
  validateKeys(
    options,
    new Set([
      'cwd',
      'config',
      'configDirectory',
      'sourceOnlyCurrentDirectory',
      'configEntryDirectories',
      'configArrayBase',
      'patterns',
      'fix',
      'singleThreaded',
      'sources',
      'buffer',
    ]),
    'options',
  );
  for (const key of ['fix', 'singleThreaded'] as const)
    if (options[key] !== undefined && typeof options[key] !== 'boolean')
      projectError(`${key} must be a boolean`);
  if (
    readSource &&
    (options.sources !== undefined || options.buffer !== undefined)
  )
    projectError(
      'CLI source snapshots cannot be combined with source overlays or a buffer',
    );
  if (typeof options.cwd !== 'string' || !path.isAbsolute(options.cwd))
    projectError('cwd must be absolute');
  const cwd = path.resolve(options.cwd);
  validateUnicode(cwd, 'cwd');
  if (
    options.sourceOnlyCurrentDirectory !== undefined &&
    (typeof options.sourceOnlyCurrentDirectory !== 'string' ||
      !path.isAbsolute(options.sourceOnlyCurrentDirectory))
  )
    projectError('sourceOnlyCurrentDirectory must be absolute');
  const sourceOnlyCurrentDirectory = path.resolve(
    options.sourceOnlyCurrentDirectory ?? options.configDirectory ?? cwd,
  );
  validateUnicode(sourceOnlyCurrentDirectory, 'sourceOnlyCurrentDirectory');
  const overlays = new Map<string, string>();
  if (options.sources !== undefined) {
    if (!Array.isArray(options.sources))
      projectError('sources must be an array');
    for (const [index, source] of options.sources.entries()) {
      if (!isRecord(source))
        projectError(`sources[${index}] must be an object`);
      validateKeys(source, new Set(['path', 'text']), `sources[${index}]`);
      if (
        typeof source.path !== 'string' ||
        source.path.length === 0 ||
        typeof source.text !== 'string'
      )
        projectError(`sources[${index}] requires a path and text`);
      validateUnicode(source.path, `sources[${index}].path`);
      validateUnicode(source.text, `sources[${index}].text`);
      const filename = path.resolve(cwd, source.path);
      if (overlays.has(filename))
        projectError(`duplicate source overlay: ${filename}`);
      overlays.set(filename, source.text);
    }
  }
  let textFilePath: string | undefined;
  if (options.buffer !== undefined) {
    if (!isRecord(options.buffer)) projectError('buffer must be an object');
    validateKeys(options.buffer, new Set(['path', 'text']), 'buffer');
    if (
      typeof options.buffer.path !== 'string' ||
      options.buffer.path.length === 0 ||
      typeof options.buffer.text !== 'string'
    )
      projectError('buffer requires a path and text');
    if (options.patterns !== undefined)
      projectError('buffer cannot be combined with patterns');
    validateUnicode(options.buffer.path, 'buffer.path');
    validateUnicode(options.buffer.text, 'buffer.text');
    textFilePath = path.resolve(cwd, options.buffer.path);
    const canonical = canonicalPath(textFilePath);
    for (const filename of overlays.keys()) {
      if (canonicalPath(filename) === canonical) overlays.delete(filename);
    }
    overlays.set(textFilePath, options.buffer.text);
  }
  const admitted = admitRustFiles({ ...options, cwd, textFilePath });
  if (readSource) {
    for (const file of admitted) {
      const text = readSource(file.path);
      if (typeof text !== 'string')
        projectError('source snapshot reader must return source text');
      validateUnicode(text, 'source snapshot');
      overlays.set(file.path, text);
    }
  }
  const selectedPaths = new Set(admitted.map((file) => file.path));
  const policyCache = new Map<AdmittedProjectPolicy, RustProjectPolicy>();
  const ruleCache = new Map<object, RustRuleConfig[]>();
  const files: ProjectInput[] = admitted.map((file) => {
    const { config } = file;
    // These settings have no Rust consumer yet. Refuse them rather than
    // accepting a config whose semantics disappear at the language boundary.
    if (Object.keys(config.settings).length > 0)
      projectError('settings are not supported yet');
    for (const [name, value] of Object.entries(config.languageOptions)) {
      if (name === 'parserOptions') continue;
      if (
        name === 'globals' &&
        isRecord(value) &&
        Object.keys(value).length === 0
      )
        continue;
      projectError(`languageOptions.${name} is not supported yet`);
    }
    let policy = policyCache.get(config.projectPolicy);
    if (!policy) {
      policy = resolvePolicy(config.projectPolicy, overlays);
      policyCache.set(config.projectPolicy, policy);
    }
    let rules = ruleCache.get(config);
    if (!rules) {
      rules = config.rules.map((rule) => {
        validateUnicode(rule.name, 'rule name');
        validateJson(rule.options, `rule ${JSON.stringify(rule.name)} options`);
        return {
          name: rule.name,
          severity: rule.severity,
          options: rule.options,
        };
      });
      ruleCache.set(config, rules);
    }
    validateUnicode(file.path, 'admitted path');
    return {
      path: file.path,
      policy,
      rules,
      ...(overlays.has(file.path) ? { text: overlays.get(file.path) } : {}),
    };
  });
  const binding: unknown = getNativeBinding();
  if (!bindingHasProjects(binding))
    projectError(
      'native addon has no configured lint exports; rebuild the Rust addon',
    );
  const names = [
    ...new Set(files.flatMap((file) => file.rules.map((rule) => rule.name))),
  ];
  const metadata: unknown = JSON.parse(
    binding.ruleCapabilitiesTsrsSync(JSON.stringify(names)),
  );
  if (
    !isRecord(metadata) ||
    Object.keys(metadata).length !== names.length ||
    names.some((name) => typeof metadata[name] !== 'boolean')
  )
    projectError('native addon returned invalid rule capabilities');
  const request = {
    schemaVersion: 1,
    cwd,
    ...(sourceOnlyCurrentDirectory === cwd
      ? {}
      : { sourceOnlyCurrentDirectory }),
    files,
    sources: [...overlays]
      .filter(([filename]) => !selectedPaths.has(filename))
      .map(([filename, text]) => ({ path: filename, text })),
    ...(options.fix === undefined ? {} : { fix: options.fix }),
    ...(options.singleThreaded === undefined
      ? {}
      : { singleThreaded: options.singleThreaded }),
  };
  const response: unknown = JSON.parse(
    binding.lintProjectsTsrsSync(JSON.stringify(request)),
  );
  if (
    !isRecord(response) ||
    !Array.isArray(response.files) ||
    response.files.length !== files.length
  )
    projectError('native addon returned incomplete project lint results');
  const capabilities = response.files.map((file: unknown, index: number) => {
    if (!isRecord(file) || typeof file.hasTypeInformation !== 'boolean')
      projectError('native addon omitted project capability');
    const policy = files[index].policy;
    if (
      file.hasTypeInformation &&
      (policy.kind === 'sourceOnly' ||
        (policy.kind === 'explicit' && policy.configFiles.length === 0))
    )
      projectError('native addon granted types to a source-only file');
    return file.hasTypeInformation;
  });
  validateRustResult(response, {
    files: files.map((file, index) => ({
      path: file.path,
      rules: file.rules.filter(
        (rule) => capabilities[index] || metadata[rule.name] === false,
      ),
    })),
    fix: options.fix,
  });
  return {
    ...response,
    files: response.files.map((file, index) => ({
      ...file,
      hasTypeInformation: capabilities[index],
    })),
  };
}

/** Project the checked Rust result to the existing public API's per-file shape. */
export function toPublicLintResults(result: RustProjectsResult): LintResult[] {
  return result.files.map((file) => {
    const messages: LintMessage[] = file.messages.map(
      ({ range: _range, ...message }) => message,
    );
    return {
      filePath: file.filePath,
      messages,
      errorCount: messages.filter((message) => message.severity === 2).length,
      warningCount: messages.filter((message) => message.severity === 1).length,
      fixableErrorCount: messages.filter(
        (message) => message.severity === 2 && message.fix,
      ).length,
      fixableWarningCount: messages.filter(
        (message) => message.severity === 1 && message.fix,
      ).length,
      ...(file.output === undefined ? {} : { output: file.output }),
    };
  });
}
