// cspell:words nonegate noext nobrace
import fs from 'node:fs';
import path from 'node:path';
import picomatch from 'picomatch';
import { normalizeConfig } from '../config/config-file-loader.js';

const EXTENSIONS = new Set([
  '.js',
  '.mjs',
  '.cjs',
  '.jsx',
  '.ts',
  '.tsx',
  '.mts',
  '.cts',
]);
const EXCLUDED_DIRECTORIES = new Set(['node_modules', '.git']);
const CONFIG_KEYS = new Set([
  'name',
  'basePath',
  'files',
  'ignores',
  'rules',
  'languageOptions',
  'plugins',
  'settings',
]);
const PLUGIN_ALIASES: Record<string, string> = {
  'eslint-plugin-import': 'import',
  'eslint-plugin-jest': 'jest',
  'eslint-plugin-jsx-a11y': 'jsx-a11y',
  'eslint-plugin-node': 'node',
  'eslint-plugin-promise': 'promise',
  'eslint-plugin-react-hooks': 'react-hooks',
  'eslint-plugin-unicorn': 'unicorn',
};

export interface AdmittedRule {
  name: string;
  severity: 1 | 2;
  options: unknown[];
}

/** The planner must resolve this policy before deciding which typed rules run. */
export interface AdmittedProjectPolicy {
  serviceRootDirectory?: string;
  explicitProject?: { patterns: string[]; baseDirectory: string };
  defaultProjectDisabled: boolean;
  projectDisabled: boolean;
  tsconfigRootDirOverride?: string;
}

/** Shared within one admission result. Consumers must treat this snapshot as immutable. */
export interface AdmittedConfig {
  /** Includes typed rules. Lack of a checker must not silently suppress these here. */
  rules: AdmittedRule[];
  plugins: string[];
  settings: Record<string, unknown>;
  languageOptions: Record<string, unknown>;
  projectPolicy: AdmittedProjectPolicy;
  matchedEntries: number[];
}

export interface AdmittedFile {
  /** Lexical identity used by the caller and reported diagnostics. */
  path: string;
  canonicalPath: string;
  config: AdmittedConfig;
}

export interface AdmissionOptions {
  cwd: string;
  /** Already evaluated flat config, including imported presets. */
  config: unknown;
  /** Module owner; different from cwd for --config outside the working tree. */
  configDirectory?: string;
  /** Per-entry fallback owners, used when a module and inline config are combined. */
  configEntryDirectories?: readonly string[];
  /** Root used to resolve authored basePath. Explicit --config uses cwd. */
  configArrayBase?: string;
  /** Literal file/directory or cwd-relative glob targets, as in lintFiles. */
  patterns?: string[];
  /** One in-memory lintText target; dependency overlays are never targets. */
  textFilePath?: string;
}

interface Pattern {
  base: string;
  negated: boolean;
  match: (relative: string) => boolean;
  directoryBlock: boolean;
  directoryOnly: boolean;
}

interface Entry {
  index: number;
  base: string;
  scoped: boolean;
  global: boolean;
  selectors?: Array<Array<(relative: string) => boolean>>;
  ignores: Pattern[];
  value: Record<string, unknown>;
}

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function slash(value: string): string {
  return path.sep === '/' ? value : value.split(path.sep).join('/');
}

function contained(relative: string): boolean {
  return (
    relative !== '..' &&
    !relative.startsWith(`..${path.sep}`) &&
    !path.isAbsolute(relative)
  );
}

/** Path coordinates are shared only within one synchronous admission snapshot. */
class AdmissionPaths {
  private readonly coordinates = new Map<string, Map<string, string>>();
  private readonly physical = new Map<string, string>();

  relative(base: string, filename: string): string {
    let files = this.coordinates.get(base);
    if (!files) {
      files = new Map();
      this.coordinates.set(base, files);
    }
    const cached = files.get(filename);
    if (cached !== undefined) return cached;
    const relative = path.relative(base, filename);
    files.set(filename, relative);
    return relative;
  }

  within(base: string, filename: string): boolean {
    return contained(this.relative(base, filename));
  }

