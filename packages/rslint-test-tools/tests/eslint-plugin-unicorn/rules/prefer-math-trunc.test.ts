import path from 'node:path';
import { runInNewContext } from 'node:vm';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-math-trunc.js
const ruleName = 'prefer-math-trunc';
const ruleTester = new RuleTester();

const jsFilename = 'src/virtual.js';

function jsCase(code: string, ...messageIds: string[]) {
  return {
    code,
    filename: jsFilename,
    errors: messageIds.map((messageId) => ({ messageId })),
  };
}

function tsCase(code: string, ...messageIds: string[]) {
  return { ...jsCase(code, ...messageIds), filename: 'src/virtual.ts' };
}

// The suite wrapper asserts message IDs. The Go suite checks ranges and edits
// for every case, and the runtime tests below exercise suggestions over IPC.
ruleTester.run(ruleName, {} as never, {
  valid: [
    // Upstream snapshots
    'const foo = 1 | 1;',
    'const foo = 0 | 1;',
    'const foo = 1.4 | +0;',
    'const foo = 1.4 | -0;',
    'const foo = 1.4 | (.5 - 0.5);',
    'const foo = 1.4 & 0xFFFFFFFF',
    'const foo = 1.4 & 0xFF',
    'const foo = 1.4 & 0x0',
    'const foo = 1.4 & 0',
    'const foo = ~3.9;',
    'const foo = 1.1 >> 1',
    'const foo = 0 << 1',
    `let foo = 0;
foo |= 1;`,
    `let foo = 1.2; // comment 1
foo |= 1; // comment 2 and 1.2 | 0`,
    'parseInt(value, 10);',
    'parseInt(String(value));',
    'parseInt(String(value), 16);',
    'Number.parseInt(String(value), 16);',
    'parseInt(String(value), "10");',
    'parseInt?.(String(value), 10);',
    'parseInt(String(...value), 10);',
    'Number.parseInt(String(...value), 10);',
    'parseInt(String(value), ...[10]);',
    'Number.parseInt(String(value), ...[10]);',
    'parseInt(String(value, extra), 10);',
    'parseInt(new String(value), 10);',
    'parseInt(String(value), 10, extra);',
    'Number.parseInt(String(value), 10, extra);',
    'Number.parseInt?.(String(value), 10);',
    'Number?.parseInt(String(value), 10);',
    'Number["parseInt"](String(value), 10);',
    'const parseInt = value => value; parseInt(String(value), 10);',
    'const Number = {parseInt() {}}; Number.parseInt(String(value), 10);',
    'const String = value => value; parseInt(String(value), 10);',
    'const String = value => value; Number.parseInt(String(value), 10);',
    // Documentation examples
    `// ✅
const foo = 37.4;
console.log(foo | 3);`,
    `// ✅
const foo = 37.4;
console.log(~foo);`,
    `// ✅
const foo = 37.4;
console.log(foo >> 3);`,
    `// ✅
const foo = 37.4;
console.log(foo << 3);`,
    `// ✅
const foo = 37.4;
console.log(foo ^ 3);`,
  ].map((code) => ({ code, filename: jsFilename })),
  invalid: [
    // Parse-int suggestions with comments
    jsCase(
      'function foo() {return(parseInt)(String(read(/* keep */)), 10);}',
      'error-parse-int',
    ),
    jsCase(
      'function foo() {throw(Number.parseInt)(String(read(/* keep */)), 10);}',
      'error-parse-int',
    ),
    jsCase(
      'async function foo() {return await(parseInt)(String(read(/* keep */)), 10);}',
      'error-parse-int',
    ),
    // Bitwise suggestions with comments
    jsCase('read(/* keep */) | 0;', 'error-bitwise'),
    jsCase('(/* keep */ value) >> 0;', 'error-bitwise'),
    jsCase('values[/* keep */ index] << 0;', 'error-bitwise'),
    jsCase('(left + /* keep */ right) ^ 0;', 'error-bitwise'),
    jsCase('(read(/* keep */), value) | 0;', 'error-bitwise'),
    jsCase(
      'function foo() {return[/* keep */ value][0] | 0;}',
      'error-bitwise',
    ),
    jsCase(
      'function foo() {\r\n  return(value // keep\r\n  ) >> 0;\r\n}',
      'error-bitwise',
    ),
    tsCase('(value as /* keep */ number) | 0;', 'error-bitwise'),
    tsCase('values[/* keep */ index]! ^ 0;', 'error-bitwise'),
    jsCase('value | /* keep */ 0;', 'error-bitwise'),
    jsCase('value | (/* keep */ 0);', 'error-bitwise'),
    jsCase('read(/* keep */) /* drop */ | 0;', 'error-bitwise'),
    jsCase('values[/* keep */ index] |= 0;', 'error-bitwise'),
    jsCase('(/* keep */ value) >>= 0;', 'error-bitwise'),
    // Double-NOT suggestions with comments
    jsCase('~~getValue(/* keep */);', 'error-bitwise-not'),
    jsCase('~~(/* keep */ value);', 'error-bitwise-not'),
    jsCase('~~values[/* keep */ index];', 'error-bitwise-not'),
    jsCase('~~(left + /* keep */ right);', 'error-bitwise-not'),
    jsCase('~~(first(/* keep */), value);', 'error-bitwise-not'),
    jsCase('~~~getValue(/* keep */);', 'error-bitwise-not'),
    jsCase(
      'function foo() {return~~getValue(/* keep */);}',
      'error-bitwise-not',
    ),
    jsCase(
      'async function foo() { return ~~(await /* keep */ getValue()); }',
      'error-bitwise-not',
    ),
    jsCase(
      'function* foo() { return ~~(yield /* keep */ getValue()); }',
      'error-bitwise-not',
    ),
    jsCase(
      'function foo() {\r\n  return~~(value // keep\r\n  );\r\n}',
      'error-bitwise-not',
    ),
    tsCase('~~(value as /* keep */ number);', 'error-bitwise-not'),
    tsCase('~~values[/* keep */ index]!;', 'error-bitwise-not'),
    jsCase('~~ /* keep */ value;', 'error-bitwise-not'),
    jsCase('~(/* keep */ ~value);', 'error-bitwise-not'),
    // Upstream snapshots
    jsCase('const foo = parseInt(String(number), 10);', 'error-parse-int'),
    jsCase(
      'const foo = Number.parseInt(String(number), 10);',
      'error-parse-int',
    ),
    jsCase('const foo = parseInt(String((number)), 10);', 'error-parse-int'),
    jsCase('const foo = parseInt(String(number + 1), 10);', 'error-parse-int'),
    jsCase('const foo = parseInt(String(number,), 10);', 'error-parse-int'),
    jsCase(
      'const foo = parseInt(String(/* comment */ number), 10);',
      'error-parse-int',
    ),
    jsCase(
      'const foo = parseInt(String(number /* comment */), 10);',
      'error-parse-int',
    ),
    jsCase(
      'const foo = parseInt(/* comment */ String(number), 10);',
      'error-parse-int',
    ),
    jsCase(
      'const foo = parseInt(String /* comment */ (number), 10);',
      'error-parse-int',
    ),
    jsCase(
      'const foo = parseInt(String(number), /* comment */ 10);',
      'error-parse-int',
    ),
    jsCase('const foo = parseInt(String((0, number)), 10);', 'error-parse-int'),
    tsCase(
      'const foo = parseInt(String(number as number), 10);',
      'error-parse-int',
    ),
    tsCase(
      'const foo = parseInt(String<() => number>(number), 10);',
      'error-parse-int',
    ),
    tsCase('const foo = parseInt(String(number!), 10);', 'error-parse-int'),
    jsCase('const foo = 1.1 | 0;', 'error-bitwise'),
    jsCase('const foo = 111 | 0;', 'error-bitwise'),
    jsCase('const foo = (1 + 2 / 3.4) | 0;', 'error-bitwise'),
    jsCase('const foo = bar((1.4 | 0) + 2);', 'error-bitwise'),
    jsCase('const foo = (0, 1.4) | 0;', 'error-bitwise'),
    jsCase('function foo() {return.1 | 0;}', 'error-bitwise'),
    jsCase('const foo = 1.4 | 0.;', 'error-bitwise'),
    jsCase('const foo = 1.4 | .0;', 'error-bitwise'),
    jsCase('const foo = 1.4 | 0.0000_0000_0000;', 'error-bitwise'),
    jsCase('const foo = 1.4 | 0b0;', 'error-bitwise'),
    jsCase('const foo = 1.4 | 0x0000_0000_0000;', 'error-bitwise'),
    jsCase('const foo = 1.4 | 0o0;', 'error-bitwise'),
    jsCase('const foo = 1.23 | 0 | 4;', 'error-bitwise'),
    jsCase('const foo = ~~3.9;', 'error-bitwise-not'),
    jsCase('const foo = ~~111;', 'error-bitwise-not'),
    jsCase('const foo = ~~(1 + 2 / 3.4);', 'error-bitwise-not'),
    jsCase('const foo = ~~1 + 2 / 3.4;', 'error-bitwise-not'),
    jsCase('const foo = ~~(0, 1.4);', 'error-bitwise-not'),
    jsCase('const foo = ~~~10.01;', 'error-bitwise-not'),
    jsCase('const foo = ~~(~10.01);', 'error-bitwise-not'),
    jsCase('const foo = ~(~~10.01);', 'error-bitwise-not'),
    jsCase('const foo = ~~-10.01;', 'error-bitwise-not'),
    jsCase('const foo = ~~~~10.01;', 'error-bitwise-not'),
    jsCase('function foo() {return~~3.9;}', 'error-bitwise-not'),
    jsCase('const foo = bar >> 0;', 'error-bitwise'),
    jsCase('const foo = (a + b) >> 0;', 'error-bitwise'),
    tsCase('const foo = ~~bar!;', 'error-bitwise-not'),
    jsCase('const foo = ~~(a ? b : c);', 'error-bitwise-not'),
    jsCase('const foo = bar << 0;', 'error-bitwise'),
    jsCase('const foo = bar ^ 0;', 'error-bitwise'),
    jsCase('function foo() {return.1 ^0;}', 'error-bitwise'),
    jsCase(
      `const foo = {a: {b: {c: 3}}};
const bar = a.b.c | 0;`,
      'error-bitwise',
    ),
    jsCase(
      `const foo = {a: {b: {c: 3}}};
const bar = a.b?.c | 0;`,
      'error-bitwise',
    ),
    jsCase(
      `const foo = {a: {b: {c: 3}}};
const bar = ~~a.b?.c;`,
      'error-bitwise-not',
    ),
    jsCase(
      `const foo = 3;
const bar = foo | 0;`,
      'error-bitwise',
    ),
    jsCase(
      `const foo = 3;
const bar = ~~foo;`,
      'error-bitwise-not',
    ),
    jsCase(
      `let foo = 2;
foo |= 0;`,
      'error-bitwise',
    ),
    jsCase(
      `const foo = {a: {b: 3.4}};
foo.a.b |= 0;`,
      'error-bitwise',
    ),
    jsCase(
      `const foo = 10.01;
const bar = ~~foo;`,
      'error-bitwise-not',
    ),
    jsCase(
      `let foo = 10.01;
foo >>= 0;`,
      'error-bitwise',
    ),
    jsCase(
      `let foo = 10.01;
foo <<= 0;`,
      'error-bitwise',
    ),
    jsCase(
      `let foo = 10.01;
foo ^= 0;`,
      'error-bitwise',
    ),
    jsCase('function foo() {return[foo][0] ^= 0;};', 'error-bitwise'),
    jsCase(
      'const foo = /* first comment */ 3.4 | 0; // A B C',
      'error-bitwise',
    ),
    jsCase(
      'const foo = /* first comment */ ~~3.4; // A B C',
      'error-bitwise-not',
    ),
    jsCase(
      `const foo = {a: {b: 3.4}};
foo /* Comment 1 */ .a /* Comment 2 */ . /* Comment 3 */ b |= /* Comment 4 */ 0 /* Comment 5 */;`,
      'error-bitwise',
    ),
    jsCase(
      `const foo = {a: {b: 3.4}};
const bar = /* Comment 1 */ ~~ a /* Comment 3 */ . /* Comment 4 */ b /* Comment 5 */;`,
      'error-bitwise-not',
    ),
    jsCase(
      'const foo = /* will keep */ 3.4 /* will remove 1 */ | /* will remove 2 */ 0;',
      'error-bitwise',
    ),
    jsCase(
      'const foo = /* will keep */ ~ /* will remove 1 */ ~ /* will remove 2 */ 3.4;',
      'error-bitwise-not',
    ),
    jsCase(
      `const foo = 3.4; // comment 1
foo |= 0; // comment 2`,
      'error-bitwise',
    ),
    jsCase(
      `const foo = 3.4; // comment 1
const bar = ~~foo; // comment 2`,
      'error-bitwise-not',
    ),
    jsCase('const foo = ~~bar | 0;', 'error-bitwise', 'error-bitwise-not'),
    jsCase('const foo = ~~(bar| 0);', 'error-bitwise-not', 'error-bitwise'),
    jsCase('const foo = bar | 0 | 0;', 'error-bitwise', 'error-bitwise'),
    jsCase(
      'const foo = ~~~~((bar | 0 | 0) >> 0 >> 0 << 0 << 0 ^ 0 ^0);',
      'error-bitwise-not',
      'error-bitwise',
      'error-bitwise',
      'error-bitwise',
      'error-bitwise',
      'error-bitwise',
      'error-bitwise',
      'error-bitwise',
      'error-bitwise',
    ),
    jsCase(
      `const foo = Array.from({length: 10}, (_, index) => (index + 1) + (index + 1) /100);
let i = 0;
while (i < foo.length) {
	foo[i++] |= 0;
}
console.log(foo);`,
      'error-bitwise',
    ),
    // Explicit suggestion regressions
    jsCase('const foo = ~~(/* keep */ 1 + 2 / 3.4);', 'error-bitwise-not'),
    jsCase('const foo = /* keep */ ~ /* drop */ ~ 3.4;', 'error-bitwise-not'),
    jsCase('const foo = x /* keep */ << 0;', 'error-bitwise'),
    jsCase('const foo = ~~(1 + 2 / 3.4);', 'error-bitwise-not'),
    jsCase('const foo = ~~(2 ** 32);', 'error-bitwise-not'),
    jsCase('const foo = x << 0;', 'error-bitwise'),
    jsCase('foo |= /* keep */ 0;', 'error-bitwise'),
    // Documentation examples
    jsCase(
      `const foo = 37.4;

// ❌
console.log(foo | 0);

// ❌
console.log(~~foo);

// ❌
console.log(foo << 0);

// ❌
console.log(foo >> 0);

// ❌
console.log(foo.bar ^ 0);

// ❌
console.log(parseInt(String(foo), 10));

// ❌
console.log(Number.parseInt(String(foo), 10));

// ✅
console.log(Math.trunc(foo));`,
      'error-bitwise',
      'error-bitwise-not',
      'error-bitwise',
      'error-bitwise',
      'error-bitwise',
      'error-parse-int',
      'error-parse-int',
    ),
    jsCase(
      `let foo = 37.4;

// ❌
foo |= 0;

// ✅
foo = Math.trunc(foo);`,
      'error-bitwise',
    ),
  ],
});

