import assert from 'node:assert/strict';
import path from 'node:path';
import { Rslint, type LintResult, type RslintConfigEntry } from '@rslint/core';
import { RuleTester } from '../rule-tester';

// Upstream: https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/no-lonely-if.js
// The Go suite also verifies exact ranges and every fix pass.
const ruleTester = new RuleTester();

ruleTester.run('no-lonely-if', {} as never, {
  valid: [
    {
      name: 'snapshots',
      code: 'if (a) {\n\tif (b) {\n\t}\n} else {}',
    },
    {
      name: 'snapshots',
      code: 'if (a) {\n\tif (b) {\n\t}\n\tfoo();\n} else {}',
    },
    {
      name: 'snapshots',
      code: 'if (a) {\n} else {\n\tif (y) {}\n}',
    },
    {
      name: 'snapshots',
      code: 'if (a) {\n\tb ? c() : d()\n}',
    },
    {
      name: 'docs',
      code: 'if (foo && bar) {\n\t// …\n}',
    },
    {
      name: 'docs',
      code: 'if (foo) {\n\t// …\n} else if (bar && baz) {\n\t// …\n}',
    },
    {
      name: 'docs',
      code: 'if (foo) {\n\t// …\n} else if (bar) {\n\tif (baz) {\n\t\t// …\n\t}\n} else {\n\t// …\n}',
    },
    {
      name: 'docs',
      code: '// Built-in rule `no-lonely-if` case https://eslint.org/docs/rules/no-lonely-if\nif (foo) {\n\t// …\n} else {\n\tif (bar) {\n\t\t// …\n\t}\n}',
    },
  ],
  invalid: [
    {
      name: 'snapshot 1',
      code: 'if (a) {\n\tif (b) {\n\t}\n}',
      output: 'if (a && b) {\n\t}',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 3,
          endColumn: 3,
        },
      ],
    },
    {
      name: 'snapshot 2',
      code: 'if (a) if (b) {\n\tfoo();\n}',
      output: 'if (a && b) {\n\tfoo();\n}',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 1,
          column: 8,
          endLine: 3,
          endColumn: 2,
        },
      ],
    },
    {
      name: 'snapshot 3',
      code: 'if (a) {\n\tif (b) foo();\n}',
      output: 'if (a && b) foo();',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'snapshot 4',
      code: 'if (a) /* comment */ {\n\tif (b) foo();\n}',
      output: '/* comment */ if (a && b) foo();',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'snapshot 5',
      code: 'if (a) if (b) foo();',
      output: 'if (a && b) foo();',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'snapshot 6',
      code: 'if (a) {\n\tif (b) foo()\n}',
      output: 'if (a && b) foo()',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'snapshot 7',
      code: 'if (a) if (b);',
      output: 'if (a && b);',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'snapshot 8',
      code: 'if (a) {\n\tif (b) {\n\t\t// Should not report\n\t}\n} else if (c) {\n\tif (d) {\n\t}\n}',
      output:
        'if (a) {\n\tif (b) {\n\t\t// Should not report\n\t}\n} else if (c && d) {\n\t}',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 6,
          column: 2,
          endLine: 7,
          endColumn: 3,
        },
      ],
    },
    {
      name: 'snapshot 9',
      code: 'function * foo() {\n\tif (a || b)\n\tif (a ?? b)\n\tif (a ? b : c)\n\tif (a = b)\n\tif (a += b)\n\tif (a -= b)\n\tif (a &&= b)\n\tif (yield a)\n\tif (a, b);\n}',
      output:
        'function * foo() {\n\tif ((a || b) && (a ?? b) && (a ? b : c) && (a = b) && (a += b) && (a -= b) && (a &&= b) && (yield a) && (a, b));\n}',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 3,
          column: 2,
          endLine: 10,
          endColumn: 12,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 4,
          column: 2,
          endLine: 10,
          endColumn: 12,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 5,
          column: 2,
          endLine: 10,
          endColumn: 12,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 6,
          column: 2,
          endLine: 10,
          endColumn: 12,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 7,
          column: 2,
          endLine: 10,
          endColumn: 12,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 8,
          column: 2,
          endLine: 10,
          endColumn: 12,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 9,
          column: 2,
          endLine: 10,
          endColumn: 12,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 10,
          column: 2,
          endLine: 10,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'snapshot 10',
      code: 'async function foo() {\n\tif (a)\n\tif (await a)\n\tif (a.b)\n\tif (a && b);\n}',
      output:
        'async function foo() {\n\tif (a && await a && a.b && a && b);\n}',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 3,
          column: 2,
          endLine: 5,
          endColumn: 14,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 4,
          column: 2,
          endLine: 5,
          endColumn: 14,
        },
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 5,
          column: 2,
          endLine: 5,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'snapshot 11',
      code: 'if (((a || b))) if (((c || d)));',
      output: 'if (((a || b)) && ((c || d)));',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'snapshot 12',
      code: 'if // 1\n(\n\t// 2\n\ta // 3\n\t\t.b // 4\n) // 5\n{\n\t// 6\n\tif (\n\t\t// 7\n\t\tc // 8\n\t\t\t.d // 9\n\t) {\n\t\t// 10\n\t\tfoo();\n\t\t// 11\n\t}\n\t// 12\n}',
      output:
        '// 6\n\tif // 1\n(\n\t// 2\n\ta // 3\n\t\t.b // 4\n && \n\t\t// 7\n\t\tc // 8\n\t\t\t.d // 9\n\t) // 5\n {\n\t\t// 10\n\t\tfoo();\n\t\t// 11\n\t\n\t// 12\n}',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 9,
          column: 2,
          endLine: 17,
          endColumn: 3,
        },
      ],
    },
    {
      name: 'snapshot 13',
      code: 'if (a) {\n\tif (b) foo()\n}\n[].forEach(bar)',
      output: 'if (a && b) foo();\n[].forEach(bar)',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'snapshot 14',
      code: 'if (a)\n\tif (b) foo()\n;[].forEach(bar)',
      output: 'if (a && b) foo()\n;[].forEach(bar)',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 3,
          endColumn: 2,
        },
      ],
    },
    {
      name: 'snapshot 15',
      code: 'if (a) {\n\tif (b) foo()\n}\n;[].forEach(bar)',
      output: 'if (a && b) foo()\n;[].forEach(bar)',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'snapshot 16',
      code: 'if (a) /* comment */ {\n\tif (b) foo()\n}',
      output: '/* comment */ if (a && b) foo()',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'documentation 1',
      code: 'if (foo) {\n\tif (bar) {\n\t\t// …\n\t}\n}',
      output: 'if (foo && bar) {\n\t\t// …\n\t}',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 2,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      name: 'documentation 2',
      code: 'if (foo) {\n\t// …\n} else if (bar) {\n\tif (baz) {\n\t\t// …\n\t}\n}',
      output: 'if (foo) {\n\t// …\n} else if (bar && baz) {\n\t\t// …\n\t}',
      errors: [
        {
          messageId: 'no-lonely-if',
          message:
            'Unexpected `if` as the only statement in a `if` block without `else`.',
          line: 4,
          column: 2,
          endLine: 6,
          endColumn: 3,
        },
      ],
    },
  ],
});

