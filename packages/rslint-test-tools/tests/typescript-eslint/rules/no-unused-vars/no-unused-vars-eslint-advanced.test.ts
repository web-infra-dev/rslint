// The following tests are adapted from the tests in eslint.
// Original Code: https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/tests/lib/rules/no-unused-vars.js
// License      : https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/LICENSE

import { assignedError, definedError, ruleTester } from './eslint-test-helpers';

ruleTester.run('no-unused-vars', {
  invalid: [
    // https://github.com/eslint/eslint/issues/7124
    {
      code: '(function (a, b, c) {});',
      errors: [
        definedError('a', '. Allowed unused args must match /c/u'),
        definedError('b', '. Allowed unused args must match /c/u'),
      ],
      options: [{ argsIgnorePattern: 'c' }],
    },
    {
      code: '(function (a, b, { c, d }) {});',
      errors: [
        definedError('a', '. Allowed unused args must match /[cd]/u'),
        definedError('b', '. Allowed unused args must match /[cd]/u'),
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ argsIgnorePattern: '[cd]' }],
    },
    {
      code: '(function (a, b, { c, d }) {});',
      errors: [
        definedError('a', '. Allowed unused args must match /c/u'),
        definedError('b', '. Allowed unused args must match /c/u'),
        definedError('d', '. Allowed unused args must match /c/u'),
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ argsIgnorePattern: 'c' }],
    },
    {
      code: '(function (a, b, { c, d }) {});',
      errors: [
        definedError('a', '. Allowed unused args must match /d/u'),
        definedError('b', '. Allowed unused args must match /d/u'),
        definedError('c', '. Allowed unused args must match /d/u'),
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ argsIgnorePattern: 'd' }],
    },
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: `
/*global
foo*/
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '',
            varName: 'foo',
          },
          messageId: 'unusedVar',
        },
      ],
    },

    // https://github.com/eslint/eslint/issues/8442
    {
      code: `
