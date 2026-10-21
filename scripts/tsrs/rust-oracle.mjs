// Targeted compatibility oracle. The original engine must live outside this repo.
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { createRequire } from 'node:module';
import { execFileSync } from 'node:child_process';
import { pathToFileURL } from 'node:url';

const [baselineArg, bindingArg, outputArg] = process.argv.slice(2);
assert(
  baselineArg && bindingArg && outputArg,
  'usage: node scripts/tsrs/rust-oracle.mjs EXTERNAL_BASELINE ADDON OUTPUT',
);
const baseline = fs.realpathSync(baselineArg);
const bindingPath = fs.realpathSync(bindingArg);
const output = path.resolve(outputArg);
const revision = execFileSync('git', ['-C', baseline, 'rev-parse', 'HEAD'], {
  encoding: 'utf8',
}).trim();
assert.equal(revision, 'e0a29f751d4def86d27931759098f27748fa3003');
assert.equal(
  execFileSync('git', ['-C', baseline, 'diff', '--name-only', 'HEAD'], {
    encoding: 'utf8',
  }),
  '',
);
const { Rslint } = await import(
  pathToFileURL(path.join(baseline, 'packages/rslint/dist/index.js')).href
);
const binding = createRequire(import.meta.url)(bindingPath);
const cwd = fs.realpathSync(
  fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rust-oracle-')),
);
const sources = {
  'types.ts':
    'export declare const bad: string; export declare const good: number; export enum External { A, B }\n',
  'main.ts': `// 🦀\nimport { bad, good } from './types';\n-bad; -good; debugger;\nfunction f() { arguments.caller; arguments.callee; }\n0 === -0; -0 !== 1;\nconst message = '\u0024{value}';\nconst x = true ? false ? 1 : 2 : 3;\n`,
  'clean.ts': 'export const safe = -1; const message = `${1}`;\n',
  'flow.ts': `async function run() { for (let i = 0; i < 2; i++) { await Promise.resolve(); continue; } }
function cleanup() { try { return 1; } finally { return 2; } }
class Constructor { constructor() { return {}; } }
try { run(); } catch (error) { throw error; }
if (!Boolean(1)) { run(); } else { cleanup(); }
`,
  'enums.ts': `import { External } from './types';
enum Mixed { A = 1, B = 'b' }
enum Other { A, B }
External.A === Other.A;
class Accessors { get value(): string { return ''; } set value(input: number) {} }
`,
  'with.js': 'with (environment) { work(); }\n',
  'bom.ts': "\uFEFFdebugger;\r\n'😀'; debugger;\n",
};
const names = [
  'no-debugger',
  'no-caller',
  'no-compare-neg-zero',
  'no-template-curly-in-string',
  'no-ternary',
  'no-nested-ternary',
  'no-with',
  '@typescript-eslint/no-unsafe-unary-minus',
  'no-continue',
  'no-await-in-loop',
  'no-unsafe-finally',
  'no-constructor-return',
  'no-useless-catch',
  'no-negated-condition',
  '@typescript-eslint/no-mixed-enums',
  '@typescript-eslint/no-unsafe-enum-comparison',
  '@typescript-eslint/related-getter-setter-pairs',
];
const rules = names.map((name, index) => ({
  name,
  severity: index % 2 ? 1 : 2,
}));
const config = [
  {
    files: ['**/*.ts', '**/*.js'],
    plugins: ['@typescript-eslint'],
    languageOptions: {
      parserOptions: { project: path.join(cwd, 'tsconfig.json') },
    },
    rules: Object.fromEntries(
      rules.map(({ name, severity }) => [name, severity]),
    ),
  },
];
const report = {
  schemaVersion: 1,
  status: 'running',
  scope: 'targeted rule parity; not complete migration or performance',
  baselineRevision: revision,
  compilerRevision: '280b426b89cd4ecd6153cf48228826bd60a99c61',
  nodeVersion: process.version,
  addonSha256: crypto
    .createHash('sha256')
    .update(fs.readFileSync(bindingPath))
    .digest('hex'),
  sources,
  rules,
};
function save() {
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, `${JSON.stringify(report, null, 2)}\n`);
}
function normalize(results) {
  return results
    .map((file) => ({
      filePath: path.relative(cwd, file.filePath).split(path.sep).join('/'),
      messages: file.messages
        .map(({ range, ...message }) => message)
        .sort(
          (a, b) =>
            a.line - b.line ||
            a.column - b.column ||
            (a.endLine ?? 0) - (b.endLine ?? 0) ||
            (a.endColumn ?? 0) - (b.endColumn ?? 0) ||
            a.ruleId.localeCompare(b.ruleId),
        ),
    }))
    .sort((a, b) => a.filePath.localeCompare(b.filePath));
}
try {
  for (const [name, text] of Object.entries(sources))
    fs.writeFileSync(path.join(cwd, name), text);
  fs.writeFileSync(
    path.join(cwd, 'tsconfig.json'),
    JSON.stringify({
      compilerOptions: {
        strict: true,
        allowJs: true,
        target: 'ES2022',
        module: 'ESNext',
        moduleResolution: 'Bundler',
      },
      files: Object.keys(sources),
    }),
  );
  const linter = new Rslint({
    cwd,
    overrideConfigFile: true,
    overrideConfig: config,
  });
  let expected;
  try {
    expected = await linter.lintFiles(Object.keys(sources));
  } finally {
    await linter.close();
  }
  report.baseline = normalize(expected);
  const request = {
    schemaVersion: 1,
    cwd,
    project: './tsconfig.json',
    files: Object.keys(sources).map((name) => ({ path: name })),
    rules,
  };
  const actual = JSON.parse(binding.lintTsrsSync(JSON.stringify(request)));
  assert.equal(actual.backend, 'tsrs-rust');
  report.candidate = normalize(actual.files);
  for (const file of actual.files) assert.deepEqual(file.appliedRules, names);
  assert.deepEqual(
    report.candidate.map((file) => file.filePath).sort(),
    Object.keys(sources).sort(),
  );
  assert.deepEqual(
    report.baseline.map((file) => file.filePath).sort(),
    Object.keys(sources).sort(),
  );
  assert.deepEqual(report.candidate, report.baseline);
  assert.deepEqual(
    [
      ...new Set(
        report.candidate.flatMap((file) => file.messages.map((m) => m.ruleId)),
      ),
    ].sort(),
    [...names].sort(),
    'Every enabled rule must produce a positive sentinel in this targeted fixture',
  );
  assert.equal(
    report.candidate.find((file) => file.filePath === 'clean.ts').messages
      .length,
    0,
  );
  report.status = 'passed';
} catch (error) {
  report.status = 'failed';
  report.error = String(error).replaceAll(cwd, '<fixture>');
  process.exitCode = 1;
} finally {
  save();
  fs.rmSync(cwd, { recursive: true, force: true });
}
console.log(`${report.status}: ${outputArg}`);
