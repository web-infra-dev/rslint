import { RuleTester } from '../rule-tester';

// All tests and documentation examples from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/handle-callback-err.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/handle-callback-err.md
// Rule-enabling comments are omitted; the tester enables node/handle-callback-err.
new RuleTester().run(
  'handle-callback-err',
  {},
  {
    valid: [
      {
        name: 'upstream valid 1',
        code: 'function test(error) {}',
      },
      {
        name: 'upstream valid 2',
        code: 'function test(err) {console.log(err);}',
      },
      {
        name: 'upstream valid 3',
        code: "function test(err, data) {if(err){ data = 'ERROR';}}",
      },
      {
        name: 'upstream valid 4',
        code: 'var test = function(err) {console.log(err);};',
      },
      {
        name: 'upstream valid 5',
        code: 'var test = function(err) {if(err){/* do nothing */}};',
      },
      {
        name: 'upstream valid 6',
        code: 'var test = function(err) {if(!err){doSomethingHere();}else{};}',
      },
      {
        name: 'upstream valid 7',
        code: 'var test = function(err, data) {if(!err) { good(); } else { bad(); }}',
      },
      {
        name: 'upstream valid 8',
        code: 'try { } catch(err) {}',
      },
      {
        name: 'upstream valid 9',
        code: 'getData(function(err, data) {if (err) {}getMoreDataWith(data, function(err, moreData) {if (err) {}getEvenMoreDataWith(moreData, function(err, allOfTheThings) {if (err) {}});});});',
      },
      {
        name: 'upstream valid 10',
        code: 'var test = function(err) {if(! err){doSomethingHere();}};',
      },
      {
        name: 'upstream valid 11',
        code: 'function test(err, data) {if (data) {doSomething(function(err) {console.error(err);});} else if (err) {console.log(err);}}',
      },
      {
        name: 'upstream valid 12',
        code: 'function handler(err, data) {if (data) {doSomethingWith(data);} else if (err) {console.log(err);}}',
      },
      {
        name: 'upstream valid 13',
        code: 'function handler(err) {logThisAction(function(err) {if (err) {}}); console.log(err);}',
      },
      {
        name: 'upstream valid 14',
        code: 'function userHandler(err) {process.nextTick(function() {if (err) {}})}',
      },
      {
        name: 'upstream valid 15',
        code: 'function help() { function userHandler(err) {function tester() { err; process.nextTick(function() { err; }); } } }',
      },
      {
        name: 'upstream valid 16',
        code: "function help(done) { var err = new Error('error'); done(); }",
      },
      {
        name: 'upstream valid 17',
        code: 'var test = err => err;',
      },
      {
        name: 'upstream valid 18',
        code: 'var test = err => !err;',
      },
      {
        name: 'upstream valid 19',
        code: 'var test = err => err.message;',
      },
      {
        name: 'upstream valid 20',
        code: 'var test = function(error) {if(error){/* do nothing */}};',
        options: ['error'],
      },
      {
        name: 'upstream valid 21',
        code: 'var test = (error) => {if(error){/* do nothing */}};',
        options: ['error'],
      },
      {
        name: 'upstream valid 22',
        code: 'var test = function(error) {if(! error){doSomethingHere();}};',
        options: ['error'],
      },
      {
        name: 'upstream valid 23',
        code: 'var test = function(err) { console.log(err); };',
        options: ['^(err|error)$'],
      },
      {
        name: 'upstream valid 24',
        code: 'var test = function(error) { console.log(error); };',
        options: ['^(err|error)$'],
      },
      {
        name: 'upstream valid 25',
        code: 'var test = function(anyError) { console.log(anyError); };',
        options: ['^.+Error$'],
      },
      {
        name: 'upstream valid 26',
        code: 'var test = function(any_error) { console.log(anyError); };',
        options: ['^.+Error$'],
      },
      {
        name: 'upstream valid 27',
        code: 'var test = function(any_error) { console.log(any_error); };',
        options: ['^.+(e|E)rror$'],
      },
      {
        name: 'documentation example 3',
        code: 'function loadData (err, data) {\n    if (err) {\n        console.log(err.stack);\n    }\n    doSomething();\n}\n\nfunction generateError (err) {\n    if (err) {}\n}',
      },
      {
        name: 'documentation example 4',
        code: 'function loadData (error, data) {\n    if (error) {\n       console.log(error.stack);\n    }\n    doSomething();\n}',
        options: ['error'],
      },
    ],
    invalid: [
      {
        name: 'upstream invalid 1',
        code: 'function test(err) {}',
        errors: [expectedAt(1, 1, 1, 22)],
      },
      {
        name: 'upstream invalid 2',
        code: 'function test(err, data) {}',
        errors: [expectedAt(1, 1, 1, 28)],
      },
      {
        name: 'upstream invalid 3',
        code: 'function test(err) {errorLookingWord();}',
        errors: [expectedAt(1, 1, 1, 41)],
      },
      {
        name: 'upstream invalid 4',
        code: 'function test(err) {try{} catch(err) {}}',
        errors: [expectedAt(1, 1, 1, 41)],
      },
      {
        name: 'upstream invalid 5',
        code: 'function test(err, callback) { foo(function(err, callback) {}); }',
        errors: [expectedAt(1, 1, 1, 66), expectedAt(1, 36, 1, 62)],
      },
      {
        name: 'upstream invalid 6',
        code: 'var test = (err) => {};',
        errors: [expectedAt(1, 12, 1, 23)],
      },
      {
        name: 'upstream invalid 7',
        code: 'var test = function(err) {};',
        errors: [expectedAt(1, 12, 1, 28)],
      },
      {
        name: 'upstream invalid 8',
        code: 'var test = function test(err, data) {};',
        errors: [expectedAt(1, 12, 1, 39)],
      },
      {
        name: 'upstream invalid 9',
        code: 'var test = function test(err) {/* if(err){} */};',
        errors: [expectedAt(1, 12, 1, 48)],
      },
      {
        name: 'upstream invalid 10',
        code: 'function test(err) {doSomethingHere(function(err){console.log(err);})}',
        errors: [expectedAt(1, 1, 1, 71)],
      },
      {
        name: 'upstream invalid 11',
        code: 'function test(error) {}',
        options: ['error'],
        errors: [expectedAt(1, 1, 1, 24)],
      },
      {
        name: 'upstream invalid 12',
        code: 'getData(function(err, data) {getMoreDataWith(data, function(err, moreData) {if (err) {}getEvenMoreDataWith(moreData, function(err, allOfTheThings) {if (err) {}});}); });',
        errors: [expectedAt(1, 9, 1, 168)],
      },
      {
        name: 'upstream invalid 13',
        code: 'getData(function(err, data) {getMoreDataWith(data, function(err, moreData) {getEvenMoreDataWith(moreData, function(err, allOfTheThings) {if (err) {}});}); });',
        errors: [expectedAt(1, 9, 1, 157), expectedAt(1, 52, 1, 153)],
      },
      {
        name: 'upstream invalid 14',
        code: 'function userHandler(err) {logThisAction(function(err) {if (err) { console.log(err); } })}',
        errors: [expectedAt(1, 1, 1, 91)],
      },
      {
        name: 'upstream invalid 15',
        code: 'function help() { function userHandler(err) {function tester(err) { err; process.nextTick(function() { err; }); } } }',
        errors: [expectedAt(1, 19, 1, 116)],
      },
      {
        name: 'upstream invalid 16',
        code: 'var test = function(anyError) { console.log(otherError); };',
        options: ['^.+Error$'],
        errors: [expectedAt(1, 12, 1, 59)],
      },
      {
        name: 'upstream invalid 17',
        code: 'var test = function(anyError) { };',
        options: ['^.+Error$'],
        errors: [expectedAt(1, 12, 1, 34)],
      },
      {
        name: 'upstream invalid 18',
        code: 'var test = function(err) { console.log(error); };',
        options: ['^(err|error)$'],
        errors: [expectedAt(1, 12, 1, 49)],
      },
      {
        name: 'documentation example 1',
        code: 'function loadData (err, data) {\n    doSomething(); // forgot to handle error\n}',
        errors: [expectedAt(1, 1, 3, 2)],
      },
      {
        name: 'documentation example 2',
        code: 'function loadData (err, data) {\n    doSomething();\n}',
        errors: [expectedAt(1, 1, 3, 2)],
      },
      {
        name: 'documented pattern ^(err|error|anySpecificError)$',
        options: ['^(err|error|anySpecificError)$'],
        code: 'function callback0(err) {}\nfunction callback1(error) {}\nfunction callback2(anySpecificError) {}',
        errors: [
          expectedAt(1, 1, 1, 27),
          expectedAt(2, 1, 2, 29),
          expectedAt(3, 1, 3, 40),
        ],
      },
      {
        name: 'documented pattern ^.+Error$',
        options: ['^.+Error$'],
        code: 'function callback0(connectionError) {}\nfunction callback1(validationError) {}',
        errors: [expectedAt(1, 1, 1, 39), expectedAt(2, 1, 2, 39)],
      },
      {
        name: 'documented pattern ^.*(e|E)rr',
        options: ['^.*(e|E)rr'],
        code: 'function callback0(err) {}\nfunction callback1(error) {}\nfunction callback2(anyError) {}\nfunction callback3(some_err) {}',
        errors: [
          expectedAt(1, 1, 1, 27),
          expectedAt(2, 1, 2, 29),
          expectedAt(3, 1, 3, 32),
          expectedAt(4, 1, 4, 32),
        ],
      },
    ],
  },
);

function expectedAt(
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) {
  return {
    messageId: 'expected',
    message: 'Expected error to be handled.',
    line,
    column,
    endLine,
    endColumn,
  };
}
