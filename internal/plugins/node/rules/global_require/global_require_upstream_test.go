package global_require_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/global_require"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func unexpectedAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unexpected", Message: "Unexpected require().",
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/global-require.js
func TestGlobalRequireUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "var x = require('y');"},
		{Code: "if (x) { x.require('y'); }"},
		{Code: `var x;
x = require('y');`},
		{Code: "var x = 1, y = require('y');"},
		{Code: "var x = require('y'), y = require('y'), z = require('z');"},
		{Code: "var x = require('y').foo;"},
		{Code: "require('y').foo();"},
		{Code: "require('y');"},
		{Code: `function x(){}


x();


if (x > y) {
	doSomething()

}

var x = require('y').foo;`},
		{Code: "var logger = require(DEBUG ? 'dev-logger' : 'logger');"},
		{Code: "var logger = DEBUG ? require('dev-logger') : require('logger');"},
		{Code: "function localScopedRequire(require) { require('y'); }"},
		{Code: "var someFunc = require('./someFunc'); someFunc(function(require) { return('bananas'); });"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Block statements.
		{Code: `if (process.env.NODE_ENV === 'DEVELOPMENT') {
	require('debug');
}`, Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(2, 2, 2, 18)}},
		{Code: "var x; if (y) { x = require('debug'); }", Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(1, 21, 1, 37)}},
		{Code: "var x; if (y) { x = require('debug').baz; }", Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(1, 21, 1, 37)}},
		{Code: "function x() { require('y') }", Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(1, 16, 1, 28)}},
		{Code: "try { require('x'); } catch (e) { console.log(e); }", Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(1, 7, 1, 19)}},
		// Non-block statements.
		{Code: "var getModule = x => require(x);", Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(1, 22, 1, 32)}},
		{Code: "var x = (x => require(x))('weird')", Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(1, 15, 1, 25)}},
		{Code: "switch(x) { case '1': require('1'); break; }", Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(1, 23, 1, 35)}},
	}
	// Upstream's shared tester uses ES2015 and CommonJS.
	for i := range valid {
		valid[i].FileName = "input.cjs"
		valid[i].TSConfig = "tsconfig.allowJs.json"
		valid[i].LanguageOptions = rule.LanguageOptions{ECMAVersion: 2015, SourceType: "commonjs"}
	}
	for i := range invalid {
		invalid[i].FileName = "input.cjs"
		invalid[i].TSConfig = "tsconfig.allowJs.json"
		invalid[i].LanguageOptions = rule.LanguageOptions{ECMAVersion: 2015, SourceType: "commonjs"}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &global_require.GlobalRequireRule, valid, invalid)
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/global-require.md
// Rule-enabling comments are omitted; the tester enables the rule.
func TestGlobalRequireDocumentation(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "var fs = require(\"fs\");"},
		{Code: `// all these variations of require() are ok
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
    z = require("z");`},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: `function foo() {

    if (condition) {
        var fs = require("fs");
    }
}`, Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(4, 18, 4, 31)}},
		{Code: `// calling require() inside of a function is not allowed
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
}`, Errors: []rule_tester.InvalidTestCaseError{unexpectedAt(3, 14, 3, 27), unexpectedAt(8, 14, 8, 30), unexpectedAt(11, 23, 11, 35), unexpectedAt(14, 27, 14, 40), unexpectedAt(17, 35, 17, 48), unexpectedAt(21, 5, 21, 26)}},
	}
	// Upstream's shared tester uses ES2015 and CommonJS.
	for i := range valid {
		valid[i].FileName = "input.cjs"
		valid[i].TSConfig = "tsconfig.allowJs.json"
		valid[i].LanguageOptions = rule.LanguageOptions{ECMAVersion: 2015, SourceType: "commonjs"}
	}
	for i := range invalid {
		invalid[i].FileName = "input.cjs"
		invalid[i].TSConfig = "tsconfig.allowJs.json"
		invalid[i].LanguageOptions = rule.LanguageOptions{ECMAVersion: 2015, SourceType: "commonjs"}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &global_require.GlobalRequireRule, valid, invalid)
}
