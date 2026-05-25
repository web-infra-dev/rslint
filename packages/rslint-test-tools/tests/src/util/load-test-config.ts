import { rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';

import { loadConfigFile, normalizeConfig } from '@rslint/core/config-loader';

// Per-test rslint config builder shared by every rule-tester under
// `packages/rslint-test-tools/tests/<plugin>/rule-tester.ts`.
//
// Why we still emit a temporary file: rslint's `lint()` JS API takes a
// config path and the underlying Go binary parses it as hujson (a JSON
// superset) — it does NOT execute JS/TS configs the way the CLI does.
// So each rule-tester loads `rslint.config.mjs`, optionally merges in
// per-test `settings`, and writes a serialized config that `lint()`
// can consume.
//
// The base config is cached by `baseConfigPath` so each .mjs is parsed
// exactly once per process, even when several rule-testers share this
// helper.
const baseConfigCache = new Map<string, Record<string, unknown>[]>();

async function getBaseConfig(
  baseConfigPath: string,
): Promise<Record<string, unknown>[]> {
  const cached = baseConfigCache.get(baseConfigPath);
  if (cached) return cached;
  const raw = await loadConfigFile(baseConfigPath);
  const normalized = normalizeConfig(raw);
  baseConfigCache.set(baseConfigPath, normalized);
  return normalized;
}

export interface BuiltTestConfig {
  configPath: string;
  cleanup: () => void;
}

export async function buildConfigForSettings(
  baseConfigPath: string,
  settings: Record<string, unknown> | undefined,
): Promise<BuiltTestConfig> {
  const base = await getBaseConfig(baseConfigPath);
  const merged = base.map((entry) => ({
    ...entry,
    settings: {
      ...((entry.settings as object | undefined) ?? {}),
      ...(settings ?? {}),
    },
  }));
  // Write the serialized config next to the base — rslint resolves the
  // relative `tsconfig.*.json` paths inside each entry against the config
  // file's directory, so the temp file must live in the same dir.
  const baseDir = path.dirname(baseConfigPath);
  const cfg = path.join(
    baseDir,
    `.rslint.test-${process.pid}-${Date.now()}-${Math.random().toString(36).slice(2)}.tmp`,
  );
  writeFileSync(cfg, JSON.stringify(merged), 'utf8');
  return {
    configPath: cfg,
    cleanup: () => {
      try {
        rmSync(cfg, { force: true });
      } catch {
        /* best-effort cleanup; never fail a test on rmdir */
      }
    },
  };
}
