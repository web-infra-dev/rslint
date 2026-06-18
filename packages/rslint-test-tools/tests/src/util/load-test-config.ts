import path from 'node:path';

import { loadConfigFile, normalizeConfig } from '@rslint/core/config-loader';

// Per-test rslint config builder shared by every rule-tester under
// `packages/rslint-test-tools/tests/<plugin>/rule-tester.ts`.
//
// Each rule-tester loads `rslint.config.mjs`, optionally merges in per-test
// `settings`, and hands the resolved config object straight to `lint()`. The
// Node API takes a config object (Go never reads config from disk), so there
// is no temp file to write or clean up.
//
// The base config is cached by `baseConfigPath` so each .mjs is parsed exactly
// once per process, even when several rule-testers share this helper.
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
  config: Record<string, unknown>[];
  configDirectory: string;
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
  // Hand the resolved config object straight to the Node API (no temp file).
  // rslint resolves each entry's relative `tsconfig.*.json` against
  // configDirectory — the base config file's directory.
  return { config: merged, configDirectory: path.dirname(baseConfigPath) };
}
