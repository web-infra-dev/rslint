// Run against the freshly built addon:
// RSLINT_NATIVE_BINDING=/absolute/path/to/rslint.node node --test crates/rslint-native/tests/rust-lint.test.mjs
import assert from 'node:assert/strict';
import fs from 'node:fs';
import { createRequire } from 'node:module';
import os from 'node:os';
import path from 'node:path';
import { test } from 'node:test';

const bindingPath = process.env.RSLINT_NATIVE_BINDING;
assert.ok(bindingPath, 'RSLINT_NATIVE_BINDING must name the addon under test');
assert.ok(
  path.isAbsolute(bindingPath),
  'RSLINT_NATIVE_BINDING must be absolute',
);
const binding = createRequire(import.meta.url)(bindingPath);

test('the addon exposes the direct Rust lint entry point', () => {
  assert.equal(typeof binding.lintTsrsSync, 'function');
});

test('invalid lint requests become JavaScript exceptions', () => {
  assert.throws(
    () => binding.lintTsrsSync('{'),
    (error) => error instanceof Error && error.message.length > 0,
  );
});

function fixture(t) {
  const cwd = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-native-tsrs-'));
  t.after(() => fs.rmSync(cwd, { recursive: true, force: true }));
  const files = [
    {
      path: path.join(cwd, 'values.ts'),
      text: 'export let bad: string = "wrong"; export let good: number = 1;\n',
    },
    {
      path: path.join(cwd, 'main.ts'),
      text: '// 🦀\nimport { bad, good } from "./values";\n-bad;\n-good;\ndebugger;\n',
    },
  ];
  for (const file of files) fs.writeFileSync(file.path, file.text);
  return {
    schemaVersion: 1,
    cwd,
    files,
    rules: [
      { name: 'no-debugger', severity: 1 },
      { name: '@typescript-eslint/no-unsafe-unary-minus', severity: 2 },
    ],
  };
}

function lint(request) {
  const raw = binding.lintTsrsSync(JSON.stringify(request));
  assert.equal(typeof raw, 'string');
  const result = JSON.parse(raw);
  assert.equal(result.schemaVersion, 1);
  assert.equal(result.backend, 'tsrs-rust');
  assert.equal(
    result.compiler.revision,
    '280b426b89cd4ecd6153cf48228826bd60a99c61',
  );
  assert.deepEqual(
    result.files.map((file) => file.filePath).sort(),
    request.files.map((file) => file.path).sort(),
    'Silent files must remain in the admitted-file result',
  );
  for (const file of result.files) {
    assert.deepEqual(
      file.appliedRules,
      request.rules.map((rule) => rule.name),
    );
  }
  return result;
}

test('the Rust compiler resolves cross-file types and reports UTF-16 ranges', (t) => {
  const request = fixture(t);
  const result = lint(request);
  const source = request.files[1].text;
  const main = result.files.find(
    (file) => file.filePath === request.files[1].path,
  );
  const values = result.files.find(
    (file) => file.filePath === request.files[0].path,
  );
  assert.deepEqual(values.messages, []);
  assert.deepEqual(
    main.messages.toSorted((a, b) => a.range[0] - b.range[0]),
    [
      {
        ruleId: '@typescript-eslint/no-unsafe-unary-minus',
        messageId: 'unaryMinus',
        message:
          'Argument of unary negation should be assignable to number | bigint but is string instead.',
        severity: 2,
        line: 3,
        column: 1,
        endLine: 3,
        endColumn: 5,
        range: [source.indexOf('-bad'), source.indexOf('-bad') + 4],
      },
      {
        ruleId: 'no-debugger',
        messageId: 'no-debugger',
        message: "Unexpected 'debugger' statement.",
        severity: 1,
        line: 5,
        column: 1,
        endLine: 5,
        endColumn: 10,
        range: [source.indexOf('debugger;'), source.indexOf('debugger;') + 9],
      },
    ],
  );

  // Keep disk contents unchanged: the next compiler must use the new snapshot,
  // including an imported file, rather than a stale project or the disk source.
  request.files[0].text =
    'export let bad: number = 1; export let good: number = 1;\n';
  const changed = lint(request);
  assert.deepEqual(
    changed.files.flatMap((file) =>
      file.messages.map((message) => message.ruleId),
    ),
    ['no-debugger'],
  );
});

test('unsupported rules and duplicate files fail instead of reducing work', (t) => {
  const request = fixture(t);
  assert.throws(() =>
    binding.lintTsrsSync(
      JSON.stringify({
        ...request,
        rules: [{ name: 'not-a-supported-rule', severity: 2 }],
      }),
    ),
  );
  assert.throws(() =>
    binding.lintTsrsSync(
      JSON.stringify({
        ...request,
        files: [...request.files, request.files[0]],
      }),
    ),
  );
});
