// The following tests are adapted from the tests in eslint.
// Original Code: https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/tests/lib/rules/no-unused-vars.js
// License      : https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/LICENSE

import {
  assignedError,
  definedError,
  ruleTester,
} from './eslint-test-helpers';

ruleTester.run('no-unused-vars', {
  invalid: [
    // ignore pattern
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
var _a;
var b;
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '. Allowed unused vars must match /^_/u',
            varName: 'b',
          },
          messageId: 'unusedVar',
        },
      ],
      options: [{ vars: 'all', varsIgnorePattern: '^_' }],
    },
    {
      // skip: top-level `var` without type annotation in TS project mode
      skip: true,
      code: `
var a;
function foo() {
  var _b;
  var c_;
}
foo();
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '. Allowed unused vars must match /^_/u',
            varName: 'c_',
          },
          messageId: 'unusedVar',
        },
      ],
      options: [{ vars: 'local', varsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(a, _b) {}
foo();
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '. Allowed unused args must match /^_/u',
            varName: 'a',
          },
          messageId: 'unusedVar',
        },
      ],
      options: [{ args: 'all', argsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(a, _b, c) {
  return a;
}
foo();
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '. Allowed unused args must match /^_/u',
            varName: 'c',
          },
          messageId: 'unusedVar',
        },
      ],
      options: [{ args: 'after-used', argsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(_a) {}
foo();
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '. Allowed unused args must match /[iI]gnored/u',
            varName: '_a',
          },
          messageId: 'unusedVar',
        },
      ],
      options: [{ args: 'all', argsIgnorePattern: '[iI]gnored' }],
    },
    {
      code: 'var [firstItemIgnored, secondItem] = items;',
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '. Allowed unused vars must match /[iI]gnored/u',
            varName: 'secondItem',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ vars: 'all', varsIgnorePattern: '[iI]gnored' }],
    },

    // https://github.com/eslint/eslint/issues/15611
    {
      code: `
const array = ['a', 'b', 'c'];
const [a, _b, c] = array;
const newArray = [a, c];
      `,
      errors: [
        // should report only `newArray`
        { ...assignedError('newArray') },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
const array = ['a', 'b', 'c', 'd', 'e'];
const [a, _b, c] = array;
      `,
      errors: [
        {
          ...assignedError(
            'a',
            '. Allowed unused elements of array destructuring must match /^_/u',
          ),
        },
        {
          ...assignedError(
            'c',
            '. Allowed unused elements of array destructuring must match /^_/u',
          ),
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
const array = ['a', 'b', 'c'];
const [a, _b, c] = array;
const fooArray = ['foo'];
const barArray = ['bar'];
const ignoreArray = ['ignore'];
      `,
      errors: [
        {
          ...assignedError(
            'a',
            '. Allowed unused elements of array destructuring must match /^_/u',
          ),
        },
        {
          ...assignedError(
            'c',
            '. Allowed unused elements of array destructuring must match /^_/u',
          ),
        },
        {
          ...assignedError(
            'fooArray',
            '. Allowed unused vars must match /ignore/u',
          ),
        },
        {
          ...assignedError(
            'barArray',
            '. Allowed unused vars must match /ignore/u',
          ),
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
      options: [
        { destructuredArrayIgnorePattern: '^_', varsIgnorePattern: 'ignore' },
      ],
    },
    {
      code: `
const array = [obj];
const [{ _a, foo }] = array;
console.log(foo);
      `,
      errors: [
        {
          ...assignedError('_a'),
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
function foo([{ _a, bar }]) {
  bar;
}
foo();
      `,
      errors: [
        {
          ...definedError('_a'),
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
let _a, b;

foo.forEach(item => {
  [a, b] = item;
});
      `,
      errors: [
        {
          ...definedError('_a'),
        },
        {
          ...assignedError('b'),
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },

    // Rest property sibling without ignoreRestSiblings
    {
      code: `
const data = { type: 'coords', x: 1, y: 2 };
const { type, ...coords } = data;
console.log(coords);
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'type',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
    },

    // Unused rest property with ignoreRestSiblings
    {
      code: `
const data = { type: 'coords', x: 2, y: 2 };
const { type, ...coords } = data;
console.log(type);
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'coords',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
      options: [{ ignoreRestSiblings: true }],
    },
    {
      code: `
let type, coords;
({ type, ...coords } = data);
console.log(type);
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'coords',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
      options: [{ ignoreRestSiblings: true }],
    },

    // Unused rest property without ignoreRestSiblings
    {
      code: `
const data = { type: 'coords', x: 3, y: 2 };
const { type, ...coords } = data;
console.log(type);
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'coords',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
    },

    // Nested array destructuring with rest property
    {
      code: `
const data = { vars: ['x', 'y'], x: 1, y: 2 };
const {
  vars: [x],
  ...coords
} = data;
console.log(coords);
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'x',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
    },

    // Nested object destructuring with rest property
    {
      code: `
const data = { defaults: { x: 0 }, x: 1, y: 2 };
const {
  defaults: { x },
  ...coords
} = data;
console.log(coords);
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'x',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
    },

    // https://github.com/eslint/eslint/issues/8119
    {
      code: '({ a, ...rest }) => {};',
      errors: [definedError('rest')],
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
      options: [{ args: 'all', ignoreRestSiblings: true }],
    },

    // https://github.com/eslint/eslint/issues/3714
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: `
/* global a$fooz,$foo */
a$fooz;
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '',
            varName: '$foo',
          },
          messageId: 'unusedVar',
        },
      ],
    },
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: `
/* globals a$fooz, $ */
a$fooz;
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '',
            varName: '$',
          },
          messageId: 'unusedVar',
        },
      ],
    },
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: '/*globals $foo*/',
      errors: [
        {
          data: {
            action: 'defined',
            additional: '',
            varName: '$foo',
          },
          messageId: 'unusedVar',
        },
      ],
    },
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: '/* global global*/',
      errors: [
        {
          data: {
            action: 'defined',
            additional: '',
            varName: 'global',
          },
          messageId: 'unusedVar',
        },
      ],
    },
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: '/*global foo:true*/',
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

    // non ascii.
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: `
/*global 変数, 数*/

変数;
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '',
            varName: '数',
          },
          messageId: 'unusedVar',
        },
      ],
    },

    // surrogate pair.
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: `
/*global 𠮷𩸽, 𠮷*/
𠮷𩸽;
      `,
      errors: [
        {
          data: {
            action: 'defined',
            additional: '',
            varName: '𠮷',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    // https://github.com/eslint/eslint/issues/4047
    {
      code: 'export default function (a) {}',
      errors: [definedError('a')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: `
export default function (a, b) {
  console.log(a);
}
      `,
      errors: [definedError('b')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: 'export default (function (a) {});',
      errors: [definedError('a')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: `
export default (function (a, b) {
  console.log(a);
});
      `,
      errors: [definedError('b')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: 'export default a => {};',
      errors: [definedError('a')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: `
export default (a, b) => {
  console.log(a);
};
      `,
      errors: [definedError('b')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
  ],
  valid: [
    // ignore pattern
    {
      code: 'var _a;',
      options: [{ vars: 'all', varsIgnorePattern: '^_' }],
    },
    {
      code: `
var a;
function foo() {
  var _b;
}
foo();
      `,
      options: [{ vars: 'local', varsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(_a) {}
foo();
      `,
      options: [{ args: 'all', argsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(a, _b) {
  return a;
}
foo();
      `,
      options: [{ args: 'after-used', argsIgnorePattern: '^_' }],
    },
    {
      code: `
var [firstItemIgnored, secondItem] = items;
console.log(secondItem);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ vars: 'all', varsIgnorePattern: '[iI]gnored' }],
    },
    {
      code: `
const [a, _b, c] = items;
console.log(a + c);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
const [[a, _b, c]] = items;
console.log(a + c);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
const {
  x: [_a, foo],
} = bar;
console.log(foo);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
function baz([_b, foo]) {
  foo;
}
baz();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
function baz({ x: [_b, foo] }) {
  foo;
}
baz();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
function baz([
  {
    x: [_b, foo],
  },
]) {
  foo;
}
baz();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
let _a, b;
foo.forEach(item => {
  [_a, b] = item;
  doSomething(b);
});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
// doesn't report _x
let _x, y;
_x = 1;
[_x, y] = foo;
y;

// doesn't report _a
let _a, b;
[_a, b] = foo;
_a = 1;
b;
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
// doesn't report _x
let _x, y;
_x = 1;
[_x, y] = foo;
y;

// doesn't report _a
let _a, b;
_a = 1;
({ _a, ...b } = foo);
b;
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
      options: [
        { destructuredArrayIgnorePattern: '^_', ignoreRestSiblings: true },
      ],
    },

    // https://github.com/eslint/eslint/issues/7124
    {
      code: `
(function (a, b, { c, d }) {
  d;
});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ argsIgnorePattern: 'c' }],
    },
    {
      code: `
(function (a, b, { c, d }) {
  c;
});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ argsIgnorePattern: 'd' }],
    },

    // https://github.com/eslint/eslint/issues/7250
    {
      code: `
(function (a, b, c) {
  c;
});
      `,
      options: [{ argsIgnorePattern: 'c' }],
    },
    {
      code: `
(function (a, b, { c, d }) {
  c;
});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ argsIgnorePattern: '[cd]' }],
    },

    // https://github.com/eslint/eslint/issues/7351
    {
      code: `
(class {
  set foo(UNUSED) {}
});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
class Foo {
  set bar(UNUSED) {}
}
console.log(Foo);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    // https://github.com/eslint/eslint/issues/8119
    {
      code: '({ a, ...rest }) => rest;',
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
      options: [{ args: 'all', ignoreRestSiblings: true }],
    },

    // https://github.com/eslint/eslint/issues/14163
    // skip: `something` is undeclared — incompatible with TypeScript project mode
    {
      skip: true,
      code: `
let foo, rest;
({ foo, ...rest } = something);
console.log(rest);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
      options: [{ ignoreRestSiblings: true }],
    },

    // Using object rest for variable omission
    {
      code: `
const data = { type: 'coords', x: 1, y: 2 };
const { type, ...coords } = data;
console.log(coords);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
      options: [{ ignoreRestSiblings: true }],
    },
  ],
});
