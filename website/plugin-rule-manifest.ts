import type {
  RspressPlugin,
  Sidebar,
  SidebarGroup,
  SidebarItem,
} from '@rspress/core';
import { execSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';

const REPO_ROOT = path.resolve(__dirname, '..');
const MANIFEST_PATH = path.resolve(__dirname, 'generated/rule-manifest.json');
const SCRIPT_PATH = path.resolve(REPO_ROOT, 'scripts/gen-rule-manifest.js');

/** Shape of each rule entry in rule-manifest.json. */
interface RuleEntry {
  name: string;
  group: string;
  status: string;
  failing_case: { name: string; url: string }[];
  /** Relative path (from repo root) to the rule's .md doc file, or null if absent. */
  docPath: string | null;
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

/**
 * Run the manifest generation script and return the parsed rule entries.
 * The script scans Go source directories for rule implementations and
 * test status, writing the result to website/generated/rule-manifest.json.
 */
function loadManifest(): RuleEntry[] {
  execSync(`node "${SCRIPT_PATH}"`, { stdio: 'inherit' });
  const manifest = JSON.parse(fs.readFileSync(MANIFEST_PATH, 'utf-8'));
  return manifest.rules;
}

/**
 * Build the /rules/ sidebar structure from manifest data.
 * Groups are sorted with eslint first, typescript-eslint second,
 * then the rest alphabetically. Rules within each group are alphabetical.
 */
function buildRulesSidebar(rules: RuleEntry[]): (SidebarGroup | SidebarItem)[] {
  const rulesWithDocs = rules.filter(r => r.docPath);

  // Collect rules into groups keyed by route slug
  const groups = new Map<string, SidebarItem[]>();
  for (const rule of rulesWithDocs) {
    const slug = groupToRouteSlug(rule.group);
    if (!groups.has(slug)) groups.set(slug, []);
    groups.get(slug)!.push({
      text: rule.name,
      link: `/rules/${slug}/${rule.name}`,
    });
  }

  return [
    { text: 'Overview', link: '/rules/' },
    ...Array.from(groups.entries())
      .sort(([a], [b]) => {
        // Pin eslint and typescript-eslint to the top; rest alphabetical
        const order: Record<string, number> = {
          eslint: 0,
          'typescript-eslint': 1,
        };
        const oa = order[a] ?? 2;
        const ob = order[b] ?? 2;
        return oa !== ob ? oa - ob : a.localeCompare(b);
      })
      .map(
        ([groupSlug, items]): SidebarGroup => ({
          text: groupSlug,
          collapsed: true,
          collapsible: true,
          items: items.sort((a, b) => a.text.localeCompare(b.text)),
        }),
      ),
  ];
}

/**
 * Rspress plugin that:
 * 1. Generates rule-manifest.json from Go source (beforeBuild via loadManifest)
 * 2. Registers <RuleConfig> as a global MDX component (markdown.globalComponents)
 * 3. Injects /rules/ sidebar and preserves the top nav bar (config hook)
 * 4. Registers a doc page for every rule that has a .md file (addPages hook),
 *    reading the source .md and inserting a <RuleConfig> component after the heading
 */
export function pluginRuleManifest(): RspressPlugin {
  // Cache manifest across hooks — generated once, reused by config and addPages
  let rules: RuleEntry[] | null = null;

  function getRules(): RuleEntry[] {
    if (!rules) {
      rules = loadManifest();
    }
    return rules;
  }

  return {
    name: 'rule-manifest',

    /**
     * Register <RuleConfig> as a global MDX component so rule doc pages
     * can use it without an explicit import statement.
     */
    markdown: {
      globalComponents: [
        path.resolve(__dirname, 'theme/components/RuleConfig.tsx'),
      ],
    },

    /**
     * Inject sidebar for /rules/ and re-supply the nav bar.
     * Setting themeConfig.sidebar explicitly disables Rspress's auto-generation
     * for the entire site, so we must also provide themeConfig.nav from _nav.json.
     */
    config(config) {
      const allRules = getRules();
      const sidebar: Sidebar = {
        ...(config.themeConfig?.sidebar as Sidebar),
        '/rules/': buildRulesSidebar(allRules),
      };

      const navPath = path.resolve(__dirname, 'docs/en/_nav.json');
      const nav = JSON.parse(fs.readFileSync(navPath, 'utf-8'));

      return {
        ...config,
        themeConfig: {
          ...config.themeConfig,
          nav,
          sidebar,
        },
      };
    },

    /**
     * Register a page for each rule that has a source .md doc.
     * We read the source content and insert a <RuleConfig> component
     * (rendered by CodeBlockRuntime with syntax highlighting + copy button)
     * right after the first heading.
     *
     * Route structure: /rules/<group-slug>/<rule-name>
     * e.g. /rules/typescript-eslint/no-explicit-any
     */
    addPages() {
      const allRules = getRules();
      return allRules
        .filter(r => r.docPath)
        .map(rule => {
          const sourceContent = fs.readFileSync(
            path.resolve(REPO_ROOT, rule.docPath!),
            'utf-8',
          );
          const fullName = getFullRuleName(rule);
          const configBlock =
            `## Configuration\n\n` +
            `<RuleConfig name="${fullName}" group="${rule.group}" />\n`;
          // Insert <RuleConfig> component right after the first line (# heading)
          const headingEnd = sourceContent.indexOf('\n');
          const content =
            headingEnd === -1
              ? `${sourceContent}\n\n${configBlock}`
              : sourceContent.slice(0, headingEnd + 1) +
                '\n' +
                configBlock +
                '\n' +
                sourceContent.slice(headingEnd + 1);
          return {
            routePath: `/rules/${groupToRouteSlug(rule.group)}/${rule.name}`,
            content,
          };
        });
    },
  };
}
