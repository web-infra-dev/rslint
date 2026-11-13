import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('prefer-exponentiation-operator', {
  valid: [
    // not Math.pow()
    'Object.pow(a, b)',
    'Math.max(a, b)',
    'Math',
    'Math(a, b)',
    'pow',
    'pow(a, b)',
    'Math.pow',
    'Math.Pow(a, b)',
    'math.pow(a, b)',
    'foo.Math.pow(a, b)',
    'new Math.pow(a, b)',
    'Math[pow](a, b)',
    'globalThis.Object.pow(a, b)',
    'globalThis.Math.max(a, b)',
    // not the global Math
    '/* globals Math:off*/ Math.pow(a, b)',
    'let Math; Math.pow(a, b);',
    'if (foo) { const Math = 1; Math.pow(a, b); }',
    'var x = function Math() { Math.pow(a, b); }',
    'function foo(Math) { Math.pow(a, b); }',
    'function foo() { Math.pow(a, b); var Math; }',
    // Mirrors upstream ecmaVersion 2019 / 6 / 2017 cases; rslint does not expose ecmaVersion-specific globalThis availability.
    {
      code: 'globalThis.Math.pow(a, b)',
      languageOptions: {
        ecmaVersion: 2019,
      },
    },
    {
      code: 'globalThis.Math.pow(a, b)',
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      code: 'globalThis.Math.pow(a, b)',
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    `
      var globalThis = bar;
      globalThis.Math.pow(a, b)
    `,
    'class C { #pow; foo() { Math.#pow(a, b); } }',
  ],
  invalid: [
    {
      code: 'Math.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: 'a**b',
    },
    {
      code: '(Math).pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'a**b',
    },
    {
      code: "Math['pow'](a, b)",
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'a**b',
    },
    {
      code: "(Math)['pow'](a, b)",
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'a**b',
    },
    {
      code: `var x=Math
.  pow( a,\x20
 b )`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 3,
          endColumn: 5,
        },
      ],
      output: 'var x=a**b',
    },
    {
      code: 'globalThis.Math.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'a**b',
    },
    {
      code: "globalThis.Math['pow'](a, b)",
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'a**b',
    },
    // able to catch some workarounds
    {
      code: 'Math[`pow`](a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'a**b',
    },
    {
      code: "Math[`${'pow'}`](a, b)",
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'a**b',
    },
    {
      code: "Math['p' + 'o' + 'w'](a, b)",
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'a**b',
    },
    // non-expression parents that don't require parens
    {
      code: 'var x = Math.pow(a, b);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'var x = a**b;',
    },
    {
      code: 'if(Math.pow(a, b)){}',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'if(a**b){}',
    },
    {
      code: 'for(;Math.pow(a, b);){}',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'for(;a**b;){}',
    },
    {
      code: 'switch(foo){ case Math.pow(a, b): break; }',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'switch(foo){ case a**b: break; }',
    },
    {
      code: '{ foo: Math.pow(a, b) }',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: '{ foo: a**b }',
    },
    {
      code: 'function foo(bar, baz = Math.pow(a, b), quux){}',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: 'function foo(bar, baz = a**b, quux){}',
    },
    {
      code: '`${Math.pow(a, b)}`',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: '`${a**b}`',
    },
    // non-expression parents that do require parens
    {
      code: 'class C extends Math.pow(a, b) {}',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'class C extends (a**b) {}',
    },
    // parents with a higher precedence
    {
      code: '+ Math.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: '+ (a**b)',
    },
    {
      code: '- Math.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: '- (a**b)',
    },
    {
      code: '! Math.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: '! (a**b)',
    },
    {
      code: 'typeof Math.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'typeof (a**b)',
    },
    {
      code: 'void Math.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'void (a**b)',
    },
    {
      code: 'Math.pow(a, b) .toString()',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: '(a**b) .toString()',
    },
    {
      code: 'Math.pow(a, b) ()',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: '(a**b) ()',
    },
    {
      code: 'Math.pow(a, b) ``',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: '(a**b) ``',
    },
    {
      code: '(class extends Math.pow(a, b) {})',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: '(class extends (a**b) {})',
    },
    // already parenthesised, shouldn't insert extra parens
    {
      code: '+(Math.pow(a, b))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: '+(a**b)',
    },
    {
      code: '(Math.pow(a, b)).toString()',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: '(a**b).toString()',
    },
    {
      code: '(class extends (Math.pow(a, b)) {})',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: '(class extends (a**b) {})',
    },
    {
      code: 'class C extends (Math.pow(a, b)) {}',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 32,
        },
      ],
      output: 'class C extends (a**b) {}',
    },
    // parents with a higher precedence, but the expression's role doesn't require parens
    {
      code: 'f(Math.pow(a, b))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'f(a**b)',
    },
    {
      code: 'f(foo, Math.pow(a, b))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'f(foo, a**b)',
    },
    {
      code: 'f(Math.pow(a, b), foo)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'f(a**b, foo)',
    },
    {
      code: 'f(foo, Math.pow(a, b), bar)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'f(foo, a**b, bar)',
    },
    {
      code: 'new F(Math.pow(a, b))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'new F(a**b)',
    },
    {
      code: 'new F(foo, Math.pow(a, b))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'new F(foo, a**b)',
    },
    {
      code: 'new F(Math.pow(a, b), foo)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'new F(a**b, foo)',
    },
    {
      code: 'new F(foo, Math.pow(a, b), bar)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 26,
        },
      ],
      output: 'new F(foo, a**b, bar)',
    },
    {
      code: 'obj[Math.pow(a, b)]',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'obj[a**b]',
    },
    {
      code: '[foo, Math.pow(a, b), bar]',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: '[foo, a**b, bar]',
    },
    // parents with a lower precedence
    {
      code: 'a * Math.pow(b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a * b**c',
    },
    {
      code: 'Math.pow(a, b) * c',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: 'a**b * c',
    },
    {
      code: 'a + Math.pow(b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a + b**c',
    },
    {
      code: 'Math.pow(a, b)/c',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: 'a**b/c',
    },
    {
      code: 'a < Math.pow(b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a < b**c',
    },
    {
      code: 'Math.pow(a, b) > c',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: 'a**b > c',
    },
    {
      code: 'a === Math.pow(b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'a === b**c',
    },
    {
      code: 'a ? Math.pow(b, c) : d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a ? b**c : d',
    },
    {
      code: 'a = Math.pow(b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a = b**c',
    },
    {
      code: 'a += Math.pow(b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'a += b**c',
    },
    {
      code: 'function *f() { yield Math.pow(a, b) }',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'function *f() { yield a**b }',
    },
    {
      code: 'a, Math.pow(b, c), d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'a, b**c, d',
    },
    // '**' is right-associative, that applies to both parent and child nodes
    {
      code: 'a ** Math.pow(b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'a ** b**c',
    },
    {
      code: 'Math.pow(a, b) ** c',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: '(a**b) ** c',
    },
    {
      code: 'Math.pow(a, b ** c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'a**b ** c',
    },
    {
      code: 'Math.pow(a ** b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: '(a ** b)**c',
    },
    {
      code: 'a ** Math.pow(b ** c, d ** e) ** f',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: 'a ** ((b ** c)**d ** e) ** f',
    },
    // doesn't remove already existing unnecessary parens around the whole expression
    {
      code: '(Math.pow(a, b))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: '(a**b)',
    },
    {
      code: 'foo + (Math.pow(a, b))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'foo + (a**b)',
    },
    {
      code: '(Math.pow(a, b)) + foo',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: '(a**b) + foo',
    },
    {
      code: '`${(Math.pow(a, b))}`',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '`${(a**b)}`',
    },
    // base and exponent with a higher precedence
    {
      code: 'Math.pow(2, 3)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: '2**3',
    },
    {
      code: 'Math.pow(a.foo, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a.foo**b',
    },
    {
      code: 'Math.pow(a, b.foo)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a**b.foo',
    },
    {
      code: 'Math.pow(a(), b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'a()**b',
    },
    {
      code: 'Math.pow(a, b())',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'a**b()',
    },
    {
      code: 'Math.pow(++a, ++b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '++a**++b',
    },
    {
      code: 'Math.pow(a++, ++b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a++**++b',
    },
    {
      code: 'Math.pow(a--, b--)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a--**b--',
    },
    {
      code: 'Math.pow(--a, b--)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '--a**b--',
    },
    // doesn't preserve unnecessary parens around base and exponent
    {
      code: 'Math.pow((a), (b))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a**b',
    },
    {
      code: 'Math.pow(((a)), ((b)))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'a**b',
    },
    {
      code: 'Math.pow((a.foo), b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'a.foo**b',
    },
    {
      code: 'Math.pow(a, (b.foo))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'a**b.foo',
    },
    {
      code: 'Math.pow((a()), b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a()**b',
    },
    {
      code: 'Math.pow(a, (b()))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a**b()',
    },
    // unary expressions are exception by the language
    {
      code: 'Math.pow(+a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: '(+a)**b',
    },
    {
      code: 'Math.pow(a, +b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: 'a**+b',
    },
    {
      code: 'Math.pow(-a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: '(-a)**b',
    },
    {
      code: 'Math.pow(a, -b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: 'a**-b',
    },
    {
      code: 'Math.pow(-2, 3)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: '(-2)**3',
    },
    {
      code: 'Math.pow(2, -3)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: '2**-3',
    },
    {
      code: 'async () => Math.pow(await a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'async () => (await a)**b',
    },
    {
      code: 'async () => Math.pow(a, await b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'async () => a**await b',
    },
    // base and exponent with a lower precedence
    {
      code: 'Math.pow(a * b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '(a * b)**c',
    },
    {
      code: 'Math.pow(a, b * c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a**(b * c)',
    },
    {
      code: 'Math.pow(a / b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '(a / b)**c',
    },
    {
      code: 'Math.pow(a, b / c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a**(b / c)',
    },
    {
      code: 'Math.pow(a + b, 3)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '(a + b)**3',
    },
    {
      code: 'Math.pow(2, a - b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '2**(a - b)',
    },
    {
      code: 'Math.pow(a + b, c + d)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: '(a + b)**(c + d)',
    },
    {
      code: 'Math.pow(a = b, c = d)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: '(a = b)**(c = d)',
    },
    {
      code: 'Math.pow(a += b, c -= d)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: '(a += b)**(c -= d)',
    },
    {
      code: 'Math.pow((a, b), (c, d))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: '(a, b)**(c, d)',
    },
    {
      code: 'function *f() { Math.pow(yield, yield) }',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: 'function *f() { (yield)**(yield) }',
    },
    {
      code: 'Math.pow((a + b), (c + d))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
      output: '(a + b)**(c + d)',
    },
    // token adjacency
    {
      code: 'a+Math.pow(b, c)+d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'a+b**c+d',
    },
    {
      code: 'a+Math.pow(++b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a+ ++b**c',
    },
    {
      code: '(a)+(Math).pow((++b), c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: '(a)+ ++b**c',
    },
    {
      code: 'Math.pow(a, b)in c',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: 'a**b in c',
    },
    {
      code: 'Math.pow(a, (b))in (c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'a**b in (c)',
    },
    {
      code: 'a+Math.pow(++b, c)in d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a+ ++b**c in d',
    },
    {
      code: 'a+Math.pow( ++b, c )in d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'a+ ++b**c in d',
    },
    {
      code: 'a+ Math.pow(++b, c) in d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'a+ ++b**c in d',
    },
    {
      code: 'a+/**/Math.pow(++b, c)/**/in d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'a+/**/++b**c/**/in d',
    },
    {
      code: 'a+(Math.pow(++b, c))in d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'a+(++b**c)in d',
    },
    {
      code: '+Math.pow(++a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: '+(++a**b)',
    },
    {
      code: 'Math.pow(a, b + c)in d',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'a**(b + c)in d',
    },
    {
      code: `Math.pow(a, b) + Math.pow(c,
 d)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 18,
          endLine: 2,
          endColumn: 4,
        },
      ],
      output: 'a**b + c**d',
    },
    {
      code: 'Math.pow(Math.pow(a, b), Math.pow(c, d))',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 40,
        },
      ],
      output: '(a**b)**c**d',
    },
    {
      code: 'Math.pow(a, b)**Math.pow(c, d)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: '(a**b)**c**d',
    },
    // no autofix branches still report
    {
      code: 'Math.pow()',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'Math.pow(a)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'Math.pow(a, b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'Math.pow(a, b, c, d)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'Math.pow(...a)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'Math.pow(...a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'Math.pow(a, ...b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'Math.pow(a, b, ...c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    // comments
    {
      code: '/* comment */Math.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: '/* comment */a**b',
    },
    {
      code: 'Math/**/.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `Math//
.pow(a, b)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `Math[//
'pow'](a, b)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 13,
        },
      ],
    },
    {
      code: "Math['pow'/**/](a, b)",
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'Math./**/pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'Math.pow/**/(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `Math.pow//
(a, b)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'Math.pow(/**/a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `Math.pow(a,//
 b)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 4,
        },
      ],
    },
    {
      code: 'Math.pow(a, b/**/)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `Math.pow(a, b//
)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'Math.pow(a, b)/* comment */;',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: 'a**b/* comment */;',
    },
    {
      code: `Math.pow(a, b)// comment
;`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: `a**b// comment
;`,
    },
    // Optional chaining
    {
      code: 'Math.pow?.(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'a**b',
    },
    {
      code: 'Math?.pow(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: 'a**b',
    },
    {
      code: 'Math?.pow?.(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'a**b',
    },
    {
      code: '(Math?.pow)(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'a**b',
    },
    {
      code: '(Math?.pow)?.(a, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'a**b',
    },
    // https://github.com/eslint/eslint/issues/17173
    {
      code: 'Math.pow(a, b as any)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'a**(b as any)',
    },
    {
      code: 'Math.pow(a as any, b)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: '(a as any)**b',
    },
    {
      code: 'Math.pow(a, b) as any',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: '(a**b) as any',
    },
    // https://github.com/eslint/eslint/issues/20987
    {
      code: 'Math.pow({a:1}.a, 2);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: '({a:1}.a**2);',
    },
    {
      code: 'Math.pow({a:1}.a, 2) + 100;',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: '({a:1}.a**2) + 100;',
    },
    {
      code: '(Math.pow({a:1}.a, 2));',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: '({a:1}.a**2);',
    },
    {
      code: '100 + Math.pow({a:1}.a, 2);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 27,
        },
      ],
      output: '100 + {a:1}.a**2;',
    },
    {
      code: 'Math.pow({a:1}.a + 100, 2);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
      output: '({a:1}.a + 100)**2;',
    },
    {
      code: 'Math.pow(function(){return 2}(), 3);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: '(function(){return 2}()**3);',
    },
    {
      code: 'Math.pow(function(){return 2}(), 3) + 100;',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: '(function(){return 2}()**3) + 100;',
    },
    {
      code: '(Math.pow(function(){return 2}(), 3));',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: '(function(){return 2}()**3);',
    },
    {
      code: '100 + Math.pow(function(){return 2}(), 3);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 42,
        },
      ],
      output: '100 + function(){return 2}()**3;',
    },
    {
      code: 'Math.pow(function(){return 2}() + 100, 3);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
      output: '(function(){return 2}() + 100)**3;',
    },
    {
      code: 'Math.pow(class{static x=2}.x, 4);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: '(class{static x=2}.x**4);',
    },
    {
      code: 'Math.pow(class{static x=2}.x, 4) + 100;',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: '(class{static x=2}.x**4) + 100;',
    },
    {
      code: '(Math.pow(class{static x=2}.x, 4));',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: '(class{static x=2}.x**4);',
    },
    {
      code: '100 + Math.pow(class{static x=2}.x, 4);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: '100 + class{static x=2}.x**4;',
    },
    {
      code: 'Math.pow(class{static x=2}.x + 100, 4);',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: '(class{static x=2}.x + 100)**4;',
    },
    // preceding semicolon
    {
      code: `foo
Math.pow(a + b, c)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
      output: `foo
;(a + b)**c`,
    },
    {
      code: `foo
Math.pow(+a, b)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
        },
      ],
      output: `foo
;(+a)**b`,
    },
    {
      code: `foo
Math.pow(-a, b)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
        },
      ],
      output: `foo
;(-a)**b`,
    },
    {
      code: `foo
Math.pow({a:1}.a, 2)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 21,
        },
      ],
      output: `foo
;({a:1}.a**2)`,
    },
    {
      code: `foo
Math.pow((a).b, c)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
      output: `foo
;(a).b**c`,
    },
    {
      code: `foo
Math.pow([a, b].find(fn), c)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 29,
        },
      ],
      output: `foo
;[a, b].find(fn)**c`,
    },
    {
      code: `foo
Math.pow(/regex/, 2)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 21,
        },
      ],
      output: `foo
;/regex/**2`,
    },
    {
      code: `foo
Math.pow(\`template literal\`, 2)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 32,
        },
      ],
      output: `foo
;\`template literal\`**2`,
    },
    {
      code: `foo
100 + Math.pow((a).b, c)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 25,
        },
      ],
      output: `foo
100 + (a).b**c`,
    },
    {
      code: `foo
Math.pow(a.b, c)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 17,
        },
      ],
      output: `foo
a.b**c`,
    },
    {
      code: 'Math.pow((a).b, c)',
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '(a).b**c',
    },
    {
      code: `foo;
Math.pow((a).b, c)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
      output: `foo;
(a).b**c`,
    },
    {
      code: `if (foo) {}
Math.pow((a).b, c)`,
      errors: [
        {
          messageId: 'useExponentiation',
          message: "Use the '**' operator instead of 'Math.pow'.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
      output: `if (foo) {}
(a).b**c`,
    },
  ],
});
