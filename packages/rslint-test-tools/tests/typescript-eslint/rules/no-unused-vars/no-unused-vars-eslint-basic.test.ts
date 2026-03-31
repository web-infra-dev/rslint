// The following tests are adapted from the tests in eslint.
// Original Code: https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/tests/lib/rules/no-unused-vars.js
// License      : https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/LICENSE

import { assignedError, definedError, ruleTester } from './eslint-test-helpers';

ruleTester.run('no-unused-vars', {
  invalid: [
    {
      code: `
function foox() {
  return foox();
}
      `,
      errors: [definedError('foox')],
    },
    {
      code: `
(function () {
  function foox() {
    if (true) {
      return foox();
    }
  }
})();
      `,
      errors: [definedError('foox')],
    },
    {
      code: 'var a = 10;',
      errors: [assignedError('a')],
    },
    {
      code: `
function f() {
  var a = 1;
  return function () {
    f((a *= 2));
  };
}
      `,
      errors: [definedError('f')],
    },
    {
      code: `
function f() {
  var a = 1;
  return function () {
    f(++a);
  };
}
      `,
      errors: [definedError('f')],
    },
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: '/*global a */',
      errors: [definedError('a', '')],
    },
    {
      code: `
function foo(first, second) {
  doStuff(function () {
    console.log(second);
  });
}
      `,
      errors: [definedError('foo')],
    },
    {
      code: 'var a = 10;',
      errors: [assignedError('a')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10;
a = 20;
      `,
      errors: [assignedError('a')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10;
(function () {
  var a = 1;
  alert(a);
})();
      `,
      errors: [assignedError('a')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10,
  b = 0,
  c = null;
alert(a + b);
      `,
      errors: [assignedError('c')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10,
  b = 0,
  c = null;
setTimeout(function () {
  var b = 2;
  alert(a + b + c);
}, 0);
      `,
      errors: [assignedError('b')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10,
  b = 0,
  c = null;
setTimeout(function () {
  var b = 2;
  var c = 2;
  alert(a + b + c);
}, 0);
      `,
      errors: [assignedError('b'), assignedError('c')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
function f() {
  var a = [];
  return a.map(function () {});
}
      `,
      errors: [definedError('f')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
function f() {
  var a = [];
  return a.map(function g() {});
}
      `,
      errors: [definedError('f')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
function foo() {
  function foo(x) {
    return x;
  }
  return function () {
    return foo;
  };
}
      `,
      errors: [
        {
          data: { action: 'defined', additional: '', varName: 'foo' },
          messageId: 'unusedVar',
        },
      ],
    },
    {
      code: `
function f() {
  var x;
  function a() {
    x = 42;
  }
  function b() {
    alert(x);
  }
}
      `,
      errors: [definedError('f'), definedError('a'), definedError('b')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
function f(a) {}
f();
      `,
      errors: [definedError('a')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
function a(x, y, z) {
  return y;
}
a();
      `,
      errors: [definedError('z')],
      options: [{ vars: 'all' }],
    },
    {
      code: 'var min = Math.min;',
      errors: [assignedError('min')],
      options: [{ vars: 'all' }],
    },
    {
      code: 'var min = { min: 1 };',
      errors: [assignedError('min')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
Foo.bar = function (baz) {
  return 1;
};
      `,
      errors: [definedError('baz')],
      options: [{ vars: 'all' }],
    },
    {
      code: 'var min = { min: 1 };',
      errors: [assignedError('min')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
function gg(baz, bar) {
  return baz;
}
gg();
      `,
      errors: [definedError('bar')],
      options: [{ vars: 'all' }],
    },
    {
      code: `
(function (foo, baz, bar) {
  return baz;
})();
      `,
      errors: [definedError('bar')],
      options: [{ args: 'after-used', vars: 'all' }],
    },
    {
      code: `
(function (foo, baz, bar) {
  return baz;
})();
      `,
      errors: [definedError('foo'), definedError('bar')],
      options: [{ args: 'all', vars: 'all' }],
    },
    {
      code: `
(function z(foo) {
  var bar = 33;
})();
      `,
      errors: [definedError('foo'), assignedError('bar')],
      options: [{ args: 'all', vars: 'all' }],
    },
    {
      code: `
(function z(foo) {
  z();
})();
      `,
      errors: [definedError('foo')],
      options: [{}],
    },
    {
      code: `
function f() {
  var a = 1;
  return function () {
    f((a = 2));
  };
}
      `,
      errors: [definedError('f'), assignedError('a')],
      options: [{}],
    },
    {
      code: "import x from 'y';",
      errors: [definedError('x')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: `
export function fn2({ x, y }) {
  console.log(x);
}
      `,
      errors: [definedError('y')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: `
export function fn2(x, y) {
  console.log(x);
}
      `,
      errors: [definedError('y')],
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },

    // exported
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
/*exported max*/ var max = 1,
  min = { min: 1 };
      `,
      errors: [assignedError('min')],
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: '/*exported x*/ var { x, y } = z;',
      errors: [assignedError('y')],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
  ],
  valid: [
    `
var foo = 5;

label: while (true) {
  console.log(foo);
  break label;
}
    `,
    `
var foo = 5;

while (true) {
  console.log(foo);
  break;
}
    `,
    {
      code: `
for (let prop in box) {
  box[prop] = parseInt(box[prop]);
}
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    `
var box = { a: 2 };
for (var prop in box) {
  box[prop] = parseInt(box[prop]);
}
    `,
    `
f({
  set foo(a) {
    return;
  },
});
    `,
    {
      code: `
a;
var a;
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10;
alert(a);
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10;
(function () {
  alert(a);
})();
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10;
(function () {
  setTimeout(function () {
    alert(a);
  }, 0);
})();
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
var a = 10;
d[a] = 0;
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
(function () {
  var a = 10;
  return a;
})();
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: '(function g() {})();',
      options: [{ vars: 'all' }],
    },
    {
      code: `
function f(a) {
  alert(a);
}
f();
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
var c = 0;
function f(a) {
  var b = a;
  return b;
}
f(c);
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
function a(x, y) {
  return y;
}
a();
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
var arr1 = [1, 2];
var arr2 = [3, 4];
for (var i in arr1) {
  arr1[i] = 5;
}
for (var i in arr2) {
  arr2[i] = 10;
}
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: 'var a = 10;',
      options: [{ vars: 'local' }],
    },
    {
      code: `
var min = 'min';
Math[min];
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
Foo.bar = function (baz) {
  return baz;
};
      `,
      options: [{ vars: 'all' }],
    },
    'myFunc(function foo() {}.bind(this));',
    'myFunc(function foo() {}.toString());',
    `
function foo(first, second) {
  doStuff(function () {
    console.log(second);
  });
}
foo();
    `,
    `
(function () {
  var doSomething = function doSomething() {};
  doSomething();
})();
    `,
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: '/*global a */ a;',
    },
    {
      code: `
var a = 10;
(function () {
  alert(a);
})();
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
function g(bar, baz) {
  return baz;
}
g();
      `,
      options: [{ vars: 'all' }],
    },
    {
      code: `
function g(bar, baz) {
  return baz;
}
g();
      `,
      options: [{ args: 'after-used', vars: 'all' }],
    },
    {
      code: `
function g(bar, baz) {
  return bar;
}
g();
      `,
      options: [{ args: 'none', vars: 'all' }],
    },
    {
      code: `
function g(bar, baz) {
  return 2;
}
g();
      `,
      options: [{ args: 'none', vars: 'all' }],
    },
    {
      code: `
function g(bar, baz) {
  return bar + baz;
}
g();
      `,
      options: [{ args: 'all', vars: 'local' }],
    },
    {
      code: `
var g = function (bar, baz) {
  return 2;
};
g();
      `,
      options: [{ args: 'none', vars: 'all' }],
    },
    `
(function z() {
  z();
})();
    `,
    {
      code: ' ',
      languageOptions: { globals: { a: true } },
    },
    {
      code: `
var who = 'Paul';
module.exports = \`Hello \${who}!\`;
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: 'export var foo = 123;',
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: 'export function foo() {}',
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: `
let toUpper = partial => partial.toUpperCase;
export { toUpper };
      `,
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: 'export class foo {}',
      languageOptions: {
        parserOptions: { ecmaVersion: 6, sourceType: 'module' },
      },
    },
    {
      code: `
class Foo {}
var x = new Foo();
x.foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const foo = 'hello!';
function bar(foobar = foo) {
  foobar.replace(/!$/, ' world!');
}
bar();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    `
function Foo() {}
var x = new Foo();
x.foo();
    `,
    `
function foo() {
  var foo = 1;
  return foo;
}
foo();
    `,
    `
function foo(foo) {
  return foo;
}
foo(1);
    `,
    `
function foo() {
  function foo() {
    return 1;
  }
  return foo();
}
foo();
    `,
    {
      code: `
function foo() {
  var foo = 1;
  return foo;
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
function foo(foo) {
  return foo;
}
foo(1);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
function foo() {
  function foo() {
    return 1;
  }
  return foo();
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
const [y = x] = [];
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
const { y = x } = {};
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
const {
  z: [y = x],
} = {};
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = [];
const { z: [y] = x } = {};
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
let y;
[y = x] = [];
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
let y;
({
  z: [y = x],
} = {});
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = [];
let y;
({ z: [y] = x } = {});
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
function foo(y = x) {
  bar(y);
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
function foo({ y = x } = {}) {
  bar(y);
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
function foo(
  y = function (z = x) {
    bar(z);
  },
) {
  y();
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
const x = 1;
function foo(
  y = function () {
    bar(x);
  },
) {
  y();
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1;
var [y = x] = [];
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1;
var { y = x } = {};
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1;
var {
  z: [y = x],
} = {};
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = [];
var { z: [y] = x } = {};
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1,
  y;
[y = x] = [];
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1,
  y;
({
  z: [y = x],
} = {});
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = [],
  y;
({ z: [y] = x } = {});
foo(y);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1;
function foo(y = x) {
  bar(y);
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1;
function foo({ y = x } = {}) {
  bar(y);
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1;
function foo(
  y = function (z = x) {
    bar(z);
  },
) {
  y();
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
var x = 1;
function foo(
  y = function () {
    bar(x);
  },
) {
  y();
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    // exported variables should work
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: "/*exported toaster*/ var toaster = 'great';",
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
/*exported toaster, poster*/ var toaster = 1;
poster = 0;
      `,
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: '/*exported x*/ var { x } = y;',
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: '/*exported x, y*/ var { x, y } = z;',
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    // Can mark variables as used via context.markVariableAsUsed()
    {
      // skip: uses ESLint markVariableAsUsed API not available in rslint
      skip: true,
      code: '/*eslint @rule-tester/use-every-a:1*/ var a;',
    },
    {
      // skip: uses ESLint markVariableAsUsed API not available in rslint
      skip: true,
      code: `
/*eslint @rule-tester/use-every-a:1*/ !function (a) {
  return 1;
};
      `,
    },
    {
      // skip: uses ESLint markVariableAsUsed API not available in rslint
      skip: true,
      code: `
/*eslint @rule-tester/use-every-a:1*/ !function () {
  var a;
  return 1;
};
      `,
    },
  ],
});
