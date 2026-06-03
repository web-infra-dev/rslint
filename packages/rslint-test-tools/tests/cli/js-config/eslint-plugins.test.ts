import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs/promises';
import path from 'node:path';
import { runRslint, createTempDir, cleanupTempDir } from './helpers.js';

// A self-contained mock ESLint plugin (mirrors tests/eslint-plugin/
// fixtures/local-plugin.mjs) so the e2e exercises the real plugin-load →
// dispatch → worker → diagnostic round-trip without an external dep.
const LOCAL_PLUGIN = `export default {
  meta: { name: 'local-plugin', version: '1.0.0' },
  rules: {
    'no-null': {
      meta: {
        type: 'suggestion',
        hasSuggestions: true,
        schema: [{ type: 'object', properties: { checkStrictEquality: { type: 'boolean' } }, additionalProperties: false }],
        defaultOptions: [{ checkStrictEquality: false }],
        messages: { error: 'Do not use \`null\`; prefer \`undefined\`.', replaceWithUndefined: 'Replace \`null\` with \`undefined\`.' },
      },
      create(context) {
        const { checkStrictEquality } = context.options[0];
        return {
          Literal(node) {
            if (node.raw !== 'null') return;
            const parent = node.parent;
            if (!checkStrictEquality && parent && parent.type === 'BinaryExpression' && (parent.operator === '===' || parent.operator === '!==')) return;
            context.report({ node, messageId: 'error', suggest: [{ messageId: 'replaceWithUndefined', fix: (fixer) => fixer.replaceText(node, 'undefined') }] });
          },
        };
      },
    },
    'prefer-array-some': {
      meta: { type: 'suggestion', fixable: 'code', schema: [], messages: { preferSome: 'Prefer \`.some(…)\` over \`.filter(…)\`.' } },
      create(context) {
        return {
          MemberExpression(node) {
            if (node.property && node.property.type === 'Identifier' && node.property.name === 'filter') {
              context.report({ node: node.property, messageId: 'preferSome', fix: (fixer) => fixer.replaceText(node.property, 'some') });
            }
          },
        };
      },
    },
  },
};
`;

const PLUGIN_CONFIG = `import local from './local-plugin.mjs';
export default [
  {
    files: ['**/*.ts'],
    eslintPlugins: { local },
    rules: { 'local/no-null': 'error', 'local/prefer-array-some': 'error' },
  },
];
`;

const TSCONFIG = JSON.stringify({
  compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
  include: ['**/*.ts'],
});

interface Diag {
  ruleName: string;
  filePath: string;
  severity: string;
  range: {
    start: { line: number; column: number };
    end: { line: number; column: number };
  };
}

async function lint(
  files: Record<string, string>,
  args: string[] = [],
): Promise<{ dir: string; exitCode: number; diags: Diag[] }> {
  const dir = await createTempDir(files);
  const res = await runRslint(['--format', 'jsonline', ...args], dir);
  const diags = res.stdout
    .trim()
    .split('\n')
    .filter((l) => l.trim().startsWith('{'))
    .map((l) => JSON.parse(l) as Diag);
  return { dir, exitCode: res.exitCode, diags };
}