// Preserve the upstream integration regressions, including companion rules.
function createLinter(rules: RslintConfigEntry['rules'] = {}) {
  return new Rslint({
    overrideConfigFile: path.resolve(
      import.meta.dirname,
      '../testdata/no-lonely-if.config.mjs',
    ),
    overrideConfig: { rules },
    fix: true,
  });
}

async function lintFixture(
  code: string,
  rules: RslintConfigEntry['rules'] = {},
) {
  const [result] = await createLinter(rules).lintText(code, {
    filePath: 'fixture.js',
  });
  return result;
}

const hasRuleMessage = (result: LintResult, ruleId: string) =>
  result.messages.some((message) => message.ruleId === ruleId);

test('fix should not produce invalid code when another rule replaces the original range', async () => {
  const code =
    "function some(value) {\n    if (value < 10) {\n        if (value < 5) {\n            console.log('this is a long string this is a long string this is a long string this is a long string', value);\n        }\n    }\n    return 0\n}";

  const eslint = createLinter({ 'fake/format': 'error' });

  const [result] = await eslint.lintText(code, { filePath: 'fixture.js' });
  const fatalMessages = result.messages.filter(
    (message) => message.ruleId === null,
  );

  assert.strictEqual(fatalMessages.length > 0, false);
  assert.strictEqual(
    result.output,
    "function some(value) {\n  if (value < 10 && value < 5) {\n    console.log(\n      'this is a long string this is a long string this is a long string this is a long string',\n      value,\n    );\n  }\n  return 0;\n}",
  );
});

test('fix should preserve text between the outer condition and block', async () => {
  const result = await lintFixture(
    'if (a) /* comment */ { if (b) { foo(); } }',
  );

  assert.strictEqual(result.output, '/* comment */ if (a && b) { foo(); }');
});

