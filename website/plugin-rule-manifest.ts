import type { RspressPlugin } from '@rspress/core';
import { execSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import * as rslintCore from '@rslint/core';
import type { RslintConfigEntry } from '@rslint/core';
import { PLUGIN_REGISTRY, groupToRouteSlug } from './theme/plugin-registry.ts';
import ruleReleases from './rule-releases.json' with { type: 'json' };

const REPO_ROOT = path.resolve(import.meta.dirname, '..');
const MANIFEST_PATH = path.resolve(
  import.meta.dirname,
  'generated/rule-manifest.json',
);
const SCRIPT_PATH = path.resolve(REPO_ROOT, 'scripts/gen-rule-manifest.js');
const RULES_DOCS_DIR = path.resolve(import.meta.dirname, 'docs/en/rules');
const RULE_VERSION_BY_ID = new Map(
  ruleReleases.flatMap(({ version, rules }) =>
    rules.map((rule) => [rule, version] as const),
  ),
);

/** Shape of each rule entry in rule-manifest.json. */
interface RuleEntry {
  name: string;
  group: string;
  status: string;
  failing_case: { name: string; url: string }[];
  /** Relative path (from repo root) to the rule's .md doc file, or null if absent. */
  docPath: string | null;
  /** Presets that include this rule, with their configured values. */
  presets: { name: string; value: unknown }[];
}

/** Resolve a preset to its config entries; a preset may ship several layers. */
function resolvePresetConfig(
  importName: string,
  presetName: string,
): RslintConfigEntry[] {
  const mod = (rslintCore as Record<string, unknown>)[importName] as
    | { configs?: Record<string, RslintConfigEntry | RslintConfigEntry[]> }
    | undefined;
  const prefix = `${importName}.configs.`;
  if (!presetName.startsWith(prefix)) return [];
  const configKey = presetName.slice(prefix.length);
  const config = mod?.configs?.[configKey];
  return config ? [config].flat() : [];
}

const PLUGINS = PLUGIN_REGISTRY.map((p) => ({
  prefix: p.prefix,
  group: p.group,
  presets: p.presets
    .map(({ name }) => {
      const entries = resolvePresetConfig(p.importName, name);
      return entries.length > 0 ? { entries, name } : null;
    })
    .filter((entry): entry is { entries: RslintConfigEntry[]; name: string } =>
      Boolean(entry),
    ),
}));

function getFullRuleName(rule: RuleEntry): string {
  if (rule.group === 'eslint') return rule.name;
  const entry = PLUGINS.find((e) => e.group === rule.group);
  const prefix = entry?.prefix || rule.group;
  return `${prefix}/${rule.name}`;
}

function parseRuleKey(ruleKey: string): { group: string; name: string } {
  for (const { prefix, group } of PLUGINS) {
    if (prefix && ruleKey.startsWith(`${prefix}/`)) {
      return { group, name: ruleKey.slice(prefix.length + 1) };
    }
  }
  return { group: 'eslint', name: ruleKey };
}

interface PresetInfo {
  name: string; // e.g. "ts.configs.recommended"
  value: unknown; // e.g. "error" or ["error", { varsIgnorePattern: "^_" }]
}

/**
 * Walk the actual preset config objects and return a map of
 * "group:name" → PresetInfo[] for every rule that is actively enabled
 * (severity !== 'off') in at least one preset.
 */
function extractPresetRules(): Map<string, PresetInfo[]> {
  const result = new Map<string, PresetInfo[]>();

  for (const { presets } of PLUGINS) {
    for (const { entries, name: presetName } of presets) {
      // The entries cascade in order, so the last one that mentions a rule
      // decides the value the preset applies.
      const effective = new Map<string, unknown>();
      for (const entry of entries) {
        for (const [ruleKey, value] of Object.entries(entry.rules ?? {})) {
          effective.set(ruleKey, value);
        }
      }
      for (const [ruleKey, value] of effective) {
        const severity = Array.isArray(value) ? value[0] : value;
        if (severity === 'off') continue;

        const { group, name } = parseRuleKey(ruleKey);
        const manifestKey = `${group}:${name}`;
        if (!result.has(manifestKey)) result.set(manifestKey, []);
        result.get(manifestKey)!.push({ name: presetName, value: value! });
      }
    }
  }
  return result;
}

/**
 * Run the manifest generation script and return the parsed rule entries.
 * The script scans Go source directories for rule implementations and
 * test status, writing the result to website/generated/rule-manifest.json.
 * After loading, enriches each entry with recommended preset information
 * and writes the result back so React components can consume it.
 */
function loadManifest(): RuleEntry[] {
  execSync(`node "${SCRIPT_PATH}"`, { stdio: 'inherit' });
  const manifest = JSON.parse(fs.readFileSync(MANIFEST_PATH, 'utf-8'));
  const rules: RuleEntry[] = manifest.rules;

  // Attach the list of presets that enable each rule, with the raw
  // configured value (severity string or [severity, options] tuple).
  const presetMap = extractPresetRules();
  for (const rule of rules) {
    const key = `${rule.group}:${rule.name}`;
    rule.presets = presetMap.get(key) || [];
  }

  // Persist the enriched manifest so runtime consumers (e.g. the rules
  // explorer in RuleStates/rule.tsx) can read presets directly.
  fs.writeFileSync(MANIFEST_PATH, JSON.stringify(manifest, null, 2) + '\n');

  return rules;
}

/**
 * Transform a rule's source .md into a .mdx string that imports and renders
 * the <RuleConfig> component right after the first heading. When the rule
 * is enabled by one or more presets, a markdown table summarizing each
 * preset and its configured value is inserted before <RuleConfig>.
 *
 * Input (source .md):
 *   # no-console
 *   ## Rule Details
 *   ...
 *
 * Output (.mdx):
 *   import RuleConfig from '@/theme/components/RuleConfig.tsx';
 *   import RuleVersionBadge from '@/theme/components/RuleVersionBadge.tsx';
 *
 *   # no-console
 *
 *   <RuleVersionBadge introducedIn="0.1.4" />
 *
 *   ## Configuration
 *
 *   | Preset                   | Configured Value |
 *   | ------------------------ | ---------------- |
 *   | ✅ js.configs.recommended | `"error"`        |
 *
 *   <RuleConfig name="no-console" group="eslint" />
 *
 *   ## Rule Details
 *   ...
 */
function buildRuleDocContent(rule: RuleEntry): string {
  const sourceContent = fs.readFileSync(
    path.resolve(REPO_ROOT, rule.docPath!),
    'utf-8',
  );
  const fullName = getFullRuleName(rule);
  const importLines = [
    `import RuleConfig from '@/theme/components/RuleConfig.tsx';`,
    `import RuleVersionBadge from '@/theme/components/RuleVersionBadge.tsx';`,
  ].join('\n');
  const introducedIn = RULE_VERSION_BY_ID.get(`${rule.group}:${rule.name}`);
  const versionBadge = introducedIn
    ? `<RuleVersionBadge introducedIn="${introducedIn}" />`
    : `<RuleVersionBadge />`;

  // Build preset table if the rule is in any preset
  let presetTable = '';
  if (rule.presets.length > 0) {
    const rows = rule.presets.map((p) => {
      const val =
        typeof p.value === 'string'
          ? `\`"${p.value}"\``
          : `\`${JSON.stringify(p.value)}\``;
      return `| ✅ ${p.name} | ${val} |`;
    });
    presetTable =
      `| Preset | Configured Value |\n` +
      `| ------ | ---------------- |\n` +
      rows.join('\n') +
      '\n\n';
  }

  const configSection =
    `## Configuration\n\n` +
    presetTable +
    `<RuleConfig name="${fullName}" group="${rule.group}" />`;

  const headingEnd = sourceContent.indexOf('\n');
  if (headingEnd === -1) {
    return `${importLines}\n\n${sourceContent}\n\n${versionBadge}\n\n${configSection}\n`;
  }
  return [
    importLines,
    '',
    sourceContent.slice(0, headingEnd),
    '',
    versionBadge,
    '',
    configSection,
    '',
    sourceContent.slice(headingEnd + 1),
  ].join('\n');
}

/**
 * Write rule doc .mdx files and `_meta.json` to docs/en/rules/ so that
 * Rspress auto-sidebar picks them up alongside other sections (/guide, /config).
 *
 * Generated structure:
 *   docs/en/rules/
 *     _meta.json                        ← section headers + Overview + group dirs
 *     index.mdx                         ← already exists (Overview page)
 *     eslint/
 *       _meta.json                      ← lists individual rules
 *       no-console.mdx                  ← imports <RuleConfig> via @/ alias
 *       ...
 *     typescript-eslint/
 *       _meta.json
 *       no-explicit-any.mdx
 *       ...
 */
function writeRuleDocsToDir(rules: RuleEntry[]): void {
  // Clean all generated files, keeping only the source-controlled index.mdx
  for (const name of fs.readdirSync(RULES_DOCS_DIR)) {
    if (name !== 'index.mdx') {
      fs.rmSync(path.join(RULES_DOCS_DIR, name), {
        recursive: true,
        force: true,
      });
    }
  }

  const rulesWithDocs = rules.filter((r) => r.docPath);

  // Group rules by slug
  const groups = new Map<string, RuleEntry[]>();
  for (const rule of rulesWithDocs) {
    const slug = groupToRouteSlug(rule.group);
    if (!groups.has(slug)) groups.set(slug, []);
    groups.get(slug)!.push(rule);
  }

  // Sort groups: eslint first, typescript-eslint second, rest alphabetical
  const sortedGroups = Array.from(groups.entries()).sort(([a], [b]) => {
    const order: Record<string, number> = {
      eslint: 0,
      'typescript-eslint': 1,
    };
    const oa = order[a] ?? 2;
    const ob = order[b] ?? 2;
    return oa !== ob ? oa - ob : a.localeCompare(b);
  });

  // Write top-level _meta.json: grouped overview + one dir per plugin
  const topMeta: unknown[] = [
    { type: 'section-header', label: 'Rule Status' },
    { type: 'file', name: 'index', label: 'Overview' },
    { type: 'section-header', label: 'Plugins' },
    ...sortedGroups.map(([slug]) => ({
      type: 'dir',
      name: slug,
      label: slug,
      collapsed: true,
    })),
  ];
  fs.writeFileSync(
    path.join(RULES_DOCS_DIR, '_meta.json'),
    JSON.stringify(topMeta, null, 2) + '\n',
  );

  // Write each group directory with _meta.json and rule .mdx files
  for (const [slug, groupRules] of sortedGroups) {
    const groupDir = path.join(RULES_DOCS_DIR, slug);
    fs.mkdirSync(groupDir, { recursive: true });

    const sorted = groupRules.sort((a, b) => a.name.localeCompare(b.name));

    const groupMeta = sorted.map((rule) => ({
      type: 'file',
      name: rule.name,
    }));
    fs.writeFileSync(
      path.join(groupDir, '_meta.json'),
      JSON.stringify(groupMeta, null, 2) + '\n',
    );

    for (const rule of sorted) {
      fs.writeFileSync(
        path.join(groupDir, `${rule.name}.mdx`),
        buildRuleDocContent(rule),
      );
    }
  }
}

/**
 * Rspress plugin that generates rule documentation pages from Go source.
 * Inside the `config` hook (run before Rspress resolves the route tree)
 * it:
 *
 * 1. Runs scripts/gen-rule-manifest.js to produce generated/rule-manifest.json
 *    and reads it back, enriching each entry with preset information.
 * 2. Writes .mdx files + _meta.json into docs/en/rules/<slug>/, so the
 *    Rspress auto-sidebar handles /rules/ the same way as /guide/ and
 *    /config/. Each <slug> is derived from `rule.group` via
 *    {@link groupToRouteSlug}.
 *
 * Each generated .mdx imports <RuleConfig> via the @/ alias to render a
 * copyable configuration snippet for the rule.
 */
export function pluginRuleManifest(): RspressPlugin {
  let rules: RuleEntry[] | null = null;

  function getRules(): RuleEntry[] {
    if (!rules) {
      rules = loadManifest();
    }
    return rules;
  }

  return {
    name: 'rule-manifest',

    config(config) {
      writeRuleDocsToDir(getRules());
      return config;
    },
  };
}