describe('CLI eslintPlugins end-to-end', () => {
  test('plugin rules produce diagnostics at precise locations', async () => {
    const { dir, diags } = await lint({
      'local-plugin.mjs': LOCAL_PLUGIN,
      'rslint.config.mjs': PLUGIN_CONFIG,
      'tsconfig.json': TSCONFIG,
      'a.ts': 'const y = null;\nconst arr = [1, 2, 3].filter((x) => x > 1);\n',
    });
    try {
      const noNull = diags.find((d) => d.ruleName === 'local/no-null');
      expect(noNull).toBeDefined();
      expect(noNull!.severity).toBe('error');
      // `const y = null;` — `null` spans columns 11–15 (1-based, end-exclusive).
      expect(noNull!.range.start.line).toBe(1);
      expect(noNull!.range.start.column).toBe(11);
      expect(noNull!.range.end.column).toBe(15);

      const preferSome = diags.find(
        (d) => d.ruleName === 'local/prefer-array-some',
      );
      expect(preferSome).toBeDefined();
      expect(preferSome!.range.start.line).toBe(2);
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('--fix applies plugin fixes to disk', async () => {
    const { dir } = await lint(
      {
        'local-plugin.mjs': LOCAL_PLUGIN,
        'rslint.config.mjs': PLUGIN_CONFIG,
        'tsconfig.json': TSCONFIG,
        'b.ts': 'const arr = [1, 2, 3].filter((x) => x > 1);\n',
      },
      ['--fix'],
    );
    try {
      const content = await fs.readFile(path.join(dir, 'b.ts'), 'utf8');
      expect(content).toContain('.some(');
      expect(content).not.toContain('.filter(');
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('mixed native + plugin rules both report on one file', async () => {
    const MIXED_CONFIG = `import local from './local-plugin.mjs';
export default [
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    eslintPlugins: { local },
    plugins: ['@typescript-eslint'],
    rules: { 'local/no-null': 'error', '@typescript-eslint/no-explicit-any': 'error' },
  },
];
`;
    const { dir, diags } = await lint({
      'local-plugin.mjs': LOCAL_PLUGIN,
      'rslint.config.mjs': MIXED_CONFIG,
      'tsconfig.json': TSCONFIG,
      'c.ts': 'const y: any = null;\n',
    });
    try {
      const ruleNames = diags.map((d) => d.ruleName);
      expect(ruleNames).toContain('local/no-null');
      expect(ruleNames).toContain('@typescript-eslint/no-explicit-any');
      // The plugin diagnostic keeps a precise location even when merged with a
      // native diagnostic on the same file: `null` in `const y: any = null;`
      // spans columns 16–20 (1-based, end-exclusive).
      const noNull = diags.find((d) => d.ruleName === 'local/no-null');
      expect(noNull!.range.start.line).toBe(1);
      expect(noNull!.range.start.column).toBe(16);
      expect(noNull!.range.end.column).toBe(20);
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('a plugin with no rules object fails fast (non-zero exit)', async () => {
    const dir = await createTempDir({
      'rslint.config.mjs': `export default [{ files: ['**/*.ts'], eslintPlugins: { bad: { meta: {} } }, rules: { 'bad/x': 'error' } }];\n`,
      'tsconfig.json': TSCONFIG,
      'd.ts': 'const x = 1;\n',
    });
    try {
      const res = await runRslint(['--format', 'jsonline', 'd.ts'], dir);
      expect(res.exitCode).not.toBe(0);
      // The failure is the specific validation error, not an unrelated crash.
      expect(res.stderr).toContain('must expose a "rules" object');
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('--fix preserves plugin fixes alongside native fixes (multi-pass merge)', async () => {
    const MIXED_FIX_CONFIG = `import local from './local-plugin.mjs';
export default [
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    eslintPlugins: { local },
    plugins: ['@typescript-eslint'],
    rules: { 'local/prefer-array-some': 'error', '@typescript-eslint/no-inferrable-types': 'error' },
  },
];
`;
    const { dir } = await lint(
      {
        'local-plugin.mjs': LOCAL_PLUGIN,
        'rslint.config.mjs': MIXED_FIX_CONFIG,
        'tsconfig.json': TSCONFIG,
        'e.ts':
          'const n: number = 10;\nconst arr = [1, 2, 3].filter((x) => x > 1);\n',
      },
      ['--fix'],
    );
    try {
      const content = await fs.readFile(path.join(dir, 'e.ts'), 'utf8');
      // Native fix (no-inferrable-types) removed the redundant annotation …
      expect(content).not.toContain(': number');
      // … AND the plugin fix (prefer-array-some) survived the SAME --fix run,
      // proving plugin diagnostics are not clobbered by the native fix pass.
      expect(content).toContain('.some(');
      expect(content).not.toContain('.filter(');
    } finally {
      await cleanupTempDir(dir);
    }
  });
});