  realpath(filename: string): string {
    const cached = this.physical.get(filename);
    if (cached !== undefined) return cached;
    try {
      const physical = fs.realpathSync.native(filename);
      this.physical.set(filename, physical);
      return physical;
    } catch (error) {
      if (record(error) && error.code === 'ENOENT') {
        const parent = path.dirname(filename);
        // Do not memoize missing paths or errors. Only successful host answers
        // belong to this invocation; a later invocation always starts fresh.
        return parent === filename
          ? filename
          : path.join(this.realpath(parent), path.basename(filename));
      }
      throw error;
    }
  }

  /** Keep a leaf symlink's authored name; only directory aliases change coordinates. */
  relativeTo(base: string, filename: string): string {
    const relative = this.relative(base, filename);
    if (contained(relative)) return slash(relative);
    const physicalBase = this.realpath(base);
    const physicalParent = this.realpath(path.dirname(filename));
    const projected = path.join(physicalParent, path.basename(filename));
    const physicalRelative = this.relative(physicalBase, projected);
    return slash(contained(physicalRelative) ? physicalRelative : relative);
  }
}

function normalizeGlob(raw: string): string {
  const value = path.posix.normalize(raw.replaceAll('\\', '/'));
  return value === '.' ? '' : value.replace(/^\.\//, '');
}

function matcher(raw: string): (relative: string) => boolean {
  // Go's authored-glob engine does not implement extended glob syntax. Reject these rather
  // than quietly apply picomatch's more permissive language.
  if (/[?*+@!]\(/.test(raw))
    throw new Error(`Rust admission does not support extglob: ${raw}`);
  const glob = normalizeGlob(raw).replace(
    /[()|]/g,
    (character) => `[${character}]`,
  );
  const match = picomatch(glob, {
    dot: true,
    nonegate: true,
    noext: true,
    regex: false,
    literalBrackets: false,
    posix: true,
  });
  return (relative) => match(relative);
}

function selector(raw: string): (relative: string) => boolean {
  let negated = false;
  while (raw.startsWith('!')) {
    negated = !negated;
    raw = raw.slice(1);
  }
  const match = matcher(raw);
  return (relative) => negated !== match(relative);
}

function authoredPattern(raw: string, base: string): Pattern {
  const negated = raw.startsWith('!');
  const body = negated ? raw.slice(1) : raw;
  return {
    base,
    negated,
    match: matcher(body),
    directoryOnly: false,
    directoryBlock:
      body !== '' &&
      !body.endsWith('/**/*') &&
      (!body.endsWith('/*') || body.endsWith('/**')),
  };
}

function assertJson(
  value: unknown,
  label: string,
  seen = new Set<object>(),
): void {
  if (value === null || typeof value === 'string' || typeof value === 'boolean')
    return;
  if (typeof value === 'number' && Number.isFinite(value)) return;
  if (typeof value !== 'object' || seen.has(value))
    throw new Error(`${label} must contain JSON values`);
  seen.add(value);
  for (const child of Object.values(value)) assertJson(child, label, seen);
  seen.delete(value);
}

function compileEntries(options: AdmissionOptions): Entry[] {
  const normalized = normalizeConfig(options.config);
  const raw = Array.isArray(options.config) ? options.config.flat() : [];
  const origin = path.resolve(options.configDirectory ?? options.cwd);
  const arrayBase = path.resolve(options.configArrayBase ?? options.cwd);
  const entryDirectories = options.configEntryDirectories;
  if (
    entryDirectories !== undefined &&
    (!Array.isArray(entryDirectories) ||
      entryDirectories.length !== normalized.length ||
      !entryDirectories.every(
        (directory) =>
          typeof directory === 'string' && path.isAbsolute(directory),
      ))
  )
    throw new Error(
      'configEntryDirectories must match the normalized config with absolute directories',
    );
  return normalized.map((value, index) => {
    const authored: unknown = raw[index];
    if (!record(authored)) throw new Error(`Invalid config entry ${index}`);
    for (const key of Object.keys(authored)) {
      if (!CONFIG_KEYS.has(key))
        throw new Error(`Rust admission does not support config key: ${key}`);
    }
    if ('eslintPlugins' in value)
      throw new Error(
        'Rust admission does not yet support JavaScript plugin objects',
      );
    const base =
      typeof value.basePath === 'string'
        ? path.resolve(arrayBase, value.basePath)
        : (entryDirectories?.[index] ?? origin);
    const ignores = Array.isArray(value.ignores) ? value.ignores : [];
    const global =
      ignores.length > 0 &&
      Object.keys(authored).every((key) =>
        ['name', 'basePath', 'ignores'].includes(key),
      );
    const selectors = Array.isArray(value.files)
      ? value.files.map((item: unknown) => {
          if (typeof item === 'string') return [selector(item)];
          if (
            !Array.isArray(item) ||
            !item.every((part) => typeof part === 'string')
          )
            throw new Error('Invalid files selector');
          return item.map((part: string) => selector(part));
        })
      : undefined;
    if (record(value.languageOptions))
      validateLanguageOptions(value.languageOptions);
    for (const key of ['settings', 'rules']) {
      if (value[key] !== undefined) assertJson(value[key], key);
    }
    return {
      index,
      base,
      scoped: 'basePath' in authored,
      global,
      selectors,
      ignores: ignores.map((item: unknown) => {
        if (typeof item !== 'string') throw new Error('Invalid ignore pattern');
        return authoredPattern(item, base);
      }),
      value,
    };
  });
}

function validateLanguageOptions(value: Record<string, unknown>): void {
  for (const key of Object.keys(value)) {
    if (
      !['globals', 'sourceType', 'ecmaVersion', 'parserOptions'].includes(key)
    )
      throw new Error(`Unsupported languageOptions.${key}`);
  }
  if (
    value.sourceType !== undefined &&
    !['script', 'module', 'commonjs'].includes(String(value.sourceType))
  )
    throw new Error('Invalid languageOptions.sourceType');
  if (
    value.ecmaVersion !== undefined &&
    value.ecmaVersion !== 'latest' &&
    !(
      typeof value.ecmaVersion === 'number' &&
      Number.isInteger(value.ecmaVersion) &&
      (value.ecmaVersion === 3 ||
        value.ecmaVersion === 5 ||
        value.ecmaVersion >= 6)
    )
  )
    throw new Error('Invalid languageOptions.ecmaVersion');
  if (value.parserOptions !== undefined) {
    if (!record(value.parserOptions))
      throw new Error('languageOptions.parserOptions must be an object');
    for (const key of Object.keys(value.parserOptions)) {
      if (!['project', 'projectService', 'tsconfigRootDir'].includes(key))
        throw new Error(`Unsupported parserOptions.${key}`);
    }
    assertJson(value.parserOptions, 'parserOptions');
  }
}

function applyPatterns(
  paths: AdmissionPaths,
  patterns: Pattern[],
  filename: string,
  directory: boolean,
  initial = false,
): boolean {
  let ignored = initial;
  for (const pattern of patterns) {
    if (pattern.directoryOnly && !directory) continue;
    const relative = paths.relativeTo(pattern.base, filename);
    if (relative === '..' || relative.startsWith('../')) continue;
    if (pattern.match(relative) || (directory && pattern.match(`${relative}/`)))
      ignored = !pattern.negated;
  }
  return ignored;
}

function absoluteDirectoryBlock(
  paths: AdmissionPaths,
  patterns: Pattern[],
  directory: string,
): boolean {
  // Authored absolute directory blocks are deliberately not cancelled by
  // negations, unlike ordinary sequential file ignores (baseline policy).
  return patterns.some(
    (pattern) =>
      !pattern.negated &&
      pattern.directoryBlock &&
      applyPatterns(paths, [pattern], directory, true),
  );
}

function quoteGitPattern(raw: string): string {
  let quoted = '';
  for (let index = 0; index < raw.length; index++) {
    const ch = raw[index];
    if (ch === '\\' && index + 1 < raw.length) {
      const literal = raw[++index];
      quoted +=
        literal === '*'
          ? '[*]'
          : literal === '?'
            ? '[?]'
            : literal === '['
              ? '[[]'
              : literal === '{'
                ? '[{]'
                : literal === '}'
                  ? '[}]'
                  : literal;
    } else if ('{}()|'.includes(ch)) quoted += `[${ch}]`;
    else quoted += ch;
  }
  return quoted;
}

function gitPatterns(directory: string): Pattern[] {
  const filename = path.join(directory, '.gitignore');
  let text: string;
  try {
    text = fs.readFileSync(filename, 'utf8');
  } catch (error) {
    if (record(error) && error.code === 'ENOENT') return [];
    throw error;
  }
  const result: Pattern[] = [];
  for (let line of text.split('\n')) {
    line = line.replace(/\r$/, '');
    while (/[ \t]$/.test(line)) {
      const slashCount = /\\*$/.exec(line.slice(0, -1))?.[0].length ?? 0;
      if (slashCount % 2) break;
      line = line.slice(0, -1);
    }
    if (!line || line.startsWith('#')) continue;
    const negated = line.startsWith('!');
    if (negated) line = line.slice(1);
    if (line.split('/').includes('..')) continue;
    const directoryOnly = line.endsWith('/');
    if (directoryOnly) line = line.slice(0, -1);
    const rooted = line.includes('/');
    if (line.startsWith('/')) line = line.slice(1);
    if (!line) continue;
    const glob = `${rooted ? '' : '**/'}${quoteGitPattern(line)}${line.endsWith('/**') ? '/*' : ''}`;
    // Git has no extglob/brace expansion; picomatch can evaluate its quoted
    // node pattern directly without normalizing away escape characters.
    const match = picomatch(glob, {
      dot: true,
      nonegate: true,
      noext: true,
      nobrace: true,
      regex: false,
      literalBrackets: false,
      posix: true,
    });
    result.push({
      base: directory,
      negated,
      directoryOnly,
      directoryBlock: false,
      match,
    });
  }
  return result;
}

function mergeObjects(
  base: Record<string, unknown>,
  next: Record<string, unknown>,
): Record<string, unknown> {
  const result = { ...base };
  for (const [key, value] of Object.entries(next)) {
    const previous = result[key];
    Object.defineProperty(result, key, {
      value:
        record(previous) && record(value)
          ? mergeObjects(previous, value)
          : structuredClone(value),
      enumerable: true,
      configurable: true,
      writable: true,
    });
  }
  return result;
}

function severity(value: unknown): 0 | 1 | 2 {
  if (value === 0 || value === 'off') return 0;
  if (value === 1 || value === 'warn') return 1;
  if (value === 2 || value === 'error') return 2;
  throw new Error('Invalid rule severity');
}

function mergeConfig(
  paths: AdmissionPaths,
  entries: Entry[],
  filename: string,
  defaultRoot: string,
  cache: Map<string, AdmittedConfig>,
): AdmittedConfig {
  let matching = entries.filter((entry) => {
    if (entry.global) return false;
    const relative = paths.relativeTo(entry.base, filename);
    return (
      !(entry.scoped && (relative === '..' || relative.startsWith('../'))) &&
      (!entry.selectors ||
        entry.selectors.some((group) =>
          group.every((match) => match(relative)),
        )) &&
      !applyPatterns(paths, entry.ignores, filename, false)
    );
  });
  if (
    !EXTENSIONS.has(path.extname(filename)) &&
    !matching.some((entry) => entry.selectors !== undefined)
  )
    matching = [];
  const key = matching.map((entry) => entry.index).join(',');
  const cached = cache.get(key);
  if (cached) return cached;
  const rules = new Map<string, { severity: 0 | 1 | 2; options: unknown[] }>();
  const plugins = new Set<string>();
  let settings: Record<string, unknown> = {};
  let languageOptions: Record<string, unknown> = {};
  let projectOrigin: string | undefined;
  const matchedEntries: number[] = [];
  for (const entry of matching) {
    matchedEntries.push(entry.index);
    if (record(entry.value.rules)) {
      for (const [name, raw] of Object.entries(entry.value.rules)) {
        const values = Array.isArray(raw) ? raw : [raw];
        rules.set(name, {
          severity: severity(values[0]),
          options:
            values.length > 1
              ? structuredClone(values.slice(1))
              : (rules.get(name)?.options ?? []),
        });
      }
    }
    if (Array.isArray(entry.value.plugins)) {
      for (const plugin of entry.value.plugins) {
        if (typeof plugin !== 'string')
          throw new Error('Invalid native plugin');
        plugins.add(PLUGIN_ALIASES[plugin] ?? plugin);
      }
    }
    if (record(entry.value.settings))
      settings = mergeObjects(settings, entry.value.settings);
    if (record(entry.value.languageOptions)) {
      languageOptions = mergeObjects(
        languageOptions,
        entry.value.languageOptions,
      );
      const parser = entry.value.languageOptions.parserOptions;
      if (record(parser) && 'project' in parser) projectOrigin = entry.base;
    }
  }
  const parser = record(languageOptions.parserOptions)
    ? languageOptions.parserOptions
    : {};
  const service = parser.projectService;
  if (service !== undefined && service !== null && typeof service !== 'boolean')
    throw new Error('parserOptions.projectService must be boolean or null');
  const root = parser.tsconfigRootDir;
  if (
    root !== undefined &&
    root !== null &&
    (typeof root !== 'string' || !path.isAbsolute(root))
  )
    throw new Error(
      'parserOptions.tsconfigRootDir must be an absolute path string',
    );
  const project = parser.project;
  const projectDisabled =
    project === false || project === null || project === '';
  if (project === true)
    throw new Error(
      'parserOptions.project: true is not supported; use projectService: true',
    );
  if (
    project !== undefined &&
    !projectDisabled &&
    typeof project !== 'string' &&
    !(
      Array.isArray(project) &&
      project.every((item) => typeof item === 'string')
    )
  )
    throw new Error('Invalid parserOptions.project');
  if (service === true && project !== undefined && !projectDisabled)
    throw new Error(
      'Remove parserOptions.project when projectService is enabled',
    );
  const explicitProject =
    project !== undefined && !projectDisabled
      ? {
          patterns:
            typeof project === 'string'
              ? [project]
              : Array.isArray(project)
                ? project.filter(
                    (item): item is string => typeof item === 'string',
                  )
                : [],
          baseDirectory:
            typeof root === 'string'
              ? path.resolve(root)
              : (projectOrigin ?? defaultRoot),
        }
      : undefined;
  const projectPolicy: AdmittedProjectPolicy = {
    defaultProjectDisabled: service !== undefined && service !== true,
    projectDisabled: projectDisabled && service !== true,
    ...(service === true
      ? {
          serviceRootDirectory:
            typeof root === 'string' ? path.resolve(root) : defaultRoot,
        }
      : {}),
    ...(explicitProject ? { explicitProject } : {}),
    ...(typeof root === 'string'
      ? { tsconfigRootDirOverride: path.resolve(root) }
      : {}),
  };
  const enabledRules: AdmittedRule[] = [];
  for (const [name, rule] of rules) {
    if (rule.severity === 0) continue;
    const separator = name.lastIndexOf('/');
    if (separator >= 0 && !plugins.has(name.slice(0, separator)))
      throw new Error(`Rule ${name} requires its native plugin declaration`);
    enabledRules.push({ name, severity: rule.severity, options: rule.options });
  }
  enabledRules.sort((a, b) => a.name.localeCompare(b.name));
  const resolved = {
    rules: enabledRules,
    plugins: [...plugins].sort(),
    settings,
    languageOptions,
    projectPolicy,
    matchedEntries,
  };
  cache.set(key, resolved);
  return resolved;
}

/**
 * Admit targets for an explicitly evaluated flat config. No compiler, addon,
 * subprocess, or rule implementation is loaded here. Automatic hierarchical
 * config-module discovery is intentionally a separate, not-yet-wired entrypoint.
 */
export function admitRustFiles(options: AdmissionOptions): AdmittedFile[] {
  if (!path.isAbsolute(options.cwd))
    throw new Error('Rust admission cwd must be absolute');
  const cwd = path.resolve(options.cwd);
  const paths = new AdmissionPaths();
  const entries = compileEntries(options);
  const globals = entries
    .filter((entry) => entry.global)
    .flatMap((entry) => entry.ignores);
  const textFilePath =
    options.textFilePath === undefined
      ? undefined
      : path.resolve(cwd, options.textFilePath);
  const targets =
    textFilePath === undefined ? (options.patterns ?? ['.']) : [textFilePath];
  if (!Array.isArray(targets))
    throw new Error('Rust admission targets must be an array');
  if (targets.length === 0)
    throw new Error('Rust admission requires non-empty targets');
  const found = new Map<string, AdmittedFile>();
  const gitCache = new Map<string, Pattern[]>();
  const configCache = new Map<string, AdmittedConfig>();
  const readGit = (directory: string): Pattern[] => {
    let cached = gitCache.get(directory);
    if (!cached) {
      cached = gitPatterns(directory);
      gitCache.set(directory, cached);
    }
    return cached;
  };
  function gitIgnored(
    patterns: Pattern[],
    filename: string,
    directory: boolean,
  ): boolean {
    const nodes: string[] = [];
    for (
      let current = directory ? filename : path.dirname(filename);
      current !== cwd && paths.within(cwd, current);
      current = path.dirname(current)
    )
      nodes.push(current);
    for (const node of nodes.reverse()) {
      if (applyPatterns(paths, patterns, node, true)) return true;
    }
    return !directory && applyPatterns(paths, patterns, filename, false);
  }
  function reopen(directory: string): boolean {
    let reopened = false;
    for (const pattern of globals) {
      if (
        applyPatterns(paths, [{ ...pattern, negated: false }], directory, true)
      )
        reopened = pattern.negated;
    }
    return reopened;
  }
  function visit(
    filename: string,
    git: Pattern[],
    targetMatch?: (relative: string) => boolean,
  ): void {
    const name = path.basename(filename);
    const virtual = filename === textFilePath;
    const stat = virtual ? undefined : fs.statSync(filename);
    if (stat?.isDirectory()) {
      if (
        EXCLUDED_DIRECTORIES.has(name) ||
        absoluteDirectoryBlock(paths, globals, filename)
      )
        return;
      if (
        filename !== cwd &&
        gitIgnored(git, filename, true) &&
        !reopen(filename)
      )
        return;
      const nextGit = [...git, ...readGit(filename)];
      for (const child of fs
        .readdirSync(filename, { withFileTypes: true })
        .sort((a, b) => a.name.localeCompare(b.name))) {
        const childPath = path.join(filename, child.name);
        if (child.isSymbolicLink()) {
          let childStat: fs.Stats;
          try {
            childStat = fs.statSync(childPath);
          } catch (error) {
            if (record(error) && error.code === 'ENOENT') continue;
            throw error;
          }
          if (childStat.isDirectory()) continue;
        }
        visit(childPath, nextGit, targetMatch);
      }
      return;
    }
    if (
      (!virtual && !stat?.isFile()) ||
      !EXTENSIONS.has(path.extname(filename).toLowerCase())
    )
      return;
    if (targetMatch && !targetMatch(slash(paths.relative(cwd, filename))))
      return;
    const ignoredByGit = gitIgnored(git, filename, false);
    if (applyPatterns(paths, globals, filename, false, ignoredByGit)) return;
    const config = mergeConfig(
      paths,
      entries,
      filename,
      path.resolve(options.configDirectory ?? cwd),
      configCache,
    );
    found.set(filename, {
      path: filename,
      canonicalPath: paths.realpath(filename),
      config,
    });
  }
  for (const target of targets) {
    if (typeof target !== 'string' || target.length === 0)
      throw new Error('Rust admission targets must be non-empty strings');
    const absolute = path.resolve(cwd, target);
    if (!paths.within(cwd, absolute))
      throw new Error(
        'Rust admission does not yet support targets outside cwd',
      );
    if (fs.existsSync(absolute) || absolute === textFilePath) {
      // Include each ancestor's .gitignore and enforce the same traversal gates
      // for literal files and subdirectories as a root traversal.
      const relative = paths.relative(cwd, absolute);
      const components = relative.split(path.sep).filter(Boolean);
      let parent = cwd;
      let git: Pattern[] = [];
      let blocked = false;
      for (const component of components.slice(0, -1)) {
        git = [...git, ...readGit(parent)];
        parent = path.join(parent, component);
        if (
          EXCLUDED_DIRECTORIES.has(component) ||
          absoluteDirectoryBlock(paths, globals, parent) ||
          (gitIgnored(git, parent, true) && !reopen(parent))
        ) {
          blocked = true;
          break;
        }
      }
      if (!blocked)
        visit(absolute, absolute === cwd ? [] : [...git, ...readGit(parent)]);
    } else if (/[*?\[{}]/.test(target)) {
      visit(cwd, [], matcher(target));
    } else throw new Error(`Rust admission target does not exist: ${target}`);
  }
  return [...found.values()].sort((a, b) => a.path.localeCompare(b.path));
}
