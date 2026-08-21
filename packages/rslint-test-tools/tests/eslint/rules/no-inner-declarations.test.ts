import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

function valid(code: string, options?: unknown[], ecmaVersion = 5) {
  return {
    code,
    ...(options ? { options: options as any } : {}),
    languageOptions: { ecmaVersion } as any,
  };
}

function error(
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) {
  return { messageId: 'moveDeclToRoot', line, column, endLine, endColumn };
}

function invalid(
  code: string,
  options: unknown[] | undefined,
  ecmaVersion: number,
  ...errors: ReturnType<typeof error>[]
) {
  return {
    code,
    ...(options ? { options: options as any } : {}),
    languageOptions: { ecmaVersion } as any,
    errors,
  };
}

// Mirrors the complete ESLint v10.8.1 upstream RuleTester suite. Native-only
// edge shapes and branch lock-ins belong in the Go extras suite.
ruleTester.run('no-inner-declarations', {
  valid: [
    valid('function doSomething() { }'),
    valid('function doSomething() { function somethingElse() { } }'),
    valid('(function() { function doSomething() { } }());'),
    valid('if (test) { var fn = function() { }; }'),
    valid('if (test) { var fn = function expr() { }; }'),
    valid('function decl() { var fn = function expr() { }; }'),
    valid('function decl(arg) { var fn; if (arg) { fn = function() { }; } }'),
    valid(
      'var x = {doSomething() {function doSomethingElse() {}}}',
      undefined,
      2015,
    ),
    valid(
      'function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }',
      undefined,
      2015,
    ),
    valid(
      'function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }',
    ),
    valid('if (test) { var foo; }'),
    valid('if (test) { let x = 1; }', ['both'], 2015),
    valid('if (test) { const x = 1; }', ['both'], 2015),
    valid('if (test) { using x = 1; }', ['both'], 2026),
    valid('if (test) { await using x = 1; }', ['both'], 2026),
    valid('function doSomething() { while (test) { var foo; } }'),
    valid('var foo;', ['both']),
    valid('var foo = 42;', ['both']),
    valid('function doSomething() { var foo; }', ['both']),
    valid('(function() { var foo; }());', ['both']),
    valid('foo(() => { function bar() { } });', undefined, 2015),
    valid('var fn = () => {var foo;}', ['both'], 2015),
    valid('var x = {doSomething() {var foo;}}', ['both'], 2015),
    valid('export var foo;', ['both'], 2015),
    valid('export function bar() {}', ['both'], 2015),
    valid('export default function baz() {}', ['both'], 2015),
    valid('exports.foo = () => {}', ['both'], 2015),
    valid('exports.foo = function(){}', ['both']),
    valid('module.exports = function foo(){}', ['both']),
    valid('class C { method() { function foo() {} } }', ['both'], 2022),
    valid('class C { method() { var x; } }', ['both'], 2022),
    valid('class C { static { function foo() {} } }', ['both'], 2022),
    valid('class C { static { var x; } }', ['both'], 2022),
    valid(
      "'use strict'\n if (test) { function doSomething() { } }",
      ['functions', { blockScopedFunctions: 'allow' }],
      2022,
    ),
    valid(
      "'use strict'\n if (test) { function doSomething() { } }",
      ['functions'],
      2022,
    ),
    valid(
      "function foo() {'use strict'\n if (test) { function doSomething() { } } }",
      ['functions', { blockScopedFunctions: 'allow' }],
      2015,
    ),
    {
      ...valid(
        'function foo() { { function bar() { } } }',
        ['functions', { blockScopedFunctions: 'allow' }],
        2022,
      ),
      // SKIP: rslint does not support an authored sourceType: module override
      // for native rules, and this upstream case has no module syntax.
      skip: true,
    },
    valid(
      'class C { method() { if(test) { function somethingElse() { } } } }',
      ['functions', { blockScopedFunctions: 'allow' }],
      2022,
    ),
    valid(
      'const C = class { method() { if(test) { function somethingElse() { } } } }',
      ['functions', { blockScopedFunctions: 'allow' }],
      2022,
    ),
  ],
  invalid: [
    invalid(
      'if (test) { function doSomething() { } }',
      ['both'],
      5,
      error(1, 13, 1, 39),
    ),
    invalid('if (foo) var a; ', ['both'], 5, error(1, 10, 1, 16)),
    invalid(
      'if (foo) /* some comments */ var a; ',
      ['both'],
      5,
      error(1, 30, 1, 36),
    ),
    invalid(
      'if (foo){ function f(){ if(bar){ var a; } } }',
      ['both'],
      5,
      error(1, 11, 1, 44),
      error(1, 34, 1, 40),
    ),
    invalid(
      'if (foo) function f(){ if(bar) var a; }',
      ['both'],
      5,
      error(1, 10, 1, 40),
      error(1, 32, 1, 38),
    ),
    invalid(
      'if (foo) { var fn = function(){} } ',
      ['both'],
      5,
      error(1, 12, 1, 33),
    ),
    invalid('if (foo)  function f(){} ', undefined, 5, error(1, 11, 1, 25)),
    invalid(
      'function bar() { if (foo) function f(){}; }',
      ['both'],
      5,
      error(1, 27, 1, 41),
    ),
    invalid(
      'function bar() { if (foo) var a; }',
      ['both'],
      5,
      error(1, 27, 1, 33),
    ),
    invalid('if (foo) { var a; }', ['both'], 5, error(1, 12, 1, 18)),
    invalid(
      'function doSomething() { do { function somethingElse() { } } while (test); }',
      undefined,
      5,
      error(1, 31, 1, 59),
    ),
    invalid(
      '(function() { if (test) { function doSomething() { } } }());',
      undefined,
      5,
      error(1, 27, 1, 53),
    ),
    invalid('while (test) { var foo; }', ['both'], 5, error(1, 16, 1, 24)),
    invalid(
      'function doSomething() { if (test) { var foo = 42; } }',
      ['both'],
      5,
      error(1, 38, 1, 51),
    ),
    invalid(
      '(function() { if (test) { var foo; } }());',
      ['both'],
      5,
      error(1, 27, 1, 35),
    ),
    invalid(
      'const doSomething = () => { if (test) { var foo = 42; } }',
      ['both'],
      2015,
      error(1, 41, 1, 54),
    ),
    invalid(
      'class C { method() { if(test) { var foo; } } }',
      ['both'],
      2015,
      error(1, 33, 1, 41),
    ),
    invalid(
      'class C { static { if (test) { var foo; } } }',
      ['both'],
      2022,
      error(1, 32, 1, 40),
    ),
    invalid(
      'class C { static { if (test) { function foo() {} } } }',
      ['both', { blockScopedFunctions: 'disallow' }],
      2022,
      error(1, 32, 1, 49),
    ),
    invalid(
      'class C { static { if (test) { if (anotherTest) { var foo; } } } }',
      ['both'],
      2022,
      error(1, 51, 1, 59),
    ),
    invalid(
      'if (test) { function doSomething() { } }',
      ['both', { blockScopedFunctions: 'allow' }],
      5,
      error(1, 13, 1, 39),
    ),
    invalid(
      'if (test) { function doSomething() { } }',
      ['both', { blockScopedFunctions: 'disallow' }],
      2022,
      error(1, 13, 1, 39),
    ),
    invalid(
      "'use strict'\n if (test) { function doSomething() { } }",
      ['both', { blockScopedFunctions: 'disallow' }],
      2022,
      error(2, 14, 2, 40),
    ),
    invalid(
      "'use strict'\n if (test) { function doSomething() { } }",
      ['both', { blockScopedFunctions: 'disallow' }],
      5,
      error(2, 14, 2, 40),
    ),
    invalid(
      "'use strict'\n if (test) { function doSomething() { } }",
      ['both', { blockScopedFunctions: 'allow' }],
      5,
      error(2, 14, 2, 40),
    ),
    invalid(
      "function foo() {'use strict'\n { function bar() { } } }",
      ['both', { blockScopedFunctions: 'disallow' }],
      2022,
      error(2, 4, 2, 22),
    ),
    invalid(
      "function foo() {'use strict'\n { function bar() { } } }",
      ['both', { blockScopedFunctions: 'disallow' }],
      5,
      error(2, 4, 2, 22),
    ),
    invalid(
      "function doSomething() { 'use strict'\n do { function somethingElse() { } } while (test); }",
      ['both', { blockScopedFunctions: 'disallow' }],
      5,
      error(2, 7, 2, 35),
    ),
    invalid(
      "{ function foo () {'use strict'\n console.log('foo called'); } }",
      ['both'],
      2022,
      error(1, 3, 2, 30),
    ),
  ],
});
