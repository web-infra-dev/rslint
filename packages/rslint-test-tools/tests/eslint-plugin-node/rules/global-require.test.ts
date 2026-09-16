import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester({
  languageOptions: { sourceType: 'commonjs' },
});

// Every case from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/global-require.js
// Documentation examples at the same tag are included below.
ruleTester.run(
  'global-require',
  {},
  {
    valid: [
      {
        name: "var x = require('y');",
        code: "var x = require('y');",
      },
      {
        name: "if (x) { x.require('y'); }",
        code: "if (x) { x.require('y'); }",
      },
      {
        name: 'var x;',
        code: `var x;
x = require('y');`,
      },
      {
        name: "var x = 1, y = require('y');",
        code: "var x = 1, y = require('y');",
      },
      {
        name: "var x = require('y'), y = require('y'), z = require('z');",
        code: "var x = require('y'), y = require('y'), z = require('z');",
      },
      {
        name: "var x = require('y').foo;",
        code: "var x = require('y').foo;",
      },
      {
        name: "require('y').foo();",
        code: "require('y').foo();",
      },
      {
        name: "require('y');",
        code: "require('y');",
      },
      {
        name: 'function x(){}',
        code: `function x(){}


x();


if (x > y) {
	doSomething()

}

var x = require('y').foo;`,
      },
      {
        name: "var logger = require(DEBUG ? 'dev-logger' : 'logger');",
        code: "var logger = require(DEBUG ? 'dev-logger' : 'logger');",
      },
      {
        name: "var logger = DEBUG ? require('dev-logger') : require('logger');",
        code: "var logger = DEBUG ? require('dev-logger') : require('logger');",
      },
      {
        name: "function localScopedRequire(require) { require('y'); }",
        code: "function localScopedRequire(require) { require('y'); }",
      },
      {
        name: "var someFunc = require('./someFunc'); someFunc(function(require) { return('bananas'); });",
        code: "var someFunc = require('./someFunc'); someFunc(function(require) { return('bananas'); });",
      },
      // Documentation examples.
      {
        name: 'var fs = require("fs");',
        code: 'var fs = require("fs");',
      },
      {
        name: '// all these variations of require() are ok',
        code: `// all these variations of require() are ok
require('x');
var y = require('y');
var z;
z = require('z').initialize();

// requiring a module and using it in a function is ok
var fs = require('fs');
function readFile(filename, callback) {
    fs.readFile(filename, callback)
}

// you can use a ternary to determine which module to require
var logger = DEBUG ? require('dev-logger') : require('logger');

// if you want you can require() at the end of your module
function doSomethingA() {}
function doSomethingB() {}
var x = require("x"),
    z = require("z");`,
      },
    ],
    invalid: [
      // Block statements.
      {
        name: "if (process.env.NODE_ENV === 'DEVELOPMENT') {",
        code: `if (process.env.NODE_ENV === 'DEVELOPMENT') {
	require('debug');
}`,
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 18,
          },
        ],
      },
      {
        name: "var x; if (y) { x = require('debug'); }",
        code: "var x; if (y) { x = require('debug'); }",
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: "var x; if (y) { x = require('debug').baz; }",
        code: "var x; if (y) { x = require('debug').baz; }",
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: "function x() { require('y') }",
        code: "function x() { require('y') }",
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: "try { require('x'); } catch (e) { console.log(e); }",
        code: "try { require('x'); } catch (e) { console.log(e); }",
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      // Non-block statements.
      {
        name: 'var getModule = x => require(x);',
        code: 'var getModule = x => require(x);',
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        name: "var x = (x => require(x))('weird')",
        code: "var x = (x => require(x))('weird')",
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 25,
          },
        ],
      },
      {
        name: "switch(x) { case '1': require('1'); break; }",
        code: "switch(x) { case '1': require('1'); break; }",
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 35,
          },
        ],
      },
      // Documentation examples.
      {
        name: 'function foo() {',
        code: `function foo() {

    if (condition) {
        var fs = require("fs");
    }
}`,
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 4,
            column: 18,
            endLine: 4,
            endColumn: 31,
          },
        ],
      },
      {
        name: '// calling require() inside of a function is not allowed',
        code: `// calling require() inside of a function is not allowed
function readFile(filename, callback) {
    var fs = require('fs');
    fs.readFile(filename, callback)
}

// conditional requires like this are also not allowed
if (DEBUG) { require('debug'); }

// a require() in a switch statement is also flagged
switch(x) { case '1': require('1'); break; }

// you may not require() inside an arrow function body
var getModule = (name) => require(name);

// you may not require() inside of a function body as well
function getModule(name) { return require(name); }

// you may not require() inside of a try/catch block
try {
    require(unsafeModule);
} catch(e) {
    console.log(e);
}`,
        errors: [
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 3,
            column: 14,
            endLine: 3,
            endColumn: 27,
          },
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 8,
            column: 14,
            endLine: 8,
            endColumn: 30,
          },
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 11,
            column: 23,
            endLine: 11,
            endColumn: 35,
          },
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 14,
            column: 27,
            endLine: 14,
            endColumn: 40,
          },
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 17,
            column: 35,
            endLine: 17,
            endColumn: 48,
          },
          {
            messageId: 'unexpected',
            message: 'Unexpected require().',
            line: 21,
            column: 5,
            endLine: 21,
            endColumn: 26,
          },
        ],
      },
    ],
  },
);
