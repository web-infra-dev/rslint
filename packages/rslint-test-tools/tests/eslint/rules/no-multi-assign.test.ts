import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-multi-assign', {
  valid: [
    // ---- ESLint upstream valid cases ----
    'var a, b, c,\nd = 0;',
    'var a = 1; var b = 2; var c = 3;\nvar d = 0;',
    'var a = 1 + (b === 10 ? 5 : 4);',
    'const a = 1, b = 2, c = 3;',
    'const a = 1;\nconst b = 2;\n const c = 3;',
    'for(var a = 0, b = 0;;){}',
    'for(let a = 0, b = 0;;){}',
    'for(const a = 0, b = 0;;){}',
    'export let a, b;',
    'export let a,\n b = 0;',
    {
      code: 'const x = {};const y = {};x.one = y.one = 1;',
      options: { ignoreNonDeclaration: true },
    },
    {
      code: 'let a, b;a = b = 1',
      options: { ignoreNonDeclaration: true },
    },
    'class C { [foo = 0] = 0 }',

    // ---- Additional edge cases ----
    '({ [a = 1]: 1 })',
    '({ a: b = 1 })',
    'let { a = 1 } = {};',
    'a = 1;',
    'a += 1;',
    'a ||= 1;',
    '[a, b] = [1, 2];',
    '({ a, b } = { a: 1, b: 2 });',
    'let x: number = 1;',
    'class C { x; }',
    'class C { x = 1; }',
    'class C { declare x: number; }',
    {
      code: "let a; let b; a = b = 'baz';",
      options: { ignoreNonDeclaration: true },
    },
    {
      code: 'let a, b; a = b ||= 1;',
      options: { ignoreNonDeclaration: true },
    },
    'var { a = 1 } = obj;',
    'var { a = b = c } = obj;',
    'var [ a = b = c ] = arr;',
    'var a = b + (c = d);',
    'var a = (b = c) + d;',
    'class C { [foo = bar] = 1 }',
    '({ [a = b]: 1 })',
    'var a = 1, b = 2;',
    'a ??= 1;',
    'var a = cond ? (b = c) : (d = e);',
    'a = cond ? b = c : d;',
    'var a = (b = 1, c = 2);',
    'fn(a = 1, b = 2);',
    'function f() { return a = 1; }',
    'function* g() { yield a = 1; }',
    'var a = b++;',
    'var a = !b;',
    'class C { handler = () => { let a; a = 1; }; }',
    'class C { handler = () => { var a; a = b; }; }',
    'let a: number;',
    'var obj = { a: b = 1 };',
    'var x = { value: a = 1 };',
    'export const PI = 3.14;',
    'var a = /* hi */ 1;',
    'var a = b as number;',
    'var a = (b satisfies number);',
    'var a = b!;',
    'var a = b?.c;',
    'var a = tag`x`;',
    'var a = fn();',
    'class C { a = 1; b = 2; static c = 3; }',
    'class C { constructor(public x = 1) {} }',
    'var a = (() => b = 1)();',
  ],
  invalid: [
    // ---- ESLint upstream invalid cases ----
    {
      code: 'var a = b = c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = b = c = d;',
      errors: [
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
      ],
    },
    {
      code: 'let foo = bar = cee = 100;',
      errors: [
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
      ],
    },
    {
      code: 'a=b=c=d=e',
      errors: [
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
      ],
    },
    {
      code: 'a=b=c',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'a\n=b\n=c',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = (b) = (((c)))',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = ((b)) = (c)',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = b = ( (c * 12) + 2)',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a =\n((b))\n = (c)',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: "a = b = '=' + c + 'foo';",
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'a = b = 7 * 12 + 5;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'const x = {};\nconst y = x.one = 1;',
      options: { ignoreNonDeclaration: true },
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'let a, b;a = b = 1',
      options: {},
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: "let x, y;x = y = 'baz'",
      options: { ignoreNonDeclaration: false },
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'const a = b = 1',
      options: { ignoreNonDeclaration: true },
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class C { field = foo = 0 }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class C { field = foo = 0 }',
      options: { ignoreNonDeclaration: true },
      errors: [{ messageId: 'unexpectedChain' }],
    },

    // ---- Additional edge cases ----
    {
      code: 'var a = (b = c);',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = ((b = c));',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'a = (b = c);',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = b += c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = b ||= c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'a = b ??= c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'a += b = c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'for (let i = j = 0;;) {}',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class C { static x = y = 1 }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class C { #x = y = 1 }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'let a: number = b = 1;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'let a = b = 1',
      options: { ignoreNonDeclaration: true },
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = b = c = d = e = f;',
      errors: [
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
      ],
    },
    {
      code: 'var a = b = c, d = e = f;',
      errors: [
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
      ],
    },
    {
      code: 'a = b = c ??= d',
      errors: [
        { messageId: 'unexpectedChain' },
        { messageId: 'unexpectedChain' },
      ],
    },
    {
      code: '`${a = b = c}`',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: '() => a = b = c',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class C { x = (y = 1); }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class A { x = class { y = z = 1 } }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = (() => { return b = c = 1; })();',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = b /* hi */ = c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a =\n  b =\n  c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'const f = () => { return a = b = 1; };',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'obj.a = obj.b = 1;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: "obj['a'] = obj['b'] = 1;",
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'this.x = this.y = 1;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'a[k] = b[k] = v;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class C { static x = (y = 1); }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = ((b) = c);',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = (b as any) = c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = b **= c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class C { x!: number = y = 1; }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: '{ using a = b = c; }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'export let a = b = c;',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: '{ let a = b = 1; }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'switch (x) { case 1: a = b = 2; break; }',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'var a = ((((b = c))));',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'a = ((b = c));',
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'let a = (b = 1);',
      options: { ignoreNonDeclaration: true },
      errors: [{ messageId: 'unexpectedChain' }],
    },
    {
      code: 'class C { x = (y = 1); }',
      options: { ignoreNonDeclaration: true },
      errors: [{ messageId: 'unexpectedChain' }],
    },
  ],
});
