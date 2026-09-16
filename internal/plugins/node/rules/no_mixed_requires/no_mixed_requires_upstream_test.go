package no_mixed_requires_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_mixed_requires"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func noMixedRequiresError(id string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	message := "Do not mix 'require' and other declarations."
	if id == "noMixCoreModuleFileComputed" {
		message = "Do not mix core, module, file and computed requires."
	}
	return rule_tester.InvalidTestCaseError{MessageId: id, Message: message, Line: line, Column: column, EndLine: endLine, EndColumn: endColumn}
}

func runNoMixedRequiresTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	for i := range valid {
		if valid[i].FileName == "" {
			valid[i].FileName = "input.cjs"
		}
		if valid[i].LanguageOptions.SourceType == "" {
			valid[i].LanguageOptions.SourceType = "commonjs"
		}
		valid[i].TSConfig = "tsconfig.allowJs.json"
	}
	for i := range invalid {
		if invalid[i].FileName == "" {
			invalid[i].FileName = "input.cjs"
		}
		if invalid[i].LanguageOptions.SourceType == "" {
			invalid[i].LanguageOptions.SourceType = "commonjs"
		}
		invalid[i].TSConfig = "tsconfig.allowJs.json"
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allowJs.json", t, &no_mixed_requires.NoMixedRequiresRule, valid, invalid)
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-mixed-requires.js
func TestNoMixedRequiresUpstream(t *testing.T) {
	runNoMixedRequiresTests(t,
		[]rule_tester.ValidTestCase{
			{Code: "var a, b = 42, c = doStuff()", Options: []any{false}},
			{Code: "var a = require(42), b = require(), c = require('y'), d = require(doStuff())", Options: []any{false}},
			{Code: "var fs = require('fs'), foo = require('foo')", Options: []any{false}},
			{Code: "var exec = require('child_process').exec, foo = require('foo')", Options: []any{false}},
			{Code: "var fs = require('fs'), foo = require('./foo')", Options: []any{false}},
			{Code: "var foo = require('foo'), foo2 = require('./foo')", Options: []any{false}},
			{Code: "var emitter = require('events').EventEmitter, fs = require('fs')", Options: []any{false}},
			{Code: "var foo = require(42), bar = require(getName())", Options: []any{false}},
			{Code: "var foo = require(42), bar = require(getName())", Options: []any{true}},
			{Code: "var fs = require('fs'), foo = require('./foo')", Options: []any{map[string]any{"grouping": false}}},
			{Code: "var foo = require('foo'), bar = require(getName())", Options: []any{false}},
			{Code: "var a;", Options: []any{true}},
			{Code: "var async = require('async'), debug = require('diagnostics')('my-module')", Options: []any{map[string]any{"allowCall": true}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "var fs = require('fs'), foo = 42", Options: []any{false}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 33)}},
			{Code: "var fs = require('fs'), foo", Options: []any{false}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 28)}},
			{Code: "var a = require(42), b = require(), c = require('y'), d = require(doStuff())", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 77)}},
			{Code: "var fs = require('fs'), foo = require('foo')", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 45)}},
			{Code: "var fs = require('fs'), foo = require('foo')", Options: []any{map[string]any{"grouping": true}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 45)}},
			{Code: "var exec = require('child_process').exec, foo = require('foo')", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 63)}},
			{Code: "var fs = require('fs'), foo = require('./foo')", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 47)}},
			{Code: "var foo = require('foo'), foo2 = require('./foo')", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 50)}},
			{Code: "var foo = require('foo'), bar = require(getName())", Options: []any{true}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 1, 1, 1, 51)}},
			{Code: "var async = require('async'), debug = require('diagnostics').someFun('my-module')", Options: []any{map[string]any{"allowCall": true}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 1, 82)}},
		},
	)
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-mixed-requires.md
func TestNoMixedRequiresDocumentation(t *testing.T) {
	runNoMixedRequiresTests(t,
		[]rule_tester.ValidTestCase{
			{Code: `// only require declarations (grouping off)
var eventEmitter = require('events').EventEmitter,
    myUtils = require('./utils'),
    util = require('util'),
    bar = require(getBarModuleName());

// only non-require declarations
var foo = 42,
    bar = 'baz';

// always valid regardless of grouping because all declarations are of the same type
var foo = require('foo' + VERSION),
    bar = require(getBarModuleName()),
    baz = require();`},
			{Code: `var async = require('async'),
    debug = require('diagnostics')('my-module'),
    eslint = require('eslint');`, Options: []any{map[string]any{"allowCall": true}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `var fs = require('fs'),        // "core"     \
    async = require('async'),  // "module"   |- these are "require declaration"s
    foo = require('./foo'),    // "file"     |
    bar = require(getName()),  // "computed" /
    baz = 42,                  // "other"
    bam;                       // "uninitialized"`, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 6, 9)}},
			{Code: `var fs = require('fs'),
    i = 0;

var async = require('async'),
    debug = require('diagnostics').someFunction('my-module'),
    eslint = require('eslint');`, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 2, 11), noMixedRequiresError("noMixRequire", 4, 1, 6, 32)}},
			{Code: `// invalid because of mixed types "core" and "module"
var fs = require('fs'),
    async = require('async');

// invalid because of mixed types "file" and "unknown"
var foo = require('foo'),
    bar = require(getBarModuleName());`, Options: []any{map[string]any{"grouping": true}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixCoreModuleFileComputed", 2, 1, 3, 30), noMixedRequiresError("noMixCoreModuleFileComputed", 6, 1, 7, 39)}},
			{Code: `var async = require('async'),
    debug = require('diagnostics').someFunction('my-module'), /* allowCall doesn't allow calling any function */
    eslint = require('eslint');`, Options: []any{map[string]any{"allowCall": true}}, Errors: []rule_tester.InvalidTestCaseError{noMixedRequiresError("noMixRequire", 1, 1, 3, 32)}},
		},
	)
}