const configDirectory = path.resolve(import.meta.dirname, '..');
const absoluteFilename = path.resolve(configDirectory, jsFilename);

function lintCode(code: string) {
  return lint({
    config: [
      { plugins: ['unicorn'], rules: { 'unicorn/prefer-math-trunc': 'error' } },
    ],
    configDirectory,
    workingDirectory: process.cwd(),
    fileContents: { [absoluteFilename]: code },
    fix: true,
  });
}

async function applyOnlySuggestion(
  code: string,
  messageId: string,
  message: string,
) {
  const result = await lintCode(code);
  expect(result.diagnostics).toHaveLength(1);
  expect(result.output ?? {}).toEqual({});
  expect(result.fixableErrorCount).toBe(0);
  const diagnostic = result.diagnostics[0];
  expect(diagnostic.suggestions).toHaveLength(1);
  const suggestion = diagnostic.suggestions?.[0];
  assert.ok(suggestion?.fixes?.length);
  expect(suggestion.messageId).toBe(messageId);
  expect(suggestion.message).toBe(message);

  let output = code;
  const fixes = [...suggestion.fixes].sort(
    (a, b) => b.startPos - a.startPos || b.endPos - a.endPos,
  );
  for (const { startPos, endPos, text } of fixes) {
    output = output.slice(0, startPos) + text + output.slice(endPos);
  }
  expect((await lintCode(output)).diagnostics).toHaveLength(0);
  return output;
}

