import path from 'node:path';
import fs from 'node:fs';
import type { ConfigDescriptor } from '@rslint/eslint-plugin-runner';
import { configEntryHasEslintPlugins } from '@rslint/eslint-plugin-runner/plugin-source';
import {
  loadConfigFile,
  normalizeConfig,
  type EslintPluginEntry,
} from './config-loader.js';
import { parseArgs, classifyArgs, isJSConfigFile } from './utils/args.js';
import {
  discoverConfigs,
  filterConfigsByParentIgnores,
  type ConfigEntry,
} from './utils/config-discovery.js';

/**
 * Single end-to-end CLI run. Always goes through engine.ts → IPC →
 * Go binary's `runCLI` regardless of intent (lint / `--init` / `--help` /
 * JSON-config fallback). The engine.ts WorkerPool is sized at 0 when no
 * eslintPlugins entries are present, so the only overhead over a direct
 * binary invocation is one short IPC handshake.
 *
 * Why one path:
 *   - The Go binary is an internal npm artifact — no user reaches it
 *     without going through cli.ts. The IPC contract is uniform.
 *   - Earlier code branched on whether eslintPlugins existed and whether
 *     the Go binary was invoked directly via execFileSync vs spawned
 *     long-running. Two paths bred subtle drift (different config
 *     loading, different stdout pipes, different exit-code translation).
 *     Folding to one path eliminates the drift surface entirely.
 */
async function runViaEngine(opts: {
  binPath: string;
  configEntries: ConfigEntry[]; // [] when JSON config / --init / --help
  goArgs: string[];
  cwd: string;
  positionalArgs: string[];
  parsedFlags: {
    fix?: boolean;
    format?: string;
    singleThreaded?: boolean;
  };
}): Promise<number> {
  // ── Wire-format `eslintPluginEntries` for Go's IPC `init` payload ──
  // Go reads `prefix + ruleNames` to register placeholder rules and
  // power the `enforcePlugins` gate.
  //
  // Cross-config dedupe: if two configs declare the SAME prefix, we
  // union their rule names into one wire entry (Go would otherwise
  // see duplicate placeholder registrations, which it tolerates but
  // is wasteful). Per-file dispatch still routes each file to its
  // own config's plugin set via `configKey`, so the union here is
  // purely about completeness of the placeholder registry.
  const prefixToRuleNames = new Map<string, Set<string>>();
  for (const cfg of opts.configEntries) {
    for (const entry of cfg.entries) {
      const ep = (entry as unknown as { eslintPlugins?: EslintPluginEntry[] })
        .eslintPlugins;
      if (!Array.isArray(ep)) continue;
      for (const e of ep) {
        let set = prefixToRuleNames.get(e.prefix);
        if (set == null) {
          set = new Set<string>();
          prefixToRuleNames.set(e.prefix, set);
        }
        for (const rn of e.ruleNames) set.add(rn);
      }
    }
  }
  const wireEntries: EslintPluginEntry[] = [];
  for (const prefix of [...prefixToRuleNames.keys()].sort()) {
    wireEntries.push({
      prefix,
      ruleNames: [...prefixToRuleNames.get(prefix)!].sort(),
    });
  }

  // ── ConfigDescriptor[] for the worker pool ─────────────────────────
  // The worker imports each config file on init and extracts plugin
  // instances directly. Only configs that DECLARE plugins need to be
  // sent — workers spin up plugin-load work per descriptor, so an
  // empty config would just waste an import.
  //
  // Path normalization: Go writes per-file `configKey` after running
  // `tspath.NormalizePath` which (a) converts backslashes to forward
  // slashes and (b) simplifies `.` / `..` segments. The worker's
  // `Map.get(configKey)` lookup must match byte-for-byte, so we apply
  // the same shape here using Node's `path.posix.normalize` (works on
  // every platform; takes posix-style input, so we feed it a
  // forward-slash version of the path).
  const workerConfigs: ConfigDescriptor[] = [];
  for (const cfg of opts.configEntries) {
    if (!cfg.entries.some(configEntryHasEslintPlugins)) continue;
    if (!cfg.configPath) continue;
    workerConfigs.push({
      configPath: cfg.configPath,
      configDirectory: path.posix.normalize(
        cfg.configDirectory.replaceAll('\\', '/'),
      ),
    });
  }

  // Lazy import — keeps engine.ts and the runner package off the
  // module-load critical path even if the run later short-circuits.
  const { runEngine } = await import('./engine.js');

  const result = await runEngine({
    binPath: opts.binPath,
    goArgs: opts.goArgs,
    eslintPluginEntries: wireEntries,
    workerConfigs,
    configs: opts.configEntries,
    cwd: opts.cwd,
    runtime: {
      forceColor: process.stdout.isTTY ? true : undefined,
      singleThreaded: opts.parsedFlags.singleThreaded,
    },
    extraInit: {
      files: opts.positionalArgs,
      format: opts.parsedFlags.format ?? 'default',
      fixMode: opts.parsedFlags.fix ?? false,
      workingDirectory: opts.cwd,
    },
  } as Parameters<typeof runEngine>[0] & { extraInit: unknown });
  return result.exitCode;
}

