import type { RspressPlugin } from '@rspress/core';
import { execSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { ts, js, reactPlugin, importPlugin } from '@rslint/core';
import type { RslintConfigEntry } from '@rslint/core';

const REPO_ROOT = path.resolve(__dirname, '..');
const MANIFEST_PATH = path.resolve(__dirname, 'generated/rule-manifest.json');
const SCRIPT_PATH = path.resolve(REPO_ROOT, 'scripts/gen-rule-manifest.js');
const RULES_DOCS_DIR = path.resolve(__dirname, 'docs/en/rules');

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

/**
 * Convert a plugin group name to a URL-safe slug.
 * e.g. "@typescript-eslint" → "typescript-eslint"
 */
function groupToRouteSlug(group: string): string {
  return group.replace(/^@/, '');
}

/**
 * Return the fully-qualified rule name as used in rslint config.
 * Core ESLint rules have no prefix; plugin rules are prefixed with the group.
 * e.g. "no-console" for eslint, "@typescript-eslint/no-explicit-any" for TS plugin.
 */
function getFullRuleName(rule: RuleEntry): string {
  if (rule.group === 'eslint') return rule.name;
  return `${rule.group}/${rule.name}`;
}

/** Each preset config paired with its full reference name (e.g. "ts.configs.recommended"). */
const PRESETS: { config: RslintConfigEntry; name: string }[] = [
  { config: ts.configs.recommended, name: 'ts.configs.recommended' },
  { config: js.configs.recommended, name: 'js.configs.recommended' },
  {
    config: reactPlugin.configs.recommended,
    name: 'reactPlugin.configs.recommended',
  },
  {
    config: importPlugin.configs.recommended,
    name: 'importPlugin.configs.recommended',
  },
];

/**
 * Parse a fully-qualified rule key from a config into the (group, name)
 * pair used by the manifest.
 */
function parseRuleKey(ruleKey: string): { group: string; name: string } {
  if (ruleKey.startsWith('@typescript-eslint/')) {
    return {
      group: '@typescript-eslint',
      name: ruleKey.slice('@typescript-eslint/'.length),
    };
  }
  if (ruleKey.startsWith('react/')) {
    return { group: 'react', name: ruleKey.slice('react/'.length) };
  }
  if (ruleKey.startsWith('import/')) {
    return {
      group: 'eslint-plugin-import',
      name: ruleKey.slice('import/'.length),
    };
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

  for (const { config, name: presetName } of PRESETS) {
    if (!config.rules) continue;
    for (const [ruleKey, value] of Object.entries(config.rules)) {
      const severity = Array.isArray(value) ? value[0] : value;
      if (severity === 'off') continue;

      const { group, name } = parseRuleKey(ruleKey);
      const manifestKey = `${group}:${name}`;
      if (!result.has(manifestKey)) result.set(manifestKey, []);
      result.get(manifestKey)!.push({ name: presetName, value: value! });
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

  // Enrich with preset data (level + configured value)
  const presetMap = extractPresetRules();
  for (const rule of rules) {
    const key = `${rule.group}:${rule.name}`;
    rule.presets = presetMap.get(key) || [];
  }

  // Write enriched manifest back for React components
  fs.writeFileSync(MANIFEST_PATH, JSON.stringify(manifest, null, 2) + '\n');

  return rules;
}

/**
 * Transform a rule's source .md into a .mdx string that imports and renders
 * the <RuleConfig> component right after the first heading.
 *
 * Input (source .md):
 *   # no-console
 *   ## Rule Details
 *   ...
 *
 * Output (.mdx):
 *   import RuleConfig from '@/theme/components/RuleConfig.tsx';
 *
 *   # no-console
 *
 *   ## Configuration
 *
 *   <RuleConfig name="no-console" group="eslint" presets={["recommended"]} />
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
  const importLine = `import RuleConfig from '@/theme/components/RuleConfig.tsx';`;

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
    return `${importLine}\n\n${sourceContent}\n\n${configSection}\n`;
  }
  return [
    importLine,
    '',
    sourceContent.slice(0, headingEnd),
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
 *     _meta.json                        ← Overview + one entry per group dir
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

  // Write top-level _meta.json: Overview + one dir per group
  const topMeta: unknown[] = [
    { type: 'file', name: 'index', label: 'Overview' },
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
 * Rspress plugin that generates rule documentation pages from Go source:
 *
 * 1. Runs scripts/gen-rule-manifest.js to produce generated/rule-manifest.json
 * 2. Writes .mdx files + _meta.json into docs/en/rules/<group>/ in beforeBuild,
 *    so Rspress auto-sidebar handles /rules/ the same way as /guide/ and /config/
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
