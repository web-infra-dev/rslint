const fs = require('node:fs');
const path = require('node:path');
const reference = path.resolve(process.argv[2]);
require('node:assert/strict').equal(
  require(path.join(reference, 'package.json')).version,
  '10.14.0',
);
const { Linter } = require(reference);
const fixer = require(path.join(reference, 'lib/linter/source-code-fixer.js'));
const cases = [
  ['wildcard', '/* eslint-disable */\nhit();', []],
  ['repeated', '/* eslint-disable a */\n/* eslint-disable a */\nhit();', ['a']],
  [
    'block before line',
    '/* eslint-disable a */\nhit(); // eslint-disable-line a',
    ['a'],
  ],
  [
    'next plus line',
    '// eslint-disable-next-line a\nhit(); // eslint-disable-line a',
    ['a'],
  ],
  ['partial first', '/* eslint-disable a, b -- reason */\nhit();', ['b']],
  [
    'partial middle',
    '/* eslint-disable a , b , c -- reason */\nhit();',
    ['a', 'c'],
  ],
  ['partial last', '/* eslint-disable a, b -- reason */\nhit();', ['a']],
  ['quoted', '/* eslint-disable "a", \'b\' -- reason */\nhit();', ['b']],
  ['enable without disable', '/* eslint-enable a */\nhit();', []],
  [
    'repeated enables',
    '/* eslint-disable a */\nhit();\n/* eslint-enable a */\n/* eslint-enable a */',
    ['a'],
  ],
  [
    'wildcard enable',
    '/* eslint-disable a */\nhit();\n/* eslint-enable */',
    ['a'],
  ],
  [
    'enable specific under wildcard',
    '/* eslint-disable */\nhit();\n/* eslint-enable a */\n/* eslint-enable */',
    ['a'],
  ],
  ['multiline next line', '/* eslint-disable-next-line\na */\nhit();', ['a']],
  ['unicode CRLF', '// 中文 😀\r\n/* eslint-disable a, b */\r\nhit();', ['b']],
  ['adjacent tokens', 'let/* eslint-disable a */x;', []],
];
// ESLint requires plugin keys for custom flat-config rules. Rewrite fixture
// rule identities consistently before evaluating; Go tests use those same keys.
const output = cases.map(([name, original, reports]) => {
  const source = original.replace(/\b[abc]\b/g, (id) => `test/${id}`);
  const linter = new Linter();
  const rules = Object.fromEntries(
    ['a', 'b', 'c'].map((id) => [
      id,
      {
        meta: { schema: [] },
        create(context) {
          return {
            ExpressionStatement(node) {
              if (reports.includes(id)) context.report({ node, message: id });
            },
          };
        },
      },
    ]),
  );
  const messages = linter.verify(
    source,
    [
      {
        plugins: { test: { rules } },
        rules: Object.fromEntries(
          ['a', 'b', 'c'].map((id) => [`test/${id}`, 'error']),
        ),
        linterOptions: { reportUnusedDisableDirectives: 'warn' },
      },
    ],
    { filename: 'test.js' },
  );
  return {
    name,
    source,
    reports: reports.map((id) => `test/${id}`),
    diagnostics: messages
      .filter((m) => m.ruleId === null)
      .map((m) => ({
        message: m.message,
        line: m.line,
        column: m.column,
        severity: m.severity,
        fix: m.fix,
      })),
    fixed: fixer.applyFixes(source, messages).output,
  };
});
fs.writeFileSync(
  path.join(__dirname, '../internal/rule/testdata/unused-directives.json'),
  JSON.stringify(output, null, 2) + '\n',
);
