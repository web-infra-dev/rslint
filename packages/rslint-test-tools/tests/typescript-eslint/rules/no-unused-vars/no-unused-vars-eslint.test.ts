// Every case of typescript-eslint v8.72.1's adaptation of the ESLint
// no-unused-vars suite, preserving inputs, options and comments. Errors carry
// the full messages and ranges produced by that release.
// Parser-level `ecmaVersion`/`sourceType` move to language options. Upstream
// runs these cases without type information; here each uses a fixture project
// with the upstream compiler options (without ambient `types`).
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.1/packages/eslint-plugin/tests/rules/no-unused-vars/no-unused-vars-eslint.test.ts
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'file.ts',
  fixtureFiles: {
    'tsconfig.json': "{\"compilerOptions\":{\"jsx\":\"preserve\",\"target\":\"es2015\",\"module\":\"commonjs\",\"strict\":true,\"types\":[],\"lib\":[\"es2015\",\"es2017\",\"esnext\"],\"experimentalDecorators\":true,\"esModuleInterop\":true},\"include\":[\"file.ts\",\"react.tsx\",\"foo.d.ts\"]}",
  },
  languageOptions: { sourceType: 'script', parserOptions: { project: './tsconfig.json' } },
});

ruleTester.run('no-unused-vars', {
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
      languageOptions: { ecmaVersion: 6 },
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
      options: ['all'],
    },
    {
      code: `
var a = 10;
alert(a);
      `,
      options: ['all'],
    },
    {
      code: `
var a = 10;
(function () {
  alert(a);
})();
      `,
      options: ['all'],
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
      options: ['all'],
    },
    {
      code: `
var a = 10;
d[a] = 0;
      `,
      options: ['all'],
    },
    {
      code: `
(function () {
  var a = 10;
  return a;
})();
      `,
      options: ['all'],
    },
    {
      code: '(function g() {})();',
      options: ['all'],
    },
    {
      code: `
function f(a) {
  alert(a);
}
f();
      `,
      options: ['all'],
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
      options: ['all'],
    },
    {
      code: `
function a(x, y) {
  return y;
}
a();
      `,
      options: ['all'],
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
      options: ['all'],
    },
    {
      code: 'var a = 10;',
      options: ['local'],
    },
    {
      code: `
var min = 'min';
Math[min];
      `,
      options: ['all'],
    },
    {
      code: `
Foo.bar = function (baz) {
  return baz;
};
      `,
      options: ['all'],
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
    '/*global a */ a;',
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
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'export var foo = 123;',
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: 'export function foo() {}',
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: `
let toUpper = partial => partial.toUpperCase;
export { toUpper };
      `,
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: 'export class foo {}',
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: `
class Foo {}
var x = new Foo();
x.foo();
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const foo = 'hello!';
function bar(foobar = foo) {
  foobar.replace(/!$/, ' world!');
}
bar();
      `,
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
function foo(foo) {
  return foo;
}
foo(1);
      `,
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const x = 1;
const [y = x] = [];
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const x = 1;
const { y = x } = {};
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const x = 1;
const {
  z: [y = x],
} = {};
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const x = [];
const { z: [y] = x } = {};
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const x = 1;
let y;
[y = x] = [];
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const x = [];
let y;
({ z: [y] = x } = {});
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const x = 1;
function foo(y = x) {
  bar(y);
}
foo();
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
const x = 1;
function foo({ y = x } = {}) {
  bar(y);
}
foo();
      `,
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
var x = 1;
var [y = x] = [];
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
var x = 1;
var { y = x } = {};
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
var x = 1;
var {
  z: [y = x],
} = {};
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
var x = [];
var { z: [y] = x } = {};
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
var x = 1,
  y;
[y = x] = [];
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
var x = [],
  y;
({ z: [y] = x } = {});
foo(y);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
var x = 1;
function foo(y = x) {
  bar(y);
}
foo();
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
var x = 1;
function foo({ y = x } = {}) {
  bar(y);
}
foo();
      `,
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
    },

    // exported variables should work
    "/*exported toaster*/ var toaster = 'great';",
    `
/*exported toaster, poster*/ var toaster = 1;
poster = 0;
    `,
    {
      code: '/*exported x*/ var { x } = y;',
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: '/*exported x, y*/ var { x, y } = z;',
      languageOptions: { ecmaVersion: 6 },
    },

    // Can mark variables as used via context.markVariableAsUsed()
    { code: '/*eslint @rule-tester/use-every-a:1*/ var a;', skip: "Requires ESLint's cross-rule markVariableAsUsed API: @rslint/rule-tester cannot define the upstream helper rule, and native rules do not share used-variable marks." },
    { code: `
/*eslint @rule-tester/use-every-a:1*/ !function (a) {
  return 1;
};
    `, skip: "Requires ESLint's cross-rule markVariableAsUsed API: @rslint/rule-tester cannot define the upstream helper rule, and native rules do not share used-variable marks." },
    { code: `
/*eslint @rule-tester/use-every-a:1*/ !function () {
  var a;
  return 1;
};
    `, skip: "Requires ESLint's cross-rule markVariableAsUsed API: @rslint/rule-tester cannot define the upstream helper rule, and native rules do not share used-variable marks." },

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
      languageOptions: { ecmaVersion: 6 },
      options: [{ vars: 'all', varsIgnorePattern: '[iI]gnored' }],
    },
    {
      code: `
const [a, _b, c] = items;
console.log(a + c);
      `,
      languageOptions: { ecmaVersion: 6 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
const [[a, _b, c]] = items;
console.log(a + c);
      `,
      languageOptions: { ecmaVersion: 6 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
const {
  x: [_a, foo],
} = bar;
console.log(foo);
      `,
      languageOptions: { ecmaVersion: 6 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
function baz([_b, foo]) {
  foo;
}
baz();
      `,
      languageOptions: { ecmaVersion: 6 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
function baz({ x: [_b, foo] }) {
  foo;
}
baz();
      `,
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
let _a, b;
foo.forEach(item => {
  [_a, b] = item;
  doSomething(b);
});
      `,
      languageOptions: { ecmaVersion: 6 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
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
      languageOptions: { ecmaVersion: 2018 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
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
      languageOptions: { ecmaVersion: 2018 },
      options: [
        { destructuredArrayIgnorePattern: '^_', ignoreRestSiblings: true },
      ],
    },

    // for-in loops (see #2342)
    `
(function (obj) {
  var name;
  for (name in obj) return;
})({});
    `,
    `
(function (obj) {
  var name;
  for (name in obj) {
    return;
  }
})({});
    `,
    `
(function (obj) {
  for (var name in obj) {
    return true;
  }
})({});
    `,
    `
(function (obj) {
  for (var name in obj) return true;
})({});
    `,

    {
      code: `
(function (obj) {
  let name;
  for (name in obj) return;
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (obj) {
  let name;
  for (name in obj) {
    return;
  }
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (obj) {
  for (let name in obj) {
    return true;
  }
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (obj) {
  for (let name in obj) return true;
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },

    {
      code: `
(function (obj) {
  for (const name in obj) {
    return true;
  }
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (obj) {
  for (const name in obj) return true;
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },

    // For-of loops
    {
      code: `
(function (iter) {
  let name;
  for (name of iter) return;
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (iter) {
  let name;
  for (name of iter) {
    return;
  }
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (iter) {
  for (let name of iter) {
    return true;
  }
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (iter) {
  for (let name of iter) return true;
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },

    {
      code: `
(function (iter) {
  for (const name of iter) {
    return true;
  }
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (iter) {
  for (const name of iter) return true;
})({});
      `,
      languageOptions: { ecmaVersion: 6 },
    },

    // Sequence Expressions (See https://github.com/eslint/eslint/issues/14325)
    {
      code: `
let x = 0;
foo = (0, x++);
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
let x = 0;
foo = (0, (x += 1));
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
let x = 0;
foo = (0, (x = x + 1));
      `,
      languageOptions: { ecmaVersion: 6 },
    },

    // caughtErrors
    {
      code: `
try {
} catch (err) {}
      `,
      options: [{ caughtErrors: 'none' }],
    },
    {
      code: `
try {
} catch (err) {
  console.error(err);
}
      `,
      options: [{ caughtErrors: 'all' }],
    },
    {
      code: `
try {
} catch (ignoreErr) {}
      `,
      options: [{ caughtErrorsIgnorePattern: '^ignore' }],
    },
    {
      code: `
try {
} catch (ignoreErr) {}
      `,
      options: [{ caughtErrors: 'all', caughtErrorsIgnorePattern: '^ignore' }],
    },

    // caughtErrors with other combinations
    {
      code: `
try {
} catch (err) {}
      `,
      options: [{ args: 'all', caughtErrors: 'none', vars: 'all' }],
    },

    // Using object rest for variable omission
    {
      code: `
const data = { type: 'coords', x: 1, y: 2 };
const { type, ...coords } = data;
console.log(coords);
      `,
      languageOptions: { ecmaVersion: 2018 },
      options: [{ ignoreRestSiblings: true }],
    },

    // https://github.com/eslint/eslint/issues/6348
    `
var a = 0,
  b;
b = a = a + 1;
foo(b);
    `,
    `
var a = 0,
  b;
b = a += a + 1;
foo(b);
    `,
    `
var a = 0,
  b;
b = a++;
foo(b);
    `,
    `
function foo(a) {
  var b = (a = a + 1);
  bar(b);
}
foo();
    `,
    `
function foo(a) {
  var b = (a += a + 1);
  bar(b);
}
foo();
    `,
    `
function foo(a) {
  var b = a++;
  bar(b);
}
foo();
    `,

    // https://github.com/eslint/eslint/issues/6576
    `
var unregisterFooWatcher;
// ...
unregisterFooWatcher = $scope.$watch('foo', function () {
  // ...some code..
  unregisterFooWatcher();
});
    `,
    `
var ref;
ref = setInterval(function () {
  clearInterval(ref);
}, 10);
    `,
    `
var _timer;
function f() {
  _timer = setTimeout(function () {}, _timer ? 100 : 0);
}
f();
    `,
    `
function foo(cb) {
  cb = (function () {
    function something(a) {
      cb(1 + a);
    }
    register(something);
  })();
}
foo();
    `,
    {
      code: `
function* foo(cb) {
  cb = yield function (a) {
    cb(1 + a);
  };
}
foo();
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
function foo(cb) {
  cb = tag\`hello\${function (a) {
    cb(1 + a);
  }}\`;
}
foo();
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    `
function foo(cb) {
  var b;
  cb = b = function (a) {
    cb(1 + a);
  };
  b();
}
foo();
    `,

    // https://github.com/eslint/eslint/issues/6646
    `
function someFunction() {
  var a = 0,
    i;
  for (i = 0; i < 2; i++) {
    a = myFunction(a);
  }
}
someFunction();
    `,

    // https://github.com/eslint/eslint/issues/7124
    {
      code: `
(function (a, b, { c, d }) {
  d;
});
      `,
      languageOptions: { ecmaVersion: 6 },
      options: [{ argsIgnorePattern: 'c' }],
    },
    {
      code: `
(function (a, b, { c, d }) {
  c;
});
      `,
      languageOptions: { ecmaVersion: 6 },
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
      languageOptions: { ecmaVersion: 6 },
      options: [{ argsIgnorePattern: '[cd]' }],
    },

    // https://github.com/eslint/eslint/issues/7351
    {
      code: `
(class {
  set foo(UNUSED) {}
});
      `,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
class Foo {
  set bar(UNUSED) {}
}
console.log(Foo);
      `,
      languageOptions: { ecmaVersion: 6 },
    },

    // https://github.com/eslint/eslint/issues/8119
    {
      code: '({ a, ...rest }) => rest;',
      languageOptions: { ecmaVersion: 2018 },
      options: [{ args: 'all', ignoreRestSiblings: true }],
    },

    // https://github.com/eslint/eslint/issues/14163
    {
      code: `
let foo, rest;
({ foo, ...rest } = something);
console.log(rest);
      `,
      languageOptions: { ecmaVersion: 2020 },
      options: [{ ignoreRestSiblings: true }],
    },

    // https://github.com/eslint/eslint/issues/10952
    { code: `
/*eslint @rule-tester/use-every-a:1*/ !function (b, a) {
  return 1;
};
    `, skip: "Requires ESLint's cross-rule markVariableAsUsed API: @rslint/rule-tester cannot define the upstream helper rule, and native rules do not share used-variable marks." },

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
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
const a = () => () => {
  a();
};
a();
      `,
      languageOptions: { ecmaVersion: 2015 },
    },

    // export * as ns from "source"
    {
      code: "export * as ns from 'source';",
      languageOptions: { ecmaVersion: 2020, sourceType: 'module' },
    },

    // import.meta
    {
      code: 'import.meta;',
      languageOptions: { ecmaVersion: 2020, sourceType: 'module' },
    },

    // https://github.com/eslint/eslint/issues/17299
    {
      code: `
var a;
a ||= 1;
      `,
      languageOptions: { ecmaVersion: 2021 },
    },
    {
      code: `
var a;
a &&= 1;
      `,
      languageOptions: { ecmaVersion: 2021 },
    },
    {
      code: `
var a;
a ??= 1;
      `,
      languageOptions: { ecmaVersion: 2021 },
    },

    // ignore class with static initialization block https://github.com/eslint/eslint/issues/17772
    {
      code: `
class Foo {
  static {}
}
      `,
      languageOptions: { ecmaVersion: 2022 },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    {
      code: `
class Foo {
  static {}
}
      `,
      languageOptions: { ecmaVersion: 2022 },
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
      languageOptions: { ecmaVersion: 2022 },
      options: [
        { ignoreClassWithStaticInitBlock: false, varsIgnorePattern: '^Foo' },
      ],
    },

    // https://github.com/eslint/eslint/issues/17568
    {
      code: `
const a = 5;
const _c = a + 5;
      `,
      languageOptions: { ecmaVersion: 6 },
      options: [
        { args: 'all', reportUsedIgnorePattern: true, varsIgnorePattern: '^_' },
      ],
    },
    {
      code: `
(function foo(a, _b) {
  return a + 5;
})(5);
      `,
      options: [
        { args: 'all', argsIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
    },
    {
      code: `
const [a, _b, c] = items;
console.log(a + c);
      `,
      languageOptions: { ecmaVersion: 6 },
      options: [
        { destructuredArrayIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
    },
  ],
  invalid: [
    {
      code: `
function foox() {
  return foox();
}
      `,
      errors: [{ messageId: "unusedVar", message: "'foox' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 14 }],
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
      errors: [{ messageId: "unusedVar", message: "'foox' is defined but never used.", line: 3, column: 12, endLine: 3, endColumn: 16 }],
    },
    {
      code: 'var a = 10;',
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 1, column: 5, endLine: 1, endColumn: 6 }],
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
      errors: [{ messageId: "unusedVar", message: "'f' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 11 }],
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
      errors: [{ messageId: "unusedVar", message: "'f' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 11 }],
    },
    {
      code: '/*global a */',
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 1, column: 10, endLine: 1, endColumn: 11 }],
    },
    {
      code: `
function foo(first, second) {
  doStuff(function () {
    console.log(second);
  });
}
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 13 }],
    },
    {
      code: 'var a = 10;',
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 1, column: 5, endLine: 1, endColumn: 6 }],
      options: ['all'],
    },
    {
      code: `
var a = 10;
a = 20;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }],
      options: ['all'],
    },
    {
      code: `
var a = 10;
(function () {
  var a = 1;
  alert(a);
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 2, column: 5, endLine: 2, endColumn: 6 }],
      options: ['all'],
    },
    {
      code: `
var a = 10,
  b = 0,
  c = null;
alert(a + b);
      `,
      errors: [{ messageId: "unusedVar", message: "'c' is assigned a value but never used.", line: 4, column: 3, endLine: 4, endColumn: 4 }],
      options: ['all'],
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
      errors: [{ messageId: "unusedVar", message: "'b' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 4 }],
      options: ['all'],
    },
    // Multiple errors
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
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
      errors: [{ messageId: "unusedVar", message: "'b' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 4 }, { messageId: "unusedVar", message: "'c' is assigned a value but never used.", line: 4, column: 3, endLine: 4, endColumn: 4 }],
      options: ['all'],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
function f() {
  var a = [];
  return a.map(function () {});
}
      `,
      errors: [{ messageId: "unusedVar", message: "'f' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 11 }],
      options: ['all'],
    },
    {
      code: `
function f() {
  var a = [];
  return a.map(function g() {});
}
      `,
      errors: [{ messageId: "unusedVar", message: "'f' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 11 }],
      options: ['all'],
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
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 13 }],
    },
    // Multiple errors
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
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
      errors: [{ messageId: "unusedVar", message: "'f' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 11 }, { messageId: "unusedVar", message: "'a' is defined but never used.", line: 4, column: 12, endLine: 4, endColumn: 13 }, { messageId: "unusedVar", message: "'b' is defined but never used.", line: 7, column: 12, endLine: 7, endColumn: 13 }],
      options: ['all'],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
function f(a) {}
f();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 2, column: 12, endLine: 2, endColumn: 13 }],
      options: ['all'],
    },
    {
      code: `
function a(x, y, z) {
  return y;
}
a();
      `,
      errors: [{ messageId: "unusedVar", message: "'z' is defined but never used.", line: 2, column: 18, endLine: 2, endColumn: 19 }],
      options: ['all'],
    },
    {
      code: 'var min = Math.min;',
      errors: [{ messageId: "unusedVar", message: "'min' is assigned a value but never used.", line: 1, column: 5, endLine: 1, endColumn: 8 }],
      options: ['all'],
    },
    {
      code: 'var min = { min: 1 };',
      errors: [{ messageId: "unusedVar", message: "'min' is assigned a value but never used.", line: 1, column: 5, endLine: 1, endColumn: 8 }],
      options: ['all'],
    },
    {
      code: `
Foo.bar = function (baz) {
  return 1;
};
      `,
      errors: [{ messageId: "unusedVar", message: "'baz' is defined but never used.", line: 2, column: 21, endLine: 2, endColumn: 24 }],
      options: ['all'],
    },
    {
      code: 'var min = { min: 1 };',
      errors: [{ messageId: "unusedVar", message: "'min' is assigned a value but never used.", line: 1, column: 5, endLine: 1, endColumn: 8 }],
      options: [{ vars: 'all' }],
    },
    {
      code: `
function gg(baz, bar) {
  return baz;
}
gg();
      `,
      errors: [{ messageId: "unusedVar", message: "'bar' is defined but never used.", line: 2, column: 18, endLine: 2, endColumn: 21 }],
      options: [{ vars: 'all' }],
    },
    {
      code: `
(function (foo, baz, bar) {
  return baz;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'bar' is defined but never used.", line: 2, column: 22, endLine: 2, endColumn: 25 }],
      options: [{ args: 'after-used', vars: 'all' }],
    },
    {
      code: `
(function (foo, baz, bar) {
  return baz;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 2, column: 12, endLine: 2, endColumn: 15 }, { messageId: "unusedVar", message: "'bar' is defined but never used.", line: 2, column: 22, endLine: 2, endColumn: 25 }],
      options: [{ args: 'all', vars: 'all' }],
    },
    // Multiple errors
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
(function z(foo) {
  var bar = 33;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 2, column: 13, endLine: 2, endColumn: 16 }, { messageId: "unusedVar", message: "'bar' is assigned a value but never used.", line: 3, column: 7, endLine: 3, endColumn: 10 }],
      options: [{ args: 'all', vars: 'all' }],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
(function z(foo) {
  z();
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 2, column: 13, endLine: 2, endColumn: 16 }],
      options: [{}],
    },
    // Multiple errors
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
function f() {
  var a = 1;
  return function () {
    f((a = 2));
  };
}
      `,
      errors: [{ messageId: "unusedVar", message: "'f' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 11 }, { messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 7, endLine: 3, endColumn: 8 }],
      options: [{}],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: "import x from 'y';",
      errors: [{ messageId: "unusedVar", message: "'x' is defined but never used.", line: 1, column: 8, endLine: 1, endColumn: 9, suggestions: [{ messageId: "removeUnusedImportDeclaration", desc: "Remove unused import declaration.", output: '' }] }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: `
export function fn2({ x, y }) {
  console.log(x);
}
      `,
      errors: [{ messageId: "unusedVar", message: "'y' is defined but never used.", line: 2, column: 26, endLine: 2, endColumn: 27 }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: `
export function fn2(x, y) {
  console.log(x);
}
      `,
      errors: [{ messageId: "unusedVar", message: "'y' is defined but never used.", line: 2, column: 24, endLine: 2, endColumn: 25 }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },

    // exported
    {
      code: `
/*exported max*/ var max = 1,
  min = { min: 1 };
      `,
      errors: [{ messageId: "unusedVar", message: "'min' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 6 }],
    },
    {
      code: '/*exported x*/ var { x, y } = z;',
      errors: [{ messageId: "unusedVar", message: "'y' is assigned a value but never used.", line: 1, column: 25, endLine: 1, endColumn: 26 }],
      languageOptions: { ecmaVersion: 6 },
    },

    // ignore pattern
    {
      code: `
var _a;
var b;
      `,
      errors: [{ messageId: "unusedVar", message: "'b' is defined but never used. Allowed unused vars must match /^_/u.", line: 3, column: 5, endLine: 3, endColumn: 6 }],
      options: [{ vars: 'all', varsIgnorePattern: '^_' }],
    },
    {
      code: `
var a;
function foo() {
  var _b;
  var c_;
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'c_' is defined but never used. Allowed unused vars must match /^_/u.", line: 5, column: 7, endLine: 5, endColumn: 9 }],
      options: [{ vars: 'local', varsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(a, _b) {}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used. Allowed unused args must match /^_/u.", line: 2, column: 14, endLine: 2, endColumn: 15 }],
      options: [{ args: 'all', argsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(a, _b, c) {
  return a;
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'c' is defined but never used. Allowed unused args must match /^_/u.", line: 2, column: 21, endLine: 2, endColumn: 22 }],
      options: [{ args: 'after-used', argsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(_a) {}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'_a' is defined but never used. Allowed unused args must match /[iI]gnored/u.", line: 2, column: 14, endLine: 2, endColumn: 16 }],
      options: [{ args: 'all', argsIgnorePattern: '[iI]gnored' }],
    },
    {
      code: 'var [firstItemIgnored, secondItem] = items;',
      errors: [{ messageId: "unusedVar", message: "'secondItem' is assigned a value but never used. Allowed unused vars must match /[iI]gnored/u.", line: 1, column: 24, endLine: 1, endColumn: 34 }],
      languageOptions: { ecmaVersion: 6 },
      options: [{ vars: 'all', varsIgnorePattern: '[iI]gnored' }],
    },

    // https://github.com/eslint/eslint/issues/15611
    {
      code: `
const array = ['a', 'b', 'c'];
const [a, _b, c] = array;
const newArray = [a, c];
      `,
      errors: [{ messageId: "unusedVar", message: "'newArray' is assigned a value but never used.", line: 4, column: 7, endLine: 4, endColumn: 15 }],
      languageOptions: { ecmaVersion: 2020 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
const array = ['a', 'b', 'c', 'd', 'e'];
const [a, _b, c] = array;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used. Allowed unused elements of array destructuring must match /^_/u.", line: 3, column: 8, endLine: 3, endColumn: 9 }, { messageId: "unusedVar", message: "'c' is assigned a value but never used. Allowed unused elements of array destructuring must match /^_/u.", line: 3, column: 15, endLine: 3, endColumn: 16 }],
      languageOptions: { ecmaVersion: 2020 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    // Multiple errors
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
const array = ['a', 'b', 'c'];
const [a, _b, c] = array;
const fooArray = ['foo'];
const barArray = ['bar'];
const ignoreArray = ['ignore'];
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used. Allowed unused elements of array destructuring must match /^_/u.", line: 3, column: 8, endLine: 3, endColumn: 9 }, { messageId: "unusedVar", message: "'c' is assigned a value but never used. Allowed unused elements of array destructuring must match /^_/u.", line: 3, column: 15, endLine: 3, endColumn: 16 }, { messageId: "unusedVar", message: "'fooArray' is assigned a value but never used. Allowed unused vars must match /ignore/u.", line: 4, column: 7, endLine: 4, endColumn: 15 }, { messageId: "unusedVar", message: "'barArray' is assigned a value but never used. Allowed unused vars must match /ignore/u.", line: 5, column: 7, endLine: 5, endColumn: 15 }],
      languageOptions: { ecmaVersion: 2020 },
      options: [
        { destructuredArrayIgnorePattern: '^_', varsIgnorePattern: 'ignore' },
      ],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
const array = [obj];
const [{ _a, foo }] = array;
console.log(foo);
      `,
      errors: [{ messageId: "unusedVar", message: "'_a' is assigned a value but never used.", line: 3, column: 10, endLine: 3, endColumn: 12 }],
      languageOptions: { ecmaVersion: 2020 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
function foo([{ _a, bar }]) {
  bar;
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'_a' is defined but never used.", line: 2, column: 17, endLine: 2, endColumn: 19 }],
      languageOptions: { ecmaVersion: 2020 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `
let _a, b;

foo.forEach(item => {
  [a, b] = item;
});
      `,
      errors: [{ messageId: "unusedVar", message: "'_a' is defined but never used.", line: 2, column: 5, endLine: 2, endColumn: 7 }, { messageId: "unusedVar", message: "'b' is assigned a value but never used.", line: 2, column: 9, endLine: 2, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2020 },
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },

    // for-in loops (see #2342)
    {
      code: `
(function (obj) {
  var name;
  for (name in obj) {
    i();
    return;
  }
})({});
      `,
      errors: [{ messageId: "unusedVar", message: "'name' is assigned a value but never used.", line: 4, column: 8, endLine: 4, endColumn: 12 }],
    },
    {
      code: `
(function (obj) {
  var name;
  for (name in obj) {
  }
})({});
      `,
      errors: [{ messageId: "unusedVar", message: "'name' is assigned a value but never used.", line: 4, column: 8, endLine: 4, endColumn: 12 }],
    },
    {
      code: `
(function (obj) {
  for (var name in obj) {
  }
})({});
      `,
      errors: [{ messageId: "unusedVar", message: "'name' is assigned a value but never used.", line: 3, column: 12, endLine: 3, endColumn: 16 }],
    },

    // For-of loops
    {
      code: `
(function (iter) {
  var name;
  for (name of iter) {
    i();
    return;
  }
})({});
      `,
      errors: [{ messageId: "unusedVar", message: "'name' is assigned a value but never used.", line: 4, column: 8, endLine: 4, endColumn: 12 }],
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (iter) {
  var name;
  for (name of iter) {
  }
})({});
      `,
      errors: [{ messageId: "unusedVar", message: "'name' is assigned a value but never used.", line: 4, column: 8, endLine: 4, endColumn: 12 }],
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: `
(function (iter) {
  for (var name of iter) {
  }
})({});
      `,
      errors: [{ messageId: "unusedVar", message: "'name' is assigned a value but never used.", line: 3, column: 12, endLine: 3, endColumn: 16 }],
      languageOptions: { ecmaVersion: 6 },
    },

    // https://github.com/eslint/eslint/issues/3617
    {
      code: `
/* global foobar, foo, bar */
foobar;
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 2, column: 19, endLine: 2, endColumn: 22 }, { messageId: "unusedVar", message: "'bar' is defined but never used.", line: 2, column: 24, endLine: 2, endColumn: 27 }],
    },
    // Multiple errors
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
/* global foobar,
   foo,
   bar
 */
foobar;
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 3, column: 4, endLine: 3, endColumn: 7 }, { messageId: "unusedVar", message: "'bar' is defined but never used.", line: 4, column: 4, endLine: 4, endColumn: 7 }],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */

    // Rest property sibling without ignoreRestSiblings
    {
      code: `
const data = { type: 'coords', x: 1, y: 2 };
const { type, ...coords } = data;
console.log(coords);
      `,
      errors: [{ messageId: "unusedVar", message: "'type' is assigned a value but never used.", line: 3, column: 9, endLine: 3, endColumn: 13 }],
      languageOptions: { ecmaVersion: 2018 },
    },

    // Unused rest property with ignoreRestSiblings
    {
      code: `
const data = { type: 'coords', x: 2, y: 2 };
const { type, ...coords } = data;
console.log(type);
      `,
      errors: [{ messageId: "unusedVar", message: "'coords' is assigned a value but never used.", line: 3, column: 18, endLine: 3, endColumn: 24 }],
      languageOptions: { ecmaVersion: 2018 },
      options: [{ ignoreRestSiblings: true }],
    },
    {
      code: `
let type, coords;
({ type, ...coords } = data);
console.log(type);
      `,
      errors: [{ messageId: "unusedVar", message: "'coords' is assigned a value but never used.", line: 3, column: 13, endLine: 3, endColumn: 19 }],
      languageOptions: { ecmaVersion: 2018 },
      options: [{ ignoreRestSiblings: true }],
    },

    // Unused rest property without ignoreRestSiblings
    {
      code: `
const data = { type: 'coords', x: 3, y: 2 };
const { type, ...coords } = data;
console.log(type);
      `,
      errors: [{ messageId: "unusedVar", message: "'coords' is assigned a value but never used.", line: 3, column: 18, endLine: 3, endColumn: 24 }],
      languageOptions: { ecmaVersion: 2018 },
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
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 4, column: 10, endLine: 4, endColumn: 11 }],
      languageOptions: { ecmaVersion: 2018 },
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
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 4, column: 15, endLine: 4, endColumn: 16 }],
      languageOptions: { ecmaVersion: 2018 },
    },

    // https://github.com/eslint/eslint/issues/8119
    {
      code: '({ a, ...rest }) => {};',
      errors: [{ messageId: "unusedVar", message: "'rest' is defined but never used.", line: 1, column: 10, endLine: 1, endColumn: 14 }],
      languageOptions: { ecmaVersion: 2018 },
      options: [{ args: 'all', ignoreRestSiblings: true }],
    },

    // https://github.com/eslint/eslint/issues/3714
    {
      code: `
/* global a$fooz,$foo */
a$fooz;
      `,
      errors: [{ messageId: "unusedVar", message: "'$foo' is defined but never used.", line: 2, column: 18, endLine: 2, endColumn: 22 }],
    },
    {
      code: `
/* globals a$fooz, $ */
a$fooz;
      `,
      errors: [{ messageId: "unusedVar", message: "'$' is defined but never used.", line: 2, column: 20, endLine: 2, endColumn: 21 }],
    },
    {
      code: '/*globals $foo*/',
      errors: [{ messageId: "unusedVar", message: "'$foo' is defined but never used.", line: 1, column: 11, endLine: 1, endColumn: 15 }],
    },
    {
      code: '/* global global*/',
      errors: [{ messageId: "unusedVar", message: "'global' is defined but never used.", line: 1, column: 11, endLine: 1, endColumn: 17 }],
    },
    {
      code: '/*global foo:true*/',
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 1, column: 10, endLine: 1, endColumn: 13 }],
    },

    // non ascii.
    {
      code: `
/*global 変数, 数*/

変数;
      `,
      errors: [{ messageId: "unusedVar", message: "'数' is defined but never used.", line: 2, column: 14, endLine: 2, endColumn: 15 }],
    },

    // surrogate pair.
    {
      code: `
/*global 𠮷𩸽, 𠮷*/
𠮷𩸽;
      `,
      errors: [{ messageId: "unusedVar", message: "'𠮷' is defined but never used.", line: 2, column: 16, endLine: 2, endColumn: 18 }],
      languageOptions: { ecmaVersion: 6 },
    },

    // https://github.com/eslint/eslint/issues/4047
    {
      code: 'export default function (a) {}',
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 1, column: 26, endLine: 1, endColumn: 27 }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: `
export default function (a, b) {
  console.log(a);
}
      `,
      errors: [{ messageId: "unusedVar", message: "'b' is defined but never used.", line: 2, column: 29, endLine: 2, endColumn: 30 }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: 'export default (function (a) {});',
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 1, column: 27, endLine: 1, endColumn: 28 }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: `
export default (function (a, b) {
  console.log(a);
});
      `,
      errors: [{ messageId: "unusedVar", message: "'b' is defined but never used.", line: 2, column: 30, endLine: 2, endColumn: 31 }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: 'export default a => {};',
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 1, column: 16, endLine: 1, endColumn: 17 }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: `
export default (a, b) => {
  console.log(a);
};
      `,
      errors: [{ messageId: "unusedVar", message: "'b' is defined but never used.", line: 2, column: 20, endLine: 2, endColumn: 21 }],
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },

    // caughtErrors
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'err' is defined but never used.", line: 3, column: 10, endLine: 3, endColumn: 13 }],
    },
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'err' is defined but never used.", line: 3, column: 10, endLine: 3, endColumn: 13 }],
      options: [{ caughtErrors: 'all' }],
    },
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'err' is defined but never used. Allowed unused caught errors must match /^ignore/u.", line: 3, column: 10, endLine: 3, endColumn: 13 }],
      options: [{ caughtErrors: 'all', caughtErrorsIgnorePattern: '^ignore' }],
    },
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'err' is defined but never used.", line: 3, column: 10, endLine: 3, endColumn: 13 }],
      options: [{ caughtErrors: 'all', varsIgnorePattern: '^err' }],
    },
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'err' is defined but never used.", line: 3, column: 10, endLine: 3, endColumn: 13 }],
      options: [{ caughtErrors: 'all', varsIgnorePattern: '^.' }],
    },

    // multiple try catch with one success
    {
      code: `
try {
} catch (ignoreErr) {}
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'err' is defined but never used. Allowed unused caught errors must match /^ignore/u.", line: 5, column: 10, endLine: 5, endColumn: 13 }],
      options: [{ caughtErrors: 'all', caughtErrorsIgnorePattern: '^ignore' }],
    },

    // multiple try catch both fail
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
try {
} catch (error) {}
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'error' is defined but never used. Allowed unused caught errors must match /^ignore/u.", line: 3, column: 10, endLine: 3, endColumn: 15 }, { messageId: "unusedVar", message: "'err' is defined but never used. Allowed unused caught errors must match /^ignore/u.", line: 5, column: 10, endLine: 5, endColumn: 13 }],
      options: [{ caughtErrors: 'all', caughtErrorsIgnorePattern: '^ignore' }],
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */

    // caughtErrors with other configs
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'err' is defined but never used.", line: 3, column: 10, endLine: 3, endColumn: 13 }],
      options: [{ args: 'all', caughtErrors: 'all', vars: 'all' }],
    },

    // no conflict in ignore patterns
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'err' is defined but never used.", line: 3, column: 10, endLine: 3, endColumn: 13 }],
      options: [
        {
          args: 'all',
          argsIgnorePattern: '^er',
          caughtErrors: 'all',
          vars: 'all',
        },
      ],
    },

    // Ignore reads for modifications to itself: https://github.com/eslint/eslint/issues/6348
    {
      code: `
var a = 0;
a = a + 1;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }],
    },
    {
      code: `
var a = 0;
a = a + a;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }],
    },
    {
      code: `
var a = 0;
a += a + 1;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }],
    },
    {
      code: `
var a = 0;
a++;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }],
    },
    {
      code: `
function foo(a) {
  a = a + 1;
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 4 }],
    },
    {
      code: `
function foo(a) {
  a += a + 1;
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 4 }],
    },
    {
      code: `
function foo(a) {
  a++;
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 4 }],
    },
    {
      code: `
var a = 3;
a = a * 5 + 6;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }],
    },
    {
      code: `
var a = 2,
  b = 4;
a = a * 2 + b;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 4, column: 1, endLine: 4, endColumn: 2 }],
    },

    // https://github.com/eslint/eslint/issues/6576 (For coverage)
    {
      code: `
function foo(cb) {
  cb = function (a) {
    cb(1 + a);
  };
  bar(not_cb);
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'cb' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 5 }],
    },
    {
      code: `
function foo(cb) {
  cb = (function (a) {
    return cb(1 + a);
  })();
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'cb' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 5 }],
    },
    {
      code: `
function foo(cb) {
  cb =
    (function (a) {
      cb(1 + a);
    },
    cb);
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'cb' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 5 }],
    },
    {
      code: `
function foo(cb) {
  cb =
    (0,
    function (a) {
      cb(1 + a);
    });
}
foo();
      `,
      errors: [{ messageId: "unusedVar", message: "'cb' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 5 }],
    },

    // https://github.com/eslint/eslint/issues/6646
    {
      code: `
while (a) {
  function foo(b) {
    b = b + 1;
  }
  foo();
}
      `,
      errors: [{ messageId: "unusedVar", message: "'b' is assigned a value but never used.", line: 4, column: 5, endLine: 4, endColumn: 6 }],
    },

    // https://github.com/eslint/eslint/issues/7124
    {
      code: '(function (a, b, c) {});',
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used. Allowed unused args must match /c/u.", line: 1, column: 12, endLine: 1, endColumn: 13 }, { messageId: "unusedVar", message: "'b' is defined but never used. Allowed unused args must match /c/u.", line: 1, column: 15, endLine: 1, endColumn: 16 }],
      options: [{ argsIgnorePattern: 'c' }],
    },
    {
      code: '(function (a, b, { c, d }) {});',
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used. Allowed unused args must match /[cd]/u.", line: 1, column: 12, endLine: 1, endColumn: 13 }, { messageId: "unusedVar", message: "'b' is defined but never used. Allowed unused args must match /[cd]/u.", line: 1, column: 15, endLine: 1, endColumn: 16 }],
      languageOptions: { ecmaVersion: 6 },
      options: [{ argsIgnorePattern: '[cd]' }],
    },
    {
      code: '(function (a, b, { c, d }) {});',
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used. Allowed unused args must match /c/u.", line: 1, column: 12, endLine: 1, endColumn: 13 }, { messageId: "unusedVar", message: "'b' is defined but never used. Allowed unused args must match /c/u.", line: 1, column: 15, endLine: 1, endColumn: 16 }, { messageId: "unusedVar", message: "'d' is defined but never used. Allowed unused args must match /c/u.", line: 1, column: 23, endLine: 1, endColumn: 24 }],
      languageOptions: { ecmaVersion: 6 },
      options: [{ argsIgnorePattern: 'c' }],
    },
    {
      code: '(function (a, b, { c, d }) {});',
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used. Allowed unused args must match /d/u.", line: 1, column: 12, endLine: 1, endColumn: 13 }, { messageId: "unusedVar", message: "'b' is defined but never used. Allowed unused args must match /d/u.", line: 1, column: 15, endLine: 1, endColumn: 16 }, { messageId: "unusedVar", message: "'c' is defined but never used. Allowed unused args must match /d/u.", line: 1, column: 20, endLine: 1, endColumn: 21 }],
      languageOptions: { ecmaVersion: 6 },
      options: [{ argsIgnorePattern: 'd' }],
    },
    {
      code: `
/*global
foo*/
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 3, column: 1, endLine: 3, endColumn: 4 }],
    },

    // https://github.com/eslint/eslint/issues/8442
    {
      code: `
(function ({ a }, b) {
  return b;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 2, column: 14, endLine: 2, endColumn: 15 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
(function ({ a }, { b, c }) {
  return b;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 2, column: 14, endLine: 2, endColumn: 15 }, { messageId: "unusedVar", message: "'c' is defined but never used.", line: 2, column: 24, endLine: 2, endColumn: 25 }],
      languageOptions: { ecmaVersion: 2015 },
    },

    // https://github.com/eslint/eslint/issues/14325
    {
      code: `
let x = 0;
(x++, (x = 0));
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 8, endLine: 3, endColumn: 9 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
(x++, (x = 0));
x = 3;
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 4, column: 1, endLine: 4, endColumn: 2 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
(x++, 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 2, endLine: 3, endColumn: 3 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
(0, x++);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 5, endLine: 3, endColumn: 6 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
(0, (1, x++));
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 9, endLine: 3, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
foo = (x++, 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 8, endLine: 3, endColumn: 9 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
foo = ((0, x++), 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 12, endLine: 3, endColumn: 13 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
((x += 1), 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 4 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
(0, (x += 1));
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 6, endLine: 3, endColumn: 7 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
(0, (1, (x += 1)));
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 10, endLine: 3, endColumn: 11 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
foo = ((x += 1), 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 9, endLine: 3, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let x = 0;
foo = ((0, (x += 1)), 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 13, endLine: 3, endColumn: 14 }],
      languageOptions: { ecmaVersion: 2015 },
    },

    // https://github.com/eslint/eslint/issues/14866
    {
      code: `
let z = 0;
((z = z + 1), (z = 2));
      `,
      errors: [{ messageId: "unusedVar", message: "'z' is assigned a value but never used.", line: 3, column: 16, endLine: 3, endColumn: 17 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
let z = 0;
((z = z + 1), (z = 2));
z = 3;
      `,
      errors: [{ messageId: "unusedVar", message: "'z' is assigned a value but never used.", line: 4, column: 1, endLine: 4, endColumn: 2 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
let z = 0;
((z = z + 1), (z = 2));
z = z + 3;
      `,
      errors: [{ messageId: "unusedVar", message: "'z' is assigned a value but never used.", line: 4, column: 1, endLine: 4, endColumn: 2 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
let x = 0;
(0, (x = x + 1));
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 6, endLine: 3, endColumn: 7 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
let x = 0;
((x = x + 1), 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 3, endLine: 3, endColumn: 4 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
let x = 0;
foo = ((0, (x = x + 1)), 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 13, endLine: 3, endColumn: 14 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
let x = 0;
foo = ((x = x + 1), 0);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 9, endLine: 3, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
let x = 0;
(0, (1, (x = x + 1)));
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 10, endLine: 3, endColumn: 11 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
(function ({ a, b }, { c }) {
  return b;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 2, column: 14, endLine: 2, endColumn: 15 }, { messageId: "unusedVar", message: "'c' is defined but never used.", line: 2, column: 24, endLine: 2, endColumn: 25 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
(function ([a], b) {
  return b;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 2, column: 13, endLine: 2, endColumn: 14 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
(function ([a], [b, c]) {
  return b;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 2, column: 13, endLine: 2, endColumn: 14 }, { messageId: "unusedVar", message: "'c' is defined but never used.", line: 2, column: 21, endLine: 2, endColumn: 22 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
(function ([a, b], [c]) {
  return b;
})();
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is defined but never used.", line: 2, column: 13, endLine: 2, endColumn: 14 }, { messageId: "unusedVar", message: "'c' is defined but never used.", line: 2, column: 21, endLine: 2, endColumn: 22 }],
      languageOptions: { ecmaVersion: 2015 },
    },

    // https://github.com/eslint/eslint/issues/9774
    {
      code: '(function (_a) {})();',
      errors: [{ messageId: "unusedVar", message: "'_a' is defined but never used.", line: 1, column: 12, endLine: 1, endColumn: 14 }],
      options: [{ args: 'all', varsIgnorePattern: '^_' }],
    },
    {
      code: '(function (_a) {})();',
      errors: [{ messageId: "unusedVar", message: "'_a' is defined but never used.", line: 1, column: 12, endLine: 1, endColumn: 14 }],
      options: [{ args: 'all', caughtErrorsIgnorePattern: '^_' }],
    },

    // https://github.com/eslint/eslint/issues/10982
    {
      code: `
var a = function () {
  a();
};
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 2, column: 5, endLine: 2, endColumn: 6 }],
    },
    {
      code: `
var a = function () {
  return function () {
    a();
  };
};
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 2, column: 5, endLine: 2, endColumn: 6 }],
    },
    {
      code: `
const a = () => () => {
  a();
};
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 2, column: 7, endLine: 2, endColumn: 8 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
let myArray = [1, 2, 3, 4].filter(x => x == 0);
myArray = myArray.filter(x => x == 1);
      `,
      errors: [{ messageId: "unusedVar", message: "'myArray' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 8 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
const a = 1;
a += 1;
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    {
      code: `
const a = () => {
  a();
};
      `,
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 2, column: 7, endLine: 2, endColumn: 8 }],
      languageOptions: { ecmaVersion: 2015 },
    },

    // https://github.com/eslint/eslint/issues/14324
    {
      code: `
let x = [];
x = x.concat(x);
      `,
      errors: [{ messageId: "unusedVar", message: "'x' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }],
      languageOptions: { ecmaVersion: 2015 },
    },
    // Multiple errors
    /* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
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
      errors: [{ messageId: "unusedVar", message: "'a' is assigned a value but never used.", line: 3, column: 1, endLine: 3, endColumn: 2 }, { messageId: "unusedVar", message: "'foo' is defined but never used.", line: 4, column: 10, endLine: 4, endColumn: 13 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    /* eslint-enable @typescript-eslint/internal/no-multiple-lines-of-errors */
    {
      code: `
let foo;
init();
foo = foo + 2;
function init() {
  foo = 1;
}
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is assigned a value but never used.", line: 4, column: 1, endLine: 4, endColumn: 4 }],
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: `
function foo(n) {
  if (n < 2) return 1;
  return n * foo(n - 1);
}
      `,
      errors: [{ messageId: "unusedVar", message: "'foo' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 13 }],
      languageOptions: { ecmaVersion: 2020 },
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
      errors: [{ messageId: "unusedVar", message: "'c' is assigned a value but never used.", line: 11, column: 1, endLine: 11, endColumn: 2 }],
      languageOptions: { ecmaVersion: 2020 },
    },

    // ignore class with static initialization block https://github.com/eslint/eslint/issues/17772
    {
      code: `
class Foo {
  static {}
}
      `,
      errors: [{ messageId: "unusedVar", message: "'Foo' is defined but never used.", line: 2, column: 7, endLine: 2, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2022 },
      options: [{ ignoreClassWithStaticInitBlock: false }],
    },
    {
      code: `
class Foo {
  static {}
}
      `,
      errors: [{ messageId: "unusedVar", message: "'Foo' is defined but never used.", line: 2, column: 7, endLine: 2, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2022 },
    },
    {
      code: `
class Foo {
  static {
    var bar;
  }
}
      `,
      errors: [{ messageId: "unusedVar", message: "'bar' is defined but never used.", line: 4, column: 9, endLine: 4, endColumn: 12 }],
      languageOptions: { ecmaVersion: 2022 },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    {
      code: 'class Foo {}',
      errors: [{ messageId: "unusedVar", message: "'Foo' is defined but never used.", line: 1, column: 7, endLine: 1, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2022 },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    {
      code: `
class Foo {
  static bar;
}
      `,
      errors: [{ messageId: "unusedVar", message: "'Foo' is defined but never used.", line: 2, column: 7, endLine: 2, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2022 },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    {
      code: `
class Foo {
  static bar() {}
}
      `,
      errors: [{ messageId: "unusedVar", message: "'Foo' is defined but never used.", line: 2, column: 7, endLine: 2, endColumn: 10 }],
      languageOptions: { ecmaVersion: 2022 },
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },

    // https://github.com/eslint/eslint/issues/17568
    {
      code: `
const _a = 5;
const _b = _a + 5;
      `,
      errors: [{ messageId: "usedIgnoredVar", message: "'_a' is marked as ignored but is used. Used vars must not match /^_/u.", line: 2, column: 7, endLine: 2, endColumn: 9 }],
      languageOptions: { ecmaVersion: 6 },
      options: [
        { args: 'all', reportUsedIgnorePattern: true, varsIgnorePattern: '^_' },
      ],
    },
    {
      code: `
const _a = 42;
foo(() => _a);
      `,
      errors: [{ messageId: "usedIgnoredVar", message: "'_a' is marked as ignored but is used. Used vars must not match /^_/u.", line: 2, column: 7, endLine: 2, endColumn: 9 }],
      languageOptions: { ecmaVersion: 6 },
      options: [
        { args: 'all', reportUsedIgnorePattern: true, varsIgnorePattern: '^_' },
      ],
    },
    {
      code: `
(function foo(_a) {
  return _a + 5;
})(5);
      `,
      errors: [{ messageId: "usedIgnoredVar", message: "'_a' is marked as ignored but is used. Used args must not match /^_/u.", line: 2, column: 15, endLine: 2, endColumn: 17 }],
      options: [
        { args: 'all', argsIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
    },
    {
      code: `
const [a, _b] = items;
console.log(a + _b);
      `,
      errors: [{ messageId: "usedIgnoredVar", message: "'_b' is marked as ignored but is used. Used elements of array destructuring must not match /^_/u.", line: 2, column: 11, endLine: 2, endColumn: 13 }],
      languageOptions: { ecmaVersion: 6 },
      options: [
        { destructuredArrayIgnorePattern: '^_', reportUsedIgnorePattern: true },
      ],
    },
    {
      code: `
let _x;
[_x] = arr;
foo(_x);
      `,
      errors: [{ messageId: "usedIgnoredVar", message: "'_x' is marked as ignored but is used. Used elements of array destructuring must not match /^_/u.", line: 3, column: 2, endLine: 3, endColumn: 4 }],
      languageOptions: { ecmaVersion: 6 },
      options: [
        {
          destructuredArrayIgnorePattern: '^_',
          reportUsedIgnorePattern: true,
          varsIgnorePattern: '[iI]gnored',
        },
      ],
    },
    {
      code: `
const [ignored] = arr;
foo(ignored);
      `,
      errors: [{ messageId: "usedIgnoredVar", message: "'ignored' is marked as ignored but is used. Used vars must not match /[iI]gnored/u.", line: 2, column: 8, endLine: 2, endColumn: 15 }],
      languageOptions: { ecmaVersion: 6 },
      options: [
        {
          destructuredArrayIgnorePattern: '^_',
          reportUsedIgnorePattern: true,
          varsIgnorePattern: '[iI]gnored',
        },
      ],
    },
    {
      code: `
try {
} catch (_err) {
  console.error(_err);
}
      `,
      errors: [{ messageId: "usedIgnoredVar", message: "'_err' is marked as ignored but is used. Used caught errors must not match /^_/u.", line: 3, column: 10, endLine: 3, endColumn: 14 }],
      options: [
        {
          caughtErrors: 'all',
          caughtErrorsIgnorePattern: '^_',
          reportUsedIgnorePattern: true,
        },
      ],
    },
    {
      code: `
try {
} catch (_) {
  _ = 'foo';
}
      `,
      errors: [{ messageId: "unusedVar", message: "'_' is assigned a value but never used. Allowed unused caught errors must match /foo/u.", line: 4, column: 3, endLine: 4, endColumn: 4 }],
      options: [{ caughtErrorsIgnorePattern: 'foo' }],
    },
    {
      code: `
try {
} catch (_) {
  _ = 'foo';
}
      `,
      errors: [{ messageId: "unusedVar", message: "'_' is assigned a value but never used. Allowed unused caught errors must match /ignored/u.", line: 4, column: 3, endLine: 4, endColumn: 4 }],
      options: [
        { caughtErrorsIgnorePattern: 'ignored', varsIgnorePattern: '_' },
      ],
    },
    {
      code: `
try {
} catch ({ message, errors: [firstError] }) {}
      `,
      errors: [{ messageId: "unusedVar", message: "'message' is defined but never used. Allowed unused caught errors must match /foo/u.", line: 3, column: 12, endLine: 3, endColumn: 19 }, { messageId: "unusedVar", message: "'firstError' is defined but never used. Allowed unused caught errors must match /foo/u.", line: 3, column: 30, endLine: 3, endColumn: 40 }],
      options: [{ caughtErrorsIgnorePattern: 'foo' }],
    },
    {
      code: `
try {
} catch ({ stack: $ }) {
  $ = 'Something broke: ' + $;
}
      `,
      errors: [{ messageId: "unusedVar", message: "'$' is assigned a value but never used. Allowed unused caught errors must match /\\w/u.", line: 4, column: 3, endLine: 4, endColumn: 4 }],
      options: [{ caughtErrorsIgnorePattern: '\\w' }],
    },
    {
      code: `
_ => {
  _ = _ + 1;
};
      `,
      errors: [{ messageId: "unusedVar", message: "'_' is assigned a value but never used. Allowed unused args must match /ignored/u.", line: 3, column: 3, endLine: 3, endColumn: 4 }],
      options: [{ argsIgnorePattern: 'ignored', varsIgnorePattern: '_' }],
    },
  ],
});
