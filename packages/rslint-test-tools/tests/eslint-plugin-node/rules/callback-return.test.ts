import { RuleTester } from '../rule-tester';

// Every case and documentation example from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/callback-return.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/callback-return.md
// Rule-enabling comments are omitted; the tester enables node/callback-return.
const ruleTester = new RuleTester();
ruleTester.run(
  'callback-return',
  {},
  {
    valid: [
      // callbacks inside of functions should return
      {
        name: 'callbacks inside of functions should return',
        code: 'function a(err) { if (err) return callback (err); }',
      },
      {
        name: 'upstream valid 2',
        code: 'function a(err) { if (err) return callback (err); callback(); }',
      },
      {
        name: 'upstream valid 3',
        code: 'function a(err) { if (err) { return callback (err); } callback(); }',
      },
      {
        name: 'upstream valid 4',
        code: 'function a(err) { if (err) { return /* confusing comment */ callback (err); } callback(); }',
      },
      {
        name: 'upstream valid 5',
        code: 'function x(err) { if (err) { callback(); return; } }',
      },
      {
        name: 'upstream valid 6',
        code: 'function x(err) { if (err) { \n log();\n callback(); return; } }',
      },
      {
        name: 'upstream valid 7',
        code: 'function x(err) { if (err) { callback(); return; } return callback(); }',
      },
      {
        name: 'upstream valid 8',
        code: 'function x(err) { if (err) { return callback(); } else { return callback(); } }',
      },
      {
        name: 'upstream valid 9',
        code: 'function x(err) { if (err) { return callback(); } else if (x) { return callback(); } }',
      },
      {
        name: 'upstream valid 10',
        code: 'function x(err) { if (err) return callback(); else return callback(); }',
      },
      { name: 'upstream valid 11', code: 'function x(cb) { cb && cb(); }' },
      {
        name: 'upstream valid 12',
        code: "function x(next) { typeof next !== 'undefined' && next(); }",
      },
      {
        name: 'upstream valid 13',
        code: "function x(next) { if (typeof next === 'function')  { return next() } }",
      },
      {
        name: 'upstream valid 14',
        code: "function x() { switch(x) { case 'a': return next(); } }",
      },
      {
        name: 'upstream valid 15',
        code: 'function x() { for(x = 0; x < 10; x++) { return next(); } }',
      },
      {
        name: 'upstream valid 16',
        code: 'function x() { while(x) { return next(); } }',
      },
      {
        name: 'upstream valid 17',
        code: 'function a(err) { if (err) { obj.method (err); } }',
      },
      // callback() all you want outside of a function
      {
        name: 'callback() all you want outside of a function',
        code: 'callback()',
      },
      { name: 'upstream valid 19', code: 'callback(); callback();' },
      { name: 'upstream valid 20', code: 'while(x) { move(); }' },
      {
        name: 'upstream valid 21',
        code: 'for (var i = 0; i < 10; i++) { move(); }',
      },
      {
        name: 'upstream valid 22',
        code: 'for (var i = 0; i < 10; i++) move();',
      },
      { name: 'upstream valid 23', code: 'if (x) callback();' },
      { name: 'upstream valid 24', code: 'if (x) { callback(); }' },
      // arrow functions
      {
        name: 'arrow functions',
        code: 'var x = err => { if (err) { callback(); return; } }',
      },
      { name: 'upstream valid 26', code: 'var x = err => callback(err)' },
      {
        name: 'upstream valid 27',
        code: 'var x = err => { setTimeout( () => { callback(); }); }',
      },
      // classes
      { name: 'classes', code: 'class x { horse() { callback(); } } ' },
      {
        name: 'upstream valid 29',
        code: 'class x { horse() { if (err) { return callback(); } callback(); } } ',
      },
      // options (only warns with the correct callback name)
      {
        name: 'options (only warns with the correct callback name)',
        code: 'function a(err) { if (err) { callback(err) } }',
        options: [['cb']],
      },
      {
        name: 'upstream valid 31',
        code: 'function a(err) { if (err) { callback(err) } next(); }',
        options: [['cb', 'next']],
      },
      {
        name: 'upstream valid 32',
        code: 'function a(err) { if (err) { return next(err) } else { callback(); } }',
        options: [['cb', 'next']],
      },
      // allow object methods (https://github.com/eslint/eslint/issues/4711)
      {
        name: 'allow object methods (https://github.com/eslint/eslint/issues/4711)',
        code: 'function a(err) { if (err) { return obj.method(err); } }',
        options: [['obj.method']],
      },
      {
        name: 'upstream valid 34',
        code: 'function a(err) { if (err) { return obj.prop.method(err); } }',
        options: [['obj.prop.method']],
      },
      {
        name: 'upstream valid 35',
        code: 'function a(err) { if (err) { return obj.prop.method(err); } otherObj.prop.method() }',
        options: [['obj.prop.method', 'otherObj.prop.method']],
      },
      {
        name: 'upstream valid 36',
        code: 'function a(err) { if (err) { callback(err); } }',
        options: [['obj.method']],
      },
      {
        name: 'upstream valid 37',
        code: 'function a(err) { if (err) { otherObj.method(err); } }',
        options: [['obj.method']],
      },
      {
        name: 'upstream valid 38',
        code: 'function a(err) { if (err) { //comment\nreturn obj.method(err); } }',
        options: [['obj.method']],
      },
      {
        name: 'upstream valid 39',
        code: 'function a(err) { if (err) { /*comment*/return obj.method(err); } }',
        options: [['obj.method']],
      },
      {
        name: 'upstream valid 40',
        code: 'function a(err) { if (err) { return obj.method(err); //comment\n } }',
        options: [['obj.method']],
      },
      {
        name: 'upstream valid 41',
        code: 'function a(err) { if (err) { return obj.method(err); /*comment*/ } }',
        options: [['obj.method']],
      },
      // only warns if object of MemberExpression is an Identifier
      {
        name: 'only warns if object of MemberExpression is an Identifier',
        code: 'function a(err) { if (err) { obj().method(err); } }',
        options: [['obj().method']],
      },
      {
        name: 'upstream valid 43',
        code: 'function a(err) { if (err) { obj.prop().method(err); } }',
        options: [['obj.prop().method']],
      },
      {
        name: 'upstream valid 44',
        code: 'function a(err) { if (err) { obj().prop.method(err); } }',
        options: [['obj().prop.method']],
      },
      // does not warn if object of MemberExpression is invoked
      {
        name: 'does not warn if object of MemberExpression is invoked',
        code: 'function a(err) { if (err) { obj().method(err); } }',
        options: [['obj.method']],
      },
      {
        name: 'upstream valid 46',
        code: 'function a(err) { if (err) { obj().method(err); } obj.method(); }',
        options: [['obj.method']],
      },
      // known bad examples that we know we are ignoring
      {
        name: 'known bad examples that we know we are ignoring',
        code: 'function x(err) { if (err) { setTimeout(callback, 0); } callback(); }',
      },
      // callback() called twice
      {
        name: 'callback() called twice',
        code: 'function x(err) { if (err) { process.nextTick(function(err) { callback(); }); } callback(); }',
      },
      // Documentation: introduction
      {
        name: 'Documentation: introduction',
        code: 'function doSomething(err, callback) {\n    if (err) {\n        return callback(err);\n    }\n    callback();\n}',
      },
      // Documentation: default names: correct
      {
        name: 'Documentation: default names: correct',
        code: 'function foo(err, callback) {\n    if (err) {\n        return callback(err);\n    }\n    callback();\n}',
      },
      // Documentation: supplied names: correct
      {
        name: 'Documentation: supplied names: correct',
        code: 'function foo(err, done) {\n    if (err) {\n        return done(err);\n    }\n    done();\n}\n\nfunction bar(err, send) {\n    if (err) {\n        return send.error(err);\n    }\n    send.success();\n}',
        options: [['done', 'send.error', 'send.success']],
      },
      // Documentation: callback passed by reference
      {
        name: 'Documentation: callback passed by reference',
        code: 'function foo(err, callback) {\n    if (err) {\n        setTimeout(callback, 0); // this is bad, but WILL NOT warn\n    }\n    callback();\n}',
      },
      // Documentation: callback in a nested function
      {
        name: 'Documentation: callback in a nested function',
        code: 'function foo(err, callback) {\n    if (err) {\n        process.nextTick(function() {\n            return callback(); // this is bad, but WILL NOT warn\n        });\n    }\n    callback();\n}',
      },
    ],
    invalid: [
      {
        name: 'upstream invalid 1',
        code: 'function a(err) { if (err) { callback (err); } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'upstream invalid 2',
        code: "function a(callback) { if (typeof callback !== 'undefined') { callback(); } }",
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 63,
            endLine: 1,
            endColumn: 73,
          },
        ],
      },
      {
        name: 'upstream invalid 3',
        code: "function a(callback) { if (typeof callback !== 'undefined') callback();  }",
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 61,
            endLine: 1,
            endColumn: 71,
          },
        ],
      },
      {
        name: 'upstream invalid 4',
        code: 'function a(callback) { if (err) { callback(); horse && horse(); } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 35,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'upstream invalid 5',
        code: 'var x = (err) => { if (err) { callback (err); } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 31,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'upstream invalid 6',
        code: 'var x = { x(err) { if (err) { callback (err); } } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 31,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'upstream invalid 7',
        code: 'function x(err) { if (err) {\n log();\n callback(err); } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 15,
          },
        ],
      },
      {
        name: 'upstream invalid 8',
        code: 'var x = { x(err) { if (err) { callback && callback (err); } } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 43,
            endLine: 1,
            endColumn: 57,
          },
        ],
      },
      {
        name: 'upstream invalid 9',
        code: 'function a(err) { callback (err); callback(); }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'upstream invalid 10',
        code: 'function a(err) { callback (err); horse(); }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'upstream invalid 11',
        code: 'function a(err) { if (err) { callback (err); horse(); return; } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'upstream invalid 12',
        code: 'var a = (err) => { callback (err); callback(); }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 34,
          },
        ],
      },
      {
        name: 'upstream invalid 13',
        code: 'function a(err) { if (err) { callback (err); } else if (x) { callback(err); return; } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'upstream invalid 14',
        code: 'function x(err) { if (err) { return callback(); }\nelse if (abc) {\ncallback(); }\nelse {\nreturn callback(); } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 3,
            column: 1,
            endLine: 3,
            endColumn: 11,
          },
        ],
      },
      {
        name: 'upstream invalid 15',
        code: 'class x { horse() { if (err) { callback(); } callback(); } } ',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
      // generally good behavior which we must not allow to keep the rule simple
      {
        name: 'generally good behavior which we must not allow to keep the rule simple',
        code: 'function x(err) { if (err) { callback() } else { callback() } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 40,
          },
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 50,
            endLine: 1,
            endColumn: 60,
          },
        ],
      },
      {
        name: 'upstream invalid 17',
        code: 'function x(err) { if (err) return callback(); else callback(); }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 52,
            endLine: 1,
            endColumn: 62,
          },
        ],
      },
      {
        name: 'upstream invalid 18',
        code: '() => { if (x) { callback(); } }',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: 'upstream invalid 19',
        code: "function b() { switch(x) { case 'horse': callback(); } }",
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 42,
            endLine: 1,
            endColumn: 52,
          },
        ],
      },
      {
        name: 'upstream invalid 20',
        code: "function a() { switch(x) { case 'horse': move(); } }",
        options: [['move']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 42,
            endLine: 1,
            endColumn: 48,
          },
        ],
      },
      // loops
      {
        name: 'loops',
        code: 'var x = function() { while(x) { move(); } }',
        options: [['move']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        name: 'upstream invalid 22',
        code: 'function x() { for (var i = 0; i < 10; i++) { move(); } }',
        options: [['move']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 47,
            endLine: 1,
            endColumn: 53,
          },
        ],
      },
      {
        name: 'upstream invalid 23',
        code: 'var x = function() { for (var i = 0; i < 10; i++) move(); }',
        options: [['move']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 51,
            endLine: 1,
            endColumn: 57,
          },
        ],
      },
      {
        name: 'upstream invalid 24',
        code: 'function a(err) { if (err) { obj.method(err); } }',
        options: [['obj.method']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'upstream invalid 25',
        code: 'function a(err) { if (err) { obj.prop.method(err); } }',
        options: [['obj.prop.method']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 50,
          },
        ],
      },
      {
        name: 'upstream invalid 26',
        code: 'function a(err) { if (err) { obj.prop.method(err); } otherObj.prop.method() }',
        options: [['obj.prop.method', 'otherObj.prop.method']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 50,
          },
        ],
      },
      {
        name: 'upstream invalid 27',
        code: 'function a(err) { if (err) { /*comment*/obj.method(err); } }',
        options: [['obj.method']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 41,
            endLine: 1,
            endColumn: 56,
          },
        ],
      },
      {
        name: 'upstream invalid 28',
        code: 'function a(err) { if (err) { //comment\nobj.method(err); } }',
        options: [['obj.method']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'upstream invalid 29',
        code: 'function a(err) { if (err) { obj.method(err); /*comment*/ } }',
        options: [['obj.method']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      {
        name: 'upstream invalid 30',
        code: 'function a(err) { if (err) { obj.method(err); //comment\n } }',
        options: [['obj.method']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 1,
            column: 30,
            endLine: 1,
            endColumn: 45,
          },
        ],
      },
      // Documentation: default names: incorrect
      {
        name: 'Documentation: default names: incorrect',
        code: 'function foo(err, callback) {\n    if (err) {\n        callback(err);\n    }\n    callback();\n}',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 3,
            column: 9,
            endLine: 3,
            endColumn: 22,
          },
        ],
      },
      // Documentation: supplied names: incorrect
      {
        name: 'Documentation: supplied names: incorrect',
        code: 'function foo(err, done) {\n    if (err) {\n        done(err);\n    }\n    done();\n}\n\nfunction bar(err, send) {\n    if (err) {\n        send.error(err);\n    }\n    send.success();\n}',
        options: [['done', 'send.error', 'send.success']],
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 3,
            column: 9,
            endLine: 3,
            endColumn: 18,
          },
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 10,
            column: 9,
            endLine: 10,
            endColumn: 24,
          },
        ],
      },
      // Documentation: if/else limitation
      {
        name: 'Documentation: if/else limitation',
        code: 'function foo(err, callback) {\n    if (err) {\n        callback(err); // this is fine, but WILL warn\n    } else {\n        callback();    // this is fine, but WILL warn\n    }\n}',
        errors: [
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 3,
            column: 9,
            endLine: 3,
            endColumn: 22,
          },
          {
            messageId: 'missingReturn',
            message: 'Expected return with your callback function.',
            line: 5,
            column: 9,
            endLine: 5,
            endColumn: 19,
          },
        ],
      },
    ],
  },
);
