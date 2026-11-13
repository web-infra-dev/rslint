// Mirrors the complete ESLint v10.8.1 upstream RuleTester suite. Native-only
// edge shapes and branch lock-ins belong in the Go extras suite.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { config } from '../config';

function valid(
  code: string,
  options?: ValidTestCase['options'],
  ecmaVersion = 5,
): ValidTestCase {
  return {
    code,
    ...(options ? { options } : {}),
    languageOptions: { ecmaVersion, sourceType: 'script' },
  };
}

function error(
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
  message: string,
): TestCase['errors'][number] {
  return {
    messageId: 'moveDeclToRoot',
    message,
    line,
    column,
    endLine,
    endColumn,
  };
}

function invalid(
  code: string,
  options: ValidTestCase['options'] | undefined,
  ecmaVersion: number,
  ...errors: TestCase['errors']
): TestCase {
  return { ...valid(code, options, ecmaVersion), errors };
}

new RuleTester(config).run('no-inner-declarations', {
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
      `'use strict'
 if (test) { function doSomething() { } }`,
      [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      2022,
    ),
    valid(
      `'use strict'
 if (test) { function doSomething() { } }`,
      ['functions'],
      2022,
    ),
    valid(
      `function foo() {'use strict'
 if (test) { function doSomething() { } } }`,
      [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      2015,
    ),
    {
      code: 'function foo() { { function bar() { } } }',
      options: [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
    valid(
      'class C { method() { if(test) { function somethingElse() { } } } }',
      [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      2022,
    ),
    valid(
      'const C = class { method() { if(test) { function somethingElse() { } } } }',
      [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      2022,
    ),
  ],
  invalid: [
    invalid(
      'if (test) { function doSomething() { } }',
      ['both'],
      5,
      error(1, 13, 1, 39, 'Move function declaration to program root.'),
    ),
    invalid(
      'if (foo) var a; ',
      ['both'],
      5,
      error(1, 10, 1, 16, 'Move variable declaration to program root.'),
    ),
    invalid(
      'if (foo) /* some comments */ var a; ',
      ['both'],
      5,
      error(1, 30, 1, 36, 'Move variable declaration to program root.'),
    ),
    invalid(
      'if (foo){ function f(){ if(bar){ var a; } } }',
      ['both'],
      5,
      error(1, 11, 1, 44, 'Move function declaration to program root.'),
      error(1, 34, 1, 40, 'Move variable declaration to function body root.'),
    ),
    invalid(
      'if (foo) function f(){ if(bar) var a; }',
      ['both'],
      5,
      error(1, 10, 1, 40, 'Move function declaration to program root.'),
      error(1, 32, 1, 38, 'Move variable declaration to function body root.'),
    ),
    invalid(
      'if (foo) { var fn = function(){} } ',
      ['both'],
      5,
      error(1, 12, 1, 33, 'Move variable declaration to program root.'),
    ),
    invalid(
      'if (foo)  function f(){} ',
      undefined,
      5,
      error(1, 11, 1, 25, 'Move function declaration to program root.'),
    ),
    invalid(
      'function bar() { if (foo) function f(){}; }',
      ['both'],
      5,
      error(1, 27, 1, 41, 'Move function declaration to function body root.'),
    ),
    invalid(
      'function bar() { if (foo) var a; }',
      ['both'],
      5,
      error(1, 27, 1, 33, 'Move variable declaration to function body root.'),
    ),
    invalid(
      'if (foo) { var a; }',
      ['both'],
      5,
      error(1, 12, 1, 18, 'Move variable declaration to program root.'),
    ),
    invalid(
      'function doSomething() { do { function somethingElse() { } } while (test); }',
      undefined,
      5,
      error(1, 31, 1, 59, 'Move function declaration to function body root.'),
    ),
    invalid(
      '(function() { if (test) { function doSomething() { } } }());',
      undefined,
      5,
      error(1, 27, 1, 53, 'Move function declaration to function body root.'),
    ),
    invalid(
      'while (test) { var foo; }',
      ['both'],
      5,
      error(1, 16, 1, 24, 'Move variable declaration to program root.'),
    ),
    invalid(
      'function doSomething() { if (test) { var foo = 42; } }',
      ['both'],
      5,
      error(1, 38, 1, 51, 'Move variable declaration to function body root.'),
    ),
    invalid(
      '(function() { if (test) { var foo; } }());',
      ['both'],
      5,
      error(1, 27, 1, 35, 'Move variable declaration to function body root.'),
    ),
    invalid(
      'const doSomething = () => { if (test) { var foo = 42; } }',
      ['both'],
      2015,
      error(1, 41, 1, 54, 'Move variable declaration to function body root.'),
    ),
    invalid(
      'class C { method() { if(test) { var foo; } } }',
      ['both'],
      2015,
      error(1, 33, 1, 41, 'Move variable declaration to function body root.'),
    ),
    invalid(
      'class C { static { if (test) { var foo; } } }',
      ['both'],
      2022,
      error(
        1,
        32,
        1,
        40,
        'Move variable declaration to class static block body root.',
      ),
    ),
    invalid(
      'class C { static { if (test) { function foo() {} } } }',
      [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      2022,
      error(
        1,
        32,
        1,
        49,
        'Move function declaration to class static block body root.',
      ),
    ),
    invalid(
      'class C { static { if (test) { if (anotherTest) { var foo; } } } }',
      ['both'],
      2022,
      error(
        1,
        51,
        1,
        59,
        'Move variable declaration to class static block body root.',
      ),
    ),
    invalid(
      'if (test) { function doSomething() { } }',
      [
        'both',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      5,
      error(1, 13, 1, 39, 'Move function declaration to program root.'),
    ),
    invalid(
      'if (test) { function doSomething() { } }',
      [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      2022,
      error(1, 13, 1, 39, 'Move function declaration to program root.'),
    ),
    invalid(
      `'use strict'
 if (test) { function doSomething() { } }`,
      [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      2022,
      error(2, 14, 2, 40, 'Move function declaration to program root.'),
    ),
    invalid(
      `'use strict'
 if (test) { function doSomething() { } }`,
      [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      5,
      error(2, 14, 2, 40, 'Move function declaration to program root.'),
    ),
    invalid(
      `'use strict'
 if (test) { function doSomething() { } }`,
      [
        'both',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      5,
      error(2, 14, 2, 40, 'Move function declaration to program root.'),
    ),
    invalid(
      `function foo() {'use strict'
 { function bar() { } } }`,
      [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      2022,
      error(2, 4, 2, 22, 'Move function declaration to function body root.'),
    ),
    invalid(
      `function foo() {'use strict'
 { function bar() { } } }`,
      [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      5,
      error(2, 4, 2, 22, 'Move function declaration to function body root.'),
    ),
    invalid(
      `function doSomething() { 'use strict'
 do { function somethingElse() { } } while (test); }`,
      [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      5,
      error(2, 7, 2, 35, 'Move function declaration to function body root.'),
    ),
    invalid(
      `{ function foo () {'use strict'
 console.log('foo called'); } }`,
      ['both'],
      2022,
      error(1, 3, 2, 30, 'Move function declaration to program root.'),
    ),
  ],
});
