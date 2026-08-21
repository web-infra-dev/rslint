import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Ported from ESLint v10.8.1 tests/lib/rules/no-magic-numbers.js, including
// the trailing `ruleTesterTypeScript` block (upstream runs the exact same
// rule through @typescript-eslint/parser there; rslint always parses
// TypeScript syntax, so both blocks apply equally here). languageOptions is
// dropped throughout: rslint's parser does not gate syntax on ecmaVersion the
// way espree does, and sourceType-only cases (legacy octal literals) are
// omitted since rslint's parser rejects that syntax regardless of
// module-ness.
ruleTester.run('no-magic-numbers', {
  valid: [
    'var x = parseInt(y, 10);',
    'var x = parseInt(y, -10);',
    'var x = Number.parseInt(y, 10);',
    'const MY_NUMBER = +42;',
    'const foo = 42;',
    {
      code: 'var foo = 42;',
      options: { enforceConst: false },
    },
    'var foo = -42;',
    {
      code: 'var foo = 0 + 1 - 2 + -2;',
      options: { ignore: [0, 1, 2, -2] },
    },
    {
      code: 'var foo = 0 + 1 + 2 + 3 + 4;',
      options: { ignore: [0, 1, 2, 3, 4] },
    },
    'var foo = { bar:10 }',
    {
      code: 'setTimeout(function() {return 1;}, 0);',
      options: { ignore: [0, 1] },
    },
    {
      code: "var data = ['foo', 'bar', 'baz']; var third = data[3];",
      options: { ignoreArrayIndexes: true },
    },
    { code: 'foo[0]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[-0]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[1]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[100]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[200.00]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[3e4]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[1.23e2]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[230e-1]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[0b110]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[0o71]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[0xABC]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[5.0000000000000001]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[4294967294]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[0n]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[-0n]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[1n]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[100n]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[0xABn]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[4294967294n]', options: { ignoreArrayIndexes: true } },

    {
      code: 'var a = <input maxLength={10} />;',
      filename: 'src/virtual.tsx',
    },
    {
      code: 'var a = <div objectProp={{ test: 1}}></div>;',
      filename: 'src/virtual.tsx',
    },

    { code: 'f(100n)', options: { ignore: ['100n'] } },
    { code: 'f(-100n)', options: { ignore: ['-100n'] } },

    {
      code: 'const { param = 123 } = sourceObject;',
      options: { ignoreDefaultValues: true },
    },
    {
      code: 'const func = (param = 123) => {}',
      options: { ignoreDefaultValues: true },
    },
    {
      code: 'const func = ({ param = 123 }) => {}',
      options: { ignoreDefaultValues: true },
    },
    {
      code: 'const [one = 1, two = 2] = []',
      options: { ignoreDefaultValues: true },
    },
    {
      code: 'var one, two; [one = 1, two = 2] = []',
      options: { ignoreDefaultValues: true },
    },

    // Optional chaining
    'var x = parseInt?.(y, 10);',
    'var x = Number?.parseInt(y, 10);',
    'var x = (Number?.parseInt)(y, 10);',
    { code: 'foo?.[777]', options: { ignoreArrayIndexes: true } },

    // ignoreClassFieldInitialValues
    {
      code: 'class C { foo = 2; }',
      options: { ignoreClassFieldInitialValues: true },
    },
    {
      code: 'class C { foo = -2; }',
      options: { ignoreClassFieldInitialValues: true },
    },
    {
      code: 'class C { static foo = 2; }',
      options: { ignoreClassFieldInitialValues: true },
    },
    {
      code: 'class C { #foo = 2; }',
      options: { ignoreClassFieldInitialValues: true },
    },
    {
      code: 'class C { static #foo = 2; }',
      options: { ignoreClassFieldInitialValues: true },
    },

    { code: 'foo[+0]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[+1]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[+0n]', options: { ignoreArrayIndexes: true } },
    { code: 'foo[+1n]', options: { ignoreArrayIndexes: true } },

    // ---- TypeScript parser suite (upstream ruleTesterTypeScript) ----
    { code: 'const FOO = 10;', options: { ignoreNumericLiteralTypes: true } },
    "type Foo = 'bar';",
    'type Foo = true;',
    { code: 'type Foo = 1;', options: { ignoreNumericLiteralTypes: true } },
    { code: 'type Foo = -1;', options: { ignoreNumericLiteralTypes: true } },
    {
      code: "type Nested = ('' | ('' | (1)));",
      options: { ignoreNumericLiteralTypes: true },
    },
    {
      code: 'type Foo = 1 | 2 | 3;',
      options: { ignoreNumericLiteralTypes: true },
    },
    {
      code: 'type Foo = 1 | -1;',
      options: { ignoreNumericLiteralTypes: true },
    },
    {
      code: `
enum foo {
  SECOND = 1000,
  NUM = '0123456789',
  NEG = -1,
  POS = +1,
}
`,
      options: { ignoreEnums: true },
    },
    {
      code: `
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 1;
  static readonly D = 1;
  readonly E = -1;
  readonly F = +1;
  private readonly G = 100n;
}
`,
      options: { ignoreReadonlyClassProperties: true },
    },
    { code: 'type Foo = Bar[0];', options: { ignoreTypeIndexes: true } },
    { code: 'type Foo = Bar[-1];', options: { ignoreTypeIndexes: true } },
    { code: 'type Foo = Bar[0xab];', options: { ignoreTypeIndexes: true } },
    { code: 'type Foo = Bar[5.6e1];', options: { ignoreTypeIndexes: true } },
    { code: 'type Foo = Bar[10n];', options: { ignoreTypeIndexes: true } },
    { code: 'type Foo = Bar[1 | -2];', options: { ignoreTypeIndexes: true } },
    { code: 'type Foo = Bar[1 & -2];', options: { ignoreTypeIndexes: true } },
    {
      code: 'type Foo = Bar[1 & number];',
      options: { ignoreTypeIndexes: true },
    },
    {
      code: 'type Foo = Bar[((1 & -2) | 3) | 4];',
      options: { ignoreTypeIndexes: true },
    },
    {
      code: 'type Foo = Parameters<Bar>[2];',
      options: { ignoreTypeIndexes: true },
    },
    { code: "type Foo = Bar['baz'];", options: { ignoreTypeIndexes: true } },
    { code: "type Foo = Bar['baz'];", options: { ignoreTypeIndexes: false } },
    {
      code: `
type Others = [['a'], ['b']];

type Foo = {
  [K in keyof Others[0]]: Others[K];
};
`,
      options: { ignoreTypeIndexes: true },
    },
    { code: 'type Foo = 1;', options: { ignore: [1] } },
    { code: 'type Foo = -2;', options: { ignore: [-2] } },
    { code: 'type Foo = 3n;', options: { ignore: ['3n'] } },
    { code: 'type Foo = -4n;', options: { ignore: ['-4n'] } },
    { code: 'type Foo = 5.6;', options: { ignore: [5.6] } },
    { code: 'type Foo = -7.8;', options: { ignore: [-7.8] } },
    { code: 'type Foo = 0x0a;', options: { ignore: [0x0a] } },
    { code: 'type Foo = -0xbc;', options: { ignore: [-0xbc] } },
    { code: 'type Foo = 1e2;', options: { ignore: [1e2] } },
    { code: 'type Foo = -3e4;', options: { ignore: [-3e4] } },
    { code: 'type Foo = 5e-6;', options: { ignore: [5e-6] } },
    { code: 'type Foo = -7e-8;', options: { ignore: [-7e-8] } },
    { code: 'type Foo = 1.1e2;', options: { ignore: [1.1e2] } },
    { code: 'type Foo = -3.1e4;', options: { ignore: [-3.1e4] } },
    { code: 'type Foo = 5.1e-6;', options: { ignore: [5.1e-6] } },
    { code: 'type Foo = -7.1e-8;', options: { ignore: [-7.1e-8] } },
    {
      code: `
interface Foo {
  bar: 1;
}
`,
      options: { ignore: [1], ignoreNumericLiteralTypes: true },
    },
    {
      code: `
enum foo {
  SECOND = 1000,
  NUM = '0123456789',
  NEG = -1,
  POS = +2,
}
`,
      options: { ignore: [1000, -1, 2], ignoreEnums: false },
    },
    {
      code: `
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 3;
  static readonly D = 4;
  readonly E = -5;
  readonly F = +6;
  private readonly G = 100n;
  private static readonly H = -2000n;
}
`,
      options: {
        ignore: [1, 2, 3, 4, -5, 6, '100n', '-2000n'],
        ignoreReadonlyClassProperties: false,
      },
    },
    {
      code: 'type Foo = Bar[0];',
      options: { ignore: [0], ignoreTypeIndexes: false },
    },
    {
      code: `
type Other = {
  [0]: 3;
};

type Foo = {
  [K in keyof Other]: \`\${K & number}\`;
};
`,
      options: { ignore: [0, 3], ignoreTypeIndexes: true },
    },
    {
      code: `
class C {
	readonly foo = +42;
	bar = +42;
}

const MY_NUMBER = +42;
`,
      options: {
        ignoreClassFieldInitialValues: true,
        ignoreReadonlyClassProperties: true,
      },
    },
  ],

  invalid: [
    {
      code: 'var foo = 42',
      options: { enforceConst: true },
      errors: [{ messageId: 'useConst' }],
    },
    {
      code: 'var foo = 0 + 1;',
      errors: [{ messageId: 'noMagic' }, { messageId: 'noMagic' }],
    },
    {
      code: 'var foo = 42n',
      options: { enforceConst: true },
      errors: [{ messageId: 'useConst' }],
    },
    {
      code: 'var foo = 0n + 1n;',
      errors: [{ messageId: 'noMagic' }, { messageId: 'noMagic' }],
    },
    {
      code: 'a = a + 5;',
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'a += 5;',
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'var foo = 0 + 1 + -2 + 2;',
      errors: [
        { messageId: 'noMagic' },
        { messageId: 'noMagic' },
        { messageId: 'noMagic' },
        { messageId: 'noMagic' },
      ],
    },
    {
      code: 'var foo = 0 + 1 + 2;',
      options: { ignore: [0, 1] },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'var foo = { bar:10 }',
      options: { detectObjects: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'var stats = {avg: 42};',
      options: { detectObjects: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'var colors = {}; colors.RED = 2; colors.YELLOW = 3; colors.BLUE = 4 + 5;',
      errors: [{ messageId: 'noMagic' }, { messageId: 'noMagic' }],
    },
    {
      code: 'function getSecondsInMinute() {return 60;}',
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'function getNegativeSecondsInMinute() {return -60;}',
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: `var Promise = require('bluebird');
var MINUTE = 60;
var HOUR = 3600;
const DAY = 86400;
var configObject = {
key: 90,
another: 10 * 10,
10: 'an "integer" key'
};
function getSecondsInDay() {
 return 24 * HOUR;
}
function getMillisecondsInDay() {
return (getSecondsInDay() *
(1000)
);
}
function callSetTimeoutZero(func) {
setTimeout(func, 0);
}
function invokeInTen(func) {
setTimeout(func, 10);
}
`,
      errors: [
        { messageId: 'noMagic', line: 7 },
        { messageId: 'noMagic', line: 7 },
        { messageId: 'noMagic', line: 11 },
        { messageId: 'noMagic', line: 15 },
        { messageId: 'noMagic', line: 19 },
        { messageId: 'noMagic', line: 22 },
      ],
    },
    {
      code: "var data = ['foo', 'bar', 'baz']; var third = data[3];",
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: "var data = ['foo', 'bar', 'baz']; var third = data[3];",
      options: {},
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: "var data = ['foo', 'bar', 'baz']; var third = data[3];",
      options: { ignoreArrayIndexes: false },
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'foo[-100]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-1.5]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-1]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-0.1]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-0b110]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-0o71]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-0x12]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[0.1]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[0.12e1]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[1.5]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[1.678e2]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[56e-1]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[5.000000000000001]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[100.9]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[4294967295]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[1e300]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[1e310]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-1e310]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-1n]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-100n]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-0x12n]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[4294967295n]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[-(-1)]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'foo[- -1n]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: '100 .toString()',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: '200[100]',
      options: { ignoreArrayIndexes: true },
      errors: [{ messageId: 'noMagic' }],
    },
    {
      code: 'var a = <div arrayProp={[1,2,3]}></div>;',
      filename: 'src/virtual.tsx',
      errors: [
        { messageId: 'noMagic', line: 1 },
        { messageId: 'noMagic', line: 1 },
        { messageId: 'noMagic', line: 1 },
      ],
    },
    {
      code: 'var min, max, mean; min = 1; max = 10; mean = 4;',
      errors: [
        { messageId: 'noMagic', line: 1 },
        { messageId: 'noMagic', line: 1 },
        { messageId: 'noMagic', line: 1 },
      ],
    },
    {
      code: 'f(100n)',
      options: { ignore: [100] },
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'f(-100n)',
      options: { ignore: ['100n'] },
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'f(100n)',
      options: { ignore: ['-100n'] },
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'f(100)',
      options: { ignore: ['100n'] },
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'const func = (param = 123) => {}',
      options: { ignoreDefaultValues: false },
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'const { param = 123 } = sourceObject;',
      options: {},
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'const { param = 123 } = sourceObject;',
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'const { param = 123 } = sourceObject;',
      options: { ignoreDefaultValues: false },
      errors: [{ messageId: 'noMagic', line: 1 }],
    },
    {
      code: 'const [one = 1, two = 2] = []',
      options: { ignoreDefaultValues: false },
      errors: [
        { messageId: 'noMagic', line: 1 },
        { messageId: 'noMagic', line: 1 },
      ],
    },
    {
      code: 'var one, two; [one = 1, two = 2] = []',
      options: { ignoreDefaultValues: false },
      errors: [
        { messageId: 'noMagic', line: 1 },
        { messageId: 'noMagic', line: 1 },
      ],
    },

    // ignoreClassFieldInitialValues
    {
      code: 'class C { foo = 2; }',
      errors: [{ messageId: 'noMagic', line: 1, column: 17 }],
    },
    {
      code: 'class C { foo = 2; }',
      options: {},
      errors: [{ messageId: 'noMagic', line: 1, column: 17 }],
    },
    {
      code: 'class C { foo = 2; }',
      options: { ignoreClassFieldInitialValues: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 17 }],
    },
    {
      code: 'class C { foo = -2; }',
      options: { ignoreClassFieldInitialValues: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 17 }],
    },
    {
      code: 'class C { static foo = 2; }',
      options: { ignoreClassFieldInitialValues: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 24 }],
    },
    {
      code: 'class C { #foo = 2; }',
      options: { ignoreClassFieldInitialValues: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 18 }],
    },
    {
      code: 'class C { static #foo = 2; }',
      options: { ignoreClassFieldInitialValues: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 25 }],
    },
    {
      code: 'class C { foo = 2 + 3; }',
      options: { ignoreClassFieldInitialValues: true },
      errors: [
        { messageId: 'noMagic', line: 1, column: 17 },
        { messageId: 'noMagic', line: 1, column: 21 },
      ],
    },
    {
      code: 'class C { 2; }',
      options: { ignoreClassFieldInitialValues: true },
      errors: [{ messageId: 'noMagic', line: 1, column: 11 }],
    },
    {
      code: 'class C { [2]; }',
      options: { ignoreClassFieldInitialValues: true },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },

    // ---- TypeScript parser suite (upstream ruleTesterTypeScript) ----
    {
      code: 'type Foo = 1;',
      options: { ignoreNumericLiteralTypes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -1;',
      options: { ignoreNumericLiteralTypes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = 1 | 2 | 3;',
      options: { ignoreNumericLiteralTypes: false },
      errors: [
        { messageId: 'noMagic', line: 1, column: 12 },
        { messageId: 'noMagic', line: 1, column: 16 },
        { messageId: 'noMagic', line: 1, column: 20 },
      ],
    },
    {
      code: 'type Foo = 1 | -1;',
      options: { ignoreNumericLiteralTypes: false },
      errors: [
        { messageId: 'noMagic', line: 1, column: 12 },
        { messageId: 'noMagic', line: 1, column: 16 },
      ],
    },
    {
      code: `
interface Foo {
  bar: 1;
}
`,
      options: { ignoreNumericLiteralTypes: true },
      errors: [{ messageId: 'noMagic', line: 3, column: 8 }],
    },
    {
      code: `
enum foo {
  SECOND = 1000,
  NUM = '0123456789',
  NEG = -1,
  POS = +1,
}
`,
      options: { ignoreEnums: false },
      errors: [
        { messageId: 'noMagic', line: 3, column: 12 },
        { messageId: 'noMagic', line: 5, column: 9 },
        { messageId: 'noMagic', line: 6, column: 9 },
      ],
    },
    {
      code: `
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 3;
  static readonly D = 4;
  readonly E = -5;
  readonly F = +6;
  private readonly G = 100n;
}
`,
      options: { ignoreReadonlyClassProperties: false },
      errors: [
        { messageId: 'noMagic', line: 3, column: 16 },
        { messageId: 'noMagic', line: 4, column: 16 },
        { messageId: 'noMagic', line: 5, column: 30 },
        { messageId: 'noMagic', line: 6, column: 23 },
        { messageId: 'noMagic', line: 7, column: 16 },
        { messageId: 'noMagic', line: 8, column: 16 },
        { messageId: 'noMagic', line: 9, column: 24 },
      ],
    },
    {
      code: 'type Foo = Bar[0];',
      options: { ignoreTypeIndexes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 16 }],
    },
    {
      code: 'type Foo = Bar[-1];',
      options: { ignoreTypeIndexes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 16 }],
    },
    {
      code: 'type Foo = Bar[0xab];',
      options: { ignoreTypeIndexes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 16 }],
    },
    {
      code: 'type Foo = Bar[5.6e1];',
      options: { ignoreTypeIndexes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 16 }],
    },
    {
      code: 'type Foo = Bar[10n];',
      options: { ignoreTypeIndexes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 16 }],
    },
    {
      code: 'type Foo = Bar[1 | -2];',
      options: { ignoreTypeIndexes: false },
      errors: [
        { messageId: 'noMagic', line: 1, column: 16 },
        { messageId: 'noMagic', line: 1, column: 20 },
      ],
    },
    {
      code: 'type Foo = Bar[1 & -2];',
      options: { ignoreTypeIndexes: false },
      errors: [
        { messageId: 'noMagic', line: 1, column: 16 },
        { messageId: 'noMagic', line: 1, column: 20 },
      ],
    },
    {
      code: 'type Foo = Bar[1 & number];',
      options: { ignoreTypeIndexes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 16 }],
    },
    {
      code: 'type Foo = Bar[((1 & -2) | 3) | 4];',
      options: { ignoreTypeIndexes: false },
      errors: [
        { messageId: 'noMagic', line: 1, column: 18 },
        { messageId: 'noMagic', line: 1, column: 22 },
        { messageId: 'noMagic', line: 1, column: 28 },
        { messageId: 'noMagic', line: 1, column: 33 },
      ],
    },
    {
      code: 'type Foo = Parameters<Bar>[2];',
      options: { ignoreTypeIndexes: false },
      errors: [{ messageId: 'noMagic', line: 1, column: 28 }],
    },
    {
      code: `
type Others = [['a'], ['b']];

type Foo = {
  [K in keyof Others[0]]: Others[K];
};
`,
      options: { ignoreTypeIndexes: false },
      errors: [{ messageId: 'noMagic', line: 5, column: 22 }],
    },
    {
      code: `
type Other = {
  [0]: 3;
};

type Foo = {
  [K in keyof Other]: \`\${K & number}\`;
};
`,
      options: { ignoreTypeIndexes: true },
      errors: [
        { messageId: 'noMagic', line: 3, column: 4 },
        { messageId: 'noMagic', line: 3, column: 8 },
      ],
    },
    {
      code: `
type Foo = {
  [K in 0 | 1 | 2]: 0;
};
`,
      options: { ignoreTypeIndexes: true },
      errors: [
        { messageId: 'noMagic', line: 3, column: 9 },
        { messageId: 'noMagic', line: 3, column: 13 },
        { messageId: 'noMagic', line: 3, column: 17 },
        { messageId: 'noMagic', line: 3, column: 21 },
      ],
    },
    {
      code: 'type Foo = 1;',
      options: { ignore: [-1] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -2;',
      options: { ignore: [2] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = 3n;',
      options: { ignore: ['-3n'] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -4n;',
      options: { ignore: ['4n'] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = 5.6;',
      options: { ignore: [-5.6] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -7.8;',
      options: { ignore: [7.8] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = 0x0a;',
      options: { ignore: [-0x0a] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -0xbc;',
      options: { ignore: [0xbc] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = 1e2;',
      options: { ignore: [-1e2] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -3e4;',
      options: { ignore: [3e4] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = 5e-6;',
      options: { ignore: [-5e-6] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -7e-8;',
      options: { ignore: [7e-8] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = 1.1e2;',
      options: { ignore: [-1.1e2] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -3.1e4;',
      options: { ignore: [3.1e4] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = 5.1e-6;',
      options: { ignore: [-5.1e-6] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = -7.1e-8;',
      options: { ignore: [7.1e-8] },
      errors: [{ messageId: 'noMagic', line: 1, column: 12 }],
    },
    {
      code: 'type Foo = { bar: 42 };',
      options: { ignoreNumericLiteralTypes: true },
      errors: [{ messageId: 'noMagic', line: 1, column: 19 }],
    },
    {
      code: 'type Foo = { bar: 2 | 3 };',
      options: { ignoreNumericLiteralTypes: true },
      errors: [
        { messageId: 'noMagic', line: 1, column: 19 },
        { messageId: 'noMagic', line: 1, column: 23 },
      ],
    },
    {
      code: 'type Foo = { bar: Bar[((1 & -2) | 3) | 4] };',
      options: { ignoreNumericLiteralTypes: true },
      errors: [
        { messageId: 'noMagic', line: 1, column: 25 },
        { messageId: 'noMagic', line: 1, column: 29 },
        { messageId: 'noMagic', line: 1, column: 35 },
        { messageId: 'noMagic', line: 1, column: 40 },
      ],
    },
  ],
});
