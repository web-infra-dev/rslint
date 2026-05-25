import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unexpected-multiline', {
  valid: [
    '(x || y).aFunction()',
    '[a, b, c].forEach(doSomething)',
    'var a = b;\n(x || y).doSomething()',
    'var a = b\n;(x || y).doSomething()',
    'var a = b\nvoid (x || y).doSomething()',
    'var a = b;\n[1, 2, 3].forEach(console.log)',
    'var a = b\nvoid [1, 2, 3].forEach(console.log)',
    '"abc\\\n(123)"',
    'var a = (\n(123)\n)',
    'f(\n(x)\n)',
    '(\nfunction () {}\n)[1]',
    'let x = function() {};\n   `hello`',
    'let x = function() {}\nx `hello`',
    'String.raw `Hi\n${2+3}!`;',
    'x\n.y\nz `Valid Test Case`',
    'f(x\n)`Valid Test Case`',
    'x.\ny `Valid Test Case`',
    '(x\n)`Valid Test Case`',
    '\n                foo\n                / bar /2\n            ',
    '\n                foo\n                / bar / mgy\n            ',
    '\n                foo\n                / bar /\n                gym\n            ',
    '\n                foo\n                / bar\n                / ygm\n            ',
    '\n                foo\n                / bar /GYM\n            ',
    '\n                foo\n                / bar / baz\n            ',
    'foo /bar/g',
    '\n                foo\n                /denominator/\n                2\n            ',
    '\n                foo\n                / /abc/\n            ',
    '\n                5 / (5\n                / 5)\n            ',

    // TypeScript generic type-argument forms (issue #11650).
    '\n                tag<generic>`\n                    multiline\n                `;\n            ',
    '\n                tag<\n                  generic\n                >`\n                    multiline\n                `;\n            ',
    '\n                tag<\n                  generic\n                >`multiline`;\n            ',

    // Optional chaining — `?.` opts the link out of the rule.
    'var a = b\n  ?.(x || y).doSomething()',
    'var a = b\n  ?.[a, b, c].forEach(doSomething)',
    'var a = b?.\n  (x || y).doSomething()',
    'var a = b?.\n  [a, b, c].forEach(doSomething)',

    // Class fields where ASI separates members.
    'class C { field1\n[field2]; }',
    'class C { field1\n*gen() {} }',
    'class C { field1 = () => {}\n[field2]; }',
    'class C { field1 = () => {}\n*gen() {} }',

    // === Real-world IIFE / chain patterns ===
    'var a = b;\n;(function() {})()',
    '(function () {\n  return 1;\n})();',
    'fetch(url)\n.then(r => r.json())\n.catch(handle)',
    '$(selector)\n  .find(".x")\n  .each(fn)',

    // === TS generics — same line / multi-line type args ===
    'fn<string>(x)',
    '(fn<string>(x)).y',
    'fn<\n  Props,\n  State\n>(x)',
    // ESLint treats `<` as the next token after `fn`, on the SAME line.
    'fn<string>\n(x)',

    // === Disambiguation cases ===
    '/foo/g.test(x)',
    'a * b / c * d',
    'a\n% b',
    'arr[0]',
    "obj['a']['b']",
    'b[c]',
    'b![c]',
    '(b as any)[c]',

    // === Template real-world ===
    'tag`hello`',
    'tag`first\nsecond\nthird`',
    'fn<T>`x`',
    '`hello`',

    // === Optional chaining variants ===
    'a?.b\n.c',
    'a?.b()',
    'a\n?.[b]',

    // === Operators that are NOT `/` ===
    '(a, b) / c / d',
    'let x = 1; x /= 2',
    'foo / bar / g',
    'foo\n/ bar /2',
    '((foo / bar)) / 2',
    '(foo / bar) / 5',

    // === ASI-already-inserted statements ===
    '[1, 2, 3]',
    '(1 + 2)',
    'function f() {\n  return\n  (1).toString()\n}',
  ],
  invalid: [
    {
      code: 'var a = b\n(x || y).doSomething()',
      errors: [{ messageId: 'function' }],
    },
    {
      code: 'var a = (a || b)\n(x || y).doSomething()',
      errors: [{ messageId: 'function' }],
    },
    {
      code: 'var a = (a || b)\n(x).doSomething()',
      errors: [{ messageId: 'function' }],
    },
    {
      code: 'var a = b\n[a, b, c].forEach(doSomething)',
      errors: [{ messageId: 'property' }],
    },
    {
      code: 'var a = b\n    (x || y).doSomething()',
      errors: [{ messageId: 'function' }],
    },
    {
      code: 'var a = b\n  [a, b, c].forEach(doSomething)',
      errors: [{ messageId: 'property' }],
    },
    {
      code: 'let x = function() {}\n `hello`',
      errors: [{ messageId: 'taggedTemplate' }],
    },
    {
      code: 'let x = function() {}\nx\n`hello`',
      errors: [{ messageId: 'taggedTemplate' }],
    },
    {
      code: 'x\n.y\nz\n`Invalid Test Case`',
      errors: [{ messageId: 'taggedTemplate' }],
    },
    {
      code: '\n                foo\n                / bar /gym\n            ',
      errors: [{ messageId: 'division' }],
    },
    {
      code: '\n                foo\n                / bar /g\n            ',
      errors: [{ messageId: 'division' }],
    },
    {
      code: '\n                foo\n                / bar /g.test(baz)\n            ',
      errors: [{ messageId: 'division' }],
    },
    {
      code: '\n                foo\n                /bar/gimuygimuygimuy.test(baz)\n            ',
      errors: [{ messageId: 'division' }],
    },
    {
      code: '\n                foo\n                /bar/s.test(baz)\n            ',
      errors: [{ messageId: 'division' }],
    },
    {
      code: 'const x = aaaa<\n  test\n>/*\ntest\n*/`foo`',
      errors: [{ messageId: 'taggedTemplate' }],
    },
    {
      code: 'class C { field1 = obj\n[field2]; }',
      errors: [{ messageId: 'property' }],
    },
    {
      code: 'class C { field1 = function() {}\n[field2]; }',
      errors: [{ messageId: 'property' }],
    },

    // === Real-world ASI traps ===
    {
      code: 'var a = foo\n(function () {})()',
      errors: [{ messageId: 'function' }],
    },
    {
      code: 'const x = arr\n[0]',
      errors: [{ messageId: 'property' }],
    },
    {
      code: 'obj.a.b.c\n[d]',
      errors: [{ messageId: 'property' }],
    },
    {
      code: 'fn().g()\n`x`',
      errors: [{ messageId: 'taggedTemplate' }],
    },
    {
      code: 'obj[k]\n(x)',
      errors: [{ messageId: 'function' }],
    },

    // === TS-only receiver wrappers across lines ===
    {
      code: 'b!\n[c]',
      errors: [{ messageId: 'property' }],
    },
    {
      code: '(b as any)\n(c)',
      errors: [{ messageId: 'function' }],
    },
    {
      code: '(x satisfies Foo)\n(y)',
      errors: [{ messageId: 'function' }],
    },
    {
      code: '(arr as const)\n[0]',
      errors: [{ messageId: 'property' }],
    },

    // === Division flag exhaustive ===
    {
      code: 'foo\n/bar/gimsuy',
      errors: [{ messageId: 'division' }],
    },
    {
      code: 'x\n/y/i',
      errors: [{ messageId: 'division' }],
    },
    {
      code: 'foo /* x */\n/ bar /g',
      errors: [{ messageId: 'division' }],
    },

    // === Templates with substitutions / blank lines / typed tags ===
    {
      code: 'tag\n`hi${x}!`',
      errors: [{ messageId: 'taggedTemplate' }],
    },
    {
      code: 'tag(x)\n`hello`',
      errors: [{ messageId: 'taggedTemplate' }],
    },
    {
      code: 'fn?.(x)\n`hello`',
      errors: [{ messageId: 'taggedTemplate' }],
    },

    // === Optional-chain CONTINUATION calls / accesses ===
    {
      code: 'a?.b\n(c)',
      errors: [{ messageId: 'function' }],
    },

    // === Multi-line breaks ===
    {
      code: 'var a = b\n\n\n[c]',
      errors: [{ messageId: 'property' }],
    },
    {
      code: 'tag(x)\n\n`hi`',
      errors: [{ messageId: 'taggedTemplate' }],
    },

    // === Class field continuations beyond upstream ===
    {
      code: 'class C { field = obj.method()\n[k]; }',
      errors: [{ messageId: 'property' }],
    },
    {
      code: 'class C { field = tag`x`\n[k]; }',
      errors: [{ messageId: 'property' }],
    },
  ],
});
