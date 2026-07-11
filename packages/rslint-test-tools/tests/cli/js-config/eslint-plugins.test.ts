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
    'replace-one-zero': {
      meta: { type: 'suggestion', fixable: 'code', schema: [], messages: { replace: 'Replace one zero.' } },
      create(context) {
        return {
          Program(node) {
            const index = context.sourceCode.text.indexOf('0');
            if (index === -1) return;
            context.report({
              node,
              messageId: 'replace',
              fix: (fixer) => fixer.replaceTextRange([index, index + 1], '1'),
            });
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
    plugins: { local },
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

describe('CLI community plugins (object-form) end-to-end', () => {
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

  test('--fix performs a final verification after the tenth write pass', async () => {
    const config = `import local from './local-plugin.mjs';
export default [{
  files: ['**/*.ts'],
  plugins: { local },
  rules: { 'local/replace-one-zero': 'error' },
}];
`;
    const dir = await createTempDir({
      'local-plugin.mjs': LOCAL_PLUGIN,
      'rslint.config.mjs': config,
      'tsconfig.json': TSCONFIG,
      'ten.ts': "const value = '0000000000';\n",
    });
    try {
      const result = await runRslint(
        ['--format', 'jsonline', '--fix', 'ten.ts'],
        dir,
      );
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('local/replace-one-zero');
      expect(await fs.readFile(path.join(dir, 'ten.ts'), 'utf8')).toBe(
        "const value = '1111111111';\n",
      );
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('mixed native + plugin rules both report on one file', async () => {
    // Native and community plugins are separate `plugins` forms (array vs
    // object), so they live in two entries; both match `**/*.ts` and merge.
    const MIXED_CONFIG = `import local from './local-plugin.mjs';
export default [
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    plugins: ['@typescript-eslint'],
    rules: { '@typescript-eslint/no-explicit-any': 'error' },
  },
  {
    files: ['**/*.ts'],
    plugins: { local },
    rules: { 'local/no-null': 'error' },
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

  test('mixed native + plugin: exit code reflects a violation from either origin', async () => {
    // Exit code is the signal CI gates on. Native and plugin diagnostics merge
    // into one origin-agnostic error count — a regression dropping plugin diags
    // from the exit code would pass every presence-only (toContain) assertion.
    const MIXED = `import local from './local-plugin.mjs';
export default [
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
    plugins: ['@typescript-eslint'],
    rules: { '@typescript-eslint/no-explicit-any': 'error' },
  },
  {
    files: ['**/*.ts'],
    plugins: { local },
    rules: { 'local/no-null': 'error' },
  },
];
`;
    const base = {
      'local-plugin.mjs': LOCAL_PLUGIN,
      'rslint.config.mjs': MIXED,
      'tsconfig.json': TSCONFIG,
    };

    // (A) plugin-only violation (null, no any) → exit 1.
    const a = await lint({ ...base, 'c.ts': 'const y = null;\n' });
    try {
      expect(a.exitCode).toBe(1);
      expect(a.diags.some((d) => d.ruleName === 'local/no-null')).toBe(true);
      expect(
        a.diags.some(
          (d) => d.ruleName === '@typescript-eslint/no-explicit-any',
        ),
      ).toBe(false);
    } finally {
      await cleanupTempDir(a.dir);
    }

    // (B) native-only violation (any, no null) → exit 1.
    const b = await lint({ ...base, 'c.ts': 'const y: any = 1;\n' });
    try {
      expect(b.exitCode).toBe(1);
      expect(
        b.diags.some(
          (d) => d.ruleName === '@typescript-eslint/no-explicit-any',
        ),
      ).toBe(true);
      expect(b.diags.some((d) => d.ruleName === 'local/no-null')).toBe(false);
    } finally {
      await cleanupTempDir(b.dir);
    }

    // (C) clean → exit 0, no diagnostics.
    const c = await lint({ ...base, 'c.ts': 'const y = 1;\n' });
    try {
      expect(c.exitCode).toBe(0);
      expect(c.diags).toHaveLength(0);
    } finally {
      await cleanupTempDir(c.dir);
    }
  });

  test('a plugin with no rules object fails fast (non-zero exit)', async () => {
    const dir = await createTempDir({
      'rslint.config.mjs': `export default [{ files: ['**/*.ts'], plugins: { bad: { meta: {} } }, rules: { 'bad/x': 'error' } }];\n`,
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
    plugins: ['@typescript-eslint'],
    rules: { '@typescript-eslint/no-inferrable-types': 'error' },
  },
  {
    files: ['**/*.ts'],
    plugins: { local },
    rules: { 'local/prefer-array-some': 'error' },
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