(function ({ a }, b) {
  return b;
})();
      `,
      errors: [definedError('a')],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
(function ({ a }, { b, c }) {
  return b;
})();
      `,
      errors: [definedError('a'), definedError('c')],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },

    // https://github.com/eslint/eslint/issues/14325
    {
      code: `
let x = 0;
x++, (x = 0);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
x++, (x = 0);
x = 3;
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
x++, 0;
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
0, x++;
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
0, (1, x++);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
foo = (x++, 0);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
foo = ((0, x++), 0);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
(x += 1), 0;
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
0, (x += 1);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
0, (1, (x += 1));
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
foo = ((x += 1), 0);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let x = 0;
foo = ((0, (x += 1)), 0);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },

    // https://github.com/eslint/eslint/issues/14866
    {
      code: `
let z = 0;
(z = z + 1), (z = 2);
      `,
      errors: [{ ...assignedError('z') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let z = 0;
(z = z + 1), (z = 2);
z = 3;
      `,
      errors: [{ ...assignedError('z') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let z = 0;
(z = z + 1), (z = 2);
z = z + 3;
      `,
      errors: [{ ...assignedError('z') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let x = 0;
0, (x = x + 1);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let x = 0;
(x = x + 1), 0;
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let x = 0;
foo = ((0, (x = x + 1)), 0);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let x = 0;
foo = ((x = x + 1), 0);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let x = 0;
0, (1, (x = x + 1));
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
(function ({ a, b }, { c }) {
  return b;
})();
      `,
      errors: [definedError('a'), definedError('c')],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
(function ([a], b) {
  return b;
})();
      `,
      errors: [definedError('a')],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
(function ([a], [b, c]) {
  return b;
})();
      `,
      errors: [definedError('a'), definedError('c')],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
(function ([a, b], [c]) {
  return b;
})();
      `,
      errors: [definedError('a'), definedError('c')],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },

    // https://github.com/eslint/eslint/issues/9774
    {
      code: '(function (_a) {})();',
      errors: [definedError('_a')],
      options: [{ args: 'all', varsIgnorePattern: '^_' }],
    },
    {
      code: '(function (_a) {})();',
      errors: [definedError('_a')],
      options: [{ args: 'all', caughtErrorsIgnorePattern: '^_' }],
    },

    // https://github.com/eslint/eslint/issues/10982
    {
      code: `
var a = function () {
  a();
};
      `,
      errors: [assignedError('a')],
    },
    {
      code: `
var a = function () {
  return function () {
    a();
  };
};
      `,
      errors: [{ ...assignedError('a') }],
    },
    {
      code: `
const a = () => () => {
  a();
};
      `,
      errors: [{ ...assignedError('a') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let myArray = [1, 2, 3, 4].filter(x => x == 0);
myArray = myArray.filter(x => x == 1);
      `,
      errors: [{ ...assignedError('myArray') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
const a = 1;
a += 1;
      `,
      errors: [{ ...assignedError('a') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
const a = () => {
  a();
};
      `,
      errors: [{ ...assignedError('a') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },

    // https://github.com/eslint/eslint/issues/14324
    {
      code: `
let x = [];
x = x.concat(x);
      `,
      errors: [{ ...assignedError('x') }],
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
let a = 'a';
a = 10;
function foo() {
  a = 11;
  a = () => {
    a = 13;
  };
}
      `,
      errors: [
        {
          ...assignedError('a'),
        },
        {
          ...definedError('foo'),
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let foo;
init();
foo = foo + 2;
function init() {
  foo = 1;
}
      `,
      errors: [{ ...assignedError('foo') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
function foo(n) {
  if (n < 2) return 1;
  return n * foo(n - 1);
}
      `,
      errors: [{ ...definedError('foo') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },
    {
      code: `
let c = 'c';
c = 10;
function foo1() {
  c = 11;
  c = () => {
    c = 13;
  };
}

c = foo1;
      `,
      errors: [{ ...assignedError('c') }],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
    },

    // ignore class with static initialization block https://github.com/eslint/eslint/issues/17772
    {
      code: `
class Foo {
  static {}
}
      `,
      errors: [{ ...definedError('Foo') }],
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
      options: [{ ignoreClassWithStaticInitBlock: false }],
    },
    {
      code: `
class Foo {
  static {}
}
      `,
      errors: [{ ...definedError('Foo') }],
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
    },
    {
      code: `
class Foo {
  static {
    var bar;
  }
}
      `,
      errors: [{ ...definedError('bar') }],
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    {
      code: 'class Foo {}',
      errors: [{ ...definedError('Foo') }],
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    {
      code: `
class Foo {
  static bar;
}
      `,
      errors: [{ ...definedError('Foo') }],
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    {
      code: `
class Foo {
  static bar() {}
}
      `,
      errors: [{ ...definedError('Foo') }],
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
  ],
  valid: [
    // https://github.com/eslint/eslint/issues/10952
    {
      // skip: uses ESLint markVariableAsUsed API not available in rslint
      skip: true,
      code: `
/*eslint @rule-tester/use-every-a:1*/ !function (b, a) {
  return 1;
};
      `,
    },

    // https://github.com/eslint/eslint/issues/10982
    `
var a = function () {
  a();
};
a();
    `,
    `
var a = function () {
  return function () {
    a();
  };
};
a();
    `,
    {
      code: `
const a = () => {
  a();
};
a();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },
    {
      code: `
const a = () => () => {
  a();
};
a();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2015 } },
    },

    // export * as ns from "source"
    {
      code: "export * as ns from 'source';",
      languageOptions: {
        parserOptions: { ecmaVersion: 2020, sourceType: 'module' },
      },
    },

    // import.meta
    {
      code: 'import.meta;',
      languageOptions: {
        parserOptions: { ecmaVersion: 2020, sourceType: 'module' },
      },
    },

    // ignore class with static initialization block https://github.com/eslint/eslint/issues/17772
    {
      code: `
class Foo {
  static {}
}
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    {
      code: `
class Foo {
  static {}
}
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
      options: [
        { ignoreClassWithStaticInitBlock: true, varsIgnorePattern: '^_' },
      ],
    },
    {
      code: `
class Foo {
  static {}
}
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2022 } },
      options: [
        { ignoreClassWithStaticInitBlock: false, varsIgnorePattern: '^Foo' },
      ],
    },
  ],
});
