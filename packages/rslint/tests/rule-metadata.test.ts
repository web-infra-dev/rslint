import { describe, expect, test } from 'rstack/test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { createRequire } from 'node:module';
import ts from 'typescript';
import { compileRuleOptionTypes } from '../plugins/generate-rule-option-types.js';
import { writeRuleDocsToDir } from '../../../website/plugin-rule-manifest.js';

const require = createRequire(import.meta.url);
const { buildManifest } = require('../../../scripts/gen-rule-manifest.js');
const {
  getCurrentRuleIds,
  getRuleIdsAtRef,
} = require('../../../scripts/sync-version-info/release.js');

function withFixture<T>(
  run: (root: string, write: (file: string, content: string) => void) => T,
): T {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rule-metadata-'));
  const write = (file: string, content: string) => {
    const target = path.join(root, file);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, content);
  };
  try {
    return run(root, write);
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

describe('rule metadata', () => {
  test('filesystem and Git discovery preserve nested rule IDs and test ownership', () => {
    withFixture((root, write) => {
      write('internal/rules/core_rule/core_rule.go', 'package core_rule');
      write('internal/rules/core_rule/helper/helper.go', 'package helper');
      write(
        'internal/plugins/node/plugin.go',
        'package node_plugin\nconst PLUGIN_NAME = "eslint-plugin-node"',
      );
      for (const rule of [
        'prefer_global',
        'prefer_global/url',
        'prefer_global/url/deep',
        'prefer_global_url',
      ]) {
        const leaf = rule.split('/').at(-1)!;
        const base = `internal/plugins/node/rules/${rule}`;
        write(`${base}/${leaf}.go`, `package ${leaf}`);
        write(
          `${base}/${leaf}.md`,
          `# ${rule.replaceAll('_', '-')}\n\nRule details.\n`,
        );
      }
      // Real helper and fixture trees must not become public rules.
      write(
        'internal/plugins/node/rules/prefer_global/testdata/fake/fake.go',
        'package fake',
      );
      write(
        'internal/plugins/node/rules/prefer_global/fixtures/fake/fake.go',
        'package fake',
      );
      write(
        'internal/plugins/node/rules/prefer_global/.hidden/fake/fake.go',
        'package fake',
      );
      write(
        'internal/plugins/node/rules/prefer_global/helpers/helpers_test.go',
        'package helpers',
      );
      const tests = 'packages/rslint-test-tools/tests';
      write(`${tests}/eslint/rules/core-rule.test.ts`, '// core');
      write(
        `${tests}/eslint-plugin-node/rules/prefer-global/base.test.ts`,
        '// parent',
      );
      write(
        `${tests}/eslint-plugin-node/rules/prefer-global/url.test.ts`,
        '// nested',
      );
      write(
        `${tests}/eslint-plugin-node/rules/prefer-global/url/extra.test.ts`,
        "it.skip('nested skip', () => {});",
      );
      write(
        `${tests}/eslint-plugin-node/rules/prefer-global/url/deep.test.ts`,
        '// deeper',
      );
      write(
        `${tests}/eslint-plugin-node/rules/prefer-global-url.test.ts`,
        '// flat',
      );
      write(
        'packages/rslint-test-tools/rstack.config.mts',
        [
          'eslint/rules/core-rule',
          'eslint-plugin-node/rules/prefer-global/base',
          'eslint-plugin-node/rules/prefer-global/url',
          'eslint-plugin-node/rules/prefer-global/url/extra',
          'eslint-plugin-node/rules/prefer-global/url/deep',
          'eslint-plugin-node/rules/prefer-global-url',
        ]
          .map((file) => `'./tests/${file}.test.ts',`)
          .join('\n'),
      );

      const manifest = buildManifest(root);
      const nodeRules = manifest.rules.filter(
        (rule: { group: string }) => rule.group === 'eslint-plugin-node',
      );
      expect(
        nodeRules.map((rule: { name: string }) => rule.name).sort(),
      ).toEqual([
        'prefer-global',
        'prefer-global-url',
        'prefer-global/url',
        'prefer-global/url/deep',
      ]);
      const nested = nodeRules.find(
        (rule: { name: string }) => rule.name === 'prefer-global/url',
      );
      expect(nested.docPath).toBe(
        'internal/plugins/node/rules/prefer_global/url/url.md',
      );
      expect(nested.status).toBe('partial-impl');
      expect(nested.failing_case).toHaveLength(1);
      expect(
        nodeRules
          .filter((rule: { name: string }) => rule.name !== nested.name)
          .every((rule: { status: string }) => rule.status === 'full'),
      ).toBe(true);

      const current = getCurrentRuleIds(root);
      expect(current).toEqual(
        [
          'eslint-plugin-node:prefer-global',
          'eslint-plugin-node:prefer-global-url',
          'eslint-plugin-node:prefer-global/url',
          'eslint-plugin-node:prefer-global/url/deep',
          'eslint:core-rule',
        ].sort(),
      );
      const git = (...args: string[]) =>
        execFileSync('git', args, { cwd: root, stdio: 'pipe' })
          .toString()
          .trim();
      git('init', '--quiet');
      git('add', '.');
      const tree = git('write-tree');
      // Before the TypeScript plugin directory existed, core rules belonged
      // to @typescript-eslint. Preserve that historical indexing convention.
      expect(getRuleIdsAtRef(tree, root)).toEqual(
        current
          .map((id: string) => id.replace(/^eslint:/, '@typescript-eslint:'))
          .sort(),
      );

      // Generate actual MDX and sidebar entries with both parent and child rules.
      write('docs/index.mdx', '# Overview');
      writeRuleDocsToDir(
        nodeRules.map((rule: { docPath: string }) => ({
          ...rule,
          docPath: path.join(root, rule.docPath),
          presets: [],
        })),
        path.join(root, 'docs'),
      );
      expect(
        fs.readFileSync(
          path.join(root, 'docs/node/prefer-global/url.mdx'),
          'utf8',
        ),
      ).toContain('name="node/prefer-global/url"');
      expect(
        fs.existsSync(path.join(root, 'docs/node/prefer-global/url/deep.mdx')),
      ).toBe(true);
      expect(
        JSON.parse(
          fs.readFileSync(path.join(root, 'docs/node/_meta.json'), 'utf8'),
        ),
      ).toContainEqual({ type: 'file', name: 'prefer-global/url' });
    });
  });

  test('generated option types preserve punctuation-distinct IDs without declaration collisions', async () => {
    const entries = [
      {
        name: 'node/prefer-global/url',
        schema: {
          type: 'array',
          items: [{ enum: ['nested'] }],
          additionalItems: false,
        },
      },
      {
        name: 'node/prefer-global-url',
        schema: {
          type: 'array',
          items: [{ type: 'boolean' }],
          additionalItems: false,
        },
      },
      {
        name: 'node/prefer_global_url',
        schema: {
          type: 'array',
          items: [{ type: 'number' }],
          additionalItems: false,
        },
      },
    ];
    const generated = await compileRuleOptionTypes(entries);
    expect(await compileRuleOptionTypes([...entries].reverse())).toEqual(
      generated,
    );
    withFixture((root, write) => {
      const source = `type RuleEntry<T extends unknown[]> = ['error', ...T];
interface RulesRecord { ${generated.recordProperties.join('\n')} }
${generated.typeDeclarations.join('\n')}
const valid: RulesRecord = {
  'node/prefer-global/url': ['error', 'nested'],
  'node/prefer-global-url': ['error', true],
  'node/prefer_global_url': ['error', 1],
};
const invalid: RulesRecord = {
  // @ts-expect-error -- the nested rule must retain its own options
  'node/prefer-global/url': ['error', true],
};
`;
      write('options.ts', source);
      const program = ts.createProgram([path.join(root, 'options.ts')], {
        strict: true,
        noEmit: true,
        types: [],
        target: ts.ScriptTarget.ES2022,
      });
      expect(
        ts
          .getPreEmitDiagnostics(program)
          .map((diagnostic) =>
            ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'),
          ),
      ).toEqual([]);
    });
  });
});