test('fix should preserve text between the outer condition and non-block consequent', async () => {
  const result = await lintFixture('if (a) /* comment */ { if (b) foo(); }');

  assert.strictEqual(result.output, '/* comment */ if (a && b) foo();');
});

test('fix should preserve comments before the inner if inside the outer block', async () => {
  const result = await lintFixture('if (a) { /* before */ if (b) foo(); }');

  assert.strictEqual(result.output, '/* before */ if (a && b) foo();');
});

test('fix should preserve comments after the inner if inside the outer block', async () => {
  const result = await lintFixture('if (a) { if (b) foo(); /* after */ }');

  assert.strictEqual(result.output, 'if (a && b) foo(); /* after */ ');
});

test('fix should keep pragma comments from before the inner if attached to the merged if', async () => {
  const result = await lintFixture(
    'if (a) { /* @keep-next */ if (b) foo(); }',
    { 'fake/pragma-attachment': 'error' },
  );

  assert.strictEqual(result.output, '/* @keep-next */ if (a && b) foo();');
  assert.strictEqual(hasRuleMessage(result, 'fake/pragma-attachment'), false);
});

test('fix should keep pragma comments from the outer condition gap attached to the merged if', async () => {
  const result = await lintFixture(
    'if (a) /* @keep-next */ { if (b) foo(); }',
    { 'fake/pragma-attachment': 'error' },
  );

  assert.strictEqual(result.output, '/* @keep-next */ if (a && b) foo();');
  assert.strictEqual(hasRuleMessage(result, 'fake/pragma-attachment'), false);
});

test('fix should preserve eslint-disable-next-line before the inner if', async () => {
  const result = await lintFixture(
    "if (a) {\n\t// eslint-disable-next-line no-console\n\tif (b) console.log('foo');\n}",
    { 'no-console': 'error' },
  );

  assert.strictEqual(
    result.output,
    "// eslint-disable-next-line no-console\n\tif (a && b) console.log('foo');",
  );
  assert.strictEqual(hasRuleMessage(result, 'no-console'), false);
});

test('fix should preserve block eslint-disable-next-line before the inner if', async () => {
  const result = await lintFixture(
    "if (a) {\n\t/* eslint-disable-next-line no-console */\n\tif (b) console.log('foo');\n}",
    { 'no-console': 'error' },
  );

  assert.strictEqual(
    result.output,
    "/* eslint-disable-next-line no-console */\n\tif (a && b) console.log('foo');",
  );
  assert.strictEqual(hasRuleMessage(result, 'no-console'), false);
});

test('fix should preserve eslint-disable-line between the outer condition and block', async () => {
  const result = await lintFixture(
    'if (true) // eslint-disable-line no-constant-condition\n{\n\tif (true) foo();\n}',
    { 'no-constant-condition': 'error' },
  );

  assert.strictEqual(
    result.output,
    'if (true && true) // eslint-disable-line no-constant-condition\n foo();',
  );
  assert.strictEqual(hasRuleMessage(result, 'no-constant-condition'), false);
});

test('fix should preserve comments inside merged conditions', async () => {
  const result = await lintFixture(
    'if (/* outer */ a) { if (b /* inner */) foo(); }',
  );

  assert.strictEqual(
    result.output,
    'if (/* outer */ a && b /* inner */) foo();',
  );
});

test('fix should preserve comments between if and opening parenthesis', async () => {
  const result = await lintFixture('if/* outer */(a) if/* inner */(b) foo();');

  assert.strictEqual(result.output, 'if/* outer */(a && /* inner */b) foo();');
});

test('fix should preserve ASI-safe semicolon insertion when keeping outer-gap text', async () => {
  const result = await lintFixture(
    'if (a) /* comment */ { if (b) foo() } [].forEach(bar)',
  );

  assert.strictEqual(
    result.output,
    '/* comment */ if (a && b) foo();[].forEach(bar)',
  );
  assert.strictEqual(
    result.messages.some((message) => message.ruleId === null),
    false,
  );
});

test('fix should preserve ASI-safe semicolon insertion when keeping trailing text from the outer block', async () => {
  const result = await lintFixture(
    'if (a) { if (b) foo() /* after */ } [].forEach(bar)',
  );

  assert.strictEqual(
    result.output,
    'if (a && b) foo() /* after */ ;[].forEach(bar)',
  );
  assert.strictEqual(
    result.messages.some((message) => message.ruleId === null),
    false,
  );
});
