import { describe, test, expect } from 'rstack/test';
import fs from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { Linter } from 'eslint';
import * as typescriptParser from '@typescript-eslint/parser';
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
    'projection-probe': {
      meta: { fixable: 'code', schema: [] },
      create(context) {
        const sc = context.sourceCode;
        return {
          AssignmentPattern(node) {
            context.report({ node, message: 'remove default', fix: fixer => fixer.replaceText(node, sc.getText(node.left)) });
          },
          SequenceExpression(node) {
            if (node.expressions[0]?.type !== 'SequenceExpression') return;
            context.report({ node: node.expressions[0], message: 'replace inner sequence', fix: fixer => fixer.replaceText(node.expressions[0], 'selected') });
          },
          'ObjectPattern[optional=true], ArrayPattern[optional=true]'(node) {
            context.report({ node, message: 'replace optional binding', fix: fixer => fixer.replaceText(node, 'parameter?: T') });
          },
          BinaryExpression(node) {
            if (node.operator !== '>=') return;
            const operator = sc.getFirstTokenBetween(node.left, node.right);
            context.report({ node: operator, message: 'replace operator', fix: fixer => fixer.replaceText(operator, '>') });
          },
          Literal(node) {
            if (node.value === 'surrogate') {
              context.report({ node, message: 'replace literal', fix: fixer => fixer.replaceText(node, '"' + String.fromCharCode(0xd800) + '"') });
            } else if (node.value === 'a😀b') {
              context.report({ node, message: 'replace scalar half', fix: fixer => fixer.replaceTextRange([node.range[0] + 2, node.range[0] + 3], 'X') });
            }
          },
          MethodDefinition(node) {
            if (node.kind !== 'constructor' || node.key.type !== 'Identifier') return;
            context.report({ node: node.key, message: 'quote key', fix: fixer => fixer.replaceText(node.key, '"constructor"') });
          },
          JSXAttribute(node) {
            if (node.name.name !== 'value' || node.value.value === 'fixed') return;
            context.report({ node: node.value, message: 'replace attribute', fix: fixer => fixer.replaceText(node.value, '"fixed"') });
          },
          TemplateElement(node) {
            if (!node.value.raw.includes('\\n') || node.parent.type !== 'TemplateLiteral' || node.parent.expressions.length) return;
            context.report({ node: node.parent, message: 'replace template', fix: fixer => fixer.replaceText(node.parent, '\`fixed\`') });
          },
          ExportNamedDeclaration(node) {
            if (node.declaration?.declare && node.exportKind !== 'type') throw new Error('wrong ambient export kind');
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
  test('project false supplies compiler node maps and applies typescript-eslint assertion fixes', async () => {
    const plugin = import.meta.resolve('@typescript-eslint/eslint-plugin');
    const { dir, exitCode, diags } = await lint(
      {
        'rslint.config.mjs': `import tsPlugin from ${JSON.stringify(plugin)};
export default [{ files: ['**/*.ts', '**/*.tsx'], languageOptions: { parserOptions: { project: false } },
  plugins: { community: tsPlugin }, rules: {
    'community/consistent-type-imports': 'error',
    'community/naming-convention': ['error', { selector: 'variable', format: ['camelCase'] }],
    'community/consistent-type-assertions': ['error', { assertionStyle: 'as' }],
    'community/no-unused-vars': 'error'
  } }];`,
        'input.ts':
          'declare const input: unknown;\nexport const value = <number>input;\n',
        'fragment.tsx':
          'import { Fragment } from "react";\nexport const view = <><span /></>;\n',
      },
      ['--fix', 'input.ts', 'fragment.tsx'],
    );
    try {
      expect(diags).toEqual([]);
      expect(exitCode).toBe(0);
      expect(await fs.readFile(path.join(dir, 'input.ts'), 'utf8')).toBe(
        'declare const input: unknown;\nexport const value = input as number;\n',
      );
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('CLI runs real type-aware @typescript-eslint rules end-to-end', async () => {
    const plugin = import.meta.resolve('@typescript-eslint/eslint-plugin');
    const { dir, exitCode, diags } = await lint({
      'rslint.config.mjs': `import tsPlugin from ${JSON.stringify(plugin)};
export default [{
  files: ['**/*.ts'],
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
  plugins: { typed: tsPlugin },
  rules: {
    'typed/await-thenable': 'error',
    'typed/no-floating-promises': 'error',
  },
}];`,
      'tsconfig.json':
        '{"compilerOptions":{"target":"esnext","module":"esnext","strict":true},"include":["*.ts"]}',
      'dep.ts':
        'export async function fetchCount(): Promise<number> { return 1; }\n',
      'index.ts': [
        'import { fetchCount } from "./dep";',
        'async function run() {',
        '  await 123;',
        '  fetchCount();',
        '  await fetchCount();',
        '}',
        'void run();',
        '',
      ].join('\n'),
    });
    try {
      expect(exitCode).toBe(1);
      expect(
        diags
          .map((diagnostic) => [
            path.basename(diagnostic.filePath),
            diagnostic.ruleName,
          ])
          .sort(),
      ).toEqual([
        ['index.ts', 'typed/await-thenable'],
        ['index.ts', 'typed/no-floating-promises'],
      ]);
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('skips standard missing-type-info errors on gap files while keeping syntax rules and surfacing real config errors', async () => {
    const req = createRequire(import.meta.url);
    const tsPluginHref = pathToFileURL(
      req.resolve('@typescript-eslint/eslint-plugin'),
    ).href;
    const sonarjsHref = pathToFileURL(
      req.resolve('eslint-plugin-sonarjs'),
    ).href;
    const { dir, exitCode, diags } = await lint(
      {
        'rslint.config.mjs': `import tsPlugin from ${JSON.stringify(tsPluginHref)};
import sonarjs from ${JSON.stringify(sonarjsHref)};
export default [{
  files: ['**/*.ts'],
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
  plugins: { typed: tsPlugin, s: sonarjs },
  rules: {
    'typed/await-thenable': 'error',
    'typed/no-floating-promises': 'error',
    'typed/naming-convention': ['error', { selector: 'variable', types: ['boolean'], format: ['PascalCase'] }],
    'typed/consistent-type-imports': 'error',
    's/no-identical-expressions': 'error',
  },
}];`,
        'tsconfig.json':
          '{"compilerOptions":{"target":"esnext","module":"esnext","strict":true},"files":["typed.ts"]}',
        'typed.ts': 'export async function run() { await 123; }\n',
        'gap.ts':
          'import { Foo } from "./typed";\nexport type Bar = Foo;\nexport const same = 1 === 1;\n',
      },
      ['typed.ts', 'gap.ts'],
    );
    try {
      expect(exitCode).toBe(1);
      expect(
        diags
          .map((diagnostic) => [
            path.basename(diagnostic.filePath),
            diagnostic.ruleName,
          ])
          .sort(),
      ).toEqual([
        ['gap.ts', 's/no-identical-expressions'],
        ['gap.ts', 'typed/consistent-type-imports'],
        ['typed.ts', 'typed/await-thenable'],
      ]);
    } finally {
      await cleanupTempDir(dir);
    }

    // Real configuration errors and runtime bugs must still surface on gap files.
    const configErr = await lint(
      {
        'rslint.config.mjs': `const custom = {
  rules: {
    'missing-aliases': {
      meta: { docs: { requiresTypeChecking: true }, schema: [] },
      create(context) {
        const settings = context.settings;
        void context.sourceCode.parserServices.program;
        if (!settings.aliases) throw new Error('Missing path aliases');
        return {};
      },
    },
    'runtime-bug': {
      meta: { docs: { requiresTypeChecking: true }, schema: [] },
      create(context) {
        void context.sourceCode.parserServices.program;
        const missing = undefined;
        missing.boom();
        return {};
      },
    },
  },
};
export default [{
  files: ['**/*.ts'],
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
  plugins: { custom },
  rules: {
    'custom/missing-aliases': 'error',
    'custom/runtime-bug': 'error',
  },
}];`,
        'tsconfig.json':
          '{"compilerOptions":{"target":"esnext","module":"esnext","strict":true},"files":["typed.ts"]}',
        'typed.ts': 'export const a = 1;\n',
        'gap.ts': 'export const b = 2;\n',
      },
      ['gap.ts'],
    );
    try {
      expect(configErr.exitCode).toBe(1);
      expect(
        configErr.diags
          .map((diagnostic) => [
            path.basename(diagnostic.filePath),
            diagnostic.ruleName,
          ])
          .sort(),
      ).toEqual([
        ['gap.ts', 'custom/missing-aliases'],
        ['gap.ts', 'custom/runtime-bug'],
      ]);
    } finally {
      await cleanupTempDir(configErr.dir);
    }
  });

  test('rules on the same file share identical parserServices identity and tsconfig scope options', async () => {
    const { dir, exitCode, diags } = await lint(
      {
        'rslint.config.mjs': `const shared = new WeakMap();
const probe = {
  rules: {
    produce: {
      meta: { schema: [] },
      create(context) {
        const services = context.sourceCode.parserServices;
        return {
          Program() {
            shared.set(services, 'marker');
          },
        };
      },
    },
    consume: {
      meta: { schema: [] },
      create(context) {
        const services = context.sourceCode.parserServices;
        return {
          'Program:exit'(node) {
            const scope = context.sourceCode.getScope(node);
            const moduleScope = scope.childScopes[0] ?? scope;
            const hasH = moduleScope.variables.some((v) => v.name === 'h' && v.references.length > 0);
            if (shared.get(services) === 'marker' && hasH) {
              context.report({ node, message: 'shared-services-and-scope-ok' });
            }
          },
        };
      },
    },
  },
};
export default [{
  files: ['**/*.tsx'],
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
  plugins: { probe },
  rules: {
    'probe/produce': 'error',
    'probe/consume': 'error',
  },
}];`,
        'tsconfig.json': JSON.stringify({
          compilerOptions: {
            target: 'esnext',
            module: 'esnext',
            jsx: 'react',
            jsxFactory: 'h',
          },
          include: ['app.tsx'],
        }),
        'app.tsx': 'import { h } from "./jsx";\nexport const view = <div />;\n',
        'jsx.ts': 'export function h() { return null; }\n',
      },
      ['app.tsx'],
    );
    try {
      expect(exitCode).toBe(1);
      expect(
        diags.map((diagnostic) => [
          path.basename(diagnostic.filePath),
          diagnostic.ruleName,
        ]),
      ).toEqual([['app.tsx', 'probe/consume']]);
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('CLI preserves firstTokens, source paths, and syntax-only parserOptions flags', async () => {
    const { dir, exitCode, diags } = await lint({
      'rslint.config.mjs': `const probe = {
  rules: {
    check: {
      meta: { schema: [] },
      create(context) {
        const services = context.sourceCode.parserServices;
        return {
          Program(node) {
            if (!services.program) {
              if (services.experimentalDecorators !== true || services.emitDecoratorMetadata !== false) {
                throw new Error('syntax-only parserOptions flags were not forwarded');
              }
              context.report({ node, message: 'syntax-flags-ok' });
              return;
            }
            const source = services.esTreeNodeToTSNodeMap.get(node);
            const firstStmt = node.body[0];
            const tsStmt = services.esTreeNodeToTSNodeMap.get(firstStmt);
            if (!tsStmt || tsStmt.getStart(source) !== firstStmt.range[0]) {
              throw new Error('firstTokens projection missing over IPC');
            }
            const byPath = services.program.getSourceFileByPath(source.path);
            if (byPath !== source) {
              throw new Error('source.path lookup mismatch over IPC');
            }
            context.report({ node, message: 'typed-projection-ok' });
          },
        };
      },
    },
  },
};
export default [
  {
    files: ['typed.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    plugins: { probe },
    rules: { 'probe/check': 'error' },
  },
  {
    files: ['syntax.ts'],
    languageOptions: { parserOptions: { project: false, experimentalDecorators: true } },
    plugins: { probe },
    rules: { 'probe/check': 'error' },
  },
];`,
      'tsconfig.json':
        '{"compilerOptions":{"target":"esnext","module":"esnext","strict":true},"files":["typed.ts"]}',
      'typed.ts': '/* leading comment */\nexport const answer = 42;\n',
      'syntax.ts': 'export const plain = 1;\n',
    });
    try {
      expect(exitCode).toBe(1);
      const messages = diags as Array<Diag & { message: string }>;
      expect(
        messages
          .map((diagnostic) => [
            path.basename(diagnostic.filePath),
            diagnostic.message,
          ])
          .sort(),
      ).toEqual([
        ['syntax.ts', 'syntax-flags-ok'],
        ['typed.ts', 'typed-projection-ok'],
      ]);
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('typed plugins use each selected Program while a project reset stays source-only', async () => {
    const plugin = pathToFileURL(
      path.resolve(
        import.meta.dirname,
        '../../eslint-plugin/fixtures/typed-plugin.mjs',
      ),
    ).href;
    const { dir, exitCode, diags } = await lint({
      'rslint.config.mjs': `import typed from ${JSON.stringify(plugin)};
export default [
 { files: ['**/*.ts'], languageOptions: { parserOptions: { project: ['./tsconfig.json', './other/tsconfig.json'] } }, plugins: { typed }, rules: { 'typed/value': 'error' } },
 { files: ['plain.ts'], languageOptions: { parserOptions: { project: false } } },
];`,
      'tsconfig.json':
        '{"compilerOptions":{"target":"esnext","module":"esnext","strict":true},"files":["dependency.ts","typed.ts"]}',
      'dependency.ts': 'export const value: string = "😀";',
      'typed.ts':
        'import { value } from "./dependency";\n' +
        'declare const optional: { item?: string } | null;\n' +
        'optional?.item?.trim();\n1, 2, 3;\nvalue;\n' +
        'declare function choose<T extends string>(value: T): 1;\n' +
        'declare function choose<T extends string>(value: T, extra: T): 2;\n' +
        'choose(false,);\n',
      'other/tsconfig.json':
        '{"compilerOptions":{"target":"esnext","module":"esnext","strict":true},"files":["dependency.ts","other.ts"]}',
      'other/dependency.ts': 'export const value: number = 1;',
      'other/other.ts': 'import { value } from "./dependency";\nvalue;\n',
      'plain.ts': 'value;\n',
    });
    try {
      expect(exitCode).toBe(1);
      const messages = diags as Array<Diag & { message: string }>;
      expect(
        messages
          .map((diagnostic) => [
            path.basename(diagnostic.filePath),
            diagnostic.message,
          ])
          .sort(),
      ).toEqual([
        ['other.ts', 'number'],
        ['plain.ts', 'source-only'],
        ['typed.ts', 'ordered overload queries'],
        ['typed.ts', 'original sequence prefix'],
        ['typed.ts', 'string'],
        ['typed.ts', 'string | undefined'],
      ]);
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test('--fix runs plugins on a long comma expression', async () => {
    const text = 'a,'.repeat(5999) + '[1].filter(Boolean);\n';
    const { dir, exitCode, diags } = await lint(
      {
        'local-plugin.mjs': LOCAL_PLUGIN,
        'rslint.config.mjs': PLUGIN_CONFIG,
        'tsconfig.json': TSCONFIG,
        'sequence.ts': text,
      },
      ['--fix'],
    );
    try {
      expect(exitCode).toBe(0);
      expect(diags).toEqual([]);
      expect(await fs.readFile(path.join(dir, 'sequence.ts'), 'utf8')).toBe(
        text.replace('.filter(', '.some('),
      );
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test.each([
    ['no plugin', ''],
    [
      'disabled plugin rules',
      `{ files: ['**/*.ts'], plugins: { local }, rules: { 'local/no-null': 'off' } },`,
    ],
    [
      'unmatched plugin files',
      `{ files: ['**/other.ts'], plugins: { local }, rules: { 'local/no-null': 'error' } },`,
    ],
  ])('%s preserves native diagnostics', async (_label, pluginConfig) => {
    const { dir, exitCode, diags } = await lint({
      'local-plugin.mjs': LOCAL_PLUGIN,
      'rslint.config.mjs': `import local from './local-plugin.mjs';
export default [
  { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
  ${pluginConfig}
];`,
      'a.ts': 'debugger;\nconst value = null;\n',
    });
    try {
      expect(exitCode).toBe(1);
      expect(diags.map((diag) => diag.ruleName)).toEqual(['no-debugger']);
      expect(diags[0].range.start.line).toBe(1);
    } finally {
      await cleanupTempDir(dir);
    }
  });

  test.each([
    ['n', 'prefer-global/url'],
    ['custom', 'a/b/c'],
    ['@scope/plugin', 'rule'],
    ['@scope/plugin/a', 'b'],
  ])(
    'multi-slash ID %s/%s survives namespace collisions, flags, suppression and fixes',
    async (prefix, name) => {
      const fullName = `${prefix}/${name}`;
      const wrongSlash = fullName.startsWith('@')
        ? fullName.indexOf('/')
        : fullName.lastIndexOf('/');
      const wrongPrefix = fullName.slice(0, wrongSlash);
      const wrongName = fullName.slice(wrongSlash + 1);
      const { dir, exitCode, diags } = await lint(
        {
          'local-plugin.mjs': LOCAL_PLUGIN,
          'rslint.config.mjs': `import base from './local-plugin.mjs';
export default [{ files: ['**/*.ts'],
  plugins: {
    ${JSON.stringify(prefix)}: { rules: { ${JSON.stringify(name)}: base.rules['prefer-array-some'] } },
    // This alternative split produces the same full ID but must not replace its owner.
    ${JSON.stringify(wrongPrefix)}: { rules: { ${JSON.stringify(wrongName)}: base.rules['no-null'] } },
  },
  rules: { ${JSON.stringify(fullName)}: 'off' },
}];`,
          'a.ts': `[1].filter(Boolean);\n// rslint-disable-next-line ${fullName}\n[2].filter(Boolean);\n`,
        },
        ['--rule', `${fullName}: error`],
      );
      try {
        expect(exitCode).toBe(1);
        expect(diags).toHaveLength(1);
        expect(diags[0].ruleName).toBe(fullName);
        expect(diags[0].range.start.line).toBe(1);
        const fixed = await runRslint(
          ['--fix', '--rule', `${fullName}: error`],
          dir,
        );
        expect(fixed.exitCode).toBe(0);
        expect(await fs.readFile(path.join(dir, 'a.ts'), 'utf8')).toBe(
          `[1].some(Boolean);\n// rslint-disable-next-line ${fullName}\n[2].filter(Boolean);\n`,
        );
      } finally {
        await cleanupTempDir(dir);
      }
    },
  );

  test('ordinary CLI preserves the JSX Fragment import', async () => {
    const plugin = import.meta.resolve('@typescript-eslint/eslint-plugin');
    const { dir, exitCode, diags } = await lint(
      {
        'rslint.config.mjs': `import community from ${JSON.stringify(plugin)};
export default [{ files: ['**/*.tsx'], plugins: { community }, rules: { 'community/no-unused-vars': 'error' } }];`,
        'tsconfig.json': JSON.stringify({
          compilerOptions: {
            jsx: 'react',
            jsxFactory: 'React.createElement',
            jsxFragmentFactory: 'Fragment',
          },
        }),
        'factory.ts': 'export declare const React: any, Fragment: any;',
        'example.tsx':
          "import { React, Fragment } from './factory'; export const view = <><div /></>;",
      },
      ['example.tsx'],
    );
    try {
      expect(diags).toEqual([]);
      expect(exitCode).toBe(0);
    } finally {
      await cleanupTempDir(dir);
    }
  });

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

  test.each([
    [
      'binding.ts',
      '\ufeff// 😀\r\nconst {source: local = 1, nested: {value = 2} = fallback} = input;\r\n',
      '\ufeff// 😀\r\nconst {source: local, nested: {value}} = input;\r\n',
    ],
    [
      'operator.ts',
      '// 😀\nconst value = left >= right; left >>= 1;\n',
      '// 😀\nconst value = left > right; left >>= 1;\n',
    ],
    ['sequence.js', '// 😀\n(a, b), c;\n', '// 😀\n(selected), c;\n'],
    ['sequence.ts', '// 😀\n(a, b), c;\n', '// 😀\n(selected), c;\n'],
    [
      'optional.ts',
      'type F = ({value}?: T) => void;\n',
      'type F = (parameter?: T) => void;\n',
    ],
    [
      'optional.tsx',
      'type F = ([value]?: T) => void;\n',
      'type F = (parameter?: T) => void;\n',
    ],
    [
      'constructor.ts',
      '// 😀\nclass C { private /* constructor */ constructor() {} static constructor() {} }\n',
      '// 😀\nclass C { private /* constructor */ "constructor"() {} static constructor() {} }\n',
    ],
    [
      'attribute.tsx',
      'const element = <A value="\\" />; /* after */ next();\n',
      'const element = <A value="fixed" />; /* after */ next();\n',
    ],
    [
      'attribute.jsx',
      'const element = <A value="first\r\nsecond" />; next();\n',
      'const element = <A value="fixed" />; next();\n',
    ],
    [
      'template.js',
      '// 😀\nconst value = `first\r\nlast`;\n',
      '// 😀\nconst value = `fixed`;\n',
    ],
    [
      'ambient.ts',
      'export declare const value: T; export declare function f<T>(x: T): T;\n',
      'export declare const value: T; export declare function f<T>(x: T): T;\n',
    ],
    [
      'surrogate.ts',
      'const value = "surrogate";\n',
      'const value = "\ud800";\n',
    ],
    ['scalar.ts', 'const value = "a😀b";\n', 'const value = "aX\ude00b";\n'],
  ])(
    'ESTree ranges survive the complete --fix pipeline: %s',
    async (filePath, text, expected) => {
      const dir = await createTempDir({
        'local-plugin.mjs': LOCAL_PLUGIN,
        'rslint.config.mjs': `import local from './local-plugin.mjs'; export default [{files: ['**/*.{js,jsx,ts,tsx}'], plugins: {local}, rules: {'local/projection-probe': 'error'}}];`,
        'tsconfig.json': JSON.stringify({
          compilerOptions: {
            target: 'ESNext',
            module: 'ESNext',
            allowJs: true,
            jsx: 'preserve',
          },
          include: ['**/*'],
        }),
        [filePath]: text,
      });
      try {
        const { default: plugin } = await import(
          /* webpackIgnore: true */ pathToFileURL(
            path.join(dir, 'local-plugin.mjs'),
          ).href
        );
        const upstream = new Linter().verifyAndFix(
          text,
          [
            {
              files: ['**/*.{js,jsx,ts,tsx}'],
              languageOptions: {
                parser: typescriptParser,
                sourceType: 'module',
                ecmaVersion: 'latest',
                parserOptions: { ecmaFeatures: { jsx: true } },
              },
              plugins: { local: plugin },
              rules: { 'local/projection-probe': 'error' },
            },
          ],
          { filename: filePath },
        );
        expect(upstream.messages).toEqual([]);
        expect(upstream.output).toBe(expected);
        const result = await runRslint(['--fix'], dir);
        expect(result.exitCode).toBe(0);
        expect(await fs.readFile(path.join(dir, filePath))).toEqual(
          Buffer.from(upstream.output),
        );
      } finally {
        await cleanupTempDir(dir);
      }
    },
  );

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
