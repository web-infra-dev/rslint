package callback_return_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/callback_return"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func missingReturnAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "missingReturn",
		Message:   "Expected return with your callback function.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/callback-return.js
func TestCallbackReturnUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &callback_return.CallbackReturnRule,
		[]rule_tester.ValidTestCase{
			// callbacks inside of functions should return
			{Code: `function a(err) { if (err) return callback (err); }`},
			{Code: `function a(err) { if (err) return callback (err); callback(); }`},
			{Code: `function a(err) { if (err) { return callback (err); } callback(); }`},
			{Code: `function a(err) { if (err) { return /* confusing comment */ callback (err); } callback(); }`},
			{Code: `function x(err) { if (err) { callback(); return; } }`},
			{Code: "function x(err) { if (err) { \n log();\n callback(); return; } }"},
			{Code: `function x(err) { if (err) { callback(); return; } return callback(); }`},
			{Code: `function x(err) { if (err) { return callback(); } else { return callback(); } }`},
			{Code: `function x(err) { if (err) { return callback(); } else if (x) { return callback(); } }`},
			{Code: `function x(err) { if (err) return callback(); else return callback(); }`},
			{Code: `function x(cb) { cb && cb(); }`},
			{Code: `function x(next) { typeof next !== 'undefined' && next(); }`},
			{Code: `function x(next) { if (typeof next === 'function')  { return next() } }`},
			{Code: `function x() { switch(x) { case 'a': return next(); } }`},
			{Code: `function x() { for(x = 0; x < 10; x++) { return next(); } }`},
			{Code: `function x() { while(x) { return next(); } }`},
			{Code: `function a(err) { if (err) { obj.method (err); } }`},
			// callback() all you want outside of a function
			{Code: `callback()`},
			{Code: `callback(); callback();`},
			{Code: `while(x) { move(); }`},
			{Code: `for (var i = 0; i < 10; i++) { move(); }`},
			{Code: `for (var i = 0; i < 10; i++) move();`},
			{Code: `if (x) callback();`},
			{Code: `if (x) { callback(); }`},
			// arrow functions
			{Code: `var x = err => { if (err) { callback(); return; } }`},
			{Code: `var x = err => callback(err)`},
			{Code: `var x = err => { setTimeout( () => { callback(); }); }`},
			// classes
			{Code: `class x { horse() { callback(); } } `},
			{Code: `class x { horse() { if (err) { return callback(); } callback(); } } `},
			// options (only warns with the correct callback name)
			{Code: `function a(err) { if (err) { callback(err) } }`, Options: []any{[]any{"cb"}}},
			{Code: `function a(err) { if (err) { callback(err) } next(); }`, Options: []any{[]any{"cb", "next"}}},
			{Code: `function a(err) { if (err) { return next(err) } else { callback(); } }`, Options: []any{[]any{"cb", "next"}}},
			// allow object methods (https://github.com/eslint/eslint/issues/4711)
			{Code: `function a(err) { if (err) { return obj.method(err); } }`, Options: []any{[]any{"obj.method"}}},
			{Code: `function a(err) { if (err) { return obj.prop.method(err); } }`, Options: []any{[]any{"obj.prop.method"}}},
			{Code: `function a(err) { if (err) { return obj.prop.method(err); } otherObj.prop.method() }`, Options: []any{[]any{"obj.prop.method", "otherObj.prop.method"}}},
			{Code: `function a(err) { if (err) { callback(err); } }`, Options: []any{[]any{"obj.method"}}},
			{Code: `function a(err) { if (err) { otherObj.method(err); } }`, Options: []any{[]any{"obj.method"}}},
			{Code: `function a(err) { if (err) { //comment
return obj.method(err); } }`, Options: []any{[]any{"obj.method"}}},
			{Code: `function a(err) { if (err) { /*comment*/return obj.method(err); } }`, Options: []any{[]any{"obj.method"}}},
			{Code: `function a(err) { if (err) { return obj.method(err); //comment
 } }`, Options: []any{[]any{"obj.method"}}},
			{Code: `function a(err) { if (err) { return obj.method(err); /*comment*/ } }`, Options: []any{[]any{"obj.method"}}},
			// only warns if object of MemberExpression is an Identifier
			{Code: `function a(err) { if (err) { obj().method(err); } }`, Options: []any{[]any{"obj().method"}}},
			{Code: `function a(err) { if (err) { obj.prop().method(err); } }`, Options: []any{[]any{"obj.prop().method"}}},
			{Code: `function a(err) { if (err) { obj().prop.method(err); } }`, Options: []any{[]any{"obj().prop.method"}}},
			// does not warn if object of MemberExpression is invoked
			{Code: `function a(err) { if (err) { obj().method(err); } }`, Options: []any{[]any{"obj.method"}}},
			{Code: `function a(err) { if (err) { obj().method(err); } obj.method(); }`, Options: []any{[]any{"obj.method"}}},
			// known bad examples that we know we are ignoring
			{Code: `function x(err) { if (err) { setTimeout(callback, 0); } callback(); }`},
			// callback() called twice
			{Code: `function x(err) { if (err) { process.nextTick(function(err) { callback(); }); } callback(); }`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `function a(err) { if (err) { callback (err); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 44)}},
			{Code: `function a(callback) { if (typeof callback !== 'undefined') { callback(); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 63, 1, 73)}},
			{Code: `function a(callback) { if (typeof callback !== 'undefined') callback();  }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 61, 1, 71)}},
			{Code: `function a(callback) { if (err) { callback(); horse && horse(); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 35, 1, 45)}},
			{Code: `var x = (err) => { if (err) { callback (err); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 31, 1, 45)}},
			{Code: `var x = { x(err) { if (err) { callback (err); } } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 31, 1, 45)}},
			{Code: `function x(err) { if (err) {
 log();
 callback(err); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(3, 2, 3, 15)}},
			{Code: `var x = { x(err) { if (err) { callback && callback (err); } } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 43, 1, 57)}},
			{Code: `function a(err) { callback (err); callback(); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 19, 1, 33)}},
			{Code: `function a(err) { callback (err); horse(); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 19, 1, 33)}},
			{Code: `function a(err) { if (err) { callback (err); horse(); return; } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 44)}},
			{Code: `var a = (err) => { callback (err); callback(); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 20, 1, 34)}},
			{Code: `function a(err) { if (err) { callback (err); } else if (x) { callback(err); return; } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 44)}},
			{Code: `function x(err) { if (err) { return callback(); }
else if (abc) {
callback(); }
else {
return callback(); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(3, 1, 3, 11)}},
			{Code: `class x { horse() { if (err) { callback(); } callback(); } } `, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 32, 1, 42)}},
			// generally good behavior which we must not allow to keep the rule simple
			{Code: `function x(err) { if (err) { callback() } else { callback() } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 40), missingReturnAt(1, 50, 1, 60)}},
			{Code: `function x(err) { if (err) return callback(); else callback(); }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 52, 1, 62)}},
			{Code: `() => { if (x) { callback(); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 18, 1, 28)}},
			{Code: `function b() { switch(x) { case 'horse': callback(); } }`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 42, 1, 52)}},
			{Code: `function a() { switch(x) { case 'horse': move(); } }`, Options: []any{[]any{"move"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 42, 1, 48)}},
			// loops
			{Code: `var x = function() { while(x) { move(); } }`, Options: []any{[]any{"move"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 33, 1, 39)}},
			{Code: `function x() { for (var i = 0; i < 10; i++) { move(); } }`, Options: []any{[]any{"move"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 47, 1, 53)}},
			{Code: `var x = function() { for (var i = 0; i < 10; i++) move(); }`, Options: []any{[]any{"move"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 51, 1, 57)}},
			{Code: `function a(err) { if (err) { obj.method(err); } }`, Options: []any{[]any{"obj.method"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 45)}},
			{Code: `function a(err) { if (err) { obj.prop.method(err); } }`, Options: []any{[]any{"obj.prop.method"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 50)}},
			{Code: `function a(err) { if (err) { obj.prop.method(err); } otherObj.prop.method() }`, Options: []any{[]any{"obj.prop.method", "otherObj.prop.method"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 50)}},
			{Code: `function a(err) { if (err) { /*comment*/obj.method(err); } }`, Options: []any{[]any{"obj.method"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 41, 1, 56)}},
			{Code: `function a(err) { if (err) { //comment
obj.method(err); } }`, Options: []any{[]any{"obj.method"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(2, 1, 2, 16)}},
			{Code: `function a(err) { if (err) { obj.method(err); /*comment*/ } }`, Options: []any{[]any{"obj.method"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 45)}},
			{Code: `function a(err) { if (err) { obj.method(err); //comment
 } }`, Options: []any{[]any{"obj.method"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(1, 30, 1, 45)}},
		},
	)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/callback-return.md
// Rule-enabling comments are omitted; the tester enables the rule.
func TestCallbackReturnDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &callback_return.CallbackReturnRule,
		[]rule_tester.ValidTestCase{
			// Documentation: introduction
			{Code: `function doSomething(err, callback) {
    if (err) {
        return callback(err);
    }
    callback();
}`},
			// Documentation: default names: correct
			{Code: `function foo(err, callback) {
    if (err) {
        return callback(err);
    }
    callback();
}`},
			// Documentation: supplied names: correct
			{Code: `function foo(err, done) {
    if (err) {
        return done(err);
    }
    done();
}

function bar(err, send) {
    if (err) {
        return send.error(err);
    }
    send.success();
}`, Options: []any{[]any{"done", "send.error", "send.success"}}},
			// Documentation: callback passed by reference
			{Code: `function foo(err, callback) {
    if (err) {
        setTimeout(callback, 0); // this is bad, but WILL NOT warn
    }
    callback();
}`},
			// Documentation: callback in a nested function
			{Code: `function foo(err, callback) {
    if (err) {
        process.nextTick(function() {
            return callback(); // this is bad, but WILL NOT warn
        });
    }
    callback();
}`},
		},
		[]rule_tester.InvalidTestCase{
			// Documentation: default names: incorrect
			{Code: `function foo(err, callback) {
    if (err) {
        callback(err);
    }
    callback();
}`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(3, 9, 3, 22)}},
			// Documentation: supplied names: incorrect
			{Code: `function foo(err, done) {
    if (err) {
        done(err);
    }
    done();
}

function bar(err, send) {
    if (err) {
        send.error(err);
    }
    send.success();
}`, Options: []any{[]any{"done", "send.error", "send.success"}}, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(3, 9, 3, 18), missingReturnAt(10, 9, 10, 24)}},
			// Documentation: if/else limitation
			{Code: `function foo(err, callback) {
    if (err) {
        callback(err); // this is fine, but WILL warn
    } else {
        callback();    // this is fine, but WILL warn
    }
}`, Errors: []rule_tester.InvalidTestCaseError{missingReturnAt(3, 9, 3, 22), missingReturnAt(5, 9, 5, 19)}},
		},
	)
}