/**
 * Load JS/TS configs into ConfigEntry[]. Tolerates individual config
 * load failures — skips broken configs with a warning and continues.
 * Returns null when every supplied config failed to load (caller falls
 * back to the JSON-config path with empty configs).
 */
async function loadJsConfigs(
  configs: Map<string, string>,
): Promise<ConfigEntry[] | null> {
  const configEntries: ConfigEntry[] = [];
  const isSingleConfig = configs.size === 1;

  for (const [configPath, configDir] of configs) {
    let rawConfig: unknown;
    try {
      rawConfig = await loadConfigFile(configPath);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      if (isSingleConfig) {
        process.stderr.write(
          `Error: failed to load config ${configPath}: ${msg}\n`,
        );
        return null;
      }
      process.stderr.write(`Warning: skipping config ${configPath}: ${msg}\n`);
      continue;
    }

    let entries: ConfigEntry['entries'];
    try {
      entries = normalizeConfig(rawConfig);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      if (isSingleConfig) {
        process.stderr.write(
          `Error: invalid config in ${configPath}: ${msg}\n`,
        );
        return null;
      }
      process.stderr.write(`Warning: skipping config ${configPath}: ${msg}\n`);
      continue;
    }

    configEntries.push({ configDirectory: configDir, configPath, entries });
  }

  if (configEntries.length === 0) return null;
  return configEntries;
}

export async function run(
  binPath: string,
  argv: string[],
  startTime: number,
): Promise<number> {
  const cwd = process.cwd();
  const args = parseArgs(argv);

  // Validate explicit --config flag
  if (args.config) {
    const configPath = path.resolve(cwd, args.config);
    if (!fs.existsSync(configPath)) {
      process.stderr.write(`Error: config file not found: ${configPath}\n`);
      return 1;
    }
  }

  // Build Go args: start-time flag BEFORE positional args, because Go's
  // flag.Parse stops at the first positional argument. If --start-time comes
  // after positionals, it gets treated as a file path. The --init flag and
  // --config (filtered out by parseArgs) are reattached here so the binary's
  // flag.Parse still sees them.
  const goArgs = [`--start-time=${startTime}`, ...args.rest];
  if (args.init) goArgs.push('--init');
  if (args.config) goArgs.push('--config', args.config);

  // Classify positional arguments into files and directories
  const { files, dirs } = classifyArgs(args.positionals, cwd);

  // Discover JS/TS configs. `--init` and `--help`/`-h` both skip discovery:
  // --init creates a fresh config (scanning existing ones is pointless), and
  // --help must reach Go's usage output even when the project's
  // rslint.config.* is broken — otherwise loadJsConfigs would `return 1`
  // before help ever prints. (--help is still forwarded to Go via args.rest,
  // so Go prints the usage text.)
  const configs =
    args.init || args.help
      ? new Map<string, string>()
      : discoverConfigs(files, dirs, cwd, args.config);

  // Filter to JS/TS configs only. If any JS config is found (even in
  // subdirectories), the entire flow uses the JS config path. A root JSON
  // config (rslint.json) will be bypassed in this case. JSON config is
  // deprecated — projects should migrate to JS.
  const jsConfigs = new Map<string, string>();
  for (const [configPath, configDir] of configs) {
    if (isJSConfigFile(configPath)) {
      jsConfigs.set(configPath, configDir);
    }
  }

  let configEntries: ConfigEntry[] = [];
  if (jsConfigs.size > 0) {
    const loaded = await loadJsConfigs(jsConfigs);
    if (loaded === null) {
      // FAIL-FAST: at least one JS/TS config was discovered AND every
      // attempt to load it failed (single config → first failure;
      // multi-config → every config failed). We MUST NOT silently fall
      // back to the JSON / no-config path: the user explicitly authored
      // JS config, so plugin rules / overrides / etc. were intended to
      // run. Silent fallback would let CI exit 0 on a broken config —
      // exactly the "fake green" outcome lint tools must avoid.
      //
      // The specific error messages (config path + cause) were already
      // written to stderr inside loadJsConfigs. We just surface the
      // failure as a non-zero exit code here.
      return 1;
    }
    // Filter out nested configs whose directory is covered by a parent
    // config's global ignores — same ESLint-v10 behavior.
    configEntries = filterConfigsByParentIgnores(loaded);
  }

  return runViaEngine({
    binPath,
    configEntries,
    goArgs,
    cwd,
    positionalArgs: args.positionals,
    parsedFlags: {
      fix: args.rest.includes('--fix'),
      format: extractFlagValue(args.rest, '--format'),
      singleThreaded: args.rest.includes('--singleThreaded'),
    },
  });
}

/**
 * Extract `--flag=value` or `--flag value` from a flag list. Returns
 * undefined when not present. Used for forwarding `--format=jsonline` etc.
 * to engine.ts → init payload.
 */
function extractFlagValue(args: string[], flag: string): string | undefined {
  for (let i = 0; i < args.length; i++) {
    if (args[i] === flag && i + 1 < args.length) return args[i + 1];
    if (args[i].startsWith(flag + '=')) return args[i].slice(flag.length + 1);
  }
  return undefined;
}
