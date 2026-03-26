// The following tests are adapted from the tests in eslint.
// Original Code: https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/tests/lib/rules/no-unused-vars.js
// License      : https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/LICENSE

import { assignedError, definedError, ruleTester } from './eslint-test-helpers';

ruleTester.run('no-unused-vars', {
  invalid: [
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
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'name',
          },
          messageId: 'unusedVar',
        },
      ],
    },
    {
      code: `
(function (obj) {
  var name;
  for (name in obj) {
  }
})({});
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'name',
          },
          messageId: 'unusedVar',
        },
      ],
    },
    {
      code: `
(function (obj) {
  for (var name in obj) {
  }
})({});
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'name',
          },
          messageId: 'unusedVar',
        },
      ],
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
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'name',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
(function (iter) {
  var name;
  for (name of iter) {
  }
})({});
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'name',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
(function (iter) {
  for (var name of iter) {
  }
})({});
      `,
      errors: [
        {
          data: {
            action: 'assigned a value',
            additional: '',
            varName: 'name',
          },
          messageId: 'unusedVar',
        },
      ],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    // https://github.com/eslint/eslint/issues/3617
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: `
/* global foobar, foo, bar */
foobar;
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
        {
          data: {
            action: 'defined',
            additional: '',
            varName: 'bar',
          },
          messageId: 'unusedVar',
        },
      ],
    },
    {
      // skip: uses ESLint /*global*/ comment directive not available in rslint
      skip: true,
      code: `
/* global foobar,
   foo,
   bar
 */
foobar;
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
        {
          data: {
            action: 'defined',
            additional: '',
            varName: 'bar',
          },
          messageId: 'unusedVar',
        },
      ],
    },

    // caughtErrors
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [definedError('err')],
    },
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [definedError('err')],
      options: [{ caughtErrors: 'all' }],
    },
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [
        definedError(
          'err',
          '. Allowed unused caught errors must match /^ignore/u',
        ),
      ],
      options: [{ caughtErrors: 'all', caughtErrorsIgnorePattern: '^ignore' }],
    },
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [definedError('err')],
      options: [{ caughtErrors: 'all', varsIgnorePattern: '^err' }],
    },
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [definedError('err')],
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
      errors: [
        definedError(
          'err',
          '. Allowed unused caught errors must match /^ignore/u',
        ),
      ],
      options: [{ caughtErrors: 'all', caughtErrorsIgnorePattern: '^ignore' }],
    },

    // multiple try catch both fail
    {
      code: `
try {
} catch (error) {}
try {
} catch (err) {}
      `,
      errors: [
        definedError(
          'error',
          '. Allowed unused caught errors must match /^ignore/u',
        ),
        definedError(
          'err',
          '. Allowed unused caught errors must match /^ignore/u',
        ),
      ],
      options: [{ caughtErrors: 'all', caughtErrorsIgnorePattern: '^ignore' }],
    },

    // caughtErrors with other configs
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [definedError('err')],
      options: [{ args: 'all', caughtErrors: 'all', vars: 'all' }],
    },

    // no conflict in ignore patterns
    {
      code: `
try {
} catch (err) {}
      `,
      errors: [definedError('err')],
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
      errors: [assignedError('a')],
    },
    {
      code: `
var a = 0;
a = a + a;
      `,
      errors: [assignedError('a')],
    },
    {
      code: `
var a = 0;
a += a + 1;
      `,
      errors: [assignedError('a')],
    },
    {
      code: `
var a = 0;
a++;
      `,
      errors: [assignedError('a')],
    },
    {
      code: `
function foo(a) {
  a = a + 1;
}
foo();
      `,
      errors: [assignedError('a')],
    },
    {
      code: `
function foo(a) {
  a += a + 1;
}
foo();
      `,
      errors: [assignedError('a')],
    },
    {
      code: `
function foo(a) {
  a++;
}
foo();
      `,
      errors: [assignedError('a')],
    },
    {
      code: `
var a = 3;
a = a * 5 + 6;
      `,
      errors: [assignedError('a')],
    },
    {
      code: `
var a = 2,
  b = 4;
a = a * 2 + b;
      `,
      errors: [assignedError('a')],
    },

    // https://github.com/eslint/eslint/issues/6576 (For coverage)
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
function foo(cb) {
  cb = function (a) {
    cb(1 + a);
  };
  bar(not_cb);
}
foo();
      `,
      errors: [assignedError('cb')],
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
      errors: [assignedError('cb')],
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
      errors: [assignedError('cb')],
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
      errors: [assignedError('cb')],
    },

    // https://github.com/eslint/eslint/issues/6646
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
while (a) {
  function foo(b) {
    b = b + 1;
  }
  foo();
}
      `,
      errors: [assignedError('b')],
    },
  ],
  valid: [
    // for-in loops (see #2342)
    {
      code: `
(function (obj) {
  var name;
  for (name in obj) return;
})({});
      `,
    },
    {
      code: `
(function (obj) {
  var name;
  for (name in obj) {
    return;
  }
})({});
      `,
    },
    {
      code: `
(function (obj) {
  for (var name in obj) {
    return true;
  }
})({});
      `,
    },
    {
      code: `
(function (obj) {
  for (var name in obj) return true;
})({});
      `,
    },

    {
      code: `
(function (obj) {
  let name;
  for (name in obj) return;
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
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
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
(function (obj) {
  for (let name in obj) {
    return true;
  }
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
(function (obj) {
  for (let name in obj) return true;
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    {
      code: `
(function (obj) {
  for (const name in obj) {
    return true;
  }
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
(function (obj) {
  for (const name in obj) return true;
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    // For-of loops
    {
      code: `
(function (iter) {
  let name;
  for (name of iter) return;
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
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
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
(function (iter) {
  for (let name of iter) {
    return true;
  }
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
(function (iter) {
  for (let name of iter) return true;
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    {
      code: `
(function (iter) {
  for (const name of iter) {
    return true;
  }
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
(function (iter) {
  for (const name of iter) return true;
})({});
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },

    // Sequence Expressions (See https://github.com/eslint/eslint/issues/14325)
    {
      code: `
let x = 0;
foo = (0, x++);
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
let x = 0;
foo = (0, (x += 1));
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: `
let x = 0;
foo = (0, (x = x + 1));
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
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
    // skip: $scope is undeclared — incompatible with TypeScript module mode
    {
      skip: true,
      code: `
var unregisterFooWatcher;
// ...
unregisterFooWatcher = $scope.$watch('foo', function () {
  // ...some code..
  unregisterFooWatcher();
});
      `,
    },
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
    // skip: `register` is undeclared — incompatible with TypeScript module mode
    {
      skip: true,
      code: `
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
    },
    {
      code: `
function* foo(cb) {
  cb = yield function (a) {
    cb(1 + a);
  };
}
foo();
      `,
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
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
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
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
    // skip: @typescript-eslint/no-unused-vars considers `a = myFunction(a)` as self-modifying
    // (reports as unused), unlike ESLint core which treats the function argument as a genuine read.
    {
      skip: true,
      code: `
function someFunction() {
  var a = 0,
    i;
  for (i = 0; i < 2; i++) {
    a = myFunction(a);
  }
}
someFunction();
      `,
    },

    // https://github.com/eslint/eslint/issues/17299
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
var a;
a ||= 1;
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2021 } },
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
var a;
a &&= 1;
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2021 } },
    },
    {
      // skip: script-mode code incompatible with TypeScript module mode
      skip: true,
      code: `
var a;
a ??= 1;
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2021 } },
    },
  ],
});