test('commented yield suggestions preserve suspension', async () => {
  const code = '(function* () { return ~~(yield read(/* keep */)); })()';
  const suggested = await applyOnlySuggestion(
    code,
    'suggestion-bitwise-not',
    'Replace `~~` with `Math.trunc`.',
  );
  expect(suggested).toBe(
    '(function* () { return Math.trunc(yield read(/* keep */)); })()',
  );
  for (const source of [code, suggested]) {
    let calls = 0;
    const iterator = runInNewContext(source, {
      read() {
        calls++;
        return 'yielded';
      },
    }) as Generator<string, number, number>;
    expect(calls).toBe(0);
    expect(iterator.next()).toEqual({ value: 'yielded', done: false });
    expect(calls).toBe(1);
    expect(iterator.next(3.9)).toEqual({ value: 3, done: true });
    expect(calls).toBe(1);
  }
});

test('truncation suggestions preserve comments and evaluate the operand once', async () => {
  const operand = 'getValue(/* keep */)';
  const cases = [
    {
      code: `~~${operand};`,
      pattern: '~~',
      messageId: 'suggestion-bitwise-not',
    },
    ...['|', '>>', '<<', '^'].map((operator) => ({
      code: `${operand} ${operator} 0;`,
      pattern: `${operator} 0`,
      messageId: 'suggestion-bitwise',
    })),
  ];
  for (const { code, pattern, messageId } of cases) {
    const suggested = await applyOnlySuggestion(
      code,
      messageId,
      `Replace \`${pattern}\` with \`Math.trunc\`.`,
    );
    expect(suggested).toBe(`Math.trunc(${operand});`);
    for (const [value, original, truncated] of [
      [3.9, 3, 3],
      [2 ** 32, 0, 2 ** 32],
      [NaN, 0, NaN],
      [-0.5, 0, -0],
    ]) {
      for (const [source, expected] of [
        [code, original],
        [suggested, truncated],
      ] as const) {
        let calls = 0;
        const result = runInNewContext(source, {
          getValue() {
            calls++;
            return value;
          },
        });
        expect(result).toBe(expected);
        expect(calls).toBe(1);
      }
    }
  }
});
