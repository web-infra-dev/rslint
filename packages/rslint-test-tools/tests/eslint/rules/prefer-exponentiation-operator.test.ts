import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const invalid = (code: string, count = 1) => ({
  code,
  errors: Array.from({ length: count }, () => ({
    messageId: 'useExponentiation',
  })),
});

ruleTester.run('prefer-exponentiation-operator', {
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
    {
      code: '/* globals Math:off*/ Math.pow(a, b)',
      skip: true,
    },
    'let Math; Math.pow(a, b);',
    'if (foo) { const Math = 1; Math.pow(a, b); }',
    'var x = function Math() { Math.pow(a, b); }',
    'function foo(Math) { Math.pow(a, b); }',
    'function foo() { Math.pow(a, b); var Math; }',
    // Mirrors upstream ecmaVersion 2019 / 6 / 2017 cases; rslint does not expose ecmaVersion-specific globalThis availability.
    {
      code: 'globalThis.Math.pow(a, b)',
      skip: true,
    },
    {
      code: 'globalThis.Math.pow(a, b)',
      skip: true,
    },
    {
      code: 'globalThis.Math.pow(a, b)',
      skip: true,
    },
    `
      var globalThis = bar;
      globalThis.Math.pow(a, b)
    `,
    'class C { #pow; foo() { Math.#pow(a, b); } }',
  ],
  invalid: [
    invalid('Math.pow(a, b)'),
    invalid('(Math).pow(a, b)'),
    invalid("Math['pow'](a, b)"),
    invalid("(Math)['pow'](a, b)"),
    invalid('var x=Math\n.  pow( a, \n b )'),
    invalid('globalThis.Math.pow(a, b)'),
    invalid("globalThis.Math['pow'](a, b)"),

    // able to catch some workarounds
    invalid('Math[`pow`](a, b)'),
    invalid("Math[`${'pow'}`](a, b)"),
    invalid("Math['p' + 'o' + 'w'](a, b)"),

    // non-expression parents that don't require parens
    invalid('var x = Math.pow(a, b);'),
    invalid('if(Math.pow(a, b)){}'),
    invalid('for(;Math.pow(a, b);){}'),
    invalid('switch(foo){ case Math.pow(a, b): break; }'),
    invalid('{ foo: Math.pow(a, b) }'),
    invalid('function foo(bar, baz = Math.pow(a, b), quux){}'),
    invalid('`${Math.pow(a, b)}`'),

    // non-expression parents that do require parens
    invalid('class C extends Math.pow(a, b) {}'),

    // parents with a higher precedence
    invalid('+ Math.pow(a, b)'),
    invalid('- Math.pow(a, b)'),
    invalid('! Math.pow(a, b)'),
    invalid('typeof Math.pow(a, b)'),
    invalid('void Math.pow(a, b)'),
    invalid('Math.pow(a, b) .toString()'),
    invalid('Math.pow(a, b) ()'),
    invalid('Math.pow(a, b) ``'),
    invalid('(class extends Math.pow(a, b) {})'),

    // already parenthesised, shouldn't insert extra parens
    invalid('+(Math.pow(a, b))'),
    invalid('(Math.pow(a, b)).toString()'),
    invalid('(class extends (Math.pow(a, b)) {})'),
    invalid('class C extends (Math.pow(a, b)) {}'),

    // parents with a higher precedence, but the expression's role doesn't require parens
    invalid('f(Math.pow(a, b))'),
    invalid('f(foo, Math.pow(a, b))'),
    invalid('f(Math.pow(a, b), foo)'),
    invalid('f(foo, Math.pow(a, b), bar)'),
    invalid('new F(Math.pow(a, b))'),
    invalid('new F(foo, Math.pow(a, b))'),
    invalid('new F(Math.pow(a, b), foo)'),
    invalid('new F(foo, Math.pow(a, b), bar)'),
    invalid('obj[Math.pow(a, b)]'),
    invalid('[foo, Math.pow(a, b), bar]'),

    // parents with a lower precedence
    invalid('a * Math.pow(b, c)'),
    invalid('Math.pow(a, b) * c'),
    invalid('a + Math.pow(b, c)'),
    invalid('Math.pow(a, b)/c'),
    invalid('a < Math.pow(b, c)'),
    invalid('Math.pow(a, b) > c'),
    invalid('a === Math.pow(b, c)'),
    invalid('a ? Math.pow(b, c) : d'),
    invalid('a = Math.pow(b, c)'),
    invalid('a += Math.pow(b, c)'),
    invalid('function *f() { yield Math.pow(a, b) }'),
    invalid('a, Math.pow(b, c), d'),

    // '**' is right-associative, that applies to both parent and child nodes
    invalid('a ** Math.pow(b, c)'),
    invalid('Math.pow(a, b) ** c'),
    invalid('Math.pow(a, b ** c)'),
    invalid('Math.pow(a ** b, c)'),
    invalid('a ** Math.pow(b ** c, d ** e) ** f'),

    // doesn't remove already existing unnecessary parens around the whole expression
    invalid('(Math.pow(a, b))'),
    invalid('foo + (Math.pow(a, b))'),
    invalid('(Math.pow(a, b)) + foo'),
    invalid('`${(Math.pow(a, b))}`'),

    // base and exponent with a higher precedence
    invalid('Math.pow(2, 3)'),
    invalid('Math.pow(a.foo, b)'),
    invalid('Math.pow(a, b.foo)'),
    invalid('Math.pow(a(), b)'),
    invalid('Math.pow(a, b())'),
    invalid('Math.pow(++a, ++b)'),
    invalid('Math.pow(a++, ++b)'),
    invalid('Math.pow(a--, b--)'),
    invalid('Math.pow(--a, b--)'),

    // doesn't preserve unnecessary parens around base and exponent
    invalid('Math.pow((a), (b))'),
    invalid('Math.pow(((a)), ((b)))'),
    invalid('Math.pow((a.foo), b)'),
    invalid('Math.pow(a, (b.foo))'),
    invalid('Math.pow((a()), b)'),
    invalid('Math.pow(a, (b()))'),

    // unary expressions are exception by the language
    invalid('Math.pow(+a, b)'),
    invalid('Math.pow(a, +b)'),
    invalid('Math.pow(-a, b)'),
    invalid('Math.pow(a, -b)'),
    invalid('Math.pow(-2, 3)'),
    invalid('Math.pow(2, -3)'),
    invalid('async () => Math.pow(await a, b)'),
    invalid('async () => Math.pow(a, await b)'),

    // base and exponent with a lower precedence
    invalid('Math.pow(a * b, c)'),
    invalid('Math.pow(a, b * c)'),
    invalid('Math.pow(a / b, c)'),
    invalid('Math.pow(a, b / c)'),
    invalid('Math.pow(a + b, 3)'),
    invalid('Math.pow(2, a - b)'),
    invalid('Math.pow(a + b, c + d)'),
    invalid('Math.pow(a = b, c = d)'),
    invalid('Math.pow(a += b, c -= d)'),
    invalid('Math.pow((a, b), (c, d))'),
    invalid('function *f() { Math.pow(yield, yield) }'),
    invalid('Math.pow((a + b), (c + d))'),

    // token adjacency
    invalid('a+Math.pow(b, c)+d'),
    invalid('a+Math.pow(++b, c)'),
    invalid('(a)+(Math).pow((++b), c)'),
    invalid('Math.pow(a, b)in c'),
    invalid('Math.pow(a, (b))in (c)'),
    invalid('a+Math.pow(++b, c)in d'),
    invalid('a+Math.pow( ++b, c )in d'),
    invalid('a+ Math.pow(++b, c) in d'),
    invalid('a+/**/Math.pow(++b, c)/**/in d'),
    invalid('a+(Math.pow(++b, c))in d'),
    invalid('+Math.pow(++a, b)'),
    invalid('Math.pow(a, b + c)in d'),

    invalid('Math.pow(a, b) + Math.pow(c,\n d)', 2),
    invalid('Math.pow(Math.pow(a, b), Math.pow(c, d))', 3),
    invalid('Math.pow(a, b)**Math.pow(c, d)', 2),

    // no autofix branches still report
    invalid('Math.pow()'),
    invalid('Math.pow(a)'),
    invalid('Math.pow(a, b, c)'),
    invalid('Math.pow(a, b, c, d)'),
    invalid('Math.pow(...a)'),
    invalid('Math.pow(...a, b)'),
    invalid('Math.pow(a, ...b)'),
    invalid('Math.pow(a, b, ...c)'),

    // comments
    invalid('/* comment */Math.pow(a, b)'),
    invalid('Math/**/.pow(a, b)'),
    invalid('Math//\n.pow(a, b)'),
    invalid("Math[//\n'pow'](a, b)"),
    invalid("Math['pow'/**/](a, b)"),
    invalid('Math./**/pow(a, b)'),
    invalid('Math.pow/**/(a, b)'),
    invalid('Math.pow//\n(a, b)'),
    invalid('Math.pow(/**/a, b)'),
    invalid('Math.pow(a,//\n b)'),
    invalid('Math.pow(a, b/**/)'),
    invalid('Math.pow(a, b//\n)'),
    invalid('Math.pow(a, b)/* comment */;'),
    invalid('Math.pow(a, b)// comment\n;'),

    // Optional chaining
    invalid('Math.pow?.(a, b)'),
    invalid('Math?.pow(a, b)'),
    invalid('Math?.pow?.(a, b)'),
    invalid('(Math?.pow)(a, b)'),
    invalid('(Math?.pow)?.(a, b)'),

    // https://github.com/eslint/eslint/issues/17173
    invalid('Math.pow(a, b as any)'),
    invalid('Math.pow(a as any, b)'),
    invalid('Math.pow(a, b) as any'),

    // https://github.com/eslint/eslint/issues/20987
    invalid('Math.pow({a:1}.a, 2);'),
    invalid('Math.pow({a:1}.a, 2) + 100;'),
    invalid('(Math.pow({a:1}.a, 2));'),
    invalid('100 + Math.pow({a:1}.a, 2);'),
    invalid('Math.pow({a:1}.a + 100, 2);'),
    invalid('Math.pow(function(){return 2}(), 3);'),
    invalid('Math.pow(function(){return 2}(), 3) + 100;'),
    invalid('(Math.pow(function(){return 2}(), 3));'),
    invalid('100 + Math.pow(function(){return 2}(), 3);'),
    invalid('Math.pow(function(){return 2}() + 100, 3);'),
    invalid('Math.pow(class{static x=2}.x, 4);'),
    invalid('Math.pow(class{static x=2}.x, 4) + 100;'),
    invalid('(Math.pow(class{static x=2}.x, 4));'),
    invalid('100 + Math.pow(class{static x=2}.x, 4);'),
    invalid('Math.pow(class{static x=2}.x + 100, 4);'),

    // preceding semicolon
    invalid('foo\nMath.pow(a + b, c)'),
    invalid('foo\nMath.pow(+a, b)'),
    invalid('foo\nMath.pow(-a, b)'),
    invalid('foo\nMath.pow({a:1}.a, 2)'),
    invalid('foo\nMath.pow((a).b, c)'),
    invalid('foo\nMath.pow([a, b].find(fn), c)'),
    invalid('foo\nMath.pow(/regex/, 2)'),
    invalid('foo\nMath.pow(`template literal`, 2)'),
    invalid('foo\n100 + Math.pow((a).b, c)'),
    invalid('foo\nMath.pow(a.b, c)'),
    invalid('Math.pow((a).b, c)'),
    invalid('foo;\nMath.pow((a).b, c)'),
    invalid('if (foo) {}\nMath.pow((a).b, c)'),
  ],
});
