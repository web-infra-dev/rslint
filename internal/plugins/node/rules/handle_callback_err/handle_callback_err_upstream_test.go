package handle_callback_err_test

// cspell:ignore rror

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/handle_callback_err"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All tests and examples from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/handle-callback-err.js

func TestHandleCallbackErrUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &handle_callback_err.HandleCallbackErrRule,
		[]rule_tester.ValidTestCase{
			// upstream valid 1.
			{Code: `function test(error) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 2.
			{Code: `function test(err) {console.log(err);}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 3.
			{Code: `function test(err, data) {if(err){ data = 'ERROR';}}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 4.
			{Code: `var test = function(err) {console.log(err);};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 5.
			{Code: `var test = function(err) {if(err){/* do nothing */}};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 6.
			{Code: `var test = function(err) {if(!err){doSomethingHere();}else{};}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 7.
			{Code: `var test = function(err, data) {if(!err) { good(); } else { bad(); }}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 8.
			{Code: `try { } catch(err) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 9.
			{Code: `getData(function(err, data) {if (err) {}getMoreDataWith(data, function(err, moreData) {if (err) {}getEvenMoreDataWith(moreData, function(err, allOfTheThings) {if (err) {}});});});`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 10.
			{Code: `var test = function(err) {if(! err){doSomethingHere();}};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 11.
			{Code: `function test(err, data) {if (data) {doSomething(function(err) {console.error(err);});} else if (err) {console.log(err);}}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 12.
			{Code: `function handler(err, data) {if (data) {doSomethingWith(data);} else if (err) {console.log(err);}}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 13.
			{Code: `function handler(err) {logThisAction(function(err) {if (err) {}}); console.log(err);}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 14.
			{Code: `function userHandler(err) {process.nextTick(function() {if (err) {}})}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 15.
			{Code: `function help() { function userHandler(err) {function tester() { err; process.nextTick(function() { err; }); } } }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 16.
			{Code: `function help(done) { var err = new Error('error'); done(); }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 17.
			{Code: `var test = err => err;`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 18.
			{Code: `var test = err => !err;`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 19.
			{Code: `var test = err => err.message;`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 20.
			{Code: `var test = function(error) {if(error){/* do nothing */}};`, Options: []any{`error`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 21.
			{Code: `var test = (error) => {if(error){/* do nothing */}};`, Options: []any{`error`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 22.
			{Code: `var test = function(error) {if(! error){doSomethingHere();}};`, Options: []any{`error`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 23.
			{Code: `var test = function(err) { console.log(err); };`, Options: []any{`^(err|error)$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 24.
			{Code: `var test = function(error) { console.log(error); };`, Options: []any{`^(err|error)$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 25.
			{Code: `var test = function(anyError) { console.log(anyError); };`, Options: []any{`^.+Error$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 26.
			{Code: `var test = function(any_error) { console.log(anyError); };`, Options: []any{`^.+Error$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// upstream valid 27.
			{Code: `var test = function(any_error) { console.log(any_error); };`, Options: []any{`^.+(e|E)rror$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
		},
		[]rule_tester.InvalidTestCase{
			// upstream invalid 1.
			{Code: `function test(err) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 22)}},
			// upstream invalid 2.
			{Code: `function test(err, data) {}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 28)}},
			// upstream invalid 3.
			{Code: `function test(err) {errorLookingWord();}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 41)}},
			// upstream invalid 4.
			{Code: `function test(err) {try{} catch(err) {}}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 41)}},
			// upstream invalid 5.
			{Code: `function test(err, callback) { foo(function(err, callback) {}); }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 66), expectedAt(1, 36, 1, 62)}},
			// upstream invalid 6.
			{Code: `var test = (err) => {};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 12, 1, 23)}},
			// upstream invalid 7.
			{Code: `var test = function(err) {};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 12, 1, 28)}},
			// upstream invalid 8.
			{Code: `var test = function test(err, data) {};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 12, 1, 39)}},
			// upstream invalid 9.
			{Code: `var test = function test(err) {/* if(err){} */};`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 12, 1, 48)}},
			// upstream invalid 10.
			{Code: `function test(err) {doSomethingHere(function(err){console.log(err);})}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 71)}},
			// upstream invalid 11.
			{Code: `function test(error) {}`, Options: []any{`error`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 24)}},
			// upstream invalid 12.
			{Code: `getData(function(err, data) {getMoreDataWith(data, function(err, moreData) {if (err) {}getEvenMoreDataWith(moreData, function(err, allOfTheThings) {if (err) {}});}); });`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 9, 1, 168)}},
			// upstream invalid 13.
			{Code: `getData(function(err, data) {getMoreDataWith(data, function(err, moreData) {getEvenMoreDataWith(moreData, function(err, allOfTheThings) {if (err) {}});}); });`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 9, 1, 157), expectedAt(1, 52, 1, 153)}},
			// upstream invalid 14.
			{Code: `function userHandler(err) {logThisAction(function(err) {if (err) { console.log(err); } })}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 91)}},
			// upstream invalid 15.
			{Code: `function help() { function userHandler(err) {function tester(err) { err; process.nextTick(function() { err; }); } } }`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 19, 1, 116)}},
			// upstream invalid 16.
			{Code: `var test = function(anyError) { console.log(otherError); };`, Options: []any{`^.+Error$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 12, 1, 59)}},
			// upstream invalid 17.
			{Code: `var test = function(anyError) { };`, Options: []any{`^.+Error$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 12, 1, 34)}},
			// upstream invalid 18.
			{Code: `var test = function(err) { console.log(error); };`, Options: []any{`^(err|error)$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 12, 1, 49)}},
		},
	)
}

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/handle-callback-err.md
// Rule-enabling comments are omitted; the tester enables the rule.

func TestHandleCallbackErrDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &handle_callback_err.HandleCallbackErrRule,
		[]rule_tester.ValidTestCase{
			// documentation example 3.
			{Code: `function loadData (err, data) {
    if (err) {
        console.log(err.stack);
    }
    doSomething();
}

function generateError (err) {
    if (err) {}
}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
			// documentation example 4.
			{Code: `function loadData (error, data) {
    if (error) {
       console.log(error.stack);
    }
    doSomething();
}`, Options: []any{`error`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json"},
		},
		[]rule_tester.InvalidTestCase{
			// documentation example 1.
			{Code: `function loadData (err, data) {
    doSomething(); // forgot to handle error
}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 3, 2)}},
			// documentation example 2.
			{Code: `function loadData (err, data) {
    doSomething();
}`, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 3, 2)}},
			// documented pattern ^(err|error|anySpecificError)$.
			{Code: `function callback0(err) {}
function callback1(error) {}
function callback2(anySpecificError) {}`, Options: []any{`^(err|error|anySpecificError)$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 27), expectedAt(2, 1, 2, 29), expectedAt(3, 1, 3, 40)}},
			// documented pattern ^.+Error$.
			{Code: `function callback0(connectionError) {}
function callback1(validationError) {}`, Options: []any{`^.+Error$`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 39), expectedAt(2, 1, 2, 39)}},
			// documented pattern ^.*(e|E)rr.
			{Code: `function callback0(err) {}
function callback1(error) {}
function callback2(anyError) {}
function callback3(some_err) {}`, Options: []any{`^.*(e|E)rr`}, FileName: `input.js`, TSConfig: "tsconfig.allowJs.json", Errors: []rule_tester.InvalidTestCaseError{expectedAt(1, 1, 1, 27), expectedAt(2, 1, 2, 29), expectedAt(3, 1, 3, 32), expectedAt(4, 1, 4, 32)}},
		},
	)
}

func expectedAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "expected", Message: "Expected error to be handled.", Line: line, Column: column, EndLine: endLine, EndColumn: endColumn}
}
